import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\router_test.go")
txt = p.read_text(encoding="utf-8", errors="ignore")
old = "\tgroup.GET(\"/experts/:id/chunk-tree\"\n\tgroup.GET(\"/system-health\", func(c *gin.Context) { c.Status(200) })\n\tgroup.POST(\"/retrieval-compare\", func(c *gin.Context) { c.Status(200) })\n\tgroup.GET(\"/audit-log\", func(c *gin.Context) { c.Status(200) }), noop)"
new = "\tgroup.GET(\"/experts/:id/chunk-tree\", noop)\n\tgroup.GET(\"/system-health\", noop)\n\tgroup.POST(\"/retrieval-compare\", noop)\n\tgroup.GET(\"/audit-log\", noop)"
if old in txt:
    txt = txt.replace(old, new)
    p.write_text(txt, encoding="utf-8")
    print("fixed block v2")
else:
    print("old v2 not found, try raw check")
    print(repr(txt[txt.find('/experts/:id/chunk-tree')-20:txt.find('/experts/:id/chunk-tree')+300]))
# ensure count check uses 25 not 22
if "got != 22" in txt:
    txt2 = p.read_text(encoding="utf-8", errors="ignore")
    txt2 = txt2.replace("got != 22", "got != 25")
    p.write_text(txt2, encoding="utf-8")
    print("fixed count 22->25")
# verify
t = p.read_text(encoding="utf-8", errors="ignore")
for i,l in enumerate(t.splitlines(),1):
    if "chunk" in l or "system-health" in l or "audit-log" in l or "compare" in l or "want 25" in l:
        print(f"{i:3d} {l}")
print("total group.", t.count("group."))
