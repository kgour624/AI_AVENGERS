# Architecture Learnings

> Dense permanent knowledge distilled from course transcripts under `Transcripts/`.
> For AI agents and engineers: read these instead of re-watching the full courses.

| File | Source transcript | Focus |
|---|---|---|
| [ultimate_go_k8s_design.md](ultimate_go_k8s_design.md) | Ultimate Go Software Design with Kubernetes 2.0 | Layered firewalls (API/App/Business/Foundation), domain lanes, auth, middleware, error handling, route-adder MUX, K8s deploy-first |
| [system_design_masterclass_adhoc.md](system_design_masterclass_adhoc.md) | System Design Master Class 1 | Distributed task scheduler, message broker on RDBMS, YouTube view counter, flash sale |
| [system_design_week3_distributed_foundations.md](system_design_week3_distributed_foundations.md) | System Design Master Class 2 | DS philosophy, load balancer LLD, remote/distributed locking (Redis/Redlock) |
| [system_design_tiered_storage_and_s3.md](system_design_tiered_storage_and_s3.md) | System Design Master Class 3 | Hot/warm/cold tiered storage, data-engineering pipeline, full S3 design |
| [microservices_masterclass_farley.md](microservices_masterclass_farley.md) | Microservices Masterclass | DRY vs coupling, monolith-first, coupled-modules anti-pattern, contracts, observability |

Related hub (product-applied rules): [`../KNOWLEDGE_HUB.md`](../KNOWLEDGE_HUB.md)

---

**How these were produced:** Transcripts were read end-to-end; only architecture principles, patterns, anti-patterns, and quotes were kept. Walkthrough fluff was dropped.
