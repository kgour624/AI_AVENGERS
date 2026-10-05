import { useState } from 'react'
import { useTemplate } from ''hooks/useTemplate'' (see below for file content)
import { useReceptionistStream } from ''hooks/useReceptionistStream'' (see below for file content)

export function TemplateOneByOne({ sessionId, idx, agenda, checkpointId }: { sessionId: string; idx: number; agenda: string; checkpointId: string }) {
  const { points, loading, ensure } = useTemplate(sessionId, idx)
  const { live, typing } = useReceptionistStream(sessionId)
  const [current, setCurrent] = useState(0)

  // Empty -> search trigger
  if (points.length === 0 && !loading) {
    return <button onClick={() => ensure(agenda)} className="bg-blue-600 text-white px-4 py-2 rounded">Search se points lao</button>
  }
  if (loading || typing) return <div className="animate-pulse p-4 border rounded">● {typing ? 'Typing...' : 'Searching...'} {live && 'Live'}</div>

  const p = points[current]
  if (!p) return <div>All done - Approve gate pe jao</div>

  return (
    <div className="border p-4 rounded bg-white">
      <div className="text-xs text-gray-500">{current+1}/{points.length} • {live? '● Live':''}</div>
      <div className="font-medium mt-2">Q: {p.q_text}</div>
      <div className="flex gap-2 mt-3">
        <button onClick={() => { p.status='approved'; setCurrent(c=>c+1) }} className="bg-green-600 text-white px-3 py-1 rounded">Haan</button>
        <button onClick={() => { p.status='dismissed'; setCurrent(c=>c+1) }} className="bg-gray-300 px-3 py-1 rounded">Nahi</button>
        <button onClick={() => ensure(agenda + ' explain ' + p.q_text)} className="bg-yellow-500 px-3 py-1 rounded">Samjhao</button>
      </div>
      <div className="text-xs mt-2 text-gray-400">Source: search • Next pe purana clear hoga</div>
    </div>
  )
}
