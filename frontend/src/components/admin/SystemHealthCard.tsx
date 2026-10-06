import { Card } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'

type Props = {
  data: any
  isLoading: boolean
  error: unknown
}

export function SystemHealthCard({ data, isLoading, error }: Props) {
  if (isLoading) return <Card><Skeleton className="h-32" /></Card>
  if (error) return <Card><p className="text-sm text-mode-refuse">Health failed to load</p></Card>
  if (!data) return <Card><p className="py-6 text-center text-sm text-text-disabled">No data</p></Card>

  const flagOn = data.flagEnabled ?? data.flag_enabled ?? false
  const flagSource = data.flagSource ?? data.flag_source ?? 'default'

  const counts = data.counts ?? {}
  const totalParents = counts.totalParents ?? counts.total_parents ?? 0
  const totalChildren = counts.totalChildren ?? counts.total_children ?? 0
  const orphans = counts.orphans ?? 0
  const experts = counts.experts ?? 0

  const runtime = data.runtime ?? {}
  const goRoutines = runtime.goRoutines ?? runtime.go_routines ?? 0
  const allocMb = Number(runtime.allocMb ?? runtime.alloc_mb ?? 0)
  const uptimeSec = runtime.uptimeSec ?? runtime.uptime_sec ?? 0

  const dbPool = data.dbPool ?? data.db_pool ?? {}
  const acquiredConns = dbPool.acquiredConns ?? dbPool.acquired_conns ?? 0
  const idleConns = dbPool.idleConns ?? dbPool.idle_conns ?? 0
  const totalConns = dbPool.totalConns ?? dbPool.total_conns ?? 0

  const checkedAt = data.checkedAt ?? data.checked_at ?? new Date().toISOString()
  const rConfig = data.retrievalConfig ?? data.retrieval_config

  return (
    <Card>
      <div className="flex flex-wrap gap-2">
        <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${flagOn ? 'bg-emerald-500/15 text-emerald-600' : 'bg-surface-secondary text-text-secondary'}`}>
          {flagOn ? 'FLAG ON' : 'FLAG OFF'} · {flagSource}
        </span>
        <span className="rounded-full bg-surface-secondary px-2 py-0.5 text-xs">{totalParents} parents</span>
        <span className="rounded-full bg-surface-secondary px-2 py-0.5 text-xs">{totalChildren} children</span>
        <span className={`rounded-full px-2 py-0.5 text-xs ${orphans > 0 ? 'bg-amber-500/15 text-amber-600' : 'bg-surface-secondary text-text-secondary'}`}>
          {orphans} orphans
        </span>
        <span className="rounded-full bg-surface-secondary px-2 py-0.5 text-xs">{experts} experts</span>
      </div>
      <div className="mt-3 grid grid-cols-2 gap-3 text-xs md:grid-cols-4">
        <div className="rounded border border-surface-border p-2">
          <p className="text-text-disabled">Go routines</p>
          <p className="font-mono font-semibold">{goRoutines}</p>
        </div>
        <div className="rounded border border-surface-border p-2">
          <p className="text-text-disabled">Alloc MB</p>
          <p className="font-mono font-semibold">{allocMb.toFixed(1)}</p>
        </div>
        <div className="rounded border border-surface-border p-2">
          <p className="text-text-disabled">DB acquired/idle/total</p>
          <p className="font-mono font-semibold">{acquiredConns}/{idleConns}/{totalConns}</p>
        </div>
        <div className="rounded border border-surface-border p-2">
          <p className="text-text-disabled">Uptime</p>
          <p className="font-mono font-semibold">{Math.floor(uptimeSec / 60)}m</p>
        </div>
      </div>
      <p className="mt-2 text-[10px] text-text-disabled">
        checked {new Date(checkedAt).toLocaleString()} · parent_top_k {rConfig?.parentTopK ?? rConfig?.parent_top_k ?? '-'} · child_top_k {rConfig?.childTopK ?? rConfig?.child_top_k ?? '-'}
      </p>
    </Card>
  )
}
