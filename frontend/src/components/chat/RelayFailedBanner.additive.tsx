import { useState } from 'react'
import { useSSEStream } from '@/hooks/useSSEStream'

type Props = {
  chatId: string
  runId: string
  failedAtIndex: number
  reason?: string
  retryUntil?: string
  expertIds: string[]
}

export default function RelayFailedBannerAdditive({ chatId, runId, failedAtIndex, reason, retryUntil, expertIds }: Props) {
  const { sendMessage } = useSSEStream()
  const [retrying, setRetrying] = useState(false)
  const expired = retryUntil ? new Date(retryUntil).getTime() < Date.now() : false

  const handleRetry = async () => {
    if (expired || retrying) return
    setRetrying(true)
    try {
      await sendMessage({
        chatId,
        message: '',
        expertIds,
        answerMode: 'collaborative',
        resumeRunId: runId,
      })
    } finally {
      setRetrying(false)
    }
  }

  return (
    <div className="rounded-lg border border-amber-500/30 bg-amber-500/10 p-3 flex items-center justify-between gap-3">
      <div className="text-sm">
        <p className="font-medium text-amber-700 dark:text-amber-300">Relay failed at step {failedAtIndex + 1}</p>
        {reason && <p className="text-xs opacity-80">{reason}</p>}
        {retryUntil && <p className="text-xs opacity-60">Retry until: {new Date(retryUntil).toLocaleString()} {expired ? '(expired)' : ''}</p>}
      </div>
      <button type="button" onClick={handleRetry} disabled={expired || retrying} className="shrink-0 rounded-md bg-amber-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-amber-700 disabled:opacity-50 disabled:cursor-not-allowed">
        {retrying ? 'Retrying…' : expired ? 'Expired' : 'Retry'}
      </button>
    </div>
  )
}
