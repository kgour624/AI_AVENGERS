import { useParams } from 'react-router-dom'
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getWorkflow, respondToApproval } from '@/api/workflows'
import { useKanbanStream } from '@/hooks/useKanbanStream'
import { useFileStream } from '@/hooks/useFileStream'
import type { FileEntry } from '@/hooks/useFileStream'
import type { KanbanTask } from '@/api/workflows'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { cn } from '@/utils/cn'

// ApprovalGate component — renders Approve/Request Changes buttons.
// Shown when workflow.status === 'paused_for_approval'.
function ApprovalGate({
  workflowId,
  approvalId,
  gateName,
  summary,
}: {
  workflowId: string
  approvalId: string
  gateName: string
  summary: string
}) {
  const [notes, setNotes] = useState('')
  const [loading, setLoading] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  // sent: this gate has been responded to successfully.
  // WHY: the gate stays mounted until the 10s workflow poll reports a status
  // other than paused_for_approval. Re-enabling the buttons in that window
  // let the user fire the same approval several times (6 duplicate POSTs were
  // seen in one 3-second window). The backend now rejects repeats with 409,
  // but the UI should not send them in the first place.
  const [sent, setSent] = useState(false)

  const respond = async (decision: 'approve' | 'request_changes') => {
    if (!approvalId) {
      setError('Approval ID not yet received from server. Please wait a moment and try again.')
      return
    }
    setLoading(decision)
    setError(null)
    try {
      await respondToApproval(workflowId, approvalId, decision, notes || undefined)
      setSent(true)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Request failed')
    } finally {
      setLoading(null)
    }
  }

  return (
    <div className="mt-4 rounded-lg border border-glow-amber/40 bg-glow-amber/10 p-4">
      <p className="text-sm font-semibold text-glow-amber">
        {'\u23f8'} Waiting for Approval
        {gateName && (
          <span className="ml-2 text-xs font-normal text-text-disabled">({gateName})</span>
        )}
      </p>
      <p className="mt-1 text-xs text-text-secondary">{summary}</p>
      <textarea
        className="mt-3 w-full rounded border border-surface-overlay bg-surface-base px-3 py-2 text-xs text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-1 focus:ring-brand"
        rows={2}
        placeholder="Optional notes (shown to experts if requesting changes)"
        value={notes}
        onChange={(e) => setNotes(e.target.value)}
        disabled={!!loading}
      />
      {error && (
        <p className="mt-1 text-xs text-mode-refuse">{error}</p>
      )}
      {sent && !error && (
        <p className="mt-1 text-xs text-mode-advise">Response sent. Resuming workflow...</p>
      )}
      <div className="mt-3 flex gap-2">
        <button
          onClick={() => respond('approve')}
          disabled={!!loading || sent}
          className="rounded bg-mode-advise px-4 py-1.5 text-xs font-semibold text-white hover:opacity-90 disabled:opacity-50"
        >
          {loading === 'approve' ? 'Approving...' : '\u2713 Approve'}
        </button>
        <button
          onClick={() => respond('request_changes')}
          disabled={!!loading || sent}
          className="rounded border border-glow-amber/60 px-4 py-1.5 text-xs font-semibold text-glow-amber hover:bg-glow-amber/10 disabled:opacity-50"
        >
          {loading === 'request_changes' ? 'Sending...' : '\u21ba Request Changes'}
        </button>
      </div>
    </div>
  )
}

// Kanban column definitions
const COLUMNS: { key: KanbanTask['status']; label: string; color: string }[] = [
  { key: 'todo',         label: 'To Do',        color: 'text-text-secondary' },
  { key: 'in_progress',  label: 'In Progress',  color: 'text-brand' },
  { key: 'under_review', label: 'Under Review', color: 'text-mode-warn' },
  { key: 'blocked',      label: 'Blocked',      color: 'text-mode-refuse' },
  { key: 'done',         label: 'Done',         color: 'text-mode-advise' },
]

function TaskCard({ task }: { task: KanbanTask }) {
  return (
    <Card className="mb-2 p-3">
      <p className="text-sm font-medium text-text-primary">{task.title}</p>
      <p className="mt-1 text-xs text-text-secondary">{task.expertName}</p>
      {task.costUsd > 0 && (
        <p className="mt-1 text-xs text-text-disabled">${task.costUsd.toFixed(4)}</p>
      )}
    </Card>
  )
}

// FilesPanel — live file browser fed by useFileStream.
// Shown once implementation/QA phases start producing code files.
// WHY separate section (not a tab): KanbanPage had no tab pattern to
// reuse; a stacked section below the board matches how ApprovalGate and
// the completion banner are already appended below the board.
function FilesPanel({ files, isConnected }: { files: FileEntry[]; isConnected: boolean }) {
  const [selectedPath, setSelectedPath] = useState<string | null>(null)

  if (files.length === 0) return null

  const selected = selectedPath ? files.find((f) => f.filePath === selectedPath) : undefined

  return (
    <div className="mt-6">
      <div className="mb-2 flex items-center justify-between">
        <span className="text-xs font-semibold uppercase tracking-wide text-text-secondary">
          Files
        </span>
        <div className="flex items-center gap-1.5">
          <span className={cn(
            'h-2 w-2 rounded-full',
            isConnected ? 'bg-mode-advise animate-pulse' : 'bg-glow-amber animate-pulse'
          )} />
          <span className="text-[10px] font-medium uppercase tracking-wider text-text-disabled">
            {isConnected ? 'Live' : 'Connecting...'}
          </span>
        </div>
      </div>
      <div className="grid grid-cols-3 gap-4">
        <Card className="col-span-1 max-h-96 overflow-y-auto p-2">
          {files.map((file) => (
            <button
              key={file.filePath}
              onClick={() => setSelectedPath(file.filePath)}
              className={cn(
                'flex w-full items-center justify-between rounded px-2 py-1.5 text-left text-xs',
                selectedPath === file.filePath
                  ? 'bg-surface-overlay text-text-primary'
                  : 'text-text-secondary hover:bg-surface-overlay/60'
              )}
            >
              <span className="truncate">{file.filePath}</span>
              <Badge variant={file.operation === 'create' ? 'success' : 'warn'}>
                {file.operation}
              </Badge>
            </button>
          ))}
        </Card>
        <Card className="col-span-2 max-h-96 overflow-y-auto p-3">
          {selected ? (
            <>
              <div className="mb-2 flex items-center justify-between">
                <p className="text-xs font-medium text-text-primary">{selected.filePath}</p>
                <span className="text-xs text-text-disabled">{selected.linesOfCode} lines</span>
              </div>
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
    </div>
  )
}

function KanbanPage() {
  const { id } = useParams<{ id: string }>()

  // Workflow metadata: title, status, cost — poll every 10s (lightweight)
  const { data: workflow } = useQuery({
    queryKey: ['workflow', id],
    queryFn: () => getWorkflow(id!),
    enabled: !!id,
    refetchInterval: 10000,
  })

  // Live Kanban state via SSE — replaces 5s polling
  // WHY SSE: instant updates when expert posts artifact or changes status.
  // 5s polling = user sees stale board for up to 5s after each update.
  const stream = useKanbanStream(id ?? null)
  const tasks = stream.tasks
  const isLoading = !stream.isConnected && tasks.length === 0 && !stream.isDone

  // Live file browser — separate SSE stream, scoped to code_artifact_produced
  // and wave_completed events only (see useFileStream).
  const fileStream = useFileStream(id ?? null)

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-semibold text-text-primary">
            {workflow?.title ?? 'Workflow'}
          </h1>
          {/* SSE connection indicator */}
          <div className="flex items-center gap-1.5">
            <span className={cn(
              'h-2 w-2 rounded-full',
              stream.isDone ? 'bg-mode-advise'
              : stream.isConnected ? 'bg-mode-advise animate-pulse'
              : 'bg-glow-amber animate-pulse'
            )} />
            <span className="text-[10px] font-medium uppercase tracking-wider text-text-disabled">
              {stream.isDone ? 'Complete' : stream.isConnected ? 'Live' : 'Connecting...'}
            </span>
          </div>
        </div>
        <div className="mt-1 flex items-center gap-3">
          {workflow && (
            <>
              <Badge variant="neutral">{workflow.currentPhase.replace(/_/g, ' ')}</Badge>
              <Badge variant={workflow.status === 'running' ? 'brand' : 'neutral'}>
                {workflow.status.replace(/_/g, ' ')}
              </Badge>
              <span className="text-xs text-text-disabled">
                ${workflow.costSpentUsd.toFixed(4)} / ${workflow.costBudgetUsd.toFixed(2)}
              </span>
            </>
          )}
        </div>
      </div>

      {/* Kanban board */}
      {isLoading ? (
        <div className="grid grid-cols-5 gap-4">
          {COLUMNS.map((col) => (
            <Skeleton key={col.key} className="h-48" />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-5 gap-4">
          {COLUMNS.map((col) => {
            const colTasks = tasks.filter((t) => t.status === col.key)
            return (
              <div key={col.key}>
                <div className="mb-2 flex items-center justify-between">
                  <span className={`text-xs font-semibold uppercase tracking-wide ${col.color}`}>
                    {col.label}
                  </span>
                  <span className="text-xs text-text-disabled">{colTasks.length}</span>
                </div>
                <div className="min-h-24 rounded-lg bg-surface-overlay p-2">
                  {colTasks.map((task) => (
                    <TaskCard key={task.id} task={task} />
                  ))}
                  {colTasks.length === 0 && (
                    <p className="py-4 text-center text-xs text-text-disabled">Empty</p>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Approval gate — interactive Approve/Request Changes controls */}
      {workflow?.status === 'paused_for_approval' && (
        <ApprovalGate
          workflowId={id!}
          approvalId={stream.approvalGate?.approvalId ?? ''}
          gateName={stream.approvalGate?.gateName ?? 'approval'}
          summary={stream.approvalGate?.summary ?? 'Review and approve to continue.'}
        />
      )}

      {/* Live file browser — populated once experts start producing code */}
      <FilesPanel files={fileStream.files} isConnected={fileStream.isConnected} />

      {/* Completion notice */}
      {stream.isDone && (
        <div className="mt-4 rounded-lg border border-mode-advise/30 bg-mode-advise/10 p-4 text-center">
          <p className="text-sm font-medium text-mode-advise">{'\u2705'} Workflow Complete</p>
        </div>
      )}
    </div>
  )
}

export const Component = KanbanPage
