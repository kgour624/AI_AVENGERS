import { useEffect, useRef, useState } from 'react'
import type { IngestionJob } from '@/api/admin'
import { useAuthStore } from '@/stores/authStore'
import { camelizeKeys } from '@/utils/casing'

export type StreamEvent =
  | { type: 'update' | 'complete' | 'failed'; job: IngestionJob; ts: string }
  | { type: 'heartbeat'; ts: string }
  | { type: 'waiting'; ts: string }
  | { type: 'connecting' }
  | { type: 'error'; message: string; ts: string }

export interface IngestionStreamState {
  job: IngestionJob | null
  lastEvent: StreamEvent | null
  isConnected: boolean
  isDone: boolean   // true when complete or failed
  error: string | null
  eventLog: Array<{ ts: string; message: string; type: string }>
}

/**
 * useIngestionStream — real-time SSE-based ingestion job monitor.
 *
 * Opens an SSE connection to /admin/experts/:id/jobs/stream.
 * Pushes live updates every 1 second from the backend.
 * No polling, no refresh needed.
 *
 * Every event (including errors) is logged to eventLog with timestamp
 * so admin can see exactly what happened and when.
 */
export function useIngestionStream(expertId: string | null): IngestionStreamState {
  const [state, setState] = useState<IngestionStreamState>({
    job: null,
    lastEvent: null,
    isConnected: false,
    isDone: false,
    error: null,
    eventLog: [],
  })

  const esRef = useRef<EventSource | null>(null)
  const retryRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (!expertId) return

    const apiBase = import.meta.env.VITE_API_URL ?? ''
    const url = `${apiBase}/api/v1/admin/experts/${expertId}/jobs/stream`

    const addLog = (type: string, message: string, ts?: string) => {
      const timestamp = ts ?? new Date().toISOString()
      setState((prev) => ({
        ...prev,
        eventLog: [
          { ts: timestamp, type, message },
          ...prev.eventLog,
        ].slice(0, 200), // keep last 200 events
      }))
    }

    const connect = () => {
      // Close any existing connection
      if (esRef.current) {
        esRef.current.close()
        esRef.current = null
      }

      setState((prev) => ({ ...prev, isConnected: false, lastEvent: { type: 'connecting' } }))
      addLog('connecting', 'Opening SSE connection...')

      // EventSource cannot send Authorization headers.
      // Pass access token as ?token= query param.
      // Backend AuthMiddleware reads this as fallback for SSE endpoints.
      // Token read at connect() time so reconnects always use latest token.
      const accessToken = useAuthStore.getState().accessToken ?? ''
      const sseUrl = accessToken ? `${url}?token=${encodeURIComponent(accessToken)}` : url

      const es = new EventSource(sseUrl, { withCredentials: true })
      esRef.current = es

      es.onopen = () => {
        setState((prev) => ({ ...prev, isConnected: true, error: null }))
        addLog('connected', 'SSE connection established')
      }

      es.onmessage = (event) => {
        try {
          const raw = JSON.parse(event.data)
          const data = camelizeKeys<StreamEvent>(raw)

          if (data.type === 'heartbeat') {
            // Heartbeat — connection alive, no state change needed
            return
          }

          if (data.type === 'waiting') {
            addLog('waiting', 'Waiting for ingestion job to start...', (data as {ts:string}).ts)
            return
          }

          if (data.type === 'update' || data.type === 'complete' || data.type === 'failed') {
            const jobData = (data as { type: string; job: IngestionJob; ts: string })
            const job = jobData.job
            const isDone = data.type === 'complete' || data.type === 'failed'

            // Build human-readable log entry
            let logMsg = ''
            if (data.type === 'complete') {
              logMsg = `✅ Complete — ${job.totalChunks} chunks stored`
            } else if (data.type === 'failed') {
              logMsg = `❌ FAILED: ${job.errorMessage || 'Unknown error'}`
            } else {
              const pct = job.totalChunks > 0
                ? Math.round((job.processedChunks / job.totalChunks) * 100)
                : 0
              logMsg = `${job.currentStage ?? job.status} — ${job.stageDetail || `${pct}%`}`
              if (job.costUsd && job.costUsd > 0) {
                logMsg += ` — $${job.costUsd.toFixed(4)}`
              }
            }

            addLog(data.type, logMsg, jobData.ts)

            setState((prev) => ({
              ...prev,
              job,
              lastEvent: data,
              isDone,
              error: data.type === 'failed' ? (job.errorMessage || 'Ingestion failed') : null,
            }))

            // Close connection when done
            if (isDone) {
              es.close()
              esRef.current = null
            }
          }
        } catch (parseErr) {
          addLog('parse_error', `Failed to parse SSE event: ${event.data}`)
        }
      }

      es.onerror = () => {
        const msg = 'SSE connection lost. Reconnecting in 3s...'
        addLog('error', msg)
        setState((prev) => ({
          ...prev,
          isConnected: false,
          error: prev.isDone ? prev.error : msg,
        }))

        es.close()
        esRef.current = null

        // Auto-reconnect after 3s (unless job is done)
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
