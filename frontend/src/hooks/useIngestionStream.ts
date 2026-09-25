import { useEffect, useRef, useState } from 'react'
import type { IngestionJob, IngestionJobEvent } from '@/api/admin'
import { useAuthStore } from '@/stores/authStore'
import { camelizeKeys, snakeifyKeys } from '@/utils/casing'

export type StreamEvent =
  | { type: 'update' | 'complete' | 'complete_with_warnings' | 'failed' | 'hello'; job: IngestionJob; ts: string }
  | { type: 'event'; event: IngestionJobEvent; ts: string }
  | { type: 'heartbeat'; ts: string }
  | { type: 'waiting'; ts: string }
  | { type: 'connecting' }
  | { type: 'error'; message: string; ts: string }

/** One display row, derived from a durable IngestionJobEvent. */
export interface IngestionLogEntry {
  seq: number
  ts: string
  type: string
  message: string
}

/** Batch-level progress of the active parallel stage (T3). */
export interface IngestionProgress {
  stage: string
  batchesDone: number
  batchesTotal: number
  workers: number
  chunksDone: number
  chunksTotal: number
}

/** Latest DB verification of what was actually stored (T2). */
export interface IngestionVerification {
  ok: boolean
  claim: Record<string, number>
  reality: Record<string, number>
  corpusTotalChunks: number
  sourceFile: string
  /**
   * parsed → duplicates → inserted → reused → stored, from the run's own
   * accounting. This is what explains the gap: `stored 277, parsed 395` looks
   * like loss until the 118 repeats of text already in the corpus are shown.
   */
  breakdown: Record<string, number>
  /** The backend's one-line rendering of the same breakdown. */
  summary: string
}

export interface IngestionStreamState {
  job: IngestionJob | null
  lastEvent: StreamEvent | null
  isConnected: boolean
  isDone: boolean
  error: string | null
  /** Durable timeline, oldest → newest. Backed by the DB, so it survives refresh. */
  events: IngestionJobEvent[]
  /** The same timeline formatted for display, newest first. */
  eventLog: IngestionLogEntry[]
  /** Latest "claim vs reality" record, or null before the store step. */
  verification: IngestionVerification | null
  /** Live batch progress for the active stage, or null. */
  progress: IngestionProgress | null
}

// maxEvents caps client memory. 1000 events is far more than a single run
// produces (a 15k-chunk run emits ~a few hundred) and keeps re-renders cheap.
const maxEvents = 1000

function num(detail: Record<string, unknown>, key: string): number {
  const value = detail[key]
  return typeof value === 'number' ? value : 0
}

function str(detail: Record<string, unknown>, key: string): string {
  const value = detail[key]
  return typeof value === 'string' ? value : ''
}

function numMap(value: unknown): Record<string, number> {
  if (value === null || typeof value !== 'object') return {}
  const out: Record<string, number> = {}
  for (const [k, v] of Object.entries(value as Record<string, unknown>)) {
    if (typeof v === 'number') out[k] = v
  }
  return out
}

/**
 * describeJobEvent turns a durable event row into one human-readable log line.
 *
 * WHY here and not in the modal: the same mapping is used for the live stream
 * and for the history replay, so it must not depend on component state.
 */
export function describeJobEvent(ev: IngestionJobEvent): { type: string; message: string } {
  // The stream camelizes the whole frame (see the onmessage handler below), so
  // by the time an event reaches here every snake_case key in `detail` is
  // camelCase — but every read in the switch below is snake_case. Left as-is,
  // that silently yields 0 for every number (`batches 0/0`, `0ms`, `corpus 0`)
  // and false for every flag (`smoke test failed`). Convert back once here: one
  // fix for all of them, independent of how many keys each branch reads.
  const d = snakeifyKeys<Record<string, unknown>>(ev.detail ?? {})
  switch (ev.kind) {
    case 'run_started': {
      const resumed = d.resumed === true
      const workers = num(d, 'workers')
      if (resumed) {
        return {
          type: 'update',
          message: `\u25b6 Resumed from checkpoint (${str(d, 'resumed_from') || 'unknown'}) \u2014 ${workers} workers`,
        }
      }
      return {
        type: 'update',
        message: `\u25b6 Run started \u2014 ${num(d, 'transcript_chars').toLocaleString()} chars, ${workers} workers`,
      }
    }
    case 'stage_started':
      return { type: 'update', message: `${ev.stage} \u2014 ${str(d, 'detail') || 'started'}` }
    case 'stage_done': {
      const ms = num(d, 'duration_ms')
      const chunks = num(d, 'chunks')
      const fromCheckpoint = d.from_checkpoint === true ? ' (from checkpoint)' : ''
      return {
        type: 'update',
        message: `${ev.stage} done in ${ms}ms${chunks ? ` \u2014 ${chunks} chunks` : ''}${fromCheckpoint}`,
      }
    }
    case 'batch_done':
      return {
        type: 'update',
        message: `${ev.stage} batch ${num(d, 'batches_done')}/${num(d, 'batches_total')} \u2014 ${num(d, 'chunks_done')}/${num(d, 'chunks_total')} chunks (${num(d, 'workers')} workers)`,
      }
    case 'chunk_stored':
      return {
        type: 'update',
        message: `stored ${num(d, 'chunks_done')}/${num(d, 'chunks_total')} chunks`,
      }
    case 'verified': {
      const reality = numMap(d.reality)
      const claim = numMap(d.claim)
      if (d.ok === true) {
        return {
          type: 'verified',
          message: `DB verified \u2014 ${reality.chunks ?? 0} chunks, ${reality.topics ?? 0} topics, ${reality.general_chunks ?? 0} general`,
        }
      }
      return {
        type: 'mismatch',
        message: `DB MISMATCH \u2014 claimed ${claim.chunks ?? 0} chunks/${claim.topics ?? 0} topics, DB has ${reality.chunks ?? 0}/${reality.topics ?? 0} (general ${reality.general_chunks ?? 0}, null embeddings ${reality.null_embeddings ?? 0})`,
      }
    }
    case 'paused':
      return {
        type: 'paused',
        message: `\u23f8 PAUSED \u2014 ${str(d, 'reason') || 'waiting for admin action'}`,
      }
    case 'failed':
      return { type: 'failed', message: `\u274c FAILED \u2014 ${str(d, 'reason') || 'unknown error'}` }
    case 'complete': {
      const ms = num(d, 'duration_ms')
      return {
        type: 'complete',
        message: `\u2705 Complete \u2014 ${num(d, 'chunks_this_run')} chunks this run, corpus ${num(d, 'corpus_total')}, ${(ms / 1000).toFixed(1)}s, smoke test ${d.smoke_test_passed === true ? 'passed' : 'failed'}`,
      }
    }
    default:
      return { type: 'update', message: `${ev.stage} \u2014 ${ev.kind}` }
  }
}

/**
 * useIngestionStream — real-time ingestion monitor.
 *
 * TWO feeds, deliberately:
 *  1. `event` frames — the durable timeline (ingestion_job_events). Pushed the
 *     moment the pipeline writes them; the DB is the source of truth, so the
 *     log survives a refresh and two admins see the identical history.
 *  2. `update`/`complete`/`failed`/`hello` frames — the job row snapshot, used
 *     only for the progress bar / stage cards / ETA / cost.
 *
 * WHY both: events carry the *history*, the snapshot carries the *current*
 * numbers. The event log is never built from snapshots (that was the old
 * implementation, and it is exactly why refreshing erased the log).
 */
export function useIngestionStream(expertId: string | null): IngestionStreamState {
  const [state, setState] = useState<IngestionStreamState>({
    job: null,
    lastEvent: null,
    isConnected: false,
    isDone: false,
    error: null,
    events: [],
    eventLog: [],
    verification: null,
    progress: null,
  })

  const esRef = useRef<EventSource | null>(null)
  const retryRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (!expertId) return

    const apiBase = import.meta.env.VITE_API_URL ?? ''
    const url = `${apiBase}/api/v1/admin/experts/${expertId}/jobs/stream`

    const connect = () => {
      if (esRef.current) {
        esRef.current.close()
        esRef.current = null
      }

      setState((prev) => ({ ...prev, isConnected: false, lastEvent: { type: 'connecting' } }))

      // EventSource cannot send Authorization headers.
      // Pass the access token as ?token= — the backend AuthMiddleware reads it
      // as an SSE-only fallback. Read at connect() time so a reconnect always
      // uses the latest token.
      const accessToken = useAuthStore.getState().accessToken ?? ''
      const sseUrl = accessToken ? `${url}?token=${encodeURIComponent(accessToken)}` : url

      const es = new EventSource(sseUrl, { withCredentials: true })
      esRef.current = es

      es.onopen = () => {
        setState((prev) => ({ ...prev, isConnected: true, error: null }))
      }

      es.onmessage = (event) => {
        let data: StreamEvent
        try {
          data = camelizeKeys<StreamEvent>(JSON.parse(event.data))
        } catch {
          // A malformed frame is not worth surfacing to the admin: the next
          // snapshot (2s) and the safety-net catch-up keep the UI correct.
          return
        }

        if (data.type === 'heartbeat' || data.type === 'waiting') {
          return
        }

        // complete_with_warnings is terminal too: the pipeline has finished and
        // will not write again. Treating it as non-terminal would leave the modal
        // waiting on a run that is already over.
        if (
          data.type === 'hello' ||
          data.type === 'update' ||
          data.type === 'complete' ||
          data.type === 'complete_with_warnings' ||
          data.type === 'failed'
        ) {
          const job = (data as { job: IngestionJob }).job
          const isDone =
            job?.status === 'complete' ||
            job?.status === 'complete_with_warnings' ||
            job?.status === 'failed'
          const isPaused = job?.status === 'paused'
          setState((prev) => ({
            ...prev,
            job,
            lastEvent: data,
            isDone,
            error: job?.status === 'failed' ? job.errorMessage || 'Ingestion failed' : null,
          }))
          // Terminal → stop the stream so we stop reconnecting to a dead run.
          // (The backend also closes; this covers the client side too.)
          if (isDone || isPaused) {
            es.close()
            esRef.current = null
          }
          return
        }

        if (data.type === 'event') {
          const ev = (data as { event: IngestionJobEvent }).event
          if (!ev || typeof ev.sequenceNumber !== 'number') return
          setState((prev) => {
            // Dedupe by sequence_number: replay + live push + the reconnect
            // catch-up can legitimately deliver the same row more than once.
            if (prev.events.some((existing) => existing.sequenceNumber === ev.sequenceNumber)) {
              return prev
            }
            const events = [...prev.events, ev].slice(-maxEvents)
            const entry = describeJobEvent(ev)
            const eventLog = [
              { seq: ev.sequenceNumber, ts: ev.createdAt, type: entry.type, message: entry.message },
              ...prev.eventLog,
            ].slice(0, maxEvents)

            // Same camelCase trap as describeJobEvent (above): the detail keys
            // arrive camelized, so read them through a snake_case view.
            const detail = snakeifyKeys<Record<string, unknown>>(ev.detail ?? {})

            let verification = prev.verification
            if (ev.kind === 'verified') {
              verification = {
                ok: detail.ok === true,
                claim: numMap(detail.claim),
                reality: numMap(detail.reality),
                corpusTotalChunks: num(detail, 'corpus_total_chunks'),
                sourceFile: str(detail, 'source_file'),
                breakdown: numMap(detail.breakdown),
                summary: str(detail, 'summary'),
              }
            }

            let progress = prev.progress
            if (ev.kind === 'batch_done') {
              progress = {
                stage: ev.stage,
                batchesDone: num(detail, 'batches_done'),
                batchesTotal: num(detail, 'batches_total'),
                workers: num(detail, 'workers'),
                chunksDone: num(detail, 'chunks_done'),
                chunksTotal: num(detail, 'chunks_total'),
              }
            }

            return { ...prev, events, eventLog, verification, progress }
          })
        }
      }

      es.onerror = () => {
        setState((prev) => ({
          ...prev,
          isConnected: false,
          error: prev.isDone ? prev.error : 'SSE connection lost. Reconnecting in 3s...',
        }))

        es.close()
        esRef.current = null

        // Auto-reconnect after 3s unless the job already finished.
        setState((prev) => {
          if (!prev.isDone) {
            retryRef.current = setTimeout(connect, 3000)
          }
          return prev
        })
      }
    }

    connect()

    return () => {
      if (esRef.current) {
        esRef.current.close()
        esRef.current = null
      }
      if (retryRef.current) {
        clearTimeout(retryRef.current)
      }
    }
  }, [expertId])

  return state
}
