import { useMemo, useState, type FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createManagedAccount,
  getAdminExperts,
  getManagedAccounts,
  issueBootstrapToken,
  setAccountExperts,
  updateManagedAccount,
  type ManagedAccount,
} from '@/api/admin'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { Modal } from '@/components/ui/Modal'
import { handleAPIError } from '@/utils/errors'
import { formatRelativeTime } from '@/utils/format'

function AdminAccounts() {
  const queryClient = useQueryClient()
  const { data: accounts, isLoading } = useQuery({
    queryKey: ['admin', 'accounts'],
    queryFn: getManagedAccounts,
  })
  const { data: experts } = useQuery({
    queryKey: ['admin', 'experts'],
    queryFn: getAdminExperts,
  })

  const [createOpen, setCreateOpen] = useState(false)
  const [assignAccount, setAssignAccount] = useState<ManagedAccount | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [bootstrapToken, setBootstrapToken] = useState<string | null>(null)

  const [fullName, setFullName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<'admin' | 'domain_expert'>('domain_expert')
  const [selectedExperts, setSelectedExperts] = useState<string[]>([])

  const createMutation = useMutation({
    mutationFn: createManagedAccount,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'accounts'] })
      setCreateOpen(false)
      setFullName('')
      setEmail('')
      setPassword('')
      setRole('domain_expert')
      setSelectedExperts([])
      setError(null)
    },
    onError: (err) => setError(handleAPIError(err)),
  })

  const toggleMutation = useMutation({
    mutationFn: ({ id, isActive }: { id: string; isActive: boolean }) =>
      updateManagedAccount(id, isActive),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin', 'accounts'] }),
  })

  const grantMutation = useMutation({
    mutationFn: ({ id, expertIds }: { id: string; expertIds: string[] }) =>
      setAccountExperts(id, expertIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'accounts'] })
      setAssignAccount(null)
    },
    onError: (err) => setError(handleAPIError(err)),
  })

  const bootstrapMutation = useMutation({
    mutationFn: issueBootstrapToken,
    onSuccess: (data) => setBootstrapToken(data.token),
    onError: (err) => setError(handleAPIError(err)),
  })

  const expertName = useMemo(() => {
    const map = new Map<string, string>()
    experts?.forEach((e) => map.set(e.id, e.name))
    return map
  }, [experts])

  function handleCreate(e: FormEvent) {
    e.preventDefault()
    setError(null)
    createMutation.mutate({
      fullName: fullName.trim(),
      email: email.trim(),
      password,
      role,
      expertIds: role === 'domain_expert' ? selectedExperts : [],
    })
  }

  function toggleExpert(id: string) {
    setSelectedExperts((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id],
    )
  }

  if (isLoading) {
    return (
      <div className="space-y-3 p-6">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-16" />
        ))}
      </div>
    )
  }

  return (
    <div className="p-6">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
        <div>
          <h1 className="text-xl font-semibold">Accounts</h1>
          <p className="text-xs text-text-secondary">
            Create admins and domain-expert accounts; assign AI experts per account.
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="secondary"
            size="sm"
            isLoading={bootstrapMutation.isPending}
            onClick={() => bootstrapMutation.mutate()}
          >
            Issue bootstrap token
          </Button>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            Create account
          </Button>
        </div>
      </div>

      {bootstrapToken && (
        <Card className="mb-4 border border-mode-advise/30 bg-mode-advise/5 p-3">
          <p className="mb-1 text-xs font-medium text-mode-advise">One-time bootstrap token</p>
          <code className="block break-all text-xs text-text-primary">{bootstrapToken}</code>
          <p className="mt-1 text-xs text-text-secondary">
            Share out-of-band. Open /bootstrap/&lt;token&gt; to enroll the first admin with TOTP.
          </p>
        </Card>
      )}

      {error && (
        <p className="mb-3 rounded-md border border-mode-refuse/30 bg-mode-refuse/10 px-3 py-2 text-xs text-mode-refuse">
          {error}
        </p>
      )}

      <div className="space-y-2">
        {accounts?.map((account) => (
          <Card key={account.id} glow="cyan" className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <p className="font-medium text-text-primary">{account.fullName}</p>
              <p className="text-xs text-text-secondary">{account.email}</p>
              <p className="text-xs text-text-disabled">
                {account.role}
                {' · '}
                {account.projectCount} projects
                {' · '}
                {account.messageCount} messages
                {account.lastLogin && (
                  <>
                    {' · Last login: '}
                    {formatRelativeTime(account.lastLogin)}
                  </>
                )}
              </p>
              {account.role === 'domain_expert' && (
                <p className="mt-1 text-xs text-text-secondary">
                  Experts:{' '}
                  {account.expertIds?.length
                    ? account.expertIds.map((id) => expertName.get(id) ?? id.slice(0, 8)).join(', ')
                    : 'none assigned'}
                </p>
              )}
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Badge variant={account.isActive ? 'brand' : 'neutral'}>
                {account.isActive ? 'Active' : 'Disabled'}
              </Badge>
              {account.totpEnabled && <Badge variant="brand">TOTP</Badge>}
              {account.role === 'domain_expert' && (
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => {
                    setAssignAccount(account)
                    setSelectedExperts(account.expertIds ?? [])
                    setError(null)
                  }}
                >
                  Assign experts
                </Button>
              )}
              <Button
                variant="secondary"
                size="sm"
                isLoading={toggleMutation.isPending && toggleMutation.variables?.id === account.id}
                onClick={() =>
                  toggleMutation.mutate({ id: account.id, isActive: !account.isActive })
                }
              >
                {account.isActive ? 'Disable' : 'Enable'}
              </Button>
            </div>
          </Card>
        ))}
        {!accounts?.length && (
          <p className="text-sm text-text-secondary">No managed accounts yet.</p>
        )}
      </div>

      <Modal isOpen={createOpen} onClose={() => setCreateOpen(false)}>
        <form onSubmit={handleCreate} className="flex w-full max-w-md flex-col gap-3 p-4">
          <h2 className="text-lg font-semibold">Create account</h2>
          <Input
            label="Full name"
            value={fullName}
            onChange={(e) => setFullName(e.target.value)}
            required
          />
          <Input
            label="Email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          <Input
            label="Temporary password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          <label className="text-xs text-text-secondary">
            Role
            <select
              className="mt-1 w-full rounded-md border border-glass-border bg-surface-base px-2 py-2 text-sm text-text-primary"
              value={role}
              onChange={(e) => setRole(e.target.value as 'admin' | 'domain_expert')}
            >
              <option value="domain_expert">Domain expert</option>
              <option value="admin">Admin</option>
            </select>
          </label>
          {role === 'domain_expert' && (
            <div className="max-h-40 overflow-y-auto rounded-md border border-glass-border p-2">
              <p className="mb-1 text-xs text-text-secondary">Assign domain experts</p>
              {experts?.map((ex) => (
                <label key={ex.id} className="flex items-center gap-2 py-1 text-sm text-text-primary">
                  <input
                    type="checkbox"
                    checked={selectedExperts.includes(ex.id)}
                    onChange={() => toggleExpert(ex.id)}
                  />
                  {ex.name} <span className="text-xs text-text-disabled">({ex.domain})</span>
                </label>
              ))}
            </div>
          )}
          {error && (
            <p className="text-xs text-mode-refuse">{error}</p>
          )}
          <Button type="submit" isLoading={createMutation.isPending}>
            Create
          </Button>
        </form>
      </Modal>

      <Modal isOpen={!!assignAccount} onClose={() => setAssignAccount(null)}>
        <div className="flex w-full max-w-md flex-col gap-3 p-4">
          <h2 className="text-lg font-semibold">
            Assign experts — {assignAccount?.fullName}
          </h2>
          <div className="max-h-56 overflow-y-auto rounded-md border border-glass-border p-2">
            {experts?.map((ex) => (
              <label key={ex.id} className="flex items-center gap-2 py-1 text-sm text-text-primary">
                <input
                  type="checkbox"
                  checked={selectedExperts.includes(ex.id)}
                  onChange={() => toggleExpert(ex.id)}
                />
                {ex.name} <span className="text-xs text-text-disabled">({ex.domain})</span>
              </label>
            ))}
          </div>
          {error && <p className="text-xs text-mode-refuse">{error}</p>}
          <Button
            isLoading={grantMutation.isPending}
            onClick={() =>
              assignAccount &&
              grantMutation.mutate({ id: assignAccount.id, expertIds: selectedExperts })
            }
          >
            Save assignments
          </Button>
        </div>
      </Modal>
    </div>
  )
}

export const Component = AdminAccounts
