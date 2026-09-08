import { create } from 'zustand'

/**
 * Reply / thread-context UI state (CT-D3, CATEGORY_TEMPLATE_HANDOFF.md
 * §5/§7). Deliberately a NEW, ISOLATED store — CT-L10 requires the
 * existing chat/message list, virtualization, and streamStore to stay
 * completely untouched by this feature. This file has zero imports
 * from streamStore.ts and is never imported BY it either.
 *
 * Scope: purely client-side UI state for "what am I replying to right
 * now, before I hit send". Once a message is actually sent with a
 * replyTo id (message/handler.go's Send, CT-C1), this store's job is
 * done for that turn — ChatPage.tsx clears it via clearReply() in the
 * same place it already clears pendingUserText after a successful send.
 *
 * WHY per-chat (keyed by chatId) rather than one flat global object:
 * a user could plausibly have two chat tabs/windows open (or navigate
 * away and back while a reply draft is pending) — keying by chatId
 * avoids one chat's reply target leaking into another's MessageInput,
 * the same reasoning streamStore.ts's activeStreams Map already
 * documents for its own per-chat keying.
 */

export interface ReplyTarget {
  /** The pinned message's id — becomes reply_to_message_id on send. */
  messageId: string
  /** Short preview text shown in the "Replying to: ..." chip. */
  preview: string
  /** Which expert originally said this (for the multi-select default). */
  expertId?: string
  expertName?: string
}

export interface ReplyState {
  target: ReplyTarget | null
  /** CT-L6: explicit opt-in only, never defaults to true. */
  includeFullThread: boolean
  /**
   * CT-L7: explicit opt-in expert loop-in on a reply. Empty set = only
   * the originally-replied-to expert receives this reply (the natural
   * default, per CATEGORY_TEMPLATE_HANDOFF.md §5 — "default = only the
   * originally-replied-to expert"). Selecting additional expert ids
   * here means the client explicitly chose to loop them in.
   */
  loopedInExpertIds: Set<string>
}

interface ReplyStore {
  replies: Map<string, ReplyState>
  /** Sets the reply target for a chat. Resets includeFullThread/loopedInExpertIds
   *  to their defaults — starting a NEW reply should not inherit toggle
   *  state from whatever the client was replying to previously. */
  setReplyTarget: (chatId: string, target: ReplyTarget) => void
  setIncludeFullThread: (chatId: string, value: boolean) => void
  toggleLoopedInExpert: (chatId: string, expertId: string) => void
  /** Clears the reply target entirely — "cancel reply" or after a successful send. */
  clearReply: (chatId: string) => void
}

function defaultReplyState(target: ReplyTarget): ReplyState {
  return { target, includeFullThread: false, loopedInExpertIds: new Set() }
}

// WHY new Map/Set per mutation (never mutated in place): same
// rationale as streamStore.ts's header comment — Zustand's default
// selector equality is Object.is, so in-place mutation of a Map/Set
// value would not reliably trigger a re-render for components reading
// via `useReplyStore(s => s.replies.get(chatId))`.
export const useReplyStore = create<ReplyStore>((set, get) => ({
  replies: new Map(),

  setReplyTarget: (chatId, target) => {
    const next = new Map(get().replies)
    next.set(chatId, defaultReplyState(target))
    set({ replies: next })
  },

  setIncludeFullThread: (chatId, value) => {
    const current = get().replies.get(chatId)
    if (!current) return
    const next = new Map(get().replies)
    next.set(chatId, { ...current, includeFullThread: value })
    set({ replies: next })
  },

  toggleLoopedInExpert: (chatId, expertId) => {
    const current = get().replies.get(chatId)
    if (!current) return
    const nextSet = new Set(current.loopedInExpertIds)
    if (nextSet.has(expertId)) {
      nextSet.delete(expertId)
    } else {
      nextSet.add(expertId)
    }
    const next = new Map(get().replies)
    next.set(chatId, { ...current, loopedInExpertIds: nextSet })
    set({ replies: next })
  },

  clearReply: (chatId) => {
    const next = new Map(get().replies)
    next.delete(chatId)
    set({ replies: next })
  },
}))
