/**
 * CodeCraftAPI domain types.
 * Used by the admin model picker when provider = 'codecraftapi'.
 *
 * WHY a separate file:
 *   CodeCraftAPI-specific types are isolated here so they don't pollute
 *   the main expert/project/auth type files. If CodeCraftAPI is ever
 *   removed, this file is the only one to delete.
 *
 * Shape is intentionally loose ([key: string]: unknown) because the
 * /v1/models response shape is not publicly documented at integration
 * time. Tighten once verified against a real API call.
 */

export interface CodeCraftModel {
  id: string
  name?: string
  description?: string
  // Additional fields from /v1/models response — optional, shape unknown until tested
  [key: string]: unknown
}
