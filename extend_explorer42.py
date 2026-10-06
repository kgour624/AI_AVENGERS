import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\pages\admin\AdminChunksExplorer.tsx")
txt = p.read_text(encoding="utf-8", errors="ignore")
# inject health + compare + audit into explorer page — append imports and render blocks
if "SystemHealthCard" not in txt:
    txt = txt.replace(
        "import { handleAPIError } from '@/utils/errors'",
        "import { handleAPIError } from '@/utils/errors'\nimport { getSystemHealth } from '@/api/admin'\nimport { SystemHealthCard } from '@/components/admin/SystemHealthCard'\nimport { ComparePanel } from '@/components/admin/ComparePanel'\nimport { AuditLogList } from '@/components/admin/AuditLogList'"
    )
    # add health query after expertsQ line
    txt = txt.replace(
        "const expertsQ=useQuery({queryKey:['admin','experts'],queryFn:getAdminExperts})",
        "const expertsQ=useQuery({queryKey:['admin','experts'],queryFn:getAdminExperts})\nconst healthQ=useQuery({queryKey:['admin','system-health'],queryFn:getSystemHealth})"
    )
    # inject health card after title div — find first </div><Card> pattern
    # simple: after the h2/p div block
    txt = txt.replace(
        "</div><Card><div className='flex flex-wrap items-end gap-3'>",
        "</div><SystemHealthCard data={healthQ.data as never} isLoading={healthQ.isLoading} error={healthQ.error}/><Card><div className='flex flex-wrap items-end gap-3'>"
    )
    # inject compare + audit after chunk-tree tab — find where tree tab ends, add extra cards
    # append before final </div>)} — add two extra sections after parents tab block
    txt = txt.replace(
        "export const Component=AdminChunksExplorer",
        "{expertId ? <ComparePanel expertId={expertId}/> : null}<AuditLogList/>export const Component=AdminChunksExplorer"
    )
    p.write_text(txt, encoding="utf-8")
    print("extended AdminChunksExplorer with health/compare/audit")
else:
    print("already extended")
print("lines", len(txt.splitlines()))
