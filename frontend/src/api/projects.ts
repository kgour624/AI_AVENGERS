import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'
import type { Project } from '@/types/project'

export interface CreateProjectRequest {
  name: string
  description?: string
}

export const getProjects = (options?: { signal?: AbortSignal }) =>
  baseAPI
    .get<ApiResponse<Project[]>>('/api/v1/projects', { signal: options?.signal })
    .then((res) => res.data.data!)

export const getProject = (id: string, options?: { signal?: AbortSignal }) =>
  baseAPI
    .get<ApiResponse<Project>>(`/api/v1/projects/${id}`, { signal: options?.signal })
    .then((res) => res.data.data!)

export const createProject = (req: CreateProjectRequest) =>
  baseAPI.post<ApiResponse<Project>>('/api/v1/projects', req).then((res) => res.data.data!)

export const getProjectExperts = (projectId: string, options?: { signal?: AbortSignal }) =>
  baseAPI
    .get<ApiResponse<Project>>(`/api/v1/projects/${projectId}`, { signal: options?.signal })
    .then((res) => res.data.data!.experts)

export const addProjectExpert = (projectId: string, expertId: string) =>
  // WHY the response type is just {status}, not ProjectExpert: verified
  // against backend-go/internal/project/service.go's AddExpert handler,
  // which responds with `response.OK(c, map[string]string{"status": "expert added"})`
  // - no expert data comes back. Phase 1's version assumed a
  // ProjectExpert object would be returned, which was never true;
  // callers (Phase 4's project-expert-management UI) must refetch the
  // project (getProjectExperts) to see the updated list, not read the
  // add call's response.
  baseAPI
    .post<ApiResponse<{ status: string }>>(`/api/v1/projects/${projectId}/experts`, {
      expertId,
    })
    .then((res) => res.data.data!)

export const removeProjectExpert = (projectId: string, expertId: string) =>
  baseAPI.delete(`/api/v1/projects/${projectId}/experts/${expertId}`).then((res) => res.data)
