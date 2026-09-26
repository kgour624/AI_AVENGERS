import { useQuery } from '@tanstack/react-query'
import { getAdminViolations, getAdminRatings } from '@/api/admin'
import { Card } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatRelativeTime } from '@/utils/format'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 11 ("Rating Analytics -
 * Low-rated responses flagged for review, Expert performance over
 * time, Topic-wise accuracy").
 *
 * WHY this page shows violations + per-expert rating aggregates, but
 * NOT "expert performance over time" (a time-series) or "topic-wise
 * accuracy": verified against the real GetRatings handler - it
 * returns one aggregate row per expert (avg score, good/bad counts),
 * with no time dimension and no per-topic breakdown anywhere in the
 * query. A time-series chart or topic breakdown would require new
 * backend aggregation queries that don't exist yet - not something to
 * fake with a chart that has no real data behind its X-axis.
 */
function AdminStats() {
  const { data: violations, isLoading: violationsLoading } = useQuery({
    queryKey: ['admin', 'violations'],
    queryFn: getAdminViolations,
  })
  const { data: ratings, isLoading: ratingsLoading } = useQuery({
    queryKey: ['admin', 'ratings'],
    queryFn: getAdminRatings,
  })

  return (
    <div className="p-6">
      <h1 className="mb-4 text-xl font-semibold">Stats & Analytics</h1>

      <h2 className="mb-2 text-sm font-medium text-text-secondary">Expert Ratings</h2>
      {ratingsLoading ? (
        <Skeleton className="h-32" />
      ) : (
        <div className="space-y-2">
          {ratings?.map((r) => (
            <Card key={r.expertName} glow="cyan" className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-text-primary">{r.expertName}</p>
                <p className="text-xs text-text-secondary">{r.domain}</p>
              </div>
              <div className="text-right">
                {/* Same bare-\uXXXX-in-JSX-text bug found repeatedly this
                    session (Header, AdminDashboard, AdminClients,
                    ProjectMemoryPanel) - this file had never been
                    audited for it until now. Wrapped in {'...'}. */}
                <p className="text-sm">
                  {'⭐'} {r.avgScore.toFixed(1)} ({r.totalRatings})
                </p>
                <p className="text-xs text-text-disabled">
                  {r.goodRatings} good{' · '}
                  {r.badRatings} bad
                </p>
              </div>
            </Card>
          ))}
        </div>
      )}

      <h2 className="mb-2 mt-6 text-sm font-medium text-text-secondary">
        China Wall Violations (most recent 100)
      </h2>
      {violationsLoading ? (
        <Skeleton className="h-32" />
      ) : violations && violations.length > 0 ? (
        <div className="space-y-1">
          {violations.map((v) => (
            <p key={v.id} className="text-sm text-text-secondary">
              <span className="text-text-disabled">{formatRelativeTime(v.createdAt)}</span>
              {' · '}
              {v.eventType}
              {' — '}
              {v.reasoning}
            </p>
          ))}
        </div>
      ) : (
        <p className="text-sm text-text-secondary">No violations recorded.</p>
      )}
    </div>
  )
}

export const Component = AdminStats
