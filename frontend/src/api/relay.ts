import { useAuthStore } from '@/stores/authStore'
import { camelizeKeys } from '@/utils/casing'
import type { PendingRelayReview, RelayFailure, RelayStepEvent } from '@/types/expert'

const base = `${import.meta.env.VITE_API_URL}/api/v1`

async function authedFetch(path: string, init?: RequestInit): Promise<Response> {
  const token = useAuthStore.getState().accessToken
  const headers: Record<string, string> = { ...(init?.headers as Record<string, string> | undefined) }
  if (token) headers['Authorization'] = `Bearer ${token}`
  return fetch(`${base}${path}`, { ...init, headers, credentials: 'include' })
}

export interface ActiveRelay {
  runId: string
  status: string
  events: RelayStepEvent[]
  pendingReview?: PendingRelayReview
  failure?: RelayFailure
}

export async function getActiveRelay(chatId: string): Promise<ActiveRelay | null> {
  const res = await authedFetch(`/chats/${chatId}/relay/active`)
  if (res.status === 404) return null
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  const json = await res.json()
  return camelizeKeys<ActiveRelay>(json)
}

export async function approveRelaySection(
  runId: string,
  sectionId: string,
  body: { edited_content?: string; resolution_source?: string },
): Promise<void> {
  const res = await authedFetch(`/collab/relay/${runId}/sections/${sectionId}/approve`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
}

export async function retryRelay(runId: string): Promise<void> {
  const res = await authedFetch(`/collab/relay/${runId}/retry`, { method: 'POST' })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
}
