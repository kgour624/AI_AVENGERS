import pathlib, re
root = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1")
p = root / "backend-go/internal/admin/system_health.go"
txt = p.read_text(encoding="utf-8")
if "SELECT EXISTS(SELECT 1 FROM system_settings WHERE key='retrieval_config')" in txt:
    print("FIX1 already")
else:
    orig = txt
    # ensure training import exists
    if '"AI_AVENGERS' not in txt and 'training' not in txt:
        # try to find import block and add
        if 'import (' in txt:
            txt = txt.replace('import (', 'import (\n\t"github.com/AI_AVENGERS/backend-go/internal/training"')
    # Patch DefaultRetrievalConfig -> Load
    if 'cfg := training.DefaultRetrievalConfig()' in txt:
        # check if source var follows
        if 'cfg := training.DefaultRetrievalConfig()\n\tsource := "default"' in txt:
            txt = txt.replace('cfg := training.DefaultRetrievalConfig()\n\tsource := "default"', 'cfg := training.DefaultRetrievalConfig()\n\tsource := "default"\n\t// FIX1: load actual DB config instead of hardcoded defaults\n\tif h.db != nil {\n\t\tif dbCfg, err := training.LoadRetrievalConfig(c.Request.Context(), h.db); err == nil {\n\t\t\tvar exists bool\n\t\t\t_ = h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM system_settings WHERE key=\'retrieval_config\')`).Scan(&exists)\n\t\t\tif exists {\n\t\t\t\tcfg = dbCfg\n\t\t\t\tsource = "system_settings"\n\t\t\t}\n\t\t}\n\t}')
        else:
            txt = txt.replace('cfg := training.DefaultRetrievalConfig()', 'cfg := training.DefaultRetrievalConfig()\n\tsource := "default"\n\tif h.db != nil {\n\t\tif dbCfg, err := training.LoadRetrievalConfig(c.Request.Context(), h.db); err == nil {\n\t\t\tvar exists bool\n\t\t\t_ = h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM system_settings WHERE key=\'retrieval_config\')`).Scan(&exists)\n\t\t\tif exists { cfg = dbCfg; source = "system_settings" }\n\t\t}\n\t}\n\t// FIX1 duplicate source guard')
            # remove duplicate if double source
            txt = txt.replace('source := "default"\n\tsource := "default"', 'source := "default"')
    # Also ensure FlagEnabled uses cfg.EnableParentChild (already does) but now reflects DB
    if txt != orig:
        p.write_text(txt, encoding="utf-8")
        print("FIX1 patched")
    else:
        print("FIX1 pattern not found, printing hints")
        for i,l in enumerate(txt.splitlines(),1):
            if "DefaultRetrievalConfig" in l or "FlagEnabled" in l or "system_settings" in l:
                print(i,l)
