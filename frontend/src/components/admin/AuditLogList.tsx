import { useQuery } from '@tanstack/react-query'
import { Card } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'
import { getAuditLog } from '@/api/admin'

export function AuditLogList() {
  const q = useQuery({ queryKey: ['admin', 'audit-log'], queryFn: () => getAuditLog({ limit: 20 }) })
  if (q.isLoading) return <Card><Skeleton className="h-24" /></Card>
  if (q.error) return <Card><p className="text-sm text-mode-refuse">Audit log unavailable</p></Card>
  const items = (q.data as any)?.items ?? []
  return (
    <Card>
      <h3 className="text-sm font-semibold">Audit Log</h3>
      <p className="mt-1 text-xs text-text-secondary">Retrieval config changes + admin audit trail.</p>
      <div className="mt-3 divide-y divide-surface-border">
        {items.length === 0 ? <p className="py-6 text-center text-sm text-text-disabled">No audit entries yet</p> : items.map((r: any) => (
          <div key={r.id} className="flex items-center justify-between py-1.5 text-xs">
            <span className="font-mono">{r.action}</span>
            <span className="text-text-disabled">{new Date(r.createdAt ?? r.created_at ?? Date.now()).toLocaleString()}</span>
          </div>
        ))}
      </div>
    </Card>
  )
}
