import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import { motion, useReducedMotion } from 'framer-motion'
import type { ExpertResponse as ExpertResponseType } from '@/types/expert'
import { ModeBadge } from '@/components/ui/Badge'
import { CitationChip } from './CitationChip'
import { CodeBlock } from './CodeBlock'
import { RatingWidget } from './RatingWidget'
import { splitContentByCitations } from '@/utils/parseCitations'
import { cn } from '@/utils/cn'
import { ARC_MOTION } from '@/design-system/motion'

/**
 * ARC-51 §9 (docs/ARC51_UI_CONTRACT.md): mode -> left-border glow
 * color, reusing the EXISTING product-semantic mode colors (§2's
 * rule: mode colors encode Gate/China Wall output, never repurposed
 * as decoration elsewhere). Not a new color system - just a new place
 * these already-meaningful colors get used.
 */
const MODE_BORDER_GLOW: Record<ExpertResponseType['mode'], string> = {
  ADVISE: 'border-l-mode-advise shadow-[-4px_0_16px_-8px_var(--color-advise)]',
  ASK: 'border-l-mode-ask shadow-[-4px_0_16px_-8px_var(--color-ask)]',
  WARN: 'border-l-mode-warn shadow-[-4px_0_16px_-8px_var(--color-warn)]',
  PUSH_BACK: 'border-l-mode-pushback shadow-[-4px_0_16px_-8px_var(--color-pushback)]',
  REFUSE: 'border-l-mode-refuse shadow-[-4px_0_16px_-8px_var(--color-refuse)]',
}

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 9 ("This is the core value
 * proposition of AI Avengers. Every design decision here affects
 * trust and usability.") and section 10's Chat Page wireframe (mode
 * badge + confidence %, markdown content with inline citations, code
 * blocks, citation list, rating, collapsible "Why did I say this?").
 *
 * `persistedMessageId` prop design decision: see RatingWidget.tsx's
 * header comment - rating only applies to already-saved messages.
 * This component is used for BOTH live-streaming ExpertResponse
 * objects (no messageId yet) and persisted Message rows rendered from
 * getMessages() (has an id) - passing persistedMessageId=undefined for
 * the former simply hides the rating UI until the turn is saved and
 * the page's message list is refetched. This mirrors exactly how the
 * ASK-mode "Why did I say this?" collapsible section only makes sense
 * once there's a decision to explain, not mid-stream.
 */
export interface ExpertResponseProps {
  response: ExpertResponseType
  persistedMessageId?: string
  isStreaming?: boolean
}

export function ExpertResponse({ response, persistedMessageId, isStreaming }: ExpertResponseProps) {
  const [showReasoning, setShowReasoning] = useState(false)

  // WHY guard on response.error before rendering the normal body: a
  // partially-failed expert (one expert's goroutine errored while
  // others succeeded, per AI_AVENGERS_SYSTEM_ARCHITECTURE.md section
  // 4.1's "Failure behavior") still has expertId/expertName but no
  // real content/mode to render meaningfully - rendering the error
  // message here instead of an empty/garbled response body.
  if (response.error) {
    return (
      <div className="rounded-lg border border-mode-refuse/30 bg-mode-refuse/5 p-4">
        <p className="text-sm font-medium text-text-primary">{response.expertName}</p>
        <p className="mt-1 text-sm text-mode-refuse">Failed to respond: {response.error}</p>
      </div>
    )
  }

  const segments = splitContentByCitations(response.content, response.citations)

  return (
    <div className={cn('rounded-lg border border-surface-border bg-surface-raised p-4', isStreaming && 'opacity-80')}>
      <div className="mb-2 flex items-center justify-between">
        <p className="text-sm font-medium text-text-primary">{response.expertName}</p>
        <div className="flex items-center gap-2">
          <ModeBadge mode={response.mode} />
          <span className="text-xs text-text-secondary">{Math.round(response.confidence * 100)}%</span>
        </div>
      </div>

      {response.warning && (
        <p className="mb-2 rounded-md bg-mode-warn/10 px-2 py-1 text-xs text-mode-warn">
          {response.warning}
        </p>
      )}

      {/* ASK mode - render clarifying questions instead of prose content. */}
      {response.mode === 'ASK' && response.questions && response.questions.length > 0 ? (
        <ul className="list-inside list-disc space-y-1 text-sm text-text-primary">
          {response.questions.map((q) => (
            <li key={q}>{q}</li>
          ))}
        </ul>
      ) : (
        <div className="prose prose-invert prose-sm max-w-none text-text-primary">
          {segments.map((segment, i) =>
            segment.type === 'text' ? (
              <ReactMarkdown key={i} components={{ code: CodeBlock }}>
                {segment.value}
              </ReactMarkdown>
            ) : (
              <CitationChip key={i} citation={segment.citation} />
            )
          )}
        </div>
      )}

      {response.citations.length > 0 && (
        <div className="mt-3 flex flex-wrap items-center gap-1.5">
          <span className="text-xs text-text-secondary">Citations:</span>
          {response.citations.map((c) => (
            <CitationChip key={c.chunkId} citation={c} />
          ))}
        </div>
      )}

      <div className="mt-3 flex items-center justify-between">
        {persistedMessageId ? (
          <RatingWidget messageId={persistedMessageId} />
        ) : (
          <span className="text-xs text-text-disabled">
            {isStreaming ? 'Streaming...' : 'Rating available after save'}
          </span>
        )}

        {response.gateStopped > 0 && (
          <button
            type="button"
            onClick={() => setShowReasoning((v) => !v)}
            className="text-xs text-text-secondary hover:text-text-primary"
          >
            {showReasoning ? '\u25bc' : '\u25b6'} Why did I stop here?
          </button>
        )}
      </div>

      {showReasoning && (
        <p className="mt-2 text-xs text-text-secondary">
          Stopped at Gate {response.gateStopped} of 5. See{' '}
          <code className="font-mono">AI_AVENGERS_SYSTEM_ARCHITECTURE.md \u00a7 10</code> for what each gate
          checks.
        </p>
      )}
    </div>
  )
}
