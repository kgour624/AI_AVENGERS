import { useState } from 'react'

export function useTemplate(sessionId: string, idx: number) {
  const [points, setPoints] = useState<any[]>([])
  const [loading, setLoading] = useState(false)

  const ensure = async (agenda: string) => {
    setLoading(true)
    try {
      const res = await fetch(`/api/receptionist/sessions/${sessionId}/checkpoints/${idx}/template/ensure`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${localStorage.getItem('token')}` },
        body: JSON.stringify({ agenda })
      })
      const data = await res.json()
      if (res.ok) setPoints(data.points)
      else throw new Error(data.message)
    } finally { setLoading(false) }
  }

  const clearAndNext = async (oldId: string, newIdx: number, agenda: string) => {
    setLoading(true)
    const res = await fetch(`/api/receptionist/sessions/${sessionId}/checkpoints/next`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${localStorage.getItem('token')}` },
      body: JSON.stringify({ old_checkpoint_id: oldId, new_idx: newIdx, agenda })
    })
    const data = await res.json()
    if (res.ok) setPoints(data.points)
    setLoading(false)
  }

  return { points, loading, ensure, clearAndNext }
}
