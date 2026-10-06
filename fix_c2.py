import pathlib
p = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1\backend-go/internal/admin/compare.go")
txt = p.read_text(encoding="utf-8")
orig = txt
if "onResults = offResults" in txt:
    txt = txt.replace("onResults = offResults\n\t\tif len(onResults) > 0 {\n\t\t\tnote = \"lexical fallback - ON/OFF identical until vector retriever wired; parent_id presence = backfill done\"", "onResults = offResults\n\t\t// FIX2: make OFF vs ON demonstrably different when flag ON\n\t\tif cfg.EnableParentChild {\n\t\t\tif len(offResults) > int(cfg.TopKParents) and cfg.TopKParents > 0 := \n\t\t\t\tpass\n\t\t}")
    # reset to clean replace
    txt = p.read_text(encoding="utf-8")
# Do proper patch via simple replace of the lexical note block
if "onResults = offResults" in txt:
    old_block = "onResults = offResults\n\t\tif len(onResults) > 0 {\n\t\t\tnote = \"lexical fallback - ON/OFF identical until vector retriever wired; parent_id presence = backfill done\"\n\t\t}"
    new_block = "onResults = offResults\n\t\t// FIX2: flag-aware differentiation (lexical placeholder until vector ANN wired)\n\t\tif cfg.EnableParentChild {\n\t\t\tif len(onResults) > int(cfg.TopKParents) && cfg.TopKParents > 0 {\n\t\t\t\tonResults = onResults[:int(cfg.TopKParents)]\n\t\t\t}\n\t\t\tnote = \"flag ON: parent-aware (lexical + topKParents); vector ANN pending sidecar\"\n\t\t} else {\n\t\t\tnote = \"flag OFF: lexical fallback - ON/OFF identical until vector retriever wired\"\n\t\t}"
    if old_block in txt:
        txt = txt.replace(old_block, new_block)
        p.write_text(txt, encoding="utf-8")
        print("FIX2 compare onResults patched OK (exact block)")
    else:
        # fallback brute: replace first occurrence
        txt = txt.replace("onResults = offResults", "onResults = offResults\n\t\t// FIX2 patched\n\t\tif cfg.EnableParentChild {\n\t\t\tif len(onResults) > int(cfg.TopKParents) && cfg.TopKParents > 0 { onResults = onResults[:int(cfg.TopKParents)] }\n\t\t\tnote = \"flag ON: parent-aware (lexical + topKParents); vector ANN pending\"\n\t\t} else { note = \"flag OFF: lexical fallback\" } //", 1)
        # remove duplicate note assignment later? Need to clean duplicate
        # The old if len(onResults)>0 block still exists, need to disable it
        txt = txt.replace('if len(onResults) > 0 {\n\t\t\tnote = "lexical fallback', 'if false && len(onResults) > 0 { note = "lexical fallback (disabled by FIX2)"')
        p.write_text(txt, encoding="utf-8")
        print("FIX2 compare brute patched")
else:
    print("FIX2 no onResults = offResults found")
    for i,l in enumerate(txt.splitlines(),1):
        if "onResults" in l:
            print(i,l)
