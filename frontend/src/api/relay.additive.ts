import { apiClient } from './base'

export type RelayRunStatus = {
  runId: string
  status: string
  failedAtIndex?: number
  failureReason?: string
  retryDeadline?: string
}

export async function getRelayRun(chatId: string, runId: string): Promise<RelayRunStatus> {
  const { data } = await apiClient.get(`/chats/${chatId}/relay/${runId}`)
  return data as RelayRunStatus
}

export async function resumeRelay(chatId: string, runId: string): Promise<RelayRunStatus> {
  const { data } = await apiClient.post(`/chats/${chatId}/relay/${runId}/resume`)
  return data as RelayRunStatus
}
