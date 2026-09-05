import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useRevalidator } from 'react-router-dom'
import { connectRepo, syncRepo } from '@/api/repo'
import { Modal } from '@/components/ui/Modal'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 10's "Connect Repository"
 * button, but the ACTUAL flow (Personal Access Token, not OAuth) is
 * dictated by the real backend gap documented in api/repo.ts - there
 * is no reachable OAuth route, so this cannot be the OAuth-redirect UI
 * the wireframe's button label implies.
 *
 * WHY the PAT input has type="password": a GitHub/GitLab PAT is a
 * secret credential, same class of sensitivity as a login password -
 * masking it on-screen (and NOT logging it, NOT storing it in any
 * client-side state that outlives this modal) follows the same
 * handling discipline as the login form's password field.
 *
 * Cross-questioned: after connectRepo succeeds, should sync start
 * automatically? Yes - explicitly chained via syncRepo(projectId) in
 * onSuccess, because ConnectRepo's own handler never triggers a sync
 * itself (confirmed from source: it only INSERTs the connection row
 * and sets sync_status='pending', with no goroutine kicked off). If
 * this component only called connectRepo, the connection would sit at
 * "pending" forever with no follow-up call ever made - it needs a
 * distinct fetch to SyncRepo to actually start syncing.
 *
 * PHASE 6 BUG FIX (same class of bug fixed in ProjectExpertManager.tsx
 * this same commit): previously called
 * queryClient.invalidateQueries({queryKey: ['projects', projectId]}),
 * which is a no-op since ProjectPage sources `project` from a router
 * loader, not a useQuery. Replaced with revalidator.revalidate().
 * The repo/status invalidation (a real useQuery, from
 * useRepoSyncStatus.ts) is correct and kept as-is.
 */
export interface RepoConnectModalProps {
  isOpen: boolean
  onClose: () => void
  projectId: string
}

export function RepoConnectModal({ isOpen, onClose, projectId }: RepoConnectModalProps) {
  const queryClient = useQueryClient()
  const revalidator = useRevalidator()
  const [provider, setProvider] = useState<'github' | 'gitlab'>('github')
  const [repoUrl, setRepoUrl] = useState('')
  const [accessToken, setAccessToken] = useState('')
  const [defaultBranch, setDefaultBranch] = useState('main')

  const mutation = useMutation({
    mutationFn: async () => {
      await connectRepo(projectId, {
        provider,
        repoUrl: repoUrl.trim(),
        accessToken,
        defaultBranch: defaultBranch.trim() || 'main',
      })
      // Chained deliberately - see header comment. ConnectRepo alone
      // never starts a sync.
      await syncRepo(projectId)
    },
    onSuccess: () => {
      revalidator.revalidate()
      queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'repo', 'status'] })
      setAccessToken('') // WHY clear this specifically on success (not repoUrl/branch): don't leave a secret token sitting in component state any longer than necessary once it's served its purpose.
      onClose()
    },
  })

  const canSubmit = repoUrl.trim().length > 0 && accessToken.length > 0

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <h3 className="mb-1 text-sm font-medium text-text-primary">Connect Repository</h3>
      <p className="mb-4 text-xs text-text-secondary">
        Requires a Personal Access Token with repo read access. OAuth connect is not available yet -
        see this project's HANDOFF.md for details.
      </p>

      <div className="flex flex-col gap-3">
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => setProvider('github')}
            className={provider === 'github' ? 'font-medium text-brand' : 'text-text-secondary'}
          >
            GitHub
          </button>
          <button
            type="button"
            onClick={() => setProvider('gitlab')}
            className={provider === 'gitlab' ? 'font-medium text-brand' : 'text-text-secondary'}
          >
            GitLab
          </button>
        </div>

        <Input
          label="Repository URL"
          placeholder="https://github.com/owner/repo"
          value={repoUrl}
          onChange={(e) => setRepoUrl(e.target.value)}
        />
        <Input
          label="Personal Access Token"
          type="password"
          value={accessToken}
          onChange={(e) => setAccessToken(e.target.value)}
        />
        <Input
          label="Default branch"
          value={defaultBranch}
          onChange={(e) => setDefaultBranch(e.target.value)}
        />

        {mutation.isError && (
          <p className="text-sm text-mode-refuse">
            {mutation.error instanceof Error ? mutation.error.message : 'Connection failed'}
          </p>
        )}
      </div>

      <div className="mt-4 flex justify-end gap-2">
        <Button variant="secondary" onClick={onClose} disabled={mutation.isPending}>
          Cancel
        </Button>
        <Button disabled={!canSubmit} isLoading={mutation.isPending} onClick={() => mutation.mutate()}>
          Connect & Sync
        </Button>
      </div>
    </Modal>
  )
}
