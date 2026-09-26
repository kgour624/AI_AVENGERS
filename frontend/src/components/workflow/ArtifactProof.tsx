import { useState } from 'react'
import { cn } from '@/utils/cn'
import type { BlackboardEvent } from '@/api/workflows'

/**
 * ArtifactProof — is deliverable ke claims ka koi saaboot hai?
 *
 * WHY this exists: the review gate says "a reviewer approved it", which is a
 * statement about process. It says nothing about whether the claims in the
 * artifact are backed by the material the expert was actually trained on. The
 * verification event (posted by the gateway's artifact verifier) carries that
 * answer, and until it had a screen the whole check was invisible — the exact
 * situation this project keeps hitting: work done, nothing shown.
 *
 * Three honesty rules the rendering follows:
 *  1. "Not checked" is never dressed as "clean". Lookups that did not run
 *     (fail_open / disabled / empty) and checks with no reference material are
 *     stated in plain words, because "0 refuted" over "0 checked" is not a pass.
 *  2. Refuted claims are the loudest thing on the card. They are the only result
 *     that says the expert contradicted its own training.
 *  3. Numbers are shown with what they were measured against (reference_chunks),
 *     so a reader can tell a strong result from a thin one.
 */

/** One claim verdict, as the backend's ClaimReport serialises. */
interface ClaimLine {
  claim: string
  verdict: string
  confidence: number
  justification?: string
}

/** The verification event's content (snake_case on the wire, camelCase here). */
export interface ArtifactVerification {
  artifactEventId: string
  claimMethod: string
  claimsSupported: number
  claimsRefuted: number
  claimsUnverifiable: number
  referenceChunks: number
  qualityOverall: number
  qualityPass: boolean
  qualityMethod: string
  qualityFeedback: string
  claims: ClaimLine[]
}

// parseArtifactVerification reads one blackboard event into the shape above,
// tolerating a missing field rather than rendering "undefined" at the reader.
export function parseArtifactVerification(event: BlackboardEvent): ArtifactVerification | null {
  const c = event.content ?? {}
  const artifactEventId = typeof c.artifactEventId === 'string' ? c.artifactEventId : ''
  if (!artifactEventId) return null

  const num = (v: unknown): number => (typeof v === 'number' ? v : 0)
  const str = (v: unknown): string => (typeof v === 'string' ? v : '')
  const claims: ClaimLine[] = Array.isArray(c.claims)
    ? (c.claims as Record<string, unknown>[])
        .filter((x) => x && typeof x === 'object')
        .map((x) => ({
          claim: str(x.claim),
          verdict: str(x.verdict),
          confidence: num(x.confidence),
          justification: str(x.justification),
        }))
    : []

  return {
    artifactEventId,
    claimMethod: str(c.claimMethod),
    claimsSupported: num(c.claimsSupported),
    claimsRefuted: num(c.claimsRefuted),
    claimsUnverifiable: num(c.claimsUnverifiable),
    referenceChunks: num(c.referenceChunks),
    qualityOverall: num(c.qualityOverall),
    qualityPass: c.qualityPass === true,
    qualityMethod: str(c.qualityMethod),
    qualityFeedback: str(c.qualityFeedback),
    claims,
  }
}

// NotCheckedReason turns a method into the sentence the screen shows. An empty
// string means the check really ran.
function notCheckedReason(v: ArtifactVerification): string {
  switch (v.claimMethod) {
    case 'verify':
      return ''
    case 'no_claims':
      return 'no material claim was found in this deliverable to check'
    case 'fail_open':
      return 'the check could not run (the verifier call failed), so nothing was confirmed'
    case 'empty':
      return 'the deliverable was empty'
    case 'disabled':
      return 'no verifier is configured on this server'
    default:
      return `the check did not run (${v.claimMethod || 'unknown state'})`
  }
}

function verdictWord(verdict: string): string {
  return verdict === 'supported' ? 'backed' : verdict === 'refuted' ? 'contradicted' : verdict
}

const MAX_VISIBLE_CLAIMS = 12

export function ArtifactProof({ verification }: { verification?: ArtifactVerification | null }) {
  const [expanded, setExpanded] = useState(false)

  // No event at all is its own state, and it is not the same as "the check ran
  // and failed" — but both must read as "not checked", never as clean.
  if (!verification) {
    return (
      <div className="mb-3 rounded border border-surface-border bg-surface-secondary/40 p-3">
        <p className="mb-1 text-[10px] font-semibold uppercase tracking-wider text-text-disabled">
          Proof
        </p>
        <p className="text-xs text-text-disabled">
          Not checked — this deliverable has no verification yet. It is checked after the review
          step, so a run still in progress will not have one.
        </p>
      </div>
    )
  }

  const notChecked = notCheckedReason(verification)
  const thinReference = verification.referenceChunks === 0
  const qualityRan = verification.qualityMethod === 'judge'

  return (
    <div className="mb-3 rounded border border-surface-border bg-surface-secondary/40 p-3">
      <p className="mb-2 text-[10px] font-semibold uppercase tracking-wider text-text-disabled">
        Proof
      </p>

      {notChecked ? (
        <p className="text-xs text-glow-amber">
          Not checked — {notChecked}.
        </p>
      ) : (
        <div className="flex flex-wrap items-center gap-2 text-xs">
          <span
            className={cn(
              'rounded px-2 py-0.5 font-medium',
              verification.claimsRefuted > 0
                ? 'bg-mode-refuse/15 text-mode-refuse'
                : verification.claimsUnverifiable > 0
                  ? 'bg-glow-amber/15 text-glow-amber'
                  : 'bg-mode-advise/15 text-mode-advise'
            )}
          >
            {verification.claimsRefuted > 0
              ? `${verification.claimsRefuted} claim${verification.claimsRefuted === 1 ? '' : 's'} contradict the training material`
              : verification.claimsUnverifiable > 0
                ? `${verification.claimsUnverifiable} claim${verification.claimsUnverifiable === 1 ? '' : 's'} not covered by the training material`
                : `${verification.claimsSupported} claim${verification.claimsSupported === 1 ? '' : 's'}, all backed`}
          </span>
          <span className="text-text-secondary">
            {verification.claimsSupported} backed · {verification.claimsRefuted} contradicted ·{' '}
            {verification.claimsUnverifiable} not covered
          </span>
        </div>
      )}

      {/* A count of zero means something different depending on what it was
          measured against, so the reference size is always stated. */}
      <p className="mt-1 text-[10px] text-text-disabled">
        Checked against {verification.referenceChunks} of this expert's training chunk
        {verification.referenceChunks === 1 ? '' : 's'}
        {thinReference && !notChecked
          ? ' — nothing to compare against, so treat this as unchecked rather than clean'
          : ''}
        .
      </p>

      {qualityRan ? (
        <p className="mt-1 text-[10px] text-text-secondary">
          Quality {verification.qualityOverall.toFixed(2)} (
          {verification.qualityPass ? 'passes' : 'below the bar'})
          {verification.qualityFeedback ? ` — ${verification.qualityFeedback}` : ''}
        </p>
      ) : (
        // Never show a score for a rubric that did not run: an absent judgement
        // is not a zero.
        verification.qualityMethod === 'no_question' && (
          <p className="mt-1 text-[10px] text-text-disabled">
            Quality not scored — this deliverable was not produced in answer to a question.
          </p>
        )
      )}

      {verification.claims.length > 0 && (
        <>
          <button
            onClick={() => setExpanded((v) => !v)}
            className="mt-2 text-[10px] font-medium uppercase tracking-wider text-brand hover:underline"
          >
            {expanded ? 'Hide claims' : `Show ${verification.claims.length} claim${verification.claims.length === 1 ? '' : 's'}`}
          </button>
          {expanded && (
            <ul className="mt-2 space-y-2">
              {verification.claims.slice(0, MAX_VISIBLE_CLAIMS).map((claim, i) => (
                <li key={i} className="border-l-2 border-surface-border pl-2">
                  <p className="text-[11px] text-text-secondary">{claim.claim}</p>
                  <p
                    className={cn(
                      'text-[10px] font-medium',
                      claim.verdict === 'refuted'
                        ? 'text-mode-refuse'
                        : claim.verdict === 'supported'
                          ? 'text-mode-advise'
                          : 'text-glow-amber'
                    )}
                  >
                    {verdictWord(claim.verdict)}
                    {claim.justification ? ` — ${claim.justification}` : ''}
                  </p>
                </li>
              ))}
            </ul>
          )}
        </>
      )}
    </div>
  )
}
