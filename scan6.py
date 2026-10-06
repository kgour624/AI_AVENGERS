import pathlib, glob, re

# list training files
print(list(glob.glob(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\training\*.go")))
for path in sorted(glob.glob(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\training\*.go")):
    txt = pathlib.Path(path).read_text(encoding="utf-8", errors="ignore")
    if "INSERT INTO" in txt:
        print(f"\n===== {pathlib.Path(path).name} =====")
        for i,l in enumerate(txt.splitlines(),1):
            if "INSERT INTO" in l:
                print(f"{i:4d} {l}")
                # print next 40 lines
                lines = txt.splitlines()
                for j in range(i, min(i+60, len(lines))):
                    print(f"{j+1:4d} {lines[j]}")
                print("---")

# also pipeline files may be ingestion*.go
for path in sorted(glob.glob(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\**\*.go"), recursive=True):
    txt = pathlib.Path(path).read_text(encoding="utf-8", errors="ignore")
    if "INSERT INTO course_chunks" in txt:
        print(f"\n===== FOUND INSERT in {path} =====")
        idx = txt.find("INSERT INTO course_chunks")
        print(txt[max(0,idx-2000):idx+6000])

# check 001 for confidence placement
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\migrations\001_initial_schema.up.sql")
txt = p.read_text(encoding="utf-8", errors="ignore")
# extract course_chunks block
import re
for m in re.finditer(r"CREATE TABLE course_chunks.*?;", txt, re.S|re.I):
    print("\n=== 001 course_chunks ===")
    print(m.group(0))
for m in re.finditer(r"CREATE TABLE expert_pages.*?;", txt, re.S|re.I):
    print("\n=== 001 expert_pages (if exists) ===")
    print(m.group(0))
# check 006 chunk_hash
p6 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\migrations\006_collaboration_layer.up.sql")
txt6 = p6.read_text(encoding="utf-8", errors="ignore")
for line in txt6.splitlines():
    if "chunk_hash" in line.lower() or "course_chunks" in line.lower():
        print(line)

p51 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\migrations\051_chunk_section_path.up.sql")
print("\n===051===")
print(p51.read_text(encoding="utf-8", errors="ignore"))

# dump admin_handler response helpers
pAdmin = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\admin_handler.go")
lines = pAdmin.read_text(encoding="utf-8", errors="ignore").splitlines()
# find ListExperts to see pagination pattern
for kw in ["func (h *AdminHandler) ListExperts", "func (h *AdminHandler) GetExpert", "func (h *AdminHandler) ListIngestion"]:
    for i,l in enumerate(lines,1):
        if kw in l:
            print(f"\n=== {kw} at {i} ===")
            for j in range(i-1, min(i+60, len(lines))):
                print(f"{j+1:4d} {lines[j]}")
