import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\router_test.go")
txt = p.read_text(encoding="utf-8", errors="ignore")
# the file currently has count 22 but missing routes; add routes
old = '        group.PUT("/retrieval-config", noop)\n\n        // Route count'
new = '        group.PUT("/retrieval-config", noop)\n\n        // Phase 1.5 Chunk Explorer (observability).\n        group.GET("/experts/:id/chunks", noop)\n        group.GET("/experts/:id/parents", noop)\n        group.GET("/experts/:id/chunk-tree", noop)\n\n        // Route count'
if old in txt:
    txt = txt.replace(old, new)
    p.write_text(txt, encoding="utf-8")
    print("patched router_test with explorer routes")
    print("has chunk-tree", "/chunk-tree" in txt)
    print(txt[txt.find("Phase 1.5"):txt.find("Route count")+500])
else:
    print("old2 not found")
    # debug dump near
    idx = txt.find('retrieval-config')
    print(repr(txt[idx-200:idx+600]))
