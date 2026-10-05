import pathlib, os, re
ml = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\ml")
print("ml exists:", ml.exists())
if ml.exists():
    for p in ml.iterdir():
        print(p.name)
    # grep Rerank
    for p in ml.rglob("*.go"):
        t = p.read_text(encoding="utf-8", errors="ignore")
        if "Rerank" in t:
            print("---", p)
            # find function
            m = re.search(r"func.*Rerank.*\n.*", t)
            if m:
                print(m.group(0)[:500])
            # print relevant snippet
            idx = t.find("Rerank")
            print(t[idx-200:idx+600].replace("\n","\\n")[:1200])

# check patch leftover
for name in ["patch2.py","patch_assembler.py"]:
    pp = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1") / name
    print(name, "exists", pp.exists())

# check assembler parent retrieval score handling
p2 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\context\assembler_parent_retrieval.go")
t2 = p2.read_text(encoding="utf-8")
print("--- assembler_parent_retrieval Rerank usage ---")
idx = t2.find("Rerank")
print(t2[idx-200:idx+800][:2000])

# check chunker defaults
p3 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\training\chunker.go")
t3 = p3.read_text(encoding="utf-8")
for kw in ["DefaultParentConfig","DefaultChildConfig","func.*ChunkMarkdown"]:
    m = re.search(kw, t3)
    print(kw, "found", bool(m))
    if m:
        print(t3[m.start()-100:m.start()+300].replace("\n","\\n")[:800])
