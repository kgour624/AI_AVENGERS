import type { ReactNode } from 'react'
import { cn } from '@/utils/cn'

/**
 * Tabs — minimal, accessible tab strip.
 *
 * WHY a local primitive: the codebase had no tab component and no `role="tab"`
 * usage anywhere; KanbanPage was a single tall stack, so later panels
 * (Deliverables/Activity/Files) fell below the fold. This is the smallest
 * thing that makes those views switchable without pulling in a dependency.
 *
 * Controlled + generic over the tab key so the page keeps a typed union.
 */
export interface TabItem<K extends string = string> {
  key: K
  label: string
  /** Optional count badge (e.g. number of files / tasks). */
  count?: number
  /** Optional attention dot (e.g. an approval is waiting on this tab). */
  alert?: boolean
}

export interface TabsProps<K extends string = string> {
  tabs: TabItem<K>[]
  active: K
  onChange: (key: K) => void
  className?: string
  /** Optional right-aligned content in the same row as the tab strip. */
  right?: ReactNode
}

export function Tabs<K extends string>({ tabs, active, onChange, className, right }: TabsProps<K>) {
  return (
    <div
      className={cn(
        'flex items-center justify-between gap-4 border-b border-surface-border',
        className
      )}
    >
      <div role="tablist" aria-label="Workflow views" className="flex items-center gap-1">
        {tabs.map((tab) => {
          const isActive = tab.key === active
          return (
            <button
              key={tab.key}
              role="tab"
              type="button"
              aria-selected={isActive}
              onClick={() => onChange(tab.key)}
              className={cn(
                '-mb-px flex items-center gap-1.5 border-b-2 px-3 py-2 text-xs font-semibold uppercase tracking-wide transition-colors',
                isActive
                  ? 'border-brand text-text-primary'
                  : 'border-transparent text-text-secondary hover:text-text-primary'
              )}
            >
              {tab.alert && (
                <span
                  className="h-1.5 w-1.5 rounded-full bg-glow-amber animate-pulse"
                  aria-hidden="true"
                />
              )}
              <span>{tab.label}</span>
              {typeof tab.count === 'number' && (
                <span
                  className={cn(
                    'rounded-full px-1.5 py-0.5 text-[10px] font-medium',
                    isActive ? 'bg-brand/15 text-brand' : 'bg-surface-overlay text-text-disabled'
                  )}
                >
                  {tab.count}
                </span>
              )}
            </button>
          )
        })}
      </div>
      {right && <div className="pb-2">{right}</div>}
    </div>
  )
}
