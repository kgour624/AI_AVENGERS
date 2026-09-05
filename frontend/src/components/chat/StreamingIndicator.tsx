import type { StreamState } from '@/stores/streamStore'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 10 ("Streaming State")
 * wireframe:
 *   "\u23f3 Processing...
 *    \ud83d\udd35 SD Expert    \u2588\u2588\u2588\u2588\u2588\u2588\u2588\u2588\u2591\u2591  Gate 3/5
 *    \ud83d\udfe2 DB Expert    \u2588\u2588\u2588\u2588\u2588\u2588\u2588\u2588\u2588\u2588  Generating..."
 *
 * GAP: the wireframe shows a per-expert progress bar and "Gate 3/5"
 * text, but SSEEvent's 'thinking' payload (section 8) only carries
 * `{ message: string, experts: number }` - a single string + a count,
 * with NO per-expert gate number or progress percentage anywhere in
 * the documented event schema. There is no real data to drive a
 * progress bar or a gate counter per expert. Rather than fabricate
 * fake numbers that look precise but mean nothing (which would be
 * actively misleading - a user seeing "Gate 3/5" would reasonably
 * expect that number to be true), this renders a generic pulsing
 * "thinking" indicator per expert response received so far, plus a
 * simple indeterminate pulse for experts not yet responded. Logged as
 * a backend gap in HANDOFF.md: if per-expert gate progress is wanted,
 * the 'thinking' SSE event needs an expertId + gateNumber field added.
 */
export function StreamingIndicator({ stream, expertCount }: { stream: StreamState; expertCount: number }) {
  const respondedCount = stream.expertResponses.length
  const pendingCount = Math.max(0, expertCount - respondedCount)

  return (
    <div className="rounded-lg border border-surface-border bg-surface-raised p-4">
      <p className="mb-2 text-sm text-text-secondary">\u23f3 Processing...</p>
      <div className="space-y-1.5">
        {stream.expertResponses.map((response, i) => (
          <div key={response.expertId ?? i} className="flex items-center gap-2 text-xs text-text-secondary">
            <span className="thinking-indicator">\u25cf</span>
            {response.expertName ?? 'Expert'} \u2014 generated
          </div>
        ))}
        {Array.from({ length: pendingCount }).map((_, i) => (
          <div key={`pending-${i}`} className="flex items-center gap-2 text-xs text-text-disabled">
            <span className="thinking-indicator">\u25cf</span>
            Thinking...
          </div>
        ))}
      </div>
    </div>
  )
}
