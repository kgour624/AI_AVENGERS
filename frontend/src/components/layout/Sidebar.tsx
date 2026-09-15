import { useState } from 'react'
import { Link, useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { motion, useReducedMotion } from 'framer-motion'
import { getProjects, disableProject, enableProject, deleteProject } from '@/api/projects'
import { queryKeys } from '@/api/queryKeys'
import { useUIStore } from '@/stores/uiStore'
import { cn } from '@/utils/cn'
import { fadeUp, staggerContainer, ARC_MOTION } from '@/design-system/motion'

// ConfirmDeleteModal: type-to-confirm for irreversible project deletion.
function ConfirmDeleteModal({
  projectName, onConfirm, onCancel, isPending,
}: { projectName: string; onConfirm: () => void; onCancel: () => void; isPending: boolean }) {
  const [typed, setTyped] = useState('')
  const match = typed.trim() === projectName.trim()
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="w-full max-w-sm rounded-xl border border-glass-border bg-surface-raised p-6 shadow-2xl">
        <h3 className="mb-2 text-sm font-semibold text-red-400">Delete Project</h3>
        <p className="mb-4 text-xs text-text-secondary">
          Permanently delete <span className="font-medium text-text-primary">{projectName}</span> and all its chats.
          Type the project name to confirm.
        </p>
        <input
          autoFocus
          value={typed}
          onChange={(e) => setTyped(e.target.value)}
          placeholder={projectName}
          className="w-full rounded-md border border-glass-border bg-surface-overlay px-3 py-2 text-sm text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-red-500/40"
        />
        <div className="mt-4 flex justify-end gap-2">
          <button onClick={onCancel} className="rounded-md px-3 py-1.5 text-xs text-text-secondary hover:text-text-primary">
            Cancel
          </button>
          <button
            disabled={!match || isPending}
            onClick={onConfirm}
            className="rounded-md bg-red-600/80 px-3 py-1.5 text-xs font-medium text-white disabled:opacity-40 hover:bg-red-600"
          >
            {isPending ? 'Deleting…' : 'Delete permanently'}
          </button>
        </div>
      </div>
    </div>
  )
}

export function Sidebar() {
  const sidebarOpen = useUIStore((s) => s.sidebarOpen)
  const setSidebarOpen = useUIStore((s) => s.setSidebarOpen)
  const { projectId: activeProjectId } = useParams<{ projectId: string }>()
  const reduceMotion = useReducedMotion()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; name: string } | null>(null)

  const { data: projects, isLoading } = useQuery({
    queryKey: queryKeys.projects.all,
    queryFn: () => getProjects(),
    enabled: sidebarOpen,
  })

  const disableMut = useMutation({
    mutationFn: (id: string) => disableProject(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.projects.all }),
  })
  const enableMut = useMutation({
    mutationFn: (id: string) => enableProject(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.projects.all }),
  })
  const deleteMut = useMutation({
    mutationFn: (id: string) => deleteProject(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.projects.all })
      if (deleteTarget?.id === activeProjectId) navigate('/')
      setDeleteTarget(null)
    },
  })

  const active = projects?.filter((p) => p.status !== 'archived') ?? []
  const passive = projects?.filter((p) => p.status === 'archived') ?? []

  const ProjectRow = ({ project, dimmed = false }: { project: { id: string; name: string; status: string }; dimmed?: boolean }) => (
    <div className="group relative flex items-center">
      <Link
        to={`/projects/${project.id}`}
        className={cn(
          'flex flex-1 items-center gap-2 truncate rounded-md px-2.5 py-2 pr-14',
          'text-xs border-l-2 border-transparent transition-all duration-150 ease-arc',
          dimmed
            ? 'text-text-disabled hover:text-text-secondary hover:bg-surface-overlay/30 hover:border-surface-border'
            : 'text-text-secondary hover:bg-surface-overlay/60 hover:text-text-primary hover:border-glow-purple/40',
          !dimmed && project.id === activeProjectId && [
            'border-glow-purple bg-surface-overlay/70 text-text-primary',
            'shadow-[inset_0_0_12px_oklch(68%_0.28_295_/_0.08)]',
          ]
        )}
      >
        <span className={cn(
          'h-1.5 w-1.5 flex-shrink-0 rounded-full',
          dimmed ? 'bg-surface-border' :
            project.id === activeProjectId ? 'bg-glow-purple shadow-[0_0_6px_oklch(68%_0.28_295_/_0.8)]' :
            'bg-surface-border group-hover:bg-glow-purple/40'
        )} />
        <span className="truncate">{project.name}</span>
      </Link>
      <div className="invisible absolute right-1 flex items-center gap-0.5 group-hover:visible">
        {dimmed ? (
          <button title="Enable" onClick={() => enableMut.mutate(project.id)}
            className="rounded px-1 py-0.5 text-[10px] text-text-disabled hover:text-glow-cyan">
            ↩
          </button>
        ) : (
          <button title="Disable" onClick={() => disableMut.mutate(project.id)}
            className="rounded px-1 py-0.5 text-[10px] text-text-disabled hover:text-amber-400">
            ⏸
          </button>
        )}
        <button title="Delete permanently" onClick={() => setDeleteTarget({ id: project.id, name: project.name })}
          className="rounded px-1 py-0.5 text-[10px] text-text-disabled hover:text-red-400">
          ✕
        </button>
      </div>
    </div>
  )

  return (
    <>
      {sidebarOpen && (
        <div className="fixed inset-0 z-30 bg-black/50 sm:hidden" aria-hidden="true" onClick={() => setSidebarOpen(false)} />
      )}
      <aside className={cn(
        'overflow-y-auto border-r border-glass-border bg-surface-base/70 backdrop-blur-xl',
        'transition-all duration-200 ease-arc flex-shrink-0',
        'fixed inset-y-0 left-0 z-40 sm:static sm:z-auto',
        sidebarOpen ? 'w-64 translate-x-0' : 'w-0 -translate-x-full sm:translate-x-0',
      )}>
        {sidebarOpen && (
          <nav className="p-4">
            <div className="mb-3 flex items-center justify-between">
              <p className="text-[10px] font-semibold uppercase tracking-[0.2em] text-text-disabled">Projects</p>
              <Link to="/" className="text-[10px] font-medium uppercase tracking-wider text-glow-purple/60 hover:text-glow-purple">
                + New
              </Link>
            </div>

            {isLoading && (
              <div className="space-y-1.5">
                {[1,2,3].map(i => <div key={i} className="h-7 rounded-md bg-surface-overlay/50 animate-pulse" style={{ opacity: 1 - i * 0.2 }} />)}
              </div>
            )}

            {!isLoading && active.length === 0 && (
              <p className="text-xs text-text-disabled italic">No active projects</p>
            )}

            <motion.ul initial={reduceMotion ? undefined : 'hidden'} animate="visible" variants={staggerContainer} className="space-y-0.5">
              {active.map((project, i) => (
                <motion.li key={project.id} variants={i < ARC_MOTION.maxStaggerItems ? fadeUp : undefined}>
                  <ProjectRow project={project} />
                </motion.li>
              ))}
            </motion.ul>

            {passive.length > 0 && (
              <div className="mt-5">
                <p className="mb-2 text-[10px] font-semibold uppercase tracking-[0.2em] text-text-disabled/50">Passive</p>
                <ul className="space-y-0.5">
                  {passive.map((project) => (
                    <li key={project.id}><ProjectRow project={project} dimmed /></li>
                  ))}
                </ul>
              </div>
            )}
          </nav>
        )}
      </aside>

      {deleteTarget && (
        <ConfirmDeleteModal
          projectName={deleteTarget.name}
          onConfirm={() => deleteMut.mutate(deleteTarget.id)}
          onCancel={() => setDeleteTarget(null)}
          isPending={deleteMut.isPending}
        />
      )}
    </>
  )
}

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
    </>
  )
}
