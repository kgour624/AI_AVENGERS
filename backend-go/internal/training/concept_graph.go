package training

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// Concept relationships (I4): the difference between an index and understanding.
//
// A flat topic list says which subjects a corpus mentions. It cannot say that
// consistent hashing is PART OF sharding, that cache stampede is a FAILURE MODE that
// arises in caching, or that SQL and NoSQL are CONTRASTED as alternatives. Those
// relations are what let an answer explain WHY rather than recite, and what lets
// retrieval follow a question to a neighbouring concept the question's own wording
// never mentioned.
//
// Extraction is an LLM call over the topic list. Writing is idempotent, and every edge
// is validated against the topics that actually exist — a model that invents a topic
// name must not be able to add a relation to a graph that retrieval then trusts.

// The closed set of relations. Closed on purpose: retrieval reads these to decide
// whether a neighbour is worth pulling in, so the behaviour has to be predictable
// rather than dependent on whatever wording the model produced.
const (
	RelationPrerequisiteOf = "prerequisite_of"
	RelationPartOf         = "part_of"
	RelationContrastsWith  = "contrasts_with"
	RelationUsedWith       = "used_with"
	RelationFailsWith      = "fails_with"
)

// ConceptRelations lists the accepted relation types, in a stable order.
func ConceptRelations() []string {
	return []string{
		RelationPrerequisiteOf,
		RelationPartOf,
		RelationContrastsWith,
		RelationUsedWith,
		RelationFailsWith,
	}
}

const (
	// maxTopicsPerCall bounds one extraction prompt. With a 140-character sample per
	// topic this is roughly 3k tokens — comfortable, and in practice a 118-topic corpus
	// fits in a single call. A larger corpus is batched, and relations spanning two
	// batches may be missed; that is a documented limitation, not a silent one.
	maxTopicsPerCall = 150

	// topicSampleChars is how much of a topic's first chunk travels in the prompt, so
	// the model sees what the course actually says rather than judging by name alone.
	topicSampleChars = 140

	// maxNeighbourTopics caps how many neighbouring concepts retrieval may pull chunks
	// from. Bounded because every neighbour adds candidates that the reranker then has
	// to sort — expansion is meant to add recall, not noise.
	maxNeighbourTopics = 5
)

// ConceptEdge is one typed, directed relationship between two topics.
type ConceptEdge struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Relation  string `json:"relation"`
	Rationale string `json:"rationale"`
}

// ConceptExtractResult reports what one extraction pass did.
type ConceptExtractResult struct {
	Topics int `json:"topics"`
	Edges  int `json:"edges"`
	Calls  int `json:"calls"`
	// Rejected counts edges the model produced that named an unknown topic or an
	// unaccepted relation. Reported rather than silently dropped: a rising number means
	// the prompt or the model is drifting.
	Rejected int `json:"rejected"`
}

// ConceptGraph stores and reads an expert's concept relationships.
type ConceptGraph struct {
	db      *pgxpool.Pool
	gateway *gateway.ModelGateway
	logger  *zap.Logger
}

// NewConceptGraph builds a concept graph reader/writer.
func NewConceptGraph(db *pgxpool.Pool, gw *gateway.ModelGateway, logger *zap.Logger) *ConceptGraph {
	return &ConceptGraph{db: db, gateway: gw, logger: logger}
}

// topicSample is one topic with a short excerpt of what the corpus says about it.
type topicSample struct {
	Topic  string
	Sample string
}

// Extract reads the topic list, asks the model how the topics relate, validates the
// answer, and stores it. Idempotent: re-running adds nothing that already exists.
func (g *ConceptGraph) Extract(ctx context.Context, expertID uuid.UUID) (ConceptExtractResult, error) {
	var result ConceptExtractResult

	topics, err := g.topicSamples(ctx, expertID)
	if err != nil {
		return result, err
	}
	result.Topics = len(topics)
	if len(topics) < 2 {
		return result, fmt.Errorf("concept graph: need at least 2 labelled topics, found %d", len(topics))
	}

	allowed := make(map[string]bool, len(topics))
	for _, t := range topics {
		allowed[t.Topic] = true
	}

	for start := 0; start < len(topics); start += maxTopicsPerCall {
		end := start + maxTopicsPerCall
		if end > len(topics) {
			end = len(topics)
		}
		batch := topics[start:end]

		response, callErr := g.requestEdges(ctx, batch)
		result.Calls++
		if callErr != nil {
			// One failed batch must not discard the ones that worked.
			g.logger.Warn("concept graph: extraction call failed",
				zap.String("expert_id", expertID.String()),
				zap.Int("batch_start", start),
				zap.Error(callErr))
			continue
		}

		edges, rejected, parseErr := parseConceptEdges(response, allowed)
		if parseErr != nil {
			g.logger.Warn("concept graph: could not parse extraction response",
				zap.String("expert_id", expertID.String()),
				zap.Int("batch_start", start),
				zap.Error(parseErr))
			continue
		}
		result.Rejected += rejected

		for _, edge := range edges {
			if _, err := g.db.Exec(ctx, `
				INSERT INTO expert_concept_edges (expert_id, from_topic, to_topic, relation, rationale)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (expert_id, from_topic, to_topic, relation) DO UPDATE SET
					rationale = EXCLUDED.rationale`,
				expertID, edge.From, edge.To, edge.Relation, edge.Rationale,
			); err != nil {
				return result, fmt.Errorf("concept graph: store edge: %w", err)
			}
			result.Edges++
		}
	}

	g.logger.Info("concept graph extracted",
		zap.String("expert_id", expertID.String()),
		zap.Int("topics", result.Topics),
		zap.Int("edges", result.Edges),
		zap.Int("calls", result.Calls),
		zap.Int("rejected", result.Rejected),
	)
	return result, nil
}

// topicSamples reads the labelled topics with a short excerpt each, strongest coverage
// first so a bounded prompt sees the most substantial topics.
func (g *ConceptGraph) topicSamples(ctx context.Context, expertID uuid.UUID) ([]topicSample, error) {
	rows, err := g.db.Query(ctx, `
		SELECT c.topic,
		       COALESCE((
		           SELECT left(k.chunk_text, $2)
		             FROM course_chunks k
		            WHERE k.expert_id = c.expert_id AND k.topic = c.topic
		            ORDER BY k.chunk_index
		            LIMIT 1
		       ), '')
		  FROM expert_capabilities c
		 WHERE c.expert_id = $1
		   AND c.topic IS NOT NULL
		   AND c.topic <> ''
		   AND c.topic <> 'general'
		 ORDER BY c.chunk_count DESC, c.topic`, expertID, topicSampleChars)
	if err != nil {
		return nil, fmt.Errorf("concept graph: read topics: %w", err)
	}
	defer rows.Close()

	var topics []topicSample
	for rows.Next() {
		var t topicSample
		if scanErr := rows.Scan(&t.Topic, &t.Sample); scanErr != nil {
			continue
		}
		topics = append(topics, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("concept graph: read topics: %w", err)
	}
	return topics, nil
}

// requestEdges asks the model how one batch of topics relate.
func (g *ConceptGraph) requestEdges(ctx context.Context, topics []topicSample) (string, error) {
	var sb strings.Builder
	sb.WriteString("You are mapping how the concepts in a technical course relate to each other.\n\n")
	sb.WriteString("Topics, each with a short excerpt of what the course says about it:\n")
	for _, t := range topics {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", t.Topic, strings.ReplaceAll(t.Sample, "\n", " ")))
	}
	sb.WriteString(`
List the relationships the material actually supports, using ONLY the topic names above.

Allowed relations:
  prerequisite_of — you must understand FROM before TO
  part_of         — FROM is a component of TO
  contrasts_with  — the course compares FROM and TO as alternatives
  used_with       — FROM and TO are typically used together
  fails_with      — FROM is a failure mode or limitation that arises in TO

Rules: do not invent topics; do not invent relations the excerpts do not suggest;
prefer a few certain edges over many speculative ones.

Return ONLY JSON:
{"edges":[{"from":"topic_name","to":"topic_name","relation":"part_of","rationale":"one sentence"}]}`)

	resp, err := g.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  sb.String(),
		MaxTokens:   4096,
		Temperature: 0.2,
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// parseConceptEdges validates a model response against the topics that exist.
//
// WHY validation is strict: the graph is read by retrieval, which pulls a neighbour's
// chunks into the candidate set. An invented topic name would create a neighbour that
// matches no corpus rows, and an unaccepted relation would make the expansion
// behaviour depend on model wording. Rejected edges are counted, not hidden, so a
// drifting prompt shows up as a number.
//
// Pure — unit-tested.
func parseConceptEdges(response string, allowedTopics map[string]bool) ([]ConceptEdge, int, error) {
	body := extractJSONObject(response)
	if body == "" {
		return nil, 0, fmt.Errorf("no JSON object in response")
	}

	var payload struct {
		Edges []struct {
			From      string `json:"from"`
			To        string `json:"to"`
			Relation  string `json:"relation"`
			Rationale string `json:"rationale"`
		} `json:"edges"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return nil, 0, err
	}

	allowedRelations := make(map[string]bool, len(ConceptRelations()))
	for _, r := range ConceptRelations() {
		allowedRelations[r] = true
	}

	rejected := 0
	seen := map[string]bool{}
	out := make([]ConceptEdge, 0, len(payload.Edges))
	for _, e := range payload.Edges {
		from := strings.TrimSpace(e.From)
		to := strings.TrimSpace(e.To)
		relation := strings.TrimSpace(strings.ToLower(e.Relation))

		if !allowedTopics[from] || !allowedTopics[to] {
			rejected++
			continue
		}
		if from == to {
			rejected++
			continue
		}
		if !allowedRelations[relation] {
			rejected++
			continue
		}
		key := from + "\x00" + to + "\x00" + relation
		if seen[key] {
			continue
		}
		seen[key] = true

		out = append(out, ConceptEdge{
			From:      from,
			To:        to,
			Relation:  relation,
			Rationale: strings.TrimSpace(e.Rationale),
		})
	}
	return out, rejected, nil
}

// Load reads an expert's stored edges, strongest relations first.
func (g *ConceptGraph) Load(ctx context.Context, expertID uuid.UUID) ([]ConceptEdge, error) {
	rows, err := g.db.Query(ctx, `
		SELECT from_topic, to_topic, relation, rationale
		  FROM expert_concept_edges
		 WHERE expert_id = $1
		 ORDER BY relation, from_topic, to_topic`, expertID)
	if err != nil {
		return nil, fmt.Errorf("concept graph: load edges: %w", err)
	}
	defer rows.Close()

	var edges []ConceptEdge
	for rows.Next() {
		var e ConceptEdge
		if scanErr := rows.Scan(&e.From, &e.To, &e.Relation, &e.Rationale); scanErr != nil {
			continue
		}
		edges = append(edges, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("concept graph: load edges: %w", err)
	}
	return edges, nil
}

// NeighbourTopics returns topics related to any of the given topics, excluding the
// given ones.
//
// Used by retrieval expansion: the topics a question's best chunks belong to are the
// seed, and their neighbours are the concepts the question did not name but probably
// means. Sorted and capped so the candidate set stays bounded and reproducible.
func (g *ConceptGraph) NeighbourTopics(ctx context.Context, expertID uuid.UUID, seed []string) ([]string, error) {
	if len(seed) == 0 {
		return nil, nil
	}

	rows, err := g.db.Query(ctx, `
		SELECT DISTINCT neighbour FROM (
		    SELECT to_topic   AS neighbour FROM expert_concept_edges
		     WHERE expert_id = $1 AND from_topic = ANY($2)
		    UNION
		    SELECT from_topic AS neighbour FROM expert_concept_edges
		     WHERE expert_id = $1 AND to_topic = ANY($2)
		) n
		WHERE neighbour <> ALL($2)
		ORDER BY neighbour
		LIMIT $3`, expertID, seed, maxNeighbourTopics)
	if err != nil {
		return nil, fmt.Errorf("concept graph: read neighbours: %w", err)
	}
	defer rows.Close()

	var neighbours []string
	for rows.Next() {
		var topic string
		if scanErr := rows.Scan(&topic); scanErr != nil {
			continue
		}
		neighbours = append(neighbours, topic)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("concept graph: read neighbours: %w", err)
	}
	return neighbours, nil
}

// NeighboursForChunks resolves the topics of the given chunk ids and returns their
// graph neighbours.
func (g *ConceptGraph) NeighboursForChunks(ctx context.Context, expertID uuid.UUID, chunkIDs []uuid.UUID) ([]string, error) {
	if len(chunkIDs) == 0 {
		return nil, nil
	}

	rows, err := g.db.Query(ctx, `
		SELECT DISTINCT topic
		  FROM course_chunks
		 WHERE expert_id = $1
		   AND id = ANY($2)
		   AND topic IS NOT NULL
		   AND topic <> ''
		   AND topic <> 'general'`, expertID, chunkIDs)
	if err != nil {
		return nil, fmt.Errorf("concept graph: read chunk topics: %w", err)
	}
	defer rows.Close()

	var seed []string
	for rows.Next() {
		var topic string
		if scanErr := rows.Scan(&topic); scanErr != nil {
			continue
		}
		seed = append(seed, topic)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("concept graph: read chunk topics: %w", err)
	}
	sort.Strings(seed)

	return g.NeighbourTopics(ctx, expertID, seed)
}
