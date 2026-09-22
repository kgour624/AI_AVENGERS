import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'

/**
 * Workflow chat (§6 of docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md).
 * Deliberately separate from api/chats.ts, same as the backend keeps
 * workflow.WorkflowChatService entirely separate from the product chat
 * (chat.Service) — different tables, different routes, no shared code path.
 *
 * Types mirror the Go structs in backend-go/internal/workflow/chat_service.go
 * exactly (WorkflowChat, WorkflowChatMessage, ChatParticipant). Field names
 * below are camelCase because baseAPI's response interceptor deep-camelizes
 * every JSON body before it reaches this file (see types/api.ts's casing
 * note) — the wire shape is snake_case, this is what arrives.
 */

export interface WorkflowChat {
  id: string
  workflowId: string
  clientId: string
  pinnedEventId?: string
  title: string
  // 0-30. Present only when this chat overrides the workflow's own setting.
  genericAllowancePct?: number
  messageCount: number
  isArchived: boolean
  createdAt: string
  updatedAt: string
}

export interface WorkflowChatMessage {
  id: string
  chatId: string
  role: 'user' | 'assistant'
  expertId?: string
  expertName?: string
  content: string
  turnNumber: number
  // Present only on an intermediate tool-call step (§7.3), never on the
  // final answer.
  toolName?: string
  toolInput?: Record<string, unknown>
  toolResult?: Record<string, unknown>
  stepNumber?: number
  citations?: unknown
  gateResult?: {
    trainingChunks?: number
    peerContributions?: number
    gate1Passed?: boolean
    genericAllowed?: boolean
    genericBlocked?: boolean
    genericAllowancePct?: number
    coverageGap?: string
  }
  tokensUsed: number
  costUsd: number
  createdAt: string
}

export interface ChatParticipant {
  expertId: string
  expertName: string
  domain: string
  addedAt: string
}

// ============================================================
// Chats — scoped to a workflow (POST/GET /workflows/:id/chats)
// ============================================================

export const createWorkflowChat = (
  workflowId: string,
  req: { title: string; pinnedEventId?: string }
) =>
  baseAPI
    .post<ApiResponse<WorkflowChat>>(`/api/v1/workflows/${workflowId}/chats`, req)
    .then((res) => res.data.data!)

export const listWorkflowChats = (workflowId: string) =>
  baseAPI
    .get<ApiResponse<{ chats: WorkflowChat[] }>>(`/api/v1/workflows/${workflowId}/chats`)
    .then((res) => res.data.data?.chats ?? [])

// ============================================================
// One chat — scoped by chat id (GET/PATCH/POST /workflow-chats/:cid/*)
// ============================================================

export const getWorkflowChat = (chatId: string) =>
  baseAPI
    .get<ApiResponse<{ chat: WorkflowChat; participants: ChatParticipant[] }>>(
      `/api/v1/workflow-chats/${chatId}`
    )
    .then((res) => res.data.data!)

// genericAllowancePct: null clears the chat's override so the workflow's
// value applies again. undefined is not sent at all — always pass null
// explicitly to clear.
export const setKnowledgeMode = (chatId: string, genericAllowancePct: number | null) =>
  baseAPI
    .patch<ApiResponse<WorkflowChat>>(`/api/v1/workflow-chats/${chatId}/knowledge-mode`, {
      genericAllowancePct,
    })
    .then((res) => res.data.data!)

export const addChatParticipant = (chatId: string, expertId: string) =>
  baseAPI
    .post<ApiResponse<{ participants: ChatParticipant[] }>>(
      `/api/v1/workflow-chats/${chatId}/participants`,
      { expertId }
    )
    .then((res) => res.data.data!.participants)

export const removeChatParticipant = (chatId: string, expertId: string) =>
  baseAPI
    .delete<ApiResponse<{ removed: string }>>(
      `/api/v1/workflow-chats/${chatId}/participants/${expertId}`
    )
    .then((res) => res.data.data!)

// expertId addresses one participant. Omit for the default responder — the
// pinned deliverable's author (§6.3).
export const sendWorkflowChatMessage = (chatId: string, message: string, expertId?: string) =>
  baseAPI
    .post<ApiResponse<WorkflowChatMessage>>(`/api/v1/workflow-chats/${chatId}/messages`, {
      message,
      expertId,
    })
    .then((res) => res.data.data!)

export const listWorkflowChatMessages = (chatId: string, limit?: number) =>
  baseAPI
    .get<ApiResponse<{ messages: WorkflowChatMessage[] }>>(
      `/api/v1/workflow-chats/${chatId}/messages`,
      { params: limit ? { limit } : undefined }
    )
    .then((res) => res.data.data?.messages ?? [])

// ============================================================
// Change requests (chat-driven coordinated redesign)
// ============================================================

export interface ChangeRequest {
  id: string
  workflowId: string
  clientId: string
  sourceChatId?: string
  sourceMessageId?: string
  changeGoal: string
  relevantExpertIds: string[]
  // pending | running | completed | cancelled
  status: string
  blackboardEventId?: string
  requestedAt: string
  startedAt?: string
  completedAt?: string
}

// proposeChange sends a free-text change goal from the workflow chat.
// The runner picks it up, determines relevant experts, re-runs their design
// sections, and presents the same approval gate as the initial design.
export const proposeChange = (chatId: string, changeGoal: string) =>
  baseAPI
    .post<ApiResponse<ChangeRequest>>(
      `/api/v1/workflow-chats/${chatId}/propose-change`,
      { changeGoal }
    )
    .then((res) => res.data.data!)
