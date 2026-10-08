import React, { useMemo, useState } from "react";
type Props = { content: string; pageSize?: number; };
export default function VirtualizedAnswer({ content, pageSize=4000 }: Props){
  const pages = useMemo(()=>{
    if(!content) return [""];
    if(content.length <= pageSize) return [content];
    const out:string[]=[]; let s=0;
    while(s<content.length){
      let e=Math.min(s+pageSize, content.length);
      if(e<content.length){ const sl=content.slice(s,e); const lb=sl.lastIndexOf("\n\n"); if(lb>pageSize*0.6) e=s+lb+2; }
      out.push(content.slice(s,e)); s=e;
    }
    return out;
  },[content,pageSize]);
  const [page,setPage]=useState(0);
  if(pages.length===1) return <div className="whitespace-pre-wrap break-words text-sm" data-testid="virtualized-answer-single">{content}</div>;
  return (
    <div data-testid="virtualized-answer-paged">
      <div className="flex items-center justify-between mb-2 text-xs text-gray-500"><span>Bada jawab - page {page+1}/{pages.length}</span><span>{content.length.toLocaleString()} chars</span></div>
      <div className="whitespace-pre-wrap break-words border rounded p-3 bg-white max-h-[60vh] overflow-auto text-sm">{pages[page]}</div>
      <div className="mt-2 flex gap-2"><button disabled={page===0} onClick={()=>setPage(p=>Math.max(0,p-1))} className="px-3 py-1 text-sm rounded border bg-white disabled:opacity-50">Prev</button><button disabled={page===pages.length-1} onClick={()=>setPage(p=>Math.min(pages.length-1,p+1))} className="px-3 py-1 text-sm rounded border bg-white disabled:opacity-50">Next</button></div>
    </div>
  );
}
