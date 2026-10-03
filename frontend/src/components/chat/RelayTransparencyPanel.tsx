import type { RelayStepEvent } from '@/types/expert'
import { cn } from '@/utils/cn'

const STEP_LABELS: Record<string, string> = {
  plan: 'Planning',
  expert_started: 'Expert started',
  retrieving_context: 'Retrieving context',
  china_wall_check: 'China Wall check',
  generating_answer: 'Generating answer',
  expert_completed: 'Expert finished',
  expert_failed: 'Expert failed',
  checking_consistency: 'Checking consistency',
  consistency_checked: 'Consistency checked',
  awaiting_human_review: 'Awaiting your review',
  human_approved: 'Approved',
  relay_failed: 'Relay stopped',
  relay_completed: 'Complete',
}

export function RelayTransparencyPanel({ steps }: { steps: RelayStepEvent[] }) {
  if (steps.length === 0) return null
  return (
    <div className="mb-2 rounded-md border border-slate-700 bg-slate-900/60 p-3">
      <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-400">
        Live pipeline
      </div>
      <ol className="space-y-1">
        {steps.map((s, i) => (
          <li key={i} className="flex items-start gap-2 text-xs">
            <span
              className={cn(
                'mt-0.5 inline-block h-2 w-2 shrink-0 rounded-full',
                s.step === 'expert_failed' || s.step === 'relay_failed'
                  ? 'bg-red-500'
                  : s.step === 'awaiting_human_review'
                    ? 'bg-amber-400'
                    : 'bg-emerald-500',
              )}
            />
            <span className="text-slate-300">
              {STEP_LABELS[s.step] ?? s.step}
              {s.expertName ? ` · ${s.expertName}` : ''}
            </span>
            {s.message && <span className="text-slate-500">— {s.message}</span>}
          </li>
        ))}
      </ol>
    </div>
  )
}
