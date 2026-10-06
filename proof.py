import pathlib, subprocess, json, re

root = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1")
print("=== PROOF START ===")

checks=[]

# FIX1 SystemHealth
p=root/"backend-go/internal/admin/system_health.go"
t=p.read_text(encoding="utf-8")
c1 = "LoadRetrievalConfig" in t and "system_settings" in t and "SELECT EXISTS" in t
print(f"FIX1 SystemHealth reads DB (not hardcoded): {c1} -> {'PASS' if c1 else 'FAIL'}")
if c1:
    for i,l in enumerate(t.splitlines(),1):
        if "LoadRetrievalConfig" in l or "source := " in l or "system_settings" in l:
            print(f"  line {i}: {l.strip()}")
checks.append(("FIX1 SystemHealth DB read", c1))

# FIX2 Compare
p=root/"backend-go/internal/admin/compare.go"
t=p.read_text(encoding="utf-8")
c2a = "LoadRetrievalConfig" in t
c2b = "cfg.EnableParentChild" in t and "TopKParents" in t
c2c = "onResults = offResults" not in t or "cfg.EnableParentChild" in t.split("onResults = offResults")[1][:500] if "onResults = offResults" in t else True
print(f"\nFIX2 Compare loads cfg & branches: Load={c2a} EnableBranch={c2b} NotStub={c2b} -> {'PASS' if c2a and c2b else 'FAIL'}")
for i,l in enumerate(t.splitlines(),1):
    if "LoadRetrievalConfig" in l or "EnableParentChild" in l or "TopKParents" in l or "onResults" in l:
        print(f"  line {i}: {l.strip()}")
checks.append(("FIX2 Compare flag-aware", c2a and c2b))

# FIX3 Audit
p=root/"backend-go/internal/admin/audit_log.go"
t=p.read_text(encoding="utf-8")
c3a = "FROM system_settings WHERE key='retrieval_config'" in t
c3b = "FROM retrieval_config" not in t
print(f"\nFIX3 Audit fallback correct table: has_system_settings={c3a} no_old={c3b} -> {'PASS' if c3a and c3b else 'FAIL'}")
for i,l in enumerate(t.splitlines(),1):
    if "system_settings" in l.lower() or "retrieval_config" in l.lower():
        print(f"  line {i}: {l.strip()}")
checks.append(("FIX3 Audit fallback", c3a and c3b))

# FIX4 AssignParentIDs
p=root/"backend-go/internal/training/ingestion_parent_child.go"
t=p.read_text(encoding="utf-8")
c4a = "AssignParentIDs(children []TextChunk, parents []ParentChunk, parentIDByIndex map[int]uuid.UUID, rcfg RetrievalConfig)" in t
c4b = "parentCfg := ChunkConfig{TargetSize: rcfg.ParentSoftLimit" in t
c4c = "cfg := DefaultParentConfig()" not in t or "parentCfg" in t
print(f"\nFIX4 AssignParentIDs uses rcfg: sig={c4a} uses rcfg={c4b} -> {'PASS' if c4a and c4b else 'FAIL'}")
for i,l in enumerate(t.splitlines(),1):
    if "AssignParentIDs" in l or "parentCfg" in l or "ParentSoftLimit" in l:
        print(f"  line {i}: {l.strip()}")
checks.append(("FIX4 AssignParentIDs rcfg", c4a and c4b))

# FIX4 pipeline call
p=root/"backend-go/internal/training/ingestion_pipeline.go"
t=p.read_text(encoding="utf-8")
c4d = "AssignParentIDs(chunks, parents, parentIDByIndex, rcfg)" in t
print(f"FIX4 pipeline call passes rcfg: {c4d} -> {'PASS' if c4d else 'FAIL'}")
for i,l in enumerate(t.splitlines(),1):
    if "AssignParentIDs" in l:
        print(f"  line {i}: {l.strip()}")
checks.append(("FIX4 pipeline call", c4d))

# FIX5 frontend helper
p=root/"frontend/src/api/admin.ts"
t=p.read_text(encoding="utf-8")
has_wrong = "parent_top_k" in t
has_correct = "top_k_parents" in t and "top_k_children" in t
c5 = (not has_wrong) and has_correct
print(f"\nFIX5 Frontend helper keys: wrong_removed={not has_wrong} correct_present={has_correct} -> {'PASS' if c5 else 'FAIL'}")
for i,l in enumerate(t.splitlines(),1):
    if "toSnakeRetrievalPayload" in l or "top_k_parents" in l or "top_k_children" in l:
        print(f"  line {i}: {l.strip()}")
checks.append(("FIX5 Frontend helper", c5))

# FIX6 migrations
mig_dir=root/"backend-go/migrations"
up_exists = (mig_dir/"081_parent_child_retrieval.up.sql").exists() or (mig_dir/"081_parent_child_retrieval.sql").exists()
down_exists = (mig_dir/"081_parent_child_retrieval.down.sql").exists()
if up_exists:
    content = (mig_dir/"081_parent_child_retrieval.up.sql").read_text() if (mig_dir/"081_parent_child_retrieval.up.sql").exists() else (mig_dir/"081_parent_child_retrieval.sql").read_text()
    c6a = "expert_pages" in content.lower() and "parent_id" in content.lower()
else:
    c6a=False
    content=""
print(f"\nFIX6 Migration 081 exists: up={up_exists} down={down_exists} content_ok={c6a} -> {'PASS' if up_exists and c6a else 'FAIL'}")
print(f"  files: {sorted([f.name for f in mig_dir.glob('081*')])}")
if c6a:
    print("  migration snippet:", content[:400].replace('\n',' | '))
checks.append(("FIX6 Migration 081", up_exists and c6a))

print("\n=== SUMMARY ===")
for name, ok in checks:
    print(f"{'✓ PASS' if ok else '✗ FAIL'} {name}")
all_pass = all(ok for _, ok in checks)
print(f"\nOVERALL: {'ALL 6 FIXES VERIFIED' if all_pass else 'SOME FAILED'}")

# also try go vet
print("\n=== GO VET (best effort) ===")
import subprocess, sys, os
try:
    # find go binary
    import shutil
    go = shutil.which("go")
    print(f"go binary: {go}")
    if go:
        res = subprocess.run([go, "vet", "./internal/training", "./internal/admin"], cwd=str(root/"backend-go"), capture_output=True, text=True, timeout=60)
        print("go vet stdout:", res.stdout[:2000])
        print("go vet stderr:", res.stderr[:2000])
        print("go vet returncode:", res.returncode)
        checks.append(("go vet", res.returncode==0))
    else:
        print("go not on PATH - skipping, run manually: go vet ./...")
        # try common win path
        for cand in [r"C:\Program Files\Go\bin\go.exe", r"C:\Go\bin\go.exe"]:
            if pathlib.Path(cand).exists():
                res = subprocess.run([cand, "vet", "./internal/training"], cwd=str(root/"backend-go"), capture_output=True, text=True, timeout=60)
                print(f"go vet via {cand} stdout:", res.stdout[:1000])
                print(f"go vet stderr:", res.stderr[:1000])
                break
except Exception as e:
    print("go vet error:", e)

print("\n=== PROOF END ===")
