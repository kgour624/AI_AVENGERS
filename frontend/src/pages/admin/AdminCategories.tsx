import { useState, type FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  getExpertCategories,
  createExpertCategory,
  updateExpertCategory,
} from '@/api/admin'
import type { ExpertCategory } from '@/types/category'
import { CATEGORY_SECTION_TYPES, TEST_CASE_BUCKETS } from '@/types/category'
import type { CategorySectionType } from '@/types/category'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { handleAPIError } from '@/utils/errors'

/**
 * CT-D1: Admin Categories page + template builder.
 * Source: CATEGORY_TEMPLATE_HANDOFF.md §7 ("Admin: new Categories page
 * | CRUD for expert_categories + template builder (add/remove/reorder
 * sections, set label/type/required per section, set default_language,
 * toggle ask_structure_permission)").
 *
 * Follows AdminClients.tsx's exact pattern (useQuery list + Card grid +
 * useMutation for writes) rather than inventing a new admin-page shape.
 *
 * A single form area doubles as both "create new" (no id selected) and
 * "edit existing" (id selected) — same pattern CreateExpertModal uses
 * for its own create-only form, but inlined on the page (no modal)
 * since a template builder with a variable number of sections needs
 * more room than a modal comfortably gives.
 */
interface SectionDraft {
  key: string
  label: string
  type: CategorySectionType
  required: boolean
}

function categoryToDraft(cat: ExpertCategory | null) {
  return {
    name: cat?.name ?? '',
    slug: cat?.slug ?? '',
    description: cat?.description ?? '',
    defaultLanguage: cat?.defaultLanguage ?? 'java',
    askStructurePermission: cat?.askStructurePermission ?? false,
    sections: (cat?.templateSchema.sections ?? []).map((s) => ({ ...s })) as SectionDraft[],
  }
}

function slugify(name: string): string {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
}

function AdminCategories() {
  const queryClient = useQueryClient()
  const { data: categories, isLoading } = useQuery({
    queryKey: ['admin', 'expert-categories'],
    queryFn: getExpertCategories,
  })

  const [editingId, setEditingId] = useState<string | null>(null)
  const [draft, setDraft] = useState(categoryToDraft(null))
  const [slugEdited, setSlugEdited] = useState(false)
  const [error, setError] = useState('')

  const createMutation = useMutation({
    mutationFn: createExpertCategory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'expert-categories'] })
      resetForm()
    },
    onError: (err) => setError(handleAPIError(err)),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, ...req }: { id: string } & Parameters<typeof updateExpertCategory>[1]) =>
      updateExpertCategory(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'expert-categories'] })
      resetForm()
    },
    onError: (err) => setError(handleAPIError(err)),
  })

  function resetForm() {
    setEditingId(null)
    setDraft(categoryToDraft(null))
    setSlugEdited(false)
    setError('')
  }

  function startEdit(cat: ExpertCategory) {
    setEditingId(cat.id)
    setDraft(categoryToDraft(cat))
    setSlugEdited(true)
    setError('')
  }

  function addSection() {
    setDraft((d) => ({
      ...d,
      sections: [...d.sections, { key: '', label: '', type: 'prose', required: true }],
    }))
  }

  function removeSection(index: number) {
    setDraft((d) => ({ ...d, sections: d.sections.filter((_, i) => i !== index) }))
  }

  function moveSection(index: number, direction: -1 | 1) {
    setDraft((d) => {
      const target = index + direction
      if (target < 0 || target >= d.sections.length) return d
      const next = [...d.sections]
      ;[next[index], next[target]] = [next[target]!, next[index]!]
      return { ...d, sections: next }
    })
  }

  function updateSection(index: number, patch: Partial<SectionDraft>) {
    setDraft((d) => ({
      ...d,
      sections: d.sections.map((s, i) => (i === index ? { ...s, ...patch } : s)),
    }))
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')

    if (!draft.name.trim() || !draft.slug.trim()) {
      setError('Name and slug are required')
      return
    }
    // WHY validate keys client-side too (backend already validates via
    // validateTemplateSchema): fail fast with a specific message before
    // a round trip, matching CreateExpertModal's existing numeric-field
    // validation pattern above the API call.
    const keys = draft.sections.map((s) => s.key.trim())
    if (keys.some((k) => k === '')) {
      setError('Every section needs a non-empty key')
      return
    }
    if (new Set(keys).size !== keys.length) {
      setError('Section keys must be unique')
      return
    }

    const payload = {
      name: draft.name.trim(),
      slug: draft.slug.trim(),
      description: draft.description.trim() || undefined,
      templateSchema: { sections: draft.sections },
      defaultLanguage: draft.defaultLanguage.trim() || undefined,
      askStructurePermission: draft.askStructurePermission,
    }

    if (editingId) {
      updateMutation.mutate({ id: editingId, ...payload })
    } else {
      createMutation.mutate(payload)
    }
  }

  const isSubmitting = createMutation.isPending || updateMutation.isPending

  if (isLoading) {
    return (
      <div className="space-y-3 p-6">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-16" />
        ))}
      </div>
    )
  }

  return (
    <div className="p-6">
      <h1 className="mb-4 text-xl font-semibold">Expert Categories</h1>
      <p className="mb-4 text-xs text-text-secondary">
        Category is a layer above an expert&apos;s domain — it owns the structured JSON
        answer template (Pattern/Idea/Code/Walkthrough/Test Cases, etc.) that applies to
        every expert placed in it. An expert with no category keeps flat-text answers.
      </p>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Existing categories list */}
        <div className="space-y-2">
          {categories?.map((cat) => (
            <Card key={cat.id} glow="purple" className="flex items-center justify-between">
              <div>
                <p className="font-medium text-text-primary">{cat.name}</p>
                <p className="text-xs text-text-secondary">/{cat.slug}</p>
                <p className="mt-1 text-xs text-text-disabled">
                  {cat.templateSchema.sections.length} section
                  {cat.templateSchema.sections.length === 1 ? '' : 's'}
                  {' \u00b7 '}lang: {cat.defaultLanguage}
                  {cat.askStructurePermission && (
                    <>
                      {' \u00b7 '}
                      <Badge variant="brand">asks structure first</Badge>
                    </>
                  )}
                </p>
              </div>
              <Button variant="secondary" size="sm" onClick={() => startEdit(cat)}>
                Edit
              </Button>
            </Card>
          ))}
          {categories?.length === 0 && (
            <p className="text-sm text-text-disabled">
              No categories yet — create one on the right.
            </p>
          )}
        </div>

        {/* Create / edit form */}
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <h2 className="text-sm font-medium text-text-primary">
            {editingId ? 'Edit Category' : 'New Category'}
          </h2>

          <Input
            label="Name"
            value={draft.name}
            onChange={(e) => {
              const value = e.target.value
              setDraft((d) => ({
                ...d,
                name: value,
                slug: slugEdited ? d.slug : slugify(value),
              }))
            }}
            placeholder="Coding"
          />
          <Input
            label="Slug"
            value={draft.slug}
            onChange={(e) => {
              setSlugEdited(true)
              setDraft((d) => ({ ...d, slug: e.target.value }))
            }}
            placeholder="coding"
          />
          <div className="flex flex-col gap-1.5">
            <label htmlFor="category-description" className="text-sm text-text-secondary">
              Description (optional)
            </label>
            <textarea
              id="category-description"
              value={draft.description}
              onChange={(e) => setDraft((d) => ({ ...d, description: e.target.value }))}
              rows={2}
              className="rounded-md border border-surface-border bg-surface-overlay px-3 py-2 text-sm text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-brand"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <Input
              label="Default Language"
              value={draft.defaultLanguage}
              onChange={(e) => setDraft((d) => ({ ...d, defaultLanguage: e.target.value }))}
              placeholder="java"
            />
            <div className="flex flex-col gap-1.5">
              <label className="text-sm text-text-secondary">Ask structure first</label>
              <label className="flex items-center gap-2 text-sm text-text-primary">
                <input
                  type="checkbox"
                  checked={draft.askStructurePermission}
                  onChange={(e) =>
                    setDraft((d) => ({ ...d, askStructurePermission: e.target.checked }))
                  }
                />
                Ask before generating (CT-L9)
              </label>
            </div>
          </div>

          {/* Template builder */}
          <div className="rounded-md border border-surface-border p-3">
            <div className="mb-2 flex items-center justify-between">
              <span className="text-sm font-medium text-text-primary">Template Sections</span>
              <Button type="button" variant="secondary" size="sm" onClick={addSection}>
                + Add Section
              </Button>
            </div>

            {draft.sections.length === 0 && (
              <p className="text-xs text-text-disabled">
                No sections — experts in this category will use flat-text answers (CT-L2).
              </p>
            )}

            <div className="flex flex-col gap-3">
              {draft.sections.map((section, i) => (
                <div key={i} className="rounded-md border border-surface-border/60 p-2">
                  <div className="grid grid-cols-2 gap-2">
                    <input
                      value={section.key}
                      onChange={(e) => updateSection(i, { key: e.target.value })}
                      placeholder="key (e.g. pattern)"
                      className="rounded-md border border-surface-border bg-surface-overlay px-2 py-1 text-xs text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-brand"
                    />
                    <input
                      value={section.label}
                      onChange={(e) => updateSection(i, { label: e.target.value })}
                      placeholder="label (e.g. Pattern)"
                      className="rounded-md border border-surface-border bg-surface-overlay px-2 py-1 text-xs text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-brand"
                    />
                  </div>
                  <div className="mt-2 flex items-center gap-2">
                    <select
                      value={section.type}
                      onChange={(e) =>
                        updateSection(i, { type: e.target.value as CategorySectionType })
                      }
                      className="flex-1 rounded-md border border-surface-border bg-surface-overlay px-2 py-1 text-xs text-text-primary focus:outline-none focus:ring-2 focus:ring-brand"
                    >
                      {CATEGORY_SECTION_TYPES.map((t) => (
                        <option key={t.value} value={t.value}>
                          {t.label}
                        </option>
                      ))}
                    </select>
                    <label className="flex items-center gap-1 text-xs text-text-secondary">
                      <input
                        type="checkbox"
                        checked={section.required}
                        onChange={(e) => updateSection(i, { required: e.target.checked })}
                      />
                      required
                    </label>
                    <button
                      type="button"
                      onClick={() => moveSection(i, -1)}
                      disabled={i === 0}
                      aria-label="Move up"
                      className="text-xs text-text-secondary hover:text-text-primary disabled:opacity-30"
                    >
                      {'\u25b2'}
                    </button>
                    <button
                      type="button"
                      onClick={() => moveSection(i, 1)}
                      disabled={i === draft.sections.length - 1}
                      aria-label="Move down"
                      className="text-xs text-text-secondary hover:text-text-primary disabled:opacity-30"
                    >
                      {'\u25bc'}
                    </button>
                    <button
                      type="button"
                      onClick={() => removeSection(i)}
                      className="text-xs text-mode-refuse hover:underline"
                    >
                      remove
                    </button>
                  </div>
                  {section.type === 'test_cases' && (
                    <p className="mt-1 text-[11px] text-text-disabled">
                      Renders fixed buckets: {TEST_CASE_BUCKETS.join(', ')} (CT-L4, not editable)
                    </p>
                  )}
                </div>
              ))}
            </div>
          </div>

          {error && <p className="text-xs text-mode-refuse">{error}</p>}
          <div className="flex justify-end gap-2">
            {editingId && (
              <Button type="button" variant="ghost" onClick={resetForm}>
                Cancel
              </Button>
            )}
            <Button type="submit" isLoading={isSubmitting}>
              {editingId ? 'Save Category' : 'Create Category'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

export const Component = AdminCategories
