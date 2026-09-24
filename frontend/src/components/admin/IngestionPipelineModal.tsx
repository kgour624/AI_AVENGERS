import { useEffect, useRef, useState } from 'react'
import { Modal } from '@/components/ui/Modal'
import { cn } from '@/utils/cn'
import { useIngestionStream } from '@/hooks/useIngestionStream'
import { resumeIngestionJob, retryIngestionJob } from '@/api/admin'

// ============================================================
// Stage definitions
// ============================================================
const STAGES = [
  { key: 'chunking',           icon: '\u2702\ufe0f', label: 'Splitting Transcript',    step: 1 },
  { key: 'topic_extraction',   icon: '\ud83c\udff7\ufe0f', label: 'Extracting Topics (LLM)', step: 2 },
  { key: 'charter_extraction', icon: '\ud83d\udcdc', label: 'Extracting Charter (LLM)', step: 3 },
  { key: 'embedding',          icon: '\ud83e\udde0', label: 'Generating Embeddings',    step: 4 },
  { key: 'storing',            icon: '\ud83d\uddc4\ufe0f', label: 'Saving to Database',      step: 5 },
  { key: 'smoke_test',         icon: '\ud83d\udd2c', label: 'Running Smoke Test',       step: 6 },
] as const

const STAGE_ORDER: Record<string, number> = {
  pending: 0, chunking: 1, topic_extraction: 2, charter_extraction: 3,
  embedding: 4, storing: 5, smoke_test: 6, complete: 7, failed: -1,
  // FIX: paused was missing — getStageStatus returned wrong result for paused jobs
  paused: -1,
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

function formatETA(seconds: number | null | undefined): string {
  if (!seconds || seconds <= 0) return ''
  if (seconds < 60) return `~${seconds}s`
  if (seconds < 3600) return `~${Math.floor(seconds / 60)}m ${seconds % 60}s`
  return `~${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`
}

function formatCost(usd: number): string {
  if (!usd || usd === 0) return ''
  return `$${usd.toFixed(4)} / \u20b9${(usd * 84).toFixed(2)}`
}

function calcProgress(stage: string, processed: number, total: number, status: string): number {
  if (status === 'complete') return 100
  if (status === 'failed') {
    const idx = STAGE_ORDER[stage] ?? 0
    return idx > 0 ? Math.max((idx / 7) * 100, 5) : 5
  }
  const stageIdx = STAGE_ORDER[stage] ?? 0
  const stageBase = ((stageIdx - 1) / 7) * 100
  const stageWidth = (1 / 7) * 100
  if (total > 0 && processed > 0) {
    return Math.min(stageBase + (processed / total) * stageWidth, 99)
  }
  return Math.max(stageBase, 0)
}

// Timestamp formatter: HH:MM:SS.mmm
function fmtTs(iso: string): string {
  try {
    const d = new Date(iso)
    const hh = d.getHours().toString().padStart(2, '0')
    const mm = d.getMinutes().toString().padStart(2, '0')
    const ss = d.getSeconds().toString().padStart(2, '0')
    const ms = d.getMilliseconds().toString().padStart(3, '0')
    return `${hh}:${mm}:${ss}.${ms}`
  } catch {
    return iso.slice(11, 23)
  }
}

// ============================================================
// Main component
// ============================================================
export interface IngestionPipelineModalProps {
  isOpen: boolean
  onClose: () => void
  expertId: string | null
  expertName: string
}

// describeIngestError turns a backend ingestion error into something an admin
// can act on. WHY: the raw strings are precise but obscure ("empty content in
// response"); these hints name the actual fix instead of making the admin guess.
function describeIngestError(msg: string): string {
  if (msg.includes('TRANSCRIPT_REQUIRED')) {
    return 'Transcript was not stored for this job. Re-upload the transcript file to restart ingestion.'
  }
  if (msg.includes('token budget') || msg.includes('empty content')) {
    return 'The model returned no visible answer — its token budget was spent on internal reasoning. Raise max_tokens (or choose a non-reasoning model), then retry.'
  }
  if (msg.includes('NOT_RESUMABLE') || msg.includes('JOB_NOT_PAUSED')) {
    return 'This job is not in a resumable state. Refresh the pipeline status and try again.'
  }
  if (msg.includes('status 402')) {
    return 'The LLM provider reports insufficient credits. Top up the account, then retry.'
  }
  if (msg.includes('status 401') || msg.includes('status 403')) {
    return 'The LLM API key was rejected. Fix the provider key in settings, then retry.'
  }
  return msg
}

export function IngestionPipelineModal({
  isOpen, onClose, expertId, expertName,
}: IngestionPipelineModalProps) {
  const stream = useIngestionStream(isOpen ? expertId : null)
  const logRef = useRef<HTMLDivElement>(null)
  const [isResuming, setIsResuming] = useState(false)
  const [resumeError, setResumeError] = useState<string | null>(null)

  // Auto-scroll log to top (newest first)
  useEffect(() => {
    if (logRef.current) {
      logRef.current.scrollTop = 0
    }
  }, [stream.eventLog.length])

  const job = stream.job
  const currentStage = job?.currentStage ?? 'pending'
  const progress = job
    ? calcProgress(currentStage, job.processedChunks, job.totalChunks, job.status)
    : 0
  const isDone = job?.status === 'complete'
  const isFailed = job?.status === 'failed'
  // FIX: paused was not handled — UI showed "Extracting..." frozen for 30+ min
  const isPaused = job?.status === 'paused'
  const isRunning = !!job && !isDone && !isFailed && !isPaused

  const handleResume = async () => {
    if (!expertId || !job?.id) return
    setIsResuming(true)
    setResumeError(null)
    try {
      await resumeIngestionJob(expertId, job.id)
      // SSE stream will auto-reconnect and show live progress
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })
        ?.response?.data?.error?.message ?? 'Resume failed'
      setResumeError(msg)
    } finally {
      setIsResuming(false)
    }
  }

  // PAUSED jobs use /retry, not /resume. The backend rejects /resume for
  // status='paused' (400 NOT_RESUMABLE), which is why the paused-state button
  // previously did nothing. Same UX, correct endpoint.
  const handleRetry = async () => {
    if (!expertId || !job?.id) return
    setIsResuming(true)
    setResumeError(null)
    try {
      await retryIngestionJob(expertId, job.id)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })
        ?.response?.data?.error?.message ?? 'Retry failed'
      setResumeError(msg)
    } finally {
      setIsResuming(false)
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <div className="flex flex-col gap-4" style={{ minWidth: 480, maxWidth: 560 }}>

        {/* Header */}
        <div className="flex items-start justify-between">
          <div>
            <h2 className="text-base font-semibold text-text-primary">Ingestion Pipeline</h2>
            <p className="mt-0.5 text-xs text-text-secondary">{expertName}</p>
          </div>
          {/* Live indicator */}
          <div className="flex items-center gap-1.5">
            <span className={cn(
              'h-2 w-2 rounded-full',
              stream.isConnected ? 'bg-mode-advise animate-pulse' : 'bg-glow-amber animate-pulse'
            )} />
            <span className="text-[10px] font-medium uppercase tracking-wider text-text-disabled">
              {stream.isConnected ? 'Live' : 'Reconnecting...'}
            </span>
          </div>
        </div>

        {/* Progress bar */}
        {job && (
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
                  'h-full rounded-full transition-all duration-500 ease-arc',
                  isDone ? 'bg-mode-advise shadow-[0_0_8px_oklch(68%_0.18_145_/_0.6)]'
                  : isFailed ? 'bg-mode-refuse'
                  : isPaused ? 'bg-glow-amber shadow-[0_0_8px_oklch(78%_0.18_80_/_0.5)]'
                  : 'bg-brand shadow-[0_0_8px_oklch(68%_0.28_295_/_0.5)]',
                )}
                style={{ width: `${progress}%` }}
              />
            </div>
            <div className="mt-1 flex items-center justify-between">
              <div className="flex gap-3">
                {job.estimatedSecondsRemaining != null && job.estimatedSecondsRemaining > 0 && !isDone && !isFailed && (
                  <span className="text-[10px] text-text-disabled">
                    {formatETA(job.estimatedSecondsRemaining)} remaining
                  </span>
                )}
              </div>
              {job.costUsd && job.costUsd > 0 && (
                <span className="rounded-full border border-glow-amber/30 bg-glow-amber/10 px-2 py-0.5 text-[10px] font-medium text-glow-amber">
                  {formatCost(job.costUsd)}
                </span>
              )}
            </div>
          </div>
        )}

        {/* T2: DB verification — the pipeline's claim checked against the
            actual rows in course_chunks (recorded as a `verified` event). */}
        {stream.verification && (
          <div
            className={cn(
              'rounded-lg border px-3 py-2',
              stream.verification.ok
                ? 'border-mode-advise/30 bg-mode-advise/5'
                : 'border-glow-amber/40 bg-glow-amber/10',
            )}
          >
            <div className="flex items-center justify-between">
              <p className={cn(
                'text-xs font-semibold',
                stream.verification.ok ? 'text-mode-advise' : 'text-glow-amber',
              )}>
                {stream.verification.ok ? '\u2705 Database verified' : '\u26a0\ufe0f Database mismatch'}
              </p>
              <span className="text-[10px] text-text-disabled">
                counted in course_chunks
              </span>
            </div>
            <div className="mt-1 grid grid-cols-3 gap-2 text-[10px] text-text-secondary">
              <span>
                chunks{' '}
                <span className="font-mono text-text-primary">
                  {stream.verification.reality.chunks ?? 0}
                </span>
                {typeof stream.verification.claim.chunks === 'number' &&
                  stream.verification.claim.chunks !== stream.verification.reality.chunks && (
                    <span className="text-glow-amber"> (claimed {stream.verification.claim.chunks})</span>
                  )}
              </span>
              <span>
                topics{' '}
                <span className="font-mono text-text-primary">
                  {stream.verification.reality.topics ?? 0}
                </span>
              </span>
              <span>
                general{' '}
                <span className="font-mono text-text-primary">
                  {stream.verification.reality.general_chunks ?? 0}
                </span>
              </span>
            </div>
            {stream.verification.corpusTotalChunks > 0 && (
              <p className="mt-1 text-[10px] text-text-disabled">
                corpus total after this run:{' '}
                <span className="font-mono">{stream.verification.corpusTotalChunks}</span> chunks
              </p>
            )}
          </div>
        )}

        {/* T3: live batch progress for the parallel stage. The stage cards below
            show WHICH step is active; this shows how many batches are in flight. */}
        {stream.progress && isRunning && (
          <div className="flex items-center justify-between rounded-lg border border-brand/25 bg-brand/5 px-3 py-2">
            <span className="text-[11px] text-text-secondary">
              {stream.progress.stage} {'\u2014'} batch{' '}
              <span className="font-mono text-text-primary">
                {stream.progress.batchesDone}/{stream.progress.batchesTotal}
              </span>
            </span>
            <span className="text-[10px] text-text-disabled">
              {stream.progress.workers} workers
            </span>
          </div>
        )}

        {/* Stage pipeline */}
        <div className="flex flex-col gap-1.5">          {STAGES.map((stage) => {
            const status = job
              ? getStageStatus(stage.key, currentStage, job.status)
              : 'waiting'
            const isActive = status === 'active'
            const isDoneStage = status === 'done'
            const isFailedStage = status === 'failed'

            return (
              <div
                key={stage.key}
                className={cn(
                  'flex items-center gap-3 rounded-lg border px-3 py-2 transition-all duration-300',
                  isDoneStage && 'border-mode-advise/20 bg-mode-advise/5',
                  isActive && 'border-brand/40 bg-brand/5 shadow-[0_0_12px_oklch(68%_0.28_295_/_0.1)]',
                  isFailedStage && 'border-mode-refuse/30 bg-mode-refuse/5',
                  status === 'waiting' && 'border-surface-border bg-surface-raised/30 opacity-40',
                )}
              >
                <div className={cn(
                  'flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full text-xs',
                  isDoneStage && 'bg-mode-advise/20 text-mode-advise',
                  isActive && 'bg-brand/20 text-brand',
                  isFailedStage && 'bg-mode-refuse/20 text-mode-refuse',
                  status === 'waiting' && 'bg-surface-overlay text-text-disabled',
                )}>
                  {isDoneStage ? '\u2713' : isFailedStage ? '\u2717' : stage.icon}
                </div>
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-medium text-text-primary">
                      {stage.step}/6 {stage.label}
                    </span>
                    {isActive && <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-brand" />}
                  </div>
                  {(isActive || isFailedStage) && job?.stageDetail && (
                    <p className="mt-0.5 text-[10px] text-text-secondary">{job.stageDetail}</p>
                  )}
                  {isActive && stream.progress && stream.progress.stage === stage.key && (
                    <p className="mt-0.5 text-[10px] text-text-disabled">
                      batch {stream.progress.batchesDone}/{stream.progress.batchesTotal}
                      {' '}{'\u00b7'} {stream.progress.workers} workers
                    </p>
                  )}
                </div>
              </div>
            )
          })}
        </div>

        {/* Failed state — with Resume button */}
        {isFailed && (
          <div className="flex flex-col gap-2">
            {/* Error details */}
            <div className="rounded-lg border border-mode-refuse/40 bg-mode-refuse/10 p-3">
              <p className="text-xs font-semibold text-mode-refuse">{'\u274c'} Failed</p>
              <p className="mt-1 font-mono text-[11px] text-mode-refuse/90 break-all">
                {job?.errorMessage || stream.error || 'Unknown error'}
              </p>
            </div>

            {/* Resume section */}
            <div className="rounded-lg border border-glow-amber/30 bg-glow-amber/5 p-3">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-xs font-semibold text-glow-amber">{'\u21ba'} Resume Available</p>
                  {job?.currentStage && job.currentStage !== 'pending' && (
                    <p className="mt-0.5 text-[10px] text-text-secondary">
                      Checkpoint: <span className="font-mono text-glow-amber/80">{job.currentStage}</span>
                      {' '}{'\u2014'} no re-upload needed
                    </p>
                  )}
                </div>
                <button
                  onClick={handleResume}
                  disabled={isResuming}
                  className={cn(
                    'rounded-md border border-glow-amber/40 bg-glow-amber/10 px-3 py-1.5',
                    'text-xs font-medium text-glow-amber',
                    'transition-all duration-150 ease-arc',
                    'hover:bg-glow-amber/20 hover:border-glow-amber/60',
                    'disabled:opacity-50 disabled:cursor-not-allowed',
                  )}
                >
                  {isResuming ? 'Resuming...' : 'Resume from checkpoint'}
                </button>
              </div>
              {resumeError && (
                <p className="mt-2 text-[10px] text-mode-refuse">
                  {describeIngestError(resumeError)}
                </p>
              )}
            </div>
          </div>
        )}

        {/* Paused state — FIX: was completely missing, UI showed frozen "Extracting..." */}
        {isPaused && (
          <div className="flex flex-col gap-2">
            <div className="rounded-lg border border-glow-amber/40 bg-glow-amber/10 p-3">
              <p className="text-xs font-semibold text-glow-amber">{'\u23f8'} Paused — charter generation failed</p>
              <p className="mt-1 text-[11px] text-glow-amber/80 break-all">
                {job?.errorMessage || 'Charter LLM failed. Waiting for admin action.'}
              </p>
              <p className="mt-1 text-[10px] text-text-secondary">
                Everything up to this step is saved at checkpoint:{' '}
                <span className="font-mono text-glow-amber/80">{job?.currentStage ?? 'unknown'}</span>
                {' '}{'\u2014'} no re-upload needed.
              </p>
            </div>
            <div className="rounded-lg border border-glow-amber/30 bg-glow-amber/5 p-3">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-xs font-semibold text-glow-amber">{'\u21ba'} Retry Available</p>
                  <p className="mt-0.5 text-[10px] text-text-secondary">
                    Fix the cause (API credits, or the model&apos;s token budget), then retry — chunking
                    and topics are not repeated.
                  </p>
                </div>
                <button
                  onClick={handleRetry}
                  disabled={isResuming}
                  className={cn(
                    'rounded-md border border-glow-amber/40 bg-glow-amber/10 px-3 py-1.5',
                    'text-xs font-medium text-glow-amber',
                    'transition-all duration-150 ease-arc',
                    'hover:bg-glow-amber/20 hover:border-glow-amber/60',
                    'disabled:opacity-50 disabled:cursor-not-allowed',
                  )}
                >
                  {isResuming ? 'Retrying...' : 'Retry now'}
                </button>
              </div>
              {resumeError && (
                <p className="mt-2 text-[10px] text-mode-refuse">
                  {describeIngestError(resumeError)}
                </p>
              )}
            </div>
          </div>
        )}

        {/* Complete box */}
        {isDone && (
          <div className="rounded-lg border border-mode-advise/30 bg-mode-advise/10 p-3 text-center">
            <p className="text-sm font-medium text-mode-advise">{'\u2705'} Ingestion Complete</p>
          </div>
        )}

        {/* Event log — sourced from ingestion_job_events (Postgres), so it is
            identical for every viewer and survives a refresh. */}
        <div>
          <div className="mb-1.5 flex items-center justify-between">
            <p className="text-[10px] font-semibold uppercase tracking-wider text-text-disabled">
              Ingestion Timeline
            </p>
            <span className="text-[10px] text-text-disabled">
              {stream.events.length} events {'\u00b7'} from database
            </span>
          </div>
          <div
            ref={logRef}
            className="h-44 overflow-y-auto rounded-lg border border-surface-border bg-surface-void p-2 font-mono text-[10px]"
          >
            {stream.eventLog.length === 0 ? (
              <p className="text-text-disabled">Waiting for the first event…</p>
            ) : (
              stream.eventLog.map((entry) => (
                <div key={entry.seq} className="flex gap-2 py-0.5">
                  <span className="flex-shrink-0 text-text-disabled">{fmtTs(entry.ts)}</span>
                  <span className={cn(
                    'break-all',
                    entry.type === 'failed' || entry.type === 'error' ? 'text-mode-refuse'
                    : entry.type === 'complete' || entry.type === 'verified' ? 'text-mode-advise'
                    : entry.type === 'paused' || entry.type === 'mismatch' ? 'text-glow-amber'
                    : 'text-text-secondary',
                  )}>
                    {entry.message}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>

      </div>
    </Modal>
  )
}
