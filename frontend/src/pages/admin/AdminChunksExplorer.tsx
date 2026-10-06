import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  getAdminExperts,
  listExpertChunks,
  listExpertParents,
  getExpertChunkTree,
  getSystemHealth,
} from '@/api/admin'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Skeleton } from '@/components/ui/Skeleton'
import { ChunkTree } from '@/components/admin/ChunkTree'
import { SystemHealthCard } from '@/components/admin/SystemHealthCard'
import { ComparePanel } from '@/components/admin/ComparePanel'
import { AuditLogList } from '@/components/admin/AuditLogList'
import { handleAPIError } from '@/utils/errors'

type Tab = 'tree' | 'chunks' | 'parents'

export function AdminChunksExplorer() {
  const [expertId, setExpertId] = useState('')
  const [tab, setTab] = useState<Tab>('tree')
  const [isChild, setIsChild] = useState('')
  const [sourceFile, setSourceFile] = useState('')
  const [page, setPage] = useState(0)
  const limit = 20

  const expertsQ = useQuery({
    queryKey: ['admin', 'experts'],
    queryFn: getAdminExperts,
  })

  const healthQ = useQuery({
    queryKey: ['admin', 'system-health'],
    queryFn: getSystemHealth,
  })

  const chunksParams = {
    limit,
    offset: page * limit,
    ...(isChild ? { is_child: isChild === 'true' } : {}),
    ...(sourceFile ? { source_file: sourceFile } : {}),
  } as any

  const chunksQ = useQuery({
    queryKey: ['admin', 'chunks', expertId, chunksParams],
    queryFn: () => listExpertChunks(expertId, chunksParams),
    enabled: !!expertId && tab === 'chunks',
  })

  const parentsQ = useQuery({
    queryKey: ['admin', 'parents', expertId, page, sourceFile],
    queryFn: () =>
      listExpertParents(expertId, {
        limit,
        offset: page * limit,
        ...(sourceFile ? { source_file: sourceFile } : {}),
      }),
    enabled: !!expertId && tab === 'parents',
  })

  const treeQ = useQuery({
    queryKey: ['admin', 'chunk-tree', expertId, sourceFile],
    queryFn: () =>
      getExpertChunkTree(expertId, {
        ...(sourceFile ? { source_file: sourceFile } : {}),
      }),
    enabled: !!expertId && tab === 'tree',
  })

  if (expertsQ.isLoading)
    return (
      <div className="p-6">
        <Skeleton className="h-24" />
      </div>
    )
  if (expertsQ.error)
    return (
      <div className="p-6 text-sm text-mode-refuse">
        {handleAPIError(expertsQ.error as any)}
      </div>
    )

  const experts = expertsQ.data ?? []

  const treeData = treeQ.data as any
  const totalParents = treeData?.totalParents ?? treeData?.total_parents ?? 0
  const totalChildren = treeData?.totalChildren ?? treeData?.total_children ?? 0

  const chunksData = chunksQ.data as any
  const chunksTotal = chunksData?.total ?? 0
  const chunksItems: any[] = chunksData?.items ?? []

  const parentsData = parentsQ.data as any
  const parentsItems: any[] = parentsData?.items ?? []

  return (
    <div className="space-y-4 p-6">
      <div>
        <h2 className="text-lg font-semibold text-text-primary">Chunk Explorer</h2>
        <p className="mt-1 text-xs text-text-secondary">Read-only parent-child inspector.</p>
      </div>

      <SystemHealthCard
        data={healthQ.data as never}
        isLoading={healthQ.isLoading}
        error={healthQ.error}
      />

      <Card>
        <div className="flex flex-wrap items-end gap-3">
          <label className="block">
            <span className="text-xs font-medium uppercase tracking-wider text-text-secondary">
              Expert
            </span>
            <select
              value={expertId}
              onChange={(e) => {
                setExpertId(e.target.value)
                setPage(0)
              }}
              className="mt-1 block w-64 rounded border border-surface-border bg-surface-secondary px-2 py-1.5 text-sm text-text-primary"
            >
              <option value="">Select expert...</option>
              {experts.map((ex: any) => (
                <option key={ex.id} value={ex.id}>
                  {ex.name} - {ex.slug}
                </option>
              ))}
            </select>
          </label>
        </div>
      </Card>

      {!expertId ? (
        <Card>
          <p className="py-8 text-center text-sm text-text-disabled">Pick an expert</p>
        </Card>
      ) : (
        <>
          <div className="flex gap-2">
            {(['tree', 'chunks', 'parents'] as Tab[]).map((t) => (
              <button
                key={t}
                onClick={() => {
                  setTab(t)
                  setPage(0)
                }}
                className={
                  'rounded-full border px-3 py-1 text-xs font-medium capitalize ' +
                  (tab === t
                    ? 'border-glow-violet bg-violet-500/10 text-violet-600'
                    : 'border-surface-border bg-surface-secondary text-text-secondary')
                }
              >
                {t === 'tree' ? 'Tree' : t}
              </button>
            ))}
          </div>

          {tab === 'tree' && (
            <Card>
              {treeQ.isLoading ? (
                <Skeleton className="h-40" />
              ) : treeQ.error ? (
                <p className="text-sm text-mode-refuse">
                  {handleAPIError(treeQ.error as any)}
                </p>
              ) : treeQ.data ? (
                <>
                  <div className="mb-3 flex gap-2 text-xs">
                    <span className="rounded-full bg-surface-secondary px-2 py-0.5">
                      {totalParents} parents
                    </span>
                    <span className="rounded-full bg-surface-secondary px-2 py-0.5">
                      {totalChildren} children
                    </span>
                  </div>
                  <ChunkTree tree={treeData.tree} orphans={treeData.orphans} />
                </>
              ) : null}
            </Card>
          )}

          {tab === 'chunks' && (
            <Card>
              {chunksQ.isLoading ? (
                <Skeleton className="h-40" />
              ) : chunksQ.error ? (
                <p className="text-sm text-mode-refuse">
                  {handleAPIError(chunksQ.error as any)}
                </p>
              ) : chunksQ.data ? (
                <>
                  <div className="mb-2 flex justify-between">
                    <span className="text-xs">{chunksTotal} rows</span>
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        variant="ghost"
                        disabled={page === 0}
                        onClick={() => setPage((p) => Math.max(0, p - 1))}
                      >
                        Prev
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => setPage((p) => p + 1)}
                      >
                        Next
                      </Button>
                    </div>
                  </div>
                  <div className="divide-y divide-surface-border">
                    {chunksItems.map((r: any) => {
                      const text = r.chunkText ?? r.chunk_text ?? ''
                      return (
                        <div key={r.id} className="py-2">
                          <p className="text-xs">{text.slice(0, 220)}</p>
                        </div>
                      )
                    })}
                  </div>
                </>
              ) : null}
            </Card>
          )}

          {tab === 'parents' && (
            <Card>
              {parentsQ.isLoading ? (
                <Skeleton className="h-40" />
              ) : parentsQ.error ? (
                <p className="text-sm text-mode-refuse">
                  {handleAPIError(parentsQ.error as any)}
                </p>
              ) : parentsQ.data ? (
                <div className="space-y-2">
                  {parentsItems.map((pr: any) => {
                    const text = pr.pageText ?? pr.page_text ?? ''
                    return (
                      <div key={pr.id} className="rounded border p-2">
                        <p className="text-xs">{text.slice(0, 200)}</p>
                      </div>
                    )
                  })}
                </div>
              ) : null}
            </Card>
          )}
        </>
      )}

      {expertId ? <ComparePanel expertId={expertId} /> : null}
      <AuditLogList />
    </div>
  )
}

export const Component = AdminChunksExplorer
