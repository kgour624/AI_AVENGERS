# Frontend Essentials — Architecture Knowledge Base

Source: Transcripts/Frontend/ (Frontend System Design Essentials.md deep 0-720, Frontend Architecture Monoliths to Micro-Frontends.mf.md, TyeScript Simplified.md, Model Complex Domains with TypeScript.md, CSS-simplified.md, reactjs.md, reactjs2.md, TypeScript Monorepos Architect.md) — Sneha Mehra. Permanent reference for frontend system design.

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

Cache = Map (set/get/has/delete) + patterns: SWR (cached instantly, refresh background) + TTL (expiry) + dedup (2 components same key → reuse in-flight, single global cache like React Query) + invalidation (after mutation clean/patch). Impl: `CacheEntry {data, timestamp, loading, error}` + `QueryProvider` (global Map) + `useQuery(key, fetcher, ttl)` → check cache first else fetch & save. Benefit: speed, instant UI, consistency, less server load.

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

## 14. React — Router & Data Loading (reactjs.md — sampled, deep pending)

- Setup: `npm create vite@latest` → `react-router-dom` + `axios` → `RouterProvider` + `createBrowserRouter`.
- Nested routes: `path: "/"` → `element: <RootLayout>` → `children: [{path: "posts", children: [{index: true}, {path: ":id"}]}]` — no leading slash in children.
- RootLayout: `container` + `nav` + `<Outlet />` + `<ScrollRestoration />` — fixes random scroll jump, wrap multi-elements in fragment.
- Colocated route object: `export const postListRoute = { element: <PostList/>, loader }` → router me `...postListRoute` spread. Keeps element+loader together, avoids file bloat. `loader({request, params})` → `axios.get(url, {signal: request.signal}).then(r=>r.data)` → `useLoaderData()`.
- `index: true` + `<Navigate to="/posts" />` for default redirect.

## 15. React — useEffect Fetch Basics (reactjs2.md — sampled, deep pending)

- `const [users, setUsers] = useState([])` — default `[]` not undefined else `users.map` crash.
- `useEffect(() => { fetch }, [])` — empty deps = run once; missing = refetch every render.
- `key={user.id}` for list, `isLoading` → `setIsLoading(true)` before fetch, `.finally(()=>setIsLoading(false))` → `isLoading ? <h2>Loading</h2> : <ul>`.
- Abort: `const controller = new AbortController(); fetch(url, {signal: controller.signal}); return ()=>controller.abort()` — prevents StrictMode double-fetch race.
- Test loading with DevTools Network → Slow 3G; fallback `public/users.json` → `fetch("users.json")` if API down.

## 16. TypeScript Monorepos (TypeScript Monorepos Architect.md — sampled, deep pending)

- Don't split early: start with fat core package, split only when need arises — easy to extract + re-export, hard to merge fragmented packages after publish.
- Exception: plugin system → many packages early + invest in plugin interface.
- NX tricks: fake `project.json` in any folder → `nx <project> <target>` → make `nx dev serve` ergonomic via noun-verb naming; targets don't need 1:1 package.
- Tool choice: Zod > Arctype — popularity = StackOverflow + LLM examples; Arctype = TS syntax at runtime but less proven.
- pnpm sell: faster install → faster CI, smaller disk via linking, less RAM; migration lift = people (training) not package count.

## Covered

- **Deep Done (1/10):** Frontend System Design Essentials.md 0-720 lines deep (7-13 above)
- **Sampled (6/10):** Micro-Frontends, Model Complex Domains, TyeScript Simplified, reactjs, reactjs2, TypeScript Monorepos — 80 lines only, deep pending
- **Empty (1/10):** CSS-simplified.md
- **Pending PDFs (2):** common pitfalls.pdf (binary)

Related: `Learnings/README.md` pattern — only principles/patterns/quotes, no fluff.
