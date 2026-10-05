import { useState } from 'react';

export default function ChangeDialog({ open, onClose, onSubmit }: { open: boolean; onClose: () => void; onSubmit: (text: string) => void }) {
  const [text, setText] = useState('');
  if (!open) return null;
  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-[1000] p-4">
      <div className="bg-gray-900 border border-cyan-800/60 p-6 rounded-2xl w-full max-w-[540px] flex flex-col shadow-2xl shadow-cyan-900/30">
        <h3 className="m-0 text-cyan-400 font-bold text-lg flex items-center gap-2">
          🔄 Change Request
        </h3>
        <p className="text-xs text-cyan-200/70 mt-1 mb-3">
          Jo galat laga usko yahan likho, secretary isi checkpoint se continue karegi (resume pointer).
        </p>
        <textarea 
          value={text} 
          onChange={e => setText(e.target.value)} 
          rows={3} 
          className="w-full mt-2 p-3 bg-gray-950 border border-cyan-800/50 rounded-xl text-cyan-100 placeholder-cyan-800/60 text-sm focus:outline-none focus:border-cyan-500 transition-colors"
          placeholder="Kya change chahiye?" 
        />
        <div className="flex gap-3 mt-5 justify-end items-center border-t border-cyan-900/40 pt-4">
          <button 
            onClick={onClose} 
            className="px-4 py-2 rounded-xl border border-cyan-800/50 text-cyan-300 hover:bg-cyan-950/50 text-sm font-medium transition-colors"
          >
            Cancel
          </button>
          <button 
            onClick={() => { onSubmit(text); setText(''); }} 
            disabled={!text.trim()} 
            className={`px-5 py-2 rounded-xl text-sm font-semibold transition-all shadow-md ${text.trim() ? 'bg-cyan-600 hover:bg-cyan-500 text-white shadow-cyan-900/30' : 'bg-gray-800 text-gray-500 cursor-not-allowed'}`}
          >
            Submit Change
          </button>
        </div>
      </div>
    </div>
  );
}
