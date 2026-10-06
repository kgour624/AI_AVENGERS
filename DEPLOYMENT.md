# AI Avengers — Complete Deployment Guide

> **Audience:** DevOps / Infrastructure Team  
> **Purpose:** End-to-end reference to clean, build, dockerise and deploy the AI Avengers platform.  
> **Author should send this file as-is — no extra explanation needed.**

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Repository Cleanup — Files to Delete Before Deploy](#2-repository-cleanup--files-to-delete-before-deploy)
3. [Files & Directories to KEEP](#3-files--directories-to-keep)
4. [Total Services at a Glance](#4-total-services-at-a-glance)
5. [Service-by-Service Details](#5-service-by-service-details)
6. [Infrastructure Requirements (Server Sizing)](#6-infrastructure-requirements-server-sizing)
7. [Environment Variables (.env)](#7-environment-variables-env)
8. [Build & Dockerise — Step-by-Step](#8-build--dockerise--step-by-step)
9. [Deploy to Production](#9-deploy-to-production)
10. [Database Migrations](#10-database-migrations)
11. [Health Checks & Monitoring](#11-health-checks--monitoring)
12. [SSL / HTTPS Setup](#12-ssl--https-setup)
13. [Rollback Procedure](#13-rollback-procedure)
14. [Cost Optimisation Tips](#14-cost-optimisation-tips)
15. [Troubleshooting](#15-troubleshooting)
16. [Quick Reference — Daily Operations](#16-quick-reference--daily-operations)

---

## 1. Project Overview

AI Avengers is a multi-tenant AI platform with:
- A **Go API backend** (business logic, LLM gateway, MCP server, orchestrator)
- A **React/Vite frontend** (static SPA served via nginx)
- A **Python ML sidecar** (embeddings + document extraction)
- An **Aider service** (AI-assisted code editing via Aider/LiteLLM)
- **PostgreSQL + pgvector** (primary DB with vector search)
- **Redis** (caching, JWT refresh tokens, rate limiting)
- **Nginx** (reverse proxy, SSL termination, rate limiting)

All services are containerised with Docker and orchestrated with Docker Compose.

---

## 2. Repository Cleanup — Files to Delete Before Deploy

These files are **developer/debug artefacts**. They should NOT be included in the production server.

### 2.1 Python Patch / Fix Scripts (root level — delete all)

```
apply_main42.py
apply_router42.py
dump.py
extend_admin_api42.py
extend_explorer42.py
fix1.py
fix_audit_dup.py
fix_audit_final.py
fix_audit_only_write.py
fix_cleanup.py
fix_explorer_tail.py
fix_frontend2.py
fix_layout.py
fix_now.py
fix_one.py
fix_parent.py
fix_router42.py
fix_router42b.py
fix_three.py
fix_two.py
fix_unicode.js
git_restore.py
monitor_training.py
patch1.py
patch2.py
patch_frontend.py
patch_main.py
patch_router.py
patch_router2.py
patch_router3.py
patch_slice42.py
plan.py
resume_job.py
run_build.py
scan.py scan2.py scan3.py scan4.py scan5.py scan6.py scan7.py
tmp_list.py
verify_fixes.py
verify_fixes2.py
parent_add.go
```

**Command to delete all at once (Linux/Mac):**
```bash
# Run from repo root
rm -f apply_main42.py apply_router42.py dump.py extend_admin_api42.py \
  extend_explorer42.py fix1.py fix_audit_dup.py fix_audit_final.py \
  fix_audit_only_write.py fix_cleanup.py fix_explorer_tail.py fix_frontend2.py \
  fix_layout.py fix_now.py fix_one.py fix_parent.py fix_router42.py \
  fix_router42b.py fix_three.py fix_two.py fix_unicode.js git_restore.py \
  monitor_training.py patch1.py patch2.py patch_frontend.py patch_main.py \
  patch_router.py patch_router2.py patch_router3.py patch_slice42.py plan.py \
  resume_job.py run_build.py scan.py scan2.py scan3.py scan4.py scan5.py \
  scan6.py scan7.py tmp_list.py verify_fixes.py verify_fixes2.py parent_add.go
```

---

### 2.2 Temporary Text / Debug Dump Files (root level — delete all)

```
.tmp_scan.txt
_final_ctx.txt
_final_ctx2.txt
_front_len.txt
_handler_head.txt
_handler_mid.txt
_handler_tail.txt
_m1.txt  _m2.txt  _m3.txt  _m4.txt  _m5.txt
_rcpt_files.txt
_recon2.txt  _recon3.txt  _recon4.txt  _recon_ctors.txt
_store_dump.txt  _store_mid.txt
_syms.txt
backend_logs.txt
build_output.txt
```

```bash
rm -f .tmp_scan.txt _final_ctx.txt _final_ctx2.txt _front_len.txt \
  _handler_head.txt _handler_mid.txt _handler_tail.txt \
  _m1.txt _m2.txt _m3.txt _m4.txt _m5.txt _rcpt_files.txt \
  _recon2.txt _recon3.txt _recon4.txt _recon_ctors.txt \
  _store_dump.txt _store_mid.txt _syms.txt backend_logs.txt build_output.txt
```

---

### 2.3 Large SQL Backup Files (DO NOT push to server — ~411 MB total)

> **WARNING:** These are local DB dumps. Store in secure backup storage (S3 / GCS), NOT in the Docker build context and NOT on the production server.

```
backup_ai_avengers.sql   (~274 MB)
backup_utf8.sql          (~137 MB)
reset.sql
```

```bash
rm -f backup_ai_avengers.sql backup_utf8.sql reset.sql
```

> If you need to restore from backup: upload to backup storage and restore using `psql` directly — do NOT include these in any Docker image.

---

### 2.4 Developer-Only Directories (delete from server / exclude from rsync)

```
Learnings/           — internal dev notes
Transcripts/         — local chat transcripts
mcp-learnings/       — dev learning notes
scratch/             — scratch/temp files
parts/               — partial file fragments
.kiro/               — IDE plugin config
.continue/           — Codeium/Continue IDE plugin
.vscode/             — VS Code workspace settings
```

> **IMPORTANT about `aider-main/`:** This directory IS needed at Docker **build time** (the `aider-service/Dockerfile` copies it during the image build). It does NOT need to be on the server after images are built. Include it during the build, exclude it afterwards.

```bash
rm -rf Learnings/ Transcripts/ mcp-learnings/ scratch/ parts/ .kiro/ .continue/ .vscode/
```

---

### 2.5 Redundant Documentation (safe to delete from server)

These are design/handoff docs for the dev team. The production server doesn't need them:

```
AI_AVENGERS_MCP_DESIGN_V1_INITIAL_LOCK.md
AI_AVENGERS_SYSTEM_ARCHITECTURE.md
CATEGORY_TEMPLATE_HANDOFF.md
CODECRAFTAPI_HANDOFF.md
CODECRAFTAPI_INTEGRATION_DESIGN.md
DOMAIN_EXPERT_COLLABORATION_DESIGN.md
EXPERT_TRAINING_ARCHITECTURE.md
FUTURE_UPDATES.md
HANDOFF.md
IMPLEMENTATION_HANDOFF.md
KNOWLEDGE_HUB.md
MCP_HANDOFF_V1.md
MIGRATIONS.md
ORCHESTRATOR_DOCS.md
TRANSCRIPT_NOISE_TABLE.md
new-ingestion.md
mcp_architecture.md
quickstart.sh
```

---

### 2.6 Files to Never Touch / Leave As-Is

```
.env              ← DO NOT delete. Fill in values before running.
.gitignore        ← keep
.gitlab-ci.yml    ← keep (CI/CD pipeline)
```

---

## 3. Files & Directories to KEEP

These are **required for the build and runtime**:

```
backend-go/              ← Go API source + Dockerfile + migrations
frontend/                ← React app source + Dockerfile + Dockerfile.dev
ml-sidecar/              ← Python ML service + Dockerfile
aider-service/           ← Python Aider wrapper + Dockerfile
aider-main/              ← Aider source (needed during aider-service Docker build ONLY)
nginx/
  └── nginx.conf         ← Nginx reverse proxy config
  └── ssl/               ← Create this directory, place cert.pem + key.pem here
docker-compose.yml       ← Development compose (reference only)
docker-compose.prod.yml  ← Production compose — USE THIS IN PROD
.env                     ← Your filled-in secrets (created from .env.prod.example)
.env.prod.example        ← Template for .env
Makefile                 ← Convenience commands
SETUP.md                 ← Initial setup notes
```

---

## 4. Total Services at a Glance

| # | Service | Technology | Internal Port | Purpose |
|---|---------|-----------|--------------|---------|
| 1 | **nginx** | nginx:alpine | 80 / 443 | Reverse proxy, SSL, rate limiting |
| 2 | **api** | Go 1.23 (Alpine) | 8080 | Main backend — REST API, LLM gateway, orchestrator, MCP |
| 3 | **frontend** | React/Vite → nginx | 80 | Static SPA (built once, served by nginx) |
| 4 | **ml-sidecar** | Python 3.11 + FastAPI | 8001 | Embeddings (bge-base-en-v1.5) + document extraction |
| 5 | **aider-service** | Python 3.11 + FastAPI + Aider | 8082 | AI code editing agent |
| 6 | **postgres** | pgvector/pgvector:pg15 | 5432 | Primary database with vector search |
| 7 | **redis** | redis:7-alpine | 6379 | Cache, sessions, rate-limit state |
| *(8)* | **migrate** | Same image as `api` | — | One-shot migration runner (exits after run) |

**Total running containers in production: 7** (`migrate` exits automatically after applying migrations)

---

## 5. Service-by-Service Details

---

### 5.1 `api` — Go Backend

**What it does:**
- REST API for all business logic (users, experts, chats, workflows, knowledge base)
- LLM Gateway — proxies calls to OpenRouter / other providers, tracks costs
- MCP (Model Context Protocol) server for external tool integrations
- Orchestrator — multi-agent workflow engine (receptionist, domain experts, implementation, QA)
- JWT authentication, RBAC, multi-tenancy

**Build complexity:** Medium
- Multi-stage Dockerfile (builder stage → minimal alpine runtime)
- Binary is statically compiled (`CGO_ENABLED=0`) — no C library dependencies at runtime
- **Runtime image includes `go`, `nodejs`, `npm`, `git`, `rsync`** — this is intentional and required:
  - The API validates AI-generated code by running `go build`/`go test` and `npm run test` inside the container (code validation sandbox)
  - This adds ~300 MB to the image but is required for the Implementation and QA workflow phases
  - Without these tools, AI-generated code cannot be verified — the platform still works but validation degrades to "unverified"
  - Security note: The sandbox runs code as non-root user with ulimits and timeouts applied

**Critical shared volume:** `workspaces` (shared with `aider-service`, both use UID=1000)

**Health check:** `GET /health` → HTTP 200

**Resource limits (prod):**
```
memory: 512M limit / 256M reserved
cpu: 1.0 core
```

**Ports:** Internal only — NOT exposed to host in prod. nginx proxies all external traffic.

---

### 5.2 `frontend` — React SPA

**What it does:**
- Vite-built React/TypeScript SPA
- Served by a minimal nginx container
- Zero server-side logic — pure static files

**Build complexity:** Low
- `npm run build` → static files → copied into nginx:alpine container
- Production Dockerfile (`frontend/Dockerfile`) builds the app and serves via internal nginx

**CRITICAL — API URL baked at build time:**
The `VITE_API_URL` environment variable is baked into the static bundle at build time (not runtime) by Vite. For production, set the URL before building:

```bash
# Edit frontend/.env.production before building
echo "VITE_API_URL=https://yourdomain.com" > frontend/.env.production
```

Or pass as a build arg:
```bash
docker build --build-arg VITE_API_URL=https://yourdomain.com -t frontend:prod ./frontend
```

**Resource limits (prod):**
```
memory: 64M limit
cpu: 0.25 core
```

---

### 5.3 `ml-sidecar` — Python ML Service

**What it does:**
- Embedding generation using `bge-base-en-v1.5` (~500 MB model weights)
- Re-ranking using `bge-reranker-base` (~280 MB model weights)
- Document text extraction supporting: PDF, DOCX, XLSX, XLS, PPTX, HTML, EPUB, RTF, ODT/ODS/ODP, CSV, TXT, SRT
- Runs as single-worker uvicorn (models load once into memory, serve all requests async)

**Build complexity:** HIGH — first container start downloads ~800 MB of ML model weights from HuggingFace

**Critical behaviour to understand:**
- `start_period: 120s` in health check — models take 1-2 minutes to load at startup. Do NOT restart the container if health check fails before 2 minutes.
- **NEVER set `--workers > 1`** — each worker loads its own copy of the ~800 MB models. 2 workers = 1.6 GB extra memory.
- Volume `ml-models:/root/.cache/huggingface` **MUST persist** between restarts — this caches the downloaded models and prevents re-downloading 800 MB every time.

**Resource limits (prod):**
```
memory: 2G limit / 1G reserved   ← DO NOT reduce below 1G — models will not load
cpu: 2.0 cores
```

**Health check:** `GET /health` → HTTP 200 (allow up to 120s on first start)

**Optional LibreOffice support (exotic doc formats):**
```bash
# Only needed for .wpd, .pages, .key, etc.
# Adds ~500MB to image — disabled by default
docker build --build-arg WITH_LIBREOFFICE=true -f ml-sidecar/Dockerfile ./ml-sidecar
```

---

### 5.4 `aider-service` — Python AI Code Editor

**What it does:**
- Wraps the Aider AI coding tool as an HTTP microservice
- Receives code editing tasks from the Go API (implementation and QA phases)
- Routes LLM calls through the Go API's LLM proxy endpoint for centralised cost tracking
- Operates on git workspaces in the shared Docker volume

**Build complexity:** Medium-High — multiple dependency pins required for stability

**Critical build details:**
- `aider-main/` directory must exist in repo root at build time (build context is `.`, not `./aider-service/`)
- `setuptools-scm<8` pin is critical — newer versions break Aider's editable install
- `SETUPTOOLS_SCM_PRETEND_VERSION=1.0.0` environment variable is required at build — Aider cannot derive a version from the copied source tree (no `.git` directory)
- `--no-build-isolation` flag is required on the pip install — prevents pip from pulling the incompatible setuptools-scm in an isolated build environment

**Critical configuration (must be in sync with `api` service):**
```
AIDER_PROXY_TOKEN   — Must be the SAME value in both api and aider-service
MODEL_GATEWAY_URL   — Must be http://api:8080/api/v1/llm/proxy
```
If `AIDER_PROXY_TOKEN` is empty or mismatched, the proxy returns 503 and all implementation/QA phases fail.

**Critical shared volume:** `workspaces` — MUST be shared between `api` and `aider-service`
- Both containers use UID 1000 / GID 1000
- The `api` creates workspace directories; `aider-service` writes and commits code inside them
- If UIDs differ, one container cannot write files created by the other

---

### 5.5 `postgres` — Database

**What it does:**
- Primary relational database for all platform data
- `pgvector` extension enabled — required for vector similarity search (RAG / knowledge retrieval)
- Stores: users, experts, chats, messages, knowledge chunks, embeddings, workflows, audit logs, MCP configs, tenants

**Image:** `pgvector/pgvector:pg15` — PostgreSQL 15 with pgvector pre-installed (NOT the standard postgres image)

**Total migrations:** 83 migration files covering the full schema

**Resource limits (prod):**
```
memory: 1G limit
cpu: 1.0 core
```

**BACKUP CRITICAL — all user data lives here:**
```bash
# Manual backup
docker exec $(docker compose -f docker-compose.prod.yml ps -q postgres) \
  pg_dump -U avengers ai_avengers | gzip > backup_$(date +%Y%m%d_%H%M).sql.gz

# Restore from backup
gunzip -c backup_20261006.sql.gz | docker exec -i \
  $(docker compose -f docker-compose.prod.yml ps -q postgres) \
  psql -U avengers ai_avengers
```

Set up automated daily backups — minimum backup retention: 7 days.

---

### 5.6 `redis` — Cache & Sessions

**What it does:**
- L1 memory cache for frequently accessed DB records
- JWT refresh token storage
- Rate limiting state (per-IP and per-user counters)
- OAuth state storage
- Append-only persistence (`--appendonly yes`) — data survives restarts

**Image:** `redis:7-alpine` (minimal ~30 MB image)

**Resource limits (prod):**
```
memory: 256M limit
cpu: 0.5 core
```

**Password required in production** — set `REDIS_PASSWORD` in `.env`. An empty password is rejected by the health check command in the prod compose.

---

### 5.7 `nginx` — Reverse Proxy

**What it does:**
- Single entry point for all external traffic (ports 80 and 443)
- Routes `/api/*` → `api` container (port 8080)
- Routes `/` → `frontend` container (port 80)
- Rate limiting (two zones):
  - `auth` zone: `/api/v1/auth/*` — 10 req/min per IP, burst 5
  - `api` zone: all other `/api/*` — 100 req/min per IP, burst 20
- SSE (Server-Sent Events) streaming support for LLM responses — no buffering, 5-minute read timeout
- SSL/TLS termination

**SSL certificate paths** (mount into container):
```
nginx/ssl/cert.pem   ← full certificate chain
nginx/ssl/key.pem    ← private key
```

Already configured in `docker-compose.prod.yml`:
```yaml
volumes:
  - ./nginx/ssl:/etc/nginx/ssl:ro
```

**To enable HTTPS redirect** (after adding certs), uncomment in `nginx/nginx.conf`:
```nginx
return 301 https://$host$request_uri;
```

---

### 5.8 `migrate` — One-Shot Migration Runner

**What it does:**
- Runs `./migrate up` once at startup
- Applies all pending database migrations (83 total as of this writing)
- Uses golang-migrate which tracks applied versions — safe to re-run (idempotent)
- **Exits with code 0** after completion — it is NOT a long-running service

**Startup dependency chain:**
```
postgres (health check passes)
     ↓
migrate (runs, exits 0)
     ↓
api (starts only after migrate completes successfully)
```

---

## 6. Infrastructure Requirements (Server Sizing)

### Minimum Production Server

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| **CPU** | 4 vCPU | 8 vCPU |
| **RAM** | 6 GB | 12 GB |
| **Disk — OS + Docker images** | 40 GB SSD | 80 GB SSD |
| **Disk — postgres data volume** | 20 GB | 100 GB (grows with user data) |
| **Disk — ML models cache** | 2 GB | 4 GB |
| **OS** | Ubuntu 22.04 LTS | Ubuntu 22.04 LTS |
| **Docker** | 24.x | 25.x+ |
| **Docker Compose** | v2.x | v2.24+ |

### Memory Breakdown (all services running simultaneously):

| Service | Memory Reserved | Memory Limit |
|---------|----------------|--------------|
| api | 256 MB | 512 MB |
| frontend | — | 64 MB |
| ml-sidecar | 1,024 MB | 2,048 MB |
| postgres | — | 1,024 MB |
| redis | — | 256 MB |
| nginx | ~20 MB | ~50 MB |
| aider-service | ~150 MB | ~512 MB |
| OS + Docker daemon | ~500 MB | — |
| **Total** | **~1.95 GB** | **~4.5 GB** |

> **Safe minimum RAM: 6 GB** to leave OS and headroom. **8 GB recommended.**

### Cloud Instance Examples (single-server deploy):

| Provider | Instance | vCPU | RAM | Cost/mo (approx.) |
|----------|---------|------|-----|--------------------|
| AWS | t3.xlarge | 4 | 16 GB | ~$120 |
| AWS | t3.large | 2 | 8 GB | ~$60 (minimum) |
| GCP | e2-standard-4 | 4 | 16 GB | ~$100 |
| DigitalOcean | s-4vcpu-8gb | 4 | 8 GB | ~$48 |
| Hetzner | CX31 | 2 | 8 GB | ~$10 (best value) |
| Hetzner | CAX21 (ARM64) | 4 | 8 GB | ~$8 (cheapest) |

> **Cost tip:** Hetzner CAX21 (ARM64) is the most cost-efficient option. All Docker images in this project build for `linux/arm64`. Build on the ARM server directly or use `docker buildx` for cross-platform builds.

---

## 7. Environment Variables (`.env`)

Copy the template and fill in all values before first deploy:

```bash
cp .env.prod.example .env
nano .env   # or your preferred editor
```

### Required variables (server will NOT start without these):

```env
# Database
DATABASE_URL=postgres://avengers:STRONG_PASS@postgres:5432/ai_avengers?sslmode=disable
POSTGRES_USER=avengers
POSTGRES_PASSWORD=STRONG_PASS_HERE          # Min 16 chars, use alphanumeric + symbols
POSTGRES_DB=ai_avengers

# Redis
REDIS_URL=redis://:STRONG_REDIS_PASS@redis:6379
REDIS_PASSWORD=STRONG_REDIS_PASS_HERE

# JWT signing secret — MINIMUM 32 characters, random string
JWT_SECRET=<generate: openssl rand -hex 32>

# LLM API Gateway
OPENROUTER_API_KEY=sk-or-v1-XXXXXXXXXXXXXXXXXXXXXXXX

# OAuth token encryption — MUST be EXACTLY 32 bytes (32 ASCII characters)
ENCRYPTION_KEY=<generate: openssl rand -hex 16>

# Shared secret between api and aider-service containers
# MUST match exactly in both services
AIDER_PROXY_TOKEN=<generate: openssl rand -hex 32>
```

### How to generate secure secrets:

```bash
# JWT_SECRET (64-char hex string)
openssl rand -hex 32

# ENCRYPTION_KEY (exactly 32 bytes = 32 printable ASCII chars)
openssl rand -hex 16

# AIDER_PROXY_TOKEN
openssl rand -hex 32

# POSTGRES_PASSWORD (strong, URL-safe)
openssl rand -base64 24 | tr -dc 'A-Za-z0-9' | head -c 24
```

### Optional variables:

```env
# GitHub / GitLab OAuth (leave empty to disable OAuth login, PAT-based auth still works)
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GITLAB_CLIENT_ID=
GITLAB_CLIENT_SECRET=

# Public domain
BASE_URL=https://yourdomain.com
FRONTEND_URL=https://yourdomain.com
CORS_ALLOWED_ORIGINS=https://yourdomain.com

# Account policy (false = admin creates all accounts — recommended for production)
SELF_REGISTRATION_ENABLED=false

# Aider model settings
AIDER_MODEL=openai/deepseek-chat
AIDER_EDIT_FORMAT=diff
AIDER_MAX_INPUT_TOKENS=64000
AIDER_MAX_OUTPUT_TOKENS=8000

# Docker registry (for pulling pre-built images)
REGISTRY=ghcr.io/YOUR_ORG
VERSION=latest
```

---

## 8. Build & Dockerise — Step-by-Step

### Prerequisites on the build/server machine:

```bash
docker --version         # Must be 24.0+
docker compose version   # Must be v2.x
git --version
```

---

### Step 1 — Clone & Clean the Repository

```bash
git clone https://github.com/YOUR_ORG/AI_AVENGERS.git
cd AI_AVENGERS

# Delete all developer artefact files (see Section 2 for full list)
rm -f apply_*.py fix_*.py patch_*.py scan*.py verify_fixes*.py monitor_training.py
rm -f dump.py extend_*.py git_restore.py plan.py resume_job.py run_build.py tmp_list.py
rm -f parent_add.go fix_unicode.js
rm -f backup_*.sql reset.sql
rm -f _*.txt *.txt
rm -rf Learnings/ Transcripts/ mcp-learnings/ scratch/ parts/ .kiro/ .continue/ .vscode/
```

---

### Step 2 — Setup Environment File

```bash
cp .env.prod.example .env
# Fill in all required values (see Section 7)
nano .env
```

---

### Step 3 — Set Frontend API URL

```bash
# Tell the frontend where the API lives (baked into the static bundle at build time)
echo "VITE_API_URL=https://yourdomain.com" > frontend/.env.production
```

---

### Step 4 — Build All Docker Images

```bash
# Build all services (no-cache ensures fresh build)
docker compose -f docker-compose.prod.yml build --no-cache

# OR build services individually (useful for debugging one service):
docker compose -f docker-compose.prod.yml build api
docker compose -f docker-compose.prod.yml build frontend
docker compose -f docker-compose.prod.yml build ml-sidecar
docker compose -f docker-compose.prod.yml build aider-service
```

> **Note on aider-service build:** Build context is `.` (repo root), NOT `./aider-service/`. The `aider-main/` directory at root level is copied during build. Ensure it exists.

**Expected build times (first build, cold cache):**

| Service | Build Time | Final Image Size |
|---------|-----------|-----------------|
| api | ~3-5 min | ~800 MB (includes Go toolchain + nodejs) |
| frontend | ~2 min | ~50 MB (nginx + static bundle) |
| ml-sidecar | ~5-8 min | ~3 GB (PyTorch + sentence-transformers) |
| aider-service | ~4-6 min | ~1.5 GB (Aider + LiteLLM + dependencies) |

> **ml-sidecar first start:** After the container starts the FIRST time, it downloads ~800 MB of ML model weights from HuggingFace. This takes 2-5 minutes depending on internet speed. It is cached in the `ml-models` Docker volume. Subsequent starts load from cache (~30-60s).

---

### Step 5 — Tag and Push to Registry (optional but recommended for CI/CD)

```bash
# Set your registry
REGISTRY=ghcr.io/YOUR_ORG
VERSION=v1.0.0

# Tag
docker tag ai-avengers-1-api:latest       ${REGISTRY}/ai-avengers/backend:${VERSION}
docker tag ai-avengers-1-frontend:latest  ${REGISTRY}/ai-avengers/frontend:${VERSION}
docker tag ai-avengers-1-ml-sidecar:latest ${REGISTRY}/ai-avengers/ml-sidecar:${VERSION}
docker tag ai-avengers-1-aider-service:latest ${REGISTRY}/ai-avengers/aider-service:${VERSION}

# Push
docker push ${REGISTRY}/ai-avengers/backend:${VERSION}
docker push ${REGISTRY}/ai-avengers/frontend:${VERSION}
docker push ${REGISTRY}/ai-avengers/ml-sidecar:${VERSION}
docker push ${REGISTRY}/ai-avengers/aider-service:${VERSION}
```

Update `.env`:
```env
REGISTRY=ghcr.io/YOUR_ORG/ai-avengers
VERSION=v1.0.0
```

---

## 9. Deploy to Production

### On the production server:

```bash
# 1. Create deployment directory
mkdir -p /opt/ai-avengers
cd /opt/ai-avengers

# 2. Copy necessary files to server
# (Use rsync, scp, or git clone — do NOT copy the large SQL backups)
# Minimum files needed on server if using pre-built images:
#   docker-compose.prod.yml
#   .env
#   nginx/nginx.conf
#   nginx/ssl/cert.pem
#   nginx/ssl/key.pem
#
# If building images ON the server, also copy:
#   backend-go/
#   frontend/
#   ml-sidecar/
#   aider-service/
#   aider-main/

# 3. Create SSL directory and place certificates
mkdir -p nginx/ssl
# Copy your SSL cert and key:
# nginx/ssl/cert.pem   ← full chain (cert + intermediates)
# nginx/ssl/key.pem    ← private key

# 4. Start all services
docker compose -f docker-compose.prod.yml up -d

# 5. Watch startup logs (Ctrl+C to exit, services keep running)
docker compose -f docker-compose.prod.yml logs -f

# 6. Verify all services are running and healthy
docker compose -f docker-compose.prod.yml ps
```

---

### Service Startup Order (managed automatically by `depends_on`):

```
postgres  ──────────────────── health check passes
redis     ──────────────────── health check passes
               ↓
         migrate  ──────────── applies 83 migrations, exits 0
               ↓
            api  ────────────── starts after migrate completes + redis healthy
               ↓
     aider-service ──────────── starts after api starts
     ml-sidecar ─────────────── starts in parallel (models loading ~120s)
     frontend ───────────────── starts after api starts
               ↓
           nginx ───────────── starts last, depends on api + frontend
```

---

### Verify Deployment:

```bash
# All containers should show "healthy" status
docker compose -f docker-compose.prod.yml ps

# API health check
curl http://localhost/health
# Expected: {"status":"ok"} or similar

# If SSL is configured
curl https://yourdomain.com/health

# Check for errors in logs
docker compose -f docker-compose.prod.yml logs api --tail=100
docker compose -f docker-compose.prod.yml logs ml-sidecar --tail=50
docker compose -f docker-compose.prod.yml logs migrate
```

---

## 10. Database Migrations

Migrations are applied automatically by the `migrate` service on every `docker compose up`. They are idempotent — golang-migrate tracks applied migration versions in a `schema_migrations` table.

### Migration Details:
- **Tool:** golang-migrate (embedded in Go binary as `./migrate` command)
- **Total migrations:** 83 (as of this deployment)
- **Location in repo:** `backend-go/migrations/`
- **Naming convention:** `NNN_description.up.sql` / `NNN_description.down.sql`
- **Tracker:** `schema_migrations` table in the database

### Manual migration commands (if automation fails):

```bash
# Apply all pending migrations
docker compose -f docker-compose.prod.yml run --rm migrate ./migrate up

# Check current migration version
docker compose -f docker-compose.prod.yml run --rm migrate ./migrate version

# Roll back the last applied migration
docker compose -f docker-compose.prod.yml run --rm migrate ./migrate down 1

# Force set version (use only to fix dirty state)
docker compose -f docker-compose.prod.yml run --rm migrate ./migrate force NNN
```

### First-time Setup — Admin User:
Migration `005_seed_admin.up.sql` creates the initial admin account. Check that file for default credentials. **Change the admin password immediately after first login via the Admin panel.**

---

## 11. Health Checks & Monitoring

### Built-in Health Check Endpoints:

| Service | Endpoint | Expected Response |
|---------|----------|------------------|
| api | `GET /health` | HTTP 200 |
| ml-sidecar | `GET http://ml-sidecar:8001/health` | HTTP 200 (wait 120s on first start) |
| nginx (external) | `GET https://yourdomain.com/health` | HTTP 200 |

### Check container health:

```bash
# Show all containers with health status
docker compose -f docker-compose.prod.yml ps

# Detailed health info
docker inspect --format='{{.Name}} {{.State.Health.Status}}' \
  $(docker compose -f docker-compose.prod.yml ps -q)
```

### Resource monitoring:

```bash
# Live stats for all containers (CPU, memory, network, disk I/O)
docker stats

# Stats for specific service
docker stats $(docker compose -f docker-compose.prod.yml ps -q api)

# Follow logs for all services
docker compose -f docker-compose.prod.yml logs -f

# Follow logs for one service
docker compose -f docker-compose.prod.yml logs -f api
docker compose -f docker-compose.prod.yml logs -f ml-sidecar
```

### External uptime monitoring (recommended):

- **UptimeRobot** (free tier) — monitor `https://yourdomain.com/health` every 5 minutes
- Set up email/SMS alerts for downtime

---

## 12. SSL / HTTPS Setup

### Option A — Let's Encrypt with Certbot (recommended)

```bash
# Install certbot on the host
sudo apt install certbot -y

# Stop nginx temporarily so certbot can use port 80
docker compose -f docker-compose.prod.yml stop nginx

# Obtain certificate (replace with your actual domain)
sudo certbot certonly --standalone -d yourdomain.com -d www.yourdomain.com

# Copy certs to nginx/ssl/
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem /opt/ai-avengers/nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem /opt/ai-avengers/nginx/ssl/key.pem
sudo chmod 644 /opt/ai-avengers/nginx/ssl/cert.pem
sudo chmod 600 /opt/ai-avengers/nginx/ssl/key.pem

# Enable HTTPS redirect in nginx/nginx.conf (uncomment this line):
# return 301 https://$host$request_uri;
# And add an HTTPS server block pointing to port 443

# Restart nginx
docker compose -f docker-compose.prod.yml start nginx

# Auto-renew (certbot handles this — just make sure the cron job runs)
sudo certbot renew --dry-run
```

**Auto-renewal hook** — add to `/etc/letsencrypt/renewal-hooks/post/`:
```bash
#!/bin/bash
cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem /opt/ai-avengers/nginx/ssl/cert.pem
cp /etc/letsencrypt/live/yourdomain.com/privkey.pem /opt/ai-avengers/nginx/ssl/key.pem
docker compose -f /opt/ai-avengers/docker-compose.prod.yml exec nginx nginx -s reload
```

---

### Option B — Cloudflare Proxy (simplest)

1. Point DNS to your server IP via Cloudflare
2. Enable "Proxied" (orange cloud) in Cloudflare DNS settings
3. Cloudflare handles SSL — set nginx to HTTP only (port 80)
4. In Cloudflare SSL settings, set to "Full" or "Full (Strict)"
5. Optionally restrict nginx to only accept traffic from Cloudflare IPs

---

## 13. Rollback Procedure

### If a deployment goes wrong:

```bash
# 1. Check what is failing
docker compose -f docker-compose.prod.yml logs api --tail=50
docker compose -f docker-compose.prod.yml ps

# 2. Stop all services
docker compose -f docker-compose.prod.yml down

# 3. If a DB migration was applied that needs to be reversed, roll it back FIRST
docker compose -f docker-compose.prod.yml run --rm migrate ./migrate down 1
# (repeat as needed for multiple migrations)

# 4. Roll back to previous image version
# In .env, change:
VERSION=v0.9.0   # previous stable version that was working

# 5. Pull the previous images from registry
docker compose -f docker-compose.prod.yml pull

# 6. Start with the previous version
docker compose -f docker-compose.prod.yml up -d

# 7. Verify recovery
docker compose -f docker-compose.prod.yml ps
curl http://localhost/health
```

> **IMPORTANT ORDER:** Always roll back DB migrations BEFORE rolling back the API image if the migration changed the schema. Rolling back the image without rolling back the migration leaves the DB in an incompatible state.

---

## 14. Cost Optimisation Tips

### High Impact

1. **ml-sidecar is the most resource-expensive service** (2 CPU, 2 GB RAM).
   - If the Knowledge Base / RAG feature is NOT needed, disable it by removing the `ml-sidecar` service from `docker-compose.prod.yml` and setting `ML_SIDECAR_URL=` empty in the API environment.
   - This saves ~2 GB RAM and 2 CPU cores.

2. **Use Hetzner CAX (ARM64) servers** — 30-40% cheaper than equivalent AWS/GCP. All images build for `linux/arm64`.

3. **LLM costs** are the biggest operational cost. Control them via the Admin panel:
   - Set conservative token limits per model
   - Enable rate limiting per user
   - Monitor usage in the admin dashboard

### Medium Impact

4. **aider-service** is only needed for the AI code editing workflow (Implementation and QA phases). If code generation features are NOT used, you can comment it out of the compose file.

5. **Redis persistence:** `--appendonly yes` writes every operation to disk. For cost-sensitive deployments where losing cache on restart is acceptable:
   ```yaml
   command: redis-server --save ""  # In-memory only, no disk write
   ```
   This reduces disk I/O significantly.

6. **Scale `ml-sidecar` horizontally** (multiple containers) instead of vertically (bigger instance) if throughput is the bottleneck. Each container is stateless for the API, just memory-heavy for model weights.

### Low Impact

7. **`ml-models` volume** — never delete this volume unless you are OK with re-downloading ~800 MB. Keep it persistent.

8. **Set AIDER_MAX_INPUT_TOKENS and AIDER_MAX_OUTPUT_TOKENS** conservatively in `.env` to limit per-task LLM spend.

9. **Clean Docker system** periodically to free disk:
   ```bash
   docker system prune -a  # Removes stopped containers + unused images + build cache
   # WARNING: this removes ALL unused images, including cached build layers
   # Only run if you have pushed images to a registry
   ```

---

## 15. Troubleshooting

### `migrate` service fails — "connection refused" or "dial error"

```
Cause: postgres container is not ready yet (still initializing).
Fix:   postgres healthcheck should handle this automatically (retries 5x).
       If it persists, check postgres logs:
       docker compose -f docker-compose.prod.yml logs postgres
```

### `api` fails to start — "migrate not completed"

```
Cause: The migrate service exited with a non-zero code (migration failed).
Fix:
  docker compose -f docker-compose.prod.yml logs migrate
  # Fix the error, then:
  docker compose -f docker-compose.prod.yml run --rm migrate ./migrate up
```

### `ml-sidecar` shows "unhealthy" after 120+ seconds

```
Cause 1: Not enough RAM to load ML models. Need minimum 1 GB reserved.
Fix:     docker stats  — check if memory limit is being hit.
         Increase memory limit in docker-compose.prod.yml.

Cause 2: HuggingFace model download failed (network issue).
Fix:     docker compose -f docker-compose.prod.yml logs ml-sidecar
         Look for download errors. Restart the container.
```

### `aider-service` returns 503 on implementation tasks

```
Cause: AIDER_PROXY_TOKEN mismatch between api and aider-service containers.
Fix:   Ensure both services have the EXACT same AIDER_PROXY_TOKEN value in .env.
       docker compose -f docker-compose.prod.yml logs aider-service | grep -i "token\|proxy\|503"
       Then restart both: docker compose -f docker-compose.prod.yml restart api aider-service
```

### "Permission denied" on workspaces volume

```
Cause: The workspaces Docker volume was created when the api container ran as root
       (old image). New image runs as UID 1000 but the volume root is still root-owned.
Fix:
  docker compose -f docker-compose.prod.yml down
  docker volume rm $(docker compose -f docker-compose.prod.yml config --volumes | grep workspaces)
  # WARNING: This deletes all in-progress workflow workspaces
  docker compose -f docker-compose.prod.yml up -d
```

### Frontend shows a blank page or stale content after deploy

```
Cause 1: VITE_API_URL was not set before building. The frontend cannot reach the API.
Fix:     Set the correct VITE_API_URL in frontend/.env.production and rebuild.

Cause 2: Browser cache showing old bundle.
Fix:     Hard refresh: Ctrl+Shift+R (or Cmd+Shift+R on Mac).
         Or: docker compose -f docker-compose.prod.yml exec nginx nginx -s reload
```

### Out of disk space

```
Fix:
  # Show disk usage by Docker
  docker system df

  # Safe cleanup (only stopped containers + dangling images)
  docker container prune -f
  docker image prune -f

  # More aggressive (removes ALL unused images — only if you have registry backup)
  docker system prune -a

  # Check volume sizes
  docker system df -v
```

### Nginx 502 Bad Gateway

```
Cause: api or frontend container is not running or not healthy.
Fix:
  docker compose -f docker-compose.prod.yml ps
  docker compose -f docker-compose.prod.yml logs api --tail=30
  docker compose -f docker-compose.prod.yml restart api
```

### Rate limit 429 errors hitting legitimate users

```
Cause: Default rate limits may be too strict for your usage pattern.
Fix:   Edit nginx/nginx.conf:
       - auth zone: increase rate= and burst=
       - api zone:  increase rate= and burst=
       Then: docker compose -f docker-compose.prod.yml exec nginx nginx -s reload
```

---

## 16. Quick Reference — Daily Operations

```bash
# ── Start / Stop ──────────────────────────────────────────────
# Start everything
docker compose -f docker-compose.prod.yml up -d

# Stop everything (keeps all data volumes intact)
docker compose -f docker-compose.prod.yml down

# Restart a single service (e.g. after config change in .env)
docker compose -f docker-compose.prod.yml restart api

# ── Logs ──────────────────────────────────────────────────────
# All services, live
docker compose -f docker-compose.prod.yml logs -f

# Single service, last 100 lines
docker compose -f docker-compose.prod.yml logs api --tail=100

# ── Status & Resources ────────────────────────────────────────
# Service status + health
docker compose -f docker-compose.prod.yml ps

# Live CPU/memory stats
docker stats

# ── Database ──────────────────────────────────────────────────
# Backup database
docker exec $(docker compose -f docker-compose.prod.yml ps -q postgres) \
  pg_dump -U avengers ai_avengers | gzip > backup_$(date +%Y%m%d_%H%M).sql.gz

# Apply new migrations
docker compose -f docker-compose.prod.yml run --rm migrate ./migrate up

# Open psql shell
docker exec -it $(docker compose -f docker-compose.prod.yml ps -q postgres) \
  psql -U avengers ai_avengers

# ── Deployment Update ─────────────────────────────────────────
# Rebuild images and redeploy (brief downtime)
docker compose -f docker-compose.prod.yml build --no-cache
docker compose -f docker-compose.prod.yml up -d --force-recreate

# Pull pre-built images from registry and redeploy
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d --force-recreate

# ── Cleanup ───────────────────────────────────────────────────
# Remove stopped containers and dangling images (safe)
docker container prune -f && docker image prune -f

# Full cleanup (WARNING: removes ALL unused images)
docker system prune -a
```

---

*AI Avengers Deployment Guide — Generated 6 October 2026*  
*Contact the development team for questions about specific service internals.*
