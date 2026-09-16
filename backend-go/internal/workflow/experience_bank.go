package workflow

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// genericMarker: LLM marks generic claims with this tag.
// Gate 3 prompt instructs LLM to use this marker.
const genericMarker = "[GENERIC]"

// ExperienceBank saves generic knowledge used in Gate 3 to
// pending_experience table for admin review.
//
// WHY pending (not direct course_chunks):
//   Direct save risks polluting training with hallucinated knowledge.
//   Admin reviews and approves before it becomes training data.
//   This is the "Experience Bank" — generic knowledge becomes
//   expert's experience after admin validation.
//
// SOLID:
//   SRP: ExperienceBank only saves, doesn't decide what to save.
//   OCP: New sources (peer, generic) added without changing save logic.
type ExperienceBank struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewExperienceBank creates a new ExperienceBank.
func NewExperienceBank(db *pgxpool.Pool, logger *zap.Logger) *ExperienceBank {
	return &ExperienceBank{db: db, logger: logger}
}

// ExtractGenericClaims finds all [GENERIC] marked sentences in LLM output.
// Called after Gate 3 LLM response to identify what needs saving.
//
// Mental execution:
//   input: "Token bucket algorithm [GENERIC] is used for rate limiting.
//           The bucket size [GENERIC] determines burst capacity.
//           This applies to our URL shortener design."
//   output: ["Token bucket algorithm is used for rate limiting.",
//            "The bucket size determines burst capacity."]
func ExtractGenericClaims(llmOutput string) []string {
	var claims []string
	// Split by sentences (period + space, or newline)
	sentences := splitSentences(llmOutput)
	for _, s := range sentences {
		if strings.Contains(s, genericMarker) {
			// Remove the [GENERIC] marker, keep the content
			clean := strings.ReplaceAll(s, genericMarker, "")
			clean = strings.TrimSpace(clean)
			if len(clean) > 20 { // skip very short fragments
				claims = append(claims, clean)
			}
		}
	}
	return claims
}

// SaveGenericClaims saves extracted generic claims to pending_experience.
// Non-fatal: log and continue if save fails.
//
// Mental execution:
//   claims = ["Token bucket algorithm is used for rate limiting."]
//   expertID = system_design_expert_uuid
//   workflowID = workflow_uuid
//   taskTitle = "Design rate limiter"
//   taskDomain = "system_design"
//
//   INSERT INTO pending_experience
//     (expert_id, workflow_id, task_title, task_domain, content, source)
//   VALUES
//     (uuid, uuid, "Design rate limiter", "system_design",
//      "Token bucket algorithm is used for rate limiting.", "generic")
func (e *ExperienceBank) SaveGenericClaims(
	ctx context.Context,
	expertID uuid.UUID,
	workflowID uuid.UUID,
	taskTitle string,
	taskDomain string,
	claims []string,
) {
	if len(claims) == 0 {
		return
	}

	for _, claim := range claims {
		_, err := e.db.Exec(ctx,
			`INSERT INTO pending_experience
				(expert_id, workflow_id, task_title, task_domain, content, source, status)
			 VALUES ($1, $2, $3, $4, $5, 'generic', 'pending')`,
			expertID, workflowID, taskTitle, taskDomain, claim,
		)
		if err != nil {
			e.logger.Warn("experience bank: save failed (non-fatal)",
				zap.String("expert_id", expertID.String()),
				zap.String("claim", truncate(claim, 50)),
				zap.Error(err),
			)
			// Non-fatal: experience bank failure never blocks workflow
		}
	}

	e.logger.Info("experience bank: generic claims saved for review",
		zap.String("expert_id", expertID.String()),
		zap.Int("claims", len(claims)),
		zap.String("task", taskTitle),
	)
}

// splitSentences splits text into sentences for generic claim extraction.
// Simple split on ". " and "\n" — good enough for claim extraction.
func splitSentences(text string) []string {
	// Replace newlines with period+space for uniform splitting
	normalized := regexp.MustCompile(`\n+`).ReplaceAllString(text, ". ")
	parts := strings.Split(normalized, ". ")
	var sentences []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			sentences = append(sentences, p)
		}
	}
	return sentences
}
