/**
 * Memory system types (L2 group memory, L3 event log).
 * Source: AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 5
 * (project_memory_l2, master_event_log tables) and section 6
 * (Memory Manager). These power the "Recent Activity (Timeline)"
 * panel on the Project page (FRONTEND_SYSTEM_DESIGN.md section 10).
 */

export type L2MemoryType = 'decision' | 'code' | 'error' | 'fix' | 'recommendation'

export interface L2Entry {
  id: string
  projectId: string
  expertId: string
  expertName: string
  memoryType: L2MemoryType
  content: string
  context?: string
  turnReference?: number
  importance: 1 | 2 | 3 | 4 | 5
  isSuperseded: boolean
  createdAt: string
}

export interface L3Event {
  id: number
  projectId: string
  expertId?: string
  expertName?: string
  chatId?: string
  eventType: string
  eventData: Record<string, unknown>
  /** WHY this event happened - Arpit's principle, present on every L3 row */
  reasoning?: string
  decisionMade?: string
  alternativesConsidered?: string[]
  createdAt: string
}
