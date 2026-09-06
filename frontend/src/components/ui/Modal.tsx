import { useEffect } from 'react'
import { createPortal } from 'react-dom'
import { AnimatePresence, motion, useReducedMotion } from 'framer-motion'
import { cn } from '@/utils/cn'
import { modalSpring } from '@/design-system/motion'

/**
 * ARC-51 §3 (docs/ARC51_UI_CONTRACT.md): internals replaced with
 * framer-motion's AnimatePresence + spring transition (modalSpring,
 * design-system/motion.ts), replacing the previous CSS-keyframe
 * isClosing/previousIsOpen state machine - AnimatePresence keeps a
 * component mounted through its exit animation automatically, which
 * is exactly what that manual ref+state dance existed to work around
 * before framer-motion was added to this project.
 *
 * WHY the public prop contract is unchanged (isOpen, onClose,
 * children, className): every existing caller (CreateProjectModal,
 * CreateExpertModal, EditCharterModal, EditProjectModal,
 * RepoConnectModal, TranscriptUploadModal, ExpertTopicsModal, and any
 * future modal) needed zero changes for this rewrite - verified by
 * reading each call site before starting, per this project's
 * Interface-First discipline for UI changes (docs/ARC51_UI_CONTRACT.md §1).
 */
export interface ModalProps {
  isOpen: boolean
  onClose: () => void
  children: React.ReactNode
  className?: string
}

export function Modal({ isOpen, onClose, children, className }: ModalProps) {
  const reduceMotion = useReducedMotion()

  // Escape key closes the modal - matches the wireframe's implied
  // keyboard-first interaction model (section 1: "keyboard-first").
  // Unchanged from the pre-ARC-51 implementation.
  useEffect(() => {
    if (!isOpen) return
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, onClose])

  const modalContainer = document.getElementById('modal-container')
  if (!modalContainer) {
    // WHY throw rather than silently render nothing: a missing portal
    // target is a setup bug (index.html must have this div), not a
    // runtime condition to degrade gracefully from - failing loudly
    // here is more useful than a modal that silently never appears.
    throw new Error('#modal-container not found in index.html')
  }

  // Reduced-motion contract (ARC51_UI_CONTRACT.md §3): fall back to an
  // instant opacity-only transition instead of the spring, rather than
  // skipping AnimatePresence entirely - it still needs to control
  // mount/unmount timing even when the motion itself is disabled.
  const variants = reduceMotion
    ? { hidden: { opacity: 0 }, visible: { opacity: 1 }, exit: { opacity: 0 } }
    : modalSpring

  return createPortal(
    <AnimatePresence>
      {isOpen && (
        <motion.div
          key="modal-backdrop"
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
          onClick={onClose}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: reduceMotion ? 0 : 0.15 }}
        >
          <motion.div
            onClick={(e) => e.stopPropagation()}
            variants={variants}
            initial="hidden"
            animate="visible"
            exit="exit"
            className={cn(
              'max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-lg border border-glass-border bg-surface-overlay/95 p-6 shadow-md backdrop-blur-xl',
              className
            )}
          >
            {children}
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>,
    modalContainer
  )
}
