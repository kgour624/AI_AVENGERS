import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'
import type { Expert } from '@/types/expert'

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
}

export const createExpert = (req: CreateExpertRequest) =>
  baseAPI
    .post<ApiResponse<{ id: string; slug: string }>>('/api/v1/admin/experts', req)
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
}

export const getIngestionJobs = (expertId: string) =>
  baseAPI
    .get<ApiResponse<IngestionJob[]>>(`/api/v1/admin/experts/${expertId}/jobs`)
    .then((res) => res.data.data!)

export interface AdminClient {
  id: string
  email: string
  fullName: string
  isActive: boolean
  lastLogin?: string
  createdAt: string
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

