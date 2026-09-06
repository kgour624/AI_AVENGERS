import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Header } from './Header'

/**
 * Main layout wrapper. Source: FRONTEND_SYSTEM_DESIGN.md section 1
 * (Main Interface Layout ASCII), section 3.
 *
 * ARC-51 §6: bg-surface-void (darker than the old bg-surface-base) -
 * Header/Sidebar are now glass panels that need a darker base to read
 * as "floating" above the page rather than blending into it.
 */
export function AppShell() {
  return (
    <div className="arc-atmosphere flex h-screen flex-col text-text-primary">
      <Header />
      <div className="flex flex-1 overflow-hidden">
        <Sidebar />
        <main className="flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
