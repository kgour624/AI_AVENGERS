import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'
import type { Chat, Message } from '@/types/project'

export const getChats = (projectId: string, options?: { signal?: AbortSignal }) =>
  baseAPI
    .get<ApiResponse<Chat[]>>(`/api/v1/projects/${projectId}/chats`, { signal: options?.signal })
    .then((res) => res.data.data!)

export const getChat = (chatId: string, options?: { signal?: AbortSignal }) =>
  baseAPI
    .get<ApiResponse<Chat>>(`/api/v1/chats/${chatId}`, { signal: options?.signal })
    .then((res) => res.data.data!)

export const createChat = (projectId: string, title: string) =>
  baseAPI
    .post<ApiResponse<Chat>>(`/api/v1/projects/${projectId}/chats`, { title })
    .then((res) => res.data.data!)

// Feature #3 fix (docs bug list): PATCH/DELETE /chats/:id already
// existed and are fully wired (chat/service.go UpdateTitle, Archive)
// - only the frontend never called them.
export const updateChatTitle = (chatId: string, title: string) =>
  baseAPI
    .patch<ApiResponse<{ status: string }>>(`/api/v1/chats/${chatId}`, { title })
    .then((res) => res.data.data!)

export const archiveChat = (chatId: string) =>
  baseAPI
    .delete<ApiResponse<{ status: string }>>(`/api/v1/chats/${chatId}`)
    .then((res) => res.data.data!)

// unarchiveChat: restores a disabled chat to active
export const unarchiveChat = (chatId: string) =>
  baseAPI
    .post<ApiResponse<{ status: string }>>(`/api/v1/chats/${chatId}/unarchive`)
    .then((res) => res.data.data!)

// permanentDeleteChat: hard deletes chat + all messages. IRREVERSIBLE.
export const permanentDeleteChat = (chatId: string) =>
  baseAPI
    .delete<ApiResponse<{ status: string }>>(`/api/v1/chats/${chatId}/permanent`)
    .then((res) => res.data.data!)

export const getMessages = (chatId: string, options?: { signal?: AbortSignal }) =>
  baseAPI
    .get<ApiResponse<Message[]>>(`/api/v1/chats/${chatId}/messages`, { signal: options?.signal })
    .then((res) => res.data.data!)
