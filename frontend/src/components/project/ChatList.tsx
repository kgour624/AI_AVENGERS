import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { getChats, createChat, updateChatTitle, archiveChat, unarchiveChat, permanentDeleteChat } from '@/api/chats'
import type { Chat } from '@/types/project'
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

function ChatRow({ chat, projectId, onChanged }: { chat: Chat; projectId: string; onChanged: () => void }) {
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

  const refresh = () => queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'chats'] })
  const activeChats = chats?.filter((c) => !c.isArchived) ?? []
  const archivedChats = chats?.filter((c) => c.isArchived) ?? []

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <p className="text-sm font-medium text-text-secondary">Chats</p>
        <Button variant="secondary" size="sm" onClick={() => setIsModalOpen(true)}>+ New Chat</Button>
      </div>

      {isLoading && (
        <div className="space-y-2">{Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-8" />)}</div>
      )}

      <div className="space-y-1">
        {activeChats.map((chat) => <ChatRow key={chat.id} chat={chat} projectId={projectId} onChanged={refresh} />)}
        {!isLoading && activeChats.length === 0 && <p className="text-sm text-text-disabled">No chats yet.</p>}
      </div>

      {archivedChats.length > 0 && (
        <div className="mt-4">
          <p className="mb-1 text-[10px] font-semibold uppercase tracking-wider text-text-disabled/50">Disabled</p>
          <div className="space-y-1">
            {archivedChats.map((chat) => <ChatRow key={chat.id} chat={chat} projectId={projectId} onChanged={refresh} />)}
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
