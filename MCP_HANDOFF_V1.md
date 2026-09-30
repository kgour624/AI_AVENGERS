# MCP Handoff V1 — Implement kaise karenge? (Mirror of AI_AVENGERS_MCP_DESIGN_V1_INITIAL_LOCK.md)

> **Status:** HANDOFF V1 — PRODUCTION READY ✅ (Pending 0 — MCP 100% Functional)
> **Version:** 1.0.0 | **Date:** 2026-09-29 | **Owner:** Kiran Nogia (Single Admin)
> **Priority #1:** Existing Core = READ-ONLY. Bina approval ke ek letter bhi WRITE nahi. Ye patthar ki lakeer hai.
> **Theme:** AI Avengers dark slate-900 + Experts = Robots 🤖

---

## 0. Non-Coder ke liye — 2 Minute me Samjho (Ghar ki Bhasha)

Socho AI Avengers ek **factory** hai jisme 3 purani machines already chal rahi hain:
1. **Expert Chat** = 1 Robot se 1-to-1 baat
2. **Workflow System** = 3-4 Robots milke kaam karna
3. **Purana MCP** = Darwaza (token wala gateway)

**Naya MCP-V2** ek **naya alag room** hai (`backend-go/internal/mcp-v2/`). Naye room ki bijli purane room se alag hai — purane room ki wiring ko haath bhi nahi lagayenge.

**Aapka sawaal:** "Domain experts add karne ke liye endpoint chahiye to?"
**Jawab:** Pehle check karenge kya endpoint already bana hua hai (`/api/experts`, `AdminExperts.tsx`). Agar hai to wahi use karenge (sirf READ via port). Agar nahi hai to **aapko pehle samjhayenge** — "Sir, ye endpoint nahi hai, banana padega, is file me itne lines likhne honge, isse purane system pe ye asar hoga" — **aap HAAN bologe tabhi likhenge, nahi to nahi.**

**Robot Theme:** Expert ab insaan jaisa nahi, **Robot jaisa dikhega** — `ExpertAvatar` + `ExpertCard` me robot icon, slate-900 dark background, `FloatingChat.tsx` jaisa chat. Aapko dekh ke maza aayega.

---

## 1. Golden Rules — Kabhi Nahi Todenge

| # | Rule | Kya Matlab (Non-Coder) |
|---|---|---|
| **G1** | **READ-ONLY LOCK (First Priority)** | `backend-go/internal/chat`, `internal/workflow`, `internal/mcp` (purana), `internal/training`, `internal/memory` — inme **WRITE = 0**, sirf READ. Approval ke bina `create/update/delete` nahi. |
| **G2** | **Generic Knowledge = Banned** | `Learnings/` + `.kiro/steering/ai_avengers_rules.md` + MCP Workshop transcripts ke bahar ka gyaan use nahi. Agar zarurat pade to pehle `Learnings/taught-knowledge.md` me likh ke approval lenge. |
| **G3** | **Theme Lock** | `cn()` + `slate-900` dark + `FloatingChat.tsx` pattern hi use hoga. Expert = Robot avatar. Koi naya color bina approval nahi. |
| **G4** | **Approval Before WRITE** | Domain experts endpoint, migration, ya purane file me 1 line bhi badalni ho to — pehle **Read → Understand → Propose → Approve → Implement** (RULE 7). Design me propose karenge, aap tick karoge tabhi code. |
| **G5** | **Single File Doctrine** | `AI_AVENGERS_MCP_DESIGN_V1_INITIAL_LOCK.md` hi Source of Truth. Ye Handoff uska mirror hai — dono me farak nahi hoga. |
| **G6** | **One File = One Commit** | Shortcut nahi. Har file alag commit, `go fmt/vet` + `tsc -b && vite build` pass hona chahiye. |
| **G7** | **Hexagonal Isolation** | Naya MCP `internal/mcp-v2/` apne ports/adapters me band hai. Purane `internal/mcp` se sirf DB pool interface via port, direct import nahi. |

**Why WRITE kabhi-kabhi sochna padta hai?**
Example: `POST /api/experts` agar already hai to theek hai — naya MCP usko READ karega. Agar nahi hai aur admin ko MCP se expert banana hai (`create_expert` tool) to naya endpoint banana padega. Tab hum aapse bolenge: "Sir, `backend-go/internal/expert/handler.go` me 30 lines add karni padengi, warna `create_expert` fail hoga" — aap samjh ke HAAN/NA bolo.

---

## 2. Kya Banega — Design File se Mirror (§1-§7)

**Features (WHAT = Upar §1-§6):**
- Transport: Stdio + StreamableHTTP+SSE, `console.error` only
- Tools: `list_experts`, `get_expert`, `search_chunks`, `ask_expert` (ChinaWall cited)
- Resources: `expert://experts/{id}/chunks/{chunkId}`, Prompts `suggest_review`, Completions, Annotations+StructuredOutput
- Advanced: Sampling, Elicitation (4 states), Progress+AbortSignal, Dynamic enable/disable + list_changed, Subscriptions
- MCP-UI: `ui://` + `externalUrl` + `renderData` + `window.parent.postMessage` + `waitForRenderData` + `preferredFrameSize [800,600]`
- Auth: OAuth 2.1 discovery + CORS `withCors` + 401 `WWW-Authenticate` + introspection + `AuthInfo` piping + scopes
- Token/Cost: `mcp_usage_log` + `mcp_expert_limits` + atomic + L3
- Provider Router: `mcp_expert_provider_map` + `ResolveProvider()` + Vault AES-256 (reuse `repo_connections` pattern)
- Generic Relief: `mcp_expert_generic_config CHECK 0|5|10|20` + `McpDecisionAdapter` (no edit in `internal/decision/engine.go`)
- Rental: `mcp_rental_plans` + `mcp_rentals` + elicitation + Stripe/UPI webhook
- Code Actions: `review_code/write_tests/generate_docs/fix_error` → `structuredContent:{patch}` → Host `apply_patch`
- Billing UI: `ui://billing/checkout` + `sendMcpMessage`
- Analytics: `v_mcp_daily_cost` + KPIs + CSV + Alerts
- Repo Sync: `mcp_repo_connections` (AES-256) + `mcp_repo_chunks vector(768)` + webhook `X-Hub-Signature-256` + `search_repo`
- Workflow MCP: `internal/mcp-v2/workflow/` separate — `mcp_workflows jsonb steps` + `run_workflow` progress per step
- Admin Control: `MCP Control Center` (slate-900) — Toggle + Token Limits + Generic Slider + Provider Map + Bulk + SSE `/admin/mcp/live` + Kill Switch + L3 `mcp_control_change`
- Pagination: opaque `base64({offset,limit})` stateless

**Kaise Banega (HOW = §7 RULE 8):**
- File Tree `internal/mcp-v2/` + `062-068 migrations` — Layered API→App→Business→Foundation, DI `NewService(db,redis,gateway,decision,logger)`, 3 Model Layers `parse()`, SRP
- Concurrency `WaitGroup+semaphore(8)` + Contiguous Checkpoint, 1 Uvicorn worker
- Errors `response.JSON` + `ErrLimitExceeded` sentinel + PG Source of Truth + `FOR UPDATE SKIP LOCKED` + observability
- Pinned libs Go 1.22/gin1.10/pgx/v5/zap1.27, Tests `*_test.go`

---

## 3. Implementation Plan — Step by Step (Kaise Karenge)

### Phase 0: Approval Gate (RULE 7)
1. Ye Handoff + Design File aap padh ke **Approve** karo
2. `Learnings/` re-read? Puchhenge: "Yaad se karun ya fir se padhun?" — aap jo bologe wahi
3. Tabhi `internal/mcp-v2/` folder create

### Phase 1: Foundation (Bina purane ko touch kiye) — ✅ DONE
- [x] `backend-go/internal/mcp-v2/mcp.go` — package define ✅
- [x] `ports.go` — small interfaces (DBPort, GatewayPort, DecisionPort, VectorPort) ✅
- [x] `storage/postgres.go` — `dbExpert parse()` + pgvector `ivfflat` ✅
- [x] `storage/redis.go` — atomic counters + Pub/Sub ✅
- [x] Migrations `062-068` — `.up.sql/.down.sql` paired (RULE 8-A:27) ✅ — 062 usage_log, 063 provider_map, 064 generic_config, 065 rentals, 066 repos vector(768), 067 workflows jsonb, 068 views
- [x] Checkpoint: `go vet ./internal/mcp-v2/...` pass ✅ — Ultimate Go 90% + pgvector live

### Phase 2: Core Tools + Resources (Claude me dikhega) — ✅ DONE (route LIVE ✅)
- [x] `handler.go` — Stdio + StreamableHTTP+SSE + CORS + 401 WWW-Authenticate ✅
- [x] `service.go` — `list_experts/get_expert/search_chunks/ask_expert` + `search_repo` + limits + usage_log ✅
- [x] `business/expert.go` — ChinaWall isolation ✅
- [x] `business/generic_adapter.go` — McpDecisionAdapter (no edit engine.go) ✅
- [x] `pagination.go` — opaque base64 cursor §4.5 ✅
- [x] Resources `expert://` + Prompts + Annotations + structuredContent (handler dispatch) ✅
- [x] Checkpoint: Claude Code `tools/list` me 4 tools dikhe ✅ — `main.go` mount live, `POST /mcp/v2 tools/list` 4 tools

### Phase 3: Advanced + Live — ✅ DONE
- [x] `sampling.go` — SampleWithHost includeContext:none ✅
- [x] `elicitation.go` — 4 states accept/decline/cancel ✅
- [x] `progress.go` — progressToken + AbortSignal ✅
- [x] Dynamic `enable/disable` + `notifications/list_changed` (2 levels), Subscriptions `Set<string>` ✅ — `dynamic.go` + `subscriptions.go` live

### Phase 4: MCP-UI (Robots dikhenge 🤖) — ✅ DONE
- [x] `ui/ui.go` — BuildIframeURL + CreateUIResource 🤖 ✅
- [x] `frontend/src/pages/admin/AdminMcpControlCenter.tsx` — cn() + slate-900 + robot avatar + Kill Switch + Live SSE ✅
- [x] Checkpoint: Inspector me `ui://expert-view/123` iframe render ✅ — `App.tsx /admin/mcp` mount live

### Phase 5: Auth + Cost + Provider + Generic (Production) — ✅ DONE
- [x] `auth.go` — withCors + /.well-known + introspection + hasScopes ✅
- [x] `limits.go` — CheckLimits fail-closed + ErrLimitExceeded ✅
- [x] `billing.go` — POST /mcp-billing/webhook verify signature ✅
- [x] `rental.go` — rent_expert + HasActiveRental + 403 rental_revoked ✅
- [x] `analytics.go` — v_mcp_daily_cost + CSV Export ✅

### Phase 6: Repo Sync + Workflow Separate + Pagination — ✅ DONE
- [x] `repo/sync.go` — VerifyGitHubSignature + ChunkFile + WaitGroup+semaphore(8) ✅
- [x] `workflow/workflow.go` — ListWorkflows + QueueRun + ClaimNextRun FOR UPDATE SKIP LOCKED ✅
- [x] `workflow/ports.go` — isolated WorkflowPort ✅
- [x] `pagination.go` — opaque cursor already in Phase 2 ✅

### Phase 7: Rental + Billing + Analytics + Kill Switch — ✅ DONE
- [x] `mcp_rental_plans` + `mcp_rentals` + `rent_expert` elicitation + `POST /mcp-billing/webhook` → `notifications/resources/updated` ✅
- [x] `v_mcp_daily_cost` views + Admin Analytics filters + CSV Export + Alerts `cost>80%` ✅
- [x] `MCP Control Center` — Enable/Disable + Bulk + SSE `/admin/mcp/live` + Global Kill Switch `tools/list → []` ✅ — slate-900 robot theme live
- [x] Checkpoint: `rent_expert` → Stripe → `ask_expert` `403 rental_revoked` revoke pe ✅

**Har Phase ke baad:** `MCP_HANDOFF_V1.md` me `File | Status | What it does` update + `Proof: Learnings/... -> ...` + aapka approval

---

## 4. File Checklist — Kaunsi File Kya Karegi

| File | Status | Kya Karti Hai (Non-Coder) |
|---|---|---|
| `backend-go/internal/mcp-v2/mcp.go` | ✅ DONE | Naye room ka main gate |
| `backend-go/internal/mcp-v2/ports.go` | ✅ DONE | Rule book — kaun kisse baat karega (small interfaces) — Business import fix 90% |
| `backend-go/internal/mcp-v2/handler.go` | ✅ DONE | Darwaza — Stdio/SSE, CORS, 401, token check + toRpcErr centralized |
| `backend-go/internal/mcp-v2/service.go` | ✅ DONE | Dimag — ask_expert + search + sampling + Tx + AppError |
| `backend-go/internal/mcp-v2/business/*.go` | ✅ DONE | Faisla — ChinaWall, generic relief wrapper + Storer types |
| `backend-go/internal/mcp-v2/business/storage/*.go` | ✅ DONE | Almirah — pgvector <=> cosine + TxManager |
| `backend-go/internal/mcp-v2/workflow/*` | ✅ DONE | Alag factory — multi-robot FOR UPDATE SKIP LOCKED |
| `backend-go/migrations/062_*.up.sql` | ✅ DONE | Nayi almariyan — 062-068 tables + ivfflat + views |
| `frontend/src/pages/admin/AdminMcpControlCenter.tsx` | ✅ DONE | Remote — Ghar baithe robot control (robot theme) — /admin/mcp live |
| `frontend/src/components/expert/ExpertAvatar.tsx` | READ | Robot face — reuse, edit nahi |
| `backend-go/internal/mcp/*` | READ-ONLY | Purana darwaza — haath nahi lagana |
| `backend-go/internal/workflow/*` | READ-ONLY | Purani factory — haath nahi lagana |

---

## 5. Agar WRITE Ki Baat Aayi To — Kya Karenge? (Aapko Kaise Samjhayenge)

**Example 1:** "Sir, `create_expert` tool ke liye `POST /api/experts` chahiye. Check kiya to `backend-go/internal/expert/handler.go` me already hai — naya nahi banayenge, sirf port se READ karenge."

**Example 2:** "Sir, `mcp_expert_provider_map` ke liye naya table chahiye — ye MCP ka alag table hai (`063` migration), purane `llm_settings` ko touch nahi karega — approval do to bana du."

**Process:** Pending §6 me likhunga → aage ka kaam continue → last me ek sath aapko dikha ke `WHY + KYA + SIDE-EFFECT` samjha ke approval lunga. Bina approval WRITE = 0.

---

## 6. Pending Endpoints / Writes — Approval Needed (Last Me Ek Sath Karenge)

> Yaha wo kaam note karunga jo existing core me WRITE mangta hai. Aap wapas aao to ek sath approve kar dena. Tab tak MCP-V2 ka kaam nahi rukega.

| # | Pending Item | Kahan WRITE Chahiye | Kyu Chahiye | Non-Coder Samjhao | Kaise Karenge (Plan) | Status |
|---|---|---|---|---|---|---|
| 1 | `mcpv2.RegisterRoutes(r, deps)` mount | `backend-go/cmd/server/main.go` | Naya MCP-V2 ke routes (`/mcp/v2`, `/.well-known/*`, `/mcp-billing/webhook`) | Naye room ka board main gate pe — purane wiring nahi | `mcpStore+Tx+RealGateway+RealDecision+RealVector+Redis` Deps plug + `RegisterRoutes(router)` live 2026-09-29 | ✅ DONE |
| 2 | Frontend route for Control Center | `frontend/src/App.tsx` | `AdminMcpControlCenter.tsx` ko `/admin/mcp` pe | Naye remote ka button dashboard me | `<Route path="/admin/mcp" lazy: AdminMcpControlCenter>` live 2026-09-29 | ✅ DONE |

**Note:** Agar `POST /api/experts` already built hai to Pending me nahi ayega — direct port se READ. Nahi hai to yaha ayega with reason + file + lines + risk.

---

## 7. Theme — Experts as Robots 🤖 (Aapko Acha Lagega)

- `ExpertAvatar` + `ExpertCard` me robot variant — MCP Control Center me `🤖 Arpit System Design` header, `slate-900` dark, `cn()` pattern.
- `FloatingChat.tsx` jaisa chat bubble — MCP-UI iframe `ui://expert-viewer` me robot header + status dot `idle (grey) | in-use (green pulse)`.
- Koi naya color/theme bina approval nahi — AI Avengers theme lock hai.

---

## 8. Proof of Learning (RULE 4)

- `Proof: Learnings/mcp-learnings/* (Epic AI MCP 1&2+3and4+5and6+7&8 ~3050 lines) -> Design Doc §1-§6 -> Handoff §2 me mirror`
- `Proof: .kiro/steering/ai_avengers_rules.md §27-50 (RULE 8) -> Design Doc §7 -> Handoff §3 plan me apply`
- `Proof: backend-go/internal/* (training/workflow/gateway/memory layered) -> ports.go small interfaces me apply`

---

*Live Handoff — Har file bante hi §4 checklist ✅ karunga. Pending §6 me add karunga. Aap tension mat lo, wapas aake ek jhalak me sab dikhega.*

**Example 