import pathlib
p = pathlib.Path("frontend/src/api/base.ts")
t = p.read_text(encoding="utf-8")
# Replace FormData block with fully safe version
old = "  if (config.data instanceof FormData) {\n    if (config.headers && typeof (config.headers as any).delete === 'function') {\n      (config.headers as any).delete('Content-Type')\n      (config.headers as any).delete('content-type')\n    } else {\n      try { delete (config.headers as any)['Content-Type'] } catch {}\n      try { delete (config.headers as any)['content-type'] } catch {}\n    }\n  }"
if old in t:
    new = """  if (config.data instanceof FormData) {
    // Let browser set multipart boundary - remove any explicit Content-Type safely
    try {
      const h: any = config.headers
      if (h) {
        if (typeof h.delete === 'function') {
          try { h.delete('Content-Type') } catch {}
          try { h.delete('content-type') } catch {}
          try { h.delete('Content-type') } catch {}
        } else {
          try { delete h['Content-Type'] } catch {}
          try { delete h['content-type'] } catch {}
          try { delete h['Content-type'] } catch {}
          try { delete h.common?.['Content-Type'] } catch {}
        }
        try { h['Content-Type'] = undefined } catch {}
      }
      if (config.headers && typeof (config.headers as any).set === 'function') {
        try { (config.headers as any).set('Content-Type', undefined as any) } catch {}
      }
    } catch {}
  }"""
    t = t.replace(old, new)
    p.write_text(t, encoding="utf-8")
    print("patched old block")
else:
    # Already patched or different
    if "Browser ko boundary" in t or "Let browser set" in t:
        print("already patched")
    else:
        print("old block not found, searching")
        idx = t.find("FormData")
        print(repr(t[idx-200:idx+800]))
