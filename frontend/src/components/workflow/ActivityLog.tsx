import { useState } from 'react'
import { cn } from '@/utils/cn'
import { Card } from '@/components/ui/Card'
import type { BlackboardEvent } from '@/api/workflows'

/**
 * ActivityLog — chronological audit trail of all blackboard_events.
 *
 * WHY: blackboard_events table has every event the workflow produces, but
 * the client could only see a filtered subset (ARTIFACT_TYPES) via
 * ArtifactsPanel. This component exposes the full log so the client can
 * audit exactly what happened, in order.
 *
 * Data source: the same blackboard query KanbanPage already runs
 * (getBlackboard) — no extra fetch needed.
 *
 * Design decisions:
 * - Sorted by sequenceNumber (ascending) — that is the canonical event order.
 * - Each row is collapsible to show full content — keeps the list scannable.
 * - Event type rendered as a coloured badge using the same colour buckets
 *   the rest of the UI uses (brand / warn / refuse / neutral).
 */

// Colour bucket per event-type prefix — keeps the log scannable at a glance.
function eventVariant(eventType: string): string {
  if (eventType.startsWith('architecture') || eventType.startsWith('module'))
    return 'text-brand border-brand/30 bg-brand/10'
  if (eventType.startsWith('api_contract') || eventType.startsWith('data_model'))
    return 'text-glow-purple border-glow-purple/30 bg-glow-purple/10'
  if (eventType.startsWith('code_artifact'))
    return 'text-mode-advise border-mode-advise/30 bg-mode-advise/10'
  if (eventType.startsWith('test_case'))
    return 'text-glow-amber border-glow-amber/30 bg-glow-amber/10'
  if (eventType.startsWith('approval') || eventType.startsWith('gate'))
    return 'text-mode-warn border-mode-warn/30 bg-mode-warn/10'
  if (eventType.startsWith('error') || eventType.startsWith('fail'))
    return 'text-mode-refuse border-mode-refuse/30 bg-mode-refuse/10'
  return 'text-text-secondary border-glass-border bg-surface-overlay/40'
}

function formatTime(iso: string): string {
  try {
    return new Date(iso).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  } catch {
    return iso
  }
}

/** One-line preview of event content — first string value found, or key count. */
function contentPreview(content: Record<string, unknown>): string {
  const entries = Object.entries(content ?? {})
  if (entries.length === 0) return '(no content)'
  for (const [, v] of entries) {
    if (typeof v === 'string' && v.trim()) {
      return v.length > 120 ? v.slice(0, 120) + '\u2026' : v
    }
  }
  return `${entries.length} field${entries.length === 1 ? '' : 's'}`
}

function EventRow({
  event,
  expertNames,
}: {
  event: BlackboardEvent
  expertNames: Map<string, string>
}) {
  const [expanded, setExpanded] = useState(false)

  const expertName =
    (event.postedByExpertId && expertNames.get(event.postedByExpertId)) ||
    (event.postedByClient ? 'client' : 'system')

  return (
    <div
      className={cn(
        'border-b border-glass-border/40 px-3 py-2 last:border-b-0',
        'transition-colors hover:bg-surface-overlay/30'
      )}
    >
      <button
        className="flex w-full items-start gap-3 text-left"
        onClick={() => setExpanded((p) => !p)}
        aria-expanded={expanded}
      >
        {/* Sequence number */}
        <span className="mt-0.5 min-w-[2rem] text-right text-[10px] font-mono text-text-disabled">
          #{event.sequenceNumber}
        </span>

        {/* Event type badge */}
        <span
          className={cn(
            'mt-0.5 shrink-0 rounded border px-1.5 py-0.5 text-[10px] font-medium',
            eventVariant(event.eventType)
          )}
        >
          {event.eventType.replace(/_/g, '\u00a0')}
        </span>

        {/* Content preview + meta */}
        <span className="min-w-0 flex-1">
          <span className="block truncate text-xs text-text-secondary">
            {contentPreview(event.content)}
          </span>
          <span className="mt-0.5 flex items-center gap-2 text-[10px] text-text-disabled">
            <span>{expertName}</span>
            <span aria-hidden="true">·</span>
            <span>{formatTime(event.postedAt)}</span>
            {event.revision > 0 && (
              <>
                <span aria-hidden="true">·</span>
                <span>rev {event.revision}</span>
              </>
            )}
          </span>
        </span>

        {/* Expand chevron */}
        <span
          className={cn(
            'mt-1 shrink-0 text-[10px] text-text-disabled transition-transform',
            expanded && 'rotate-180'
          )}
          aria-hidden="true"
        >
          ▾
        </span>
      </button>

      {/* Expanded full content */}
      {expanded && (
        <pre className="mt-2 overflow-x-auto rounded bg-surface-base px-3 py-2 text-[11px] text-text-secondary">
          {JSON.stringify(event.content, null, 2)}
        </pre>
      )}
    </div>
  )
}

export function ActivityLog({
  events,
  expertNames,
}: {
  events: BlackboardEvent[]
  expertNames: Map<string, string>
}) {
  const [open, setOpen] = useState(false)

  if (events.length === 0) return null

  // Canonical order: ascending sequenceNumber.
  const sorted = [...events].sort((a, b) => a.sequenceNumber - b.sequenceNumber)

  return (
    <div className="mt-6">
      {/* Section header — same style as ArtifactsPanel / FilesPanel */}
      <div className="mb-2 flex items-center justify-between">
        <span className="text-xs font-semibold uppercase tracking-wide text-text-secondary">
          Activity Log ({events.length})
        </span>
        <button
          onClick={() => setOpen((p) => !p)}
          className="text-[10px] font-medium uppercase tracking-wider text-brand hover:underline"
        >
          {open ? 'Hide' : 'Show'}
        </button>
      </div>

      {open && (
        <Card className="overflow-hidden p-0">
          <div className="max-h-[28rem] overflow-y-auto">
            {sorted.map((event) => (
              <EventRow key={event.id} event={event} expertNames={expertNames} />
            ))}
          </div>
        </Card>
      )}
    </div>
  )
}
