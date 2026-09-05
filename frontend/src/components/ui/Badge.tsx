import type { ResponseMode } from '@/types/expert'
import { getModeBadgeConfig } from '@/types/expert'
import { cn } from '@/utils/cn'

/**
 * Renders a ResponseMode as a colored chip.
 * WHY this is a thin wrapper around getModeBadgeConfig rather than
 * its own switch statement: keeps exactly one place
 * (types/expert.ts::getModeBadgeConfig) that has to change if the
 * backend ever adds a 6th ResponseMode - the `never` exhaustiveness
 * check there will fail to compile, and this component inherits that
 * safety automatically without needing its own separate check.
 */
export interface ModeBadgeProps {
  mode: ResponseMode
  className?: string
}

export function ModeBadge({ mode, className }: ModeBadgeProps) {
  const config = getModeBadgeConfig(mode)

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-xs font-medium',
        config.color,
        config.bgClass,
        className
      )}
    >
      <span aria-hidden="true">{config.icon}</span>
      {config.label}
    </span>
  )
}

/** Generic pill badge for non-mode use cases (e.g. depth level, domain tag). */
export interface BadgeProps {
  children: React.ReactNode
  variant?: 'neutral' | 'brand'
  className?: string
}

export function Badge({ children, variant = 'neutral', className }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
        variant === 'neutral' && 'bg-surface-overlay text-text-secondary',
        variant === 'brand' && 'bg-brand/10 text-brand',
        className
      )}
    >
      {children}
    </span>
  )
}
