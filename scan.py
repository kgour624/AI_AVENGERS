import pathlib, re, glob
migs = sorted(glob.glob(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\migrations\*.up.sql"))
for path in migs:
    txt = pathlib.Path(path).read_text(encoding="utf-8", errors="ignore")
    if "course_chunks" in txt.lower() or "expert_pages" in txt.lower():
        print(f"\n=== {pathlib.Path(path).name} ===")
        # extract relevant lines
        for line in txt.splitlines():
            low = line.lower()
            if "course_chunks" in low or "expert_pages" in low:
                print(line.strip())
        # also print full if small
        if pathlib.Path(path).name in ("081_parent_child_retrieval.up.sql","082_parent_child_compat_view.up.sql","003_repo_chunks.up.sql","008_chunk_hash_unique.up.sql"):
            print("---FULL---")
            print(txt)
            print("---END FULL---")

# also check admin_handler imports and existing pattern for List
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\admin_handler.go")
txt = p.read_text(encoding="utf-8", errors="ignore")
# find a simple list handler example
import re
for name in ["ListExpertCategories","GetExpertCategory","ListGateThresholds"]:
    idx = txt.find(f"func (h *AdminHandler) {name}")
    if idx != -1:
        print(f"\n=== HANDLER {name} ===")
        print(txt[idx:idx+2500])

# check admin.ts patterns
p2 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\api\admin.ts")
txt2 = p2.read_text(encoding="utf-8", errors="ignore")
# print around RetrievalConfig
for i, line in enumerate(txt2.splitlines()):
    if "retrieval" in line.lower() or "RetrievalConfig" in line:
        print(f"{i+1:4d} {line}")
print("\n=== TAIL admin.ts last 200 lines ===")
lines = txt2.splitlines()
for i,l in enumerate(lines[-200:], start=len(lines)-199):
    print(f"{i:4d} {l}")

# check AdminRagSettings lines
p3 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\pages\admin\AdminRagSettings.tsx")
print("\n=== AdminRagSettings full ===")
print(p3.read_text(encoding="utf-8", errors="ignore"))
