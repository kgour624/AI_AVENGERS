import { useQuery } from '@tanstack/react-query'
import { getExpertTopics } from '@/api/experts'
import { Modal } from '@/components/ui/Modal'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'

export interface ExpertTopicsModalProps {
  isOpen: boolean
  onClose: () => void
  expertId: string
  expertName: string
}

const DEPTH_LABELS = ['', 'Basic', 'Intermediate', 'Advanced', 'Expert', 'Master']

/**
 * Feature #4 fix (docs bug list): ExpertCard already had an
 * onViewTopics callback prop wired to a "View Topics" button, but
 * ExpertsPage.tsx never passed a handler - clicking it did nothing.
 *
 * KNOWN SCOPE CORRECTION (not what the original bug report claimed):
 * verified against expert/handler.go's real GetTopics SQL - it
 * selects only topic/depth_level/chunk_count/complexity_ceiling.
 * can_handle/cannot_handle/example_questions are NOT selected despite
 * existing as columns on expert_capabilities (same gap already
 * documented in types/expert.ts's ExpertTopic interface). This modal
 * only shows what the endpoint actually returns - it does not
 * fabricate the missing fields.
 */
export function ExpertTopicsModal({ isOpen, onClose, expertId, expertName }: ExpertTopicsModalProps) {
  const { data: topics, isLoading } = useQuery({
    queryKey: ['experts', expertId, 'topics'],
    queryFn: () => getExpertTopics(expertId),
    enabled: isOpen,
  })

  return (
    <Modal isOpen={isOpen} onClose={onClose} className="max-w-xl">
      <h2 className="mb-1 text-lg font-semibold text-text-primary">{expertName}</h2>
      <p className="mb-4 text-xs text-text-secondary">Topics by depth level</p>

      {isLoading && (
        <div className="space-y-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-12" />
          ))}
        </div>
      )}

      {!isLoading && topics?.length === 0 && (
        <p className="text-sm text-text-secondary">No capability data yet for this expert.</p>
      )}

      <div className="max-h-96 space-y-2 overflow-y-auto">
        {topics?.map((topic) => (
          <div
            key={topic.topic}
            className="flex items-center justify-between rounded-md border border-surface-border bg-surface-raised px-3 py-2"
          >
            <div>
              <p className="text-sm text-text-primary">{topic.topic}</p>
              <p className="text-xs text-text-disabled">{topic.chunkCount} chunks</p>
            </div>
            <div className="flex items-center gap-2">
              <Badge variant="neutral">{DEPTH_LABELS[topic.depthLevel]}</Badge>
              <span className="text-xs text-text-disabled">{topic.depthLevel}/5</span>
            </div>
          </div>
        ))}
      </div>
    </Modal>
  )
}
