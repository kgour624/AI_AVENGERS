import { clsx, type ClassValue } from 'clsx'

/**
 * Tailwind class merge helper.
 * WHY a helper instead of raw template strings: conditional classes
 * (`isActive && 'bg-brand'`) produce `false` in the string when the
 * condition is falsy unless filtered - clsx handles this correctly.
 * We intentionally skip tailwind-merge (dedup of conflicting utility
 * classes) for this Phase 1 scaffold since no component yet produces
 * conflicting class combinations; revisit if that changes.
 */
export function cn(...inputs: ClassValue[]): string {
  return clsx(...inputs)
}
