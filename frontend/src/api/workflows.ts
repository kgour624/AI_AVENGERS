import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'

// ============================================================
// Workflow types — match backend workflow.Workflow struct
// ============================================================

export interface Workflow {
  id: string
  clientId: string
  projectId: string
  title: string
  status: 'draft' | 'running' | 'paused_for_approval' | 'paused_for_client_input' | 'completed' | 'cancelled' | 'failed'
  currentPhase: 'intake' | 'high_level_design' | 'detailed_design' | 'implementation' | 'qa' | 'handoff' | 'completed'
  selectedExpertIds: string[]
  costBudgetUsd: number
  costSpentUsd: number
  costSoftLimitPct: number
  costHardLimitPct: number
  // genericAllowancePct: 0-30. 0 = experts use trained + peer knowledge only.
  genericAllowancePct: number
  // mode (phase 3D): 'scratch' builds something new; 'existing_codebase' works
  // inside a connected client repository, where the readable files are a
  // human-approved working set.
  mode: 'scratch' | 'existing_codebase'
  createdAt: string
  updatedAt: string
}

export interface KanbanTask {
  id: string
  workflowId: string
  assignedExpertId: string
  expertName: string
  domain: string
  title: string
  description: string
  status: 'todo' | 'in_progress' | 'under_review' | 'blocked' | 'done' | 'cancelled'
  costUsd: number
  startedAt?: string
  completedAt?: string
  updatedAt: string
}

export interface FileArtifact {
  filename: string
  filePath: string
  content: string
  language: string
  commitSha: string
  linesOfCode: number
  phase: string
  validationPassed: boolean
  validationError: string
  operation: 'create' | 'modify'
}

export interface BlackboardEvent {
  id: string
  workflowId: string
  sequenceNumber: number
  eventType: string
  postedByExpertId?: string
  postedByClient: boolean
  toExpertId?: string
  content: Record<string, unknown>
  revision: number
  postedAt: string
}

// ============================================================
// API functions
// ============================================================

export const listWorkflows = () =>
  baseAPI
    .get<ApiResponse<Workflow[]>>('/api/v1/workflows')
    .then((res) => res.data.data ?? [])

export const createWorkflow = (req: {
  projectId: string
  title: string
  selectedExpertIds: string[]
  costBudgetUsd?: number
  requirementText?: string
  /** Omitted or 'scratch' keeps the original behaviour. */
  mode?: 'scratch' | 'existing_codebase'
}) =>
  baseAPI
    .post<ApiResponse<Workflow>>('/api/v1/workflows', req)
    .then((res) => res.data.data!)

export const getWorkflow = (workflowId: string) =>
  baseAPI
    .get<ApiResponse<Workflow>>(`/api/v1/workflows/${workflowId}`)
    .then((res) => res.data.data!)

export const startWorkflow = (workflowId: string) =>
  baseAPI
    .post<ApiResponse<Workflow>>(`/api/v1/workflows/${workflowId}/start`)
    .then((res) => res.data.data!)

export const runWorkflow = (workflowId: string) =>
  baseAPI
    .post<ApiResponse<{ status: string; workflowId: string; message: string }>>(
      `/api/v1/workflows/${workflowId}/run`
    )
    .then((res) => res.data.data!)

export const getKanban = (workflowId: string) =>
  baseAPI
    .get<ApiResponse<{ workflowId: string; tasks: KanbanTask[] }>>(
      `/api/v1/workflows/${workflowId}/kanban`
    )
    .then((res) => res.data.data!)

export const getBlackboard = (workflowId: string, since = 0, types?: string[]) => {
  const params = new URLSearchParams({ since: String(since) })
  if (types?.length) params.set('types', types.join(','))
  return baseAPI
    .get<ApiResponse<{ workflowId: string; events: BlackboardEvent[]; count: number }>>(
      `/api/v1/workflows/${workflowId}/blackboard?${params}`
    )
    .then((res) => res.data.data!)
}

export const respondToApproval = (
  workflowId: string,
  approvalId: string,
  decision: 'approve' | 'approve_with_notes' | 'request_changes' | 'reject_and_restart_phase' | 'cancel_workflow',
  notes?: string,
  genericAllowancePct?: number
) =>
  baseAPI
    .post<ApiResponse<{ status: string; decision: string }>>(
      `/api/v1/workflows/${workflowId}/approvals/${approvalId}/respond`,
      { decision, notes, genericAllowancePct }
    )
    .then((res) => res.data.data!)

export const cancelWorkflow = (workflowId: string) =>
  baseAPI
    .post<ApiResponse<{ status: string }>>(`/api/v1/workflows/${workflowId}/cancel`)
    .then((res) => res.data.data!)

export const retryTask = (workflowId: string, taskId: string) =>
  baseAPI
    .post<ApiResponse<{ status: string }>>(`/api/v1/workflows/${workflowId}/tasks/${taskId}/retry`)
    .then((res) => res.data.data!)

/**
 * Phase 3D — the human-approved working set for an existing-codebase workflow.
 *
 * The backend stores suggestions and approvals in ONE table with three states,
 * so "the manifest" is simply the approved rows rather than a second list that
 * could drift out of step. Nothing here lets an expert's suggestion become
 * readable on its own: only a decision does that.
 */
export type CodebaseFileStatus = 'pending' | 'approved' | 'rejected'

export interface CodebaseFile {
  id: string
  workflowId: string
  path: string
  status: CodebaseFileStatus
  /** 'expert' = the system proposed it; 'client' = the client added it directly. */
  source: 'expert' | 'client'
  reason: string
  score: number
  hopDepth: number
  decidedAt: string | null
  createdAt: string
}

export const getCodebaseFiles = (workflowId: string) =>
  baseAPI
    .get<ApiResponse<{ files: CodebaseFile[]; count: number; approved: number }>>(
      `/api/v1/workflows/${workflowId}/codebase/files`
    )
    .then((res) => res.data.data!)

/**
 * Asks the repository ranker for candidate files. Existing rows are left alone
 * whatever their status, so a file the client already rejected cannot be
 * silently re-proposed by a later run.
 */
export const suggestCodebaseFiles = (workflowId: string, limit?: number) =>
  baseAPI
    .post<ApiResponse<{ suggestions: CodebaseFile[]; count: number }>>(
      `/api/v1/workflows/${workflowId}/codebase/suggest`,
      { limit }
    )
    .then((res) => res.data.data!)

export const addCodebaseFile = (workflowId: string, path: string) =>
  baseAPI
    .post<ApiResponse<CodebaseFile>>(`/api/v1/workflows/${workflowId}/codebase/files`, { path })
    .then((res) => res.data.data!)

export const decideCodebaseFile = (
  workflowId: string,
  fileId: string,
  decision: 'approve' | 'reject'
) =>
  baseAPI
    .post<ApiResponse<CodebaseFile>>(
      `/api/v1/workflows/${workflowId}/codebase/files/${fileId}/decide`,
      { decision }
    )
    .then((res) => res.data.data!)

export const decideCodebaseFilesBulk = (
  workflowId: string,
  fileIds: string[],
  decision: 'approve' | 'reject'
) =>
  baseAPI
    .post<ApiResponse<{ changed: number }>>(
      `/api/v1/workflows/${workflowId}/codebase/files/decide-bulk`,
      { fileIds, decision }
    )
    .then((res) => res.data.data!)

/** The approved paths only — the set an expert is allowed to read. */
export const getCodebaseManifest = (workflowId: string) =>
  baseAPI
    .get<ApiResponse<{ paths: string[]; count: number }>>(
      `/api/v1/workflows/${workflowId}/codebase/manifest`
    )
    .then((res) => res.data.data!)
