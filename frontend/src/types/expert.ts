/**
 * Expert + response-mode domain models.
 * Source: FRONTEND_SYSTEM_DESIGN.md section 13, cross-referenced against
 * the `experts` table and DecisionResult/ExpertResponse Go structs in
 * AI_AVENGERS_SYSTEM_ARCHITECTURE.md sections 5, 10.
 */

export type ResponseMode = 'ASK' | 'WARN' | 'PUSH_BACK' | 'REFUSE' | 'ADVISE'

/** Which of the 5 gates stopped processing. 0 means it reached Gate 5 (generated). */
export type GateStopped = 0 | 1 | 2 | 3 | 4 | 5

/**
 * PHASE 4 CORRECTION: verified against the real
 * backend-go/internal/expert/handler.go ExpertPublic struct (the
 * public GET /experts and GET /experts/:id endpoint, used by
 * api/experts.ts::getExperts/getExpert). It does NOT return
 * isActive, isTraining, or totalRatings at all - only the admin-only
 * endpoint (GetIn ListExperts, admin_handler.go) returns those.
 * Phase 1's Expert type assumed every one of these fields exists on
 * every expert response - they would have silently been `undefined`
 * at runtime on the public endpoints despite compiling cleanly.
 *
 * Split into a base `PublicExpert` (matches ExpertPublic exactly) and
 * `Expert` which extends it with the admin-only fields as optional -
 * so admin/experts.ts's getAdminExperts can still populate them, but
 * api/experts.ts's getExperts/getExpert (public) don't fabricate
 * fields the backend never sends.
 */
export interface PublicExpert {
  id: string
  name: string
  slug: string
  domain: string
  description: string
  totalChunks: number
  totalTopics: number
  avgDepthLevel: number
  avgRating: number
  createdAt: string
}

export interface Expert extends PublicExpert {
  avatarUrl?: string
  totalRatings?: number
  isActive?: boolean
  isTraining?: boolean
}

/**
 * PHASE 4 CORRECTION: verified against expert/handler.go's real
 * GetTopics query - `SELECT topic, depth_level, chunk_count,
 * complexity_ceiling FROM expert_capabilities`. There is NO
 * can_handle/cannot_handle/example_questions in this SELECT despite
 * those columns existing in the expert_capabilities table (per
 * AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 5) - the handler simply
 * never selects them. CapabilityCard.tsx (Phase 2) was built assuming
 * these arrays would always be populated; they will always be empty
 * from the real public endpoint today. Made them explicitly optional
 * with a documented reason, rather than required fields that are
 * always empty in practice (which would be misleading about what the
 * type guarantees).
 */
export interface ExpertTopic {
  topic: string
  depthLevel: 1 | 2 | 3 | 4 | 5
  chunkCount: number
  complexityCeiling: 'basic' | 'intermediate' | 'advanced' | 'expert' | 'master'
  /** NOT returned by the current GET /experts/:id/topics handler - see comment above. Backend gap, not a frontend rendering choice. */
  canHandle?: string[]
  cannotHandle?: string[]
  exampleQuestions?: string[]
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
  /**
   * Tailwind background+border utility classes for the badge chip.
   * WHY added here in Phase 2 (was missing in the Phase 1 version of
   * this function): FRONTEND_SYSTEM_DESIGN.md section 9's actual
   * MODE_CONFIG for ExpertResponse.tsx has 4 fields per mode - color,
   * icon, label, AND bgClass (e.g. 'bg-mode-advise/10 border-mode-advise/30').
   * Building Badge.tsx against the Phase 1 3-field version would have
   * meant either duplicating this switch a second time inside the
   * component (drifting from this single source of truth the moment
   * someone edits one but not the other) or silently dropping the
   * background styling the doc's wireframes clearly show. Extended
   * the existing exhaustive switch instead of creating a second one.
   */
  bgClass: string
}

export function getModeBadgeConfig(mode: ResponseMode): ModeBadgeConfig {
  switch (mode) {
    case 'ADVISE':
      return {
        color: 'text-mode-advise',
        icon: '\u2705',
        label: 'ADVISE',
        bgClass: 'bg-mode-advise/10 border-mode-advise/30',
      }
    case 'ASK':
      return {
        color: 'text-mode-ask',
        icon: '\u2753',
        label: 'ASK',
        bgClass: 'bg-mode-ask/10 border-mode-ask/30',
      }
    case 'WARN':
      return {
        color: 'text-mode-warn',
        icon: '\u26A0\uFE0F',
        label: 'WARN',
        bgClass: 'bg-mode-warn/10 border-mode-warn/30',
      }
    case 'PUSH_BACK':
      return {
        color: 'text-mode-pushback',
        icon: '\uD83D\uDD04',
        label: 'PUSH BACK',
        bgClass: 'bg-mode-pushback/10 border-mode-pushback/30',
      }
    case 'REFUSE':
      return {
        color: 'text-mode-refuse',
        icon: '\uD83D\uDEAB',
        label: 'REFUSE',
        bgClass: 'bg-mode-refuse/10 border-mode-refuse/30',
      }
    default: {
      // Exhaustiveness check - compile error if a new mode is added
      // without updating this function.
      const _exhaustive: never = mode
      return _exhaustive
    }
  }
}
