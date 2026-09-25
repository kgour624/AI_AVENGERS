/**
 * Expert + response-mode domain models.
 * Source: FRONTEND_SYSTEM_DESIGN.md section 13, cross-referenced against
 * the `experts` table and DecisionResult/ExpertResponse Go structs in
 * AI_AVENGERS_SYSTEM_ARCHITECTURE.md sections 5, 10.
 */

export type ResponseMode = 'ASK' | 'WARN' | 'PUSH_BACK' | 'REFUSE' | 'ADVISE'

/**
 * Which of the 5 gates stopped processing. 0 means it reached Gate 5
 * (generated).
 *
 * -1 (CT-C4, CATEGORY_TEMPLATE_HANDOFF.md §6): a sentinel emitted ONLY
 * by decision/engine.go's gateStructurePermission — means this
 * response IS the structure-permission ASK itself ("Chahiye
 * structure/boilerplate ya sirf logic likh doon?"), not a real Gate 1
 * clarification ASK. Added here as a real member of the union (not
 * silently widened to `number`) so every consumer must explicitly
 * handle it via TypeScript's exhaustiveness checking, matching this
 * file's existing `getModeBadgeConfig`'s `never`-branch pattern below.
 */
export type GateStopped = -1 | 0 | 1 | 2 | 3 | 4 | 5

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
  /**
   * Fix (2026-09-08): backend's admin ListExperts now selects
   * reasoning_charter (see admin_handler.go's adminExpertRow) — it was
   * saved correctly by the ingestion pipeline all along, just never
   * read back on this list endpoint. Optional because the PUBLIC
   * GET /experts endpoint still never returns this (admin-only field,
   * same reasoning as modelTier/temperature/etc. below).
   */
  reasoningCharter?: string
  // Migration 006 config fields — present on admin endpoint, absent on public.
  // WHY optional: public GET /experts never returns these. Admin GET /admin/experts does.
  // Keeping them optional on Expert means both endpoints can use the same type
  // without fabricating fields that don't exist on the public response.
  modelTier?: 'cheap' | 'strong' | 'fast'
  temperature?: number
  topP?: number
  loopPattern?: 'ota' | 'react' | 'plan_execute'
  maxLoopIterations?: number
  allowedTools?: string[]
  trainingStatus?: 'draft' | 'ingesting' | 'trained' | 'deprecated'
  /**
   * Fix (2026-09-08): backend's admin ListExperts now selects
   * category_id (see admin_handler.go's adminExpertRow) — previously
   * missing entirely, so the admin panel had no way to see which
   * category an expert is in, or fix a mismatch introduced by a
   * direct-SQL category retrofit (done once, manually, before this
   * admin category-edit UI existed). null = flat-text expert, no
   * category (CT-L2). Optional (not `| null` required) for the same
   * reason as every other admin-only field on this type: the public
   * GET /experts endpoint never returns it at all.
   */
  categoryId?: string | null
}

/**
 * Human-readable label for trainingStatus badge.
 * WHY a function not a map: avoids importing a map at module level
 * when most consumers only need one value at a time.
 */
export function trainingStatusLabel(status: Expert['trainingStatus']): string {
  switch (status) {
    case 'trained':    return '\u2705 Trained'
    case 'ingesting':  return '\u23f3 Ingesting'
    case 'deprecated': return '\u26a0\ufe0f Deprecated'
    case 'draft':
    default:           return '\u270f\ufe0f Draft'
  }
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
  /**
   * Measured answer depth from a capability evaluation pass (I2): 1 definitions,
   * 2 mechanics/trade-offs, 3 failure modes. 0 means never measured.
   *
   * A DIFFERENT SCALE from depthLevel and never comparable to it: depthLevel counts
   * chunk coverage (1-5), measuredLevel counts what the expert actually answered
   * (1-3). The UI shows them side by side, labelled, for exactly that reason.
   */
  measuredLevel?: 0 | 1 | 2 | 3
  /** How many questions were asked about this topic in the last pass. */
  evalCases?: number
  /** How many of those were answered, cited and judged supported. */
  evalPassed?: number
  /**
   * can_handle / cannot_handle are only meaningful once evalCases > 0.
   *
   * They used to be generated from a few hundred characters per topic; they are now
   * overwritten by an evaluation pass, so render them ONLY when measured. Showing
   * them for an unmeasured topic would publish the old guess as a claim.
   */
  canHandle?: string[]
  cannotHandle?: string[]
  exampleQuestions?: string[]
  /** When this topic's capability was last measured, if ever. */
  lastEvaluatedAt?: string | null
}

/**
 * Citation links a claim in an expert's answer to a source chunk from their training material.
 * Feature #23 (2026-09-23): Added sourceName and chunkIndex so the citation modal can show
 * which transcript the citation came from and its position in that transcript.
 */
export interface Citation {
  chunkId: string
  text: string
  score: number
  /**
   * Human-readable transcript filename (e.g. "react_hooks_part1.txt").
   * Empty string if chunk has no source_file (legacy data, repo chunks, etc.).
   * Frontend displays "Unknown Source" when empty.
   */
  sourceName?: string
  /**
   * 0-based position in the original transcript.
   * Frontend displays as 1-based: "Chunk #43" (index 42 + 1).
   * Helps users locate the exact position: "this is chunk #42 out of 150".
   */
  chunkIndex?: number
}

/**
 * One section of a categorized expert's structured JSON answer
 * (CT-B, CATEGORY_TEMPLATE_HANDOFF.md §4). Mirrors
 * chinawall.TemplateSectionResult exactly — `type` intentionally
 * reuses the same 3 values as CategoryTemplateSection['type']
 * (category.ts) since both describe the SAME section, just at
 * different points in the pipeline (category owns the schema
 * definition; this is the generated, citation-processed result).
 */
export interface TemplateSectionResult {
  key: string
  label: string
  type: 'prose' | 'code' | 'test_cases'
  content: string
  /**
   * 2026-09-08 RCA: typed nullable to match reality - backend fixed
   * to always send [] (see chinawall's extractCitations), but rows
   * persisted before that fix, and any future backend regression of
   * the same shape, can still send null. Every consumer
   * (splitContentByCitations, ExpertResponse.tsx) already guards for
   * this; the type now says so honestly instead of promising a
   * guarantee the wire data does not actually keep.
   */
  citations: Citation[] | null
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
  /**
   * CT-B4: populated ONLY when this expert has a category with a
   * non-empty template_schema. undefined/omitted for every flat-text
   * expert response (CT-L2) — consumers must check
   * `templateSections && templateSections.length > 0` before
   * rendering structured UI, falling back to plain `content`
   * otherwise (same pattern the backend uses at every layer).
   */
  templateSections?: TemplateSectionResult[]
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
