import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { handleAPIError } from '@/utils/errors'
import { compareRetrieval } from '@/api/admin'

export function ComparePanel({ expertId }: { expertId: string }) {
  const [query, setQuery] = useState('')
  const m = useMutation({
    mutationFn: () => compareRetrieval({ expert_id: expertId, query, top_k: 5 }),
  })
  const data = m.data as any
  const offResults: any[] = data?.offResults ?? data?.off_results ?? []
  const onResults: any[] = data?.onResults ?? data?.on_results ?? []
  const note = data?.note ?? ''
  const flagWasOn = data?.flagWasOn ?? data?.flag_was_on ?? false

  return (
    <Card>
      <h3 className="text-sm font-semibold">A/B Retrieval — OFF vs ON</h3>
      <p className="mt-1 text-xs text-text-secondary">Same query, OFF (children only) vs ON (flag + parent expansion). Read-only, flag not mutated.</p>
      <div className="mt-3 flex gap-2">
        <input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="query e.g. risk assessment" className="flex-1 rounded border border-surface-border bg-surface-secondary px-2 py-1.5 text-sm" />
        <Button size="sm" disabled={!expertId || !query.trim() || m.isPending} onClick={() => m.mutate()}>{m.isPending ? '…' : 'Compare'}</Button>
      </div>
      {m.error ? <p className="mt-2 text-xs text-mode-refuse">{handleAPIError(m.error as unknown as Error)}</p> : null}
      {data ? (
        <div className="mt-3 space-y-2">
          <p className="rounded bg-surface-secondary px-2 py-1 text-xs text-text-disabled">{note} · flag was {flagWasOn ? 'ON' : 'OFF'}</p>
          <div className="grid gap-2 md:grid-cols-2">
            <div>
              <p className="text-xs font-medium">OFF ({offResults.length})</p>
              {offResults.map((r: any) => (
                <p key={r.id} className="mt-1 line-clamp-2 text-xs text-text-secondary">
                  c{r.chunkIndex ?? r.chunk_index}: {(r.chunkText ?? r.chunk_text ?? '').slice(0, 140)}{(r.parentId ?? r.parent_id) ? ' →P' : ' (orphan)'}
                </p>
              ))}
            </div>
            <div>
              <p className="text-xs font-medium">ON ({onResults.length})</p>
              {onResults.map((r: any) => (
                <p key={r.id} className="mt-1 line-clamp-2 text-xs text-text-secondary">
                  {(r.chunkText ?? r.chunk_text ?? '').slice(0, 140)}
                </p>
              ))}
            </div>
          </div>
        </div>
      ) : null}
    </Card>
  )
}
