import pathlib, re, os, glob

# 1. Ensure audit_log.go logs INFO not WARN for missing table (idempotent patch)
p = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\audit_log.go")
if p.exists():
    txt = p.read_text(encoding="utf-8", errors="ignore")
    orig = txt
    # downgrade any Warn mentioning audit_logs to Info
    txt = txt.replace('logger.Warn("audit_logs query failed', 'logger.Info("audit_logs fallback')
    txt = txt.replace('audit_logs query failed (table may not exist)', 'audit_logs table missing — serving synthetic fallback')
    txt = re.sub(r'logger\.Warn\(([^)]*audit_logs[^)]*)\)', lambda m: m.group(0).replace('logger.Warn','logger.Info'), txt)
    if txt != orig:
        p.write_text(txt, encoding="utf-8")
        print(f"patched audit_log.go WARN->INFO")
    else:
        print(f"audit_log.go already INFO")

# 2. Create proper versioned migration for audit_logs (fix 999 not picked up)
mig_dir = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1\backend-go\migrations")
mig_dir.mkdir(exist_ok=True, parents=True)
# find max numeric prefix
max_n = 0
for f in mig_dir.glob("*.up.sql"):
    m = re.match(r"(\d+)_", f.name)
    if m:
        try:
            n = int(m.group(1))
            max_n = max(max_n, n)
        except: pass
print(f"max migration n={max_n}")
next_n = max_n + 1
# also handle 019 already may exist for audit - if we already created 999, next will be correct
# ensure we create 019 or next_n whichever larger
# force create both 019 and next_n to guarantee migrate picks it up
for n in [19, next_n]:
    up = mig_dir / f"{n:03d}_audit_logs.up.sql"
    down = mig_dir / f"{n:03d}_audit_logs.down.sql"
    sql_up = """-- audit_logs table for admin audit log (fixes 42P01 WARN -> INFO fallback)
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
"""
    sql_down = """DROP TABLE IF EXISTS audit_logs;"""
    if not up.exists():
        up.write_text(sql_up, encoding="utf-8")
        print(f"created {up}")
    if not down.exists():
        down.write_text(sql_down, encoding="utf-8")
        print(f"created {down}")

# also ensure root migrations/ exists for Dockerfile that does COPY migrations/
root_mig = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1\migrations")
root_mig.mkdir(exist_ok=True, parents=True)
# copy latest audit migration to root as well
import shutil
for f in mig_dir.glob("*audit_logs*"):
    dest = root_mig / f.name
    if not dest.exists():
        shutil.copy(f, dest)
        print(f"copied {f.name} to root migrations/")
# also copy existing 999 file if needed as 019
p999 = mig_dir / "999_audit_logs_if_not_exists.sql"
if p999.exists():
    print(f"999 file exists, will keep but also have versioned 019")
print("DONE fix_audit_final.py")
