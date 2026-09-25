package training

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// Depth layers (I5, option A): what KIND of content each chunk is.
//
// The expert's declared depth was derived from how many chunks mention a topic, so
// fourteen introductory definitions scored the same as fourteen failure modes and
// trade-offs. This classifies each existing chunk instead:
//
//	1 = WHAT / WHY        a definition, a purpose, when to use the thing
//	2 = HOW / TRADE-OFFS  mechanics, comparisons, costs and benefits
//	3 = FAILURE / EDGE    what breaks, limitations, debugging, war stories
//
// Option A means NOTHING IS INVENTED. The course already contains whatever depth it
// contains; this only says which kind each passage is. That is why the result is
// trustworthy enough to show a buyer, and why a topic with no layer-3 chunks is a
// finding about the course rather than a guess about the expert.
//
// Classification is on demand and resumable — it costs model calls, so it is not
// hidden inside ingest, and a half-classified corpus reports how much is left instead
// of pretending to be finished.

const (
	// LayerDefinition, LayerMechanics and LayerFailure are the three content kinds.
	LayerDefinition = 1
	LayerMechanics  = 2
	LayerFailure    = 3

	// layerBatchSize is how many passages travel in one classification call. Small
	// enough that each answer is checkable, large enough that a 1600-chunk corpus does
	// not take 1600 calls.
	layerBatchSize = 20

	// layerCallChunkChars caps one passage in the prompt.
	layerCallChunkChars = 700

	// layerMaxChunksPerRun bounds one call to Classify, so a button press is a bounded
	// amount of work and the response can say how much is left. The caller runs it again
	// to continue.
	layerMaxChunksPerRun = 200
)

// DepthLayerName renders a layer for the screen.
func DepthLayerName(layer int) string {
	switch layer {
	case LayerDefinition:
		return "what / why"
	case LayerMechanics:
		return "how / trade-offs"
	case LayerFailure:
		return "failure / edge cases"
	default:
		return "not classified"
	}
}

// DepthClassifyResult reports one classification pass.
type DepthClassifyResult struct {
	Classified int `json:"classified"`
	// Remaining is how many chunks still have no layer, so the caller knows whether to
	// run again rather than guessing.
	Remaining int `json:"remaining"`
	Calls     int `json:"calls"`
	// Rejected counts entries the model produced that named an out-of-range passage or
	// layer. Shown, not hidden: a rising number means the prompt or the model is drifting.
	Rejected int `json:"rejected"`
}

// TopicLayerCoverage is one topic's content kinds.
type TopicLayerCoverage struct {
	Topic      string `json:"topic"`
	Definition int    `json:"definition"`
	Mechanics  int    `json:"mechanics"`
	Failure    int    `json:"failure"`
	Total      int    `json:"total"`
}

// DepthLayerReport answers "is this expert deep?" from the corpus itself.
type DepthLayerReport struct {
	ExpertID     uuid.UUID `json:"expert_id"`
	TotalChunks  int       `json:"total_chunks"`
	Classified   int       `json:"classified"`
	Unclassified int       `json:"unclassified"`
	Definition   int       `json:"definition"`
	Mechanics    int       `json:"mechanics"`
	Failure      int       `json:"failure"`
	// Topics that have no failure-mode content, and none at mechanics level. These are
	// the actionable numbers: they name where the expert can only answer "what is it".
	TopicsWithoutFailure   int                  `json:"topics_without_failure"`
	TopicsWithoutMechanics int                  `json:"topics_without_mechanics"`
	Topics                 []TopicLayerCoverage `json:"topics"`
	Findings               []string             `json:"findings"`
}

// DepthClassifier labels chunks by content kind and reports coverage.
type DepthClassifier struct {
	db      *pgxpool.Pool
	gateway *gateway.ModelGateway
	logger  *zap.Logger
}

// NewDepthClassifier builds a depth classifier.
func NewDepthClassifier(db *pgxpool.Pool, gw *gateway.ModelGateway, logger *zap.Logger) *DepthClassifier {
	return &DepthClassifier{db: db, gateway: gw, logger: logger}
}

// Classify labels up to layerMaxChunksPerRun unclassified chunks, then reports how many
// are still left. Resumable by design: a second call continues where this one stopped.
func (c *DepthClassifier) Classify(ctx context.Context, expertID uuid.UUID) (DepthClassifyResult, error) {
	var result DepthClassifyResult

	type pending struct {
		ID   uuid.UUID
		Text string
	}
	rows, err := c.db.Query(ctx, `
		SELECT id, chunk_text
		  FROM course_chunks
		 WHERE expert_id = $1 AND layer IS NULL
		 ORDER BY chunk_index
		 LIMIT $2`, expertID, layerMaxChunksPerRun)
	if err != nil {
		return result, fmt.Errorf("depth layers: read unclassified chunks: %w", err)
	}
	var todo []pending
	for rows.Next() {
		var p pending
		if scanErr := rows.Scan(&p.ID, &p.Text); scanErr != nil {
			continue
		}
		todo = append(todo, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("depth layers: read unclassified chunks: %w", err)
	}

	for start := 0; start < len(todo); start += layerBatchSize {
		end := start + layerBatchSize
		if end > len(todo) {
			end = len(todo)
		}
		batch := todo[start:end]

		texts := make([]string, 0, len(batch))
		for _, p := range batch {
			texts = append(texts, p.Text)
		}

		response, callErr := c.requestLayers(ctx, texts)
		result.Calls++
		if callErr != nil {
			// One failed batch must not discard the batches that worked.
			c.logger.Warn("depth layers: classification call failed",
				zap.String("expert_id", expertID.String()),
				zap.Error(callErr))
			continue
		}

		layers, rejected, parseErr := parseDepthLayers(response, len(batch))
		if parseErr != nil {
			c.logger.Warn("depth layers: unparseable response",
				zap.String("expert_id", expertID.String()),
				zap.Error(parseErr))
			continue
		}
		result.Rejected += rejected

		for index, layer := range layers {
			if _, err := c.db.Exec(ctx,
				`UPDATE course_chunks SET layer = $1 WHERE id = $2`, layer, batch[index].ID); err != nil {
				return result, fmt.Errorf("depth layers: store layer: %w", err)
			}
			result.Classified++
		}
	}

	if err := c.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM course_chunks WHERE expert_id = $1 AND layer IS NULL`, expertID,
	).Scan(&result.Remaining); err != nil {
		return result, fmt.Errorf("depth layers: count remaining: %w", err)
	}

	c.logger.Info("depth layers classified",
		zap.String("expert_id", expertID.String()),
		zap.Int("classified", result.Classified),
		zap.Int("remaining", result.Remaining),
		zap.Int("calls", result.Calls),
		zap.Int("rejected", result.Rejected),
	)
	return result, nil
}

// requestLayers asks the model what kind of content each passage is.
func (c *DepthClassifier) requestLayers(ctx context.Context, texts []string) (string, error) {
	var sb strings.Builder
	sb.WriteString("For each numbered passage from a technical course, decide what KIND of content it is.\n\n")
	sb.WriteString(`1 = WHAT / WHY       — explains what something is, why it matters, or when to use it
2 = HOW / TRADE-OFFS — explains how it works, compares options, or weighs costs and benefits
3 = FAILURE / EDGE   — explains what breaks, limitations, debugging, or a real incident

If a passage mixes kinds, choose the dominant one. Judge only the passage.

Passages:
`)
	for i, text := range texts {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, clipPromptText(strings.ReplaceAll(text, "\n", " "), layerCallChunkChars)))
	}
	sb.WriteString(`
Return ONLY JSON, one entry per passage:
{"layers":[{"n":1,"layer":2}]}`)

	resp, err := c.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  sb.String(),
		MaxTokens:   2048,
		Temperature: 0,
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// parseDepthLayers reads the model's answer into 0-based indexes.
//
// Pure — unit-tested. Entries naming a passage outside the batch are DROPPED, not
// clamped: a layer applied to the wrong chunk is worse than a chunk left unclassified,
// because an unclassified chunk is visibly pending while a mislabelled one is silently
// wrong for as long as the corpus lives.
func parseDepthLayers(response string, count int) (map[int]int, int, error) {
	body := extractJSONObject(response)
	if body == "" {
		return nil, 0, fmt.Errorf("no JSON object in response")
	}

	var payload struct {
		Layers []struct {
			N     int `json:"n"`
			Layer int `json:"layer"`
		} `json:"layers"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return nil, 0, err
	}

	rejected := 0
	out := make(map[int]int, len(payload.Layers))
	for _, entry := range payload.Layers {
		index := entry.N - 1
		if index < 0 || index >= count {
			rejected++
			continue
		}
		if entry.Layer < LayerDefinition || entry.Layer > LayerFailure {
			rejected++
			continue
		}
		if _, dup := out[index]; dup {
			// The model repeating itself is not an invention; keep the first answer.
			continue
		}
		out[index] = entry.Layer
	}
	return out, rejected, nil
}

// Report reads the corpus's content-kind coverage.
func (c *DepthClassifier) Report(ctx context.Context, expertID uuid.UUID) (*DepthLayerReport, error) {
	report := &DepthLayerReport{ExpertID: expertID, Topics: []TopicLayerCoverage{}, Findings: []string{}}

	if err := c.db.QueryRow(ctx, `
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE layer IS NULL),
		       COUNT(*) FILTER (WHERE layer = $2),
		       COUNT(*) FILTER (WHERE layer = $3),
		       COUNT(*) FILTER (WHERE layer = $4)
		  FROM course_chunks
		 WHERE expert_id = $1`, expertID, LayerDefinition, LayerMechanics, LayerFailure,
	).Scan(&report.TotalChunks, &report.Unclassified, &report.Definition, &report.Mechanics, &report.Failure); err != nil {
		return nil, fmt.Errorf("depth layers: read coverage: %w", err)
	}
	report.Classified = report.TotalChunks - report.Unclassified

	rows, err := c.db.Query(ctx, `
		SELECT topic,
		       COUNT(*) FILTER (WHERE layer = $2),
		       COUNT(*) FILTER (WHERE layer = $3),
		       COUNT(*) FILTER (WHERE layer = $4),
		       COUNT(*)
		  FROM course_chunks
		 WHERE expert_id = $1
		   AND topic IS NOT NULL AND topic <> '' AND topic <> 'general'
		 GROUP BY topic
		 ORDER BY COUNT(*) DESC`, expertID, LayerDefinition, LayerMechanics, LayerFailure)
	if err != nil {
		return nil, fmt.Errorf("depth layers: read topic coverage: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t TopicLayerCoverage
		if scanErr := rows.Scan(&t.Topic, &t.Definition, &t.Mechanics, &t.Failure, &t.Total); scanErr != nil {
			continue
		}
		if t.Failure == 0 {
			report.TopicsWithoutFailure++
		}
		if t.Mechanics == 0 {
			report.TopicsWithoutMechanics++
		}
		report.Topics = append(report.Topics, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("depth layers: read topic coverage: %w", err)
	}

	report.Findings = layerFindings(report)
	return report, nil
}

// layerFindings turns the coverage numbers into sentences, in the order worth acting on.
//
// Pure — unit-tested. The point of the whole feature is that these are statements about
// the COURSE ("this topic has no failure-mode content"), not a score someone has to
// interpret.
func layerFindings(r *DepthLayerReport) []string {
	findings := make([]string, 0, 4)

	if r.TotalChunks == 0 {
		return []string{"No chunks stored for this expert yet."}
	}
	if r.Unclassified > 0 {
		findings = append(findings, fmt.Sprintf(
			"%d of %d chunks are not classified yet — run the classification to see the full picture.",
			r.Unclassified, r.TotalChunks))
	}
	if r.Classified == 0 {
		return findings
	}

	if r.TopicsWithoutFailure > 0 {
		findings = append(findings, fmt.Sprintf(
			"%d topic(s) have NO failure-mode or edge-case content — the expert will explain them but cannot answer \"what breaks under load\".",
			r.TopicsWithoutFailure))
	}
	if r.TopicsWithoutMechanics > 0 {
		findings = append(findings, fmt.Sprintf(
			"%d topic(s) have no mechanics content — they will answer \"what is it\" but not \"how does it work\".",
			r.TopicsWithoutMechanics))
	}

	definitionShare := r.Definition * 100 / r.Classified
	if definitionShare >= 60 {
		findings = append(findings, fmt.Sprintf(
			"%d%% of classified content is definitions (what/why). The corpus is broad but shallow: mostly explaining what things are.",
			definitionShare))
	}
	if r.Failure > 0 && r.Failure*100/r.Classified <= 5 {
		findings = append(findings, fmt.Sprintf(
			"Only %d%% of content covers failure modes and edge cases — the part of a course that makes an expert useful in an incident.",
			r.Failure*100/r.Classified))
	}
	if len(findings) == 0 {
		findings = append(findings, "Content kinds look balanced across the corpus.")
	}
	return findings
}
