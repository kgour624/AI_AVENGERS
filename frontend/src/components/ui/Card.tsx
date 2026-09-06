import type { HTMLAttributes } from 'react'
import { cn } from '@/utils/cn'

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  /**
   * ARC-51 §4 (docs/ARC51_UI_CONTRACT.md): optional hover border-glow.
   * Default 'none' renders PIXEL-IDENTICAL to the pre-ARC-51 Card -
   * every existing caller that doesn't pass this prop is unaffected.
   */
  glow?: 'cyan' | 'purple' | 'none'
}

const GLOW_CLASSES: Record<NonNullable<CardProps['glow']>, string> = {
  none: '',
  cyan: 'transition-[border-color,box-shadow] duration-150 ease-arc hover:border-glow-cyan/50 hover:shadow-[0_0_24px_-8px_var(--glow-cyan)]',
  purple:
    'transition-[border-color,box-shadow] duration-150 ease-arc hover:border-glow-purple/50 hover:shadow-[0_0_24px_-8px_var(--glow-purple)]',
}

export function Card({ className, children, glow = 'none', ...props }: CardProps) {
  return (
    <div
      className={cn(
        'rounded-lg border border-surface-border bg-surface-raised p-4',
        GLOW_CLASSES[glow],
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}
