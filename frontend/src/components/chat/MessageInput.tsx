import { useCallback, useEffect, useRef, useState, type KeyboardEvent } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useDropzone } from 'react-dropzone'
import { getExpertAnswerFormats } from '@/api/experts'
import { ExpertPicker } from '@/components/expert/ExpertPicker'
import { Button } from '@/components/ui/Button'
import { Tooltip } from '@/components/ui/Tooltip'
import { useReplyStore } from '@/stores/replyStore'
import type { ProjectExpert } from '@/types/project'
import { cn } from '@/utils/cn'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 9:
 * "1. Expert multi-select (checkboxes) 2. File upload (drag and drop
 * via react-dropzone) 3. Textarea with auto-resize 4. Send button
 * (disabled while streaming) 5. Keyboard shortcut: Cmd+Enter to send"
 *
 * Feature #6 fix (docs bug list): Added persistent expert selection
 * with "Lock Selection" toggle. When locked, selected experts persist
 * across messages within the same chat session. Selection is stored
 * in localStorage per chat ID.
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

// localStorage keys for persistent selection
function getSelectionKey(chatId: string) {
  return `chat_${chatId}_expert_selection`
}

function getLockKey(chatId: string) {
  return `chat_${chatId}_selection_locked`
}

// Load persisted selection from localStorage
function loadPersistedSelection(chatId: string): Set<string> {
  try {
    const stored = localStorage.getItem(getSelectionKey(chatId))
    if (stored) {
      return new Set(JSON.parse(stored))
    }
  } catch (err) {
    console.warn('Failed to load persisted expert selection', err)
  }
  return new Set()
}

// Load lock state from localStorage
function loadLockState(chatId: string): boolean {
  try {
    const stored = localStorage.getItem(getLockKey(chatId))
    return stored === 'true'
  } catch (err) {
    console.warn('Failed to load lock state', err)
  }
  return false
}

// Save selection to localStorage
function saveSelection(chatId: string, selectedIds: Set<string>) {
  try {
    localStorage.setItem(getSelectionKey(chatId), JSON.stringify(Array.from(selectedIds)))
  } catch (err) {
    console.warn('Failed to save expert selection', err)
  }
}

// Save lock state to localStorage
function saveLockState(chatId: string, locked: boolean) {
  try {
    localStorage.setItem(getLockKey(chatId), locked.toString())
  } catch (err) {
    console.warn('Failed to save lock state', err)
  }
}

export interface MessageInputProps {
  chatId: string
  experts: ProjectExpert[]
  onSend: (
    message: string,
    expertIds: string[],
    file?: File,
    replyToMessageId?: string,
    includeFullThread?: boolean,
    templateName?: string
  ) => void
  isSending: boolean
}

export function MessageInput({ chatId, experts, onSend, isSending }: MessageInputProps) {
  const [message, setMessage] = useState('')
  const [attachedFile, setAttachedFile] = useState<File | null>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  // Feature #6: Persistent expert selection with lock toggle
  const [isLocked, setIsLocked] = useState(() => loadLockState(chatId))
  const [selectedIds, setSelectedIds] = useState<Set<string>>(() => {
    // Load persisted selection if locked, otherwise start empty
    return isLocked ? loadPersistedSelection(chatId) : new Set()
  })

  // CT-D4: reply state, isolated per-chat via replyStore (CT-L10 —
  // does not touch streamStore or the message list at all).
  // T-CAT: per-question answer format. Formats come from the FIRST selected
  // expert's category — if several experts are selected with different
  // categories, the first is authoritative for this picker (the backend still
  // applies each expert's own category default when none is chosen).
  const firstSelectedId = Array.from(selectedIds)[0]
  const { data: formatInfo } = useQuery({
    queryKey: ['expert-answer-formats', firstSelectedId],
    queryFn: () => getExpertAnswerFormats(firstSelectedId!),
    enabled: Boolean(firstSelectedId),
  })
  const formats = formatInfo?.formats ?? []
  const defaultFormat = formatInfo?.default ?? ''
  const [selectedFormat, setSelectedFormat] = useState('')

  // Reset the picker whenever the expert set changes; the backend default
  // applies until the user picks something.
  useEffect(() => {
    setSelectedFormat('')
  }, [firstSelectedId])

  const replyState = useReplyStore((s) => s.replies.get(chatId))
  // replyTarget: hoisted so the null-check is a single, well-narrowed binding.
  // TS does not carry `replyState.target` narrowing into the .filter()
  // callbacks below, but a stable local + optional chaining does.
  const replyTarget = replyState?.target ?? null
  const setIncludeFullThread = useReplyStore((s) => s.setIncludeFullThread)
  const toggleLoopedInExpert = useReplyStore((s) => s.toggleLoopedInExpert)
  const clearReply = useReplyStore((s) => s.clearReply)

  // Persist selection whenever it changes (if locked)
  useEffect(() => {
    if (isLocked) {
      saveSelection(chatId, selectedIds)
    }
  }, [chatId, selectedIds, isLocked])

  // Persist lock state whenever it changes
  useEffect(() => {
    saveLockState(chatId, isLocked)
  }, [chatId, isLocked])

  // Handle selection changes
  function handleSelectionChange(newSelection: Set<string>) {
    setSelectedIds(newSelection)
  }

  // Toggle lock state
  function toggleLock() {
    const newLockState = !isLocked
    setIsLocked(newLockState)
    
    if (newLockState) {
      // When locking, save current selection
      saveSelection(chatId, selectedIds)
    } else {
      // When unlocking, optionally clear selection
      // (keeping it for now - user can manually deselect if needed)
    }
  }

  const onDrop = useCallback((acceptedFiles: File[]) => {
    const file = acceptedFiles[acceptedFiles.length - 1]
    if (file) setAttachedFile(file)
  }, [])

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    noClick: true, // WHY noClick: the visible [📎] attach button (not the whole textarea) opens the file picker - see the button below calling `open()` via getInputProps's ref trick would be redundant; instead we let users click the paperclip OR drag anywhere onto the input area.
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

    // T-CAT: only send a format when the user (or the category default) actually
    // chose one; empty = backend uses the category default.
    const formatToSend =
      (selectedFormat && formats.some((f) => f === selectedFormat) ? selectedFormat : undefined) ??
      formats.find((f) => f.toLowerCase() === defaultFormat.toLowerCase()) ??
      undefined

    onSend(
      trimmed,
      Array.from(selectedIds),
      attachedFile ?? undefined,
      replyTarget?.messageId,
      replyState?.includeFullThread,
      formatToSend
    )
    setMessage('')
    setAttachedFile(null)
    
    // Feature #6: Only clear selection if NOT locked
    if (!isLocked) {
      setSelectedIds(new Set())
    }
    
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

  // Feature #19: Keyboard shortcuts help content
  // WHY multi-line tooltip: shows all shortcuts at once, not just one
  // WHY platform-specific: Mac users see ⌘, Windows/Linux users see Ctrl
  const isMac = typeof navigator !== 'undefined' && navigator.platform.toUpperCase().indexOf('MAC') >= 0
  const modKey = isMac ? '⌘' : 'Ctrl'
  const keyboardShortcuts = [
    `${modKey}+Enter: Send message`,
    'Shift+Enter: New line',
    'Esc: Cancel reply (when replying)',
  ].join('\n')

  return (
    <div className="border-t border-surface-border bg-surface-raised p-3">
      <div className="flex items-center justify-between">
        <ExpertPicker experts={experts} selectedIds={selectedIds} onChange={handleSelectionChange} />
        
        <div className="flex items-center gap-3">
          {/* Feature #19: Keyboard Shortcut Help
              WHY corner placement: non-intrusive, discoverable
              WHY ⌨️ icon: universal symbol for keyboard shortcuts
              WHY tooltip not modal: quick reference, no interruption */}
          <Tooltip content={<pre className="text-xs whitespace-pre-wrap">{keyboardShortcuts}</pre>}>
            <button
              type="button"
              className="text-text-secondary hover:text-text-primary transition-colors"
              aria-label="Keyboard shortcuts"
            >
              ⌨️
            </button>
          </Tooltip>

          {/* Feature #6: Lock Selection toggle */}
          <label className="flex items-center gap-2 text-xs text-text-secondary hover:text-text-primary cursor-pointer">
            <input
              type="checkbox"
              checked={isLocked}
              onChange={toggleLock}
              className="cursor-pointer"
            />
            <span className="flex items-center gap-1">
              {isLocked ? '🔒' : '🔓'}
              <span>Lock Selection</span>
            </span>
          </label>
        </div>
      </div>

      {replyState && replyTarget && (
        <div className="mt-2 flex flex-col gap-2 rounded-md border border-brand/30 bg-brand/5 p-2">
          <div className="flex items-center justify-between">
            <p className="text-xs text-text-secondary">
              Replying to{replyTarget.expertName ? ` ${replyTarget.expertName}` : ''}:{' '}
              <span className="text-text-primary">
                &ldquo;{replyTarget.preview}&rdquo;
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

          {experts.filter((e) => e.expertId !== replyTarget?.expertId).length > 0 && (
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-xs text-text-disabled">Loop in:</span>
              {experts
                .filter((e) => e.expertId !== replyTarget?.expertId)
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

        {formats.length >= 2 && (
          <div className="mb-1 flex items-center gap-2 text-xs text-text-secondary">
            <span>Answer format</span>
            <select
              value={selectedFormat || defaultFormat}
              onChange={(e) => setSelectedFormat(e.target.value)}
              className="rounded-md border border-surface-border bg-surface-overlay px-2 py-1 text-xs text-text-primary focus:outline-none focus:ring-2 focus:ring-brand"
            >
              {formats.map((f) => (
                <option key={f} value={f}>
                  {f}
                </option>
              ))}
            </select>
          </div>
        )}

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
              : isLocked
              ? 'Ask a follow-up question... (selection locked, ⌘+Enter to send)'
              : 'Ask a follow-up question... (⌘+Enter to send)'
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
          📎
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
          📎 {attachedFile.name}{' '}
          <button type="button" onClick={() => setAttachedFile(null)} className="underline">
            remove
          </button>
        </p>
      )}
      
      {isLocked && selectedIds.size > 0 && (
        <p className="mt-1 text-xs text-text-disabled">
          🔒 Selection locked: {selectedIds.size} expert{selectedIds.size !== 1 ? 's' : ''} will be used for all messages
        </p>
      )}
    </div>
  )
}
