import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getAdminSettings, updateAdminSetting, type SystemSetting } from '@/api/admin'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { handleAPIError } from '@/utils/errors'

/**
 * Bug 1.6 fix (docs bug list): this page was a static placeholder.
 * Per its own (now removed) header comment and HANDOFF.md, that WAS a
 * deliberate decision, not an oversight - GET/PATCH /admin/settings
 * operate on an arbitrary JSONB `value` per key, and the 4 known keys
 * (china_wall, context, models, cost_budget - seed data in
 * AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 5) each have a
 * COMPLETELY DIFFERENT shape inside that blob. A generic "edit raw
 * JSON" textarea would let a typo silently break China Wall
 * enforcement in production with zero validation.
 *
 * Fix: 4 bespoke, typed forms - one per known key - exactly what
 * HANDOFF.md itself recommended instead of a generic editor.
 */

interface ChinaWallValue {
  rerankerThreshold: number
  maxRetries: number
  stripUncited: boolean
}
interface ContextValue {
  maxTokens: number
  recentMessages: number
  semanticTopK: number
  courseChunks: number
}
interface ModelsValue {
  cheap: string
  strong: string
  fast: string
}
interface CostBudgetValue {
  monthlyLimitUsd: number
  alertThreshold: number
}

function useSettingSave(key: string) {
  const queryClient = useQueryClient()
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState('')
  const [saved, setSaved] = useState(false)

  const save = async (value: Record<string, unknown>) => {
    setIsSaving(true)
    setError('')
    setSaved(false)
    try {
      await updateAdminSetting(key, value)
      await queryClient.invalidateQueries({ queryKey: ['admin', 'settings'] })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch (err) {
      setError(handleAPIError(err))
    } finally {
      setIsSaving(false)
    }
  }

  return { save, isSaving, error, saved }
}

function SaveRow({
  isSaving,
  error,
  saved,
  onSave,
}: {
  isSaving: boolean
  error: string
  saved: boolean
  onSave: () => void
}) {
  return (
    <>
      {error && <p className="mt-2 text-xs text-mode-refuse">{error}</p>}
      <div className="mt-3 flex items-center gap-2">
        <Button size="sm" isLoading={isSaving} onClick={onSave}>
          Save
        </Button>
        {saved && <span className="text-xs text-mode-advise">Saved</span>}
      </div>
    </>
  )
}

function ChinaWallForm({ value }: { value: ChinaWallValue }) {
  const [form, setForm] = useState(value)
  const { save, isSaving, error, saved } = useSettingSave('china_wall')

  return (
    <Card>
      <h2 className="mb-1 font-medium text-text-primary">China Wall</h2>
      <p className="mb-3 text-xs text-text-secondary">
        Reranker threshold + citation enforcement. A wrong value here directly changes when
        experts refuse to answer or strip claims.
      </p>
      <div className="grid grid-cols-2 gap-3">
        <Input
          label="Reranker threshold (0-1)"
          type="number"
          step="0.01"
          min={0}
          max={1}
          value={form.rerankerThreshold}
          onChange={(e) => setForm({ ...form, rerankerThreshold: Number(e.target.value) })}
        />
        <Input
          label="Max retries"
          type="number"
          min={0}
          value={form.maxRetries}
          onChange={(e) => setForm({ ...form, maxRetries: Number(e.target.value) })}
        />
      </div>
      <label className="mt-3 flex items-center gap-2 text-sm text-text-secondary">
        <input
          type="checkbox"
          checked={form.stripUncited}
          onChange={(e) => setForm({ ...form, stripUncited: e.target.checked })}
        />
        Strip uncited claims
      </label>
      <SaveRow
        isSaving={isSaving}
        error={error}
        saved={saved}
        onSave={() => save(form as unknown as Record<string, unknown>)}
      />
    </Card>
  )
}

function ContextForm({ value }: { value: ContextValue }) {
  const [form, setForm] = useState(value)
  const { save, isSaving, error, saved } = useSettingSave('context')

  return (
    <Card>
      <h2 className="mb-1 font-medium text-text-primary">Context Budget</h2>
      <p className="mb-3 text-xs text-text-secondary">
        Token budget allocation for the context assembler. Changes apply to every new turn.
      </p>
      <div className="grid grid-cols-2 gap-3">
        <Input
          label="Max tokens"
          type="number"
          min={0}
          value={form.maxTokens}
          onChange={(e) => setForm({ ...form, maxTokens: Number(e.target.value) })}
        />
        <Input
          label="Recent messages"
          type="number"
          min={0}
          value={form.recentMessages}
          onChange={(e) => setForm({ ...form, recentMessages: Number(e.target.value) })}
        />
        <Input
          label="Semantic top K"
          type="number"
          min={0}
          value={form.semanticTopK}
          onChange={(e) => setForm({ ...form, semanticTopK: Number(e.target.value) })}
        />
        <Input
          label="Course chunks"
          type="number"
          min={0}
          value={form.courseChunks}
          onChange={(e) => setForm({ ...form, courseChunks: Number(e.target.value) })}
        />
      </div>
      <SaveRow
        isSaving={isSaving}
        error={error}
        saved={saved}
        onSave={() => save(form as unknown as Record<string, unknown>)}
      />
    </Card>
  )
}

function ModelsForm({ value }: { value: ModelsValue }) {
  const [form, setForm] = useState(value)
  const { save, isSaving, error, saved } = useSettingSave('models')

  return (
    <Card>
      <h2 className="mb-1 font-medium text-text-primary">Model Routing</h2>
      <p className="mb-3 text-xs text-text-secondary">
        OpenRouter model identifiers per tier. Typos here fail silently at call time, not here -
        double check the exact OpenRouter slug before saving.
      </p>
      <div className="flex flex-col gap-3">
        <Input
          label="Cheap (tagging, metadata)"
          value={form.cheap}
          onChange={(e) => setForm({ ...form, cheap: e.target.value })}
        />
        <Input
          label="Strong (answer generation)"
          value={form.strong}
          onChange={(e) => setForm({ ...form, strong: e.target.value })}
        />
        <Input
          label="Fast (quick checks)"
          value={form.fast}
          onChange={(e) => setForm({ ...form, fast: e.target.value })}
        />
      </div>
      <SaveRow
        isSaving={isSaving}
        error={error}
        saved={saved}
        onSave={() => save(form as unknown as Record<string, unknown>)}
      />
    </Card>
  )
}

function CostBudgetForm({ value }: { value: CostBudgetValue }) {
  const [form, setForm] = useState(value)
  const { save, isSaving, error, saved } = useSettingSave('cost_budget')

  return (
    <Card>
      <h2 className="mb-1 font-medium text-text-primary">Cost Budget</h2>
      <p className="mb-3 text-xs text-text-secondary">
        Monitoring/alerting only - does not hard-stop LLM calls when exceeded.
      </p>
      <div className="grid grid-cols-2 gap-3">
        <Input
          label="Monthly limit (USD)"
          type="number"
          min={0}
          value={form.monthlyLimitUsd}
          onChange={(e) => setForm({ ...form, monthlyLimitUsd: Number(e.target.value) })}
        />
        <Input
          label="Alert threshold (0-1)"
          type="number"
          step="0.01"
          min={0}
          max={1}
          value={form.alertThreshold}
          onChange={(e) => setForm({ ...form, alertThreshold: Number(e.target.value) })}
        />
      </div>
      <SaveRow
        isSaving={isSaving}
        error={error}
        saved={saved}
        onSave={() => save(form as unknown as Record<string, unknown>)}
      />
    </Card>
  )
}

function getValue<T>(settings: SystemSetting[] | undefined, key: string): T | undefined {
  return settings?.find((s) => s.key === key)?.value as T | undefined
}

function AdminSettings() {
  const { data: settings, isLoading } = useQuery({
    queryKey: ['admin', 'settings'],
    queryFn: getAdminSettings,
  })

  const chinaWall = getValue<ChinaWallValue>(settings, 'china_wall')
  const context = getValue<ContextValue>(settings, 'context')
  const models = getValue<ModelsValue>(settings, 'models')
  const costBudget = getValue<CostBudgetValue>(settings, 'cost_budget')

  return (
    <div className="p-6">
      <h1 className="mb-4 text-xl font-semibold">Settings</h1>

      {isLoading && (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-40" />
          ))}
        </div>
      )}

      {!isLoading && (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          {chinaWall && <ChinaWallForm value={chinaWall} />}
          {context && <ContextForm value={context} />}
          {models && <ModelsForm value={models} />}
          {costBudget && <CostBudgetForm value={costBudget} />}
          {!chinaWall && !context && !models && !costBudget && (
            <p className="text-sm text-text-secondary">
              No known settings keys found in the database (expected: china_wall, context, models,
              cost_budget - see AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 5's seed INSERT).
            </p>
          )}
        </div>
      )}
    </div>
  )
}

export const Component = AdminSettings
