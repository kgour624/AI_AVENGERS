package workflow

// The workflow-chat tool loop — gather -> act -> verify, §7 of
// docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md. This is what makes the chat
// capable rather than only conversational (§7.1): it reads the design, searches
// it, and proposes changes, instead of only describing what it might contain.
//
// TOOL CALL PROTOCOL — reused, not invented
//
// The model signals a tool call exactly as agent_loop.go already teaches it to:
// a <tool_call>{...}</tool_call> block containing JSON. extractToolCallBlocks
// and extractFirstJSONObject (agent_loop.go) are reused verbatim — the same
// tolerance for a stray trailing character or an array-wrapped object that was
// needed there is needed here, for the same reason: LLM tool-call output is not
// deterministic.
//
// A second protocol, not agent_loop's: agent_loop's tools are fixed by a Go
// switch (PostArtifact / AskExpert). This loop's tool set is a database-backed
// registry per §7.2, filtered per expert by AllowedTools, so it cannot be a
// switch — the set of cases is not known until the expert row is read.
//
// MUTATING TOOLS DO NOT WRITE THE DESIGN (§7.5, this step)
//
// A mutating tool here creates a proposal — a blackboard event — and nothing
// else. It does not open Aider, does not create a branch, does not commit. That
// half of §7.5 (client approves -> Aider applies -> merge) depends on the
// authoring turn this design introduces in §9, which is a later step. Building
// the apply path before an authoring turn exists to apply into would be
// building on ground that is not there yet — the same mistake as wiring the
// whole Aider pipeline before a single LLM call through it had ever succeeded.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/gateway"
)

// toolLoopMaxStepsDefault is used when system_settings has no
// tool_loop_max_steps row (migration 018 seeds one; this is the same fallback
// pattern genericCeiling uses in chat_service.go — a missing settings row must
// never mean "unbounded").
const toolLoopMaxStepsDefault = 8

// designConflictGate is the approval_requests.gate_name for a conflict that
// could not be auto-resolved and must go to the client.
//
// Already legal: migration 018 widened approval_requests_gate_check to include
// 'design_conflict'. No migration is needed — same reasoning as amendmentGate
// in amendment.go (the constraint was widened for a code path that was missing).
const designConflictGate = "design_conflict"

// maxConflictHopsPerStatement bounds how many times the SAME statement_id
// may be raised via raise_conflict before the deterministic deadlock guard
// (A19) fires: instead of delegating to yet another rank-based resolver
// (which can ping-pong — expert A raises, resolver picks a side, the
// original expert disagrees and raises again, ...), the repeat raise is
// escalated straight to the client. A genuine first-time conflict on a
// statement still goes through normal rank-based auto-resolve.
const maxConflictHopsPerStatement = 2

// Tool is one capability the model may invoke (§7.2). Everything about a tool
// is data except the Go closure that runs it: the registry is built once at
// startup and filtered per expert from AllowedTools, so granting or withdrawing
// a capability is a database change, never a code change.
type Tool struct {
	Name        string
	Description string
	// InputSchema is shown to the model so it knows the call shape. Kept as a
	// human-readable string rather than a JSON Schema document: nothing in this
	// codebase validates against JSON Schema today, and the loop's own input
	// checks (per tool, below) are the actual validation. A full schema
	// validator is future work, not invented here to look complete.
	InputSchema string
	// Mutating: true means the tool's effect is a proposal requiring approval,
	// never a direct write. See the file comment above.
	Mutating bool
	Handler  func(ctx context.Context, loop *toolLoopContext, input json.RawMessage) (any, error)
}

// toolLoopContext is the per-call state every tool handler needs. Passed
// explicitly rather than closed over, so a handler's dependencies are visible
// at its call site instead of hidden in a constructor.
type toolLoopContext struct {
	db            *pgxpool.Pool
	store         *blackboard.Store
	sections      *DesignSectionStore
	gates         *GateSystem
	gw            *gateway.ModelGateway
	workflowID    uuid.UUID
	expert        workflowExpert
	chatID        uuid.UUID
	workspaceRoot string
}

// ToolRegistry holds every tool this codebase knows how to run.
//
// WHY a struct with a slice, not a package-level map: a map of functions
// initialised at package-load time cannot easily take constructor
// dependencies (workspaceRoot below). Building the registry in
// NewToolRegistry keeps the tool set open to that without a global.
type ToolRegistry struct {
	tools  map[string]Tool
	logger *zap.Logger
}

// NewToolRegistry builds the full catalogue (§7.4): six read-only, five
// mutating. workspaceRoot is needed by read_design/search_design, which read
// the merged design workspace on disk (aider_runner.go's own
// {workspaceRoot}/{workflowID}/main/ layout — reused, not a second convention).
func NewToolRegistry(workspaceRoot string, logger *zap.Logger) *ToolRegistry {
	r := &ToolRegistry{tools: map[string]Tool{}, logger: logger}

	r.register(Tool{
		Name:        "list_sections",
		Description: "List every design section that exists, and who owns it.",
		InputSchema: `{}`,
		Handler:     toolListSections,
	})
	r.register(Tool{
		Name: "list_files",
		Description: "List EVERY file that actually exists in this workflow's workspace, including generated " +
			"code (design spine, sections, source files, tests). Use this before read_design: list_sections only " +
			"lists design sections, so code files must be discovered here or you will guess a path that does not exist.",
		InputSchema: `{}`,
		Handler:     toolListFiles,
	})
	r.register(Tool{
		Name:        "read_design",
		Description: "Read ANY file in the workflow workspace by its path (from list_files, list_sections, or \"final.md\").",
		InputSchema: `{"path": "string, required"}`,
		Handler:     toolReadDesign,
	})
	r.register(Tool{
		Name:        "search_design",
		Description: "Search across the spine and every section for a query string.",
		InputSchema: `{"query": "string, required"}`,
		Handler:     toolSearchDesign,
	})
	r.register(Tool{
		Name:        "read_blackboard",
		Description: "Read decision history for this workflow, optionally filtered by event type.",
		InputSchema: `{"event_types": "array of string, optional"}`,
		Handler:     toolReadBlackboard,
	})
	r.register(Tool{
		Name:        "get_acceptance",
		Description: "Read acceptance criteria, optionally filtered to one section.",
		InputSchema: `{"section": "string, optional"}`,
		Handler:     toolGetAcceptance,
	})
	r.register(Tool{
		Name:        "read_decisions",
		Description: "Read the decision log, newest first.",
		InputSchema: `{}`,
		Handler:     toolReadDecisions,
	})

	r.register(Tool{
		Name: "propose_amendment",
		Description: "Propose a change to a design file. old_text must be copied EXACTLY from the " +
			"file (use read_design first) and must appear there only once — it is replaced verbatim " +
			"once the client approves. Leave old_text empty to append instead of replace.",
		InputSchema: `{"target": "string, required", "old_text": "string, required (empty means append)", "new_text": "string, required", "reason": "string, required"}`,
		Mutating:    true,
		Handler:     toolProposeAmendment,
	})
	r.register(Tool{
		Name: "propose_acceptance",
		Description: "Propose a new acceptance criterion for a section. The id and owner are assigned " +
			"for you; verify must be a real command that exits zero or non-zero.",
		InputSchema: `{"section": "string, required", "statement": "string, required", "verify": "string, required", "done_when": "string, required"}`,
		Mutating:    true,
		Handler:     toolProposeAcceptance,
	})
	r.register(Tool{
		Name:        "raise_conflict",
		Description: "Raise a disagreement with an existing design statement.",
		InputSchema: `{"statement_id": "string, required", "position": "string, required", "reason": "string, required"}`,
		Mutating:    true,
		Handler:     toolRaiseConflict,
	})
	r.register(Tool{
		Name:        "ask_expert",
		Description: "Consult a peer expert with a specific question.",
		InputSchema: `{"expert_id": "string (uuid), required", "question": "string, required"}`,
		Mutating:    true,
		Handler:     toolAskExpert,
	})
	r.register(Tool{
		Name:        "ask_client",
		Description: "Escalate a question to the client and pause for their answer.",
		InputSchema: `{"summary": "string, required"}`,
		Mutating:    true,
		Handler:     toolAskClient,
	})

	return r
}

func (r *ToolRegistry) register(t Tool) {
	if _, dup := r.tools[t.Name]; dup {
		// A duplicate name here is a programming error, caught at startup
		// rather than at the first confusing runtime dispatch.
		panic(fmt.Sprintf("tool_loop: duplicate tool registered: %s", t.Name))
	}
	r.tools[t.Name] = t
}

// lookup returns the named tool if it exists AND this expert may call it —
// the same registry ∩ AllowedTools rule as ForExpert, applied to one name
// instead of the whole set. Used at dispatch time so a model that requests an
// unlisted or unpermitted tool gets a named error rather than a panic or a
// silent no-op.
func (r *ToolRegistry) lookup(name string, expert workflowExpert) (Tool, bool) {
	t, ok := r.tools[name]
	if !ok {
		return Tool{}, false
	}
	if !t.Mutating {
		return t, true
	}
	for _, allowed := range expert.AllowedTools {
		if allowed == name {
			return t, true
		}
	}
	return Tool{}, false
}

// ForExpert returns the tools a given expert may call: registry ∩
// expert.AllowedTools (§7.2). An expert with an empty AllowedTools — every
// expert that exists today — gets every READ-ONLY tool and no mutating one.
// That is the safe default, not a special case: nothing has to change for an
// expert that has never been given explicit permissions.
func (r *ToolRegistry) ForExpert(expert workflowExpert) []Tool {
	allowed := make(map[string]bool, len(expert.AllowedTools))
	for _, name := range expert.AllowedTools {
		allowed[name] = true
	}

	var out []Tool
	for _, t := range r.tools {
		if !t.Mutating {
			out = append(out, t)
			continue
		}
		if allowed[t.Name] {
			out = append(out, t)
		}
	}
	return out
}

// promptCatalogue renders the tool list for the system prompt: name,
// description, input shape. Sorted by name so the prompt is stable across
// calls — a map iterates in random order, and a prompt that changes shape
// between otherwise-identical requests would defeat the gateway's cache key.
func promptCatalogue(tools []Tool) string {
	names := make([]string, len(tools))
	byName := make(map[string]Tool, len(tools))
	for i, t := range tools {
		names[i] = t.Name
		byName[t.Name] = t
	}
	sortStrings(names)

	var sb strings.Builder
	sb.WriteString("=== TOOLS AVAILABLE ===\n")
	sb.WriteString("To call one, output EXACTLY:\n")
	sb.WriteString(`<tool_call>{"tool": "<name>", "input": {<args>}}</tool_call>` + "\n\n")
	for _, name := range names {
		t := byName[name]
		kind := "read-only"
		if t.Mutating {
			kind = "MUTATING — creates a proposal, does not write anything directly"
		}
		sb.WriteString(fmt.Sprintf("%s (%s)\n  %s\n  input: %s\n\n", t.Name, kind, t.Description, t.InputSchema))
	}
	sb.WriteString("When you have your answer and no more tools are needed, reply with your\n")
	sb.WriteString("final answer as plain text and no tool_call block.\n")
	return sb.String()
}

// sortStrings avoids importing "sort" for four call sites' worth of one-liner;
// insertion sort is fine for the handful of tool names in play.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}

// toolCall is the parsed shape of one <tool_call>{...}</tool_call> block.
type toolCall struct {
	Tool  string          `json:"tool"`
	Input json.RawMessage `json:"input"`
}

// normalizeDeepSeekToolTokens rewrites DeepSeek's native reserved-token
// tool-call format into our canonical <tool_call>{...}</tool_call> blocks.
//
// WHY this exists:
// DeepSeek-chat is trained with reserved tokens for tool calling:
//   <｜tool▁calls▁begin｜>
//   <｜tool▁call▁begin｜>function_name<｜tool▁separator｜>{"arg":"val"}
//   <｜tool▁call▁end｜>
//   <｜tool▁calls▁end｜>
// When the model drifts from our custom <tool_call>{...}</tool_call>
// prompt instruction (non-deterministic — happens especially on
// tool-heavy queries like list_sections / search_design), it emits
// these reserved tokens instead. extractToolCallBlocks only looks for
// <tool_call>, so the native format produced zero parsed calls and the
// raw garbled text was returned to the client as the final answer.
//
// This function is called BEFORE extractToolCallBlocks so the rest of
// the parse pipeline is unchanged.
//
// Other providers (Anthropic, Gemini, Cavoti, CodeCraftAPI, OpenRouter)
// use OpenAI-compatible text completion and never emit these tokens —
// the function is a no-op for them (the marker strings are absent).
//
// Mental execution:
//   input:  "<｜tool▁calls▁begin｜><｜tool▁call▁begin｜>list_sections<｜tool▁separator｜>{}<｜tool▁call▁end｜><｜tool▁calls▁end｜>"
//   output: "<tool_call>{\"tool\":\"list_sections\",\"input\":{}}</tool_call>"
//
//   input:  "Here is my answer."   (any non-DeepSeek provider)
//   output: "Here is my answer."   (unchanged — no markers present)
func normalizeDeepSeekToolTokens(content string) string {
	// Fast path: if none of the DeepSeek reserved markers are present,
	// return immediately without allocating. This is the common case for
	// every provider that is not DeepSeek.
	const callsBegin = "<｜tool▁calls▁begin｜>"
	if !strings.Contains(content, callsBegin) {
		return content
	}

	const (
		callsEnd  = "<｜tool▁calls▁end｜>"
		callBegin = "<｜tool▁call▁begin｜>"
		callEnd   = "<｜tool▁call▁end｜>"
		separator = "<｜tool▁separator｜>"
	)

	var out strings.Builder
	remaining := content

	for {
		// Find the outer wrapper start.
		wrapStart := strings.Index(remaining, callsBegin)
		if wrapStart == -1 {
			// No more DeepSeek blocks — write the rest as-is.
			out.WriteString(remaining)
			break
		}
		// Write any prose before the block.
		out.WriteString(remaining[:wrapStart])

		// Find the outer wrapper end.
		wrapEnd := strings.Index(remaining[wrapStart:], callsEnd)
		var blockContent string
		if wrapEnd == -1 {
			// Malformed: no closing wrapper. Treat everything from here as
			// the block content and stop after this iteration.
			blockContent = remaining[wrapStart+len(callsBegin):]
			remaining = ""
		} else {
			blockContent = remaining[wrapStart+len(callsBegin) : wrapStart+wrapEnd]
			remaining = remaining[wrapStart+wrapEnd+len(callsEnd):]
		}

		// Each individual call is delimited by callBegin / callEnd.
		for {
			cStart := strings.Index(blockContent, callBegin)
			if cStart == -1 {
				break
			}
			blockContent = blockContent[cStart+len(callBegin):]

			cEnd := strings.Index(blockContent, callEnd)
			var oneCall string
			if cEnd == -1 {
				oneCall = blockContent
				blockContent = ""
			} else {
				oneCall = blockContent[:cEnd]
				blockContent = blockContent[cEnd+len(callEnd):]
			}

			// oneCall is: "tool_name<｜tool▁separator｜>{...json...}"
			// Split on the separator to get name and args.
			sepIdx := strings.Index(oneCall, separator)
			var toolName, argsRaw string
			if sepIdx == -1 {
				// No separator — the whole thing might be a JSON object
				// already (some DeepSeek variants omit the name prefix).
				toolName = ""
				argsRaw = strings.TrimSpace(oneCall)
			} else {
				toolName = strings.TrimSpace(oneCall[:sepIdx])
				argsRaw = strings.TrimSpace(oneCall[sepIdx+len(separator):])
			}

			// argsRaw should be a JSON object. If it is not, skip this call
			// rather than emitting malformed JSON into the pipeline.
			if !strings.HasPrefix(argsRaw, "{") {
				continue
			}

			// Rewrite into our canonical format.
			// {"tool":"<name>","input":<args>}
			// If toolName is empty (no separator variant), the JSON object
			// itself may already contain a "tool" key — pass it through as
			// the input and let extractFirstJSONObject + json.Unmarshal
			// handle it; the toolCall struct will pick up the "tool" field.
			var canonical string
			if toolName != "" {
				// Escape the tool name for safe JSON embedding.
				nameBytes, err := json.Marshal(toolName)
				if err != nil {
					continue
				}
				canonical = fmt.Sprintf(`{"tool":%s,"input":%s}`, string(nameBytes), argsRaw)
			} else {
				// No name prefix — pass the raw JSON through; it may already
				// be in {"tool":"...","input":{...}} shape.
				canonical = argsRaw
			}
			out.WriteString("<tool_call>")
			out.WriteString(canonical)
			out.WriteString("</tool_call>")

			if cEnd == -1 {
				break
			}
		}
	}
	return out.String()
}

// parseToolCalls extracts every tool call from a model response, using the same
// two-stage tolerant parse agent_loop.go uses (extractToolCallBlocks then
// extractFirstJSONObject) for the same reason: tool-call output is not
// guaranteed to be clean JSON.
//
// normalizeDeepSeekToolTokens is called first so that DeepSeek's native
// reserved-token format is rewritten into our canonical <tool_call>{...}</tool_call>
// blocks before extractToolCallBlocks runs. For all other providers the
// normalizer is a no-op (fast path: no reserved-token markers present).
func parseToolCalls(content string) []toolCall {
	content = normalizeDeepSeekToolTokens(content)
	blocks := extractToolCallBlocks(content)
	var calls []toolCall
	for _, block := range blocks {
		jsonBlock := extractFirstJSONObject(block)
		if jsonBlock == "" {
			continue
		}
		var c toolCall
		if err := json.Unmarshal([]byte(jsonBlock), &c); err != nil {
			continue
		}
		if c.Tool == "" {
			continue
		}
		calls = append(calls, c)
	}
	return calls
}

// stripToolCallBlocks removes every <tool_call>...</tool_call> block, leaving
// any prose the model wrote around them. Used when a response mixes a tool call
// with commentary — the commentary is not the final answer (there IS a tool
// call, so the loop continues), but it also should not be silently discarded
// from what gets persisted as this step's content.
//
// normalizeDeepSeekToolTokens is called first for the same reason as in
// parseToolCalls: DeepSeek's native tokens must be rewritten before the
// <tool_call> stripper runs, otherwise they pass through as prose and appear
// in the persisted reasoning step.
func stripToolCallBlocks(content string) string {
	content = normalizeDeepSeekToolTokens(content)
	const open, close = "<tool_call>", "</tool_call>"
	var sb strings.Builder
	for {
		start := strings.Index(content, open)
		if start == -1 {
			sb.WriteString(content)
			break
		}
		sb.WriteString(content[:start])
		rest := content[start:]
		end := strings.Index(rest, close)
		if end == -1 {
			sb.WriteString(content)
			break
		}
		content = rest[end+len(close):]
	}
	return strings.TrimSpace(sb.String())
}

// ============================================================
// Read-only tool implementations
// ============================================================

func toolListSections(ctx context.Context, l *toolLoopContext, _ json.RawMessage) (any, error) {
	secs, err := l.sections.ListSections(ctx, l.workflowID)
	if err != nil {
		return nil, err
	}
	type row struct {
		Path   string `json:"path"`
		Expert string `json:"owner_expert_id"`
	}
	out := make([]row, 0, len(secs))
	for _, s := range secs {
		out = append(out, row{Path: s.SectionPath, Expert: s.ExpertID.String()})
	}
	return out, nil
}

// designPath resolves a design-relative path (as returned by list_sections, or
// "final.md") to an absolute path under the workflow's merged workspace, and
// refuses anything that would escape it.
//
// WHY this exists: the path in a read_design/search_design call comes from
// model output, not from a fixed list the server chose. Nothing else in this
// codebase validates a path against escaping its root — WriteTempFile
// (validation/sandbox.go) only ever writes filenames THIS process generated —
// so this is new, and it is the one place in the tool loop where an
// unconstrained string from the model reaches the filesystem.
func designPath(workspaceRoot string, workflowID uuid.UUID, relPath string) (string, error) {
	if relPath == "" {
		return "", fmt.Errorf("path is required")
	}
	root := filepath.Join(workspaceRoot, workflowID.String(), "main")
	// filepath.Join already collapses ".." segments against the preceding
	// element, but a path with MORE ".." segments than depth (e.g.
	// "../../../etc/passwd") collapses to something outside root — hence the
	// HasPrefix check below rather than trusting Join alone.
	full := filepath.Join(root, relPath)
	rootWithSep := root + string(filepath.Separator)
	if full != root && !strings.HasPrefix(full, rootWithSep) {
		return "", fmt.Errorf("path escapes the design workspace: %s", relPath)
	}
	return full, nil
}

// workspaceFiles lists the files that ACTUALLY exist in a workflow's merged
// workspace, relative to its main/ root, excluding VCS/tool caches.
//
// WHY this exists (production incident): the workflow chat could only list
// *design sections* (blackboard rows). For a run whose output is generated code
// (hld/*.go), that list is empty, so the expert guessed paths from the client's
// message, hit a file that was listed in the UI but missing on disk, and told
// the client it could not read any of its own files. Listing the disk truth is
// what makes "ask me about the files you produced" work.
func workspaceFiles(workspaceRoot string, workflowID uuid.UUID) []string {
	root := filepath.Join(workspaceRoot, workflowID.String(), "main")
	out := []string{}
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || strings.HasPrefix(name, ".aider") {
				return filepath.SkipDir
			}
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	sortStrings(out)
	return out
}

func toolListFiles(_ context.Context, l *toolLoopContext, _ json.RawMessage) (any, error) {
	files := workspaceFiles(l.workspaceRoot, l.workflowID)
	return map[string]any{"files": files, "count": len(files)}, nil
}

func toolReadDesign(ctx context.Context, l *toolLoopContext, input json.RawMessage) (any, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return nil, fmt.Errorf("read_design: invalid input: %w", err)
	}
	full, err := designPath(l.workspaceRoot, l.workflowID, args.Path)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(full)
	if os.IsNotExist(err) {
		// Not an error the model needs a stack trace for: the design workspace
		// may not exist yet (authoring has not produced it — §9), or the path
		// is wrong. Either way "not found" is the correct, actionable answer.
		return map[string]string{"error": fmt.Sprintf("no such design file: %s", args.Path)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read_design: %w", err)
	}
	return map[string]string{"path": args.Path, "content": string(content)}, nil
}

func toolSearchDesign(ctx context.Context, l *toolLoopContext, input json.RawMessage) (any, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return nil, fmt.Errorf("search_design: invalid input: %w", err)
	}
	if strings.TrimSpace(args.Query) == "" {
		return nil, fmt.Errorf("search_design: query is required")
	}

	root := filepath.Join(l.workspaceRoot, l.workflowID.String(), "main")
	type hit struct {
		Path string `json:"path"`
		Line int    `json:"line"`
		Text string `json:"text"`
	}
	var hits []hit
	query := strings.ToLower(args.Query)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil // unreadable entry or a directory: skip, keep walking
		}
		if !strings.HasSuffix(path, ".md") && !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			rel = path
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.Contains(strings.ToLower(line), query) {
				hits = append(hits, hit{Path: rel, Line: i + 1, Text: strings.TrimSpace(line)})
			}
		}
		return nil
	})
	if os.IsNotExist(err) || (err == nil && root == "") {
		return []hit{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("search_design: %w", err)
	}
	if hits == nil {
		hits = []hit{}
	}
	return hits, nil
}

func toolReadBlackboard(ctx context.Context, l *toolLoopContext, input json.RawMessage) (any, error) {
	var args struct {
		EventTypes []string `json:"event_types"`
	}
	// Empty/absent input is valid — it means "all types" — so a parse error is
	// only real when the field IS present and malformed.
	if len(input) > 0 {
		_ = json.Unmarshal(input, &args)
	}

	if len(args.EventTypes) == 0 {
		events, err := l.store.GetSince(ctx, l.workflowID, 0, 50)
		if err != nil {
			return nil, err
		}
		return events, nil
	}
	events, err := l.store.GetByType(ctx, l.workflowID, args.EventTypes, 0)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// toolGetAcceptance and toolReadDecisions read ACCEPTANCE.md / DECISIONS.md from
// the same merged workspace read_design uses. Both files are produced by the
// authoring turn (§9), which is a later step — so today these correctly return
// "not found" rather than failing, exactly like read_design does for any path
// before authoring has run once.
func toolGetAcceptance(ctx context.Context, l *toolLoopContext, input json.RawMessage) (any, error) {
	var args struct {
		Section string `json:"section"`
	}
	if len(input) > 0 {
		_ = json.Unmarshal(input, &args)
	}
	return toolReadDesign(ctx, l, mustJSON(map[string]string{"path": "ACCEPTANCE.md"}))
}

func toolReadDecisions(ctx context.Context, l *toolLoopContext, _ json.RawMessage) (any, error) {
	return toolReadDesign(ctx, l, mustJSON(map[string]string{"path": "DECISIONS.md"}))
}

// mustJSON marshals a small, known-safe literal for delegating one tool's
// implementation to another's. Panics only if a hardcoded map fails to
// marshal, which cannot happen for map[string]string.
func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("tool_loop: mustJSON: %v", err))
	}
	return b
}

// ============================================================
// Mutating tool implementations — every one posts a proposal event and writes
// nothing else. See the file comment for why.
// ============================================================

// toolProposeAmendment records a proposal through proposeAmendment
// (amendment.go), which posts the event AND creates the pending approval row.
//
// It used to post only the event, which meant the chat could describe a change
// that nothing could ever act on — the client had no list to approve from and no
// id to approve by. §7.5's step 1 and step 2 are one write, not two features.
func toolProposeAmendment(ctx context.Context, l *toolLoopContext, input json.RawMessage) (any, error) {
	var args struct {
		Target  string `json:"target"`
		OldText string `json:"old_text"`
		NewText string `json:"new_text"`
		Reason  string `json:"reason"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return nil, fmt.Errorf("propose_amendment: invalid input: %w", err)
	}
	if args.Target == "" || args.NewText == "" || args.Reason == "" {
		return nil, fmt.Errorf("propose_amendment: target, new_text and reason are required")
	}

	// Checked here as well as at apply time, so a model that names a file that
	// does not exist is told immediately rather than after a client has approved
	// something unappliable.
	allowed, err := amendableTargets(ctx, l.sections, l.workflowID)
	if err != nil {
		return nil, fmt.Errorf("propose_amendment: %w", err)
	}
	if !allowed[args.Target] {
		return nil, fmt.Errorf("propose_amendment: %q is not an amendable file — use list_sections, %q or %q",
			args.Target, finalMD, acceptanceMD)
	}

	proposed, err := proposeAmendment(ctx, l.db, l.store, ProposeAmendmentRequest{
		WorkflowID: l.workflowID,
		ExpertID:   l.expert.ID,
		ExpertName: l.expert.Name,
		ChatID:     l.chatID,
		Kind:       amendmentKindStatement,
		Target:     args.Target,
		OldText:    args.OldText,
		NewText:    args.NewText,
		Reason:     args.Reason,
	})
	if err != nil {
		return nil, fmt.Errorf("propose_amendment: %w", err)
	}
	return map[string]any{
		"proposed":    true,
		"approval_id": proposed.ApprovalID,
		"event_id":    proposed.EventID,
		"note": "Recorded as a pending amendment. Nothing is written until the client " +
			"approves it, and the exact old text you gave must still be in the file at that point.",
	}, nil
}

// toolProposeAcceptance builds the criterion block itself rather than leaving it
// to the model.
//
// The id, the owner token and the section path are all DERIVED: the id is
// AC-{section_no}-{next}, the section number comes from
// workflow_design_sections, and the owner token is the expert slug already
// embedded in the section path. None of it is asked for, because all three must
// match what documentCheck (authoring.go) validates and what acceptanceLineRe
// parses — and a model filling in its own id is how two AC-20-01 entries end up
// in one file.
func toolProposeAcceptance(ctx context.Context, l *toolLoopContext, input json.RawMessage) (any, error) {
	var args struct {
		Section   string `json:"section"`
		Statement string `json:"statement"`
		Verify    string `json:"verify"`
		DoneWhen  string `json:"done_when"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return nil, fmt.Errorf("propose_acceptance: invalid input: %w", err)
	}
	if args.Section == "" || args.Statement == "" || args.Verify == "" || args.DoneWhen == "" {
		return nil, fmt.Errorf("propose_acceptance: section, statement, verify and done_when are all required")
	}

	sections, err := l.sections.ListSections(ctx, l.workflowID)
	if err != nil {
		return nil, fmt.Errorf("propose_acceptance: list sections: %w", err)
	}
	sectionNo := 0
	for _, sec := range sections {
		if sec.SectionPath == args.Section {
			sectionNo = sec.SectionNo
			break
		}
	}
	if sectionNo == 0 {
		return nil, fmt.Errorf("propose_acceptance: %q is not a design section in this workflow — use list_sections", args.Section)
	}

	// The next index is read from the file as it stands, so a criterion proposed
	// now does not collide with one an authoring turn wrote a minute ago.
	existing := ""
	if full, err := designPath(l.workspaceRoot, l.workflowID, acceptanceMD); err == nil {
		if data, readErr := os.ReadFile(full); readErr == nil {
			existing = string(data)
		}
	}
	acID := fmt.Sprintf("AC-%d-%02d", sectionNo, acceptanceIDsForSection(existing, sectionNo)+1)

	block := formatAcceptanceBlock(
		acID, ownerTokenFromSectionPath(args.Section), args.Section,
		args.Statement, args.Verify, args.DoneWhen,
	)

	proposed, err := proposeAmendment(ctx, l.db, l.store, ProposeAmendmentRequest{
		WorkflowID: l.workflowID,
		ExpertID:   l.expert.ID,
		ExpertName: l.expert.Name,
		ChatID:     l.chatID,
		Kind:       amendmentKindAcceptance,
		Target:     acceptanceMD,
		OldText:    "", // a new criterion is an append, so there is nothing to replace
		NewText:    block,
		Reason:     fmt.Sprintf("%s adds %s for %s", l.expert.Name, acID, args.Section),
	})
	if err != nil {
		return nil, fmt.Errorf("propose_acceptance: %w", err)
	}
	return map[string]any{
		"proposed":     true,
		"approval_id":  proposed.ApprovalID,
		"event_id":     proposed.EventID,
		"criterion_id": acID,
		"criterion":    block,
		"note":         "Recorded as a pending amendment. Nothing is written until the client approves it.",
	}, nil
}

// toolRaiseConflict handles a disagreement between two experts.
//
// FLOW (doc §16.2):
//
//  1. Post design_conflict_raised so the event is always on the blackboard.
//
//  2. Classify the conflict:
//     - Spine statement (C-NNN in final.md's statement table) → architectural
//       decision → always goes to the client. Rank cannot override a product
//       decision.
//     - Section detail (anything else) → try rank-based auto-resolve.
//
//  3. Auto-resolve: walk expert_categories.authoring_rank ascending (lower
//     number = higher authority). For each candidate:
//     a. Skip the two conflicting experts themselves — they already disagree.
//     b. Check HasKnowledge: does this expert's RAG have relevant chunks
//        (score >= gate1UsableThreshold = 0.40)? If not, skip with a log.
//     c. First candidate that passes → ask it via one LLM call (same as
//        toolAskExpert). Its answer is the resolution.
//     d. Post design_conflict_resolved with the winner's answer.
//
//  4. If no candidate has knowledge → escalate to client with a message
//     explaining why no expert could auto-resolve.
//
// WHY the two conflicting experts are skipped in step 3: they already stated
// their positions. Asking one of them to "resolve" the conflict they raised
// would just return their own position — not a resolution.
func toolRaiseConflict(ctx context.Context, l *toolLoopContext, input json.RawMessage) (any, error) {
	var args struct {
		StatementID string `json:"statement_id"`
		Position    string `json:"position"`
		Reason      string `json:"reason"`
		// OtherExpertID is the expert this one disagrees with.
		// Optional — if absent, auto-resolve still runs but skips only the
		// calling expert.
		OtherExpertID string `json:"other_expert_id"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return nil, fmt.Errorf("raise_conflict: invalid input: %w", err)
	}
	if args.StatementID == "" || args.Position == "" {
		return nil, fmt.Errorf("raise_conflict: statement_id and position are required")
	}

	// Step 1: always post the conflict event first.
	ev, err := l.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:       l.workflowID,
		EventType:        "design_conflict_raised",
		PostedByExpertID: &l.expert.ID,
		Content: map[string]any{
			"chat_id":         l.chatID,
			"statement_id":    args.StatementID,
			"my_position":     args.Position,
			"reason":          args.Reason,
			"other_expert_id": args.OtherExpertID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("raise_conflict: post event: %w", err)
	}

	// Step 2: classify — spine statement or section detail?
	// A spine statement ID starts with "C-" (e.g. C-014). Everything else
	// is a section-level detail. This is the deterministic classifier from
	// §16.2: location, not an LLM judgement.
	isSpineStatement := strings.HasPrefix(strings.TrimSpace(args.StatementID), "C-")
	if isSpineStatement {
		// Architectural decision — always goes to the client.
		approvalID, escErr := escalateConflictToClient(ctx, l, ev.ID, args.StatementID,
			"spine statement — architectural decision requires client approval",
			nil,
		)
		if escErr != nil {
			return nil, fmt.Errorf("raise_conflict: escalate spine: %w", escErr)
		}
		return map[string]any{
			"raised":      true,
			"event_id":    ev.ID,
			"escalated":   true,
			"approval_id": approvalID.String(),
			"note":        "This is a spine-level architectural statement (C-NNN). It requires client approval — rank cannot override a product decision.",
		}, nil
	}

	// A19 deadlock/cycle guard: if this exact statement_id has already been
	// raised maxConflictHopsPerStatement times (this raise included), a
	// further rank-based delegation would just repeat the same ping-pong
	// instead of converging. Escalate straight to the client — a
	// deterministic default action, not another LLM judgement.
	// Fail-open on a read error: a transient blackboard read failure must
	// not force an escalation for what may be a genuine first-time conflict.
	if raisedCount, ok := conflictRaisedCountForStatement(ctx, l, args.StatementID); ok &&
		raisedCount >= maxConflictHopsPerStatement {
		approvalID, escErr := escalateConflictToClient(ctx, l, ev.ID, args.StatementID,
			fmt.Sprintf("statement %s raised %d times without converging — deadlock guard (A19)", args.StatementID, raisedCount),
			nil,
		)
		if escErr != nil {
			return nil, fmt.Errorf("raise_conflict: escalate after hop limit: %w", escErr)
		}
		return map[string]any{
			"raised":      true,
			"event_id":    ev.ID,
			"escalated":   true,
			"approval_id": approvalID.String(),
			"note":        fmt.Sprintf("This statement has been raised %d times without resolving. Escalated to client instead of delegating again (deadlock guard).", raisedCount),
		}, nil
	}

	// Step 3: section detail — try rank-based auto-resolve.
	// Build the skip set: the two conflicting experts.
	skipIDs := map[uuid.UUID]bool{l.expert.ID: true}
	if args.OtherExpertID != "" {
		if otherID, parseErr := uuid.Parse(args.OtherExpertID); parseErr == nil {
			skipIDs[otherID] = true
		}
	}

	// Load all workflow experts with their category rank, ordered by rank ASC
	// (lower number = higher authority). NULL rank comes last.
	type rankedExpert struct {
		expert workflowExpert
		rank   *int // nil when authoring_rank IS NULL
	}
	rows, queryErr := l.db.Query(ctx,
		`SELECT e.id, e.name, e.domain,
		        COALESCE(e.reasoning_charter, ''),
		        COALESCE(e.loop_pattern, 'ota'),
		        COALESCE(e.max_loop_iterations, 5),
		        COALESCE(e.allowed_tools, '[]'::jsonb),
		        ec.authoring_rank
		 FROM workflow_experts we
		 JOIN experts e ON e.id = we.expert_id
		 LEFT JOIN expert_categories ec ON ec.id = e.category_id
		 WHERE we.workflow_id = $1
		   AND e.is_active = TRUE AND e.deleted_at IS NULL
		 ORDER BY ec.authoring_rank ASC NULLS LAST`,
		l.workflowID,
	)
	if queryErr != nil {
		// DB error — fall back to client escalation rather than failing the tool.
		approvalID, escErr := escalateConflictToClient(ctx, l, ev.ID, args.StatementID,
			fmt.Sprintf("could not load ranked experts (db error: %s)", queryErr.Error()),
			nil,
		)
		if escErr != nil {
			return nil, fmt.Errorf("raise_conflict: escalate after rank query fail: %w", escErr)
		}
		return map[string]any{
			"raised":      true,
			"event_id":    ev.ID,
			"escalated":   true,
			"approval_id": approvalID.String(),
			"note":        fmt.Sprintf("Could not load ranked experts (db error: %s) — escalated to client.", queryErr.Error()),
		}, nil
	}
	defer rows.Close()

	var candidates []rankedExpert
	for rows.Next() {
		var re rankedExpert
		var toolsJSON []byte
		if scanErr := rows.Scan(
			&re.expert.ID, &re.expert.Name, &re.expert.Domain,
			&re.expert.ReasoningCharter, &re.expert.LoopPattern,
			&re.expert.MaxLoopIterations, &toolsJSON, &re.rank,
		); scanErr != nil {
			continue
		}
		if len(toolsJSON) > 0 {
			_ = json.Unmarshal(toolsJSON, &re.expert.AllowedTools)
		}
		candidates = append(candidates, re)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		approvalID, escErr := escalateConflictToClient(ctx, l, ev.ID, args.StatementID,
			fmt.Sprintf("could not read ranked experts (rows error: %s)", rowsErr.Error()),
			nil,
		)
		if escErr != nil {
			return nil, fmt.Errorf("raise_conflict: escalate after rank rows fail: %w", escErr)
		}
		return map[string]any{
			"raised":      true,
			"event_id":    ev.ID,
			"escalated":   true,
			"approval_id": approvalID.String(),
			"note":        fmt.Sprintf("Could not read ranked experts (rows error: %s) — escalated to client.", rowsErr.Error()),
		}, nil
	}

	// Walk candidates in rank order. Skip conflicting experts. Check knowledge.
	conflictTopic := fmt.Sprintf("%s: %s", args.StatementID, args.Reason)
	var skippedMessages []string

	for _, candidate := range candidates {
		if skipIDs[candidate.expert.ID] {
			continue // one of the two conflicting experts — skip
		}

		// Knowledge check: does this expert have relevant training?
		// B4: pass domain so HasKnowledge uses the domain's usable threshold.
		if !l.gates.HasKnowledge(ctx, candidate.expert.ID, conflictTopic, candidate.expert.Domain) {
			msg := fmt.Sprintf("%s (rank %v): skipped — no relevant training chunks for this topic",
				candidate.expert.Name, rankStr(candidate.rank))
			skippedMessages = append(skippedMessages, msg)
			continue
		}

		// This expert has knowledge — ask it to resolve the conflict.
		resolutionPrompt := fmt.Sprintf(
			"Two experts disagree on: %s\n\nStatement ID: %s\nReason for conflict: %s\n\n"+
				"Based on your training, which position is correct and why? Be concise.",
			args.StatementID, args.StatementID, args.Reason,
		)

		gr, gateErr := l.gates.RunGates(ctx, l.workflowID, candidate.expert,
			conflictTopic, []workflowExpert{candidate.expert}, 0)
		if gateErr != nil {
			skippedMessages = append(skippedMessages, fmt.Sprintf(
				"%s: skipped — gate system error: %s", candidate.expert.Name, gateErr.Error()))
			continue
		}

		resp, llmErr := l.gw.Call(ctx, gateway.LLMRequest{
			Model:      gateway.ModelFast,
			WorkflowID: &l.workflowID,
			SystemPrompt: fmt.Sprintf("You are %s, a domain expert in %s. %s\n\n%s",
				candidate.expert.Name, candidate.expert.Domain,
				candidate.expert.ReasoningCharter,
				FormatGateContext(gr)),
			UserPrompt: resolutionPrompt,
			MaxTokens:  500,
		})
		if llmErr != nil {
			skippedMessages = append(skippedMessages, fmt.Sprintf(
				"%s: skipped — LLM error: %s", candidate.expert.Name, llmErr.Error()))
			continue
		}

		// Resolution found — post the event and return.
		_, _ = l.store.Post(ctx, blackboard.PostRequest{
			WorkflowID:       l.workflowID,
			EventType:        "design_conflict_resolved",
			PostedByExpertID: &candidate.expert.ID,
			Content: map[string]any{
				"conflict_event_id": ev.ID,
				"statement_id":      args.StatementID,
				"resolver":          candidate.expert.Name,
				"resolver_rank":     rankStr(candidate.rank),
				"resolution":        resp.Content,
				"skipped_experts":   skippedMessages,
			},
		})
		return map[string]any{
			"raised":          true,
			"event_id":        ev.ID,
			"auto_resolved":   true,
			"resolved_by":     candidate.expert.Name,
			"resolution":      resp.Content,
			"skipped_experts": skippedMessages,
			"note":            "Auto-resolved by rank. Client will see this in DECISIONS.md.",
		}, nil
	}

	// Step 4: no candidate had knowledge — escalate to client.
	approvalID, escErr := escalateConflictToClient(ctx, l, ev.ID, args.StatementID,
		"no expert with sufficient training found to auto-resolve",
		skippedMessages,
	)
	if escErr != nil {
		return nil, fmt.Errorf("raise_conflict: escalate unresolved: %w", escErr)
	}
	return map[string]any{
		"raised":          true,
		"event_id":        ev.ID,
		"escalated":       true,
		"approval_id":     approvalID.String(),
		"skipped_experts": skippedMessages,
		"note":            "No expert with relevant training could auto-resolve this conflict. Escalated to client for manual decision.",
	}, nil
}

// conflictRaisedCountForStatement counts how many design_conflict_raised
// events already exist on this workflow's blackboard for the given
// statement_id (A19 deadlock guard). Returns ok=false on a read/parse
// failure so the caller can fail open instead of misreading "no history"
// as zero.
func conflictRaisedCountForStatement(ctx context.Context, l *toolLoopContext, statementID string) (int, bool) {
	events, err := l.store.GetByType(ctx, l.workflowID, []string{"design_conflict_raised"}, 0)
	if err != nil {
		return 0, false
	}
	count := 0
	for _, ev := range events {
		var c struct {
			StatementID string `json:"statement_id"`
		}
		if json.Unmarshal(ev.Content, &c) != nil {
			continue
		}
		if c.StatementID == statementID {
			count++
		}
	}
	return count, true
}

// escalateConflictToClient records the escalation and opens a client decision
// gate without pausing the workflow.
//
// Two bugs this closes:
//  1. The old path posted design_conflict_escalated with PostedByClient=false
//     and no expert ID, which violates blackboard_events_poster_check
//     (migration 006). The insert was ignored via `_, _ =`, so the event never
//     landed.
//  2. No approval_requests row was created, so the frontend never received an
//     approval_id and never rendered Approve / Request-changes buttons.
//
// Pattern mirrors amendment proposal (amendment.go): createApprovalRequest +
// blackboard events, no PauseForApproval — a chat-time conflict must not freeze
// a design wave in progress.
func escalateConflictToClient(
	ctx context.Context,
	l *toolLoopContext,
	conflictEventID uuid.UUID,
	statementID string,
	reason string,
	skippedExperts []string,
) (uuid.UUID, error) {
	content := map[string]any{
		"conflict_event_id": conflictEventID.String(),
		"statement_id":      statementID,
		"reason":            reason,
	}
	if len(skippedExperts) > 0 {
		content["skipped_experts"] = skippedExperts
	}

	// System escalation → posted_by_client=true (satisfies poster CHECK).
	escEv, err := l.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:         l.workflowID,
		EventType:          "design_conflict_escalated",
		PostedByClient:     true,
		Content:            content,
		ReferencesEventIDs: []uuid.UUID{conflictEventID},
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("post escalated event: %w", err)
	}

	summary := fmt.Sprintf("Design conflict on %s needs your decision: %s", statementID, reason)
	approvalID, err := createApprovalRequest(ctx, l.db, AskClientRequest{
		WorkflowID: l.workflowID,
		GateName:   designConflictGate,
		Summary:    summary,
		ArtifactContent: map[string]any{
			"conflict_event_id":   conflictEventID.String(),
			"escalation_event_id": escEv.ID.String(),
			"statement_id":        statementID,
			"reason":              reason,
			"skipped_experts":     skippedExperts,
			"chat_id":             l.chatID.String(),
			"raised_by_expert_id": l.expert.ID.String(),
			"raised_by_expert":    l.expert.Name,
		},
		CitedEventIDs: []uuid.UUID{conflictEventID, escEv.ID},
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("create approval: %w", err)
	}

	// question_to_client carries approval_id so the kanban SSE / frontend can
	// set approvalGate and render the decision buttons (same contract as
	// Tools.AskClient — without this the if(approvalId) guard never fires).
	_, err = l.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:     l.workflowID,
		EventType:      "question_to_client",
		PostedByClient: true,
		Content: map[string]any{
			"gate_name":   designConflictGate,
			"summary":     summary,
			"approval_id": approvalID.String(),
		},
		ReferencesEventIDs: []uuid.UUID{conflictEventID, escEv.ID},
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("post question_to_client: %w", err)
	}

	return approvalID, nil
}

// rankStr formats a nullable rank for log messages.
func rankStr(r *int) string {
	if r == nil {
		return "unranked"
	}
	return fmt.Sprintf("%d", *r)
}

// toolAskExpert deliberately does NOT call Tools.AskExpert.
//
// Tools.AskExpert (tools.go) posts a question then BLOCKS for up to 60 seconds
// waiting for a review_comment event from a live agent_loop iteration that
// answers it (tools.go's Subscribe call, driven by the design-phase wave loop).
// In a workflow chat there is no second live loop running in parallel to
// produce that answer — the peer expert is not mid-iteration anywhere. Calling
// AskExpert here would not consult a peer; it would silently burn 60 seconds and
// return TimedOut on every single call.
//
// Instead the peer is asked directly with one LLM call, gated the same way the
// primary responder is (RunGates), so a peer's answer inside the tool loop
// obeys the same knowledge rules as the primary answer and is transparent to
// the client per §7.6 ("shows its work").
func toolAskExpert(ctx context.Context, l *toolLoopContext, input json.RawMessage) (any, error) {
	var args struct {
		ExpertID string `json:"expert_id"`
		Question string `json:"question"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return nil, fmt.Errorf("ask_expert: invalid input: %w", err)
	}
	toID, err := uuid.Parse(args.ExpertID)
	if err != nil {
		return nil, fmt.Errorf("ask_expert: invalid expert_id: %w", err)
	}
	if args.Question == "" {
		return nil, fmt.Errorf("ask_expert: question is required")
	}

	peer, err := loadChatExpertByID(ctx, l.db, toID)
	if err != nil {
		return nil, fmt.Errorf("ask_expert: %w", err)
	}

	gr, err := l.gates.RunGates(ctx, l.workflowID, peer, args.Question, []workflowExpert{peer, l.expert}, 0)
	if err != nil {
		return nil, fmt.Errorf("ask_expert: gates: %w", err)
	}

	systemPrompt := fmt.Sprintf(
		"You are %s, a domain expert in %s, briefly answering a peer expert's question.\n\n%s",
		peer.Name, peer.Domain, FormatGateContext(gr),
	)
	resp, err := l.gw.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelFast,
		WorkflowID:   &l.workflowID,
		SystemPrompt: systemPrompt,
		UserPrompt:   args.Question,
		MaxTokens:    500,
	})
	if err != nil {
		return nil, fmt.Errorf("ask_expert: %w", err)
	}
	return map[string]any{"expert": peer.Name, "answer": resp.Content}, nil
}

// loadChatExpertByID reads one expert into workflowExpert, matching
// WorkflowChatService.loadChatExpert's query shape. Duplicated rather than
// exported from chat_service.go: both are 8-line reads of the same six columns,
// and a shared helper would need a receiver on WorkflowChatService that this
// package-level function has no reason to take.
func loadChatExpertByID(ctx context.Context, db *pgxpool.Pool, expertID uuid.UUID) (workflowExpert, error) {
	var e workflowExpert
	var toolsJSON []byte
	err := db.QueryRow(ctx,
		`SELECT id, name, domain,
		        COALESCE(reasoning_charter, ''),
		        COALESCE(loop_pattern, 'ota'),
		        COALESCE(max_loop_iterations, 5),
		        COALESCE(allowed_tools, '[]'::jsonb)
		 FROM experts
		 WHERE id = $1 AND is_active = TRUE AND deleted_at IS NULL`,
		expertID,
	).Scan(&e.ID, &e.Name, &e.Domain, &e.ReasoningCharter,
		&e.LoopPattern, &e.MaxLoopIterations, &toolsJSON)
	if err != nil {
		return workflowExpert{}, fmt.Errorf("%w: %s", ErrExpertNotFound, expertID)
	}
	if len(toolsJSON) > 0 {
		_ = json.Unmarshal(toolsJSON, &e.AllowedTools)
	}
	return e, nil
}

func toolAskClient(ctx context.Context, l *toolLoopContext, input json.RawMessage) (any, error) {
	var args struct {
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return nil, fmt.Errorf("ask_client: invalid input: %w", err)
	}
	if args.Summary == "" {
		return nil, fmt.Errorf("ask_client: summary is required")
	}

	// NOT Tools.AskClient. That call posts the question AND calls
	// Engine.PauseForApproval, which sets the WORKFLOW's status to
	// paused_for_approval. A workflow chat is not the workflow's run state
	// (§6.1's whole point is that the chat is a separate concern) — a client
	// asking a question inside a chat must not freeze the design phase, any
	// wave in progress, or any other chat on the same workflow. Posting the
	// event directly gives the same client-facing escalation without that
	// side effect.
	ev, err := l.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:       l.workflowID,
		EventType:        "question_to_client",
		PostedByExpertID: &l.expert.ID,
		Content: map[string]any{
			"chat_id":   l.chatID,
			"summary":   args.Summary,
			"gate_name": "ad_hoc",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ask_client: %w", err)
	}
	return map[string]any{"asked": true, "event_id": ev.ID}, nil
}
