import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'

/**
 * Repo integration endpoints.
 *
 * CRITICAL FINDING (read backend-go/internal/repo/service.go and
 * cmd/server/main.go's router registration in full before writing
 * this file): the OAuth-initiation endpoint (`GetOAuthURL`, meant to
 * live at GET /repo/oauth/:provider per the service's own
 * RedirectURL construction) and the OAuth callback endpoint are BOTH
 * completely absent from the actual registered routes in main.go's
 * buildRouter. Only three repo routes are wired:
 *   POST   /projects/:id/repo         (ConnectRepo)
 *   POST   /projects/:id/repo/sync    (SyncRepo)
 *   GET    /projects/:id/repo/status  (GetSyncStatus)
 * ConnectRepo's request body requires the client to already have a
 * plaintext `access_token` (`AccessToken string json:"access_token"
 * binding:"required"`). There is currently NO way to obtain that
 * token through a real OAuth redirect - the button implied by the
 * design doc's wireframe ("Connect Repository" -> OAuth flow) cannot
 * work against the backend as it is actually wired today.
 *
 * DECISION: connectRepo below sends accessToken as a manually-entered
 * Personal Access Token (PAT), not an OAuth-derived token - because
 * that is the only thing that can honestly work against the real,
 * reachable API surface. This is flagged as a significant backend gap
 * in HANDOFF.md (add the OAuth routes, or officially adopt PAT-based
 * connection as the real design), not silently worked around by
 * fabricating a fake OAuth redirect that leads nowhere.
 */
export interface ConnectRepoRequest {
  provider: 'github' | 'gitlab'
  repoUrl: string
  accessToken: string
  defaultBranch?: string
}

export const connectRepo = (projectId: string, req: ConnectRepoRequest) =>
  baseAPI
    .post<ApiResponse<{ connectionId: string; status: string }>>(
      `/api/v1/projects/${projectId}/repo`,
      req
    )
    .then((res) => res.data.data!)

export const syncRepo = (projectId: string) =>
  baseAPI
    .post<ApiResponse<{ status: string }>>(`/api/v1/projects/${projectId}/repo/sync`)
    .then((res) => res.data.data!)

/**
 * Matches Service.GetSyncStatus's real return shape exactly - it
 * returns `{connected: false}` (nothing else) when no connection
 * exists yet, or the full object when one does. Modeled as a union
 * rather than making every field optional, so callers are forced to
 * check `connected` before reading the rest - matching how the real
 * data actually behaves rather than papering over the two shapes.
 */
export type RepoSyncStatus =
  | { connected: false }
  | {
      connected: true
      status: 'pending' | 'syncing' | 'complete' | 'failed'
      repoUrl: string
      repoName: string
      totalChunks: number
      lastSyncAt: string | null
    }

export const getRepoSyncStatus = (projectId: string) =>
  baseAPI
    .get<ApiResponse<RepoSyncStatus>>(`/api/v1/projects/${projectId}/repo/status`)
    .then((res) => res.data.data!)
