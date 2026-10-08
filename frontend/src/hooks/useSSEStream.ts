import { useAuthStore } from '@/stores/authStore'
import { useStreamStore } from '@/stores/streamStore'
import { refreshSession } from '@/api/base'
import { camelizeKeys } from '@/utils/casing'
import type { ExpertResponse, SynthesisResult, CollabSection } from '@/types/expert'

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
  /** One or more attachments. Each is extracted to text by the backend. */
  files?: File[]
  /**
   * T-GEN: generic knowledge ceiling for this message (0-30). 0/absent keeps
   * the strict China Wall; >0 lets the expert answer uncovered parts from
   * general knowledge, tagged [GENERIC].
   */
  genericAllowancePct?: number
  /**
   * CT-D4/CT-C1: optional reply target. undefined = fresh question,
   * the existing behavior for every message sent before this feature.
   * When set, becomes reply_to_message_id on the backend request
   * (message/handler.go's SendMessageRequest).
   */
  replyToMessageId?: string
  /**
   * T-CAT: optional named answer format of the expert's category ("Code",
   * "Approach"). undefined = the category's default, which is what every
   * pre-existing caller sends.
   */
  templateName?: string
  /** CT-L6: explicit opt-in only, ignored if replyToMessageId is unset. */
  includeFullThread?: boolean
  /**
   * Collaborative Relay: '' / 'independent' (default, every pre-existing
   * caller) runs each selected expert in parallel exactly as before.
   * 'collaborative' requires 2+ expertIds and instead runs them
   * SEQUENTIALLY as one merged answer (message/handler.go's AnswerMode).
   * Unset/unknown values fail safe to independent mode on the backend,
   * never to an error.
   */
  answerMode?: 'independent' | 'collaborative'
  resumeRunId?: string
  /** Optional AbortSignal from caller to cancel stream on WiFi drop / unmount. */
  signal?: AbortSignal
}

type SSEEvent =
  | { type: 'thinking'; data: { message: string; experts: number } }
  | { type: 'chunk'; data: { content: string; expertId: string } }
  | { type: 'complete'; data: ExpertResponse }
  | { type: 'synthesis'; data: SynthesisResult }
  // Collaborative Relay: one finished section, sent as each expert's turn
  // mode never sends this event.
  | { type: 'collab_section'; data: CollabSection }
  | { type: 'relay_step'; data: import('@/types/expert').RelayStepEvent }
  | { type: 'done'; data: { turnNumber: number; durationMs: number } }
  | { type: 'error'; data: { message: string } }

function applyEvent(chatId: string, event: SSEEvent) {
  const { appendChunk, completeStream, setError } = useStreamStore.getState()

  switch (event.type) {
    case 'thinking':
      appendChunk(chatId, { status: 'thinking' })
      break
    case 'chunk': {
      // Accumulate streaming tokens into streamingContent.
      // This is what the user sees while the answer is being typed out.
      // content:"" heartbeat tokens are filtered in message/handler.go
      // before reaching SSE, but guard here too for safety.
      if (!event.data.content) break
      const current = useStreamStore.getState().activeStreams.get(chatId)
      const prev = current?.streamingContent ?? ''
      appendChunk(chatId, {
        status: 'streaming',
        streamingContent: prev + event.data.content,
      })
      break
    }
    case 'complete': {
      const current = useStreamStore.getState().activeStreams.get(chatId)
      const responses = current ? [...current.expertResponses] : []
      responses.push(event.data)
      // Clear streamingContent — ExpertResponse component takes over rendering.
      appendChunk(chatId, { status: 'streaming', expertResponses: responses, streamingContent: '' })
      break
    }
    case 'synthesis':
      appendChunk(chatId, { synthesis: event.data })
      break
    case 'collab_section': {
      // Collaborative Relay: append in the order sections arrive, which IS
      // relay order - never re-sort this.
      const current = useStreamStore.getState().activeStreams.get(chatId)
      const sections = current ? [...current.collabSections, event.data] : [event.data]
      appendChunk(chatId, { status: 'streaming', collabSections: sections })
      break
    }
    case 'done':
      completeStream(chatId)
      break
    case 'error':
      setError(chatId, event.data.message)
      break
    case 'relay_step':
      useStreamStore.getState().appendRelayStep(chatId, event.data)
      break
  }
}

export function useSSEStream() {
  const { startStream, setError } = useStreamStore()

  const sendMessage = async (options: SendMessageOptions) => {
    const {
      chatId,
      message,
      expertIds,
      files,
      replyToMessageId,
      includeFullThread,
      templateName,
      genericAllowancePct,
      answerMode,
      resumeRunId,
    } = options
    const hasFiles = Boolean(files && files.length > 0)

    startStream(chatId)

    let body: BodyInit
    const headers: HeadersInit = {}

    if (hasFiles) {
      const formData = new FormData()
      formData.append('message', message)
      formData.append('expert_ids', JSON.stringify(expertIds))
      // Repeated "file" entries — the backend loops over form.File["file"].
      for (const f of files ?? []) formData.append('file', f)
      if (genericAllowancePct) formData.append('generic_allowance_pct', String(genericAllowancePct))
      // CT-D4/CT-C1: multipart branch also needs reply fields threaded
      // through — message/handler.go's Send() reads these from
      // c.PostForm just like message/expert_ids on this same branch.
      if (replyToMessageId) formData.append('reply_to_message_id', replyToMessageId)
      if (includeFullThread) formData.append('include_full_thread', 'true')
      if (templateName) formData.append('template_name', templateName)
      if (answerMode) formData.append('answer_mode', answerMode)
      if (resumeRunId) formData.append('resume_run_id', resumeRunId)
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
        ...(templateName ? { template_name: templateName } : {}),
        ...(genericAllowancePct ? { generic_allowance_pct: genericAllowancePct } : {}),
        ...(answerMode ? { answer_mode: answerMode } : {}),
        ...(resumeRunId ? { resume_run_id: resumeRunId } : {}),
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
        signal: options.signal,
      })
    } catch (networkError: any) {
      // WiFi drop / AbortSignal cancel: show as abort, backend already cancelled via ctx
      if (networkError?.name === 'AbortError') {
        setError(chatId, 'ABORTED: WiFi band / request cancel — server processing bhi band ho gaya (cost bach gaya)')
        return
      }
      setError(chatId, `Network error: ${String(networkError)}`)
      return
    }

    // The stream uses raw fetch, so the axios 401-refresh interceptor never runs
    // for it. When the session token expires mid-answer the stream was simply
    // cut with "HTTP 401" and the client lost the response. Refresh once and
    // retry the same request, which is what kept failing in production.
    if (response.status === 401) {
      try {
        await refreshSession()
        const refreshed = useAuthStore.getState().accessToken
        if (refreshed) headers['Authorization'] = `Bearer ${refreshed}`
        const retry = await fetch(
          `${import.meta.env.VITE_API_URL}/api/v1/chats/${chatId}/messages`,
          { method: 'POST', headers, body, credentials: 'include', signal: options.signal },
        )
        if (retry.ok) {
          response = retry
        }
      } catch {
        // fall through to the normal error path below
      }
    }

    if (!response.ok) {
      // Fix: Jawab shuru hone se pehle Error -> read server JSON for proper code/message, show Error Card with Retry
      let msg = `HTTP ${response.status}`;
      let code = `HTTP_${response.status}`;
      try {
        const text = await response.text();
        if (text) {
          const j = JSON.parse(text);
          // backend uses {code, message} or {error}
          if (j.code) code = j.code;
          if (j.message) msg = j.message;
          else if (j.error) msg = j.error;
          else if (j.data?.message) msg = j.data.message;
          else msg = text.slice(0, 400);
        }
      } catch {}
      // Map to user-friendly Hindi hint for retryable errors (DB down, validation)
      if (response.status >= 500) msg = `${msg} — Retry kar sakte hain`;
      setError(chatId, `${code}: ${msg}`)
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
