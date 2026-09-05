import type { Expert } from '@/types/expert'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'

/**
 * Expert selection/discovery card.
 * Source: FRONTEND_SYSTEM_DESIGN.md section 10 ("Expert Discovery
 * Page" wireframe - name, domain, depth, topics, rating, chunk count,
 * [View Topics] [Add to Project] actions).
 */
export interface ExpertCardProps {
  expert: Expert
  onViewTopics?: (expertId: string) => void
  onAddToProject?: (expertId: string) => void
  /** Hides the "Add to Project" action when this expert is already in the current project. */
  isAdded?: boolean
}

function depthLabel(avgDepthLevel: number): string {
  // Matches the wireframe's "Depth: Expert (4/5)" / "Depth: Advanced (3/5)" format.
  if (avgDepthLevel >= 4.5) return 'Master'
  if (avgDepthLevel >= 3.5) return 'Expert'
  if (avgDepthLevel >= 2.5) return 'Advanced'
  if (avgDepthLevel >= 1.5) return 'Intermediate'
  return 'Basic'
}

function ratingStars(avgRating: number): string {
  // WHY round to nearest half-star via Math.round(x*2)/2 rather than
  // floor: an expert with 4.8 should visually read as \u2b50x5, not
  // \u2b50x4 - flooring would understate a genuinely excellent expert.
  const rounded = Math.round(avgRating * 2) / 2
  const full = Math.floor(rounded)
  const half = rounded - full === 0.5
  return '\u2b50'.repeat(full) + (half ? '\u00bd' : '')
}

export function ExpertCard({ expert, onViewTopics, onAddToProject, isAdded }: ExpertCardProps) {
  return (
    <Card>
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="font-medium text-text-primary">{expert.name}</p>
          <p className="mt-1 text-sm text-text-secondary">
            Domain: {expert.domain} | Depth: {depthLabel(expert.avgDepthLevel)} (
            {expert.avgDepthLevel.toFixed(1)}/5)
          </p>
        </div>
        <Badge variant="brand">{expert.totalTopics} topics</Badge>
      </div>

      <p className="mt-3 text-sm text-text-secondary">
        Rating: {ratingStars(expert.avgRating)} ({expert.avgRating.toFixed(1)}) |{' '}
        {expert.totalChunks.toLocaleString()} chunks
      </p>

      <div className="mt-4 flex gap-2">
        <Button variant="secondary" size="sm" onClick={() => onViewTopics?.(expert.id)}>
          View Topics
        </Button>
        {!isAdded && (
          <Button variant="primary" size="sm" onClick={() => onAddToProject?.(expert.id)}>
            Add to Project
          </Button>
        )}
        {isAdded && <Badge variant="neutral">Already in project</Badge>}
      </div>
    </Card>
  )
}
