import pathlib

# We need to capture exact file contents for safe edits
# 1. Dump admin_handler tail needed
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\admin_handler.go")
lines = p.read_text(encoding="utf-8", errors="ignore").splitlines()
# find the DeleteExpert end
for i,l in enumerate(lines,1):
    if "func (h *AdminHandler) DeleteExpert" in l:
        print(f"DeleteExpert at {i}")
        for j in range(i, min(len(lines), i+30)):
            print(f"{j:4d} {lines[j-1]}")

# Need full file for python injection approach
# Also check imports
print("\n=== imports ===")
for i,l in enumerate(lines[:40],1):
    print(f"{i:4d} {l}")

# Check chunker.go types
p2 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\training\chunker.go")
txt2 = p2.read_text(encoding="utf-8", errors="ignore")
print("\n=== chunker types ===")
for i,l in enumerate(txt2.splitlines(),1):
    if "type " in l or "ParentChunk" in l or "TextChunk" in l:
        print(f"{i:4d} {l}")
        if i>120:
            break
# print first 200 lines
print(txt2[:5000])

# Check response helpers available
p3 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\response\response.go")
print("\n=== response.go ===")
print(p3.read_text(encoding="utf-8", errors="ignore")[:4000])

# Check router_test for counting
p4 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\router_test.go")
print("\n=== router_test ===")
print(p4.read_text(encoding="utf-8", errors="ignore"))

# Check admin.ts RetrievalConfig details
p5 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\api\admin.ts")
txt5 = p5.read_text(encoding="utf-8", errors="ignore")
# find RetrievalConfig block
idx = txt5.find("export interface RetrievalConfig")
print("\n=== RetrievalConfig ===")
print(txt5[idx:idx+3000])

# Check parent_store for expert_pages schema
p6 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\training\parent_store.go")
print("\n=== parent_store.go ===")
print(p6.read_text(encoding="utf-8", errors="ignore"))

p7 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\training\store_parent_child.go")
print("\n=== store_parent_child.go ===")
print(p7.read_text(encoding="utf-8", errors="ignore"))
