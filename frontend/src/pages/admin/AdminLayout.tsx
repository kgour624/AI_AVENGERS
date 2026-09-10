import { NavLink, Outlet, Link, useNavigate } from 'react-router-dom'
import { cn } from '@/utils/cn'
import { useAuthStore } from '@/stores/authStore'
import { logout as logoutApi } from '@/api/auth'
import { queryClient } from '@/api/queryKeys'

const NAV_ITEMS = [
  { to: '/admin',                 label: 'Overview',        end: true  },
  { to: '/admin/experts',         label: 'Experts',         end: false },
  { to: '/admin/categories',      label: 'Categories',      end: false },
  { to: '/admin/domain-profiles', label: 'Domain Profiles', end: false },
  { to: '/admin/clients',         label: 'Clients',         end: false },
  { to: '/admin/stats',           label: 'Stats',           end: false },
  { to: '/admin/settings',        label: 'Settings',        end: false },
  { to: '/admin/llm-settings',    label: 'LLM Settings',    end: false },
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
/**
 * ARC-51 §10 (docs/ARC51_UI_CONTRACT.md): same glass language as the
 * client-facing Header/Sidebar (§6), but deliberately LOWER drama -
 * no ambient glow, no stagger motion on the nav list (it's a fixed 5
 * items, not a dynamic data list) - §1's page-by-page notes are
 * explicit that admin needs density/information, not spectacle. Every
 * NavLink/Outlet/cn() logic line is untouched, only classNames changed.
 */
export function AdminLayout() {
  return (
    <div className="arc-atmosphere flex h-screen text-text-primary">
      <aside className="w-56 flex-shrink-0 border-r border-glass-border bg-surface-raised/70 p-4 backdrop-blur-xl">
        <p className="mb-4 font-semibold tracking-wide [text-shadow:0_0_12px_var(--glow-purple)]">
          {'\u26A1'} AI Avengers Admin
        </p>
        <nav className="flex flex-col gap-1">
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                cn(
                  'rounded-md border-l-2 border-transparent px-2 py-1.5 text-sm text-text-secondary transition-colors duration-150 ease-arc hover:bg-surface-panel-hover hover:text-text-primary',
                  isActive && 'border-glow-purple bg-surface-panel-hover text-text-primary'
                )
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>

        <Link
          to="/"
          className="mt-4 block border-t border-glass-border pt-3 text-sm text-text-secondary hover:text-text-primary"
        >
          {'\u2190'} Back to App
        </Link>
      </aside>
      <main className="min-w-0 flex-1 overflow-y-auto">
        <Outlet />
      </main>
    </div>
  )
}

// react-router-dom's `lazy` loader convention expects a default export
// providing `Component` (or named exports it destructures) - see App.tsx
// route definitions using `lazy: () => import('./pages/admin/AdminLayout')`.
export const Component = AdminLayout
