import { useState, type FormEvent } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { createExpert } from '@/api/admin'
import { handleAPIError } from '@/utils/errors'

export interface CreateExpertModalProps {
  isOpen: boolean
  onClose: () => void
  onCreated: () => void
}

function slugify(name: string): string {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
}

/**
 * Bug 1.5 fix (docs bug list): backend route + (now-added) API
 * function existed, but AdminExperts.tsx had no way to create a new
 * expert at all - only "Upload Transcript" for experts that already
 * exist. Slug auto-fills from name (editable) since CreateExpert's
 * real handler requires it and rejects duplicates with 409.
 */
export function CreateExpertModal({ isOpen, onClose, onCreated }: CreateExpertModalProps) {
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [slugEdited, setSlugEdited] = useState(false)
  const [domain, setDomain] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleClose = () => {
    setName('')
    setSlug('')
    setSlugEdited(false)
    setDomain('')
    setDescription('')
    setError('')
    onClose()
  }

  const handleNameChange = (value: string) => {
    setName(value)
    if (!slugEdited) setSlug(slugify(value))
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!name.trim() || !slug.trim() || !domain.trim()) {
      setError('Name, slug, and domain are all required')
      return
    }
    setIsSubmitting(true)
    setError('')
    try {
      await createExpert({
        name: name.trim(),
        slug: slug.trim(),
        domain: domain.trim(),
        description: description.trim() || undefined,
      })
      onCreated()
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
        <h2 className="text-lg font-semibold text-text-primary">New Expert</h2>
        <Input
          label="Name"
          value={name}
          onChange={(e) => handleNameChange(e.target.value)}
          placeholder="Arpit Bhiyani \u2014 System Design"
          autoFocus
        />
        <Input
          label="Slug"
          value={slug}
          onChange={(e) => {
            setSlugEdited(true)
            setSlug(e.target.value)
          }}
          placeholder="arpit-system-design"
        />
        <Input
          label="Domain"
          value={domain}
          onChange={(e) => setDomain(e.target.value)}
          placeholder="system_design"
        />
        <div className="flex flex-col gap-1.5">
          <label htmlFor="expert-description" className="text-sm text-text-secondary">
            Description (optional)
          </label>
          <textarea
            id="expert-description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={2}
            className="rounded-md border border-surface-border bg-surface-overlay px-3 py-2 text-sm text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-brand"
          />
        </div>
        {error && <p className="text-xs text-mode-refuse">{error}</p>}
        <div className="flex justify-end gap-2">
          <Button type="button" variant="ghost" onClick={handleClose}>
            Cancel
          </Button>
          <Button type="submit" isLoading={isSubmitting}>
            Create Expert
          </Button>
        </div>
      </form>
    </Modal>
  )
}
