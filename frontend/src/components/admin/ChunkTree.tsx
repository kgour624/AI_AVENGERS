import { useState } from 'react'

function hasFence(s: string) {
  return (s || '').includes('```')
}

function Badge({ children, tone = 'default' }: { children: React.ReactNode; tone?: 'default' | 'amber' | 'violet' | 'cyan' }) {
  const map: Record<string, string> = {
    default: 'bg-surface-secondary border-surface-border text-text-secondary',
    amber: 'bg-amber-500/15 border-amber-500/30 text-amber-600',
    violet: 'bg-violet-500/15 border-violet-500/30 text-violet-600',
    cyan: 'bg-cyan-500/15 border-cyan-500/30 text-cyan-600',
  }
  return <span className={`inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-medium uppercase tracking-wider ${map[tone]}`}>{children}</span>
}

export function ChunkTree(props: {
  tree: { parent: any; children: any[] }[]
  orphans: any[]
  compact?: boolean
}) {
  const tree = props.tree ?? []
  const orphans = props.orphans ?? []

  return (
    <div className="space-y-3">
      {tree.map((node, i) => (
        <ParentNode key={node.parent?.id ?? i} parent={node.parent} children={node.children} compact={props.compact} />
      ))}
      {orphans.length > 0 && (
        <div className="rounded-lg border border-amber-500/20 bg-amber-500/5 p-3">
          <div className="mb-2 flex items-center gap-2">
            <Badge tone="amber">orphans {orphans.length}</Badge>
            <span className="text-xs text-text-secondary">is_child=true but parent_id NULL — before backfill or atomic fence orphan</span>
          </div>
          <div className="space-y-2">
            {orphans.slice(0, 20).map((ch, i) => (
              <ChildRow key={ch.id ?? i} row={ch} />
            ))}
            {orphans.length > 20 && <p className="text-xs text-text-disabled">+ {orphans.length - 20} more orphans</p>}
          </div>
        </div>
      )}
      {tree.length === 0 && orphans.length === 0 && (
        <p className="py-6 text-center text-sm text-text-disabled">No parents/children — ingestion not parent-child or expert empty</p>
      )}
    </div>
  )
}

function ParentNode(props: { parent: any; children: any[]; compact?: boolean }) {
  const [open, setOpen] = useState(true)
  const p = props.parent ?? {}
  const children = props.children ?? []
  const pageIndex = p.pageIndex ?? p.page_index ?? 0
  const tokenCount = p.tokenCount ?? p.token_count ?? 0
  const childCount = p.childCount ?? p.child_count ?? children.length
  const sectionPath = p.sectionPath ?? p.section_path ?? ''
  const pageText = p.pageText ?? p.page_text ?? ''
  const sourceFile = p.sourceFile ?? p.source_file ?? ''

  return (
    <div className="rounded-lg border border-surface-border bg-surface-raised shadow-card overflow-hidden">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center justify-between gap-3 px-3 py-2 text-left hover:bg-surface-secondary/40"
      >
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-xs font-semibold text-text-primary">P{pageIndex}</span>
            <Badge tone="violet">{tokenCount} tok</Badge>
            <Badge tone="cyan">{childCount} children</Badge>
            {sectionPath ? <Badge>{sectionPath}</Badge> : null}
            {hasFence(pageText) ? <Badge tone="amber">code fence</Badge> : null}
            {sourceFile ? <span className="max-w-[18ch] truncate text-xs text-text-disabled">{sourceFile.split('/').pop()}</span> : null}
          </div>
          {!props.compact && (
            <p className="mt-1 line-clamp-2 text-xs leading-relaxed text-text-secondary">{pageText.slice(0, 260)}</p>
          )}
        </div>
        <span className="shrink-0 text-xs text-text-disabled">{open ? '▾' : '▸'}</span>
      </button>
      {open && (
        <div className="border-t border-surface-border bg-surface-secondary/20">
          {children.length === 0 ? (
            <p className="px-3 py-2 text-xs text-text-disabled">no children — orphan parent or parent-only page</p>
          ) : (
            <div className="divide-y divide-surface-border/60">
              {children.map((ch: any, i: number) => (
                <ChildRow key={ch.id ?? i} row={ch} parentIndex={pageIndex} />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function ChildRow(props: { row: any; parentIndex?: number }) {
  const r = props.row ?? {}
  const chunkText = r.chunkText ?? r.chunk_text ?? ''
  const chunkIndex = r.chunkIndex ?? r.chunk_index ?? 0
  const tokenEstimate = r.tokenEstimate ?? r.token_estimate ?? 0
  const parentIndex = r.parentIndex ?? r.parent_index ?? props.parentIndex
  const sectionPath = r.sectionPath ?? r.section_path ?? ''
  const chunkHash = r.chunkHash ?? r.chunk_hash ?? ''
  const sourceFile = r.sourceFile ?? r.source_file ?? ''
  const fence = hasFence(chunkText)
  const hash8 = chunkHash ? chunkHash.slice(0, 8) : ''

  return (
    <div className="flex gap-2 px-3 py-2">
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-1.5">
          <span className="text-xs font-medium text-text-primary">c{chunkIndex}</span>
          <Badge tone="default">{tokenEstimate} tok</Badge>
          {parentIndex != null ? <Badge tone="violet">→P{parentIndex}</Badge> : <Badge tone="amber">orphan</Badge>}
          {sectionPath ? <Badge>{sectionPath}</Badge> : null}
          {fence ? <Badge tone="amber">fence</Badge> : null}
          {hash8 ? <span className="font-mono text-[10px] text-text-disabled">{hash8}</span> : null}
          {sourceFile ? <span className="max-w-[14ch] truncate text-[10px] text-text-disabled">{sourceFile.split('/').pop()}</span> : null}
        </div>
        <p className="mt-1 line-clamp-2 text-xs leading-relaxed text-text-secondary">{chunkText.slice(0, 220)}</p>
      </div>
    </div>
  )
}
