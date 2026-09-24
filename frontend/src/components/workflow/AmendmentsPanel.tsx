import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  listAmendments,
  respondToAmendment,
  type Amendment,
  type AmendmentStatus,
} from '@/api/amendments'
import { queryKeys } from '@/api/queryKeys'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { handleAPIError } from '@/utils/errors'

/**
 * AmendmentsPanel — §7.5's approval surface
 * (docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md).
 *
 * NOT KanbanPage.tsx's ApprovalGate. That component only ever renders while
 * workflow.status === 'paused_for_approval' and reads its one approval id off
 * the kanban_approval SSE event — approving an amendment must never pause the
 * workflow (amendment_handler.go's whole reason for being a separate
 * endpoint), so there is no SSE event to key off and no gate to wait for.
 * This is a plain polled list instead: GET /workflows/:id/amendments.
 */

const STATUS_TABS: { key: AmendmentStatus | 'all'; label: string }[] = [
  { key: 'pending', label: 'Pending' },
  { key: 'approved', label: 'Approved' },
  { key: 'rejected', label: 'Rejected' },
  { key: 'all', label: 'All' },
]

function DiffBlock({ amendment }: { amendment: Amendment }) {
  if (!amendment.oldText) {
    // An append (a new criterion or statement row) has nothing to diff against.
    return (
      <pre className="mt-2 overflow-x-auto whitespace-pre-wrap rounded bg-mode-advise/10 p-2 text-[11px] text-mode-advise">
        + {amendment.newText}
      </pre>
    )
  }
  return (
    <div className="mt-2 space-y-1">
      <pre className="overflow-x-auto whitespace-pre-wrap rounded bg-mode-refuse/10 p-2 text-[11px] text-mode-refuse">
        - {amendment.oldText}
      </pre>
      <pre className="overflow-x-auto whitespace-pre-wrap rounded bg-mode-advise/10 p-2 text-[11px] text-mode-advise">
        + {amendment.newText}
      </pre>
    </div>
  )
}

function AmendmentCard({ workflowId, amendment }: { workflowId: string; amendment: Amendment }) {
  const queryClient = useQueryClient()
  const [editing, setEditing] = useState(false)
  const [editedText, setEditedText] = useState(amendment.newText)
  const [notes, setNotes] = useState('')
  const [error, setError] = useState<string | null>(null)

  const invalidate = () => {
    // Every status tab this amendment could now belong to.
    queryClient.invalidateQueries({ queryKey: ['workflows', workflowId, 'amendments'] })
  }

  const respond = useMutation({
    mutationFn: (args: { decision: 'approve' | 'approve_with_edit' | 'reject' }) =>
      respondToAmendment(workflowId, amendment.approvalId, args.decision, {
        editedText: args.decision === 'approve_with_edit' ? editedText : undefined,
        notes: notes || undefined,
      }),
    onSuccess: invalidate,
    onError: (err) => setError(handleAPIError(err)),
  })

  const isPending = amendment.status === 'pending'

  return (
    <Card className="p-3">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className="text-sm font-medium text-text-primary">{amendment.summary}</p>
          <p className="mt-0.5 text-xs text-text-disabled">
            {amendment.target}
            {amendment.proposedByExpert && ` \u00b7 ${amendment.proposedByExpert}`}
          </p>
        </div>
        <Badge
          variant={
            amendment.status === 'approved'
              ? 'success'
              : amendment.status === 'rejected'
                ? 'danger'
                : 'warning'
          }
        >
          {amendment.status}
        </Badge>
      </div>

      {amendment.reason && (
        <p className="mt-2 text-xs text-text-secondary">{amendment.reason}</p>
      )}

      {editing ? (
        <textarea
          value={editedText}
          onChange={(e) => setEditedText(e.target.value)}
          rows={4}
          className="mt-2 w-full rounded border border-surface-overlay bg-surface-base px-2 py-1.5 font-mono text-[11px] text-text-primary focus:outline-none focus:ring-1 focus:ring-brand"
        />
      ) : (
        <DiffBlock amendment={amendment} />
      )}

      {isPending && (
        <textarea
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          placeholder="Optional notes"
          rows={1}
          className="mt-2 w-full rounded border border-surface-overlay bg-surface-base px-2 py-1 text-[11px] text-text-primary placeholder:text-text-disabled focus:outline-none"
        />
      )}

      {error && <p className="mt-1 text-xs text-mode-refuse">{error}</p>}

      {isPending && (
        <div className="mt-2 flex gap-2">
          {editing ? (
            <>
              <Button
                size="sm"
                isLoading={respond.isPending}
                onClick={() => respond.mutate({ decision: 'approve_with_edit' })}
              >
                Approve edited
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setEditing(false)}>
                Cancel edit
              </Button>
            </>
          ) : (
            <>
              <Button
                size="sm"
                isLoading={respond.isPending}
                onClick={() => respond.mutate({ decision: 'approve' })}
              >
                \u2713 Approve
              </Button>
              <Button size="sm" variant="secondary" onClick={() => setEditing(true)}>
                Edit & approve
              </Button>
              <Button
                size="sm"
                variant="danger"
                isLoading={respond.isPending}
                onClick={() => respond.mutate({ decision: 'reject' })}
              >
                Reject
              </Button>
            </>
          )}
        </div>
      )}
    </Card>
  )
}

export function AmendmentsPanel({ workflowId }: { workflowId: string }) {
  const [tab, setTab] = useState<AmendmentStatus | 'all'>('pending')

  const { data: amendments = [], isLoading } = useQuery({
    queryKey: queryKeys.workflows.amendments(workflowId, tab),
    queryFn: () => listAmendments(workflowId, tab),
    // Amendments arrive from a chat tool call with no push notification of
    // their own — poll while this panel is mounted, same interval the
    // Kanban page already uses for the workflow/blackboard queries.
    refetchInterval: 10000,
  })

  return (
    <div className="mt-6">
      <div className="mb-2 flex items-center justify-between">
        <span className="text-xs font-semibold uppercase tracking-wide text-text-secondary">
          Amendments
        </span>
        <div className="flex gap-1">
          {STATUS_TABS.map((t) => (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className={
                tab === t.key
                  ? 'rounded px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide text-brand'
                  : 'rounded px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide text-text-disabled hover:text-text-secondary'
              }
            >
              {t.label}
            </button>
          ))}
        </div>
      </div>

      {isLoading && <p className="text-xs text-text-disabled">Loading...</p>}
      {!isLoading && amendments.length === 0 && (
        <Card className="p-4">
          <p className="text-center text-xs text-text-disabled">
            {tab === 'pending'
              ? 'No amendments waiting on you.'
              : `No ${tab === 'all' ? '' : tab} amendments.`}
          </p>
        </Card>
      )}

      <div className="space-y-2">
        {amendments.map((a) => (
          <AmendmentCard key={a.approvalId} workflowId={workflowId} amendment={a} />
        ))}
      </div>
    </div>
  )
}
