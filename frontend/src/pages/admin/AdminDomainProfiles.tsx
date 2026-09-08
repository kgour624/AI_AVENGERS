import { useState, type FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getDomainProfiles, updateDomainProfile } from '@/api/admin'
import type { DomainProfile } from '@/types/domainProfile'
import { COVERAGE_MODES, CITATION_MODES, STRIP_MODES } from '@/types/domainProfile'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { handleAPIError } from '@/utils/errors'

/**
 * Admin Domain Profiles page — China Wall per-domain config editor.
 * Follows AdminCategories.tsx's exact pattern (useQuery list + form
 * area that doubles as create/edit) rather than inventing a new
 * admin-page shape.
 *
 * Scope (Option B, confirmed with admin before implementing): every
 * DomainProfile field is editable here, not just maxTokensFlat/
 * maxTokensStructured — the admin API itself (admin_handler.go's
 * UpdateDomainProfile) exposes the whole struct, so the UI matches
 * that surface exactly. customRules is intentionally NOT editable
 * here — it is AI-updated based on conversation patterns
 * (domain_profile.go's CustomRules doc comment), not an admin-panel
 * field, and the backend's UpdateDomainProfile request has no field
 * for it either.
 *
 * WHY no "create new domain" flow distinct from "edit": PATCH
 * /admin/domain-profiles/:domain has upsert semantics on the backend
 * (UpdateDomainProfile starts from BaseProfile's defaults when the
 * domain has no stored row yet) — so entering a brand-new domain name
 * in the same form and saving IS how a new profile gets created. No
 * separate POST endpoint exists to duplicate.
 */
function profileToDraft(p: DomainProfile | null) {
  return {
    domain: p?.domain ?? '',
    gate1Skip: p?.gate1Skip ?? false,
    coverageMode: p?.coverageMode ?? 'LITERAL_MATCH',
    citationMode: p?.citationMode ?? 'STRICT',
    stripMode: p?.stripMode ?? 'FULL_STRIP',
    systemPromptExt: p?.systemPromptExt ?? '',
    domainKeywordsText: (p?.domainKeywords ?? []).join(', '),
    maxTokensFlat: p?.maxTokensFlat ?? 0,
    maxTokensStructured: p?.maxTokensStructured ?? 0,
  }
}

function AdminDomainProfiles() {
  const queryClient = useQueryClient()
  const { data: profiles, isLoading } = useQuery({
    queryKey: ['admin', 'domain-profiles'],
    queryFn: getDomainProfiles,
  })

  const [editingDomain, setEditingDomain] = useState<string | null>(null)
  const [draft, setDraft] = useState(profileToDraft(null))
  const [error, setError] = useState('')

  const updateMutation = useMutation({
    mutationFn: ({ domain, ...req }: { domain: string } & Parameters<typeof updateDomainProfile>[1]) =>
      updateDomainProfile(domain, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'domain-profiles'] })
      resetForm()
    },
    onError: (err) => setError(handleAPIError(err)),
  })

  function resetForm() {
    setEditingDomain(null)
    setDraft(profileToDraft(null))
    setError('')
  }

  function startEdit(p: DomainProfile) {
    setEditingDomain(p.domain)
    setDraft(profileToDraft(p))
    setError('')
  }

  function startNew() {
    setEditingDomain('__new__')
    setDraft(profileToDraft(null))
    setError('')
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')

    const domain = draft.domain.trim()
    if (!domain) {
      setError('Domain is required')
      return
    }
    if (draft.maxTokensFlat < 0 || draft.maxTokensStructured < 0) {
      setError('Max tokens must be 0 (use default) or a positive number')
      return
    }

    updateMutation.mutate({
      domain,
      gate1Skip: draft.gate1Skip,
      coverageMode: draft.coverageMode,
      citationMode: draft.citationMode,
      stripMode: draft.stripMode,
      systemPromptExt: draft.systemPromptExt,
      domainKeywords: draft.domainKeywordsText
        .split(',')
        .map((k) => k.trim())
        .filter((k) => k !== ''),
      maxTokensFlat: draft.maxTokensFlat,
      maxTokensStructured: draft.maxTokensStructured,
    })
  }

  const isSubmitting = updateMutation.isPending
  const isNewDomain = editingDomain === '__new__'

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
      <h1 className="mb-4 text-xl font-semibold">Domain Profiles</h1>
      <p className="mb-4 text-xs text-text-secondary">
        Controls how the China Wall (4-layer citation enforcement) behaves per expert
        domain — changes take effect for the very next question, no redeploy needed.
        Includes per-domain LLM response token caps (maxTokensFlat / maxTokensStructured)
        so a domain needing longer answers (prose + code + test cases) can be tuned
        independently of every other domain.
      </p>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Existing profiles list */}
        <div className="space-y-2">
          <div className="mb-2 flex justify-end">
            <Button variant="secondary" size="sm" onClick={startNew}>
              + New Domain Profile
            </Button>
          </div>
          {profiles?.map((p) => (
            <Card key={p.domain} glow="purple" className="flex items-center justify-between">
              <div>
                <p className="font-medium text-text-primary">{p.domain}</p>
                <p className="mt-1 text-xs text-text-disabled">
                  {p.coverageMode} {'\u00b7'} {p.citationMode} {'\u00b7'} {p.stripMode}
                  {p.gate1Skip && (
                    <>
                      {' \u00b7 '}
                      <Badge variant="brand">gate1 skip</Badge>
                    </>
                  )}
                </p>
                <p className="mt-1 text-[11px] text-text-disabled">
                  maxTokens: flat {p.maxTokensFlat || 'default'}, structured{' '}
                  {p.maxTokensStructured || 'default'}
                </p>
              </div>
              <Button variant="secondary" size="sm" onClick={() => startEdit(p)}>
                Edit
              </Button>
            </Card>
          ))}
          {profiles?.length === 0 && (
            <p className="text-sm text-text-disabled">
              No domain profiles yet — create one on the right.
            </p>
          )}
        </div>

        {/* Create / edit form */}
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <h2 className="text-sm font-medium text-text-primary">
            {editingDomain && !isNewDomain ? 'Edit Domain Profile' : 'New Domain Profile'}
          </h2>

          <Input
            label="Domain"
            value={draft.domain}
            onChange={(e) => setDraft((d) => ({ ...d, domain: e.target.value }))}
            placeholder="dsa"
            disabled={!!editingDomain && !isNewDomain}
          />

          <div className="grid grid-cols-3 gap-3">
            <div className="flex flex-col gap-1.5">
              <label className="text-sm text-text-secondary">Coverage Mode</label>
              <select
                value={draft.coverageMode}
                onChange={(e) =>
                  setDraft((d) => ({
                    ...d,
                    coverageMode: e.target.value as DomainProfile['coverageMode'],
                  }))
                }
                className="rounded-md border border-surface-border bg-surface-overlay px-2 py-1.5 text-xs text-text-primary focus:outline-none focus:ring-2 focus:ring-brand"
              >
                {COVERAGE_MODES.map((m) => (
                  <option key={m.value} value={m.value}>
                    {m.value}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className="text-sm text-text-secondary">Citation Mode</label>
              <select
                value={draft.citationMode}
                onChange={(e) =>
                  setDraft((d) => ({
                    ...d,
                    citationMode: e.target.value as DomainProfile['citationMode'],
                  }))
                }
                className="rounded-md border border-surface-border bg-surface-overlay px-2 py-1.5 text-xs text-text-primary focus:outline-none focus:ring-2 focus:ring-brand"
              >
                {CITATION_MODES.map((m) => (
                  <option key={m.value} value={m.value}>
                    {m.value}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className="text-sm text-text-secondary">Strip Mode</label>
              <select
                value={draft.stripMode}
                onChange={(e) =>
                  setDraft((d) => ({ ...d, stripMode: e.target.value as DomainProfile['stripMode'] }))
                }
                className="rounded-md border border-surface-border bg-surface-overlay px-2 py-1.5 text-xs text-text-primary focus:outline-none focus:ring-2 focus:ring-brand"
              >
                {STRIP_MODES.map((m) => (
                  <option key={m.value} value={m.value}>
                    {m.value}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {COVERAGE_MODES.concat(CITATION_MODES as never, STRIP_MODES as never).length > 0 && (
            <p className="text-[11px] text-text-disabled">
              {COVERAGE_MODES.find((m) => m.value === draft.coverageMode)?.label}
              <br />
              {CITATION_MODES.find((m) => m.value === draft.citationMode)?.label}
              <br />
              {STRIP_MODES.find((m) => m.value === draft.stripMode)?.label}
            </p>
          )}

          <label className="flex items-center gap-2 text-sm text-text-primary">
            <input
              type="checkbox"
              checked={draft.gate1Skip}
              onChange={(e) => setDraft((d) => ({ ...d, gate1Skip: e.target.checked }))}
            />
            Gate 1 Skip (skip vagueness check — answer every question directly)
          </label>

          <div className="flex flex-col gap-1.5">
            <label htmlFor="domain-system-prompt-ext" className="text-sm text-text-secondary">
              System Prompt Extension
            </label>
            <textarea
              id="domain-system-prompt-ext"
              value={draft.systemPromptExt}
              onChange={(e) => setDraft((d) => ({ ...d, systemPromptExt: e.target.value }))}
              rows={2}
              placeholder="Always include time and space complexity..."
              className="rounded-md border border-surface-border bg-surface-overlay px-3 py-2 text-sm text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-brand"
            />
          </div>

          <Input
            label="Domain Keywords (comma-separated)"
            value={draft.domainKeywordsText}
            onChange={(e) => setDraft((d) => ({ ...d, domainKeywordsText: e.target.value }))}
            placeholder="algorithm, data structure, complexity"
          />

          <div className="grid grid-cols-2 gap-3">
            <Input
              label="Max Tokens (Flat)"
              type="number"
              min={0}
              value={draft.maxTokensFlat}
              onChange={(e) =>
                setDraft((d) => ({ ...d, maxTokensFlat: Number(e.target.value) || 0 }))
              }
              placeholder="0 = default (1500)"
            />
            <Input
              label="Max Tokens (Structured)"
              type="number"
              min={0}
              value={draft.maxTokensStructured}
              onChange={(e) =>
                setDraft((d) => ({ ...d, maxTokensStructured: Number(e.target.value) || 0 }))
              }
              placeholder="0 = default (3500)"
            />
          </div>
          <p className="text-[11px] text-text-disabled">
            0 means “use the backend default.” Structured answers (categorized experts) pack
            prose + code + test cases into one JSON object and routinely need a higher cap
            than flat-text answers for the same domain.
          </p>

          {error && <p className="text-xs text-mode-refuse">{error}</p>}
          <div className="flex justify-end gap-2">
            {editingDomain && (
              <Button type="button" variant="ghost" onClick={resetForm}>
                Cancel
              </Button>
            )}
            <Button type="submit" isLoading={isSubmitting}>
              Save Domain Profile
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

export const Component = AdminDomainProfiles
