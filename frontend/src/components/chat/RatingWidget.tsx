import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { rateMessage } from '@/api/messages'
import { cn } from '@/utils/cn'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 10: "\u2b50\u2b50\u2b50\u2b50\u2b50  [Rate this response]".
 *
 * WHY this only ever receives a messageId (never renders against a
 * live in-flight ExpertResponse): rating a response that hasn't been
 * persisted yet has no id to rate against. ExpertResponse (the SSE
 * payload type) deliberately has no messageId field - adding a fake
 * one just to make this component "work" during streaming would let a
 * user click a star that silently does nothing meaningful, or worse,
 * throws once the request actually fires. ChatPage only renders this
 * for messages loaded from getMessages() (i.e. already persisted),
 * never for the live stream - documented as a deliberate scope
 * decision in HANDOFF.md, not a missing feature.
 *
 * Cross-questioned: what if the user clicks a star while a previous
 * rating request is still in flight? useMutation's `isPending` gates
 * the buttons so a second click can't fire a second request before
 * the first resolves - avoids a race where two different scores could
 * both land depending on response order.
 */
export function RatingWidget({ messageId }: { messageId: string }) {
  const [selectedScore, setSelectedScore] = useState<number | null>(null)

  const mutation = useMutation({
    mutationFn: (score: 1 | 2 | 3 | 4 | 5) => rateMessage(messageId, { score }),
    onSuccess: (_data, score) => setSelectedScore(score),
  })

  return (
    <div className="flex items-center gap-1">
      {([1, 2, 3, 4, 5] as const).map((score) => (
        <button
          key={score}
          type="button"
          disabled={mutation.isPending || selectedScore !== null}
          onClick={() => mutation.mutate(score)}
          aria-label={`Rate ${score} stars`}
          className={cn(
            'text-sm disabled:cursor-not-allowed',
            selectedScore !== null && score <= selectedScore ? 'opacity-100' : 'opacity-40 hover:opacity-70'
          )}
        >
          \u2b50
        </button>
      ))}
      {selectedScore !== null && <span className="ml-1 text-xs text-text-secondary">Thanks!</span>}
      {mutation.isError && <span className="ml-1 text-xs text-mode-refuse">Failed to submit rating</span>}
    </div>
  )
}
