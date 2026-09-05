import type { LoaderFunctionArgs } from 'react-router-dom'
import { useLoaderData, Outlet } from 'react-router-dom'
import { getProject } from '@/api/projects'
import type { Project } from '@/types/project'

async function loader({ params, request }: LoaderFunctionArgs) {
  const projectId = params.projectId as string
  const project = await getProject(projectId, { signal: request.signal })
  return { project }
}

export const projectRoute = { element: <ProjectPage />, loader }

export default function ProjectPage() {
  const { project } = useLoaderData() as { project: Project }

  return (
    <div className="p-6">
      <h1 className="text-xl font-semibold">{project.name}</h1>
      <p className="mt-1 text-sm text-text-secondary">{project.description}</p>
      {/* Chat list, timeline, repo status - Phase 2 */}
      <Outlet />
    </div>
  )
}
