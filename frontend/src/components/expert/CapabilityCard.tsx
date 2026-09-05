import type { ExpertTopic } from '@/types/expert'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'

/**
 * Renders one topic's capability profile.
 * Source: AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 5
 * (expert_capabilities table: topic, depth_level, chunk_count,
 * complexity_ceiling, can_handle, cannot_handle, example_questions)
 * and FRONTEND_SYSTEM_DESIGN.md section 3
 * (components/expert/CapabilityCard.tsx).
 *
 * This component did not have an explicit wireframe in the frontend
 * design doc (only listed by filename in section 3's tree) - built
 * against the backend's actual DB columns instead of guessing, since
 * every field on ExpertTopic maps 1:1 to a real column.
 */
export function CapabilityCard({ topic }: { topic: ExpertTopic }) {
  return (
    <Card>
      <div className="flex items-center justify-between">
        <p className="font-medium text-text-primary">{topic.topic}</p>
        <Badge variant="brand">
          Depth {topic.depthLevel}/5 \u00b7 {topic.complexityCeiling}
        </Badge>
      </div>
      <p className="mt-1 text-xs text-text-secondary">{topic.chunkCount} chunks</p>

      {topic.canHandle.length > 0 && (
        <div className="mt-3">
          <p className="text-xs font-medium text-mode-advise">Can handle</p>
          <ul className="mt-1 list-inside list-disc text-sm text-text-secondary">
            {topic.canHandle.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </div>
      )}

      {topic.cannotHandle.length > 0 && (
        <div className="mt-3">
          <p className="text-xs font-medium text-mode-refuse">Cannot handle</p>
          <ul className="mt-1 list-inside list-disc text-sm text-text-secondary">
            {topic.cannotHandle.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </div>
      )}

      {topic.exampleQuestions.length > 0 && (
        <div className="mt-3">
          <p className="text-xs font-medium text-text-secondary">Example questions</p>
          <ul className="mt-1 list-inside list-disc text-sm text-text-secondary">
            {topic.exampleQuestions.map((q) => (
              <li key={q}>{q}</li>
            ))}
          </ul>
        </div>
      )}
    </Card>
  )
}
