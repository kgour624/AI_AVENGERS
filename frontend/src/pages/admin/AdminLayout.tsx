import { Outlet } from 'react-router-dom'

/**
 * Admin layout wrapper - lazy-loaded, per FRONTEND_SYSTEM_DESIGN.md
 * section 5 ("code split, only loads for admin users"). This module
 * is only ever imported via the router's `lazy: () => import(...)`,
 * never a static import, or the whole point of the code-split is lost.
 */
export function AdminLayout() {
  return (
    <div className="flex h-screen bg-surface-base text-text-primary">
      <aside className="w-56 flex-shrink-0 border-r border-surface-border bg-surface-raised p-4">
        <p className="mb-4 font-semibold">\u26A1 AI Avengers Admin</p>
      </aside>
      <main className="flex-1 overflow-y-auto">
        <Outlet />
      </main>
    </div>
  )
}

// react-router-dom's `lazy` loader convention expects a default export
// providing `Component` (or named exports it destructures) - see App.tsx
// route definitions using `lazy: () => import('./pages/admin/AdminLayout')`.
export const Component = AdminLayout
