import type { LoaderFunctionArgs } from 'react-router-dom'
import { useLoaderData, Outlet } from 'react-router-dom'
import { getProject } from '@/api/projects'
import { getProjectTimeline } from '@/api/memory'
import type { Project } from '@/types/project'
import type { L3Event } from '@/types/memory'
import { RepoStatus } from '@/components/project/RepoStatus'
import { ChatList } from '@/components/project/ChatList'
import { ProjectExpertManager } from '@/components/project/ProjectExpertManager'
import { ProjectTimeline } from '@/components/project/ProjectTimeline'

/**
 * PHASE 6 CORRECTION: ProjectTimeline is now wired in. This was
 * blocked in Phases 2-4 by an incorrect finding (documented and
 * corrected in HANDOFF.md's consolidated blocker list) that the
 * timeline endpoint didn't exist at all - re-checking main.go directly
 * showed it IS registered and reachable, just with hardcoded
 * limit=50/offset=0 and no pagination params. Fetched in parallel
 * with the project itself, consistent with the "parallel loader
 * fetch" pattern used everywhere else in this codebase (ChatPage,
 * ExpertsPage) - Promise.all rather than a sequential await, since
 * neither fetch depends on the other's result.
 */
async function loader({ params, request }: LoaderFunctionArgs) {
  const projectId = params.projectId as string
  const signal = request.signal

  const [project, timeline] = await Promise.all([
    getProject(projectId, { signal }),
    getProjectTimeline(projectId, { signal }),
  ])

  return { project, timeline }
}

export const projectRoute = { element: <ProjectPage />, loader }

export default function ProjectPage() {
  const { project, timeline } = useLoaderData() as { project: Project; timeline: L3Event[] }

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
        <p className="mb-1 text-sm font-medium text-text-secondary">Recent Activity</p>
        <ProjectTimeline events={timeline} />
      </div>

      <div className="mt-6">
        <RepoStatus projectId={project.id} />
      </div>

      <Outlet />
    </div>
  )
}
