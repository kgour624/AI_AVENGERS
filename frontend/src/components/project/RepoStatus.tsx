import { useState } from 'react'
import { useRepoSyncStatus } from '@/hooks/useRepoSyncStatus'
import { Button } from '@/components/ui/Button'
import { RepoConnectModal } from './RepoConnectModal'

/**
 * PHASE 5 REWRITE: replaces Phase 2's static, non-interactive version
 * (which only read Project.repoConnected/repoUrl/repoLastSync from the
 * project detail response and had no working onConnect handler at
 * all) with a real, polling-backed status display + working connect
 * flow. Source: FRONTEND_SYSTEM_DESIGN.md section 10's "GitHub/GitLab"
 * panel, now driven by GET /projects/:id/repo/status via
 * useRepoSyncStatus rather than the parent Project object's static
 * repo fields (which, per Service.ConnectRepo, DO get updated on
 * connect - but repo/status is the live source of truth for sync
 * progress, which the Project object has no field for at all).
 *
 * Still uses PAT-based connectRepo per the documented OAuth gap in
 * api/repo.ts - RepoConnectModal's copy explains this to the user
 * directly rather than presenting a button that implies OAuth.
 */
export function RepoStatus({ projectId }: { projectId: string }) {
  const [isModalOpen, setIsModalOpen] = useState(false)
  const { data: status, isLoading } = useRepoSyncStatus(projectId)

  if (isLoading) {
    return <div className="h-16 animate-pulse rounded-lg bg-surface-overlay" />
  }

  if (!status?.connected) {
    return (
      <>
        <div className="flex items-center justify-between rounded-lg border border-surface-border bg-surface-raised p-4">
          <p className="text-sm text-text-secondary">Not connected</p>
          <Button variant="secondary" size="sm" onClick={() => setIsModalOpen(true)}>
            Connect Repository
          </Button>
        </div>
        <RepoConnectModal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} projectId={projectId} />
      </>
    )
  }

  const statusIcon =
    status.status === 'complete' ? '\u2705' : status.status === 'failed' ? '\u274c' : '\u23f3'

  return (
    <div className="rounded-lg border border-surface-border bg-surface-raised p-4">
      <p className="text-sm text-text-primary">
        {statusIcon} {status.repoName} \u00b7 {status.status} \u00b7 {status.totalChunks} chunks
      </p>
      {status.lastSyncAt && (
        <p className="mt-1 text-xs text-text-secondary">Last synced: {status.lastSyncAt}</p>
      )}
    </div>
  )
}
