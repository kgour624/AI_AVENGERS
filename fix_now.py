import pathlib
p = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1\frontend\src\pages\admin\AdminRagSettings.tsx")
code = r'''import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { getRetrievalConfig, updateRetrievalConfig } from "@/api/admin";

function normalize(raw: any) {
  if (!raw) return {};
  const n: any = { ...raw };
  const epc = raw.enableParentChild ?? raw.enable_parent_child;
  if (epc !== undefined) { n.enable_parent_child = epc; n.enableParentChild = epc; }
  const ptk = raw.parentTopK ?? raw.parent_top_k ?? raw.parentTop_k;
  if (ptk !== undefined) { n.parent_top_k = ptk; n.parentTopK = ptk; }
  const ctk = raw.childTopK ?? raw.child_top_k;
  if (ctk !== undefined) { n.child_top_k = ctk; n.childTopK = ctk; }
  return n;
}

function Toggle({ label, checked, onChange }: { label: string; checked: boolean; onChange: (v:boolean)=>void }) {
  return <label style={{display:"flex",alignItems:"center",gap:8,cursor:"pointer"}}><input type="checkbox" checked={checked} onChange={e=>onChange(e.target.checked)} /><span>{label}</span></label>;
}
function Num({ label, value, onChange }: { label:string; value:number; onChange:(v:number)=>void }) {
  return <label style={{display:"flex",flexDirection:"column",gap:4}}><span>{label}</span><input type="number" value={value} onChange={e=>onChange(Number(e.target.value))} style={{padding:6,borderRadius:6,border:"1px solid #333",background:"#111",color:"#fff"}} /></label>;
}

export default function AdminRagSettings(){
  const [form,setForm]=useState<any>({});
  const [loading,setLoading]=useState(true);
  const [saving,setSaving]=useState(false);
  const [msg,setMsg]=useState("");
  useEffect(()=>{(async()=>{
    try{ const res:any = await getRetrievalConfig(); const raw=res?.data??res; setForm(normalize(raw)); }
    catch(e:any){ setMsg(e?.message||"load failed"); }
    finally{ setLoading(false); }
  })()},[]);
  const get=(c:string,s:string,fb:any)=>(form as any)[c] ?? (form as any)[s] ?? fb;
  const set=(c:string,s:string,v:any)=>setForm((p:any)=>({...p,[c]:v,[s]:v}));
  const handleSave=async()=>{
    setSaving(true); setMsg("");
    try{
      const enabled = Boolean(get("enableParentChild","enable_parent_child",false));
      const parentTopK = Number(get("parentTopK","parent_top_k",5));
      const childTopK = Number(get("childTopK","child_top_k",5));
      const payload:any={
        enable_parent_child: enabled,
        parent_top_k: parentTopK,
        child_top_k: childTopK,
        enableParentChild: enabled,
        parentTopK: parentTopK,
        childTopK: childTopK,
      };
      await updateRetrievalConfig(payload);
      setMsg("Saved \u2713 - retrieval_config updated");
    }catch(e:any){ setMsg(e?.response?.data?.msg||e?.message||"save failed"); }
    finally{ setSaving(false); }
  };
  if(loading) return <div style={{padding:24}}>Loading retrieval config...</div>;
  const enabled=Boolean(get("enableParentChild","enable_parent_child",false));
  const parentTopK=Number(get("parentTopK","parent_top_k",5));
  const childTopK=Number(get("childTopK","child_top_k",5));
  return(
    <div style={{padding:24,maxWidth:720,display:"flex",flexDirection:"column",gap:16}}>
      <h1 style={{fontSize:22,fontWeight:700}}>RAG Settings</h1>
      <Link to="/admin/chunks" style={{color:"#7aa5ff",textDecoration:"underline"}}>-> Open Chunk Explorer (read-only parent-child inspector)</Link>
      <div style={{border:"1px solid #333",borderRadius:12,padding:16,background:"#0f1117",display:"flex",flexDirection:"column",gap:14}}>
        <Toggle label="Enable Parent-Child" checked={enabled} onChange={v=>set("enableParentChild","enable_parent_child",v)} />
        <div style={{fontSize:12,opacity:0.7}}>FLAG {enabled?"ON":"OFF"} - {enabled?"parent-child retrieval active":"default lexical fallback"}</div>
        <Num label="parent_top_k" value={parentTopK} onChange={v=>set("parentTopK","parent_top_k",v)} />
        <Num label="child_top_k" value={childTopK} onChange={v=>set("childTopK","child_top_k",v)} />
        <button onClick={handleSave} disabled={saving} style={{padding:"10px 14px",borderRadius:8,background:saving?"#444":"#4f46e5",color:"#fff",border:"none",cursor:"pointer",fontWeight:600}}>{saving?"Saving...":"Save retrieval config"}</button>
        {msg && <div style={{fontSize:13,color:msg.includes("\u2713")?"#22c55e":"#f87171"}}>{msg}</div>}
      </div>
      <div style={{fontSize:12,opacity:0.6}}>Debug: enableParentChild={String((form as any).enableParentChild)} enable_parent_child={String((form as any).enable_parent_child)}</div>
    </div>
  );
}
'''
p.write_text(code, encoding='utf-8')
print(f"OVERWRITTEN {p} size={p.stat().st_size}")

# also patch admin.ts to make api resilient - write without reading, just append-safe patch via python replace if exists
import pathlib as pl2
q = pl2.Path(r"c:\Users\sharm\AI_AVENGERS-1\frontend\src\api\admin.ts")
if q.exists():
    txt = q.read_text(encoding='utf-8')
    # ensure updateRetrievalConfig sends snake keys - if function exists, we keep it but add dual payload helper comment
    # we do minimal safe append: add helper at end if not present
    if "toSnakeRetrievalPayload" not in txt:
        txt = txt + r'''

// --- FIX: camelCase trap resilient helper ---
export function toSnakeRetrievalPayload(form:any){
  return {
    enable_parent_child: Boolean(form.enableParentChild ?? form.enable_parent_child ?? false),
    parent_top_k: Number(form.parentTopK ?? form.parent_top_k ?? 5),
    child_top_k: Number(form.childTopK ?? form.child_top_k ?? 5),
  };
}
'''
        q.write_text(txt, encoding='utf-8')
        print(f"PATCHED admin.ts size={q.stat().st_size}")
    else:
        print("admin.ts already patched")
else:
    print("admin.ts not found")
