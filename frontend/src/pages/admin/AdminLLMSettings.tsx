import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getLLMSettings, updateLLMSettings } from '@/api/admin'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card } from '@/components/ui/Card'
import { handleAPIError } from '@/utils/errors'

const PROVIDERS = [
  { value: 'openrouter', label: 'OpenRouter (multi-model gateway)', desc: 'Single key, access to Claude, GPT, Gemini, DeepSeek and more' },
  { value: 'deepseek',   label: 'DeepSeek (direct)',               desc: 'Direct DeepSeek API — use your own DeepSeek key' },
  { value: 'anthropic',  label: 'Anthropic (direct)',              desc: 'Direct Anthropic API — use your own Claude key' },
  { value: 'gemini',     label: 'Google Gemini (direct)',          desc: 'Direct Gemini API — use your own Google AI key' },
]

function AdminLLMSettings() {
  const queryClient = useQueryClient()
  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'llm-settings'],
    queryFn: getLLMSettings,
  })

  const [provider, setProvider] = useState('')
  const [keys, setKeys] = useState<Record<string, string>>({})
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const mutation = useMutation({
    mutationFn: () =>
      updateLLMSettings({
        provider: provider || data?.activeProvider || 'openrouter',
        apiKeys: keys,
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

  const activeProvider = provider || data?.activeProvider || 'openrouter'

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
                    placeholder={existing ? 'Enter new key to replace' : 'sk-...'}
                    value={keys[p.value] ?? ''}
                    onChange={(e) =>
                      setKeys((prev) => ({ ...prev, [p.value]: e.target.value }))
                    }
                  />
                )
              })}
            </div>
          </Card>

          {error && <p className="text-sm text-mode-refuse">{error}</p>}
          {success && <p className="text-sm text-mode-advise">{success}</p>}

          <Button
            onClick={() => mutation.mutate()}
            isLoading={mutation.isPending}
            className="self-start"
          >
            Save Settings
          </Button>
        </div>
      )}
    </div>
  )
}

export const Component = AdminLLMSettings
