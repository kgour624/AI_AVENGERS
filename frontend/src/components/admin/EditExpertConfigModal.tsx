import { useState, type FormEvent } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { updateExpert } from '@/api/admin'
import { handleAPIError } from '@/utils/errors'
import type { Expert } from '@/types/expert'

export interface EditExpertConfigModalProps {
  isOpen: boolean
  onClose: () => void
  expert: Expert
  onSaved: () => void
}

/**
 * A12: Edit the 6 migration 006 config fields on an existing expert.
 * Also exposes trainingStatus for manual admin override.
 *
 * Pre-fills from expert prop. Calls PATCH /admin/experts/:id.
 * Same field layout as CreateExpertModal's Advanced Config section.
 */
export function EditExpertConfigModal({
  isOpen,
  onClose,
  expert,
  onSaved,
}: EditExpertConfigModalProps) {
  const [modelTier, setModelTier] = useState<'cheap' | 'strong' | 'fast'>(
    expert.modelTier ?? 'strong'
  )
  const [loopPattern, setLoopPattern] = useState<'ota' | 'react' | 'plan_execute'>(
    expert.loopPattern ?? 'react'
  )
  const [temperature, setTemperature] = useState(
    String(expert.temperature ?? 0.3)
  )
  const [topP, setTopP] = useState(
    String(expert.topP ?? 0.5)
  )
  const [maxLoopIterations, setMaxLoopIterations] = useState(
    String(expert.maxLoopIterations ?? 5)
  )
  const [allowedToolsRaw, setAllowedToolsRaw] = useState(
    (expert.allowedTools ?? []).join(', ')
  )
  const [trainingStatus, setTrainingStatus] = useState<
    'draft' | 'ingesting' | 'trained' | 'deprecated'
  >(expert.trainingStatus ?? 'draft')

  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleClose = () => {
    setError('')
    onClose()
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()

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

    const allowedTools = allowedToolsRaw
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean)

    setIsSubmitting(true)
    setError('')
    try {
      await updateExpert(expert.id, {
        modelTier,
        loopPattern,
        temperature: tempNum,
        topP: topPNum,
        maxLoopIterations: maxIter,
        allowedTools,
        trainingStatus,
      })
      onSaved()
      handleClose()
    } catch (err) {
      setError(handleAPIError(err))
    } finally {
      setIsSubmitting(false)
    }
  }

  const selectClass =
    'rounded-md border border-surface-border bg-surface-overlay px-3 py-2 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-brand'
  const inputClass =
    'rounded-md border border-surface-border bg-surface-overlay px-3 py-2 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-brand'

  return (
    <Modal isOpen={isOpen} onClose={handleClose}>
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <h2 className="text-lg font-semibold text-text-primary">
          Edit Config — {expert.name}
        </h2>

        {/* Model Tier */}
        <div className="flex flex-col gap-1.5">
          <label className="text-sm text-text-secondary">
            Model Tier
            <span className="ml-1 text-xs text-text-disabled">
              (cheap=DeepSeek, strong=Claude, fast=Gemini)
            </span>
          </label>
          <select
            value={modelTier}
            onChange={(e) => setModelTier(e.target.value as typeof modelTier)}
            className={selectClass}
          >
            <option value="strong">strong — Claude (answer generation)</option>
            <option value="cheap">cheap — DeepSeek (tagging, metadata)</option>
            <option value="fast">fast — Gemini (quick checks)</option>
          </select>
        </div>

        {/* Loop Pattern */}
        <div className="flex flex-col gap-1.5">
          <label className="text-sm text-text-secondary">
            Loop Pattern
            <span className="ml-1 text-xs text-text-disabled">
              (how this expert reasons)
            </span>
          </label>
          <select
            value={loopPattern}
            onChange={(e) => setLoopPattern(e.target.value as typeof loopPattern)}
            className={selectClass}
          >
            <option value="react">react — Reason+Act (exploratory)</option>
            <option value="ota">ota — Observe-Think-Act (code generation)</option>
            <option value="plan_execute">plan_execute — Plan then Execute</option>
          </select>
        </div>

        {/* Temperature + Top-P */}
        <div className="grid grid-cols-2 gap-3">
          <div className="flex flex-col gap-1.5">
            <label className="text-sm text-text-secondary">
              Temperature
              <span className="ml-1 text-xs text-text-disabled">(0.0–2.0)</span>
            </label>
            <input
              type="number"
              min={0}
              max={2}
              step={0.1}
              value={temperature}
              onChange={(e) => setTemperature(e.target.value)}
              className={inputClass}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-sm text-text-secondary">
              Top-P
              <span className="ml-1 text-xs text-text-disabled">(0.0–1.0)</span>
            </label>
            <input
              type="number"
              min={0}
              max={1}
              step={0.05}
              value={topP}
              onChange={(e) => setTopP(e.target.value)}
              className={inputClass}
            />
          </div>
        </div>

        {/* Max Loop Iterations */}
        <div className="flex flex-col gap-1.5">
          <label className="text-sm text-text-secondary">
            Max Loop Iterations
            <span className="ml-1 text-xs text-text-disabled">(1–50)</span>
          </label>
          <input
            type="number"
            min={1}
            max={50}
            step={1}
            value={maxLoopIterations}
            onChange={(e) => setMaxLoopIterations(e.target.value)}
            className={inputClass}
          />
        </div>

        {/* Allowed Tools */}
        <div className="flex flex-col gap-1.5">
          <label className="text-sm text-text-secondary">
            Allowed Tools
            <span className="ml-1 text-xs text-text-disabled">
              (comma-separated)
            </span>
          </label>
          <input
            type="text"
            value={allowedToolsRaw}
            onChange={(e) => setAllowedToolsRaw(e.target.value)}
            placeholder="PostArtifact, AskExpert, ReadBlackboard"
            className={inputClass}
          />
        </div>

        {/* Training Status — manual override */}
        <div className="flex flex-col gap-1.5">
          <label className="text-sm text-text-secondary">
            Training Status
            <span className="ml-1 text-xs text-text-disabled">
              (auto-set by smoke test, override here if needed)
            </span>
          </label>
          <select
            value={trainingStatus}
            onChange={(e) => setTrainingStatus(e.target.value as typeof trainingStatus)}
            className={selectClass}
          >
            <option value="draft">✏️ draft — no transcripts yet</option>
            <option value="ingesting">⏳ ingesting — pipeline running</option>
            <option value="trained">✅ trained — smoke test passed, publicly visible</option>
            <option value="deprecated">⚠️ deprecated — retired, not visible</option>
          </select>
        </div>

        {error && <p className="text-xs text-mode-refuse">{error}</p>}

        <div className="flex justify-end gap-2">
          <Button type="button" variant="ghost" onClick={handleClose}>
            Cancel
          </Button>
          <Button type="submit" isLoading={isSubmitting}>
            Save Config
          </Button>
        </div>
      </form>
    </Modal>
  )
}
