import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\router_test.go")
txt = p.read_text(encoding="utf-8", errors="ignore")
lines = txt.splitlines()
for i in range(58, 75):
    print(f"{i+1:3d} {repr(lines[i])}")
# Fix corrupted chunk-tree block
old_snippet = '    group.GET("/experts/:id/chunk-tree"'
if old_snippet in txt:
    # check if already broken (missing , noop)
    if 'group.GET("/experts/:id/chunk-tree", noop)' not in txt:
        txt = txt.replace(
            '    group.GET("/experts/:id/chunk-tree"\n    group.GET("/system-health", func(c *gin.Context) { c.Status(200) })\n    group.POST("/retrieval-compare", func(c *gin.Context) { c.Status(200) })\n    group.GET("/audit-log", func(c *gin.Context) { c.Status(200) }), noop)',
            '    group.GET("/experts/:id/chunk-tree", noop)\n    group.GET("/system-health", noop)\n    group.POST("/retrieval-compare", noop)\n    group.GET("/audit-log", noop)'
        )
        p.write_text(txt, encoding="utf-8")
        print("fixed router_test block")
    else:
        print("already correct chunk-tree line")
else:
    print("old_snippet not found")
# verify
txt2 = p.read_text(encoding="utf-8", errors="ignore")
for i,l in enumerate(txt2.splitlines(),1):
    if "system-health" in l or "chunk-tree" in l or "audit-log" in l or "retrieval-compare" in l:
        print(f"{i:3d} {l}")
print("want 25?", "want 25" in txt2, "total group.", txt2.count("group."))
