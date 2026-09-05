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

export const getMessages = (chatId: string, options?: { signal?: AbortSignal }) =>
  baseAPI
    .get<ApiResponse<Message[]>>(`/api/v1/chats/${chatId}/messages`, { signal: options?.signal })
    .then((res) => res.data.data!)
