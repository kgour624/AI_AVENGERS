import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'

// MCP tokens let an external coding agent (Claude Code, Cursor) use the domain
// experts. They are admin-only: a client or expert role can never reach them.
export interface McpTokenRecord {
  id: string
  label: string
  domains: string[]
  tools: string[]
  revoked: boolean
  last_used_at?: string
  request_count: number
  created_at: string
}

export const listMcpTokens = () =>
  baseAPI.get<ApiResponse<McpTokenRecord[]>>('/api/v1/admin/mcp-tokens').then((res) => res.data.data!)

// The plaintext token is returned ONLY here. After this reply the server holds
// nothing but its hash, so the UI must show it immediately and never again.
export const createMcpToken = (body: { label: string; domains: string[]; tools: string[] }) =>
  baseAPI
    .post<ApiResponse<{ token: string; record: McpTokenRecord }>>('/api/v1/admin/mcp-tokens', body)
    .then((res) => res.data.data!)

export const revokeMcpToken = (id: string) =>
  baseAPI.post<ApiResponse<{ status: string }>>(`/api/v1/admin/mcp-tokens/${id}/revoke`).then((res) => res.data.data!)
