/**
 * Deep snake_case <-> camelCase key conversion for API boundary.
 *
 * WHY this exists: see the casing note at the top of types/api.ts.
 * The backend's JSON responses use snake_case keys; every TypeScript
 * interface in this codebase is camelCase per FRONTEND_SYSTEM_DESIGN.md
 * section 13. Without this conversion, field access would silently
 * return `undefined` at runtime.
 *
 * Design choices, and why:
 * - Recurses into arrays and plain objects only. Does NOT recurse into
 *   `Date` instances, `File`, `Blob`, etc. - converting keys on those
 *   would corrupt them (e.g. Date has no own enumerable keys to convert,
 *   but a naive isObject check could still misbehave on other exotic
 *   objects, so we explicitly guard on "plain object" via constructor
 *   check).
 * - Does NOT convert keys that are already valid identifiers with no
 *   underscore (fast path, avoids unnecessary regex on every key).
 * - Skips null/undefined values without throwing.
 */

function isPlainObject(value: unknown): value is Record<string, unknown> {
  if (value === null || typeof value !== 'object') return false
  const proto = Object.getPrototypeOf(value)
  return proto === Object.prototype || proto === null
}

function snakeToCamel(key: string): string {
  if (!key.includes('_')) return key
  return key.replace(/_([a-z0-9])/g, (_, c: string) => c.toUpperCase())
}

function camelToSnake(key: string): string {
  return key.replace(/([A-Z])/g, '_$1').toLowerCase()
}

export function camelizeKeys<T = unknown>(input: unknown): T {
  if (Array.isArray(input)) {
    return input.map((item) => camelizeKeys(item)) as unknown as T
  }
  if (isPlainObject(input)) {
    const out: Record<string, unknown> = {}
    for (const [key, value] of Object.entries(input)) {
      out[snakeToCamel(key)] = camelizeKeys(value)
    }
    return out as unknown as T
  }
  return input as T
}

export function snakeifyKeys<T = unknown>(input: unknown): T {
  if (Array.isArray(input)) {
    return input.map((item) => snakeifyKeys(item)) as unknown as T
  }
  if (isPlainObject(input)) {
    const out: Record<string, unknown> = {}
    for (const [key, value] of Object.entries(input)) {
      out[camelToSnake(key)] = snakeifyKeys(value)
    }
    return out as unknown as T
  }
  return input as T
}
