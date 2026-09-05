import type { LoaderFunctionArgs } from 'react-router-dom'
import { useLoaderData } from 'react-router-dom'
import { getProjects } from '@/api/projects'
import type { Project } from '@/types/project'
import { ProjectCard } from '@/components/project/ProjectCard'

/**
 * Co-located loader pattern. Source: FRONTEND_SYSTEM_DESIGN.md section 5
 * ("Route only defines path. Component owns its data requirements.").
 */
async function loader({ request }: LoaderFunctionArgs) {
  const projects = await getProjects({ signal: request.signal })
  return { projects }
}

export const projectsRoute = { element: <ProjectsPage />, loader }

export default function ProjectsPage() {
  const { projects } = useLoaderData() as { projects: Project[] }

  return (
    <div className="p-6">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-xl font-semibold">Your Projects</h1>
      </div>
      {projects.length === 0 ? (
        <p className="text-text-secondary">No projects yet.</p>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {projects.map((p) => (
            <ProjectCard key={p.id} project={p} />
          ))}
        </div>
      )}
    </div>
  )
}
