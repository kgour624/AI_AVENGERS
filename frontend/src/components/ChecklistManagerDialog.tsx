import { useState } from 'react';

type Props = {
  open: boolean;
  onClose: () => void;
  onSave: (texts: string[], noChecklist: boolean) => void;
};

export default function ChecklistManagerDialog({ open, onClose, onSave }: Props) {
  const [items, setItems] = useState<string[]>(['PRD document', 'Design Document']);
  const [noChecklist, setNoChecklist] = useState(false);

  if (!open) return null;

  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-[1000] p-4">
      <div className="bg-gray-900 border border-cyan-800/60 p-6 rounded-2xl w-full max-w-[580px] max-h-[85vh] flex flex-col shadow-2xl shadow-cyan-900/30">
        <h3 className="m-0 text-cyan-400 font-bold text-lg flex items-center gap-2">
          📋 Checklist Manager (CRUD)
        </h3>
        <p className="text-xs text-cyan-200/70 mt-1 mb-4">
          Har point final response me consider hoga. Points add/remove karke Save karo.
        </p>

        <label className="flex items-center gap-3.5 p-3 rounded-xl border border-cyan-800/50 bg-gray-950/60 cursor-pointer hover:border-cyan-700/70 transition-colors">
          <input 
            type="checkbox" 
            checked={noChecklist} 
            onChange={e => setNoChecklist(e.target.checked)}
            className="accent-cyan-500 w-4 h-4 rounded cursor-pointer"
          />
          <span className="text-sm font-medium text-cyan-100">
            No-checklist (single General Discussion flow)
          </span>
        </label>

        {!noChecklist && (
          <div className="mt-4 flex-1 overflow-y-auto space-y-2.5 pr-1">
            {items.map((v, i) => (
              <div key={i} className="flex gap-2 items-center">
                <input 
                  value={v} 
                  onChange={e => { const c=[...items]; c[i]=e.target.value; setItems(c); }} 
                  placeholder={`Checkpoint ${i+1}`} 
                  className="flex-1 p-2.5 bg-gray-950 border border-cyan-800/50 rounded-xl text-cyan-100 placeholder-cyan-800/60 text-sm focus:outline-none focus:border-cyan-500 transition-colors"
                />
                <button 
                  onClick={() => setItems(items.filter((_, idx) => idx !== i))} 
                  className="p-2.5 text-xs text-red-400 hover:text-red-300 hover:bg-red-950/40 border border-red-900/30 rounded-xl transition-colors font-bold"
                  title="Remove point"
                >
                  ✕
                </button>
              </div>
            ))}
            <button 
              onClick={() => setItems([...items, ''])} 
              className="w-full mt-2 py-2 border border-dashed border-cyan-800/70 hover:border-cyan-600 rounded-xl text-cyan-400 hover:text-cyan-300 text-xs font-medium transition-colors bg-cyan-950/20"
            >
              + Add Checkpoint Point
            </button>
          </div>
        )}

        <div className="flex gap-3 mt-6 justify-end items-center border-t border-cyan-900/40 pt-4">
          <button 
            onClick={onClose} 
            className="px-4 py-2 rounded-xl border border-cyan-800/50 text-cyan-300 hover:bg-cyan-950/50 text-sm font-medium transition-colors"
          >
            Cancel
          </button>
          <button 
            onClick={() => { const clean = items.map(s=>s.trim()).filter(Boolean); onSave(clean, noChecklist); }} 
            className="bg-cyan-600 hover:bg-cyan-500 text-white px-5 py-2 rounded-xl text-sm font-semibold transition-all shadow-md shadow-cyan-900/30"
          >
            Save Checklist
          </button>
        </div>
      </div>
    </div>
  );
}
