package training

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/gateway"
)

// Capability evaluation (I2): measure what an expert can actually answer.
//
// WHY this exists: an expert's capability used to be DECLARED. depth_level came
// from how many chunks mention a topic, and can_handle / cannot_handle came from
// an LLM that saw about 600 characters per topic. The row that answers "what does
// this expert know?" was therefore a guess, and Gate 2 reads it.
//
// The opposite approach costs more and cannot be faked: write real questions from
// the expert's own corpus, ask them, and record what happened. Three separate
// things are measured per case, because they fail differently and need different
// fixes:
//
//	retrieval  did the production retrieval path surface the chunk that answers
//	           the question? (hit@k, plus the rank, so MRR-style trends are
//	           available from stored data)
//	grounding  did the answer cite a retrieved chunk, and is it supported by the
//	           context it was given? (deterministic citation check + an LLM judge)
//	refusal    did the expert decline? A refusal on an answerable question is a
//	           capability failure; a refusal on an unanswerable one is correct
//	           behaviour, which is why adversarial cases are a separate suite.
//
// The verdict is written back onto expert_capabilities, so every existing reader
// sees a measurement instead of a claim without changing its query.

// Answer depths. These are the levels a question is written at, and the scale
// measured_level uses. Deliberately NOT the 1-5 depth_level scale: that one counts
// chunk coverage, this one counts what the expert can do.
const (
	CapabilityLevelDefinition = 1 // what it is, when to use it
	CapabilityLevelMechanics  = 2 // how it works, trade-offs
	CapabilityLevelFailure    = 3 // what breaks, edge cases, failure modes
)

// FailureReason values. Stored per case so a topic report can say WHAT went wrong
// rather than only that something did — an actionable eval, not a score.
const (
	FailureRetrievalMissed = "retrieval_missed" // expected chunk not in top-k
	FailureNotCited        = "not_cited"        // answer used no context marker
	FailureUngrounded      = "ungrounded"       // judge says not supported by context
	FailureRefused         = "refused"          // expert declined an answerable question
	FailureEmpty           = "empty"            // no usable answer text
)

// insufficientContextMarker is what the answering prompt must return when the
// context does not contain the answer. A marker is used rather than trusting the
// model's phrasing, so "did it refuse?" is a string comparison and not a judgement.
const insufficientContextMarker = "INSUFFICIENT_CONTEXT"

const (
	// defaultEvalTopics bounds one pass. WHY bounded: each topic costs one
	// generation call plus two calls per case, so an unbounded pass over 118 topics
	// would be ~800 LLM calls in a single admin action. The budget is recorded on
	// the run row, so a score is always read next to the sample it came from.
	defaultEvalTopics = 10

	// defaultEvalTopK mirrors the retrieval depth the workflow path uses, so the
	// measurement reflects production rather than a privately chosen k.
	defaultEvalTopK = 5

	// casesPerTopic: one question per level, each bound to a different chunk so a
	// topic's report is not derived from a single passage.
	casesPerTopic = 3

	// promptPassageChars caps how much of a chunk enters a prompt. Chunks run
	// 375-600 tokens; this keeps three passages plus instructions inside a small
	// model's comfortable window.
	promptPassageChars = 900

	// capabilityPassRateThreshold is how much of a level must pass for that level
	// to count as achieved. Majority, so one unlucky case does not hide a topic
	// that the expert genuinely handles.
	capabilityPassRateThreshold = 0.5

	// Mirror the existing can/cannot-handle caps on expert_capabilities.
	canHandleLimit    = 8
	cannotHandleLimit = 5
)

// CapabilityRetriever is the retrieval this evaluator measures.
//
// WHY an interface with exactly the production method: the measurement is only
// meaningful if it exercises the same path the chat does. Wiring the real
// assembler means a change to that path breaks this interface at compile time
// instead of silently leaving the eval measuring an older, simpler retrieval. It
// also lets the scoring be tested without a database or a model.
type CapabilityRetriever interface {
	GetCourseChunksForWorkflow(ctx context.Context, expertID uuid.UUID, taskDescription string, topK int) ([]chinawall.CourseChunk, error)
	// GetCourseChunksExpanded is the same retrieval with concept-graph expansion. It is
	// a separate method so a pass can measure each path, and so the difference between
	// two runs can be attributed to expansion alone.
	GetCourseChunksExpanded(ctx context.Context, expertID uuid.UUID, taskDescription string, topK int) ([]chinawall.CourseChunk, error)
}

// CapabilityEvaluator generates capability questions from a corpus, asks them, and
// records the evidence.
type CapabilityEvaluator struct {
	db        *pgxpool.Pool
	gateway   *gateway.ModelGateway
	retriever CapabilityRetriever
	answerer  *groundedAnswerer
	logger    *zap.Logger
}

// NewCapabilityEvaluator builds an evaluator. retriever may be nil, in which case
// RunEval fails loudly rather than reporting a retrieval score it never measured.
func NewCapabilityEvaluator(db *pgxpool.Pool, gw *gateway.ModelGateway, retriever CapabilityRetriever, logger *zap.Logger) *CapabilityEvaluator {
	return &CapabilityEvaluator{
		db: db, gateway: gw, retriever: retriever,
		answerer: newGroundedAnswerer(gw, logger), logger: logger,
	}
}

// CapabilityEvalRequest is one evaluation pass's budget.
type CapabilityEvalRequest struct {
	// Topics caps how many topics are evaluated, strongest coverage first.
	Topics int
	// TopK is the retrieval depth scored against.
	TopK int
	// Regenerate replaces the stored question set before running. WHY opt-in: the
	// question set is the golden set, and silently regenerating it would destroy
	// the baseline that makes two runs comparable.
	Regenerate bool

	// GraphExpansion retrieves with concept-graph expansion. Recorded on the run, and
	// comparisons are only made against a run with the SAME value — otherwise the
	// difference would be attributed to whichever change was made last.
	GraphExpansion bool
}

// CapabilityCase is one stored question with its ground truth.
type CapabilityCase struct {
	ID                 uuid.UUID
	Topic              string
	Level              int
	Question           string
	ExpectedChunkID    uuid.UUID
	ExpectedSourceFile string
}

// CapabilityCaseResult is the evidence for one case in one run.
type CapabilityCaseResult struct {
	CaseID        uuid.UUID
	Question      string
	Topic         string
	Level         int
	RetrievedIDs  []uuid.UUID
	HitRank       int
	Cited         bool
	Grounded      string
	Refused       bool
	Passed        bool
	Answer        string
	FailureReason string
}

// TopicCapabilityReport is one topic's measured verdict.
type TopicCapabilityReport struct {
	Topic          string   `json:"topic"`
	CoverageChunks int      `json:"coverage_chunks"`
	Cases          int      `json:"cases"`
	Passed         int      `json:"passed"`
	MeasuredLevel  int      `json:"measured_level"`
	FailureReasons []string `json:"failure_reasons"`
	CanHandle      []string `json:"can_handle"`
	CannotHandle   []string `json:"cannot_handle"`
}

// RetrievalMetrics is how well retrieval did across one run's questions.
//
// WHY this exists: changing retrieval without measuring it is how a "fix" becomes a
// regression nobody notices. The per-case evidence was already stored, so these
// numbers are derived from it rather than recorded separately — one source, no drift.
type RetrievalMetrics struct {
	Cases int     `json:"cases"`
	Hits  int     `json:"hits"`
	MRR   float64 `json:"mrr"`
}

// HitRate is the share of questions whose source chunk was retrieved in the top k.
func (m RetrievalMetrics) HitRate() float64 {
	if m.Cases == 0 {
		return 0
	}
	return float64(m.Hits) / float64(m.Cases)
}

// CapabilityEvalReport is a whole pass, ready for the admin screen.
type CapabilityEvalReport struct {
	RunID         uuid.UUID               `json:"run_id"`
	ExpertID      uuid.UUID               `json:"expert_id"`
	Status        string                  `json:"status"`
	TopicsTotal   int                     `json:"topics_total"`
	CasesTotal    int                     `json:"cases_total"`
	CasesPassed   int                     `json:"cases_passed"`
	RetrievalHits int                     `json:"retrieval_hits"`
	Grounded      int                     `json:"grounded"`
	Refused       int                     `json:"refused"`
	TopK          int                     `json:"top_k"`
	StartedAt     time.Time               `json:"started_at"`
	CompletedAt   *time.Time              `json:"completed_at"`
	TopicReports  []TopicCapabilityReport `json:"topics"`
	Findings      []string                `json:"findings"`

	// Metrics for this run, and the previous run's, so a retrieval change can be
	// proven or rolled back instead of believed.
	Metrics  RetrievalMetrics  `json:"metrics"`
	Previous *RetrievalMetrics `json:"previous_metrics,omitempty"`
	// GraphExpansion records which retrieval path this pass used. The baseline above
	// is only ever the previous pass with the same value.
	GraphExpansion bool `json:"graph_expansion"`
}

// RunEval performs one pass: ensure the question set, then ask every question and
// record the evidence.
//
// It is designed to run in the background (the caller spawns a goroutine), because
// a pass costs one generation call per topic plus two calls per case and would
// outlive any HTTP request.
func (e *CapabilityEvaluator) RunEval(ctx context.Context, expertID uuid.UUID, req CapabilityEvalRequest) (uuid.UUID, error) {
	if e.retriever == nil {
		return uuid.Nil, fmt.Errorf("capability eval: no retriever wired — retrieval cannot be measured")
	}
	if req.Topics <= 0 {
		req.Topics = defaultEvalTopics
	}
	if req.TopK <= 0 {
		req.TopK = defaultEvalTopK
	}

	runID, err := e.createRun(ctx, expertID, req.TopK, req.GraphExpansion)
	if err != nil {
		return uuid.Nil, err
	}

	// Any failure past this point must land on the run row, or the admin screen
	// shows a pass that is "running" forever.
	if err := e.executeRun(ctx, runID, expertID, req); err != nil {
		e.failRun(ctx, runID, err)
		return runID, err
	}
	return runID, nil
}

func (e *CapabilityEvaluator) createRun(ctx context.Context, expertID uuid.UUID, topK int, graphExpansion bool) (uuid.UUID, error) {
	var runID uuid.UUID
	if err := e.db.QueryRow(ctx, `
		INSERT INTO expert_capability_eval_runs (expert_id, status, top_k, graph_expansion)
		VALUES ($1, 'running', $2, $3)
		RETURNING id`, expertID, topK, graphExpansion).Scan(&runID); err != nil {
		return uuid.Nil, fmt.Errorf("capability eval: create run: %w", err)
	}
	return runID, nil
}

func (e *CapabilityEvaluator) failRun(ctx context.Context, runID uuid.UUID, cause error) {
	if _, err := e.db.Exec(ctx, `
		UPDATE expert_capability_eval_runs
		   SET status = 'failed', error_message = $1, completed_at = now()
		 WHERE id = $2`, cause.Error(), runID); err != nil {
		e.logger.Warn("capability eval: could not record run failure (non-fatal)",
			zap.String("run_id", runID.String()), zap.Error(err))
	}
}

func (e *CapabilityEvaluator) executeRun(ctx context.Context, runID, expertID uuid.UUID, req CapabilityEvalRequest) error {
	if req.Regenerate {
		if _, err := e.db.Exec(ctx,
			`DELETE FROM expert_capability_cases WHERE expert_id = $1`, expertID); err != nil {
			return fmt.Errorf("capability eval: clear question set: %w", err)
		}
	}

	// Topics that carry the most chunks, ignoring the unlabelled bucket: a topic
	// called "general" has no subject to write a question about.
	topics, err := e.topTopics(ctx, expertID, req.Topics)
	if err != nil {
		return err
	}
	if len(topics) == 0 {
		return fmt.Errorf("capability eval: no labelled topics to evaluate — run topic extraction first")
	}

	for _, topic := range topics {
		if err := e.ensureCasesForTopic(ctx, expertID, topic); err != nil {
			// One topic failing to generate must not abort the pass; the report
			// says which topics produced no cases.
			e.logger.Warn("capability eval: question generation failed for topic",
				zap.String("expert_id", expertID.String()),
				zap.String("topic", topic),
				zap.Error(err))
		}
	}

	cases, err := e.loadCases(ctx, expertID)
	if err != nil {
		return err
	}
	if len(cases) == 0 {
		return fmt.Errorf("capability eval: no questions could be generated from this corpus")
	}

	results := make([]CapabilityCaseResult, 0, len(cases))
	for _, c := range cases {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		result := e.evaluateCase(ctx, expertID, c, req.TopK, req.GraphExpansion)
		if err := e.storeResult(ctx, runID, expertID, c, result); err != nil {
			return err
		}
		results = append(results, result)
	}

	return e.summarise(ctx, runID, expertID, topics, req.TopK, results)
}

// topTopics lists the expert's topics by coverage, excluding the unlabelled bucket.
func (e *CapabilityEvaluator) topTopics(ctx context.Context, expertID uuid.UUID, limit int) ([]string, error) {
	rows, err := e.db.Query(ctx, `
		SELECT topic
		  FROM expert_capabilities
		 WHERE expert_id = $1
		   AND topic IS NOT NULL
		   AND topic <> ''
		   AND topic <> 'general'
		 ORDER BY chunk_count DESC, topic
		 LIMIT $2`, expertID, limit)
	if err != nil {
		return nil, fmt.Errorf("capability eval: list topics: %w", err)
	}
	defer rows.Close()

	var topics []string
	for rows.Next() {
		var topic string
		if scanErr := rows.Scan(&topic); scanErr != nil {
			continue
		}
		topics = append(topics, topic)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("capability eval: list topics: %w", err)
	}
	return topics, nil
}

// topicCoverage reads how many chunks carry each topic.
//
// Reported next to the measured level because coverage and capability are
// different facts: "14 chunks, answers definitions only" and "14 chunks, handles
// failure modes" have the same coverage and very different value.
func (e *CapabilityEvaluator) topicCoverage(ctx context.Context, expertID uuid.UUID) (map[string]int, error) {
	rows, err := e.db.Query(ctx, `
		SELECT topic, chunk_count
		  FROM expert_capabilities
		 WHERE expert_id = $1 AND topic IS NOT NULL AND topic <> ''`, expertID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]int{}
	for rows.Next() {
		var topic string
		var count int
		if scanErr := rows.Scan(&topic, &count); scanErr != nil {
			continue
		}
		out[topic] = count
	}
	return out, rows.Err()
}

// runMetrics derives retrieval metrics from one run's stored per-case evidence.
//
// WHY derived rather than stored on the run row: the per-case evidence is already
// durable, and a second copy could disagree with it. MRR comes from the stored rank of
// each hit, so every number here is reproducible from the rows the report shows.
func (e *CapabilityEvaluator) runMetrics(ctx context.Context, runID uuid.UUID) (RetrievalMetrics, error) {
	var m RetrievalMetrics
	if err := e.db.QueryRow(ctx, `
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE hit_rank > 0),
		       COALESCE(AVG(CASE WHEN hit_rank > 0 THEN 1.0 / hit_rank END), 0)
		  FROM expert_capability_results
		 WHERE run_id = $1`, runID,
	).Scan(&m.Cases, &m.Hits, &m.MRR); err != nil {
		return m, fmt.Errorf("capability eval: read run metrics: %w", err)
	}
	return m, nil
}

// topicPassage is a chunk selected to host a question.
type topicPassage struct {
	ID         uuid.UUID
	Text       string
	SourceFile string
}

// ensureCasesForTopic generates and stores the questions for one topic, unless the
// topic already has them.
func (e *CapabilityEvaluator) ensureCasesForTopic(ctx context.Context, expertID uuid.UUID, topic string) error {
	var existing int
	if err := e.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM expert_capability_cases WHERE expert_id = $1 AND topic = $2`,
		expertID, topic).Scan(&existing); err != nil {
		return fmt.Errorf("count existing cases: %w", err)
	}
	if existing > 0 {
		return nil
	}

	passages, err := e.topicPassages(ctx, expertID, topic, casesPerTopic)
	if err != nil {
		return err
	}
	if len(passages) == 0 {
		return fmt.Errorf("no chunks for topic %q", topic)
	}

	questions, err := e.generateQuestions(ctx, topic, passages)
	if err != nil {
		return err
	}

	for _, q := range questions {
		if _, err := e.db.Exec(ctx, `
			INSERT INTO expert_capability_cases
				(expert_id, topic, level, question, expected_chunk_id, expected_source_file)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (expert_id, question) DO NOTHING`,
			expertID, topic, q.Level, q.Question, q.PassageID, q.SourceFile); err != nil {
			return fmt.Errorf("store case: %w", err)
		}
	}
	return nil
}

// topicPassages picks up to n chunks spread across the topic.
//
// WHY spread and not the first n: the first chunks of a topic are its introduction,
// so questions written from them would all test definitions. Sampling across the
// topic is what makes the three difficulty levels meaningful.
func (e *CapabilityEvaluator) topicPassages(ctx context.Context, expertID uuid.UUID, topic string, n int) ([]topicPassage, error) {
	rows, err := e.db.Query(ctx, `
		SELECT id, chunk_text, COALESCE(source_file,'')
		  FROM course_chunks
		 WHERE expert_id = $1 AND topic = $2
		 ORDER BY chunk_index`, expertID, topic)
	if err != nil {
		return nil, fmt.Errorf("read chunks for topic %q: %w", topic, err)
	}
	defer rows.Close()

	var all []topicPassage
	for rows.Next() {
		var p topicPassage
		if scanErr := rows.Scan(&p.ID, &p.Text, &p.SourceFile); scanErr != nil {
			continue
		}
		all = append(all, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read chunks for topic %q: %w", topic, err)
	}
	if len(all) <= n {
		return all, nil
	}

	// Evenly spaced pick, always including the first and last.
	picked := make([]topicPassage, 0, n)
	for i := 0; i < n; i++ {
		idx := i * (len(all) - 1) / (n - 1)
		picked = append(picked, all[idx])
	}
	return picked, nil
}

// generatedQuestion is one question bound to the passage that answers it.
type generatedQuestion struct {
	Level      int
	Question   string
	PassageID  uuid.UUID
	SourceFile string
}

// generateQuestions writes one question per passage, at the difficulty that
// passage can carry.
func (e *CapabilityEvaluator) generateQuestions(ctx context.Context, topic string, passages []topicPassage) ([]generatedQuestion, error) {
	var sb strings.Builder
	sb.WriteString("You are writing exam questions for a course expert.\n")
	sb.WriteString(fmt.Sprintf("The topic is %q. Below are passages from the course.\n\n", topic))
	for i, p := range passages {
		sb.WriteString(fmt.Sprintf("Passage %d:\n%s\n\n", i+1, clipPromptText(p.Text, promptPassageChars)))
	}
	sb.WriteString(`Write ONE question per passage, in English, that ITS OWN passage answers.
Assign each question a level:
  1 = a definition question (what it is, when to use it)
  2 = a mechanics question (how it works, trade-offs, why one option over another)
  3 = a failure question (what breaks under load, edge cases, failure modes)
Pick the highest level the passage genuinely supports; do NOT invent details the passage does not contain.

Return ONLY JSON:
{"questions":[{"passage":1,"level":2,"question":"..."}]}`)

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  sb.String(),
		MaxTokens:   2048,
		Temperature: 0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("generate questions for %q: %w", topic, err)
	}

	parsed, err := parseGeneratedQuestions(resp.Content, len(passages))
	if err != nil {
		return nil, fmt.Errorf("parse questions for %q: %w", topic, err)
	}

	out := make([]generatedQuestion, 0, len(parsed))
	for _, q := range parsed {
		p := passages[q.PassageIdx]
		out = append(out, generatedQuestion{
			Level:      q.Level,
			Question:   q.Question,
			PassageID:  p.ID,
			SourceFile: p.SourceFile,
		})
	}
	return out, nil
}

// parsedQuestion is a validated model response entry.
type parsedQuestion struct {
	PassageIdx int // 0-based
	Level      int
	Question   string
}

// parseGeneratedQuestions validates the model's JSON.
//
// Pure on purpose: dropping out-of-range passage indexes and clamping levels is
// exactly the kind of silent-corruption logic that deserves a test. An entry that
// names a passage outside the range is DROPPED rather than clamped — a question
// bound to the wrong passage would score retrieval against content that cannot
// answer it, which would look like a retrieval failure.
func parseGeneratedQuestions(response string, passageCount int) ([]parsedQuestion, error) {
	body := extractJSONObject(response)
	if body == "" {
		return nil, fmt.Errorf("no JSON object in response")
	}

	var payload struct {
		Questions []struct {
			Passage  int    `json:"passage"`
			Level    int    `json:"level"`
			Question string `json:"question"`
		} `json:"questions"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return nil, err
	}

	out := make([]parsedQuestion, 0, len(payload.Questions))
	seen := make(map[int]bool, len(payload.Questions))
	for _, q := range payload.Questions {
		idx := q.Passage - 1
		if idx < 0 || idx >= passageCount {
			continue
		}
		if seen[idx] {
			continue // one question per passage; keep the first
		}
		question := strings.TrimSpace(q.Question)
		if question == "" {
			continue
		}
		level := q.Level
		if level < CapabilityLevelDefinition || level > CapabilityLevelFailure {
			level = CapabilityLevelDefinition
		}
		seen[idx] = true
		out = append(out, parsedQuestion{PassageIdx: idx, Level: level, Question: question})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no usable questions")
	}
	return out, nil
}

// loadCases reads the expert's stored questions.
func (e *CapabilityEvaluator) loadCases(ctx context.Context, expertID uuid.UUID) ([]CapabilityCase, error) {
	rows, err := e.db.Query(ctx, `
		SELECT id, topic, level, question, expected_chunk_id, expected_source_file
		  FROM expert_capability_cases
		 WHERE expert_id = $1
		 ORDER BY topic, level, created_at`, expertID)
	if err != nil {
		return nil, fmt.Errorf("capability eval: load cases: %w", err)
	}
	defer rows.Close()

	var cases []CapabilityCase
	for rows.Next() {
		var c CapabilityCase
		if scanErr := rows.Scan(&c.ID, &c.Topic, &c.Level, &c.Question, &c.ExpectedChunkID, &c.ExpectedSourceFile); scanErr != nil {
			continue
		}
		cases = append(cases, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("capability eval: load cases: %w", err)
	}
	return cases, nil
}

// evaluateCase asks one question and scores the three components.
func (e *CapabilityEvaluator) evaluateCase(ctx context.Context, expertID uuid.UUID, c CapabilityCase, topK int, graphExpansion bool) CapabilityCaseResult {
	result := CapabilityCaseResult{CaseID: c.ID, Question: c.Question, Topic: c.Topic, Level: c.Level}

	var (
		chunks []chinawall.CourseChunk
		err    error
	)
	if graphExpansion {
		chunks, err = e.retriever.GetCourseChunksExpanded(ctx, expertID, c.Question, topK)
	} else {
		chunks, err = e.retriever.GetCourseChunksForWorkflow(ctx, expertID, c.Question, topK)
	}
	if err != nil {
		result.FailureReason = FailureRetrievalMissed
		e.logger.Warn("capability eval: retrieval failed for case (recorded as a miss)",
			zap.String("case_id", c.ID.String()), zap.Error(err))
		return result
	}

	result.RetrievedIDs = make([]uuid.UUID, 0, len(chunks))
	for i, chunk := range chunks {
		result.RetrievedIDs = append(result.RetrievedIDs, chunk.ID)
		if chunk.ID == c.ExpectedChunkID && result.HitRank == 0 {
			result.HitRank = i + 1
		}
	}

	// Build the context with [n] markers, in the order retrieval returned it, so the
	// citation check and the judge see the same numbering the answer does.
	texts := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		texts = append(texts, chunk.Text)
	}
	contextBlock := buildNumberedContext(texts, promptPassageChars)
	if strings.TrimSpace(contextBlock) == "" {
		result.FailureReason = FailureRetrievalMissed
		return result
	}

	answer, err := e.answerer.Answer(ctx, c.Question, contextBlock)
	if err != nil {
		result.FailureReason = FailureEmpty
		e.logger.Warn("capability eval: answering failed",
			zap.String("case_id", c.ID.String()), zap.Error(err))
		return result
	}
	result.Answer = answer
	result.Cited = countCitationMarkers(answer) > 0
	result.Refused = strings.Contains(answer, insufficientContextMarker)

	// One definition of "well-formed enough to judge", shared with the ingest smoke
	// test so a weaker copy cannot let a weaker expert through a training gate.
	if reason := refusalOrCitationFailure(answer); reason != "" {
		result.FailureReason = reason
		return result
	}

	result.Grounded = e.answerer.Judge(ctx, c.Question, contextBlock, answer)
	if result.Grounded != "supported" {
		result.FailureReason = FailureUngrounded
		return result
	}

	// Retrieval is scored last on purpose: a case that was answered correctly
	// despite a miss is still a retrieval failure worth seeing, and a case that
	// passed everything else deserves to have that recorded.
	if result.HitRank == 0 {
		result.FailureReason = FailureRetrievalMissed
		return result
	}

	result.Passed = true
	return result
}

// parseJudgeVerdict reads the judge's JSON, defaulting to unverifiable.
func parseJudgeVerdict(response string) string {
	body := extractJSONObject(response)
	if body == "" {
		return "unverifiable"
	}
	var payload struct {
		Verdict string `json:"verdict"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return "unverifiable"
	}
	switch strings.ToLower(strings.TrimSpace(payload.Verdict)) {
	case "supported":
		return "supported"
	case "refuted":
		return "refuted"
	default:
		return "unverifiable"
	}
}

// countCitationMarkers counts distinct [n] markers in an answer.
//
// Pure. A marker is a bracketed integer; anything else (a bare "[figure]", a
// markdown link) is not a citation, so the check cannot be satisfied by accident.
func countCitationMarkers(answer string) int {
	count := 0
	for i := 0; i < len(answer); i++ {
		if answer[i] != '[' {
			continue
		}
		j := i + 1
		digits := 0
		for j < len(answer) && answer[j] >= '0' && answer[j] <= '9' {
			j++
			digits++
		}
		if digits > 0 && j < len(answer) && answer[j] == ']' {
			count++
			i = j
		}
	}
	return count
}

// extractJSONObject returns the outermost {...} slice of a model response, or "".
// Models wrap JSON in prose or fences regardless of instructions.
func extractJSONObject(response string) string {
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || end <= start {
		return ""
	}
	return response[start : end+1]
}

// clipPromptText bounds how much of a chunk enters a prompt.
func clipPromptText(text string, limit int) string {
	if limit <= 0 || len(text) <= limit {
		return text
	}
	return text[:limit]
}

// storeResult persists one case's evidence.
func (e *CapabilityEvaluator) storeResult(ctx context.Context, runID, expertID uuid.UUID, c CapabilityCase, r CapabilityCaseResult) error {
	if _, err := e.db.Exec(ctx, `
		INSERT INTO expert_capability_results
			(run_id, case_id, expert_id, topic, level, retrieved_ids, hit_rank,
			 cited, grounded, refused, passed, answer, failure_reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		runID, c.ID, expertID, c.Topic, c.Level, r.RetrievedIDs, r.HitRank,
		r.Cited, r.Grounded, r.Refused, r.Passed, r.Answer, r.FailureReason,
	); err != nil {
		return fmt.Errorf("capability eval: store result: %w", err)
	}
	return nil
}

// summarise aggregates the pass, writes the measured verdict back onto the
// capability rows, and closes the run.
func (e *CapabilityEvaluator) summarise(
	ctx context.Context,
	runID, expertID uuid.UUID,
	topics []string,
	topK int,
	results []CapabilityCaseResult,
) error {
	byTopic := make(map[string][]CapabilityCaseResult, len(topics))
	for _, r := range results {
		byTopic[r.Topic] = append(byTopic[r.Topic], r)
	}

	coverage, err := e.topicCoverage(ctx, expertID)
	if err != nil {
		// Coverage is context for the report, not the measurement itself; losing it
		// must not fail a pass that already spent real LLM calls.
		e.logger.Warn("capability eval: could not read topic coverage (non-fatal)",
			zap.String("expert_id", expertID.String()), zap.Error(err))
		coverage = map[string]int{}
	}

	reports := make([]TopicCapabilityReport, 0, len(topics))
	passedTotal, hits, grounded, refused := 0, 0, 0, 0
	for _, r := range results {
		if r.Passed {
			passedTotal++
		}
		if r.HitRank > 0 {
			hits++
		}
		if r.Grounded == "supported" {
			grounded++
		}
		if r.Refused {
			refused++
		}
	}

	for _, topic := range topics {
		topicResults := byTopic[topic]
		if len(topicResults) == 0 {
			continue
		}
		report := buildTopicReport(topic, coverage[topic], topicResults)
		reports = append(reports, report)
		if err := e.writeBackCapability(ctx, expertID, report); err != nil {
			e.logger.Warn("capability eval: could not write measured capability (non-fatal)",
				zap.String("expert_id", expertID.String()),
				zap.String("topic", topic),
				zap.Error(err))
		}
	}

	if _, err := e.db.Exec(ctx, `
		UPDATE expert_capability_eval_runs
		   SET status = 'complete',
		       topics_total = $1, cases_total = $2, cases_passed = $3,
		       retrieval_hits = $4, grounded = $5, refused = $6,
		       completed_at = now()
		 WHERE id = $7`,
		len(reports), len(results), passedTotal, hits, grounded, refused, runID,
	); err != nil {
		return fmt.Errorf("capability eval: close run: %w", err)
	}

	e.logger.Info("capability eval complete",
		zap.String("expert_id", expertID.String()),
		zap.String("run_id", runID.String()),
		zap.Int("topics", len(reports)),
		zap.Int("cases", len(results)),
		zap.Int("passed", passedTotal),
		zap.Int("retrieval_hits", hits),
		zap.Int("refused", refused),
		zap.Int("top_k", topK),
	)
	return nil
}

// buildTopicReport turns one topic's case results into the measured verdict.
//
// coverageChunks is passed in rather than looked up so this stays pure: it is the
// judgement the whole feature rests on, and separating it from the database is what
// makes it testable directly.
func buildTopicReport(topic string, coverageChunks int, results []CapabilityCaseResult) TopicCapabilityReport {
	casesByLevel := map[int]int{}
	passedByLevel := map[int]int{}
	failureSet := map[string]bool{}
	report := TopicCapabilityReport{
		Topic:          topic,
		CoverageChunks: coverageChunks,
		FailureReasons: []string{},
	}

	for _, r := range results {
		report.Cases++
		casesByLevel[r.Level]++
		if r.Passed {
			report.Passed++
			passedByLevel[r.Level]++
			// The question that was answered becomes the claim the expert makes.
			if len(report.CanHandle) < canHandleLimit {
				report.CanHandle = append(report.CanHandle, r.Question)
			}
			continue
		}
		if r.FailureReason != "" {
			failureSet[r.FailureReason] = true
			if len(report.CannotHandle) < cannotHandleLimit {
				// The failed question is the honest limit, with the reason, so the
				// list says "cannot answer this, because retrieval never found it"
				// rather than just "cannot do caching".
				report.CannotHandle = append(report.CannotHandle,
					fmt.Sprintf("%s (%s)", r.Question, r.FailureReason))
			}
		}
	}

	report.MeasuredLevel = measuredLevelFrom(passedByLevel, casesByLevel)
	for reason := range failureSet {
		report.FailureReasons = append(report.FailureReasons, reason)
	}
	sort.Strings(report.FailureReasons)
	return report
}

// measuredLevelFrom decides how deep an expert is on one topic.
//
// Level L counts as achieved only when every level up to L has cases AND at least
// capabilityPassRateThreshold of them passed. WHY cumulative: passing a failure-mode
// question while missing definitions is not depth, it is luck. WHY a rate and not
// all: generated questions vary in quality, so one unlucky case must not hide a
// topic the expert genuinely handles.
//
// Pure — unit-tested.
func measuredLevelFrom(passedByLevel, casesByLevel map[int]int) int {
	measured := 0
	for level := CapabilityLevelDefinition; level <= CapabilityLevelFailure; level++ {
		cases := casesByLevel[level]
		if cases == 0 {
			break // nothing was asked at this depth, so nothing is claimed
		}
		if float64(passedByLevel[level])/float64(cases) < capabilityPassRateThreshold {
			break
		}
		measured = level
	}
	return measured
}

// writeBackCapability replaces the declared capability with the measured one.
//
// depth_level is deliberately NOT touched: it means chunk coverage and is still
// true. measured_level is the new answer to "can this expert handle a question
// about this topic", and keeping both lets the screen show declared next to
// measured — which is the whole point of measuring.
func (e *CapabilityEvaluator) writeBackCapability(ctx context.Context, expertID uuid.UUID, report TopicCapabilityReport) error {
	_, err := e.db.Exec(ctx, `
		UPDATE expert_capabilities
		   SET measured_level = $1,
		       eval_cases = $2,
		       eval_passed = $3,
		       can_handle = $4,
		       cannot_handle = $5,
		       last_evaluated_at = now(),
		       updated_at = now()
		 WHERE expert_id = $6 AND topic = $7`,
		nullIfZero(report.MeasuredLevel), report.Cases, report.Passed,
		report.CanHandle, report.CannotHandle, expertID, report.Topic)
	return err
}

// nullIfZero stores 0 as NULL: a measured level of 0 means "no level achieved",
// which is the absence of a measurement, not a measurement of zero.
func nullIfZero(v int) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

// Report reads the most recent completed run for an expert.
func (e *CapabilityEvaluator) Report(ctx context.Context, expertID uuid.UUID) (*CapabilityEvalReport, error) {
	report := &CapabilityEvalReport{ExpertID: expertID, TopicReports: []TopicCapabilityReport{}, Findings: []string{}}

	err := e.db.QueryRow(ctx, `
		SELECT id, status, topics_total, cases_total, cases_passed,
		       retrieval_hits, grounded, refused, top_k, started_at, completed_at,
		       graph_expansion
		  FROM expert_capability_eval_runs
		 WHERE expert_id = $1
		 ORDER BY started_at DESC
		 LIMIT 1`, expertID,
	).Scan(&report.RunID, &report.Status, &report.TopicsTotal, &report.CasesTotal,
		&report.CasesPassed, &report.RetrievalHits, &report.Grounded, &report.Refused,
		&report.TopK, &report.StartedAt, &report.CompletedAt,
		&report.GraphExpansion)
	if err != nil {
		// No run yet is not an error for the caller: the screen shows "never
		// measured", which is a different state from "measured and empty".
		return nil, nil
	}

	rows, err := e.db.Query(ctx, `
		SELECT topic, level, passed, hit_rank, grounded, refused, failure_reason, question
		  FROM expert_capability_results
		 WHERE run_id = $1
		 ORDER BY topic, level`, report.RunID)
	if err != nil {
		return nil, fmt.Errorf("capability eval: read results: %w", err)
	}
	defer rows.Close()

	byTopic := map[string][]CapabilityCaseResult{}
	order := []string{}
	for rows.Next() {
		var topic, grounded, failureReason, question string
		var level, hitRank int
		var passed, refused bool
		if scanErr := rows.Scan(&topic, &level, &passed, &hitRank, &grounded, &refused, &failureReason, &question); scanErr != nil {
			continue
		}
		if _, seen := byTopic[topic]; !seen {
			order = append(order, topic)
		}
		byTopic[topic] = append(byTopic[topic], CapabilityCaseResult{
			Topic: topic, Level: level, Passed: passed, HitRank: hitRank,
			Grounded: grounded, Refused: refused, FailureReason: failureReason,
			Question: question,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("capability eval: read results: %w", err)
	}

	coverage := map[string]int{}
	if loaded, covErr := e.topicCoverage(ctx, expertID); covErr == nil {
		coverage = loaded
	}

	if metrics, mErr := e.runMetrics(ctx, report.RunID); mErr == nil {
		report.Metrics = metrics
	} else {
		e.logger.Warn("capability eval: could not derive run metrics",
			zap.String("run_id", report.RunID.String()), zap.Error(mErr))
	}

	// The baseline is the previous completed run IN THE SAME RETRIEVAL MODE. Comparing
	// a graph-expanded pass against a plain one would attribute the difference to
	// whichever change happened to be made last, which is exactly the mistake this
	// whole measurement exists to prevent.
	var previousRunID uuid.UUID
	if prevErr := e.db.QueryRow(ctx, `
		SELECT id FROM expert_capability_eval_runs
		 WHERE expert_id = $1 AND id <> $2 AND status = 'complete'
		   AND graph_expansion = $3
		 ORDER BY started_at DESC
		 LIMIT 1`, expertID, report.RunID, report.GraphExpansion).Scan(&previousRunID); prevErr == nil {
		if previous, mErr := e.runMetrics(ctx, previousRunID); mErr == nil {
			report.Previous = &previous
		}
	}

	for _, topic := range order {
		report.TopicReports = append(report.TopicReports, buildTopicReport(topic, coverage[topic], byTopic[topic]))
	}
	report.Findings = evalFindings(report)
	return report, nil
}

// evalFindings states what the numbers mean and what to do — an eval that only
// reports a score leaves the admin to guess which part is broken.
func evalFindings(r *CapabilityEvalReport) []string {
	findings := []string{}
	if r.CasesTotal == 0 {
		return []string{"No questions were evaluated in this run."}
	}

	passRate := r.CasesPassed * 100 / r.CasesTotal
	findings = append(findings, fmt.Sprintf("%d%% of questions answered (%d/%d), measured against the top %d retrieved chunks.",
		passRate, r.CasesPassed, r.CasesTotal, r.TopK))

	// The comparison against the previous pass, stated plainly and with the
	// direction spelled out. This is what makes a retrieval change provable: without
	// it, "the retrieval fix helped" is an opinion, and a regression looks the same
	// as an improvement.
	if r.Previous != nil && r.Previous.Cases > 0 {
		nowPct := r.Metrics.HitRate() * 100
		wasPct := r.Previous.HitRate() * 100
		delta := nowPct - wasPct
		switch {
		case delta >= 1:
			findings = append(findings, fmt.Sprintf(
				"Source-chunk retrieval improved: %.0f%% of questions (was %.0f%%), MRR %.2f (was %.2f).",
				nowPct, wasPct, r.Metrics.MRR, r.Previous.MRR))
		case delta <= -1:
			findings = append(findings, fmt.Sprintf(
				"Source-chunk retrieval REGRESSED: %.0f%% of questions (was %.0f%%), MRR %.2f (was %.2f) — check what changed in retrieval before trusting this corpus.",
				nowPct, wasPct, r.Metrics.MRR, r.Previous.MRR))
		default:
			findings = append(findings, fmt.Sprintf(
				"Source-chunk retrieval unchanged: %.0f%% (was %.0f%%), MRR %.2f (was %.2f).",
				nowPct, wasPct, r.Metrics.MRR, r.Previous.MRR))
		}
	}

	missed := r.CasesTotal - r.RetrievalHits
	if missed > 0 {
		findings = append(findings, fmt.Sprintf(
			"%d question(s) never had their source chunk retrieved — the corpus may not cover them, or the wording does not match how the course phrases it.",
			missed))
	}
	ungrounded := r.CasesTotal - r.Grounded
	if ungrounded > 0 {
		findings = append(findings, fmt.Sprintf(
			"%d answer(s) were not supported by the retrieved context — retrieval found something, the answer went beyond it.",
			ungrounded))
	}
	if r.Refused > 0 {
		findings = append(findings, fmt.Sprintf(
			"%d answerable question(s) were refused — the expert declined something its own corpus contains.",
			r.Refused))
	}
	return findings
}
