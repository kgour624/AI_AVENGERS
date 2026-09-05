import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Header } from './Header'

/** Main layout wrapper. Source: FRONTEND_SYSTEM_DESIGN.md section 1 (Main Interface Layout ASCII), section 3. */
export function AppShell() {
  return (
    <div className="flex h-screen flex-col bg-surface-base text-text-primary">
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
