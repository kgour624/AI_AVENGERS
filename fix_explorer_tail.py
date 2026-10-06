import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\pages\admin\AdminChunksExplorer.tsx")
txt = p.read_text(encoding="utf-8", errors="ignore")
print("before tail", txt[-600:])
# remove misplaced insertion outside function
txt = txt.replace("{expertId ? <ComparePanel expertId={expertId}/> : null}<AuditLogList/>", "")
# now inject inside return before final </div>)} 
# compact file ends with ...</>)} </div>)}\n export...
old = "</>)}"
new = "</>)}\n      {expertId ? <ComparePanel expertId={expertId}/> : null}\n      <AuditLogList/>\n"
# find the outer closing pattern:  )}</div> vs </div>)}
if "</div>)" in txt:
    # insert before the last </div>)
    # locate last occurrence
    idx = txt.rfind("</div>)")
    if idx != -1:
        before = txt[:idx]
        after = txt[idx:]
        # check if already has ComparePanel inside
        if "ComparePanel" not in before[-500:]:
            txt = before + "      {expertId ? <ComparePanel expertId={expertId}/> : null}\n      <AuditLogList/>\n      " + after
            print("injected inside")
        else:
            print("already inside")
else:
    print("no </div>) found")
# also ensure SystemHealthCard placement is correct — it was inserted after title div, check
if "SystemHealthCard" in txt:
    print("has SystemHealthCard")
else:
    print("missing SystemHealthCard")
# ensure imports include SystemHealthCard etc
if "ComparePanel" not in txt:
    print("missing ComparePanel import, need fix imports")
    # re-add imports if needed
    if "compareRetrieval" not in txt.lower():
        txt = txt.replace("from '@/api/admin'", "from '@/api/admin'\nimport { getSystemHealth } from '@/api/admin'")
p.write_text(txt, encoding="utf-8")
print("after tail", txt[-800:])
print("lines", len(txt.splitlines()), "len", len(txt))
