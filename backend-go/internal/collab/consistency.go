package collab

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// CheckConsistency compares one freshly-finished section against all prior
// sections using ONE cheap LLM call, and returns whether they conflict plus a
// plain-language explanation a non-technical human can read (ans3).
//
// Failure posture: if the LLM call errors, we fail OPEN (conflict=false) so the
// transparency feature itself never blocks progression — but the caller logs a
// consistency_check_failed event so the miss is visible, not swallowed.
func CheckConsistency(
	ctx context.Context,
	gw *gateway.ModelGateway,
	logger *zap.Logger,
	newTitle, newContent string,
	newExpert uuid.UUID,
	prior []Section,
) (conflict bool, explanation string, err error) {
	if gw == nil || len(prior) == 0 {
		return false, "", nil
	}

	var sb strings.Builder
	sb.WriteString("You are checking whether the newest expert section contradicts any earlier section of the SAME collaborative answer.\n\n")
	sb.WriteString("NEWEST SECTION (title: " + newTitle + "):\n" + truncateForCheck(newContent) + "\n\n")
	sb.WriteString("PRIOR SECTIONS (already approved):\n")
	for _, p := range prior {
		sb.WriteString("- " + p.SectionTitle + " (" + p.ExpertName + "): " + truncateForCheck(p.Content) + "\n")
	}
	sb.WriteString("\nReturn JSON only:\n")
	sb.WriteString(`{"conflict": true_or_false, "explanation": "if conflict, a short plain-language explanation of the disagreement for a non-technical reader; otherwise empty string"}`)

	resp, err := gw.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  sb.String(),
		MaxTokens:   300,
		Temperature: 0.1,
	})
	if err != nil {
		return false, "", err
	}

	clean := strings.TrimSpace(resp.Content)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)
	start := strings.Index(clean, "{")
	end := strings.LastIndex(clean, "}")
	if start == -1 || end == -1 || end < start {
		return false, "", nil // unparseable -> fail open
	}
	var parsed struct {
		Conflict    bool   `json:"conflict"`
		Explanation string `json:"explanation"`
	}
	if jsonErr := json.Unmarshal([]byte(clean[start:end+1]), &parsed); jsonErr != nil {
		return false, "", nil // unparseable -> fail open
	}
	return parsed.Conflict, parsed.Explanation, nil
}

func truncateForCheck(s string) string {
	if len(s) > 3000 {
		return s[:3000] + "...[truncated]"
	}
	return s
}
