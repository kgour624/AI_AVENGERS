import { useState } from 'react'
import { Link } from 'react-router-dom'
import type { Project } from '@/types/project'
import { deleteProject, disableProject, enableProject } from '@/api/projects'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { ExpertAvatar } from '@/components/expert/ExpertAvatar'
import { formatRelativeTime } from '@/utils/format'
import { handleAPIError } from '@/utils/errors'

const MAX_AVATAR_STACK = 4

/**
 * ProjectCard — shows project summary + quick-action buttons.
 *
 * Delete and archive/unarchive are now available directly from the
 * ProjectsPage grid so the user does not have to navigate into a
 * project just to remove it.
 *
 * WHY callbacks not router navigation:
 *   ProjectsPage sources its list from a React Router loader
 *   (useLoaderData). The correct invalidation mechanism is
 *   revalidator.revalidate(), not queryClient.invalidate() — same
 *   pattern ProjectPage already uses for its own delete button.
 *   Callbacks let the parent (ProjectsPage) own the revalidation.
 *
 * WHY window.confirm for delete:
 *   Matches the existing pattern in ProjectPage.handleDelete and
 *   ChatPage.handleDeleteMessage — consistent UX, no new modal needed.
 */
export function ProjectCard({
  project,
  onDeleted,
  onArchiveToggled,
}: {
  project: Project
  onDeleted?: () => void
  onArchiveToggled?: () => void
}) {
  const experts = project.experts ?? []
  const [isActing, setIsActing] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)

  const isArchived = project.status === 'archived'

  async function handleDelete() {
    if (!window.confirm(`Delete "${project.name}"? This cannot be undone.`)) return
    setIsActing(true)
    setActionError(null)
    try {
      await deleteProject(project.id)
      onDeleted?.()
    } catch (err) {
      setActionError(handleAPIError(err))
    } finally {
      setIsActing(false)
    }
  }

  async function handleArchiveToggle() {
    setIsActing(true)
    setActionError(null)
    try {
      if (isArchived) {
        await enableProject(project.id)
      } else {
        await disableProject(project.id)
      }
      onArchiveToggled?.()
    } catch (err) {
      setActionError(handleAPIError(err))
    } finally {
      setIsActing(false)
    }
  }

  return (
    <Card glow="purple">
      <div className="flex items-start justify-between gap-2">
        <p className="font-medium text-text-primary">{project.name}</p>
        <div className="flex items-center gap-1.5">
          {/* Quick-action buttons — visible on hover */}
          <div className="flex items-center gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
            <button
              aria-label={isArchived ? 'Restore project' : 'Archive project'}
              title={isArchived ? 'Restore' : 'Archive'}
              disabled={isActing}
              onClick={handleArchiveToggle}
              className="rounded p-0.5 text-xs text-text-disabled hover:text-glow-amber disabled:opacity-40"
            >
              {isArchived ? '\u21a9' : '\u23f8'}
            </button>
            <button
              aria-label="Delete project"
              title="Delete project"
              disabled={isActing}
              onClick={handleDelete}
              className="rounded p-0.5 text-xs text-text-disabled hover:text-mode-refuse disabled:opacity-40"
            >
              {'\u2715'}
            </button>
          </div>
          {/* Live pulse dot */}
          {!isArchived && (
            <span className="relative mt-0.5 flex h-2 w-2 flex-shrink-0" aria-hidden="true">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-glow-purple opacity-60" />
              <span className="relative inline-flex h-2 w-2 rounded-full bg-glow-purple" />
            </span>
          )}
        </div>
      </div>

      {project.architectureType && (
        <span className="mt-1 inline-block rounded-full border border-glass-border bg-surface-overlay px-2 py-0.5 text-xs text-glow-cyan">
          {'\u26a1'} {project.architectureType}
        </span>
      )}

      {isArchived && (
        <span className="mt-1 inline-block rounded-full border border-glow-amber/30 bg-glow-amber/10 px-2 py-0.5 text-xs text-glow-amber">
          Archived
        </span>
      )}

      <p className="mt-2 text-sm text-text-secondary">Experts: {experts.length}</p>

      {experts.length > 0 && (
        <div className="mt-1 flex -space-x-2" aria-hidden="true">
          {experts.slice(0, MAX_AVATAR_STACK).map((e) => (
            <ExpertAvatar
              key={e.expertId}
              domain={e.domain}
              status="idle"
              size="sm"
              className="rounded-full ring-2 ring-surface-raised"
            />
          ))}
          {experts.length > MAX_AVATAR_STACK && (
            <span className="flex h-8 w-8 items-center justify-center rounded-full bg-surface-overlay text-xs text-text-secondary ring-2 ring-surface-raised">
              +{experts.length - MAX_AVATAR_STACK}
            </span>
          )}
        </div>
      )}

      <p className="mt-2 text-sm text-text-secondary">
        Last active: {formatRelativeTime(project.updatedAt)}
      </p>

      {actionError && (
        <p className="mt-1 text-xs text-mode-refuse">{actionError}</p>
      )}

      <Link to={`/projects/${project.id}`} className="mt-4 inline-block">
        <Button variant="secondary" size="sm">
          Open Project
        </Button>
      </Link>
    </Card>
  )
}
