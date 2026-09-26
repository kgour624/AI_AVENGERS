import type { ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getMessageExplanation } from '@/api/messages'
import { Modal } from '@/components/ui/Modal'
import { Skeleton } from '@/components/ui/Skeleton'

/**
 * AnswerExplanation — "why does this answer look like this?" (C9).
 *
 * WHY this screen exists: everything below the answer — which gates ran, where
 * one stopped, which chunks were used, how the claims were checked, what the
 * quality judge said, whether the provenance chain verified — has been recorded
 * since C8/C9 and was reachable only as JSON from a shell. The people who have to
 * trust the answer are the ones who cannot read a shell.
 *
 * The one rule this screen keeps: a section that did not run says so. An empty
 * citation list is "no sources recorded", not "no sources needed"; a missing
 * quality verdict is "not judged", never a zero; a missing integrity block is
 * "no chain was built for this answer", not "verified". Anything else would make
 * absence look like good news.
 */

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="mb-4">
      <p className="mb-1 text-[10px] font-semibold uppercase tracking-wider text-text-disabled">
        {title}
      </p>
      {children}
    </div>
  )
}

function NotRecorded({ what }: { what: string }) {
  return <p className="text-xs text-text-disabled">{what}</p>
}

export function AnswerExplanation({ messageId, onClose }: { messageId: string; onClose: () => void }) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['messages', messageId, 'explanation'],
    queryFn: () => getMessageExplanation(messageId),
  })

  return (
    <Modal isOpen onClose={onClose}>
      <h3 className="mb-4 text-sm font-medium text-text-primary">Why this answer</h3>

      {isLoading ? (
        <Skeleton className="h-40" />
      ) : isError || !data ? (
        // Not the same as "nothing was recorded": the lookup itself failed, and
        // saying so keeps the reader from reading a failed fetch as a clean bill.
        <p className="text-sm text-mode-refuse">
          Could not load the explanation. That is a failure to read it, not evidence about the
          answer.
        </p>
      ) : (
        <div className="max-h-[70vh] overflow-y-auto">
          <Section title="Answer">
            <p className="text-xs text-text-secondary">
              Mode {data.message.mode || 'unknown'} · stopped at gate {data.message.gateStopped}
              {typeof data.message.confidence === 'number' &&
                ` · confidence ${data.message.confidence.toFixed(2)}`}
            </p>
          </Section>

          <Section title="Why it went this way">
            {data.decision?.route || data.decision?.reason ? (
              <p className="text-xs text-text-secondary">
                {data.decision.route ?? ''}
                {data.decision.reason ? ` — ${data.decision.reason}` : ''}
                {data.decision.domain ? ` (domain: ${data.decision.domain})` : ''}
              </p>
            ) : (
              <NotRecorded what="No routing decision was recorded for this answer." />
            )}
          </Section>

          <Section title="Gates">
            {data.gates.length === 0 ? (
              <NotRecorded what="No gate steps were recorded." />
            ) : (
              <ul className="space-y-1">
                {data.gates.map((g) => (
                  <li key={g.gate} className="text-xs">
                    <span className={g.passed ? 'text-mode-advise' : 'text-mode-refuse'}>
                      {g.passed ? 'passed' : 'stopped'}
                    </span>
                    <span className="text-text-secondary">
                      {' '}
                      gate {g.gate}
                      {g.name ? ` (${g.name})` : ''}
                    </span>
                    {g.detail && <span className="block text-[10px] text-text-disabled">{g.detail}</span>}
                  </li>
                ))}
              </ul>
            )}
          </Section>

          <Section title="Sources">
            {data.sources.length === 0 ? (
              <NotRecorded what="No sources were recorded — this answer is not backed by the expert's material." />
            ) : (
              <ul className="space-y-1">
                {data.sources.map((s) => (
                  <li key={s.chunkId} className="text-xs text-text-secondary">
                    {s.topic || 'untitled'}
                    {s.sourceFile ? ` · ${s.sourceFile}` : ''}
                    {typeof s.rerankScore === 'number' ? ` · score ${s.rerankScore.toFixed(2)}` : ''}
                  </li>
                ))}
              </ul>
            )}
          </Section>

          <Section title="Quality">
            {typeof data.quality?.score === 'number' ? (
              <p className="text-xs text-text-secondary">
                {data.quality.score.toFixed(2)}
                {data.quality.verdict ? ` · ${data.quality.verdict}` : ''}
                {data.quality.method ? ` · via ${data.quality.method}` : ''}
                {typeof data.quality.attempts === 'number' ? ` · ${data.quality.attempts} attempt(s)` : ''}
              </p>
            ) : (
              <NotRecorded what="This answer was not scored by the quality judge." />
            )}
          </Section>

          <Section title="Claim verification">
            {data.claims ? (
              <pre className="overflow-x-auto whitespace-pre-wrap text-[10px] text-text-secondary">
                {JSON.stringify(data.claims, null, 2)}
              </pre>
            ) : (
              <NotRecorded what="No claim verification was recorded for this answer." />
            )}
          </Section>

          <Section title="Provenance">
            {data.integrity ? (
              <p className="text-xs text-text-secondary">
                {data.integrity.verified === true
                  ? 'chain verified'
                  : data.integrity.verified === false
                    ? 'chain present but NOT verified'
                    : 'chain present, verification state unknown'}
                {data.integrity.chain ? ` · ${data.integrity.chain}` : ''}
              </p>
            ) : (
              <NotRecorded what="No provenance chain exists for this answer." />
            )}
          </Section>

          {data.refusal && (
            <Section title="Refusal">
              <pre className="overflow-x-auto whitespace-pre-wrap text-[10px] text-text-secondary">
                {JSON.stringify(data.refusal, null, 2)}
              </pre>
            </Section>
          )}
        </div>
      )}
    </Modal>
  )
}
