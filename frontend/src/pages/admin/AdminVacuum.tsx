import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { baseAPI } from '@/api/base'

export default function AdminVacuum() {
  const qc = useQueryClient()
  const [pattern, setPattern] = useState('')
  const [previewText, setPreviewText] = useState('uh so guys you know this is a test lecture transcript with thank you so much and can you see my screen?')
  const [diffResult, setDiffResult] = useState<any>(null)
  const [previewResult, setPreviewResult] = useState<any>(null)
  const [uploadFile, setUploadFile] = useState<File | null>(null)
  const [selectedJob, setSelectedJob] = useState<string | null>(null)
  const [jobStatus, setJobStatus] = useState('')
  const [jobLimit] = useState(10)
  const [jobOffset, setJobOffset] = useState(0)
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [downloadingId, setDownloadingId] = useState<string | null>(null)

  const { data: patterns } = useQuery({
    queryKey: ['vacuum-patterns'],
    queryFn: async () => (await baseAPI.get('/api/v1/admin/vacuum/patterns')).data.data,
  })
  const { data: candidates } = useQuery({
    queryKey: ['vacuum-candidates'],
    queryFn: async () => (await baseAPI.get('/api/v1/admin/vacuum/candidates?status=pending')).data.data,
  })
  const { data: jobs } = useQuery({
    queryKey: ['vacuum-jobs', jobStatus, jobLimit, jobOffset],
    queryFn: async () => {
      const params = new URLSearchParams({ limit: String(jobLimit), offset: String(jobOffset) })
      if (jobStatus) params.set('status', jobStatus)
      return (await baseAPI.get(`/api/v1/admin/vacuum/jobs?${params.toString()}`)).data.data
    },
    refetchInterval: 5000,
  })
  const { data: brain } = useQuery({
    queryKey: ['vacuum-brain'],
    queryFn: async () => (await baseAPI.get('/api/v1/admin/vacuum/brain/version')).data,
    refetchInterval: 10000,
  })
  const { data: metrics } = useQuery({
    queryKey: ['vacuum-metrics'],
    queryFn: async () => (await baseAPI.get('/api/v1/admin/vacuum/metrics')).data.data,
    refetchInterval: 8000,
  })
  const { data: stats } = useQuery({
    queryKey: ['vacuum-stats'],
    queryFn: async () => (await baseAPI.get('/api/v1/admin/vacuum/stats')).data.data,
    refetchInterval: 5000,
  })
  const { data: chunks } = useQuery({
    queryKey: ['vacuum-chunks', selectedJob],
    queryFn: async () => (await baseAPI.get(`/api/v1/admin/vacuum/jobs/${selectedJob}/chunks`)).data.data,
    enabled: !!selectedJob,
  })

  const createMut = useMutation({
    mutationFn: async () => (await baseAPI.post('/api/v1/admin/vacuum/patterns', { pattern, patternType: 'PHRASE', category: 'filler' })).data,
    onSuccess: () => { setPattern(''); qc.invalidateQueries({ queryKey: ['vacuum-patterns'] }); qc.invalidateQueries({ queryKey: ['vacuum-brain'] }) },
  })
  const diffMut = useMutation({
    mutationFn: async () => {
      console.log('[vacuum diff] POST /preview/diff ->', previewText.slice(0,120))
      const res = await baseAPI.post('/api/v1/admin/vacuum/preview/diff', { text: previewText })
      console.log('[vacuum diff] response', res.data)
      return res.data
    },
    onSuccess: (d:any) => {
      const data=d?.data??d
      // baseAPI's response interceptor deep-camelizes keys, so read camelCase (dsaCleaned, combinedCleaned, ...).
      console.log('[vacuum diff] success', data)
      setDiffResult(data)
    },
    onError: (e:any) => {
      const msg = e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || 'Diff preview failed'
      console.error('[vacuum diff] error', e?.response?.status, e?.response?.data, e)
      alert('Diff failed: ' + msg)
    },
  })
  const autoPromoteMut = useMutation({
    mutationFn: async () => (await baseAPI.post('/api/v1/admin/vacuum/auto-promote')).data,
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['vacuum-candidates'] }); qc.invalidateQueries({ queryKey: ['vacuum-brain'] }); qc.invalidateQueries({ queryKey: ['vacuum-metrics'] }) },
  })
  const { data: evalData } = useQuery({
    queryKey: ['vacuum-eval'],
    queryFn: async () => (await baseAPI.get('/api/v1/admin/vacuum/eval?limit=10')).data.data,
    refetchInterval: 15000,
  })
  const { data: driftData } = useQuery({
    queryKey: ['vacuum-drift'],
    queryFn: async () => (await baseAPI.get('/api/v1/admin/vacuum/drift')).data,
    refetchInterval: 30000,
  })
  const approveMut = useMutation({
    mutationFn: async (id: string) => (await baseAPI.post(`/api/v1/admin/vacuum/candidates/${id}/approve`)).data,
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['vacuum-candidates'] }); qc.invalidateQueries({ queryKey: ['vacuum-patterns'] }); qc.invalidateQueries({ queryKey: ['vacuum-brain'] }); qc.invalidateQueries({ queryKey: ['vacuum-metrics'] }) },
  })
  const rejectMut = useMutation({
    mutationFn: async (id: string) => (await baseAPI.post(`/api/v1/admin/vacuum/candidates/${id}/reject`)).data,
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['vacuum-candidates'] }); qc.invalidateQueries({ queryKey: ['vacuum-metrics'] }) },
  })
  const bulkApproveMut = useMutation({
    mutationFn: async () => (await baseAPI.post('/api/v1/admin/vacuum/candidates/bulk-approve', { ids: Array.from(selected) })).data.data,
    onSuccess: () => { setSelected(new Set()); qc.invalidateQueries({ queryKey: ['vacuum-candidates'] }); qc.invalidateQueries({ queryKey: ['vacuum-patterns'] }); qc.invalidateQueries({ queryKey: ['vacuum-brain'] }); qc.invalidateQueries({ queryKey: ['vacuum-metrics'] }) },
  })
  const bulkRejectMut = useMutation({
    mutationFn: async () => (await baseAPI.post('/api/v1/admin/vacuum/candidates/bulk-reject', { ids: Array.from(selected) })).data.data,
    onSuccess: () => { setSelected(new Set()); qc.invalidateQueries({ queryKey: ['vacuum-candidates'] }); qc.invalidateQueries({ queryKey: ['vacuum-metrics'] }) },
  })
  const previewMut = useMutation({
    mutationFn: async () => (await baseAPI.post('/api/v1/admin/vacuum/preview', { text: previewText })).data.data,
    onSuccess: (d) => { setPreviewResult(d); qc.invalidateQueries({ queryKey: ['vacuum-candidates'] }) },
  })
  const uploadMut = useMutation({
    mutationFn: async () => {
      const fd = new FormData()
      fd.append('file', uploadFile as File)
      return (await baseAPI.post('/api/v1/admin/vacuum/upload', fd, { headers: { 'Content-Type': 'multipart/form-data' } })).data.data
    },
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['vacuum-jobs'] }); qc.invalidateQueries({ queryKey: ['vacuum-stats'] }); setUploadFile(null) },
  })
  const pickMut = useMutation({
    mutationFn: async () => (await baseAPI.post('/api/v1/admin/vacuum/jobs/pick-execute')).data.data,
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['vacuum-jobs'] }); qc.invalidateQueries({ queryKey: ['vacuum-stats'] }) },
  })
  const executeMut = useMutation({
    mutationFn: async (id: string) => (await baseAPI.post(`/api/v1/admin/vacuum/jobs/${id}/execute`)).data.data,
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['vacuum-jobs'] }); qc.invalidateQueries({ queryKey: ['vacuum-stats'] }) },
  })
  const retryMut = useMutation({
    mutationFn: async (id: string) => (await baseAPI.post(`/api/v1/admin/vacuum/jobs/${id}/retry`)).data.data,
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['vacuum-jobs'] }); qc.invalidateQueries({ queryKey: ['vacuum-stats'] }) },
  })

  const handleDownload = async (id: string, fileName: string) => {
    if (downloadingId) return
    setDownloadingId(id)
    try {
      const res = await baseAPI.get(`/api/v1/admin/vacuum/jobs/${id}/download`, { responseType: 'blob' })
      const blob = res.data instanceof Blob ? res.data : new Blob([res.data])
      if (blob.type.includes('json')) {
        const txt = await blob.text()
        try { const j = JSON.parse(txt); if (j?.error || j?.success === false) throw new Error(j.error?.message || j.message || 'Download failed') } catch {}
      }
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url; a.download = fileName + '.cleaned.txt'
      document.body.appendChild(a); a.click(); a.remove()
      window.URL.revokeObjectURL(url)
      qc.invalidateQueries({ queryKey: ['vacuum-jobs'] })
      qc.invalidateQueries({ queryKey: ['vacuum-stats'] })
    } catch (e: any) {
      let m = e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || 'Download failed'
      if (e?.response?.data instanceof Blob) { try { const txt = await e.response.data.text(); const j = JSON.parse(txt); m = j?.error?.message || j?.message || m } catch {} }
      alert(m)
    } finally { setDownloadingId(null) }
  }

  const toggle = (id: string) => {
    const n = new Set(selected)
    if (n.has(id)) n.delete(id); else n.add(id)
    setSelected(n)
  }
  const allIds = (candidates || []).map((c: any) => c.id)
  const allSelected = allIds.length > 0 && allIds.every((id: string) => selected.has(id))
  const toggleAll = () => {
    if (allSelected) setSelected(new Set())
    else setSelected(new Set(allIds))
  }
  const brainVersion = (brain as any)?.version ?? (brain as any)?.data?.version ?? (brain as any)?.brainVersion ?? '-'
  const brainPatterns = (brain as any)?.patterns ?? (brain as any)?.data?.patterns ?? '-'

  return (
    <div className="p-6 space-y-6 max-w-6xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Vacuum Brain — Admin Cleaner</h1>
          <p className="text-xs text-slate-500">Phase 7: Split View + Self-Learning (approve → brain v++ hot-reload 30s) + Prometheus/LangFuse</p>
        </div>
        <div className="text-right">
          <div className="px-3 py-1 rounded bg-violet-600 text-white font-mono text-xs">brain v{brainVersion} · {brainPatterns} patterns</div>
          <div className="text-xs text-slate-400">hot-reload 30s{metrics ? ` · pending ${metrics.candidatesPending ?? 0} · done ${metrics.vacuumJobsDone ?? 0}` : ''}</div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="border rounded p-4 space-y-3">
          <h2 className="font-semibold">Add Kachra Pattern</h2>
          <div className="flex gap-2">
            <input value={pattern} onChange={e => setPattern(e.target.value)} placeholder="e.g. So guys, uh, [Music]" className="flex-1 border rounded px-3 py-2 text-sm" />
            <button onClick={() => createMut.mutate()} disabled={!pattern.trim()} className="px-4 py-2 bg-slate-900 text-white rounded text-sm">Add</button>
          </div>
          <div className="max-h-64 overflow-auto text-sm divide-y">
            {(patterns || []).map((p: any) => (
              <div key={p.id} className="py-1 flex justify-between"><span>{p.pattern} <span className="text-slate-400">({p.category}/{p.patternType})</span></span><span className="text-slate-400">{p.hitCount}</span></div>
            ))}
            {(!patterns || patterns.length === 0) && <div className="text-slate-400 py-2">No patterns yet — add fuel.</div>}
          </div>
        </div>

        <div className="border rounded p-4 space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="font-semibold">Candidate Kachra — Self-Learning</h2>
            <div className="flex gap-2 items-center text-xs">
              <label className="flex items-center gap-1"><input type="checkbox" checked={allSelected} onChange={toggleAll} /> all</label>
              <span className="text-slate-500">{selected.size} sel</span>
              <button disabled={selected.size===0 || bulkApproveMut.isPending} onClick={()=>bulkApproveMut.mutate()} className="px-2 py-1 bg-emerald-600 text-white rounded disabled:opacity-40">Bulk Approve → v++</button>
              <button disabled={selected.size===0 || bulkRejectMut.isPending} onClick={()=>bulkRejectMut.mutate()} className="px-2 py-1 bg-slate-200 rounded disabled:opacity-40">Bulk Reject</button>
            </div>
          </div>
          <div className="max-h-64 overflow-auto text-sm divide-y border rounded">
            {(candidates || []).map((c: any) => (
              <div key={c.id} className="py-2 px-2 flex items-start justify-between gap-2 hover:bg-slate-50">
                <div className="flex gap-2 items-start"><input type="checkbox" checked={selected.has(c.id)} onChange={()=>toggle(c.id)} className="mt-1" /><div><div className="font-mono font-medium text-sm">{c.pattern} <span className="text-slate-400 text-xs">· {c.category} · ×{c.hitCount}</span></div><div className="text-slate-500 text-xs truncate max-w-[260px]">{c.contextSnippet}</div></div></div>
                <div className="flex gap-1 shrink-0">
                  <button onClick={() => approveMut.mutate(c.id)} className="px-2 py-1 bg-green-600 text-white rounded text-xs">Approve</button>
                  <button onClick={() => rejectMut.mutate(c.id)} className="px-2 py-1 bg-slate-200 rounded text-xs">Reject</button>
                </div>
              </div>
            ))}
            {(!candidates || candidates.length === 0) && <div className="text-slate-400 py-2 text-center">No pending candidates — run Preview or upload to generate.</div>}
          </div>
        </div>
      </div>

      <div className="border rounded p-4 space-y-3">
        <div className="flex justify-between items-center"><h2 className="font-semibold">Split View — Raw vs Clean (Preview)</h2><div className="flex gap-2"><button onClick={()=>diffMut.mutate()} disabled={!previewText.trim() || diffMut.isPending} className="px-3 py-1 bg-indigo-600 text-white rounded text-xs disabled:opacity-40">{diffMut.isPending?'Diff...':'DSA vs LLM vs Combined'}</button><button onClick={()=>autoPromoteMut.mutate()} disabled={autoPromoteMut.isPending} className="px-3 py-1 bg-emerald-600 text-white rounded text-xs disabled:opacity-40">Auto-Promote ≥3hits 0.90</button></div></div>
        <div className="grid md:grid-cols-3 gap-3">
          <div><div className="text-xs font-semibold text-slate-600">RAW</div><textarea value={previewText} onChange={e => setPreviewText(e.target.value)} placeholder="Paste raw transcript snippet..." rows={6} className="w-full border rounded px-3 py-2 text-sm font-mono" /><div className="flex gap-2 mt-2"><button onClick={() => previewMut.mutate()} disabled={!previewText.trim()} className="flex-1 px-4 py-2 bg-violet-600 text-white rounded text-sm disabled:opacity-40">{previewMut.isPending ? 'Cleaning...' : 'Preview Clean →'}</button></div>{previewResult && <div className="text-xs text-slate-500 mt-1">hits {previewResult.hits?.length ?? 0} · p1Hits {previewResult.p1Hits?.length ?? 0} · v{previewResult.brainVersion} · sha {previewResult.sha256Out?.slice(0,8)}</div>}
            {diffResult && <div className="text-[11px] mt-1 p-1 bg-indigo-50 rounded">DSA hits {(diffResult.dsaHits||[]).length} → Combined {(diffResult.combinedHits||[]).length} (+{diffResult.llmAddedHits||0} LLM) · filler thr {diffResult.fillerThreshold}</div>}
            {driftData && (driftData as any)?.data?.drift && <div className="text-[11px] mt-1 p-1 bg-red-50 text-red-700 rounded">⚠ Drift: recent {(driftData as any).data.recentAvg?.toFixed?.(2)} vs older {(driftData as any).data.olderAvg?.toFixed?.(2)}</div>}
            {evalData && <div className="text-[11px] mt-1">eval last {(evalData as any)?.eval?.length??0} · avg score {(evalData as any)?.avgEvalScore?.toFixed?.(3) ?? '-'}</div>}
          </div>
          <div><div className="text-xs font-semibold text-emerald-700">DSA CLEANED</div>{!(diffResult||previewResult) ? <div className="h-[140px] flex items-center justify-center text-slate-400 text-sm border border-dashed rounded bg-white">Run Preview Clean</div> : <><pre className="whitespace-pre-wrap text-sm bg-emerald-50 border rounded p-2 min-h-[120px] max-h-[180px] overflow-auto">{(diffResult?.dsaCleaned ?? previewResult.cleaned) || '(empty)'}</pre><div className="flex flex-wrap gap-1 mt-2">{((diffResult?.dsaHits||previewResult?.hits||[]) as any[]).map((h:any,i:number)=>(<span key={i} className="px-2 py-0.5 bg-slate-100 text-slate-700 rounded text-xs font-mono">{h.Pattern||h.pattern||h}</span>))}</div></>}</div>
          <div><div className="text-xs font-semibold text-indigo-700">COMBINED (DSA+LLM)</div>{!diffResult ? <div className="h-[140px] flex items-center justify-center text-slate-400 text-sm border border-dashed rounded bg-white">Run DSA vs LLM vs Combined</div> : <><pre className="whitespace-pre-wrap text-sm bg-indigo-50 border rounded p-2 min-h-[120px] max-h-[180px] overflow-auto">{diffResult.combinedCleaned || '(empty)'}</pre><div className="flex flex-wrap gap-1 mt-2">{(diffResult.combinedHits||[]).map((h:any,i:number)=>(<span key={i} className="px-2 py-0.5 bg-violet-100 text-violet-700 rounded text-xs font-mono">{h.Pattern||h.pattern}</span>))}</div><div className="text-[10px] text-slate-400 mt-1">LLM added {diffResult.llmAddedHits||0} spans · thr {diffResult.fillerThreshold}</div></>}</div>
        </div>
      </div>

      <div className="border rounded p-4 space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="font-semibold">Upload Transcript (Input→DB→S3)</h2>
          <button onClick={() => pickMut.mutate()} className="px-3 py-1 bg-slate-900 text-white rounded text-xs">Pick &amp; Execute (DTS sem5)</button>
        </div>
        <div className="flex gap-2">
          <input type="file" accept=".txt,.md" onChange={e => setUploadFile(e.target.files?.[0] || null)} className="text-sm" />
          <button onClick={() => uploadMut.mutate()} disabled={!uploadFile} className="px-4 py-1 bg-indigo-600 text-white rounded text-sm disabled:opacity-40">Upload &amp; Enqueue</button>
        </div>
        {pickMut.data && <div className="text-xs text-green-600">Picked {pickMut.data.length} jobs — verified &amp; chunked.</div>}
      </div>

      {stats && (
        <div className="border rounded p-3 flex flex-wrap gap-4 text-xs items-center">
          <span className="font-semibold">Stats:</span>
          {Object.entries(stats.byStatus || {}).map(([k,v]: any) => (<span key={k}>{k}: <b>{String(v)}</b></span>))}
          <span>verified_chunks: <b>{stats.verifiedChunks ?? 0}</b></span>
          <span>avg_duration_ms: <b>{stats.avgDurationMs ?? 0}</b></span>
        </div>
      )}
      <div className="border rounded p-4">
        <div className="flex items-center justify-between mb-2">
          <h2 className="font-semibold">File Jobs (DTS queue — paginated)</h2>
          <div className="flex gap-2 items-center">
            <select value={jobStatus} onChange={e=>{setJobStatus(e.target.value); setJobOffset(0)}} className="border rounded px-2 py-1 text-xs">
              <option value="">all</option>
              <option value="scheduled">scheduled</option>
              <option value="verifying">verifying</option>
              <option value="done">done</option>
              <option value="failed">failed</option>
              <option value="quarantined">quarantined</option>
            </select>
            <button disabled={jobOffset===0} onClick={()=>setJobOffset(o=>Math.max(0,o-jobLimit))} className="px-2 py-1 bg-slate-100 rounded text-xs disabled:opacity-40">Prev</button>
            <span className="text-xs text-slate-500">offset {jobOffset}</span>
            <button disabled={!jobs || jobs.length < jobLimit} onClick={()=>setJobOffset(o=>o+jobLimit)} className="px-2 py-1 bg-slate-100 rounded text-xs disabled:opacity-40">Next</button>
          </div>
        </div>
        <div className="text-xs bg-amber-50 border border-amber-200 rounded px-2 py-1 mb-2 text-amber-800">verifying = background me LLM 78s tak chal raha hai (106 chunks parallel). Is dauran Retry mat dabao — complete hote hi Download aa jayega.</div>
        <div className="text-sm divide-y">
          {(jobs || []).map((j: any) => (
            <div key={j.id} className="py-2 flex justify-between items-center">
              <div className="cursor-pointer" onClick={() => setSelectedJob(j.id)}><span className="font-medium">{j.fileName}</span> — <span className={j.status==='done' ? "text-emerald-600 font-semibold" : j.status==='verifying' ? "text-amber-600 font-semibold" : j.status==='failed' ? "text-red-600" : ""}>{j.status}</span> / phase {j.phase} {j.status==='done' && '✓'} {j.durationMs!=null && <span className="text-slate-500 text-xs"> · {j.durationMs}ms</span>} {j.preservationVerified && <span className="text-emerald-600 text-xs"> · preserved ✓</span>} {j.llmClassifierLabel && <span className="text-violet-600 text-xs"> · {j.llmClassifierLabel}</span>} {j.llmHeadingCount>0 && <span className="text-slate-500 text-xs"> · headings:{j.llmHeadingCount}</span>} {j.downloadCount>0 && <span className="text-slate-500 text-xs"> · ↓{j.downloadCount}</span>} {j.lastError && <span className="text-red-500 text-xs"> · {j.lastError.slice(0,40)}</span>}</div>
              <div className="flex gap-2 items-center"><span className="text-slate-400 text-xs">{j.s3Key}</span>
                {j.status==='verifying' && <span className="px-2 py-1 bg-amber-100 text-amber-800 rounded text-xs flex items-center gap-1 border border-amber-300"><span className="w-3 h-3 border-2 border-amber-600 border-t-transparent rounded-full animate-spin inline-block"></span> Processing... {j.phase}/6</span>}
                {j.status==='scheduled' && <button onClick={() => executeMut.mutate(j.id)} disabled={executeMut.isPending} className="px-2 py-1 bg-amber-500 text-white rounded text-xs disabled:opacity-40">Execute</button>}
                {(j.status==='failed' || j.status==='quarantined') && <button onClick={() => { if(confirm('Retry this job?')) retryMut.mutate(j.id) }} disabled={retryMut.isPending} className="px-2 py-1 bg-emerald-600 text-white rounded text-xs disabled:opacity-40">Retry</button>}
                {j.status==='done' && <button disabled={downloadingId===j.id} onClick={()=>handleDownload(j.id, j.fileName)} className="px-3 py-1 bg-blue-600 text-white rounded text-xs disabled:opacity-50 shadow hover:bg-blue-700">{downloadingId===j.id ? 'Downloading...' : 'Download'}</button>}
                <button onClick={() => setSelectedJob(j.id)} className="px-2 py-1 bg-slate-100 rounded text-xs">Chunks</button>
              </div>
            </div>
          ))}
          {(!jobs || jobs.length === 0) && <div className="text-slate-400">No jobs.</div>}
        </div>
        {selectedJob && (
          <div className="mt-3 border rounded p-2 text-xs">
            <div className="font-semibold mb-1">Chunks for {selectedJob} ({(chunks||[]).length}) — verified contiguous 0..n</div>
            {(chunks||[]).map((ch:any)=>(<div key={ch.chunkIndex} className="py-1 border-t flex justify-between"><span>#{ch.chunkIndex} [{ch.charStart}-{ch.charEnd}] {ch.status} {ch.llmLabel && <span className="text-violet-600">· {ch.llmLabel} {(ch.llmConfidence||0).toFixed(2)}</span>}</span><span className="text-slate-400 font-mono">{ch.hashOutput?.slice(0,16)}</span></div>))}
            {(!chunks||chunks.length===0) && <div className="text-slate-400">No chunks yet — execute the job.</div>}
          </div>
        )}
      </div>
    </div>
  )
}

export const Component = AdminVacuum
