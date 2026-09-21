import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'

/**
 * §17 (code-feedback loop) and §18 (git-push export) —
 * docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md. Both act on a client's OWN git
 * remote using the token from api/repo.ts's connect flow — see
 * backend-go/internal/workflow/client_remote.go for why the remote is
 * validated against a host allowlist before that token is ever read.
 */

export interface GitExportRequest {
  provider: 'github' | 'gitlab'
  repoUrl: string
  branch?: string
  createRepo?: boolean
  // Only applies when createRepo is true. Omit to default to private —
  // this repo is a client's design and acceptance criteria.
  private?: boolean
}

export interface GitExportResult {
  repoUrl: string
  branch: string
  commitSha: string
  repoCreated: boolean
  eventId?: string
}

export const exportHarnessToGit = (workflowId: string, req: GitExportRequest) =>
  baseAPI
    .post<ApiResponse<GitExportResult>>(`/api/v1/workflows/${workflowId}/export/git`, req)
    .then((res) => res.data.data!)

export interface CodeFeedbackRequest {
  provider: 'github' | 'gitlab'
  repoUrl: string
  branch?: string
}

// Asynchronous: the comparison runs in the background. Results arrive as
// code_feedback_ingested / code_feedback_question blackboard events, read
// through getBlackboard (api/workflows.ts) — there is no separate ingest
// status endpoint.
export const ingestCodeFeedback = (workflowId: string, req: CodeFeedbackRequest) =>
  baseAPI
    .post<ApiResponse<{ ingestStarted: boolean; note: string }>>(
      `/api/v1/workflows/${workflowId}/code-feedback`,
      req
    )
    .then((res) => res.data.data!)
