import { useState, type ReactNode } from 'react'
import { cn } from '@/utils/cn'

/**
 * Simple hover tooltip. Used by CitationChip (Phase 3) to show chunk
 * text on hover - built now as a ui/ primitive since it has no
 * dependency on chat-specific state.
 */
export interface TooltipProps {
  content: string
  children: ReactNode
  className?: string
}

export function Tooltip({ content, children, className }: TooltipProps) {
  const [visible, setVisible] = useState(false)

  return (
    <span
      className="relative inline-block"
      onMouseEnter={() => setVisible(true)}
      onMouseLeave={() => setVisible(false)}
    >
      {children}
      {visible && (
        <span
          role="tooltip"
          className={cn(
            'absolute bottom-full left-1/2 z-50 mb-1 -translate-x-1/2 whitespace-nowrap rounded-md bg-surface-overlay px-2 py-1 text-xs text-text-primary shadow-md',
            className
          )}
        >
          {content}
        </span>
      )}
    </span>
  )
}
