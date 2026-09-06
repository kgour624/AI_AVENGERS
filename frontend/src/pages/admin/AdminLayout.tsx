import { NavLink, Outlet } from 'react-router-dom'
import { cn } from '@/utils/cn'

const NAV_ITEMS = [
  { to: '/admin', label: 'Overview', end: true },
  { to: '/admin/experts', label: 'Experts', end: false },
  { to: '/admin/clients', label: 'Clients', end: false },
  { to: '/admin/stats', label: 'Stats', end: false },
  { to: '/admin/settings', label: 'Settings', end: false },
]

/**
 * Admin layout wrapper - lazy-loaded, per FRONTEND_SYSTEM_DESIGN.md
 * section 5 ("code split, only loads for admin users"). This module
 * is only ever imported via the router's `lazy: () => import(...)`,
 * never a static import, or the whole point of the code-split is lost.
 *
 * Bug 1.4 fix (docs bug list): this sidebar had no navigation at all -
 * an admin could only reach a sub-page (Experts/Clients/Stats/
 * Settings) by typing the URL directly, since none of the 5 real
 * routes registered in App.tsx's /admin subtree were ever linked here.
 */
export function AdminLayout() {
  return (
    <div className="flex h-screen bg-surface-base text-text-primary">
      <aside className="w-56 flex-shrink-0 border-r border-surface-border bg-surface-raised p-4">
        <p className="mb-4 font-semibold">{'\u26A1'} AI Avengers Admin</p>
        <nav className="flex flex-col gap-1">
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                cn(
                  'rounded-md px-2 py-1.5 text-sm text-text-secondary hover:bg-surface-overlay hover:text-text-primary',
                  isActive && 'bg-surface-overlay text-text-primary'
                )
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
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
