import { useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useRevalidator } from 'react-router-dom'
import { addProjectExpert, removeProjectExpert } from '@/api/projects'
import { getExperts } from '@/api/experts'
import type { ProjectExpert } from '@/types/project'
import { ExpertBadge } from '@/components/expert/ExpertBadge'
import { Button } from '@/components/ui/Button'
import { Modal } from '@/components/ui/Modal'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 10 ("Experts in this
 * project" panel: badges + [+ Add Expert] button).
 *
 * PHASE 6 BUG FIX (found during final QA audit of cache invalidation
 * consistency across the codebase): this component's onSuccess used
 * to call queryClient.invalidateQueries({queryKey: ['projects', projectId]}),
 * expecting that to refresh the `experts` prop after add/remove. That
 * was a SILENT NO-OP - ProjectPage.tsx gets `project` (and therefore
 * `experts`) from a React Router LOADER via useLoaderData(), not
 * from any useQuery call. There is no TanStack Query cache entry
 * keyed ['projects', projectId] for anything to invalidate - nothing
 * in the app is subscribed to that key for project detail data. The
 * badge list would never actually update after adding or removing an
 * expert without a full page reload, despite the mutation succeeding
 * and no error being shown - the worst kind of bug, silently wrong
 * rather than loudly broken.
 *
 * Fix: use useRevalidator() (React Router's mechanism for re-running
 * the current route's loader) instead, matching the exact pattern
 * ChatPage.tsx already uses correctly for the same class of problem
 * (SSE completion needing to refresh loader-sourced data).
 */
export function ProjectExpertManager({ projectId, experts }: { projectId: string; experts: ProjectExpert[] }) {
  const revalidator = useRevalidator()
  const [isModalOpen, setIsModalOpen] = useState(false)

  // Defensive default: GetByID (backend-go/internal/project/service.go)
  // does `experts, _ := s.GetExperts(...)` - discarding the query
  // error means a transient DB error leaves Experts nil, which the
  // Project struct's `experts,omitempty` json tag then omits from the
  // wire entirely. Must not crash the add/remove-expert UI over it.
  const safeExperts = experts ?? []

  const { data: allExperts } = useQuery({
    queryKey: ['experts'],
    queryFn: () => getExperts(),
    enabled: isModalOpen, // WHY only fetch when the modal opens: no reason to load the full expert catalog on every project page view if the user never opens the add-expert modal.
  })

  const addMutation = useMutation({
    mutationFn: (expertId: string) => addProjectExpert(projectId, expertId),
    onSuccess: () => revalidator.revalidate(),
  })

  const removeMutation = useMutation({
    mutationFn: (expertId: string) => removeProjectExpert(projectId, expertId),
    onSuccess: () => revalidator.revalidate(),
  })

  const addedIds = new Set(safeExperts.map((e) => e.expertId))

  return (
    <div>
      <div className="flex flex-wrap items-center gap-2">
        {safeExperts.map((e) => (
          <div key={e.expertId} className="group relative">
            <ExpertBadge expert={e} />
            <button
              type="button"
              onClick={() => removeMutation.mutate(e.expertId)}
              aria-label={`Remove ${e.expertName}`}
              className="absolute -right-1 -top-1 hidden h-4 w-4 rounded-full bg-mode-refuse text-[10px] text-white group-hover:flex group-hover:items-center group-hover:justify-center"
            >
              {'×'}
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
