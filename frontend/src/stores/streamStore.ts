import { create } from 'zustand'
import type { ExpertResponse, SynthesisResult } from '@/types/expert'

/**
 * Active SSE stream state, keyed by chatId.
 * Source: FRONTEND_SYSTEM_DESIGN.md section 6.
 *
 * Gap fixed vs the design doc: the doc's StreamStore interface listing
 * (section 6) does NOT declare a `setError` method, but its own
 * useSSEStream implementation (section 8) calls
 * `useStreamStore.getState().setError(chatId, String(error))` in the
 * catch block. Without adding setError here, that call would throw
 * "setError is not a function" at runtime the first time a fetch()
 * fails (e.g. network drop mid-stream) - the exact scenario a stream
 * store exists to handle gracefully. Added below.
 *
 * WHY store state is always a NEW object per update (never mutated
 * in place): components read via a selector like
 * `useStreamStore(s => s.activeStreams.get(chatId))`. Zustand's
 * default equality check is Object.is on the selector's return value.
 * If we mutated the StreamState object in place and only replaced the
 * Map, `.get(chatId)` would still return a referentially-identical
 * object on the next read in some call orders, and React would skip
 * re-rendering. Every mutation method below therefore constructs a
 * brand-new Map AND a brand-new StreamState for the touched key.
 */

export type StreamStatus = 'thinking' | 'streaming' | 'complete' | 'error'

export interface StreamState {
  status: StreamStatus
  expertResponses: Partial<ExpertResponse>[]
  synthesis: SynthesisResult | null
  error: string | null
}

function emptyStreamState(): StreamState {
  return { status: 'thinking', expertResponses: [], synthesis: null, error: null }
}

interface StreamStore {
  activeStreams: Map<string, StreamState>
  startStream: (chatId: string) => void
  appendChunk: (chatId: string, patch: Partial<StreamState>) => void
  completeStream: (chatId: string) => void
  setError: (chatId: string, error: string) => void
  clearStream: (chatId: string) => void
}

export const useStreamStore = create<StreamStore>((set, get) => ({
  activeStreams: new Map(),

  startStream: (chatId) => {
    const next = new Map(get().activeStreams)
    next.set(chatId, emptyStreamState())
    set({ activeStreams: next })
  },

  appendChunk: (chatId, patch) => {
    const current = get().activeStreams.get(chatId)
    // Guard: an SSE event can arrive after clearStream() ran (e.g. the
    // component unmounted and cleaned up, but a queued microtask from
    // the reader loop still fires). Silently ignore rather than
    // resurrecting a stream nobody is listening to.
    if (!current) return

    const next = new Map(get().activeStreams)
    next.set(chatId, { ...current, ...patch })
    set({ activeStreams: next })
  },

  completeStream: (chatId) => {
    const current = get().activeStreams.get(chatId)
    if (!current) return

    const next = new Map(get().activeStreams)
    next.set(chatId, { ...current, status: 'complete' })
    set({ activeStreams: next })
  },

  setError: (chatId, error) => {
    const current = get().activeStreams.get(chatId) ?? emptyStreamState()
    const next = new Map(get().activeStreams)
    next.set(chatId, { ...current, status: 'error', error })
    set({ activeStreams: next })
  },

  clearStream: (chatId) => {
    const next = new Map(get().activeStreams)
    next.delete(chatId)
    set({ activeStreams: next })
  },
}))
