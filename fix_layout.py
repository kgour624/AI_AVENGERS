import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\pages\admin\AdminLayout.tsx")
txt = p.read_text(encoding="utf-8", errors="ignore")
needle = "{ to: '/admin/rag',             label: 'RAG Retrieval',   end: false },"
if needle in txt:
    txt = txt.replace(needle, needle + "\n  { to: '/admin/chunks',          label: 'Chunk Explorer',  end: false },")
    p.write_text(txt, encoding="utf-8")
    print("patched layout nav")
else:
    print("needle not found")
    for i,l in enumerate(txt.splitlines(),1):
        if "rag" in l.lower():
            print(i, repr(l))
