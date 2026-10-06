import pathlib, re
root = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1")
p = root / "backend-go/internal/admin/audit_log.go"
txt = p.read_text(encoding="utf-8")
# Fix WHERE id = '000...' -> WHERE key='retrieval_config'
txt2 = txt.replace("WHERE id = '00000000-0000-0000-0000-000000000001'", "WHERE key='retrieval_config'")
txt2 = re.sub(r"WHERE id\s*=\s*'000[^']*'", "WHERE key='retrieval_config'", txt2)
if txt2 != txt:
    p.write_text(txt2, encoding="utf-8")
    print("FIX3 final WHERE patched")
else:
    print("FIX3 no change needed or pattern mismatch")
    for i,l in enumerate(txt.splitlines(),1):
        if "system_settings" in l:
            print(f"{i}: {l}")

# Also fix proof.py unicode bug to allow exit 0
pp = root / "proof.py"
t = pp.read_text(encoding="utf-8")
t = t.replace("print(f\"{'\\u2713 PASS' if ok else '\\u2717 FAIL'} {name}\")", "print(f\"{'PASS' if ok else 'FAIL'} {name}\")")
t = t.replace("'\\u2713 PASS'", "'PASS'")
pp.write_text(t, encoding="utf-8")
print("proof unicode fixed")

# Check imports for system_health and compare
import pathlib as pl
for name in ["system_health.go", "compare.go"]:
    f = root / f"backend-go/internal/admin/{name}"
    tt = f.read_text(encoding="utf-8")
    # print first 600 chars
    start = tt.find("import (")
    print(f"\n--- {name} import block ---")
    print(tt[start:start+800] if start!=-1 else tt[:800])
    # check module mismatch
    if "github.com/AI_AVENGERS" in tt:
        print(f"WARNING {name} has wrong GH import")
        # try to fix by reading go.mod module name
        mod = (root / "backend-go/go.mod").read_text(encoding="utf-8").splitlines()[0]
        print(f"go.mod first line: {mod}")
        # module is ai_avengers/backend -> import should be ai_avengers/backend/internal/training
        mod_name = mod.split()[1] if len(mod.split())>1 else "ai_avengers/backend"
        tt = tt.replace("github.com/AI_AVENGERS/backend-go/internal/training", f"{mod_name}/internal/training")
        tt = tt.replace("github.com/AI_AVENGERS/backend/internal/training", f"{mod_name}/internal/training")
        f.write_text(tt, encoding="utf-8")
        print(f"Fixed import in {name} to {mod_name}/internal/training")
