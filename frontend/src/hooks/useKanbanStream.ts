import { useEffect, useRef, useState } from 'react'
import type { KanbanTask } from '@/api/workflows'
import { useAuthStore } from '@/stores/authStore'
import { camelizeKeys } from '@/utils/casing'

export interface ApprovalGate {
  approvalId: string
  gateName: string
  summary: string
}

export interface KanbanStreamState {
  // isFailed distinguishes a workflow that ENDED from one that SUCCEEDED. The
  // backend closes the kanban stream for both, sending event_type
  // 'workflow_completed' or 'workflow_failed' — previously only isDone was read,
  // so a failed workflow rendered the green "Workflow Complete" bar.
  isFailed: boolean
  tasks: KanbanTask[]
  isConnected: boolean
  isDone: boolean
  approvalGate: ApprovalGate | null
  eventLog: Array<{ ts: string; type: string; message: string }>
}

/**
 * useKanbanStream — real-time SSE-based Kanban board updater.
 *
 * Replaces 5s polling in KanbanPage.tsx with live SSE stream.
 * Same pattern as useIngestionStream — consistency over novelty.
 *
 * SSE endpoint: GET /api/v1/workflows/:id/kanban/stream
 * Events handled:
 *   kanban_plan     → build initial task list from plan
 *   kanban_task     → update task status in map
 *   kanban_artifact → update task produced_artifact
 *   kanban_approval → log approval gate
 *   kanban_done     → close connection, set isDone=true
 *   kanban_error    → log error
 *
 * Auto-reconnect: 3s after disconnect (same as useIngestionStream).
 * Auth: ?token= query param (EventSource cannot send headers).
 */
export function useKanbanStream(workflowId: string | null): KanbanStreamState {
  const [state, setState] = useState<KanbanStreamState>({
    tasks: [],
    isConnected: false,
    isDone: false,
    isFailed: false,
    approvalGate: null,
    eventLog: [],
  })

  const esRef = useRef<EventSource | null>(null)
  const retryRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  // tasksMapRef: assignedExpertId -> KanbanTask
  // WHY ref not state: map mutations don't need to trigger re-renders.
  // Only setState triggers re-render, with a fresh array from the map.
  const tasksMapRef = useRef<Map<string, KanbanTask>>(new Map())

  useEffect(() => {
    if (!workflowId) return

    const apiBase = import.meta.env.VITE_API_URL ?? ''
    const url = `${apiBase}/api/v1/workflows/${workflowId}/kanban/stream`

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
          // Backend sends: {type: "kanban_task", data: {...}}
          const parsed = camelizeKeys<{ type: string; data: Record<string, unknown> }>(raw)
          const { type, data } = parsed

          switch (type) {
            case 'kanban_done': {
              // data.eventType is 'workflow_failed' when the run failed. Without
              // reading it, a failure was announced to the screen as completion.
              const failed = String(data?.eventType ?? '') === 'workflow_failed'
              setState((prev) => ({
                ...prev,
                isDone: true,
                isFailed: failed,
                isConnected: false,
                approvalGate: null,
              }))
              addLog(failed ? 'error' : 'done', failed ? 'Workflow failed' : 'Workflow completed')
              es.close(); esRef.current = null
              break
            }

            case 'kanban_error': {
              addLog('error', String(data?.message ?? JSON.stringify(data)))
              break
            }

            case 'kanban_plan': {
              // task_plan_ready event: build initial task list
              // data.content.tasks = [{expert_id, expert_name, title, description}, ...]
              const content = data?.content as {
                tasks?: Array<{
                  expertId: string
                  expertName: string
                  title: string
                  description: string
                }>
              } | undefined
              const planTasks = content?.tasks ?? []
              tasksMapRef.current.clear()
              planTasks.forEach((t, i) => {
                const task: KanbanTask = {
                  id: `plan-${i}-${t.expertId}`,
                  workflowId,
                  assignedExpertId: t.expertId,
                  expertName: t.expertName,
                  domain: '',
                  title: t.title,
                  description: t.description,
                  status: 'todo',
                  costUsd: 0,
                  updatedAt: new Date().toISOString(),
                }
                tasksMapRef.current.set(t.expertId, task)
              })
              setState((prev) => ({
                ...prev,
                tasks: [...tasksMapRef.current.values()],
              }))
              addLog('plan', `${planTasks.length} tasks planned`)
              break
            }

            case 'kanban_task': {
              // task_status_changed event
              // data.content = {expert_id, status}
              const content = data?.content as {
                expertId?: string
                status?: string
              } | undefined
              const expertId = content?.expertId ?? (data?.postedByExpertId as string | undefined)
              const status = content?.status
              if (expertId && status) {
                const existing = tasksMapRef.current.get(expertId)
                if (existing) {
                  const updated: KanbanTask = {
                    ...existing,
                    status: status as KanbanTask['status'],
                    updatedAt: new Date().toISOString(),
                  }
                  tasksMapRef.current.set(expertId, updated)
                  setState((prev) => ({
                    ...prev,
                    tasks: [...tasksMapRef.current.values()],
                  }))
                }
              }
              addLog('task', `${expertId ?? 'unknown'} → ${status ?? 'unknown'}`)
              break
            }

            case 'kanban_artifact': {
              // artifact posted by an expert
              const expertId = data?.postedByExpertId as string | undefined
              const eventType = data?.eventType as string | undefined
              if (expertId) {
                const existing = tasksMapRef.current.get(expertId)
                if (existing && existing.status !== 'done') {
                  const updated: KanbanTask = {
                    ...existing,
                    status: 'under_review',
                    updatedAt: new Date().toISOString(),
                  }
                  tasksMapRef.current.set(expertId, updated)
                  setState((prev) => ({
                    ...prev,
                    tasks: [...tasksMapRef.current.values()],
                  }))
                }
              }
              addLog('artifact', `${eventType ?? 'artifact'} by ${expertId ?? 'unknown'}`)
              break
            }

            case 'kanban_approval': {
              // Parse approval gate details so KanbanPage can render Approve/Reject buttons.
              const content = data?.content as {
                approvalId?: string
                gateName?: string
                summary?: string
              } | undefined
              const approvalId = content?.approvalId ?? ''
              const gateName = content?.gateName ?? 'approval'
              const summary = content?.summary ?? 'Review and approve to continue.'
              if (approvalId) {
                setState((prev) => ({
                  ...prev,
                  approvalGate: { approvalId, gateName, summary },
                }))
              }
              addLog('approval', `Gate: ${gateName}`)
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
