import pathlib
base = r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin"
for name in ['system_health.go','compare.go','audit_log.go']:
    p = pathlib.Path(base + "\\" + name)
    txt = p.read_text(encoding="utf-8", errors="ignore")
    has_pool = "h.pool" in txt
    has_db = "h.db" in txt
    has_rec = "RetrievalConfig" in txt
    has_err = "Error()" in txt
    print(name, "pool?", has_pool, "db?", has_db, "has RetrievalConfig?", has_rec, "has Error()?", has_err)
    if has_pool:
        print(" BAD still has h.pool")
    # check training import
    print(" has training import?", "training" in txt)
# also show that audit_log has auditParseErr Error method
audit = pathlib.Path(base + "\\audit_log.go")
print(audit.read_text(encoding="utf-8", errors="ignore").count("auditParseErr"))
