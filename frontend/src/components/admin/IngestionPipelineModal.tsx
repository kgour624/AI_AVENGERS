import { Modal } from '@/components/ui/Modal'
import { cn } from '@/utils/cn'
import type { IngestionJob } from '@/api/admin'

// ============================================================
// Stage definitions
// ============================================================
const STAGES = [
  { key: 'chunking',          icon: '\u2702\ufe0f', label: 'Splitting Transcript',    step: 1 },
  { key: 'topic_extraction',  icon: '\ud83c\udff7\ufe0f', label: 'Extracting Topics (LLM)', step: 2 },
  { key: 'charter_extraction',icon: '\ud83d\udcdc', label: 'Extracting Charter (LLM)', step: 3 },
  { key: 'embedding',         icon: '\ud83e\udde0', label: 'Generating Embeddings',    step: 4 },
  { key: 'storing',           icon: '\ud83d\uddc4\ufe0f', label: 'Saving to Database',      step: 5 },
  { key: 'smoke_test',        icon: '\ud83d\udd2c', label: 'Running Smoke Test',       step: 6 },
] as const

const STAGE_ORDER: Record<string, number> = {
  pending: 0, chunking: 1, topic_extraction: 2, charter_extraction: 3,
  embedding: 4, storing: 5, smoke_test: 6, complete: 7, failed: -1,
}

type StageStatus = 'waiting' | 'active' | 'done' | 'failed'

function getStageStatus(stageKey: string, currentStage: string, jobStatus: string): StageStatus {
  if (jobStatus === 'failed' && currentStage === stageKey) return 'failed'
  const current = STAGE_ORDER[currentStage] ?? 0
  const mine = STAGE_ORDER[stageKey] ?? 0
  if (mine < current) return 'done'
  if (mine === current) return jobStatus === 'complete' ? 'done' : 'active'
  return 'waiting'
}

// ============================================================
// ETA formatter
// ============================================================
function formatETA(seconds: number | null | undefined): string {
  if (!seconds || seconds <= 0) return ''
  if (seconds < 60) return `~${seconds}s remaining`
  if (seconds < 3600) {
    const m = Math.floor(seconds / 60)
    const s = seconds % 60
    return `~${m}m ${s}s remaining`
  }
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  return `~${h}h ${m}m remaining`
}

// ============================================================
// Cost formatter
// ============================================================
function formatCost(usd: number): string {
  if (!usd || usd === 0) return ''
  const inr = usd * 84
  return `$${usd.toFixed(4)} / \u20b9${inr.toFixed(2)}`
}

// ============================================================
// Progress % calculation
// ============================================================
function calcProgress(job: IngestionJob): number {
  const stageIdx = STAGE_ORDER[job.currentStage ?? 'pending'] ?? 0
  if (job.status === 'complete') return 100
  if (job.status === 'failed') return stageIdx > 0 ? (stageIdx / 7) * 100 : 0
  // Within-stage progress
  const stageBase = ((stageIdx - 1) / 7) * 100
  const stageWidth = (1 / 7) * 100
  if (job.totalChunks > 0 && job.processedChunks > 0) {
    const within = (job.processedChunks / job.totalChunks) * stageWidth
    return Math.min(stageBase + within, 99)
  }
  return Math.max(stageBase, 0)
}

// ============================================================
// Speed calculation
// ============================================================
function calcSpeed(job: IngestionJob): string {
  if (!job.startedAt || job.processedChunks === 0) return ''
  const elapsed = (Date.now() - new Date(job.startedAt).getTime()) / 1000
  if (elapsed <= 0) return ''
  const speed = job.processedChunks / elapsed
  return `${speed.toFixed(1)} chunks/sec`
}

// ============================================================
// Main component
// ============================================================
export interface IngestionPipelineModalProps {
  isOpen: boolean
  onClose: () => void
  job: IngestionJob | null
  expertName: string
}

export function IngestionPipelineModal({
  isOpen, onClose, job, expertName,
}: IngestionPipelineModalProps) {
  if (!job) return null

  const currentStage = job.currentStage ?? 'pending'
  const progress = calcProgress(job)
  const eta = formatETA(job.estimatedSecondsRemaining)
  const cost = formatCost(job.costUsd ?? 0)
  const speed = calcSpeed(job)
  const isDone = job.status === 'complete'
  const isFailed = job.status === 'failed'

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <div className="flex flex-col gap-5">

        {/* Header */}
        <div>
          <h2 className="text-base font-semibold text-text-primary">
            Ingestion Pipeline
          </h2>
          <p className="mt-0.5 text-xs text-text-secondary">{expertName}</p>
          {job.resumedFromCheckpoint && (
            <span className="mt-1 inline-flex items-center gap-1 rounded-full border border-glow-amber/30 bg-glow-amber/10 px-2 py-0.5 text-[10px] font-medium text-glow-amber">
              \u21ba Resumed from checkpoint
            </span>
          )}
        </div>

        {/* Progress bar */}
        <div>
          <div className="mb-1.5 flex items-center justify-between">
            <span className="text-xs text-text-secondary">
              {job.processedChunks > 0
                ? `${job.processedChunks.toLocaleString()} / ${job.totalChunks.toLocaleString()} chunks`
                : job.totalChunks > 0
                ? `${job.totalChunks.toLocaleString()} chunks total`
                : 'Calculating...'}
            </span>
            <span className="text-xs font-medium text-text-primary">{Math.round(progress)}%</span>
          </div>
          <div className="h-2 w-full overflow-hidden rounded-full bg-surface-overlay">
            <div
              className={cn(
                'h-full rounded-full transition-all duration-700 ease-arc',
                isDone
                  ? 'bg-mode-advise shadow-[0_0_8px_oklch(68%_0.18_145_/_0.6)]'
                  : isFailed
                  ? 'bg-mode-refuse'
                  : 'bg-brand shadow-[0_0_8px_oklch(68%_0.28_295_/_0.5)]',
              )}
              style={{ width: `${progress}%` }}
            />
          </div>
          {/* Meta row */}
          <div className="mt-1.5 flex items-center justify-between">
            <div className="flex items-center gap-3">
              {eta && !isDone && !isFailed && (
                <span className="text-[10px] text-text-disabled">{eta}</span>
              )}
              {speed && !isDone && !isFailed && (
                <span className="text-[10px] text-text-disabled">{speed}</span>
              )}
            </div>
            {cost && (
              <span className="rounded-full border border-glow-amber/30 bg-glow-amber/10 px-2 py-0.5 text-[10px] font-medium text-glow-amber">
                {cost}
              </span>
            )}
          </div>
        </div>

        {/* Stage pipeline */}
        <div className="flex flex-col gap-2">
          {STAGES.map((stage, idx) => {
            const status = getStageStatus(stage.key, currentStage, job.status)
            const isActive = status === 'active'
            const isDoneStage = status === 'done'
            const isFailedStage = status === 'failed'

            return (
              <div
                key={stage.key}
                className={cn(
                  'flex items-start gap-3 rounded-lg border p-3 transition-all duration-300',
                  isDoneStage && 'border-mode-advise/20 bg-mode-advise/5',
                  isActive && 'border-brand/40 bg-brand/5 shadow-[0_0_12px_oklch(68%_0.28_295_/_0.1)]',
                  isFailedStage && 'border-mode-refuse/30 bg-mode-refuse/5',
                  status === 'waiting' && 'border-surface-border bg-surface-raised/30 opacity-50',
                )}
              >
                {/* Stage icon / status indicator */}
                <div className="relative flex-shrink-0">
                  <div
                    className={cn(
                      'flex h-8 w-8 items-center justify-center rounded-full text-sm',
                      isDoneStage && 'bg-mode-advise/20 text-mode-advise',
                      isActive && 'bg-brand/20 text-brand',
                      isFailedStage && 'bg-mode-refuse/20 text-mode-refuse',
                      status === 'waiting' && 'bg-surface-overlay text-text-disabled',
                    )}
                  >
                    {isDoneStage ? '\u2713' : isFailedStage ? '\u2717' : stage.icon}
                  </div>
                  {/* Active pulse ring */}
                  {isActive && (
                    <div className="absolute inset-0 rounded-full border-2 border-brand/60 arc-avatar-ring-analyzing" />
                  )}
                </div>

                {/* Stage info */}
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-medium text-text-primary">
                      {stage.step}/6 {stage.label}
                    </span>
                    {isActive && (
                      <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-brand" />
                    )}
                  </div>
                  {/* Stage detail (only for active/failed) */}
                  {(isActive || isFailedStage) && job.stageDetail && (
                    <p className="mt-0.5 text-[10px] text-text-secondary">{job.stageDetail}</p>
                  )}
                  {isFailedStage && job.errorMessage && (
                    <p className="mt-0.5 text-[10px] text-mode-refuse">{job.errorMessage}</p>
                  )}
                </div>

                {/* Connector line to next stage */}
                {idx < STAGES.length - 1 && (
                  <div
                    className={cn(
                      'absolute left-[27px] mt-8 h-2 w-0.5',
                      isDoneStage ? 'bg-mode-advise/40' : 'bg-surface-border',
                    )}
                    style={{ top: `${idx * 56 + 48}px` }}
                  />
                )}
              </div>
            )
          })}
        </div>

        {/* Final status */}
        {isDone && (
          <div className="rounded-lg border border-mode-advise/30 bg-mode-advise/10 p-3 text-center">
            <p className="text-sm font-medium text-mode-advise">\u2705 Ingestion Complete</p>
            <p className="mt-0.5 text-xs text-text-secondary">
              Expert is now {job.status === 'complete' ? 'trained and publicly visible' : 'in draft — upload more transcripts'}
            </p>
          </div>
        )}
        {isFailed && (
          <div className="rounded-lg border border-mode-refuse/30 bg-mode-refuse/10 p-3 text-center">
            <p className="text-sm font-medium text-mode-refuse">\u274c Ingestion Failed</p>
            <p className="mt-0.5 text-xs text-text-secondary">
              {job.errorMessage || 'Check server logs for details. You can retry by uploading the transcript again.'}
            </p>
          </div>
        )}
      </div>
    </Modal>
  )
}
