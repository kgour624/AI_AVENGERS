import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Header } from './Header'

export function AppShell() {
  return (
    <div className="arc-atmosphere flex h-screen flex-col overflow-hidden text-text-primary">
      {/* Scan-line sweep — decorative, aria-hidden, reduced-motion safe */}
      <div aria-hidden="true" className="scan-line" />
      <Header />
      <div className="flex flex-1 overflow-hidden">
        <Sidebar />
        {/* min-w-0 (2026-09-09 round 11): without this, a flex item
            defaults to min-width:auto - it will not shrink below its
            content's intrinsic width, so any wide descendant (long
            unwrapped text, a table, etc. rendered by a page inside
            <Outlet/>) expands THIS box past the viewport, and because
            this box is a flex sibling of Header in the SAME outer
            flex-column above, the whole page gets a horizontal
            scrollbar - shifting Header (and its Admin/Logout buttons)
            off-screen along with it. min-w-0 lets this box actually
            shrink to the available space; any overflow then stays
            CONTAINED inside main's own overflow-hidden, never
            propagating up to the page. */}
        <main className="flex min-w-0 flex-1 flex-col overflow-hidden min-h-0">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
