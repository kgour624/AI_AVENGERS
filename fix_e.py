import pathlib
root = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1")
p = root / "frontend/src/api/admin.ts"
txt = p.read_text(encoding="utf-8")
orig = txt
# brute fix helper keys
changed=False
if "parent_top_k" in txt:
    txt = txt.replace("parent_top_k", "top_k_parents")
    changed=True
if "child_top_k" in txt:
    txt = txt.replace("child_top_k", "top_k_children")
    changed=True
# Also ensure helper includes enable_parent_child etc if incomplete
# Check helper function definition
if 'toSnakeRetrievalPayload' in txt:
    # verify helper now has correct keys, print snippet
    for i,l in enumerate(txt.splitlines(),1):
        if "toSnakeRetrievalPayload" in l or "top_k_parents" in l or "top_k_children" in l:
            print(i,l)
if changed:
    p.write_text(txt, encoding="utf-8")
    print("FIX5 patched")
else:
    print("FIX5 no parent_top_k found, checking brute")
    # still check
    if "top_k_parents" in orig:
        print("FIX5 already correct")
    else:
        print("FIX5 nothing")
# migrations fix
mig_dir = root / "backend-go/migrations"
files = list(mig_dir.glob("*.sql"))
print([f.name for f in files][:20])
has_parent=False
for f in files:
    t=f.read_text(encoding="utf-8", errors="ignore").lower()
    if "parent_id" in t:
        has_parent=True
        print(f"parent_id in {f.name}")
print("has_parent_id", has_parent)
if not has_parent:
    content="""-- 081_parent_child_retrieval: additive parent-child support
CREATE TABLE IF NOT EXISTS expert_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    page_index INT NOT NULL,
    page_text TEXT NOT NULL,
    section_path TEXT,
    token_count INT NOT NULL DEFAULT 0,
    source_file TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(expert_id, page_index, source_file)
);
CREATE INDEX IF NOT EXISTS idx_expert_pages_expert_id ON expert_pages(expert_id);
ALTER TABLE course_chunks ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES expert_pages(id) ON DELETE SET NULL;
ALTER TABLE course_chunks ADD COLUMN IF NOT EXISTS parent_index INT DEFAULT -1;
ALTER TABLE course_chunks ADD COLUMN IF NOT EXISTS is_child BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX IF NOT EXISTS idx_course_chunks_parent_id ON course_chunks(expert_id, parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_course_chunks_is_child ON course_chunks(expert_id, is_child);
"""
    (mig_dir/"081_parent_child_retrieval.sql").write_text(content, encoding="utf-8")
    (mig_dir/"081_parent_child_retrieval.up.sql").write_text(content, encoding="utf-8")
    print("FIX6 migration created")
else:
    print("FIX6 already exists")
