import { useState, type FormEvent } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { createProject } from '@/api/projects'
import { handleAPIError } from '@/utils/errors'

export interface CreateProjectModalProps {
  isOpen: boolean
  onClose: () => void
  onCreated: (projectId: string) => void
}

/**
 * Bug 1.1 fix (docs bug list): createProject already existed in
 * api/projects.ts, and every UI primitive this needs (Modal, Input,
 * Button) already existed too - nothing was ever wired together. A
 * first-time user saw "No projects yet" with no button, form, or
 * modal anywhere in the UI to create one - a genuine dead end.
 */
export function CreateProjectModal({ isOpen, onClose, onCreated }: CreateProjectModalProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleClose = () => {
    setName('')
    setDescription('')
    setError('')
    onClose()
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!name.trim()) {
      setError('Project name is required')
      return
    }
    setIsSubmitting(true)
    setError('')
    try {
      const project = await createProject({
        name: name.trim(),
        description: description.trim() || undefined,
      })
      onCreated(project.id)
      handleClose()
    } catch (err) {
      setError(handleAPIError(err))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={handleClose}>
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <h2 className="text-lg font-semibold text-text-primary">New Project</h2>
        <Input
          label="Name"
          name="name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="E-Commerce App"
          autoFocus
          maxLength={500}
        />
        <div className="flex flex-col gap-1.5">
          <label htmlFor="description" className="text-sm text-text-secondary">
            Description (optional)
          </label>
          <textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
            className="rounded-md border border-surface-border bg-surface-overlay px-3 py-2 text-sm text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-brand"
          />
        </div>
        {error && <p className="text-xs text-mode-refuse">{error}</p>}
        <div className="flex justify-end gap-2">
          <Button type="button" variant="ghost" onClick={handleClose}>
            Cancel
          </Button>
          <Button type="submit" isLoading={isSubmitting}>
            Create Project
          </Button>
        </div>
      </form>
    </Modal>
  )
}
