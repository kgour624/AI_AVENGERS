import { useQuery } from '@tanstack/react-query'
import { getRepoSyncStatus } from '@/api/repo'

/**
 * Polls repo sync status while a sync is in progress.
 * WHY polling: SyncRepo's real implementation runs `go s.runSync(...)`
 * in a detached goroutine (verified from source) with no push
 * mechanism - the same pattern as admin transcript ingestion
 * (useIngestionStatus.ts), and for the same reason: polling
 * GET /projects/:id/repo/status is the only way to observe progress.
 * Stops polling once status leaves pending/syncing.
 */
export function useRepoSyncStatus(projectId: string) {
  return useQuery({
    queryKey: ['projects', projectId, 'repo', 'status'],
    queryFn: () => getRepoSyncStatus(projectId),
    refetchInterval: (query) => {
      const data = query.state.data
      if (!data || !data.connected) return false
      return data.status === 'pending' || data.status === 'syncing' ? 3000 : false
    },
  })
}
