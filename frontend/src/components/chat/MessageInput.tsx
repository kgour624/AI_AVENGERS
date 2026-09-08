import { useCallback, useRef, useState, type KeyboardEvent } from 'react'
import { useDropzone } from 'react-dropzone'
import { ExpertPicker } from '@/components/expert/ExpertPicker'
import { Button } from '@/components/ui/Button'
import { useReplyStore } from '@/stores/replyStore'
import type { ProjectExpert } from '@/types/project'
import { cn } from '@/utils/cn'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 9:
 * "1. Expert multi-select (checkboxes) 2. File upload (drag and drop
 * via react-dropzone) 3. Textarea with auto-resize 4. Send button
 * (disabled while streaming) 5. Keyboard shortcut: Cmd+Enter to send"
 *
 * Cross-questioned before writing:
 * - What if the user tries to send with zero experts selected? The
 *   backend has no concept of a message with no addressee - Send is
 *   disabled until at least one expert is checked, rather than
 *   letting a request go out that the backend would have to reject.
 * - What if the user tries to send while a previous message is still
 *   streaming? Send is disabled while `isSending` (passed from the
 *   parent, which owns streamStore's status) - a second concurrent
 *   sendMessage() call for the same chat would be nonsensical (which
 *   turn does the SSE response belong to?).
 * - What about pasting a huge block of text - should the textarea
 *   grow forever? Capped at max-height with overflow-y:auto past a
 *   point (via Tailwind's max-h + overflow), so a 500-line paste
 *   doesn't push the send button off-screen.
 * - Attaching a second file after already attaching one: the doc's
 *   SendMessageOptions only supports a single optional `file`, not an
 *   array - matches that exactly (react-dropzone's onDrop takes the
 *   LAST dropped file if multiple are dropped at once, since only one
 *   slot exists), rather than silently building multi-file support
 *   the backend contract doesn't actually accept.
 */
export interface MessageInputProps {
  chatId: string
  experts: ProjectExpert[]
  onSend: (
    message: string,
    expertIds: string[],
    file?: File,
    replyToMessageId?: string,
    includeFullThread?: boolean
  ) => void
  isSending: boolean
}

export function MessageInput({ chatId, experts, onSend, isSending }: MessageInputProps) {
  const [message, setMessage] = useState('')
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [attachedFile, setAttachedFile] = useState<File | null>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  // CT-D4: reply state, isolated per-chat via replyStore (CT-L10 —
  // does not touch streamStore or the message list at all).
  const replyState = useReplyStore((s) => s.replies.get(chatId))
  const setIncludeFullThread = useReplyStore((s) => s.setIncludeFullThread)
  const toggleLoopedInExpert = useReplyStore((s) => s.toggleLoopedInExpert)
  const clearReply = useReplyStore((s) => s.clearReply)

  const onDrop = useCallback((acceptedFiles: File[]) => {
    const file = acceptedFiles[acceptedFiles.length - 1]
    if (file) setAttachedFile(file)
  }, [])

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    noClick: true, // WHY noClick: the visible [\ud83d\udcce] attach button (not the whole textarea) opens the file picker - see the button below calling `open()` via getInputProps's ref trick would be redundant; instead we let users click the paperclip OR drag anywhere onto the input area.
    multiple: false,
  })

  function autoResize() {
    const el = textareaRef.current
    if (!el) return
    el.style.height = 'auto'
    // WHY cap at 200px rather than letting scrollHeight grow unbounded:
    // see the "huge paste" scenario in the header comment.
    el.style.height = `${Math.min(el.scrollHeight, 200)}px`
  }

  function handleSend() {
    const trimmed = message.trim()
    if (!trimmed || selectedIds.size === 0 || isSending) return

    onSend(
      trimmed,
      Array.from(selectedIds),
      attachedFile ?? undefined,
      replyState?.target.messageId,
      replyState?.includeFullThread
    )
    setMessage('')
    setAttachedFile(null)
    // CT-D4: clear the reply target on successful send — same moment
    // ChatPage.tsx clears pendingUserText, so "replying to" state never
    // outlives the turn it was drafted for.
    clearReply(chatId)
    // Reset height after clearing content - otherwise a tall textarea
    // from a long message stays tall even after the text is gone.
    requestAnimationFrame(() => {
      if (textareaRef.current) textareaRef.current.style.height = 'auto'
    })
  }

  function handleKeyDown(e: KeyboardEvent<HTMLTextAreaElement>) {
    // Cmd+Enter (Mac) / Ctrl+Enter (Windows/Linux) sends, per section 9.
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault()
      handleSend()
    }
  }

  const canSend = message.trim().length > 0 && selectedIds.size > 0 && !isSending

  return (
    <div className="border-t border-surface-border bg-surface-raised p-3">
      <ExpertPicker experts={experts} selectedIds={selectedIds} onChange={setSelectedIds} />

      {replyState && (
        <div className="mt-2 flex flex-col gap-2 rounded-md border border-brand/30 bg-brand/5 p-2">
          <div className="flex items-center justify-between">
            <p className="text-xs text-text-secondary">
              Replying to{replyState.target.expertName ? ` ${replyState.target.expertName}` : ''}:{' '}
              <span className="text-text-primary">
                &ldquo;{replyState.target.preview}&rdquo;
              </span>
            </p>
            <button
              type="button"
              onClick={() => clearReply(chatId)}
              className="text-xs text-text-secondary hover:text-text-primary"
            >
              Cancel reply
            </button>
          </div>

          <label className="flex items-center gap-2 text-xs text-text-secondary">
            <input
              type="checkbox"
              checked={replyState.includeFullThread}
              onChange={(e) => setIncludeFullThread(chatId, e.target.checked)}
            />
            Include full thread (off by default, pins only the message above)
          </label>

          {experts.filter((e) => e.expertId !== replyState.target.expertId).length > 0 && (
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-xs text-text-disabled">Loop in:</span>
              {experts
                .filter((e) => e.expertId !== replyState.target.expertId)
                .map((expert) => (
                  <label
                    key={expert.expertId}
                    className="flex items-center gap-1 text-xs text-text-secondary"
                  >
                    <input
                      type="checkbox"
                      checked={replyState.loopedInExpertIds.has(expert.expertId)}
                      onChange={() => toggleLoopedInExpert(chatId, expert.expertId)}
                    />
                    {expert.expertName}
                  </label>
                ))}
            </div>
          )}
        </div>
      )}

      <div
        {...getRootProps()}
        className={cn(
          'mt-2 flex items-end gap-2 rounded-lg border p-2',
          isDragActive ? 'border-brand bg-brand/5' : 'border-surface-border'
        )}
      >
        <input {...getInputProps()} />

        <textarea
          ref={textareaRef}
          value={message}
          onChange={(e) => {
            setMessage(e.target.value)
            autoResize()
          }}
          onKeyDown={handleKeyDown}
          placeholder={
            selectedIds.size === 0
              ? 'Select at least one expert to ask a question...'
              : 'Ask a follow-up question... (\u2318+Enter to send)'
          }
          rows={1}
          className="flex-1 resize-none bg-transparent text-sm text-text-primary placeholder:text-text-disabled focus:outline-none"
        />

        <button
          type="button"
          onClick={() => document.getElementById('message-file-input')?.click()}
          aria-label="Attach file"
          className="text-text-secondary hover:text-text-primary"
        >
          \ud83d\udcce
        </button>
        {/* Hidden native input, separate from react-dropzone's own (which has noClick set) - lets the paperclip button trigger a real file picker without also enabling click-anywhere-to-upload on the whole textarea row. */}
        <input
          id="message-file-input"
          type="file"
          className="hidden"
          onChange={(e) => {
            const file = e.target.files?.[0]
            if (file) setAttachedFile(file)
          }}
        />

        <Button onClick={handleSend} disabled={!canSend} size="sm">
          Send
        </Button>
      </div>

      {attachedFile && (
        <p className="mt-1 text-xs text-text-secondary">
          \ud83d\udcce {attachedFile.name}{' '}
          <button type="button" onClick={() => setAttachedFile(null)} className="underline">
            remove
          </button>
        </p>
      )}
    </div>
  )
}
