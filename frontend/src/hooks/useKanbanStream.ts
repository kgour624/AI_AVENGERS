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
  isFailed: boolean
  tasks: KanbanTask[]
  isConnected: boolean
  isDone: boolean
  approvalGate: ApprovalGate | null
  eventLog: Array<{ ts: string; type: string; message: string }>
  isAwaitingApprovalId: boolean // naya
}

export function useKanbanStream(workflowId: string | null, workflowStatus?: string): KanbanStreamState {
  const [state, setState] = useState<KanbanStreamState>({
    tasks: [], isConnected: false, isDone: false, isFailed: false,
    approvalGate: null, eventLog: [], isAwaitingApprovalId: false
  })
  const esRef = useRef<EventSource | null>(null)
  const retryRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const tasksMapRef = useRef<Map<string, KanbanTask>>(new Map())

  useEffect(() => {
    if (!workflowId) return
    const apiBase = import.meta.env.VITE_API_URL ?? ''
    const url = `${apiBase}/api/v1/workflows/${workflowId}/kanban/stream`
    const addLog = (type: string, message: string) => {
      const ts = new Date().toISOString()
      setState(prev => ({ ...prev, eventLog: [{ ts, type, message }, ...prev.eventLog].slice(0, 100) }))
    }
    const connect = () => {
      if (esRef.current) { esRef.current.close(); esRef.current = null }
      // FIX: isConnected false karo par approvalGate ko preserve rakho - null mat karo
      setState(prev => ({ ...prev, isConnected: false }))
      addLog('connecting', 'Opening SSE connection...')
      const token = useAuthStore.getState().accessToken ?? ''
      const sseUrl = token ? `${url}?token=${encodeURIComponent(token)}` : url
      const es = new EventSource(sseUrl, { withCredentials: true })
      esRef.current = es
      es.onopen = () => {
        setState(prev => ({ ...prev, isConnected: true }))
        addLog('connected', 'SSE connection established')
      }
      es.onmessage = (event) => {
        try {
          const raw = JSON.parse(event.data)
          const parsed = camelizeKeys<{ type: string; data: Record<string, unknown> }>(raw)
          const { type, data } = parsed
          switch (type) {
            case 'kanban_done': {
              const failed = String(data?.eventType ?? '') === 'workflow_failed'
              setState(prev => ({ ...prev, isDone: true, isFailed: failed, isConnected: false, approvalGate: null, isAwaitingApprovalId: false }))
              addLog(failed ? 'error' : 'done', failed ? 'Workflow failed' : 'Workflow completed')
              es.close(); esRef.current = null
              break
            }
            case 'kanban_approval': {
              // FIX: normalize all variants: approvalId | approval_id | id
              const content = data?.content as Record<string, any> | undefined
              // camelizeKeys already converts approval_id->approvalId, but handle both + raw id
              const approvalId = String(content?.approvalId ?? content?.approval_id ?? content?.id ?? (data as any)?.approvalId ?? '').trim()
              const gateName = String(content?.gateName ?? content?.gate_name ?? 'approval')
              const summary = String(content?.summary ?? 'Review and approve to continue.')
              if (approvalId) {
                setState(prev => ({ ...prev, approvalGate: { approvalId, gateName, summary }, isAwaitingApprovalId: false }))
              } else {
                // id abhi tak nahi aaya par workflow paused hai -> loading state
                setState(prev => ({ ...prev, isAwaitingApprovalId: workflowStatus === 'paused_for_approval' }))
              }
              addLog('approval', `Gate: ${gateName} id=${approvalId || 'pending'}`)
              break
            }
            // ... kanban_plan, kanban_task, kanban_artifact wale cases same rakho ...
            case 'kanban_plan': {
              const content = data?.content as { tasks?: Array<{ expertId: string; expertName: string; title: string; description: string }> } | undefined
              const planTasks = content?.tasks ?? []
              tasksMapRef.current.clear()
              planTasks.forEach((t, i) => {
                tasksMapRef.current.set(t.expertId, {
                  id: `plan-${i}-${t.expertId}`, workflowId, assignedExpertId: t.expertId,
                  expertName: t.expertName, domain: '', title: t.title, description: t.description,
                  status: 'todo', costUsd: 0, updatedAt: new Date().toISOString(),
                })
              })
              setState(prev => ({ ...prev, tasks: [...tasksMapRef.current.values()] }))
              addLog('plan', `${planTasks.length} tasks planned`)
              break
            }
            case 'kanban_task': {
              const content = data?.content as { expertId?: string; status?: string } | undefined
              const expertId = content?.expertId ?? (data?.postedByExpertId as string | undefined)
              const status = content?.status
              if (expertId && status) {
                const existing = tasksMapRef.current.get(expertId)
                if (existing) {
                  tasksMapRef.current.set(expertId, { ...existing, status: status as any, updatedAt: new Date().toISOString() })
                  setState(prev => ({ ...prev, tasks: [...tasksMapRef.current.values()] }))
                }
              }
              addLog('task', `${expertId ?? 'unknown'} → ${status ?? 'unknown'}`)
              break
            }
            case 'kanban_artifact': {
              const expertId = data?.postedByExpertId as string | undefined
              if (expertId) {
                const existing = tasksMapRef.current.get(expertId)
                if (existing && existing.status !== 'done') {
                  tasksMapRef.current.set(expertId, { ...existing, status: 'under_review', updatedAt: new Date().toISOString() })
                  setState(prev => ({ ...prev, tasks: [...tasksMapRef.current.values()] }))
                }
              }
              addLog('artifact', `${(data?.eventType as string) ?? 'artifact'} by ${expertId ?? 'unknown'}`)
              break
            }
            default: break
          }
        } catch { addLog('parse_error', `Failed to parse: ${event.data.slice(0, 80)}`) }
      }
      es.onerror = () => {
        addLog('error', 'SSE disconnected. Reconnecting in 3s...')
        // FIX: approvalGate preserve rakho, null mat karo
        setState(prev => ({ ...prev, isConnected: false, isAwaitingApprovalId: workflowStatus === 'paused_for_approval' && !prev.approvalGate }))
        es.close(); esRef.current = null
        setState(prev => {
          if (!prev.isDone) retryRef.current = setTimeout(connect, 3000)
          return prev
        })
      }
    }
    connect()
    return () => { esRef.current?.close(); if (retryRef.current) clearTimeout(retryRef.current) }
  }, [workflowId, workflowStatus])

  // REST fallback: agar paused_for_approval hai par gate nahi hai to poll karo
  useEffect(() => {
    if (workflowStatus !== 'paused_for_approval' || state.approvalGate || !workflowId) return
    let cancelled = false
    const fetchPending = async () => {
      try {
        const token = useAuthStore.getState().accessToken
        const apiBase = import.meta.env.VITE_API_URL ?? ''
        const res = await fetch(`${apiBase}/api/v1/workflows/${workflowId}/approvals/pending`, {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        })
        if (!res.ok) return
        const json = await res.json()
        const data = camelizeKeys<any>(json.data ?? json)
        const id = data.approvalId ?? data.approval_id ?? data.id
        if (id && !cancelled) {
          setState(prev => ({ ...prev, approvalGate: { approvalId: String(id), gateName: data.gateName ?? 'approval', summary: data.summary ?? '' }, isAwaitingApprovalId: false }))
        }
      } catch {}
    }
    fetchPending()
    const iv = setInterval(fetchPending, 3000)
    return () => { cancelled = true; clearInterval(iv) }
  }, [workflowStatus, state.approvalGate, workflowId])

  return state
}
