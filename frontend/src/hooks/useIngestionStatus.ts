import { useQuery } from '@tanstack/react-query'
import { getIngestionJobs } from '@/api/admin'

/**
 * Polls ingestion job status for a given expert.
 *
 * FIX: 'paused' status added to polling condition.
 * Previously: only 'pending' | 'running' triggered polling.
 * Bug: when backend paused a job, polling stopped but UI kept showing
 * the last 'running' state (e.g. "Extracting...") indefinitely.
 * Fix: poll on 'paused' too so UI reflects the paused state immediately.
 */
export function useIngestionStatus(expertId: string | null) {
  return useQuery({
    queryKey: ['admin', 'experts', expertId, 'jobs'],
    queryFn: () => getIngestionJobs(expertId!),
    enabled: expertId !== null,
    refetchInterval: (query) => {
      const jobs = query.state.data
      const latest = jobs?.[0]
      if (!latest) return false
      // Poll while job is active OR paused (paused needs UI update too)
      return latest.status === 'pending'
        || latest.status === 'running'
        || latest.status === 'paused'
        ? 3000
        : false
    },
  })
}
