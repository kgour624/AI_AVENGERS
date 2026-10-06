import pathlib, glob, re

# scan all migrations for course_chunks alters
print("=== ALL course_chunks ALTERS ===")
for path in sorted(glob.glob(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\migrations\*.up.sql")):
    txt = pathlib.Path(path).read_text(encoding="utf-8", errors="ignore")
    if "course_chunks" in txt:
        name = pathlib.Path(path).name
        lines = [l.strip() for l in txt.splitlines() if "course_chunks" in l.lower() or "course_chunk" in l.lower()]
        print(f"\n{name}:")
        for l in lines:
            print("  ", l)
        # also if small file print full
        if len(txt.splitlines()) < 60:
            print("--- full ---")
            print(txt)
            print("--- end ---")

# check current course_chunks schema expected after all migrations
print("\n=== CHECK 001 + 081 + other course_chunks columns expected ===")
# read relevant migrations 003 etc
for target in ["003_repo_chunks.up.sql","008_chunk_hash_unique.up.sql","081_parent_child_retrieval.up.sql","082_parent_child_compat_view.up.sql"]:
    p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\migrations") / target
    if p.exists():
        print(f"\n--- {target} ---")
        print(p.read_text(encoding="utf-8", errors="ignore"))

# check training chunker/child store for column usage
for fname in ["chunker.go","child_store.go"]:
    p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\training") / fname
    if p.exists():
        txt = p.read_text(encoding="utf-8", errors="ignore")
        print(f"\n=== {fname} mentions ===")
        for i,l in enumerate(txt.splitlines(),1):
            if "course_chunks" in l.lower() or "expert_pages" in l.lower() or "parent_id" in l.lower() or "chunk_" in l.lower():
                print(f"{i:4d} {l}")

# dump admin_handler retrieval section fully
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\admin_handler.go")
lines = p.read_text(encoding="utf-8", errors="ignore").splitlines()
print(f"\n=== admin_handler total {len(lines)} lines, dump 4210-4461 ===")
for i in range(4209, 4461):
    print(f"{i+1:4d} {lines[i]}")

# dump frontend admin.ts around retrieval and tail
p2 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\api\admin.ts")
lines2 = p2.read_text(encoding="utf-8", errors="ignore").splitlines()
print(f"\n=== admin.ts total {len(lines2)} ===")
for i,l in enumerate(lines2):
    if "RetrievalConfig" in l or "retrieval" in l.lower():
        print(f"{i+1:4d} {l}")
print("\n--- last 250 lines ---")
for i in range(max(0,len(lines2)-250), len(lines2)):
    print(f"{i+1:4d} {lines2[i]}")

# dump AdminRagSettings
p3 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\pages\admin\AdminRagSettings.tsx")
txt3 = p3.read_text(encoding="utf-8", errors="ignore")
print(f"\n=== AdminRagSettings.tsx {len(txt3.splitlines())} lines ===")
print(txt3)

# dump response helpers
p4 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\response")
print("\n=== response dir ===")
print(list(p4.glob("*.go")))
if (p4 / "response.go").exists():
    print((p4 / "response.go").read_text(encoding="utf-8", errors="ignore")[:3000])
