import { useQuery } from '@tanstack/react-query'
import { getProjectMemory } from '@/api/memory'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatRelativeTime } from '@/utils/format'

const TYPE_LABELS: Record<string, string> = {
  decision: 'Decision',
  code: 'Code',
  error: 'Error',
  fix: 'Fix',
  recommendation: 'Recommendation',
}

/**
 * Feature #5 fix (docs bug list): consolidated cross-expert decisions
 * (L2 group memory) had no UI view at all - only the L3 timeline was
 * rendered (ProjectTimeline.tsx). This is deliberately a separate
 * component/section rather than merged into ProjectTimeline: L2 is
 * "what has been decided" (deduplicated, importance-ranked, one row
 * per decision), L3 is "what happened" (append-only log of every
 * event) - different questions, same underlying data model, but
 * merging them into one list would blur that distinction the backend
 * itself maintains as two separate tables/concepts.
 */
export function ProjectMemoryPanel({ projectId }: { projectId: string }) {
  const { data: entries, isLoading } = useQuery({
    queryKey: ['projects', projectId, 'memory'],
    queryFn: () => getProjectMemory(projectId),
  })

  if (isLoading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-16" />
        ))}
      </div>
    )
  }

  if (!entries || entries.length === 0) {
    return <p className="text-sm text-text-secondary">No cross-expert decisions recorded yet.</p>
  }

  return (
    <div className="space-y-2">
      {entries.map((entry) => (
        <div key={entry.id} className="rounded-md border border-surface-border bg-surface-raised p-3">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-brand">{entry.expertName}</span>
            <span className="text-xs text-text-disabled">
              {TYPE_LABELS[entry.memoryType] ?? entry.memoryType}
              {' \u00b7 '}
              {formatRelativeTime(entry.createdAt)}
            </span>
          </div>
          <p className="mt-1 text-sm text-text-primary">{entry.content}</p>
          {entry.context && <p className="mt-1 text-xs text-text-secondary">{entry.context}</p>}
        </div>
      ))}
    </div>
  )
}
