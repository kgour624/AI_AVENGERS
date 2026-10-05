import { useEffect, useState } from 'react'
import { useReceptionistStream } from ''hooks/useReceptionistStream'' (see below for file content)

// NOTES SIDEBAR - Right fixed, "Baad me karenge"
export function NotesSidebar({ sessionId }: { sessionId: string }) {
  const [notes, setNotes] = useState<any[]>([])
  const [text, setText] = useState("")
  const { events } = useReceptionistStream(sessionId)
  useEffect(()=>{ fetch(`/api/receptionist/sessions/${sessionId}/notes`, {headers:{Authorization:`Bearer ${localStorage.getItem('token')}`}}).then(r=>r.json()).then(d=>setNotes(d.notes||[])) },[sessionId])
  useEffect(()=>{ const last = events.filter(e=>e.type==='note').slice(-1)[0]; if(last) setNotes(n=>[...n, last.payload]) },[events])
  const add = async ()=>{ if(!text) return; const r=await fetch(`/api/receptionist/sessions/${sessionId}/notes`,{method:'POST',headers:{'Content-Type':'application/json',Authorization:`Bearer ${localStorage.getItem('token')}`},body:JSON.stringify({text})}); const d=await r.json(); setNotes(n=>[...n,d]); setText("") }
  return (
    <div className="w-64 border-l p-3 bg-yellow-50 h-full overflow-auto">
      <div className="font-medium text-sm">📝 Notes - Baad me karenge</div>
      <div className="flex gap-1 mt-2"><input value={text} onChange={e=>setText(e.target.value)} placeholder="Note likho..." className="flex-1 border px-2 py-1 text-sm"/><button onClick={add} className="bg-blue-600 text-white px-2 rounded">Add</button></div>
      <div className="mt-3 text-xs text-gray-500">{notes.length} notes - final se pehle discuss karna hai</div>
      {notes.map(n=> <div key={n.id} className="bg-white border p-2 rounded mt-2 text-sm">{n.text}</div>)}
    </div>
  )
}

// CHECKLIST PROGRESS - Contiguous bar
export function ChecklistProgress({ sessionId }: { sessionId: string }) {
  const [p, setP] = useState({total:0, contiguous:0, percent:0, next_idx:0})
  useEffect(()=>{ fetch(`/api/receptionist/sessions/${sessionId}/progress`,{headers:{Authorization:`Bearer ${localStorage.getItem('token')}`}}).then(r=>r.json()).then(setP) },[sessionId])
  return (
    <div className="border p-3 rounded bg-white">
      <div className="flex justify-between text-xs"><span>Progress {p.contiguous}/{p.total}</span><span>{p.percent}%</span></div>
      <div className="w-full bg-gray-200 h-2 rounded mt-1"><div className="bg-green-600 h-2 rounded" style={{width:`${p.percent}%`}}></div></div>
      <div className="text-xs text-gray-500 mt-1">Contiguous: gap pe break, out-of-order safe • Next idx {p.next_idx}</div>
    </div>
  )
}

// END BUTTON + FINAL VIEW - Single button, pyara bye
export function EndConversation({ sessionId, onEnded }: { sessionId: string; onEnded:(md:string,hing:string)=>void }) {
  const [loading, setLoading] = useState(false)
  const [finalMD, setFinalMD] = useState("")
  const [hinglish, setHinglish] = useState("")
  const end = async ()=>{
    setLoading(true)
    const r=await fetch(`/api/receptionist/sessions/${sessionId}/end`,{method:'POST',headers:{Authorization:`Bearer ${localStorage.getItem('token')}`}})
    const d=await r.json()
    setLoading(false)
    if(r.ok){ setFinalMD(d.final_md); setHinglish(d.hinglish); onEnded(d.final_md,d.hinglish) }
    else alert(d.message)
  }
  if(finalMD) return (
    <div className="border p-4 rounded bg-green-50">
      <div className="prose max-w-none whitespace-pre-wrap border p-3 bg-white rounded">{finalMD}</div>
      {hinglish && <div className="bg-yellow-50 border p-2 rounded mt-2 text-sm">💡 Hinglish: {hinglish}</div>}
      <div className="text-center mt-3 text-lg">🙏 Dhanyawaad, aapka PRD taiyaar hai! Phir milenge, khayal rakhna</div>
    </div>
  )
  return <button onClick={end} disabled={loading} className="w-full bg-purple-600 text-white py-2 rounded disabled:bg-gray-400">{loading?'Synthesis bana rahi hu...':'End Conversation'}</button>
}

// ReceptionistSection.tsx me layout:
// <div className="flex gap-4"><div className="flex-1"><LiveBar/><ChecklistProgress/><ConversationOneByOne/><ExpertGrid/><EndConversation/></div><NotesSidebar sessionId={session.id}/></div>
