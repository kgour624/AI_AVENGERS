import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useRepoSyncStatus } from '@/hooks/useRepoSyncStatus'
import { getRepoFile, getRepoTree, type RepoTreeEntry } from '@/api/repo'
import { Card } from '@/components/ui/Card'
import { cn } from '@/utils/cn'

/**
 * RepoTreePanel — read-only browser for the CLIENT's own repository (Phase 3A).
 *
 * Deliberately distinct from workflow/FilesPanel.tsx: that panel shows code the
 * workflow GENERATED (live, via SSE). This one shows the client's existing
 * codebase as it was stored at sync time. Two different sources, so two
 * components — sharing one would couple "generated output" to "ingested input"
 * and start lying the moment either changes shape.
 *
 * The backend returns a flat, path-ordered list; folders are reconstructed here
 * in a single pass, matching how FilesPanel already builds its tree.
 *
 * Visibility: the panel renders nothing until the repo is connected AND its
 * first sync has completed, because until then there is no tree to show and an
 * empty panel would read as "your repository is empty".
 */

interface TreeNode {
  name: string
  path: string
  isDir: boolean
  children: TreeNode[]
  entry?: RepoTreeEntry
}

/** Builds a sorted directory tree from the flat path list. Pure. */
export function buildRepoTree(files: RepoTreeEntry[]): TreeNode[] {
  const root: TreeNode = { name: '', path: '', isDir: true, children: [] }

  for (const entry of files) {
    const parts = entry.path.split('/').filter(Boolean)
    if (parts.length === 0) continue

    let cursor = root
    parts.forEach((part, index) => {
      const isLast = index === parts.length - 1
      const path = parts.slice(0, index + 1).join('/')
      let node = cursor.children.find((child) => child.path === path)
      if (!node) {
        node = { name: part, path, isDir: !isLast, children: [] }
        cursor.children.push(node)
      }
      if (isLast) node.entry = entry
      cursor = node
    })
  }

  sortNodes(root.children)
  return root.children
}

function sortNodes(nodes: TreeNode[]) {
  nodes.sort((a, b) => {
    if (a.isDir !== b.isDir) return a.isDir ? -1 : 1
    return a.name.localeCompare(b.name)
  })
  for (const node of nodes) {
    if (node.children.length > 0) sortNodes(node.children)
  }
}

function formatBytes(bytes?: number): string {
  if (bytes === undefined || bytes === null) return ''
  if (bytes < 1024) return `${bytes} B`
  return `${(bytes / 1024).toFixed(1)} KB`
}

interface TreeRowProps {
  node: TreeNode
  depth: number
  collapsed: Set<string>
  onToggle: (path: string) => void
  selectedPath: string | null
  onSelect: (path: string) => void
}

function TreeRow({ node, depth, collapsed, onToggle, selectedPath, onSelect }: TreeRowProps) {
  const indent = { paddingLeft: `${depth * 12 + 8}px` }

  if (node.isDir) {
    const isCollapsed = collapsed.has(node.path)
    return (
      <div>
        <button
          type="button"
          onClick={() => onToggle(node.path)}
          style={indent}
          className="flex w-full items-center gap-1.5 rounded px-2 py-1 text-left text-xs text-text-secondary hover:bg-surface-overlay/60"
        >
          <span className="text-text-disabled" aria-hidden="true">
            {isCollapsed ? '\u25b8' : '\u25be'}
          </span>
          <span className="truncate font-medium">{node.name}</span>
        </button>
        {!isCollapsed &&
          node.children.map((child) => (
            <TreeRow
              key={child.path}
              node={child}
              depth={depth + 1}
              collapsed={collapsed}
              onToggle={onToggle}
              selectedPath={selectedPath}
              onSelect={onSelect}
            />
          ))}
      </div>
    )
  }

  // has_content false is shown, not hidden: the client should see that their
  // repository contains the file even when we chose not to store its bytes.
  const canOpen = node.entry?.has_content ?? false
  const isSelected = selectedPath === node.path

  return (
    <button
      type="button"
      onClick={() => canOpen && onSelect(node.path)}
      disabled={!canOpen}
      title={canOpen ? node.path : `${node.path} — content not stored (binary, unsupported, or too large)`}
      style={indent}
      className={cn(
        'flex w-full items-center justify-between gap-2 rounded px-2 py-1 text-left text-xs',
        isSelected
          ? 'bg-surface-overlay text-text-primary'
          : 'text-text-secondary hover:bg-surface-overlay/60',
        !canOpen && 'cursor-not-allowed text-text-disabled hover:bg-transparent'
      )}
    >
      <span className="truncate">{node.name}</span>
      {!canOpen && (
        <span className="shrink-0 text-[10px] uppercase tracking-wider text-text-disabled">
          metadata
        </span>
      )}
    </button>
  )
}

export function RepoTreePanel({ projectId }: { projectId: string }) {
  const { data: status } = useRepoSyncStatus(projectId)
  const [selectedPath, setSelectedPath] = useState<string | null>(null)
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())

  const isSynced = status?.connected === true && status.status === 'complete'

  const { data: tree, isLoading } = useQuery({
    queryKey: ['projects', projectId, 'repo', 'tree'],
    queryFn: () => getRepoTree(projectId),
    enabled: isSynced,
  })

  const { data: file, error: fileError, isFetching: fileLoading } = useQuery({
    queryKey: ['projects', projectId, 'repo', 'file', selectedPath],
    queryFn: () => getRepoFile(projectId, selectedPath as string),
    enabled: isSynced && !!selectedPath,
    // A 404/409 is a real answer about this file, not a transient failure.
    retry: false,
  })

  const nodes = useMemo(() => buildRepoTree(tree?.files ?? []), [tree])

  if (!isSynced) return null

  const onToggle = (path: string) => {
    setCollapsed((prev) => {
      const next = new Set(prev)
      if (next.has(path)) next.delete(path)
      else next.add(path)
      return next
    })
  }

  const fileCount = tree?.files.length ?? 0

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <span className="text-xs font-semibold uppercase tracking-wide text-text-secondary">
          Repository files
          {fileCount > 0 ? ` (${fileCount})` : ''}
        </span>
        {tree?.commit_sha && (
          <span className="text-[10px] font-medium uppercase tracking-wider text-text-disabled">
            at {tree.commit_sha.slice(0, 7)} · read-only
          </span>
        )}
      </div>

      {isLoading ? (
        <div className="h-40 animate-pulse rounded-lg bg-surface-overlay" />
      ) : fileCount === 0 ? (
        <Card className="p-6">
          <p className="text-center text-xs text-text-disabled">
            No file tree stored yet. Run a sync to index this repository.
          </p>
        </Card>
      ) : (
        <div className="grid grid-cols-3 gap-4">
          <Card className="col-span-1 max-h-[28rem] overflow-y-auto p-2">
            {nodes.map((node) => (
              <TreeRow
                key={node.path}
                node={node}
                depth={0}
                collapsed={collapsed}
                onToggle={onToggle}
                selectedPath={selectedPath}
                onSelect={setSelectedPath}
              />
            ))}
          </Card>

          <Card className="col-span-2 max-h-[28rem] overflow-y-auto p-3">
            {!selectedPath ? (
              <p className="py-4 text-center text-xs text-text-disabled">
                Select a file to view it. Files marked “metadata” were listed but
                their content was not stored.
              </p>
            ) : fileLoading ? (
              <p className="py-4 text-center text-xs text-text-disabled">Loading file…</p>
            ) : fileError || !file ? (
              <p className="py-4 text-center text-xs text-mode-refuse">
                This file’s content could not be loaded.
              </p>
            ) : (
              <>
                <div className="mb-2 flex items-center justify-between gap-3">
                  <p className="truncate text-xs font-medium text-text-primary">{file.path}</p>
                  <div className="flex shrink-0 items-center gap-2">
                    {file.language && (
                      <span className="text-[10px] uppercase tracking-wider text-text-disabled">
                        {file.language}
                      </span>
                    )}
                    <span className="text-xs text-text-disabled">{formatBytes(file.size_bytes)}</span>
                  </div>
                </div>
                <pre className="overflow-x-auto whitespace-pre text-xs text-text-secondary">
                  {file.content}
                </pre>
              </>
            )}
          </Card>
        </div>
      )}
    </div>
  )
}
