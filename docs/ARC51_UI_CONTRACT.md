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

*(updated as each section from §7 completes)*
