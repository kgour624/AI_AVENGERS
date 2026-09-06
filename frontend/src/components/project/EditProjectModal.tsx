import { useState, type FormEvent } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { updateProject } from '@/api/projects'
import { handleAPIError } from '@/utils/errors'

export interface EditProjectModalProps {
  isOpen: boolean
  onClose: () => void
  projectId: string
  currentName: string
  currentDescription: string
  onSaved: () => void
}

/**
 * Feature #2 fix (docs bug list): no way to rename a project or edit
 * its description existed anywhere in the UI, despite PATCH
 * /projects/:id being fully functional on the backend.
 */
export function EditProjectModal({
  isOpen,
  onClose,
  projectId,
  currentName,
  currentDescription,
  onSaved,
}: EditProjectModalProps) {
  const [name, setName] = useState(currentName)
  const [description, setDescription] = useState(currentDescription)
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!name.trim()) {
      setError('Project name is required')
      return
    }
    setIsSubmitting(true)
    setError('')
    try {
      await updateProject(projectId, { name: name.trim(), description: description.trim() })
      onSaved()
      onClose()
    } catch (err) {
      setError(handleAPIError(err))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <h2 className="text-lg font-semibold text-text-primary">Edit Project</h2>
        <Input label="Name" value={name} onChange={(e) => setName(e.target.value)} maxLength={500} />
        <div className="flex flex-col gap-1.5">
          <label htmlFor="edit-project-description" className="text-sm text-text-secondary">
            Description
          </label>
          <textarea
            id="edit-project-description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
            className="rounded-md border border-surface-border bg-surface-overlay px-3 py-2 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-brand"
          />
        </div>
        {error && <p className="text-xs text-mode-refuse">{error}</p>}
        <div className="flex justify-end gap-2">
          <Button type="button" variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" isLoading={isSubmitting}>
            Save
          </Button>
        </div>
      </form>
    </Modal>
  )
}
