import { useEffect, useRef, useState } from 'react'
import type { FileArtifact } from '@/api/workflows'
import { useAuthStore } from '@/stores/authStore'
import { camelizeKeys } from '@/utils/casing'

export interface FileEntry extends FileArtifact {
  expertId: string
}

export interface FileStreamState {
  files: FileEntry[]
  isConnected: boolean
  isDone: boolean
  lastWave: { waveIndex: number; phase: string; expertIds: string[]; mergedFiles: string[] } | null
  eventLog: Array<{ ts: string; type: string; message: string }>
}

/**
 * useFileStream — real-time SSE-based file/code browser updater.
 *
 * Sibling hook to useKanbanStream — same connection lifecycle, same
 * reconnect/auth pattern, different SSE endpoint and event set.
 *
 * SSE endpoint: GET /api/v1/workflows/:id/files/stream
 * Events handled:
 *   file_artifact → expert produced/modified a code file (create|modify)
 *   file_wave     → a wave's code was merged into main/
 *   file_done     → close connection, set isDone=true
 *   file_error    → log error
 *
 * Auto-reconnect: 3s after disconnect (same as useKanbanStream).
 * Auth: ?token= query param (EventSource cannot send headers).
 *
 * WHY per-file "live" but per-wave "batch":
 *   file_artifact fires once per file, right after that expert's Aider
 *   run finishes (not character-by-character — there is no typing
 *   animation in this architecture, see aider_runner.go
 *   publishCodeArtifacts()). file_wave fires once per wave, after
 *   WorkspaceMerger.MergeWave() merges all of that wave's experts into
 *   main/. Both are "as soon as the underlying work completes", not a
 *   simulated live-typing effect.
 */
export function useFileStream(workflowId: string | null): FileStreamState {
  const [state, setState] = useState<FileStreamState>({
    files: [],
    isConnected: false,
    isDone: false,
    lastWave: null,
    eventLog: [],
  })

  const esRef = useRef<EventSource | null>(null)
  const retryRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  // filesMapRef: filePath -> FileEntry
  // WHY ref not state: map mutations don't need to trigger re-renders.
  // Only setState triggers re-render, with a fresh array from the map.
  const filesMapRef = useRef<Map<string, FileEntry>>(new Map())

  useEffect(() => {
    if (!workflowId) return

    const apiBase = import.meta.env.VITE_API_URL ?? ''
    const url = `${apiBase}/api/v1/workflows/${workflowId}/files/stream`

    const addLog = (type: string, message: string) => {
      const ts = new Date().toISOString()
      setState((prev) => ({
        ...prev,
        eventLog: [{ ts, type, message }, ...prev.eventLog].slice(0, 100),
      }))
    }

    const connect = () => {
      if (esRef.current) { esRef.current.close(); esRef.current = null }
      setState((prev) => ({ ...prev, isConnected: false }))
      addLog('connecting', 'Opening SSE connection...')

      // EventSource cannot send Authorization headers.
      // Pass access token as ?token= query param.
      // Backend AuthMiddleware reads this as fallback for SSE endpoints.
      const token = useAuthStore.getState().accessToken ?? ''
      const sseUrl = token ? `${url}?token=${encodeURIComponent(token)}` : url
      const es = new EventSource(sseUrl, { withCredentials: true })
      esRef.current = es

      es.onopen = () => {
        setState((prev) => ({ ...prev, isConnected: true }))
        addLog('connected', 'SSE connection established')
      }

      es.onmessage = (event) => {
        try {
          const raw = JSON.parse(event.data)
          // Backend sends: {type: "file_artifact", data: {...}}
          const parsed = camelizeKeys<{ type: string; data: Record<string, unknown> }>(raw)
          const { type, data } = parsed

          switch (type) {
            case 'file_done': {
              setState((prev) => ({ ...prev, isDone: true, isConnected: false }))
              addLog('done', 'Workflow completed')
              es.close(); esRef.current = null
              break
            }

            case 'file_error': {
              addLog('error', String(data?.message ?? JSON.stringify(data)))
              break
            }

            case 'file_artifact': {
              // code_artifact_produced event
              // data.content = {filename, file_path, content, language,
              //   commit_sha, lines_of_code, phase, validation_passed,
              //   validation_error, operation}
              const expertId = data?.postedByExpertId as string | undefined
              const content = data?.content as Partial<FileArtifact> | undefined
              if (expertId && content?.filePath) {
                const entry: FileEntry = { ...content, expertId } as FileEntry
                filesMapRef.current.set(content.filePath, entry)
                setState((prev) => ({
                  ...prev,
                  files: [...filesMapRef.current.values()],
                }))
                addLog('artifact', `${content.operation ?? 'modify'} ${content.filePath}`)
              }
              break
            }

            case 'file_wave': {
              // wave_completed event
              // data.content = {wave_index, phase, expert_ids, merged_files}
              const content = data?.content as {
                waveIndex?: number
                phase?: string
                expertIds?: string[]
                mergedFiles?: string[]
              } | undefined
              if (content) {
                setState((prev) => ({
                  ...prev,
                  lastWave: {
                    waveIndex: content.waveIndex ?? 0,
                    phase: content.phase ?? '',
                    expertIds: content.expertIds ?? [],
                    mergedFiles: content.mergedFiles ?? [],
                  },
                }))
                addLog('wave', `Wave ${content.waveIndex ?? 0} merged (${content.expertIds?.length ?? 0} experts)`)
              }
              break
            }

            default:
              break
          }
        } catch {
          addLog('parse_error', `Failed to parse: ${event.data.slice(0, 80)}`)
        }
      }

      es.onerror = () => {
        addLog('error', 'SSE disconnected. Reconnecting in 3s...')
        setState((prev) => ({ ...prev, isConnected: false }))
        es.close(); esRef.current = null
        // Auto-reconnect unless workflow is done
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
      esRef.current?.close()
      if (retryRef.current) clearTimeout(retryRef.current)
    }
  }, [workflowId])

  return state
}
