import { useState } from 'react'
import { useReceptionistStream } from ''hooks/useReceptionistStream'' (see below for file content)

type Point = { id: string; q_text: string; status: string; judge_explained?: boolean }

export function ConversationOneByOne({ sessionId, checkpointId, points: initPoints, live, typing }: 
  { sessionId: string; checkpointId: string; points: Point[]; live: boolean; typing: boolean }) {
  const [points, setPoints] = useState<Point[]>(initPoints)
  const [currentIdx, setCurrentIdx] = useState(() => initPoints.findIndex(p => p.status==='draft' || p.status==='judge_explained') !== -1 ? initPoints.findIndex(p => p.status==='draft') : 0)
  const [explain, setExplain] = useState<string>("")
  const [showAddMore, setShowAddMore] = useState(false)
  const { events } = useReceptionistStream(sessionId)

  const cur = points[currentIdx]
  const allDone = points.every(p => p.status==='approved' || p.status==='dismissed')

  const answer = async (ans: "Haan"|"Nahi"|"Samjhao") => {
    setExplain("")
    const res = await fetch(`/api/receptionist/sessions/${sessionId}/checkpoints/${checkpointId}/answer`, {
      method: 'POST',
      headers: { 'Content-Type':'application/json', Authorization: `Bearer ${localStorage.getItem('token')}` },
      body: JSON.stringify({ point_id: cur.id, answer: ans })
    })
    const data = await res.json()
    if (data.judge && data.explain) {
      setExplain(data.explain) // Smart Judge ne samjhaya - ab fir se Haan/Nahi dikhao
      setPoints(data.points)
      return
    }
    setPoints(data.points)
    if (data.all_discussed) setShowAddMore(true)
    else if (currentIdx < points.length-1) setCurrentIdx(i=>i+1)
  }

  const addMore = async () => {
    const res = await fetch(`/api/receptionist/sessions/${sessionId}/checkpoints/${checkpointId}/add-more`, {
      method:'POST', headers:{'Content-Type':'application/json', Authorization:`Bearer ${localStorage.getItem('token')}`},
      body: JSON.stringify({ agenda: "LLD" })
    })
    const data = await res.json()
    setPoints(data.points)
    setShowAddMore(false)
    setCurrentIdx(points.length) // naye points pe jao
  }

  const approve = async () => {
    const res = await fetch(`/api/receptionist/sessions/${sessionId}/checkpoints/${checkpointId}/approve`, {
      method:'POST', headers:{Authorization:`Bearer ${localStorage.getItem('token')}`}
    })
    const data = await res.json()
    alert(`Approved ${data.approved_count} points, ab experts ko bhej sakte ho!`)
  }

  if (allDone) {
    return (
      <div className="border p-4 rounded bg-green-50">
        <div className="font-medium">Sab points discuss ho gaye! {live && '● Live'}</div>
        <div className="flex gap-2 mt-3">
          <button onClick={addMore} className="bg-yellow-500 px-3 py-1 rounded">Add More Points (search se)</button>
          <button onClick={approve} className="bg-green-600 text-white px-4 py-1 rounded">Ye {points.filter(p=>p.status==='approved').length} points leke jau? Approve</button>
        </div>
      </div>
    )
  }

  if (!cur) return <div>Loading... {typing && 'Typing...'}</div>

  return (
    <div className="border p-4 rounded bg-white shadow">
      <div className="text-xs text-gray-500">{currentIdx+1}/{points.length} • {live?'● Live':''} {typing && 'Typing...'}</div>
      <div className="font-medium mt-2">Q: {cur.q_text}</div>
      {cur.status==='judge_explained' && <div className="text-xs text-orange-600 mt-1">⚠️ Ek baar samjhaya, fir Nahi to hata denge</div>}
      {explain && <div className="bg-yellow-50 border p-3 rounded mt-3 text-sm">💡 {explain}<div className="text-xs mt-1">Fir bhi Nahi bolna hai to hata denge, force nahi.</div></div>}
      <div className="flex gap-2 mt-3">
        <button onClick={()=>answer("Haan")} className="bg-green-600 text-white px-4 py-1 rounded">Haan</button>
        <button onClick={()=>answer("Nahi")} className="bg-gray-300 px-4 py-1 rounded">Nahi</button>
        <button onClick={()=>answer("Samjhao")} className="bg-blue-500 text-white px-4 py-1 rounded">Samjhao</button>
      </div>
      <div className="text-xs mt-2 text-gray-400">Pressure: {events.filter(e=>e.type==='pressure').slice(-1)[0]?.payload.score ?? 0}/100 • Mother-like care</div>
    </div>
  )
}
