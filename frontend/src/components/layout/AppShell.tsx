import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Header } from './Header'

export function AppShell() {
  return (
    <div className="arc-atmosphere flex h-screen flex-col text-text-primary">
      {/* Scan-line sweep — decorative, aria-hidden, reduced-motion safe */}
      <div aria-hidden="true" className="scan-line" />
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
