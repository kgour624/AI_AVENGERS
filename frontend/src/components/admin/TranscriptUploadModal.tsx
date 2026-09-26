import { useState, useCallback } from 'react'
import { useDropzone } from 'react-dropzone'
import { useMutation } from '@tanstack/react-query'
import { ingestTranscript } from '@/api/admin'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { cn } from '@/utils/cn'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 11 ("Upload Transcript
 * Modal" wireframe: drag-and-drop .txt/.md, [Cancel] [Start Ingestion]).
 *
 * UPDATED (D3/D4): the backend now accepts a full document allowlist — every
 * upload is converted to text before ingestion (PDF/DOCX/XLSX/... via the ML
 * sidecar; .txt/.md decoded inline). The server validates the extension and
 * returns 400 UNSUPPORTED_FORMAT with the supported list, and the dropzone's
 * `accept` mirrors that same list so a wrong file is caught before the network
 * round-trip. The client list is a UX filter; the server list is the boundary.
 */
export const ACCEPTED_DOCUMENT_EXTENSIONS = [
  '.txt', '.text', '.md', '.markdown',
  '.pdf',
  '.docx', '.docm', '.doc',
  '.xlsx', '.xlsm', '.xls', '.csv', '.tsv',
  '.pptx', '.ppt',
  '.rtf', '.epub',
  '.html', '.htm', '.xhtml', '.xml',
  '.odt', '.ods', '.odp',
  '.json', '.log',
  '.srt', '.vtt',
]

/** Human-readable summary for the dropzone hint. */
const ACCEPT_HINT =
  'PDF, Word, Excel, PowerPoint, CSV, text/markdown, HTML, RTF, EPUB, ODT, subtitles'

/**
 * Extensions that are converted by the document extractor before training.
 * The rest (.txt/.md/.json/.log/.xml) are plain text and go straight in.
 * Mirrors the backend's split (plain text is decoded in the API process so it
 * keeps working when the ML sidecar is unavailable).
 */
const NEEDS_CONVERSION = new Set([
  'pdf', 'doc', 'docx', 'docm',
  'xls', 'xlsx', 'xlsm', 'csv', 'tsv',
  'ppt', 'pptx',
  'rtf', 'epub', 'html', 'htm', 'xhtml', 'xml',
  'odt', 'ods', 'odp',
  'srt', 'vtt',
])

export interface TranscriptUploadModalProps {
  isOpen: boolean
  onClose: () => void
  expertId: string
  onIngestStarted: (jobId: string) => void
}

/**
 * Surfaces the backend's own message — a rejected upload returns
 * 400 { error: { code: "UNSUPPORTED_FORMAT", message: "… Supported: .csv, …" } },
 * which is far more useful than axios's "Request failed with status code 400".
 */
function describeUploadError(err: unknown): string {
  const apiMessage = (err as { response?: { data?: { error?: { message?: string } } } })
    ?.response?.data?.error?.message
  if (apiMessage) return apiMessage
  if (err instanceof Error && err.message) return err.message
  return 'Upload failed'
}

export function TranscriptUploadModal({
  isOpen,
  onClose,
  expertId,
  onIngestStarted,
}: TranscriptUploadModalProps) {
  const [file, setFile] = useState<File | null>(null)
  // Append is the default and the safe choice: it adds this document to whatever
  // the expert already knows. Replacing throws the existing corpus away first,
  // and it is the ONLY way to apply a changed chunker without doubling the
  // corpus — so it is offered, but never as the default.
  const [replaceExisting, setReplaceExisting] = useState(false)

  const onDrop = useCallback((accepted: File[]) => {
    const f = accepted[0]
    if (f) setFile(f)
  }, [])

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    multiple: false,
    // WHY the map spans several MIME types: react-dropzone matches on MIME, and
    // the office/document MIME types vary by OS and browser (a .docx can arrive
    // as application/vnd.openxmlformats-officedocument.wordprocessingml.document
    // or as application/octet-stream). Grouping by families keeps the file
    // picker usable without pretending to be a security boundary.
    accept: {
      'text/plain': ['.txt', '.text', '.log'],
      'text/markdown': ['.md', '.markdown'],
      'text/csv': ['.csv', '.tsv'],
      'text/html': ['.html', '.htm', '.xhtml'],
      'text/xml': ['.xml'],
      'application/json': ['.json'],
      'text/vtt': ['.vtt'],
      'application/x-subrip': ['.srt'],
      'application/pdf': ['.pdf'],
      'application/msword': ['.doc'],
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx', '.docm'],
      'application/vnd.ms-excel': ['.xls'],
      'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': ['.xlsx', '.xlsm'],
      'application/vnd.ms-powerpoint': ['.ppt'],
      'application/vnd.openxmlformats-officedocument.presentationml.presentation': ['.pptx'],
      'application/rtf': ['.rtf'],
      'application/epub+zip': ['.epub'],
      'application/vnd.oasis.opendocument.text': ['.odt'],
      'application/vnd.oasis.opendocument.spreadsheet': ['.ods'],
      'application/vnd.oasis.opendocument.presentation': ['.odp'],
      // Catch-all for the office/document formats browsers report as generic
      // binary (common on Windows). The server still does the real validation.
      'application/octet-stream': ACCEPTED_DOCUMENT_EXTENSIONS,
    },
  })

  const mutation = useMutation({
    mutationFn: () => {
      if (!file) throw new Error('No file selected')
      // WHY the 50MB check is duplicated here even though the backend
      // also enforces it: failing fast client-side avoids uploading a
      // multi-hundred-MB file over the network only to have it rejected
      // after the fact - the backend's check is the real boundary, this
      // is purely a UX optimization, not a substitute for it.
      if (file.size > 50 * 1024 * 1024) {
        throw new Error('File must be under 50MB')
      }
      return ingestTranscript(expertId, file, replaceExisting)
    },
    onSuccess: (data) => {
      onIngestStarted(data.jobId)
      setFile(null)
      onClose()
    },
  })

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <h3 className="mb-4 text-sm font-medium text-text-primary">Upload Transcript</h3>

      <div
        {...getRootProps()}
        className={cn(
          'flex flex-col items-center justify-center rounded-lg border-2 border-dashed p-8 text-center',
          isDragActive ? 'border-brand bg-brand/5' : 'border-surface-border'
        )}
      >
        <input {...getInputProps()} />
        <p className="text-2xl">\ud83d\udcc4</p>
        <p className="mt-2 text-sm text-text-secondary">
          {file ? file.name : 'Drag & drop a document here or click to browse'}
        </p>
        {!file && (
          <p className="mt-1 text-[11px] text-text-disabled">{ACCEPT_HINT}</p>
        )}
      </div>

      {file && (
        <p className="mt-2 text-[11px] text-text-disabled">
          {NEEDS_CONVERSION.has(file.name.split('.').pop()?.toLowerCase() ?? '')
            ? 'This document will be converted to text before training starts.'
            : 'Plain text — ingested as-is.'}
        </p>
      )}

      <label className="mt-3 flex items-start gap-2 cursor-pointer">
        <input
          type="checkbox"
          className="mt-0.5"
          checked={replaceExisting}
          onChange={(e) => setReplaceExisting(e.target.checked)}
        />
        <span className="text-[11px] text-text-secondary">
          Replace the existing corpus instead of adding to it.
          <span className="mt-0.5 block text-text-disabled">
            Use this when the expert's content has been re-chunked or you are re-uploading
            the same course: adding a second copy leaves both in the corpus and the expert
            retrieves from duplicates. Everything this expert has learned from previously
            uploaded documents is removed.
          </span>
        </span>
      </label>

      {mutation.isError && (
        <p className="mt-2 text-sm text-mode-refuse">{describeUploadError(mutation.error)}</p>
      )}

      <div className="mt-4 flex justify-end gap-2">
        <Button variant="secondary" onClick={onClose} disabled={mutation.isPending}>
          Cancel
        </Button>
        <Button onClick={() => mutation.mutate()} disabled={!file} isLoading={mutation.isPending}>
          {replaceExisting ? 'Replace & Train' : 'Start Ingestion'}
        </Button>
      </div>
    </Modal>
  )
}
