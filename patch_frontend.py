import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\api\admin.ts")
txt = p.read_text(encoding="utf-8", errors="ignore")
old = "export const deleteExpert = (expertId: string) =>\n  baseAPI.delete<ApiResponse<{ status: string }>>(`/api/v1/admin/experts/${expertId}`).then((res) => res.data.data!)"
new = """// Chunk Explorer (Phase 1.5 Observability) — read-only parent-child inspector
export interface ChunkExplorerRow {
  id: string
  expert_id: string
  chunk_index: number
  chunk_text: string
  topic?: string | null
  subtopic?: string | null
  source_file?: string | null
  chunk_hash?: string | null
  section_path: string
  parent_id?: string | null
  parent_index?: number | null
  is_child: boolean
  token_estimate: number
  created_at: string
}
export interface ParentExplorerRow {
  id: string
  expert_id: string
  page_index: number
  page_text: string
  section_path: string
  token_count: number
  source_file?: string | null
  created_at: string
  child_count: number
}
export interface ChunkExplorerListResponse {
  items: ChunkExplorerRow[]
  total: number
  limit: number
  offset: number
}
export interface ParentExplorerListResponse {
  items: ParentExplorerRow[]
  total: number
  limit: number
  offset: number
}
export interface ChunkTreeResponse {
  tree: { parent: ParentExplorerRow; children: ChunkExplorerRow[] }[]
  orphans: ChunkExplorerRow[]
  total_children: number
  total_parents: number
  parents_shown: number
  children_shown: number
}
export type ChunkExplorerParams = {
  limit?: number
  offset?: number
  is_child?: boolean
  parent_id?: string
  source_file?: string
}
export const listExpertChunks = (expertId: string, params: ChunkExplorerParams = {}) =>
  baseAPI
    .get<ApiResponse<ChunkExplorerListResponse>>(`/api/v1/admin/experts/${expertId}/chunks`, { params })
    .then((res) => res.data.data!)
export const listExpertParents = (expertId: string, params: { limit?: number; offset?: number; source_file?: string } = {}) =>
  baseAPI
    .get<ApiResponse<ParentExplorerListResponse>>(`/api/v1/admin/experts/${expertId}/parents`, { params })
    .then((res) => res.data.data!)
export const getExpertChunkTree = (expertId: string, params: { source_file?: string } = {}) =>
  baseAPI
    .get<ApiResponse<ChunkTreeResponse>>(`/api/v1/admin/experts/${expertId}/chunk-tree`, { params })
    .then((res) => res.data.data!)

export const deleteExpert = (expertId: string) =>
  baseAPI.delete<ApiResponse<{ status: string }>>(`/api/v1/admin/experts/${expertId}`).then((res) => res.data.data!)"""
if old in txt:
    txt = txt.replace(old, new)
    p.write_text(txt, encoding="utf-8")
    print("patched admin.ts")
else:
    print("old not found")
    print(txt[txt.rfind("deleteExpert")-200:txt.rfind("deleteExpert")+400])
