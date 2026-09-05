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
 * Cross-questioned: what file types does the backend actually accept?
 * AdminHandler.IngestTranscript only checks `header.Size > 50*1024*1024`
 * (size limit) - it does NOT validate file extension/MIME type
 * server-side at all. The wireframe says ".txt, .md" but there's no
 * backend enforcement of that. Restricting the dropzone's `accept` to
 * .txt/.md client-side is still worth doing (matches user intent from
 * the wireframe, prevents an accidental wrong-file upload), but it's
 * NOT a security boundary - documented so nobody mistakes this for
 * actual validation.
 */
export interface TranscriptUploadModalProps {
  isOpen: boolean
  onClose: () => void
  expertId: string
  onIngestStarted: (jobId: string) => void
}

export function TranscriptUploadModal({
  isOpen,
  onClose,
  expertId,
  onIngestStarted,
}: TranscriptUploadModalProps) {
  const [file, setFile] = useState<File | null>(null)

  const onDrop = useCallback((accepted: File[]) => {
    const f = accepted[0]
    if (f) setFile(f)
  }, [])

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    multiple: false,
    accept: { 'text/plain': ['.txt'], 'text/markdown': ['.md'] },
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
      return ingestTranscript(expertId, file)
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
          {file ? file.name : 'Drag & drop transcript here or click to browse (.txt, .md)'}
        </p>
      </div>

      {mutation.isError && (
        <p className="mt-2 text-sm text-mode-refuse">
          {mutation.error instanceof Error ? mutation.error.message : 'Upload failed'}
        </p>
      )}

      <div className="mt-4 flex justify-end gap-2">
        <Button variant="secondary" onClick={onClose} disabled={mutation.isPending}>
          Cancel
        </Button>
        <Button onClick={() => mutation.mutate()} disabled={!file} isLoading={mutation.isPending}>
          Start Ingestion
        </Button>
      </div>
    </Modal>
  )
}
