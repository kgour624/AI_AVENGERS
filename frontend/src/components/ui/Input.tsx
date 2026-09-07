import { forwardRef, type InputHTMLAttributes } from 'react'
import { cn } from '@/utils/cn'

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  error?: string
  label?: string
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, error, label, id, ...props }, ref) => {
    const inputId = id ?? props.name
    const errorId = inputId ? `${inputId}-error` : undefined

    return (
      <div className="flex flex-col gap-1.5">
        {label && (
          <label
            htmlFor={inputId}
            className="text-xs font-medium uppercase tracking-wider text-text-secondary"
          >
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={inputId}
          aria-invalid={error ? true : undefined}
          aria-describedby={error ? errorId : undefined}
          className={cn(
            // Base
            'w-full rounded-md px-3 py-2 text-sm text-text-primary',
            // Background — glass
            'bg-surface-overlay/60 backdrop-blur-sm',
            // Border — subtle default, neon on focus
            'border border-surface-border',
            'transition-[border-color,box-shadow,background-color] duration-150 ease-arc',
            // Focus — electric violet neon ring
            'focus:outline-none',
            'focus:border-brand/70',
            'focus:bg-surface-overlay/80',
            'focus:shadow-[0_0_0_3px_oklch(68%_0.28_295_/_0.20),_0_0_12px_oklch(68%_0.28_295_/_0.15)]',
            // Placeholder
            'placeholder:text-text-disabled',
            // Error state
            error && [
              'border-mode-refuse/60',
              'focus:border-mode-refuse/80',
              'focus:shadow-[0_0_0_3px_oklch(62%_0.22_25_/_0.20)]',
            ],
            className
          )}
          {...props}
        />
        {error && (
          <p id={errorId} className="text-xs text-mode-refuse">
            {error}
          </p>
        )}
      </div>
    )
  }
)
Input.displayName = 'Input'
