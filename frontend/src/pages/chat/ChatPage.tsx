import type { LoaderFunctionArgs } from 'react-router-dom'
import { useLoaderData } from 'react-router-dom'
import { getChat, getMessages } from '@/api/chats'
import { getProjectExperts } from '@/api/projects'
import type { ChatLoaderData } from '@/types/project'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 5 - the exact pattern shown
 * for ChatPage: parallel fetch of chat + messages + experts, abort
 * signal wired through from the router's request.
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
  const { chat, messages } = useLoaderData() as ChatLoaderData

  return (
    <div className="flex h-full flex-col p-6">
      <h2 className="mb-4 text-lg font-medium">{chat.title}</h2>
      <div className="flex-1 space-y-3 overflow-y-auto">
        {messages.map((m) => (
          <div key={m.id} className="rounded-md border border-surface-border p-3 text-sm">
            {m.content}
          </div>
        ))}
      </div>
      {/* MessageInput, ExpertResponse rendering, SSE wiring - Phase 2 */}
    </div>
  )
}
