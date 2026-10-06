import pathlib, glob, re

# Check all migrations for columns added to course_chunks by name
cols = []
for path in sorted(glob.glob(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\migrations\*.up.sql")):
    txt = pathlib.Path(path).read_text(encoding="utf-8", errors="ignore")
    if "ALTER TABLE course_chunks" in txt:
        print(f"\n=== {pathlib.Path(path).name} ===")
        for line in txt.splitlines():
            if "course_chunks" in line.lower():
                print(line.strip())

# Also check for any migration that adds chunk_hash, section_path etc elsewhere
print("\n=== SEARCH cols ===")
for path in sorted(glob.glob(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\migrations\*.up.sql")):
    txt = pathlib.Path(path).read_text(encoding="utf-8", errors="ignore")
    for kw in ["chunk_hash","section_path","confidence","token_count","is_child","parent_id","parent_index"]:
        if kw in txt.lower():
            # print file and snippet
            for line in txt.splitlines():
                if kw in line.lower():
                    print(f"{pathlib.Path(path).name}: {line.strip()}")
                    break

# Check ingestion pipeline inserts columns
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\training\pipeline.go")
txt = p.read_text(encoding="utf-8", errors="ignore")
print("\n=== pipeline.go INSERT INTO course_chunks snippet ===")
# find sb.WriteString insert
idx = txt.find("INSERT INTO course_chunks")
print(txt[max(0,idx-2000):idx+6000])
print("\n=== pipeline end scan ===")

# check store/retrieval for selects
for path in sorted(glob.glob(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\training\*.go")):
    txt = pathlib.Path(path).read_text(encoding="utf-8", errors="ignore")
    if "SELECT" in txt and "course_chunks" in txt.lower():
        print(f"\n=== SELECT in {pathlib.Path(path).name} ===")
        for i,l in enumerate(txt.splitlines(),1):
            if "course_chunks" in l.lower() or "expert_pages" in l.lower():
                print(f"{i:4d} {l}")

# check backend-go/internal/retrieval or similar
for path in sorted(glob.glob(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\**\*.go"), recursive=True):
    txt = pathlib.Path(path).read_text(encoding="utf-8", errors="ignore")
    if "expert_pages" in txt.lower() and "SELECT" in txt:
        print(f"\n=== expert_pages SELECT in {path} ===")
        for i,l in enumerate(txt.splitlines(),1):
            if "expert_pages" in l.lower():
                print(f"{i:4d} {l}")
