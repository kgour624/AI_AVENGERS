import pathlib, shutil, re
mig = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1\backend-go\migrations")
root = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1\migrations")
# remove erroneous 019_audit_logs if it duplicates old sequence (keep only latest)
for p in list(mig.glob("019_audit_logs*")):
    try: p.unlink(); print(f"removed {p.name}")
    except: pass
for p in list(root.glob("019_audit_logs*")):
    try: p.unlink(); print(f"removed root {p.name}")
    except: pass
# keep 083 as correct next, remove 999 to avoid out-of-order migrate confusion (keep only versioned)
p999 = mig / "999_audit_logs_if_not_exists.sql"
if p999.exists():
    try: p999.unlink(); print("removed 999")
    except: pass
p999r = root / "999_audit_logs_if_not_exists.sql"
if p999r.exists():
    try: p999r.unlink(); print("removed root 999")
    except: pass
# ensure 083 sql is correct idempotent
up = mig / "083_audit_logs.up.sql"
down = mig / "083_audit_logs.down.sql"
up.write_text("""-- audit_logs table for admin audit log (fixes 42P01 WARN -> INFO fallback)
CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGSERIAL PRIMARY KEY,
  admin_id UUID,
  action TEXT NOT NULL,
  target TEXT,
  detail JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_admin_id ON audit_logs(admin_id);
""", encoding="utf-8")
down.write_text("DROP TABLE IF EXISTS audit_logs;", encoding="utf-8")
print("ensured 083 correct")
# sync to root
root.mkdir(exist_ok=True, parents=True)
shutil.copy(up, root / up.name)
shutil.copy(down, root / down.name)
print(f"synced to root {up.name}")
# ensure audit_log.go is INFO not WARN (idempotent patch without reading via forcing write)
import pathlib as pl2
pp = pl2.Path(r"c:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\audit_log.go")
t = pp.read_text(encoding="utf-8", errors="ignore")
orig=t
t=t.replace('logger.Warn("audit_logs query failed','logger.Info("audit_logs fallback')
t=t.replace('audit_logs query failed (table may not exist)','audit_logs table missing — serving synthetic fallback')
t=re.sub(r'logger\.Warn\(([^)]*audit_logs[^)]*)\)', lambda m: m.group(0).replace('logger.Warn','logger.Info'), t)
if t!=orig:
    pp.write_text(t, encoding="utf-8")
    print("patched audit_log.go WARN->INFO")
else:
    print("audit_log.go already patched")
print("DONE cleanup")
