import { Link, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getProjects } from '@/api/projects'
import { queryKeys } from '@/api/queryKeys'
import { useUIStore } from '@/stores/uiStore'
import { cn } from '@/utils/cn'

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

  const { data: projects, isLoading } = useQuery({
    queryKey: queryKeys.projects.all,
    queryFn: () => getProjects(),
    enabled: sidebarOpen,
  })

  return (
    <aside
      className={cn(
        'flex-shrink-0 overflow-y-auto border-r border-surface-border bg-surface-raised transition-all',
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
            <Link to="/" className="text-xs text-brand hover:underline">
              + New
            </Link>
          </div>

          {isLoading && <p className="text-xs text-text-disabled">Loading...</p>}
          {!isLoading && projects?.length === 0 && (
            <p className="text-xs text-text-disabled">No projects yet</p>
          )}

          <ul className="space-y-0.5">
            {projects?.map((project) => (
              <li key={project.id}>
                <Link
                  to={`/projects/${project.id}`}
                  className={cn(
                    'block truncate rounded-md px-2 py-1.5 text-text-secondary hover:bg-surface-overlay hover:text-text-primary',
                    project.id === activeProjectId && 'bg-surface-overlay text-text-primary'
                  )}
                  title={project.name}
                >
                  {project.id === activeProjectId ? '\u25b8 ' : ''}
                  {project.name}
                </Link>
              </li>
            ))}
          </ul>
        </nav>
      )}
    </aside>
  )
}
