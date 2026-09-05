import { useState } from 'react'
import type { LoaderFunctionArgs } from 'react-router-dom'
import { useLoaderData, useRevalidator } from 'react-router-dom'
import { getChat, getMessages } from '@/api/chats'
import { getProjectExperts } from '@/api/projects'
import type { ChatLoaderData } from '@/types/project'
import type { ExpertResponse as ExpertResponseType } from '@/types/expert'
import { useSSEStream } from '@/hooks/useSSEStream'
import { useStreamStore } from '@/stores/streamStore'
import { MessageInput } from '@/components/chat/MessageInput'
import { ExpertResponse } from '@/components/chat/ExpertResponse'
import { SynthesisPanel } from '@/components/chat/SynthesisPanel'
import { StreamingIndicator } from '@/components/chat/StreamingIndicator'
import { persistedMessageToExpertResponse } from '@/utils/adaptMessage'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 5 (loader pattern) and
 * section 10 (Chat Page wireframe: history + live streaming turn +
 * synthesis + MessageInput at the bottom).
 *
 * Cross-questioned before wiring streamStore + loader data together:
 * - What happens to the live stream's state when the user navigates
 *   away and back to the SAME chat mid-stream? React Router's loader
 *   re-runs on navigation, refetching messages/chat/experts fresh -
 *   but streamStore's Map is keyed by chatId and persists independent
 *   of this component's mount/unmount (it's a module-level Zustand
 *   store, not component state). So a stream that was mid-flight
 *   survives a quick navigate-away-and-back. This is intentional, not
 *   an accident of implementation - matches the product's "persistent
 *   context" principle (section 1: "Kiro-like: persistent context").
 * - What happens once the stream reaches 'complete'? We call
 *   revalidator.revalidate() to re-run the loader and fetch the newly
 *   persisted message(s), THEN clearStream() to remove the transient
 *   SSE state - otherwise the completed turn would render twice
 *   (once from the live stream, once from the refetched message list)
 *   until the user reloads. Ordering matters here: clearing the
 *   stream BEFORE the revalidated messages arrive would cause a
 *   visible flash where the completed turn briefly disappears -
 *   avoided by only clearing after revalidation resolves.
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
  const { sendMessage } = useSSEStream()
  const revalidator = useRevalidator()
  const [pendingUserText, setPendingUserText] = useState<string | null>(null)

  const stream = useStreamStore((s) => s.activeStreams.get(chat.id))
  const clearStream = useStreamStore((s) => s.clearStream)

  const isStreamActive = stream !== undefined && stream.status !== 'complete' && stream.status !== 'error'

  async function handleSend(text: string, expertIds: string[], file?: File) {
    setPendingUserText(text)
    await sendMessage({ chatId: chat.id, message: text, expertIds, file })

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
    <div className="flex h-full flex-col">
      <h2 className="border-b border-surface-border px-6 py-4 text-lg font-medium">{chat.title}</h2>

      <div className="flex-1 space-y-3 overflow-y-auto p-6">
        {messages.map((m) => {
          if (m.role === 'user') {
            return (
              <div key={m.id} className="ml-auto max-w-[80%] rounded-md bg-brand/10 p-3 text-sm text-text-primary">
                {m.content}
              </div>
            )
          }
          const asExpertResponse = persistedMessageToExpertResponse(m, experts)
          if (!asExpertResponse) {
            // Shouldn't happen for a well-formed assistant row, but a
            // malformed one (missing expertId) degrades to plain text
            // rather than being silently dropped from the transcript.
            return (
              <div key={m.id} className="rounded-md border border-surface-border p-3 text-sm">
                {m.content}
              </div>
            )
          }
          return <ExpertResponse key={m.id} response={asExpertResponse} persistedMessageId={m.id} />
        })}

        {pendingUserText && (
          <div className="ml-auto max-w-[80%] rounded-md bg-brand/10 p-3 text-sm text-text-primary">
            {pendingUserText}
          </div>
        )}

        {stream && isStreamActive && (
          <>
            {stream.status === 'thinking' && (
              <StreamingIndicator stream={stream} expertCount={experts.length} />
            )}
            {stream.expertResponses.map((partial, i) =>
              // Live SSE 'complete' events send a full ExpertResponse per
              // the documented SSEEvent type (hooks/useSSEStream.ts) - the
              // Partial<> wrapper on StreamState.expertResponses exists
              // for the 'thinking' phase where nothing has arrived yet,
              // not because a 'complete' payload is ever partial itself.
              // This cast reflects that: by the time an entry exists in
              // this array at all, it came from a 'complete' event.
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

      <MessageInput experts={experts} onSend={handleSend} isSending={isStreamActive} />
    </div>
  )
}
