import { useState } from 'react'

type ExpertResult = { id: string; expert_id: string; answer_md: string; hinglish_summary: string; latency_ms: number; cost_usd: number; rating?: number }

export function ExpertCard({ result, onRate }: { result: ExpertResult; onRate: (id: string, rating: number)=>void }) {
  const [rating, setRating] = useState<number| null>(result.rating ?? null)
  const [hover, setHover] = useState(0)

  const rate = async (r: number) => {
    setRating(r)
    await fetch(`/api/receptionist/expert-responses/${result.id}/rate`, {
      method:'POST', headers:{'Content-Type':'application/json', Authorization:`Bearer ${localStorage.getItem('token')}`},
      body: JSON.stringify({ rating: r })
    })
    onRate(result.id, r) // parent will handle APPROVED->Next else REOPENED
  }

  return (
    <div className="border rounded bg-white shadow p-4 mb-3">
      <div className="flex justify-between text-xs text-gray-500">
        <span>Expert {result.expert_id.slice(0,6)} • {result.latency_ms}ms • ${result.cost_usd.toFixed(6)}</span>
        <span className="bg-green-100 px-2 py-0.5 rounded">EN markdown</span>
      </div>
      <div className="prose prose-sm max-w-none mt-2 whitespace-pre-wrap border p-3 rounded bg-gray-50">{result.answer_md}</div>
      {result.hinglish_summary && <div className="bg-yellow-50 border p-2 rounded mt-2 text-sm">💡 Hinglish: {result.hinglish_summary}</div>}
      <div className="flex items-center gap-1 mt-3">
        <span className="text-xs mr-2">Rate:</span>
        {[1,2,3,4,5].map(s => (
          <button key={s} onMouseEnter={()=>setHover(s)} onMouseLeave={()=>setHover(0)} onClick={()=>rate(s)}
            className={`text-xl ${ (hover?s<=hover: rating && s<=rating) ? 'text-yellow-400' : 'text-gray-300'}`}>★</button>
        ))}
        {rating && <span className="text-xs ml-2">{rating<=3?'Reopen same point':'Approved & Next'}</span>}
      </div>
    </div>
  )
}

// Parent Grid: ExpertWaitGrid + ExpertCard list
export function ExpertGrid({ sessionId, checkpointId, results, live }: { sessionId: string; checkpointId: string; results: ExpertResult[]; live: boolean }) {
  if (results.length===0) return <div className="animate-pulse p-4 border rounded">{live?'● Live Typing... Experts se baat (0/3, 7s)':'Loading experts...'}</div>
  return <div>{results.map(r=> <ExpertCard key={r.id} result={r} onRate={(id,rating)=> {
    if(rating>3) { /* contiguous next idx */ } else { /* reopen logic, pressure+10 */ }
  }} />)}</div>
}
