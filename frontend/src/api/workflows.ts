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
  // genericAllowancePct: 0-30. Sent with 'request_changes' to re-run the
  // design with a bounded generic-knowledge allowance. Omit to leave the
  // workflow's current setting untouched.
  genericAllowancePct?: number
) =>
  baseAPI
    .post<ApiResponse<{ status: string; decision: string }>>(
      `/api/v1/workflows/${workflowId}/approvals/${approvalId}/respond`,
      { decision, notes, genericAllowancePct }
    )
    .then((res) => res.data.data!)
