import { forwardRef, type InputHTMLAttributes } from 'react'
import { cn } from '@/utils/cn'

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  error?: string
  label?: string
}

// WHY error is a prop on Input itself rather than a separate wrapper
// component: LoginPage/RegisterPage need per-field validation errors
// (e.g. "Invalid email") directly under the field. Wiring aria-invalid
// and aria-describedby here means every consumer gets accessible error
// association for free instead of re-implementing it per form.
export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, error, label, id, ...props }, ref) => {
    const inputId = id ?? props.name
    const errorId = inputId ? `${inputId}-error` : undefined

    return (
      <div className="flex flex-col gap-1.5">
        {label && (
          <label htmlFor={inputId} className="text-sm text-text-secondary">
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={inputId}
          aria-invalid={error ? true : undefined}
          aria-describedby={error ? errorId : undefined}
          className={cn(
            'rounded-md border border-surface-border bg-surface-overlay px-3 py-2 text-sm text-text-primary placeholder:text-text-disabled',
            'focus:outline-none focus:ring-2 focus:ring-brand',
            error && 'border-mode-refuse focus:ring-mode-refuse',
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
