package chinawall

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// DefaultArtifactQualityFloor is the overall score an artifact must reach to be
// called good enough. It is the same 0.7 the chat path uses, because it is the
// same rubric (B6): the point of sharing the judge is that an artifact and an
// answer are not held to two different standards.
const DefaultArtifactQualityFloor = 0.7

// ArtifactVerdict is the outcome of verifying one produced artifact.
//
// The counts are separate fields rather than one score because they answer
// different questions: a "supported" claim is evidence, a "refuted" one is a
// contradiction that must not ship, and "unverifiable" means the expert's own
// training material does not cover it — which is a coverage finding, not an
// error. Collapsing them into a single number would hide exactly the distinction
// P9 exists to preserve.
type ArtifactVerdict struct {
	Claims          []ClaimReport  `json:"claims"`
	Supported       int            `json:"claims_supported"`
	Refuted         int            `json:"claims_refuted"`
	Unverifiable    int            `json:"claims_unverifiable"`
	Quality         QualityVerdict `json:"quality"`
	// ClaimMethod: "verify" | "no_claims" | "fail_open" | "empty" | "disabled"
	ClaimMethod string `json:"claim_method"`
	// ReferenceChunks is how much of the expert's own material was available to
	// check against. Zero refuted claims over zero reference chunks is not the
	// same result as zero refuted claims over twenty, and only these two numbers
	// together say which one it was.
	ReferenceChunks int `json:"reference_chunks"`
}

// ArtifactVerifier judges a produced artifact with the same two instruments the
// chat path uses: atomic claim verification against the expert's own training
// material (B8/P9), and the quality rubric (B6).
type ArtifactVerifier struct {
	gateway *gateway.ModelGateway
	logger  *zap.Logger
	floor   float64
}

// NewArtifactVerifier builds the verifier. A nil gateway is allowed: Verify then
// reports Method "disabled" instead of pretending to have checked anything.
func NewArtifactVerifier(gw *gateway.ModelGateway, logger *zap.Logger) *ArtifactVerifier {
	return &ArtifactVerifier{gateway: gw, logger: logger, floor: DefaultArtifactQualityFloor}
}

// HasModel reports whether a model is wired, so a caller can skip the work and
// say so rather than describe a verification that never happened.
func (v *ArtifactVerifier) HasModel() bool { return v != nil && v.gateway != nil }

// Verify checks an artifact's claims against reference chunks and scores it.
//
// It never returns an error. Verification is an enhancement layered on top of the
// review gate (A7) and the adversarial debate (C7), exactly as it is on top of the
// citation wall in chat: a broken verifier must not block work that a human
// reviewer already approved. Every failure is reported through ClaimMethod so the
// screen can say "not checked" instead of implying a clean bill of health.
//
// question may be empty — an artifact is not always produced in answer to a
// question. When it is empty the quality rubric is skipped (coverage cannot be
// judged without a question) and only the claims are verified.
func (v *ArtifactVerifier) Verify(ctx context.Context, question, text string, chunks []CourseChunk) ArtifactVerdict {
	out := ArtifactVerdict{ClaimMethod: "disabled", ReferenceChunks: len(chunks)}
	if !v.HasModel() {
		return out
	}
	if strings.TrimSpace(text) == "" {
		out.ClaimMethod = "empty"
		return out
	}

	ref := chunks
	if len(ref) > claimVerifyMaxChunks {
		ref = ref[:claimVerifyMaxChunks]
	}

	candidates := extractMaterialClaims(text)
	if len(candidates) == 0 {
		out.ClaimMethod = "no_claims"
	} else {
		reports, err := verifyClaimsWithLLM(ctx, v.gateway, candidates, ref)
		if err != nil {
			v.logger.Warn("artifact claim verify failed — fail-open",
				zap.Error(err),
				zap.Int("candidates", len(candidates)),
			)
			out.ClaimMethod = "fail_open"
		} else {
			out.Claims = normalizeClaimReports(text, reports)
			out.ClaimMethod = "verify"
		}
	}

	out.Supported = countVerdict(out.Claims, ClaimSupported)
	out.Refuted = countVerdict(out.Claims, ClaimRefuted)
	out.Unverifiable = countVerdict(out.Claims, ClaimUnverifiable)

	if strings.TrimSpace(question) == "" {
		out.Quality = QualityVerdict{Method: "no_question"}
		return out
	}
	out.Quality = judgeTextWithLLM(ctx, v.gateway, v.logger, question, text, ref, v.floor)
	return out
}
