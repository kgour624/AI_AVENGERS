import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'
import type { Expert } from '@/types/expert'
import type { User } from '@/types/auth'

/** Admin-only endpoints. Route guard (AuthGuard) + backend 403 both enforce access - see security notes in FRONTEND_SYSTEM_DESIGN.md section 15. */

export const getAdminExperts = () =>
  baseAPI.get<ApiResponse<Expert[]>>('/api/v1/admin/experts').then((res) => res.data.data!)

export const ingestTranscript = (expertId: string, file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  return baseAPI
    .post<ApiResponse<{ jobId: string }>>(`/api/v1/admin/experts/${expertId}/ingest`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    .then((res) => res.data.data!)
}

export const getIngestionJobStatus = (expertId: string) =>
  baseAPI
    .get<ApiResponse<{ status: string; processedChunks: number; totalChunks: number }>>(
      `/api/v1/admin/experts/${expertId}/jobs`
    )
    .then((res) => res.data.data!)

export const getAdminClients = () =>
  baseAPI.get<ApiResponse<User[]>>('/api/v1/admin/clients').then((res) => res.data.data!)

export interface AdminStats {
  totalExperts: number
  totalClients: number
  totalProjects: number
  monthlyCostUsd: number
  violationRate7d: number
}

export const getAdminStats = () =>
  baseAPI.get<ApiResponse<AdminStats>>('/api/v1/admin/stats').then((res) => res.data.data!)
