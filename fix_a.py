import pathlib, re
root = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1")
p = root / "backend-go/internal/training/ingestion_parent_child.go"
txt = p.read_text(encoding="utf-8")
if "AssignParentIDs(children []TextChunk, parents []ParentChunk, parentIDByIndex map[int]uuid.UUID, rcfg RetrievalConfig)" not in txt:
    txt = txt.replace("func AssignParentIDs(children []TextChunk, parents []ParentChunk, parentIDByIndex map[int]uuid.UUID) []TextChunk {", "func AssignParentIDs(children []TextChunk, parents []ParentChunk, parentIDByIndex map[int]uuid.UUID, rcfg RetrievalConfig) []TextChunk {")
    txt = txt.replace("cfg := DefaultParentConfig()\n\tpHard := cfg.effectiveHard()\n\tpSoft := cfg.effectiveSoft()", "parentCfg := ChunkConfig{TargetSize: rcfg.ParentSoftLimit, MinSize: 900, MaxSize: rcfg.ParentHardLimit, SoftLimit: rcfg.ParentSoftLimit, HardLimit: rcfg.ParentHardLimit, OverlapTokens: rcfg.OverlapTokens, AtomicCodeFence: rcfg.AtomicCodeFence}\n\tif parentCfg.SoftLimit == 0 {\n\t\tparentCfg = DefaultParentConfig()\n\t}\n\tpHard := parentCfg.effectiveHard()\n\tpSoft := parentCfg.effectiveSoft()")
    txt = txt.replace("child index -> parent index by replaying same grouping", "child index -> parent index by replaying same grouping with rcfg limits (FIX4)")
    p.write_text(txt, encoding="utf-8")
    print("FIX4 Assign patched")
else:
    print("FIX4 already")
p2 = root / "backend-go/internal/training/ingestion_pipeline.go"
txt2 = p2.read_text(encoding="utf-8")
if "AssignParentIDs(children, parents, parentIDByIndex, rcfg)" not in txt2 and "AssignParentIDs(chunks, parents, parentIDByIndex, rcfg)" not in txt2:
    orig = txt2
    txt2 = re.sub(r"AssignParentIDs\(([^)]+),\s*parentIDByIndex\)", lambda m: f"AssignParentIDs({m.group(1)}, parentIDByIndex, rcfg)", txt2)
    if txt2 != orig:
        p2.write_text(txt2, encoding="utf-8")
        print("FIX4 pipeline patched")
    else:
        print("FIX4 pipeline not found")
        for i,l in enumerate(txt2.splitlines(),1):
            if "AssignParentIDs" in l:
                print(i,l)
else:
    print("FIX4 pipeline already")
