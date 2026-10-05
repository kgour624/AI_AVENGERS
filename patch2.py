import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\context\assembler.go")
t = p.read_text(encoding="utf-8")
old = "        embedding, err := a.embedder.EmbedSingle(ctx, question)\n        if err != nil {\n                return nil, fmt.Errorf(\"embed failed: %w\", err)\n        }\n\n        // Feature #23: source_file and chunk_index are selected"
new = "        embedding, err := a.embedder.EmbedSingle(ctx, question)\n        if err != nil {\n                return nil, fmt.Errorf(\"embed failed: %w\", err)\n        }\n        // Phase 2: Parent-Child fast-path (flag-gated, additive). If parents exist, return them directly.\n        if pcs, perr := a.tryParentChildRetrieval(ctx, expertID, embedding, question, limit); perr == nil && len(pcs) > 0 {\n                if a.logger != nil {\n                        a.logger.Info(\"parent-child retrieval hit\", zap.Int(\"parents\", len(pcs)), zap.Int(\"limit\", limit))\n                }\n                return pcs, nil\n        }\n\n        // Feature #23: source_file and chunk_index are selected"
if old in t:
    t = t.replace(old, new, 1)
    p.write_text(t, encoding="utf-8")
    print("patched assembler.go successfully")
else:
    print("old snippet not found")
    idx = t.find("embedding, err := a.embedder.EmbedSingle")
    print(repr(t[idx-500:idx+900]))
