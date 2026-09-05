import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'
import type { L3Event } from '@/types/memory'

/**
 * Source: confirmed reachable in cmd/server/main.go's first (real)
 * buildRouter copy - `projects.GET("/:id/timeline", handleGetProjectTimeline(memManager))`,
 * which calls `memManager.GetTimeline(ctx, projectID, 50, 0)`.
 *
 * WHY no pagination params are exposed on this function: the real
 * handler hardcodes limit=50, offset=0 and accepts no query-string
 * parameters at all (verified from main.go directly) - there is
 * nothing for this function to pass through. If pagination is added
 * server-side later, this function's signature will need query params
 * added to match; documented here so that future change has an
 * obvious frontend counterpart to update.
 */
export const getProjectTimeline = (projectId: string, options?: { signal?: AbortSignal }) =>
  baseAPI
    .get<ApiResponse<L3Event[]>>(`/api/v1/projects/${projectId}/timeline`, {
      signal: options?.signal,
    })
    .then((res) => res.data.data!)
