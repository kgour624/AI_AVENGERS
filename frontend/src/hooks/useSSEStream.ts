import { useAuthStore } from '@/stores/authStore'
import { useStreamStore } from '@/stores/streamStore'
import { camelizeKeys } from '@/utils/casing'
import type { ExpertResponse, SynthesisResult } from '@/types/expert'

/**
 * SSE streaming via fetch() + ReadableStream (NOT EventSource).
 * Source: FRONTEND_SYSTEM_DESIGN.md section 8.
 *
 * WHY not EventSource: EventSource only supports GET requests and
 * cannot attach an Authorization header or send a file body - this
 * endpoint is POST, needs JWT auth, and optionally accepts a file
 * upload. fetch() + manual stream reading is the only way to get all
 * three.
 *
 * BUG FIXED vs the doc's illustrative snippet - partial SSE frames:
 * The doc's read loop does `decoder.decode(value).split('\n')` on
 * every single `reader.read()` result and processes each line
 * immediately. Traced through what actually happens on the wire: TCP
 * chunk boundaries do not align with SSE `data: ...\n\n` frame
 * boundaries. A single `read()` call can return a chunk that ends
 * mid-line, e.g. `...data: {"type": "complete", "exp` with the rest
 * of that JSON object arriving in the NEXT read(). The doc's snippet
 * would try `JSON.parse('{"type": "complete", "exp')`, get a parse
 * error, and the comment literally says "Malformed JSON - skip" -
 * silently DROPPING a real, complete event just because it happened
 * to straddle two reads. This is not a rare edge case; it's routine
 * over real networks, especially for larger citation/content payloads.
 *
 * Fix: maintain a rolling `buffer` string across reads. Only split
 * off and process a line once we've accumulated up to a `\n`; any
 * trailing partial line stays in the buffer for the next iteration.
 */

export interface SendMessageOptions {
  chatId: string
  message: string
  expertIds: string[]
  file?: File
  /**
   * CT-D4/CT-C1: optional reply target. undefined = fresh question,
   * the existing behavior for every message sent before this feature.
   * When set, becomes reply_to_message_id on the backend request
   * (message/handler.go's SendMessageRequest).
   */
  replyToMessageId?: string
  /** CT-L6: explicit opt-in only, ignored if replyToMessageId is unset. */
  includeFullThread?: boolean
}

type SSEEvent =
  | { type: 'thinking'; data: { message: string; experts: number } }
  | { type: 'complete'; data: ExpertResponse }
  | { type: 'synthesis'; data: SynthesisResult }
  | { type: 'done'; data: { turnNumber: number; durationMs: number } }
  | { type: 'error'; data: { message: string } }

function applyEvent(chatId: string, event: SSEEvent) {
  const { appendChunk, completeStream, setError } = useStreamStore.getState()

  switch (event.type) {
    case 'thinking':
      appendChunk(chatId, { status: 'thinking' })
      break
    case 'complete': {
      const current = useStreamStore.getState().activeStreams.get(chatId)
      const responses = current ? [...current.expertResponses] : []
      responses.push(event.data)
      appendChunk(chatId, { status: 'streaming', expertResponses: responses })
      break
    }
    case 'synthesis':
      appendChunk(chatId, { synthesis: event.data })
      break
    case 'done':
      completeStream(chatId)
      break
    case 'error':
      setError(chatId, event.data.message)
      break
  }
}

export function useSSEStream() {
  const { startStream, setError } = useStreamStore()

  const sendMessage = async (options: SendMessageOptions) => {
    const { chatId, message, expertIds, file, replyToMessageId, includeFullThread } = options

    startStream(chatId)

    let body: BodyInit
    const headers: HeadersInit = {}

    if (file) {
      const formData = new FormData()
      formData.append('message', message)
      formData.append('expert_ids', JSON.stringify(expertIds))
      formData.append('file', file)
      // CT-D4/CT-C1: multipart branch also needs reply fields threaded
      // through — message/handler.go's Send() reads these from
      // c.PostForm just like message/expert_ids on this same branch.
      if (replyToMessageId) formData.append('reply_to_message_id', replyToMessageId)
      if (includeFullThread) formData.append('include_full_thread', 'true')
      body = formData
      // WHY no Content-Type set for FormData: the browser sets the
      // multipart boundary itself. Setting it manually (as plain
      // 'multipart/form-data' with no boundary) breaks server-side
      // parsing - a mistake easy to make when copying the JSON branch.
    } else {
      body = JSON.stringify({
        message,
        expert_ids: expertIds,
        ...(replyToMessageId ? { reply_to_message_id: replyToMessageId } : {}),
        ...(includeFullThread ? { include_full_thread: true } : {}),
      })
      headers['Content-Type'] = 'application/json'
    }

    const token = useAuthStore.getState().accessToken
    if (token) headers['Authorization'] = `Bearer ${token}`

    let response: Response
    try {
      response = await fetch(`${import.meta.env.VITE_API_URL}/api/v1/chats/${chatId}/messages`, {
        method: 'POST',
        headers,
        body,
        credentials: 'include',
      })
    } catch (networkError) {
      // fetch() itself throws on network failure (offline, DNS, CORS
      // preflight rejection) - distinct from an HTTP error status.
      setError(chatId, `Network error: ${String(networkError)}`)
      return
    }

    if (!response.ok) {
      setError(chatId, `HTTP ${response.status}`)
      return
    }
    if (!response.body) {
      setError(chatId, 'No response body')
      return
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    try {
      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        // stream: true keeps multi-byte UTF-8 characters that straddle
        // chunk boundaries intact instead of emitting replacement chars.
        buffer += decoder.decode(value, { stream: true })

        // Only process complete lines; keep any trailing partial line
        // in the buffer for the next read() iteration.
        let newlineIndex: number
        while ((newlineIndex = buffer.indexOf('\n')) !== -1) {
          const line = buffer.slice(0, newlineIndex)
          buffer = buffer.slice(newlineIndex + 1)

          if (!line.startsWith('data: ')) continue
          const jsonStr = line.slice(6).trim()
          if (!jsonStr) continue

          try {
            const rawEvent = JSON.parse(jsonStr) as { type: string; data?: unknown }
            const event = camelizeKeys<SSEEvent>(rawEvent)
            applyEvent(chatId, event)
            if (event.type === 'done') return
          } catch {
            // A genuinely malformed line (not a partial-frame artifact,
            // since we now only parse complete lines) - skip it, but
            // this should be rare enough to be worth investigating if
            // it ever shows up in practice.
          }
        }
      }
    } catch (streamError) {
      setError(chatId, `Stream read error: ${String(streamError)}`)
    }
  }

  return { sendMessage }
}
