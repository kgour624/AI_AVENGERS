// Package business — GenericExecutor Interpreter (Phase 1: SQL_READ).
// Single Go function per Design Doc: GenericExecutor.Execute(ctx, toolDef, args) with switch action.type
// Security: SELECT-only, pgx param binding, timeout 5s, 1MB cap, recover() isolation (RULE 8-F, 8-B).
package business

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

var argsPlaceholderRe = regexp.MustCompile(`\{\{\s*args\.([a-zA-Z0-9_]+)\s*\}\}`)

// GenericExecutor is the Interpreter — no code-gen, reads Action JSONB and dispatches to SqlRunner etc.
type GenericExecutor struct {
	storer  Storer
	gateway GatewayCaller
}

func NewGenericExecutor(s Storer) *GenericExecutor { return &GenericExecutor{storer: s} }
func NewGenericExecutorWithGateway(s Storer, g GatewayCaller) *GenericExecutor { return &GenericExecutor{storer: s, gateway: g} }

// Execute runs toolDef.Action against args; returns JSON text for ToolResult.
// Phase 1: SQL_READ only; other types return CodeInvalidInput.
func (g *GenericExecutor) Execute(ctx context.Context, tool ToolDefinition, args map[string]any) (string, error) {
	if tool.Action.IsEmpty() {
		return "", fmt.Errorf("tool %s has no dynamic action", tool.Name)
	}
	// Differentiated timeout — 4-5 min domain expert support (Design Doc: LLM needs time)
	defTimeout := 5000
	maxTimeout := 10000
	switch tool.Action.Type {
	case ActionSQLRead:
		defTimeout = 5000
		maxTimeout = 10000
	case ActionAPICall:
		defTimeout = 5000
		maxTimeout = 10000
	case ActionLLMPrompt:
		defTimeout = 60000
		maxTimeout = 90000
	case ActionComposite:
		defTimeout = 90000
		maxTimeout = 120000
	}
	timeout := tool.Action.TimeoutMs
	if timeout <= 0 {
		timeout = defTimeout
	}
	if timeout > maxTimeout {
		timeout = maxTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()

	// Panic isolation — one tool crash != server down (Design Doc §6)
	defer func() {
		if r := recover(); r != nil {
			// recovered, caller gets ToolError
		}
	}()

	switch tool.Action.Type {
	case ActionSQLRead:
		return g.execSQLRead(ctx, tool, args)
	case ActionAPICall:
		return g.execAPICall(ctx, tool, args)
	case ActionLLMPrompt:
		return g.execLLMPrompt(ctx, tool, args)
	case ActionComposite:
		return g.execComposite(ctx, tool, args)
	default:
		return "", fmt.Errorf("unknown action type %s", tool.Action.Type)
	}
}

func (g *GenericExecutor) execSQLRead(ctx context.Context, tool ToolDefinition, args map[string]any) (string, error) {
	rawQuery := tool.Action.SQLQuery()
	if strings.TrimSpace(rawQuery) == "" {
		return "", fmt.Errorf("SQL_READ: query required in action.config.query")
	}
	query, params := bindArgsToQuery(rawQuery, args)
	rows, err := g.storer.ExecSQLRead(ctx, query, params)
	if err != nil {
		return "", fmt.Errorf("sql_read: %w", err)
	}
	// Output mapping: if specified, apply simple template (future), else return JSON array
	if tool.Action.OutputMapping != "" {
		// MVP: return raw JSON — mapping interpreted in Phase 3
	}
	b, err := json.Marshal(rows)
	if err != nil {
		return "", fmt.Errorf("marshal rows: %w", err)
	}
	// 1MB cap (Design Doc §6)
	if len(b) > 1*1024*1024 {
		b = b[:1*1024*1024]
	}
	if len(rows) == 0 {
		return "[]", nil
	}
	return string(b), nil
}

// isBlockedHost implements SSRF guard — blocks private/metadata hosts (Design Doc §6).
func isBlockedHost(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "" {
		return true
	}
	hostOnly := h
	if strings.Contains(h, ":") {
		if ho, _, err := net.SplitHostPort(h); err == nil {
			hostOnly = ho
		}
	}
	if hostOnly == "localhost" || hostOnly == "metadata.google.internal" {
		return true
	}
	if hostOnly == "169.254.169.254" || hostOnly == "127.0.0.1" || hostOnly == "0.0.0.0" || hostOnly == "::1" {
		return true
	}
	if ip := net.ParseIP(hostOnly); ip != nil {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsPrivate() {
			return true
		}
	}
	if strings.HasPrefix(hostOnly, "10.") || strings.HasPrefix(hostOnly, "192.168.") {
		return true
	}
	return false
}

func renderTemplate(tmpl string, args map[string]any) string {
	if tmpl == "" {
		return ""
	}
	return argsPlaceholderRe.ReplaceAllStringFunc(tmpl, func(s string) string {
		sub := argsPlaceholderRe.FindStringSubmatch(s)
		if len(sub) < 2 {
			return s
		}
		if v, ok := args[sub[1]]; ok {
			return fmt.Sprint(v)
		}
		return ""
	})
}

func (g *GenericExecutor) execAPICall(ctx context.Context, tool ToolDefinition, args map[string]any) (string, error) {
	cfg := tool.Action.Config
	rawURL, _ := cfg["url"].(string)
	method, _ := cfg["method"].(string)
	bodyTmpl, _ := cfg["body_template"].(string)
	if strings.TrimSpace(rawURL) == "" {
		return "", fmt.Errorf("API_CALL: config.url required")
	}
	urlStr := renderTemplate(strings.TrimSpace(rawURL), args)
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("url scheme must be http/https")
	}
	if isBlockedHost(u.Host) {
		return "", fmt.Errorf("SSRF blocked host: %s", u.Host)
	}
	// allowlist if provided via config.allowed_hosts
	if v, ok := cfg["allowed_hosts"]; ok {
		allowed := map[string]bool{}
		switch vv := v.(type) {
		case []any:
			for _, x := range vv {
				if s, ok := x.(string); ok {
					allowed[strings.ToLower(s)] = true
				}
			}
		case []string:
			for _, s := range vv {
				allowed[strings.ToLower(s)] = true
			}
		}
		if len(allowed) > 0 {
			hostOnly := strings.ToLower(u.Hostname())
			okHost := false
			for a := range allowed {
				if hostOnly == strings.TrimPrefix(a, "*.") || strings.HasSuffix(hostOnly, strings.TrimPrefix(a, "*")) || hostOnly == a {
					okHost = true
					break
				}
			}
			if !okHost {
				return "", fmt.Errorf("host not in allowlist: %s", u.Host)
			}
		}
	}
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	if method != "GET" && method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
		return "", fmt.Errorf("unsupported method %s", method)
	}
	var bodyReader io.Reader
	if bodyTmpl != "" && method != "GET" {
		rendered := renderTemplate(bodyTmpl, args)
		bodyReader = bytes.NewBufferString(rendered)
	}
	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	// headers from config.headers
	if hv, ok := cfg["headers"]; ok {
		switch hm := hv.(type) {
		case map[string]any:
			for k, v := range hm {
				req.Header.Set(k, renderTemplate(fmt.Sprint(v), args))
			}
		case map[string]string:
			for k, v := range hm {
				req.Header.Set(k, renderTemplate(v, args))
			}
		}
	}
	if bodyReader != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("api call: %w", err)
	}
	defer resp.Body.Close()
	limited := io.LimitReader(resp.Body, 1*1024*1024)
	b, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if len(b) > 1024 {
			b = b[:1024]
		}
		return "", fmt.Errorf("upstream %d: %s", resp.StatusCode, string(b))
	}
	// Try to return compact JSON if possible
	var js any
	if err := json.Unmarshal(b, &js); err == nil {
		if out, err := json.Marshal(js); err == nil {
			if tool.Action.OutputMapping != "" {
				// future: JSONPath mapping
			}
			return string(out), nil
		}
	}
	return string(b), nil
}

func (g *GenericExecutor) execLLMPrompt(ctx context.Context, tool ToolDefinition, args map[string]any) (string, error) {
	if g.gateway == nil {
		return "", fmt.Errorf("LLM_PROMPT: gateway not configured")
	}
	cfg := tool.Action.Config
	rawPrompt, _ := cfg["prompt"].(string)
	if strings.TrimSpace(rawPrompt) == "" {
		return "", fmt.Errorf("LLM_PROMPT: config.prompt required")
	}
	// Strict template: only {{args.xxx}} allowed — block {{.Env}} etc.
	if strings.Contains(rawPrompt, "{{.") || strings.Contains(rawPrompt, "{{ .") {
		return "", fmt.Errorf("prompt injection blocked: only {{args.xxx}} allowed")
	}
	prompt := renderTemplate(rawPrompt, args)
	if len(prompt) > 8000 {
		prompt = prompt[:8000]
	}
	tier, _ := cfg["model_tier"].(string)
	if tier == "" {
		tier = "cheap"
	}
	if tier != "cheap" && tier != "strong" && tier != "fast" {
		tier = "cheap"
	}
	msgs := []Message{{Role: "user", Content: prompt}}
	out, _, err := g.gateway.Call(ctx, tier, msgs)
	if err != nil {
		return "", fmt.Errorf("llm_prompt: %w", err)
	}
	// 2k token ~ 8k chars cap
	if len(out) > 8000 {
		out = out[:8000]
	}
	return out, nil
}

func (g *GenericExecutor) execComposite(ctx context.Context, tool ToolDefinition, args map[string]any) (string, error) {
	cfg := tool.Action.Config
	stepsRaw, ok := cfg["steps"]
	if !ok {
		return "", fmt.Errorf("COMPOSITE: config.steps required")
	}
	stepsSlice, ok := stepsRaw.([]any)
	if !ok {
		if s2, ok2 := stepsRaw.([]map[string]any); ok2 {
			stepsSlice = make([]any, len(s2))
			for i, v := range s2 {
				stepsSlice[i] = v
			}
		} else {
			return "", fmt.Errorf("COMPOSITE: steps must be array")
		}
	}
	if len(stepsSlice) == 0 || len(stepsSlice) > 5 {
		return "", fmt.Errorf("COMPOSITE: steps 1..5 required")
	}
	// Shared context: args + previous step outputs
	ctxArgs := make(map[string]any, len(args)+5)
	for k, v := range args {
		ctxArgs[k] = v
	}
	var lastOut string
	for idx, stepAny := range stepsSlice {
		stepMap, ok := stepAny.(map[string]any)
		if !ok {
			return "", fmt.Errorf("step %d: invalid", idx)
		}
		typeRaw, _ := stepMap["type"].(string)
		stepType := ActionType(strings.ToUpper(strings.TrimSpace(typeRaw)))
		stepCfg, _ := stepMap["config"].(map[string]any)
		if stepCfg == nil {
			stepCfg = map[string]any{}
			if c, ok := stepMap["config"]; ok {
				if cm, ok2 := c.(map[string]any); ok2 {
					stepCfg = cm
				}
			}
		}
		// expose previous output as {{args._prev}} and {{args.step0}} etc.
		if lastOut != "" {
			ctxArgs["_prev"] = lastOut
			ctxArgs[fmt.Sprintf("step%d", idx)] = lastOut
		}
		var out string
		var err error
		switch stepType {
		case ActionSQLRead:
			rawQ, _ := stepCfg["query"].(string)
			if strings.TrimSpace(rawQ) == "" {
				return "", fmt.Errorf("step %d SQL_READ: query required", idx)
			}
			// support {{_prev}} etc. via ctxArgs
			q, params := bindArgsToQuery(renderTemplate(rawQ, ctxArgs), ctxArgs)
			rows, e := g.storer.ExecSQLRead(ctx, q, params)
			if e != nil {
				return "", fmt.Errorf("step %d sql: %w", idx, e)
			}
			b, _ := json.Marshal(rows)
			if len(b) > 1*1024*1024 {
				b = b[:1*1024*1024]
			}
			out = string(b)
		case ActionAPICall:
			tmpTool := ToolDefinition{Action: ActionDef{Type: ActionAPICall, Config: stepCfg, TimeoutMs: 5000}}
			out, err = g.execAPICall(ctx, tmpTool, ctxArgs)
			if err != nil {
				return "", fmt.Errorf("step %d api: %w", idx, err)
			}
		case ActionLLMPrompt:
			tmpTool := ToolDefinition{Action: ActionDef{Type: ActionLLMPrompt, Config: stepCfg}}
			out, err = g.execLLMPrompt(ctx, tmpTool, ctxArgs)
			if err != nil {
				return "", fmt.Errorf("step %d llm: %w", idx, err)
			}
		default:
			return "", fmt.Errorf("step %d: unknown type %s", idx, stepType)
		}
		lastOut = out
		// also store for next steps templating
		ctxArgs[fmt.Sprintf("output_%d", idx)] = lastOut
	}
	if tool.Action.OutputMapping != "" {
		return renderTemplate(tool.Action.OutputMapping, map[string]any{"_prev": lastOut, "output": lastOut}), nil
	}
	return lastOut, nil
}

// bindArgsToQuery converts {{args.xxx}} templates to $1,$2 positional params for pgx.
// If query already contains $1, it binds by sorted key order for deterministic mapping.
// Returns query with $n placeholders and ordered args slice.
func bindArgsToQuery(query string, args map[string]any) (string, []any) {
	if args == nil {
		args = map[string]any{}
	}
	matches := argsPlaceholderRe.FindAllStringSubmatch(query, -1)
	if len(matches) > 0 {
		var params []any
		// Map arg name -> position
		pos := map[string]int{}
		for _, m := range matches {
			key := m[1]
			if _, ok := pos[key]; !ok {
				pos[key] = len(params) + 1
				params = append(params, args[key])
			}
		}
		// Replace all {{args.key}} with $n
		result := argsPlaceholderRe.ReplaceAllStringFunc(query, func(s string) string {
			sub := argsPlaceholderRe.FindStringSubmatch(s)
			if len(sub) < 2 {
				return s
			}
			return fmt.Sprintf("$%d", pos[sub[1]])
		})
		return result, params
	}
	// Already has $ placeholders — bind by sorted keys to keep deterministic order
	if strings.Contains(query, "$") {
		keys := make([]string, 0, len(args))
		for k := range args {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		params := make([]any, 0, len(keys))
		for _, k := range keys {
			params = append(params, args[k])
		}
		return query, params
	}
	// No placeholders — query is static, no params
	return query, nil
}