import { formatDistanceToNow } from 'date-fns'

/** WHY a wrapper instead of calling formatDistanceToNow directly everywhere: keeps the `addSuffix: true` option (giving "2 hours ago" instead of bare "2 hours") consistent across every call site, matching the wireframe's "Last: 2h ago" format. */
export function formatRelativeTime(dateStr: string): string {
  return formatDistanceToNow(new Date(dateStr), { addSuffix: true })
}

export function formatCostUsd(amount: number): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount)
}
