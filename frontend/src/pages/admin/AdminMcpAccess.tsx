import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { createMcpToken, listMcpTokens, revokeMcpToken, type McpTokenRecord } from '@/api/mcp'

/**
 * MCP Access — the tokens external coding agents (Claude Code and friends) use to
 * reach the domain experts.
 *
 * WHY this screen exists: a token can be minted from a CLI, which is fine for an
 * engineer and impossible for the person running the product. Everything an
 * operator needs is here: create, copy, see what a token may reach, and revoke.
 *
 * The token is shown EXACTLY ONCE, in the panel below the form. After that the
 * server only holds its hash, so the UI never pretends it can show it again.
 */
function AdminMcpAccess() {
  const queryClient = useQueryClient()
  const [label, setLabel] = useState('')
  const [domains, setDomains] = useState('')
  const [tools, setTools] = useState('')
  const [newToken, setNewToken] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  const { data: tokens, isLoading } = useQuery({
    queryKey: ['admin', 'mcp-tokens'],
    queryFn: listMcpTokens,
  })

  const refresh = () => queryClient.invalidateQueries({ queryKey: ['admin', 'mcp-tokens'] })

  const createMutation = useMutation({
    mutationFn: () =>
      createMcpToken({
        label: label.trim(),
        domains: splitList(domains),
        tools: splitList(tools),
      }),
    onSuccess: (data) => {
      setNewToken(data.token)
      setCopied(false)
      setLabel('')
      setDomains('')
      setTools('')
      refresh()
    },
  })

  const revokeMutation = useMutation({
    mutationFn: (id: string) => revokeMcpToken(id),
    onSuccess: refresh,
  })

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-lg font-semibold text-text-primary">MCP Access</h1>
        <p className="mt-1 text-sm text-text-secondary">
          Tokens that let an external coding agent (Claude Code, Cursor) use your domain experts:
          read their standards, ask them questions, and have a change reviewed.
        </p>
      </div>

      <Card>
        <h2 className="text-sm font-semibold text-text-primary">New token</h2>
        <div className="mt-3 grid gap-3 sm:grid-cols-3">
          <Input
            label="Label"
            value={label}
            onChange={(e) => setLabel(e.target.value)}
            placeholder="e.g. Sneha laptop"
          />
          <Input
            label="Domains (optional)"
            value={domains}
            onChange={(e) => setDomains(e.target.value)}
            placeholder="frontend, system design"
          />
          <Input
            label="Tools (optional)"
            value={tools}
            onChange={(e) => setTools(e.target.value)}
            placeholder="get_standards, ask_expert"
          />
        </div>
        <p className="mt-2 text-xs text-text-disabled">
          Leave domains or tools empty to allow all of them. Restricting a token is what keeps one
          developer&apos;s agent out of domains it has no business in.
        </p>
        <div className="mt-3">
          <Button
            disabled={!label.trim() || createMutation.isPending}
            isLoading={createMutation.isPending}
            onClick={() => createMutation.mutate()}
          >
            Create token
          </Button>
        </div>

        {createMutation.isError && (
          <p className="mt-3 text-xs text-mode-refuse">
            Could not create the token. Check the label and try again.
          </p>
        )}

        {newToken && (
          <div className="mt-4 rounded-md border border-brand/40 bg-brand/5 p-3">
            <p className="text-xs font-medium text-text-primary">
              Copy this now — it is shown only once. The server stores only its hash.
            </p>
            <code className="mt-2 block break-all rounded bg-black/40 p-2 text-xs text-text-secondary">
              {newToken}
            </code>
            <div className="mt-2 flex items-center gap-3">
              <Button
                size="sm"
                variant="secondary"
                onClick={async () => {
                  await navigator.clipboard?.writeText(newToken)
                  setCopied(true)
                }}
              >
                {copied ? 'Copied' : 'Copy'}
              </Button>
              <button
                type="button"
                className="text-xs text-text-disabled hover:text-text-primary"
                onClick={() => setNewToken(null)}
              >
                Dismiss
              </button>
            </div>
          </div>
        )}
      </Card>

      <Card>
        <h2 className="text-sm font-semibold text-text-primary">Existing tokens</h2>
        {isLoading && <p className="mt-3 text-xs text-text-disabled">Loading tokens...</p>}
        {!isLoading && (tokens?.length ?? 0) === 0 && (
          <p className="mt-3 text-xs text-text-disabled">No tokens yet.</p>
        )}
        <div className="mt-3 space-y-2">
          {tokens?.map((token: McpTokenRecord) => (
            <div
              key={token.id}
              className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-border-subtle bg-white/5 px-3 py-2"
            >
              <div className="min-w-0 flex-1">
                <p className="truncate text-xs font-medium text-text-primary">
                  {token.label}
                  {token.revoked && <span className="ml-2 text-mode-refuse">revoked</span>}
                </p>
                <p className="mt-0.5 text-[11px] text-text-disabled">
                  domains: {token.domains?.length ? token.domains.join(', ') : 'all'}
                  {' · '}tools: {token.tools?.length ? token.tools.join(', ') : 'all'}
                  {' · '}requests: {token.request_count}
                  {token.last_used_at ? ` · last used ${token.last_used_at.slice(0, 10)}` : ' · never used'}
                </p>
              </div>
              {!token.revoked && (
                <button
                  type="button"
                  disabled={revokeMutation.isPending}
                  onClick={() => revokeMutation.mutate(token.id)}
                  className="flex-shrink-0 rounded-md border border-border-subtle px-2.5 py-1 text-xs text-text-secondary transition hover:border-mode-refuse/40 hover:text-mode-refuse disabled:opacity-40"
                >
                  Revoke
                </button>
              )}
            </div>
          ))}
        </div>
        {revokeMutation.isError && (
          <p className="mt-2 text-xs text-mode-refuse">
            Revoke failed — the token may already be revoked.
          </p>
        )}
      </Card>
    </div>
  )
}

// The router lazy-loads a named `Component` (same contract every admin page
// follows), so the page is exported that way rather than as a default.
export const Component = AdminMcpAccess

/** "a, b ,c" -> ["a","b","c"]; empty means "no restriction". */
function splitList(raw: string): string[] {
  return raw
    .split(',')
    .map((part) => part.trim())
    .filter(Boolean)
}
