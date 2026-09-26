import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'
import { useAuthStore } from '@/stores/authStore'

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
  /** True when this workflow was asked to produce working code as well. */
  deliverCode?: boolean
  // failureReason: why the workflow stopped, set only when status === 'failed'.
  // The engine has always recorded it; the read API did not return it, so the
  // screen could show a FAILED badge and nothing else.
  failureReason?: string | null
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
  /**
   * deliverCode asks the implementation phase for working code as well as the
   * design documents. Omitted/false is the design-only path (§9), which is what
   * every workflow did before this option existed.
   */
  deliverCode?: boolean
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

/**
 * Phase 3F — delivery as a patch against the pinned base revision.
 *
 * The client's repository is never written to. The patch is generated against
 * the exact commit the work was based on, so it can be reviewed and applied (or
 * rejected) by the client themselves.
 */
export interface CodebaseChangedFile {
  /** git name-status code: A, M, D, or R with a similarity score. */
  status: string
  path: string
}

export interface CodebaseDelivery {
  workflowId: string
  baseCommitSha: string
  baselineRef: string
  changedFiles: CodebaseChangedFile[]
  patchBytes: number
  generatedAt: string | null
}

export const getCodebasePatch = (workflowId: string) =>
  baseAPI
    .get<ApiResponse<CodebaseDelivery>>(`/api/v1/workflows/${workflowId}/codebase/patch`)
    .then((res) => res.data.data!)

export const generateCodebasePatch = (workflowId: string) =>
  baseAPI
    .post<ApiResponse<CodebaseDelivery>>(`/api/v1/workflows/${workflowId}/codebase/patch`)
    .then((res) => res.data.data!)

/**
 * Downloads the patch as a file.
 *
 * WHY fetch rather than baseAPI: this is a binary-ish attachment, and routing it
 * through the shared axios instance would push it through the camelCase response
 * transform. Auth still goes through the same access token.
 */
/**
 * getCodebasePatchText fetches the same patch the download button serves, as
 * text, so the review screen can show it. WHY a separate raw fetch: the download
 * path is an attachment meant to be saved, and reading it as a blob would show
 * the reviewer nothing.
 */
export async function getCodebasePatchText(workflowId: string): Promise<string> {
  const token = useAuthStore.getState().accessToken
  const base = import.meta.env.VITE_API_URL ?? ''
  const res = await fetch(`${base}/api/v1/workflows/${workflowId}/codebase/patch/download`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) {
    throw new Error(`patch ${res.status}`)
  }
  return res.text()
}

export async function downloadCodebasePatch(workflowId: string): Promise<Blob> {
  const token = useAuthStore.getState().accessToken
  const base = import.meta.env.VITE_API_URL ?? ''
  const res = await fetch(`${base}/api/v1/workflows/${workflowId}/codebase/patch/download`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) {
    throw new Error('patch download failed')
  }
  return res.blob()
}
