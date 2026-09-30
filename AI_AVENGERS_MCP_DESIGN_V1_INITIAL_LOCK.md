# AI Avengers — New MCP Server Design Doc — V1 INITIAL LOCK

> **Status:** V1 LOCKED — SOURCE OF TRUTH
> **Version:** 1.0.0 | **Date:** 2026-09-04 | **Owner:** System Design Architect
> **Course Compliance:** 100% Epic AI MCP Workshop (1&2 + 3and4 + 5and6 + 7&8) — ~3050 lines — No Generic Knowledge
> **Architecture Pattern:** Modular Monolith (Hexagonal / Ports & Adapters) — Separate Module
> **Repo Mode:** Local existing project, completely separate module — Zero touch on existing systems

---

## 0. Enforcement Rules — Pathar Ki Lakeer

| # | Rule | Enforcement |
|---|---|---|
| R1 | **Read-Only Lock** | Expert Chat System, Workflow System, Existing MCP Server (`mcp_tokens` gateway) — NO write, NO edit, not even 1 letter without explicit admin permission |
| R2 | **Separate Module** | New MCP lives in isolated hexagonal module `backend-go/internal/mcp-v2/` with own ports/adapters. No imports from old MCP except shared DB pool if needed via port interface |
| R3 | **Course-Only** | Every feature MUST map to workshop transcript (1&2, 3and4, 5and6, 7&8). No generic MCP knowledge allowed. Design Doc is audited against transcripts |
| R4 | **Design First** | This doc is Source of Truth. Handoff file `MCP_HANDOFF_V1.md` MUST mirror this doc 1:1. Every completed feature updates handoff |
| R5 | **Single Responsibility** | One component = one file = one commit. No shortcut |
| R6 | **Continuation Rule (Line 207+)** | Sections 3.7-3.15 above are STUBS — full verbatim spec is in §4.2 Continuation (Line 207+). Implementer MUST read both stub + continuation together. Ignoring continuation = violation |
| R7 | **Single File Doctrine (No V1/V2 Split)** | This file `AI_AVENGERS_MCP_DESIGN_V1_INITIAL_LOCK.md` is the ONLY Source of Truth for ALL MCP decisions. There is NO separate V1 vs V2 file — all decisions append in SAME file from Line 271+ with continuation markers. Implementer MUST read entire file end-to-end |

---

## 1. Product Goal — Why This MCP Exists

**Primary Mission (Admin Locked):**
> Domain Experts trained on transcripts (e.g., Arpit System Design) are currently locked inside AI Avengers. Via MCP they must be usable inside **Claude Code / Cursor** as native tools so that Claude can **write code, review code, write tests, and collaborate** using cited expert knowledge.

**User:** Single admin (Kiran Nogia) + Claude Code/Cursor as MCP Client
**Trigger:** `tools/call` from Claude (user prompt like "Ask Arpit expert to review this PR")
**Value:** Claude stops hallucinating, uses domain-constrained cited knowledge for code tasks

**V1 is Mission-Critical + Future-Proof:** V1 already covers 100% mission + advanced features that make it production-grade (UI, live updates, full OAuth).

---

## 2. Architecture Overview — Hexagonal

```
[Claude Code / Cursor] --STDIO/StreamableHTTP+S SSE--> [MCP-V2 Module] --Port--> [Core Domain: Experts, Chunks, pgvector]

MCP-V2 Module (Hexagonal)
├── Primary Ports (Driven by Claude): Tools, Resources, Prompts, Completions
├── Secondary Ports (Drives Host): Sampling, Elicitation, Notifications
├── UI Port: MCP-UI (externalUrl + renderData + postMessage bridge)
├── Auth Port: OAuth 2.1 (introspection, scopes)
└── Adapters: DB Adapter (pgx+pgvector), ML Sidecar Adapter (bge), Worker Adapter (Cloudflare/Stdio)
```

**Locked Decision:**
- WHEN new MCP is built DO isolate in `internal/mcp-v2/` with hexagonal ports BECAUSE existing Expert Chat/Workflow/MCP are read-only and modular monolith forbids tight coupling EXCEPT shared Postgres pool via interface

---

## 3. V1 Initial Lock — Capability Map

### 3.1 Transport Layer

| Feature | Spec | V1 Lock |
|---|---|---|
| Stdio | `StdioServerTransport` via `child_process.spawn` stdin/stdout | YES — For Claude Code/Cursor local |
| Streamable HTTP + SSE | POST /mcp + SSE after `initialize` for server-initiated `sampling/elicitation/notifications` | YES — For remote/hosted mode |
| Rule | `console.log` forbidden on stdio, use `console.error` | LOCKED |

**WHEN** Host is Claude Code local DO use STDIO BECAUSE it is zero-config EXCEPT when remote team needs OAuth THEN use Streamable HTTP

### 3.2 Core — Fundamentals (1&2)

**A. Tools (Model-Controlled) — Claude decides when to call**
- `list_experts` — List active experts (name, domain, slug) — no input — returns `resource_link[]`
- `get_expert` — `id: z.string()` — returns expert charter + capability
- `search_chunks` — `query: z.string(), expertId: z.string(), topK?: z.number()` — semantic search via pgvector — returns `embedded resource[]`
- `ask_expert` — `expertId: z.string(), question: z.string()` — China Wall enforced, returns cited answer with `structuredContent`
- Every tool has `title, description (when to use + example), inputSchema (Zod), annotations, outputSchema`

**B. Resources & Templates (App-Controlled)**
- `expert://experts` — list all experts (when list small)
- `expert://experts/{expertId}` — single expert
- `expert://experts/{expertId}/chunks/{chunkId}` — single chunk (template with `list: undefined` for large sets)
- `list` callback only for small sets (<30 tags), NOT for chunks (1000+)
- `text` vs `blob` — chunks as `application/json` text, images as base64 if needed

**C. Prompts (User-Controlled Slash)**
- `suggest_review` — `argsSchema: {expertId: z.string(), filePath: z.string()}` — returns `messages: [{role:"user", content:{type:"text", text:"Review {{file}} as {{expert}} citing chunks"}}]`

**D. Completions**
- `completable(z.string(), async (value) => filter+slice(0,100))` for `expertId`, `chunkId`
- `complete: { expertId: async (value)=>... }` for templates

**E. Annotations & Structured Output**
- `annotations: {readOnlyHint:true, openWorldHint:false}` for search/list, `{destructiveHint:false, idempotentHint:true}` for ask
- `outputSchema: {expert: expertSchema}` + `return {content:[{type:"text", text: JSON.stringify(structuredContent)}], structuredContent}` — backward compat

**F. Error Handling**
- `throw new Error` -> SDK sets `isError:true` for LLM retry; manual `isError:false` for `canceled`

### 3.3 Advanced (3&4) — Not in Old MCP

**A. Sampling — Server borrows Claude's LLM**
- Flow: `ask_expert` after DB fetch -> `server.server.createMessage({messages, systemPrompt, maxTokens:100})` -> Host shows Allow -> LLM returns JSON -> Server parses + logs
- Check: `getClientCapabilities()?.sampling` else `sendLoggingMessage`
- `void sampling()` fire-and-forget, `includeContext:"none"` never `allServers`

**B. Elicitation — Confirmation Form**
- Use for `delete_expert` etc: `elicitInput({message:"Are you sure delete expert X?", requestedSchema: {type:"object", properties:{confirmed:{type:"boolean"}}, required:["confirmed"]}})` -> 4 states `accept+true/accept+false/decline/cancel`

**C. Progress + Cancellation — Long Tasks**
- `review_codebase` (30s) -> `progressToken` from `_meta` -> `sendNotification({method:"notifications/progress", params:{progressToken, progress, total:1}})` + `signal: AbortSignal` -> `signal.addEventListener("abort", ()=>kill)`

### 3.4 Dynamic & Live (3&4)

- **Dynamic:** `prompt.disable()/enable()` when `experts.length===0` -> auto `notifications/prompts/list_changed` + same for tools/resources
- **Subscriptions:** `SubscribeRequestSchema` handler + `Set<string>` (or Durable Object storage) -> `db.subscribe` -> `notification({method:"notifications/resources/updated", params:{uri}})`
- **List Changed 2 Levels:** TYPE available vs INSTANCE list (template `list` callback) — notify on every create/delete for list-backed templates

### 3.5 MCP-UI (5&6) — Biggest Gap

- **External URL + Dynamic BaseUrl:** `new URL("/ui/expert-viewer", baseUrl).toString()` where `baseUrl = new URL(request.url).origin` piped via `agent.fetch(request, {props:{baseUrl}})` -> `this.props.baseUrl` with `invariant`
- **RenderData Secure Flow:** `return {content:[createUIResource({uri:"ui://expert-view/123", content:{type:"externalUrl", iframeUrl}, encoding:"text"})], _meta:{ui:{renderData: expert}}}` -> Host stores -> Iframe `waitForRenderData({schema})` -> `window.parent.postMessage({type:"ui-lifecycle-iframe-ready"})` -> Host `postMessage({type:"ui-lifecycle-iframe-render-data", payload:{renderData}})` -> Zod parse. Direct `/ui/expert-viewer` shows only spinner.
- **Lifecycle:** `useEffect(()=>{window.parent.postMessage({type:"ui-lifecycle-iframe-ready"}, "*")},[])` mandatory, `window.parent` not `window`
- **PostMessage Bridge:** Generic `sendMcpMessage<T extends "link"|"tool"|"prompt">(type, payload, {schema?})` with `crypto.randomUUID()` correlation, dual filter `type==="ui-message-response" && messageId===id`, `removeEventListener` to avoid leak, Zod cross-boundary, TS overrides
- **Remote DOM:** `document.createElement("ui-stack")/ui-text/ui-button` with `framework:"react"` + `root.appendChild`
- **Sizing:** `uiMetadata:{preferredFrameSize:[800,600]}` + runtime `ui-size-change` via `ref.current.clientHeight` (not scrollHeight)

### 3.6 Auth — OAuth 2.1 (7&8) — Production

- **Discovery:** `/.well-known/oauth-protected-resource` and `/.well-known/oauth-protected-resource/mcp` + `/.well-known/oauth-authorization-server` (proxy to auth server for legacy clients)
- **CORS:** `withCors` for `/.well-known/*` -> `Access-Control-Allow-Origin:*`, `Allow-Methods: GET, HEAD, OPTIONS`, `Allow-Headers: MCP-Protocol-Version`
- **401 Trigger:** `if(!headers.has("authorization")) return 401 { "WWW-Authenticate": `Bearer realm="AI Avengers", resource_metadata="${new URL("/.well-known/oauth-protected-resource/mcp", request.url).toString()}"` }`
- **401 Enhanced:** `hasAuth ? error="invalid_token" + error_description : null` via `.filter(Boolean).join(", ")`
- **Introspection:** `POST /oauth/introspect` with `application/x-www-form-urlencoded` body `token`, `authHeader.replace(/^Bearer\s+/i,"")`, Zod `introspectResponseSchema = z.object({active:z.boolean(), client_id:z.string(), scope:z.string(), sub:z.string()})` + discriminated union for `active:false`, return `null` if not `active` or `!response.ok`
- **Resolve AuthInfo:** `type AuthInfo = SDKAuthInfo & {extra:{userId:string}}` -> `{token, clientId, scopes: scope.split(" "), extra:{userId: sub}} satisfies AuthInfo`
- **Props Piping:** `resolveAuthInfo(request)` -> `agent.fetch(request, {props:{authInfo, baseUrl}})` -> `class MCP extends McpAgent<{authInfo}> { get authInfo(){invariant(this.props?.authInfo)} }` -> `getDbClient(authToken)`
- **Scopes:** `SUPPORTED_SCOPES = ["expert:read","chunk:read","expert:write"] as const` + `validateScopes(authInfo, scopes)` + `agent.hasScopes(...scopes)` -> protect tools/resources/prompts/sampling; `MINIMAL_SCOPE_COMBINATIONS` -> `403 insufficient_scope` with `WWW-Authenticate: error="insufficient_scope", error_description="Need one of: ..."`
- **WhoAmI:** `requireUser()` via token-aware DB -> `whoami` tool + `expert://user` resource

---

## 4. Out of Scope for V1 — UPDATED (Single File Doctrine @ Line 271+)

> **UPDATE:** GitHub/GitLab Repo Sync and Workflow MCP are NO LONGER out-of-scope — they are IN-SCOPE and documented in §4.3 and §4.4 below (Line 271+). Only truly deferred items remain.

- Expert Chat UI reuse — existing Expert Chat untouched (MCP uses its own MCP-UI `ui://` iframes, no reuse — clarified per doubt)
- ~~Pagination `cursor` — spec pending, not needed~~ → MOVED TO §4.5 IN-SCOPE (Line 271+ — cursor pagination locked)
- ~~Workflow orchestration — stays in existing Workflow System~~ → MOVED TO §4.4 IN-SCOPE (separate `internal/mcp-v2/workflow/` module, see Line 271+)
- ~~GitHub/GitLab repo sync — future V2~~ → MOVED TO §4.3 IN-SCOPE (end-to-end, see Line 271+)

### 4.1 Detailed Backlog — V2+ (Documented in V1 file for traceability, NOT in scope for V1 implementation)

> **Note:** Line 162 se aage ka saara content V1 me implement nahi hoga — ye backlog ke roop me isi file me lock kiya ja raha hai taaki Design Doc hi single Source of Truth rahe. Implementation V2/V3 me hogi.

#### 3.7 Admin Control (Full — Single Admin Kiran Nogia) — ⏩ CONTINUATION in §4.2 (Line 207+) — MUST READ BELOW
- WHEN admin action DO require `admin:write` scope + JWT `role=admin` BECAUSE single-admin finance-grade ownership EXCEPT read-only expert browsing which needs `expert:read`
- Tools: `create_expert`, `update_expert`, `delete_expert` (with elicitation confirm), `ingest_transcript` — all gated by `validateScopes`
- Audit: every admin `tools/call` -> `master_event_log` append-only + `mcp_audit` table

#### 3.8 Security — China Wall + Read-Only Enforcement — ⏩ CONTINUATION in §4.2 (Line 207+) — MUST READ BELOW
- Reuse existing 4-layer China Wall + 5-gate DecisionEngine (ASK|WARN|PUSH_BACK|REFUSE|ADVISE) via Port interface — NO mutation of existing module
- Cross-expert isolation: `expertId` scoped pgvector search (`WHERE expert_id = $1`) — no leakage
- Read-only lock: MCP-V2 has no write access to `chats/messages/project_memory` — only `experts/course_chunks` read via DB Port

#### 3.9 DecisionEngine Generic Relief (5-10% — Wrapper Only) — ⏩ CONTINUATION in §4.2 (Line 207+) — MUST READ BELOW
- WHEN expert has no chunks for query DO wrapper calls DecisionEngine gate-2 with `relaxed threshold 0.30` BECAUSE 5-10% generic relief allowed EXCEPT never bypass China Wall or cite fake chunk
- Implementation: `GenericReliefAdapter` wraps existing engine, no edit in `backend-go/internal/decision/` — hexagonal port only
- Flag in `system_settings` `generic_relief_enabled` + per-expert `allow_generic` boolean

#### 3.10 Per-Expert Per-Platform Token Tracking — ⏩ CONTINUATION in §4.2 (Line 207+) — MUST READ BELOW
- Table `mcp_usage` (`id, user_id, expert_id, platform: claude|cursor|inspector, tokens_in, tokens_out, created_at`) — separate from `mcp_tokens`
- Every `ask_expert` + `sampling` call logs tokens via `sendLoggingMessage` + DB adapter
- Resource `expert://usage/{expertId}` for admin analytics

#### 3.11 Multi-Provider Routing (cheap|strong|fast) — ⏩ CONTINUATION in §4.2 (Line 207+) — MUST READ BELOW
- Reuse global `llm_settings` + OpenRouter gateway: `cheap:deepseek/deepseek-chat`, `strong:anthropic/claude-3-5-sonnet`, `fast:google/gemini-flash-1.5`
- `ask_expert` param `tier?: cheap|strong|fast` -> gateway `streamWithFailover()` — default `cheap` for search, `strong` for cited answer
- No new LLM integration — port to existing `backend-go/internal/llm/`

#### 3.12 Rental Marketplace (Future) — ⏩ CONTINUATION in §4.2 (Line 207+) — MUST READ BELOW
- Experts as rentable SKUs — `rental_entitlements` table — MCP tool `rent_expert` triggers elicitation + billing
- Out of scope V1 — only interface stub `RentalPort` documented

#### 3.13 Code Actions — Write/Review/Test via Cited Knowledge — ⏩ CONTINUATION in §4.2 (Line 207+) — MUST READ BELOW
- `review_code` tool: `code: z.string(), expertId: z.string()` -> pgvector search -> sampling to generate review with citations -> `structuredContent: {review, citations[]}`
- `generate_tests` similar — both use `progressToken` for long tasks

#### 3.14 MCP-UI Billing & Entitlements — ⏩ CONTINUATION in §4.2 (Line 207+) — MUST READ BELOW
- UI resource `ui://billing` with `externalUrl` + `renderData: {usage, quota}` — same `waitForRenderData` + `postMessage` bridge
- Scope `billing:read` required — `403 insufficient_scope` if missing

#### 3.15 Usage Analytics + Workflow MCP (Separate) — ⏩ CONTINUATION in §4.2 (Line 207+) — MUST READ BELOW
- Analytics `get_usage_stats` tool + `expert://analytics` resource — reads `mcp_usage` + `mcp_audit`
- Workflow MCP: completely separate module `internal/mcp-workflow/` — multi-expert orchestration — NOT mixed with this MCP-V2 — documented here only for traceability

---

### 4.2 CONTINUATION — Detailed Verbatim Spec for §3.7-3.15 (Line 207+ — MUST READ WITH STUBS ABOVE)

> **RULE R6 ENFORCEMENT:** Upar §3.7-3.15 sirf STUBS hain. Neeche ka §4.2 hi verbatim Source of Truth hai. Implementer jab bhi 3.7-3.15 implement kare, MUST read Stub (upar) + Continuation (neeche) dono ek saath. Neeche ko ignore karna = Design Violation.

#### 3.7 Token Usage & Cost Engine (Per Expert Per Platform) — Detailed Spec
- WHEN any ask_expert or sampling call happens DO record tokens via ModelGateway.Call response InputTokens + OutputTokens + CostUSD BECAUSE admin must know per-expert spend EXCEPT never trust LLM estimate, use provider usage field.
- Tables: `mcp_usage_log {id, expert_id, platform:"claude|cursor|inspector", provider:"openrouter|codecraftapi|deepseek", model:"claude-3-5-sonnet|gpt-6|deepseek-chat", tier:"cheap|strong|fast", input_tokens, output_tokens, total_tokens, cost_usd, generic_used bool, generic_percent int, duration_ms, created_at}` + `mcp_expert_limits {expert_id, platform, daily_token_limit, monthly_token_limit, current_daily, current_monthly, blocked_until}` (append to L3 also)
- Limit Enforcement: Before every LLM call if current + estimated > limit → return `{isError:true, text:"Expert token limit exceeded for platform. Contact admin."}` + `notifications/message {level:"warning"}`
- Multi-Platform Isolation: Same expertId can be called from Claude and Cursor in same second. Tokens are platform tagged, not shared. Go atomic counters + DB aggregation hourly.

#### 3.8 Multi-Provider Router + LLM Settings Integration (High Complexity Locked) — Detailed Spec
- Core Rule: AI Avengers already has 3 tiers cheap | strong | fast mapped to providers via system_settings.models and llm_settings. New MCP MUST reuse same mapping, not create parallel config.
- WHEN admin enables multi-provider DO expose 3 tier selectors per MCP expert BECAUSE tiers are source of truth EXCEPT only providers that are active in llm_settings are shown in MCP UI.
- Config Table: `mcp_expert_provider_map {expert_id, platform:"claude|cursor|generic", tier:"cheap|strong|fast", provider:"openrouter|codecraftapi|deepseek|openai", model:"gpt-6|deepseek-chat|claude-3-5-sonnet", api_key_ref:"vault_ref", is_active bool}` — One row per expert per platform per tier. Default inherits from global llm_settings if no row.
- Example: system-design-expert → Claude platform → strong: codecraftapi/gpt-6, cheap: codecraftapi/gpt-4o-mini ; Same expert → Cursor platform → strong: deepseek/deepseek-chat, cheap: deepseek/deepseek-coder . Same expert, different keys, no code change.
- Resolution Flow: ask_expert(expertId, platform) → ResolveProvider(expertId, platform, tier=strong) → checks mcp_expert_provider_map → fallback to system_settings.models[tier] → if fallback provider not enabled in llm_settings → error Provider not enabled in LLM settings (never show disabled provider in dropdown).
- Vault: api_key_ref never stored plaintext. Uses existing encrypted at rest (AES-256) pattern from repo_connections.access_token. Rotation without restart.
- Validation: WHEN provider selected DO check llm_settings[tier].provider == selectedProvider BECAUSE mismatched provider will 401 at call time EXCEPT admin can create override with explicit allow_override=true + reason L3 log.

#### 3.9 Generic Relief Control (Reuse DecisionEngine Gate-2, No Change) — Detailed Spec
- Locked: Reuse DecisionEngine.gate2 (Knowledge Coverage) as-is. No mutation to existing internal/decision/engine.go. New MCP wraps it via adapter.
- Config Table: `mcp_expert_generic_config {expert_id, allow_generic bool, generic_percent int CHECK 0|5|10|20, time_bound_until timestamptz, reason text, set_by uuid, created_at}` — Admin slider per expert.
- Flow: ask_expert → DecisionEngine.Process() → if GateStopped==2 REFUSE and allow_generic==true → Adapter triggers ModelGateway.Call with generic_percent budget: prompt = Original question + "You may use up to {{generic_percent}}% general knowledge outside cited chunks. Tag generic parts [GENERIC:10%]" → Response tagged [DOMAIN:90%][GENERIC:10%] → mcp_usage_log.generic_used=true, generic_percent stored + L3 master_event_log {event_type:"mcp_generic_used", expert_id, generic_percent, reason}.
- Guard: WHEN generic_percent >0 DO require reason + time_bound_until BECAUSE permanent generic defeats domain constraint EXCEPT admin can set foreverwithreason="Emergency fallback" and it shows red badge in UI.
- No Engine Change: Wrapper McpDecisionAdapter.CheckWithGeneric(expert, question, context, genericConfig) internally calls engine.gate2 then optionally second generic LLM call. Existing chat system gate untouched.

#### 3.10 Expert Rental Marketplace (End-to-End) — Detailed Spec
- Tables: `mcp_rental_plans {id, expert_id, name:"Hourly|Daily|PerCall", price_usd, token_limit, generic_allow, allowed_platforms: ["claude","cursor"]}` + `mcp_rentals {id, renter_client_id, expert_id, plan_id, status:"active|expired|revoked", started_at, expires_at, usage_tokens, usage_cost}` + `mcp_rental_usage_log FK to mcp_usage_log`.
- Flow: Admin lists expert for rent (Admin Panel → Expert → Rent Settings → Price, Limit, Platforms) → Renter sees marketplace in Claude/Cursor list_experts with for_rent:true → rent_expert tool (expertId, planId) → elicitation confirm → Stripe/UPI via MCP-UI Billing (3.12) → mcp_rentals active → hasScopes checks rental active before ask_expert.
- Revoke: Admin can revoke instantly → next ask_expert returns 403 rental_revoked + WWW-Authenticate: error="insufficient_scope".

#### 3.11 Code Actions (Power Pack) — Detailed Spec
- Locked: Domain expert knowledge + direct code execution in Host.
- Tools: review_code {expertId, repoUrl|filePath, diff?} → returns cited review comments with annotations: {readOnlyHint:true} + resource_link to fix diff; write_tests {expertId, filePath, framework:"jest|go-test"} → sampling borrows Host LLM to generate tests citing expert patterns; generate_docs {expertId, topic} → prompts with citations; fix_error {expertId, errorLog, filePath} → elicitation confirm before writing file.
- Host Execution: Code Actions never run code inside MCP server. They return structuredContent: {patch: "...diff"} + MCP-UI resource {type:"externalUrl", iframeUrl:"/ui/code-preview?patch=..."} where Host shows Apply patch? button → sendMcpMessage("tool", {toolName:"apply_patch", params:{patch}}) → Host writes file.

#### 3.12 MCP-UI Billing (Direct Monetization) — Detailed Spec
- UI: ui://billing/checkout iframe with sendMcpMessage("tool", {toolName:"create_checkout", params:{expertId, planId}}) + sendMcpMessage("prompt", {promptName:"explain_price", args:{expertId}}) → Host opens Stripe/UPI.
- Backend: POST /mcp-billing/webhook verifies signature → updates mcp_rentals → notifications/resources/updated to renter's mcp://rentals/{id}.
- Every paid call increments mcp_rental_usage_log.cost_usd and checks mcp_expert_limits.

#### 3.13 Usage Analytics (Complete Source of Truth) — Detailed Spec
- Views: v_mcp_daily_cost {date, expert_id, platform, provider, model, tier, calls, total_tokens, generic_calls, cost_usd} + v_mcp_rental_revenue + v_mcp_generic_audit.
- Admin Panel Page: MCP Analytics → Filters: Expert, Platform, Provider, Tier, Date, Generic only → KPIs: Total calls, Total tokens, Total cost, Avg latency, Generic %, Revenue, Top expert by cost/revenue, Platform split (Claude vs Cursor pie), Provider split, Tier split.
- Audit Log: Every row in mcp_usage_log + mcp_rental_usage_log + L3 searchable: Who (model) → Which expert → When → How many % generic → How long → Why (reason from generic config) → Cost. Export CSV.
- Alerts: if daily cost > budget*0.8 → warning notification, if generic % > 15% for 3 days → flag for review (China Wall style).

#### 3.14 Workflow MCP (Fully Separate, Highly Collaborative Like Claude) — Detailed Spec
- Locked: Must NOT touch existing Workflow System. New module internal/mcp-v2/workflow/ with own tables mcp_workflows {id, name, steps: jsonb [{expertId, tool:"ask_expert|review_code", inputTemplate}], status} + mcp_workflow_runs {id, workflow_id, platform, status, steps_results: jsonb}.
- Tools: run_workflow {workflowId, input} → Progress + Cancellation (3.3) → runs step1: system-design-expert → review_code, step2: code-quality-expert → write_tests sequentially with sampling borrowing Host LLM between steps → notifications/progress per step.
- Collaboration: Each step's structuredContent feeds next step's messages. Host sees MCP-UI progress board ui://workflow/{runId} with live ui-size-change updates. No shared state with old workflow.

#### 3.15 Admin Control Plane (Ghar Baithe Baccha Control — Production Grade) — Detailed Spec
- This is the remote-control page for V1 + V2 (Sections 3.7-3.14) without ever bula ke samjhana.
- Page: Admin → MCP Control Center (reuses existing cn() + FloatingChat.tsx + dark slate-900 pattern from facts).
- Controls per Expert row: Enable/Disable toggle (instant prompt.disable()/enable() + list_changed) | Status dot: idle|in-use (platform) | Token Limit inputs (Daily/Monthly per platform) | Generic Slider 0/5/10/20% + Reason + Time-bound (1hr/1day/forever) | Provider Map: Claude→ [Tier→Provider→Model dropdown (only enabled providers)] , Cursor→ [...] | Rent: List for rent | Price | Plan | Code Actions toggles | Billing: View revenue.
- Bulk: Select 10 experts → Bulk Enable/Disable, Bulk set generic %, Bulk set token limit.
- Live: SSE GET /admin/mcp/live shows which expert on which platform is currently answering (from mcp_usage_log where duration is null).
- Kill Switch: Global MCP Kill Switch → all tools/list returns [] + 403 mcp_disabled.
- Every toggle writes L3 master_event_log {event_type:"mcp_control_change", expert_id, field, old, new, reason, set_by} + audit_types for rollback in one click.

### 4.3 GitHub/GitLab Repo Sync — End-to-End (Line 271+ — IN-SCOPE, NOT Future)

> **Doctrine:** Yahi file Source of Truth hai. Repo Sync abhi se banana hai, jahaan need hai wahaan end-to-end.

- **Goal:** Claude/Cursor me expert `review_code` / `write_tests` / `fix_error` karte time live GitHub/GitLab repo ka code + diff directly use ho — bina manual paste ke.
- **Tables (reuse existing `repo_connections` pattern):** `mcp_repo_connections {id, expert_id, provider:"github|gitlab", repo_url, branch, access_token_encrypted AES-256, webhook_secret_ref, last_synced_at, status}` — encryption same as `repo_connections.access_token` (AES-256 at rest), rotation without restart.
- **Connect Flow:** Admin → MCP Control Center → Repo Sync → Connect GitHub/GitLab (OAuth App) → `elicitation` confirm → token encrypted save → `master_event_log {event_type:"mcp_repo_connected"}` + test webhook ping.
- **Sync Flow:** Webhook `POST /mcp-repo/webhook` verifies `X-Hub-Signature-256` / GitLab token → enqueue job → ML sidecar clones shallow → chunk + embed via `bge-base-en-v1.5` → `course_chunks` style `mcp_repo_chunks {repo_connection_id, file_path, chunk_text, embedding vector(768)}` + `ivfflat` index.
- **Tools:** `list_repos {expertId}` → `resource_link[]`; `sync_repo {repoConnectionId, branch?}` → `progressToken` + `notifications/progress` (long task) + `AbortSignal` cancel; `search_repo {expertId, query}` → pgvector `WHERE repo_connection_id = $1` scoped; `review_code` now auto-calls `search_repo` + `search_chunks` then `sampling` to generate cited review with `resource_link` to fix diff.
- **Security:** WHEN repo token used DO scope to `repo:read` only BECAUSE MCP never pushes code EXCEPT `apply_patch` via Host `sendMcpMessage("tool", {toolName:"apply_patch"})` after user Allow in Host — never server-side git push.
- **UI:** `ui://repo-sync/{expertId}` externalUrl + `renderData: {repos, lastSync}` + `waitForRenderData` + `postMessage` bridge, `preferredFrameSize [900,600]`

### 4.4 Workflow MCP — Fully Separate Module (Line 271+ — IN-SCOPE, NOT Deferred)

> **Locked:** Must NOT touch existing Workflow System (`backend-go/internal/workflow/`). New module `internal/mcp-v2/workflow/` with own tables. Yahi file me decision lock — no V2 wait.

- **Architecture:** `internal/mcp-v2/workflow/` hexagonal — Ports: `WorkflowPort {CreateWorkflow, RunWorkflow, CancelRun}` + `WorkflowStorePort`; Adapters: `PostgresWorkflowAdapter`, `MCPSamplingAdapter`. Zero import from old workflow except shared DB pool via interface (R2).
- **Tables:** `mcp_workflows {id, name, description, steps jsonb [{step:1, expertId:"uuid", tool:"ask_expert|review_code|write_tests|generate_docs", inputTemplate:"Review {{input.repoUrl}} as {{expert}}", dependsOn: null|stepId }], status:"active|disabled", created_by, created_at}` + `mcp_workflow_runs {id, workflow_id, platform:"claude|cursor", status:"running|completed|failed|canceled", input jsonb, steps_results jsonb [{step, expertId, tool, output structuredContent, citations[], duration_ms}], started_at, completed_at}` + L3 `master_event_log {event_type:"mcp_workflow_run"}`
- **Tools:** `list_workflows {}` → active list; `get_workflow {workflowId}` → steps; `run_workflow {workflowId, input}` → long task with `progressToken` + `notifications/progress` per step + `signal: AbortSignal` → `signal.addEventListener("abort", ()=> cancelRun)` ; `get_workflow_run {runId}` → resource `mcp://workflow-runs/{runId}`
- **Collaboration Flow (Highly Collaborative Like Claude):** `run_workflow` step1: system-design-expert → `review_code` (sampling borrows Host LLM) → `structuredContent {review, citations}` feeds step2 inputTemplate → step2: code-quality-expert → `write_tests` sequentially. Each step's `structuredContent` is next step's `messages`. No shared state with old workflow.
- **Live UI:** `ui://workflow/{runId}` externalUrl progress board with `renderData: {workflow, run}` + `waitForRenderData` + `ui-size-change` via `clientHeight`, live step dots idle→running→done.
- **Admin Control:** MCP Control Center → Workflows tab → Create/Edit (steps drag-drop), Enable/Disable toggle → `prompt.disable()/enable()` + `notifications/tools/list_changed`, Kill Switch for workflows.
- **WHEN** workflow step fails DO mark `steps_results[step].error` + `isError:true` + continue or abort based on `onError:"continue|stop"` BECAUSE partial results are useful EXCEPT never leave run in hanging `running` state.

### 4.5 Pagination Cursor — Detailed Spec (Line 271+ — IN-SCOPE, Spec Locked)

> **Why Now:** Pehle spec pending isliye deferred tha — ab lock kar diya taaki koi deferred na bache. Yahi file Source of Truth (R7).

- **Where Needed:** `list_experts`, `search_chunks {query, expertId, topK, cursor?}`, `resources/list expert://experts/{expertId}/chunks/*`, `list_workflows` — jaha list >20 ya chunks 1000+.
- **Cursor Format:** Opaque `base64(JSON.stringify({offset:number, limit:number, qhash:string, expertId?:string}))` — never expose raw offset. `limit` default 20, max 50.
- **Flow:** `tools/call list_experts {cursor, limit}` → decode cursor → `SELECT ... LIMIT $1 OFFSET $2` (pgvector `ivfflat` + `ORDER BY score`) → return `{items, nextCursor: offset+limit<total ? btoa({offset:offset+limit}) : undefined, total}` + `structuredContent`. Invalid cursor → `throw new Error("invalid_cursor")` → `isError:true`.
- **WHEN** cursor tampered DO return `400 invalid_cursor` BECAUSE offset skip can bypass scope EXCEPT valid cursor missing → start at 0.
- **UI:** `ui://expert-list` / `ui://chunks` iframe pagination buttons → `sendMcpMessage("tool", {toolName:"list_experts", params:{cursor:nextCursor}})` + `ui-size-change` via `clientHeight`.
- **No New Table:** Cursor stateless — no DB column. Hourly count via `mcp_usage_log` only.

---

## 5. Dependencies & Checkpoints

**Dependencies:** Postgres+pgvector, Redis (L1 if needed), Python ML sidecar (bge-base-en-v1.5 + reranker), OpenRouter gateway, Existing auth DB (read via port)
**Checkpoint V1:** Claude Code `STDIO` -> `initialize` -> `tools/list` shows `list_experts, search_chunks, ask_expert` -> `tools/call ask_expert` returns cited answer with `structuredContent` -> `view_expert` returns `ui://` iframe that renders in Inspector/Goose -> `401` without token triggers well-known flow

---

## 6. Handoff Link

This doc is mirrored by `MCP_HANDOFF_V1.md` — every component completion updates handoff with `File | Status | What it does`. Divergence is not allowed.

---

---

## 7. Implementation Design — HOW TO BUILD (Line 309+ — RULE 8 Compliant) — Added 2026-09-28

> **Answer to your Q:** Upar §1-§6 (Line 1-309) WHAT/FEATURES hai — wo 100% **MCP Workshop (1&2+3and4+5and6+7&8)** par bana hai (Course Compliance), **AI Avengers Rules (RULE 8 coding standards) par pura nahi**. Neeche §7 HOW hai jo 100% `.kiro/steering/ai_avengers_rules.md` RULE 7 + RULE 8 par lock hai. Upar WHAT, neeche HOW — dono milke single Source of Truth (R7).
> **Proof:** Learnings/mcp-learnings/* -> MCP concepts -> §1-§6 me apply | RULE 8 (ai_avengers_rules.md:27-50) -> code standards -> §7 me apply

### 7.0 Workflow (RULE 7 — Read→Understand→Propose→Approve→Implement)
- **Read:** `backend-go/cmd/server/main.go` (route composition), `backend-go/internal/<domain>/` (training/workflow/gateway/memory pattern), `backend-go/migrations/*.up.sql`, `frontend/src/utils/cn.ts` + `FloatingChat.tsx`, `ml-sidecar/main.py`, `Learnings/`
- **Understand:** Layered API->App->Business->Foundation, Domain Vertical Lanes, Postgres Source of Truth, Redis Pub/Sub wakeup only
- **Propose:** Neeche §7.1-7.8 design — tumse explicit approval ke baad hi `internal/mcp-v2/` create + migrations + code
- **Approve→Implement:** One file = one commit (R5), Handoff `MCP_HANDOFF_V1.md` mirror, `go fmt/vet`, `tsc -b && vite build`, `py_compile`

### 7.1 File Tree (RULE 8-A: Structure & Layering — §27-30)
```
backend-go/
├── cmd/server/main.go                 // MODIFY: add mcpv2.RegisterRoutes(r, deps) — composition only (RULE 8-A:27)
├── migrations/
│   ├── 062_mcp_usage_log.up.sql       // mcp_usage_log + mcp_expert_limits + indexes
│   ├── 062_mcp_usage_log.down.sql
│   ├── 063_mcp_provider_map.up.sql    // mcp_expert_provider_map (vault_ref AES-256)
│   ├── 064_mcp_generic_config.up.sql  // mcp_expert_generic_config CHECK 0|5|10|20
│   ├── 065_mcp_rentals.up.sql         // mcp_rental_plans + mcp_rentals + mcp_rental_usage_log
│   ├── 066_mcp_repos.up.sql           // mcp_repo_connections (enc) + mcp_repo_chunks vector(768) ivfflat
│   ├── 067_mcp_workflows.up.sql       // mcp_workflows + mcp_workflow_runs jsonb
│   └── 068_mcp_views.up.sql           // v_mcp_daily_cost etc
└── internal/mcp-v2/                   // NEW — Isolated Hexagonal Module (R2) — NO import from internal/mcp/*
    ├── mcp.go                         // Package mcpv2 — first file = package name (RULE 8-A:30)
    ├── ports.go                       // Small Interfaces at Boundaries (RULE 8-B:32) — Discover, don't design
    ├── handler.go                     // API Layer: Stdio + StreamableHTTP/SSE, withCors, WWW-Authenticate, introspection
    ├── service.go                     // App Layer: Business orchestration (ask_expert, search, review_code)
    ├── business/
    │   ├── expert.go                  // Business Model (core) — no net/http import (RULE 8-A:28)
    │   └── generic_adapter.go         // McpDecisionAdapter.CheckWithGeneric — wraps decision.Engine via port
    ├── storage/
    │   ├── postgres.go                // Storage Model dbExpert/dbChunk (native pgx types) + parse() → business
    │   └── redis.go                   // Atomic counters (mcp_expert_limits) + Pub/Sub notifications
    ├── workflow/                      // Sub-module internal/mcp-v2/workflow/ — own ports/adapters (R2)
    │   ├── workflow.go
    │   ├── ports.go
    │   └── storage.go
    └── ui/                            // MCP-UI assets: externalUrl builder, renderData schemas
frontend/src/pages/admin/
└── AdminMcpControlCenter.tsx          // Reuse cn() + dark slate-900 + FloatingChat pattern (facts)
```
- **Layering:** API (handler.go Gin) -> App (service.go) -> Business (business/*, no net/http) -> Foundation (storage/* pgx/zap) — imports only downwards (RULE 8-A:28)
- **Vertical Lanes:** mcp-v2 never imports training/workflow/gateway; cross-domain only via `ports.go` small interfaces business-to-business (RULE 8-A:29)
- **No common/utils:** No `mcpv2/common` — each pkg purpose-named (RULE 8-A:30)

### 7.2 Ports & Adapters (Hexagonal) + DI (RULE 8-B:31-34)
- **Constructor DI mandatory:** `NewService(db DBPool, redis RedisClient, gateway GatewayPort, decision DecisionPort, logger *zap.Logger) *Service` — no globals (RULE 8-B:31)
- **Small Interfaces (ports.go):**
```go
type DBPort interface { ListExperts(ctx) ([]Expert, error); SearchChunks(ctx, expertID string, embedding []float32, limit int, cursor string) ([]Chunk, string, error) }
type GatewayPort interface { Call(ctx context.Context, tier string, messages []Message) (output string, usage Usage, costUSD float64, err error) } // wraps internal/gateway
type DecisionPort interface { Process(ctx, expert, question, chunks) (GateResult, error) }
type VectorPort interface { Embed(ctx, text string) ([]float32, error) } // ml-sidecar bge-base-en-v1.5
```
- Only input params use interface, returns are concrete structs (RULE 8-B:32)
- **Three Model Layers (RULE 8-B:33):** App `type AskRequest struct { ExpertID string `json:"expertId"` }` -> Business `type Expert struct { ID uuid.UUID; Charter string }` -> Storage `type dbExpert struct { id pgtype.UUID; charter pgtype.Text }` + `func (d dbExpert) parse() (Expert, error) { if !d.id.Valid {return error} }` integrity check each transition
- **SRP (RULE 8-B:34):** handler only HTTP/SSE, service only orchestration, business only ChinaWall check, storage only SQL — TokenLogger separate from LimitEnforcer

### 7.3 Naming (RULE 8-C:35-37)
- Go: `package mcpv2`, exported `ListExperts`, `McpDecisionAdapter`, unexported `resolveProvider`, stage const `StatusActive = "active"` for DB
- TS/React: `AdminMcpControlCenter.tsx` PascalCase, hook `useMcpLive`, API `listExperts()` camelCase verbs
- Python (if sidecar job): `def embed_chunks()` snake_case, `class EmbeddingClient` PascalCase

### 7.4 Concurrency (RULE 8-D:38-40)
- Go `sync.WaitGroup + semaphore chan struct{ cap:8 }` for repo sync `sync_repo` (shallow clone + chunk embed batches) — ordered when dependency, else parallel (RULE 8-D:38)
- **Contiguous Checkpoint:** `sync_repo` concurrent batches finish out-of-order → checkpoint only contiguous completed prefix (file offset) retained (RULE 8-D:39) — same as ingestion_pipeline.go
- Python sidecar stays 1 Uvicorn worker + `ThreadPoolExecutor(max_workers=2)` for embedding/rerank (RULE 8-D:40)

### 7.5 Error Handling & Logging (RULE 8-E:41-45)
- HTTP: `response.JSON(c, 200, data)` + stable codes `invalid_cursor`, `token_limit_exceeded` — service `fmt.Errorf("search chunks: %w", err)` + `zap.String("expert_id", id)` fields (RULE 8-E:41)
- Sentinel: `var ErrLimitExceeded = errors.New("token limit exceeded")` + `ErrRepoSyncCanceled` (intentional abort not crash) (RULE 8-E:42)
- Best-effort vs fail-closed: `mcp_usage_log` append + Redis Pub/Sub best-effort; `mcp_expert_limits` blocked_until update fail-closed (RULE 8-E:43)
- Centralized: middleware logs once, handler never double-logs; unknown → 500 without leak (RULE 8-E:44)
- Python sidecar: `raise HTTPException(status_code=400)` + structured logs (RULE 8-E:45)

### 7.6 System Design & Data (RULE 8-F:46-48)
- **Postgres = Source of Truth** (all mcp_* tables), Redis Pub/Sub = `notifications/resources/updated` wakeup only — durable `mcp_usage_log` append-only + mutable `mcp_expert_limits` hot snapshot (RULE 8-F:46)
- **FOR UPDATE SKIP LOCKED:** `SELECT ... FROM mcp_workflow_runs WHERE status='queued' ORDER BY created_at LIMIT 5 FOR UPDATE SKIP LOCKED` for `run_workflow` workers — mandatory high throughput (RULE 8-F:47)
- **Observability (RULE 8-F:48):** Measure `picked_at - scheduled_at`, `duration_ms`, `input_tokens/output_tokens/cost_usd` per call — never scale on queue length, scale on `tier=strong avg duration` + `provider latency`
- **Indexes:** `mcp_repo_chunks embedding vector(768) ivfflat`, `mcp_usage_log (expert_id, platform, created_at)`, `mcp_workflows status`

### 7.7 Libraries & Tests (RULE 8-G:49-50) — Pinned
- Go 1.22, gin 1.10, pgx/v5, pgvector-go, go-redis/v9, jwt/v5, otp, uuid, crypto, zap 1.27 — no latest (RULE 8-G:49)
- Tests: `service_test.go` with `testing` + mock ports, `handler_test.go` for CORS/401/WWW-Authenticate, `Frontend npm run build` check, `python -m py_compile` for sidecar helper (RULE 8-G:50)

### 7.8 Permissions (RULE 7 + RULES file last line)
- Har naya file/create/update/delete se PEHLE user se approval — `Remember Once Learned` (RULE 6): already learned workshop+RULES so next task पूछूंगा "re-read karun ya yaad se?"
- Proof har task ke baad: `Proof: Learnings/mcp-learnings/* + .kiro/steering/ai_avengers_rules.md §27-50 -> §7.x me apply`

---

*V1 LOCKED — No changes without admin approval. Next upgrades V2, V3 via discussion. §7 added 2026-09-28 as HOW layer — WHAT (§1-§6) + HOW (§7) = Single Source of Truth (R7).*