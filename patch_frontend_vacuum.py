import pathlib
p = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1\frontend\src\App.tsx")
t = p.read_text(encoding="utf-8")
if "AdminVacuum" not in t:
    t = t.replace(
        "{ path: 'mcp', lazy: () => import('@/pages/admin/AdminMcpControlCenter') },",
        "{ path: 'mcp', lazy: () => import('@/pages/admin/AdminMcpControlCenter') },\n              { path: 'vacuum', lazy: () => import('@/pages/admin/AdminVacuum') },"
    )
    p.write_text(t, encoding="utf-8")
    print("patched App.tsx")
else:
    print("already patched")

q = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1\frontend\src\pages\admin\AdminLayout.tsx")
s = q.read_text(encoding="utf-8")
if "vacuum" not in s.lower():
    # add nav link near mcp
    s = s.replace("MCP", "MCP")
    if 'to="/admin/mcp"' in s or "mcp" in s.lower():
        s = s.replace('"/admin/mcp"', '"/admin/mcp"')
        # naive append after mcp NavLink line
        s = s.replace("AdminMcpControlCenter", "AdminMcpControlCenter")
    # Instead inject vacuum link after mcp link by string replace on JSX
    if '"/admin/vacuum"' not in s:
        s = s.replace(
            'to="/admin/mcp"',
            'to="/admin/mcp"\n              />\n            </NavLink>\n            <NavLink to="/admin/vacuum" className={navClass}>Vacuum'
        )
        # The above may break; fallback: just append vacuum line before closing nav
        if '"/admin/vacuum"' not in s:
            s = s.replace("</nav>", '  <NavLink to="/admin/vacuum" className={({isActive}:any)=> isActive? \"active\":\"\" }>Vacuum</NavLink>\n</nav>')
    q.write_text(s, encoding="utf-8")
    print("patched AdminLayout")
else:
    print("layout already has vacuum")
