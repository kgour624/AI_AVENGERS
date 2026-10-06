import pathlib, glob, re

# scan training dir for INSERT
for path in sorted(glob.glob(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\training\*.go")):
    txt = pathlib.Path(path).read_text(encoding="utf-8", errors="ignore")
    if "INSERT INTO" in txt:
        print(f"\n===== {pathlib.Path(path).name} INSERT scan =====")
        for i,l in enumerate(txt.splitlines(),1):
            if "INSERT INTO" in l or "course_chunks" in l.lower() or "expert_pages" in l.lower():
                print(f"{i:4d} {l}")
        # print surrounding block
        import re
        for m in re.finditer(r"INSERT INTO[^\n]*course_chunks.*?(?:\);|String\(\)|args\.\.\.)", txt, re.S):
            print("---MATCH---")
            print(txt[max(0,m.start()-500):m.end()+1500][:4000])
            print("---END MATCH---")

# also check pipeline embed section
for path in sorted(glob.glob(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\training\*.go")):
    name = pathlib.Path(path).name
    txt = pathlib.Path(path).read_text(encoding="utf-8", errors="ignore")
    if "chunk_hash" in txt.lower() or "section_path" in txt.lower():
        print(f"\n===== {name} hash/section mentions =====")
        for i,l in enumerate(txt.splitlines(),1):
            if "chunk_hash" in l.lower() or "section_path" in l.lower() or "parent_id" in l.lower() or "is_child" in l.lower():
                print(f"{i:4d} {l}")

# check migrations for chunk_hash, section_path adds
for path in sorted(glob.glob(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\migrations\*.up.sql")):
    txt = pathlib.Path(path).read_text(encoding="utf-8", errors="ignore")
    low = txt.lower()
    if "chunk_hash" in low or "section_path" in low or "confidence" in low or "is_child" in low:
        print(f"\n===== MIG {pathlib.Path(path).name} =====")
        print(txt.strip()[:5000])
