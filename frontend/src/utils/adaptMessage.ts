import type { Message, ProjectExpert } from '@/types/project'
import type { ExpertResponse } from '@/types/expert'

/**
 * Converts a persisted Message row (assistant role) into the
 * ExpertResponse shape ExpertResponse.tsx renders.
 *
 * WHY this adapter exists rather than making ExpertResponse.tsx accept
 * either type directly: ExpertResponse (the SSE payload type) and
 * Message (the persisted DB row type) genuinely have different shapes
 * with different gaps (see types/project.ts's documented data-loss
 * gap for warning/questions) - collapsing them into one type would
 * either make ExpertResponse's fields all-optional everywhere
 * (weakening type safety for the live-stream case, which DOES have
 * them) or silently paper over the fact that reloaded messages
 * genuinely have less information than live ones. An explicit adapter
 * makes that information loss visible at the one call site that needs
 * to know about it, instead of hidden inside a shared type.
 *
 * Returns null for user-role messages or a message with no expertId,
 * since ExpertResponse has no representation of "who said this" for a
 * user turn - callers should render those as plain chat bubbles
 * instead (see ChatPage.tsx).
 */
export function persistedMessageToExpertResponse(
  message: Message,
  experts: ProjectExpert[]
): ExpertResponse | null {
  if (message.role !== 'assistant' || !message.expertId) return null

  const expert = experts.find((e) => e.expertId === message.expertId)

  return {
    expertId: message.expertId,
    expertName: expert?.expertName ?? 'Unknown expert',
    domain: expert?.domain ?? '',
    // WHY default to 'ADVISE' rather than leaving mode undefined: the
    // backend always sets decision_mode for a real assistant row (it's
    // NOT NULL in the schema per the DecisionResult flow) - a missing
    // value here would indicate a genuinely unexpected data state, not
    // a normal case to silently paper over. Defaulting to ADVISE is a
    // pragmatic fallback so the UI doesn't crash on that unexpected
    // state, but this default should never actually be hit in practice.
    mode: message.decisionMode ?? 'ADVISE',
    content: message.content,
    citations: message.citations ?? [],
    confidence: message.confidence ?? 0,
    gateStopped: (message.gateStopped ?? 0) as 0 | 1 | 2 | 3 | 4 | 5,
    // warning and questions are NOT reconstructable from a persisted
    // Message - see the documented gap in types/project.ts. Left
    // undefined here rather than guessed at.
  }
}
