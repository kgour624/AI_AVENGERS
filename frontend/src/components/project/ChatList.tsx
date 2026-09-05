import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { getChats, createChat } from '@/api/chats'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { Modal } from '@/components/ui/Modal'

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
          <Link
            key={chat.id}
            to={`/projects/${projectId}/chats/${chat.id}`}
            className="block rounded-md px-2 py-1.5 text-sm text-text-secondary hover:bg-surface-overlay hover:text-text-primary"
          >
            {chat.title}
          </Link>
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
