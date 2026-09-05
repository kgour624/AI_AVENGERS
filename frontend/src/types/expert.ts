/**
 * Expert + response-mode domain models.
 * Source: FRONTEND_SYSTEM_DESIGN.md section 13, cross-referenced against
 * the `experts` table and DecisionResult/ExpertResponse Go structs in
 * AI_AVENGERS_SYSTEM_ARCHITECTURE.md sections 5, 10.
 */

export type ResponseMode = 'ASK' | 'WARN' | 'PUSH_BACK' | 'REFUSE' | 'ADVISE'

/** Which of the 5 gates stopped processing. 0 means it reached Gate 5 (generated). */
export type GateStopped = 0 | 1 | 2 | 3 | 4 | 5

export interface Expert {
  id: string
  name: string
  slug: string
  domain: string
  description: string
  avatarUrl?: string
  totalChunks: number
  totalTopics: number
  avgDepthLevel: number
  avgRating: number
  totalRatings: number
  isActive: boolean
  isTraining: boolean
  createdAt: string
}

export interface ExpertTopic {
  topic: string
  depthLevel: 1 | 2 | 3 | 4 | 5
  chunkCount: number
  complexityCeiling: 'basic' | 'intermediate' | 'advanced' | 'expert' | 'master'
  canHandle: string[]
  cannotHandle: string[]
  exampleQuestions: string[]
}

export interface Citation {
  chunkId: string
  text: string
  score: number
}

export interface ExpertResponse {
  expertId: string
  expertName: string
  domain: string
  mode: ResponseMode
  content: string
  citations: Citation[]
  confidence: number
  gateStopped: GateStopped
  warning?: string
  /** Only populated when mode === 'ASK' */
  questions?: string[]
  /** Populated if this expert's processing failed but others succeeded */
  error?: string
}

export interface Contradiction {
  topic: string
  expertA: string
  positionA: string
  expertB: string
  positionB: string
}

export interface SynthesisResult {
  agreements: string[]
  contradictions: Contradiction[]
  summary: string
}

/**
 * Exhaustive mode -> badge config mapping.
 * WHY the `never` branch: TypeScript Simplified transcript pattern -
 * if the backend ever adds a 6th ResponseMode, this switch fails to
 * compile until every call site is updated, instead of silently
 * falling through at runtime with a missing badge.
 */
export interface ModeBadgeConfig {
  color: string
  icon: string
  label: string
}

export function getModeBadgeConfig(mode: ResponseMode): ModeBadgeConfig {
  switch (mode) {
    case 'ADVISE':
      return { color: 'mode-advise', icon: '\u2705', label: 'ADVISE' }
    case 'ASK':
      return { color: 'mode-ask', icon: '\u2753', label: 'ASK' }
    case 'WARN':
      return { color: 'mode-warn', icon: '\u26A0\uFE0F', label: 'WARN' }
    case 'PUSH_BACK':
      return { color: 'mode-pushback', icon: '\uD83D\uDD04', label: 'PUSH BACK' }
    case 'REFUSE':
      return { color: 'mode-refuse', icon: '\uD83D\uDEAB', label: 'REFUSE' }
    default: {
      // Exhaustiveness check - compile error if a new mode is added
      // without updating this function.
      const _exhaustive: never = mode
      return _exhaustive
    }
  }
}
