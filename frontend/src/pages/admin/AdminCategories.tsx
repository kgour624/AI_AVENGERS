import { useState, type FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  getExpertCategories,
  createExpertCategory,
  updateExpertCategory,
  getAdminExperts,
  getCategoryExperts,
  setCategoryExperts,
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

interface TemplateDraft {
  name: string
  sections: SectionDraft[]
}

function categoryToDraft(cat: ExpertCategory | null) {
  const schema = cat?.templateSchema
  const templates = (schema?.templates ?? []).map((t) => ({
    name: t.name,
    sections: (t.sections ?? []).map((s) => ({ ...s })) as SectionDraft[],
  }))
  return {
    name: cat?.name ?? '',
    slug: cat?.slug ?? '',
    description: cat?.description ?? '',
    defaultLanguage: cat?.defaultLanguage ?? 'java',
    askStructurePermission: cat?.askStructurePermission ?? false,
    sections: ((schema?.sections ?? []) as SectionDraft[]).map((s) => ({ ...s })),
    // T-CAT: multiple named formats. Off = the original single template.
    useMultiple: templates.length > 0,
    templates: templates as TemplateDraft[],
    defaultTemplate: schema?.default ?? '',
  }
}

function slugify(name: string): string {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
}

/** One section row editor, shared by the single-template and multi-format editors. */
function SectionRow({
  section,
  index,
  total,
  onPatch,
  onMove,
  onRemove,
}: {
  section: SectionDraft
  index: number
  total: number
  onPatch: (ix: number, patch: Partial<SectionDraft>) => void
  onMove: (ix: number, dir: -1 | 1) => void
  onRemove: (ix: number) => void
}) {
  return (
    <div className="rounded-md border border-surface-border/60 p-2">
      <div className="grid grid-cols-2 gap-2">
        <input
          value={section.key}
          onChange={(e) => onPatch(index, { key: e.target.value })}
          placeholder="key (e.g. pattern)"
          className="rounded-md border border-surface-border bg-surface-overlay px-2 py-1 text-xs text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-brand"
        />
        <input
          value={section.label}
          onChange={(e) => onPatch(index, { label: e.target.value })}
          placeholder="label (e.g. Pattern)"
          className="rounded-md border border-surface-border bg-surface-overlay px-2 py-1 text-xs text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-brand"
        />
      </div>
      <div className="mt-2 flex items-center gap-2">
        <select
          value={section.type}
          onChange={(e) => onPatch(index, { type: e.target.value as CategorySectionType })}
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
            onChange={(e) => onPatch(index, { required: e.target.checked })}
          />
          required
        </label>
        <button
          type="button"
          onClick={() => onMove(index, -1)}
          disabled={index === 0}
          aria-label="Move up"
          className="text-xs text-text-secondary hover:text-text-primary disabled:opacity-30"
        >
          {'\u25b2'}
        </button>
        <button
          type="button"
          onClick={() => onMove(index, 1)}
          disabled={index === total - 1}
          aria-label="Move down"
          className="text-xs text-text-secondary hover:text-text-primary disabled:opacity-30"
        >
          {'\u25bc'}
        </button>
        <button
          type="button"
          onClick={() => onRemove(index)}
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
  )
}

/**
 * T-CAT: assign experts to a category from the category page.
 *
 * WHY this exists (production incident): an expert only answers with a
 * category's format when experts.category_id points at it, but the category
 * page had no way to set that — so the template was built and never used, and
 * every answer silently fell back to flat text. Assignment lived only inside
 * each expert's edit modal, where it was easy to miss.
 */
function CategoryExpertsPicker({ categoryId }: { categoryId: string }) {
  const queryClient = useQueryClient()
  const { data: allExperts } = useQuery({ queryKey: ['admin', 'experts'], queryFn: getAdminExperts })
  const { data: members, isLoading } = useQuery({
    queryKey: ['admin', 'category-experts', categoryId],
    queryFn: () => getCategoryExperts(categoryId),
  })
  const [selected, setSelected] = useState<Set<string> | null>(null)
  const [msg, setMsg] = useState('')

  // null = "not edited yet, show what the server has"; any Set = the user's
  // pending selection. Keeps the checkbox list in sync after a save without an
  // effect that would fight the user's clicks.
  const currentIds = selected ?? new Set((members ?? []).map((m) => m.id))

  const save = useMutation({
    mutationFn: () => setCategoryExperts(categoryId, Array.from(currentIds)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'category-experts', categoryId] })
      queryClient.invalidateQueries({ queryKey: ['admin', 'experts'] })
      setSelected(null)
      setMsg('Saved')
    },
    onError: (err) => setMsg(handleAPIError(err)),
  })

  function toggle(id: string) {
    const next = new Set(currentIds)
    if (next.has(id)) next.delete(id)
    else next.add(id)
    setSelected(next)
    setMsg('')
  }

  return (
    <div className="rounded-md border border-surface-border p-3">
      <p className="mb-1 text-sm font-medium text-text-primary">Experts in this category</p>
      <p className="mb-2 text-[11px] text-text-disabled">
        Only assigned experts answer with this category&apos;s format. Experts left out keep
        flat-text answers.
      </p>
      {isLoading ? (
        <Skeleton className="h-10" />
      ) : (
        <div className="max-h-48 space-y-1 overflow-y-auto">
          {(allExperts ?? []).map((ex) => (
            <label key={ex.id} className="flex items-center gap-2 text-xs text-text-secondary">
              <input type="checkbox" checked={currentIds.has(ex.id)} onChange={() => toggle(ex.id)} />
              {ex.name}
              {ex.categoryId === categoryId && <Badge variant="brand">in category</Badge>}
            </label>
          ))}
          {(allExperts ?? []).length === 0 && (
            <p className="text-xs text-text-disabled">No experts yet.</p>
          )}
        </div>
      )}
      <div className="mt-2 flex items-center gap-2">
        <Button
          type="button"
          size="sm"
          variant="secondary"
          isLoading={save.isPending}
          onClick={() => save.mutate()}
        >
          Save experts
        </Button>
        <span className="text-xs text-text-disabled">{currentIds.size} selected</span>
        {msg && <span className="text-xs text-text-secondary">{msg}</span>}
      </div>
    </div>
  )
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

  // --- T-CAT: named-format (template) helpers -------------------------------
  const patchTemplates = (fn: (t: TemplateDraft[]) => TemplateDraft[]) =>
    setDraft((d) => ({ ...d, templates: fn(d.templates) }))

  function addTemplate() {
    patchTemplates((ts) => [...ts, { name: '', sections: [{ key: '', label: '', type: 'prose', required: true }] }])
  }

  function removeTemplate(ti: number) {
    patchTemplates((ts) => ts.filter((_, i) => i !== ti))
  }

  function updateTemplateName(ti: number, name: string) {
    patchTemplates((ts) => ts.map((t, i) => (i === ti ? { ...t, name } : t)))
  }

  function addTemplateSection(ti: number) {
    patchTemplates((ts) =>
      ts.map((t, i) =>
        i === ti ? { ...t, sections: [...t.sections, { key: '', label: '', type: 'prose', required: true }] } : t,
      ),
    )
  }

  function patchTemplateSection(ti: number, si: number, patch: Partial<SectionDraft>) {
    patchTemplates((ts) =>
      ts.map((t, i) =>
        i === ti ? { ...t, sections: t.sections.map((s, j) => (j === si ? { ...s, ...patch } : s)) } : t,
      ),
    )
  }

  function moveTemplateSection(ti: number, si: number, dir: -1 | 1) {
    patchTemplates((ts) =>
      ts.map((t, i) => {
        if (i !== ti) return t
        const target = si + dir
        if (target < 0 || target >= t.sections.length) return t
        const next = [...t.sections]
        ;[next[si], next[target]] = [next[target]!, next[si]!]
        return { ...t, sections: next }
      }),
    )
  }

  function removeTemplateSection(ti: number, si: number) {
    patchTemplates((ts) =>
      ts.map((t, i) => (i === ti ? { ...t, sections: t.sections.filter((_, j) => j !== si) } : t)),
    )
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
    // Validate whichever shape is active. The same rules apply to every
    // format's sections, so a variant cannot be saved half-configured.
    const validateSections = (sections: SectionDraft[], where: string): string | null => {
      const keys = sections.map((s) => s.key.trim())
      if (keys.some((k) => k === '')) return `Every section needs a non-empty key (${where})`
      if (new Set(keys).size !== keys.length) return `Section keys must be unique (${where})`
      return null
    }

    if (draft.useMultiple) {
      if (draft.templates.length === 0) {
        setError('Add at least one format, or turn off multiple formats')
        return
      }
      const names = draft.templates.map((t) => t.name.trim())
      if (names.some((n) => n === '')) {
        setError('Every format needs a name (e.g. Code, Approach)')
        return
      }
      if (new Set(names.map((n) => n.toLowerCase())).size !== names.length) {
        setError('Format names must be unique')
        return
      }
      for (const [i, t] of draft.templates.entries()) {
        const problem = validateSections(t.sections, `format "${t.name || i + 1}"`)
        if (problem) {
          setError(problem)
          return
        }
      }
    } else {
      const problem = validateSections(draft.sections, 'template')
      if (problem) {
        setError(problem)
        return
      }
    }

    const templateSchema = draft.useMultiple
      ? {
          templates: draft.templates.map((t) => ({
            name: t.name.trim(),
            sections: t.sections.map((s) => ({ ...s, key: s.key.trim() })),
          })),
          default:
            draft.defaultTemplate.trim() || draft.templates[0]?.name.trim() || undefined,
        }
      : { sections: draft.sections.map((s) => ({ ...s, key: s.key.trim() })) }

    const payload = {
      name: draft.name.trim(),
      slug: draft.slug.trim(),
      description: draft.description.trim() || undefined,
      templateSchema,
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
                  {(cat.templateSchema.templates?.length ?? 0) > 0
                    ? `${cat.templateSchema.templates!.length} format${
                        cat.templateSchema.templates!.length === 1 ? '' : 's'
                      }`
                    : `${cat.templateSchema.sections?.length ?? 0} section${
                        (cat.templateSchema.sections?.length ?? 0) === 1 ? '' : 's'
                      }`}
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

          {/* Mode switch (T-CAT): one template, or several named formats. */}
          <label className="flex items-center gap-2 text-sm text-text-primary">
            <input
              type="checkbox"
              checked={draft.useMultiple}
              onChange={(e) => {
                const on = e.target.checked
                setDraft((d) => ({
                  ...d,
                  useMultiple: on,
                  // Seed the first format from the single template so switching
                  // modes never loses the sections the admin already wrote.
                  templates:
                    on && d.templates.length === 0
                      ? [
                          {
                            name: 'Code',
                            sections:
                              d.sections.length > 0
                                ? d.sections
                                : [{ key: '', label: '', type: 'prose', required: true }],
                          },
                        ]
                      : d.templates,
                  defaultTemplate:
                    on && !d.defaultTemplate ? (d.templates[0]?.name ?? 'Code') : d.defaultTemplate,
                }))
              }}
            />
            Multiple answer formats (e.g. Code / Approach) — same expert, per-question shape
          </label>

          {!draft.useMultiple && (
            <>
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

            </>
          )}

          {draft.useMultiple && (
            <div className="rounded-md border border-surface-border p-3">
              <div className="mb-2 flex items-center justify-between">
                <span className="text-sm font-medium text-text-primary">Answer formats</span>
                <Button type="button" variant="secondary" size="sm" onClick={addTemplate}>
                  + Add Format
                </Button>
              </div>
              <p className="mb-2 text-[11px] text-text-disabled">
                The client can pick one per question; the default is used when they don&apos;t.
              </p>
              <div className="flex flex-col gap-4">
                {draft.templates.map((tpl, ti) => (
                  <div key={ti} className="rounded-md border border-surface-border/60 p-2">
                    <div className="flex items-center gap-2">
                      <input
                        value={tpl.name}
                        onChange={(e) => updateTemplateName(ti, e.target.value)}
                        placeholder="format name (e.g. Code)"
                        className="flex-1 rounded-md border border-surface-border bg-surface-overlay px-2 py-1 text-xs text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-brand"
                      />
                      <label className="flex items-center gap-1 text-xs text-text-secondary">
                        <input
                          type="radio"
                          name="default-template"
                          checked={
                            tpl.name.trim() !== '' &&
                            draft.defaultTemplate.trim().toLowerCase() === tpl.name.trim().toLowerCase()
                          }
                          onChange={() => setDraft((d) => ({ ...d, defaultTemplate: tpl.name }))}
                        />
                        default
                      </label>
                      <button
                        type="button"
                        onClick={() => removeTemplate(ti)}
                        className="text-xs text-mode-refuse hover:underline"
                      >
                        remove format
                      </button>
                    </div>
                    <div className="mt-2 flex flex-col gap-3">
                      {tpl.sections.map((s, si) => (
                        <SectionRow
                          key={si}
                          section={s}
                          index={si}
                          total={tpl.sections.length}
                          onPatch={(ix, patch) => patchTemplateSection(ti, ix, patch)}
                          onMove={(ix, dir) => moveTemplateSection(ti, ix, dir)}
                          onRemove={(ix) => removeTemplateSection(ti, ix)}
                        />
                      ))}
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className="mt-2"
                      onClick={() => addTemplateSection(ti)}
                    >
                      + Add Section
                    </Button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {editingId && <CategoryExpertsPicker categoryId={editingId} />}

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
