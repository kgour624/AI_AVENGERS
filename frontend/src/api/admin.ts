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

export const ingestTranscript = (expertId: string, file: File) => {
  const formData = new FormData()
  // WHY "transcript", not "file": confirmed against
  // AdminHandler.IngestTranscript's real `c.Request.FormFile("transcript")`.
  formData.append('transcript', file)
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
}) =>
  baseAPI
    .post<ApiResponse<{ status: string; provider: string; note: string }>>(
      '/api/v1/admin/llm-settings',
      req
    )
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

