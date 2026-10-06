import pathlib, re
root = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1")
p = root / "backend-go/internal/admin/audit_log.go"
txt = p.read_text(encoding="utf-8")
orig = txt
if "FROM system_settings WHERE key='retrieval_config'" in txt:
    print("FIX3 already")
else:
    if "FROM retrieval_config" in txt:
        txt = txt.replace("FROM retrieval_config", "FROM system_settings")
        txt = re.sub(r"WHERE id='000[^']*'", "WHERE key='retrieval_config'", txt)
        # ensure value column selected if not present
        if "SELECT value" not in txt and "FROM system_settings WHERE key='retrieval_config'" in txt:
            txt = txt.replace("SELECT updated_at,updated_by FROM system_settings WHERE key='retrieval_config'", "SELECT value, updated_at, updated_by FROM system_settings WHERE key='retrieval_config'")
            # if scan expects 2 vars, we keep 3 col query but need to adjust scan
            # Check scan line: if it scans into updatedAt, updatedBy only, expand
            # Find pattern: .Scan(&updatedAt, &updatedBy)
            if ".Scan(&updatedAt, &updatedBy)" in txt or ".Scan(&updated" in txt:
                # patch to scan value into throwaway
                txt = txt.replace(".Scan(&updatedAt, &updatedBy)", ".Scan(&val, &updatedAt, &updatedBy)")
                # Ensure val var declared before: add var val json.RawMessage
                if "var val" not in txt:
                    txt = txt.replace("var updatedAt", "var val json.RawMessage\n\tvar updatedAt")
                    if "encoding/json" not in txt:
                        txt = txt.replace("import (", "import (\n\t\"encoding/json\"")
        p.write_text(txt, encoding="utf-8")
        print("FIX3 patched")
    else:
        print("FIX3 pattern not found")
        for i,l in enumerate(txt.splitlines(),1):
            if "retrieval_config" in l.lower() or "system_settings" in l.lower():
                print(i,l)
