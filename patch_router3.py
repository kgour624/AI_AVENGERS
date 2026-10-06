import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\router_test.go")
txt = p.read_text(encoding="utf-8", errors="ignore")
# file has no blank double newline between retrieval and Route count now; search exact
old = 'group.PUT("/retrieval-config", noop)\n\n\t// Route count'
new = 'group.PUT("/retrieval-config", noop)\n\n\t// Phase 1.5 Chunk Explorer (observability).\n\tgroup.GET("/experts/:id/chunks", noop)\n\tgroup.GET("/experts/:id/parents", noop)\n\tgroup.GET("/experts/:id/chunk-tree", noop)\n\n\t// Route count'
if old in txt:
    txt = txt.replace(old, new)
    p.write_text(txt, encoding="utf-8")
    print("patched")
else:
    # try without tab prefix
    print("still not found; dump bytes around")
    import re
    m = re.search(r'PUT\(".*"\,.*?\)\s*// Route count', txt, re.S)
    print(m.group(0)[:300] if m else "no match")
    for i,l in enumerate(txt.splitlines(),1):
        if "Route count" in l:
            print(i, repr(txt.splitlines()[i-3]), repr(txt.splitlines()[i-2]), repr(l))
