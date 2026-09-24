
# Microservices Masterclass — Dense Architecture Reference

Source: compiled transcript (Dave Farley / "Engineering Room" style episodes + interviews with
Randy Shoup, Gregor Hohpe, James Lewis, Spotify). 16 segments. This is permanent, reusable
system-design knowledge, not a summary of one talk.

---

## A. DRY vs Coupling — the core tradeoff

- DRY (Wikipedia/original): "Every piece of knowledge must have a single, unambiguous,
  authoritative representation within a system." Nearly always good advice — **but not always**.
- For bigger/complex systems, **coupling is an equally implacable enemy, at least as important
  as duplication.** The tradeoff for DRY is *increased coupling* — commonly the biggest
  stumbling block for teams adopting microservices.
- Duplication is bad because: you must find all copies to change behaviour; copies **drift
  apart** into slightly-different versions; systems become a pain to maintain.
- **Fractal rule of thumb:** within a module/class/file, accept no more than a couple of lines
  of duplication (DRY is essentially mandatory at fine grain). In the same repo, aim to be
  largely DRY. Across module boundaries, be cautious not to generalise too soon, yet not too
  tolerant of duplication. Finding the sweet spot "is the real skill of software development."
- At the **service** level: the interface/API should usually be DRY (services should be the one
  place to get a job done); the *implementation* is more problematic because of developmental
  coupling.
- Across **true microservices** (separate repos, independent deploys): "I would treat DRY with
  great suspicion and great care... I would much prefer duplication to reuse. The cost of
  coupling now are too high." Developmentally coupling services together to achieve reuse is
  "a big mistake that I see played out all the time."
- Bottom line: **"DRY is a useful guideline and a rotten rule."** Once systems/orgs get beyond
  small and simple, **coupling is the real enemy.**

## B. The platform / common-services trap

- Commonest form of the DRY-in-microservices mistake: code called **"platform" or "common
  services"** that several other services depend on.
- It begins trying to be "architecturally DRY" and ends forcing everyone into **lockstep**,
  because now changes to the platform break everyone.
- Diagnostic: **if your team is forced to take versions of shared code for any reason other
  than that it does something new you want, you are suffering from this problem.**
- Platform/common-services code "should take loose coupling and good abstraction more seriously
  than any other part of the system, but often they don't."

## C. Startup → growth → hyperscale evolution

- **Startup mode** (Randy Shoup): looking for product-market fit, must iterate fast; team small;
  problem bounds unknown; no bounded contexts yet. **Start with a monolith — 95–99% of teams.**
  You have not pre-decided subcomponents; can refactor the entire system; all calls local (no
  network); single artifact → easy roll out/rollback.
- The problems that microservices / event-driven architectures solve **don't exist yet** in a
  startup. "I think ninety percent of the software on the planet really should be done in a
  monolith."
- **Every big company started as a monolith**: eBay, Netflix, Twitter. It's an **S-curve** —
  only in the concave-down part do you need microservices.
- **Growth/scaling phase**: monolith problems appear → then leverage other architectures.
- Don't copy the hyperscalers: "it's a very small percentage of the industry that has the
  hyperscale problems." eBay today = Kubernetes + tons of repos, but it has **27 years and
  4,000 engineers** — "but you're not starting."
- Architecture *will* change as you grow; eBay changed its architecture **~5 times** along its
  trajectory — that is legitimate.
- Modular monoliths scale too: **Facebook, Shopify** (huge Ruby modular monolith; Shopify
  engineering blog documents making it much more modular after hitting issues).

## D. Definition of real microservices

Canonical list (microservices.io / James Lewis & Martin Fowler, 2014):
**small, focused on one task, aligned with a bounded context, autonomous, independently
deployable, loosely coupled.** It says **nothing about technology** — not REST, not HTTP, not
XML/JSON. (REST can be a microservice or not; not all services are microservices.)
- **Independently deployable is the defining characteristic.** All other attributes exist mainly
  to make this possible. It means you can deploy a service **without testing it alongside every
  other service before release.** "This is the decoupling step... table stakes."
- This enables organisation into many small **autonomous** teams; autonomy is one of the
  strongest predictors of success in the DORA/Accelerate findings; "smaller autonomous teams
  build better software faster" (Team Topologies: max ~8 people; Fred Brooks, 1970).
- Microservices **trade consistency for independence** — the most distributed approach to
  development.
- If you can't promote/deploy independently you have a **distributed monolith** (Beth Scurry:
  "If you can't promote microservices independently of other microservices, you don't have
  microservices. What you've got is a distributed monolith."). A distributed monolith isn't
  wrong — it just doesn't deliver the microservice benefits.
- Microservices are fundamentally a **socio-technical strategy** — more about team organisation
  and dynamics than about software architecture. A technical solution to a social problem (and
  vice versa).
- "Small" heuristic: it fits in James Lewis's head; can you throw it away and rewrite it in a
  few days / a couple of weeks? If rewriting it scares you, it's too big.
- "Most teams that claim to be practicing microservices aren't" — repeatedly asserted (Farley,
  Simon Brown, Shoup).

## E. Coupled modules — the worst-of-both-worlds anti-pattern

- Multi-repo where pieces are nonetheless **sufficiently coupled that you don't trust their
  interactions unless you test them all together**. "I think that this is the dumb one."
- Result: a monolithic system (you must test everything together) **with all the impedance and
  friction of a microservice system, and none of the benefits**. Not independently deployable
  **and** not independently developable.
- Loses monolith optimization techniques: **incremental builds**, **static typing** to catch
  mistakes earlier than CI, refactoring tools across shared code.
- "Couple modules is just a dumb approach... the worst of all worlds with none of the benefits.
  But it's incredibly common. In fact, most teams that claim to practice microservices are in
  reality doing coupled modules."
- Litmus test: **can you deploy your microservice without testing it with others first?**

## F. Repo strategy (monorepo vs multi-repo)

- What is a repo? **A useful working scope for our software.** Version control's three big
  advantages: (1) step back to safety when you make a mistake; (2) share changes with others;
  (3) a shared safe place / backup. "Version control is just a tool... not likely to be
  inherently evil." "Monorepo" shouldn't be a synonym for evil.
- Three models:
  1. **Monorepo** — everything in one repo. Records a definitive version of the whole system
     ("v1 of A works with v2 of B/C and v3 of D"); easy to share even complex coupled changes;
     build/test all together; IDE/CI/acceptance tests catch breakage; whole-system backup.
  2. **Multi-repo, decoupled = real microservices** — the most scalable way to build software,
     but at a price (below).
  3. **Multi-repo, coupled modules** — the dumb one (see E).
- In true multi-repo microservices, **nothing records the conversations between pieces, nor
  which versions work together**. There is no definitive statement "vX of A works with vY of B";
  versions change constantly; production is a moving target; "they never met until they met in
  production." Easy to reach states where you **cannot step back → roll forward** instead (giving
  up a VCS advantage). Sharing/coordinating coupled changes across repos is slower and harder.
- **Design rule: the scope of a repository (and of a deployment pipeline) is the independently
  deployable unit — i.e. the releaseability of the software.** A pipeline evaluates an
  independently deployable unit, easiest when that unit is in a single repo.
- Independently deployable units **need not be microservices** — a big subsystem or a library
  can sensibly have its own repo + pipeline. Monoliths are fine; real microservices are fine;
  coupled modules are "only very inefficient monoliths in disguise."
- "If the contents of your repo aren't independently deployable, I think you're almost certainly
  wasting time and effort."

## G. What architecture is / what it does

- **Grady Booch: "Software architecture is about the significant decisions."** Key tech choices
  you can't easily change later; the overriding modularity strategy (monolith / microservices /
  in-between); how to avoid the **blast-radius effect** (change here, everything blows up).
- Architecture is the **skeleton** of the overall system; a means to an end; there is **no one
  right architecture** — good at some scales, bad at others.
- "Architecture's nearly all about **management of complexity** — it lets us build systems
  beyond the scale we can hold in our heads," compartmentalized so each piece fits in a head.
- Means: modularity, cohesion, separation of concerns, good lines of abstraction, loose coupling.
- Defining quality in modern systems = **our ability to change it.** Good architecture supports
  iterative, fast delivery to production and makes the code "a habitable space that you can
  change." You don't get good design free by hacking.
- Architecture **will change** over time (legitimately, e.g. ~5 times at eBay). At every scale you
  are changing software, so optimise for **changeability**.
- **Architect = option trader:** which choices do you take away (to reduce complexity per
  Gregor's law), and what choices do you gain in return? Architects are the **flexibility vs
  complexity balancers**. Constraints are a tool that creates freedoms elsewhere.
- Some parts of a design matter more than others: **boundaries, modules, APIs** matter more than
  internal implementation detail — those important seams need architectural thinking.

## H. Evolutionary design

- Agile principle (IDO/agile manifesto): **"a continuous approach to good design enhances
  agility."** Teams that went from Big Design Up Front to no design discovered that's also bad.
- **Agility and architecture go together** ("agile is the steering wheel, architecture is the
  engine") — both deal with high rates of change and uncertainty.
- **Start with a model and assume it is wrong.** That is the step to engineering: work so that
  when you discover where it's wrong you can correct it. Contrast with Big Design Up Front, which
  assumed the blueprints were right and were the permanent target.
- **Gregor's law: "Excessive complexity is nature's punishment for organizations who are unable
  to make decisions."** Deferring decisions buys changeability but costs complexity.
- Example: standardising on **common APIs + JSON over HTTP** takes away choices but *enables*
  language/deployment freedom — at the cost of API layer, partial failures, retries, idempotency,
  out-of-sequence messages, and JSON null-vs-missing-vs-empty semantics.
- **Insurance example**: a really elegant calculation engine solved the *wrong* problem because
  the team didn't know the business variability point (business wanted the inverse: input premium
  → output coverage). You cannot make a good technical decision at a lower level without
  understanding the business strategy above you — the **architect elevator**. Otherwise you guess,
  and guess wrong repeatedly.

## I. Modular monolith as the default

- **Recommended starting point for 95%+ of teams: one repo, even one process** (modular monolith)
  — keep service-oriented design internally.
- Rationale: you don't yet know the right services/messages; breaking up too soon destroys the
  fast feedback you need to learn. In one repo/process the IDE catches ~90% of cross-service
  breakage instantly; refactoring crosses the whole system.
- **Farley's preferred path:** start with best-guess services interacting via ports & adapters
  and best-guess messaging; evaluate everything together in a shared pipeline; live with it until
  the messaging "doesn't change very often"; **only then pull services out into separate repos as
  microservices.** Use event storming to find the first guess.
- Risk to manage: a careless developer can ignore service boundaries inside the bigger-scope repo
  → add tests that reject cross-boundary code sharing.
- Anti-pattern flip side: a 10–15-year-old brittle Java "legacy mess" declared "we're converting to
  microservices" by putting JSON-over-HTTPS on the *existing bad modularity* → wrong boundaries,
  lockstep, brittle, fragile, slow. It needs a mindset shift, not network links.
- Modularity is always good, but you don't always need inter-process communication or multithreading
  everywhere — both "amp up complexity by an order of magnitude."
- **The distributed monolith is a legitimate, often preferable architecture** for some systems
  (e.g. in one repo, collection of services); it makes dependency management easier.

## J. Interface design for independence

- Two strategies make interfaces tolerant of different versions:
  1. **So stable they never need to change** (Amazon S3, TCP/IP).
  2. **Flexible enough to absorb change without breaking** (web browsers ↔ HTML).
- Interfaces must be **abstract, hide implementation detail, and allow replacement** with better
  versions. Align services with bounded contexts; **always translate at boundaries**; use
  **ports & adapters** at service edges to insulate the core and validate inputs.
- **Function calls are tightly coupled by design**, but *design* can decouple them (translation +
  validation + minimal abstract params). Conversely a REST API carrying implementation-coupled
  data with no translation is still tightly coupled. **The technology doesn't solve coupling — the
  design does.**
- **Eric Evans insight:** the language/protocol of the messages between services is itself a
  **separate bounded context** (more accurately, a *collection* clustered around different
  conversations) — so translate **to** and **from** the message. Aim for messages that are either
  **more stable** (change slower than the services) or **more independent** than the services.
- Keep interfaces **minimal**: expose the smallest thing consumers need; don't add "just in case";
  design actively for simplicity and likely-future sense.
- Data-format ladder (increasing decoupling): fixed binary offsets (2-byte ID broke at >1024
  customers) → class → interface → versioned contract supporting old + new → tagged/delimited
  forms (XML/JSON/YAML) where consumers depend on **tags not byte positions**, so fields can be
  added/reordered and old code keeps working (if not positional and it doesn't assume every tag).
- **Contract = externally visible functions + published data/events + operational characteristics**
  (moving toward reactive systems). If it's a genuine microservice, these are the only places that
  can break interactions — so this is the scope of contract testing.
- **Bounded context (Eric Evans):** a defined part of software where particular terms, definitions
  and rules apply consistently — the natural "fire breaks" in the problem domain. Aligning services
  to them makes the system naturally less coupled, with cleaner, less chatty interfaces.
- Model messages at the **level of the problem domain** (ubiquitous language) — e.g. "Order Book",
  "Dispatch Book", "Book Dispatched" carrying books and customer accounts, not SELECTs/records. A
  non-technical domain expert should broadly understand the conversations.
- On top of services + messages sits a **thin layer of core concepts / durable unique IDs (keys)**
  that glue contexts together for traceability — keep it intentionally simple.
- **Versioning techniques:** support old and new API versions in parallel; run old/new versions of
  a service; communicate deprecation; allow gradual migration; use beta programmes; never force
  lockstep change (you may not even know your consumers).

## K. Version pinning of shared libs

- If a shared library dependency is resolved as **"give me the latest,"** you are developmentally
  coupled again (my change forces yours).
- Fix: **pin explicit versions**, treating shared code as an external third-party dependency
  ("my service uses v7, yours v6"). Each team decides when to upgrade. "But let's be clear, we're
  no longer DRY" — VCS is used to allow two versions in use.
- Alternative: deploy the shared behaviour as its own microservice and design a loosely coupled
  interface around it — same design-over-technology caveat as in (J).

## L. TDD and contract testing as design feedback

- **TDD isn't really about testing — it's design feedback.** If a test is hard to write or
  understand, you have a coupling/abstraction problem.
- TDD is great at exposing **over-coupled dependencies**, drives better interface abstraction of
  dependencies, helps spot generality earlier, and keeps code DRY **without inventing crappy
  tactical abstractions**.
- **Contract testing isn't really about testing either** — it *specifies the dependencies between
  components* so they are both human-understandable and machine-checkable (compliance auto-discovered).
- **PACT + PACT broker:** the broker tracks every version of every component, its dependencies, and
  which environments each version is deployed in. You run a contract test between A v73 and B v212
  **once**; before promoting A into an environment you just look up the matrix to see whether a
  passing test exists against the version it depends on there. Scales to many components/dependencies
  → fast promotion safety.
- **Specmatic:** contract-to-contract compatibility testing with a shared contract repository,
  storing contracts as an IDL (typically **OpenAPI**), comparing versions for backward compatibility
  at build/commit time without writing code; can block non-backwards-compatible merges; consumers can
  propose new contract versions and work ahead without breaking anything. Zero-code approach for many
  cases (auto-generated contract tests + stubs). Depends on contracts being explicit — "no sneaky back
  doors, no sharing data via data stores."
- **CI/CD framing:** evaluate *exactly what you release* (the sequence of bytes), at least daily;
  the pipeline must be **definitive for release**. The correct scope of a deployment pipeline is an
  **independently deployable unit**. Two (and only two) real solutions to scaling teams:
  (i) treat it as a monolith — shared repo, CI, one fast pipeline evaluating everything together
  (works surprisingly well and surprisingly scalably); or (ii) make each unit independently
  deployable (microservices). The common big mistake is coupled modules.

## M. Observability and platform bootstrap at scale

- Observability is a **property of a system**, not a tool/technique; not binary. Definition: the
  ability to collect data about a program's execution, modules, internal states, and the
  communication between components. **Signal-to-noise matters — too much is nearly as bad as too
  little.** Measurement is what distinguishes engineering from craft.
- **Empirical learning:** some lessons can only be learned by observing production, not from
  pre-release tests.
- Three telemetry types:
  - **Metrics** — a measurement at a moment in time (CPU, trades/sec, users of a feature).
  - **Logs** — a sequence of events, usually human-readable; for post-failure cause analysis, bug
    recreation, predicting problems.
  - **Traces** — map interactions across a distributed system (breadcrumbs: which services were
    visited fulfilling a function).
- Distributed difficulties: **consistency is not a strength** of microservices (autonomous teams);
  tracing across services is hard; time zones/clocks ("did A happen before B?"); multiple copies of
  a service in an ever-changing cloud of relationships; async/event-based reordering.
- Patterns that help:
  - **Time synchronisation** — establish a common baseline; store all timestamps in **UTC**;
    synchronise clocks (NTP ~millisecond resolution; may be insufficient for microsecond ordering).
  - **Snapshots / health checks** — poll services for a pass/fail health status (Elmax financial
    exchange polled every ~30s), with **hierarchical health checks** that propagate to
    dependencies, allowing traversal of the service graph to build overall system health and report
    causes of failure.
  - **Log aggregation** — a central collector assembles a system-wide, ordered account (possible
    with PubSub infra or purpose-built tools).
  - **Correlation IDs** — model IDs and their relationships/hierarchy (Social Security Number
    analogy); every message must carry the relevant keys that contextualise the conversation so
    distributed interactions can be stitched together later.
- **Platform bootstrap for autonomy at scale:** big microservice-focused organisations give teams
  autonomy but fund internal **engineering/platform teams that bootstrap** product/service teams —
  pull a service template from an internal repo and get **observability, monitoring, and
  deployability free** into a standardised production environment.
- **Spotify example:** standardise technology stacks ("**Golden Technologies**"), target **85%
  adoption, deliberately not 100%** to preserve room for experimentation/innovation; standardise
  infrastructure, leave flexibility at the application level; **guide rails** over full autonomy;
  distinguish **accidental fragmentation** (individual preference/CV-padding) from **intentional
  fragmentation** (e.g. Rust for transcoding — good for the problem but unsupported by infra teams,
  losing all automation). Reasons for standardisation: automation, and org fluidity (move ownership
  / restructure teams without relearning stacks). Platform team builds an internal platform on top
  of the cloud; dedupe that large investment.

## N. Anti-patterns, failure modes, and quotes

Anti-patterns:
- **Coupled modules / distributed monolith built by accident** (E) — the most common; "the worst of
  all worlds."
- **Microservice soup:** "2,000 microservices and 10 developers... deploy them all at once every two
  years... that is why we never said do that." "You're not winning at anything if you've ended up
  there."
- **Shared database / shared data model** behind services — violates autonomy; means you can't change
  schema in one service without changing the other; data normalisation represents coupling; bounded
  context alignment exists precisely to avoid this. Prefer sharing only via messages (no back doors).
- **Order/customer modelling example:** the order service should store only the **account ID** (the
  most loosely coupled relationship). If it needs the email, ask the customer-details service at the
  point of need. To prevent orders for non-existent accounts, have the customer service **notify** the
  order service of new accounts, so the order service keeps its own duplicate valid-ID list — a
  deliberate duplication that **decouples** and lets each service be tested alone. Translate and
  abstract conversations at the **business-problem level**, not the technical-integration level.
- **Naive tactical solutions feel easier at the start but are much harder to maintain;** modelling the
  real problem takes more thought up front.
- **Semantic diffusion:** microservices has been devalued to mean "a small lump of code that
  communicates via XML over HTTP."
- **Network naivety:** Corba/DCOM failed for the same reason a generation earlier — people ignored the
  cost of remote calls (one loop called another service **byte-at-a-time over the network: ~1,000×
  slower**; each byte sent in its own ~1KB packet). At DoorDash (Matt Ranney) the **average fan-out
  from one request exceeded 1,000 service messages** — a performance and diagnosis nightmare.
- Once services are independently deployed, we still pay a **technical debt** — more sophisticated
  design, more effort/coordination to change multiple services or debug complex systems — but now it's
  the "better kind" because we gain autonomy. If they aren't truly independently deployable, it is
  debt with **no** benefit.

Notable quotes / facts:
- "DRY is a useful guideline and a rotten rule."
- "Once systems and organisations get beyond the small and simple, coupling is the real enemy."
- "You need to be this tall to take the microservices ride." (Martin Fowler, 2014)
- "If you can't promote microservices independently... you've got a distributed monolith." (Beth Scurry)
- "Microservices are an inorganizational decoupling play... If you don't need to decouple your
  organization developmentally, they're probably not the right thing to do."
- **Conway's law** (Mervyn Conway, 1967): an organisation designs a system whose structure copies the
  organisation's communication structure. Organise teams by technical boundary → tiered/ layered
  systems; modular autonomous teams → modular software.
- **Greenspun's 10th law (paraphrased):** any sufficiently complicated microservices implementation
  contains a half-arsed, bug-ridden implementation of half of Erlang.
- Considered "hard on the outside, soft in the middle": hard interfaces defined by standard protocols
  (e.g. W3C web standards), soft autonomy inside teams; fractal all the way from methods → classes →
  libraries → services → collaborative-service groups → business capabilities.
- **Business capability ≠ bounded context:** business capabilities are large (e.g. warehouse
  management includes people, processes and tools, only some of which is software). Services organised
  around a business capability = a *set* of services for some of those tools.
- No mainstream programming language/OS surfaces the "service" concept; every service system requires
  you to invent your own conventions/protocols for the service edge.
- Microservices' origins (James Lewis): convergent ideas (Stan North's replaceable-component
  architectures; Fred George's tiny services; Adrian Cockcroft's fine-grained SOA at Netflix; Alan
  Kay's message-passing/encapsulation); cloud + automated environments made pushing complexity into
  infrastructure worthwhile. "XP is doing all the things and turning it up to 11. I think
  microservices is similar."
- James Lewis's term that never caught on: "organizational onion" (fractal org/service structure).
- Matt Wynne's satire "Mortgage Driven Development" — choosing tech because it looks good on a CV.

## Cross-cutting principle (the spine of the whole masterclass)

**Boundaries are decisions about coupling, and coupling is the primary economic/architectural force.**
Choose your scope of *version control* and *evaluation* to match the **independently deployable
unit**; keep interfaces stable-or-flexible and translated; prefer duplication over coupling between
services; start with a modular monolith and evolve toward microservices only when the S-curve and
organisation scale demand it; treat observability and platform bootstrap as first-class rather than
afterthoughts.
