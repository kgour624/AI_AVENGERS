import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getUsage, getUsageBudgets, getUsageAlerts, type UsageGroupBy } from '@/api/admin'
import { Card } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'

/**
 * Usage & Cost (C5).
 *
 * WHY this screen exists: usage has been recorded per call since C5 — provider,
 * model, tokens and cost all land in one table through the gateway's single choke
 * point — but nothing ever showed it. So the honest answer to "what is this
 * costing, and where is it going" was "ask the database". Nothing here is
 * invented: every number is a SUM over recorded calls.
 *
 * Grouping is whitelisted by the backend (model / expert / project / tenant /
 * use_case), and cost per group comes from the same aggregation the budget
 * checks read, so this screen and a breached budget can never disagree.
 */

const GROUPS: { value: UsageGroupBy; label: string }[] = [
  { value: 'model', label: 'Model' },
  { value: 'use_case', label: 'What it was spent on' },
  { value: 'expert', label: 'Expert' },
  { value: 'project', label: 'Project' },
  { value: 'tenant', label: 'Client' },
]

function money(v: number): string {
  return `$${v.toFixed(v < 1 ? 4 : 2)}`
}

function AdminUsage() {
  const [groupBy, setGroupBy] = useState<UsageGroupBy>('model')

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'usage', groupBy],
    queryFn: () => getUsage(groupBy),
  })
  const { data: budgets } = useQuery({
    queryKey: ['admin', 'usage-budgets'],
    queryFn: getUsageBudgets,
  })
  const { data: alerts } = useQuery({
    queryKey: ['admin', 'usage-alerts'],
    queryFn: getUsageAlerts,
    refetchInterval: 30000,
  })

  const rows = data?.rows ?? []
  const totalCalls = rows.reduce((n, r) => n + r.calls, 0)
  const totalCost = rows.reduce((n, r) => n + r.costUsd, 0)

  return (
    <div className="p-6">
      <h1 className="mb-1 text-xl font-semibold">Usage & Cost</h1>
      <p className="mb-6 text-sm text-text-secondary">
        Every call the gateway makes is recorded with its provider, model and token counts, so these
        numbers are measured rather than estimated.
      </p>

      {/* Alerts first: a breached budget is the one thing on this page that needs
          acting on, so it does not sit under a table. */}
      {alerts && alerts.length > 0 && (
        <div className="mb-5 rounded-lg border border-mode-refuse/30 bg-mode-refuse/10 p-3">
          <p className="text-sm font-medium text-mode-refuse">
            {alerts.length} budget{alerts.length === 1 ? '' : 's'} at or over threshold
          </p>
          <ul className="mt-1 space-y-1">
            {alerts.map((a, i) => (
              <li key={i} className="text-xs text-text-secondary">
                {a.breached ? 'Over limit' : 'Near limit'} — {money(a.spendUsd)} of{' '}
                {money(a.limitUsd)} ({a.percent.toFixed(0)}%)
              </li>
            ))}
          </ul>
        </div>
      )}

      <div className="mb-4 flex flex-wrap items-center gap-3">
        <span className="text-xs text-text-secondary">Group by</span>
        <div className="flex flex-wrap gap-1">
          {GROUPS.map((g) => (
            <button
              key={g.value}
              onClick={() => setGroupBy(g.value)}
              className={`rounded px-2 py-1 text-xs ${
                groupBy === g.value
                  ? 'bg-surface-overlay text-text-primary'
                  : 'text-text-secondary hover:bg-surface-overlay/60'
              }`}
            >
              {g.label}
            </button>
          ))}
        </div>
      </div>

      {isLoading ? (
        <Skeleton className="h-40" />
      ) : (
        <>
          <div className="mb-4 grid grid-cols-3 gap-4">
            <Card className="p-4">
              <p className="text-[10px] uppercase tracking-wider text-text-disabled">Calls</p>
              <p className="mt-1 text-lg font-medium text-text-primary">{totalCalls}</p>
            </Card>
            <Card className="p-4">
              <p className="text-[10px] uppercase tracking-wider text-text-disabled">Cost</p>
              <p className="mt-1 text-lg font-medium text-text-primary">{money(totalCost)}</p>
            </Card>
            <Card className="p-4">
              <p className="text-[10px] uppercase tracking-wider text-text-disabled">
                Average per call
              </p>
              <p className="mt-1 text-lg font-medium text-text-primary">
                {totalCalls > 0 ? money(totalCost / totalCalls) : '—'}
              </p>
            </Card>
          </div>

          <Card className="p-4">
            {rows.length === 0 ? (
              <p className="text-xs text-text-disabled">
                No usage recorded in this period. Calls appear here as soon as the gateway serves
                them.
              </p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="text-left text-text-disabled">
                      <th className="py-1 pr-3 font-medium">Where</th>
                      <th className="py-1 pr-3 font-medium">Calls</th>
                      <th className="py-1 pr-3 font-medium">Input tokens</th>
                      <th className="py-1 pr-3 font-medium">Output tokens</th>
                      <th className="py-1 font-medium">Cost</th>
                    </tr>
                  </thead>
                  <tbody>
                    {rows.map((r) => (
                      <tr key={r.key}>
                        <td className="py-1 pr-3 text-text-secondary">{r.label}</td>
                        <td className="py-1 pr-3 text-text-secondary">{r.calls}</td>
                        <td className="py-1 pr-3 text-text-secondary">{r.inputTokens}</td>
                        <td className="py-1 pr-3 text-text-secondary">{r.outputTokens}</td>
                        <td className="py-1 text-text-primary">{money(r.costUsd)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Card>
        </>
      )}

      {/* Budgets: the limits the alerts above are measured against. */}
      <div className="mt-5">
        <h2 className="mb-2 text-sm font-medium text-text-secondary">Budgets</h2>
        {!budgets || (!budgets.global && budgets.tenants.length === 0) ? (
          <Card className="p-4">
            <p className="text-xs text-text-disabled">
              No budget is set, so spending is reported but never capped.
            </p>
          </Card>
        ) : (
          <Card className="p-4">
            <table className="w-full text-xs">
              <thead>
                <tr className="text-left text-text-disabled">
                  <th className="py-1 pr-3 font-medium">Scope</th>
                  <th className="py-1 pr-3 font-medium">Spent</th>
                  <th className="py-1 pr-3 font-medium">Limit</th>
                  <th className="py-1 font-medium">Used</th>
                </tr>
              </thead>
              <tbody>
                {budgets.global && (
                  <tr>
                    <td className="py-1 pr-3 text-text-secondary">Global</td>
                    <td className="py-1 pr-3 text-text-secondary">{money(budgets.global.spendUsd)}</td>
                    <td className="py-1 pr-3 text-text-secondary">{money(budgets.global.limitUsd)}</td>
                    <td className="py-1 text-text-primary">{budgets.global.percent.toFixed(0)}%</td>
                  </tr>
                )}
                {budgets.tenants.map((t) => (
                  <tr key={t.tenantId ?? 'tenant'}>
                    <td className="py-1 pr-3 text-text-secondary">{t.tenantId ?? 'Client'}</td>
                    <td className="py-1 pr-3 text-text-secondary">{money(t.spendUsd)}</td>
                    <td className="py-1 pr-3 text-text-secondary">{money(t.limitUsd)}</td>
                    <td className="py-1 text-text-primary">{t.percent.toFixed(0)}%</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
        )}
      </div>
    </div>
  )
}

export const Component = AdminUsage
