import type { ProjectExpert } from '@/types/project'
import { cn } from '@/utils/cn'

/**
 * Compact "active expert" indicator chip.
 * Source: FRONTEND_SYSTEM_DESIGN.md section 1 main-interface wireframe
 * ("[\ud83d\udd35 SD Expert] [\ud83d\udfe2 DB Expert] [\ud83d\udfe1 Backend]")
 * and section 10 Chat Page ("Active: [\ud83d\udd35 SD \u2713] [\ud83d\udfe2 DB \u2713] [Backend]").
 *
 * WHY the color dot uses a rotating palette keyed by expertId's hash,
 * not a fixed per-domain color: the design doc's wireframes show
 * different colored dots per expert (blue/green/yellow) but never
 * specifies WHICH color maps to WHICH domain - there is no rule given.
 * Rather than inventing an arbitrary domain->color mapping that the
 * backend has no concept of (domain is a free-text string, not an
 * enum), this derives a stable-but-arbitrary color from the expert's
 * own id so the same expert always gets the same dot color across
 * renders, without pretending there's a real semantic mapping that
 * doesn't exist in the spec.
 */
const DOT_COLORS = ['bg-mode-ask', 'bg-mode-advise', 'bg-mode-warn', 'bg-mode-pushback', 'bg-brand']

function colorForId(id: string): string {
  let hash = 0
  for (let i = 0; i < id.length; i++) {
    hash = (hash * 31 + id.charCodeAt(i)) | 0
  }
  return DOT_COLORS[Math.abs(hash) % DOT_COLORS.length]!
}

export interface ExpertBadgeProps {
  expert: Pick<ProjectExpert, 'expertId' | 'expertName'>
  isSelected?: boolean
  onClick?: () => void
}

export function ExpertBadge({ expert, isSelected, onClick }: ExpertBadgeProps) {
  const dotColor = colorForId(expert.expertId)

  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={isSelected}
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium transition-colors',
        isSelected
          ? 'border-brand bg-brand/10 text-text-primary'
          : 'border-surface-border bg-surface-overlay text-text-secondary hover:text-text-primary'
      )}
    >
      <span className={cn('h-2 w-2 rounded-full', dotColor)} aria-hidden="true" />
      {expert.expertName}
      {isSelected && <span aria-hidden="true">\u2713</span>}
    </button>
  )
}
