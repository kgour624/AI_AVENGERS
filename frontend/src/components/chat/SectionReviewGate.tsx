import { useState } from 'react'
import type { PendingRelayReview } from '@/types/expert'
import { approveRelaySection } from '@/api/relay'

export function SectionReviewGate({
  review,
  onDone,
}: {
  review: PendingRelayReview
  onDone: () => void
}) {
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState(review.content)
  const [saving, setSaving] = useState(false)
  const [resolution, setResolution] = useState<'human_merged' | undefined>(undefined)

  const submit = async (editedContent?: string, resolutionSource?: string) => {
    setSaving(true)
    try {
      await approveRelaySection(review.runId, review.sectionId, {
        edited_content: editedContent,
        resolution_source: resolutionSource,
      })
      onDone?.()
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="my-2 rounded-md border border-amber-500/40 bg-slate-900 p-4">
      <div className="mb-1 text-sm font-semibold text-amber-300">
        Review before the next expert — {review.expertName} · {review.sectionTitle}
      </div>

      {review.conflict && (
        <div className="mb-3 rounded border border-red-500/40 bg-red-950/30 p-3 text-xs text-red-200">
          <div className="font-semibold">This section conflicts with an earlier one:</div>
          <p className="mt-1">{review.conflictExplanation ?? 'No explanation available.'}</p>
          <div className="mt-2 flex gap-2">
            <button onClick={() => submit(undefined, 'expert_a')}
              className="rounded bg-slate-700 px-2 py-1 hover:bg-slate-600">Use earlier</button>
            <button onClick={() => submit(undefined, 'expert_b')}
              className="rounded bg-slate-700 px-2 py-1 hover:bg-slate-600">Use this</button>
          </div>
        </div>
      )}

      <div className="max-h-72 overflow-y-auto rounded border border-slate-700 bg-slate-950 p-3 text-sm text-slate-200 whitespace-pre-wrap">
        {editing ? (
          <textarea
            className="h-64 w-full resize-y bg-slate-950 text-slate-100 outline-none"
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
          />
        ) : (
          draft
        )}
      </div>

      <div className="mt-3 flex items-center justify-end gap-2">
        {editing ? (
          <>
            <button onClick={() => { setEditing(false); setDraft(review.content) }}
              className="rounded bg-slate-700 px-3 py-1 text-sm hover:bg-slate-600">Cancel</button>
            <button onClick={() => { setEditing(false); setResolution('human_merged') }}
              className="rounded bg-slate-600 px-3 py-1 text-sm hover:bg-slate-500">Save draft</button>
          </>
        ) : (
          <button onClick={() => setEditing(true)}
            className="rounded bg-slate-700 px-3 py-1 text-sm hover:bg-slate-600">Edit</button>
        )}
        <button
          disabled={saving}
          onClick={() => submit(editing ? draft : undefined, resolution)}
          className="rounded bg-emerald-600 px-4 py-1 text-sm font-medium text-white hover:bg-emerald-500 disabled:opacity-50"
        >
          {saving ? 'Saving…' : 'Approve & Continue'}
        </button>
      </div>
    </div>
  )
}
