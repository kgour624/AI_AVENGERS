import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createWorkflowChat,
  listWorkflowChats,
  getWorkflowChat,
  listWorkflowChatMessages,
  sendWorkflowChatMessage,
  addChatParticipant,
  removeChatParticipant,
  setKnowledgeMode,
  proposeChange,
} from '@/api/workflowChat'
import { queryKeys } from '@/api/queryKeys'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { handleAPIError } from '@/utils/errors'
import { cn } from '@/utils/cn'

/**
 * WorkflowChatPanel — §6 of docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md.
 *
 * Deliberately its own component, not a reuse of ChatPage/ExpertResponse (the
 * PRODUCT chat's UI): the backend keeps WorkflowChatService entirely separate
 * from chat.Service (different tables, no shared code path), and the product
 * chat's components assume SSE streaming (useSSEStream, streamStore) which
 * this chat does not use — WorkflowChatService.Send is synchronous, it runs
 * its own tool loop server-side and returns the final message in one response.
 *
 * GENERIC_OPTIONS mirrors KanbanPage.tsx's ApprovalGate — same 0/5/10/20/30
 * ladder, same 30% ceiling (MaxGenericAllowancePct / migration 017).
 */
const GENERIC_OPTIONS = [0, 5, 10, 20, 30]

export interface WorkflowChatExpertOption {
  id: string
  name: string
}

interface Props {
  workflowId: string
  // Experts actually in this workflow (from the Kanban task list, so the
  // participant picker only ever offers experts this workflow really has —
  // not the global expert catalogue).
  availableExperts: WorkflowChatExpertOption[]
}

function CreateChatForm({
  workflowId,
  onCreated,
}: {
  workflowId: string
  onCreated: (chatId: string) => void
}) {
  const [title, setTitle] = useState('')
  const queryClient = useQueryClient()

  const createMut = useMutation({
    mutationFn: () => createWorkflowChat(workflowId, { title: title.trim() || 'New conversation' }),
    onSuccess: (chat) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workflows.chats(workflowId) })
      setTitle('')
      onCreated(chat.id)
    },
  })

  return (
    <div className="flex gap-2">
      <Input
        placeholder="e.g. Questions about the API contract"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        className="flex-1"
      />
      <Button size="sm" isLoading={createMut.isPending} onClick={() => createMut.mutate()}>
        + New chat
      </Button>
    </div>
  )
}

function ParticipantsRow({
  chatId,
  availableExperts,
}: {
  chatId: string
  availableExperts: WorkflowChatExpertOption[]
}) {
  const queryClient = useQueryClient()
  const { data } = useQuery({
    queryKey: queryKeys.workflowChats.detail(chatId),
    queryFn: () => getWorkflowChat(chatId),
  })
  const participants = data?.participants ?? []
  const genericPct = data?.chat.genericAllowancePct

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: queryKeys.workflowChats.detail(chatId) })

  const addMut = useMutation({
    mutationFn: (expertId: string) => addChatParticipant(chatId, expertId),
    onSuccess: invalidate,
  })
  const removeMut = useMutation({
    mutationFn: (expertId: string) => removeChatParticipant(chatId, expertId),
    onSuccess: invalidate,
  })
  const modeMut = useMutation({
    mutationFn: (pct: number | null) => setKnowledgeMode(chatId, pct),
    onSuccess: invalidate,
  })

  const notYetIn = availableExperts.filter(
    (e) => !participants.some((p) => p.expertId === e.id)
  )

  return (
    <div className="flex flex-wrap items-center gap-2 border-b border-surface-border px-3 py-2">
      {participants.map((p) => (
        <Badge key={p.expertId} variant="brand">
          {p.expertName}
          <button
            type="button"
            onClick={() => removeMut.mutate(p.expertId)}
            className="ml-1.5 text-text-disabled hover:text-mode-refuse"
            aria-label={`Remove ${p.expertName}`}
          >
            \u00d7
          </button>
        </Badge>
      ))}

      {notYetIn.length > 0 && (
        <select
          value=""
          onChange={(e) => e.target.value && addMut.mutate(e.target.value)}
          className="rounded border border-surface-overlay bg-surface-base px-2 py-1 text-xs text-text-secondary"
        >
          <option value="">+ add expert...</option>
          {notYetIn.map((e) => (
            <option key={e.id} value={e.id}>{e.name}</option>
          ))}
        </select>
      )}

      <div className="ml-auto flex items-center gap-1.5">
        <span className="text-[10px] uppercase tracking-wide text-text-disabled">Knowledge</span>
        <select
          value={genericPct ?? ''}
          onChange={(e) => modeMut.mutate(e.target.value === '' ? null : Number(e.target.value))}
          className="rounded border border-surface-overlay bg-surface-base px-2 py-1 text-xs text-text-secondary"
        >
          <option value="">workflow default</option>
          {GENERIC_OPTIONS.map((pct) => (
            <option key={pct} value={pct}>
              {pct === 0 ? '0% — trained only' : `${pct}% generic`}
            </option>
          ))}
        </select>
      </div>
    </div>
  )
}

// ProposeChangeBar sits below the send row, collapsed by default.
function ProposeChangeBar({ chatId, workflowId }: { chatId: string; workflowId: string }) {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [goal, setGoal] = useState('')
  const [submitted, setSubmitted] = useState<string | null>(null)
  const [proposeError, setProposeError] = useState<string | null>(null)

  const proposeMut = useMutation({
    mutationFn: () => proposeChange(chatId, goal.trim()),
    onSuccess: (cr) => {
      setSubmitted(cr.id)
      setGoal('')
      setOpen(false)
      setProposeError(null)
      queryClient.invalidateQueries({ queryKey: queryKeys.workflowChats.messages(chatId) })
      queryClient.invalidateQueries({ queryKey: queryKeys.workflows.blackboard(workflowId) })
    },
    onError: (err) => setProposeError(handleAPIError(err)),
  })

  if (submitted) {
    return (
      <div className="flex items-center gap-2 border-t border-glow-purple/20 bg-glow-purple/5 px-3 py-2">
        <span className="flex-1 text-xs text-glow-purple">
          Change request submitted. Experts will re-run their sections and you will be asked to approve.
        </span>
        <button type="button" onClick={() => setSubmitted(null)}
          className="text-[10px] text-text-disabled hover:text-text-secondary">dismiss</button>
      </div>
    )
  }

  if (!open) {
    return (
      <div className="border-t border-surface-border px-3 py-1.5">
        <button type="button" onClick={() => setOpen(true)}
          className="text-[11px] text-text-disabled hover:text-glow-purple transition-colors">
          Propose a design change
        </button>
      </div>
    )
  }

  return (
    <div className="border-t border-glow-purple/30 bg-glow-purple/5 p-2">
      <p className="mb-1 text-[10px] font-semibold uppercase tracking-wide text-glow-purple">Propose a design change</p>
      <p className="mb-2 text-[11px] text-text-disabled">
        Describe what you want changed. Relevant experts will re-run their sections and you will approve the result.
      </p>
      <div className="flex gap-2">
        <textarea
          value={goal}
          onChange={(e) => setGoal(e.target.value)}
          onKeyDown={(e) => {
            if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
              e.preventDefault()
              if (goal.trim()) proposeMut.mutate()
            }
          }}
          placeholder="e.g. Add a bulk-move endpoint to the file manager API"
          rows={2}
          className="flex-1 resize-none rounded border border-glow-purple/30 bg-surface-base px-2 py-1.5 text-xs text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-1 focus:ring-glow-purple"
        />
        <div className="flex flex-col gap-1">
          <Button size="sm" isLoading={proposeMut.isPending}
            disabled={!goal.trim() || proposeMut.isPending}
            onClick={() => proposeMut.mutate()}>Submit</Button>
          <button type="button"
            onClick={() => { setOpen(false); setGoal(''); setProposeError(null) }}
            className="rounded px-2 py-1 text-[11px] text-text-disabled hover:text-text-secondary">Cancel</button>
        </div>
      </div>
      {proposeError && <p className="mt-1.5 text-xs text-mode-refuse">{proposeError}</p>}
    </div>
  )
}

function MessageBubble({
  message,
}: {
  message: import('@/api/workflowChat').WorkflowChatMessage
}) {
  if (message.role === 'user') {
    const isChangeReq = message.content.startsWith('[CHANGE REQUEST]')
    return (
      <div className={cn(
        'ml-auto max-w-[85%] rounded-md p-3 text-sm text-text-primary',
        isChangeReq ? 'border border-glow-purple/40 bg-glow-purple/10' : 'bg-brand/10'
      )}>
        {isChangeReq && (
          <span className="mb-1 block text-[10px] font-semibold uppercase tracking-wide text-glow-purple">Change request</span>
        )}
        {isChangeReq ? message.content.replace('[CHANGE REQUEST] ', '') : message.content}
      </div>
    )
  }

  // An intermediate tool-call step (§7.3) — shown compactly, not as a normal
  // answer bubble, so the client can see what the expert did without the tool
  // I/O crowding out the actual answer.
  if (message.toolName) {
    return (
      <div className="max-w-[85%] rounded-md border border-glass-border bg-surface-overlay/40 px-3 py-2 text-xs text-text-disabled">
        <span className="font-mono text-text-secondary">{message.toolName}</span>
        {message.content && <p className="mt-1 text-text-secondary">{message.content}</p>}
      </div>
    )
  }

  const generic = message.gateResult
  return (
    <div className="max-w-[85%] rounded-md border border-glass-border bg-surface-raised/80 p-3 text-sm text-text-primary">
      <div className="mb-1 flex items-center gap-2">
        <span className="text-xs font-semibold text-glow-purple">
          {message.expertName ?? 'Expert'}
        </span>
        {generic && (
          <Badge variant={generic.genericBlocked === false ? 'warning' : 'neutral'}>
            {generic.genericBlocked === false
              ? `generic ${generic.genericAllowancePct ?? 0}%`
              : 'trained only'}
          </Badge>
        )}
      </div>
      <p className="whitespace-pre-wrap">{message.content}</p>
    </div>
  )
}

function ChatThread({
  chatId,
  workflowId,
  availableExperts,
}: {
  chatId: string
  workflowId: string
  availableExperts: WorkflowChatExpertOption[]
}) {
  const queryClient = useQueryClient()
  const [draft, setDraft] = useState('')
  const [toExpertId, setToExpertId] = useState('')
  const [error, setError] = useState<string | null>(null)

  const { data: messages = [], isLoading } = useQuery({
    queryKey: queryKeys.workflowChats.messages(chatId),
    queryFn: () => listWorkflowChatMessages(chatId),
  })

  // Send is synchronous server-side (the tool loop runs before it responds),
  // so there is no streaming state to manage here — just a pending mutation.
  const sendMut = useMutation({
    mutationFn: () => sendWorkflowChatMessage(chatId, draft.trim(), toExpertId || undefined),
    onSuccess: () => {
      setDraft('')
      setError(null)
      queryClient.invalidateQueries({ queryKey: queryKeys.workflowChats.messages(chatId) })
      // A tool call may have proposed an amendment or posted a blackboard
      // event — both lists should reflect it once the answer lands.
      queryClient.invalidateQueries({ queryKey: queryKeys.workflows.amendments(workflowId, 'pending') })
      queryClient.invalidateQueries({ queryKey: queryKeys.workflows.blackboard(workflowId) })
    },
    onError: (err) => setError(handleAPIError(err)),
  })

  return (
    // min-h-0 on both the column and the scroll area: a flex child defaults to
    // min-height:auto, so without it the message list grows the page instead of
    // scrolling and a long chat has no scrollbar at all. Cap the panel height so
    // the composer stays reachable.
    <div className="flex h-full max-h-[70vh] min-h-0 flex-col">
      <ParticipantsRow chatId={chatId} availableExperts={availableExperts} />

      <div className="min-h-0 flex-1 space-y-2 overflow-y-auto p-3">
        {isLoading && <p className="text-xs text-text-disabled">Loading...</p>}
        {!isLoading && messages.length === 0 && (
          <p className="text-center text-xs text-text-disabled">
            Ask a question about this deliverable.
          </p>
        )}
        {messages.map((m) => <MessageBubble key={m.id} message={m} />)}
      </div>

      {error && (
        <p className="border-t border-mode-refuse/30 bg-mode-refuse/5 px-3 py-2 text-xs text-mode-refuse">
          {error}
        </p>
      )}

      <div className="flex items-end gap-2 border-t border-surface-border p-2">
        {availableExperts.length > 0 && (
          <select
            value={toExpertId}
            onChange={(e) => setToExpertId(e.target.value)}
            className="rounded border border-surface-overlay bg-surface-base px-1.5 py-1.5 text-xs text-text-secondary"
            title="Which expert answers (default: the deliverable's author)"
          >
            <option value="">default</option>
            {availableExperts.map((e) => (
              <option key={e.id} value={e.id}>{e.name}</option>
            ))}
          </select>
        )}
        <textarea
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={(e) => {
            if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
              e.preventDefault()
              if (draft.trim()) sendMut.mutate()
            }
          }}
          placeholder="Ask about this deliverable... (\u2318+Enter to send)"
          rows={1}
          className="flex-1 resize-none rounded border border-surface-overlay bg-surface-base px-2 py-1.5 text-xs text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-1 focus:ring-brand"
        />
        <Button
          size="sm"
          isLoading={sendMut.isPending}
          disabled={!draft.trim() || sendMut.isPending}
          onClick={() => sendMut.mutate()}
        >
          Send
        </Button>
      </div>

      <ProposeChangeBar chatId={chatId} workflowId={workflowId} />
    </div>
  )
}

export function WorkflowChatPanel({ workflowId, availableExperts }: Props) {
  const [activeChatId, setActiveChatId] = useState<string | null>(null)

  const { data: chats = [], isLoading } = useQuery({
    queryKey: queryKeys.workflows.chats(workflowId),
    queryFn: () => listWorkflowChats(workflowId),
  })

  const activeChat = activeChatId ? chats.find((c) => c.id === activeChatId) : undefined

  return (
    <div className="mt-6">
      <div className="mb-2 flex items-center justify-between">
        <span className="text-xs font-semibold uppercase tracking-wide text-text-secondary">
          Chat with the experts ({chats.length})
        </span>
      </div>

      <Card className="p-0">
        <div className="grid grid-cols-3" style={{ minHeight: 360 }}>
          <div className="col-span-1 flex flex-col border-r border-surface-border">
            <div className="border-b border-surface-border p-2">
              <CreateChatForm workflowId={workflowId} onCreated={setActiveChatId} />
            </div>
            <div className="flex-1 overflow-y-auto">
              {isLoading && <p className="p-3 text-xs text-text-disabled">Loading...</p>}
              {!isLoading && chats.length === 0 && (
                <p className="p-3 text-center text-xs text-text-disabled">
                  No conversations yet — start one above.
                </p>
              )}
              {chats.map((c) => (
                <button
                  key={c.id}
                  onClick={() => setActiveChatId(c.id)}
                  className={cn(
                    'block w-full px-3 py-2 text-left text-xs',
                    activeChatId === c.id
                      ? 'bg-surface-overlay text-text-primary'
                      : 'text-text-secondary hover:bg-surface-overlay/60'
                  )}
                >
                  <span className="block truncate font-medium">{c.title}</span>
                  <span className="block text-[10px] text-text-disabled">
                    {c.messageCount} message{c.messageCount === 1 ? '' : 's'}
                  </span>
                </button>
              ))}
            </div>
          </div>

          <div className="col-span-2">
            {activeChat ? (
              <ChatThread
                key={activeChat.id}
                chatId={activeChat.id}
                workflowId={workflowId}
                availableExperts={availableExperts}
              />
            ) : (
              <div className="flex h-full items-center justify-center">
                <p className="text-xs text-text-disabled">Select or start a conversation</p>
              </div>
            )}
          </div>
        </div>
      </Card>
    </div>
  )
}
