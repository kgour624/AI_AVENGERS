import type { ResponseMode } from '@/types/expert'
import { getModeBadgeConfig } from '@/types/expert'
import { cn } from '@/utils/cn'

export interface ModeBadgeProps {
  mode: ResponseMode
  className?: string
}

export function ModeBadge({ mode, className }: ModeBadgeProps) {
  const config = getModeBadgeConfig(mode)
  return (
    <span className={cn(
      'inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-xs font-medium',
      'backdrop-blur-sm',
      config.color, config.bgClass,
      className
    )}>
      <span aria-hidden="true">{config.icon}</span>
      {config.label}
    </span>
  )
}

export interface BadgeProps {
  children: React.ReactNode
  variant?: 'neutral' | 'brand' | 'success' | 'warn' | 'danger'
  className?: string
}

export function Badge({ children, variant = 'neutral', className }: BadgeProps) {
  return (
    <span className={cn(
      'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium backdrop-blur-sm',
      variant === 'neutral' && 'border-surface-border bg-surface-overlay/60 text-text-secondary',
      variant === 'brand'   && [
        'border-glow-purple/30 bg-glow-purple/10 text-glow-purple',
        'shadow-[0_0_8px_oklch(68%_0.28_295_/_0.2)]',
      ],
      variant === 'success' && 'border-mode-advise/30 bg-mode-advise/10 text-mode-advise',
      variant === 'warn'    && 'border-mode-warn/30 bg-mode-warn/10 text-mode-warn',
      variant === 'danger'  && 'border-mode-refuse/30 bg-mode-refuse/10 text-mode-refuse',
      className
    )}>
      {children}
    </span>
  )
}
