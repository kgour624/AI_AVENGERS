import type { StreamState } from '@/stores/streamStore'
import { ExpertAvatar } from '@/components/expert/ExpertAvatar'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 10 ("Streaming State")
 * wireframe. GAP (unchanged from before ARC-51, still true): the
 * wireframe shows a per-expert progress bar and "Gate 3/5" text, but
 * SSEEvent's 'thinking' payload (section 8) only carries
 * `{ message: string, experts: number }` - a count, with NO per-expert
 * gate number anywhere in the documented event schema. ARC-51 §9
 * upgrades the VISUALS (real ExpertAvatar analyzing/responded states
 * instead of plain bullet dots) but does not fabricate the missing
 * data - pending experts still render a generic, identity-less
 * avatar (empty domain string -> ExpertAvatar's 'core' fallback
 * shape), never a fake name/gate number. Completed responses use
 * their REAL domain + mode from the SSE payload (both ARE present on
 * a completed ExpertResponse), so their avatar's color/shape is
 * genuinely earned, not decorative guessing.
 */
export function StreamingIndicator({ stream, expertCount }: { stream: StreamState; expertCount: number }) {
  const respondedCount = stream.expertResponses.length
  const pendingCount = Math.max(0, expertCount - respondedCount)

  return (
    <div className="rounded-lg border border-glass-border bg-surface-raised/70 p-4 backdrop-blur-xl">
      <p className="mb-3 text-sm text-text-secondary">{'\u26a1'} Neural Processing{'\u2026'}</p>
      <div className="flex flex-wrap gap-4">
        {stream.expertResponses.map((response, i) => (
          <div key={response.expertId ?? i} className="flex flex-col items-center gap-1">
            <ExpertAvatar domain={response.domain ?? ''} status="responded" mode={response.mode} size="sm" />
            <span className="max-w-[72px] truncate text-xs text-text-secondary">
              {response.expertName ?? 'Expert'}
            </span>
          </div>
        ))}
        {Array.from({ length: pendingCount }).map((_, i) => (
          <div key={`pending-${i}`} className="flex flex-col items-center gap-1">
            <ExpertAvatar domain="" status="analyzing" size="sm" />
            <span className="text-xs text-text-disabled">Thinking{'\u2026'}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
