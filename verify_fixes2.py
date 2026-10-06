import pathlib, re
base = r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin"
# Check for duplicate types or missing imports
for name in ['system_health.go','compare.go','audit_log.go','admin_handler.go']:
    p = pathlib.Path(base + "\\" + name)
    txt = p.read_text(encoding="utf-8", errors="ignore")
    print("---", name)
    # check module prefix: should be ai_avengers/backend
    if "ai_avengers/backend" not in txt and name != "admin_handler.go":
        print("  import prefix check: needs ai_avengers/backend")
    # check if knowledge vs training RetrievalConfig used correctly
    if "knowledge.RetrievalConfig" in txt or "training.RetrievalConfig" in txt:
        print("  has qualified RetrievalConfig")
    # check any stray h.pool
    if "h.pool" in txt:
        print("  STILL HAS h.pool!!!")
# Check go.mod module name
mod = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\go.mod")
print(mod.read_text(encoding="utf-8", errors="ignore")[:200])
