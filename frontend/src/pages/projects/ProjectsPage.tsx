import { useMemo, useState } from 'react'
import type { LoaderFunctionArgs } from 'react-router-dom'
import { useLoaderData, useRevalidator, useNavigate } from 'react-router-dom'
import { motion, useReducedMotion } from 'framer-motion'
import { getProjects } from '@/api/projects'
import type { Project } from '@/types/project'
import { ProjectCard } from '@/components/project/ProjectCard'
import { CreateProjectModal } from '@/components/project/CreateProjectModal'
import { ProjectsEmptyHero } from '@/components/project/ProjectsEmptyHero'
import { Button } from '@/components/ui/Button'
import { fadeUp, staggerContainer, ARC_MOTION } from '@/design-system/motion'

/**
 * Co-located loader pattern. Source: FRONTEND_SYSTEM_DESIGN.md section 5
 * ("Route only defines path. Component owns its data requirements.").
 */
async function loader({ request }: LoaderFunctionArgs) {
  // Defensive fallback: getProjects() resolves via a TS non-null
  // assertion that the backend's data field is never undefined/null -
  // that is a compile-time promise only, not a runtime guarantee.
  // The component calls .length/.map on this directly, so it must
  // always receive a real array here.
  const projects = (await getProjects({ signal: request.signal })) ?? []
  return { projects }
}

export const projectsRoute = { element: <ProjectsPage />, loader }

export default function ProjectsPage() {
  const { projects } = useLoaderData() as { projects: Project[] }
  const revalidator = useRevalidator()
  const navigate = useNavigate()
  const reduceMotion = useReducedMotion()
  const [isCreateOpen, setIsCreateOpen] = useState(false)

  // ARC-51 S14 telemetry HUD: only real, derivable metrics - see
  // commit message for what was deliberately NOT fabricated here.
  const activeExpertsCount = useMemo(
    () => new Set(projects.flatMap((p) => (p.experts ?? []).map((e) => e.expertId))).size,
    [projects]
  )
  const connectedRepoCount = useMemo(() => projects.filter((p) => p.repoConnected).length, [projects])

  // Shared by both the header's "+ New Project" button/modal AND the
  // empty-state hero's CTA/templates - same real revalidate+navigate
  // flow either way, no duplicated logic between the two entry points.
  const handleCreated = (projectId: string) => {
    // This page sources `projects` from a loader (useLoaderData), not a
    // useQuery - revalidate() is React Router's real mechanism to
    // re-run it. Matches the pattern this codebase already fixed once
    // before for exactly this loader-vs-query mistake
    // (ProjectExpertManager/RepoConnectModal, Phase 6).
    revalidator.revalidate()
    navigate(`/projects/${projectId}`)
  }

  return (
    <div className="flex-1 overflow-y-auto p-6">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-xl font-semibold">Your Projects</h1>
        {/* Bug 1.1 fix (docs bug list): no Create Project entry point
            existed anywhere in the UI - see CreateProjectModal.tsx.
            Kept even in the non-empty state (hero only shows when
            projects.length === 0) so returning users have a fast path
            without scrolling. */}
        {projects.length > 0 && (
          <Button onClick={() => setIsCreateOpen(true)}>+ New Project</Button>
        )}
      </div>

      {projects.length > 0 && (
        <motion.div
          initial={reduceMotion ? undefined : { opacity: 0, y: -8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: ARC_MOTION.panel, ease: ARC_MOTION.ease }}
          className="mb-6 flex gap-6 rounded-lg border border-glass-border bg-surface-raised/50 px-4 py-3 backdrop-blur-xl"
        >
          <div>
            <p className="text-xs uppercase tracking-wide text-text-disabled">Active Experts</p>
            <p className="text-lg font-semibold text-glow-cyan">{activeExpertsCount}</p>
          </div>
          <div className="w-px bg-glass-border" />
          <div>
            <p className="text-xs uppercase tracking-wide text-text-disabled">Repository Links</p>
            <p className="text-lg font-semibold text-glow-purple">{connectedRepoCount}</p>
          </div>
        </motion.div>
      )}

      {projects.length === 0 ? (
        <ProjectsEmptyHero onCreated={handleCreated} />
      ) : (
        <motion.div
          initial={reduceMotion ? undefined : 'hidden'}
          animate="visible"
          variants={staggerContainer}
          className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3"
        >
          {projects.map((p, i) => (
            <motion.div key={p.id} variants={i < ARC_MOTION.maxStaggerItems ? fadeUp : undefined}>
              <ProjectCard project={p} />
            </motion.div>
          ))}
        </motion.div>
      )}

      <CreateProjectModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        onCreated={handleCreated}
      />
    </div>
  )
}
