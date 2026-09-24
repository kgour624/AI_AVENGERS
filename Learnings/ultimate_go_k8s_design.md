
# Ultimate Go Software Design with Kubernetes 2.0 — Architecture Knowledge Base

Source: Bill Kennedy / Ardan Labs transcript (garage-sale "service" domain). Permanent reference for system-design architect work.

## 0. Three Levels of Firewalls (master mental model)

Everything is about isolating parts so humans can hold mental models:

1. **Packaging** (micro): every Go package is a static API; "we don't organize code so much as we organize APIs."
2. **Horizontal layering**: API → App → Business → Foundation (+ vendor, + Zarf). Horizontal firewalls where data is *shaped*.
3. **Domains** (vertical firewalls/lanes): cut across every horizontal layer; each domain of data has its own app/business/storage APIs and conceptual DB.

Quotes: Brian Kernighan (C's missing firewalls); John Ousterhout (complexity accumulates; can't keep all factors in mind).

**Root goal:** code maintainable, manageable, debuggable **by somebody you don't know yet**.

## 1. Layered Architecture

| Layer | Responsibility | Must NOT |
|---|---|---|
| **API** (protocol) | Decode/encode, route bind, MUX, map app error codes → protocol codes | Business/app logic; validation |
| **App** | Validation, orchestration of business calls, app models/errors, auth, middleware, metrics | Protocol imports (`net/http`); business rules |
| **Business** | Core rules + core data model; defines storage interface it needs | Import app/API; use `errors` package (wrap only) |
| **Foundation** | Project "stdlib": logger, web, keystore, validation, config, debug, metrics | Logging, importing logger, importing upward |

- **vendor/** — own third-party for diffs, license, supply-chain.
- **Zarf/** — Docker/K8s/keys deployment config ("don't get burned by containers").

**Import rule:** *"Imports can only go down, they can't go up."* Foundation ← Business ← App ← API. Only two deliberate exceptions: `foundation/web` (is HTTP) and `app/.../auth_client` (must speak HTTP to auth).

Each horizontal layer repeats: **API (support) + domain + tests**.

## 2. Domains and Vertical Firewalls

- Domain = group of related data (user, product, home, sale) → vertical lane through API/app/business/storage.
- Conceptually **own database instance** even if one physical DB; isolate via table-name prefixes or Postgres schemas.
- Sense of microservices **without** microservice complexity; scale one service horizontally; build different binaries with route subsets.
- Cross-domain: business may call other business API if relationship permits; else **delegate/event system** (serial same-goroutine by default). App-to-app doesn't make sense.
- **"Layer, not group."**

## 3. Auth Placement

- **Auth is an app-layer concern, not business.** By business layer, authN/authZ already applied.
- Centralized **auth service**: generate token, authenticate, authorize. Sales never does OPA directly.
- Auth client (app) wraps auth over HTTP; exposes only authenticate/authorize.
- JWT (RFC 7519) + OPA (Rego). Never roll your own. RSA private/public; keyLookup interface.
- Disabling users with valid tokens: DB check that user still enabled if needed.
- `authorize user` middleware: parse user ID from URL, load user into context, OPA authorize — get/update/delete only (not create).

## 4. Middleware

- **Two halves:** app-layer middleware (no protocol) + protocol-layer middleware (map codes, respond).
- Names: `authenticate`/`authorize` = auth **service** calls; `bearer` = local JWT; `basic` = local basic-auth.
- Layout outer→inner; **executes right→left**.
- **Panics outermost.** Metrics counts goroutines/requests/errors/panics. Error middleware centralized.
- Variable `ATH` not `auth` (package shadow).

## 5. Web / Foundation Patterns

- **Steal the MUX:** embed `http.ServeMux` into `app`. Custom handler returns error; adapter to std signature.
- Logger: **function type** `func(ctx, msg, args...)` not interface. Foundation cannot import a logger.
- Shutdown channel **removed** (unused); log and return on response-write failure.
- Package = purpose API, not containment. First file named after package. No `common`/`utils`/`helpers`/`models`.
- **Every package has its own type system.** Concrete first; discover interfaces (Rob Pike). Interfaces as **input only, never return**. Prefer one-method interfaces.
- Polymorphism: code changes behavior based on concrete data it operates on.

## 6. Route Adder / Polymorphic MUX / Build

- Interface `route adder`: `add(mux, config)`.
- Build packages: `build/all`, `crud`, `reporting` → distinct service binaries with API subsets.
- Each route group owns `route.go` with routes binding.

## 7. Data Models & Storage

**Three model layers only — "You can't have one data model":**
1. App models (JSON tags OK — encoding not protocol)
2. Business models (core)
3. Storage models (`dbUser`, native DB types)

- Each transition requires **`parse`** (integrity mechanism).
- **No separate data layer.** migrate + sqldb under `business/API`; storage under each business domain.
- Business defines storage interface (device-like; larger OK).
- Domain values: struct with unexported field + parse, never `type Role string`.
- Data → value semantics; API → pointer semantics.
- **Transactions:** begin at start of request; put tx in context; substitute pool. Initiated at **app layer**. Interfaces: `beginner`, `commitRollbacker`. If no tx in context, **fail** — don't fall back to pool.
- Same DB instance required for cross-domain transactions.

## 8. Error Handling

> **"Once you handle it, it's not an error anymore."**

Handle = log once + inspect + respond consistently + stop propagating. Non-handlers wrap and send up. Centralize in middleware, never handlers. App defines error type + protocol-agnostic codes; API maps to HTTP. Unknown → 500, no leak. Business doesn't use `errors` package. Don't DRY error variables across packages.

## 9. Core Philosophies

1. **"Do not make things easy to do. Make things easy to understand."**
2. **"Every encapsulation must define a new semantic where one is absolutely precise."**
3. **"Engineer with clear and obvious layers of concern and purpose."**

- Programming mode → engineering mode (refactor to health).
- Uncertainty = stop and learn, not guess.
- Consistency wins the API game.
- Data-oriented design: everything is data transformation.
- **Trust but verify at compile time** via parse/type shaping — not re-validate every layer at runtime.

## 10. Kubernetes / Ops

- **Deploy-first.** Staging must look like production day one.
- Zarf + Kustomize (base + overlays). Kind for local images.
- Never hardcode image names. Only `main.go` imports config.
- Liveness cheap; readiness must exercise real DB queries (not just ping).
- Bounded load-shedding shutdown; sync with `terminationGracePeriodSeconds`.
- **Never bind default ServeMux** (supply-chain — any init can register).
- GOMAXPROCS = cores under K8s CPU limits. Go programs are CPU-bound; never more OS threads than cores.

## 11. Logging

- Stdlib slog. No logger singletons; no logger in context. Pass from main.
- Don't use levels to turn logging off — same signal in dev and prod; levels for **alerting**.
- Log errors once, at handler of the error. Messages lowercase.

## 12. Testing

- Business unit tests against **real DB** (Docker), not mocks.
- API integration tests through MUX.
- `must`-prefixed only in tests.

## 13. Concurrency

- No orphan goroutines (parent waits). Only accepted orphan: read-only debug/metrics server.
- Server.Shutdown with bounded timeout, then Close.

## 14. Condensed Rule Card

1. Three firewalls: package → horizontal layers → domain lanes.
2. API → App → Business → Foundation; imports down only.
3. Auth in App, centralized auth service (JWT+OPA).
4. Middleware two halves; panics outermost.
5. Own the MUX; logger function type; no default ServeMux.
6. Route adder + build packages for API subsets.
7. Storage under business domains; three model layers; parse for integrity.
8. Handle error once in middleware; never leak internals.
9. Consistency + precision + clarity over convenience.
10. K8s deploy-first, staging≈prod, GOMAXPROCS==cores.

## Key Quotes

- "Once you handle it, it's not an error anymore."
- "Do not make things easy to do. Make things easy to understand."
- "Imports can only go down, they can't go up."
- "We don't build packages that contain, we build packages that have purpose."
- "Don't design with interfaces, discover them."
- "Trust but verify" — at compile time.
- "You can't have one data model."
- "Auth is an app layer concern, not a business layer concern."
- "You cannot, in any code you're putting in production, bind the default server MUX to any port."
- "Somebody else, somebody you don't even know right now, should be able to come in, maintain, manage, and debug this code base."
