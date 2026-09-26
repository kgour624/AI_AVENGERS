import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getLLMSettings,
  updateLLMSettings,
  getCodeCraftModels,
  getEmbeddingSettings,
  updateEmbeddingSettings,
  getModelLimits,
  updateModelLimits,
  getLLMHealth,
} from '@/api/admin'
import type { ModelLimit } from '@/api/admin'
import type { CodeCraftModel } from '@/types/codecraftapi'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card } from '@/components/ui/Card'
import { handleAPIError } from '@/utils/errors'

/**
 * LLM + Embedding Settings page.
 *
 * Changes from original:
 * 1. Added CodeCraftAPI as 5th provider option.
 * 2. CodeCraftAPI model picker section (3 dropdowns: cheap/strong/fast).
 *    Shown only when codecraftapi is the active provider.
 *    "Fetch Available Models" is a MANUAL button — NOT called on every render.
 *    WHY: avoids hitting the backend proxy on every page load; admin
 *    explicitly requests the model list when they need it.
 * 3. Embedding settings section (always visible, separate from LLM provider).
 *    Toggle: sidecar (default) vs codecraftapi.
 *    Embedding model picker (dropdown when models fetched, text input fallback).
 *    Re-ingestion warning banner (visible when codecraftapi selected).
 * 4. Separate Save buttons for LLM settings and embedding settings.
 *    WHY separate: they write to different system_settings keys and have
 *    different side effects (LLM change = next call; embedding change = must re-ingest).
 * 5. Added Cavoti as 6th provider option (backend already wired: see
 *    internal/gateway/providers/cavoti.go + model_gateway.go + config.go).
 *    Cavoti model picker uses manual text Inputs, NOT a "Fetch Available
 *    Models" button like CodeCraftAPI — the backend has no
 *    GET /admin/cavoti/models proxy endpoint (only codecraftapi/models
 *    exists), so there is nothing to fetch from yet. Same fallback
 *    pattern already used below for the embedding model Input when no
 *    catalog has been fetched.
 */

const PROVIDERS = [
  { value: 'openrouter',   label: 'OpenRouter (multi-model gateway)',   desc: 'Single key, access to Claude, GPT, Gemini, DeepSeek and more' },
  { value: 'deepseek',     label: 'DeepSeek (direct)',                  desc: 'Direct DeepSeek API — use your own DeepSeek key' },
  { value: 'anthropic',    label: 'Anthropic (direct)',                 desc: 'Direct Anthropic API — use your own Claude key' },
  { value: 'gemini',       label: 'Google Gemini (direct)',             desc: 'Direct Gemini API — use your own Google AI key' },
  { value: 'codecraftapi', label: 'CodeCraftAPI (multi-model gateway)', desc: 'Single key, access to multiple AI models. Also supports embeddings.' },
  { value: 'cavoti',       label: 'Cavoti (multi-model gateway)',       desc: 'OpenAI-compatible, sk- prefix API key. Multiple models via one key.' },
]

function AdminLLMSettings() {
  const queryClient = useQueryClient()

  // LLM settings from backend
  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'llm-settings'],
    queryFn: getLLMSettings,
  })

  // Embedding settings from backend
  const { data: embData } = useQuery({
    queryKey: ['admin', 'embedding-settings'],
    queryFn: getEmbeddingSettings,
  })

  // LLM provider + key state
  const [provider, setProvider] = useState('')
  const [keys, setKeys] = useState<Record<string, string>>({})
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  // CodeCraftAPI model selection state
  const [ccModels, setCcModels] = useState<CodeCraftModel[]>([])
  const [ccModelCheap, setCcModelCheap] = useState('')
  const [ccModelStrong, setCcModelStrong] = useState('')
  const [ccModelFast, setCcModelFast] = useState('')
  const [fetchingModels, setFetchingModels] = useState(false)
  const [fetchModelsError, setFetchModelsError] = useState('')

  // Cavoti model selection state — manual text inputs, no catalog fetch.
  // WHY no "Fetch Available Models" button here: backend has no
  // GET /admin/cavoti/models proxy (only /admin/codecraftapi/models
  // exists). Admin types the model ID from the Cavoti dashboard.
  const [cavotiModelCheap, setCavotiModelCheap] = useState('')
  const [cavotiModelStrong, setCavotiModelStrong] = useState('')
  const [cavotiModelFast, setCavotiModelFast] = useState('')

  // Provider health (G4). Refetched on a timer: a breaker that opened a minute
  // ago is exactly the thing an admin opens this page to see, and a stale
  // snapshot would say "working" while every call is being skipped.
  const { data: health } = useQuery({
    queryKey: ['admin', 'llm-health'],
    queryFn: getLLMHealth,
    refetchInterval: 15000,
  })

  // Model token limits state. Drafts are keyed by provider|tier and only hold
  // the rows the admin has actually touched, so an untouched table saves nothing.
  const { data: limitsData } = useQuery({
    queryKey: ['admin', 'model-limits'],
    queryFn: getModelLimits,
  })
  const [limitDrafts, setLimitDrafts] = useState<Record<string, ModelLimit>>({})
  const [newLimitProvider, setNewLimitProvider] = useState('anthropic')
  const [newLimitTier, setNewLimitTier] = useState('strong')
  const [limitsError, setLimitsError] = useState('')
  const [limitsSuccess, setLimitsSuccess] = useState('')

  const limitKey = (l: { provider: string; tier: string }) => `${l.provider}|${l.tier}`

  // Server rows plus anything drafted on top, so a newly added row is visible
  // before it is saved.
  const limitRows: ModelLimit[] = []
  const seenLimitKeys = new Set<string>()
  for (const row of limitsData?.limits ?? []) {
    limitRows.push(limitDrafts[limitKey(row)] ?? row)
    seenLimitKeys.add(limitKey(row))
  }
  for (const draft of Object.values(limitDrafts)) {
    if (!seenLimitKeys.has(limitKey(draft))) limitRows.push(draft)
  }

  const editLimit = (row: ModelLimit, patch: Partial<ModelLimit>) =>
    setLimitDrafts((prev) => ({ ...prev, [limitKey(row)]: { ...row, ...patch } }))

  const limitsMutation = useMutation({
    mutationFn: () => updateModelLimits(Object.values(limitDrafts)),
    onSuccess: () => {
      setLimitsError('')
      setLimitsSuccess('Token limits saved. The next LLM call uses them — no restart needed.')
      setLimitDrafts({})
      queryClient.invalidateQueries({ queryKey: ['admin', 'model-limits'] })
    },
    onError: (err: unknown) => {
      setLimitsSuccess('')
      setLimitsError(handleAPIError(err))
    },
  })

  // Embedding settings state
  const [embeddingProvider, setEmbeddingProvider] = useState<'sidecar' | 'codecraftapi' | ''>('')
  const [embeddingModel, setEmbeddingModel] = useState('')
  const [embError, setEmbError] = useState('')
  const [embSuccess, setEmbSuccess] = useState('')

  const activeProvider = provider || data?.activeProvider || 'openrouter'
  const activeEmbeddingProvider = embeddingProvider || embData?.embeddingProvider || 'sidecar'

  // LLM settings mutation
  const mutation = useMutation({
    mutationFn: () =>
      updateLLMSettings({
        provider: activeProvider,
        apiKeys: keys,
        codecraftapiModelCheap:  ccModelCheap  || undefined,
        codecraftapiModelStrong: ccModelStrong || undefined,
        codecraftapiModelFast:   ccModelFast   || undefined,
        cavotiModelCheap:  cavotiModelCheap  || undefined,
        cavotiModelStrong: cavotiModelStrong || undefined,
        cavotiModelFast:   cavotiModelFast   || undefined,
      }),
    onSuccess: () => {
      setSuccess('Settings saved. Takes effect on next LLM call.')
      setError('')
      setKeys({})
      queryClient.invalidateQueries({ queryKey: ['admin', 'llm-settings'] })
    },
    onError: (err) => {
      setError(handleAPIError(err))
      setSuccess('')
    },
  })

  // Embedding settings mutation
  const embMutation = useMutation({
    mutationFn: () => {
      const ep = activeEmbeddingProvider as 'sidecar' | 'codecraftapi'
      const em = embeddingModel || embData?.embeddingModel || ''
      if (ep === 'codecraftapi' && !em) {
        throw new Error('Select an embedding model before saving.')
      }
      return updateEmbeddingSettings({
        embeddingProvider: ep,
        embeddingModel: ep === 'codecraftapi' ? em : undefined,
      })
    },
    onSuccess: () => {
      setEmbSuccess('Embedding settings saved. Takes effect on next Embed() call.')
      setEmbError('')
      queryClient.invalidateQueries({ queryKey: ['admin', 'embedding-settings'] })
    },
    onError: (err) => {
      setEmbError(handleAPIError(err))
      setEmbSuccess('')
    },
  })

  // Fetch CodeCraftAPI model list — MANUAL trigger only, not on every render.
  // WHY: avoids hitting the backend proxy on every page load.
  async function handleFetchModels() {
    setFetchingModels(true)
    setFetchModelsError('')
    try {
      const models = await getCodeCraftModels()
      setCcModels(models)
    } catch (err) {
      setFetchModelsError(handleAPIError(err))
    } finally {
      setFetchingModels(false)
    }
  }

  return (
    <div className="p-6 max-w-2xl">
      <h1 className="text-xl font-semibold text-text-primary mb-1">LLM Settings</h1>
      <p className="text-sm text-text-secondary mb-6">
        Change provider or API keys without restarting the server.
      </p>

      {isLoading ? (
        <p className="text-sm text-text-secondary">Loading...</p>
      ) : (
        <div className="flex flex-col gap-6">

          {/* Provider selector */}
          <Card>
            <p className="text-sm font-medium text-text-primary mb-3">Active Provider</p>
            <div className="flex flex-col gap-2">
              {PROVIDERS.map((p) => (
                <label
                  key={p.value}
                  className={`flex items-start gap-3 rounded-lg border p-3 cursor-pointer ${
                    activeProvider === p.value
                      ? 'border-brand bg-brand/5'
                      : 'border-surface-border'
                  }`}
                >
                  <input
                    type="radio"
                    name="provider"
                    value={p.value}
                    checked={activeProvider === p.value}
                    onChange={() => setProvider(p.value)}
                    className="mt-0.5"
                  />
                  <div>
                    <p className="text-sm font-medium text-text-primary">{p.label}</p>
                    <p className="text-xs text-text-secondary">{p.desc}</p>
                  </div>
                </label>
              ))}
            </div>
          </Card>

          {/* API Keys */}
          <Card>
            <p className="text-sm font-medium text-text-primary mb-1">API Keys</p>
            <p className="text-xs text-text-secondary mb-4">
              Leave blank to keep existing key. Keys are masked after saving.
            </p>
            <div className="flex flex-col gap-3">
              {PROVIDERS.map((p) => {
                const existing = data?.apiKeysConfigured?.[p.value]
                return (
                  <Input
                    key={p.value}
                    label={`${p.label.split(' (')[0]} API Key${
                      existing ? ` (current: ${existing})` : ''
                    }`}
                    type="password"
                    placeholder={
                      existing
                        ? 'Enter new key to replace'
                        : p.value === 'codecraftapi'
                        ? 'cc_...'
                        : 'sk-...'
                    }
                    value={keys[p.value] ?? ''}
                    onChange={(e) =>
                      setKeys((prev) => ({ ...prev, [p.value]: e.target.value }))
                    }
                  />
                )
              })}
            </div>
          </Card>

          {/* CodeCraftAPI Model Selection — only shown when codecraftapi is selected */}
          {activeProvider === 'codecraftapi' && (
            <Card>
              <p className="text-sm font-medium text-text-primary mb-1">
                CodeCraftAPI Model Selection
              </p>
              <p className="text-xs text-text-secondary mb-3">
                Fetch the available models, then assign one per tier.
              </p>

              <Button
                onClick={handleFetchModels}
                isLoading={fetchingModels}
                className="mb-3 self-start"
              >
                Fetch Available Models
              </Button>

              {fetchModelsError && (
                <p className="text-xs text-mode-refuse mb-3">{fetchModelsError}</p>
              )}

              {ccModels.length > 0 && (
                <div className="flex flex-col gap-3">
                  {([
                    {
                      label: 'Cheap tier (tagging, metadata, coverage checks)',
                      value: ccModelCheap,
                      setter: setCcModelCheap,
                    },
                    {
                      label: 'Strong tier (answer generation, charter extraction)',
                      value: ccModelStrong,
                      setter: setCcModelStrong,
                    },
                    {
                      label: 'Fast tier (quick checks, simple tasks)',
                      value: ccModelFast,
                      setter: setCcModelFast,
                    },
                  ] as const).map(({ label, value, setter }) => (
                    <div key={label}>
                      <p className="text-xs text-text-secondary mb-1">{label}</p>
                      <select
                        className="w-full rounded-md border border-surface-border bg-surface-secondary text-text-primary text-sm px-3 py-2"
                        value={value}
                        onChange={(e) => setter(e.target.value)}
                      >
                        <option value="">Select a model...</option>
                        {ccModels.map((m) => (
                          <option key={m.id} value={m.id}>
                            {(m.name as string | undefined) ?? m.id}
                          </option>
                        ))}
                      </select>
                    </div>
                  ))}
                </div>
              )}
            </Card>
          )}

          {/* Cavoti Model Selection — only shown when cavoti is selected.
              WHY manual text Inputs, not a dropdown fed by a "Fetch Available
              Models" button (unlike CodeCraftAPI above): the backend has no
              GET /admin/cavoti/models proxy endpoint. Admin copies the model
              ID from the Cavoti dashboard's Models tab and pastes it here. */}
          {activeProvider === 'cavoti' && (
            <Card>
              <p className="text-sm font-medium text-text-primary mb-1">
                Cavoti Model Selection
              </p>
              <p className="text-xs text-text-secondary mb-3">
                Enter the model ID for each tier, copied from the Cavoti dashboard's Models tab.
              </p>
              <div className="flex flex-col gap-3">
                <Input
                  label="Cheap tier (tagging, metadata, coverage checks)"
                  placeholder="e.g. deepseek-chat"
                  value={cavotiModelCheap}
                  onChange={(e) => setCavotiModelCheap(e.target.value)}
                />
                <Input
                  label="Strong tier (answer generation, charter extraction)"
                  placeholder="e.g. claude-3-5-sonnet"
                  value={cavotiModelStrong}
                  onChange={(e) => setCavotiModelStrong(e.target.value)}
                />
                <Input
                  label="Fast tier (quick checks, simple tasks)"
                  placeholder="e.g. gemini-flash"
                  value={cavotiModelFast}
                  onChange={(e) => setCavotiModelFast(e.target.value)}
                />
              </div>
            </Card>
          )}

          {error && <p className="text-sm text-mode-refuse">{error}</p>}
          {success && <p className="text-sm text-mode-advise">{success}</p>}

          <Button
            onClick={() => mutation.mutate()}
            isLoading={mutation.isPending}
            className="self-start"
          >
            Save LLM Settings
          </Button>

          {/* Embedding Settings — always visible, separate from LLM provider */}
          <div className="border-t border-surface-border pt-6">
            <h2 className="text-base font-semibold text-text-primary mb-1">Embedding Settings</h2>
            <p className="text-sm text-text-secondary mb-4">
              Choose where embeddings are generated. Used during transcript ingestion and semantic search.
            </p>

            <Card>
              <p className="text-sm font-medium text-text-primary mb-3">Embedding Provider</p>
              <div className="flex flex-col gap-2 mb-4">
                {(['sidecar', 'codecraftapi'] as const).map((ep) => (
                  <label
                    key={ep}
                    className={`flex items-start gap-3 rounded-lg border p-3 cursor-pointer ${
                      activeEmbeddingProvider === ep
                        ? 'border-brand bg-brand/5'
                        : 'border-surface-border'
                    }`}
                  >
                    <input
                      type="radio"
                      name="embeddingProvider"
                      value={ep}
                      checked={activeEmbeddingProvider === ep}
                      onChange={() => setEmbeddingProvider(ep)}
                      className="mt-0.5"
                    />
                    <div>
                      <p className="text-sm font-medium text-text-primary">
                        {ep === 'sidecar'
                          ? 'Python sidecar (local, default)'
                          : 'CodeCraftAPI'}
                      </p>
                      <p className="text-xs text-text-secondary">
                        {ep === 'sidecar'
                          ? 'bge-base-en-v1.5 running locally. Zero cost. Always available.'
                          : "Use CodeCraftAPI's embedding endpoint. Requires CodeCraftAPI key."}
                      </p>
                    </div>
                  </label>
                ))}
              </div>

              {/* Embedding model picker — only when codecraftapi selected */}
              {activeEmbeddingProvider === 'codecraftapi' && (
                <div className="mb-4">
                  <p className="text-xs text-text-secondary mb-1">Embedding model</p>
                  {ccModels.length > 0 ? (
                    <select
                      className="w-full rounded-md border border-surface-border bg-surface-secondary text-text-primary text-sm px-3 py-2"
                      value={embeddingModel || embData?.embeddingModel || ''}
                      onChange={(e) => setEmbeddingModel(e.target.value)}
                    >
                      <option value="">Select a model...</option>
                      {ccModels.map((m) => (
                        <option key={m.id} value={m.id}>
                          {(m.name as string | undefined) ?? m.id}
                        </option>
                      ))}
                    </select>
                  ) : (
                    <Input
                      label=""
                      placeholder="Fetch models above, or type model ID manually"
                      value={embeddingModel || embData?.embeddingModel || ''}
                      onChange={(e) => setEmbeddingModel(e.target.value)}
                    />
                  )}
                </div>
              )}

              {/* Re-ingestion warning — always visible when codecraftapi embedding is selected.
                  WHY always visible (not just on change): admin must see this BEFORE saving,
                  not after. Hiding it until they click Save would be too late. */}
              {activeEmbeddingProvider === 'codecraftapi' && (
                <div className="rounded-md border border-mode-warn/40 bg-mode-warn/5 p-3">
                  <p className="text-xs text-mode-warn font-medium">
                    {'⚠️'} Changing embedding provider requires re-ingesting ALL transcripts.
                  </p>
                  <p className="text-xs text-text-secondary mt-1">
                    Existing vectors were generated by the previous provider and are incompatible
                    with a different embedding space. After saving, re-upload all transcripts.
                  </p>
                </div>
              )}
            </Card>

            {embError && <p className="text-sm text-mode-refuse mt-3">{embError}</p>}
            {embSuccess && <p className="text-sm text-mode-advise mt-3">{embSuccess}</p>}

            <Button
              onClick={() => embMutation.mutate()}
              isLoading={embMutation.isPending}
              className="mt-3 self-start"
            >
              Save Embedding Settings
            </Button>
          </div>

          {/* Provider health: is anything being skipped, and is anything slow? */}
          <Card>
            <p className="text-sm font-medium text-text-primary mb-1">Provider health</p>
            <p className="text-xs text-text-secondary mb-3">
              A provider that keeps failing is skipped for a short while instead of being
              asked again on every request — the request goes to the fallback provider when one
              is configured. Latency is the recent p50/p95 per provider; percentiles rather than
              an average, because one slow call in twenty is exactly what an average hides.
            </p>

            {(!health || (health.breakers.length === 0 && health.latency.length === 0)) && (
              <p className="text-xs text-text-disabled">
                No calls have been made since the server started, so there is nothing to report
                yet.
              </p>
            )}

            {health && health.breakers.length > 0 && (
              <div className="mb-4 overflow-x-auto">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="text-left text-text-disabled">
                      <th className="py-1 pr-3 font-medium">Provider</th>
                      <th className="py-1 pr-3 font-medium">State</th>
                      <th className="py-1 pr-3 font-medium">Failures</th>
                      <th className="py-1 font-medium">Last error</th>
                    </tr>
                  </thead>
                  <tbody>
                    {health.breakers.map((b) => (
                      <tr key={b.provider}>
                        <td className="py-1 pr-3 text-text-secondary">{b.provider}</td>
                        <td className="py-1 pr-3">
                          <span
                            className={`rounded px-1.5 py-0.5 text-[10px] font-medium ${
                              b.state === 'open'
                                ? 'bg-mode-refuse/15 text-mode-refuse'
                                : b.state === 'half_open'
                                  ? 'bg-glow-amber/15 text-glow-amber'
                                  : 'bg-mode-advise/15 text-mode-advise'
                            }`}
                          >
                            {b.state === 'open'
                              ? `skipped — next try in ${b.retryAfterSeconds ?? 0}s`
                              : b.state === 'half_open'
                                ? 'testing recovery'
                                : 'working'}
                          </span>
                        </td>
                        <td className="py-1 pr-3 text-text-secondary">{b.consecutiveFailures}</td>
                        <td className="py-1 text-text-disabled">
                          {b.lastError ? b.lastError.slice(0, 120) : '—'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            {health && health.latency.length > 0 && (
              <div className="overflow-x-auto">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="text-left text-text-disabled">
                      <th className="py-1 pr-3 font-medium">Provider</th>
                      <th className="py-1 pr-3 font-medium">Calls</th>
                      <th className="py-1 pr-3 font-medium">p50</th>
                      <th className="py-1 pr-3 font-medium">p95</th>
                      <th className="py-1 font-medium">Slowest</th>
                    </tr>
                  </thead>
                  <tbody>
                    {health.latency.map((l) => (
                      <tr key={l.provider}>
                        <td className="py-1 pr-3 text-text-secondary">{l.provider}</td>
                        <td className="py-1 pr-3 text-text-secondary">{l.callsTotal}</td>
                        <td className="py-1 pr-3 text-text-secondary">
                          {(l.p50Ms / 1000).toFixed(1)}s
                        </td>
                        <td className="py-1 pr-3 text-text-secondary">
                          {(l.p95Ms / 1000).toFixed(1)}s
                        </td>
                        <td className="py-1 text-text-secondary">{(l.maxMs / 1000).toFixed(1)}s</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            {health?.note && <p className="mt-2 text-[10px] text-text-disabled">{health.note}</p>}
          </Card>

          {/* Per-model token limits.
              WHY this screen and not env vars: the right number depends on the
              model, and the same deployment serves several providers through
              different keys. A limit that is too low does not fail loudly — a
              reasoning model simply returns an empty answer — so the fix has to
              be reachable without a redeploy. */}
          <Card>
            <p className="text-sm font-medium text-text-primary mb-1">Model token limits</p>
            <p className="text-xs text-text-secondary mb-3">
              The most this system will ask a model for. Output is the number that matters for
              reasoning models: they spend part of it thinking before writing anything, so too
              low comes back as an empty reply. <span className="text-text-primary">0</span> means
              "use the provider's own maximum". An input limit is enforced only once you set it.
            </p>

            {limitRows.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="text-left text-text-disabled">
                      <th className="py-1 pr-3 font-medium">Provider</th>
                      <th className="py-1 pr-3 font-medium">Tier</th>
                      <th className="py-1 pr-3 font-medium">Max input tokens</th>
                      <th className="py-1 font-medium">Max output tokens</th>
                    </tr>
                  </thead>
                  <tbody>
                    {limitRows.map((row) => (
                      <tr key={limitKey(row)}>
                        <td className="py-1 pr-3 text-text-secondary">{row.provider}</td>
                        <td className="py-1 pr-3 text-text-secondary">
                          {row.tier === '*' ? 'all tiers' : row.tier}
                        </td>
                        <td className="py-1 pr-3">
                          <input
                            type="number"
                            min={0}
                            value={row.maxInputTokens}
                            onChange={(e) =>
                              editLimit(row, { maxInputTokens: Number(e.target.value) || 0 })
                            }
                            className="w-28 rounded border border-surface-border bg-surface-secondary px-2 py-1 text-text-primary"
                          />
                        </td>
                        <td className="py-1">
                          <input
                            type="number"
                            min={0}
                            value={row.maxOutputTokens}
                            onChange={(e) =>
                              editLimit(row, { maxOutputTokens: Number(e.target.value) || 0 })
                            }
                            className="w-28 rounded border border-surface-border bg-surface-secondary px-2 py-1 text-text-primary"
                          />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <p className="text-xs text-text-disabled">
                No limits configured — every provider is using its own maximum.
              </p>
            )}

            <div className="mt-3 flex flex-wrap items-end gap-2">
              <div>
                <p className="text-xs text-text-secondary mb-1">Provider</p>
                <select
                  className="rounded-md border border-surface-border bg-surface-secondary text-text-primary text-sm px-3 py-2"
                  value={newLimitProvider}
                  onChange={(e) => setNewLimitProvider(e.target.value)}
                >
                  {(limitsData?.providers ?? ['anthropic']).map((p) => (
                    <option key={p} value={p}>{p}</option>
                  ))}
                </select>
              </div>
              <div>
                <p className="text-xs text-text-secondary mb-1">Tier</p>
                <select
                  className="rounded-md border border-surface-border bg-surface-secondary text-text-primary text-sm px-3 py-2"
                  value={newLimitTier}
                  onChange={(e) => setNewLimitTier(e.target.value)}
                >
                  {(limitsData?.tiers ?? ['strong', 'fast', 'cheap', '*']).map((t) => (
                    <option key={t} value={t}>{t === '*' ? 'all tiers' : t}</option>
                  ))}
                </select>
              </div>
              <Button
                variant="secondary"
                onClick={() =>
                  editLimit(
                    { provider: newLimitProvider, tier: newLimitTier, maxInputTokens: 0, maxOutputTokens: 0 },
                    {}
                  )
                }
              >
                Add row
              </Button>
            </div>

            {limitsError && <p className="text-sm text-mode-refuse mt-3">{limitsError}</p>}
            {limitsSuccess && <p className="text-sm text-mode-advise mt-3">{limitsSuccess}</p>}

            <Button
              onClick={() => limitsMutation.mutate()}
              isLoading={limitsMutation.isPending}
              disabled={Object.keys(limitDrafts).length === 0}
              className="mt-3"
            >
              Save Token Limits
            </Button>
          </Card>

        </div>
      )}
    </div>
  )
}

export const Component = AdminLLMSettings
