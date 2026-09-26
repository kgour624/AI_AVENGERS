package workflow

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/chinawall"
)

// ArtifactChunkSource loads the reference material an artifact is checked
// against. Declared here by the consumer, for the same reason as the runner's
// other seams: the verifier should not learn how retrieval works, and the
// assembler that already serves the gates and the chat is the one source of
// "an expert's training material".
type ArtifactChunkSource interface {
	GetCourseChunksForWorkflow(ctx context.Context, expertID uuid.UUID, query string, topK int) ([]chinawall.CourseChunk, error)
}

// verifiedArtifactTypes are the artifacts worth a claim-by-claim check.
//
// WHY code artifacts are absent: code is already validated by compiling and
// type-checking it, which is a far stronger check for source text than "is this
// sentence backed by a course chunk". Running the claim verifier over code would
// produce noise, and noise that looks like evidence is worse than none.
var verifiedArtifactTypes = map[string]bool{
	"architecture_decision":  true,
	"data_model_proposed":    true,
	"api_contract_proposed":  true,
	"module_design_proposed": true,
	"design_section_written": true,
}

// artifactVerificationChunks caps how much of the expert's material is pulled for
// one check. Same order as the claim verifier's own reference cap, so the query
// does not fetch rows the verifier would discard.
const artifactVerificationChunks = 6

// ArtifactVerificationService checks what a wave produced, using the same claim
// verification and quality rubric the chat path applies to an answer (G3).
//
// WHY it is separate from the cross-verifier: the cross-verifier is a GATE — it
// can block an artifact and force a revision. This is a MEASUREMENT, deliberately
// fail-open, and keeping the two apart means "the reviewer approved it" and "the
// evidence backs it" stay two different statements on screen instead of one
// blurred one.
type ArtifactVerificationService struct {
	store    *blackboard.Store
	verifier *chinawall.ArtifactVerifier
	chunks   ArtifactChunkSource
	logger   *zap.Logger
}

// NewArtifactVerificationService builds the service. Any nil part makes the whole
// step a no-op rather than a failure: verification must never be the reason a
// workflow stops.
func NewArtifactVerificationService(
	store *blackboard.Store,
	verifier *chinawall.ArtifactVerifier,
	chunks ArtifactChunkSource,
	logger *zap.Logger,
) *ArtifactVerificationService {
	return &ArtifactVerificationService{store: store, verifier: verifier, chunks: chunks, logger: logger}
}

// VerifyWaveArtifacts verifies every design artifact posted during one wave and
// records the verdict on the blackboard.
//
// It reuses the same event window the cross-verifier just looked at (events after
// lastSeqBefore), so it costs one query and cannot drift from what was reviewed.
// Every failure path logs and returns: a workflow whose artifacts were not checked
// must still run, but it must never claim they were.
func (s *ArtifactVerificationService) VerifyWaveArtifacts(
	ctx context.Context,
	workflowID uuid.UUID,
	experts []workflowExpert,
	lastSeqBefore int64,
) {
	if s == nil || s.store == nil || s.verifier == nil || s.chunks == nil {
		return
	}

	events, err := s.store.GetSince(ctx, workflowID, lastSeqBefore, 200)
	if err != nil {
		s.logger.Warn("artifact verify: fetch events failed (non-fatal)", zap.Error(err))
		return
	}

	nameByExpert := make(map[uuid.UUID]string, len(experts))
	for _, e := range experts {
		nameByExpert[e.ID] = e.Name
	}

	for _, event := range events {
		if !verifiedArtifactTypes[event.EventType] {
			continue
		}
		if event.PostedByExpertID == nil {
			continue
		}
		expertID := *event.PostedByExpertID

		text, question := artifactTextAndQuestion(event.Content)
		if strings.TrimSpace(text) == "" {
			continue
		}

		refChunks, err := s.chunks.GetCourseChunksForWorkflow(ctx, expertID, verificationQuery(text), artifactVerificationChunks)
		if err != nil {
			// Non-fatal, and worth saying out loud: a verdict computed against no
			// reference would read as "nothing refuted" when the truth is
			// "nothing checked".
			s.logger.Warn("artifact verify: could not load reference chunks (non-fatal)",
				zap.String("expert_id", expertID.String()),
				zap.String("event_type", event.EventType),
				zap.Error(err),
			)
			refChunks = nil
		}

		verdict := s.verifier.Verify(ctx, question, text, refChunks)
		s.logger.Info("artifact verified",
			zap.String("workflow_id", workflowID.String()),
			zap.String("expert", nameByExpert[expertID]),
			zap.String("artifact_type", event.EventType),
			zap.String("claim_method", verdict.ClaimMethod),
			zap.Int("supported", verdict.Supported),
			zap.Int("refuted", verdict.Refuted),
			zap.Int("unverifiable", verdict.Unverifiable),
			zap.Float64("quality_overall", verdict.Quality.Overall),
		)

		// PostedByClient=true marks it a system event: the blackboard's CHECK
		// constraint requires an event with no expert id to be a client/system
		// post, and an artifact verification is nobody's opinion but the
		// verifier's.
		if _, postErr := s.store.Post(ctx, blackboard.PostRequest{
			WorkflowID:      workflowID,
			EventType:       "artifact_verification",
			PostedByClient:  true,
			ReferencesEventIDs: []uuid.UUID{event.ID},
			Content: map[string]interface{}{
				"artifact_event_id":   event.ID.String(),
				"artifact_type":       event.EventType,
				"expert_id":           expertID.String(),
				"expert_name":         nameByExpert[expertID],
				"claim_method":        verdict.ClaimMethod,
				"claims_supported":    verdict.Supported,
				"claims_refuted":      verdict.Refuted,
				"claims_unverifiable": verdict.Unverifiable,
				"reference_chunks":    verdict.ReferenceChunks,
				"quality_overall":     verdict.Quality.Overall,
				"quality_pass":        verdict.Quality.Pass,
				"quality_method":      verdict.Quality.Method,
				"quality_feedback":    verdict.Quality.Feedback,
				"claims":              verdict.Claims,
			},
		}); postErr != nil {
			s.logger.Warn("artifact verify: could not record the verdict (non-fatal)",
				zap.String("workflow_id", workflowID.String()),
				zap.Error(postErr),
			)
		}
	}
}

// artifactTextAndQuestion pulls the verifiable text and the question the artifact
// answers out of an artifact event's content.
//
// Artifacts are free-form: each event type posts its own shape, and the shapes
// are written by a model. So the known keys are tried in order of how specific
// they are, and anything else falls back to concatenating the event's own string
// values in a stable order — better a slightly odd reference text than a silent
// "nothing to verify". question is best-effort: an artifact produced without a
// question has none, and the verifier then reports that it skipped the rubric
// rather than inventing a question to score against.
func artifactTextAndQuestion(raw json.RawMessage) (text, question string) {
	if len(raw) == 0 {
		return "", ""
	}
	var content map[string]interface{}
	if err := json.Unmarshal(raw, &content); err != nil {
		// Some events post a bare string.
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			return strings.TrimSpace(s), ""
		}
		return "", ""
	}

	for _, key := range []string{"question", "task", "task_title", "title", "summary"} {
		if v, ok := content[key].(string); ok && strings.TrimSpace(v) != "" {
			question = strings.TrimSpace(v)
			break
		}
	}

	for _, key := range []string{"text", "content", "body", "decision", "rationale", "description"} {
		if v, ok := content[key].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v), question
		}
	}

	// Deterministic fallback: sorted keys, so the same artifact always produces
	// the same reference text (a verifier whose input changes between runs would
	// make its own verdicts incomparable).
	keys := make([]string, 0, len(content))
	for k := range content {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		if v, ok := content[k].(string); ok && strings.TrimSpace(v) != "" {
			sb.WriteString(v)
			sb.WriteString("\n\n")
		}
	}
	return strings.TrimSpace(sb.String()), question
}

// verificationQuery builds the retrieval query for an artifact: the opening of
// the artifact itself, which is the best available description of its subject.
func verificationQuery(text string) string {
	const maxQueryRunes = 400
	trimmed := strings.TrimSpace(text)
	if len(trimmed) <= maxQueryRunes {
		return trimmed
	}
	// Cut on a rune boundary so a multi-byte character is never split.
	runes := []rune(trimmed)
	return string(runes[:maxQueryRunes])
}
