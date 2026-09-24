
# System Design Master Class 2 — Week 3 Day 1: Distributed Systems Foundations

Source: `C:\Users\sharm\.local\share\gencode\tool-output\tool_0d0e68e71001ZR5M0i36w0ITcd` (lines 1–1048 = Week 3 Day 1; file also compiles 5 other sessions — see section H).

## A. Philosophy of distributed systems
- A distributed system abstracts away its own distributedness: one coherent system to the outside, many machines inside. "Outside" is contextual (user, microservice, portal).
- Love/hate duality: you love them when you understand them and hate them when you understand them.
- **Murphy's law** is the operating principle: anything that can go wrong will. If you don't handle a failure scenario, you don't guarantee correctness → inconsistent data / corruption.
- **Day-zero architecture:** start with one server, one DB, one cache. Scale each component independently, observe under load, rectify, re-architect, repeat (microscope-style zoom from boxes to code).
- **Horizontal scalability** must hold at EVERY layer (DB, cache, stateful/stateless services, API, LB, CDN). Non-scalable layer = single point of failure.
- **Unit economics:** know load per server to get linear amplification; otherwise you under/over-provision and leak money. Every LB decision traces to money/utilization.

## B. Load balancing algorithms
- **Round Robin (RR):** default. index=(index+1) mod N. Use for uniform infrastructure + uniform, low-variance load (<1s responses).
- **Weighted RR (WRR):** weights per machine (4:8:4 → 1:2:1). Use for heterogeneous hardware (e.g., mismatched reserved instances) to avoid under-utilizing big boxes. Config exposed by cloud LBs.
- **Least Connections (LC):** track active connections per backend, pick fewest. Open connection ≈ busy server. Use for high-variance workloads (analytics: 1s to hours; RR would pile long queries on one box). Redshift/BigQuery use this.
- Reject CPU/memory-based routing: LB must decide in **microseconds**; querying resources adds milliseconds. LB stays light/stateless; connection count is the cheap proxy for busyness.
- LC ties resolved randomly; local optimization is fine.

## C. LB low-level + high-level design
- Terms: **LB server** (the load balancer process) vs **backend server** (black box).
- LB = normal TCP server on a static IP:port; holds in-memory list/map of backend `ip:port`.
- **Dual TCP connections** per request: client↔LB and LB↔backend. LB keeps an in-memory **source-conn → target-conn map**.
- Forwarding is **protocol-agnostic byte copy** (read one socket, write the other); handles HTTP/WS/RTMP/gRPC. TCP only closes when client/server/network terminates it (keep-alive, gRPC).
- **Single-LB limit:** 2 TCP conns per request → 10k conn limit serves only ~5k clients → LB tier must scale.
- Four brainstorm factors: **config, monitoring, availability, extensibility.**
  - Config: DB is source of truth; pull into in-memory copy at startup (never query DB per request). Staleness is the new problem.
  - **Orchestrator** does health checks via a configured endpoint (`/health`); must check the service process, not the VM. Orchestrator polls backends (backends don't know orchestrator IP). On failure it updates DB.
  - Config freshness: polling is wrong (staleness window → error spike). Use **reactive push**: second-port HTTP "pull now" endpoint, or **real-time pub/sub (Redis) + CDC**. Goal near-zero staleness.
  - Monitoring: orchestrator watches LB servers too; needs historical metrics → use **Prometheus** (telemetry agents on all VMs), like an autoscaling group. Build generic monitoring, not LB-specific.
  - Scaling LB tier: **DNS**, never client-held IP list. Domain abstracts multiple IPs; DNS does round-robin/WRR across LB IPs. **DNS is not a proxy** — client resolves, caches (TTL), connects directly. Private DNS exists (**CoreDNS**); real flow: public DNS → API Gateway → private CoreDNS → LB. DNS is a tiny single point with hot standby.
  - Orchestrator HA: multiple nodes + **leader election**; stateless (state in DB/Prometheus), 1–2s failover. Recursion bottoms out at a minimal monitoring service with leader election.
- LB = reverse proxy over a **homogeneous** set; API Gateway ≠ LB. Every service is a 3–10ms network hop; don't add boxes. Auth is backend business logic.
- Session handling: use **consistent hashing at LB**, not a separate session service. Consistent hashing = in-memory array.
- Papers: **Maglev**, **Chubby**.

## D. Source-level insights (toy Go LB)
- `lb.go`: `LB.run` listens on 9090; `listener.accept` loop; wraps client conn as source conn; `proxy()` picks backend via `strategy.getNextBackend()`, `dial`s backend, then `io.Copy` both directions in **goroutines** (copy is blocking; ends on EOF).
- `strategy.go`: RR = index+1 mod N, plus static, hash/consistent-hash strategies.
- Track requests-per-backend as a statistic; connection counts for LC updated on connect/terminate.
- Toy was 5–6× "faster" than Nginx/HAProxy, but real proxies do far more; lesson from reading Nginx source.
- "Difficult to scale, not to build." Build everything locally to learn edge cases.

## E. Remote / distributed locking
- Need: coordinate multiple processes/machines to preserve data correctness (e.g., two distributed txns updating one row).
- **Nearest-shared-storage principle:** threads→RAM (mutex/semaphore/atomics); processes→disk (`apt-get` `dpkg.log` lock file); machines→**network** (remote lock manager).
- Lock manager: tracks which resource is owned by whom (key Q7 → consumer). Motivating example: hypothetical unprotected remote queue; acquire lock → read → release. Real-world: MongoDB distributed txn across shards uses an internal lock manager.
- **Core properties of the lock-manager store: key-value access + atomic instructions + TTL/auto-expiry.** Redis has all three → **Redlock** (but any store with these properties works; MySQL + cron TTL is fine).
- Acquire: `SET key val NX EX 300` (atomic); consumer ID; while-loop retry. Naive release `DEL` can delete a lock you don't own.
- Failure modes:
  1. Holder dies before delete → perpetual lock → fixed by TTL.
  2. TTL expires mid-operation; B acquires; A deletes B's lock → add ownership check before delete.
  3. Check-then-delete race (TTL expires between commands) → make atomic with a **Lua script via `EVAL`**.
  4. Single Redis = SPOF (whole system stalls).
- **Distributed lock (quorum):** N independent Redis nodes (e.g., 5); acquire on a **majority (>50%)**; else release partial locks. Mutual exclusion holds while a majority is alive. Costs throughput → used only where high correctness is required; most startups never need it (cloud abstracts it). Residual edge case: majority holder's nodes die before it completes; another gets a majority → two owners. Minimized, never fully eliminated. Simulate with artificial delays.
- Alternatives: **Zookeeper** (persistent leader; notification = real-time pub/sub over persistent TCP — same properties as Redis PubSub); choose by required properties + team stack comfort. **Fencing tokens were NOT named** in this session.
- Creative network-layer lock: a TCP server accepting only one connection at a time; dead consumer's connection drops automatically, no TTL needed.
- If resource is single-machine, use local mutex — no Redis.
- Locking reduces throughput; justified where **correctness > throughput**.
- References: Martin Kleppmann's Redlock critique; Chubby.

## F. Example systems used
Online/offline indicator (heartbeat/TTL/pull-vs-push, Redis vs DynamoDB); airline check-in (single-process row locking ↔ distributed lock manager); the hypothetical Q7 remote queue (3-consumer mutual exclusion); `apt-get` (disk lock); threads/processes/machines ladder; MongoDB sharded distributed txn; analytics queries (LC); reserved-instance blunder (WRR).

## G. Cross-cutting principles & anti-patterns
Principles: day-zero iterate; assume every failure; all-layers horizontal; unit economics; nearest shared storage; design for core properties not tech affinity; reactive not polling; minimize network hops; microsecond LB decisions; reuse generic infra (Prometheus/DNS/pubsub); HA via leader election at every tier; build vs scale; read papers + source; everything is a trade-off; simulate edge conditions.
Anti-patterns: "this can't happen"; only-remote-machines "distributed"; LB querying DB or CPU/mem; config polling; DNS-as-proxy; random microservices/boxes; tech tribalism; perpetual/non-atomic locks; distributed locks in high-throughput systems; overcomplicated day zero; trusting toy benchmarks.

## H. Other sessions in the same file (context)
- **Week 5 Day 1 — Distributed cache + tiered storage/ETL/CDC:** CDC (Airbyte/Debezium) as robust DB→analytics sync vs API events; single-node cache internals (malloc wrapper/zmalloc, totalSize vs maxmemory); eviction LRU/LFU/random (Redis LRU bits; random = zero overhead, trades optimality; LRU for recency e.g. news/trends, LFU for past-frequency e.g. stock analogy/scan queries); TTL lazy deletion + sample 20 keys until <25% expired (CLT; ~5M keys; Redis single-threaded optimizes speed over completeness); consistent hashing = ownership only, not a service, not data movement (simple array + binary search).
- **Week 2 Day 1 — Relational DB & locking:** pessimistic (shared/read vs exclusive/write), deadlock detection kills the cycle-forming txn (DB never deadlocks), optimistic locking/CAS, single-threadedness, correctness>throughput, transactions/isolation, soft deletes, MySQL source (signals), indexes.
- **Intro + online/offline + approach to system design + connection pooling (~2986–3865):** heartbeat TTL 30s; Redis vs DynamoDB (managed vs self-host, persistence, advanced structures, persistent connections, expiry hooks, team size); TCP connection cost = 3-way setup + 2-way teardown = 5 trips; connection pool pre-establishes connections; DB max-connection limits from RAM/NIC (cloud formulas); idle TTL; scale risk at spikes.
- **Week 4 Day 1 — Social networks (~3868–4873):** tech-stack waves (2008–2012 Facebook PHP/MySQL; post-2012 Instagram Django/Celery/queues/Redis/AWS, 14M users 3 engineers); CDN usage; hashtag service; Redis PubSub per-hashtag/group = overkill + data loss (no persistence).
- **Week 6 Day 1 — Cost-efficient order storage + S3 (~4874–end):** tiered storage for performance at low cost; S3 partition manager (master + workers + leader election), partition servers, S3 API servers; S3 does NOT use hash/consistent hashing (cedes placement control) → range-based routing/partitioning for tenant isolation; hot-partition handling; elastic storage.
