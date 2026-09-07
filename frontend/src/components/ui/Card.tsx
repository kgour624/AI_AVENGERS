import type { HTMLAttributes } from 'react'
import { cn } from '@/utils/cn'

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  glow?: 'cyan' | 'purple' | 'violet' | 'amber' | 'none'
  glass?: boolean
}

const GLOW_CLASSES: Record<NonNullable<CardProps['glow']>, string> = {
  none: '',
  cyan: [
    'transition-[border-color,box-shadow] duration-200 ease-arc',
    'hover:border-glow-cyan/40',
    'hover:shadow-[0_0_24px_-4px_oklch(78%_0.18_200_/_0.4),_0_0_60px_-12px_oklch(78%_0.18_200_/_0.2)]',
  ].join(' '),
  purple: [
    'transition-[border-color,box-shadow] duration-200 ease-arc',
    'hover:border-glow-purple/40',
    'hover:shadow-[0_0_24px_-4px_oklch(68%_0.28_295_/_0.4),_0_0_60px_-12px_oklch(68%_0.28_295_/_0.2)]',
  ].join(' '),
  violet: [
    'transition-[border-color,box-shadow] duration-200 ease-arc',
    'hover:border-glow-violet/40',
    'hover:shadow-[0_0_24px_-4px_oklch(65%_0.30_320_/_0.4),_0_0_60px_-12px_oklch(65%_0.30_320_/_0.2)]',
  ].join(' '),
  amber: [
    'transition-[border-color,box-shadow] duration-200 ease-arc',
    'hover:border-glow-amber/40',
    'hover:shadow-[0_0_24px_-4px_oklch(80%_0.18_80_/_0.4),_0_0_60px_-12px_oklch(80%_0.18_80_/_0.2)]',
  ].join(' '),
}

export function Card({ className, children, glow = 'none', glass = false, ...props }: CardProps) {
  return (
    <div
      className={cn(
        'rounded-lg border p-4',
        // Base: glass morphism or solid
        glass
          ? 'border-glass-border bg-surface-raised/60 backdrop-blur-xl shadow-card'
          : 'border-surface-border bg-surface-raised shadow-card',
        GLOW_CLASSES[glow],
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}
