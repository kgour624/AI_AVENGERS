import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { rateMessage } from '@/api/messages'
import { Button } from '@/components/ui/Button'
import { cn } from '@/utils/cn'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 10: "⭐⭐⭐⭐⭐  [Rate this response]".
 *
 * Feature #7 enhancement: Added optional text feedback, feedback type
 * dropdown, and submit button. Shows "Thank you" toast after submission.
 *
 * WHY this only ever receives a messageId (never renders against a
 * live in-flight ExpertResponse): rating a response that hasn't been
 * persisted yet has no id to rate against. ExpertResponse (the SSE
 * payload type) deliberately has no messageId field - adding a fake
 * one just to make this component "work" during streaming would let a
 * user click a star that silently does nothing meaningful, or worse,
 * throws once the request actually fires. ChatPage only renders this
 * for messages loaded from getMessages() (i.e. already persisted),
 * never for the live stream - documented as a deliberate scope
 * decision in HANDOFF.md, not a missing feature.
 *
 * Cross-questioned: what if the user clicks a star while a previous
 * rating request is still in flight? useMutation's `isPending` gates
 * the buttons so a second click can't fire a second request before
 * the first resolves - avoids a race where two different scores could
 * both land depending on response order.
 */

const FEEDBACK_TYPES = [
  { value: 'accepted', label: 'Accepted as-is' },
  { value: 'rejected', label: 'Not helpful' },
  { value: 'modified', label: 'Modified before using' },
  { value: 'ignored', label: 'Ignored' },
] as const

export function RatingWidget({ messageId }: { messageId: string }) {
  const [selectedScore, setSelectedScore] = useState<number | null>(null)
  const [showFeedbackForm, setShowFeedbackForm] = useState(false)
  const [feedbackText, setFeedbackText] = useState('')
  const [feedbackType, setFeedbackType] = useState<'accepted' | 'rejected' | 'modified' | 'ignored'>('accepted')
  const [showThankYou, setShowThankYou] = useState(false)

  const mutation = useMutation({
    mutationFn: (data: { score: 1 | 2 | 3 | 4 | 5; feedback?: string; feedbackType?: typeof feedbackType }) =>
      rateMessage(messageId, data),
    onSuccess: (_data, variables) => {
      setSelectedScore(variables.score)
      setShowFeedbackForm(false)
      setShowThankYou(true)
      // Hide thank you message after 3 seconds
      setTimeout(() => setShowThankYou(false), 3000)
    },
  })

  function handleStarClick(score: 1 | 2 | 3 | 4 | 5) {
    setSelectedScore(score)
    setShowFeedbackForm(true)
  }

  function handleSubmit() {
    if (selectedScore === null) return
    
    mutation.mutate({
      score: selectedScore as 1 | 2 | 3 | 4 | 5,
      feedback: feedbackText.trim() || undefined,
      feedbackType: feedbackText.trim() ? feedbackType : undefined,
    })
  }

  function handleSkipFeedback() {
    if (selectedScore === null) return
    mutation.mutate({ score: selectedScore as 1 | 2 | 3 | 4 | 5 })
  }

  // Already rated - show thank you or just the stars
  if (showThankYou) {
    return (
      <div className="flex items-center gap-2 rounded-md border border-mode-advise/30 bg-mode-advise/10 px-3 py-2">
        <span className="text-sm text-mode-advise">✓ Thank you for your feedback!</span>
      </div>
    )
  }

  // Show feedback form after star selection
  if (showFeedbackForm && selectedScore !== null) {
    return (
      <div className="space-y-2 rounded-md border border-surface-border bg-surface-raised p-3">
        <div className="flex items-center gap-1">
          <span className="text-xs text-text-secondary">Your rating:</span>
          {([1, 2, 3, 4, 5] as const).map((score) => (
            <span
              key={score}
              className={cn(
                'text-sm',
                score <= selectedScore ? 'opacity-100' : 'opacity-30'
              )}
            >
              {'⭐'}
            </span>
          ))}
        </div>

        <div className="space-y-2">
          <label className="block">
            <span className="text-xs text-text-secondary">Feedback type (optional)</span>
            <select
              value={feedbackType}
              onChange={(e) => setFeedbackType(e.target.value as typeof feedbackType)}
              className="mt-1 w-full rounded-md border border-surface-border bg-surface-overlay px-2 py-1.5 text-sm text-text-primary focus:border-brand focus:outline-none"
            >
              {FEEDBACK_TYPES.map((type) => (
                <option key={type.value} value={type.value}>
                  {type.label}
                </option>
              ))}
            </select>
          </label>

          <label className="block">
            <span className="text-xs text-text-secondary">Additional feedback (optional)</span>
            <textarea
              value={feedbackText}
              onChange={(e) => setFeedbackText(e.target.value)}
              placeholder="Tell us more about your experience..."
              rows={3}
              className="mt-1 w-full resize-none rounded-md border border-surface-border bg-surface-overlay px-2 py-1.5 text-sm text-text-primary placeholder:text-text-disabled focus:border-brand focus:outline-none"
            />
          </label>
        </div>

        <div className="flex items-center justify-end gap-2">
          <Button
            variant="secondary"
            size="sm"
            onClick={handleSkipFeedback}
            disabled={mutation.isPending}
          >
            Skip
          </Button>
          <Button
            size="sm"
            onClick={handleSubmit}
            disabled={mutation.isPending}
            isLoading={mutation.isPending}
          >
            Submit
          </Button>
        </div>

        {mutation.isError && (
          <p className="text-xs text-mode-refuse">Failed to submit rating. Please try again.</p>
        )}
      </div>
    )
  }

  // Initial state - show stars
  return (
    <div className="flex items-center gap-1">
      <span className="text-xs text-text-secondary">Rate:</span>
      {([1, 2, 3, 4, 5] as const).map((score) => (
        <button
          key={score}
          type="button"
          disabled={mutation.isPending}
          onClick={() => handleStarClick(score)}
          aria-label={`Rate ${score} stars`}
          className="text-sm opacity-40 hover:opacity-100 disabled:cursor-not-allowed"
        >
          {'⭐'}
        </button>
      ))}
    </div>
  )
}
