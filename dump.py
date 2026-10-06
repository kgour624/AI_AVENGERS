import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\migrations\001_initial_schema.up.sql")
txt = p.read_text(encoding="utf-8", errors="ignore")
# find course_chunks table definition
import re
for m in re.finditer(r"CREATE TABLE.*?course_chunks.*?\);", txt, re.S|re.I):
    print(m.group(0)[:5000])
    print("---END---")
# also expert_pages
p2 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\migrations\081_parent_child_retrieval.up.sql")
print(p2.read_text(encoding="utf-8", errors="ignore"))
print("==081 END==")
p3 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\api\admin.ts")
txt3 = p3.read_text(encoding="utf-8", errors="ignore")
# find last 5 functions
lines = txt3.splitlines()
for i,l in enumerate(lines):
    if "RetrievalConfig" in l:
        print(f"{i+1}: {l}")
# print tail
print("---TAIL admin.ts---")
print("\n".join(lines[-120:]))
print("---HEAD AdminRagSettings---")
p4 = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\pages\admin\AdminRagSettings.tsx")
print(p4.read_text(encoding="utf-8", errors="ignore"))
