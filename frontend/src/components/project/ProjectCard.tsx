import { Link } from 'react-router-dom'
import type { Project } from '@/types/project'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { formatRelativeTime } from '@/utils/format'

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
  return (
    <Card glow="purple">
      <p className="font-medium text-text-primary">{project.name}</p>
      <p className="mt-1 text-sm text-text-secondary">Experts: {project.experts.length}</p>
      <p className="mt-1 text-sm text-text-secondary">
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
