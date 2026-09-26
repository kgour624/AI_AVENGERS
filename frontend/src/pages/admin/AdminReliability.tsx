import { useQuery } from '@tanstack/react-query'
import { getReliabilityStatus, getReliabilityEvents } from '@/api/admin'
import { Card } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatRelativeTime } from '@/utils/format'

/**
 * Reliability (C10).
 *
 * WHY this screen exists: the reliability service has published an SLO, an error
 * budget and an append-only audit trail since C10 — through a JSON endpoint that
 * only a curl user could read. The person who has to decide whether the system is
 * healthy should not need a shell.
 *
 * What is shown is exactly what the service reports, including the honest limits
 * of it: the availability counters are process-lifetime, and the window below is
 * the target the budget is published for, not the period actually measured. The
 * page says so rather than implying a rolling 30-day view that does not exist.
 */

function verdictColour(status: string): string {
  switch (status) {
    case 'ok':
      return 'bg-mode-advise/15 text-mode-advise'
    case 'degraded':
      return 'bg-mode-refuse/15 text-mode-refuse'
    default:
      return 'bg-surface-overlay text-text-secondary'
  }
}

function formatUptime(seconds: number): string {
  if (seconds <= 0) return '—'
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (d > 0) return `${d}d ${h}h`
  if (h > 0) return `${h}h ${m}m`
  return `${m}m`
}

function AdminReliability() {
  const { data: status, isLoading } = useQuery({
    queryKey: ['admin', 'reliability', 'status'],
    queryFn: getReliabilityStatus,
    refetchInterval: 30000,
  })
  const { data: events } = useQuery({
    queryKey: ['admin', 'reliability', 'events'],
    queryFn: () => getReliabilityEvents(50),
    refetchInterval: 30000,
  })

  const slo = status?.slo

  return (
    <div className="p-6">
      <div className="mb-1 flex items-center gap-3">
        <h1 className="text-xl font-semibold">Reliability</h1>
        {status && (
          <span className={`rounded px-2 py-0.5 text-xs font-medium ${verdictColour(status.status)}`}>
            {status.status}
          </span>
        )}
        {status?.version && <span className="text-xs text-text-disabled">{status.version}</span>}
      </div>
      <p className="mb-6 text-sm text-text-secondary">
        Dependency health, the availability target and the audit trail. Refreshes every 30 seconds.
      </p>

      {isLoading ? (
        <Skeleton className="h-40" />
      ) : (
        <>
          {slo && (
            <div className="mb-5 grid grid-cols-4 gap-4">
              <Card className="p-4">
                <p className="text-[10px] uppercase tracking-wider text-text-disabled">
                  Availability
                </p>
                <p className="mt-1 text-lg font-medium text-text-primary">
                  {(slo.availability * 100).toFixed(2)}%
                </p>
                <p className="mt-0.5 text-[10px] text-text-disabled">
                  target {(slo.availabilityTarget * 100).toFixed(2)}%
                </p>
              </Card>
              <Card className="p-4">
                <p className="text-[10px] uppercase tracking-wider text-text-disabled">
                  Error budget left
                </p>
                <p className="mt-1 text-lg font-medium text-text-primary">
                  {(slo.errorBudgetRemaining * 100).toFixed(0)}%
                </p>
                <p className="mt-0.5 text-[10px] text-text-disabled">
                  published over {slo.errorBudgetWindowDays}d
                </p>
              </Card>
              <Card className="p-4">
                <p className="text-[10px] uppercase tracking-wider text-text-disabled">
                  LLM calls / errors
                </p>
                <p className="mt-1 text-lg font-medium text-text-primary">
                  {slo.llmCallsTotal} / {slo.llmErrorsTotal}
                </p>
                <p className="mt-0.5 text-[10px] text-text-disabled">
                  {(slo.errorRate * 100).toFixed(2)}% error rate
                </p>
              </Card>
              <Card className="p-4">
                <p className="text-[10px] uppercase tracking-wider text-text-disabled">Uptime</p>
                <p className="mt-1 text-lg font-medium text-text-primary">
                  {formatUptime(slo.uptimeSeconds)}
                </p>
                <p className="mt-0.5 text-[10px] text-text-disabled">this process</p>
              </Card>
            </div>
          )}

          {/* The counters are process-lifetime, so a reader must not mistake them
              for a rolling window. */}
          <p className="mb-5 text-[10px] text-text-disabled">
            Counters are for the life of this process, not a rolling window — a restart resets
            them, and the budget figure is the target they are measured against.
          </p>

          <h2 className="mb-2 text-sm font-medium text-text-secondary">Dependencies</h2>
          <Card className="mb-5 p-4">
            {!status || status.components.length === 0 ? (
              <p className="text-xs text-text-disabled">No dependency probes are registered.</p>
            ) : (
              <table className="w-full text-xs">
                <thead>
                  <tr className="text-left text-text-disabled">
                    <th className="py-1 pr-3 font-medium">Component</th>
                    <th className="py-1 pr-3 font-medium">Status</th>
                    <th className="py-1 font-medium">Error</th>
                  </tr>
                </thead>
                <tbody>
                  {status.components.map((c) => (
                    <tr key={c.name}>
                      <td className="py-1 pr-3 text-text-secondary">{c.name}</td>
                      <td className="py-1 pr-3">
                        <span
                          className={
                            c.status === 'healthy' ? 'text-mode-advise' : 'text-mode-refuse'
                          }
                        >
                          {c.status}
                        </span>
                      </td>
                      <td className="py-1 text-text-disabled">{c.error ?? '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </Card>

          <h2 className="mb-2 text-sm font-medium text-text-secondary">Audit trail</h2>
          <Card className="p-4">
            {!events || events.length === 0 ? (
              <p className="text-xs text-text-disabled">
                No reliability events yet. Health transitions and error-budget breaches are
                recorded here as they happen.
              </p>
            ) : (
              <ul className="space-y-2">
                {events.map((e) => (
                  <li key={e.id} className="border-l-2 border-surface-border pl-2">
                    <p className="text-xs text-text-secondary">
                      <span
                        className={
                          e.severity === 'error'
                            ? 'text-mode-refuse'
                            : e.severity === 'warn'
                              ? 'text-glow-amber'
                              : 'text-text-primary'
                        }
                      >
                        {e.kind}
                      </span>
                      {e.component ? ` · ${e.component}` : ''}
                      <span className="ml-2 text-[10px] text-text-disabled">
                        {formatRelativeTime(e.createdAt)}
                      </span>
                    </p>
                    {/* detail is an arbitrary JSON blob; an explicit null check
                        (not a truthiness guard) keeps its type out of the JSX
                        position, where 'unknown' is not a renderable value. */}
                    {e.detail != null && (
                      <pre className="mt-0.5 overflow-x-auto whitespace-pre-wrap text-[10px] text-text-disabled">
                        {typeof e.detail === 'string' ? e.detail : JSON.stringify(e.detail)}
                      </pre>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </Card>
        </>
      )}
    </div>
  )
}

export const Component = AdminReliability
