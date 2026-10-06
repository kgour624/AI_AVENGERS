import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { baseAPI } from '@/api/base'

export function AdminRagSettings() {
  const [form, setForm] = useState<any>({
    enableParentChild: false,
    atomicCodeFence: true,
    childSoftLimit: 150,
    childHardLimit: 220,
    parentSoftLimit: 1200,
    parentHardLimit: 1500,
    overlapTokens: 20,
    topKChildren: 50,
    topKParents: 3,
    rerankTopN: 10,
    maxHops: 3,
    rrfK: 60
  })
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [msg, setMsg] = useState('')

  useEffect(() => {
    ;(async () => {
      try {
        const res = await baseAPI.get('/api/v1/admin/retrieval-config')
        const raw = res.data?.data || res.data || {}
        setForm({
          enableParentChild: raw.enableParentChild ?? raw.enable_parent_child ?? false,
          atomicCodeFence: raw.atomicCodeFence ?? raw.atomic_code_fence ?? true,
          childSoftLimit: raw.childSoftLimit ?? raw.child_soft_limit ?? 150,
          childHardLimit: raw.childHardLimit ?? raw.child_hard_limit ?? 220,
          parentSoftLimit: raw.parentSoftLimit ?? raw.parent_soft_limit ?? 1200,
          parentHardLimit: raw.parentHardLimit ?? raw.parent_hard_limit ?? 1500,
          overlapTokens: raw.overlapTokens ?? raw.overlap_tokens ?? 20,
          topKChildren: raw.topKChildren ?? raw.top_k_children ?? 50,
          topKParents: raw.topKParents ?? raw.top_k_parents ?? 3,
          rerankTopN: raw.rerankTopN ?? raw.rerank_top_n ?? 10,
          maxHops: raw.maxHops ?? raw.max_hops ?? 3,
          rrfK: raw.rrfK ?? raw.rrf_k ?? 60
        })
      } catch (e: any) {
        setMsg('Failed to load settings')
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  const set = (key: string, val: any) => setForm((p: any) => ({ ...p, [key]: val }))

  const handleSave = async () => {
    setSaving(true)
    setMsg('')
    try {
      const payload = {
        enable_parent_child: form.enableParentChild,
        atomic_code_fence: form.atomicCodeFence,
        child_soft_limit: form.childSoftLimit,
        child_hard_limit: form.childHardLimit,
        parent_soft_limit: form.parentSoftLimit,
        parent_hard_limit: form.parentHardLimit,
        overlap_tokens: form.overlapTokens,
        top_k_children: form.topKChildren,
        top_k_parents: form.topKParents,
        rerank_top_n: form.rerankTopN,
        max_hops: form.maxHops,
        rrf_k: form.rrfK,
        // Dual payload to bypass frontend camelizer bugs
        enableParentChild: form.enableParentChild,
        atomicCodeFence: form.atomicCodeFence,
        childSoftLimit: form.childSoftLimit,
        childHardLimit: form.childHardLimit,
        parentSoftLimit: form.parentSoftLimit,
        parentHardLimit: form.parentHardLimit,
        overlapTokens: form.overlapTokens,
        topKChildren: form.topKChildren,
        topKParents: form.topKParents,
        rerankTopN: form.rerankTopN,
        maxHops: form.maxHops,
        rrfK: form.rrfK
      }
      await baseAPI.put('/api/v1/admin/retrieval-config', payload)
      setMsg('Saved ✓')
    } catch (e: any) {
      setMsg(e?.response?.data?.msg || 'Save failed')
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <div className="p-8 text-neutral-400">Loading...</div>

  return (
    <div className="p-8 max-w-5xl flex flex-col gap-6 font-sans text-white">
      <div>
        <h1 className="text-2xl font-bold mb-4">RAG Retrieval</h1>
        <Link 
          to="/admin/chunks"
          className="inline-block px-4 py-1.5 rounded-full border border-purple-600 text-purple-400 text-sm font-medium hover:bg-purple-900/30 transition-colors mb-3"
        >
          Open Chunk Explorer →
        </Link>
        <p className="text-sm text-neutral-400">
          Parent-Child: children 150/220 rerank 512-safe to 3 parents 1200/1500, overlap 20, RRF 60. OFF = zero regression
        </p>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
        
        {/* Card 1: Master switch & chunking */}
        <div 
          className="relative rounded-2xl p-6 bg-[#0a0a0f] border border-[#1a1a24] flex flex-col gap-6"
          style={{ boxShadow: '0 0 40px rgba(138, 43, 226, 0.05)' }}
        >
          <div className="flex flex-col gap-3">
            <h2 className="text-lg font-semibold text-white">Master switch & chunking</h2>
            <p className="text-sm text-neutral-400 mb-2">ChunkMarkdown parents+children, word*1.3, atomic fence.</p>
            
            <div className="rounded-xl border border-[#222] bg-[#111116] p-4 flex items-center justify-between">
              <div className="flex flex-col">
                <span className="font-medium text-white">Enable Parent-Child</span>
                <span className="text-xs text-neutral-500 mt-0.5">OFF = zero regression</span>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input type="checkbox" className="sr-only peer" checked={form.enableParentChild} onChange={e => set('enableParentChild', e.target.checked)} />
                <div className="w-11 h-6 bg-[#2a2a35] peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-neutral-400 peer-checked:after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-purple-600"></div>
              </label>
            </div>

            <div className="rounded-xl border border-[#222] bg-[#111116] p-4 flex items-center justify-between">
              <div className="flex flex-col">
                <span className="font-medium text-white">Atomic code fence</span>
                <span className="text-xs text-neutral-500 mt-0.5">Never split inside triple backticks</span>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input type="checkbox" className="sr-only peer" checked={form.atomicCodeFence} onChange={e => set('atomicCodeFence', e.target.checked)} />
                <div className="w-11 h-6 bg-[#2a2a35] peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-neutral-400 peer-checked:after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-purple-600"></div>
              </label>
            </div>
          </div>

          <div className="flex flex-col gap-5 mt-2">
            <SliderRow label="CHILD SOFT" value={form.childSoftLimit} min={50} max={300} onChange={v => set('childSoftLimit', v)} />
            <SliderRow label="CHILD HARD" value={form.childHardLimit} min={100} max={400} onChange={v => set('childHardLimit', v)} />
            <SliderRow label="PARENT SOFT" value={form.parentSoftLimit} min={500} max={2000} step={100} onChange={v => set('parentSoftLimit', v)} />
            <SliderRow label="PARENT HARD" value={form.parentHardLimit} min={600} max={2500} step={100} onChange={v => set('parentHardLimit', v)} />
            <SliderRow label="OVERLAP TOKENS" value={form.overlapTokens} min={0} max={100} step={5} onChange={v => set('overlapTokens', v)} />
          </div>
        </div>

        {/* Card 2: Retrieval & ranking */}
        <div 
          className="relative rounded-2xl p-6 bg-[#0a0a0f] border border-[#1a1a24] flex flex-col gap-6"
          style={{ boxShadow: '0 0 40px rgba(0, 255, 255, 0.03)' }}
        >
          <div className="flex flex-col gap-2">
            <h2 className="text-lg font-semibold text-white">Retrieval & ranking</h2>
            <p className="text-sm text-neutral-400 mb-2">ANN → Rerank → Agentic loop settings.</p>
          </div>

          <div className="flex flex-col gap-5">
            <SliderRow label="TOP-K CHILDREN (Scan N)" value={form.topKChildren} min={10} max={100} step={5} onChange={v => set('topKChildren', v)} />
            <SliderRow label="RERANK TOP-N (Rerank M)" value={form.rerankTopN} min={5} max={50} onChange={v => set('rerankTopN', v)} />
            <SliderRow label="TOP-K PARENTS (Return K)" value={form.topKParents} min={1} max={10} onChange={v => set('topKParents', v)} />
            <SliderRow label="MAX AGENTIC HOPS" value={form.maxHops} min={1} max={10} onChange={v => set('maxHops', v)} />
            <SliderRow label="RRF K (Fusion constant)" value={form.rrfK} min={20} max={100} step={5} onChange={v => set('rrfK', v)} />
          </div>

          <div className="mt-auto pt-6 border-t border-[#222]">
            <h3 className="text-xs font-bold text-neutral-500 mb-3 tracking-wider">DRY-RUN PREVIEW (LIVE)</h3>
            <div className="bg-[#111116] p-4 rounded-xl border border-cyan-900/30 text-sm font-mono text-cyan-200">
              Children ~{form.childSoftLimit}-{form.childHardLimit}, 
              Parents ~{form.parentSoftLimit}-{form.parentHardLimit},<br/>
              Scan {form.topKChildren} → rerank {form.rerankTopN} → return {form.topKParents}
            </div>
          </div>
        </div>

      </div>

      <div className="flex items-center gap-4 mt-2">
        <button 
          onClick={handleSave} 
          disabled={saving}
          className="px-6 py-2.5 rounded-lg text-white font-medium shadow-lg transition-all"
          style={{ background: 'linear-gradient(90deg, #8a2be2, #00d2ff)' }}
        >
          {saving ? 'Saving...' : 'Save retrieval config'}
        </button>
        {msg && <span className={`text-sm ${msg.includes('failed') ? 'text-red-400' : 'text-green-400'}`}>{msg}</span>}
      </div>
    </div>
  )
}

function SliderRow({ label, value, min, max, step=1, onChange }: { label:string, value:number, min:number, max:number, step?:number, onChange:(v:number)=>void }) {
  const percent = Math.min(100, Math.max(0, ((value - min) / (max - min)) * 100))
  
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-2">
        <span className="text-xs font-bold tracking-wider text-neutral-400">{label}</span>
      </div>
      <div className="flex items-center gap-4">
        <div className="relative flex-1 h-1.5 bg-[#2a2a35] rounded-full">
          <div className="absolute left-0 top-0 h-full bg-cyan-500 rounded-full" style={{ width: `${percent}%` }}></div>
          <input 
            type="range" 
            min={min} 
            max={max} 
            step={step}
            value={value} 
            onChange={e => onChange(Number(e.target.value))}
            className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
          />
          <div 
            className="absolute top-1/2 -mt-2 w-4 h-4 bg-cyan-400 rounded-full shadow-[0_0_10px_rgba(0,255,255,0.4)] pointer-events-none"
            style={{ left: `calc(${percent}% - 8px)` }}
          ></div>
        </div>
        <div className="w-14 h-7 bg-[#222] rounded flex items-center justify-center border border-[#333]">
          <span className="text-xs text-neutral-300">{value}</span>
        </div>
      </div>
    </div>
  )
}

export const Component = AdminRagSettings
