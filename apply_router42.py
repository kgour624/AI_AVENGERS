import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\router_test.go")
txt = p.read_text(encoding="utf-8", errors="ignore")
print("before has GetSystemHealth", "GetSystemHealth" in txt)
# bump want 22 -> 25
for a,b in [("want 22","want 25"), ("Expected 22","Expected 25"), ("len(routes) != 22","len(routes) != 25"), ("== 22","== 25")]:
    if a in txt:
        txt = txt.replace(a,b)
        print(f"replaced {a} -> {b}")
# add 3 routes if chunk-tree exists
if "GetSystemHealth" not in txt and "/experts/:id/chunk-tree" in txt:
    txt = txt.replace('"/experts/:id/chunk-tree"', '"/experts/:id/chunk-tree"\n\tgroup.GET("/system-health", func(c *gin.Context) { c.Status(200) })\n\tgroup.POST("/retrieval-compare", func(c *gin.Context) { c.Status(200) })\n\tgroup.GET("/audit-log"')
    # fix missing close
    if 'group.GET("/audit-log"' in txt and 'group.GET("/audit-log", func(c *gin.Context)' not in txt:
        txt = txt.replace('group.GET("/audit-log"', 'group.GET("/audit-log", func(c *gin.Context) { c.Status(200) })')
    p.write_text(txt, encoding="utf-8")
    print("patched router_test with 3 routes + bump to 25")
else:
    # if count not 22 pattern, just ensure routes added
    if "GetSystemHealth" not in txt:
        # append after any admin group
        txt2 = txt.replace('c.Status(200) })', 'c.Status(200) })\n\tgroup.GET("/system-health", func(c *gin.Context) { c.Status(200) })\n\tgroup.POST("/retrieval-compare", func(c *gin.Context) { c.Status(200) })\n\tgroup.GET("/audit-log", func(c *gin.Context) { c.Status(200) })', 1)
        p.write_text(txt2, encoding="utf-8")
        print("fallback patch router_test")
    else:
        print("already has routes")
for i,l in enumerate(txt.splitlines(),1):
    if "want" in l.lower() or "system-health" in l or "retrieval-compare" in l or "audit-log" in l:
        print(i, l.strip())
