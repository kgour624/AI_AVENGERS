import { forwardRef, type ButtonHTMLAttributes } from 'react'
import { cn } from '@/utils/cn'

/**
 * WHY forwardRef: react-hook-form-style libraries and focus-management
 * code (e.g. "focus the send button after Cmd+Enter") need a real DOM
 * ref, not just props. Every ui/ primitive in this file set follows
 * the same forwardRef pattern for that reason.
 */
export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger'
export type ButtonSize = 'sm' | 'md' | 'lg'

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  size?: ButtonSize
  isLoading?: boolean
}

const variantClasses: Record<ButtonVariant, string> = {
  primary: 'bg-brand text-white hover:bg-brand-hover',
  secondary: 'bg-surface-overlay text-text-primary hover:bg-surface-border',
  ghost: 'bg-transparent text-text-secondary hover:text-text-primary hover:bg-surface-overlay',
  danger: 'bg-mode-refuse text-white hover:bg-mode-refuse-hover',
}

const sizeClasses: Record<ButtonSize, string> = {
  sm: 'px-2.5 py-1.5 text-xs',
  md: 'px-4 py-2 text-sm',
  lg: 'px-6 py-3 text-base',
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', isLoading, disabled, children, ...props }, ref) => {
    return (
      <button
        ref={ref}
        // WHY also disable while isLoading, not just when `disabled` is
        // explicitly passed: without this, a double-click during an
        // in-flight request (e.g. login submit) would fire the request
        // twice. Cross-questioned this scenario before finalizing.
        disabled={disabled || isLoading}
        className={cn(
          // ARC-51 §5: press/hover polish. transition-[...] lists exact
          // properties (background-color, transform, box-shadow) rather
          // than transition-all, so this never accidentally animates a
          // layout-affecting property. active:scale-95 is the "spring
          // press" feel; disabled buttons don't scale since clicking
          // them does nothing anyway.
          'inline-flex items-center justify-center gap-2 rounded-md font-medium transition-[background-color,transform,box-shadow] duration-150 ease-arc active:scale-95',
          'disabled:cursor-not-allowed disabled:opacity-50 disabled:active:scale-100',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-offset-2 focus-visible:ring-offset-surface-base',
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
