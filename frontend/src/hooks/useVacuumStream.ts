import { useEffect, useRef, useState } from 'react'
import { useAuthStore } from '@/stores/authStore'
import { camelizeKeys } from '@/utils/casing'

/**
 * useVacuumStream — live per-stage progress for one vacuum (file_jobs) job.
 *
 * WHY EventSource and not the fetch()+ReadableStream pattern used by
 * useSSEStream: this stream is a read-only server push (GET) with no request
 * body, which is exactly EventSource's use case, and EventSource gives us free
 * auto-reconnect semantics. The one thing it cannot do is set an Authorization
 * header, so — exactly like useKanbanStream/useFileStream — the access token is
 * passed as a `?token=` query param, which AuthMiddleware accepts for SSE.
 *
 * Refresh-safe recovery: because the backend seeds every new connection with the
 * job's current status/phase read from the DB, a page refresh simply opens a new
 * EventSource against the same job id and picks the live run back up — no blank
 * screen and no need to re-trigger anything.
 */
export interface VacuumProgress {
  jobId: string
  phase: number
  stage: string
  status: string
  percent: number
  message?: string
  done: number
  total: number
  at: number
}

const TERMINAL_STAGES = new Set(['done', 'failed', 'canceled', 'quarantined'])

export function useVacuumStream(jobId: string | null | undefined, onEvent?: (p: VacuumProgress) => void) {
  const [progress, setProgress] = useState<VacuumProgress | null>(null)
  const [connected, setConnected] = useState(false)
  // ended = the job reached a terminal stage (done/failed/canceled/quarantined)
  // and this stream will never deliver anything again. Deliberately NOT the same
  // as !connected: that also means "still retrying to connect", and the UI shows
  // a very different label for the two states.
  const [ended, setEnded] = useState(false)
  const esRef = useRef<EventSource | null>(null)
  const onEventRef = useRef(onEvent)
  onEventRef.current = onEvent

  useEffect(() => {
    if (!jobId) {
      setProgress(null)
      setConnected(false)
      setEnded(false)
      return
    }
    let closed = false
    // Once a terminal event lands, the stream is finished for good. Without this
    // flag the browser's onerror (it fires when we close the EventSource, and
    // when the server ends the response right after a terminal seed) kept
    // scheduling reconnects forever against a job with nothing left to send —
    // the badge then sat on "connecting…" although the run was already over.
    let terminalSeen = false
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null

    const connect = () => {
      if (closed) return
      const token = useAuthStore.getState().accessToken ?? ''
      const base = `${import.meta.env.VITE_API_URL}/api/v1/admin/vacuum/jobs/${jobId}/stream`
      const url = token ? `${base}?token=${encodeURIComponent(token)}` : base
      const es = new EventSource(url, { withCredentials: true })
      esRef.current = es

      es.onopen = () => setConnected(true)
      es.onmessage = (event) => {
        try {
          const parsed = camelizeKeys<VacuumProgress>(JSON.parse(event.data))
          setProgress(parsed)
          onEventRef.current?.(parsed)
          if (TERMINAL_STAGES.has(parsed.stage)) {
            terminalSeen = true
            setEnded(true)
            es.close()
            esRef.current = null
            setConnected(false)
          }
        } catch {
          // Keepalive comments and partial frames are ignored.
        }
      }
      es.onerror = () => {
        setConnected(false)
        es.close()
        esRef.current = null
        if (!closed && !terminalSeen) reconnectTimer = setTimeout(connect, 3000)
      }
    }

    connect()
    return () => {
      closed = true
      if (reconnectTimer) clearTimeout(reconnectTimer)
      esRef.current?.close()
      esRef.current = null
      setConnected(false)
    }
  }, [jobId])

  return { progress, connected, ended }
}
