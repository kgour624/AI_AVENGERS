import { useMemo, useRef, useState, useEffect } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import ReactMarkdown from 'react-markdown'
import type { LoaderFunctionArgs } from 'react-router-dom'
import { useLoaderData, useRevalidator, useParams, Link } from 'react-router-dom'
import { getChat, getMessages } from '@/api/chats'
import { getProjectExperts } from '@/api/projects'
import { deleteMessage, updateMessage } from '@/api/messages'
import type { ChatLoaderData } from '@/types/project'
import type { ExpertResponse as ExpertResponseType } from '@/types/expert'
import { useSSEStream } from '@/hooks/useSSEStream'
import { useStreamStore } from '@/stores/streamStore'
import { MessageInput } from '@/components/chat/MessageInput'
import { ExpertResponse } from '@/components/chat/ExpertResponse'
import { SynthesisPanel } from '@/components/chat/SynthesisPanel'
import { StreamingIndicator } from '@/components/chat/StreamingIndicator'
import { CodeBlock } from '@/components/chat/CodeBlock'
import { persistedMessageToExpertResponse } from '@/utils/adaptMessage'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 5 (loader pattern) and
 * section 10 (Chat Page wireframe: history + live streaming turn +
 * synthesis + MessageInput at the bottom).
 *
 * PHASE 6 ADDITION - virtualized message history: section 12
 * explicitly calls for "@tanstack/react-virtual for 100+ messages" but
 * no prior phase actually wired it up; this was plain .map() rendering
 * until now. Cross-questioned two real problems before implementing,
 * not just dropped the library in:
 *
 * 1. ExpertResponse items have genuinely variable height (markdown
 *    content length, presence of code blocks, citation lists, warning
 *    banners) - a fixed-size virtualizer (assuming every row is the
 *    same height) would produce wrong scroll positions and visible
 *    gaps/overlaps. Used `measureElement` (dynamic size measurement)
 *    instead of a fixed estimateSize, with a generous initial estimate
 *    (120px) that gets corrected once each row actually renders and
 *    is measured.
 * 2. Auto-scroll-to-bottom on new content must NOT yank the viewport
 *    down if the user has deliberately scrolled UP to read earlier
 *    history while a new SSE chunk streams in - that's a genuinely
 *    hostile UX pattern seen in worse chat UIs. Tracked via
 *    `isNearBottomRef` (updated on scroll), and only auto-scroll when
 *    that ref was true immediately before the update that triggered
 *    it, not unconditionally.
 *
 * The virtualizer only wraps the PERSISTED `messages` list (the part
 * that can genuinely grow to hundreds of items over a long chat's
 * lifetime) - the live streaming turn (pendingUserText, thinking
 * indicator, in-flight ExpertResponses, synthesis, error) renders as
 * normal DOM below the virtualized area, since it's always a small,
 * bounded, ephemeral set of items where virtualization would add
 * complexity for zero benefit.
 */
async function loader({ params, request }: LoaderFunctionArgs) {
  const signal = request.signal
  const projectId = params.projectId as string
  const chatId = params.chatId as string

  const [chat, messages, experts] = await Promise.all([
    getChat(chatId, { signal }),
    getMessages(chatId, { signal }),
    getProjectExperts(projectId, { signal }),
  ])

  return { chat, messages, experts }
}

export const chatRoute = { element: <ChatPage />, loader }

export default function ChatPage() {
  const { chat, messages, experts } = useLoaderData() as ChatLoaderData
  // WHY useParams here even though the loader already has projectId:
  // the loader's params are not exposed to the component via
  // useLoaderData - this is the standard React Router way for a
  // component to read its own route's params independently.
  const { projectId } = useParams<{ projectId: string }>()
  const { sendMessage } = useSSEStream()
  const revalidator = useRevalidator()
  const [pendingUserText, setPendingUserText] = useState<string | null>(null)
  const [editingMessageId, setEditingMessageId] = useState<string | null>(null)
  const [editingContent, setEditingContent] = useState('')
  const [deletingMessageId, setDeletingMessageId] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [showSearch, setShowSearch] = useState(false)

  const filteredMessages = useMemo(() => {
    if (!searchQuery.trim()) return messages
    const q = searchQuery.toLowerCase()
    return messages.filter((m) => m.content.toLowerCase().includes(q))
  }, [messages, searchQuery])

  async function handleDeleteMessage(messageId: string) {
    if (!window.confirm('Delete this message?')) return
    setDeletingMessageId(messageId)
    try {
      await deleteMessage(messageId)
      await revalidator.revalidate()
    } catch (err) {
      console.error('delete message failed', err)
    } finally {
      setDeletingMessageId(null)
    }
  }

  async function handleSaveEdit(messageId: string) {
    if (!editingContent.trim()) return
    try {
      await updateMessage(messageId, editingContent.trim())
      await revalidator.revalidate()
      setEditingMessageId(null)
    } catch (err) {
      console.error('update message failed', err)
    }
  }

  const stream = useStreamStore((s) => s.activeStreams.get(chat.id))
  const clearStream = useStreamStore((s) => s.clearStream)

  const isStreamActive = stream !== undefined && stream.status !== 'complete' && stream.status !== 'error'

  const scrollContainerRef = useRef<HTMLDivElement>(null)
  const isNearBottomRef = useRef(true)

  const expertsById = useMemo(() => new Map(experts.map((e) => [e.expertId, e])), [experts])

  const virtualizer = useVirtualizer({
    count: filteredMessages.length,
    getScrollElement: () => scrollContainerRef.current,
    estimateSize: () => 120,
    measureElement: (el) => el.getBoundingClientRect().height,
    overscan: 5,
  })

  // Track whether the user is scrolled near the bottom, so streaming
  // updates only auto-scroll when that was already true - see header
  // comment for why unconditional auto-scroll is a real UX hazard here.
  useEffect(() => {
    const el = scrollContainerRef.current
    if (!el) return
    function handleScroll() {
      const threshold = 80 // px - "close enough to bottom" tolerance
      isNearBottomRef.current = el!.scrollHeight - el!.scrollTop - el!.clientHeight < threshold
    }
    el.addEventListener('scroll', handleScroll)
    return () => el.removeEventListener('scroll', handleScroll)
  }, [])

  useEffect(() => {
    if (isNearBottomRef.current && scrollContainerRef.current) {
      scrollContainerRef.current.scrollTop = scrollContainerRef.current.scrollHeight
    }
    // Re-run whenever the streaming turn changes shape (new expert
    // response arrives, synthesis appears, etc.) or history length
    // changes (revalidated after completion) - both are moments new
    // content might now be below the fold.
  }, [messages.length, stream?.expertResponses.length, stream?.synthesis, pendingUserText])

  async function handleSend(
    text: string,
    expertIds: string[],
    file?: File,
    replyToMessageId?: string,
    includeFullThread?: boolean
  ) {
    setPendingUserText(text)
    await sendMessage({ chatId: chat.id, message: text, expertIds, file, replyToMessageId, includeFullThread })

    // WHY re-read stream status via getState() rather than the `stream`
    // closure variable: sendMessage's promise resolves after the SSE
    // loop returns, but this handler function's own closure captured
    // `stream` from BEFORE sendMessage ran - by the time we get here,
    // the store has almost certainly already moved past that stale
    // snapshot. Reading fresh state avoids acting on outdated status.
    const finalStatus = useStreamStore.getState().activeStreams.get(chat.id)?.status
    if (finalStatus === 'complete') {
      await revalidator.revalidate()
      clearStream(chat.id)
      setPendingUserText(null)
    }
    // WHY no action on 'error': the error message stays visible via
    // StreamingIndicator/stream.error until the user tries again or
    // navigates away - clearing it immediately would hide a failure
    // the user needs to see and act on.
  }

  return (
    <div className="flex h-full flex-col overflow-hidden">
      <div className="flex items-center gap-3 border-b border-surface-border px-6 py-4">
        <Link
          to={`/projects/${projectId}`}
          className="text-sm text-text-secondary hover:text-text-primary"
        >
          {'\u2190'} Back to Project
        </Link>
        <h2 className="flex-1 text-lg font-medium">{chat.title}</h2>
        {showSearch ? (
          <div className="flex items-center gap-2">
            <input
              autoFocus
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search messages..."
              className="rounded-md border border-surface-border bg-surface-overlay px-3 py-1.5 text-sm text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-brand/40"
            />
            <button
              onClick={() => { setShowSearch(false); setSearchQuery('') }}
              className="text-xs text-text-secondary hover:text-text-primary"
            >{'\u2715'}</button>
          </div>
        ) : (
          <button
            onClick={() => setShowSearch(true)}
            className="text-sm text-text-secondary hover:text-text-primary"
            title="Search messages"
          >{'\ud83d\udd0d'}</button>
        )}
        {searchQuery && (
          <span className="text-xs text-text-disabled">{filteredMessages.length} result{filteredMessages.length !== 1 ? 's' : ''}</span>
        )}
      </div>

      <div ref={scrollContainerRef} className="flex-1 overflow-y-auto p-6">
        <div style={{ height: virtualizer.getTotalSize(), position: 'relative' }}>
          {virtualizer.getVirtualItems().map((virtualRow) => {
            const m = messages[virtualRow.index]!
            return (
              <div
                key={m.id}
                ref={virtualizer.measureElement}
                data-index={virtualRow.index}
                className="pb-3"
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  transform: `translateY(${virtualRow.start}px)`,
                }}
              >
                {m.role === 'user' ? (
                  <div className="group ml-auto max-w-[80%]">
                    {editingMessageId === m.id ? (
                      <div className="flex flex-col gap-2">
                        <textarea
                          className="w-full rounded-md bg-brand/10 p-3 text-sm text-text-primary resize-none border border-brand/30 focus:outline-none focus:border-brand"
                          rows={3}
                          value={editingContent}
                          onChange={(e) => setEditingContent(e.target.value)}
                          onKeyDown={(e) => {
                            if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSaveEdit(m.id) }
                            if (e.key === 'Escape') setEditingMessageId(null)
                          }}
                          autoFocus
                        />
                        <div className="flex gap-2 justify-end">
                          <button onClick={() => setEditingMessageId(null)} className="text-xs text-text-secondary hover:text-text-primary px-2 py-1">Cancel</button>
                          <button onClick={() => handleSaveEdit(m.id)} className="text-xs bg-brand/20 hover:bg-brand/30 text-text-primary px-3 py-1 rounded">Save</button>
                        </div>
                      </div>
                    ) : (
                      <div className="relative">
                        <div className="rounded-md bg-brand/10 p-3 text-sm text-text-primary">{m.content}</div>
                        <div className="absolute -top-6 right-0 hidden group-hover:flex gap-1">
                          <button
                            onClick={() => { setEditingMessageId(m.id); setEditingContent(m.content) }}
                            className="text-xs text-text-secondary hover:text-text-primary bg-surface-raised border border-surface-border rounded px-2 py-0.5"
                          >Edit</button>
                          <button
                            onClick={() => handleDeleteMessage(m.id)}
                            disabled={deletingMessageId === m.id}
                            className="text-xs text-mode-refuse hover:text-mode-refuse/80 bg-surface-raised border border-surface-border rounded px-2 py-0.5 disabled:opacity-50"
                          >{deletingMessageId === m.id ? '...' : 'Delete'}</button>
                        </div>
                      </div>
                    )}
                  </div>
                ) : (
                  (() => {
                    const asExpertResponse = persistedMessageToExpertResponse(m, experts)
                    if (!asExpertResponse) {
                      // Shouldn't happen for a well-formed assistant row, but
                      // a malformed one (missing expertId) degrades to plain
                      // text rather than being silently dropped.
                      return (
                        <div className="rounded-md border border-surface-border p-3 text-sm">
                          {m.content}
                        </div>
                      )
                    }
                    return <ExpertResponse response={asExpertResponse} persistedMessageId={m.id} chatId={chat.id} />
                  })()
                )}
              </div>
            )
          })}
        </div>

        {/* Live streaming turn - deliberately NOT virtualized, see header comment. */}
        <div className="mt-3 space-y-3">
          {pendingUserText && (
            <div className="ml-auto max-w-[80%] rounded-md bg-brand/10 p-3 text-sm text-text-primary">
              {pendingUserText}
            </div>
          )}

          {stream && isStreamActive && (
            <>
              {/* Show thinking indicator only when no content has arrived yet */}
              {stream.status === 'thinking' && !stream.streamingContent && (
                <StreamingIndicator stream={stream} expertCount={experts.length} />
              )}

              {/* Live streaming content — shown as tokens arrive via chunk events.
                  Replaced by ExpertResponse once complete event fires. */}
              {stream.streamingContent && (
                <div className="rounded-lg border-l-4 border-l-mode-advise border-y border-r border-glass-border bg-surface-raised/80 p-4 backdrop-blur-xl">
                  <div className="prose prose-invert prose-sm max-w-none text-text-primary">
                    <ReactMarkdown components={{ code: CodeBlock }}>
                      {stream.streamingContent}
                    </ReactMarkdown>
                  </div>
                  <p className="mt-2 text-xs text-text-disabled animate-pulse">Streaming...</p>
                </div>
              )}

              {stream.expertResponses.map((partial, i) =>
                partial.expertId ? (
                  <ExpertResponse key={partial.expertId} response={partial as ExpertResponseType} isStreaming />
                ) : (
                  <p key={i} className="text-xs text-text-disabled">Malformed response - missing expertId</p>
                )
              )}
            </>
          )}

          {stream?.synthesis && <SynthesisPanel synthesis={stream.synthesis} />}

          {stream?.status === 'error' && (
            <p className="rounded-md border border-mode-refuse/30 bg-mode-refuse/5 p-3 text-sm text-mode-refuse">
              {stream.error}
            </p>
          )}
        </div>
      </div>

      <MessageInput chatId={chat.id} experts={experts} onSend={handleSend} isSending={isStreamActive} />
    </div>
  )
}
