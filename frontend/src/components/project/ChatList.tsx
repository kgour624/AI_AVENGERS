import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { getChats, createChat, updateChatTitle, archiveChat, unarchiveChat, permanentDeleteChat, getMessages } from '@/api/chats'
import type { Chat, Message } from '@/types/project'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { Modal } from '@/components/ui/Modal'

function ConfirmDeleteChatModal({
  chatTitle, onConfirm, onCancel, isPending,
}: { chatTitle: string; onConfirm: () => void; onCancel: () => void; isPending: boolean }) {
  const [typed, setTyped] = useState('')
  const match = typed.trim() === chatTitle.trim()
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="w-full max-w-sm rounded-xl border border-glass-border bg-surface-raised p-6 shadow-2xl">
        <h3 className="mb-2 text-sm font-semibold text-red-400">Delete Chat</h3>
        <p className="mb-4 text-xs text-text-secondary">
          Permanently delete{' '}
          <span className="font-medium text-text-primary">{chatTitle}</span>{' '}
          and all its messages. Type the chat title to confirm.
        </p>
        <input
          autoFocus
          value={typed}
          onChange={(e) => setTyped(e.target.value)}
          placeholder={chatTitle}
          className="w-full rounded-md border border-glass-border bg-surface-overlay px-3 py-2 text-sm text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-2 focus:ring-red-500/40"
        />
        <div className="mt-4 flex justify-end gap-2">
          <button onClick={onCancel} className="rounded-md px-3 py-1.5 text-xs text-text-secondary hover:text-text-primary">Cancel</button>
          <button
            disabled={!match || isPending}
            onClick={onConfirm}
            className="rounded-md bg-red-600/80 px-3 py-1.5 text-xs font-medium text-white disabled:opacity-40 hover:bg-red-600"
          >
            {isPending ? 'Deleting...' : 'Delete permanently'}
          </button>
        </div>
      </div>
    </div>
  )
}

function ChatRow({
  chat, projectId, onChanged, selected, onToggle,
}: {
  chat: Chat
  projectId: string
  onChanged: () => void
  selected: boolean
  onToggle: (chatId: string) => void
}) {
  const [isEditing, setIsEditing] = useState(false)
  const [title, setTitle] = useState(chat.title)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  const renameMutation = useMutation({
    mutationFn: () => updateChatTitle(chat.id, title.trim()),
    onSuccess: () => { setIsEditing(false); onChanged() },
  })
  const archiveMutation = useMutation({ mutationFn: () => archiveChat(chat.id), onSuccess: onChanged })
  const unarchiveMutation = useMutation({ mutationFn: () => unarchiveChat(chat.id), onSuccess: onChanged })
  const deleteMutation = useMutation({
    mutationFn: () => permanentDeleteChat(chat.id),
    onSuccess: () => { setShowDeleteConfirm(false); onChanged() },
  })

  if (isEditing) {
    return (
      <div className="flex items-center gap-1 px-2 py-1">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          autoFocus
          className="flex-1 rounded-md border border-surface-border bg-surface-overlay px-2 py-1 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-brand"
          onKeyDown={(e) => {
            if (e.key === 'Enter' && title.trim()) renameMutation.mutate()
            if (e.key === 'Escape') setIsEditing(false)
          }}
        />
        <button className="text-xs text-brand" disabled={!title.trim() || renameMutation.isPending} onClick={() => renameMutation.mutate()}>Save</button>
        <button className="text-xs text-text-secondary" onClick={() => setIsEditing(false)}>Cancel</button>
      </div>
    )
  }

  return (
    <>
      <div className={`group flex items-center gap-1 rounded-md px-2 py-1.5 hover:bg-surface-overlay${chat.isArchived ? ' opacity-50' : ''}`}>
        {/* Per-chat selection. What is selected here is what the export button
            below turns into one markdown file — nothing else is downloaded. */}
        <input
          type="checkbox"
          checked={selected}
          onChange={() => onToggle(chat.id)}
          aria-label={`Select chat ${chat.title}`}
          title="Select for export"
          className="h-3.5 w-3.5 flex-shrink-0 cursor-pointer accent-brand"
        />
        <Link
          to={`/projects/${projectId}/chats/${chat.id}`}
          className="flex-1 truncate text-sm text-text-secondary hover:text-text-primary"
        >
          {chat.isArchived && <span className="mr-1 text-[10px] text-text-disabled">[off]</span>}
          {chat.title}
        </Link>
        <div className="invisible flex items-center gap-0.5 group-hover:visible">
          <button aria-label="Rename" title="Rename" onClick={() => setIsEditing(true)}
            className="text-xs text-text-disabled hover:text-text-primary">
            {'\u270f\ufe0f'}
          </button>
          {chat.isArchived ? (
            <button aria-label="Enable" title="Enable" disabled={unarchiveMutation.isPending}
              onClick={() => unarchiveMutation.mutate()}
              className="text-xs text-text-disabled hover:text-glow-cyan">
              {'\u21a9'}
            </button>
          ) : (
            <button aria-label="Disable" title="Disable" disabled={archiveMutation.isPending}
              onClick={() => archiveMutation.mutate()}
              className="text-xs text-text-disabled hover:text-amber-400">
              {'\u23f8'}
            </button>
          )}
          <button aria-label="Delete permanently" title="Delete permanently"
            onClick={() => setShowDeleteConfirm(true)}
            className="text-xs text-text-disabled hover:text-red-400">
            {'\u2715'}
          </button>
        </div>
      </div>
      {showDeleteConfirm && (
        <ConfirmDeleteChatModal
          chatTitle={chat.title}
          onConfirm={() => deleteMutation.mutate()}
          onCancel={() => setShowDeleteConfirm(false)}
          isPending={deleteMutation.isPending}
        />
      )}
    </>
  )
}

/**
 * One chat as markdown: the question and every answer, in turn order.
 *
 * WHY a flat transcript and not a "Q:"/"A:" summary: the export is what the user
 * reads outside the app, so it keeps the same content they saw — expert name and
 * mode per answer included.
 */
function renderChatMarkdown(chat: Chat, messages: Message[]): string {
  const lines: string[] = [`# ${chat.title}`, '']
  if (messages.length === 0) {
    lines.push('_No messages in this chat._')
    return lines.join('\n')
  }
  for (const m of messages) {
    const who = m.role === 'user' ? 'You' : m.role === 'assistant' ? 'Expert' : m.role
    const meta: string[] = []
    if (m.expertId) meta.push(`expert: ${m.expertId}`)
    if (m.decisionMode) meta.push(`mode: ${m.decisionMode}`)
    lines.push(`## ${who}${meta.length ? ` (${meta.join(', ')})` : ''}`, '', m.content ?? '', '')
  }
  return lines.join('\n')
}

function downloadMarkdown(filename: string, markdown: string) {
  const blob = new Blob([markdown], { type: 'text/markdown;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

export function ChatList({ projectId }: { projectId: string }) {
  const queryClient = useQueryClient()
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [newTitle, setNewTitle] = useState('')

  const { data: chats, isLoading } = useQuery({
    queryKey: ['projects', projectId, 'chats'],
    queryFn: () => getChats(projectId),
  })

  const createMutation = useMutation({
    mutationFn: (title: string) => createChat(projectId, title),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'chats'] })
      setNewTitle('')
      setIsModalOpen(false)
    },
  })

  // Export selection. Kept as a Set so a chat disappearing from the list cannot
  // leave a stale id behind that the download would then try to fetch.
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [isExporting, setIsExporting] = useState(false)

  const refresh = () => queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'chats'] })

  const toggleSelected = (chatId: string) =>
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (next.has(chatId)) next.delete(chatId)
      else next.add(chatId)
      return next
    })

  // Selected chats (and their answers) are fetched one by one and rendered into
  // a single markdown file. Chats are exported in the order they are shown, so
  // the file reads like the list the user was looking at.
  const handleExport = async () => {
    if (allChats.length === 0 || selectedIds.size === 0) return
    setIsExporting(true)
    try {
      const chosen = allChats.filter((c) => selectedIds.has(c.id))
      const parts: string[] = []
      for (const chat of chosen) {
        const messages = await getMessages(chat.id)
        parts.push(renderChatMarkdown(chat, messages))
      }
      const stamp = new Date().toISOString().slice(0, 10)
      downloadMarkdown(`expert-chats-${stamp}.md`, parts.join('\n\n---\n\n'))
    } finally {
      setIsExporting(false)
    }
  }
  const activeChats = chats?.filter((c) => !c.isArchived) ?? []
  const archivedChats = chats?.filter((c) => c.isArchived) ?? []
  const allChats = [...activeChats, ...archivedChats]

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <p className="text-sm font-medium text-text-secondary">Chats</p>
        <Button variant="secondary" size="sm" onClick={() => setIsModalOpen(true)}>+ New Chat</Button>
      </div>

      {selectedIds.size > 0 && (
        <div className="mb-2 flex items-center justify-between gap-2 rounded-md border border-glass-border bg-surface-overlay px-2 py-1.5">
          <span className="text-xs text-text-secondary">{selectedIds.size} selected</span>
          <div className="flex items-center gap-2">
            <button
              type="button"
              className="text-xs text-text-disabled hover:text-text-primary"
              onClick={() => setSelectedIds(new Set())}
            >
              Clear
            </button>
            <Button size="sm" isLoading={isExporting} onClick={handleExport}>
              Download .md
            </Button>
          </div>
        </div>
      )}

      {isLoading && (
        <div className="space-y-2">{Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-8" />)}</div>
      )}

      <div className="space-y-1">
        {activeChats.map((chat) => <ChatRow key={chat.id} chat={chat} projectId={projectId} onChanged={refresh} selected={selectedIds.has(chat.id)} onToggle={toggleSelected} />)}
        {!isLoading && activeChats.length === 0 && <p className="text-sm text-text-disabled">No chats yet.</p>}
      </div>

      {archivedChats.length > 0 && (
        <div className="mt-4">
          <p className="mb-1 text-[10px] font-semibold uppercase tracking-wider text-text-disabled/50">Disabled</p>
          <div className="space-y-1">
            {archivedChats.map((chat) => <ChatRow key={chat.id} chat={chat} projectId={projectId} onChanged={refresh} selected={selectedIds.has(chat.id)} onToggle={toggleSelected} />)}
          </div>
        </div>
      )}

      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)}>
        <h3 className="mb-4 text-sm font-medium text-text-primary">New Chat</h3>
        <Input label="Title" value={newTitle} onChange={(e) => setNewTitle(e.target.value)}
          placeholder="e.g. Database Schema Discussion" autoFocus />
        <div className="mt-4 flex justify-end gap-2">
          <Button variant="secondary" onClick={() => setIsModalOpen(false)}>Cancel</Button>
          <Button disabled={newTitle.trim().length === 0} isLoading={createMutation.isPending}
            onClick={() => createMutation.mutate(newTitle.trim())}>Create</Button>
        </div>
      </Modal>
    </div>
  )
}
