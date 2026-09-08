/**
 * Expert Category + structured template domain models (CT-D0/D1).
 * Source: backend-go/internal/category/registry.go (Category,
 * TemplateSchema, TemplateSection, SectionType) and
 * backend-go/internal/admin/admin_handler.go (categoryRow,
 * templateSchemaInput — the wire shapes for the admin CRUD endpoints).
 * Cross-referenced against CATEGORY_TEMPLATE_HANDOFF.md §2/§3.
 */

/** Matches category.SectionType exactly — CT-L4/CT-L5 constrain what a section can be. */
export type CategorySectionType = 'prose' | 'code' | 'test_cases'

export interface CategoryTemplateSection {
  key: string
  label: string
  type: CategorySectionType
  required: boolean
}

export interface CategoryTemplateSchema {
  /** Empty/omitted sections = no structured template, flat-text fallback (CT-L2). */
  sections: CategoryTemplateSection[]
}

/**
 * Wire shape for GET /admin/expert-categories and GET .../:id.
 * WHY templateSchema is typed as CategoryTemplateSchema (not raw JSON):
 * the backend's categoryRow.TemplateSchema is json.RawMessage on the
 * Go side, but baseAPI's response interceptor (see api/base.ts) runs
 * camelizeKeys on the WHOLE response body including nested JSON, so
 * by the time this reaches a component it is already a parsed,
 * camelCased object — not a string needing a second JSON.parse.
 */
export interface ExpertCategory {
  id: string
  name: string
  slug: string
  description: string
  templateSchema: CategoryTemplateSchema
  defaultLanguage: string
  askStructurePermission: boolean
  createdAt: string
  updatedAt: string
}

/**
 * Hardcoded test-case buckets (CT-L4) — mirrors chinawall.TestCaseBuckets
 * exactly. Never admin-configurable; shown here only for the template
 * builder's "Test Cases" section preview so admin sees what will
 * actually render, without pretending it's an editable field.
 */
export const TEST_CASE_BUCKETS = ['BASE', 'EDGE', 'CORNER', 'STRESS'] as const

export const CATEGORY_SECTION_TYPES: { value: CategorySectionType; label: string }[] = [
  { value: 'prose', label: 'Prose (explanation text, citation-processed)' },
  { value: 'code', label: 'Code (fenced block, default_language, citation-exempt)' },
  { value: 'test_cases', label: 'Test Cases (fixed BASE/EDGE/CORNER/STRESS buckets)' },
]
