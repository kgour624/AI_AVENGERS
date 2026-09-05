import type { L3Event } from '@/types/memory'
import { formatRelativeTime } from '@/utils/format'

/**
 * L3 master event log visualization.
 * Source: FRONTEND_SYSTEM_DESIGN.md section 10 ("Recent Activity
 * (Timeline)" panel: "Today / 14:23 [DB] Decided: PostgreSQL sharding")
 * and AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 5 (master_event_log
 * table - append-only, has reasoning/decisionMade/eventType columns).
 *
 * WHY "Today"/"Yesterday" day-grouping headers, not just a flat list:
 * matches the wireframe exactly, which explicitly shows these group
 * headers. Implemented with date-fns's isToday/isYesterday rather
 * than hand-rolled date math, since date-fns is already a listed
 * dependency (section 2) specifically for this kind of formatting.
 */
import { isToday, isYesterday, format } from 'date-fns'

function dayGroupLabel(dateStr: string): string {
  const date = new Date(dateStr)
  if (isToday(date)) return 'Today'
  if (isYesterday(date)) return 'Yesterday'
  return format(date, 'MMM d')
}

export function ProjectTimeline({ events }: { events: L3Event[] }) {
  if (events.length === 0) {
    return <p className="text-sm text-text-secondary">No activity yet.</p>
  }

  // Group consecutive events by day label, preserving the events'
  // existing sort order (assumed newest-first, matching
  // AI_AVENGERS_SYSTEM_ARCHITECTURE.md's idx_l3_project index which
  // orders by created_at DESC - NOT re-sorted here, since re-sorting
  // client-side would silently mask a backend ordering bug instead of
  // surfacing it).
  const groups: { label: string; events: L3Event[] }[] = []
  for (const event of events) {
    const label = dayGroupLabel(event.createdAt)
    const lastGroup = groups[groups.length - 1]
    if (lastGroup && lastGroup.label === label) {
      lastGroup.events.push(event)
    } else {
      groups.push({ label, events: [event] })
    }
  }

  return (
    <div className="space-y-4">
      {groups.map((group) => (
        <div key={group.label}>
          <p className="mb-1 text-xs font-medium uppercase tracking-wide text-text-disabled">
            {group.label}
          </p>
          <div className="space-y-1">
            {group.events.map((event) => (
              <p key={event.id} className="text-sm text-text-secondary">
                <span className="text-text-disabled">{format(new Date(event.createdAt), 'HH:mm')}</span>{' '}
                {event.expertName && <span className="font-medium text-text-primary">[{event.expertName}]</span>}{' '}
                {event.decisionMade ?? event.eventType}
              </p>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}
