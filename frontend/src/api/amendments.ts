import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'

/**
 * §7.5's write-back path (docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md).
 * Deliberately NOT api/workflows.ts's respondToApproval — that endpoint
 * resumes the WORKFLOW's run state on approve (Engine.Resume), which an
 * amendment must never do (an amendment does not pause anything). See
 * backend-go/internal/workflow/amendment_handler.go's file comment for the
 * three reasons this needed its own pair of endpoints.
 *
 * Types mirror AmendmentView / AmendmentOutcome in
 * backend-go/internal/workflow/amendment.go exactly.
 */

export type AmendmentStatus = 'pending' | 'approved' | 'rejected'

export interface Amendment {
  approvalId: string
  status: AmendmentStatus
  summary: string
  kind: 'statement' | 'acceptance'
  target: string
  oldText: string
  newText: string
  reason: string
  proposedByExpert?: string
  chatId?: string
  requestedAt: string
  respondedAt?: string
}

export interface AmendmentOutcome {
  approvalId: string
  status: string
  applied: boolean
  target?: string
  commitSha?: string
  decisionId?: string
}

export const listAmendments = (
  workflowId: string,
  status: AmendmentStatus | 'all' = 'pending'
) =>
  baseAPI
    .get<ApiResponse<{ amendments: Amendment[] }>>(
      `/api/v1/workflows/${workflowId}/amendments`,
      { params: { status } }
    )
    .then((res) => res.data.data?.amendments ?? [])

export const respondToAmendment = (
  workflowId: string,
  approvalId: string,
  decision: 'approve' | 'approve_with_edit' | 'reject',
  opts?: { editedText?: string; notes?: string }
) =>
  baseAPI
    .post<ApiResponse<AmendmentOutcome>>(
      `/api/v1/workflows/${workflowId}/amendments/${approvalId}/respond`,
      { decision, editedText: opts?.editedText, notes: opts?.notes }
    )
    .then((res) => res.data.data!)
