import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { getProjects } from '@/api/projects'
import { getExperts } from '@/api/experts'
import { createWorkflow, startWorkflow, runWorkflow } from '@/api/workflows'
import { useRepoSyncStatus } from '@/hooks/useRepoSyncStatus'
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
  const [requirement, setRequirement] = useState('')
  const [projectId, setProjectId] = useState('')
  // Environment (phase 3D). 'existing_codebase' makes the workflow work inside
  // the project's connected repository, where the readable files are a set the
  // client approves — not the whole repo.
  const [mode, setMode] = useState<'scratch' | 'existing_codebase'>('scratch')
  // G8: ask the implementation phase for working code. Default OFF — the design
  // documents are what §9 chose, and code generation needs the code service to be
  // configured; a workflow that asks for code and cannot get it fails loudly, so
  // this must be a deliberate choice rather than a default.
  const [deliverCode, setDeliverCode] = useState(false)
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

  const isExisting = mode === 'existing_codebase'

  // Existing-codebase work needs a connected, synced repository AND a human
  // approval step before anything can run — so this mode is created as a draft
  // and started from its Codebase tab, never launched from here.
  //
  // WHY the check is duplicated on the client: the backend already refuses the
  // run (SeedWorkspace fails with "no approved files to read"), which is the
  // correct place for the guarantee but arrives after the client has committed
  // to a workflow. Reading the project's repo status up front turns a
  // guaranteed failure into a sentence, and `enabled` keeps the query from
  // firing until a project is selected.
  const { data: repoStatus } = useRepoSyncStatus(projectId, isExisting)
  const repoReady = !!repoStatus?.connected && repoStatus.status === 'complete'
  const repoBlocked = isExisting && !!projectId && !repoReady

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
        requirementText: requirement.trim() || undefined,
        mode,
        deliverCode,
      })
      // Launching straight into a run is right for scratch and wrong for an
      // existing codebase: the approved readable set is what the runner seeds the
      // workspace from, and it cannot exist before the client has seen this
      // screen's Codebase tab. A draft is started from there.
      if (isExisting) return wf
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
    setRequirement('')
    setProjectId('')
    setMode('scratch')
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
          <label className="mb-1 block text-xs font-medium text-text-secondary">
            Requirement
            <span className="ml-1 text-text-disabled">(optional — the details behind the title)</span>
          </label>
          <textarea
            value={requirement}
            onChange={(e) => setRequirement(e.target.value)}
            rows={4}
            placeholder="Paste or type the full requirement. Experts work from this instead of the title alone."
            className="w-full resize-y rounded-md border border-glass-border bg-surface-overlay px-3 py-2 text-sm text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-brand/40"
          />
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-text-secondary">Environment</label>
          <div className="grid grid-cols-2 gap-2">
            <button
              type="button"
              onClick={() => setMode('scratch')}
              className={
                mode === 'scratch'
                  ? 'rounded-md border border-brand/60 bg-surface-overlay px-3 py-2 text-left text-xs text-text-primary'
                  : 'rounded-md border border-glass-border px-3 py-2 text-left text-xs text-text-secondary hover:bg-surface-overlay/60'
              }
            >
              <span className="block font-medium">From scratch</span>
              <span className="block text-[10px] text-text-disabled">Build something new</span>
            </button>
            <button
              type="button"
              onClick={() => setMode('existing_codebase')}
              className={
                mode === 'existing_codebase'
                  ? 'rounded-md border border-brand/60 bg-surface-overlay px-3 py-2 text-left text-xs text-text-primary'
                  : 'rounded-md border border-glass-border px-3 py-2 text-left text-xs text-text-secondary hover:bg-surface-overlay/60'
              }
            >
              <span className="block font-medium">Existing codebase</span>
              <span className="block text-[10px] text-text-disabled">
                Work inside the project's connected repository
              </span>
            </button>
          </div>
          {isExisting && (
            <p className="mt-1 text-[10px] text-text-disabled">
              {!projectId
                ? 'Select a project first — this mode reads a repository that must already be connected to it.'
                : repoBlocked && repoStatus?.connected
                  ? `The repository is connected but not synced yet (${repoStatus.status}). Wait for the sync, then create the workflow.`
                  : repoBlocked
                    ? 'This project has no repository connected. Connect GitHub or GitLab on the project page first.'
                    : 'Created as a draft: you approve which files the experts may read, then press Start on the Codebase tab.'}
            </p>
          )}
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

        <label className="mb-3 flex items-start gap-2 cursor-pointer">
          <input
            type="checkbox"
            className="mt-0.5"
            checked={deliverCode}
            onChange={(e) => setDeliverCode(e.target.checked)}
          />
          <span className="text-[11px] text-text-secondary">
            Also produce working code, not only the design documents.
            <span className="mt-0.5 block text-text-disabled">
              Use this when the deliverable is code — in an existing-codebase workflow the patch
              to your repository is the point. Needs the code service to be running: a workflow
              that asks for code and cannot produce it fails with that reason instead of quietly
              shipping design only.
            </span>
          </span>
        </label>

        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={handleClose} disabled={createMut.isPending}>Cancel</Button>
          <Button
            onClick={() => createMut.mutate()}
            disabled={
              createMut.isPending ||
              !title.trim() ||
              !projectId ||
              selectedExpertIds.length === 0 ||
              repoBlocked
            }
          >
            {createMut.isPending
              ? 'Creating...'
              : isExisting
                ? 'Create & choose files'
                : 'Launch Workflow'}
          </Button>
        </div>
      </div>
    </Modal>
  )
}
