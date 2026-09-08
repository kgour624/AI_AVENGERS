/**
 * China Wall DomainProfile domain model (admin-configurable, no
 * redeploy needed). Source: backend-go/internal/chinawall/domain_profile.go
 * (DomainProfile struct) and backend-go/internal/admin/admin_handler.go
 * (domainProfileRow — the wire shape for the admin CRUD endpoints).
 *
 * WHY this exists (2026-09-08): MaxTokensFlat/MaxTokensStructured were
 * added to let admin tune the LLM response token cap per domain without
 * a backend redeploy — a hardcoded limit caused a real truncated-JSON
 * incident for one domain (see HANDOFF.md's 2026-09-08 round-3 RCA).
 * Exposed here alongside every other DomainProfile field (Option B
 * scope, not just the two token fields) since the admin API itself
 * makes the whole struct editable, not a narrow max-tokens-only route.
 */

export type CoverageMode = 'APPLY_PRINCIPLES' | 'LITERAL_MATCH'
export type CitationMode = 'LOOSE' | 'STRICT'
export type StripMode = 'CODE_EXEMPT' | 'FULL_STRIP'

export interface DomainRule {
  id: string
  description: string
  condition: string
  action: string
}

/**
 * Wire shape for GET /admin/domain-profiles, GET .../:domain, and the
 * PATCH .../:domain response. Matches admin_handler.go's
 * domainProfileRow field-for-field.
 */
export interface DomainProfile {
  domain: string
  gate1Skip: boolean
  coverageMode: CoverageMode
  citationMode: CitationMode
  stripMode: StripMode
  systemPromptExt: string
  domainKeywords: string[]
  customRules: DomainRule[]
  /** 0 = use DefaultMaxTokensFlat (1500) backend-side. */
  maxTokensFlat: number
  /** 0 = use DefaultMaxTokensStructured (3500) backend-side. */
  maxTokensStructured: number
}

export const COVERAGE_MODES: { value: CoverageMode; label: string }[] = [
  { value: 'APPLY_PRINCIPLES', label: 'Apply Principles (DSA, coding — principles transfer to new problems)' },
  { value: 'LITERAL_MATCH', label: 'Literal Match (medical, legal, finance — facts must be in transcript)' },
]

export const CITATION_MODES: { value: CitationMode; label: string }[] = [
  { value: 'LOOSE', label: 'Loose (cite principles, code blocks exempt)' },
  { value: 'STRICT', label: 'Strict (every claim must cite a chunk)' },
]

export const STRIP_MODES: { value: StripMode; label: string }[] = [
  { value: 'CODE_EXEMPT', label: 'Code Exempt (fenced code kept unconditionally)' },
  { value: 'FULL_STRIP', label: 'Full Strip (every uncited sentence removed)' },
]
