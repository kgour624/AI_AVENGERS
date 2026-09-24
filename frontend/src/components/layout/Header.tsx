import { Link, useNavigate } from 'react-router-dom'
import { useUIStore } from '@/stores/uiStore'
import { useAuthStore } from '@/stores/authStore'
import { logout as logoutApi } from '@/api/auth'
import { queryClient } from '@/api/queryKeys'
import { GlobalChatSearch } from '@/components/layout/GlobalChatSearch'

export function Header() {
  const navigate = useNavigate()
  const toggleSidebar = useUIStore((s) => s.toggleSidebar)
  const user = useAuthStore((s) => s.user)
  const isAdmin = useAuthStore((s) => s.isAdmin())
  const clearAuth = useAuthStore((s) => s.clearAuth)

  const handleLogout = async () => {
    try { await logoutApi() } catch { /* best-effort */ }
    clearAuth()
    queryClient.clear()
    navigate('/login', { replace: true })
  }

  return (
    <header className="relative flex h-14 flex-shrink-0 items-center justify-between gap-2 overflow-x-auto px-2 sm:gap-4 sm:px-4 border-b border-glass-border bg-surface-base/80 backdrop-blur-xl shadow-[0_1px_0_oklch(68%_0.28_295_/_0.15)]">
      <div className="flex flex-shrink-0 items-center gap-2 sm:gap-5">
        <button onClick={toggleSidebar} aria-label="Toggle sidebar"
          className="text-text-disabled transition-colors duration-150 ease-arc hover:text-glow-cyan text-lg">
          {'\u2630'}
        </button>
        <Link to="/" className="holo-text text-sm font-bold tracking-[0.15em] uppercase whitespace-nowrap">
          {'\u26A1'}<span className="hidden sm:inline"> AI Avengers</span>
        </Link>
        <Link to="/experts"
          className="text-xs font-medium uppercase tracking-wider text-text-disabled transition-colors duration-150 ease-arc hover:text-glow-cyan whitespace-nowrap">
          Experts
        </Link>
      </div>
      <div className="flex flex-shrink-0 items-center gap-2 sm:gap-4">
        {/* #27: global chat search across every project. */}
        <GlobalChatSearch />
        {isAdmin && (
          <Link to="/admin"
            className="text-xs font-medium uppercase tracking-wider text-glow-purple/70 transition-colors duration-150 ease-arc hover:text-glow-purple hover:drop-shadow-[0_0_8px_oklch(68%_0.28_295_/_0.6)] whitespace-nowrap">
            Admin
          </Link>
        )}
        {user && (
          <span className="hidden text-xs text-text-secondary sm:inline" title={user.email}>{user.fullName}</span>
        )}
        <button onClick={handleLogout}
          className="text-xs font-medium uppercase tracking-wider text-text-disabled transition-colors duration-150 ease-arc hover:text-mode-refuse whitespace-nowrap">
          Logout
        </button>
      </div>
    </header>
  )
}
