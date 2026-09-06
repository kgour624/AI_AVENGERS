import { useState } from 'react'
import type { LoaderFunctionArgs } from 'react-router-dom'
import { useLoaderData, useRevalidator, useNavigate } from 'react-router-dom'
import { getProjects } from '@/api/projects'
import type { Project } from '@/types/project'
import { ProjectCard } from '@/components/project/ProjectCard'
import { CreateProjectModal } from '@/components/project/CreateProjectModal'
import { Button } from '@/components/ui/Button'

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
  const revalidator = useRevalidator()
  const navigate = useNavigate()
  const [isCreateOpen, setIsCreateOpen] = useState(false)

  return (
    <div className="p-6">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-xl font-semibold">Your Projects</h1>
        {/* Bug 1.1 fix (docs bug list): no Create Project entry point
            existed anywhere in the UI - see CreateProjectModal.tsx. */}
        <Button onClick={() => setIsCreateOpen(true)}>+ New Project</Button>
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

      <CreateProjectModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        onCreated={(projectId) => {
          // This page sources `projects` from a loader (useLoaderData),
          // not a useQuery - revalidate() is React Router's real
          // mechanism to re-run it. Matches the pattern this codebase
          // already fixed once before for exactly this loader-vs-query
          // mistake (ProjectExpertManager/RepoConnectModal, Phase 6).
          revalidator.revalidate()
          navigate(`/projects/${projectId}`)
        }}
      />
    </div>
  )
}
