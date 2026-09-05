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
  createdAt: string
}

/** Route loader return shape for /projects/:projectId/chats/:chatId */
export interface ChatLoaderData {
  chat: Chat
  messages: Message[]
  experts: ProjectExpert[]
}
