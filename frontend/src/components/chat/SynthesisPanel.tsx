import type { SynthesisResult } from '@/types/expert'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 1, 10 wireframes:
 * "\u26a1 SYNTHESIS / \u2705 Agreement: ... / \u26a0\ufe0f Contradiction: ... YOU DECIDE"
 *
 * WHY contradictions render with an explicit "YOU DECIDE" callout
 * (not just listed as data): this matches the exact wireframe text
 * and the product's core design principle from
 * AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 1 ("Let the client
 * decide" - the system never silently picks a winner between
 * contradicting experts).
 */
export function SynthesisPanel({ synthesis }: { synthesis: SynthesisResult }) {
  return (
    <div className="rounded-lg border border-brand/30 bg-brand/5 p-4">
      <p className="mb-2 text-sm font-medium text-brand">\u26a1 SYNTHESIS</p>

      {synthesis.agreements.map((agreement, i) => (
        <p key={i} className="text-sm text-mode-advise">
          \u2705 Agreement: {agreement}
        </p>
      ))}

      {synthesis.contradictions.map((c, i) => (
        <div key={i} className="mt-2 text-sm text-mode-warn">
          <p>
            \u26a0\ufe0f Contradiction on {c.topic}: {c.expertA} says "{c.positionA}", {c.expertB} says "
            {c.positionB}" \u2014 <span className="font-medium">YOU DECIDE</span>
          </p>
        </div>
      ))}

      {synthesis.summary && <p className="mt-3 text-sm text-text-secondary">{synthesis.summary}</p>}
    </div>
  )
}
