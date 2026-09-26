import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { listWorkflows } from '@/api/workflows'
import { queryKeys } from '@/api/queryKeys'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Skeleton } from '@/components/ui/Skeleton'
import { CreateWorkflowModal } from './CreateWorkflowModal'
import { cn } from '@/utils/cn'

const STATUS_VARIANT: Record<string, 'brand' | 'neutral' | 'warning' | 'danger'> = {
  draft:                   'neutral',
  running:                 'brand',
  paused_for_approval:     'warning',
  paused_for_client_input: 'warning',
  completed:               'neutral',
  cancelled:               'neutral',
  failed:                  'danger',
}

function WorkflowsPage() {
  const [isCreateOpen, setIsCreateOpen] = useState(false)

  const { data: workflows = [], isLoading } = useQuery({
    queryKey: queryKeys.workflows.all,
    queryFn: listWorkflows,
    refetchInterval: (query) => {
      const hasActive = query.state.data?.some(
        (w) => w.status === 'running' || w.status === 'paused_for_approval' || w.status === 'paused_for_client_input'
      )
      return hasActive ? 5000 : false
    },
  })

  return (
    <div className="p-6">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Workflows</h1>
          <p className="mt-0.5 text-xs text-text-secondary">Multi-agent collaborative sessions</p>
        </div>
        <Button onClick={() => setIsCreateOpen(true)}>+ New Workflow</Button>
      </div>

      {isLoading && (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-20" />)}
        </div>
      )}

      {!isLoading && workflows.length === 0 && (
        <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-glass-border py-16 text-center">
          <p className="text-sm font-medium text-text-secondary">No workflows yet</p>
          <p className="mt-1 text-xs text-text-disabled">
            Launch a workflow to let your domain experts collaborate on a requirement.
          </p>
          <Button className="mt-4" onClick={() => setIsCreateOpen(true)}>+ New Workflow</Button>
        </div>
      )}

      <div className="space-y-3">
        {workflows.map((wf) => (
          <Card key={wf.id} glow="purple">
            <div className="flex items-start justify-between gap-4">
              <div className="min-w-0 flex-1">
                <Link
                  to={`/workflows/${wf.id}/kanban`}
                  className="text-sm font-medium text-text-primary transition-colors hover:text-brand"
                >
                  {wf.title}
                </Link>
                <p className="mt-0.5 text-xs text-text-disabled">
                  {wf.currentPhase.replace(/_/g, ' ')}
                  {' · '}
                  ${wf.costSpentUsd.toFixed(4)} / ${wf.costBudgetUsd.toFixed(2)}
                </p>
              </div>
              <div className="flex flex-shrink-0 items-center gap-2">
                <Badge variant={STATUS_VARIANT[wf.status] ?? 'neutral'}>
                  {wf.status.replace(/_/g, ' ')}
                </Badge>
                <Link
                  to={`/workflows/${wf.id}/kanban`}
                  className={cn(
                    'rounded-md border border-glass-border px-2.5 py-1 text-xs text-text-secondary',
                    'transition-all hover:border-brand/40 hover:text-brand'
                  )}
                >
                  View Kanban
                </Link>
              </div>
            </div>
          </Card>
        ))}
      </div>

      <CreateWorkflowModal isOpen={isCreateOpen} onClose={() => setIsCreateOpen(false)} />
    </div>
  )
}

export const Component = WorkflowsPage
