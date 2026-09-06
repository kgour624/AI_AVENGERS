# ARC-51 UI Redesign — Interface-First Contract

> **Status:** APPROVED DESIGN — SECTION-BY-SECTION ROLLOUT IN PROGRESS
> **Author:** System Design Architect
> **Last Updated:** 2026-09-06
> **Approved by:** Kiran Nogia (admin) — "ARC-51 / J.A.R.V.I.S. Command Center" direction, + 2 amendments (holographic/cybernetic avatar silhouettes, framer-motion for spring physics where CSS isn't smooth enough).

---

## 1. What This Contract Locks Down

This is a **presentation-layer-only** redesign. Per Interface-First discipline (see `docs/INTERFACE_FIRST_CONTRACT.md`), every new visual element's contract is defined here BEFORE implementation, and every existing functional contract (API calls, hooks, stores, loaders) is explicitly declared **out of scope and untouched** below — not assumed safe, declared safe, with the exact file list that proves it.

**Rule for every rollout section:** change `className`, add wrapper components (motion/glow/avatar), add new CSS/tokens. Never change: a `queryFn`, a `mutationFn`, a loader, a Zustand store's state shape, an API function's request/response handling, or a route path. If a section's redesign seems to require one of those, STOP and flag it — don't silently do it.

---

## 2. Design Tokens (additive, not replacing the existing mode-color system)

The existing OKLCH response-mode colors (`--color-advise`, `--color-ask`, `--color-warn`, `--color-pushback`, `--color-refuse`) are **product semantics, not decoration** — they encode Gate output meaning (China Wall / Decision Engine). ARC-51 does not replace them; it adds a layer around them.

### New CSS custom properties (`design-system/tokens.css`)

| Token | Value | WHEN used |
|---|---|---|
| `--surface-void` | `oklch(9% 0.012 270)` | Page background (darker than existing `--surface-base`) |
| `--surface-panel` | existing `--surface-raised`, unchanged | Glass card base |
| `--surface-panel-hover` | `oklch(18% 0.015 270)` | Card hover state |
| `--glow-cyan` | `oklch(75% 0.15 200)` | System/neutral accents, HUD lines |
| `--glow-purple` | `oklch(65% 0.20 300)` | Brand identity (logo, active nav, primary CTA) |
| `--glass-border` | `oklch(100% 0 0 / 0.08)` | Panel borders (replaces flat `--surface-border` ONLY on glass panels, not on plain cards like admin tables) |
| `--ease-arc` | `cubic-bezier(0.16, 1, 0.3, 1)` | Every CSS transition/animation in ARC-51 components |

### Tailwind extensions (`tailwind.config.ts`)

Mirrors the tokens above as utility classes (`bg-surface-void`, `text-glow-cyan`, `border-glass-border`, etc.), following the exact pattern the file's own header comment already documents ("Tailwind's JIT compiler needs literal color values"). **Also fixes the pre-existing `-hover` gap** (bug found in an earlier session: `brand-hover`, `mode-refuse-hover` etc. were referenced in components but never defined in this config) — folded into this same pass since it's the same file.

### Locked decision

**WHEN** adding any new visual token
**DO** add it here first, then to `tokens.css`, then to `tailwind.config.ts`, in that order
**BECAUSE** this is the same discipline as the API Interface-First contract — a token used in a component before it exists in the config either silently fails (unknown Tailwind class → no CSS generated) or drifts from the CSS source of truth
**EXCEPT** none.

---

## 3. Motion Contract

### Library boundary (explicit, per admin's amendment)

| Use case | Tool | WHY |
|---|---|---|
| Button/card hover, press, simple fades | CSS transitions (`--ease-arc`) | Cheap, no JS, GPU-accelerated via `transform`/`opacity` only |
| Modal open/close | **framer-motion** (`AnimatePresence`) | Admin's explicit ask — CSS keyframe approach (current `Modal.tsx`) is replaced with spring physics for real fluidity |
| Staggered list/grid entrance (project cards, expert cards) | **framer-motion** (`motion.div` + `staggerChildren`) | Same reason — CSS stagger via `animation-delay` is janky when list length is dynamic |
| Expert avatar status rings (idle/analyzing/responded) | CSS (`conic-gradient` + `@keyframes rotate`) | Runs continuously per-card; framer-motion here would be unnecessary JS overhead for a purely decorative infinite loop |
| Route transitions | CSS fade (existing `AppShell`/`Outlet` boundary) | framer-motion's `AnimatePresence` for ROUTE-level transitions requires restructuring the router's element tree — explicitly deferred, not done in this pass, to avoid touching routing logic (out of scope per §1) |

### New dependency

`framer-motion` added to `frontend/package.json`. This is the one new runtime dependency this redesign introduces — flagged explicitly rather than silently added, per package-recommendation discipline.

### Timing constants (single source of truth, `design-system/motion.ts`)

```ts
export const ARC_MOTION = {
  micro: 0.12,   // button press, hover
  panel: 0.28,   // modal/panel enter-exit
  stagger: 0.04, // per-item delay, list entrance
  maxStaggerItems: 6, // beyond this, stagger delay is clamped to avoid a sluggish feel on long lists
  ease: [0.16, 1, 0.3, 1], // matches --ease-arc
}
```

### Reduced motion

**WHEN** `prefers-reduced-motion: reduce` is set
**DO** every framer-motion component here checks `useReducedMotion()` and swaps to instant opacity-only transitions; every CSS animation gets a `@media (prefers-reduced-motion: reduce)` override to `animation: none`
**BECAUSE** accessibility is non-negotiable, not a nice-to-have
**EXCEPT** none.

---

## 4. Component Contracts (backward-compatible prop additions only)

### `Modal.tsx`
- **Before:** CSS-keyframe open/close (`modal-enter`/`modal-closing` classes), manual `isClosing` state machine.
- **After:** Internals replaced with `AnimatePresence` + `motion.div` spring transition. **Public prop contract (`isOpen`, `onClose`, `children`, `className`) is unchanged** — every existing caller (11 call sites: CreateProjectModal, CreateExpertModal, EditCharterModal, EditProjectModal, RepoConnectModal, TranscriptUploadModal, ExpertTopicsModal, etc.) needs zero changes.

### `Card.tsx`
- New optional prop: `glow?: 'cyan' | 'purple' | 'none'` (default `'none'`). Adds hover border-glow. **Every existing usage that doesn't pass `glow` renders pixel-identical to today** (default is a no-op).

### `Button.tsx`
- No prop changes. `active:scale-95` + `--ease-arc` transition added to the existing `variantClasses`/base className string only.

### New component: `ExpertAvatar.tsx` (`components/expert/`)
- **Props:** `{ domain: string; status: 'idle' | 'analyzing' | 'responded'; mode?: ResponseMode; size?: 'sm' | 'md' | 'lg' }`
- **Contract:** pure presentational, no data fetching, no API calls — receives `domain`/`status`/`mode` from whatever already has that data (ExpertCard, ExpertResponse, StreamingIndicator). Does not replace `ExpertCard`/`ExpertBadge` — composes inside them.
- **Domain → visual signature mapping:** since `domain` is backend free text (not an enum — same finding as HANDOFF's existing `ExpertBadge` color-hash gap), signatures are chosen by **keyword match** (`system_design` → hex-lattice, `database`/`db` → stacked-cylinder, `security` → shield-lattice, `architecture` → blueprint-grid) with a **generic neural-core silhouette fallback** for any unmatched domain string — never a blank/broken avatar.

### New file: `design-system/motion.ts`
Exports `ARC_MOTION` constants (§3) and shared framer-motion variant objects (`fadeUp`, `staggerContainer`, `modalSpring`) so every component pulls from one definition instead of redefining spring configs inline.

---

## 5. Expert Avatar Visual Spec (holographic/cybernetic amendment)

Per admin's correction: pure geometric shapes (hexagon/cylinder/shield) alone read as "icons," not "living agents." Fix:

- Every avatar SVG is a **two-layer composite**: (1) the domain geometric base shape (unchanged concept), **behind** (2) a **simplified holographic humanoid/robot-head silhouette** (visor-line face, no detailed features — stays abstract/iconic, not literal robot clip-art) rendered in a lower-opacity cyan wireframe, scan-line effect optional via CSS `mask` gradient animation.
- Status ring (outer, per §3's motion contract) sits OUTSIDE this composite, never overlapping the silhouette itself — keeps the "is this agent active" signal legible even if the inner art gets busy.
- Delivered as **inline SVG React components** (not image files) — lets status/mode colors be applied via `currentColor`/CSS variables rather than baking colors into static assets, consistent with the token system in §2.

---

## 6. Login Page — Full Redesign Contract

### Functional contract (UNCHANGED — verified against current `LoginPage.tsx`)

| Element | Current behavior | Contract |
|---|---|---|
| Form fields | `email`, `password` (controlled state) | Unchanged — same two fields, same trim-on-submit, same disabled-while-submitting |
| Submit | Calls `login({email, password})` from `api/auth.ts` | **Unchanged call site** — zero API contract change |
| Success | `setAuth(user, tokenPair.accessToken)` then `navigate('/', {replace: true})` | Unchanged |
| Failure | `setError(handleAPIError(err))`, password/email NOT cleared | Unchanged — this was a deliberate UX decision (see file's own header comment), preserved |
| Register link | `<Link to="/register">` | Unchanged |

### Visual contract (new)

- Full-viewport ARC-51 scene: `--surface-void` background, animated ambient glow following a slow drifting radial gradient (CSS `@keyframes`, GPU-cheap — NOT mouse-tracked, to avoid input-lag risk on low-end hardware).
- **Domain expert constellation**: 4-6 `ExpertAvatar` instances (from §4/§5, status=`idle`, using real domain strings this product actually has — system_design/database/security/architecture, not invented ones) arranged in a loose orbit around the login card, each with independent slow-drift animation — purely decorative, `aria-hidden`, zero data fetching (no API call to list real experts on an unauthenticated page — that would be a scope violation of §1, since `/experts` requires a JWT).
- Login card itself: glass panel (§2 tokens), framer-motion entrance (fade+scale, `ARC_MOTION.panel`), form fields keep the existing `Input`/`Button` components unchanged (only their already-token-driven classNames render differently once §2's tokens exist — no `Input.tsx`/`Button.tsx` logic changes needed here specifically).
- Heading copy: "AI AVENGERS" wordmark with glow-purple treatment + subtitle "Neural Command Center — Multi-Agent Domain Expert Simulator" (evolves the existing "Domain Expert Team Simulator" line, doesn't remove its meaning).

---

## 7. Section-by-Section Rollout Order

Per this project's own verification rule — one section at a time, report VERIFIED/GAP, then next:

| Order | Section | Depends on |
|---|---|---|
| 1 | Design tokens (`tokens.css`, `tailwind.config.ts`) + `framer-motion` dependency + `design-system/motion.ts` | none — foundation |
| 2 | Login page (explicit priority this session) | 1 |
| 3 | `Modal.tsx` framer-motion conversion (used everywhere — do once, benefits every modal immediately) | 1 |
| 4 | `ExpertAvatar.tsx` component | 1 |
| 5 | `Card.tsx` glow prop + Button press/hover polish | 1 |
| 6 | Header/Sidebar glass treatment | 1, 5 |
| 7 | Projects empty-state HUD hero + quick-start templates (real `createProject` calls, not fake buttons) | 1, 5 |
| 8 | Expert cards (`ExpertCard`, `ExpertsPage`) wire in `ExpertAvatar` | 4 |
| 9 | Chat interface (StreamingIndicator neural-processing visual, ExpertResponse card polish, citation chip hover) | 1, 4, 5 |
| 10 | Admin panel glass pass (lower drama, higher density per §1's page-by-page notes) | 1, 5, 6 |

No section starts until the previous one is verified. This document is updated with a §8 progress log as each section completes.

---

## 8. Progress Log

### §1 Design Tokens — VERIFIED ✅
- `tokens.css`: added `--surface-void`, `--surface-panel-hover`, `--glow-cyan`, `--glow-purple`, `--glass-border`, `--ease-arc`. Existing mode-color tokens untouched.
- `tailwind.config.ts`: added ARC-51 utility colors + **also fixed the pre-existing `-hover` gap** (`brand-hover`, `mode-*-hover` were referenced by `Button.tsx`/`RouteError.tsx`/`AppErrorBoundary.tsx` since an earlier session but never defined here — real bug, fixed opportunistically in the same pass since it's the same file/same root cause as ARC-51's own token additions).
- `frontend/package.json`: added `framer-motion` dependency.
- `design-system/motion.ts` created: `ARC_MOTION` constants, `modalSpring`, `fadeUp`, `staggerContainer` variants.
- **Not build-verified** (no Node toolchain access this session) — same standing caveat as every prior frontend change in this repo. Run `npm install && npm run build` to confirm.

### §4 ExpertAvatar — VERIFIED ✅, REORDERED
- **Deviation from §7's stated order, flagged here rather than silently done:** §6 (Login page) was written to depend on `ExpertAvatar` for its decorative constellation. Rather than ship a throwaway placeholder for Login and rebuild it properly two sections later (order 4 in §7), `ExpertAvatar.tsx` was pulled forward and built for real, immediately after §1's tokens (its only real dependency). No section was skipped — order 3 (Modal.tsx) is still next, unaffected by this reorder.
- `components/expert/ExpertAvatar.tsx`: domain→shape keyword matching (hexagon/cylinder/shield/blueprint/generic-core fallback — same free-text-domain caveat as `ExpertBadge`'s existing color-hash), holographic visor-silhouette overlay per admin's amendment, status ring (idle/analyzing/responded) in CSS (not framer-motion, per §3's boundary — continuous decorative loop).
- `animations.css`: added `arc-ring-rotate`, `arc-ring-pulse-once`, with `prefers-reduced-motion` overrides.
- Not yet wired into `ExpertCard.tsx`/`ExpertResponse.tsx` (that's still §8 in §7's order) — only consumed by the Login page's decorative constellation so far.

### §2 Login Page — VERIFIED ✅
- Functional contract cross-checked line-by-line against the pre-existing `LoginPage.tsx` before touching anything: same `email`/`password` state, same trim-on-submit, same `login()` call site, same `setAuth`+`navigate` success path, same error handling (password/email NOT cleared on failure — preserved deliberate UX decision from the file's own original header comment). **Zero API/store/routing changes.**
- Visual: `surface-void` background, two CSS ambient drifting glows (not mouse-tracked), 4-avatar decorative constellation (`aria-hidden`, no API call — `/experts` needs a JWT this page doesn't have), glass login card with `framer-motion` entrance spring (`modalSpring`-equivalent inline, respects `useReducedMotion`).
- **Bug caught and fixed in this same section:** the em dash in the new subtitle copy was written as a bare `\u2014` in JSX children text (same class of bug fixed repeatedly elsewhere this session — Header.tsx, AdminDashboard.tsx, AdminClients.tsx, ProjectMemoryPanel.tsx). Wrapped in `{'\u2014'}` before this was reported back, since I re-read my own output before calling it done.

### §3 Modal.tsx — VERIFIED ✅
- Replaced the CSS-keyframe `isClosing`/`previousIsOpen` state machine with `AnimatePresence` + `motion.div` using the shared `modalSpring` variant (`design-system/motion.ts`). Backdrop gets a light `backdrop-blur-sm`, panel gets `backdrop-blur-xl` + `border-glass-border` (§2 tokens).
- **Public prop contract unchanged** (`isOpen`, `onClose`, `children`, `className`) — confirmed safe by construction (no prop added/removed/retyped), so every existing caller (`CreateProjectModal`, `CreateExpertModal`, `EditCharterModal`, `EditProjectModal`, `RepoConnectModal`, `TranscriptUploadModal`, `ExpertTopicsModal`) needs zero changes.
- Reduced-motion contract honored: `useReducedMotion()` swaps the spring variant for an opacity-only one at `duration: 0`, rather than skipping `AnimatePresence` (it still needs to control mount/unmount timing even with motion disabled).
- Cleanup: removed the now-fully-unused `.modal-enter`/`.modal-closing` classes and `@keyframes modal-in`/`modal-out` from `animations.css` — verified nothing else referenced them before deleting, not assumed.

### §5 Card + Button polish — VERIFIED ✅
- `tailwind.config.ts`: added `transitionTimingFunction.arc` so `ease-arc` is usable as a utility class everywhere (exposes `--ease-arc` from §1's tokens.css addition).
- `Card.tsx`: new optional `glow?: 'cyan'|'purple'|'none'` prop, default `'none'` — renders pixel-identical to the pre-ARC-51 Card when omitted. Every existing `<Card>` usage (AdminClients, AdminExperts, AdminDashboard, ProjectCard, etc.) needed zero changes.
- `Button.tsx`: `active:scale-95` press feedback + `ease-arc` transition, scoped to exact properties (`background-color,transform,box-shadow`) rather than `transition-all`, to never accidentally animate a layout-affecting property. Disabled buttons don't scale (`disabled:active:scale-100`).

### §6 Header/Sidebar glass — VERIFIED ✅
- `AppShell.tsx`: `bg-surface-base` → `bg-surface-void` (§2 token) so the glass Header/Sidebar read as floating above the page.
- `Header.tsx`/`Sidebar.tsx`: `border-glass-border` + `bg-surface-raised/70 backdrop-blur-xl` glass treatment. **Every handler/hook is untouched** (`toggleSidebar`, `handleLogout`, `useQuery(queryKeys.projects.all)`, `useParams` active-project detection) — confirmed by construction, since only `className` strings and a motion wrapper were added, no logic lines changed.
- Sidebar's project list now uses `motion.ul`/`motion.li` with the shared `staggerContainer`/`fadeUp` variants (§3), capped at `ARC_MOTION.maxStaggerItems` per motion.ts's own documented consumer-side cap.
- Active project indicator changed from a `\u25b8` text-prefix to a `border-glow-purple` left-border (same signal, different visual language — not a regression, a presentational swap).

### §7 Projects empty-state HUD hero — VERIFIED ✅
- `ProjectsEmptyHero.tsx` (new): glowing rotating AI-core graphic (CSS, reuses `arc-ring-rotate`), heading/subtitle, `\u26A1 Initialize First Project` CTA (opens the existing `CreateProjectModal` — zero duplicate create-logic), and 3 quick-start blueprint templates.
- **Non-negotiable rule enforced**: templates call the real `createProject()` API directly (same function the modal uses) with a pre-filled name/description — verified this is not a decorative mockup by tracing the exact call path before considering this done.
- `ProjectsPage.tsx`: empty state renders `ProjectsEmptyHero`; non-empty state's grid now uses `motion.div`/`staggerContainer`/`fadeUp` (same pattern as Sidebar's §6 list). `handleCreated` unified so both the header's "+ New Project" button and the hero's CTA/templates share one revalidate+navigate path — no duplicated logic between the two entry points.
- `ProjectCard.tsx`: added `glow="purple"` to its existing `<Card>` usage (backward-compatible prop from §5).
- **Note on commit granularity**: this section was split into 4 small commits (new component → type doc-comment → page wiring → ProjectCard polish) after an earlier single large multi-file commit attempt was interrupted mid-flight and needed to be redone from scratch — smaller commits going forward reduce how much re-verification is needed if a single tool call fails partway through.

### §8 ExpertAvatar wired into ExpertCard/ExpertsPage — VERIFIED ✅
- `ExpertCard.tsx`: added `<ExpertAvatar domain={expert.domain} status="idle" size="md" />` next to the name, `glow="cyan"` on its `<Card>`. Status is always `'idle'` here deliberately — this is a static browsing list with no live SSE stream reaching it; `'analyzing'`/`'responded'` states are reserved for §9 (chat interface) where they're actually true.
- `ExpertsPage.tsx`: list wrapped in the same `staggerContainer`/`fadeUp` pattern as §6/§7. `onViewTopics` wiring (feature #4 fix, earlier session) untouched.

### Next: §9 Chat interface — StreamingIndicator (real ExpertAvatar analyzing state), ExpertResponse card polish, citation chip hover (order 9).
