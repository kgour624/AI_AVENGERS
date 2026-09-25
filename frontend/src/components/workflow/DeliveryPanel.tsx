import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { exportHarnessToGit, ingestCodeFeedback, type GitExportResult } from '@/api/delivery'
import { useRepoSyncStatus } from '@/hooks/useRepoSyncStatus'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { handleAPIError } from '@/utils/errors'

/**
 * DeliveryPanel — §17 (code-feedback loop) and §18 (git-push export),
 * docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md.
 *
 * Both call POST endpoints that act on a CLIENT's own git remote using the
 * token from the project's repo connection (RepoConnectModal.tsx / api/repo.ts)
 * — this panel does not collect a token itself, it only names a target repo;
 * the backend reads the connected token and refuses (client_remote.go's
 * allowlist) before ever touching it if the host is not github.com/gitlab.com.
 *
 * Export is synchronous — the request resolves with the commit SHA. Ingest is
 * asynchronous — the response only confirms the background job started; the
 * actual findings arrive as blackboard events (code_feedback_ingested,
 * code_feedback_question), read through the existing Deliverables panel
 * (ArtifactsPanel in KanbanPage.tsx already renders any event type it does
 * not specifically know about via its content dump — see renderContentValue).
 */

type Provider = 'github' | 'gitlab'

function ExportForm({ workflowId }: { workflowId: string }) {
  const [provider, setProvider] = useState<Provider>('github')
  const [repoUrl, setRepoUrl] = useState('')
  const [branch, setBranch] = useState('main')
  const [createRepo, setCreateRepo] = useState(false)
  const [result, setResult] = useState<GitExportResult | null>(null)

  const mutation = useMutation({
    mutationFn: () =>
      exportHarnessToGit(workflowId, {
        provider,
        repoUrl: repoUrl.trim(),
        branch: branch.trim() || undefined,
        createRepo,
      }),
    onSuccess: (res) => setResult(res),
  })

  return (
    <div>
      <p className="text-xs font-medium text-text-primary">Push the harness to a repo</p>
      <p className="mt-0.5 text-[11px] text-text-disabled">
        Pushes final.md, ACCEPTANCE.md, DECISIONS.md and every design section as
        real commits. Never overwrites existing history — an occupied branch is
        refused, not force-pushed.
      </p>

      <div className="mt-2 flex gap-2">
        <button
          type="button"
          onClick={() => setProvider('github')}
          className={provider === 'github' ? 'text-xs font-medium text-brand' : 'text-xs text-text-secondary'}
        >
          GitHub
        </button>
        <button
          type="button"
          onClick={() => setProvider('gitlab')}
          className={provider === 'gitlab' ? 'text-xs font-medium text-brand' : 'text-xs text-text-secondary'}
        >
          GitLab
        </button>
      </div>

      <div className="mt-2 flex gap-2">
        <Input
          placeholder="owner/repo or https://github.com/owner/repo"
          value={repoUrl}
          onChange={(e) => setRepoUrl(e.target.value)}
          className="flex-1"
        />
        <Input
          placeholder="branch"
          value={branch}
          onChange={(e) => setBranch(e.target.value)}
          className="w-28"
        />
      </div>

      <label className="mt-2 flex items-center gap-1.5 text-[11px] text-text-secondary">
        <input type="checkbox" checked={createRepo} onChange={(e) => setCreateRepo(e.target.checked)} />
        Create the repository if it does not exist (private by default)
      </label>

      {mutation.isError && (
        <p className="mt-2 text-xs text-mode-refuse">{handleAPIError(mutation.error)}</p>
      )}
      {result && (
        <p className="mt-2 text-xs text-mode-advise">
          Pushed {result.commitSha.slice(0, 8)} to {result.repoUrl}#{result.branch}
          {result.repoCreated && ' (repo created)'}
        </p>
      )}

      <Button
        className="mt-2"
        size="sm"
        disabled={!repoUrl.trim() || mutation.isPending}
        isLoading={mutation.isPending}
        onClick={() => mutation.mutate()}
      >
        Push to repo
      </Button>
    </div>
  )
}

function CodeFeedbackForm({ workflowId }: { workflowId: string }) {
  const [provider, setProvider] = useState<Provider>('github')
  const [repoUrl, setRepoUrl] = useState('')
  const [branch, setBranch] = useState('')
  const [started, setStarted] = useState(false)

  const mutation = useMutation({
    mutationFn: () =>
      ingestCodeFeedback(workflowId, {
        provider,
        repoUrl: repoUrl.trim(),
        branch: branch.trim() || undefined,
      }),
    onSuccess: () => setStarted(true),
  })

  return (
    <div>
      <p className="text-xs font-medium text-text-primary">Check what got built</p>
      <p className="mt-0.5 text-[11px] text-text-disabled">
        Clones the repo Claude Code built, runs each acceptance criterion's
        Verify command against it, and compares its endpoints/tables to the
        design. Findings are routed to the expert who owns that section — watch
        the Amendments panel below.
      </p>

      <div className="mt-2 flex gap-2">
        <button
          type="button"
          onClick={() => setProvider('github')}
          className={provider === 'github' ? 'text-xs font-medium text-brand' : 'text-xs text-text-secondary'}
        >
          GitHub
        </button>
        <button
          type="button"
          onClick={() => setProvider('gitlab')}
          className={provider === 'gitlab' ? 'text-xs font-medium text-brand' : 'text-xs text-text-secondary'}
        >
          GitLab
        </button>
      </div>

      <div className="mt-2 flex gap-2">
        <Input
          placeholder="the client's built repo"
          value={repoUrl}
          onChange={(e) => setRepoUrl(e.target.value)}
          className="flex-1"
        />
        <Input
          placeholder="branch (default)"
          value={branch}
          onChange={(e) => setBranch(e.target.value)}
          className="w-32"
        />
      </div>

      {mutation.isError && (
        <p className="mt-2 text-xs text-mode-refuse">{handleAPIError(mutation.error)}</p>
      )}
      {started && !mutation.isError && (
        <p className="mt-2 text-xs text-mode-advise">
          Started — this runs in the background. Check the Deliverables and
          Amendments panels for results.
        </p>
      )}

      <Button
        className="mt-2"
        size="sm"
        variant="secondary"
        disabled={!repoUrl.trim() || mutation.isPending}
        isLoading={mutation.isPending}
        onClick={() => {
          setStarted(false)
          mutation.mutate()
        }}
      >
        Check the build
      </Button>
    </div>
  )
}

export function DeliveryPanel({
  workflowId,
  projectId,
}: {
  workflowId: string
  // projectId lets the panel say "connect a repo first" instead of letting the
  // push fail with "no usable git provider credential for this project". The
  // connect flow lives on the project page, so the fix is a link away.
  projectId?: string
}) {
  const { data: repoStatus } = useRepoSyncStatus(projectId ?? '', !!projectId)
  const needsRepo = !!projectId && !!repoStatus && !repoStatus.connected

  return (
    <div className="mt-6">
      <div className="mb-2 flex items-center justify-between">
        <span className="text-xs font-semibold uppercase tracking-wide text-text-secondary">
          Delivery
        </span>
      </div>

      {needsRepo && (
        <div className="mb-3 rounded-lg border border-glass-border bg-surface-raised p-3">
          <p className="text-[11px] text-text-secondary">
            This project has no connected repository, so there is no credential to push with.{' '}
            <Link to={`/projects/${projectId}`} className="text-brand underline">
              Connect GitHub or GitLab on the project page
            </Link>{' '}
            first — the repository URL below must be one that connection can reach. A Personal
            Access Token works even when the one-click OAuth app is not configured.
          </p>
        </div>
      )}
      <div className="grid grid-cols-2 gap-4">
        <Card className="p-4">
          <ExportForm workflowId={workflowId} />
        </Card>
        <Card className="p-4">
          <CodeFeedbackForm workflowId={workflowId} />
        </Card>
      </div>
    </div>
  )
}
