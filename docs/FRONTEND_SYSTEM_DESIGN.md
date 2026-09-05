# AI Avengers — Frontend System Design

> **Production-Grade React Frontend for Multi-Agent AI Platform**  
> Stack: React 18 + TypeScript (strict) + Vite + TanStack Query + Zustand + Tailwind CSS

---

## Document Status

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | DESIGN COMPLETE — READY FOR IMPLEMENTATION |
| Author | System Design Architect |
| Last Updated | 2026-09-05 |
| Backend API | See `AI_AVENGERS_SYSTEM_ARCHITECTURE.md` |
| Target Repo | gitlab.com/zepto-group3/ai_avengers |
| Branch | main |

---

## Table of Contents

1. [Product Vision — What It Looks Like](#1-product-vision)
2. [Technology Stack — Decisions with Reasoning](#2-technology-stack)
3. [Project Structure](#3-project-structure)
4. [Design System — Tokens, Colors, Typography](#4-design-system)
5. [Routing Architecture](#5-routing-architecture)
6. [State Management Architecture](#6-state-management)
7. [API Client Layer](#7-api-client-layer)
8. [SSE Streaming Architecture](#8-sse-streaming)
9. [Component Architecture](#9-component-architecture)
10. [Page Designs — Every Screen](#10-page-designs)
11. [Admin Panel Design](#11-admin-panel)
12. [Performance Strategy](#12-performance-strategy)
13. [TypeScript Domain Models](#13-typescript-domain-models)
14. [Error Handling Strategy](#14-error-handling)
15. [Security Considerations](#15-security)
16. [Locked Decisions](#16-locked-decisions)

---

## 1. Product Vision

### What It Looks Like

AI Avengers frontend is a **Cursor IDE + Linear + Kiro hybrid**:

- **Cursor-like:** Code-aware, expert responses with citations, diff-style suggestions
- **Linear-like:** Clean, dark, professional, no clutter, keyboard-first
- **Kiro-like:** Project-centric (not chat-centric), persistent context

### Main Interface Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  AI Avengers                    [Project: E-Commerce App]  [Admin]  │
├──────────────┬──────────────────────────────────────────────────────┤
│              │                                                       │
│  PROJECTS    │   ┌─────────────────────────────────────────────┐   │
│  ──────────  │   │  Active Experts                             │   │
│  > E-Commerce│   │  [🔵 SD Expert] [🟢 DB Expert] [🟡 Backend] │   │
│    Auth App  │   └─────────────────────────────────────────────┘   │
│    API Design│                                                       │
│              │   ┌─────────────────────────────────────────────┐   │
│  CHATS       │   │  Turn 23 — System Design Expert             │   │
│  ──────────  │   │  ┌─────────────────────────────────────┐   │   │
│  > DB Schema │   │  │ 🟢 ADVISE  94% confident            │   │   │
│    API Design│   │  │                                     │   │   │
│    Auth Flow │   │  │ Sharding ke liye consistent         │   │   │
│              │   │  │ hashing use karo [CHUNK_abc]        │   │   │
│  [+ New Chat]│   │  │ kyunki virtual nodes ensure         │   │   │
│              │   │  │ even distribution [CHUNK_def]       │   │   │
│              │   │  └─────────────────────────────────────┘   │   │
│              │   │                                             │   │
│              │   │  Turn 23 — DB Expert                        │   │
│              │   │  ┌─────────────────────────────────────┐   │   │
│              │   │  │ 🟡 WARN                             │   │   │
│              │   │  │ Sharding se pehle indexing try      │   │   │
│              │   │  │ karo [CHUNK_xyz]                    │   │   │
│              │   │  └─────────────────────────────────────┘   │   │
│              │   │                                             │   │
│              │   │  ⚡ SYNTHESIS                               │   │
│              │   │  ✅ Agreement: PostgreSQL use karo          │   │
│              │   │  ⚠️  Contradiction: SD says shard now,      │   │
│              │   │     DB says index first — YOU DECIDE        │   │
│              │   └─────────────────────────────────────────────┘   │
│              │                                                       │
│              │   ┌─────────────────────────────────────────────┐   │
│              │   │ Experts: [SD ✓] [DB ✓] [Backend] [DSA]     │   │
│              │   │ ┌──────────────────────────────────────┐   │   │
│              │   │ │ How should I design the user table?  │   │   │
│              │   │ │                          [📎] [Send] │   │   │
│              │   │ └──────────────────────────────────────┘   │   │
│              │   └─────────────────────────────────────────────┘   │
└──────────────┴──────────────────────────────────────────────────────┘
```

### Response Mode Colors (from transcript — OKLCH tokens)

| Mode | Color | Meaning |
|---|---|---|
| 🟢 ADVISE | `oklch(65% 0.15 145)` Green | Cited answer |
| 🔵 ASK | `oklch(65% 0.15 220)` Blue | Need more info |
| 🟡 WARN | `oklch(70% 0.18 60)` Yellow | Proceed with caution |
| 🟠 PUSH_BACK | `oklch(68% 0.18 40)` Orange | Challenge necessity |
| 🔴 REFUSE | `oklch(60% 0.2 25)` Red | Cannot answer |

---

## 2. Technology Stack

### Decisions with Reasoning

#### React 18 + TypeScript (strict)

**WHY React 18:**
- Concurrent features — `useTransition` for non-blocking SSE updates
- Suspense for data loading — works with React Router loaders
- Most popular, best ecosystem, transcript confirmed

**WHY TypeScript strict:**
- Transcript taught: `never` for exhaustive ResponseMode checks
- Compile-time safety for API responses — catch bugs before runtime
- Domain modeling — `ValidationResult` as types not exceptions

#### Vite (Build Tool)

**WHY Vite over CRA/Next.js:**
- Fast HMR — instant feedback during development
- `import.meta.env.VITE_API_URL` for environment variables (transcript pattern)
- No SSR needed — AI Avengers is a SPA (authenticated app)
- Smaller bundle, faster builds

#### TanStack Query (React Query v5)

**WHY TanStack Query:**
- Transcript taught: data fetching is not just fetch-and-render
- Built-in caching, background refetch, stale-while-revalidate
- Optimistic updates for ratings
- Deduplication — multiple components requesting same expert data = 1 request
- Works perfectly with React Router loaders

#### Zustand (Client State)

**WHY Zustand over Redux:**
- Transcript: don't over-engineer state management
- Lightweight, no boilerplate
- Perfect for: selected experts, UI state, streaming state
- Atomic stores — transcript's NanoStores pattern adapted

#### Tailwind CSS + CSS Custom Properties

**WHY Tailwind:**
- Fast development, consistent spacing
- Transcript taught OKLCH tokens — define in `@layer base` as CSS variables
- Dark theme by default — `dark:` variants
- Purged in production — small bundle

#### React Router v6 (Data Router)

**WHY React Router loaders:**
- Transcript pattern: `export const chatRoute = { element, loader }` — co-located
- Parallel data loading — loader fetches chat + messages + experts simultaneously
- Abort signals — cancel on navigation (transcript's AbortController pattern)
- `useLoaderData()` — type-safe data access

#### Additional Libraries

```
react-dropzone       — File upload (drag and drop)
react-markdown       — Render expert responses with markdown
prism-react-renderer — Code syntax highlighting in responses
date-fns             — Date formatting
zod                  — Runtime API response validation (transcript: unknown → typed)
axios                — HTTP client with interceptors
```

---

## 3. Project Structure

```
frontend/
├── src/
│   ├── main.tsx                    # Entry point
│   ├── App.tsx                     # Router setup
│   │
│   ├── design-system/
│   │   ├── tokens.css              # OKLCH color tokens
│   │   ├── typography.css          # Font scale
│   │   └── animations.css         # Transitions
│   │
│   ├── types/                      # TypeScript domain models
│   │   ├── api.ts                  # All API response types
│   │   ├── expert.ts               # Expert, ResponseMode, Citation
│   │   ├── project.ts              # Project, Chat, Message
│   │   ├── memory.ts               # L2Entry, L3Event
│   │   └── auth.ts                 # User, TokenPair
│   │
│   ├── api/                        # API client layer
│   │   ├── base.ts                 # Axios instance with interceptors
│   │   ├── auth.ts                 # login, register, refresh
│   │   ├── experts.ts              # getExperts, getExpertTopics
│   │   ├── projects.ts             # CRUD + expert management
│   │   ├── chats.ts                # CRUD + messages
│   │   ├── messages.ts             # sendMessage (SSE), rateMessage
│   │   └── admin.ts                # Admin endpoints
│   │
│   ├── hooks/                      # Custom hooks
│   │   ├── useSSEStream.ts         # SSE streaming hook
│   │   ├── useAuth.ts              # Auth state
│   │   ├── useExpertSelection.ts   # Expert picker state
│   │   └── useProjectMemory.ts     # L2/L3 memory queries
│   │
│   ├── stores/                     # Zustand stores
│   │   ├── authStore.ts            # User, tokens
│   │   ├── uiStore.ts              # Sidebar open, theme
│   │   └── streamStore.ts          # Active SSE streams
│   │
│   ├── components/                 # Reusable components
│   │   ├── ui/                     # Design system primitives
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   ├── Badge.tsx           # ResponseMode badges
│   │   │   ├── Card.tsx
│   │   │   ├── Modal.tsx
│   │   │   ├── Tooltip.tsx
│   │   │   ├── Skeleton.tsx        # Loading states
│   │   │   └── ScrollArea.tsx
│   │   │
│   │   ├── expert/
│   │   │   ├── ExpertCard.tsx      # Expert selection card
│   │   │   ├── ExpertBadge.tsx     # Active expert indicator
│   │   │   ├── ExpertPicker.tsx    # Multi-select expert panel
│   │   │   └── CapabilityCard.tsx  # Expert topics + depth
│   │   │
│   │   ├── chat/
│   │   │   ├── MessageBubble.tsx   # Single message
│   │   │   ├── ExpertResponse.tsx  # Expert response with mode badge
│   │   │   ├── CitationChip.tsx    # [CHUNK_abc] citation
│   │   │   ├── SynthesisPanel.tsx  # Multi-expert synthesis
│   │   │   ├── RatingWidget.tsx    # 1-5 stars + feedback
│   │   │   ├── StreamingIndicator.tsx # Thinking animation
│   │   │   └── MessageInput.tsx    # Input + file upload + expert select
│   │   │
│   │   ├── project/
│   │   │   ├── ProjectCard.tsx
│   │   │   ├── ProjectTimeline.tsx # L3 event log visualization
│   │   │   └── RepoStatus.tsx      # GitHub/GitLab sync status
│   │   │
│   │   └── layout/
│   │       ├── AppShell.tsx        # Main layout wrapper
│   │       ├── Sidebar.tsx         # Left sidebar
│   │       ├── Header.tsx          # Top bar
│   │       └── AuthGuard.tsx       # Protected route wrapper
│   │
│   ├── pages/                      # Route pages
│   │   ├── auth/
│   │   │   ├── LoginPage.tsx
│   │   │   └── RegisterPage.tsx
│   │   │
│   │   ├── projects/
│   │   │   ├── ProjectsPage.tsx    # Project list
│   │   │   └── ProjectPage.tsx     # Single project + chats
│   │   │
│   │   ├── chat/
│   │   │   └── ChatPage.tsx        # Main chat interface
│   │   │
│   │   ├── experts/
│   │   │   └── ExpertsPage.tsx     # Expert discovery
│   │   │
│   │   └── admin/
│   │       ├── AdminLayout.tsx
│   │       ├── AdminDashboard.tsx
│   │       ├── AdminExperts.tsx
│   │       ├── AdminClients.tsx
│   │       ├── AdminStats.tsx
│   │       └── AdminSettings.tsx
│   │
│   └── utils/
│       ├── cn.ts                   # Tailwind class merging
│       ├── format.ts               # Date, cost formatting
│       └── validation.ts           # Zod schemas
│
├── public/
├── index.html
├── vite.config.ts
├── tailwind.config.ts
├── tsconfig.json
├── .env.development
├── .env.production
└── package.json
```

---

## 4. Design System

### Color Tokens (OKLCH — from CSS-simplified transcript)

```css
/* src/design-system/tokens.css */
/* WHY OKLCH: Perceptually uniform lightness.
   Transcript taught: HSL lightness is not human-perception based.
   OKLCH gives consistent perceived brightness across hues. */

:root {
  /* Response Mode Colors */
  --color-advise:    oklch(65% 0.15 145);   /* Green */
  --color-ask:       oklch(65% 0.15 220);   /* Blue */
  --color-warn:      oklch(70% 0.18 60);    /* Yellow */
  --color-pushback:  oklch(68% 0.18 40);    /* Orange */
  --color-refuse:    oklch(60% 0.20 25);    /* Red */

  /* Hover states using relative colors (transcript pattern) */
  --color-advise-hover:   oklch(from var(--color-advise) calc(l - 8%) c h);
  --color-ask-hover:      oklch(from var(--color-ask) calc(l - 8%) c h);

  /* Surface colors — dark theme */
  --surface-base:    oklch(12% 0.01 270);   /* Darkest bg */
  --surface-raised:  oklch(16% 0.01 270);   /* Card bg */
  --surface-overlay: oklch(20% 0.01 270);   /* Modal bg */
  --surface-border:  oklch(25% 0.01 270);   /* Borders */

  /* Text */
  --text-primary:    oklch(95% 0.01 270);   /* Main text */
  --text-secondary:  oklch(65% 0.01 270);   /* Muted text */
  --text-disabled:   oklch(40% 0.01 270);   /* Disabled */

  /* Brand */
  --color-brand:     oklch(65% 0.20 270);   /* Purple */
  --color-brand-hover: oklch(from var(--color-brand) calc(l - 8%) c h);

  /* Semantic */
  --color-success:   var(--color-advise);
  --color-warning:   var(--color-warn);
  --color-error:     var(--color-refuse);
  --color-info:      var(--color-ask);

  /* Spacing scale */
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-6: 24px;
  --space-8: 32px;

  /* Border radius */
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;
  --radius-full: 9999px;

  /* Shadows */
  --shadow-sm: 0 1px 3px oklch(0% 0 0 / 0.3);
  --shadow-md: 0 4px 12px oklch(0% 0 0 / 0.4);
}
```

### Typography

```css
/* src/design-system/typography.css */
:root {
  --font-sans: 'Inter', system-ui, sans-serif;
  --font-mono: 'JetBrains Mono', 'Fira Code', monospace;

  --text-xs:   11px;
  --text-sm:   13px;
  --text-base: 14px;
  --text-lg:   16px;
  --text-xl:   20px;
  --text-2xl:  24px;

  --leading-tight:  1.25;
  --leading-normal: 1.5;
  --leading-relaxed: 1.75;
}
```

### Tailwind Config

```typescript
// tailwind.config.ts
export default {
  darkMode: 'class',
  content: ['./src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // Map CSS tokens to Tailwind classes
        'surface-base':    'oklch(12% 0.01 270)',
        'surface-raised':  'oklch(16% 0.01 270)',
        'surface-overlay': 'oklch(20% 0.01 270)',
        'surface-border':  'oklch(25% 0.01 270)',
        'text-primary':    'oklch(95% 0.01 270)',
        'text-secondary':  'oklch(65% 0.01 270)',
        'brand':           'oklch(65% 0.20 270)',
        'mode-advise':     'oklch(65% 0.15 145)',
        'mode-ask':        'oklch(65% 0.15 220)',
        'mode-warn':       'oklch(70% 0.18 60)',
        'mode-pushback':   'oklch(68% 0.18 40)',
        'mode-refuse':     'oklch(60% 0.20 25)',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      },
    },
  },
}
```

---

## 5. Routing Architecture

### Route Structure (React Router v6 Data Router)

```typescript
// src/App.tsx
// WHY Data Router: Transcript pattern — loaders co-located with components.
// Parallel data loading, abort signals, type-safe useLoaderData().

const router = createBrowserRouter([
  // Public routes
  { path: '/login',    element: <LoginPage /> },
  { path: '/register', element: <RegisterPage /> },

  // Protected routes
  {
    element: <AuthGuard />,  // Redirects to /login if not authenticated
    children: [
      {
        element: <AppShell />,  // Layout with sidebar
        children: [

          // Projects list
          {
            path: '/',
            ...projectsRoute,  // { element, loader } co-located
          },

          // Single project
          {
            path: '/projects/:projectId',
            ...projectRoute,
            children: [
              // Chat window
              {
                path: 'chats/:chatId',
                ...chatRoute,
              },
            ],
          },

          // Expert discovery
          {
            path: '/experts',
            ...expertsRoute,
          },
        ],
      },
    ],
  },

  // Admin routes — code split, only loads for admin users
  {
    path: '/admin',
    lazy: () => import('./pages/admin/AdminLayout'),
    children: [
      { index: true,          lazy: () => import('./pages/admin/AdminDashboard') },
      { path: 'experts',      lazy: () => import('./pages/admin/AdminExperts') },
      { path: 'clients',      lazy: () => import('./pages/admin/AdminClients') },
      { path: 'stats',        lazy: () => import('./pages/admin/AdminStats') },
      { path: 'settings',     lazy: () => import('./pages/admin/AdminSettings') },
    ],
  },
])
```

### Route Loader Pattern (from reactjs.md transcript)

```typescript
// src/pages/chat/ChatPage.tsx
// WHY co-located loader: Transcript taught this pattern explicitly.
// Router only defines path. Component owns its data requirements.

import type { LoaderFunctionArgs } from 'react-router-dom'
import { getChat, getMessages } from '@/api/chats'
import { getProjectExperts } from '@/api/projects'

// Loader — runs before component renders
async function loader({ params, request }: LoaderFunctionArgs) {
  const signal = request.signal  // Abort on navigation (transcript pattern)
  const { projectId, chatId } = params as { projectId: string; chatId: string }

  // Parallel fetch — all three simultaneously
  const [chat, messages, experts] = await Promise.all([
    getChat(chatId, { signal }),
    getMessages(chatId, { signal }),
    getProjectExperts(projectId, { signal }),
  ])

  return { chat, messages, experts }
}

// Export route object — transcript pattern
export const chatRoute = { element: <ChatPage />, loader }

// Component
export default function ChatPage() {
  const { chat, messages, experts } = useLoaderData() as ChatLoaderData
  // ...
}
```

---

## 6. State Management

### Two-Layer State Architecture

```
Server State (TanStack Query)     Client State (Zustand)
─────────────────────────────     ──────────────────────
Experts list                      Selected experts for turn
Project list                      Sidebar open/closed
Chat messages                     Active streaming state
Project memory (L2)               UI preferences
Admin stats                       Auth tokens
```

### TanStack Query Setup

```typescript
// src/api/queryKeys.ts
// WHY structured query keys: Cache invalidation is predictable.
// Transcript: caching strategy is critical for performance.

export const queryKeys = {
  experts: {
    all: ['experts'] as const,
    detail: (id: string) => ['experts', id] as const,
    topics: (id: string) => ['experts', id, 'topics'] as const,
  },
  projects: {
    all: ['projects'] as const,
    detail: (id: string) => ['projects', id] as const,
    chats: (id: string) => ['projects', id, 'chats'] as const,
    memory: (id: string) => ['projects', id, 'memory'] as const,
    timeline: (id: string) => ['projects', id, 'timeline'] as const,
  },
  chats: {
    messages: (id: string) => ['chats', id, 'messages'] as const,
  },
}
```

### Zustand Stores

```typescript
// src/stores/authStore.ts
interface AuthStore {
  user: User | null
  accessToken: string | null
  setAuth: (user: User, token: string) => void
  clearAuth: () => void
  isAdmin: () => boolean
}

// src/stores/streamStore.ts
// WHY separate stream store: SSE state is transient, not server state.
// Transcript: streaming responses feel instant — manage carefully.
interface StreamStore {
  activeStreams: Map<string, StreamState>  // chatId → state
  startStream: (chatId: string) => void
  appendChunk: (chatId: string, chunk: SSEChunk) => void
  completeStream: (chatId: string) => void
  clearStream: (chatId: string) => void
}

interface StreamState {
  status: 'thinking' | 'streaming' | 'complete' | 'error'
  expertResponses: Partial<ExpertResponse>[]
  synthesis: SynthesisResult | null
  error: string | null
}
```

---

## 7. API Client Layer

### Base Axios Instance (from reactjs.md transcript)

```typescript
// src/api/base.ts
// WHY base instance: Transcript taught this pattern explicitly.
// All API calls share baseURL, auth headers, error handling.

import axios from 'axios'
import { useAuthStore } from '@/stores/authStore'

export const baseAPI = axios.create({
  baseURL: import.meta.env.VITE_API_URL,  // Transcript: env vars for API URL
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

// Request interceptor — attach JWT
baseAPI.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// Response interceptor — handle 401, refresh token
baseAPI.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Attempt token refresh
      try {
        const newToken = await refreshToken()
        useAuthStore.getState().setAuth(
          useAuthStore.getState().user!,
          newToken
        )
        // Retry original request
        error.config.headers.Authorization = `Bearer ${newToken}`
        return baseAPI(error.config)
      } catch {
        useAuthStore.getState().clearAuth()
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)
```

### API Functions (co-located pattern)

```typescript
// src/api/experts.ts
import { baseAPI } from './base'
import type { Expert, ExpertTopic } from '@/types/expert'

export const getExperts = (options?: { signal?: AbortSignal }) =>
  baseAPI.get<Expert[]>('/api/v1/experts', { signal: options?.signal })
    .then(res => res.data)

export const getExpertTopics = (id: string, options?: { signal?: AbortSignal }) =>
  baseAPI.get<{ topics: ExpertTopic[] }>(`/api/v1/experts/${id}/topics`, {
    signal: options?.signal
  }).then(res => res.data)

// src/api/messages.ts
// NOTE: sendMessage uses SSE — NOT axios. See useSSEStream hook.
export const rateMessage = (messageId: string, rating: RateRequest) =>
  baseAPI.post(`/api/v1/messages/${messageId}/rate`, rating)
    .then(res => res.data)
```

---

## 8. SSE Streaming Architecture

### useSSEStream Hook

```typescript
// src/hooks/useSSEStream.ts
// WHY custom hook: SSE is fundamentally different from REST.
// fetch() with ReadableStream — not axios, not EventSource.
// Transcript: streaming responses feel instant.

import { useStreamStore } from '@/stores/streamStore'

interface SendMessageOptions {
  chatId: string
  message: string
  expertIds: string[]
  file?: File
}

export function useSSEStream() {
  const { startStream, appendChunk, completeStream, clearStream } = useStreamStore()

  const sendMessage = async (options: SendMessageOptions) => {
    const { chatId, message, expertIds, file } = options

    startStream(chatId)

    // Build request body
    let body: BodyInit
    let headers: HeadersInit = {}

    if (file) {
      // Multipart for file upload
      const formData = new FormData()
      formData.append('message', message)
      formData.append('expert_ids', JSON.stringify(expertIds))
      formData.append('file', file)
      body = formData
    } else {
      body = JSON.stringify({ message, expert_ids: expertIds })
      headers['Content-Type'] = 'application/json'
    }

    const token = useAuthStore.getState().accessToken
    headers['Authorization'] = `Bearer ${token}`

    try {
      const response = await fetch(
        `${import.meta.env.VITE_API_URL}/api/v1/chats/${chatId}/messages`,
        { method: 'POST', headers, body }
      )

      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      if (!response.body) throw new Error('No response body')

      // Read SSE stream
      const reader = response.body.getReader()
      const decoder = new TextDecoder()

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const text = decoder.decode(value, { stream: true })
        const lines = text.split('\n')

        for (const line of lines) {
          if (!line.startsWith('data: ')) continue
          const jsonStr = line.slice(6).trim()
          if (!jsonStr) continue

          try {
            const event = JSON.parse(jsonStr) as SSEEvent
            appendChunk(chatId, event)

            if (event.type === 'done') {
              completeStream(chatId)
              return
            }
          } catch {
            // Malformed JSON — skip
          }
        }
      }
    } catch (error) {
      useStreamStore.getState().setError(chatId, String(error))
    }
  }

  return { sendMessage }
}

// SSE Event types — match backend exactly
type SSEEvent =
  | { type: 'thinking'; data: { message: string; experts: number } }
  | { type: 'complete'; data: ExpertResponse }
  | { type: 'synthesis'; data: SynthesisResult }
  | { type: 'done'; data: { turn_number: number; duration_ms: number } }
  | { type: 'error'; data: { message: string } }
```

---

## 9. Component Architecture

### ExpertResponse Component

```typescript
// src/components/chat/ExpertResponse.tsx
// WHY this component is critical:
// This is the core value proposition of AI Avengers.
// Every design decision here affects trust and usability.

interface ExpertResponseProps {
  response: ExpertResponse
  isStreaming?: boolean
}

// Response mode → visual treatment
const MODE_CONFIG: Record<ResponseMode, {
  color: string
  icon: string
  label: string
  bgClass: string
}> = {
  ADVISE:     { color: 'text-mode-advise',    icon: '✅', label: 'ADVISE',     bgClass: 'bg-mode-advise/10 border-mode-advise/30' },
  ASK:        { color: 'text-mode-ask',       icon: '❓', label: 'ASK',        bgClass: 'bg-mode-ask/10 border-mode-ask/30' },
  WARN:       { color: 'text-mode-warn',      icon: '⚠️', label: 'WARN',       bgClass: 'bg-mode-warn/10 border-mode-warn/30' },
  PUSH_BACK:  { color: 'text-mode-pushback',  icon: '🔄', label: 'PUSH BACK',  bgClass: 'bg-mode-pushback/10 border-mode-pushback/30' },
  REFUSE:     { color: 'text-mode-refuse',    icon: '🚫', label: 'REFUSE',     bgClass: 'bg-mode-refuse/10 border-mode-refuse/30' },
}
```

### MessageInput Component

```typescript
// src/components/chat/MessageInput.tsx
// Features:
// 1. Expert multi-select (checkboxes)
// 2. File upload (drag and drop via react-dropzone)
// 3. Textarea with auto-resize
// 4. Send button (disabled while streaming)
// 5. Keyboard shortcut: Cmd+Enter to send
```

### CitationChip Component

```typescript
// src/components/chat/CitationChip.tsx
// Renders [CHUNK_abc] as clickable chip
// On hover: shows chunk text in tooltip
// On click: opens citation detail modal
// WHY: Citations are the trust mechanism of AI Avengers
```

---

## 10. Page Designs — Every Screen

### Login Page

```
┌─────────────────────────────────────────────────────┐
│                                                     │
│              ⚡ AI Avengers                         │
│         Domain Expert Team Simulator                │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │  Email                                      │   │
│  │  ┌───────────────────────────────────────┐  │   │
│  │  │ you@company.com                       │  │   │
│  │  └───────────────────────────────────────┘  │   │
│  │                                             │   │
│  │  Password                                   │   │
│  │  ┌───────────────────────────────────────┐  │   │
│  │  │ ••••••••                              │  │   │
│  │  └───────────────────────────────────────┘  │   │
│  │                                             │   │
│  │  [        Sign In        ]                  │   │
│  │                                             │   │
│  │  Don't have an account? Register            │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### Projects Page (Home)

```
┌─────────────────────────────────────────────────────────────────────┐
│  ⚡ AI Avengers                                    [+ New Project]  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Your Projects                                                      │
│                                                                     │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐  │
│  │ E-Commerce App   │  │ Auth Service     │  │ API Design       │  │
│  │                  │  │                  │  │                  │  │
│  │ Experts: 3       │  │ Experts: 2       │  │ Experts: 1       │  │
│  │ Chats: 12        │  │ Chats: 5         │  │ Chats: 3         │  │
│  │ Last: 2h ago     │  │ Last: 1d ago     │  │ Last: 3d ago     │  │
│  │                  │  │                  │  │                  │  │
│  │ [Open Project]   │  │ [Open Project]   │  │ [Open Project]   │  │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Project Page (with Chat List)

```
┌─────────────────────────────────────────────────────────────────────┐
│  ⚡ AI Avengers  /  E-Commerce App                  [+ New Chat]   │
├──────────────┬──────────────────────────────────────────────────────┤
│              │                                                       │
│  CHATS       │   Project: E-Commerce App                            │
│  ──────────  │                                                       │
│  > DB Schema │   Experts in this project:                           │
│    API Design│   ┌──────────────────────────────────────────────┐  │
│    Auth Flow │   │ [🔵 SD Expert] [🟢 DB Expert] [🟡 Backend]   │  │
│    Sharding  │   │                              [+ Add Expert]   │  │
│              │   └──────────────────────────────────────────────┘  │
│  [+ New Chat]│                                                       │
│              │   Recent Activity (Timeline)                         │
│              │   ┌──────────────────────────────────────────────┐  │
│              │   │ Today                                        │  │
│              │   │ 14:23 [DB] Decided: PostgreSQL sharding      │  │
│              │   │ 13:45 [SD] Recommended: Monolith first       │  │
│              │   │ Yesterday                                    │  │
│              │   │ 16:12 [DB] Decided: pgvector for embeddings  │  │
│              │   └──────────────────────────────────────────────┘  │
│              │                                                       │
│              │   GitHub/GitLab                                       │
│              │   ┌──────────────────────────────────────────────┐  │
│              │   │ Not connected  [Connect Repository]          │  │
│              │   └──────────────────────────────────────────────┘  │
└──────────────┴──────────────────────────────────────────────────────┘
```

### Chat Page (Main Interface)

```
┌─────────────────────────────────────────────────────────────────────┐
│  ⚡ AI Avengers  /  E-Commerce  /  DB Schema Chat                  │
├──────────────┬──────────────────────────────────────────────────────┤
│              │                                                       │
│  CHATS       │  ┌─────────────────────────────────────────────┐    │
│  ──────────  │  │ Active: [🔵 SD ✓] [🟢 DB ✓] [Backend]      │    │
│  > DB Schema │  └─────────────────────────────────────────────┘    │
│    API Design│                                                       │
│    Auth Flow │  ┌─────────────────────────────────────────────┐    │
│              │  │ USER — Turn 23                              │    │
│              │  │ How should I design the user table?         │    │
│              │  └─────────────────────────────────────────────┘    │
│              │                                                       │
│              │  ┌─────────────────────────────────────────────┐    │
│              │  │ 🟢 ADVISE  DB Expert  94%                   │    │
│              │  │                                             │    │
│              │  │ Use UUID primary key [CHUNK_abc] because    │    │
│              │  │ distributed systems need globally unique    │    │
│              │  │ identifiers [CHUNK_def].                    │    │
│              │  │                                             │    │
│              │  │ ```sql                                      │    │
│              │  │ CREATE TABLE users (                        │    │
│              │  │   id UUID PRIMARY KEY DEFAULT gen_random_uuid(), │ │
│              │  │   email VARCHAR(255) UNIQUE NOT NULL,       │    │
│              │  │   ...                                       │    │
│              │  │ );                                          │    │
│              │  │ ```                                         │    │
│              │  │                                             │    │
│              │  │ Citations: [CHUNK_abc] [CHUNK_def]          │    │
│              │  │ ⭐⭐⭐⭐⭐  [Rate this response]             │    │
│              │  │ [▶ Why did I say this?]                     │    │
│              │  └─────────────────────────────────────────────┘    │
│              │                                                       │
│              │  ┌─────────────────────────────────────────────┐    │
│              │  │ 🟡 WARN  SD Expert                          │    │
│              │  │ Consider CQRS for this table BECAUSE        │    │
│              │  │ read/write patterns differ [CHUNK_xyz]      │    │
│              │  └─────────────────────────────────────────────┘    │
│              │                                                       │
│              │  ┌─────────────────────────────────────────────┐    │
│              │  │ ⚡ SYNTHESIS                                 │    │
│              │  │ ✅ Both experts agree: UUID primary key      │    │
│              │  │ ⚠️  Contradiction: DB says simple schema,    │    │
│              │  │    SD suggests CQRS — YOU DECIDE             │    │
│              │  └─────────────────────────────────────────────┘    │
│              │                                                       │
│              │  ┌─────────────────────────────────────────────┐    │
│              │  │ Experts: [SD ✓] [DB ✓] [Backend] [DSA]     │    │
│              │  │ ┌───────────────────────────────────────┐   │    │
│              │  │ │ Ask a follow-up question...           │   │    │
│              │  │ │                           [📎] [Send] │   │    │
│              │  │ └───────────────────────────────────────┘   │    │
│              │  └─────────────────────────────────────────────┘    │
└──────────────┴──────────────────────────────────────────────────────┘
```

### Streaming State (while experts are thinking)

```
│  ┌─────────────────────────────────────────────────┐    │
│  │ ⏳ Processing...                                │    │
│  │                                                 │    │
│  │ 🔵 SD Expert    ████████░░  Gate 3/5            │    │
│  │ 🟢 DB Expert    ██████████  Generating...       │    │
│  │                                                 │    │
│  └─────────────────────────────────────────────────┘    │
```

### Expert Discovery Page

```
┌─────────────────────────────────────────────────────────────────────┐
│  ⚡ AI Avengers  /  Experts                                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Available Domain Experts                                           │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 🔵 Arpit Bhiyani — System Design                            │  │
│  │ Domain: system_design  |  Depth: Expert (4/5)               │  │
│  │ Topics: sharding, caching, kafka, database, microservices    │  │
│  │ Rating: ⭐⭐⭐⭐⭐ (4.8)  |  1,247 chunks                    │  │
│  │ [View Topics]  [Add to Project]                              │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 🟢 DB Expert — Database Architecture                        │  │
│  │ Domain: database  |  Depth: Advanced (3/5)                  │  │
│  │ Topics: postgresql, indexing, sharding, replication          │  │
│  │ Rating: ⭐⭐⭐⭐½ (4.5)  |  892 chunks                      │  │
│  │ [View Topics]  [Add to Project]                              │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 11. Admin Panel Design

### Admin Dashboard

```
┌─────────────────────────────────────────────────────────────────────┐
│  ⚡ AI Avengers Admin                              [← Back to App]  │
├──────────────┬──────────────────────────────────────────────────────┤
│              │                                                       │
│  ADMIN       │  System Overview                                      │
│  ──────────  │                                                       │
│  Dashboard   │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ │
│  Experts     │  │ Experts  │ │ Clients  │ │ Projects │ │  Cost  │ │
│  Clients     │  │    12    │ │   234    │ │   891    │ │ $23.45 │ │
│  Stats       │  │  active  │ │  total   │ │  total   │ │ /month │ │
│  Violations  │  └──────────┘ └──────────┘ └──────────┘ └────────┘ │
│  Settings    │                                                       │
│              │  China Wall Violations (last 7 days)                  │
│              │  ████████████░░░░░░░░  12 violations (2.1%)          │
│              │                                                       │
│              │  Expert Ratings                                       │
│              │  Arpit SD    ⭐⭐⭐⭐⭐ 4.8  (1,247 ratings)          │
│              │  DB Expert   ⭐⭐⭐⭐½  4.5  (892 ratings)            │
│              │                                                       │
└──────────────┴──────────────────────────────────────────────────────┘
```

### Admin Expert Management

```
│  Experts                                    [+ Create Expert]       │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ Arpit Bhiyani — System Design                               │  │
│  │ Status: ✅ Active  |  Chunks: 1,247  |  Rating: 4.8         │  │
│  │ [Upload Transcript]  [Edit Charter]  [Disable]              │  │
│  │                                                              │  │
│  │ Last ingestion: 2026-09-04  Status: ✅ Complete (1,247 chunks)│  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  Upload Transcript Modal:                                           │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ ┌────────────────────────────────────────────────────────┐  │  │
│  │ │                                                        │  │  │
│  │ │   📄 Drag & drop transcript here                       │  │  │
│  │ │      or click to browse (.txt, .md)                    │  │  │
│  │ │                                                        │  │  │
│  │ └────────────────────────────────────────────────────────┘  │  │
│  │                                                              │  │
│  │ [Cancel]                          [Start Ingestion]         │  │
│  └──────────────────────────────────────────────────────────────┘  │
```

---

## 12. Performance Strategy

### Application Level (from Frontend System Design transcript)

```typescript
// 1. Code splitting — admin routes lazy loaded
{ path: '/admin', lazy: () => import('./pages/admin/AdminLayout') }

// 2. React Router loaders — parallel data fetching
// All data loads before component renders — no waterfall
const [chat, messages, experts] = await Promise.all([...])

// 3. TanStack Query — stale-while-revalidate
const { data: experts } = useQuery({
  queryKey: queryKeys.experts.all,
  queryFn: getExperts,
  staleTime: 5 * 60 * 1000,  // 5 minutes fresh
  gcTime: 30 * 60 * 1000,    // 30 minutes in cache
})

// 4. Virtualization — long message lists
// Use @tanstack/react-virtual for 100+ messages

// 5. Skeleton screens — perceived performance
// Show skeleton while loader runs
```

### Infrastructure Level (from transcript)

```
Backend sets these headers (already implemented in Go):
  Cache-Control: public, max-age=300, stale-while-revalidate=60  → Expert list
  Cache-Control: private, max-age=60                             → User data
  Cache-Control: no-store                                        → Auth tokens
  ETag: md5(response_body)                                       → All GET responses
  Vary: Authorization                                            → User-specific data

Frontend benefits automatically:
  Expert list → served from browser cache (0ms)
  304 Not Modified → no data transfer when unchanged
  SWR → instant response + background refresh
```

---

## 13. TypeScript Domain Models

```typescript
// src/types/expert.ts
// WHY strict types: Transcript taught — never for exhaustive checks,
// unknown for API responses, ValidationResult as types not exceptions.

export type ResponseMode = 'ASK' | 'WARN' | 'PUSH_BACK' | 'REFUSE' | 'ADVISE'

export interface Expert {
  id: string
  name: string
  slug: string
  domain: string
  description: string
  totalChunks: number
  totalTopics: number
  avgDepthLevel: number
  avgRating: number
  createdAt: string
}

export interface Citation {
  chunkId: string
  text: string
  score: number
}

export interface ExpertResponse {
  expertId: string
  expertName: string
  domain: string
  mode: ResponseMode
  content: string
  citations: Citation[]
  confidence: number
  gateStopped: 0 | 1 | 2 | 3 | 4 | 5
  warning?: string
  questions?: string[]  // For ASK mode
  error?: string
}

export interface SynthesisResult {
  agreements: string[]
  contradictions: Contradiction[]
  summary: string
}

export interface Contradiction {
  topic: string
  expertA: string
  positionA: string
  expertB: string
  positionB: string
}

// Exhaustive check — transcript's never pattern
// If backend adds new ResponseMode, TypeScript immediately flags it
export function getModeBadgeConfig(mode: ResponseMode) {
  switch (mode) {
    case 'ADVISE':    return { color: 'mode-advise',    icon: '✅', label: 'ADVISE' }
    case 'ASK':       return { color: 'mode-ask',       icon: '❓', label: 'ASK' }
    case 'WARN':      return { color: 'mode-warn',      icon: '⚠️', label: 'WARN' }
    case 'PUSH_BACK': return { color: 'mode-pushback',  icon: '🔄', label: 'PUSH BACK' }
    case 'REFUSE':    return { color: 'mode-refuse',    icon: '🚫', label: 'REFUSE' }
    default: {
      const _exhaustive: never = mode  // Compile error if new mode added
      return _exhaustive
    }
  }
}

// src/types/project.ts
export interface Project {
  id: string
  clientId: string
  name: string
  description: string
  status: 'active' | 'completed' | 'archived'
  repoUrl?: string
  repoProvider?: 'github' | 'gitlab'
  repoConnected: boolean
  experts: ProjectExpert[]
  createdAt: string
  updatedAt: string
}

export interface Chat {
  id: string
  projectId: string
  title: string
  messageCount: number
  isArchived: boolean
  createdAt: string
  updatedAt: string
}

export interface Message {
  id: string
  chatId: string
  role: 'user' | 'assistant'
  content: string
  turnNumber: number
  expertId?: string
  decisionMode?: ResponseMode
  citations?: Citation[]
  confidence?: number
  createdAt: string
}

// API response wrapper — matches Go backend exactly
export interface APIResponse<T> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
    details?: Record<string, string>
  }
  meta: {
    requestId: string
    timestamp: string
  }
}
```

---

## 14. Error Handling Strategy

```typescript
// src/utils/errors.ts
// WHY: Transcript taught — never silent catch, always log or rethrow.
// Frontend equivalent: never swallow errors silently.

// Global error boundary
export class AppErrorBoundary extends React.Component {
  // Catches render errors, shows fallback UI
}

// API error handling
export function handleAPIError(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const apiError = error.response?.data as APIResponse<never>
    return apiError?.error?.message ?? 'An error occurred'
  }
  if (error instanceof Error) return error.message
  return 'Unknown error'
}

// Route error element
export function RouteError() {
  const error = useRouteError()
  // Show user-friendly error with retry option
}
```

---

## 15. Security Considerations

```
1. JWT stored in memory (Zustand) — NOT localStorage
   WHY: XSS cannot steal in-memory tokens
   Refresh token in httpOnly cookie (set by backend)

2. API calls always through base.ts interceptor
   WHY: Single place to add auth headers, handle 401

3. Admin routes — double check: route guard + API returns 403
   WHY: Defense in depth

4. File uploads — validate type + size client-side
   WHY: Fast feedback, reduce server load
   Backend validates again (never trust client)

5. No sensitive data in URL params
   WHY: URLs appear in logs, browser history
```

---

## 16. Locked Decisions

| # | Decision | Reason |
|---|---|---|
| 1 | React 18 + TypeScript strict | Concurrent features + compile-time safety |
| 2 | Vite (not Next.js) | SPA — no SSR needed, faster builds |
| 3 | TanStack Query for server state | Caching, SWR, deduplication |
| 4 | Zustand for client state | Lightweight, no boilerplate |
| 5 | Tailwind + OKLCH tokens | Perceptually uniform colors (transcript) |
| 6 | React Router v6 Data Router | Co-located loaders, parallel fetch |
| 7 | SSE via fetch() not EventSource | POST support, auth headers |
| 8 | JWT in memory not localStorage | XSS protection |
| 9 | `never` for ResponseMode exhaustive check | Transcript pattern |
| 10 | Monolith (not micro-frontends) for v1 | Transcript: fat core first |
| 11 | pnpm | Faster installs, smaller disk |
| 12 | Admin routes lazy loaded | Code split — admin code only for admins |
