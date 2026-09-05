import { useQuery } from '@tanstack/react-query'
import { getAdminStats } from '@/api/admin'
import { Card } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatCostUsd } from '@/utils/format'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 11 ("Admin Dashboard"
 * wireframe: Experts/Clients/Projects/Cost stat cards + violations bar
 * + expert ratings list), rewired against the REAL AdminStats shape
 * confirmed from backend-go/internal/admin/admin_handler.go (see
 * api/admin.ts's header comment for the full list of corrections).
 *
 * WHY llmTotalCost (not a "monthlyCostUsd") is what's shown: the real
 * GetStats handler returns a running total from the ModelGateway's
 * in-memory counter (gwStats["total_cost"]), not a monthly figure -
 * there is no month-boundary reset logic anywhere in the backend.
 * Labeling it "Total LLM cost" rather than "$X/month" (as the original
 * wireframe implied) avoids implying a time-boundary that doesn't
 * actually exist server-side.
 */
export default function AdminDashboard() {
  const { data: stats, isLoading, isError } = useQuery({
    queryKey: ['admin', 'stats'],
    queryFn: getAdminStats,
  })

  if (isLoading) {
    return (
      <div className="grid grid-cols-2 gap-4 p-6 sm:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-24" />
        ))}
      </div>
    )
  }

  if (isError || !stats) {
    return <p className="p-6 text-sm text-mode-refuse">Failed to load system stats.</p>
  }

  return (
    <div className="p-6">
      <h1 className="mb-6 text-xl font-semibold">System Overview</h1>

      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <Card>
          <p className="text-xs text-text-secondary">Experts</p>
          <p className="text-2xl font-semibold">{stats.experts.total}</p>
          <p className="text-xs text-text-disabled">{stats.experts.active} active</p>
        </Card>
        <Card>
          <p className="text-xs text-text-secondary">Clients</p>
          <p className="text-2xl font-semibold">{stats.clients}</p>
        </Card>
        <Card>
          <p className="text-xs text-text-secondary">Projects</p>
          <p className="text-2xl font-semibold">{stats.projects}</p>
        </Card>
        <Card>
          <p className="text-xs text-text-secondary">Total LLM cost</p>
          <p className="text-2xl font-semibold">{formatCostUsd(stats.llmTotalCost)}</p>
          <p className="text-xs text-text-disabled">{stats.llmTotalCalls} calls</p>
        </Card>
      </div>

      <Card className="mt-6">
        <p className="text-sm text-text-secondary">China Wall violations</p>
        <p className="mt-1 text-lg font-medium">{stats.violations}</p>
        <p className="text-xs text-text-disabled">
          {/* WHY no "(2.1%)" rate shown: violations here is a raw lifetime
              COUNT(*), not a rate over any time window - fabricating a
              percentage would require a denominator (total responses in
              some period) the backend doesn't provide. */}
          Total messages: {stats.messages}
        </p>
      </Card>

      <Card className="mt-4">
        <p className="text-sm text-text-secondary">Average rating across experts</p>
        <p className="mt-1 text-lg font-medium">{stats.avgRating} \u2b50</p>
      </Card>
    </div>
  )
}

export const Component = AdminDashboard
