import type { LoaderFunctionArgs } from 'react-router-dom'
import { useLoaderData } from 'react-router-dom'
import { getExperts } from '@/api/experts'
import type { Expert } from '@/types/expert'

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
      <div className="space-y-3">
        {experts.map((e) => (
          <div key={e.id} className="rounded-lg border border-surface-border bg-surface-raised p-4">
            <p className="font-medium">{e.name}</p>
            <p className="text-sm text-text-secondary">Domain: {e.domain}</p>
          </div>
        ))}
      </div>
    </div>
  )
}
