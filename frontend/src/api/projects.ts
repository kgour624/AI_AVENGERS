import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'
import type { Project, ProjectExpert } from '@/types/project'

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
  baseAPI
    .post<ApiResponse<ProjectExpert>>(`/api/v1/projects/${projectId}/experts`, { expertId })
    .then((res) => res.data.data!)

export const removeProjectExpert = (projectId: string, expertId: string) =>
  baseAPI.delete(`/api/v1/projects/${projectId}/experts/${expertId}`).then((res) => res.data)
