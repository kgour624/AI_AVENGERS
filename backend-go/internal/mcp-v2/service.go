package mcpv2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"ai_avengers/backend/internal/mcp-v2/business"
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

// ListTools fetches all active tool definitions — Redis cache 5m TTL (Phase 3) to avoid DB on every tools/list.
func (s *Service) ListTools(ctx context.Context) ([]ToolDefinition, error) {
	if s.deps.Redis != nil {
		if cached, err := s.deps.Redis.Get(ctx, "mcp_v2:tools"); err == nil && cached != "" {
			var tools []ToolDefinition
			if err := json.Unmarshal([]byte(cached), &tools); err == nil {
				return tools, nil
			}
		}
	}
	tools, err := s.deps.DB.ListTools(ctx)
	if err != nil {
		return nil, NewInternal("list tools", err)
	}
	if s.deps.Redis != nil && len(tools) > 0 {
		if b, err := json.Marshal(tools); err == nil {
			_ = s.deps.Redis.Set(ctx, "mcp_v2:tools", string(b), 300)
		}
	}
	return tools, nil
}

func (s *Service) CreateTool(ctx context.Context, t ToolDefinition) error {
	if t.Name == "" {
		return NewInvalidInput("name required", nil)
	}
	if t.DisplayName == "" {
		return NewInvalidInput("display_name required", nil)
	}
	if t.Description == "" {
		return NewInvalidInput("description required", nil)
	}
	if t.InputSchema == nil || len(t.InputSchema) == 0 {
		return NewInvalidInput("input_schema required", nil)
	}
	if !t.Action.IsEmpty() {
		if err := ValidateAction(t.Action); err != nil {
			return NewInvalidInput(err.Error(), nil)
		}
	}
	if err := s.deps.DB.CreateTool(ctx, t); err != nil {
		return NewInternal("create tool", err)
	}
	if s.deps.Redis != nil {
		_ = s.deps.Redis.Del(ctx, "mcp_v2:tools")
	}
	return nil
}

// ValidateAction validates Dynamic Engine ActionDef — Phase 3 all types live.
func ValidateAction(a business.ActionDef) error {
	switch a.Type {
	case business.ActionSQLRead:
		q := strings.TrimSpace(a.SQLQuery())
		if q == "" {
			return fmt.Errorf("SQL_READ: config.query required")
		}
		if len(q) > 8000 {
			return fmt.Errorf("query too long")
		}
		if a.TimeoutMs > 10000 {
			return fmt.Errorf("SQL_READ timeout_ms max 10000")
		}
	case business.ActionAPICall:
		if a.Config == nil {
			return fmt.Errorf("API_CALL: config required")
		}
		urlStr, _ := a.Config["url"].(string)
		if strings.TrimSpace(urlStr) == "" {
			return fmt.Errorf("API_CALL: config.url required")
		}
		if len(urlStr) > 2048 {
			return fmt.Errorf("url too long")
		}
		if m, ok := a.Config["method"].(string); ok && m != "" {
			mm := strings.ToUpper(strings.TrimSpace(m))
			if mm != "GET" && mm != "POST" && mm != "PUT" && mm != "PATCH" && mm != "DELETE" {
				return fmt.Errorf("invalid method %s", m)
			}
		}
		if a.TimeoutMs > 10000 {
			return fmt.Errorf("API_CALL timeout_ms max 10000")
		}
	case business.ActionLLMPrompt:
		if a.Config == nil {
			return fmt.Errorf("LLM_PROMPT: config required")
		}
		p, _ := a.Config["prompt"].(string)
		if strings.TrimSpace(p) == "" {
			p, _ = a.Config["prompt_template"].(string)
		}
		if strings.TrimSpace(p) == "" {
			return fmt.Errorf("LLM_PROMPT: config.prompt required")
		}
		if strings.Contains(p, "{{.") {
			return fmt.Errorf("prompt injection: only {{args.xxx}} allowed")
		}
		if a.TimeoutMs > 90000 {
			return fmt.Errorf("LLM_PROMPT timeout_ms max 90000")
		}
	case business.ActionComposite:
		if a.Config == nil {
			return fmt.Errorf("COMPOSITE: config required")
		}
		steps, ok := a.Config["steps"]
		if !ok {
			return fmt.Errorf("COMPOSITE: config.steps required")
		}
		var n int
		switch v := steps.(type) {
		case []any:
			n = len(v)
		case []map[string]any:
			n = len(v)
		default:
			return fmt.Errorf("COMPOSITE: steps must be array")
		}
		if n == 0 || n > 5 {
			return fmt.Errorf("COMPOSITE: steps 1..5 required")
		}
		if a.TimeoutMs > 120000 {
			return fmt.Errorf("COMPOSITE timeout_ms max 120000")
		}
	default:
		return fmt.Errorf("unknown action type %s", a.Type)
	}
	if a.TimeoutMs < 0 {
		return fmt.Errorf("timeout_ms must be >=0")
	}
	return nil
}

// GetTool fetches single tool — for dispatch.
func (s *Service) GetTool(ctx context.Context, name string) (ToolDefinition, error) {
	t, err := s.deps.DB.GetTool(ctx, name)
	if err != nil {
		return ToolDefinition{}, NewNotFound("tool not found")
	}
	return t, nil
}

// ExecuteGeneric delegates to GenericExecutor interpreter (Hexagon: Service -> Business).
// Phase 3: wires Gateway for LLM_PROMPT/COMPOSITE, Redis cache invalidated on CreateTool.
func (s *Service) ExecuteGeneric(ctx context.Context, toolName string, args map[string]any) (string, error) {
	t, err := s.deps.DB.GetTool(ctx, toolName)
	if err != nil {
		return "", NewNotFound("tool not found: " + toolName)
	}
	if t.Action.IsEmpty() {
		return "", NewInvalidInput("tool has no dynamic action", nil)
	}
	var exec *business.GenericExecutor
	if s.deps.Gateway != nil {
		exec = business.NewGenericExecutorWithGateway(s.deps.DB, s.deps.Gateway)
	} else {
		exec = business.NewGenericExecutor(s.deps.DB)
	}
	result, err := exec.Execute(ctx, t, args)
	if err != nil {
		return "", NewInternal("generic execute", err)
	}
	return result, nil
}

// GenerateSQL is Magic Generate 🪄 — text-to-SQL via cheap LLM + schema context.
func (s *Service) GenerateSQL(ctx context.Context, prompt string) (string, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", NewInvalidInput("prompt required", nil)
	}
	// Fetch schema via storage (SELECT from information_schema) — uses ExecSQLRead with SELECT guard
	schemaRows, err := s.deps.DB.ExecSQLRead(ctx, `SELECT table_name, column_name, data_type FROM information_schema.columns WHERE table_schema='public' AND table_name NOT LIKE 'pg_%' ORDER BY table_name, ordinal_position LIMIT 200`, nil)
	if err != nil {
		// Fallback: minimal schema hint without DB
		schemaRows = []map[string]any{}
	}
	schemaJSON, _ := json.Marshal(schemaRows)
	if len(schemaJSON) > 6000 {
		schemaJSON = schemaJSON[:6000]
	}
	sysMsg := `You are a Postgres SQL generator. Output ONLY a single SELECT query, no explanation, no markdown. 
Rules: SELECT only, use $1,$2 placeholders for user args (do NOT inline values), allow {{args.xxx}} as alternative placeholder, max 1 statement, no DELETE/UPDATE/DROP/INSERT, no ; comments, limit 100.
Schema (table_name,column_name,data_type): ` + string(schemaJSON)
	if s.deps.Gateway == nil {
		return "", NewInternal("gateway not configured", nil)
	}
	msgs := []Message{{Role: "system", Content: sysMsg}, {Role: "user", Content: prompt}}
	out, _, err := s.deps.Gateway.Call(ctx, "cheap", msgs)
	if err != nil {
		return "", NewInternal("generate sql gateway", err)
	}
	out = strings.TrimSpace(out)
	// Strip markdown fences if LLM returns ```sql
	out = strings.TrimPrefix(out, "```sql")
	out = strings.TrimPrefix(out, "```")
	out = strings.TrimSuffix(out, "```")
	out = strings.TrimSpace(out)
	if out == "" {
		return "", NewInternal("empty sql generated", nil)
	}
	return out, nil
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
	// 2. Retrieve chunks (expert-isolated) + DecisionEngine gate2 via port (no mutation) — DB 5s timeout, embed fallback
	var emb []float32
	{
		embCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		e, err2 := s.deps.Vector.Embed(embCtx, question)
		if err2 != nil {
			emb = nil
		} else {
			emb = e
		}
	}
	searchCtx, searchCancel := context.WithTimeout(ctx, 5*time.Second)
	defer searchCancel()
	chunks, _, _, err := s.deps.DB.SearchChunks(searchCtx, expertID, emb, 5, "")
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
	// 3. Gateway call — LLM needs 90s (differentiated timeout, cheap retry)
	// Cancellation guard — paise bachaao, LLM call hi skip
	if err := ctx.Err(); err != nil {
		return "", nil, NewCancelled("request cancelled", err)
	}
	start := time.Now()
	msgs := []Message{{Role: "user", Content: question}}
	for _, c := range chunks {
		msgs = append(msgs, Message{Role: "user", Content: "CITATION: " + c.Text})
	}
	gwCtx, gwCancel := context.WithTimeout(ctx, 90*time.Second)
	defer gwCancel()
	output, usage, err := s.deps.Gateway.Call(gwCtx, tier, msgs)
	if err != nil && errors.Is(err, context.Canceled) {
		return "", nil, NewCancelled("request cancelled", err)
	}
	if ctx.Err() != nil && err != nil {
		return "", nil, NewCancelled("request cancelled", ctx.Err())
	}
	if err != nil {
		if tier == "cheap" {
			retryCtx, retryCancel := context.WithTimeout(ctx, 5*time.Second)
			defer retryCancel()
			output, usage, err = s.deps.Gateway.Call(retryCtx, tier, msgs)
		}
		if err != nil {
			return "", nil, NewInternal("gateway", err)
		}
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

// GetStandards — grounded fetch of expert coding/quality standards for a track.
// Reuses AskExpert's Hexagon pattern: Validate -> Embed -> SearchChunks (expert-isolated) -> Gateway synthesis -> Usage log.
func (s *Service) GetStandards(ctx context.Context, expertID, track string) (string, []Chunk, error) {
	if expertID == "" {
		return "", nil, NewInvalidInput("expertId required", nil)
	}
	if track == "" {
		track = "general"
	}
	if _, err := s.deps.DB.GetExpert(ctx, expertID); err != nil {
		return "", nil, NewNotFound("expert not found")
	}
	query := "coding standards quality guidelines " + track
	var emb []float32
	{
		embCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		e, err2 := s.deps.Vector.Embed(embCtx, query)
		if err2 != nil {
			emb = nil
		} else {
			emb = e
		}
	}
	searchCtx, searchCancel := context.WithTimeout(ctx, 5*time.Second)
	defer searchCancel()
	chunks, _, _, err := s.deps.DB.SearchChunks(searchCtx, expertID, emb, 8, "")
	if err != nil {
		return "", nil, NewInternal("search standards chunks", err)
	}
	if len(chunks) == 0 {
		return "No standards found for track: " + track, nil, nil
	}
	msgs := []Message{{Role: "user", Content: "You are an expert standards assistant. Using ONLY the citations below, summarize the coding/quality standards for track '" + track + "'. Cite sources. If a standard is not in citations, say 'not covered'."}}
	for _, c := range chunks {
		msgs = append(msgs, Message{Role: "user", Content: "CITATION: " + c.Text})
	}
	if err := ctx.Err(); err != nil {
		return "", nil, NewCancelled("request cancelled", err)
	}
	start := time.Now()
	gwCtx, gwCancel := context.WithTimeout(ctx, 90*time.Second)
	defer gwCancel()
	output, usage, err := s.deps.Gateway.Call(gwCtx, "cheap", msgs)
	if err != nil && errors.Is(err, context.Canceled) {
		return "", nil, NewCancelled("request cancelled", err)
	}
	if ctx.Err() != nil && err != nil {
		return "", nil, NewCancelled("request cancelled", ctx.Err())
	}
	if err != nil {
		// cheap retry 5s once
		retryCtx, retryCancel := context.WithTimeout(ctx, 5*time.Second)
		defer retryCancel()
		output, usage, err = s.deps.Gateway.Call(retryCtx, "cheap", msgs)
		if err != nil {
			return "", nil, NewInternal("gateway get_standards", err)
		}
	}
	duration := int(time.Since(start).Milliseconds())
	if s.deps.TxStarter != nil {
		txCtx, err := s.deps.TxStarter.Begin(ctx)
		if err == nil {
			errLog := s.deps.DB.InsertUsageLog(txCtx, UsageLog{ExpertID: expertID, Platform: "generic", Tier: "cheap", InputTokens: usage.InputTokens, OutputTokens: usage.OutputTokens, TotalTokens: usage.InputTokens + usage.OutputTokens, CostUSD: usage.CostUSD, DurationMs: duration})
			errInc := s.deps.DB.IncLimits(txCtx, expertID, "generic", usage.InputTokens+usage.OutputTokens)
			if errLog == nil && errInc == nil {
				_ = storage.Commit(txCtx)
			} else {
				_ = storage.Rollback(txCtx)
			}
		} else {
			_ = s.deps.DB.InsertUsageLog(ctx, UsageLog{ExpertID: expertID, Platform: "generic", Tier: "cheap", InputTokens: usage.InputTokens, OutputTokens: usage.OutputTokens, TotalTokens: usage.InputTokens + usage.OutputTokens, CostUSD: usage.CostUSD, DurationMs: duration})
			_ = s.deps.DB.IncLimits(ctx, expertID, "generic", usage.InputTokens+usage.OutputTokens)
		}
	} else {
		_ = s.deps.DB.InsertUsageLog(ctx, UsageLog{ExpertID: expertID, Platform: "generic", Tier: "cheap", InputTokens: usage.InputTokens, OutputTokens: usage.OutputTokens, TotalTokens: usage.InputTokens + usage.OutputTokens, CostUSD: usage.CostUSD, DurationMs: duration})
		_ = s.deps.DB.IncLimits(ctx, expertID, "generic", usage.InputTokens+usage.OutputTokens)
	}
	if s.deps.Redis != nil {
		_, _ = s.deps.Redis.IncrTokens(ctx, fmt.Sprintf("mcp:limits:%s:generic:daily", expertID), int64(usage.InputTokens+usage.OutputTokens))
	}
	return output, chunks, nil
}

// ReviewChange — grounded code-review of a unified diff against expert standards.
func (s *Service) ReviewChange(ctx context.Context, expertID, diff, changeDescription, filePath string) (string, []Chunk, error) {
	if expertID == "" || diff == "" {
		return "", nil, NewInvalidInput("expertId and diff required", nil)
	}
	if _, err := s.deps.DB.GetExpert(ctx, expertID); err != nil {
		return "", nil, NewNotFound("expert not found")
	}
	query := diff
	if changeDescription != "" {
		query = changeDescription + "\n\n" + diff
	}
	if filePath != "" {
		query = "File: " + filePath + "\n" + query
	}
	var emb2 []float32
	{
		embCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		e, err2 := s.deps.Vector.Embed(embCtx, query)
		if err2 != nil {
			emb2 = nil
		} else {
			emb2 = e
		}
	}
	searchCtx2, searchCancel2 := context.WithTimeout(ctx, 5*time.Second)
	defer searchCancel2()
	chunks, _, _, err := s.deps.DB.SearchChunks(searchCtx2, expertID, emb2, 8, "")
	if err != nil {
		return "", nil, NewInternal("search review chunks", err)
	}
	prompt := "You are a senior code reviewer embodying this expert's standards. Review the diff below against ONLY the citations. Return: 1) Verdict (approve/request_changes), 2) Cited violations with file:line if available, 3) Suggested fix grounded in citations. If no violation, say 'No standards violation found'."
	msgs := []Message{{Role: "user", Content: prompt}, {Role: "user", Content: "DIFF:\n" + diff}}
	if changeDescription != "" {
		msgs = append(msgs, Message{Role: "user", Content: "CHANGE INTENT: " + changeDescription})
	}
	if filePath != "" {
		msgs = append(msgs, Message{Role: "user", Content: "FILE: " + filePath})
	}
	for _, c := range chunks {
		msgs = append(msgs, Message{Role: "user", Content: "CITATION: " + c.Text})
	}
	if err := ctx.Err(); err != nil {
		return "", nil, NewCancelled("request cancelled", err)
	}
	start := time.Now()
	gwCtx2, gwCancel2 := context.WithTimeout(ctx, 90*time.Second)
	defer gwCancel2()
	output, usage, err := s.deps.Gateway.Call(gwCtx2, "strong", msgs)
	if err != nil && errors.Is(err, context.Canceled) {
		return "", nil, NewCancelled("request cancelled", err)
	}
	if ctx.Err() != nil && err != nil {
		return "", nil, NewCancelled("request cancelled", ctx.Err())
	}
	if err != nil {
		return "", nil, NewInternal("gateway review_change", err)
	}
	duration := int(time.Since(start).Milliseconds())
	if s.deps.TxStarter != nil {
		txCtx, err := s.deps.TxStarter.Begin(ctx)
		if err == nil {
			errLog := s.deps.DB.InsertUsageLog(txCtx, UsageLog{ExpertID: expertID, Platform: "generic", Tier: "strong", InputTokens: usage.InputTokens, OutputTokens: usage.OutputTokens, TotalTokens: usage.InputTokens + usage.OutputTokens, CostUSD: usage.CostUSD, DurationMs: duration})
			errInc := s.deps.DB.IncLimits(txCtx, expertID, "generic", usage.InputTokens+usage.OutputTokens)
			if errLog == nil && errInc == nil {
				_ = storage.Commit(txCtx)
			} else {
				_ = storage.Rollback(txCtx)
			}
		} else {
			_ = s.deps.DB.InsertUsageLog(ctx, UsageLog{ExpertID: expertID, Platform: "generic", Tier: "strong", InputTokens: usage.InputTokens, OutputTokens: usage.OutputTokens, TotalTokens: usage.InputTokens + usage.OutputTokens, CostUSD: usage.CostUSD, DurationMs: duration})
			_ = s.deps.DB.IncLimits(ctx, expertID, "generic", usage.InputTokens+usage.OutputTokens)
		}
	} else {
		_ = s.deps.DB.InsertUsageLog(ctx, UsageLog{ExpertID: expertID, Platform: "generic", Tier: "strong", InputTokens: usage.InputTokens, OutputTokens: usage.OutputTokens, TotalTokens: usage.InputTokens + usage.OutputTokens, CostUSD: usage.CostUSD, DurationMs: duration})
		_ = s.deps.DB.IncLimits(ctx, expertID, "generic", usage.InputTokens+usage.OutputTokens)
	}
	if s.deps.Redis != nil {
		_, _ = s.deps.Redis.IncrTokens(ctx, fmt.Sprintf("mcp:limits:%s:generic:daily", expertID), int64(usage.InputTokens+usage.OutputTokens))
	}
	return output, chunks, nil
}



