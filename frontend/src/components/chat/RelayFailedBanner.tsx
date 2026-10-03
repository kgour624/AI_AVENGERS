import { useState } from 'react'
import type { RelayFailure } from '@/types/expert'
import { retryRelay } from '@/api/relay'

export function RelayFailedBanner({ failure }: { failure: RelayFailure }) {
  const [retrying, setRetrying] = useState(false)
  const expired = new Date(failure.retryUntil).getTime() < Date.now()

  return (
    <div className="my-2 rounded-md border border-red-500/40 bg-red-950/30 p-4 text-sm">
      <div className="font-semibold text-red-300">
        The relay stopped at expert {failure.failedAtIndex + 1}.
      </div>
      <p className="mt-1 text-red-200">{failure.reason}</p>
      <p className="mt-1 text-xs text-red-300/70">
        {expired
          ? 'Retry window elapsed — this answer failed permanently.'
          : `You can retry from where it stopped until ${new Date(failure.retryUntil).toLocaleString()}.`}
      </p>
      {!expired && (
        <button
          disabled={retrying}
          onClick={async () => {
            setRetrying(true)
            try { await retryRelay(failure.runId) } finally { setRetrying(false) }
          }}
          className="mt-2 rounded bg-red-600 px-3 py-1 text-white hover:bg-red-500 disabled:opacity-50"
        >
          {retrying ? 'Retrying…' : 'Retry from failed section'}
        </button>
      )}
    </div>
  )
}
