import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getAdminExperts } from '@/api/admin'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { TranscriptUploadModal } from '@/components/admin/TranscriptUploadModal'
import { useIngestionStatus } from '@/hooks/useIngestionStatus'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 11 ("Admin Expert
 * Management" wireframe: per-expert card with status/chunks/rating,
 * [Upload Transcript] [Edit Charter] [Disable] actions, upload modal).
 *
 * "Edit Charter" is NOT wired up in this pass - AdminHandler.UpdateExpert
 * does accept a reasoning_charter field, so the endpoint exists, but a
 * charter-editing UI (likely a large textarea + the WHY-principle
 * explanation from the architecture doc) is enough of its own scope
 * that bundling it into this same commit would risk rushing it. Left
 * as a clearly-labeled disabled button rather than silently omitted,
 * so it's visible as a known gap, not an invisible one.
 */
function ExpertRow({ expertId }: { expertId: string }) {
  const { data: jobs } = useIngestionStatus(expertId)
  const latestJob = jobs?.[0]

  if (!latestJob) return null

  return (
    <p className="mt-1 text-xs text-text-secondary">
      Last ingestion: {latestJob.status === 'complete' ? '\u2705' : latestJob.status === 'failed' ? '\u274c' : '\u23f3'}{' '}
      {latestJob.status} ({latestJob.processedChunks}/{latestJob.totalChunks} chunks)
    </p>
  )
}

function AdminExperts() {
  const { data: experts, isLoading } = useQuery({
    queryKey: ['admin', 'experts'],
    queryFn: getAdminExperts,
  })
  const queryClient = useQueryClient()
  const [uploadTargetId, setUploadTargetId] = useState<string | null>(null)

  return (
    <div className="p-6">
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-xl font-semibold">Experts</h1>
      </div>

      {isLoading && (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-20" />
          ))}
        </div>
      )}

      <div className="space-y-3">
        {experts?.map((expert) => (
          <Card key={expert.id}>
            <div className="flex items-center justify-between">
              <p className="font-medium text-text-primary">{expert.name}</p>
              <Badge variant={expert.isActive ? 'brand' : 'neutral'}>
                {expert.isActive ? '\u2705 Active' : '\u23f8 Disabled'}
              </Badge>
            </div>
            <p className="mt-1 text-xs text-text-secondary">
              Chunks: {expert.totalChunks} | Rating: {expert.avgRating}
            </p>
            <ExpertRow expertId={expert.id} />

            <div className="mt-3 flex gap-2">
              <Button variant="secondary" size="sm" onClick={() => setUploadTargetId(expert.id)}>
                Upload Transcript
              </Button>
              <Button variant="ghost" size="sm" disabled title="Charter editing UI - not yet built">
                Edit Charter
              </Button>
            </div>
          </Card>
        ))}
      </div>

      {uploadTargetId && (
        <TranscriptUploadModal
          isOpen={uploadTargetId !== null}
          onClose={() => setUploadTargetId(null)}
          expertId={uploadTargetId}
          onIngestStarted={() => {
            // Invalidate this expert's job-status query so useIngestionStatus
            // picks up the newly-created job on its next poll, rather than
            // waiting for the existing (possibly stale/disabled) query state.
            queryClient.invalidateQueries({ queryKey: ['admin', 'experts', uploadTargetId, 'jobs'] })
          }}
        />
      )}
    </div>
  )
}

export const Component = AdminExperts
