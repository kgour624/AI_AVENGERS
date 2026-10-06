import pathlib, glob, re

# Detailed migration scan for course_chunks and expert_pages
for path in sorted(glob.glob(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\migrations\*.up.sql")):
    name = pathlib.Path(path).name
    txt = pathlib.Path(path).read_text(encoding="utf-8", errors="ignore")
    if "course_chunks" in txt.lower() or "expert_pages" in txt.lower():
        print(f"\n===== {name} =====")
        print(txt.strip())
        print("===== END =====")

# training inserts
for fname in sorted(glob.glob(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\training\*.go")):
    txt = pathlib.Path(fname).read_text(encoding="utf-8", errors="ignore")
    if "course_chunks" in txt.lower() or "expert_pages" in txt.lower():
        print(f"\n===== TRAINING {pathlib.Path(fname).name} =====")
        for i,l in enumerate(txt.splitlines(),1):
            if "course_chunks" in l.lower() or "expert_pages" in l.lower() or "INSERT INTO" in l or "parent_id" in l.lower() or "is_child" in l.lower() or "chunk_" in l.lower():
                print(f"{i:4d} {l}")
        # also print whole file if small
        if len(txt.splitlines()) < 200:
            print("---FULL---")
            print(txt)

# admin_handler details around retrieval
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\admin_handler.go")
lines = p.read_text(encoding="utf-8", errors="ignore").splitlines()
print(f"\n===== admin_handler retrieval block 4211-4330 =====")
for i in range(4210, 4330):
    print(f"{i+1:4d} {lines[i]}")
print(f"\n===== gate thresholds start 4342-4365 =====")
for i in range(4341, 4370):
    if i < len(lines):
        print(f"{i+1:4d} {lines[i]}")

# main.go route block
p2 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\cmd\server\main.go")
lines2 = p2.read_text(encoding="utf-8", errors="ignore").splitlines()
for i,l in enumerate(lines2,1):
    if "adminGroup" in l and ("GET" in l or "POST" in l or "PUT" in l or "PATCH" in l or "DELETE" in l):
        print(f"{i:4d} {l}")

# frontend api admin.ts tail
p3 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\api\admin.ts")
lines3 = p3.read_text(encoding="utf-8", errors="ignore").splitlines()
print(f"\n===== admin.ts lines {len(lines3)} =====")
for i,l in enumerate(lines3,1):
    if "RetrievalConfig" in l or "retrieval-config" in l:
        print(f"{i:4d} {l}")
print("\n--- last 80 lines ---")
for i in range(max(0,len(lines3)-80), len(lines3)):
    print(f"{i+1:4d} {lines3[i]}")

# AdminRagSettings full dump
p4 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\pages\admin\AdminRagSettings.tsx")
print("\n===== AdminRagSettings.tsx =====")
print(p4.read_text(encoding="utf-8", errors="ignore"))
