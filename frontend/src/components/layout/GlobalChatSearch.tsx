import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { searchChats, type ChatSearchResult } from '@/api/chats'

const MIN_QUERY = 2
const DEBOUNCE_MS = 250

/**
 * GlobalChatSearch (#27) — header-level search across ALL of the user's chats
 * (every project), matching chat titles and message content.
 *
 * WHY here (not in ChatPage): the per-chat search in ChatPage only looks at the
 * chat you are already in. With dozens of chats named "New Chat" there was no
 * way to find one from anywhere else. This is the entry point that returns
 * chat name + project name and jumps straight to the hit.
 *
 * Behavior: debounced query (250ms), min 2 chars, Escape / click-outside
 * closes, Enter opens the top hit. Backend: GET /api/v1/search/chats.
 */
export function GlobalChatSearch() {
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const [term, setTerm] = useState('')
  const [debounced, setDebounced] = useState('')
  const rootRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    const t = setTimeout(() => setDebounced(term.trim()), DEBOUNCE_MS)
    return () => clearTimeout(t)
  }, [term])

  useEffect(() => {
    if (!open) return
    inputRef.current?.focus()
    const onDown = (e: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDown)
    return () => document.removeEventListener('mousedown', onDown)
  }, [open])

  const active = open && debounced.length >= MIN_QUERY
  const { data, isFetching } = useQuery({
    queryKey: ['global-chat-search', debounced],
    queryFn: ({ signal }) => searchChats(debounced, { signal }),
    enabled: active,
    staleTime: 30_000,
    retry: false,
  })
  const results = data ?? []

  const close = () => {
    setOpen(false)
    setTerm('')
    setDebounced('')
  }

  const go = (hit: ChatSearchResult) => {
    close()
    navigate(`/projects/${hit.projectId}/chats/${hit.chatId}`)
  }

  return (
    <div ref={rootRef} className="relative">
      <button
        type="button"
        aria-label="Search chats"
        title="Search all chats"
        onClick={() => setOpen((v) => !v)}
        className="text-sm text-text-disabled transition-colors duration-150 ease-arc hover:text-glow-cyan"
      >
        {'\uD83D\uDD0D'}
      </button>

      {open && (
        <div className="absolute right-0 top-8 z-50 w-80 rounded-lg border border-glass-border bg-surface-raised shadow-card">
          <div className="border-b border-surface-border p-2">
            <input
              ref={inputRef}
              value={term}
              onChange={(e) => setTerm(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Escape') {
                  e.stopPropagation()
                  close()
                }
                if (e.key === 'Enter') {
                  const first = results[0]
                  if (first) go(first)
                }
              }}
              placeholder="Search all chats..."
              aria-label="Search all chats"
              className="w-full rounded border border-surface-overlay bg-surface-base px-2 py-1 text-xs text-text-primary placeholder:text-text-disabled focus:outline-none focus:ring-1 focus:ring-brand"
            />
          </div>

          <div className="max-h-80 overflow-y-auto p-1">
            {!active && (
              <p className="p-3 text-center text-[11px] text-text-disabled">
                Type at least {MIN_QUERY} characters
              </p>
            )}
            {active && results.length === 0 && (
              <p className="p-3 text-center text-[11px] text-text-disabled">
                {isFetching ? 'Searching...' : `No chats match "${debounced}".`}
              </p>
            )}

            {results.map((hit) => (
              <button
                key={hit.chatId}
                type="button"
                onClick={() => go(hit)}
                className="block w-full rounded px-2 py-1.5 text-left hover:bg-surface-overlay/60"
              >
                <span className="flex items-center gap-1.5">
                  <span className="truncate text-xs text-text-primary">
                    {hit.chatTitle || 'Untitled chat'}
                  </span>
                  {hit.isArchived && (
                    <span className="shrink-0 text-[9px] uppercase tracking-wider text-text-disabled">
                      [off]
                    </span>
                  )}
                </span>
                <span className="mt-0.5 flex items-center gap-1.5 text-[10px] text-text-disabled">
                  <span className="truncate">{hit.projectName}</span>
                  {hit.contentMatches > 0 && (
                    <span className="shrink-0">
                      {'\u00b7'} {hit.contentMatches} message{hit.contentMatches === 1 ? '' : 's'}
                    </span>
                  )}
                </span>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
