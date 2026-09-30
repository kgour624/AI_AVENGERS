package mcpv2

import (
	"context"
	"fmt"
	"time"

	"ai_avengers/backend/internal/mcp-v2/business/storage"
)

// Sentinel errors (RULE 8-E:42) — kept for backward compat, new code uses AppError (Gap 4 — Ultimate Go §8)
var (
	ErrLimitExceeded = fmt.Errorf("token limit exceeded")
	ErrNotFound      = fmt.Errorf("not found")
)

// ListExperts returns active experts — App layer validates, Business never imports App (Gap 1 fix).
func (s *Service) ListExperts(ctx context.Context) ([]Expert, error) {
	// App validates, then delegates to Business.Storer via DBPort (App imports Business — §1)
	experts, err := s.deps.DB.ListExperts(ctx)
	if err != nil {
		return nil, NewInternal("list experts", err)
	}
	return experts, nil
}

// GetExpert returns single expert charter — maps to AppError (Gap 4 — App codes).
func (s *Service) GetExpert(ctx context.Context, expertID string) (Expert, error) {
	if expertID == "" {
		return Expert{}, NewInvalidInput("expertId required", nil)
	}
	e, err := s.deps.DB.GetExpert(ctx, expertID)
	if err != nil {
		return Expert{}, NewNotFound("expert not found")
	}
	return e, nil
}

// SearchChunks does expert-scoped pgvector search (WHERE expert_id=$1 isolation — China Wall).
// App validates, Business enforces isolation via Storer (Gap 1).
func (s *Service) SearchChunks(ctx context.Context, expertID, query string, limit int, cursor string) ([]Chunk, string, error) {
	if expertID == "" || query == "" {
		return nil, "", NewInvalidInput("expertId and query required", nil)
	}
	emb, err := s.deps.Vector.Embed(ctx, query)
	if err != nil {
		return nil, "", NewInternal("embed", err)
	}
	chunks, nextCursor, _, err := s.deps.DB.SearchChunks(ctx, expertID, emb, limit, cursor)
	if err != nil {
		return nil, "", NewInternal("search chunks", err)
	}
	return chunks, nextCursor, nil
}

// AskExpert — China Wall + limits + gateway + usage log (PG Source of Truth + atomic via Redis).
// Gap 3: Tx BEGIN at App (if TxStarter wired), Gap 4: AppError codes + wrap only in Business.
func (s *Service) AskExpert(ctx context.Context, expertID, question, platform, tier string) (answer string, citations []Chunk, err error) {
	if expertID == "" || question == "" {
		return "", nil, NewInvalidInput("expertId and question required", nil)
	}
	if tier == "" {
		tier = "strong"
	}
	if platform == "" {
		platform = "generic"
	}
	// 1. Limits check (fail-closed before LLM call) — RULE 8-E:43 — now returns AppError CodeLimitExceeded (Gap 4)
	limits, _ := s.deps.DB.GetExpertLimits(ctx, expertID, platform)
	if limits.CurrentDaily >= limits.DailyLimit || limits.CurrentMonthly >= limits.MonthlyLimit {
		return "", nil, NewLimitExceeded(fmt.Sprintf("daily %d/%d monthly %d/%d", limits.CurrentDaily, limits.DailyLimit, limits.CurrentMonthly, limits.MonthlyLimit))
	}
	// 2. Retrieve chunks (expert-isolated) + DecisionEngine gate2 via port (no mutation)
	emb, err := s.deps.Vector.Embed(ctx, question)
	if err != nil {
		return "", nil, NewInternal("embed question", err)
	}
	chunks, _, _, err := s.deps.DB.SearchChunks(ctx, expertID, emb, 5, "")
	if err != nil {
		return "", nil, NewInternal("search chunks", err)
	}
	if s.deps.Decision != nil {
		gate, err := s.deps.Decision.Check(ctx, expertID, question, chunks)
		if err == nil && gate.GateStopped == 2 && !gate.Allowed {
			return "", nil, NewInvalidInput("knowledge not covered — generic relief check required", nil)
		}
		if err != nil {
			return "", nil, NewInternal("decision check", err)
		}
	}
	// 3. Gateway call (cheap|strong|fast reuse system_settings via GatewayPort)
	start := time.Now()
	msgs := []Message{{Role: "user", Content: question}}
	for _, c := range chunks {
		msgs = append(msgs, Message{Role: "user", Content: "CITATION: " + c.Text})
	}
	output, usage, err := s.deps.Gateway.Call(ctx, tier, msgs)
	if err != nil {
		return "", nil, NewInternal("gateway", err)
	}
	duration := int(time.Since(start).Milliseconds())
	// 4. Usage log + limits inc — Gap 3: App BEGIN tx, Business substitutes pool (Ultimate Go §7)
	// If TxStarter is wired, do InsertUsageLog + IncLimits in one tx; else non-tx fallback.
	if s.deps.TxStarter != nil {
		txCtx, err := s.deps.TxStarter.Begin(ctx)
		if err == nil {
			errLog := s.deps.DB.InsertUsageLog(txCtx, UsageLog{
				ExpertID: expertID, Platform: platform, Tier: tier,
				InputTokens: usage.InputTokens, OutputTokens: usage.OutputTokens,
				TotalTokens: usage.InputTokens + usage.OutputTokens, CostUSD: usage.CostUSD,
				DurationMs: duration,
			})
			errInc := s.deps.DB.IncLimits(txCtx, expertID, platform, usage.InputTokens+usage.OutputTokens)
			if errLog == nil && errInc == nil {
				_ = storage.Commit(txCtx)
			} else {
				_ = storage.Rollback(txCtx)
			}
		} else {
			// Fallback non-tx best-effort if BEGIN fails
			_ = s.deps.DB.InsertUsageLog(ctx, UsageLog{ExpertID: expertID, Platform: platform, Tier: tier, InputTokens: usage.InputTokens, OutputTokens: usage.OutputTokens, TotalTokens: usage.InputTokens + usage.OutputTokens, CostUSD: usage.CostUSD, DurationMs: duration})
			_ = s.deps.DB.IncLimits(ctx, expertID, platform, usage.InputTokens+usage.OutputTokens)
		}
	} else {
		_ = s.deps.DB.InsertUsageLog(ctx, UsageLog{
			ExpertID: expertID, Platform: platform, Tier: tier,
			InputTokens: usage.InputTokens, OutputTokens: usage.OutputTokens,
			TotalTokens: usage.InputTokens + usage.OutputTokens, CostUSD: usage.CostUSD,
			DurationMs: duration,
		})
		_ = s.deps.DB.IncLimits(ctx, expertID, platform, usage.InputTokens+usage.OutputTokens)
	}
	// Redis atomic counter (best-effort, wakeup — not tx)
	if s.deps.Redis != nil {
		_, _ = s.deps.Redis.IncrTokens(ctx, fmt.Sprintf("mcp:limits:%s:%s:daily", expertID, platform), int64(usage.InputTokens+usage.OutputTokens))
	}
	return output, chunks, nil
}