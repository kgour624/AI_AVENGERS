import { useMemo, useState } from 'react'
import type { FileEntry, FileStreamState } from '@/hooks/useFileStream'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { cn } from '@/utils/cn'

/**
 * FilesPanel — live browser for the code the workflow generated.
 *
 * Extracted from KanbanPage.tsx and promoted to a first-class tab
 * (feature #28). WHY: the panel used to be a stacked section that returned
 * `null` until the first artifact arrived, so it was invisible below the fold
 * and looked unimplemented. It is now always rendered (with an empty state)
 * and shows the files as a directory TREE instead of a flat path list.
 *
 * Data comes from useFileStream (SSE /workflows/:id/files/stream) — this
 * component is presentational; it holds only view state (selection, collapsed
 * folders). Backend is unchanged.
 */
export interface FilesPanelProps {
  files: FileEntry[]
  isConnected: boolean
  isDone: boolean
  lastWave: FileStreamState['lastWave']
}

// TreeNode is a directory (isDir) or a leaf file. `path` is the full path so
// collapse state can key on it without re-walking the tree.
interface TreeNode {
  name: string
  path: string
  isDir: boolean
  children: TreeNode[]
  file?: FileEntry
}

// buildTree — pure. Turns flat file paths into a sorted directory tree.
// Splitting on '/' is enough: the backend stores POSIX-style relative paths
// (e.g. "src/api/handler.go") on every code_artifact_produced event.
export function buildTree(files: FileEntry[]): TreeNode[] {
  const root: TreeNode = { name: '', path: '', isDir: true, children: [] }

  for (const file of files) {
    const parts = file.filePath.split('/').filter(Boolean)
    if (parts.length === 0) continue

    let cursor = root
    parts.forEach((part, index) => {
      const isLast = index === parts.length - 1
      const path = parts.slice(0, index + 1).join('/')
      let node = cursor.children.find((c) => c.path === path)
      if (!node) {
        node = { name: part, path, isDir: !isLast, children: [] }
        cursor.children.push(node)
      }
      if (isLast) node.file = file
      cursor = node
    })
  }

  sortTree(root.children)
  return root.children
}

function sortTree(nodes: TreeNode[]) {
  nodes.sort((a, b) => {
    if (a.isDir !== b.isDir) return a.isDir ? -1 : 1
    return a.name.localeCompare(b.name)
  })
  for (const node of nodes) {
    if (node.children.length > 0) sortTree(node.children)
  }
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

  const isSelected = selectedPath === node.path
  return (
    <button
      type="button"
      onClick={() => onSelect(node.path)}
      style={indent}
      className={cn(
        'flex w-full items-center justify-between gap-2 rounded px-2 py-1 text-left text-xs',
        isSelected
          ? 'bg-surface-overlay text-text-primary'
          : 'text-text-secondary hover:bg-surface-overlay/60'
      )}
    >
      <span className="truncate">{node.name}</span>
      <Badge variant={node.file?.operation === 'create' ? 'success' : 'warning'}>
        {node.file?.operation ?? 'modify'}
      </Badge>
    </button>
  )
}

export function FilesPanel({ files, isConnected, isDone, lastWave }: FilesPanelProps) {
  const [selectedPath, setSelectedPath] = useState<string | null>(null)
  // Collapsed folder paths. Default: everything expanded.
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())

  const tree = useMemo(() => buildTree(files), [files])

  const onToggle = (path: string) => {
    setCollapsed((prev) => {
      const next = new Set(prev)
      if (next.has(path)) next.delete(path)
      else next.add(path)
      return next
    })
  }

  const selected = selectedPath ? files.find((f) => f.filePath === selectedPath) : undefined

  const statusLabel = isDone ? 'Complete' : isConnected ? 'Live' : 'Connecting...'
  const statusDot = isDone
    ? 'bg-mode-advise'
    : isConnected
      ? 'bg-mode-advise animate-pulse'
      : 'bg-glow-amber animate-pulse'

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <span className="text-xs font-semibold uppercase tracking-wide text-text-secondary">
          Files ({files.length})
        </span>
        <div className="flex items-center gap-3">
          {lastWave && (
            <span className="text-[10px] font-medium uppercase tracking-wider text-text-disabled">
              Wave {lastWave.waveIndex + 1}
              {lastWave.phase ? ` \u00b7 ${lastWave.phase.replace(/_/g, ' ')}` : ''}
              {lastWave.mergedFiles.length > 0 ? ` \u00b7 ${lastWave.mergedFiles.length} merged` : ''}
            </span>
          )}
          {/* Per-stream indicator: this panel can be 'Live' while the Kanban
              stream is done, and vice-versa — show its own state honestly. */}
          <div className="flex items-center gap-1.5">
            <span className={cn('h-2 w-2 rounded-full', statusDot)} />
            <span className="text-[10px] font-medium uppercase tracking-wider text-text-disabled">
              {statusLabel}
            </span>
          </div>
        </div>
      </div>

      {files.length === 0 ? (
        <Card className="p-6">
          <p className="text-center text-xs text-text-disabled">
            {isDone
              ? 'This workflow produced no code files.'
              : 'No files yet \u2014 generated code will appear here once implementation starts.'}
          </p>
        </Card>
      ) : (
        <div className="grid grid-cols-3 gap-4">
          <Card className="col-span-1 max-h-[28rem] overflow-y-auto p-2">
            {tree.map((node) => (
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
            {selected ? (
              <>
                <div className="mb-2 flex items-center justify-between gap-3">
                  <p className="truncate text-xs font-medium text-text-primary">{selected.filePath}</p>
                  <div className="flex shrink-0 items-center gap-2">
                    {selected.language && (
                      <span className="text-[10px] uppercase tracking-wider text-text-disabled">
                        {selected.language}
                      </span>
                    )}
                    <span className="text-xs text-text-disabled">{selected.linesOfCode} lines</span>
                  </div>
                </div>
                {selected.phase && (
                  <p className="mb-2 text-[10px] uppercase tracking-wider text-text-disabled">
                    {selected.phase.replace(/_/g, ' ')}
                    {selected.commitSha ? ` \u00b7 ${selected.commitSha.slice(0, 7)}` : ''}
                  </p>
                )}
                {!selected.validationPassed && selected.validationError && (
                  <p className="mb-2 text-xs text-mode-refuse">{selected.validationError}</p>
                )}
                <pre className="overflow-x-auto whitespace-pre text-xs text-text-secondary">
                  {selected.content}
                </pre>
              </>
            ) : (
              <p className="py-4 text-center text-xs text-text-disabled">
                Select a file to view its content
              </p>
            )}
          </Card>
        </div>
      )}
    </div>
  )
}
