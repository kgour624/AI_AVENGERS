import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card } from '@/components/ui/Card'
import { cn } from '@/utils/cn'
import { getExpertTopics } from '@/api/experts'
import type { Expert, ExpertTopic } from '@/types/expert'

/**
 * ExpertCapabilitiesTable - Feature #26
 *
 * Shows the expert's actual capability rows: every topic the corpus was tagged
 * with, how many chunks cover it, and the coverage band the backend assigned.
 *
 * WHY this reads from GET /experts/:id/topics and not from the expert's aggregate
 * fields: the aggregate ("118 topics, average depth 2.0") is the least
 * informative view of the same data, and it hid the only question worth asking —
 * WHICH topics are thin. The per-topic rows were already available; this view
 * just had no way to reach them, so the admin could not see that most topics were
 * a handful of chunks each.
 *
 * WHY there are no stars and no "Basic/Advanced/Master" words:
 * the backend derives depth_level from how many CHUNKS cover a topic
 * (capability_builder.calculateDepthLevel: <5 → 1, 5-14 → 2, 15-29 → 3,
 * 30-49 → 4, ≥50 → 5). It is a coverage band, not a judgement of how hard the
 * content is, so this view labels it as coverage. A five-star row would claim the
 * course is research-level when the only fact behind it is "more than 50 chunks
 * mention this topic" — and worse, the number moves if the chunk size changes,
 * with no change to the course at all.
 */

/**
 * THIN_TOPIC_CHUNKS mirrors the floor in the backend's own depth table: below 5
 * chunks the pipeline assigns depth level 1, i.e. it declines to claim anything
 * about that topic. So this is not a new threshold invented by the UI — it is the
 * point where the pipeline itself stops making a claim.
 */
const THIN_TOPIC_CHUNKS = 5

/** The capability bucket for chunks the topic extractor could not label. */
const GENERAL_TOPIC = 'general'

const TOPIC_PAGE_SIZE = 25

/**
 * coverageBand renders the same bands the backend uses, but as the facts behind
 * them (chunk counts) rather than as quality words.
 */
function coverageBand(chunkCount: number): { label: string; className: string } {
  if (chunkCount >= 50) return { label: '50+ chunks', className: 'text-purple-400' }
  if (chunkCount >= 30) return { label: '30-49 chunks', className: 'text-blue-400' }
  if (chunkCount >= 15) return { label: '15-29 chunks', className: 'text-green-400' }
  if (chunkCount >= THIN_TOPIC_CHUNKS) return { label: '5-14 chunks', className: 'text-yellow-400' }
  return { label: 'under 5 chunks', className: 'text-gray-400' }
}

interface ExpertTopicsSummary {
  /** Topics the capability table actually holds. */
  listedTopics: number
  /** Chunks those topics account for. */
  coveredChunks: number
  /** Topics with too few chunks for the pipeline to claim anything about them. */
  thinTopics: number
  /** Chunks the topic extractor could not label — they answer nothing specific. */
  unlabelledChunks: number
  /** Largest chunk count, used only to scale the coverage bars. */
  peakChunks: number
  sorted: ExpertTopic[]
}

function summariseTopics(topics: ExpertTopic[]): ExpertTopicsSummary {
  const sorted = [...topics].sort((a, b) => b.chunkCount - a.chunkCount)
  let coveredChunks = 0
  let thinTopics = 0
  let unlabelledChunks = 0
  for (const topic of topics) {
    coveredChunks += topic.chunkCount
    if (topic.chunkCount < THIN_TOPIC_CHUNKS) thinTopics++
    if (topic.topic === GENERAL_TOPIC) unlabelledChunks += topic.chunkCount
  }
  return {
    listedTopics: topics.length,
    coveredChunks,
    thinTopics,
    unlabelledChunks,
    peakChunks: sorted[0]?.chunkCount ?? 0,
    sorted,
  }
}

interface ExpertCapabilitiesTableProps {
  expert: Expert
}

export function ExpertCapabilitiesTable({ expert }: ExpertCapabilitiesTableProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const [showAll, setShowAll] = useState(false)

  // Fetched only when the section is opened: this is a per-expert table read that
  // nobody asked for until they expand it.
  const { data: topics, isLoading, isError } = useQuery({
    queryKey: ['experts', expert.id, 'topics'],
    queryFn: () => getExpertTopics(expert.id),
    enabled: isExpanded,
    staleTime: 60_000,
  })

  const summary = useMemo(() => summariseTopics(topics ?? []), [topics])
  const visibleTopics = showAll ? summary.sorted : summary.sorted.slice(0, TOPIC_PAGE_SIZE)
  const hiddenTopics = summary.sorted.length - visibleTopics.length

  return (
    <div className="mt-3">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="flex items-center gap-2 text-sm font-medium text-text-secondary hover:text-text-primary transition-colors"
      >
        <span>{isExpanded ? '\u25bc' : '\u25b6'}</span>
        <span>Capabilities Overview</span>
      </button>

      {isExpanded && (
        <Card className="mt-2 p-4" glow="cyan">
          <div className="space-y-3">
            {isLoading && (
              <p className="text-xs text-text-disabled">Loading topics from the corpus...</p>
            )}

            {isError && (
              <p className="text-xs text-glow-amber">
                Could not load the topic list. The aggregate counts below come from the expert
                record and may be stale.
              </p>
            )}

            {topics && (
              <>
                {/* Summary Stats — measured from the capability rows themselves,
                    not from the expert's cached totals. */}
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-xs text-text-secondary">Topics in capability table</p>
                    <p className="text-lg font-semibold text-text-primary">
                      {summary.listedTopics}
                    </p>
                    {summary.listedTopics !== expert.totalTopics && (
                      <p className="text-[10px] text-text-disabled">
                        expert record says {expert.totalTopics}
                      </p>
                    )}
                  </div>
                  <div>
                    <p className="text-xs text-text-secondary">Chunks these topics cover</p>
                    <p className="text-lg font-semibold text-text-primary">
                      {summary.coveredChunks}
                    </p>
                    {summary.coveredChunks !== expert.totalChunks && (
                      <p className="text-[10px] text-text-disabled">
                        expert record says {expert.totalChunks} total
                      </p>
                    )}
                  </div>
                </div>

                {/* Coverage distribution — the numbers an admin can act on, and
                    the reason this view exists at all. Stated as facts, with what
                    they mean, rather than as a single averaged score. */}
                <div className="rounded border border-border bg-bg-tertiary p-3">
                  <p className="text-xs font-medium text-text-primary">Coverage</p>
                  <ul className="mt-1 space-y-1 text-[11px] text-text-secondary">
                    <li>
                      <span className="font-mono text-text-primary">{summary.thinTopics}</span> topic
                      {summary.thinTopics === 1 ? '' : 's'} have fewer than {THIN_TOPIC_CHUNKS} chunks —
                      too thin for the pipeline to claim anything about them.
                    </li>
                    {summary.unlabelledChunks > 0 && (
                      <li className="text-glow-amber">
                        <span className="font-mono">{summary.unlabelledChunks}</span> chunks are in the
                        &ldquo;{GENERAL_TOPIC}&rdquo; bucket — topic extraction did not label them, so
                        they answer nothing specific.
                      </li>
                    )}
                    <li className="text-text-disabled">
                      Coverage is measured in chunks, not difficulty. The band the backend reports is
                      derived from how many chunks mention a topic, so it reflects how the corpus was
                      split as much as what the course teaches.
                    </li>
                  </ul>
                </div>

                {/* Per-topic rows */}
                <div className="overflow-hidden rounded border border-border">
                  <div className="grid grid-cols-[1fr_auto_auto] gap-2 border-b border-border bg-bg-tertiary px-2 py-1 text-[10px] uppercase tracking-wide text-text-disabled">
                    <span>Topic</span>
                    <span className="text-right">Chunks</span>
                    <span className="w-24 text-right">Coverage</span>
                  </div>
                  {visibleTopics.map((topic) => {
                    const band = coverageBand(topic.chunkCount)
                    const width = summary.peakChunks > 0
                      ? Math.max(2, Math.round((topic.chunkCount / summary.peakChunks) * 100))
                      : 0
                    return (
                      <div
                        key={topic.topic}
                        className="grid grid-cols-[1fr_auto_auto] items-center gap-2 border-b border-border px-2 py-1 last:border-b-0"
                      >
                        <span
                          className={cn(
                            'truncate text-xs',
                            topic.topic === GENERAL_TOPIC ? 'text-text-disabled' : 'text-text-secondary',
                          )}
                          title={topic.topic}
                        >
                          {topic.topic}
                        </span>
                        <span className="text-right text-xs font-mono text-text-primary">
                          {topic.chunkCount}
                        </span>
                        <span className="w-24">
                          <span className="block h-1.5 w-full overflow-hidden rounded bg-bg-secondary">
                            <span
                              className="block h-full rounded bg-brand/70"
                              style={{ width: `${width}%` }}
                            />
                          </span>
                          <span className={cn('mt-0.5 block text-right text-[9px]', band.className)}>
                            {band.label}
                          </span>
                        </span>
                      </div>
                    )
                  })}
                </div>

                {hiddenTopics > 0 && (
                  <button
                    onClick={() => setShowAll(true)}
                    className="text-[11px] text-brand hover:text-brand-hover"
                  >
                    Show all {summary.sorted.length} topics ({hiddenTopics} more)
                  </button>
                )}
              </>
            )}

            {/* Rating is left as the expert record reports it. It is a human
                signal, unrelated to the corpus, so it stays out of the coverage
                section above. */}
            <div>
              <p className="text-xs text-text-secondary mb-1">Average Rating</p>
              <div className="flex items-center gap-2">
                <span className="text-lg font-semibold text-text-primary">
                  {expert.avgRating.toFixed(2)}
                </span>
                <span className="text-xs text-text-secondary">
                  ({expert.totalRatings ?? 0} ratings)
                </span>
              </div>
            </div>
          </div>
        </Card>
      )}
    </div>
  )
}
