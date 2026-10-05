import { useState, useRef } from 'react';
import ChecklistManagerDialog from '../components/ChecklistManagerDialog';
import AskExpertsDialog from '../components/AskExpertsDialog';
import ChangeDialog from '../components/ChangeDialog';
import FinalResponseView from '../components/FinalResponseView';
import { useReceptionistSSE } from '../hooks/useReceptionistSSE';
import { baseAPI } from '@/api/base';

type Lang = 'EN'|'HI'|'IN_EN'|'HINGLISH';

export default function ReceptionistSection() {
  const [lang, setLang] = useState<Lang>('HINGLISH');
  const [persona, setPersona] = useState('software engineer');
  const [agenda, setAgenda] = useState('');
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [phase, setPhase] = useState<string>('INIT');

  const [checkOpen, setCheckOpen] = useState(false);
  const [askOpen, setAskOpen] = useState(false);
  const [changeOpen, setChangeOpen] = useState(false);

  const [messages, setMessages] = useState<{ role: 'user' | 'secretary'; text: string }[]>([]);
  const [input, setInput] = useState('');
  const [activeCheckpoint, setActiveCheckpoint] = useState<string | null>(null);
  const [checkpoints, setCheckpoints] = useState<any[]>([]);
  const [summary, setSummary] = useState<string>('');
  const [finalDoc, setFinalDoc] = useState<string>('');
  const [summaryApproved, setSummaryApproved] = useState(false);
  const [expertResponses, setExpertResponses] = useState<{id:string,name:string,markdown:string,rating?:number}[]>([]);
  const fileRef = useRef<HTMLInputElement>(null);

  const isPhase2 = phase === 'PHASE_2_CONVERSATION';
  const checklistCommitted = checkpoints.some(c => c.status === 'committed' || c.status === 'COMMITTED');
  const canAskExperts = isPhase2 && checklistCommitted;
  const canChat = isPhase2;

  const { data: live, connected } = useReceptionistSSE(sessionId);

  // FUTURE-PROOF: Single auth path — baseAPI = Authorization header + single-flight refresh + camelize.
  // Pehle raw fetch tha → Authorization missing + refresh ka new token retry pe use nahi hota → 401 loop (screenshot #2: sessions 401 → refresh 200 → sessions 401).
  // Ab baseAPI.request use → backend AuthMiddleware + frontend interceptor ek saath kaam karte hain.
  const api = async (path: string, opts: any = {}) => {
    const method = (opts.method || 'GET').toLowerCase();
    const headers: Record<string, string> = opts.headers || {};
    let data: any = undefined;
    let isForm = false;
    if (opts.body !== undefined) {
      if (typeof FormData !== 'undefined' && opts.body instanceof FormData) {
        data = opts.body;
        isForm = true;
        // let browser/axios set multipart boundary
        delete headers['Content-Type'];
        delete headers['content-type'];
      } else if (typeof opts.body === 'string') {
        try { data = JSON.parse(opts.body); } catch { data = opts.body; }
      } else {
        data = opts.body;
      }
    }
    const res = await baseAPI.request({ url: path, method, data, headers: isForm ? headers : headers });
    const d: any = res.data;
    if (d && typeof d === 'object' && 'success' in d && 'data' in d) return d.data ?? d;
    return d;
  };

  const [creating, setCreating] = useState(false);
  const createSession = async () => {
    if (creating) return;
    if (!persona.trim()) { alert('Persona (Response Type) required'); return; }
    setCreating(true);
    try {
      const d = await api('/api/receptionist/sessions', { method: 'POST', body: JSON.stringify({ language: lang, persona: persona.trim(), agenda }) });
      const sid = d?.id || d?.session?.id || d?.data?.id || d?.data?.session?.id;
      if (!sid) {
        alert('Session creation failed: no valid session ID returned');
        return;
      }
      setSessionId(sid);
      setPhase('PHASE_1_SETUP');
      setMessages([{ role: 'secretary', text: `Session created. Persona locked: "${persona}" | Lang: ${lang} | Agenda: ${agenda || '(empty)'}. Ab Checklist set karo.` }]);
      setCheckOpen(true);
    } catch (e: any) {
      const msg = e?.response?.data?.error || e?.response?.data?.message || e?.message || 'Create failed';
      alert(msg);
    } finally { setCreating(false); }
  };

  const handleChecklistSave = async (texts: string[], noChecklist: boolean) => {
    if (!sessionId) return;
    try {
      await api(`/api/receptionist/sessions/${sessionId}/setup`, { 
        method: 'POST', 
        body: JSON.stringify({ checklist: texts, texts, no_checklist: noChecklist }) 
      });
      setCheckOpen(false);
      setPhase('PHASE_2_CONVERSATION');
      const s = await api(`/api/receptionist/sessions/${sessionId}`);
      const cps = s?.checkpoints || s?.session?.checkpoints || [];
      setCheckpoints(cps);
      const first = cps.find((c: any) => c.status === 'IN_PROGRESS' || c.status === 'DISCUSSING' || c.status === 'in_progress' || c.status === 'PENDING' || c.status === 'pending') || cps[0];
      if (first) setActiveCheckpoint(first.id);
      setMessages(m => [...m, { role: 'secretary', text: `Checklist saved (${noChecklist ? 'General Discussion' : texts.length + ' points'}). Prior discussion se start karte hain — aapka motive kya hai?` }]);
    } catch (e: any) {
      alert(e?.message || 'Checklist save failed');
    }
  };

  const isSendEnabled = canChat && Boolean(activeCheckpoint) && Boolean(input.trim());

  const sendChat = async () => {
    if (!input.trim() || !sessionId || !activeCheckpoint) return;
    const text = input.trim();
    setMessages(m => [...m, { role: 'user', text }]);
    setInput('');
    try {
      const res = await api(`/api/receptionist/sessions/${sessionId}/chat`, { method: 'POST', body: JSON.stringify({ checkpoint_id: activeCheckpoint, message: text }) });
      setMessages(m => [...m, { role: 'secretary', text: res.reply || res.summary || res.message || "Samjha, is point pe note kar liya. Agla checkpoint?" }]);
    } catch (e: any) {
      alert(e?.message || 'Failed to send chat message');
    }
  };

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !sessionId) return;
    const form = new FormData();
    form.append('file', file);
    form.append('checkpoint_id', activeCheckpoint || '');
    setMessages(m => [...m, { role: 'user', text: `📎 Uploaded: ${file.name}` }]);
    try {
      await fetch(`/api/receptionist/sessions/${sessionId}/upload`, { method: 'POST', body: form });
      setMessages(m => [...m, { role: 'secretary', text: `File "${file.name}" padh liya aur context me add kar diya.` }]);
    } catch { setMessages(m => [...m, { role: 'secretary', text: `File parse failed, dobara try karo.` }]); }
    if (fileRef.current) fileRef.current.value = '';
  };

  const handleAskExperts = async (expertIds: string[], question: string) => {
    if (!sessionId) return;
    setAskOpen(false);
    setMessages(m => [...m, { role: 'user', text: `Ask-Experts (${expertIds.length}): ${question}` }]);
    setMessages(m => [...m, { role: 'secretary', text: `Experts se parallel puch raha hun (${expertIds.length} concurrent)...` }]);
    try {
      const d = await api(`/api/receptionist/sessions/${sessionId}/ask-experts`, { method: 'POST', body: JSON.stringify({ checkpoint_id: activeCheckpoint, expert_ids: expertIds, question }) });
      const results = d.results || [];
      setExpertResponses(results.map((x:any)=>({id:x.expert_id, name:x.expert_name||x.expert_id.slice(0,8), markdown:x.relevant})));
      setMessages(m => [...m, { role: 'secretary', text: `Expert answers aa gaye. Har answer ko rating do (≤3 pe re-ask hoga).` }]);
    } catch (e: any) { setMessages(m => [...m, { role: 'secretary', text: `Ask-Experts failed: ${e.message}` }]); }
  };

  const handleChangeSubmit = async (text: string) => {
    if (!sessionId || !activeCheckpoint) return;
    setChangeOpen(false);
    await api(`/api/receptionist/sessions/${sessionId}/change`, { method: 'POST', body: JSON.stringify({ checkpoint_id: activeCheckpoint, change_text: text }) });
    setMessages(m => [...m, { role: 'user', text: `Change: ${text}` }]);
    setMessages(m => [...m, { role: 'secretary', text: `Change note kar liya, isi checkpoint se continue karte hain.` }]);
  };

  const rateCheckpoint = async (stars: number) => {
    if (!sessionId || !activeCheckpoint) return;
    const summaryForCommit = messages.slice(-4).map(x => x.text).join(' | ').slice(0, 600);
    const d = await api(`/api/receptionist/sessions/${sessionId}/checkpoints/${activeCheckpoint}/rate`, { method: 'POST', body: JSON.stringify({ stars, summary: summaryForCommit }) });
    if (d.committed) {
      setMessages(m => [...m, { role: 'secretary', text: `Rating ${stars} >3 — committed. Next checkpoint.` }]);
      const s = await api(`/api/receptionist/sessions/${sessionId}`);
      setCheckpoints(s.checkpoints || []);
      const next = (s.checkpoints || []).find((c: any) => c.status === 'IN_PROGRESS' || c.status === 'PENDING');
      if (next) setActiveCheckpoint(next.id);
      else { setPhase('PHASE_3_WRAPUP'); setMessages(m => [...m, { role: 'secretary', text: `Sab checkpoints done. Kuch aur add karna hai ya Done dabau?` }]); }
    } else {
      setMessages(m => [...m, { role: 'secretary', text: `Rating ${stars} ≤3 — is point ko dobara karte hain, batao kya kami lagi?` }]);
    }
  };

  const handleDone = async () => {
    if (!sessionId) return;
    await api(`/api/receptionist/sessions/${sessionId}/done`, { method: 'POST' });
    setPhase('PHASE_4_SUMMARY_LOOP');
    const d = await api(`/api/receptionist/sessions/${sessionId}/summary`);
    setSummary(d.summary || '');
    setMessages(m => [...m, { role: 'secretary', text: `Done pressed. Summary ready — neeche padho aur rating do.` }]);
  };

  const handleAddMore = async (texts: string[]) => {
    if (!sessionId) return;
    await api(`/api/receptionist/sessions/${sessionId}/add-checkpoints`, { method: 'POST', body: JSON.stringify({ texts }) });
    const s = await api(`/api/receptionist/sessions/${sessionId}`);
    setCheckpoints(s.checkpoints || []);
    setPhase('PHASE_2_CONVERSATION');
  };

  const rateSummary = async (stars: number) => {
    if (!sessionId) return;
    const d = await api(`/api/receptionist/sessions/${sessionId}/summary/rate`, { method: 'POST', body: JSON.stringify({ stars, feedback: stars <= 3 ? 'needs improvement' : '' }) });
    if (d.approved) {
      setSummaryApproved(true);
      setMessages(m => [...m, { role: 'secretary', text: `Summary rating ${stars} >3 — Approve enabled.` }]);
    } else {
      const nd = await api(`/api/receptionist/sessions/${sessionId}/summary`);
      setSummary(nd.summary || summary);
      setMessages(m => [...m, { role: 'secretary', text: `Rating ${stars} ≤3 — batao kahan sudhar karna hai, phir se summary banaunga.` }]);
    }
  };

  const handleApprove = async () => {
    if (!sessionId) return;
    const d = await api(`/api/receptionist/sessions/${sessionId}/approve`, { method: 'POST' });
    setFinalDoc(d.final_response || '');
    setPhase('COMPLETED');
  };

  const handleDownload = () => { if (!sessionId) return; window.open(`/api/receptionist/sessions/${sessionId}/download`, '_blank'); };
  const handleDelete = async () => { if (!sessionId) return; await api(`/api/receptionist/sessions/${sessionId}`, { method: 'DELETE' }); setSessionId(null); setPhase('INIT'); setFinalDoc(''); setSummary(''); setMessages([]); };

  return (
    <div className="max-w-[1100px] mx-auto p-4 font-sans text-gray-200">
      <h2 className="m-0 text-cyan-400 font-bold">🧑💼 Receptionist — Brilliant Secretary (Admin Only)</h2>
      <div className={`text-xs mt-1 ${connected ? 'text-green-500 font-medium' : 'text-red-500'}`}>
        {connected ? '🔴 Live' : '○ Connecting...'}
      </div>

      {phase === 'INIT' && (
        <div className="mt-4 p-4 border border-cyan-800/50 bg-cyan-950/20 rounded-xl shadow-sm">
          <h3 className="text-cyan-100 font-semibold mb-2">Setup — Language + Response Type + Agenda</h3>
          <div className="flex gap-3 flex-wrap mt-2">
            <label className="text-cyan-200">Language <select value={lang} onChange={e => setLang(e.target.value as Lang)} className="ml-2 p-1.5 rounded-lg bg-gray-900 border border-cyan-800 text-cyan-100"><option value="EN">English</option><option value="HI">Hindi</option><option value="IN_EN">Indian-English</option><option value="HINGLISH">Hinglish</option></select></label>
            <label className="flex-1 text-cyan-200">Response Type (Character Lock) <input value={persona} onChange={e => setPersona(e.target.value)} placeholder="software engineer" className="w-full mt-1 p-2 border border-cyan-800 bg-gray-900 rounded-lg text-cyan-100 placeholder-cyan-800" /></label>
          </div>
          <textarea value={agenda} onChange={e => setAgenda(e.target.value)} placeholder="Agenda — final goal kya hai? + Requirement likho" rows={3} className="w-full mt-3 p-2 border border-cyan-800 bg-gray-900 rounded-lg text-cyan-100 placeholder-cyan-800" />
          <button onClick={createSession} disabled={creating} className={`mt-3 px-4 py-2 rounded-lg transition-colors shadow-md shadow-cyan-900/20 font-medium ${creating ? 'bg-cyan-800 text-cyan-300 cursor-not-allowed' : 'bg-cyan-600 hover:bg-cyan-500 text-white'}`}>{creating ? 'Creating...' : 'Create Session'}</button>
        </div>
      )}

      {sessionId && (phase === 'PHASE_1_SETUP' || phase === 'PHASE_2_CONVERSATION' || phase === 'PHASE_3_WRAPUP') && (
        <div className="mt-3 flex gap-2 flex-wrap items-center">
          <button onClick={() => setCheckOpen(true)} className="px-3 py-1.5 rounded-lg border border-cyan-700 bg-gray-900 hover:bg-cyan-900/30 text-cyan-300 transition-colors">📋 Checklist</button>
          <span className="text-xs text-cyan-600/70">{checkpoints.length ? `${checkpoints.filter((c:any)=>c.status==='COMMITTED').length}/${checkpoints.length} committed` : 'No checklist yet'}</span>
          {checkpoints.length > 0 && activeCheckpoint && (() => {
            const cp = checkpoints.find((c:any)=>c.id===activeCheckpoint);
            if (!cp) return null;
            return (
              <span className="text-xs bg-cyan-900/30 border border-cyan-800/50 px-2 py-1 rounded-lg flex items-center gap-1.5 text-cyan-200">
                Active: {cp.text?.slice(0,40)}
                {cp.status === 'reopened' && <span className="bg-orange-500/20 text-orange-400 border border-orange-500/30 px-1.5 py-0.5 rounded text-[10px]">REOPENED v{cp.version} - HOT Restored</span>}
                {cp.status === 'amended' && <span className="bg-red-500/20 text-red-400 border border-red-500/30 px-1.5 py-0.5 rounded text-[10px]">AMENDED v{cp.version}</span>}
                {cp.status === 'committed' && <span className="bg-green-500/20 text-green-400 border border-green-500/30 px-1.5 py-0.5 rounded text-[10px]">COMMITTED</span>}
              </span>
            );
          })()}
        </div>
      )}

      {sessionId && !isPhase2 && phase !== 'INIT' && phase !== 'COMPLETED' && (
        <div className="bg-yellow-900/20 border border-yellow-700/50 p-3 rounded-lg text-sm text-yellow-500 mt-3 shadow-sm">
          ⚠️ Checklist abhi ready nahi hai. Pehle points ko approve karo, tabhi expert se baat hogi.
        </div>
      )}

      {sessionId && (
        <div className="mt-3 border border-cyan-800/50 rounded-xl overflow-hidden flex flex-col h-[520px] bg-gray-950/80 shadow-lg shadow-cyan-900/10">
          <div className="flex-1 overflow-y-auto p-3 flex flex-col gap-2">
            {messages.map((m, i) => (
              <div key={i} className={`self-${m.role === 'user' ? 'end' : 'start'} p-2.5 px-3.5 rounded-xl max-w-[80%] text-[13px] whitespace-pre-wrap ${m.role === 'user' ? 'bg-cyan-700 text-white rounded-br-sm' : 'bg-gray-800 text-cyan-50 border border-cyan-900/50 rounded-bl-sm'}`}>{m.text}</div>
            ))}
            {live?.notes && <div className="text-[11px] text-cyan-700/60 mt-1">Live notes: {live.notes.length}</div>}
          </div>

          <div className="flex gap-2 p-2 border-t border-cyan-900/50 flex-wrap bg-gray-900 items-center">
            <button onClick={() => setAskOpen(true)} disabled={!canAskExperts} title={!canAskExperts ? "CHECKLIST_NOT_READY: Pehle Phase 1 setup complete karo" : "Experts se baat karo"} className={`px-2.5 py-1.5 rounded-lg border text-sm transition-colors ${canAskExperts ? 'bg-gray-800 hover:bg-gray-700 border-gray-700 text-gray-300' : 'bg-gray-900 border-gray-800 text-gray-600 cursor-not-allowed'}`}>Ask-Experts</button>
            <button onClick={() => setChangeOpen(true)} className="px-2.5 py-1.5 rounded-lg bg-gray-800 hover:bg-gray-700 border border-gray-700 text-gray-300 text-sm transition-colors">Change</button>
            <button onClick={() => { const a=prompt('Agenda edit', agenda); if(a!==null) setAgenda(a); }} className="px-2.5 py-1.5 rounded-lg bg-gray-800 hover:bg-gray-700 border border-gray-700 text-gray-300 text-sm transition-colors">Agenda</button>
            <button onClick={() => setCheckOpen(true)} className="px-2.5 py-1.5 rounded-lg bg-gray-800 hover:bg-gray-700 border border-gray-700 text-gray-300 text-sm transition-colors">Checklist</button>
            <button onClick={handleDone} className="px-3 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-white font-medium text-sm transition-colors shadow-sm">Done</button>
            <button onClick={async () => { const d=await api(`/api/receptionist/sessions/${sessionId}/summary`); setSummary(d.summary||''); }} className="px-2.5 py-1.5 rounded-lg border border-cyan-700 hover:bg-cyan-900/30 text-cyan-300 text-sm transition-colors">Summary</button>
            <input ref={fileRef} type="file" accept=".pdf,.md,.png,.jpg,.jpeg" onChange={handleFileUpload} className="text-xs text-cyan-500 file:mr-2 file:py-1.5 file:px-3 file:rounded-lg file:border-0 file:text-xs file:bg-gray-800 file:text-cyan-300 hover:file:bg-gray-700" />
          </div>

          <div className="flex gap-2 p-2 border-t border-cyan-900/50 bg-gray-900">
            <input value={input} onChange={e => setInput(e.target.value)} onKeyDown={e => e.key === 'Enter' && isSendEnabled && sendChat()} disabled={!canChat} placeholder={canChat ? (activeCheckpoint ? 'Type message for active checkpoint...' : 'No active checkpoint...') : 'Chat Phase 2 me hi khulega'} title={!canChat ? "CHECKLIST_NOT_READY" : ""} className={`flex-1 p-2.5 border rounded-lg focus:outline-none ${canChat ? 'border-cyan-800 bg-gray-950 text-cyan-100 placeholder-cyan-800/60 focus:border-cyan-600' : 'border-gray-800 bg-gray-900 text-gray-500 cursor-not-allowed'}`} />
            <button type="button" onClick={sendChat} disabled={!isSendEnabled} className={`px-5 py-2.5 rounded-lg font-medium transition-colors ${isSendEnabled ? 'bg-cyan-600 hover:bg-cyan-500 text-white cursor-pointer shadow-md shadow-cyan-900/30' : 'bg-gray-800 text-gray-500 cursor-not-allowed'}`}>Send</button>
          </div>

          {expertResponses.length > 0 && phase === 'PHASE_2_CONVERSATION' && (
            <div className="p-3 border-t border-cyan-900/50 bg-gray-900 flex flex-col gap-3 max-h-[300px] overflow-y-auto">
              {expertResponses.map(er => (
                <div key={er.id} className="bg-gray-950 border border-cyan-800/50 p-3 rounded-lg shadow-sm">
                  <div className="text-cyan-300 font-bold mb-2 flex justify-between items-center">
                    <span>{er.name}</span>
                    <div className="flex gap-1">
                      {[1,2,3,4,5].map(s => <button key={s} disabled={!isPhase2} onClick={() => {
                        rateCheckpoint(s);
                        setExpertResponses(prev => prev.map(p => p.id === er.id ? {...p, rating: s} : p));
                      }} className={`px-2 py-0.5 rounded text-xs border ${er.rating===s ? 'bg-cyan-600 text-white' : 'border-cyan-700 text-cyan-400 hover:bg-cyan-800 transition-colors'}`}>{s}★</button>)}
                    </div>
                  </div>
                  <pre className="text-[12.5px] text-cyan-50 whitespace-pre-wrap font-sans leading-relaxed">{er.markdown}</pre>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {phase === 'PHASE_3_WRAPUP' && (
        <div className="mt-4 p-4 border border-cyan-500/50 bg-cyan-950/20 rounded-xl">
          <b className="text-cyan-300">Wrap-up:</b> <span className="text-cyan-100">Sab checkpoints done. Kuch aur add karna hai?</span>
          <div className="flex gap-2 mt-3">
            <button onClick={() => { const t=prompt('New checkpoint text?'); if(t) handleAddMore([t]); }} className="px-3 py-1.5 rounded-lg border border-cyan-700 hover:bg-cyan-900/30 text-cyan-300 transition-colors">+ Add More</button>
            <button onClick={handleDone} className="px-3 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-white transition-colors">No, Done → Summary</button>
          </div>
        </div>
      )}

      {(phase === 'PHASE_4_SUMMARY_LOOP' || summary) && (
        <div className="mt-4 p-4 border border-cyan-800/50 rounded-xl bg-gray-900 shadow-lg">
          <h3 className="m-0 text-cyan-100 font-semibold">Summary (aapki bhasha me)</h3>
          <pre className="whitespace-pre-wrap text-xs bg-gray-950 p-4 rounded-lg max-h-[260px] overflow-y-auto border border-cyan-900/50 text-cyan-50 mt-3 font-mono leading-relaxed">{summary || 'Generating...'}</pre>
          <div className="flex gap-2 mt-3 items-center flex-wrap">
            <span className="text-xs text-cyan-300">Rate summary:</span>
            {[1,2,3,4,5].map(s => <button key={s} onClick={() => rateSummary(s)} className="px-2 py-1 rounded-md border border-cyan-700 hover:bg-cyan-800/50 text-cyan-300 text-xs transition-colors">{s}★</button>)}
            <span className="text-[11px] text-cyan-600/70 ml-2">≤3 = feedback → reopen, &gt;3 = Approve enabled</span>
            {summaryApproved && <button onClick={handleApprove} className="ml-3 bg-cyan-600 hover:bg-cyan-500 text-white px-3 py-1.5 rounded-lg text-sm font-medium transition-colors">Approve → Final Response</button>}
          </div>
        </div>
      )}

      {phase === 'COMPLETED' && finalDoc && (
        <div className="mt-4">
          <FinalResponseView finalDoc={finalDoc} onDownload={handleDownload} onDelete={handleDelete} />
        </div>
      )}

      <ChecklistManagerDialog open={checkOpen} onClose={() => setCheckOpen(false)} onSave={handleChecklistSave} />
      <AskExpertsDialog open={askOpen} onClose={() => setAskOpen(false)} onAsk={handleAskExperts} />
      <ChangeDialog open={changeOpen} onClose={() => setChangeOpen(false)} onSubmit={handleChangeSubmit} />
    </div>
  );
}
