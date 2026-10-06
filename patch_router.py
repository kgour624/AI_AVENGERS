import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\router_test.go")
txt = p.read_text(encoding="utf-8", errors="ignore")
# add 3 chunk explorer routes before count assertion
old = '        // Phase 1.5 RAG retrieval config.\n        group.GET("/retrieval-config", noop)\n        group.PUT("/retrieval-config", noop)'
new = '        // Phase 1.5 RAG retrieval config.\n        group.GET("/retrieval-config", noop)\n        group.PUT("/retrieval-config", noop)\n\n        // Phase 1.5 Chunk Explorer (observability).\n        group.GET("/experts/:id/chunks", noop)\n        group.GET("/experts/:id/parents", noop)\n        group.GET("/experts/:id/chunk-tree", noop)'
if old in txt:
    txt = txt.replace(old, new)
# update count 19 -> 22
txt = txt.replace('want 19', 'want 22')
txt = txt.replace('got != 19', 'got != 22')
p.write_text(txt, encoding="utf-8")
print(txt)
