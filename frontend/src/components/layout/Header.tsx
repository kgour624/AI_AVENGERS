import { useUIStore } from '@/stores/uiStore'
import { useAuthStore } from '@/stores/authStore'
import { Link } from 'react-router-dom'

export function Header() {
  const toggleSidebar = useUIStore((s) => s.toggleSidebar)
  const isAdmin = useAuthStore((s) => s.isAdmin())

  return (
    <header className="flex h-12 flex-shrink-0 items-center justify-between border-b border-surface-border bg-surface-raised px-4">
      <div className="flex items-center gap-3">
        <button
          onClick={toggleSidebar}
          aria-label="Toggle sidebar"
          className="text-text-secondary hover:text-text-primary"
        >
          \u2630
        </button>
        <Link to="/" className="font-semibold text-text-primary">
          \u26A1 AI Avengers
        </Link>
      </div>
      {isAdmin && (
        <Link to="/admin" className="text-sm text-text-secondary hover:text-text-primary">
          Admin
        </Link>
      )}
    </header>
  )
}
