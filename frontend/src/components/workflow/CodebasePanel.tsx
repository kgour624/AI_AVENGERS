import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  addCodebaseFile,
  decideCodebaseFile,
  decideCodebaseFilesBulk,
  downloadCodebasePatch,
  generateCodebasePatch,
  getCodebaseFiles,
  getCodebasePatch,
  runWorkflow,
  startWorkflow,
  suggestCodebaseFiles,
  type CodebaseFile,
  type Workflow,
} from '@/api/workflows'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'

/**
 * CodebasePanel — the human approval loop for an existing-codebase workflow
 * (phase 3D).
 *
 * THE RULE THIS PANEL ENFORCES: the system proposes files, the client decides,
 * and only approved files become readable by an expert. Nothing here approves
 * anything automatically — "Approve all" is still a client click, and the
 * backend refuses to overwrite a decision that was already made.
 *
 * Suggestions the client rejected stay rejected even if a later ranking run
 * proposes the same path again, so this list is the record of decisions, not a
 * queue that keeps re-surfacing the same file.
 */

function statusBadge(status: CodebaseFile['status']) {
  if (status === 'approved') return <Badge variant="success">APPROVED</Badge>
  if (status === 'rejected') return <Badge variant="danger">REJECTED</Badge>
  return <Badge variant="warning">PENDING</Badge>
}

export function CodebasePanel({
  workflowId,
  workflowStatus,
}: {
  workflowId: string
  // workflowStatus drives the start control below. Passed in rather than fetched
  // again because KanbanPage already polls the workflow every 10s.
  workflowStatus?: Workflow['status']
}) {
  const queryClient = useQueryClient()
  const [newPath, setNewPath] = useState('')
  const [actionError, setActionError] = useState<string | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['workflow', workflowId, 'codebase'],
    queryFn: () => getCodebaseFiles(workflowId),
  })

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ['workflow', workflowId, 'codebase'] })

  const suggestMutation = useMutation({
    mutationFn: () => suggestCodebaseFiles(workflowId),
    onSuccess: invalidate,
    onError: () => setActionError('Could not generate suggestions. Add requirement text first.'),
  })

  const addMutation = useMutation({
    mutationFn: (path: string) => addCodebaseFile(workflowId, path),
    onSuccess: () => {
      setNewPath('')
      setActionError(null)
      invalidate()
    },
    onError: () => setActionError("That path is not part of the repository's stored tree."),
  })

  const decideMutation = useMutation({
    mutationFn: ({ id, decision }: { id: string; decision: 'approve' | 'reject' }) =>
      decideCodebaseFile(workflowId, id, decision),
    onSuccess: invalidate,
    onError: () => setActionError('Could not record that decision.'),
  })

  const bulkMutation = useMutation({
    mutationFn: ({ ids, decision }: { ids: string[]; decision: 'approve' | 'reject' }) =>
      decideCodebaseFilesBulk(workflowId, ids, decision),
    onSuccess: invalidate,
    onError: () => setActionError('Could not record those decisions.'),
  })

  // Delivery (3F). Kept on the same tab as the approvals because a patch is the
  // consequence of that working set, and a reviewer needs both in one place.
  const { data: delivery } = useQuery({
    queryKey: ['workflow', workflowId, 'codebase', 'patch'],
    queryFn: () => getCodebasePatch(workflowId),
    retry: false,
  })

  const patchMutation = useMutation({
    mutationFn: () => generateCodebasePatch(workflowId),
    onSuccess: () => {
      setActionError(null)
      queryClient.invalidateQueries({ queryKey: ['workflow', workflowId, 'codebase', 'patch'] })
    },
    onError: () => setActionError('Could not generate a patch yet. Start the workflow first.'),
  })

  const downloadMutation = useMutation({
    mutationFn: async () => {
      const blob = await downloadCodebasePatch(workflowId)
      const url = URL.createObjectURL(blob)
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = `workflow-${workflowId.slice(0, 8)}.patch`
      anchor.click()
      // Revoke on the next tick: revoking immediately can cancel the download
      // in some browsers before it has read the blob.
      setTimeout(() => URL.revokeObjectURL(url), 0)
    },
    onError: () => setActionError('Could not download the patch.'),
  })

  // Starting the run is the step AFTER approval, and it belongs here rather than
  // in the create dialog.
  //
  // WHY: an existing-codebase workflow must have an approved readable set before
  // the runner seeds a workspace. The create dialog used to create-and-run in one
  // click, so every such workflow started with an empty approved set and failed
  // immediately ("no approved files to read") — with the tab that would have let
  // the client fix it not even visible. Creating a draft and starting it from the
  // screen where the files are approved removes that trap.
  const startMutation = useMutation({
    mutationFn: async () => {
      await startWorkflow(workflowId)
      await runWorkflow(workflowId)
    },
    onSuccess: () => {
      setActionError(null)
      queryClient.invalidateQueries({ queryKey: ['workflow', workflowId] })
    },
    onError: () =>
      setActionError('Could not start the workflow. It must still be a draft with at least one approved file.'),
  })

  if (isLoading) {
    return <div className="h-32 animate-pulse rounded-lg bg-surface-overlay" />
  }

  const files = data?.files ?? []
  const pending = files.filter((file) => file.status === 'pending')
  const decided = files.filter((file) => file.status !== 'pending')
  const approvedCount = data?.approved ?? 0
  const busy = decideMutation.isPending || bulkMutation.isPending

  return (
    <div className="space-y-4">
      {workflowStatus === 'draft' && (
        <Card className="p-4">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <span className="text-xs font-semibold uppercase tracking-wide text-text-secondary">
                Start
              </span>
              <p className="mt-1 text-[11px] text-text-disabled">
                {approvedCount > 0
                  ? `The experts will read only the ${approvedCount} approved file${
                      approvedCount === 1 ? '' : 's'
                    }, and a patch will be generated against the commit this started from.`
                  : 'Approve at least one file first — a run with nothing approved has no code to read and would fail.'}
              </p>
            </div>
            <Button
              size="sm"
              disabled={approvedCount === 0}
              isLoading={startMutation.isPending}
              onClick={() => startMutation.mutate()}
            >
              Start workflow
            </Button>
          </div>
        </Card>
      )}

      <Card className="p-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <span className="text-xs font-semibold uppercase tracking-wide text-text-secondary">
              Readable files
            </span>
            <Badge variant={approvedCount > 0 ? 'success' : 'warning'}>
              {approvedCount} approved
            </Badge>
            {pending.length > 0 && (
              <span className="text-xs text-text-disabled">{pending.length} awaiting your review</span>
            )}
          </div>
          <Button
            size="sm"
            variant="secondary"
            onClick={() => suggestMutation.mutate()}
            isLoading={suggestMutation.isPending}
          >
            Find relevant files
          </Button>
        </div>

        <div className="mt-3 flex gap-2">
          <input
            value={newPath}
            onChange={(e) => setNewPath(e.target.value)}
            placeholder="src/services/invoicing.ts"
            className="flex-1 rounded-md border border-glass-border bg-surface-overlay px-3 py-1.5 text-xs text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-brand/40"
          />
          <Button
            size="sm"
            onClick={() => newPath.trim() && addMutation.mutate(newPath.trim())}
            disabled={!newPath.trim()}
            isLoading={addMutation.isPending}
          >
            Add &amp; approve
          </Button>
        </div>

        {actionError && <p className="mt-2 text-xs text-mode-refuse">{actionError}</p>}
      </Card>

      {pending.length > 0 && (
        <Card className="p-4">
          <div className="mb-2 flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wide text-text-secondary">
              Awaiting review ({pending.length})
            </span>
            <div className="flex gap-2">
              <Button
                size="sm"
                variant="secondary"
                disabled={busy}
                onClick={() =>
                  bulkMutation.mutate({ ids: pending.map((f) => f.id), decision: 'reject' })
                }
              >
                Reject all
              </Button>
              <Button
                size="sm"
                disabled={busy}
                onClick={() =>
                  bulkMutation.mutate({ ids: pending.map((f) => f.id), decision: 'approve' })
                }
              >
                Approve all
              </Button>
            </div>
          </div>

          <ul className="divide-y divide-surface-border">
            {pending.map((file) => (
              <li key={file.id} className="flex items-start justify-between gap-3 py-2">
                <div className="min-w-0">
                  <p className="truncate text-xs font-medium text-text-primary">{file.path}</p>
                  {/* The reason is shown, not hidden: a human can only judge a
                      suggestion if they can see why it was made. */}
                  <p className="mt-0.5 text-[11px] text-text-disabled">
                    {file.reason || 'no reason recorded'}
                    {file.score > 0 ? ` · score ${file.score.toFixed(2)}` : ''}
                  </p>
                </div>
                <div className="flex shrink-0 gap-2">
                  <Button
                    size="sm"
                    variant="secondary"
                    disabled={busy}
                    onClick={() => decideMutation.mutate({ id: file.id, decision: 'reject' })}
                  >
                    Reject
                  </Button>
                  <Button
                    size="sm"
                    disabled={busy}
                    onClick={() => decideMutation.mutate({ id: file.id, decision: 'approve' })}
                  >
                    Approve
                  </Button>
                </div>
              </li>
            ))}
          </ul>
        </Card>
      )}

      {decided.length > 0 && (
        <Card className="p-4">
          <span className="text-xs font-semibold uppercase tracking-wide text-text-secondary">
            Decisions ({decided.length})
          </span>
          <ul className="mt-2 divide-y divide-surface-border">
            {decided.map((file) => (
              <li key={file.id} className="flex items-center justify-between gap-3 py-2">
                <div className="flex min-w-0 items-center gap-2">
                  {statusBadge(file.status)}
                  <span className="truncate text-xs text-text-secondary">{file.path}</span>
                </div>
                <div className="flex shrink-0 items-center gap-2">
                  <span className="text-[10px] uppercase tracking-wider text-text-disabled">
                    {file.source === 'client' ? 'added by you' : 'suggested'}
                  </span>
                  <Button
                    size="sm"
                    variant="ghost"
                    disabled={busy}
                    onClick={() =>
                      decideMutation.mutate({
                        id: file.id,
                        decision: file.status === 'approved' ? 'reject' : 'approve',
                      })
                    }
                  >
                    {file.status === 'approved' ? 'Revoke' : 'Approve'}
                  </Button>
                </div>
              </li>
            ))}
          </ul>
        </Card>
      )}

      {files.length === 0 && (
        <Card className="p-6">
          <p className="text-center text-xs text-text-disabled">
            No files in the working set yet. Use “Find relevant files” to get ranked
            suggestions from this repository, or add a path yourself.
          </p>
        </Card>
      )}

      <Card className="p-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="min-w-0">
            <span className="text-xs font-semibold uppercase tracking-wide text-text-secondary">
              Delivery
            </span>
            <p className="mt-0.5 text-[11px] text-text-disabled">
              {delivery?.baseCommitSha
                ? `Patch against commit ${delivery.baseCommitSha.slice(0, 7)}`
                : 'A patch is generated against the commit your work started from.'}
            </p>
          </div>
          <div className="flex shrink-0 gap-2">
            <Button
              size="sm"
              variant="secondary"
              onClick={() => patchMutation.mutate()}
              isLoading={patchMutation.isPending}
            >
              {delivery?.generatedAt ? 'Regenerate patch' : 'Generate patch'}
            </Button>
            <Button
              size="sm"
              disabled={!delivery?.generatedAt || delivery.patchBytes === 0}
              onClick={() => downloadMutation.mutate()}
              isLoading={downloadMutation.isPending}
            >
              Download .patch
            </Button>
          </div>
        </div>

        {delivery?.generatedAt && (
          <div className="mt-3">
            <p className="text-[11px] text-text-disabled">
              {delivery.changedFiles.length} file{delivery.changedFiles.length === 1 ? '' : 's'} changed
              {' · '}
              {delivery.patchBytes} bytes
            </p>
            {delivery.changedFiles.length > 0 && (
              <ul className="mt-2 max-h-40 overflow-y-auto rounded border border-glass-border">
                {delivery.changedFiles.map((file) => (
                  <li
                    key={`${file.status}-${file.path}`}
                    className="flex items-center gap-2 border-b border-glass-border px-2 py-1 last:border-b-0"
                  >
                    <span className="w-6 shrink-0 text-[10px] font-semibold uppercase text-text-disabled">
                      {file.status}
                    </span>
                    <span className="truncate text-xs text-text-secondary">{file.path}</span>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}
      </Card>
    </div>
  )
}
