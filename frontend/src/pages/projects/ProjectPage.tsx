import type { LoaderFunctionArgs } from 'react-router-dom'
import { useLoaderData, Outlet } from 'react-router-dom'
import { getProject } from '@/api/projects'
import type { Project } from '@/types/project'
import { RepoStatus } from '@/components/project/RepoStatus'

async function loader({ params, request }: LoaderFunctionArgs) {
  const projectId = params.projectId as string
  const project = await getProject(projectId, { signal: request.signal })
  return { project }
}

export const projectRoute = { element: <ProjectPage />, loader }

/**
 * WHY ProjectTimeline is NOT wired in here yet: it needs L3Event[]
 * from a separate endpoint (GET /projects/:id/timeline per
 * AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 15), which this loader
 * does not currently fetch. Adding a second parallel fetch here is
 * straightforward, but deferred to avoid silently mixing an
 * unspecified pagination contract (the endpoint is documented as
 * "paginated" in HANDOFF.md with no page-size/cursor shape given) into
 * this loader without first confirming that contract - tracked as a
 * named follow-up in HANDOFF.md rather than guessed at here.
 */
export default function ProjectPage() {
  const { project } = useLoaderData() as { project: Project }

  return (
    <div className="p-6">
      <h1 className="text-xl font-semibold">{project.name}</h1>
      <p className="mt-1 text-sm text-text-secondary">{project.description}</p>

      <div className="mt-6">
        <RepoStatus project={project} />
      </div>

      {/* Chat list, timeline - chat list is Phase 3 (needs getChats wiring + UI); timeline deferred per comment above */}
      <Outlet />
    </div>
  )
}
