import { Link } from 'react-router-dom'
import type { Project } from '@/types/project'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { ExpertAvatar } from '@/components/expert/ExpertAvatar'
import { formatRelativeTime } from '@/utils/format'

const MAX_AVATAR_STACK = 4

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 10 ("Projects Page (Home)"
 * wireframe - name, Experts: N, Chats: N, Last: Xh ago, [Open Project]).
 *
 * Gap found: Project (types/project.ts) has no `chatCount` field and
 * no `lastActivityAt` field - only `experts: ProjectExpert[]` (whose
 * .length gives the expert count) and updatedAt. The wireframe's
 * "Chats: 12" and "Last: 2h ago" have no backing field anywhere in
 * either design doc's data model (section 5's `projects` table has no
 * chat count column either - chats are a separate table with no
 * denormalized count on projects). Rather than fabricate a chatCount
 * prop with no real data source, this card omits that line and uses
 * `updatedAt` for "Last active" (the closest real field available).
 * Flagged in HANDOFF.md as a backend gap: either add a chat_count to
 * the project list response, or drop that line from the spec.
 */
export function ProjectCard({ project }: { project: Project }) {
  // Defensive default preserved from the earlier crash fix - now
  // backed by real data end-to-end (service.go's List() batch-loads
  // experts instead of never populating them).
  const experts = project.experts ?? []

  return (
    <Card glow="purple">
      <div className="flex items-start justify-between gap-2">
        <p className="font-medium text-text-primary">{project.name}</p>
        {/* Live pulse dot - decorative "active" cue, Tailwind core
            animate-ping (no config changes needed). */}
        <span className="relative mt-1.5 flex h-2 w-2 flex-shrink-0" aria-hidden="true">
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-glow-purple opacity-60" />
          <span className="relative inline-flex h-2 w-2 rounded-full bg-glow-purple" />
        </span>
      </div>

      {/* Architecture Type badge - only rendered when the project
          actually has one set. Never fabricated: architectureType
          comes from a real (now-selected) DB column, absent for any
          project that never had it set via EditProjectModal. */}
      {project.architectureType && (
        <span className="mt-1 inline-block rounded-full border border-glass-border bg-surface-overlay px-2 py-0.5 text-xs text-glow-cyan">
          {'\u26a1'} {project.architectureType}
        </span>
      )}

      <p className="mt-2 text-sm text-text-secondary">Experts: {experts.length}</p>

      {experts.length > 0 && (
        <div className="mt-1 flex -space-x-2" aria-hidden="true">
          {experts.slice(0, MAX_AVATAR_STACK).map((e) => (
            <ExpertAvatar
              key={e.expertId}
              domain={e.domain}
              status="idle"
              size="sm"
              className="rounded-full ring-2 ring-surface-raised"
            />
          ))}
          {experts.length > MAX_AVATAR_STACK && (
            <span className="flex h-8 w-8 items-center justify-center rounded-full bg-surface-overlay text-xs text-text-secondary ring-2 ring-surface-raised">
              +{experts.length - MAX_AVATAR_STACK}
            </span>
          )}
        </div>
      )}

      <p className="mt-2 text-sm text-text-secondary">
        Last active: {formatRelativeTime(project.updatedAt)}
      </p>
      <Link to={`/projects/${project.id}`} className="mt-4 inline-block">
        <Button variant="secondary" size="sm">
          Open Project
        </Button>
      </Link>
    </Card>
  )
}
