import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { addProjectExpert, removeProjectExpert } from '@/api/projects'
import { getExperts } from '@/api/experts'
import { useQuery } from '@tanstack/react-query'
import type { ProjectExpert } from '@/types/project'
import { ExpertBadge } from '@/components/expert/ExpertBadge'
import { Button } from '@/components/ui/Button'
import { Modal } from '@/components/ui/Modal'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 10 ("Experts in this
 * project" panel: badges + [+ Add Expert] button).
 *
 * Cross-questioned: after addProjectExpert succeeds, what does the UI
 * show? Per the correction commit earlier this phase, AddExpert's real
 * response is just {status}, no expert data - so this component
 * cannot optimistically render the newly-added expert from the
 * mutation's own response. Instead it invalidates the project detail
 * query (which includes .experts) so the badge list refetches and
 * shows the real, server-confirmed state - slightly slower to update
 * than an optimistic UI, but correct rather than guessing at what the
 * new ProjectExpert row would contain (addedAt, isActive, etc. that
 * this component has no way to know client-side).
 */
export function ProjectExpertManager({ projectId, experts }: { projectId: string; experts: ProjectExpert[] }) {
  const queryClient = useQueryClient()
  const [isModalOpen, setIsModalOpen] = useState(false)

  const { data: allExperts } = useQuery({
    queryKey: ['experts'],
    queryFn: () => getExperts(),
    enabled: isModalOpen, // WHY only fetch when the modal opens: no reason to load the full expert catalog on every project page view if the user never opens the add-expert modal.
  })

  function invalidateProject() {
    queryClient.invalidateQueries({ queryKey: ['projects', projectId] })
  }

  const addMutation = useMutation({
    mutationFn: (expertId: string) => addProjectExpert(projectId, expertId),
    onSuccess: invalidateProject,
  })

  const removeMutation = useMutation({
    mutationFn: (expertId: string) => removeProjectExpert(projectId, expertId),
    onSuccess: invalidateProject,
  })

  const addedIds = new Set(experts.map((e) => e.expertId))

  return (
    <div>
      <div className="flex flex-wrap items-center gap-2">
        {experts.map((e) => (
          <div key={e.expertId} className="group relative">
            <ExpertBadge expert={e} />
            <button
              type="button"
              onClick={() => removeMutation.mutate(e.expertId)}
              aria-label={`Remove ${e.expertName}`}
              className="absolute -right-1 -top-1 hidden h-4 w-4 rounded-full bg-mode-refuse text-[10px] text-white group-hover:flex group-hover:items-center group-hover:justify-center"
            >
              \u00d7
            </button>
          </div>
        ))}
        <Button variant="ghost" size="sm" onClick={() => setIsModalOpen(true)}>
          + Add Expert
        </Button>
      </div>

      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)}>
        <h3 className="mb-4 text-sm font-medium text-text-primary">Add Expert to Project</h3>
        <div className="space-y-2">
          {allExperts
            ?.filter((e) => !addedIds.has(e.id))
            .map((e) => (
              <div key={e.id} className="flex items-center justify-between rounded-md border border-surface-border p-2">
                <div>
                  <p className="text-sm text-text-primary">{e.name}</p>
                  <p className="text-xs text-text-secondary">{e.domain}</p>
                </div>
                <Button
                  size="sm"
                  isLoading={addMutation.isPending && addMutation.variables === e.id}
                  onClick={() => addMutation.mutate(e.id)}
                >
                  Add
                </Button>
              </div>
            ))}
        </div>
      </Modal>
    </div>
  )
}
