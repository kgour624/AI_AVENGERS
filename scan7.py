import pathlib, glob, re

# Check current pipeline inserts - store_parent_child and parent_store
for name in ["store_parent_child.go","parent_store.go","ingestion_pipeline.go","ingestion_parent_child.go","child_store.go"]:
    p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\training") / name
    if not p.exists():
        continue
    txt = p.read_text(encoding="utf-8", errors="ignore")
    print(f"\n===== {name} =====")
    for i,l in enumerate(txt.splitlines(),1):
        if "INSERT INTO" in l or "course_chunks" in l.lower() or "expert_pages" in l.lower() or "parent_id" in l.lower():
            print(f"{i:4d} {l}")
    if len(txt.splitlines()) < 400:
        print("---FULL---")
        print(txt)
        print("---END FULL---")

# list admin_handler functions for reference
pAdmin = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\admin_handler.go")
lines = pAdmin.read_text(encoding="utf-8", errors="ignore").splitlines()
print("\n===== admin_handler function list =====")
for i,l in enumerate(lines,1):
    if l.strip().startswith("func (h *AdminHandler)"):
        print(f"{i:4d} {l.strip()}")

# check main.go exact route registration block
pMain = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\cmd\server\main.go")
lines2 = pMain.read_text(encoding="utf-8", errors="ignore").splitlines()
# print around 1228-1252
for i in range(1225, 1255):
    print(f"{i+1:4d} {lines2[i]}")

# check frontend AdminLayout nav
pLay = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\pages\admin\AdminLayout.tsx")
print("\n===== AdminLayout =====")
print(pLay.read_text(encoding="utf-8", errors="ignore"))

# check App.tsx routes
pApp = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\App.tsx")
txtApp = pApp.read_text(encoding="utf-8", errors="ignore")
for i,l in enumerate(txtApp.splitlines(),1):
    if "admin" in l.lower() or "rag" in l.lower() or "Admin" in l:
        print(f"{i:4d} {l}")
