
# System Design Master Class — Ad Hoc Systems Week (instructor Sneha Mehra / Arpit-style)

Source transcript: course day covering Distributed Task Scheduler (main), Message Broker on RDBMS, YouTube View Counter, Flash Sale. Theme: "break problem into phases; solve one thing well then extend."

---
## 1. DISTRIBUTED TASK SCHEDULER (deep dive)

### A. Problem statement + SLA
- Schedule any task; execute at a given time. Examples: "run at 3pm every Sunday" (recurring), "run once on 20 Mar 8am" (one-time).
- SLA: **start executing within 30 seconds** of scheduled minute. Cannot guarantee completion (a task may run an hour). So guarantee = **start**, not finish.
- Scheduling granularity = **minute level** (no "10:01:32"); fair assumption, no need for second precision.
- Two task kinds: **one-time fixed-time execution** and **recurring**. Recurring = "one-time task executed multiple times."
- Real-world parallels: AWS CloudWatch Events; open-source **D2/D-Cron** (instructor read its codebase to learn how SLA is met); Celery.

### B. Phased breakdown (Store → Pick → Execute)
1. **Store** task in a relational DB (choose RDBMS deliberately: "no need of over-engineering"; any store giving same guarantees works).
2. **Pick** eligible tasks.
3. **Execute** them.
Deliberately solve **one-time execution first**, then extend to recurring.

### C. Schema / data model + WHY
Initial bare-minimum: `id`, `command` (all execution details, JSON/shell), `scheduled_at` (store **Unix epoch UTC**, not local time).
Added during implementation: `picked_at`, `started_at`, `completed_at`, `failed_at`, `status`, `type` (task type for routing/scaling).
- **Why Unix epoch UTC**: avoids timezone ambiguity.
- **status column is derivable** from the timestamps; instructor seeds the "don't store derivable data" principle (analogy: not storing S3 file path when it can be derived). Kept but warned it can go **inconsistent** (crash while `in_progress` leaves a stuck status). Either keeping or removing is acceptable.
- **No `frequency` column** for one-time only; multi-tenancy / frequency / auth are **"vitamins, not painkillers"** — add later.
- **Separate `completed_at` and `failed_at`** because failures **fail fast** (runtime exception in ~2s vs 1hr success); combining would skew the average used to detect failure. Use historical avg completion × **1.5** as timeout to declare failure.
- Use **weighted average** (recent tasks weighted higher) when per-type duration changes (e.g. task went 5min → 10min after added business logic).

### D. Components and why split
- **Task API** — accepts schedule requests, writes to DB; keeps user stats/UI. Deliberately NOT doing picking (concern: single responsibility, distinct scaling axis: HTTP requests vs number of tasks).
- **Pickers (task pullers)** — continuously fire transactional query, pull batch (limit 10), push to queue, update status. Separate from executors because:
  1. **DB connection limit**: if each of N tasks had its own machine picking+executing, need N concurrent DB connections → DB dies. Batching to 10/50/100 reduces connections.
  2. **Specialized executors**: executors optimized for GPU / high-RAM / high-disk / Java / Golang; a generic executor becomes "bloated." picker routes to per-type queues/topics.
- **Message queue** between them: SQS / RabbitMQ / Kafka.
- **Executors/workers** — pull from queue, execute, update DB. Homogeneous set per queue.
- **Orchestrator** — autoscales pickers and executors.

### E. Concurrency pattern
Query evolution:
1. `SELECT * FROM tasks WHERE status='scheduled'` — wrong: matches tasks a year away.
2. Add `scheduled_at < now() + 30s`.
3. Add `ORDER BY scheduled_at LIMIT 1` — predictable, but one-at-a-time too slow and no batching.
4. **`FOR UPDATE`** — exclusive row lock stops multiple pickers returning same row; but others **wait**.
5. **`FOR UPDATE SKIP LOCKED`** — locked rows are skipped; other transactions proceed → high throughput. (Same construct taught week 2 in airline/seat-checking system.)
Final picker query: `SELECT * FROM tasks WHERE scheduled_at < now()+30s AND picked_at IS NULL ORDER BY scheduled_at LIMIT 10 FOR UPDATE SKIP LOCKED;` then update `status` and `picked_at`, push to queue.
- Batch window `now()+X` is **configurable**.

### F. Observability / measuring SLA
- SLA check: **`started_at - scheduled_at < 30s`**.
- Three measured phases (store `picked_at`, `started_at`):
  1. **Pick time** = picked_at − scheduled_at (time to read + enqueue).
  2. **Queue wait** = started_at − picked_at (time waiting in queue).
  3. **Execution** = completed_at − started_at.
- Measure actuals (e.g. 5ms to pick 10 tasks), never guess. First measure, then act.
- **Don't scale on vanity metrics** (queue length alone). Scale on **task type**, count of each type, and avg duration per type; use a `type` column to group.
- No "worker ID" needed — you don't care which worker, only when execution started/ended.
- Executor scaling depends on types (10 one-hour tasks ≠ 10 one-second tasks).

### G. Failure modes & mitigations
- **Task breach**: analyze per-phase timings to find bottleneck.
- **Abrupt failure** (machine crash): completed_at/failed_at stays null → orchestrator declares failed after timeout (avg×1.5).
- **DB failure**: passive replica / standby, leader election (AWS RDS multi-AZ). One replica per shard.
- **SQS/broker failure**: rely on cloud, but add **dead letter queue (DLQ)** as fallback; consumers read DLQ too.
- **Orchestrator failure**: master-worker orchestrators with **leader election** ("best use case of leader election is auto-recovery"). Who monitors the monitor? Leader election.
- **Kafka parallelism limit**: parallelism = #partitions, can't have too many/few, rebalancing is expensive. Solution: **Kafka + SQS fan-out** — consumers read Kafka partition and fan out (round-robin) to on-demand SQS queues → high fan-out. Common for massive notifications (MoEngage). Cost: on-demand wins for micro tasks; constant infra wins for heavy tasks.
- **Guarantee no concurrent duplicate execution** of same job: very hard; most systems don't; handle on user side.

### H. Extension patterns (one-time → recurring)
- Two tables: **task** (the schedule) and **job** (single execution: id, task_id, scheduled_at, picked_at, started_at, fixed_at...).
- Picker sees recurring task → computes **next N (e.g. 10) occurrences** → upserts future job rows. Sliding-window: each pick computes next N and inserts only missing rows (unique key on timestamp / simple **upsert**).
- **Why N>1, not 1**: if a picker is down or SLA is breached, a single-next-entry scheme loses executions (state change not guaranteed to happen before next frequency). Buffer preserves the guarantee "if scheduled N times, it runs N times." N is configurable; **N=1** is correct if the desired guarantee is "stop after a failure." 
- Google Calendar analogy: not infinite rows — creates ~3 months of entries, extends as time moves; uses **range queries + copy-on-write/override** for exceptions.
- Other vitamins: logs (Fluentd → Elasticsearch → S3, ELK/EFK), user notification (async result callback), priority queues (P0/P1/P2 separate infra).
- **Sharding**: only when workload warrants ("shard can become a club"). Assign shards to pickers via consistent hashing / range / hash partitioning.
- Resource isolation on workers: **cgroups / Docker**.
- Assumption: a task runs on a single node; a Spark job becomes an API call waiting for completion.

### I. Patterns/principles taught
Get a "brain" (orchestrator) to scale up **and down** (cost). Look at next 5 min forecast to spin infra in time. Single responsibility per layer. "Recurring is just fixed-time scheduling solved well + add-on."

### J. Anti-patterns / wrong turns
- Storing derivable `status` (debated; consistency risk).
- Adding frequency/multi-tenancy/auth early.
- Merging picker+executor; **"I felt like it" without a rational reason** (intuition must be backed).
- Passing API server also do the picking (couples scaling policies, complex codebase).
- Scaling on queue length (vanity metric).
- Treating schema design as immutable / phases as exclusive; designing whole schema in one shot.

---
## 2. MESSAGE BROKER ON RELATIONAL DB

### A. Problem + constraints
Build an SQS-like broker: **FIFO**, consumer reads and deletes, **receipt handle** on delete, **high throughput**. Framing: this is the **first half of the distributed task scheduler**.

### B/C. Data model + WHY
Messages table in MySQL: `id`, `message`, `created_at`, `picked_at`, `receipt_handle` (and delete state). Index on `created_at`; **unique index on `receipt_handle`**.
Consumer query: `SELECT * FROM messages WHERE picked_at IS NULL AND deleted_at IS NULL ORDER BY created_at LIMIT n FOR UPDATE SKIP LOCKED;` then `UPDATE messages SET picked_at=now(), receipt_handle=<UUID> WHERE id=...`.
- **ORDER BY created_at** = FIFO.
- **SKIP LOCKED** = one consumer reading one message blocks others only on that row.

### D/E. Components & concurrency
- HTTP APIs: put / get / delete. Multiple parallel consumers. Exclusive lock on read (that's why `picked_at` set).
- **Never expose delete-by-ID**; delete by **receipt handle** (random string/OTP-like). Ensures only the consumer that read the message can delete it; stateless alternative to tracking who-read-what. On resurface a **new receipt handle** is issued, so old consumer can no longer delete. Requires **idempotent consumers**.

### F. Visibility timeout
Cron job on a separate machine: `UPDATE messages SET picked_at=NULL WHERE picked_at < now() - 10min`. Message **reappears at the front** because of `ORDER BY created_at`. Alternative accepted: add OR clause (`picked_at IS NULL OR picked_at < now()-timeout`) to the select — but someone still must null `picked_at`.

### G/H/I. Underlying-storage properties & extension
Properties required of ANY storage layer: ordered access by created_at + exclusive lock. Constructs differ: mutex/semaphore/atomic ops. Can build on SQLite/LevelDB/RocksDB/MySQL. **Celery supports RDBMS as a broker** and does exactly this. "Message-queue-as-database" (Martin Kleppmann) referenced. Lesson: identify patterns, map unknown → known (relational world), first-principles.

---
## 3. YOUTUBE VIEW COUNTER

### A. Problem + SLA
Count video views; a **rule engine** filters fake/spammy views (self-views, duplicate view within 5 min, under-age watching age-restricted). Views = money (ad revenue/payout), so filtering matters.

### B/C/D/E. Design
- Views stream into **Kafka** (high write throughput). Message = video_id, user_id, IP, etc. (everything rule engine needs).
- Kafka **consumer reads batches**; passes batch to rule engine; gets back genuine views.
- **Partition Kafka by video_id** so a consumer's batch is for one video → can `count += n` (batch writes). Micro `count++` writes would "die a brutal death."
- If not partitioned by video, insert an **adapter pattern** (from social-network week).

### I. The high-level (counterintuitive) pattern — reusability
After expensive rule-engine filtering, **re-ingest the valid views into a second Kafka topic** rather than only updating the DB. Downstream systems (analytics, search, payment/payout) consume the pre-filtered stream instead of re-running the expensive rule engine. **"Storage is cheap"** — persist expensive computations so others reuse them; this is **day-zero architecture for adtech** (e.g. InMobi). Makes cross-team/infrastructure impact. Separate consumers update the counts DB.
- Rule: wherever a rule engine consumes from a message stream, re-ingest output into another stream instead of directly updating the use case.

### J. Anti-patterns
- Directly doing count++ per message (micro writes); not partitioning by video ID; forgetting rule-engine recomputation cost.

---
## 4. FLASH SALE

### A. Problem
Limited inventory (e.g. 100,000 MI8 phones), fixed start time, everyone rushes; only inventory many can buy. **Spoiler: NOT a distributed-transaction problem** — it's a **high-throughput + contention problem** (contention ⇒ locking). Spoiler: it mirrors real-world human behavior.

### B. Phased breakdown
1. Pre-create/stock inventory rows ("fill the store").
2. Turn sale on (config flag `is_flash_sale_started=true`) → buy button appears.
3. Users "grab" an item (add to cart) → block/lock the item.
4. Payment flow (normal, separate from flash sale).
5. Expiry cron releases abandoned items.

### C. Schema / data model + WHY
- Products table: product_id, name, total inventory.
- **Units/inventory table — one row per physical unit** (e.g. 10,000 rows): `id`, `item_id` (FK), `picked_at`, `picked_by` (user_id), `purchased_at`, `purchased_by`. "Stocking the store with physical units = creating 10,000 entries."
- **Why one row per unit, not aggregated count**: need **per-item timeout** (users pick at different times → different expiry). Aggregation can't track per-item pick time/owner. If storing expiry, store only `picked_at` (expiry = picked_at + 10 min, don't store derivable). `picked_by` and `picked_at` nullable.

### D/E. Components & concurrency pattern
Grab query: `SELECT * FROM units WHERE item_id=720 AND picked_at IS NULL ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED;` then `UPDATE units SET picked_at=now(), picked_by=<user> WHERE id=...`.
- Exclusive lock = **"hand on the phone"** in the real-world store; other shoppers skip to another unit. SKIP LOCKED keeps throughput.
- **Reduce/block count when picking (add-to-cart), NOT after payment.** Otherwise 50k people reach payment for 10k inventory → refunds. User gets 10–15 min to pay.
- **Payment is NOT part of the flash-sale transaction** — contention ends when item is picked. Payment is regular microservice/gateway flow; use **webhooks** (Razorpay etc. with payment ID) to set purchased_at/purchased_by on success; on failure reset picked_at/picked_by = NULL after n failures.

### F. Expiry / observability
Cron "bouncer": `UPDATE units SET picked_at=NULL, picked_by=NULL WHERE picked_at < now() - 12min`. Continuously releases abandoned carts.

### G. Failure modes / product tradeoff
- Oversell + refund some customers, OR undersell. Product decision, usually prefer undersell.
- Rate limiter = the **door** at front-end proxy (requests never reach backend DB). Out-of-capacity → "cute" wait messages. Flash sales are separate systems with their own heavily-provisioned DB/own infra.
- Frequent small flash sales > one huge (load + brand awareness).

### H/I. Extensions & analogies
Same pattern: **BookMyShow, IRCTC Tatkal, hotel/travel booking** (fixed inventory, contention, temporary lock, timeout release). Archival: keep only ~1 week of unit rows, archive rest to S3 (for IRCTC partition by train). Alternative implementation: queue-based (DynamoDB streams / SQS) — put N tokens in a queue, "whoever draws a ticket gets in"; and since message broker = relational DB, the same works. "There is no one way to build a system."

### J. Anti-patterns
- Treating it as a distributed transaction.
- Reducing count after payment (oversell).
- Storing aggregated count (loses per-item timeout).
- No rate limiter / letting requests reach DB.

---
## CROSS-CUTTING DESIGN PRINCIPLES (apply to all systems)
1. **Break the problem into phases** (store → pick → execute). Don't get overwhelmed; a scary problem is usually simple when decomposed.
2. **Solve one thing really well, then extend** (one-time → recurring). Recurring = fixed-time base + add-on.
3. **Don't store derivable data** (status, S3 path, expiry timestamp).
4. **Vitamins vs painkillers**: don't add non-essential features (multi-tenancy, priority, logging) before the core works.
5. **Backed intuition**: every design choice needs a rational reason; "I felt like it" is unacceptable in a design doc.
6. **Separate concerns / clear responsibility per layer** — decouple producers from executors; avoid coupling two scaling axes in one service.
7. **Measure before claiming SLA** — observability is step one; instrument each phase.
8. **Scale on real signals, not vanity metrics** (task types, durations, unit economics — not just queue length); scale up AND down for cost.
9. **Reuse expensive computations** — re-ingest filtered streams; storage is cheap; cross-team impact (view counter pattern).
10. **Map the unknown system onto a known system** (usually relational DB), extract the properties you need, then swap the storage layer. This is first-principles thinking.
11. **No silver bullet**: Kafka and consistent hashing are not universal; every tool has strengths and limits. Read between the lines for extensibility.
12. **Patterns repeat**: DTS ↔ message broker (read/lock/delete/resurface); flash sale ↔ DTS (FOR UPDATE SKIP LOCKED); flash sale ↔ booking systems; CQRS-ish pick/execute.
13. **Leader election = auto-recovery** ("who monitors the monitor").
14. **Sharding for load** (small data, many queries) is different from sharding for data size.
15. **Real-world analogies** (human behavior in a store/park) reliably produce correct distributed design.
16. **Cost awareness**: on-demand (SQS) vs constant infra depends on task heaviness; prefer managed services.
17. **Guarantees drive config**: buffer size N and whether to continue after failure depend on the promised guarantee.
18. **Clarify assumptions explicitly** (minute granularity, single-node tasks, no retries).
19. **Use correct terminology** — knowing the term (e.g. "SKIP LOCKED") matters even if you know the idea.
20. **No solution in tech is magical**; ask critical questions of every tool.
