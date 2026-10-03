import { useState } from 'react'
import { Button } from '@/components/ui/Button'
import { downloadCombinedDesignDoc } from '@/api/workflows'
import type { DesignDocPointer } from '@/hooks/useFileStream'

/**
 * DownloadUnifiedDesignButton — serves the handoff-phase unified design
 * document (final_design.md / redesign_N.md) straight from the backend.
 *
 * WHY a separate component from DownloadDesignPackageButton:
 * - That button is client-side: it builds markdown from in-memory blackboard
 *   events and saves via Blob. No network request.
 * - This button is server-side: the document is assembled once at the handoff
 *   gate (design_doc.go:writeCombinedDesignDoc), written to workspace/main/,
 *   and served from disk via GET /workflows/:id/design-doc (publish.go).
 *   The click must hit the network, handle 404 vs 500 distinctly, and honor
 *   the filename the server chose (final_design.md vs redesign_2.md).
 *
 * Availability is driven by the file stream's latestDesignDoc pointer, which
 * arrives as file_design_doc SSE (files_sse.go) and is replayed from seq 0 on
 * reconnect/refresh. Until that pointer arrives the button is disabled rather
 * than hidden so the layout does not shift when handoff completes.
 */
export function DownloadUnifiedDesignButton({
  workflowId,
  designDoc,
}: {
  workflowId: string
  designDoc: DesignDocPointer | null
}) {
  const [isDownloading, setIsDownloading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const ready = designDoc !== null
  // Label reflects versioning: v0 = final_design.md, v>=1 = redesign_N.md.
  const label =
    designDoc == null
      ? 'Unified design — not ready'
      : designDoc.version === 0
        ? 'Download Unified Design'
        : `Download Unified Design (rev. ${designDoc.version})`
  const title = ready
    ? `Download ${designDoc.fileName} (v${designDoc.version}) — assembled at handoff`
    : 'Unified design document has not been generated yet (pre-handoff)'

  const handleDownload = async () => {
    if (!workflowId) return
    setIsDownloading(true)
    setError(null)
    try {
      const blob = await downloadCombinedDesignDoc(workflowId)
      // Prefer the pointer's filename; fallback to Content-Disposition is handled
      // by the browser when anchor.download is not set, but setting it makes the
      // save dialog correct even if the blob URL is revoked quickly.
      const fileName = designDoc?.fileName || 'final_design.md'
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = fileName
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      setTimeout(() => URL.revokeObjectURL(url), 1000)
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Download failed'
      // Axios wraps HTTP errors; 404 means the pointer exists but doc not yet
      // generated — keep the message human vs a raw stack.
      if (msg.includes('404')) {
        setError('Unified design not yet generated for this workflow.')
      } else {
        setError(msg)
      }
    } finally {
      setIsDownloading(false)
    }
  }

  return (
    <span className="inline-flex flex-col items-end gap-1">
      <Button
        variant="secondary"
        size="sm"
        isLoading={isDownloading}
        disabled={!ready}
        onClick={handleDownload}
        title={title}
      >
        {!isDownloading && (
          <span aria-hidden="true" className="mr-1">
            ⬇
          </span>
        )}
        {label}
      </Button>
      {error && <span className="text-[10px] text-mode-refuse">{error}</span>}
    </span>
  )
}
