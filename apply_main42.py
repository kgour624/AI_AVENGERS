import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\cmd\server\main.go")
txt = p.read_text(encoding="utf-8", errors="ignore")
old = 'adminGroup.GET("/experts/:id/chunk-tree", adminHandler.GetExpertChunkTree)'
new = 'adminGroup.GET("/experts/:id/chunk-tree", adminHandler.GetExpertChunkTree)\n\t\tadminGroup.GET("/system-health", adminHandler.GetSystemHealth)\n\t\tadminGroup.POST("/retrieval-compare", adminHandler.CompareRetrieval)\n\t\tadminGroup.GET("/audit-log", adminHandler.GetAuditLog)'
if old in txt and "GetSystemHealth" not in txt:
    txt = txt.replace(old, new)
    p.write_text(txt, encoding="utf-8")
    print("patched main.go with 3 slice 4.2 routes")
else:
    print("already patched or old not found", "GetSystemHealth" in txt)
