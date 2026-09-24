package chinawall

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// ClaimVerdict is the atomic-claim fact-check result (B8, Arpit masterclass).
// P3: on uncertainty default to unverifiable — never guess "supported".
type ClaimVerdict string

const (
	ClaimSupported     ClaimVerdict = "supported"
	ClaimRefuted       ClaimVerdict = "refuted"
	ClaimUnverifiable  ClaimVerdict = "unverifiable"
)

// ClaimReport is one material claim mapped to evidence (P9 span anchors).
// SpanStart/SpanEnd are byte offsets into the cleaned answer (pre-annotation).
// ChunkIDs are evidence units from the retrieved set; empty → P9 unverified.
type ClaimReport struct {
	Claim         string       `json:"claim"`
	Verdict       ClaimVerdict `json:"verdict"`
	Confidence    float64      `json:"confidence"`
	Justification string       `json:"justification,omitempty"`
	SpanStart     int          `json:"span_start"`
	SpanEnd       int          `json:"span_end"`
	ChunkIDs      []uuid.UUID  `json:"chunk_ids,omitempty"`
}

// ClaimVerifyResult is the Layer-4.5 output (B8).
type ClaimVerifyResult struct {
	Claims       []ClaimReport
	// Annotated answer: unverifiable/refuted claims labelled inline; a
	// short verification footer for observability. Empty Claims → Answer
	// is returned unchanged (no annotations).
	Answer string
	// Method: "verify" | "fail_open" | "disabled" | "empty"
	Method string
}

const (
	// claimVerifyMaxClaims caps extract output (cost + latency bound).
	claimVerifyMaxClaims = 12
	// claimVerifyMaxChunks caps reference material sent to the judge.
	claimVerifyMaxChunks = 6
	// claimMinMaterialRunes: shorter sentences are structure/transition, skip.
	claimMinMaterialRunes = 40
)

// verifyClaims runs the atomic-claim pipeline on a cleaned flat-path
// answer (markers already stripped). extract → verify each → annotate.
//
// Fail-open on LLM/parse error: return the original answer with
// Method=fail_open so a broken verifier never blocks a cited response
// (quality is an enhancement on the citation wall; same policy as B6).
// Kill switch: ClaimVerifyEnabled=false → passthrough.
func (e *Enforcer) verifyClaims(
	ctx context.Context,
	answer string,
	chunks []CourseChunk,
	citations []Citation,
) ClaimVerifyResult {
	passthrough := ClaimVerifyResult{Answer: answer, Method: "disabled", Claims: nil}
	if e == nil || !e.cfg.ClaimVerifyEnabled {
		return passthrough
	}
	if strings.TrimSpace(answer) == "" {
		return ClaimVerifyResult{Answer: answer, Method: "empty"}
	}
	if e.gateway == nil {
		return ClaimVerifyResult{Answer: answer, Method: "fail_open"}
	}

	// Deterministic candidates first (no LLM) — material prose sentences.
	candidates := extractMaterialClaims(answer)
	if len(candidates) == 0 {
		return ClaimVerifyResult{Answer: answer, Method: "verify", Claims: []ClaimReport{}}
	}

	// Restrict reference to cited chunks when available; else top-N retrieved.
	refChunks := chunksForVerification(chunks, citations)

	reports, err := e.llmVerifyClaims(ctx, candidates, refChunks)
	if err != nil {
		e.logger.Warn("claim verify failed — fail-open", zap.Error(err))
		return ClaimVerifyResult{Answer: answer, Method: "fail_open"}
	}

	// P9/P3 post-process: no span → unverifiable; empty chunk evidence on
	// "supported" → demote to unverifiable; clamp confidence; cap list.
	normalized := normalizeClaimReports(answer, reports)
	annotated := annotateUnverified(answer, normalized)

	e.logger.Info("claim verify",
		zap.Int("candidates", len(candidates)),
		zap.Int("reports", len(normalized)),
		zap.Int("supported", countVerdict(normalized, ClaimSupported)),
		zap.Int("refuted", countVerdict(normalized, ClaimRefuted)),
		zap.Int("unverifiable", countVerdict(normalized, ClaimUnverifiable)),
	)
	return ClaimVerifyResult{
		Claims: normalized,
		Answer: annotated,
		Method: "verify",
	}
}

// extractMaterialClaims splits the answer into candidate atomic claims.
// Skips fenced code, headings, short transitions. Pure — unit-tested.
func extractMaterialClaims(answer string) []string {
	// Reuse sentence split that preserves markdown (same as Layer 4).
	units := splitSentences(answer)
	var out []string
	inCode := false
	seen := map[string]bool{}
	for _, u := range units {
		t := strings.TrimSpace(u)
		if strings.HasPrefix(t, "```") {
			inCode = !inCode
			continue
		}
		if inCode || t == "" {
			continue
		}
		if isHeadingOrTransition(t) {
			continue
		}
		if runeLen(t) < claimMinMaterialRunes {
			continue
		}
		// Drop pure list markers / markers we already inject.
		if strings.HasPrefix(t, "[UNVERIFIED]") || strings.HasPrefix(t, "[REFUTED]") {
			continue
		}
		key := strings.ToLower(t)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, t)
		if len(out) >= claimVerifyMaxClaims {
			break
		}
	}
	return out
}

func (e *Enforcer) llmVerifyClaims(
	ctx context.Context,
	candidates []string,
	chunks []CourseChunk,
) ([]ClaimReport, error) {
	ref := buildClaimReference(chunks)
	var candSB strings.Builder
	for i, c := range candidates {
		candSB.WriteString(fmt.Sprintf("[%d] %s\n", i, c))
	}

	systemPrompt := `You are a strict claim verifier for a domain-expert system.
Given CANDIDATE claims from an answer and REFERENCE training chunks, decide for EACH claim:
- supported: reference explicitly backs the claim (quote a short span)
- refuted: reference contradicts the claim
- unverifiable: reference is silent or insufficient (DEFAULT when unsure)

Rules:
1. Atomic: treat each candidate as-is; do not merge or invent claims.
2. No outside knowledge — only the REFERENCE.
3. Uncertain → unverifiable (never guess supported).
4. span_start/span_end: character offsets into the claim text itself (0-based, end exclusive) covering the core assertion; if unknown use 0,0.
5. chunk_ids: list of REFERENCE chunk UUIDs that support/refute; empty when unverifiable.
6. Output ONLY a JSON array, no markdown, no preamble:
[{"i":0,"verdict":"supported|refuted|unverifiable","confidence":0.0,"justification":"...","span_start":0,"span_end":0,"chunk_ids":["uuid-or-empty"]}]
7. One object per candidate index. confidence in [0,1].`

	userPrompt := fmt.Sprintf("REFERENCE:\n%s\n\nCANDIDATES:\n%s", ref, candSB.String())

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelCheap, // batched, cheap (P4/P11)
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    900,
		Temperature:  0.1,
	})
	if err != nil {
		return nil, err
	}
	return parseClaimVerifyResponse(resp.Content, candidates)
}

type claimVerifyLLMItem struct {
	I             int      `json:"i"`
	Verdict       string   `json:"verdict"`
	Confidence    float64  `json:"confidence"`
	Justification string   `json:"justification"`
	SpanStart     int      `json:"span_start"`
	SpanEnd       int      `json:"span_end"`
	ChunkIDs      []string `json:"chunk_ids"`
}

func parseClaimVerifyResponse(raw string, candidates []string) ([]ClaimReport, error) {
	raw = stripJSONFence(raw)
	// Accept a bare array or {"claims":[...]} wrapper.
	var items []claimVerifyLLMItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		var wrap struct {
			Claims []claimVerifyLLMItem `json:"claims"`
		}
		if err2 := json.Unmarshal([]byte(raw), &wrap); err2 != nil {
			return nil, fmt.Errorf("claim verify unmarshal: %w", err)
		}
		items = wrap.Claims
	}

	byIdx := map[int]claimVerifyLLMItem{}
	for _, it := range items {
		byIdx[it.I] = it
	}

	out := make([]ClaimReport, 0, len(candidates))
	for i, claim := range candidates {
		it, ok := byIdx[i]
		if !ok {
			// Missing index → P3 unverifiable
			out = append(out, ClaimReport{
				Claim:      claim,
				Verdict:    ClaimUnverifiable,
				Confidence: 0,
				Justification: "missing from verifier output",
			})
			continue
		}
		v := normalizeVerdict(it.Verdict)
		conf := it.Confidence
		if conf < 0 {
			conf = 0
		}
		if conf > 1 {
			conf = 1
		}
		ids := parseChunkIDList(it.ChunkIDs)
		ss, se := it.SpanStart, it.SpanEnd
		if ss < 0 || se < ss || se > len(claim) {
			ss, se = 0, 0
		}
		out = append(out, ClaimReport{
			Claim:         claim,
			Verdict:       v,
			Confidence:    conf,
			Justification: strings.TrimSpace(it.Justification),
			SpanStart:     ss,
			SpanEnd:       se,
			ChunkIDs:      ids,
		})
	}
	return out, nil
}

// normalizeClaimReports enforces P9 (no evidence span/chunk → unverifiable)
// and locates each claim inside the full answer for answer-level spans.
func normalizeClaimReports(answer string, reports []ClaimReport) []ClaimReport {
	out := make([]ClaimReport, 0, len(reports))
	for _, r := range reports {
		// Locate claim text in answer for answer-level span anchors.
		idx := strings.Index(answer, r.Claim)
		if idx >= 0 {
			r.SpanStart = idx
			r.SpanEnd = idx + len(r.Claim)
		} else {
			// P9: untraceable claim in the shipped answer → unverified.
			r.SpanStart, r.SpanEnd = 0, 0
			if r.Verdict == ClaimSupported {
				r.Verdict = ClaimUnverifiable
				r.Justification = joinJust(r.Justification, "claim span not found in answer")
			}
		}
		// Supported with zero chunk evidence → demote (P9).
		if r.Verdict == ClaimSupported && len(r.ChunkIDs) == 0 {
			r.Verdict = ClaimUnverifiable
			r.Justification = joinJust(r.Justification, "no chunk evidence")
		}
		// Span anchors required for supported (P9).
		if r.Verdict == ClaimSupported && r.SpanStart == r.SpanEnd {
			r.Verdict = ClaimUnverifiable
			r.Justification = joinJust(r.Justification, "empty span")
		}
		out = append(out, r)
	}
	return out
}

// annotateUnverified labels refuted/unverifiable claims in the answer
// body and appends a short verification footer. Supported claims stay
// untouched. Deterministic rewrite — unit-tested.
func annotateUnverified(answer string, reports []ClaimReport) string {
	if len(reports) == 0 {
		return answer
	}
	out := answer
	// Replace longer claims first so nested/overlapping text is stable.
	ordered := make([]ClaimReport, len(reports))
	copy(ordered, reports)
	for i := 0; i < len(ordered); i++ {
		for j := i + 1; j < len(ordered); j++ {
			if len(ordered[j].Claim) > len(ordered[i].Claim) {
				ordered[i], ordered[j] = ordered[j], ordered[i]
			}
		}
	}

	var unverifiable, refuted int
	for _, r := range ordered {
		switch r.Verdict {
		case ClaimUnverifiable:
			unverifiable++
			if r.Claim != "" && strings.Contains(out, r.Claim) {
				out = strings.Replace(out, r.Claim, "[UNVERIFIED] "+r.Claim, 1)
			}
		case ClaimRefuted:
			refuted++
			if r.Claim != "" && strings.Contains(out, r.Claim) {
				out = strings.Replace(out, r.Claim, "[REFUTED] "+r.Claim, 1)
			}
		}
	}

	if unverifiable == 0 && refuted == 0 {
		return out
	}
	var footer strings.Builder
	footer.WriteString("\n\n---\nVerification: ")
	parts := []string{}
	if supported := countVerdict(reports, ClaimSupported); supported > 0 {
		parts = append(parts, fmt.Sprintf("%d supported", supported))
	}
	if unverifiable > 0 {
		parts = append(parts, fmt.Sprintf("%d unverifiable (labelled)", unverifiable))
	}
	if refuted > 0 {
		parts = append(parts, fmt.Sprintf("%d refuted (labelled)", refuted))
	}
	footer.WriteString(strings.Join(parts, ", "))
	footer.WriteString(". Labels mark claims without span-anchored evidence in the training material.")
	return out + footer.String()
}

func chunksForVerification(chunks []CourseChunk, citations []Citation) []CourseChunk {
	if len(citations) > 0 {
		want := map[uuid.UUID]bool{}
		for _, c := range citations {
			want[c.ChunkID] = true
		}
		var hit []CourseChunk
		for _, ch := range chunks {
			if want[ch.ID] {
				hit = append(hit, ch)
			}
		}
		if len(hit) > 0 {
			if len(hit) > claimVerifyMaxChunks {
				hit = hit[:claimVerifyMaxChunks]
			}
			return hit
		}
	}
	n := len(chunks)
	if n > claimVerifyMaxChunks {
		n = claimVerifyMaxChunks
	}
	if n == 0 {
		return nil
	}
	return chunks[:n]
}

func buildClaimReference(chunks []CourseChunk) string {
	if len(chunks) == 0 {
		return "(no reference chunks)"
	}
	var sb strings.Builder
	for i, c := range chunks {
		sb.WriteString(fmt.Sprintf("[CHUNK_%s]\n%s\n\n", c.ID.String(), truncateForLog(c.Text, 500)))
		if i+1 >= claimVerifyMaxChunks {
			break
		}
	}
	return sb.String()
}

func normalizeVerdict(s string) ClaimVerdict {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "supported", "support", "true", "yes":
		return ClaimSupported
	case "refuted", "refute", "false", "contradicted":
		return ClaimRefuted
	default:
		return ClaimUnverifiable
	}
}

func parseChunkIDList(raw []string) []uuid.UUID {
	var out []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, s := range raw {
		s = strings.TrimSpace(s)
		s = strings.TrimPrefix(s, "CHUNK_")
		s = strings.Trim(s, "[]")
		id, err := uuid.Parse(s)
		if err != nil || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func countVerdict(rs []ClaimReport, v ClaimVerdict) int {
	n := 0
	for _, r := range rs {
		if r.Verdict == v {
			n++
		}
	}
	return n
}

func joinJust(a, b string) string {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	return a + "; " + b
}

func stripJSONFence(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimPrefix(raw, "```JSON")
		raw = strings.TrimPrefix(raw, "```")
		if i := strings.LastIndex(raw, "```"); i >= 0 {
			raw = raw[:i]
		}
		raw = strings.TrimSpace(raw)
	}
	// Prefer array if both present.
	if start := strings.Index(raw, "["); start >= 0 {
		if end := strings.LastIndex(raw, "]"); end > start {
			// Heuristic: if object wrapper has no array, fall through.
			candidate := raw[start : end+1]
			if strings.Contains(candidate, "{") {
				return candidate
			}
		}
	}
	if start := strings.Index(raw, "{"); start >= 0 {
		if end := strings.LastIndex(raw, "}"); end > start {
			return raw[start : end+1]
		}
	}
	return raw
}

func runeLen(s string) int {
	n := 0
	for _, r := range s {
		if !unicode.IsSpace(r) {
			n++
		}
	}
	return n
}
