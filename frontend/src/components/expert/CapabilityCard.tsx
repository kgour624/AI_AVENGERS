import type { ExpertTopic } from '@/types/expert'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'

/**
 * Renders one topic's capability profile.
 *
 * PHASE 4 CORRECTION: canHandle/cannotHandle/exampleQuestions are now
 * optional on ExpertTopic (see types/expert.ts) because the real
 * GET /experts/:id/topics handler never selects those columns - they
 * will be undefined in practice today, not just empty arrays. Guards
 * below check `topic.canHandle && topic.canHandle.length > 0` instead
 * of the Phase 2 version's `topic.canHandle.length > 0`, which would
 * throw "Cannot read properties of undefined" the moment this
 * component renders against the real backend response.
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

      {topic.canHandle && topic.canHandle.length > 0 && (
        <div className="mt-3">
          <p className="text-xs font-medium text-mode-advise">Can handle</p>
          <ul className="mt-1 list-inside list-disc text-sm text-text-secondary">
            {topic.canHandle.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </div>
      )}

      {topic.cannotHandle && topic.cannotHandle.length > 0 && (
        <div className="mt-3">
          <p className="text-xs font-medium text-mode-refuse">Cannot handle</p>
          <ul className="mt-1 list-inside list-disc text-sm text-text-secondary">
            {topic.cannotHandle.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </div>
      )}

      {topic.exampleQuestions && topic.exampleQuestions.length > 0 && (
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
