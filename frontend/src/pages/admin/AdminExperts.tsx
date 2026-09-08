import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getAdminExperts } from '@/api/admin'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { TranscriptUploadModal } from '@/components/admin/TranscriptUploadModal'
import { CreateExpertModal } from '@/components/admin/CreateExpertModal'
import { EditCharterModal } from '@/components/admin/EditCharterModal'
import { EditExpertConfigModal } from '@/components/admin/EditExpertConfigModal'
import { IngestionPipelineModal } from '@/components/admin/IngestionPipelineModal'
import { useIngestionStatus } from '@/hooks/useIngestionStatus'
import { trainingStatusLabel } from '@/types/expert'

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
const STAGE_LABELS: Record<string, string> = {
  pending: 'Waiting to start',
  chunking: 'Step 1/6: Splitting transcript',
  topic_extraction: 'Step 2/6: Extracting topics (LLM)',
  charter_extraction: 'Step 3/6: Extracting charter (LLM)',
  embedding: 'Step 4/6: Generating embeddings',
  storing: 'Step 5/6: Saving to database',
  smoke_test: 'Step 6/6: Running smoke test',
  complete: 'Complete',
  failed: 'Failed',
}

function ExpertRow({
  expertId,
  expertName,
  onViewProgress,
}: {
  expertId: string
  expertName: string
  onViewProgress: (expertId: string, name: string) => void
}) {
  const { data: jobs } = useIngestionStatus(expertId)
  const latestJob = jobs?.[0]

  if (!latestJob) return null

  const isActive = latestJob.status === 'running' || latestJob.status === 'pending'
  const stageLabel = STAGE_LABELS[latestJob.currentStage ?? latestJob.status] ?? latestJob.status

  return (
    <div className="mt-1 flex items-center justify-between">
      <p className="text-xs text-text-secondary">
        {latestJob.status === 'complete' ? '\u2705' : latestJob.status === 'failed' ? '\u274c' : '\u23f3'}{' '}
        {stageLabel}
        {latestJob.processedChunks > 0 && (
          <span className="ml-1 text-text-disabled">
            ({latestJob.processedChunks}/{latestJob.totalChunks})
          </span>
        )}
        {latestJob.costUsd && latestJob.costUsd > 0 && (
          <span className="ml-2 text-glow-amber/70">${latestJob.costUsd.toFixed(3)}</span>
        )}
      </p>
      {(isActive || latestJob.status === 'failed') && (
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onViewProgress(expertId, expertName)}
          className="text-[10px] px-2 py-0.5"
        >
          View Progress
        </Button>
      )}
    </div>
  )
}

function AdminExperts() {
  const { data: experts, isLoading } = useQuery({
    queryKey: ['admin', 'experts'],
    queryFn: getAdminExperts,
  })
  const queryClient = useQueryClient()
  const [uploadTargetId, setUploadTargetId] = useState<string | null>(null)
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [charterTargetId, setCharterTargetId] = useState<string | null>(null)
  const [configTargetId, setConfigTargetId] = useState<string | null>(null)
  // Pipeline progress modal — expertId drives SSE connection inside modal
  const [pipelineExpertId, setPipelineExpertId] = useState<string | null>(null)
  const [pipelineExpertName, setPipelineExpertName] = useState<string>('')

  return (
    <div className="p-6">
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-xl font-semibold">Experts</h1>
        {/* Bug 1.5 fix (docs bug list): no way to create a new expert
            existed anywhere on this page - only per-existing-expert
            actions (Upload Transcript, disabled Edit Charter). */}
        <Button onClick={() => setIsCreateOpen(true)}>+ New Expert</Button>
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
          <Card key={expert.id} glow="cyan">
            <div className="flex items-center justify-between">
              <p className="font-medium text-text-primary">{expert.name}</p>
              <div className="flex items-center gap-2">
                {/* A12: training status badge */}
                <Badge variant="neutral" className="text-xs">
                  {trainingStatusLabel(expert.trainingStatus)}
                </Badge>
                <Badge variant={expert.isActive ? 'brand' : 'neutral'}>
                  {expert.isActive ? '\u2705 Active' : '\u23f8 Disabled'}
                </Badge>
              </div>
            </div>
            <p className="mt-1 text-xs text-text-secondary">
              Chunks: {expert.totalChunks} | Rating: {expert.avgRating}
              {expert.modelTier && (
                <span className="ml-2 text-text-disabled">
                  | {expert.modelTier} / {expert.loopPattern}
                </span>
              )}
            </p>
            <ExpertRow
              expertId={expert.id}
              expertName={expert.name}
              onViewProgress={(job, name) => {
                setPipelineJob(job)
                setPipelineExpertName(name)
              }}
            />

            <div className="mt-3 flex flex-wrap gap-2">
              <Button variant="secondary" size="sm" onClick={() => setUploadTargetId(expert.id)}>
                Upload Transcript
              </Button>
              {/* Feature #1 fix (docs bug list): was permanently disabled. */}
              <Button variant="ghost" size="sm" onClick={() => setCharterTargetId(expert.id)}>
                Edit Charter
              </Button>
              {/* A12: new config edit button */}
              <Button variant="ghost" size="sm" onClick={() => setConfigTargetId(expert.id)}>
                Edit Config
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

      <CreateExpertModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        onCreated={() => queryClient.invalidateQueries({ queryKey: ['admin', 'experts'] })}
      />

      {charterTargetId && (
        <EditCharterModal
          isOpen={charterTargetId !== null}
          onClose={() => setCharterTargetId(null)}
          expertId={charterTargetId}
          expertName={experts?.find((e) => e.id === charterTargetId)?.name ?? ''}
          currentCharter=""
          onSaved={() => queryClient.invalidateQueries({ queryKey: ['admin', 'experts'] })}
        />
      )}

      {/* Pipeline progress modal */}
      <IngestionPipelineModal
        isOpen={pipelineJob !== null}
        onClose={() => setPipelineJob(null)}
        job={pipelineJob}
        expertName={pipelineExpertName}
      />

      {/* A12: Edit Config modal — pre-fills from expert object */}
      {configTargetId && (() => {
        const expert = experts?.find((e) => e.id === configTargetId)
        if (!expert) return null
        return (
          <EditExpertConfigModal
            isOpen
            onClose={() => setConfigTargetId(null)}
            expert={expert}
            onSaved={() => {
              queryClient.invalidateQueries({ queryKey: ['admin', 'experts'] })
              setConfigTargetId(null)
            }}
          />
        )
      })()}
    </div>
  )
}

export const Component = AdminExperts
