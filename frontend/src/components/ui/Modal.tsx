import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { cn } from '@/utils/cn'

/**
 * Modal with a proper closing animation, using the .modal-closing CSS
 * class defined in design-system/animations.css.
 *
 * WHY the isClosing/previousIsOpen dance (not just `if (!isOpen) return
 * null`): traced through what a naive implementation would do - the
 * moment isOpen flips false, the component unmounts immediately and
 * the .modal-out animation never gets a chance to play (there's
 * nothing left in the DOM to animate). This pattern - track whether
 * we just transitioned from open->closed via a ref, set isClosing,
 * wait for the CSS animation's onAnimationEnd, THEN actually unmount -
 * is the same pattern taught in Transcripts/Frontend/TyeScript
 * Simplified.md's calendar-modal project (use ref for "previous open
 * state", useLayoutEffect so the closing class is applied before
 * paint, only return null once isOpen AND isClosing are both false).
 */
export interface ModalProps {
  isOpen: boolean
  onClose: () => void
  children: React.ReactNode
  className?: string
}

export function Modal({ isOpen, onClose, children, className }: ModalProps) {
  const [isClosing, setIsClosing] = useState(false)
  const previousIsOpen = useRef(isOpen)

  useEffect(() => {
    if (!isOpen && previousIsOpen.current) {
      setIsClosing(true)
    }
    previousIsOpen.current = isOpen
  }, [isOpen])

  // Escape key closes the modal - matches the wireframe's implied
  // keyboard-first interaction model (section 1: "keyboard-first").
  useEffect(() => {
    if (!isOpen) return
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, onClose])

  if (!isOpen && !isClosing) return null

  const modalContainer = document.getElementById('modal-container')
  if (!modalContainer) {
    // WHY throw rather than silently render nothing: a missing portal
    // target is a setup bug (index.html must have this div), not a
    // runtime condition to degrade gracefully from - failing loudly
    // here is more useful than a modal that silently never appears.
    throw new Error('#modal-container not found in index.html')
  }

  return createPortal(
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={onClose}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        onAnimationEnd={() => {
          if (isClosing) setIsClosing(false)
        }}
        className={cn(
          'max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-lg border border-surface-border bg-surface-overlay p-6 shadow-md',
          isClosing ? 'modal-closing' : 'modal-enter',
          className
        )}
      >
        {children}
      </div>
    </div>,
    modalContainer
  )
}
