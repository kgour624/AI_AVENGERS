import { cn } from '@/utils/cn'

/** Loading placeholder. WHY a dedicated component instead of inline `animate-pulse` divs: section 12 ("Skeleton screens - perceived performance") calls this out as a deliberate pattern, used consistently wherever a loader is in flight. */
export function Skeleton({ className }: { className?: string }) {
  return <div className={cn('animate-pulse rounded-md bg-surface-overlay', className)} />
}
