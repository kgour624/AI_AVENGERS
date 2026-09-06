import { useState } from 'react'
import type { LoaderFunctionArgs } from 'react-router-dom'
import { useLoaderData } from 'react-router-dom'
import { motion, useReducedMotion } from 'framer-motion'
import { getExperts } from '@/api/experts'
import type { Expert } from '@/types/expert'
import { ExpertCard } from '@/components/expert/ExpertCard'
import { ExpertTopicsModal } from '@/components/expert/ExpertTopicsModal'
import { fadeUp, staggerContainer, ARC_MOTION } from '@/design-system/motion'

async function loader({ request }: LoaderFunctionArgs) {
  const experts = await getExperts({ signal: request.signal })
  return { experts }
}

export const expertsRoute = { element: <ExpertsPage />, loader }

export default function ExpertsPage() {
  const { experts } = useLoaderData() as { experts: Expert[] }
  const reduceMotion = useReducedMotion()
  // Feature #4 fix (docs bug list): ExpertCard already had an
  // onViewTopics callback prop, never wired to anything here.
  const [topicsTargetId, setTopicsTargetId] = useState<string | null>(null)

  return (
    <div className="p-6">
      <h1 className="mb-6 text-xl font-semibold">Available Domain Experts</h1>
      {experts.length === 0 ? (
        <p className="text-text-secondary">No experts available yet.</p>
      ) : (
        <motion.div
          initial={reduceMotion ? undefined : 'hidden'}
          animate="visible"
          variants={staggerContainer}
          className="space-y-3"
        >
          {experts.map((expert, i) => (
            <motion.div key={expert.id} variants={i < ARC_MOTION.maxStaggerItems ? fadeUp : undefined}>
              <ExpertCard expert={expert} onViewTopics={(expertId) => setTopicsTargetId(expertId)} />
            </motion.div>
          ))}
        </motion.div>
      )}

      {topicsTargetId && (
        <ExpertTopicsModal
          isOpen={topicsTargetId !== null}
          onClose={() => setTopicsTargetId(null)}
          expertId={topicsTargetId}
          expertName={experts.find((e) => e.id === topicsTargetId)?.name ?? ''}
        />
      )}
    </div>
  )
}
