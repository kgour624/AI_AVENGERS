import { useParams, useNavigate } from 'react-router-dom'
import { useMemo, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getBlackboard, getWorkflow, respondToApproval, cancelWorkflow, retryTask } from '@/api/workflows'
import { useKanbanStream } from '@/hooks/useKanbanStream'
import { useFileStream } from '@/hooks/useFileStream'
import type { BlackboardEvent, KanbanTask } from '@/api/workflows'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { Tabs, type TabItem } from '@/components/ui/Tabs'
import { cn } from '@/utils/cn'
import { WorkflowChatPanel } from '@/components/workflow/WorkflowChatPanel'
import { ArtifactProof, parseArtifactVerification, type ArtifactVerification } from '@/components/workflow/ArtifactProof'
import { AmendmentsPanel } from '@/components/workflow/AmendmentsPanel'
import { DeliveryPanel } from '@/components/workflow/DeliveryPanel'
import { DownloadDesignPackageButton } from '@/components/workflow/DownloadDesignPackageButton'
import { ActivityLog } from '@/components/workflow/ActivityLog'
import { FilesPanel } from '@/components/workflow/FilesPanel'
import { CodebasePanel } from '@/components/workflow/CodebasePanel'

// CancelWorkflowButton — renders a cancel button with confirmation dialog.
// WHY confirmation: cancelling a workflow is destructive and cannot be undone.
// Cost already spent is not refunded, and all work is lost.
function CancelWorkflowButton({ workflowId, workflowTitle }: { workflowId: string; workflowTitle: string }) {
  const [showConfirm, setShowConfirm] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const handleCancel = async () => {
    setLoading(true)
    setError(null)
    try {
      await cancelWorkflow(workflowId)
      // Invalidate queries to refresh workflow list
      queryClient.invalidateQueries({ queryKey: ['workflow', workflowId] })
      queryClient.invalidateQueries({ queryKey: ['workflows'] })
      // Navigate back to workflows list
      navigate('/workflows')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to cancel workflow')
      setLoading(false)
    }
  }

  if (!showConfirm) {
    return (
      <button
        onClick={() => setShowConfirm(true)}
        className="flex items-center gap-1.5 rounded-md border border-mode-refuse/60 bg-mode-refuse/10 px-3 py-1.5 text-xs font-medium text-mode-refuse hover:bg-mode-refuse/20 transition-colors"
        title="Cancel this workflow"
      >
        <span>{'\u23f9'}</span>
        <span>Cancel Workflow</span>
      </button>
    )
  }

  return (
    <div className="flex items-center gap-2 rounded-md border border-mode-refuse/60 bg-mode-refuse/10 px-3 py-1.5">
      <span className="text-xs font-medium text-mode-refuse">Cancel "{workflowTitle}"?</span>
      {error && (
        <span className="text-xs text-mode-refuse">{error}</span>
      )}
      <button
        onClick={handleCancel}
        disabled={loading}
        className="rounded bg-mode-refuse px-2 py-1 text-xs font-semibold text-white hover:opacity-90 disabled:opacity-50"
      >
        {loading ? 'Cancelling...' : 'Yes, Cancel'}
      </button>
      <button
        onClick={() => {
          setShowConfirm(false)
          setError(null)
        }}
        disabled={loading}
        className="rounded border border-surface-border px-2 py-1 text-xs font-medium text-text-secondary hover:bg-surface-overlay disabled:opacity-50"
      >
        No, Keep Running
      </button>
    </div>
  )
}

// ApprovalGate component — renders Approve/Request Changes buttons.
// Shown when workflow.status === 'paused_for_approval'.
// GENERIC_OPTIONS: the allowance values the client can pick.
// Capped at 30 to match MaxGenericAllowancePct in the backend and the
// workflows_generic_allowance_pct_check constraint (migration 017) — trained
// knowledge must always remain the majority contributor.
const GENERIC_OPTIONS = [0, 5, 10, 20, 30]

function ApprovalGate({
  workflowId,
  approvalId,
  gateName,
  summary,
  currentGenericPct,
}: {
  workflowId: string
  approvalId: string
  gateName: string
  summary: string
  currentGenericPct: number
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

  // genericPct: only sent with 'request_changes'. Approving never changes the
  // dial — approval means "this output is acceptable as it stands".
  const [genericPct, setGenericPct] = useState(currentGenericPct)

  // Feature #18: View Full Artifact - expandable section for reviewing complete deliverable
  const [showArtifact, setShowArtifact] = useState(false)

  // Fetch latest artifact from blackboard for full content review
  const { data: blackboard } = useQuery({
    queryKey: ['blackboard', workflowId],
    queryFn: () => getBlackboard(workflowId),
    enabled: !!workflowId,
  })

  // Find the most recent artifact event before this approval
  const latestArtifact = blackboard?.events
    ?.filter((e) => ARTIFACT_TYPES.has(e.eventType))
    ?.sort((a, b) => new Date(b.postedAt).getTime() - new Date(a.postedAt).getTime())[0]

  const respond = async (decision: 'approve' | 'request_changes') => {
    if (!approvalId) {
      setError('Approval ID not yet received from server. Please wait a moment and try again.')
      return
    }
    setLoading(decision)
    setError(null)
    try {
      await respondToApproval(
        workflowId,
        approvalId,
        decision,
        notes || undefined,
        decision === 'request_changes' ? genericPct : undefined
      )
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

      {/* Generic-knowledge dial. Applies only to "Request Changes": the design
          is produced again with this ceiling. 0 keeps experts on trained +
          peer knowledge only, which is the default posture. */}
      <div className="mt-3 rounded border border-surface-overlay bg-surface-base/60 p-3">
        <p className="text-xs font-medium text-text-primary">
          Not satisfied? Allow some generic knowledge and re-run the design
        </p>
        <p className="mt-1 text-[11px] text-text-disabled">
          Experts currently use ONLY their trained material and each other's work.
          Raising this lets them fill gaps with general knowledge, tagged [GENERIC].
          Capped at 30% so trained knowledge stays the majority.
        </p>
        <div className="mt-2 flex items-center gap-2">
          <label htmlFor="generic-pct" className="text-[11px] text-text-secondary">
            Generic allowance
          </label>
          <select
            id="generic-pct"
            value={genericPct}
            onChange={(e) => setGenericPct(Number(e.target.value))}
            disabled={!!loading || sent}
            className="rounded border border-surface-overlay bg-surface-base px-2 py-1 text-xs text-text-primary disabled:opacity-50"
          >
            {GENERIC_OPTIONS.map((pct) => (
              <option key={pct} value={pct}>
                {pct === 0 ? '0% — trained knowledge only' : `${pct}%`}
              </option>
            ))}
          </select>
          {currentGenericPct > 0 && (
            <span className="text-[11px] text-glow-amber">
              currently {currentGenericPct}%
            </span>
          )}
        </div>
      </div>

      {/* Feature #18: View Full Artifact - expandable section for reviewing complete deliverable */}
      {latestArtifact && (
        <div className="mt-3 rounded border border-surface-overlay bg-surface-base/60 p-3">
          <button
            onClick={() => setShowArtifact(!showArtifact)}
            className="flex w-full items-center justify-between text-left text-xs font-medium text-text-primary hover:text-brand"
          >
            <span>📄 View Full Artifact</span>
            <span className="text-text-disabled">{showArtifact ? '▲' : '▼'}</span>
          </button>
          {showArtifact && (
            <div className="mt-2 rounded border border-surface-overlay bg-surface-base p-2">
              <div className="mb-2 flex items-center justify-between">
                <span className="text-[10px] font-semibold uppercase tracking-wider text-text-disabled">
                  {latestArtifact.eventType.replace(/_/g, ' ')}
                </span>
                <span className="text-[10px] text-text-disabled">
                  {new Date(latestArtifact.postedAt).toLocaleString()}
                </span>
              </div>
              <div className="max-h-96 overflow-y-auto">
                {Object.entries(latestArtifact.content ?? {}).map(([key, value]) => (
                  <div key={key} className="mb-3">
                    <p className="mb-1 text-[10px] font-semibold uppercase tracking-wider text-text-disabled">
                      {key.replace(/_/g, ' ').replace(/([A-Z])/g, ' $1')}
                    </p>
                    <pre className="overflow-x-auto whitespace-pre-wrap rounded bg-surface-overlay/60 p-2 text-xs text-text-secondary">
                      {renderContentValue(value)}
                    </pre>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
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
          {loading === 'request_changes'
            ? 'Sending...'
            : `\u21ba Request Changes & Re-run (${genericPct}% generic)`}
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

function TaskCard({ task, onSelect, workflowId }: { task: KanbanTask; onSelect: () => void; workflowId: string }) {
  const [retrying, setRetrying] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const queryClient = useQueryClient()

  const handleRetry = async (e: React.MouseEvent) => {
    e.stopPropagation() // Prevent card click
    setRetrying(true)
    setError(null)
    try {
      await retryTask(workflowId, task.id)
      // Invalidate queries to refresh Kanban board
      queryClient.invalidateQueries({ queryKey: ['workflow', workflowId] })
      queryClient.invalidateQueries({ queryKey: ['blackboard', workflowId] })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Retry failed')
    } finally {
      setRetrying(false)
    }
  }

  const isFailed = task.status === 'blocked' || task.status === 'cancelled'

  return (
    <Card
      onClick={onSelect}
      className={cn(
        'mb-2 cursor-pointer p-3 hover:border-glow-purple/40',
        isFailed && 'border-mode-refuse/40 bg-mode-refuse/5'
      )}
      title="Show this expert's deliverables"
    >
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium text-text-primary">{task.title}</p>
          <p className="mt-1 text-xs text-text-secondary">{task.expertName}</p>
          {task.costUsd > 0 && (
            <p className="mt-1 text-xs text-text-disabled">${task.costUsd.toFixed(4)}</p>
          )}
          {error && (
            <p className="mt-1 text-xs text-mode-refuse">{error}</p>
          )}
        </div>
        {isFailed && (
          <button
            onClick={handleRetry}
            disabled={retrying}
            className="shrink-0 rounded border border-brand/60 bg-brand/10 px-2 py-1 text-xs font-medium text-brand hover:bg-brand/20 disabled:opacity-50"
            title="Retry this failed task"
          >
            {retrying ? '...' : '🔄 Retry'}
          </button>
        )}
      </div>
    </Card>
  )
}

// Blackboard event types that represent an expert deliverable worth reading.
// Same set the backend forwards as kanban_artifact (see kanban_sse.go).
const ARTIFACT_TYPES = new Set([
  'architecture_decision',
  'data_model_proposed',
  'api_contract_proposed',
  'module_design_proposed',
  'code_artifact_produced',
  'test_case_proposed',
  'requirement_captured',
])

// renderContentValue turns one artifact content field into readable text.
// Strings are shown as-is (they are usually prose or code); anything else is
// pretty-printed JSON rather than "[object Object]".
function renderContentValue(value: unknown): string {
  if (typeof value === 'string') return value
  return JSON.stringify(value, null, 2)
}

// ArtifactsPanel — reads what the experts actually produced.
//
// WHY this exists: the Kanban board only ever showed a task's title and
// expert name. The deliverable itself (the PRD, the architecture decision,
// the API contract) lives in blackboard_events and had no screen at all —
// getBlackboard() existed in the API client but no page called it. So a
// finished workflow looked empty even though the work was done.
function ArtifactsPanel({
  events,
  expertNames,
  filterExpertId,
  onClearFilter,
  verifications,
}: {
  events: BlackboardEvent[]
  expertNames: Map<string, string>
  filterExpertId: string | null
  onClearFilter: () => void
  // Keyed by the artifact event id the verification refers to. A missing entry
  // means "not checked" and the proof block says so — it never shows a pass for
  // a check that never ran.
  verifications: Map<string, ArtifactVerification>
}) {
  const [selectedId, setSelectedId] = useState<string | null>(null)

  const artifacts = events.filter(
    (e) =>
      ARTIFACT_TYPES.has(e.eventType) &&
      (!filterExpertId || e.postedByExpertId === filterExpertId)
  )

  if (events.length === 0) return null

  const selected = selectedId
    ? artifacts.find((a) => a.id === selectedId)
    : artifacts[artifacts.length - 1] // default to the newest

  return (
    <div className="mt-6">
      <div className="mb-2 flex items-center justify-between">
        <span className="text-xs font-semibold uppercase tracking-wide text-text-secondary">
          Deliverables ({artifacts.length})
        </span>
        {filterExpertId && (
          <button
            onClick={onClearFilter}
            className="text-[10px] font-medium uppercase tracking-wider text-brand hover:underline"
          >
            Showing {expertNames.get(filterExpertId) ?? 'one expert'} — show all
          </button>
        )}
      </div>

      {artifacts.length === 0 ? (
        <Card className="p-4">
          <p className="text-center text-xs text-text-disabled">
            No deliverables yet for this selection.
          </p>
        </Card>
      ) : (
        <div className="grid grid-cols-3 gap-4">
          <Card className="col-span-1 max-h-96 overflow-y-auto p-2">
            {artifacts.map((a) => (
              <button
                key={a.id}
                onClick={() => setSelectedId(a.id)}
                className={cn(
                  'mb-1 w-full rounded px-2 py-1.5 text-left text-xs',
                  selected?.id === a.id
                    ? 'bg-surface-overlay text-text-primary'
                    : 'text-text-secondary hover:bg-surface-overlay/60'
                )}
              >
                <span className="block truncate font-medium">
                  {a.eventType.replace(/_/g, ' ')}
                </span>
                <span className="block truncate text-[10px] text-text-disabled">
                  {(a.postedByExpertId && expertNames.get(a.postedByExpertId)) || 'system'}
                </span>
              </button>
            ))}
          </Card>

          <Card className="col-span-2 max-h-96 overflow-y-auto p-3">
            {selected ? (
              <>
                <div className="mb-2 flex items-center justify-between">
                  <p className="text-xs font-medium text-text-primary">
                    {selected.eventType.replace(/_/g, ' ')}
                  </p>
                  <span className="text-xs text-text-disabled">
                    {(selected.postedByExpertId && expertNames.get(selected.postedByExpertId)) ||
                      'system'}
                  </span>
                </div>
                <ArtifactProof verification={verifications.get(selected.id)} />
                {Object.entries(selected.content ?? {}).map(([key, value]) => (
                  <div key={key} className="mb-3">
                    {/* baseAPI's response interceptor camelizes every key, so a
                        field posted as "file_path" arrives as "filePath" —
                        split on capitals as well as underscores. */}
                    <p className="mb-1 text-[10px] font-semibold uppercase tracking-wider text-text-disabled">
                      {key.replace(/_/g, ' ').replace(/([A-Z])/g, ' $1')}
                    </p>
                    <pre className="overflow-x-auto whitespace-pre-wrap text-xs text-text-secondary">
                      {renderContentValue(value)}
                    </pre>
                  </div>
                ))}
              </>
            ) : (
              <p className="py-4 text-center text-xs text-text-disabled">
                Select a deliverable to read it
              </p>
            )}
          </Card>
        </div>
      )}
    </div>
  )
}

// Workspace tabs (feature #28). 'board' is the default so the task board and
// the approval gate stay front-and-centre; the other three are read-only views
// that previously sat below the fold in one long scroll.
type WorkspaceTab = 'board' | 'deliverables' | 'activity' | 'files' | 'codebase'

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

  // Deliverables (blackboard artifacts). Polled on the same 10s interval as
  // the workflow metadata above — the SSE stream reports that an artifact was
  // posted but drops its content, and this panel needs the content itself.
  const { data: blackboard } = useQuery({
    queryKey: ['blackboard', id],
    queryFn: () => getBlackboard(id!),
    enabled: !!id,
    refetchInterval: 10000,
  })

  // Clicking a task card narrows the deliverables panel to that expert.
  const [filterExpertId, setFilterExpertId] = useState<string | null>(null)

  // Explicit tuple type: .map() with an array literal infers string[], not
  // [string, string], which does not satisfy the Map constructor.
  const expertNames = new Map<string, string>(
    tasks.map((t) => [t.assignedExpertId, t.expertName] as [string, string])
  )

  // The set of experts this workflow actually has, for the chat's participant
  // picker (WorkflowChatPanel) — derived from the Kanban task list rather
  // than a second fetch of the global expert catalogue, and de-duplicated
  // because the same expert can own more than one task across waves.
  const availableExperts = Array.from(
    new Map(tasks.map((t) => [t.assignedExpertId, { id: t.assignedExpertId, name: t.expertName }])).values()
  )

  // Feature #28: workspace tab strip + its counts.
  // events is hoisted so the counts, the empty-state guard and the panels all
  // read the same list (no three copies of `blackboard?.events ?? []`).
  const [activeTab, setActiveTab] = useState<WorkspaceTab>('board')
  const events = blackboard?.events ?? []

  // Verification verdicts, keyed by the artifact they belong to, so the
  // Deliverables screen can show the proof next to the deliverable. A missing
  // entry is shown as "not checked" — the panel never invents a pass.
  const artifactVerifications = useMemo(() => {
    const map = new Map<string, ArtifactVerification>()
    for (const e of events) {
      if (e.eventType !== 'artifact_verification') continue
      const parsed = parseArtifactVerification(e)
      if (parsed) map.set(parsed.artifactEventId, parsed)
    }
    return map
  }, [events])
  // Mirrors ArtifactsPanel's filter so the tab count matches what it renders.
  const artifactCount = events.filter(
    (e) => ARTIFACT_TYPES.has(e.eventType) && (!filterExpertId || e.postedByExpertId === filterExpertId)
  ).length
  const tabs: TabItem<WorkspaceTab>[] = [
    // alert: an approval is waiting — surfaced even when another tab is open.
    { key: 'board', label: 'Board', count: tasks.length, alert: workflow?.status === 'paused_for_approval' },
    { key: 'deliverables', label: 'Deliverables', count: artifactCount },
    { key: 'activity', label: 'Activity', count: events.length },
    { key: 'files', label: 'Files', count: fileStream.files.length },
  ]

  // The Codebase tab exists only for existing-codebase workflows: a scratch
  // workflow has no repository, so an approval queue would be meaningless.
  if (workflow?.mode === 'existing_codebase') {
    tabs.push({ key: 'codebase', label: 'Codebase' })
  }

  // Root scroll container. AppShell's <main> (frontend/src/components/
  // layout/AppShell.tsx) is `overflow-hidden` by contract — every page it
  // renders must provide its OWN scroll box, exactly as ChatPage.tsx already
  // does with its `overflow-y-auto` message list. This page never had one:
  // the Kanban board + ApprovalGate + Deliverables + Files + (now) the
  // Chat/Amendments/Delivery panels made the page taller than the viewport,
  // and main's overflow-hidden silently clipped everything past the fold —
  // Deliverables onward in one screenshot, the chat's textarea/send button in
  // another. The content was always in the DOM; it was just clipped.
  // h-full fills main's height, which is definite (inherited via flexbox from
  // the h-screen root) — the same mechanism ChatPage's h-full already relies
  // on. overflow-y-auto is what actually lets this page scroll instead of
  // clip. WorkflowChatPanel/AmendmentsPanel/DeliveryPanel each already manage
  // their own bounded internal scroll areas, so this nests without conflict.
  return (
    <div className="h-full overflow-y-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-semibold text-text-primary">
            {workflow?.title ?? 'Workflow'}
          </h1>
          <div className="flex items-center gap-3">
            {/* Cancel Workflow Button - only show when workflow is running */}
            {workflow && (workflow.status === 'running' || workflow.status === 'paused_for_approval') && (
              <CancelWorkflowButton workflowId={id!} workflowTitle={workflow.title} />
            )}
            {/* Run state indicator. "Failed" is its own state: the stream closes
                for a failure exactly as it does for success, so a two-way test
                (done or not) labelled every failed workflow "Complete". */}
            <div className="flex items-center gap-1.5">
              <span className={cn(
                'h-2 w-2 rounded-full',
                stream.isFailed || workflow?.status === 'failed' ? 'bg-mode-refuse'
                : stream.isDone ? 'bg-mode-advise'
                : stream.isConnected ? 'bg-mode-advise animate-pulse'
                : 'bg-glow-amber animate-pulse'
              )} />
              <span className="text-[10px] font-medium uppercase tracking-wider text-text-disabled">
                {stream.isFailed || workflow?.status === 'failed'
                  ? 'Failed'
                  : stream.isDone
                    ? 'Complete'
                    : stream.isConnected
                      ? 'Live'
                      : 'Connecting...'}
              </span>
            </div>
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

      {/* Workspace tabs (feature #28). Board is the default so the task board
          and the approval gate stay front-and-centre; Deliverables/Activity/
          Files are read-only views that used to sit below the fold. */}
      <Tabs tabs={tabs} active={activeTab} onChange={setActiveTab} />

      <div className="mt-4">
        {activeTab === 'board' && (
          <>
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
                          <TaskCard
                            key={task.id}
                            task={task}
                            onSelect={() => setFilterExpertId(task.assignedExpertId)}
                            workflowId={id!}
                          />
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
                currentGenericPct={workflow?.genericAllowancePct ?? 0}
              />
            )}
          </>
        )}

        {activeTab === 'deliverables' &&
          (events.length === 0 ? (
            <Card className="p-6">
              <p className="text-center text-xs text-text-disabled">
                No deliverables yet. They appear here as experts post their work.
              </p>
            </Card>
          ) : (
            <ArtifactsPanel
              events={events}
              expertNames={expertNames}
              filterExpertId={filterExpertId}
              onClearFilter={() => setFilterExpertId(null)}
              verifications={artifactVerifications}
            />
          ))}

        {activeTab === 'activity' &&
          (events.length === 0 ? (
            <Card className="p-6">
              <p className="text-center text-xs text-text-disabled">
                No activity yet. Every blackboard event will be listed here.
              </p>
            </Card>
          ) : (
            <ActivityLog events={events} expertNames={expertNames} />
          ))}

        {activeTab === 'files' && (
          <FilesPanel
            files={fileStream.files}
            isConnected={fileStream.isConnected}
            isDone={fileStream.isDone}
            lastWave={fileStream.lastWave}
          />
        )}

        {activeTab === 'codebase' && id && (
          <CodebasePanel workflowId={id} workflowStatus={workflow?.status} />
        )}
      </div>

      {/* Deliverable chat (§6) — separate from the product chat, own tables,
          own tool loop. Available once the workflow has produced at least
          one task/expert to talk to. */}
      {id && availableExperts.length > 0 && (
        <WorkflowChatPanel workflowId={id} availableExperts={availableExperts} />
      )}

      {/* Amendments (§7.5) — proposals from the chat's mutating tools and from
          code-feedback findings, waiting on this client's approve/edit/reject. */}
      {id && <AmendmentsPanel workflowId={id} />}

      {/* Delivery (§17, §18) — push the harness out, or check what got built. */}
      {id && <DeliveryPanel workflowId={id} projectId={workflow?.projectId} />}

      {/* A failed workflow has to state its cause. failure_reason has been
          written by the engine on every failure; the read API simply never
          returned it, so this screen could do no better than a red FAILED
          badge. Shown only for a failed status, so the banner can never be
          mistaken for a live warning. */}
      {workflow?.status === 'failed' && (
        <div className="mt-4 rounded-lg border border-mode-refuse/30 bg-mode-refuse/10 p-4">
          <p className="text-sm font-medium text-mode-refuse">Workflow failed</p>
          <p className="mt-1 text-xs text-text-secondary">
            {workflow.failureReason?.trim() ||
              'No reason was recorded for this failure. Check the Activity tab for the last events.'}
          </p>
        </div>
      )}

      {/* Completion notice + Download Design Package.
          Gated on a stream that did not fail AND a status that is not failed:
          the stream closes for a failure too, and showing the green bar over a
          failed run is how "it says complete but nothing was produced" started. */}
      {!stream.isFailed && workflow?.status !== 'failed' && (stream.isDone || workflow?.status === 'completed') && (
        <div className="mt-4 rounded-lg border border-mode-advise/30 bg-mode-advise/10 p-4">
          <div className="flex items-center justify-between gap-4">
            <p className="text-sm font-medium text-mode-advise">{'\u2705'} Workflow Complete</p>
            <DownloadDesignPackageButton
              workflowTitle={workflow?.title ?? 'workflow'}
              events={events}
              expertNames={expertNames}
            />
          </div>
        </div>
      )}
    </div>
  )
}

export const Component = KanbanPage
