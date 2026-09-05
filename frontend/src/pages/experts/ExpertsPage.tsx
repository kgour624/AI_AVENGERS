import type { LoaderFunctionArgs } from 'react-router-dom'
import { useLoaderData } from 'react-router-dom'
import { getExperts } from '@/api/experts'
import type { Expert } from '@/types/expert'
import { ExpertCard } from '@/components/expert/ExpertCard'

async function loader({ request }: LoaderFunctionArgs) {
  const experts = await getExperts({ signal: request.signal })
  return { experts }
}

export const expertsRoute = { element: <ExpertsPage />, loader }

export default function ExpertsPage() {
  const { experts } = useLoaderData() as { experts: Expert[] }

  return (
    <div className="p-6">
      <h1 className="mb-6 text-xl font-semibold">Available Domain Experts</h1>
      {experts.length === 0 ? (
        <p className="text-text-secondary">No experts available yet.</p>
      ) : (
        <div className="space-y-3">
          {experts.map((expert) => (
            <ExpertCard key={expert.id} expert={expert} />
          ))}
        </div>
      )}
    </div>
  )
}
