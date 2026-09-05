import { useQuery } from '@tanstack/react-query'
import { getIngestionJobs } from '@/api/admin'

/**
 * Polls ingestion job status for a given expert.
 * WHY polling (refetchInterval) rather than a one-shot fetch: the
 * backend runs ingestion in a detached goroutine (see
 * AdminHandler.IngestTranscript's `go func() { ... }()`) with no SSE
 * or WebSocket push for job progress - the ONLY way to observe
 * progress is to poll GET /admin/experts/:id/jobs repeatedly. Verified
 * this by reading the handler: there is no other mechanism.
 *
 * Cross-questioned: should this poll forever, even after the job
 * completes? refetchInterval returns `false` once the most recent job
 * is no longer 'pending'/'running', stopping the poll - otherwise this
 * would keep hitting the endpoint every few seconds indefinitely for a
 * job that finished hours ago, for as long as the admin page stays open.
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
      return latest.status === 'pending' || latest.status === 'running' ? 3000 : false
    },
  })
}
