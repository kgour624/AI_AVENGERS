import type { LoaderFunctionArgs } from 'react-router-dom'
import { useLoaderData, Outlet } from 'react-router-dom'
import { getProject } from '@/api/projects'
import type { Project } from '@/types/project'
import { RepoStatus } from '@/components/project/RepoStatus'
import { ChatList } from '@/components/project/ChatList'
import { ProjectExpertManager } from '@/components/project/ProjectExpertManager'

async function loader({ params, request }: LoaderFunctionArgs) {
  const projectId = params.projectId as string
  const project = await getProject(projectId, { signal: request.signal })
  return { project }
}

export const projectRoute = { element: <ProjectPage />, loader }

/**
 * WHY ProjectTimeline is STILL not wired in here (carried over from
 * Phase 2/3, now confirmed rather than just suspected): read the full
 * backend-go/internal/project/service.go source in Phase 4's
 * correction pass - there is no timeline/L3-event handler or route in
 * that file at all, not even an unconfirmed-pagination one. This
 * isn't "the contract is unclear", it's "the endpoint does not exist
 * yet" - wiring it up would mean calling a URL that 404s. Logged as a
 * concrete backend TODO in HANDOFF.md, not a frontend gap.
 */
export default function ProjectPage() {
  const { project } = useLoaderData() as { project: Project }

  return (
    <div className="p-6">
      <h1 className="text-xl font-semibold">{project.name}</h1>
      <p className="mt-1 text-sm text-text-secondary">{project.description}</p>

      <div className="mt-4">
        <p className="mb-1 text-sm font-medium text-text-secondary">Experts in this project</p>
        <ProjectExpertManager projectId={project.id} experts={project.experts} />
      </div>

      <div className="mt-6">
        <ChatList projectId={project.id} />
      </div>

      <div className="mt-6">
        <RepoStatus project={project} />
      </div>

      <Outlet />
    </div>
  )
}
