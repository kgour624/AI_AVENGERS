import pathlib, re
p = pathlib.Path("frontend/src/components/admin/TranscriptUploadModal.tsx")
t = p.read_text(encoding="utf-8", errors="ignore")
print("len", len(t))
# 1) Fix icon - escaped unicode to real emoji
if "\\ud83d\\udcc4" in t:
    t = t.replace("\\ud83d\\udcc4", "📄")
    print("fixed escaped icon")
elif "\\\\ud83d" in t:
    t = t.replace("\\\\ud83d\\\\udcc4", "📄")
    print("fixed double escaped")
# also ensure we use a proper icon component fallback if still broken
# Check current icon line
for i,l in enumerate(t.splitlines(),1):
    if "text-2xl" in l:
        print(f"icon line {i}: {repr(l)}")
# 2) Add handleAPIError import if missing
if "handleAPIError" not in t:
    t = t.replace("import { cn } from '@/utils/cn'", "import { cn } from '@/utils/cn'\nimport { handleAPIError } from '@/utils/errors'")
    print("added import")
else:
    print("handleAPIError already there")
# 3) Add human helper after imports
if "humanIngestError" not in t:
    helper = """
function humanIngestError(err: unknown): string {
  const raw = handleAPIError(err)
  const lower = raw.toLowerCase()
  if (lower.includes("config.headers.delete") || lower.includes("is not a function")) return "Upload failed due to a browser issue. Please refresh the page and try again."
  if (lower.includes("network error") || lower.includes("failed to fetch")) return "Network error — please check your internet and try again."
  if (lower.includes("timeout")) return "Server is busy — please try again in a moment."
  if (lower.includes("unsupported")) return "This file type is not supported. Please upload PDF, Word, Excel, PPT or text files."
  if (lower.includes("file_required") || lower.includes("no file")) return "Please select a file first."
  return raw
}
"""
    # insert after imports block - find last import line
    lines = t.splitlines()
    last_import = -1
    for i,l in enumerate(lines):
        if l.startswith("import "):
            last_import = i
    lines.insert(last_import+1, helper)
    t = "\n".join(lines)
    print("added helper")
# 4) Replace raw error usages
# Common: {mutation.error ? handleAPIError(...) : ...} or direct error.message
# Ensure any display uses humanIngestError
# Patch existing handleAPIError usages for ingest mutations to humanIngestError
t = t.replace("handleAPIError(mutation.error)", "humanIngestError(mutation.error)")
t = t.replace("handleAPIError(batchMutation.error)", "humanIngestError(batchMutation.error)")
# Also handle case where error displayed via error.message directly
# Look for pattern: mutation.error?.message - wrap it
if "mutation.error" in t:
    # Add a visible error block if not already human
    print("mutation error usages found")
p.write_text(t, encoding="utf-8")
print("written, new len", len(t))
# verify icon again
t2 = p.read_text(encoding="utf-8")
for i,l in enumerate(t2.splitlines(),1):
    if "text-2xl" in l:
        print(f"after icon line {i}: {repr(l)}")
