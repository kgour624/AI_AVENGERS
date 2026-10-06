import pathlib, re
root = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1")
p = root / "backend-go/internal/admin/compare.go"
txt = p.read_text(encoding="utf-8")
orig = txt
if "cfg, _ := training.LoadRetrievalConfig" in txt:
    print("FIX2 cfg already")
else:
    if "cfg := training.DefaultRetrievalConfig()" in txt:
        txt = txt.replace("cfg := training.DefaultRetrievalConfig()", "cfg, _ := training.LoadRetrievalConfig(c.Request.Context(), h.db)\n\tif cfg.ParentSoftLimit == 0 && cfg.ChildSoftLimit == 0 { cfg = training.DefaultRetrievalConfig() }")
        print("FIX2 cfg patched with LoadRetrievalConfig")
    elif "training.DefaultRetrievalConfig()" in txt:
        txt = txt.replace("training.DefaultRetrievalConfig()", "(func() training.RetrievalConfig { cfg, _ := training.LoadRetrievalConfig(c.Request.Context(), h.db); if cfg.ParentSoftLimit==0 && cfg.ChildSoftLimit==0 { return training.DefaultRetrievalConfig() }; return cfg }())")
        print("FIX2 cfg alt patched")
if "onResults := offResults" in txt:
    txt = txt.replace("onResults := offResults", "onResults := offResults // FIX2 stub removed\n\tif cfg.EnableParentChild {\n\t\tif len(offResults) > int(cfg.TopKParents) && cfg.TopKParents > 0 {\n\t\t\tonResults = offResults[:int(cfg.TopKParents)]\n\t\t}\n\t\tnote = \"flag ON: parent-aware (lexical + topKParents); vector ANN pending sidecar\"\n\t} else {\n\t\tnote = \"flag OFF: lexical fallback\"\n\t}")
    print("FIX2 onResults stub patched")
else:
    print("FIX2 onResults not found")
    for i,l in enumerate(txt.splitlines(),1):
        if "onResults" in l or "offResults" in l:
            print(i,l)
if txt != orig:
    # ensure training import
    if '"training"' not in txt and 'training.' in txt:
        if 'import (' in txt and 'internal/training' not in txt:
            txt = txt.replace('import (', 'import (\n\t"github.com/AI_AVENGERS/backend-go/internal/training"')
            # fix module path guess: search existing import for training
    p.write_text(txt, encoding="utf-8")
    print("FIX2 written")
else:
    print("FIX2 no change")
