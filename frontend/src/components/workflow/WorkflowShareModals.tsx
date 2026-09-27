import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/Button'
import {
  createPublication,
  downloadWorkflowFile,
  listPublications,
  listWorkflowFiles,
  pushToPublication,
  type Publication,
  type WorkflowFileEntry,
} from '@/api/workflows'

/**
 * FilesModal — every file the workflow produced, with a download each.
 *
 * WHY it reads from the workflow's event log (same source as the FILES tab) and
 * not from a cache: a design-only workflow that later had code generated, or a
 * redesigned file, must show the NEWEST content. The list is rebuilt on every
 * open, so nothing stale can be downloaded.
 */
export function WorkflowFilesModal({
  workflowId,
  open,
  onClose,
}: {
  workflowId: string
  open: boolean
  onClose: () => void
}) {
  const [files, setFiles] = useState<WorkflowFileEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    let cancelled = false
    setLoading(true)
    setError(null)
    listWorkflowFiles(workflowId)
      .then((f) => {
        if (!cancelled) setFiles(f)
      })
      .catch(() => {
        if (!cancelled) setError('Could not load this workflow\u2019s files.')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [open, workflowId])

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="flex max-h-[80vh] w-full max-w-xl flex-col rounded-xl border border-border-subtle bg-card p-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-text-primary">Files produced</h3>
          <button type="button" className="text-xs text-text-secondary hover:text-text-primary" onClick={onClose}>
            Close
          </button>
        </div>
        {loading && <p className="mt-4 text-xs text-text-disabled">Loading files...</p>}
        {error && <p className="mt-4 text-xs text-mode-refuse">{error}</p>}
        {!loading && !error && files.length === 0 && (
          <p className="mt-4 text-xs text-text-disabled">This workflow has not produced any files yet.</p>
        )}
        <div className="mt-3 min-h-0 flex-1 space-y-2 overflow-y-auto">
          {files.map((f) => (
            <div
              key={f.path}
              className="flex items-center justify-between gap-3 rounded-md border border-border-subtle bg-white/5 px-3 py-2"
            >
              <span className="min-w-0 flex-1 truncate text-xs text-text-primary" title={f.path}>
                {f.name ?? f.path.split('/').pop()}
                <span className="ml-2 text-text-disabled">{f.path}</span>
              </span>
              <button
                type="button"
                className="flex-shrink-0 rounded-md border border-border-subtle px-2.5 py-1 text-xs text-text-secondary transition hover:border-brand/40 hover:text-brand"
                onClick={async () => {
                  const blob = await downloadWorkflowFile(workflowId, f.path)
                  const url = URL.createObjectURL(blob)
                  const a = document.createElement('a')
                  a.href = url
                  a.download = f.path.split('/').pop() ?? 'file.txt'
                  a.click()
                  URL.revokeObjectURL(url)
                }}
              >
                Download
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

/**
 * PublicShareModal — pick files, publish them under a fresh token (Submit), or
 * add more files to an existing token (Push).
 *
 * WHY push keeps the same token: the landing page integrates one URL per
 * published design. If pushing created a new token, the landing page would have
 * to be changed every time an admin adds a file. Submit creates the token once;
 * push is the "keep adding to the same array" half.
 */
export function PublicShareModal({
  workflowId,
  open,
  onClose,
}: {
  workflowId: string
  open: boolean
  onClose: () => void
}) {
  const [files, setFiles] = useState<WorkflowFileEntry[]>([])
  const [selected, setSelected] = useState<string[]>([])
  const [publications, setPublications] = useState<Publication[]>([])
  const [title, setTitle] = useState('')
  const [busy, setBusy] = useState(false)
  const [message, setMessage] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    let cancelled = false
    Promise.all([listWorkflowFiles(workflowId), listPublications(workflowId)])
      .then(([f, p]) => {
        if (cancelled) return
        setFiles(f)
        setPublications(p)
      })
      .catch(() => {
        if (!cancelled) setMessage('Could not load files or publications.')
      })
    return () => {
      cancelled = true
    }
  }, [open, workflowId])

  if (!open) return null

  const toggle = (path: string) =>
    setSelected((prev) => (prev.includes(path) ? prev.filter((p) => p !== path) : [...prev, path]))

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="flex max-h-[85vh] w-full max-w-2xl flex-col rounded-xl border border-border-subtle bg-card p-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-text-primary">Publish files to the landing page</h3>
          <button type="button" className="text-xs text-text-secondary hover:text-text-primary" onClick={onClose}>
            Close
          </button>
        </div>

        <input
          className="mt-3 w-full rounded-md border border-border-subtle bg-white/5 px-3 py-2 text-xs text-text-primary outline-none focus:border-brand/40"
          placeholder="Friendly name shown to the landing page (e.g. WhatsApp Design)"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
        />

        <div className="mt-3 min-h-0 flex-1 space-y-2 overflow-y-auto">
          {files.map((f) => (
            <label
              key={f.path}
              className="flex cursor-pointer items-center gap-3 rounded-md border border-border-subtle bg-white/5 px-3 py-2"
            >
              <input type="checkbox" checked={selected.includes(f.path)} onChange={() => toggle(f.path)} />
              <span className="min-w-0 flex-1 truncate text-xs text-text-primary">{f.path}</span>
            </label>
          ))}
          {files.length === 0 && (
            <p className="text-xs text-text-disabled">No files available to publish yet.</p>
          )}
        </div>

        {message && <p className="mt-2 text-xs text-text-secondary">{message}</p>}

        <div className="mt-3 flex items-center gap-2">
          <Button
            size="sm"
            disabled={busy || selected.length === 0}
            onClick={async () => {
              setBusy(true)
              setMessage(null)
              try {
                const created = await createPublication(workflowId, { title, paths: selected })
                setMessage(
                  `Published. Landing page URL: /api/v1/public/designs/${created.token}`
                )
                setPublications(await listPublications(workflowId))
              } catch {
                setMessage('Submit failed — nothing was published.')
              } finally {
                setBusy(false)
              }
            }}
          >
            Submit
          </Button>
          <Button
            size="sm"
            variant="secondary"
            disabled={busy || selected.length === 0 || publications.length === 0}
            onClick={async () => {
              setBusy(true)
              setMessage(null)
              try {
                const target = publications[0]
                if (!target) return
                const res = await pushToPublication(target.id, selected)
                setMessage(`Pushed ${res.files_added} file(s) into the existing share.`)
                setPublications(await listPublications(workflowId))
              } catch {
                setMessage('Push failed — the existing share is unchanged.')
              } finally {
                setBusy(false)
              }
            }}
          >
            Push
          </Button>
        </div>

        <div className="mt-4 border-t border-border-subtle pt-3">
          <p className="text-xs font-medium text-text-secondary">Active shares</p>
          {publications.length === 0 && (
            <p className="mt-1 text-xs text-text-disabled">Nothing published yet.</p>
          )}
          {publications.map((p) => (
            <div key={p.id} className="mt-2 flex items-center justify-between gap-3 text-xs">
              <span className="min-w-0 flex-1 truncate text-text-primary">
                {p.title} <span className="text-text-disabled">{p.file_count} file(s)</span>
              </span>
              <button
                type="button"
                className="flex-shrink-0 rounded-md border border-border-subtle px-2 py-1 text-text-secondary hover:border-brand/40 hover:text-brand"
                onClick={() => {
                  const url = `${window.location.origin}/api/v1/public/designs/${p.token}`
                  void navigator.clipboard?.writeText(url)
                  setMessage('Public URL copied.')
                }}
              >
                Copy URL
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
