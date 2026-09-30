# Frontend Essentials — Architecture Knowledge Base

Source: Transcripts/Frontend/ (Frontend System Design Essentials.md, Frontend Architecture Monoliths to Micro-Frontends.mf.md, TyeScript Simplified.md, Model Complex Domains with TypeScript.md, CSS-simplified.md) — Sneha Mehra. Permanent reference for frontend system design.

## 1. 7 Essential Areas (Shift: Component-first → Architecture mindset)

Frontend system design = data flow + failure cases + trade-offs (perf vs flexibility vs delivery), not just UI.

| Area | Core Problem | Key Patterns |
|---|---|---|
| **1. Data Mutation** | UI → Server changes feel slow/unreliable | Optimistic UI (instant feedback), rollback on fail, state sync |
| **2. Data Fetching** | Network load + responsive UI | Caching (memory + persistent), pagination/infinite, dedup/cancel/debounce |
| **3. Rendering** | Where to render matters for speed + SEO | CSR, SSR, SSG, Island Architecture, Streaming SSR — choose per route |
| **4. State / Data Modeling** | Duplication → bugs | Normalization, selectors, memoization, persistence (localStorage/IndexedDB/Redux) |
| **5. Infrastructure** | Delivery > code | CDN, HTTP caching, compression, CI/CD |
| **6. Cross-functional** | Not extras, core | a11y, i18n, security, observability, error handling, SEO |
| **7. Perf (Perceived)** | Feels slow even after code-opts | Skeleton screens, preload, lazy load, bundling |

## 2. HTTP Caching (Infra-level, Biggest Perceived Perf Win)

**Two levels:** App-level (code split, virtualize) vs Infra-level (CDN, compression, caching). Infra often wins more.

4 headers = complete system:

1. **Cache-Control** — who + how long fresh. `public max-age=300 stale-while-revalidate=60` → 5min fresh, 60s serve-stale-while-fetch. `private` = browser only, `no-store` = never cache (auth), `no-cache` = store but revalidate always.
2. **ETag** — fingerprint (e.g. MD5 of body). Server returns `ETag: abc123`, browser sends `If-None-Match: abc123` → `304 Not Modified` (0 body) if unchanged. Saves bandwidth.
3. **Last-Modified** — timestamp version. `Last-Modified: <date>` / `If-Modified-Since: <date>` → same 304 flow. Use when data has `updated_at`.
4. **Vary** — cache key by request header. `Vary: Accept-Language, Authorization` → separate cache per language/user. Prevents user B getting user A's board.

Together: `Cache-Control` = freshness, `ETag/Last-Modified` = validation, `Vary` = correctness.

## 3. Micro-Frontends — State Decoupling

**Problem:** Host passes state + handlers as props to micro-frontend → tight coupling.

**Fix:** Micro-frontend owns its state via `nanostores` atom.

```ts
// stores/count.state.ts
import { atom } from 'nanostores'
export const count = atom(0) // useStore(count) in React

// Module Federation: expose it
// rsbuild.json / federation.config → exposes: { "count.state": "./src/stores/count.state.ts" }

// Host listens, does NOT mutate
import { count } from 'analytics/count.state' // async import at runtime
count.listen(value => console.log(value))
```

**Rule:** Don't expose raw `atom` — expose `listenToCount(callback)` wrapper. Otherwise host can `count.set(100)` and mutate micro's state.

## 4. TypeScript — Type Safety

- **PropTypes** = runtime check, only warns in console, need to watch console. Last resort.
- **TypeScript** = build-time check, preferred 100% if possible.

| `any` | `unknown` | `as` casting |
|---|---|---|
| Gives up 100%, infects all downstream code | Forces check — must narrow before use | Override TS: "trust me this is Todo". `data as Todo` — no validation, last resort only. Better use type guards. |

- `any` → never use, `unknown` → safe fetch result, `as` → only when you + backend team know shape and guard is too heavy.
- `fetch(...).then(r=>r.json())` returns `any` — cast or validate.

## 5. Domain Modeling (Garden Example)

Validation rules as data, not `throw`:

- `ValidationResult = {valid: boolean, reason?: string}` — return error, don't throw.
- Rules: `inBounds`, `notOnTop`, `max 80% full`, `sun zone`, `antagonist adjacency` (tomato+potato → block).
- Store rule data at category level (garlic vs tomato), not instance level.

## 6. CSS-simplified.md

Empty file — no knowledge extracted.

## 7. Frontend Cache — useQuery Pattern (Deep lines 240-400)

Cache = Map (set/get/has/delete) + patterns: SWR (cached instantly, refresh background) + TTL (expiry) + dedup (2 components same key → reuse in-flight, single global cache like React Query) + invalidation (after mutation clean/patch). Impl: `CacheEntry {data, timestamp, loading, error}` + `QueryProvider` (global Map) + `useQuery(key, fetcher, ttl)` → check cache first else fetch & save. Benefit: speed, instant UI, consistency.

## 8. Data Mutation (Deep)

CRUD → POST/PATCH/PUT/DELETE via fetch. Add: POST /api/cards {title, columnId} → id = total+1 → push to column 201. Update: PATCH partial vs PUT full replace (merge challenge). Delete: remove now or confirm. With cache: mutation → addCard updates normalized store (cards + columns cardIds) → re-render. Must handle rollback on fail.

## 9. SSR Deep (lines 320-480)

CSR = blank #root, JS builds HTML → 2-5s blank on slow 4G, view-source empty. SSR = Express fetches board+user → normalize → `renderToPipeableStream(<BoardProvider initialData>)` → stream HTML + `<script>window.__INITIAL_DATA__</script>` → browser paints instantly → `hydrateRoot` attaches listeners. Must match or React warns. Streaming = chunked HTML. Build: vite build + ssr entry.server → dist/server/entry-server.js.

## 10. Performance Measurement (Deep)

Core Web Vitals: LCP (main content), CLS (jump), INP (lag). Others: TTFB (first byte), TTI (usable). Lab: Lighthouse (LCP/CLS/INP), PageSpeed Insights (URL), React Profiler (extra renders). Workflow: baseline → apply lazy/SSR/prefetch → re-measure. Lab = fast laptop, field = real user 3G (75th percentile). Dev 5173 not accurate, prod 4000 SSR = 100 score.

## 11. Realtime Updates (Deep)

Polling (blunt, waste) vs SSE vs WebSocket. SSE: `new EventSource(/api/board/:id/events)` + `addEventListener('card_assigned')`, server EventEmitter (on/off/emit) + keep-alive + sendEvent(type, JSON) → broadcast all tabs. MSW moved to Express 4000 proxy /api→4000. One-way, lightweight, auto-reconnect. WebSocket two-way for chat/games. Prod: heartbeat, HTTP/1.1 limit ~6 → share one EventSource, HTTP/2 multiplexed.

## 12. Prefetch (Deep)

Fetch before click on hover/focus/touch → cache TTL 1-2min → instant. Ex: user picker 5/page slow 2s → hover avatar triggers usePrefetch same key as useQuery → next click hits cache, next page not cached shows dots. Rules: keys must match exactly, trigger only on intent +100ms delay, sensible TTL, debounce spam guard.

## 13. Accessibility (Deep)

Build in from start, not retrofit. Why: WCAG 2.1 AA legal, +50% users, SEO. Patterns: progressive enhancement, semantic HTML (article/h3/button/nav), ARIA, tab order, focus ring. Tests: jest-axe toHaveNoViolations (li in ul/ol), eslint-plugin-jsx-a11y, Lighthouse a11y. Keyboard: Tab, focus moves to next card after delete, shortcuts.


## 14. React — Router & Data Loading (reactjs.md — sampled)
- Setup: `npm create vite@latest` -> `react-router-dom` + `axios` -> `RouterProvider` + `createBrowserRouter`.
- Nested routes: `path: "/"` -> `element: <RootLayout>` -> `children: [{path: "posts", children: [{index: true}, {path: ":id"}]}]` — no leading slash in children.
- RootLayout: `container` + `nav` + `<Outlet />` + `<ScrollRestoration />` — fixes random scroll jump, wrap multi-elements in fragment.
- Colocated route object: `export const postListRoute = { element: <PostList/>, loader }` -> router me `...postListRoute` spread. Loader file bloat se bacho.
- `index: true` + `<Navigate to="/posts" />` for default redirect.

## 15. React — useEffect Fetch Basics (reactjs2.md — sampled)
- `const [users, setUsers] = useState([])` — default `[]` not undefined else `users.map` crash.
- `useEffect(() => { fetch }, [])` — empty deps = run once; missing deps = refetch every render.
- `key={user.id}` for list, `isLoading` -> `.finally(()=>setIsLoading(false))`.
- Abort: `const controller = new AbortController(); return ()=>controller.abort()` — prevents StrictMode double-fetch race.
- Test loading with DevTools Network -> Slow 3G; fallback `public/users.json`.

## 16. TypeScript Monorepos (TypeScript Monorepos Architect.md — sampled)
- Don't split early: start fat core, split only when need — easy to extract + re-export, hard to merge after publish.
- Exception: plugin system -> many packages early + invest in plugin interface.
- NX tricks: fake `project.json` in any folder -> `nx <project> <target>` -> noun-verb commands.
- Zod > Arctype — popularity = StackOverflow + LLM help; pnpm = faster + smaller via linking.

## Frontend System Design Essentials.md Learnings

> Source: `Transcripts/Frontend/Frontend System Design Essentials.md` — pura 0-end deep padha (29 modules). Is file ka saara seekha hua yahi ek section me hai.

**Jo seekha:**
- **7 Areas mindset:** Component-first → Architecture (data flow, failure, trade-offs)
- **Cache useQuery:** Map + SWR + TTL + dedup (single global cache) + invalidation; `QueryProvider` + `CacheEntry`
- **Mutation:** POST/PATCH/PUT/DELETE → `addCard` store update + rollback on fail; Delete me `card + position` save karke `insertCard` se restore
- **SSR:** CSR blank → SSR `Express + renderToPipeableStream + hydrateRoot + window.__INITIAL_DATA__` + streaming, build `dist/server`
- **Perf Measurement:** LCP/CLS/INP, TTFB/TTI, Lighthouse/PageSpeed/React Profiler, lab vs field 75th percentile
- **Realtime:** Polling vs SSE `EventSource + EventEmitter keep-alive` vs WebSocket (two-way)
- **Prefetch:** hover pe same cache key se `usePrefetch`, TTL 1-2min, 100ms delay
- **a11y:** WCAG 2.1 AA, semantic HTML, jest-axe, eslint-jsx-a11y, keyboard Tab + focus
- **Optimistic Updates:** `previousAssignee` save → UI instant → PATCH → fail pe rollback
- **Drag & Drop:** `POST /api/cards/:id/move` + `@atlaskit/pragmatic-drag-and-drop` (`draggable` + `dropTargetForElements`) + keyboard menu `Move to [Column]`
- **Routing + Lazy:** `/your-work`, `/board/:id`, `/settings` → `createBrowserRouter` vs `StaticRouter` same routes → `lazy(() => import)` → 433kb → ~200kb + chunks + `<Suspense>`
- **Error Boundaries:** class `getDerivedStateFromError` + `componentDidCatch` → layering: UserSelect → Card → App
- **Normalization:** nested `cards[].assignee` → `{columnOrder, columnsById, cardsById, usersById}` → `BoardProvider ingestBoard/upsertUser` → search sirf `columns` clear, `usersById` persistent
- **Structure:** Feature-based (default) vs Component-based vs Atomic vs MVVM → start simple, `shared` tab extract, ESLint enforce
- **Security:** XSS `<img onerror>` + `dangerouslySetInnerHTML` → `DOMPurify.sanitize` + CSP `default-src 'self'` → defense in depth
- **Request Mgmt:** Race `i→in→inst` out-of-order → `AbortController` + cleanup → Debounce 300ms (search) vs Throttle 200ms (scroll/drag)
- **Pagination:** offset `?page=0&pageSize=5` → `{items, pageInfo{hasMore}}` → `onMenuScrollToBottom` append; cursor for infinite feed
- **Skeleton:** 4 perceived perf (Skeleton, Optimistic, Prefetch, SSR) → match layout, `animate-pulse`, timebox 100ms, Slow 3G test
- **Testing:** Pyramid → Vitest+JSDOM `renderWithProviders` + Playwright `POST /api/test/reset` → flows load/assign/delete, smoke tests
- **Mental Models:** Build→Deploy→Runtime + Pillars (Modeling→Fetching→Mutation→Perf→Cross-functional) checklist
- **AI:** junior dev jaisa → small steps me break, delegate+review, supporting tasks offload

## TyeScript Simplified.md Learnings

> Source: `Transcripts/Frontend/TyeScript Simplified.md` — pura deep padha (0-640+ lines). Is file ka saara seekha hua yahi ek section me hai.

**Jo seekha:**
- **PropTypes vs TS:** PropTypes runtime check, console warning only, `.isRequired`, `oneOf/oneOfType/arrayOf/shape/exact/node` — last resort. TS build-time check 100% prefer.
- **any vs unknown vs as:** `any` infects full code, `unknown` forces narrowing before use, `as Todo` casting tells TS trust me — no validation, last resort only, better type guards. `fetch().json()` returns `any`.
- **void vs undefined:** `void` = function returns nothing (no usable return), `undefined` = explicit `return undefined` usable as value. Hover shows difference.
- **never + Exhaustive Check:** `never` = impossible to reach. Switch `default: const x: never = priority` → if new value `super-low` added without case, TS error. Pattern for union completeness.
- **TS in React (Vite):** `npm create vite -- template react-ts`, `.tsx` not `.jsx`, `tsconfig.json`, `import type { ReactNode }`, `Vite.env.d.ts`, `child props: {name: string}` inline or `type ChildProps`, `React.FC` adds extra props (not preferred), `children: ReactNode` optional `?`, HTML props: `ComponentProps<'button'> & {outline?: boolean}`, custom component `ComponentProps<typeof Child>`.
- **Function Overloads:** Multiple signatures before impl: `function sum(nums: number[]): number; function sum(a: number, b: number): number; function sum(a: number | number[], b?: number) { if(Array.isArray(a)) reduce else a+b }` — impl must be generic.
- **JS → TS Conversion:** Top-down (feature first, quickly spirals 1→10→60 files) vs Bottom-up (leaf functions first, easier). For bundler Vite check `vitejs` TS docs or create fresh TS project and copy config.
- **Calendar Project Patterns:** `date-fns` (startOfWeek/startOfMonth/endOfWeek/endOfMonth/eachDayOfInterval/addMonths/subMonths/isSameMonth/isBefore/isToday/parse/format), `useState<Date>(new Date())` + `useMemo(() => calcDays, [selectedMonth])` to avoid calc each render, `key={day.getTime()}`.
- **Helpers:** `formatDate(date, options?: Intl.DateTimeFormatOptions)` using `new Intl.DateTimeFormat(undefined, options).format(date)` for locale; `CC(...classes: unknown[]) => classes.filter(c=>typeof c==='string').join(' ')` for conditional classes.
- **Modal:** `createPortal(<div className="modal"><div className="overlay" onClick={onClose}><div className="body">{children}</div></div>, document.querySelector('#modal-container') as HTMLElement)`, `useEffect` for `Escape` key add/remove listener, `formId = useId()` for label/htmlFor.
- **Context + Union Types:** `Event = {id: string, name: string, date: Date, color: typeof eventColors[number], ...} & ({allDay: true, startTime?: never, endTime?: never} | {allDay: false, startTime: string, endTime: string})` — checks allDay determines required fields. `UnionOmit<T,K>` custom: `T extends unknown ? Omit<T,K> : never` to distribute Omit over union (normal Omit breaks unions).
- **Events Context:** `createContext<EventsContext | null>(null)`, `useEvents()` throws if null, provider `useState<Event[]>` + `addEvent(e: UnionOmit<Event,'id'>) => setEvents([...events, {...e, id: crypto.randomUUID()}])`, `updateEvent(id, details) => map`, `deleteEvent(id) => filter`. Wrap entire `Calendar` in provider.
- **useLocalStorage:** `useState(() => { json = localStorage.getItem(key); if null return initial; parsed = JSON.parse(json) as Event[]; return parsed.map(e=> e.date instanceof Date ? e : {...e, date: new Date(e.date)}) })` + `useEffect(() => localStorage.setItem(key, JSON.stringify(value)), [value,key])`, `as const` for return.
- **Closing Animation:** `isClosing` state + `prevIsOpen = useRef<boolean>()` + `useLayoutEffect(() => { if(!isOpen && prevIsOpen.current) setIsClosing(true); prevIsOpen.current = isOpen }, [isOpen])` + `if(!isOpen && !isClosing) return null` + `className={CC('modal', isClosing && 'closing')}` + `onAnimationEnd={() => setIsClosing(false)}`.
- **Overflow (bonus):** `OverflowContainer` takes `items: Event[], renderItem: (e)=>ReactNode, renderOverflow: (hiddenCount)=>ReactNode, className` — imperative measure container size vs children, hide overflow and show `+N more` button.

"## Model Complex Domains with TypeScript.md Learnings

> Source: `Transcripts/Frontend/Model Complex Domains with TypeScript.md` — pura 0-720 deep padha. Is file ka saara seekha hua yahi ek section me hai.

**Jo seekha:**
- **Why Domain Modeling:** Business ambiguity handle karna = staff skill. Product + dev + AI same ubiquitous language use kare. DDD toolbox hai, rigid nahi (Eric Evans).
- **Why TS:** Expressive but readable, protobuf/JSON schema ka source of truth, SDK generate ho sakta. Garden jaise complex domain ke liye best.
- **Garden Example:** 40 raised beds, seed catalog YAML, seed packets (quantity, viability expire), hydroponic start, planting distance/size, seasonality.
- **Entities vs Values vs Resources:** `types` pkg me entities (DB), values (no ID, embedded), resources (API request/response). Generic `Packet<Metadata>` + `presentation {icon, accentColor}` + `metadata {quantity, plantingDistance, daysToHarvest, expiresAt}`. API != DB → route handler me adapt.
- **New Field Flow:** `daysToHarvest (>0)`, `plantingDistance`, `expiresAt: string` wire → entity `Date` → `parseSeedPacket(rawYAML)` me `new Date(2024 + viabilityInYears,1,1)` → TSConfig `rootDir: src` fix + `rm -rf dist && build` se `NOT NULL constraint fails` gaya.
- **Relationships:** `SeedPacket 1—* Plant *—1 Bed *—1 Garden`. Plant = join table + extra (`position: XYCoordinate`, `bed`, `seedPacket`). TypeORM `@Entity`, `@Column !`, `@ManyToOne/@OneToMany`. Value object `Temperature{value, unit:'C'|'F'}` embedded with prefix, `MonthlyTemperatureRange{month, min:Temperature, max:Temperature}`.
- **Validation Rules:** `ValidationResult {valid, reason?}` return not throw. Rules: `inBounds`, `notOnTop`, `max 80% full`, `sun zone`, `antagonist adjacency` (tomato+potato). Data **category level** pe store, instance pe nahi. `Indicator {id, effect{targetItemTypeId, effect: beneficial/harmful/neutral, description}}` directional, bi-directional possible.
- **Rule Formalizing:** `preventAntagonistPlantAdjacency` → `indicators.flatMap(effects)` → if `harmful && target==item.category && bed.placements.some(p.category==source)` → valid=false. 30 rules bhi modular, unit testable.

## Covered"
- **Deep Done:** Frontend System Design Essentials.md FULLY (upar wala section)
- **Sampled (6):** Micro-Frontends, Model Complex Domains, TyeScript Simplified, reactjs, reactjs2, TypeScript Monorepos — 80 lines only
- **Empty:** CSS-simplified.md — empty, done
- **Pending (1):** Frontend system design common pitfalls.pdf — binary

Related: `Learnings/README.md` pattern — only principles/patterns/quotes, no fluff."