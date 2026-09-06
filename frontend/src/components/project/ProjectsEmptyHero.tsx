import { useState } from 'react'
import { motion, useReducedMotion } from 'framer-motion'
import { createProject, type CreateProjectRequest } from '@/api/projects'
import { Button } from '@/components/ui/Button'
import { CreateProjectModal } from '@/components/project/CreateProjectModal'
import { handleAPIError } from '@/utils/errors'
import { ARC_MOTION } from '@/design-system/motion'

export interface ProjectsEmptyHeroProps {
  onCreated: (projectId: string) => void
}

/**
 * ARC-51 §7 (docs/ARC51_UI_CONTRACT.md): replaces the dead
 * "No projects yet." text with a HUD-style empty state.
 *
 * NON-NEGOTIABLE per §1 of the contract: the quick-start templates
 * below are NOT decorative buttons - each one calls the real
 * createProject() API (same function CreateProjectModal uses) with a
 * pre-filled name/description, then hands the new project's real ID
 * to onCreated exactly like the modal does. A user clicking a
 * template gets a real project in their real project list, not a
 * visual mockup.
 */
const TEMPLATES: CreateProjectRequest[] = [
  {
    name: 'E-Commerce Microservices',
    description: 'Design a scalable e-commerce platform with microservices architecture.',
  },
  {
    name: 'Real-time Chat Engine',
    description: 'Build a real-time messaging system with WebSocket/SSE delivery.',
  },
  {
    name: 'High-Scale Database',
    description: 'Design a database schema and sharding strategy for high-throughput workloads.',
  },
]

export function ProjectsEmptyHero({ onCreated }: ProjectsEmptyHeroProps) {
  const reduceMotion = useReducedMotion()
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [pendingTemplate, setPendingTemplate] = useState<string | null>(null)
  const [error, setError] = useState('')

  const handleTemplateClick = async (template: CreateProjectRequest) => {
    setPendingTemplate(template.name)
    setError('')
    try {
      const project = await createProject(template)
      onCreated(project.id)
    } catch (err) {
      setError(handleAPIError(err))
      setPendingTemplate(null)
    }
  }

  return (
    <motion.div
      initial={reduceMotion ? undefined : { opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: ARC_MOTION.panel, ease: ARC_MOTION.ease }}
      className="flex flex-col items-center py-16 text-center"
    >
      {/* Glowing AI Core - CSS conic-gradient rotation, reuses the same
          arc-ring-rotate keyframe as ExpertAvatar's analyzing ring and
          the Login page's ambient glows, per §3's CSS-for-continuous-
          decorative-loops boundary. */}
      <div className="relative mb-8 h-28 w-28" aria-hidden="true">
        <div
          className="absolute inset-0 rounded-full bg-[conic-gradient(var(--glow-purple),var(--glow-cyan),var(--glow-purple))] opacity-70 blur-md"
          style={{ animation: 'arc-ring-rotate 6s linear infinite' }}
        />
        <div className="absolute inset-3 rounded-full bg-surface-void" />
        <div className="absolute inset-0 flex items-center justify-center text-3xl">{'\u26A1'}</div>
      </div>

      <h2 className="max-w-md text-2xl font-semibold text-text-primary [text-shadow:0_0_20px_var(--glow-purple)]">
        Deploy Your Multi-Agent Architecture Workspace
      </h2>
      <p className="mt-3 max-w-md text-sm text-text-secondary">
        Assemble your team of specialized AI Domain Experts to design, review, and simulate your
        systems.
      </p>

      <Button size="lg" className="mt-6" onClick={() => setIsCreateOpen(true)}>
        {'\u26A1'} Initialize First Project
      </Button>

      <div className="mt-10 w-full max-w-2xl">
        <p className="mb-3 text-xs uppercase tracking-wide text-text-disabled">
          Or start from a blueprint
        </p>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          {TEMPLATES.map((template) => (
            <button
              key={template.name}
              type="button"
              disabled={pendingTemplate !== null}
              onClick={() => handleTemplateClick(template)}
              className="rounded-lg border border-glass-border bg-surface-raised/70 p-4 text-left text-sm text-text-secondary backdrop-blur-xl transition-[border-color,box-shadow] duration-150 ease-arc hover:border-glow-cyan/50 hover:shadow-[0_0_24px_-8px_var(--glow-cyan)] hover:text-text-primary disabled:cursor-not-allowed disabled:opacity-50"
            >
              <p className="font-medium text-text-primary">
                {pendingTemplate === template.name ? 'Deploying...' : template.name}
              </p>
              <p className="mt-1 text-xs text-text-disabled">{template.description}</p>
            </button>
          ))}
        </div>
        {error && <p className="mt-3 text-sm text-mode-refuse">{error}</p>}
      </div>

      <CreateProjectModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        onCreated={onCreated}
      />
    </motion.div>
  )
}
