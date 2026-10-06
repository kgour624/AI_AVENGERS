import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\admin_handler.go")
txt = p.read_text(encoding="utf-8", errors="ignore")
# simple append before last DeleteExpert section? Instead use known anchor after response.OK(c, req) in UpdateRetrievalConfig
anchor = "response.OK(c, req)\n}"
idx = txt.rfind(anchor)
# we want the first occurrence after retrieval config? Find around 4338.
# Use rfind but ensure we get the retrieval one, not later.
# Let's scan for UpdateRetrievalConfig anchor specifically.
search = "func (h *AdminHandler) UpdateRetrievalConfig"
a = txt.find(search)
end = txt.find(anchor, a)
if end == -1:
    raise SystemExit("anchor miss")
pre = txt[:end+len(anchor)]
post = txt[end+len(anchor):]
snippet = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\chunk_explorer_snippet.go").read_text(encoding="utf-8")
new = pre + "\n" + snippet + "\n" + post
p.write_text(new, encoding="utf-8")
print("patched admin_handler, snippet inserted, length", len(new.splitlines()))
print("contains ListExpertChunks:", "ListExpertChunks" in new)
