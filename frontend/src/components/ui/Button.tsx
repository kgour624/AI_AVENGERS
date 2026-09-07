import { forwardRef, type ButtonHTMLAttributes } from 'react'
import { cn } from '@/utils/cn'

export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger'
export type ButtonSize = 'sm' | 'md' | 'lg'

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  size?: ButtonSize
  isLoading?: boolean
}

const variantClasses: Record<ButtonVariant, string> = {
  // Primary — electric violet with neon glow
  primary: [
    'bg-brand text-white',
    'border border-brand/60',
    'shadow-[0_0_12px_oklch(68%_0.28_295_/_0.35)]',
    'hover:bg-brand-hover',
    'hover:shadow-[0_0_20px_oklch(68%_0.28_295_/_0.55),_0_0_40px_oklch(68%_0.28_295_/_0.2)]',
    'hover:border-brand/80',
  ].join(' '),

  // Secondary — glass morphism
  secondary: [
    'bg-surface-overlay/60 text-text-primary',
    'border border-glass-border',
    'backdrop-blur-md',
    'hover:bg-surface-float/70',
    'hover:border-glow-purple/30',
    'hover:shadow-[0_0_12px_oklch(68%_0.28_295_/_0.15)]',
  ].join(' '),

  // Ghost — transparent with subtle hover
  ghost: [
    'bg-transparent text-text-secondary',
    'border border-transparent',
    'hover:text-text-primary',
    'hover:bg-surface-overlay/50',
    'hover:border-glass-border',
  ].join(' '),

  // Danger — red neon
  danger: [
    'bg-mode-refuse/20 text-mode-refuse',
    'border border-mode-refuse/40',
    'hover:bg-mode-refuse/30',
    'hover:border-mode-refuse/70',
    'hover:shadow-[0_0_16px_oklch(62%_0.22_25_/_0.4)]',
  ].join(' '),
}

const sizeClasses: Record<ButtonSize, string> = {
  sm: 'px-3 py-1.5 text-xs gap-1.5',
  md: 'px-4 py-2 text-sm gap-2',
  lg: 'px-6 py-3 text-base gap-2.5',
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', isLoading, disabled, children, ...props }, ref) => {
    return (
      <button
        ref={ref}
        disabled={disabled || isLoading}
        className={cn(
          'inline-flex items-center justify-center rounded-md font-medium',
          'transition-[background-color,border-color,box-shadow,transform,opacity]',
          'duration-150 ease-arc',
          'active:scale-95',
          'disabled:cursor-not-allowed disabled:opacity-40 disabled:active:scale-100',
          'disabled:shadow-none',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand',
          'focus-visible:ring-offset-2 focus-visible:ring-offset-surface-void',
          variantClasses[variant],
          sizeClasses[size],
          className
        )}
        {...props}
      >
        {isLoading && (
          <span
            className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-current border-t-transparent"
            aria-hidden="true"
          />
        )}
        {children}
      </button>
    )
  }
)
Button.displayName = 'Button'
