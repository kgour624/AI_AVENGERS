import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'
import type { Expert } from '@/types/expert'
import type { ExpertCategory, CategoryTemplateSchema } from '@/types/category'
import type { DomainProfile } from '@/types/domainProfile'

/**
 * Admin-only endpoints. Route guard (AuthGuard) + backend 403 both
 * enforce access - see security notes in FRONTEND_SYSTEM_DESIGN.md
 * section 15.
 *
 * PHASE 4 CORRECTION: every function below was rewritten after
 * actually reading backend-go/internal/admin/admin_handler.go, not
 * just the design docs' illustrative examples. Concrete bugs found
 * and fixed:
 *
 * 1. ingestTranscript was sending the file under the FormData key
 *    "file" - the real handler reads `c.Request.FormFile("transcript")`.
 *    Wrong key means the backend's IngestTranscript would 400 with
 *    FILE_REQUIRED on every real call. This was flagged as an
 *    unverified assumption in Phase 1's HANDOFF.md entry - now fixed
 *    with the real field name confirmed from source.
 * 2. AdminStats was a completely invented shape (totalExperts,
 *    monthlyCostUsd, violationRate7d) that does not match
 *    AdminHandler.GetStats's real response at all. Real shape has
 *    nested `experts: {total, active}`, a raw `violations` COUNT (not
 *    a rate), and `avg_rating` returned as a STRING (Go's
 *    fmt.Sprintf("%.2f", ...), not a JSON number) - the type below
 *    reflects that exactly, including the string typing for avgRating.
 * 3. getAdminClients was typed to return the auth `User[]` type, which
 *    has role/totpEnabled fields the admin clients list endpoint never
 *    returns, and was missing isActive/lastLogin which it DOES return.
 *    Added a dedicated AdminClient type instead of reusing User.
 * 4. addProjectExpert (projects.ts, separate file) assumed the backend
 *    returns a ProjectExpert object - AddExpert's real handler returns
 *    only {status: "expert added"}. Fixed there too, see that file's
 *    diff in this same commit.
 * 5. Added getViolations/getRatings - endpoints that exist in the real
 *    handler (GetViolations, GetRatings) but were never added to this
 *    file in Phase 1, so Admin Violations/Ratings pages (Phase 4) would
 *    have had no way to fetch their data at all.
 */

export const getAdminExperts = () =>
  baseAPI.get<ApiResponse<Expert[]>>('/api/v1/admin/experts').then((res) => res.data.data!)

// Feature #1 fix (docs bug list): PATCH /admin/experts/:id already
// existed and accepts reasoningCharter (verified against
// admin_handler.go's UpdateExpert - it also accepts name/isActive,
// but description is silently ignored despite being bindable; only
// wiring the field this feature actually needs).
export const updateExpertCharter = (expertId: string, reasoningCharter: string) =>
  baseAPI
    .patch<ApiResponse<{ status: string }>>(`/api/v1/admin/experts/${expertId}`, {
      reasoningCharter,
    })
    .then((res) => res.data.data!)

// RegenerateCharter: calls POST /admin/experts/:id/regenerate-charter
// Used when an expert has blank reasoning_charter (e.g. LLM credits
// ran out during ingestion). Regenerates charter from stored transcript
// without re-processing any chunks. Returns immediately — generation
// happens in background (~30s). Poll getAdminExperts to see result.
export const regenerateCharter = (expertId: string) =>
  baseAPI
    .post<ApiResponse<{ status: string; message: string }>>(
      `/api/v1/admin/experts/${expertId}/regenerate-charter`,
    )
    .then((res) => res.data.data!)

// Bug 1.5 fix (docs bug list): the real backend route
// (POST /admin/experts -> AdminHandler.CreateExpert) was already
// registered, but no frontend function ever called it and no UI
// button/modal existed either - confirmed by reading
// admin_handler.go's CreateExpert directly: it requires name/slug/
// domain (description optional) and returns {id, slug}.
export interface CreateExpertRequest {
  name: string
  slug: string
  domain: string
  description?: string
  // CT-D2: optional expert_categories FK. Omit/undefined = flat-text
  // fallback (CT-L2, nullable at the DB level). snake_case category_id
  // on the wire — baseAPI's snakeifyKeys request interceptor converts
  // this camelCase key automatically, same as every other field here.
  categoryId?: string
  // Migration 006 config fields — all optional, backend defaults apply when omitted.
  // Defaults: modelTier='strong', temperature=0.30, topP=0.50,
  //           loopPattern='react', maxLoopIterations=5, allowedTools=[]
  modelTier?: 'cheap' | 'strong' | 'fast'
  temperature?: number
  topP?: number
  loopPattern?: 'ota' | 'react' | 'plan_execute'
  maxLoopIterations?: number
  allowedTools?: string[]
}

export const createExpert = (req: CreateExpertRequest) =>
  baseAPI
    .post<ApiResponse<{ id: string; slug: string }>>('/api/v1/admin/experts', req)
    .then((res) => res.data.data!)

/**
 * General-purpose PATCH for all expert fields.
 * All fields optional — only provided fields are updated.
 * WHY separate from updateExpertCharter:
 *   updateExpertCharter is a narrow single-field function used by
 *   EditCharterModal. updateExpert is the general PATCH used by
 *   EditExpertConfigModal. Keeping them separate avoids breaking
 *   the charter modal's call site.
 */
export interface UpdateExpertRequest {
  name?: string
  description?: string
  isActive?: boolean
  reasoningCharter?: string
  // Migration 006 config fields
  modelTier?: 'cheap' | 'strong' | 'fast'
  temperature?: number
  topP?: number
  loopPattern?: 'ota' | 'react' | 'plan_execute'
  maxLoopIterations?: number
  allowedTools?: string[]
  trainingStatus?: 'draft' | 'ingesting' | 'trained' | 'deprecated'
  // CT-A4 / 2026-09-08 admin-UI fix: backend's UpdateExpert already
  // accepted category_id (validated against h.categoryReg.Get) - this
  // was simply never wired into the admin UI's request shape. No
  // pointer-to-pointer "explicit clear to NULL" support here, matching
  // this handler's OWN documented convention (see admin_handler.go's
  // UpdateExpert comment on CategoryID: "only supports set-if-present,
  // never explicit-clear") - passing a category ID here changes the
  // expert's category, omitting this field leaves it unchanged.
  categoryId?: string
}

export const updateExpert = (expertId: string, req: UpdateExpertRequest) =>
  baseAPI
    .patch<ApiResponse<{ status: string }>>(`/api/v1/admin/experts/${expertId}`, req)
    .then((res) => res.data.data!)

// replaceExisting=true REPLACES the expert's whole corpus instead of adding to
// it. Default false (append) per the design doc §5.4.
//
// WHY the flag has to exist: append mode dedups on chunk_hash, so re-ingesting
// the same course through a changed chunker adds a second copy with different
// hashes rather than replacing the first — the corpus doubles and retrieval gets
// noisier while the job still reports success. Replacing is the only way to move
// an expert onto new chunking, so the choice is explicit here.
export const ingestTranscript = (expertId: string, file: File, replaceExisting = false) => {
  const formData = new FormData()
  // WHY "transcript", not "file": confirmed against
  // AdminHandler.IngestTranscript's real `c.Request.FormFile("transcript")`.
  formData.append('transcript', file)
  formData.append('replace_existing', replaceExisting ? 'true' : 'false')
  return baseAPI
    .post<ApiResponse<{ jobId: string; status: string; message: string; expertId: string }>>(
      `/api/v1/admin/experts/${expertId}/ingest`,
      formData,
      { headers: { 'Content-Type': 'multipart/form-data' } }
    )
    .then((res) => res.data.data!)
}

export interface IngestionJob {
  id: string
  status: string
  sourcePath: string
  totalChunks: number
  processedChunks: number
  errorMessage?: string
  startedAt?: string
  completedAt?: string
  createdAt: string
  // Migration 007 fields
  currentStage?: string
  stageDetail?: string
  costUsd?: number
  estimatedSecondsRemaining?: number
  resumedFromCheckpoint?: boolean
}

export const getIngestionJobs = (expertId: string) =>
  baseAPI
    .get<ApiResponse<IngestionJob[]>>(`/api/v1/admin/experts/${expertId}/jobs`)
    .then((res) => res.data.data!)

export const resumeIngestionJob = (expertId: string, jobId: string) =>
  baseAPI
    .post<ApiResponse<{ jobId: string; status: string; checkpointStage: string; message: string }>>(
      `/api/v1/admin/experts/${expertId}/jobs/${jobId}/resume`
    )
    .then((res) => res.data.data!)

// retryIngestionJob: the PAUSED-job path (charter LLM failed).
// WHY separate from resume: the backend contract distinguishes
// paused -> /retry from failed -> /resume. The paused branch of the ingestion
// modal used to call /resume, which the backend rejects for paused jobs
// (400 NOT_RESUMABLE) — the whole reason "Resume" appeared to do nothing.
export const retryIngestionJob = (expertId: string, jobId: string) =>
  baseAPI
    .post<ApiResponse<{ jobId: string; status: string; message: string }>>(
      `/api/v1/admin/experts/${expertId}/jobs/${jobId}/retry`
    )
    .then((res) => res.data.data!)

/**
 * T1 (training transparency): one row of the durable ingestion timeline
 * (table ingestion_job_events). Unlike the old browser-only log, these come
 * from the database: they survive a refresh and are identical for every viewer.
 *
 * `kind` is one of:
 *   run_started | stage_started | stage_done | batch_done | chunk_stored |
 *   verified | paused | failed | complete
 *
 * `detail` is kind-specific (batch index, counts, durations, the verification
 * ledger). It stays loosely typed here because the renderer narrows it per
 * kind — inventing a single strict shape would be a lie for at least half of
 * the kinds.
 */
export interface IngestionJobEvent {
  id: string
  jobId: string
  expertId: string
  sequenceNumber: number
  stage: string
  kind: string
  detail: Record<string, unknown>
  createdAt: string
}

// getIngestionJobEvents: timeline history for one job. Needed in addition to
// the SSE stream because the stream closes at a terminal state, so opening the
// modal on an already-finished run has no live feed to replay.
//
// jobId is a QUERY PARAM (not a path segment) — the backend note explains why:
// GET /experts/:id/jobs/stream already occupies that route tree position.
export const getIngestionJobEvents = (expertId: string, jobId: string, after = 0, limit = 200) =>
  baseAPI
    .get<ApiResponse<{ events: IngestionJobEvent[]; lastSequence: number; timeline: boolean }>>(
      `/api/v1/admin/experts/${expertId}/jobs/events`,
      { params: { jobId, after, limit } }
    )
    .then((res) => res.data.data!)

// Feature #7 fix (docs bug list): projectCount/messageCount added -
// previously ListClients returned neither, so there was no data for
// the frontend to show beyond the enable/disable toggle. Both are now
// real backend aggregates (admin_handler.go's ListClients), not
// frontend-side guesses.
export interface AdminClient {
  id: string
  email: string
  fullName: string
  isActive: boolean
  lastLogin?: string
  createdAt: string
  projectCount: number
  messageCount: number
}

export const getAdminClients = () =>
  baseAPI.get<ApiResponse<AdminClient[]>>('/api/v1/admin/clients').then((res) => res.data.data!)

export const updateAdminClient = (clientId: string, isActive: boolean) =>
  baseAPI
    .patch<ApiResponse<{ status: string }>>(`/api/v1/admin/clients/${clientId}`, { isActive })
    .then((res) => res.data.data!)

/**
 * Matches AdminHandler.GetStats's real response exactly (source:
 * backend-go/internal/admin/admin_handler.go). avgRating is a string
 * because the backend formats it server-side with fmt.Sprintf("%.2f", ...)
 * rather than sending a raw float - typing it as `number` here would
 * be technically wrong even though it looks like a number.
 */
export interface AdminStats {
  experts: { total: number; active: number }
  clients: number
  projects: number
  messages: number
  totalChunks: number
  violations: number
  avgRating: string
  llmTotalCalls: number
  llmTotalCost: number
}

export const getAdminStats = () =>
  baseAPI.get<ApiResponse<AdminStats>>('/api/v1/admin/stats').then((res) => res.data.data!)

export interface ViolationEntry {
  id: number
  projectId: string
  expertId?: string
  clientId: string
  eventType: string
  reasoning: string
  createdAt: string
}

export const getAdminViolations = () =>
  baseAPI.get<ApiResponse<ViolationEntry[]>>('/api/v1/admin/violations').then((res) => res.data.data!)

export interface RatingSummary {
  expertName: string
  domain: string
  totalRatings: number
  avgScore: number
  goodRatings: number
  badRatings: number
}

export const getAdminRatings = () =>
  baseAPI.get<ApiResponse<RatingSummary[]>>('/api/v1/admin/ratings').then((res) => res.data.data!)

// Bug 1.6 fix (docs bug list): GetSettings/UpdateSetting endpoints
// already existed on the backend, but no frontend function ever
// called them. Verified against admin_handler.go directly: GET
// returns [{key, value, description, updatedAt}] with `value` as
// arbitrary JSON per key; PATCH /admin/settings/:key expects
// {value: <that key's JSON>}. Nested keys inside `value` (e.g.
// china_wall's reranker_threshold) round-trip through baseAPI's
// camelizeKeys/snakeifyKeys interceptors like everything else, so
// AdminSettings.tsx can work in camelCase throughout.
export interface SystemSetting {
  key: string
  value: Record<string, unknown>
  description: string
  updatedAt: string
}

export const getAdminSettings = () =>
  baseAPI.get<ApiResponse<SystemSetting[]>>('/api/v1/admin/settings').then((res) => res.data.data!)

export const updateAdminSetting = (key: string, value: Record<string, unknown>) =>
  baseAPI
    .patch<ApiResponse<{ status: string; key: string }>>(`/api/v1/admin/settings/${key}`, { value })
    .then((res) => res.data.data!)

// LLM Settings — provider + API keys from admin panel
export interface LLMSettingsResponse {
  activeProvider: string
  availableProviders: string[]
  apiKeysConfigured: Record<string, string> // masked keys
  note: string
}

export const getLLMSettings = () =>
  baseAPI
    .get<ApiResponse<LLMSettingsResponse>>('/api/v1/admin/llm-settings')
    .then((res) => res.data.data!)

export const updateLLMSettings = (req: {
  provider: string
  apiKeys: Record<string, string>
  // CodeCraftAPI per-tier model IDs — camelCase here, snakeifyKeys
  // interceptor in base.ts converts to snake_case before sending:
  //   codecraftapiModelCheap  → codecraftapi_model_cheap
  //   codecraftapiModelStrong → codecraftapi_model_strong
  //   codecraftapiModelFast   → codecraftapi_model_fast
  // Backend json tags match exactly: json:"codecraftapi_model_cheap" etc.
  codecraftapiModelCheap?: string
  codecraftapiModelStrong?: string
  codecraftapiModelFast?: string
  // Cavoti per-tier model IDs — same snakeifyKeys conversion as above:
  //   cavotiModelCheap  → cavoti_model_cheap
  //   cavotiModelStrong → cavoti_model_strong
  //   cavotiModelFast   → cavoti_model_fast
  // Backend json tags match exactly: json:"cavoti_model_cheap" etc.
  // (see UpdateLLMSettings request struct in admin_handler.go).
  cavotiModelCheap?: string
  cavotiModelStrong?: string
  cavotiModelFast?: string
}) =>
  baseAPI
    .post<ApiResponse<{ status: string; provider: string; note: string }>>(
      '/api/v1/admin/llm-settings',
      req
    )
    .then((res) => res.data.data!)

// ============================================================
// USAGE / COST (C5)
// ============================================================
// The backend has aggregated usage since C5 but no screen ever showed it, so
// "what is this costing and where is it going" had no answer in the product.
export interface UsageRow {
  key: string
  label: string
  calls: number
  inputTokens: number
  outputTokens: number
  costUsd: number
}

export type UsageGroupBy = 'tenant' | 'project' | 'expert' | 'model' | 'use_case'

export const getUsage = (groupBy: UsageGroupBy = 'model') =>
  baseAPI
    .get<ApiResponse<{ groupBy: string; rows: UsageRow[] }>>('/api/v1/admin/usage', {
      params: { group_by: groupBy },
    })
    .then((res) => res.data.data!)

export interface UsageBudgetStatus {
  tenantId?: string
  spendUsd: number
  limitUsd: number
  percent: number
  alertThreshold: number
  alert: boolean
  breached: boolean
}

export const getUsageBudgets = () =>
  baseAPI
    .get<ApiResponse<{ global: UsageBudgetStatus | null; tenants: UsageBudgetStatus[] }>>(
      '/api/v1/admin/usage/budgets'
    )
    .then((res) => res.data.data!)

export const getUsageAlerts = () =>
  baseAPI
    .get<ApiResponse<UsageBudgetStatus[]>>('/api/v1/admin/usage/alerts')
    .then((res) => res.data.data!)

// ============================================================
// RELIABILITY (C10)
// ============================================================
export interface ReliabilityComponent {
  name: string
  status: string
  error?: string
}

export interface ReliabilitySLO {
  llmCallsTotal: number
  llmErrorsTotal: number
  availability: number
  errorRate: number
  availabilityTarget: number
  errorBudgetRemaining: number
  status: string
  errorBudgetWindowDays: number
  uptimeSeconds: number
}

export interface ReliabilityStatus {
  status: string
  version: string
  components: ReliabilityComponent[]
  slo?: ReliabilitySLO
  generatedAt: string
}

export interface ReliabilityEvent {
  id: string
  kind: string
  severity: string
  component?: string
  detail: unknown
  createdAt: string
}

export const getReliabilityStatus = () =>
  baseAPI
    .get<ApiResponse<ReliabilityStatus>>('/api/v1/admin/reliability/status')
    .then((res) => res.data.data!)

export const getReliabilityEvents = (limit = 50) =>
  baseAPI
    .get<ApiResponse<ReliabilityEvent[]>>('/api/v1/admin/reliability/events', { params: { limit } })
    .then((res) => res.data.data!)

// ============================================================
// LLM PROVIDER HEALTH (G4)
// ============================================================
// Two questions the LLM Settings screen has to answer: is a provider being
// skipped right now (breaker), and is one of them answering slowly (latency).
export interface BreakerStatus {
  provider: string
  // closed = working · open = skipped until retryAfterSeconds · half_open = one
  // probe is being allowed through to find out whether it recovered.
  state: 'closed' | 'open' | 'half_open'
  consecutiveFailures: number
  openedAt?: string
  retryAfterSeconds?: number
  lastError?: string
}

export interface ProviderLatency {
  provider: string
  callsTotal: number
  samplesInWindow: number
  p50Ms: number
  p95Ms: number
  maxMs: number
  lastMs: number
}

export interface LLMHealthResponse {
  breakers: BreakerStatus[]
  latency: ProviderLatency[]
  note: string
}

export const getLLMHealth = () =>
  baseAPI
    .get<ApiResponse<LLMHealthResponse>>('/api/v1/admin/llm-health')
    .then((res) => res.data.data!)

// ============================================================
// MODEL TOKEN LIMITS
// ============================================================
// Per provider+tier caps on what this system will ask a model for — the same
// keys the gateway selects a model by ('strong' | 'fast' | 'cheap' | '*').
//
// WHY this exists: models differ in how much room they need. A reasoning model
// spends part of its completion budget thinking before it writes any visible
// text, so a small max_tokens returns an EMPTY answer — which is how a workflow
// died at intake with "planning failed ... empty content in response". 0 on any
// field means "not configured": the provider's own maximum applies.
export interface ModelLimit {
  provider: string
  tier: string
  maxInputTokens: number
  maxOutputTokens: number
}

export interface ModelLimitsResponse {
  limits: ModelLimit[]
  providers: string[]
  tiers: string[]
  note: string
}

export const getModelLimits = () =>
  baseAPI
    .get<ApiResponse<ModelLimitsResponse>>('/api/v1/admin/llm-settings/model-limits')
    .then((res) => res.data.data!)

// updateModelLimits upserts the given rows and leaves the rest untouched, so
// saving one edited row does not require resending the whole table.
export const updateModelLimits = (limits: ModelLimit[]) =>
  baseAPI
    .put<ApiResponse<{ saved: number }>>('/api/v1/admin/llm-settings/model-limits', { limits })
    .then((res) => res.data.data!)

// ============================================================
// EXPERT CATEGORIES (CT-D1, CATEGORY_TEMPLATE_HANDOFF.md §8)
// ============================================================
// Matches admin_handler.go's ListExpertCategories/CreateExpertCategory/
// GetExpertCategory/UpdateExpertCategory exactly — read directly from
// source before writing these, not guessed from the design doc alone
// (same discipline as every other function in this file per the
// header comment's "PHASE 4 CORRECTION" note).

export const getExpertCategories = () =>
  baseAPI
    .get<ApiResponse<ExpertCategory[]>>('/api/v1/admin/expert-categories')
    .then((res) => res.data.data!)

export const getExpertCategory = (categoryId: string) =>
  baseAPI
    .get<ApiResponse<ExpertCategory>>(`/api/v1/admin/expert-categories/${categoryId}`)
    .then((res) => res.data.data!)

export interface CreateExpertCategoryRequest {
  name: string
  slug: string
  description?: string
  templateSchema: CategoryTemplateSchema
  /** Omit to use the backend's "java" default (CT-L5). */
  defaultLanguage?: string
  askStructurePermission?: boolean
}

export const createExpertCategory = (req: CreateExpertCategoryRequest) =>
  baseAPI
    .post<ApiResponse<{ id: string; slug: string }>>('/api/v1/admin/expert-categories', req)
    .then((res) => res.data.data!)

export interface UpdateExpertCategoryRequest {
  name?: string
  description?: string
  templateSchema?: CategoryTemplateSchema
  defaultLanguage?: string
  askStructurePermission?: boolean
}

export const updateExpertCategory = (categoryId: string, req: UpdateExpertCategoryRequest) =>
  baseAPI
    .patch<ApiResponse<{ status: string }>>(`/api/v1/admin/expert-categories/${categoryId}`, req)
    .then((res) => res.data.data!)

// ============================================================
// DOMAIN PROFILES (China Wall per-domain config, admin-configurable)
// ============================================================
// Matches admin_handler.go's ListDomainProfiles/GetDomainProfile/
// UpdateDomainProfile exactly — read directly from source before
// writing these (same discipline as every other function in this
// file). Includes maxTokensFlat/maxTokensStructured (2026-09-08
// addition) alongside every other DomainProfile field (Option B scope).

export const getDomainProfiles = () =>
  baseAPI
    .get<ApiResponse<DomainProfile[]>>('/api/v1/admin/domain-profiles')
    .then((res) => res.data.data!)

export const getDomainProfile = (domain: string) =>
  baseAPI
    .get<ApiResponse<DomainProfile>>(`/api/v1/admin/domain-profiles/${domain}`)
    .then((res) => res.data.data!)

// All fields optional — only provided fields are changed on the
// backend (UpdateDomainProfile's *T pointer fields). customRules is
// deliberately NOT included here: admin_handler.go's UpdateDomainProfile
// request struct has no custom_rules field — those are AI-updated based
// on conversation patterns (domain_profile.go's CustomRules doc
// comment), not admin-panel-edited, so there is no PATCH surface for
// them to wire up here.
export interface UpdateDomainProfileRequest {
  gate1Skip?: boolean
  coverageMode?: DomainProfile['coverageMode']
  citationMode?: DomainProfile['citationMode']
  stripMode?: DomainProfile['stripMode']
  systemPromptExt?: string
  domainKeywords?: string[]
  maxTokensFlat?: number
  maxTokensStructured?: number
}

export const updateDomainProfile = (domain: string, req: UpdateDomainProfileRequest) =>
  baseAPI
    .patch<ApiResponse<DomainProfile>>(`/api/v1/admin/domain-profiles/${domain}`, req)
    .then((res) => res.data.data!)

// ============================================================
// CODECRAFTAPI MODEL CATALOG (proxy via backend)
// ============================================================
// WHY proxy: API key must never leave the server.
// Frontend never calls CodeCraftAPI directly.
// Backend GET /admin/codecraftapi/models proxies to CodeCraftAPI /v1/models.

import type { CodeCraftModel } from '@/types/codecraftapi'

export const getCodeCraftModels = () =>
  baseAPI
    .get<ApiResponse<unknown>>('/api/v1/admin/codecraftapi/models')
    .then((res) => {
      const d = res.data.data
      // Backend already extracts the data array from CodeCraftAPI's
      // { object: "list", data: [...] } envelope, so d should be a
      // plain array. This defensive check handles any edge case where
      // the shape is unexpected.
      if (Array.isArray(d)) return d as CodeCraftModel[]
      if (d && typeof d === 'object' && Array.isArray((d as Record<string, unknown>).data)) {
        return (d as Record<string, unknown>).data as CodeCraftModel[]
      }
      return [] as CodeCraftModel[]
    })

// ============================================================
// EMBEDDING SETTINGS
// ============================================================

export interface EmbeddingSettings {
  embeddingProvider: 'sidecar' | 'codecraftapi'
  embeddingModel: string
  availableProviders: string[]
  note: string
}

export const getEmbeddingSettings = () =>
  baseAPI
    .get<ApiResponse<EmbeddingSettings>>('/api/v1/admin/embedding-settings')
    .then((res) => res.data.data!)

export interface UpdateEmbeddingSettingsRequest {
  embeddingProvider: 'sidecar' | 'codecraftapi'
  /** Required when embeddingProvider is 'codecraftapi'. */
  embeddingModel?: string
}

export const updateEmbeddingSettings = (req: UpdateEmbeddingSettingsRequest) =>
  baseAPI
    .post<ApiResponse<{ status: string }>>('/api/v1/admin/embedding-settings', req)
    .then((res) => res.data.data!)

// ============================================================
// MANAGED ACCOUNTS (admin + domain_expert) + expert grants
// ============================================================

export type ManagedRole = 'admin' | 'domain_expert' | 'client'

export interface ManagedAccount {
  id: string
  email: string
  fullName: string
  role: ManagedRole
  isActive: boolean
  totpEnabled: boolean
  lastLogin?: string
  createdAt: string
  expertIds: string[]
  projectCount: number
  messageCount: number
}

export const getManagedAccounts = () =>
  baseAPI
    .get<ApiResponse<ManagedAccount[]>>('/api/v1/admin/accounts')
    .then((res) => res.data.data!)

export interface CreateManagedAccountRequest {
  email: string
  password: string
  fullName: string
  role: ManagedRole
  expertIds?: string[]
}

export const createManagedAccount = (req: CreateManagedAccountRequest) =>
  baseAPI
    .post<ApiResponse<ManagedAccount>>('/api/v1/admin/accounts', {
      email: req.email,
      password: req.password,
      full_name: req.fullName,
      role: req.role,
      expert_ids: req.expertIds ?? [],
    })
    .then((res) => res.data.data!)

// Partial update — only provided fields change. Mirrors the backend's
// UpdateManagedAccountRequest (nil = unchanged).
export interface UpdateManagedAccountFields {
  fullName?: string
  email?: string
  role?: ManagedRole
  isActive?: boolean
  password?: string
}

export const updateManagedAccount = (accountId: string, fields: UpdateManagedAccountFields) =>
  baseAPI
    .patch<ApiResponse<{ status: string }>>(`/api/v1/admin/accounts/${accountId}`, {
      full_name: fields.fullName,
      email: fields.email,
      role: fields.role,
      is_active: fields.isActive,
      password: fields.password,
    })
    .then((res) => res.data.data!)

// Soft-deletes (and disables) an account; removes its expert grants.
export const deleteManagedAccount = (accountId: string) =>
  baseAPI
    .delete<ApiResponse<{ status: string }>>(`/api/v1/admin/accounts/${accountId}`)
    .then((res) => res.data.data!)

export const setAccountExperts = (accountId: string, expertIds: string[]) =>
  baseAPI
    .put<ApiResponse<{ status: string }>>(`/api/v1/admin/accounts/${accountId}/experts`, {
      expert_ids: expertIds,
    })
    .then((res) => res.data.data!)

export const issueBootstrapToken = () =>
  baseAPI
    .post<ApiResponse<{ token: string; message: string }>>('/api/v1/admin/bootstrap-tokens')
    .then((res) => res.data.data!)



// ============================================================
// Ingestion corpus check + repair (Phase D/E)
// ============================================================

/** One ingestion_runs row: what a run actually stored. */
export interface IngestionRunAudit {
  jobId: string
  sourceFile: string
  verificationStatus: 'not_checked' | 'verified' | 'mismatch' | 'repaired'
  mismatchReason: string
  parsed: number
  duplicates: number
  inserted: number
  reused: number
  storedForFile: number
  generalStored: number
  fallbackChunks: number
  nullEmbeddings: number
  createdAt: string
}

/** "Is this expert's corpus what it claims to be?" — numbers plus sentences. */
export interface IngestionAudit {
  expertId: string
  corpusChunks: number
  corpusTopics: number
  nullEmbeddings: number
  generalChunks: number
  declaredChunks: number
  declaredTopics: number
  statsDrift: boolean
  capabilityRows: number
  staleCapabilityRows: number
  missingCapabilities: boolean
  mismatchedRuns: number
  uncheckedRuns: number
  runs: IngestionRunAudit[]
  findings: string[]
}

/** Why a file holds fewer rows than its run parsed. */
export type FileIntegrityVerdict = 'complete' | 'duplicates_merged' | 'tail_missing' | 'unchecked'

export interface IngestionFileIntegrity {
  sourceFile: string
  storedRows: number
  expectedFromLedger: number
  minIndex: number
  maxIndex: number
  distinctIndices: number
  missingIndices: number
  shortfall: number
  tailMissing: boolean
  nullEmbeddings: number
  generalChunks: number
  fallbackChunks: number
  verdict: FileIntegrityVerdict
}

export interface IngestionDiagnostics {
  jobId: string
  expertId: string
  runs: IngestionRunAudit[]
  files: IngestionFileIntegrity[]
  findings: string[]
}

export type ReconcileAction = 'expert_stats' | 'capabilities' | 'embeddings' | 'resolve_run'

export interface ReconcileActionResult {
  action: ReconcileAction
  applied: boolean
  dryRun: boolean
  changed: number
  detail: string
}

export interface ReconcileResponse {
  expertId: string
  dryRun: boolean
  applied: boolean
  results: ReconcileActionResult[]
  availableActions: ReconcileAction[]
}

/** Actions that rebuild derived state. resolve_run is deliberately separate. */
export const RECONCILE_REPAIR_ACTIONS: ReconcileAction[] = [
  'expert_stats',
  'capabilities',
  'embeddings',
]

/**
 * The one judgement action: "this discrepancy is acceptable". Kept apart from the
 * repairs because it is a human decision, not a derived-state fix.
 */
export const RECONCILE_RESOLVE_ACTIONS: ReconcileAction[] = ['resolve_run']

export const getIngestionAudit = (expertId: string) =>
  baseAPI
    .get<ApiResponse<IngestionAudit>>(`/api/v1/admin/experts/${expertId}/ingestion/audit`)
    .then((res) => res.data.data!)

export const getIngestionDiagnostics = (expertId: string, jobId: string) =>
  baseAPI
    .get<ApiResponse<IngestionDiagnostics>>(
      `/api/v1/admin/experts/${expertId}/ingestion/diagnostics`,
      { params: { job: jobId } },
    )
    .then((res) => res.data.data!)

/**
 * Runs the requested repairs. The backend defaults dry_run to true, and this
 * wrapper does too — a repair call must be explicit about writing.
 */
export const reconcileIngestion = (
  expertId: string,
  actions: ReconcileAction[],
  dryRun: boolean,
  jobId?: string,
) =>
  baseAPI
    .post<ApiResponse<ReconcileResponse>>(
      `/api/v1/admin/experts/${expertId}/ingestion/reconcile`,
      { actions, dry_run: dryRun, job_id: jobId },
    )
    .then((res) => res.data.data!)

// ============================================================
// Capability measurement (I2)
// ============================================================

/** One topic's measured verdict from an evaluation pass. */
export interface CapabilityEvalTopicReport {
  topic: string
  /** Chunks covering the topic — coverage, not difficulty. */
  coverageChunks: number
  cases: number
  passed: number
  /** 1 definitions, 2 mechanics/trade-offs, 3 failure modes. 0 = not achieved. */
  measuredLevel: number
  failureReasons: string[]
  canHandle: string[]
  cannotHandle: string[]
}

/** How well retrieval did across one pass's questions. */
export interface RetrievalMetrics {
  cases: number
  /** Questions whose source chunk was retrieved in the top k. */
  hits: number
  /** Mean reciprocal rank of those hits. */
  mrr: number
}

/** A whole evaluation pass. */
export interface CapabilityEvalReport {
  runId: string
  expertId: string
  status: 'running' | 'complete' | 'failed'
  topicsTotal: number
  casesTotal: number
  casesPassed: number
  /** Cases whose source chunk was actually retrieved. */
  retrievalHits: number
  /** Cases the judge found supported by the retrieved context. */
  grounded: number
  /** Answerable questions the expert declined. */
  refused: number
  topK: number
  startedAt: string
  completedAt: string | null
  topics: CapabilityEvalTopicReport[]
  findings: string[]
  /**
   * This pass's retrieval metrics, and the previous pass's.
   *
   * WHY both: the findings already state the change in words, but the raw numbers
   * are here so a retrieval change can be judged rather than taken on trust. Absent
   * previous = this is the first pass, not "no change".
   */
  metrics: RetrievalMetrics
  previousMetrics?: RetrievalMetrics
  /**
   * Which retrieval path this pass used. Shown on screen because the comparison above
   * is only ever against a pass with the same value — a reader who does not know the
   * mode cannot judge the delta.
   */
  graphExpansion: boolean
  /** Which retrieval path this run used, alongside graphExpansion. */
  layerPreference?: boolean
  /** Why a failed pass failed, so the screen that offered the button can explain it. */
  errorMessage?: string
  /** Absent until both modes have completed a pass. */
  comparison?: CapabilityModeComparison
  /**
   * The same comparison for the depth preference: plain retrieval against retrieval
   * nudged by the depth the question was asked at. Separate from `comparison` so two
   * experiments are never read as one.
   */
  preferenceComparison?: CapabilityModeComparison
}

/**
 * "Did following the concept links help?" — measured on the questions BOTH modes
 * actually asked, never on the two passes' totals (a pass can legitimately ask a
 * different number of questions, and comparing totals would blame retrieval for what
 * is really a different exam).
 */
export interface CapabilityModeComparison {
  commonCases: number
  withoutHits: number
  withHits: number
  withoutPassed: number
  withPassed: number
  withoutMrr: number
  withMrr: number
  verdict: 'improved' | 'unchanged' | 'regressed' | 'inconclusive'
  detail: string
}

export interface CapabilityEvalResponse {
  expertId: string
  /** false = never measured, which is a different state from measured-and-empty. */
  measured: boolean
  report?: CapabilityEvalReport
}

export interface CapabilityEvalStartOptions {
  topics?: number
  topK?: number
  regenerate?: boolean
  /**
   * Retrieve with concept links. Off = the path production chat uses. The two are
   * recorded separately on each pass, so running both is what makes the comparison
   * attributable instead of a guess.
   */
  graphExpansion?: boolean
  /**
   * layerPreference nudges the ranking toward the depth the question was asked at.
   * Measured, never assumed: run Measure and then this, and the screen compares them
   * on the same stored questions.
   */
  layerPreference?: boolean
}

/** Reads the latest pass. Returns measured:false when there has never been one. */
export const getCapabilityEval = (expertId: string) =>
  baseAPI
    .get<ApiResponse<CapabilityEvalResponse>>(`/api/v1/admin/experts/${expertId}/capability-eval`)
    .then((res) => res.data.data!)

/**
 * Starts a measurement pass. Returns immediately — the pass costs one generation
 * call per topic plus two per question, so it runs in the background and the caller
 * polls getCapabilityEval.
 *
 * dry-run has no equivalent here: measuring is read-only for the corpus, but it
 * does overwrite can_handle / cannot_handle with the measured result.
 */
export const startCapabilityEval = (expertId: string, options: CapabilityEvalStartOptions = {}) =>
  baseAPI
    .post<ApiResponse<{ expertId: string; status: string }>>(
      `/api/v1/admin/experts/${expertId}/capability-eval`,
      {
        topics: options.topics ?? 0,
        top_k: options.topK ?? 0,
        regenerate: options.regenerate ?? false,
        graph_expansion: options.graphExpansion ?? false,
        layer_preference: options.layerPreference ?? false,
      },
    )
    .then((res) => res.data.data!)

// ============================================================
// Concept links (I4)
// ============================================================

/** How two of an expert's topics relate. */
export interface ConceptEdge {
  from: string
  to: string
  relation: string
  rationale: string
}

export interface ConceptGraphResponse {
  expertId: string
  count: number
  edges: ConceptEdge[]
  relations: string[]
}

export interface ConceptExtractResult {
  topics: number
  edges: number
  calls: number
  /** Links the model produced that named a topic or relation that does not exist. */
  rejected: number
}

/** Reads the stored concept links. Cheap and read-only. */
export const getExpertConcepts = (expertId: string) =>
  baseAPI
    .get<ApiResponse<ConceptGraphResponse>>(`/api/v1/admin/experts/${expertId}/concepts`)
    .then((res) => res.data.data!)

/**
 * Asks the model how this expert's topics relate and stores the answer. Bounded (a
 * handful of calls over the topic list), so unlike the capability measurement it
 * returns a result directly.
 */
export const extractExpertConcepts = (expertId: string) =>
  baseAPI
    .post<ApiResponse<{ expertId: string; result: ConceptExtractResult }>>(
      `/api/v1/admin/experts/${expertId}/concepts`,
    )
    .then((res) => res.data.data!.result)

// ============================================================
// Depth layers (I5)
// ============================================================

/**
 * What KIND of content each chunk is. Derived from the course, never generated:
 * 1 what/why, 2 how/trade-offs, 3 failure/edge. A topic with no layer-3 chunks will
 * explain it but cannot answer "what breaks under load".
 */
export interface TopicLayerCoverage {
  topic: string
  definition: number
  mechanics: number
  failure: number
  total: number
}

export interface DepthLayerReport {
  expertId: string
  totalChunks: number
  classified: number
  unclassified: number
  definition: number
  mechanics: number
  failure: number
  topicsWithoutFailure: number
  topicsWithoutMechanics: number
  topics: TopicLayerCoverage[]
  findings: string[]
}

export interface DepthClassificationJob {
  id: string
  expertId: string
  status: 'queued' | 'running' | 'complete' | 'failed'
  total: number
  classified: number
  remaining: number
  calls: number
  rejected: number
  errorMessage: string
  createdAt: string
  updatedAt: string
  completedAt?: string | null
}

export interface DepthLayersState {
  report: DepthLayerReport
  job?: DepthClassificationJob | null
}

export interface DepthClassifyResult {
  classified: number
  /** How many chunks still have no layer — run again to continue. */
  remaining: number
  calls: number
  /** Answers that named a passage or layer that does not exist. */
  rejected: number
}

export const getExpertDepthLayers = (expertId: string) =>
  baseAPI
    .get<ApiResponse<DepthLayersState>>(`/api/v1/admin/experts/${expertId}/depth-layers`)
    .then((res) => res.data.data!)

/**
 * Classifies one bounded batch and reports what is left. Resumable on purpose: it costs
 * model calls, so the admin decides how far to go rather than one click launching an
 * unbounded pass over the whole corpus.
 */
export const classifyExpertDepthLayers = (expertId: string) =>
  baseAPI
    .post<ApiResponse<{ expertId: string; job: DepthClassificationJob }>>(
      `/api/v1/admin/experts/${expertId}/depth-layers`,
    )
    .then((res) => res.data.data!.job)
