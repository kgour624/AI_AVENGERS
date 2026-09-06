import { Link, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { motion, useReducedMotion } from 'framer-motion'
import { getProjects } from '@/api/projects'
import { queryKeys } from '@/api/queryKeys'
import { useUIStore } from '@/stores/uiStore'
import { cn } from '@/utils/cn'
import { fadeUp, staggerContainer, ARC_MOTION } from '@/design-system/motion'

/**
 * Left sidebar: real project list + active-project highlight.
 *
 * Bug 1.2 fix (docs bug list): this was a static `<p>Projects</p>`
 * stub with a comment promising "Phase 2" wiring (`useQuery(queryKeys.
 * projects.all)`) that never actually happened in any later phase -
 * no list, no active state, no links, despite getProjects() and
 * queryKeys.projects.all already existing exactly as the old comment
 * described.
 */
export function Sidebar() {
  const sidebarOpen = useUIStore((s) => s.sidebarOpen)
  const { projectId: activeProjectId } = useParams<{ projectId: string }>()
  const reduceMotion = useReducedMotion()

  const { data: projects, isLoading } = useQuery({
    queryKey: queryKeys.projects.all,
    queryFn: () => getProjects(),
    enabled: sidebarOpen,
  })

  return (
    // ARC-51 §6: glass panel, same width-collapse logic as before
    // (sidebarOpen ? 'w-64' : 'w-0') - only the border/background
    // classes and the list's entrance motion changed.
    <aside
      className={cn(
        'flex-shrink-0 overflow-y-auto border-r border-glass-border bg-surface-raised/70 backdrop-blur-xl transition-all duration-150 ease-arc',
        sidebarOpen ? 'w-64' : 'w-0'
      )}
    >
      {sidebarOpen && (
        <nav className="p-4 text-sm">
          <div className="mb-2 flex items-center justify-between">
            <p className="text-xs uppercase tracking-wide text-text-disabled">Projects</p>
            {/* Points at ProjectsPage, where the actual Create Project
                modal lives (bug 1.1 fix) - avoids a second, duplicate
                create-project form living here too. */}
            <Link to="/" className="text-xs text-glow-purple hover:underline">
              + New
            </Link>
          </div>

          {isLoading && <p className="text-xs text-text-disabled">Loading...</p>}
          {!isLoading && projects?.length === 0 && (
            <p className="text-xs text-text-disabled">No projects yet</p>
          )}

          {/* ARC-51 §3/§6: stagger entrance, capped at
              ARC_MOTION.maxStaggerItems staggered items - beyond that,
              remaining items render inside the same motion.ul (which
              already finished its own stagger timing) without
              individual per-item delay, so a long project list never
              feels sluggish to enter (see motion.ts's staggerContainer
              doc comment for why this cap is enforced here, by the
              consumer, not inside the shared variants object). */}
          <motion.ul
            initial={reduceMotion ? undefined : 'hidden'}
            animate="visible"
            variants={staggerContainer}
            className="space-y-0.5"
          >
            {projects?.map((project, i) => (
              <motion.li
                key={project.id}
                variants={i < ARC_MOTION.maxStaggerItems ? fadeUp : undefined}
              >
                <Link
                  to={`/projects/${project.id}`}
                  className={cn(
                    'block truncate rounded-md border-l-2 border-transparent px-2 py-1.5 text-text-secondary transition-colors duration-150 ease-arc hover:bg-surface-panel-hover hover:text-text-primary',
                    project.id === activeProjectId &&
                      'border-glow-purple bg-surface-panel-hover text-text-primary'
                  )}
                  title={project.name}
                >
                  {project.name}
                </Link>
              </motion.li>
            ))}
          </motion.ul>
        </nav>
      )}
    </aside>
  )
}
