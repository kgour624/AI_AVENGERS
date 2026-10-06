import pathlib
main = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\cmd\server\main.go")
txt = main.read_text(encoding="utf-8", errors="ignore")
add = """		adminGroup.GET("/system-health", adminHandler.GetSystemHealth)
		adminGroup.POST("/retrieval-compare", adminHandler.CompareRetrieval)
		adminGroup.GET("/audit-log", adminHandler.GetAuditLog)"""
needles = [
    'adminGroup.GET("/chunk-tree", adminHandler.GetExpertChunkTree)',
]
inserted=False
for n in needles:
    if n in txt and "GetSystemHealth" not in txt:
        txt = txt.replace(n, n + "\n" + add)
        inserted=True
        break
if inserted:
    main.write_text(txt, encoding="utf-8")
    print("patched main.go slice 4.2 routes")
else:
    print("not inserted", "GetSystemHealth" in txt, "needle" in txt)

# bump router_test count 22 -> 25 if needed
rt = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\router_test.go")
t = rt.read_text(encoding="utf-8", errors="ignore")
if "GetSystemHealth" not in t:
    # find count assertion e.g. want 22 or len(routes)==22
    for old, new in [("22", "25"), ("want 22","want 25"), ("expected 22","expected 25")]:
        if old in t:
            print("found", repr(old))
    # fallback: try to add handlers
    # look for last chunk-tree line
    if '/chunk-tree' in t:
        t = t.replace('/chunk-tree', '/chunk-tree"\n\tgroup.GET("/system-health", func(c *gin.Context) { c.Status(200) })\n\tgroup.POST("/retrieval-compare", func(c *gin.Context) { c.Status(200) })\n\tgroup.GET("/audit-log', '/chunk-tree')
        # actually simpler: duplicate pattern
    print("router_test manual patch needed - show snippet")
    # just print surrounding
    for i,l in enumerate(t.splitlines(),1):
        if "chunk" in l.lower() or "want" in l.lower() or "22" in l:
            print(i, l)
else:
    print("router_test already has health")
