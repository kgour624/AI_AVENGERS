import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useRepoSyncStatus } from '@/hooks/useRepoSyncStatus'
import { syncRepo } from '@/api/repo'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
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
 * Feature #5 fix (docs bug list): Added manual "Sync Now" button,
 * error message display, and status badges for better UX.
 */

const STATUS_CONFIG = {
  pending: { variant: 'warning' as const, label: 'PENDING' },
  syncing: { variant: 'info' as const, label: 'SYNCING' },
  complete: { variant: 'success' as const, label: 'SYNCED' },
  failed: { variant: 'danger' as const, label: 'FAILED' },
}

export function RepoStatus({ projectId }: { projectId: string }) {
  const [isModalOpen, setIsModalOpen] = useState(false)
  const { data: status, isLoading } = useRepoSyncStatus(projectId)
  const queryClient = useQueryClient()

  const syncMutation = useMutation({
    mutationFn: () => syncRepo(projectId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['projects', projectId, 'repo', 'status'],
      })
    },
  })

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

  const statusConfig = STATUS_CONFIG[status.status]
  const isSyncing = status.status === 'syncing'

  return (
    <div className="rounded-lg border border-surface-border bg-surface-raised p-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Badge variant={statusConfig.variant}>{statusConfig.label}</Badge>
          <span className="text-sm font-medium text-text-primary">{status.repoName}</span>
          <span className="text-xs text-text-disabled">·</span>
          <span className="text-xs text-text-secondary">{status.totalChunks} chunks</span>
        </div>
        <Button
          variant="secondary"
          size="sm"
          onClick={() => syncMutation.mutate()}
          disabled={isSyncing || syncMutation.isPending}
          isLoading={syncMutation.isPending}
        >
          {isSyncing ? 'Syncing...' : 'Sync Now'}
        </Button>
      </div>

      {status.lastSyncAt && (
        <p className="mt-2 text-xs text-text-secondary">Last synced: {status.lastSyncAt}</p>
      )}

      {status.status === 'failed' && status.errorMessage && (
        <p className="mt-2 text-xs text-mode-refuse">
          <span className="font-medium">Error:</span> {status.errorMessage}
        </p>
      )}

      {syncMutation.isError && (
        <p className="mt-2 text-xs text-mode-refuse">
          Failed to trigger sync. Please try again.
        </p>
      )}
    </div>
  )
}
