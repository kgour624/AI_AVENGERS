import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\cmd\server\main.go")
txt = p.read_text(encoding="utf-8", errors="ignore")
needle = 'adminGroup.PUT("/retrieval-config", adminHandler.UpdateRetrievalConfig)'
if needle not in txt:
    print("needle not found; dumping lines with retrieval")
    for i,l in enumerate(txt.splitlines(),1):
        if "retrieval" in l.lower():
            print(i, repr(l))
    raise SystemExit
repl = needle + '\n                // Chunk Explorer (Phase 1.5 Observability) - read-only parent-child inspector\n                adminGroup.GET("/experts/:id/chunks", adminHandler.ListExpertChunks)\n                adminGroup.GET("/experts/:id/parents", adminHandler.ListExpertParents)\n                adminGroup.GET("/experts/:id/chunk-tree", adminHandler.GetExpertChunkTree)'
txt = txt.replace(needle, repl)
p.write_text(txt, encoding="utf-8")
print("patched main.go")
txt2 = p.read_text(encoding="utf-8")
print("has ListExpertChunks", "ListExpertChunks" in txt2)
