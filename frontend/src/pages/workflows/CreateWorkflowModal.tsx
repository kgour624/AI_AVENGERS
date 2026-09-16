import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { getProjects } from '@/api/projects'
import { getExperts } from '@/api/experts'
import { createWorkflow, startWorkflow, runWorkflow } from '@/api/workflows'
import { queryKeys } from '@/api/queryKeys'
import { cn } from '@/utils/cn'

interface Props {
  isOpen: boolean
  onClose: () => void
}

export function CreateWorkflowModal({ isOpen, onClose }: Props) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [title, setTitle] = useState('')
  const [projectId, setProjectId] = useState('')
  const [selectedExpertIds, setSelectedExpertIds] = useState<string[]>([])
  const [budget, setBudget] = useState('10')
  const [error, setError] = useState<string | null>(null)

  const { data: projects = [] } = useQuery({
    queryKey: queryKeys.projects.all,
    queryFn: getProjects,
    enabled: isOpen,
  })

  const { data: experts = [] } = useQuery({
    queryKey: queryKeys.experts.all,
    queryFn: getExperts,
    enabled: isOpen,
  })

  const createMut = useMutation({
    mutationFn: async () => {
      if (!title.trim()) throw new Error('Title is required')
      if (!projectId) throw new Error('Select a project')
      if (selectedExpertIds.length === 0) throw new Error('Select at least one expert')
      const wf = await createWorkflow({
        projectId,
        title: title.trim(),
        selectedExpertIds,
        costBudgetUsd: parseFloat(budget) || 10,
      })
      await startWorkflow(wf.id)
      await runWorkflow(wf.id)
      return wf
    },
    onSuccess: (wf) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workflows.all })
      onClose()
      navigate(`/workflows/${wf.id}/kanban`)
    },
    onError: (err: unknown) => {
      setError(err instanceof Error ? err.message : 'Failed to create workflow')
    },
  })

  const toggleExpert = (id: string) =>
    setSelectedExpertIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    )

  const handleClose = () => {
    setTitle('')
    setProjectId('')
    setSelectedExpertIds([])
    setBudget('10')
    setError(null)
    onClose()
  }

  return (
    <Modal isOpen={isOpen} onClose={handleClose}>
      <div className="flex flex-col gap-4" style={{ minWidth: 440, maxWidth: 520 }}>
        <h2 className="text-base font-semibold text-text-primary">New Workflow</h2>

        <div>
          <label className="mb-1 block text-xs font-medium text-text-secondary">Title</label>
          <input
            autoFocus
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="e.g. Build URL shortener"
            className="w-full rounded-md border border-glass-border bg-surface-overlay px-3 py-2 text-sm text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-brand/40"
          />
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-text-secondary">Project</label>
          <select
            value={projectId}
            onChange={(e) => setProjectId(e.target.value)}
            className="w-full rounded-md border border-glass-border bg-surface-overlay px-3 py-2 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-brand/40"
          >
            <option value="">Select a project...</option>
            {projects.filter((p) => p.status !== 'archived').map((p) => (
              <option key={p.id} value={p.id}>{p.name}</option>
            ))}
          </select>
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-text-secondary">
            Domain Experts
            <span className="ml-1 text-text-disabled">({selectedExpertIds.length} selected)</span>
          </label>
          <div className="max-h-48 overflow-y-auto rounded-md border border-glass-border bg-surface-overlay p-2">
            {experts.length === 0 && (
              <p className="py-2 text-center text-xs text-text-disabled">No trained experts available</p>
            )}
            {experts.map((expert) => {
              const selected = selectedExpertIds.includes(expert.id)
              return (
                <button
                  key={expert.id}
                  type="button"
                  onClick={() => toggleExpert(expert.id)}
                  className={cn(
                    'mb-1 flex w-full items-center gap-2 rounded-md px-2.5 py-2 text-left text-xs transition-all',
                    selected
                      ? 'border border-brand/40 bg-brand/10 text-text-primary'
                      : 'border border-transparent text-text-secondary hover:bg-surface-raised hover:text-text-primary'
                  )}
                >
                  <span className={cn(
                    'flex h-3.5 w-3.5 flex-shrink-0 items-center justify-center rounded border text-[9px]',
                    selected ? 'border-brand bg-brand text-white' : 'border-surface-border'
                  )}>
                    {selected && '\u2713'}
                  </span>
                  <span className="font-medium">{expert.name}</span>
                  <span className="ml-auto text-text-disabled">{expert.domain}</span>
                </button>
              )
            })}
          </div>
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-text-secondary">Cost Budget (USD)</label>
          <input
            type="number" min="1" step="1"
            value={budget}
            onChange={(e) => setBudget(e.target.value)}
            className="w-32 rounded-md border border-glass-border bg-surface-overlay px-3 py-2 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-brand/40"
          />
        </div>

        {error && (
          <p className="rounded-md border border-mode-refuse/30 bg-mode-refuse/10 px-3 py-2 text-xs text-mode-refuse">
            {error}
          </p>
        )}

        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={handleClose} disabled={createMut.isPending}>Cancel</Button>
          <Button
            onClick={() => createMut.mutate()}
            disabled={createMut.isPending || !title.trim() || !projectId || selectedExpertIds.length === 0}
          >
            {createMut.isPending ? 'Launching...' : 'Launch Workflow'}
          </Button>
        </div>
      </div>
    </Modal>
  )
}
