import type { HTMLAttributes } from 'react'
import { cn } from '@/utils/cn'

/** Thin wrapper enforcing consistent scroll styling (used by chat message list, sidebar). Not a virtualized list - see Phase 3/6 note about @tanstack/react-virtual for 100+ messages (section 12). */
export function ScrollArea({ className, children, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cn('overflow-y-auto', className)} {...props}>
      {children}
    </div>
  )
}
