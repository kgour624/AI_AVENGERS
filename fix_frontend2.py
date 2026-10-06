import pathlib
# Patch RagSettings link and Layout nav and App routes
rag = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\pages\admin\AdminRagSettings.tsx")
txt = rag.read_text(encoding="utf-8", errors="ignore")
if "AdminChunksExplorer" not in txt and "/admin/chunks" not in txt:
    # add Link import
    if "from 'react-router-dom'" not in txt:
        txt = txt.replace("import { handleAPIError } from '@/utils/errors'", "import { handleAPIError } from '@/utils/errors'\nimport { Link } from 'react-router-dom'")
    else:
        txt = txt.replace("import { handleAPIError }", "import { Link } from 'react-router-dom'\nimport { handleAPIError }")
    old_h1 = '<h1 className="text-xl font-semibold">RAG Retrieval</h1>'
    new_h1 = '<h1 className="text-xl font-semibold">RAG Retrieval</h1>\n      <Link to="/admin/chunks" className="mt-2 inline-flex rounded-full border border-glow-violet bg-violet-500/10 px-3 py-1 text-xs font-medium text-violet-600">Open Chunk Explorer -></Link>'
    if old_h1 in txt:
        txt = txt.replace(old_h1, new_h1)
    rag.write_text(txt, encoding="utf-8")
    print("patched RagSettings")
else:
    print("RagSettings already patched")

lay = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\pages\admin\AdminLayout.tsx")
tlay = lay.read_text(encoding="utf-8", errors="ignore")
if "/admin/chunks" not in tlay:
    old_nav = "{ to: '/admin/rag', label: 'RAG Retrieval' },"
    new_nav = "{ to: '/admin/rag', label: 'RAG Retrieval' },\n  { to: '/admin/chunks', label: 'Chunk Explorer' },"
    if old_nav in tlay:
        tlay = tlay.replace(old_nav, new_nav)
        lay.write_text(tlay, encoding="utf-8")
        print("patched layout")
    else:
        print("old_nav not found")
        # dump around
        for i,l in enumerate(tlay.splitlines(),1):
            if "rag" in l.lower():
                print(i,l)

app = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\frontend\src\App.tsx")
tapp = app.read_text(encoding="utf-8", errors="ignore")
if "AdminChunksExplorer" not in tapp and "/admin/chunks" not in tapp:
    old_route = "{ path: 'rag', lazy: () => import('@/pages/admin/AdminRagSettings') },"
    new_route = "{ path: 'rag', lazy: () => import('@/pages/admin/AdminRagSettings') },\n              { path: 'chunks', lazy: () => import('@/pages/admin/AdminChunksExplorer') },"
    if old_route in tapp:
        tapp = tapp.replace(old_route, new_route)
        app.write_text(tapp, encoding="utf-8")
        print("patched App.tsx")
    else:
        print("old_route not found")
        for i,l in enumerate(tapp.splitlines(),1):
            if "rag" in l.lower():
                print(i,l)
else:
    print("App already has chunks route")
