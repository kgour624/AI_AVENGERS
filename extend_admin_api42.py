import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\api\admin.ts")
txt = p.read_text(encoding="utf-8", errors="ignore")
add = """
export type SystemHealthResponse = {
  flag_enabled: boolean
  flag_source: string
  retrieval_config: RetrievalConfig | null
  runtime: { go_routines: number; alloc_mb: number; num_cpu: number; uptime_sec: number }
  db_pool: { acquired_conns: number; idle_conns: number; total_conns: number }
  counts: { total_parents: number; total_children: number; orphans: number; experts: number }
  checked_at: string
}
export const getSystemHealth = async (): Promise<SystemHealthResponse> => {
  const res = await baseAPI.get<ApiResponse<SystemHealthResponse>>('/api/v1/admin/system-health')
  return res.data.data
}
export const compareRetrieval = async (body: { expert_id: string; query: string; top_k?: number }) => {
  const res = await baseAPI.post<ApiResponse<unknown>>('/api/v1/admin/retrieval-compare', body)
  return res.data.data
}
export const getAuditLog = async (params?: { limit?: number; offset?: number; action?: string }) => {
  const res = await baseAPI.get<ApiResponse<{ items: unknown[]; total: number }>>('/api/v1/admin/audit-log', { params })
  return res.data.data
}
"""
if "getSystemHealth" not in txt:
    p.write_text(txt + add, encoding="utf-8")
    print("extended admin.ts with health/compare/audit")
else:
    print("already extended")
