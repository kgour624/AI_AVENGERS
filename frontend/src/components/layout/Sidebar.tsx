import { Link, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { motion, useReducedMotion } from 'framer-motion'
import { getProjects } from '@/api/projects'
import { queryKeys } from '@/api/queryKeys'
import { useUIStore } from '@/stores/uiStore'
import { cn } from '@/utils/cn'
import { fadeUp, staggerContainer, ARC_MOTION } from '@/design-system/motion'

export function Sidebar() {
  const sidebarOpen = useUIStore((s) => s.sidebarOpen)
  const setSidebarOpen = useUIStore((s) => s.setSidebarOpen)
  const { projectId: activeProjectId } = useParams<{ projectId: string }>()
  const reduceMotion = useReducedMotion()

  const { data: projects, isLoading } = useQuery({
    queryKey: queryKeys.projects.all,
    queryFn: () => getProjects(),
    enabled: sidebarOpen,
  })

  return (
    <>
      {/* Mobile-only backdrop (2026-09-09 responsive fix): below sm,
          the sidebar switches from push-layout to a fixed overlay (see
          <aside> classes below) - this backdrop lets a tap outside the
          drawer close it, and dims the page content behind it so it's
          clear the drawer is a temporary overlay, not part of the flow.
          Absent entirely on sm+ (push-layout there, no overlay concept). */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-30 bg-black/50 sm:hidden"
          aria-hidden="true"
          onClick={() => setSidebarOpen(false)}
        />
      )}
      <aside className={cn(
        'overflow-y-auto',
        'border-r border-glass-border',
        'bg-surface-base/70 backdrop-blur-xl',
        'transition-all duration-200 ease-arc',
        // Mobile (< sm): fixed overlay drawer, full height, slides via
        // width/translate - NOT a push layout, so it never squeezes the
        // page content underneath on a narrow viewport.
        // Desktop (sm+): original push-layout behavior, unchanged.
        'fixed inset-y-0 left-0 z-40 sm:static sm:z-auto',
        sidebarOpen ? 'w-64 translate-x-0' : 'w-0 -translate-x-full sm:translate-x-0',
        'flex-shrink-0'
      )}>
        {sidebarOpen && (
        <nav className="p-4">
          <div className="mb-3 flex items-center justify-between">
            <p className="text-[10px] font-semibold uppercase tracking-[0.2em] text-text-disabled">Projects</p>
            <Link to="/"
              className="text-[10px] font-medium uppercase tracking-wider text-glow-purple/60 transition-colors duration-150 ease-arc hover:text-glow-purple">
              + New
            </Link>
          </div>

          {isLoading && (
            <div className="space-y-1.5">
              {[1,2,3].map(i => (
                <div key={i} className="h-7 rounded-md bg-surface-overlay/50 animate-pulse" style={{ opacity: 1 - i * 0.2 }} />
              ))}
            </div>
          )}

          {!isLoading && projects?.length === 0 && (
            <p className="text-xs text-text-disabled italic">No projects yet</p>
          )}

          <motion.ul
            initial={reduceMotion ? undefined : 'hidden'}
            animate="visible"
            variants={staggerContainer}
            className="space-y-0.5"
          >
            {projects?.map((project, i) => (
              <motion.li key={project.id} variants={i < ARC_MOTION.maxStaggerItems ? fadeUp : undefined}>
                <Link
                  to={`/projects/${project.id}`}
                  title={project.name}
                  className={cn(
                    'group flex items-center gap-2 truncate rounded-md px-2.5 py-2',
                    'text-xs text-text-secondary',
                    'border-l-2 border-transparent',
                    'transition-all duration-150 ease-arc',
                    'hover:bg-surface-overlay/60 hover:text-text-primary hover:border-glow-purple/40',
                    project.id === activeProjectId && [
                      'border-glow-purple bg-surface-overlay/70 text-text-primary',
                      'shadow-[inset_0_0_12px_oklch(68%_0.28_295_/_0.08)]',
                    ]
                  )}
                >
                  <span className={cn(
                    'h-1.5 w-1.5 flex-shrink-0 rounded-full transition-colors duration-150',
                    project.id === activeProjectId
                      ? 'bg-glow-purple shadow-[0_0_6px_oklch(68%_0.28_295_/_0.8)]'
                      : 'bg-surface-border group-hover:bg-glow-purple/40'
                  )} />
                  <span className="truncate">{project.name}</span>
                </Link>
              </motion.li>
            ))}
          </motion.ul>
        </nav>
      )}
    </aside>
  )
}
