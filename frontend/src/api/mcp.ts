import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'

// MCP tokens let an external coding agent (Claude Code, Cursor) use the domain
// experts. They are admin-only: a client or expert role can never reach them.
export interface McpTokenRecord {
  id: string
  label: string
  domains: string[]
  expert_ids: string[]
  tools: string[]
  revoked: boolean
  last_used_at?: string
  request_count: number
  created_at: string
}

// --- mcp-v2: dynamic tools (single source of truth from Postgres) ---
export interface ToolAction {
  type: 'SQL_READ' | 'API_CALL' | 'LLM_PROMPT' | 'COMPOSITE'
  config?: Record<string, any>
  output_mapping?: string
  timeout_ms?: number
}
// Phase 3: LLM_PROMPT + COMPOSITE live, Redis cache 5m TTL on ListTools

export interface ToolDefinition {
  id: string
  name: string // canonical e.g. 'search_course_content'
  // add if missing:
  input_schema?: any
  display_name: string // UI label e.g. 'Search Course Content'
  description: string
  is_active: boolean
  action?: ToolAction
  created_at?: string
  updated_at?: string
}

export const listMcpTokens = () =>
  baseAPI.get<ApiResponse<McpTokenRecord[]>>('/api/v1/admin/mcp-tokens').then((res) => res.data.data!)

// The plaintext token is returned ONLY here. After this reply the server holds
// nothing but its hash, so the UI must show it immediately and never again.
export const createMcpToken = (body: { label: string; domains?: string[]; expert_ids: string[]; tools: string[] }) =>
  baseAPI
    .post<ApiResponse<{ token: string; record: McpTokenRecord }>>('/api/v1/admin/mcp-tokens', body)
    .then((res) => res.data.data!)

export const revokeMcpToken = (id: string) =>
  baseAPI.post<ApiResponse<{ status: string }>>(`/api/v1/admin/mcp-tokens/${id}/revoke`).then((res) => res.data.data!)

export const deleteMcpToken = (id: string) =>
  baseAPI.delete<ApiResponse<{ status: string }>>(`/api/v1/admin/mcp-tokens/${id}`).then((res) => res.data.data!)

// --- mcp-v2: fetch tools dynamically from Postgres via mcp-v2 hexagon ---
// GET /api/v1/mcp-v2/tools -> { tools: ToolDefinition[] }
// Wired in backend-go/internal/mcp-v2/handler.go:HandleListTools() -> service.ListTools() -> storage.ListTools()
export const fetchMcpV2Tools = (): Promise<ToolDefinition[]> =>
  baseAPI.get<{ tools: ToolDefinition[] }>('/api/v1/mcp-v2/tools').then((res) => res.data.tools)

export const createMcpV2Tool = (body: Omit<ToolDefinition, 'id' | 'is_active'> & { input_schema: any; action?: ToolAction }) =>
  baseAPI.post('/api/v1/mcp-v2/tools', body).then((res) => res.data)

export const generateMcpV2SQL = (prompt: string) =>
  baseAPI.post<{ sql: string }>('/api/v1/mcp-v2/tools/generate-sql', { prompt }).then((res) => res.data.sql)

// Optional alias if you prefer naming consistency with React Query keys
export const listMcpV2Tools = fetchMcpV2Tools
