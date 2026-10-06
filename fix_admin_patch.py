import pathlib
p = pathlib.Path("frontend/src/api/admin.ts")
t = p.read_text(encoding="utf-8")
need = "{ headers: { 'Content-Type': 'multipart/form-data' } }"
if need in t:
    c = t.count(need)
    print(f"found {c} occurrences")
    t = t.replace(need, "undefined /* browser sets boundary */")
    # Cleanup: the call was baseAPI.post(url, formData, undefined ...) -> should be just (url, formData)
    t = t.replace(",\n      formData,\n      undefined /* browser sets boundary */", ",\n      formData")
    t = t.replace(", formData, undefined /* browser sets boundary */", ", formData")
    # Fix leftover undefined as third arg on new line with comment
    # Search for pattern: formData\n    ) -> should stay, but if we left , formData then extra comma handling
    # Ensure no double undefined left
    p.write_text(t, encoding="utf-8")
    print("patched admin.ts")
    # verify
    t2 = p.read_text(encoding="utf-8")
    for i,l in enumerate(t2.splitlines(),1):
        if "multipart" in l or "Content-Type" in l or "browser sets" in l:
            print(i,repr(l[:200]))
else:
    print("not found")
    for i,l in enumerate(t.splitlines(),1):
        if "multipart" in l or "Content-Type" in l:
            print(i,repr(l))
