import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { getChats, createChat, updateChatTitle, archiveChat, unarchiveChat, permanentDeleteChat } from '@/api/chats'
import type { Chat } from '@/types/project'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { Modal } from '@/components/ui/Modal'

/**
 * Feature #3 fix (docs bug list): inline rename/archive per chat row.
 * Rename uses an inline text input (not a separate modal) since it's
 * a single-field edit directly on the row that's already visible -
 * a modal would be unnecessary ceremony for one text field.
 */
function ChatRow({
  chat,
  projectId,
  onChanged,
}: {
  chat: Chat
  projectId: string
  onChanged: () => void
}) {
  const [isEditing, setIsEditing] = useState(false)
  const [title, setTitle] = useState(chat.title)

  const renameMutation = useMutation({
    mutationFn: () => updateChatTitle(chat.id, title.trim()),
    onSuccess: () => {
      setIsEditing(false)
      onChanged()
    },
  })

  const archiveMutation = useMutation({
    mutationFn: () => archiveChat(chat.id),
    onSuccess: onChanged,
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
        <button
          className="text-xs text-brand"
          disabled={!title.trim() || renameMutation.isPending}
          onClick={() => renameMutation.mutate()}
        >
          Save
        </button>
        <button className="text-xs text-text-secondary" onClick={() => setIsEditing(false)}>
          Cancel
        </button>
      </div>
    )
  }

  return (
    <div className="group flex items-center gap-1 rounded-md px-2 py-1.5 hover:bg-surface-overlay">
      <Link
        to={`/projects/${projectId}/chats/${chat.id}`}
        className="flex-1 truncate text-sm text-text-secondary hover:text-text-primary"
      >
        {chat.title}
      </Link>
      <button
        aria-label="Rename chat"
        title="Rename"
        className="invisible text-xs text-text-disabled hover:text-text-primary group-hover:visible"
        onClick={() => setIsEditing(true)}
      >
        {'\u270f\ufe0f'}
      </button>
      <button
        aria-label="Archive chat"
        title="Archive"
        disabled={archiveMutation.isPending}
        className="invisible text-xs text-text-disabled hover:text-mode-refuse group-hover:visible"
        onClick={() => {
          if (window.confirm(`Archive "${chat.title}"?`)) archiveMutation.mutate()
        }}
      >
        {'\ud83d\uddc4\ufe0f'}
      </button>
    </div>
  )
}

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 10 ("Project Page (with
 * Chat List)" wireframe - CHATS sidebar list + [+ New Chat] button).
 *
 * WHY this is a separate component rather than inlined in ProjectPage:
 * it owns its own query (chat list can change independently of the
 * parent project data the route loader already fetched) and its own
 * create-chat mutation/modal state - bundling that into ProjectPage
 * directly would make an already content-heavy component responsible
 * for two independent concerns.
 */
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

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <p className="text-sm font-medium text-text-secondary">Chats</p>
        <Button variant="secondary" size="sm" onClick={() => setIsModalOpen(true)}>
          + New Chat
        </Button>
      </div>

      {isLoading && (
        <div className="space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-8" />
          ))}
        </div>
      )}

      <div className="space-y-1">
        {chats?.map((chat) => (
          <ChatRow
            key={chat.id}
            chat={chat}
            projectId={projectId}
            onChanged={() => queryClient.invalidateQueries({ queryKey: ['projects', projectId, 'chats'] })}
          />
        ))}
        {chats?.length === 0 && <p className="text-sm text-text-disabled">No chats yet.</p>}
      </div>

      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)}>
        <h3 className="mb-4 text-sm font-medium text-text-primary">New Chat</h3>
        <Input
          label="Title"
          value={newTitle}
          onChange={(e) => setNewTitle(e.target.value)}
          placeholder="e.g. Database Schema Discussion"
          autoFocus
        />
        <div className="mt-4 flex justify-end gap-2">
          <Button variant="secondary" onClick={() => setIsModalOpen(false)}>
            Cancel
          </Button>
          <Button
            disabled={newTitle.trim().length === 0}
            isLoading={createMutation.isPending}
            onClick={() => createMutation.mutate(newTitle.trim())}
          >
            Create
          </Button>
        </div>
      </Modal>
    </div>
  )
}
