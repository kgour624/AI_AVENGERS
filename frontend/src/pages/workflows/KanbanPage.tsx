import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getKanban, getWorkflow } from '@/api/workflows'
import type { KanbanTask } from '@/api/workflows'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'

// Kanban column definitions
const COLUMNS: { key: KanbanTask['status']; label: string; color: string }[] = [
  { key: 'todo',         label: 'To Do',        color: 'text-text-secondary' },
  { key: 'in_progress',  label: 'In Progress',  color: 'text-brand' },
  { key: 'under_review', label: 'Under Review', color: 'text-mode-warn' },
  { key: 'blocked',      label: 'Blocked',      color: 'text-mode-refuse' },
  { key: 'done',         label: 'Done',         color: 'text-mode-advise' },
]

function TaskCard({ task }: { task: KanbanTask }) {
  return (
    <Card className="mb-2 p-3">
      <p className="text-sm font-medium text-text-primary">{task.title}</p>
      <p className="mt-1 text-xs text-text-secondary">{task.expertName}</p>
      {task.costUsd > 0 && (
        <p className="mt-1 text-xs text-text-disabled">${task.costUsd.toFixed(4)}</p>
      )}
    </Card>
  )
}

function KanbanPage() {
  const { id } = useParams<{ id: string }>()

  const { data: workflow } = useQuery({
    queryKey: ['workflow', id],
    queryFn: () => getWorkflow(id!),
    enabled: !!id,
    refetchInterval: 5000,
  })

  const { data: kanban, isLoading } = useQuery({
    queryKey: ['workflow', id, 'kanban'],
    queryFn: () => getKanban(id!),
    enabled: !!id,
    refetchInterval: 5000, // poll every 5s — SSE upgrade is Phase F scope
  })

  const tasks = kanban?.tasks ?? []

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-xl font-semibold text-text-primary">
          {workflow?.title ?? 'Workflow'}
        </h1>
        <div className="mt-1 flex items-center gap-3">
          {workflow && (
            <>
              <Badge variant="neutral">{workflow.currentPhase.replace(/_/g, ' ')}</Badge>
              <Badge variant={workflow.status === 'running' ? 'brand' : 'neutral'}>
                {workflow.status.replace(/_/g, ' ')}
              </Badge>
              <span className="text-xs text-text-disabled">
                ${workflow.costSpentUsd.toFixed(4)} / ${workflow.costBudgetUsd.toFixed(2)}
              </span>
            </>
          )}
        </div>
      </div>

      {/* Kanban board */}
      {isLoading ? (
        <div className="grid grid-cols-5 gap-4">
          {COLUMNS.map((col) => (
            <Skeleton key={col.key} className="h-48" />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-5 gap-4">
          {COLUMNS.map((col) => {
            const colTasks = tasks.filter((t) => t.status === col.key)
            return (
              <div key={col.key}>
                <div className="mb-2 flex items-center justify-between">
                  <span className={`text-xs font-semibold uppercase tracking-wide ${col.color}`}>
                    {col.label}
                  </span>
                  <span className="text-xs text-text-disabled">{colTasks.length}</span>
                </div>
                <div className="min-h-24 rounded-lg bg-surface-overlay p-2">
                  {colTasks.map((task) => (
                    <TaskCard key={task.id} task={task} />
                  ))}
                  {colTasks.length === 0 && (
                    <p className="text-center text-xs text-text-disabled py-4">Empty</p>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

export const Component = KanbanPage
