/**
 * Project, chat, message domain models.
 * Source: FRONTEND_SYSTEM_DESIGN.md section 13, cross-referenced against
 * the `projects`/`chats`/`messages` tables in
 * AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 5.
 */

import type { Citation, ResponseMode } from './expert'

export type ProjectStatus = 'active' | 'completed' | 'archived'
export type RepoProvider = 'github' | 'gitlab'

export interface ProjectExpert {
  expertId: string
  expertName: string
  domain: string
  addedAt: string
  isActive: boolean
}

export interface Project {
  id: string
  clientId: string
  name: string
  description: string
  status: ProjectStatus
  repoUrl?: string
  repoProvider?: RepoProvider
  repoConnected: boolean
  repoLastSync?: string
  experts: ProjectExpert[]
  createdAt: string
  updatedAt: string
}

export interface Chat {
  id: string
  projectId: string
  title: string
  messageCount: number
  isArchived: boolean
  createdAt: string
  updatedAt: string
}

export type MessageRole = 'user' | 'assistant'

export interface Message {
  id: string
  chatId: string
  role: MessageRole
  content: string
  turnNumber: number
  expertId?: string
  decisionMode?: ResponseMode
  citations?: Citation[]
  confidence?: number
  /**
   * WHICH of the 5 gates stopped processing (0 = reached Gate 5).
   * Added in Phase 3 - this was missing from the Phase 1 version of
   * this interface despite AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 5's
   * `messages` table having a real `gate_stopped INTEGER` column.
   * Without this field, ExpertResponse.tsx's "Why did I stop here?"
   * explainer would have no data to work with once rendering a
   * persisted (reloaded) message rather than a live SSE response.
   */
  gateStopped?: number
  createdAt: string
}

/**
 * KNOWN DATA-LOSS GAP (documented, not silently worked around):
 * `ExpertResponse.warning` (Gate 3 WARN reasoning) and
 * `ExpertResponse.questions` (Gate 1 ASK clarifying questions) have NO
 * corresponding column anywhere in the `messages` table schema
 * (AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 5). That means once a
 * turn completes and the page is reloaded, a WARN message's warning
 * text and an ASK message's question list are permanently gone -
 * only the persisted `content`/`decisionMode`/`citations` survive.
 * This is a real backend gap, not a frontend rendering choice: adding
 * `warning_text` and `clarifying_questions` (or similar) columns to
 * `messages` is the actual fix, tracked in HANDOFF.md rather than
 * faked here with empty defaults that would look correct but aren't.
 */

/** Route loader return shape for /projects/:projectId/chats/:chatId */
export interface ChatLoaderData {
  chat: Chat
  messages: Message[]
  experts: ProjectExpert[]
}
