import { useState, useEffect, type FormEvent } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { updateExpertCharter } from '@/api/admin'
import { handleAPIError } from '@/utils/errors'

export interface EditCharterModalProps {
  isOpen: boolean
  onClose: () => void
  expertId: string
  expertName: string
  currentCharter: string
  onSaved: () => void
}

/**
 * Feature #1 fix (docs bug list): "Edit Charter" was a permanently
 * disabled button with title="Charter editing UI - not yet built".
 * The backend endpoint (PATCH /admin/experts/:id) already accepted
 * reasoningCharter - only the UI was missing.
 */
export function EditCharterModal({
  isOpen,
  onClose,
  expertId,
  expertName,
  currentCharter,
  onSaved,
}: EditCharterModalProps) {
  const [charter, setCharter] = useState(currentCharter)
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  // BUG FIX (2026-09-08): useState(currentCharter) only seeds state on
  // this component's initial mount. AdminExperts.tsx renders one
  // EditCharterModal instance and just changes its props (isOpen,
  // currentCharter, etc.) each time "Edit Charter" is clicked for a
  // DIFFERENT expert — React does not re-run useState's initializer on
  // a prop change, so `charter` stayed stuck at whatever the FIRST
  // opened expert's currentCharter was (often "" the very first time,
  // before the reasoningCharter backend fix even landed). This effect
  // re-syncs local state whenever the prop actually changes OR the
  // modal is (re)opened, so switching between experts — or reopening
  // the same expert after an edit — always starts from the real saved
  // value, not stale local state.
  useEffect(() => {
    setCharter(currentCharter)
  }, [currentCharter, isOpen])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    setError('')
    try {
      await updateExpertCharter(expertId, charter)
      onSaved()
      onClose()
    } catch (err) {
      setError(handleAPIError(err))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} className="max-w-2xl">
      <form onSubmit={handleSubmit} className="flex flex-col gap-3">
        <h2 className="text-lg font-semibold text-text-primary">Edit Reasoning Charter</h2>
        <p className="text-xs text-text-secondary">{expertName}</p>
        <p className="text-xs text-text-disabled">
          Every rule should include the WHY (Arpit's principle) - e.g. "Never recommend
          microservices for team &lt; 10 BECAUSE operational overhead exceeds development
          velocity at that scale."
        </p>
        <textarea
          value={charter}
          onChange={(e) => setCharter(e.target.value)}
          rows={14}
          className="rounded-md border border-surface-border bg-surface-overlay px-3 py-2 font-mono text-sm text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-brand"
        />
        {error && <p className="text-xs text-mode-refuse">{error}</p>}
        <div className="flex justify-end gap-2">
          <Button type="button" variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" isLoading={isSubmitting}>
            Save Charter
          </Button>
        </div>
      </form>
    </Modal>
  )
}
