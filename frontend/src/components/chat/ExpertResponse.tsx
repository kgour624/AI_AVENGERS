import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import { motion, useReducedMotion } from 'framer-motion'
import type { ExpertResponse as ExpertResponseType } from '@/types/expert'
import { ModeBadge } from '@/components/ui/Badge'
import { CitationChip } from './CitationChip'
import { CodeBlock } from './CodeBlock'
import { RatingWidget } from './RatingWidget'
import { useReplyStore } from '@/stores/replyStore'
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
  /**
   * CT-D5: required to key replyStore (CT-D3). Optional in the type
   * only because some historical call sites may not have it yet —
   * the reply button silently does nothing without it, never throws.
   */
  chatId?: string
}

export function ExpertResponse({ response, persistedMessageId, isStreaming, chatId }: ExpertResponseProps) {
  const [showReasoning, setShowReasoning] = useState(false)
  const [copied, setCopied] = useState(false)
  const reduceMotion = useReducedMotion()
  const setReplyTarget = useReplyStore((s) => s.setReplyTarget)

  function handleCopyResponse() {
    const text =
      response.templateSections && response.templateSections.length > 0
        ? response.templateSections.map((s) => s.label + ':\n' + s.content).join('\n\n')
        : response.content
    navigator.clipboard
      .writeText(text)
      .then(() => {
        setCopied(true)
        setTimeout(() => setCopied(false), 1500)
      })
      .catch((err) => {
        console.warn('copy to clipboard failed', err)
      })
  }

  function handleReplyClick() {
    if (!chatId || !persistedMessageId) return
    const preview =
      response.content.length > 80 ? `${response.content.slice(0, 80)}…` : response.content
    setReplyTarget(chatId, {
      messageId: persistedMessageId,
      preview: preview || '(structured answer)',
      expertId: response.expertId,
      expertName: response.expertName,
    })
  }

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

  // Build chunkId -> 1-based index map so inline citations match the
  // bottom citations list. Both use the same numbering.
  // 2026-09-08 RCA: `?? []` guard - see splitContentByCitations' own
  // doc comment (parseCitations.ts) for the full incident this
  // defends against (a nil backend slice marshaling as JSON null).
  const citationIndexMap = new Map(
    (response.citations ?? []).map((c, i) => [c.chunkId, i + 1])
  )

  return (
    <motion.div
      initial={reduceMotion ? undefined : { opacity: 0, y: 8 }}
      animate={{ opacity: isStreaming ? 0.8 : 1, y: 0 }}
      transition={{ duration: ARC_MOTION.panel, ease: ARC_MOTION.ease }}
      className={cn(
        'rounded-lg border-l-4 border-y border-r border-glass-border bg-surface-raised/80 p-4 backdrop-blur-xl',
        MODE_BORDER_GLOW[response.mode]
      )}
    >
      <div className="mb-2 flex items-center justify-between">
        <p className="text-sm font-medium text-text-primary">{response.expertName}</p>
        <div className="flex items-center gap-2">
          <ModeBadge mode={response.mode} />
          <span className="text-xs text-text-secondary">{Math.round(response.confidence * 100)}%</span>
          <button
            type="button"
            onClick={handleCopyResponse}
            aria-label="Copy response"
            className="text-xs text-text-secondary hover:text-text-primary"
          >
            {copied ? 'Copied!' : 'Copy'}
          </button>
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
      ) : response.templateSections && response.templateSections.length > 0 ? (
        <div className="flex flex-col gap-3">
          {response.templateSections.map((section) => {
            const sectionSegments = splitContentByCitations(section.content, section.citations)
            return (
              <div key={section.key}>
                <h4 className="mb-1 text-xs font-semibold uppercase tracking-wide text-text-secondary">
                  {section.label}
                </h4>
                {section.type === 'code' ? (
                  <ReactMarkdown components={{ code: CodeBlock }}>
                    {'```\n' + section.content + '\n```'}
                  </ReactMarkdown>
                ) : (
                  <div className="prose prose-invert prose-sm max-w-none text-text-primary">
                    {sectionSegments.map((segment, i) =>
                      segment.type === 'text' ? (
                        <ReactMarkdown key={i} components={{ code: CodeBlock }}>
                          {segment.value}
                        </ReactMarkdown>
                      ) : (
                        <CitationChip key={i} citation={segment.citation} index={i + 1} />
                      )
                    )}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      ) : (
        <div className="prose prose-invert prose-sm max-w-none text-text-primary">
          {segments.map((segment, i) =>
            segment.type === 'text' ? (
              <ReactMarkdown key={i} components={{ code: CodeBlock }}>
                {segment.value}
              </ReactMarkdown>
            ) : (
              <CitationChip
                key={i}
                citation={segment.citation}
                index={citationIndexMap.get(segment.citation.chunkId) ?? i + 1}
              />
            )
          )}
        </div>
      )}

      {response.citations && response.citations.length > 0 && (
        <div className="mt-3 flex flex-wrap items-center gap-1.5">
          <span className="text-xs text-text-secondary">Citations:</span>
          {response.citations.map((c, i) => (
            <CitationChip key={c.chunkId} citation={c} index={i + 1} />
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

        <div className="flex items-center gap-3">
          {/* CT-D5: reply button. Disabled until this response is
              persisted (same gating RatingWidget already uses above). */}
          {chatId && persistedMessageId && (
            <button
              type="button"
              onClick={handleReplyClick}
              className="text-xs text-text-secondary hover:text-text-primary"
            >
              Reply
            </button>
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
      </div>

      {showReasoning && (
        <div className="mt-2 rounded-md border border-surface-border bg-surface-overlay p-3 text-xs text-text-secondary space-y-1">
          <p><span className="text-text-primary font-medium">Gate stopped:</span> {response.gateStopped} of 5</p>
          <p><span className="text-text-primary font-medium">Mode:</span> {response.mode}</p>
          <p><span className="text-text-primary font-medium">Confidence:</span> {Math.round(response.confidence * 100)}%</p>
          {response.gateStopped === 0 && <p className="text-mode-advise">{'\u2713'} Passed all 5 gates — full cited answer generated.</p>}
          {response.gateStopped === 1 && <p className="text-mode-ask">Stopped at Gate 1: insufficient information to answer. Clarifying questions raised.</p>}
          {response.gateStopped === 2 && <p className="text-mode-refuse">Stopped at Gate 2: question not covered by training material.</p>}
          {response.gateStopped === 3 && <p className="text-mode-warn">Gate 3 warning: answer may conflict with expert charter rules. Proceeded with caution.</p>}
          {response.gateStopped === 4 && <p className="text-mode-pushback">Stopped at Gate 4: solution may be over-engineered for current needs.</p>}
          {response.gateStopped === 5 && <p className="text-mode-refuse">Stopped at Gate 5: China Wall — could not produce a fully cited answer after retries.</p>}
          <p className="text-text-disabled">See AI_AVENGERS_SYSTEM_ARCHITECTURE.md {'\u00a7'}10 for gate details.</p>
        </div>
      )}
    </motion.div>
  )
}
