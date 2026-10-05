import { useEffect, useState } from 'react';
import { baseAPI } from '@/api/base';

type Expert = { id: string; name: string; domain: string; description?: string; totalChunks?: number };

type Props = {
  open: boolean;
  onClose: () => void;
  onAsk: (expertIds: string[], question: string) => void;
};

export default function AskExpertsDialog({ open, onClose, onAsk }: Props) {
  const [experts, setExperts] = useState<Expert[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [question, setQuestion] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!open) return;
    setLoading(true);
    setError('');
    // FIX: /api/experts -> /api/v1/experts + baseAPI (JWT + vite proxy api:8080)
    // Pehle raw fetch('/api/experts') => 404 + no Authorization => hamesha [] => "No experts found"
    baseAPI.get('/api/v1/experts')
      .then(res => {
        const raw: any = res.data;
        const list: any[] = Array.isArray(raw?.data) ? raw.data : Array.isArray(raw) ? raw : raw?.experts || [];
        const normalized: Expert[] = list.map((e: any) => ({
          id: e.id,
          name: e.name || e.slug || 'Unknown',
          domain: e.domain || e.category || 'general',
          description: e.description,
          totalChunks: e.totalChunks ?? e.total_chunks,
        }));
        setExperts(normalized);
      })
      .catch((err) => {
        console.error('AskExperts fetch failed', err);
        setError(err?.response?.data?.error || err.message || 'Failed to load experts');
        setExperts([]);
      })
      .finally(() => setLoading(false));
  }, [open]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-[1000] p-4">
      <div className="bg-gray-900 border border-cyan-800/60 p-6 rounded-xl w-full max-w-[640px] max-h-[85vh] flex flex-col shadow-xl shadow-cyan-900/20">
        <h3 className="m-0 text-cyan-400 font-bold text-lg">Ask Experts (English to Expert, Client Language to You)</h3>
        {loading && <div className="p-3 text-cyan-600/70 text-[13px]">Loading domain experts...</div>}
        {error && !loading && <div className="p-2 text-red-400 text-xs bg-red-950/30 border border-red-900/50 rounded-lg mt-2">{error}</div>}

        <div className="mt-4 border border-cyan-800/50 rounded-lg max-h-[220px] overflow-y-auto p-2 bg-gray-950/50">
          {experts.map(ex => (
            <label key={ex.id} className="flex gap-3 p-2.5 border-b border-cyan-900/30 cursor-pointer hover:bg-cyan-900/20 transition-colors rounded-lg group">
              <input type="checkbox" checked={selected.has(ex.id)} onChange={e => { const n=new Set(selected); if(e.target.checked) n.add(ex.id); else n.delete(ex.id); setSelected(n); }} className="accent-cyan-500 mt-0.5" />
              <span className="flex-1">
                <b className="text-cyan-100 group-hover:text-cyan-50 transition-colors">{ex.name}</b> 
                <span className="text-cyan-600/80 text-xs ml-2">({ex.domain})</span> 
                {ex.totalChunks ? <span className="text-cyan-700/60 text-[11px] ml-2">- {ex.totalChunks} chunks</span> : null}
              </span>
            </label>
          ))}
          {!loading && experts.length===0 && <div className="p-3 text-cyan-700/60 text-sm">No experts found</div>}
        </div>

        {selected.size > 0 && (
          <div className="mt-4 bg-cyan-950/20 p-3 rounded-lg border border-cyan-900/30">
            <div className="text-xs text-cyan-300 mb-2">
              {selected.size===1 ? 'Single window' : `${selected.size} windows (concurrent fan-out)`} — Question English me jayega, jawab aapki language me ayega.
            </div>
            <textarea value={question} onChange={e=>setQuestion(e.target.value)} placeholder="Type your question for selected experts..." rows={3} className="w-full p-2.5 border border-cyan-800 bg-gray-950 rounded-lg text-cyan-100 placeholder-cyan-800/60 focus:outline-none focus:border-cyan-600" />
          </div>
        )}

        <div className="flex gap-3 mt-5 justify-end">
          <button onClick={onClose} className="px-4 py-2 rounded-lg text-cyan-300 hover:text-cyan-100 hover:bg-cyan-900/30 transition-colors">Cancel</button>
          <button disabled={selected.size===0 || !question.trim()} onClick={()=> { onAsk(Array.from(selected), question); setQuestion(''); }} className="bg-cyan-600 hover:bg-cyan-500 text-white px-5 py-2 rounded-lg font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-md shadow-cyan-900/20">Ask</button>
        </div>
      </div>
    </div>
  );
}
