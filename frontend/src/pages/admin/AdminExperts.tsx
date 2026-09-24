import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getAdminExperts, regenerateCharter } from '@/api/admin'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { TranscriptUploadModal } from '@/components/admin/TranscriptUploadModal'
import { CreateExpertModal } from '@/components/admin/CreateExpertModal'
import { EditCharterModal } from '@/components/admin/EditCharterModal'
import { EditExpertConfigModal } from '@/components/admin/EditExpertConfigModal'
import { IngestionPipelineModal } from '@/components/admin/IngestionPipelineModal'
import { ExpertCapabilitiesTable } from '@/components/admin/ExpertCapabilitiesTable'
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
  paused: 'Paused — waiting for admin action',
}

// RegenFeedback renders the regenerate-charter result under the button.
// WHY a component (not inline JSX): the inline `regenFeedback[id] && (...).ok`
// pattern does not narrow inside the nested template literal, which TS flags as
// "possibly undefined". A typed prop narrows cleanly.
function RegenFeedback({ feedback }: { feedback?: { ok: boolean; msg: string } }) {
  if (!feedback) return null
  return (
    <p className={`text-[10px] ${feedback.ok ? 'text-green-400' : 'text-red-400'}`}>
      {feedback.msg}
    </p>
  )
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

  // FIX: paused added — was missing, so "View Progress" never showed for paused jobs
  const isActive = latestJob.status === 'running' || latestJob.status === 'pending' || latestJob.status === 'paused'
  const stageLabel = STAGE_LABELS[latestJob.currentStage ?? latestJob.status] ?? latestJob.status

  return (
    <div className="mt-1 flex items-center justify-between">
      <p className="text-xs text-text-secondary">
        {latestJob.status === 'complete' ? '\u2705' : latestJob.status === 'failed' ? '\u274c' : latestJob.status === 'paused' ? '\u23f8' : '\u23f3'}{' '}
        {stageLabel}
        {latestJob.processedChunks > 0 && (
          <span className="ml-1 text-text-disabled">
            ({latestJob.processedChunks}/{latestJob.totalChunks})
          </span>
        )}
        {(latestJob.costUsd ?? 0) > 0 && (
          <span className="ml-2 text-glow-amber/70">${(latestJob.costUsd ?? 0).toFixed(3)}</span>
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
  // Regenerate charter loading state: expertId -> true while in-flight
  const [regeneratingIds, setRegeneratingIds] = useState<Set<string>>(new Set())
  // Regenerate charter success/error feedback: expertId -> message
  const [regenFeedback, setRegenFeedback] = useState<Record<string, { ok: boolean; msg: string }>>({})

  const handleRegenerateCharter = async (expertId: string) => {
    setRegeneratingIds((prev) => new Set(prev).add(expertId))
    setRegenFeedback((prev) => ({ ...prev, [expertId]: { ok: true, msg: '' } }))
    try {
      const result = await regenerateCharter(expertId)
      setRegenFeedback((prev) => ({
        ...prev,
        [expertId]: { ok: true, msg: result.message ?? 'Charter regeneration started.' },
      }))
      // Poll experts list after 35s so the UI reflects the new charter
      setTimeout(() => {
        queryClient.invalidateQueries({ queryKey: ['admin', 'experts'] })
        setRegenFeedback((prev) => {
          const next = { ...prev }
          delete next[expertId]
          return next
        })
      }, 35_000)
    } catch (err: unknown) {
      const msg =
        err instanceof Error ? err.message : 'Charter regeneration failed. Check API credits.'
      setRegenFeedback((prev) => ({ ...prev, [expertId]: { ok: false, msg } }))
    } finally {
      setRegeneratingIds((prev) => {
        const next = new Set(prev)
        next.delete(expertId)
        return next
      })
    }
  }

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
              onViewProgress={(id, name) => {
                setPipelineExpertId(id)
                setPipelineExpertName(name)
              }}
            />

            {/* Feature #26: Capabilities Overview */}
            <ExpertCapabilitiesTable expert={expert} />

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
              {/* Regenerate Charter: shown when charter is blank or training failed.
                  One-click fix for experts whose LLM credits ran out during ingestion.
                  Calls POST /admin/experts/:id/regenerate-charter — background job,
                  no chunk re-processing. Auto-refreshes expert list after 35s. */}
              {(!expert.reasoningCharter || expert.reasoningCharter.trim() === '') && (
                <div className="flex flex-col gap-1">
                  <Button
                    variant="secondary"
                    size="sm"
                    disabled={regeneratingIds.has(expert.id)}
                    onClick={() => handleRegenerateCharter(expert.id)}
                    className="border-amber-500/50 text-amber-400 hover:bg-amber-500/10"
                  >
                    {regeneratingIds.has(expert.id) ? '\u23f3 Generating...' : '\u26a1 Regenerate Charter'}
                  </Button>
                  <RegenFeedback feedback={regenFeedback[expert.id]} />
                </div>
              )}
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
          // BUG FIX (2026-09-08): key={charterTargetId} forces React to
          // unmount+remount a fresh EditCharterModal instance whenever a
          // DIFFERENT expert's "Edit Charter" is clicked, instead of
          // reusing the same instance with just-updated props. This is
          // a belt-and-suspenders safety net alongside the modal's own
          // useEffect fix — either one alone fixes the blank-textarea
          // bug, but together they make the component robust even if
          // one of the two fixes is ever reverted independently.
          key={charterTargetId}
          isOpen={charterTargetId !== null}
          onClose={() => setCharterTargetId(null)}
          expertId={charterTargetId}
          expertName={experts?.find((e) => e.id === charterTargetId)?.name ?? ''}
          // Fix (2026-09-08): was hardcoded to "" — backend now returns
          // reasoningCharter on GET /admin/experts (see types/expert.ts),
          // so the modal opens pre-filled with the real saved charter
          // instead of always looking empty even after a successful
          // training run.
          currentCharter={experts?.find((e) => e.id === charterTargetId)?.reasoningCharter ?? ''}
          onSaved={() => queryClient.invalidateQueries({ queryKey: ['admin', 'experts'] })}
        />
      )}

      {/* Pipeline progress modal — SSE-driven, no polling */}
      <IngestionPipelineModal
        isOpen={pipelineExpertId !== null}
        onClose={() => { setPipelineExpertId(null); setPipelineExpertName('') }}
        expertId={pipelineExpertId}
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
