import type { Project } from '@/types/project'
import { Button } from '@/components/ui/Button'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 10 ("GitHub/GitLab" panel:
 * "Not connected [Connect Repository]").
 *
 * Cross-questioned: what does the connected state look like? The
 * wireframe only shows the disconnected state explicitly. Inferred
 * the connected state from Project's own repoUrl/repoProvider/
 * repoLastSync fields (which DO exist in the data model, section 5's
 * `projects` table columns repo_url/repo_provider/repo_last_sync) -
 * showing provider + last sync time is the minimum useful information
 * given those fields exist, without inventing UI elements (e.g. a
 * sync-progress bar) that have no backing field to drive them.
 */
export function RepoStatus({
  project,
  onConnect,
}: {
  project: Project
  onConnect?: () => void
}) {
  if (!project.repoConnected) {
    return (
      <div className="flex items-center justify-between rounded-lg border border-surface-border bg-surface-raised p-4">
        <p className="text-sm text-text-secondary">Not connected</p>
        <Button variant="secondary" size="sm" onClick={onConnect}>
          Connect Repository
        </Button>
      </div>
    )
  }

  return (
    <div className="rounded-lg border border-surface-border bg-surface-raised p-4">
      <p className="text-sm text-text-primary">
        {project.repoProvider} \u00b7 {project.repoUrl}
      </p>
      {project.repoLastSync && (
        <p className="mt-1 text-xs text-text-secondary">Last synced: {project.repoLastSync}</p>
      )}
    </div>
  )
}
