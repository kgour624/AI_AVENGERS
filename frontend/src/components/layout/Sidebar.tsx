import { useUIStore } from '@/stores/uiStore'
import { cn } from '@/utils/cn'

/** Left sidebar: projects/chats nav. Content is a Phase-2 concern (data fetching); this is the structural shell only. */
export function Sidebar() {
  const sidebarOpen = useUIStore((s) => s.sidebarOpen)

  return (
    <aside
      className={cn(
        'flex-shrink-0 overflow-y-auto border-r border-surface-border bg-surface-raised transition-all',
        sidebarOpen ? 'w-64' : 'w-0'
      )}
    >
      {sidebarOpen && (
        <nav className="p-4 text-sm text-text-secondary">
          {/* PROJECTS / CHATS lists populated in Phase 2 via useQuery(queryKeys.projects.all) */}
          <p className="px-2 text-xs uppercase tracking-wide text-text-disabled">Projects</p>
        </nav>
      )}
    </aside>
  )
}
