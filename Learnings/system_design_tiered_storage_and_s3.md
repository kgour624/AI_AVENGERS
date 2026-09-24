
Source: course transcript file `tool_0d0e6ca97001I6hpuspNTUU9gE` (Genspark tool-output). Primary session = "System Design Master Class 3 — Week 6 Day 1: Cost-efficient order storage + S3 design". Speaker: Sneha Mehra. The file is 4 concatenated sessions; sections I summarizes the other three.

# A. Multi-tiered storage philosophy
- Every object/row has a DATA LIFECYCLE: accessed heavily when fresh, then access decays (orders, tweets, messages, posts). Access pattern is predictable for ~99.99% of cases; anomalies (a post going viral years later) are exceptions — design for normal customers, handle extremes separately.
- Root problem: huge tables => huge indexes => indexes no longer fit in RAM => disk spills => DB degrades. Multi-tiering keeps each tier's data (and indexes) small enough to live in memory.
- Nomenclature hot/warm/cold (temperature metaphor) is soft, not rigid.
  - HOT: main transactional DB (MySQL/Postgres/DynamoDB). Low latency, strongly consistent, end-user reads+writes. Expensive at scale.
  - WARM: read-only, non-transactional, horizontally scalable/distributed, supports frequent reads (e.g. last 6 months of orders). Slower than hot. Data is denormalized/aggregated into one JSON per order (order+payment+logistics) so one API call serves everything.
  - COLD: very infrequent reads (once a year: accounting/legal/lawsuits). Blob storage = S3; within S3 even the Glacier tier (cheapest, slowest). Read-only.
- Why archive: reduces hot DB size → smaller indexes → fit in RAM → performance; and drastically lowers storage cost.
- When to archive: driven by a per-table DATA DELETION POLICY (e.g. Snapchat messages 2 days; orders ~6 months / sometimes 1 month). Policy is business-driven and rarely changes.
- Determine lifecycle from metrics (access logging) or, failing that, a reasoned guess.

# B. Order system architecture end-to-end
Components and why each exists:
- Order service + transactional DB = hot tier. Many internal services (payment, logistics, customer support) read it → high read traffic.
- Scaling sequence for the hot DB: (1) vertical scale, (2) read replicas to offload reads, (3) shard the master (mutually exclusive subsets). But note: once read replicas + sharding are reached, ask the root-cause question — "do I need all this data in the hot DB?" → tiering is the cheaper answer.
- Dumper: periodic process that reads data from hot DB (and from payment/logistics DBs) into a STAGING store (S3/blob). Named "dumper" generically; "cron" here just means "a repetitive job", not necessarily a literal cron — the loader can also delete from hot after confirming warm copy exists.
- Staging storage: raw landing zone; can hold raw / processed / semi-processed / optimized layers. Multiple systems' tables land here as-is. (Staging can even be treated as a cold tier.)
- Loader (a Spark job) = ETL: reads the three files (orders, payments, logistics), joins/aggregates by order ID, writes the merged JSON into warm storage. Classic ETL (extract-transform-load); Spark abstracts distributed compute (like MapReduce but in-memory). No data loss: every column of every table is copied; flattening/denormalization happens downstream.
- Warm store: read-only, queryable as-is; given an order ID returns everything in one shot.
- Cold path: same dumper→staging→loader pattern to S3.
- Analytics path (OLAP, e.g. Redshift/BigQuery): a SEPARATE pipeline fed directly from the hot DB via CDC (not waiting 6 months). OLAP lets you query fast; pre-join/pre-aggregate before storing. If real-time analytics is required on hot data, add a real-time flush path.
- Data lake: built on S3 (HDFS-compliant), format = Apache Hudi ("Hoody" in transcript). Query via PrestoDB / AWS Athena / Hive, powered by Spark.
- Policy cache caveat: if deletion policy changes, services may fetch updated policy — but policy changes are infrequent, so a redeploy/restart is fine; don't over-engineer.

# C. Hot vs warm routing via ID (timestamp-in-ID)
- Service often only has an order ID; must decide hot vs warm without querying both (double calls are expensive).
- Solution: encode TIME in the ID. Using the known deletion policy, infer whether the record is still hot or already warm from the ID's timestamp. On a miss, fall back to the other tier.
- Concrete: Snowflake ID layout — 41 bits epoch-millis timestamp (leftmost) + 10 bits machine ID + 12 bits per-machine sequence (static counter, atomic increment); ~63 bits + sign. Decentralized (no central ID service). Machine 0..1023; collision only if >4096 IDs/ms/machine (~400k req/s). Discord/Sony use same layout; Discord sets custom epoch = 1 Jan 2015. Instagram variant: 41 bits timestamp (custom epoch 1 Jan 2011) + 13 bits DB shard ID + 10 bits per-shard sequence (left shift 23), generated inside the DB via stored procedure; Instagram had ~8000 logical DBs on only 3 Postgres servers.
- NOTE / correction: session 1 (line 129) says "first 23 bits were epoch milliseconds" — that is a transcription slip; the correct figure from the ID-generation session is 41 bits.
- IDs with time as the most-significant bits are roughly/uniformly sortable → enable ID/cursor-based pagination instead of LIMIT/OFFSET (constant-time deep pagination; can't do complex WHERE clauses). Not strictly monotonic (auto-increment is). Avoid Snowflake for strict-ordering needs like finance/banks.

# D. CDC vs app-emitted events
- Prefer CDC (Change Data Capture) over app/API-emitted Kafka events when moving raw data.
- Reason: not every table/update will have a Kafka event emitted by the app; you're dealing with raw data and need completeness regardless of which tables/updates occur. CDC reads directly from the DB and is context-free.
- Tools: Airbyte, Debezium (transcript "DBZM") read CDC and emit to destinations. CDC can fan out to multiple sinks (warm-tier dumper AND analytics/OLAP) from one source — clean architecture.
- CDC ≈ superset of replication: DBs with binlog/commit-log give CDC the log; DBs without them force CDC to iterate rows/scan tables. Normal master→replica replication itself works by reading the binlog/commit log.
- For analytics into OLAP, route via CDC rather than making the API layer emit Kafka events.

# E. Cold-tier access patterns
- Pointed reads on S3: possible if the stored format carries index/meta info (Word-Dictionary-style index+data, or Hudi). Tools that expose "query directly on S3": AWS Athena, Presto, Hive.
- Good for INFREQUENT one-off reads.
- Bulk/frequent reads (e.g. legal analytics for one accused person's history): don't hammer S3 (slow AND expensive — S3 cost + Presto infra cost). Instead LOAD the relevant historical data into a TEMPORARY queryable DB, run bulk queries, then delete the temp DB (data still safe on S3).
- OLAP (Redshift/BigQuery) is the fast-queryable analytical store; S3 is not built to serve that query frequency.
- Compression/encryption: don't rely on S3 for compression — use columnar format (e.g. Parquet) which compresses internally. Per-bucket custom encryption keys supported.

# F. Cost-driven design principles
- Business first, engineering second, performance third: pick cheapest commodity hardware (HDD) and extract max performance within that constraint.
- "60% of infrastructure cost comes from data analytics" → every architect should know data-engineering tools.
- Tiering reduces cost via smaller hot DBs.
- Minimize hops and infra size to protect margin.
- Don't over-engineer rare events (policy change once a year → redeploy, don't build real-time policy streaming).
- Top-down capacity planning is WRONG. Start from hardware limit and derive upward: e.g. 32 Gbps = max intra-datacenter network; divide by avg file size (1 GB → 32 files/s; 100 MB avg → ~320 req/s) → multiply by racks → that's your SLA/rate-limit guarantee. Unit economics drive how you limit per-token / per-key / per-bucket / per-partition.
- S3 "dirt cheap" is a trap at org lifespan scale (data never deleted; egress is hard).

# G. Full S3 / object storage design
Core identity:
- S3 = a gigantic KEY-VALUE store: key = file path (bucket/key), value = file content (up to terabytes). Directories/folders are LOGICAL/virtual, not physical. Bucket names are globally unique across all AWS accounts.
- Cannot use a transactional DB (MySQL) for storage layer — GBs/TBs of data.
- Requirement: storage must be ELASTIC (grow on demand). Client always PRE-TELLS S3 the file size before streaming bytes (reserve-then-write); never "append forever".

Storage layer:
- Choose HDD (magnetic, commodity) over SSD for cost/margin; S3 stores multiple exabytes (~7–8 EB at time of talk).
- HDDs can't do random writes well (head seek) → use a LOG-STRUCTURED, append-only file system: writes always go to the next sector sequentially. Local FS (FAT/NTFS/EXT/BTRFS) normally owns placement via inode/directory metadata; here we use a log-structured FS and only need to know which disk holds the file.
- Disks chained as a LINKED LIST with a HEAD POINTER naming the active disk. Unit = STORAGE RACK: 20–30 disks × 10–20 TB each = ~200–600 TB/rack; multiple racks per DC. Rack device-driver (Dell/Verbatim/WD) abstracts the head pointer and disk switching.
- Disk-full threshold = 70% (not 100%): keeps headroom for in-flight writes and for merge/compaction temp space, preserving 99.999999999-style availability. A monitor/rack driver switches the head.
- DELETE = remove entry + cleanup. MERGE & COMPACTION: scan disks, keep live files, skip deleted/stale entries, defragment, reclaim space. Heavy process (terabytes; O(n) per file).
- UPDATE = PUT the entire object again; same path → NEW consecutive location (never random writes). This is both a performance necessity and the mechanism behind OBJECT VERSIONING: old versions remain; enabling versioning = don't drop tail entries during merge/compaction; disabling = keep only latest during compaction. Need metadata mapping bucket+key → all version locations, persisted in a DB.
- READ path: random reads ARE allowed (head moves back); S3 is a WRITE-HEAVY system, not read-heavy (write:read ratio high; reads are slow so people don't read S3 transactionally; read-heavy = ~90:10 read ratio). Most content is backups/rarely-accessed media.

Routing / partitioning:
- Naive hash-based routing (incl. consistent hashing) is WRONG where tenant isolation is needed. Hashing distributes uniformly but YOU lose control → two hot tenants (e.g. Prime "Rings of Power" + Netflix "Sacred Games") can land on the same physical disk → HOT PARTITION problem: one tenant's load throttles another's even within per-tenant rate limits.
- S3 uses RANGE-BASED PARTITIONING on bucket/key (key range → physical rack/disk). Gives control, locality (a tenant's objects cluster), and tenant isolation; hot ranges can be SPLIT (L→L-M + N-R) and moved via disk/data movement. Documented: S3 uses range-based, not hash-based, partitioning.
- Hot-partition mitigations (general): (1) range-based partitioning + tenant isolation, (2) move/dump-load a logical shard off a hot node (Elasticsearch: shards + head plugin drag-and-drop rebalance; Instagram: 8000 logical DBs on few physical servers), (3) observe metrics to know WHICH partition is hot (observability is essential).
- Assigning one rack per tenant = wasteful; multiplex small tenants, and only migrate a bucket to a dedicated area once it crosses a threshold (e.g. >500 GB). Never give an entire rack to a tenant (kills margin).

API / control plane:
- S3 API servers (web servers) do NOT talk directly to physical storage (too many disks/racks; every server would need to know all storage; no tenant separation).
- PARTITION MAP TABLE: metadata mapping each key range/partition → owning partition server and physical location. Source of truth; must itself be replicated (with failover).
- PARTITION SERVER: owns partitions, knows how to talk to physical hardware/storage layer, reads and returns data. One partition ↔ one partition server; one partition server ↔ many partitions. Ownership is LOGICAL (cheap to reassign; no physical data move). Partition can span multiple racks; defragmentation tries to co-locate.
- PARTITION MANAGER = the orchestrator: health-checks partition servers and rebalances ownership.
- LEADER ELECTION between partition managers: one MASTER + multiple WORKERS; workers do health checks/balancing, master manages workers, and if master dies a worker is promoted. Solves "who monitors the monitor" and enables self-healing without humans.
- FLOW: user → S3 API server → consult partition map table → talk DIRECTLY to the owning partition server → storage layer → return. Deliberately does NOT proxy through the partition manager.
  - Why (counter to microservice dogma): blobs are huge (e.g. 1 GB); routing the payload through an extra proxy hop is costly and doubles request volume on the manager. Partition-map schema is stable/rarely changes, so sharing the metadata store directly is safe. This is a deliberately "anti-microservice" pragmatic decision.
- Durability (write-and-forget = trust S3): guarantee durability ONLY by redundancy. Levels: (1) RAID within a disk (write to 2 sectors, halves usable capacity, survives sector corruption); (2) disk within a rack (default: write to 2 disks, outlive disk failure); (3) across racks; (4) across data centers same region (async); (5) across geography/regions (async, chargeable; prescribed at least a daily backup; part of business-continuity plan investors demand). Within-datacenter replication is typically free/out-of-box; cross-region is costly.
- Data integrity: bit-flips corrupt data in transit (BAT→CAT). Use CHECKSUMS at EVERY layer: user→API, API→partition server, partition→disk, disk→partition→API→user, plus AWS SDK and HTTP checksums. Checksum = cheap XOR of 32-byte blocks. Reject the write if integrity fails. Often overlooked; essential because S3 is source of truth.
- Multipart upload / file limits: underlying log-structured FS caps single-file size; S3's logical limit may exceed the FS limit (e.g. FS 1 GB, S3 1 TB) via INHERENT CHUNKING — store as many FS files and keep "file path 1..n at location" metadata (in inode). Multipart: reserve space, upload chunks at offsets, then complete. Chunking also enables resume-after-failure. But chunking for read parallelism is rejected: S3 is write-heavy and files are read sequentially; if you need parallelism, create multiple full copies (use the durability replicas) rather than chunk-spread — simpler, avoids per-byte chunk lookups and metadata blowup.
- Hot single key across racks: don't build auto-rebalance for it. S3 enforces per-key / per-bucket / per-token / per-partition limits derived from hardware. For a genuinely huge customer, put a cache/CDN in front as an exception. Exceptional files can use a STATIC partition map fallback checked before range partitioning.
- Consistency/concurrency: S3 is simplified because objects are immutable — once written, not updated; any update is a new file. No concurrent-write handling needed; partition moves are safe because data is read-only. (The reference paper covers "strong consistency" but S3 has not published its model.)
- Security: S3 SIGNED URL = a certificate; validated locally at the edge (decrypt + check signature/validity) with NO DB call. Private key held by the customer (shipped in SDK); signatures are short-lived — hard to MITM.

Reference sources cited in the lecture:
- Paper: "Windows Azure Storage: A Highly Available Cloud Storage Service with Strong Consistency" — source of partition-map-table / partition-manager naming (Windows' equivalent of S3).
- Paper: "Building a Database on S3" — how to build queryable structures on S3 (Word Dictionary).
- Paper: Facebook "Scuba" (Facebook blob storage).
- Book: "Database Internals" by Alex Petrov.
- Spark book (Effective Spark) referenced in the other session.

# H. Anti-patterns / common wrong turns
- Keeping all historical data in the hot DB; never archiving by lifecycle.
- Routing every request to BOTH hot and warm instead of using timestamp-in-ID.
- Using hash-based routing / consistent hashing where tenant isolation is required (loses control; hot-partition collisions).
- Making S3 (or any blob store) behave like a transactional DB (auto-rebalancing hot keys, micro-reads, chunk-parallel reads).
- Assuming S3 is read-heavy when it is write-heavy.
- Relying on app-emitted Kafka events for data extraction instead of CDC (incomplete coverage of tables/updates).
- Proxying multi-GB blobs through extra microservice hops ("everything must go through a service").
- Over-engineering infrequent changes (policy) with real-time pipelines.
- Top-down capacity planning / promising throughput without knowing hardware limits.
- Giving a whole storage rack per tenant; not multiplexing.
- Optimizing for anomalies instead of common cases.
- Forgetting checksums/integrity and metadata (partition map, versions) replication.
- Building microservices reflexively — "good engineers think in problems and solutions, not in microservices"; simple systems scale.
- Ignoring observability — you cannot fix a hot partition you can't identify.

# I. Other systems/topics in the same file (other sessions)
- Distributed ID generators (session 2): auto-increment failure modes (clock, collisions, single point), Flickr/Amazon approaches, Snowflake (41/10/12, custom epochs, Discord 2015, Sony Go impl), Instagram DB-stored-procedure variant (41/13/10, custom epoch 2011), why UUIDs (16 bytes) bloat indexes 4× and hurt DB performance (indexes must fit RAM), index internals, ID/cursor-based pagination vs LIMIT/OFFSET.
- Social networks / Week 4 Day 2 (session 3): private image sharing, on-demand image optimization, photo tagging + cross-team collaboration and extensibility, CDN cache invalidation via Kafka on photo update, unread/new-message indicator (complex due to scale) and a high-level outage-prevention pattern.
- Storage index / Week 5 Day 2 (session 4): build a Word Dictionary (scalable storage+compute) with NO traditional database — index+data files (embedded-DB on S3), pointed reads; then build a super-fast KV store on commodity magnetic HDD — magnetic disk mechanics, head movement/seek, and BitCask/log-structured storage (also referenced as the basis of Kafka internals).
