/**
 * Project, chat, message domain models.
 * Source: FRONTEND_SYSTEM_DESIGN.md section 13, cross-referenced against
 * the `projects`/`chats`/`messages` tables in
 * AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 5.
 */

import type { Citation, ResponseMode, TemplateSectionResult } from './expert'

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
  techStack?: string[]
  architectureType?: string
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
  /**
   * Gate 3 WARN reasoning text. Bug 4.1 fix (docs bug list): backend
   * migration 004 added `warning_text`/`clarifying_questions` columns
   * to `messages` and chat/service.go + message/handler.go now save
   * and select them (see those files' "Bug 3 fix" comments) - but
   * this interface and utils/adaptMessage.ts were never updated to
   * read them back, so a WARN/ASK message's extra data still
   * vanished on reload even though the backend had it all along.
   * Optional because most messages (ADVISE/REFUSE/PUSH_BACK) never
   * set these.
   */
  warningText?: string
  /** Gate 1 ASK clarifying questions. See warningText comment above. */
  clarifyingQuestions?: string[]
  /**
   * Fix (2026-09-08 RCA round 7, migration 011): a categorized
   * expert's structured answer (Pattern/Idea/Code/Walkthrough/Test
   * Cases) is now persisted to messages.template_sections and
   * returned here on reload/history fetch - previously undefined on
   * every persisted message, which meant utils/adaptMessage.ts could
   * only ever produce content (empty for a structured response) on
   * reload, rendering as completely blank despite the live SSE
   * stream having shown it correctly moments earlier.
   */
  templateSections?: TemplateSectionResult[]
  createdAt: string
}

/** Route loader return shape for /projects/:projectId/chats/:chatId */
export interface ChatLoaderData {
  chat: Chat
  messages: Message[]
  experts: ProjectExpert[]
}
