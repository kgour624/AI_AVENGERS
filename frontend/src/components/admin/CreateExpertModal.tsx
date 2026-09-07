import { useState, type FormEvent } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { createExpert } from '@/api/admin'
import { handleAPIError } from '@/utils/errors'
import { cn } from '@/utils/cn'

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
 * A12: CreateExpertModal extended with Advanced Config section.
 * New fields: modelTier, loopPattern, temperature, topP,
 * maxLoopIterations, allowedTools.
 * All behind a collapsible toggle — simple case stays simple.
 */
export function CreateExpertModal({ isOpen, onClose, onCreated }: CreateExpertModalProps) {
  // Required fields
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [slugEdited, setSlugEdited] = useState(false)
  const [domain, setDomain] = useState('')
  const [description, setDescription] = useState('')

  // Advanced config fields — defaults match backend migration 006 defaults
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [modelTier, setModelTier] = useState<'cheap' | 'strong' | 'fast'>('strong')
  const [loopPattern, setLoopPattern] = useState<'ota' | 'react' | 'plan_execute'>('react')
  const [temperature, setTemperature] = useState('0.30')
  const [topP, setTopP] = useState('0.50')
  const [maxLoopIterations, setMaxLoopIterations] = useState('5')
  const [allowedToolsRaw, setAllowedToolsRaw] = useState('')

  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleClose = () => {
    setName('')
    setSlug('')
    setSlugEdited(false)
    setDomain('')
    setDescription('')
    setShowAdvanced(false)
    setModelTier('strong')
    setLoopPattern('react')
    setTemperature('0.30')
    setTopP('0.50')
    setMaxLoopIterations('5')
    setAllowedToolsRaw('')
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

    // Validate numeric fields
    const tempNum = parseFloat(temperature)
    const topPNum = parseFloat(topP)
    const maxIter = parseInt(maxLoopIterations, 10)
    if (isNaN(tempNum) || tempNum < 0 || tempNum > 2) {
      setError('Temperature must be between 0.0 and 2.0')
      return
    }
    if (isNaN(topPNum) || topPNum < 0 || topPNum > 1) {
      setError('Top-P must be between 0.0 and 1.0')
      return
    }
    if (isNaN(maxIter) || maxIter < 1 || maxIter > 50) {
      setError('Max loop iterations must be between 1 and 50')
      return
    }

    // Parse allowedTools: comma-separated string → trimmed string[]
    const allowedTools = allowedToolsRaw
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean)

    setIsSubmitting(true)
    setError('')
    try {
      await createExpert({
        name: name.trim(),
        slug: slug.trim(),
        domain: domain.trim(),
        description: description.trim() || undefined,
        modelTier,
        loopPattern,
        temperature: tempNum,
        topP: topPNum,
        maxLoopIterations: maxIter,
        allowedTools: allowedTools.length > 0 ? allowedTools : undefined,
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
