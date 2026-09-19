# AI Avengers — Setup Guide

## Prerequisites

- Docker & Docker Compose
- OpenRouter API key (https://openrouter.ai)

## Quick Start

```bash
# 1. Clone and setup
git clone https://gitlab.com/zepto-group3/ai_avengers.git
cd ai_avengers
chmod +x quickstart.sh
./quickstart.sh

# 2. Edit .env with your values
nano .env

# 3. Run again to start
./quickstart.sh
```

## Manual Setup

### 1. Environment Variables

```bash
cp .env.prod.example .env
# Edit .env with your values
```

Required:
- `DATABASE_URL` — PostgreSQL connection string
- `JWT_SECRET` — minimum 32 characters
- `OPENROUTER_API_KEY` — from openrouter.ai
- `ENCRYPTION_KEY` — exactly 32 characters (for OAuth token encryption)

### 2. Start Services

```bash
# Development
docker-compose up -d

# Production
docker-compose -f docker-compose.prod.yml up -d
```

### 3. Run Migrations

```bash
# From backend-go directory
DATABASE_URL=postgres://... go run cmd/migrate/main.go up

# Or via Docker
docker-compose exec api go run cmd/migrate/main.go up
```

### 4. Create Admin User

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U avengers -d ai_avengers

# Create admin user (replace with your values)
INSERT INTO users (email, hashed_password, full_name, role)
VALUES (
    'admin@yourdomain.com',
    -- Generate hash: htpasswd -bnBC 12 "" 'your_password' | tr -d ':\n'
    '$2y$12$HASH_HERE',
    'Admin',
    'admin'
);
```

### 5. Setup TOTP for Admin

```bash
# After creating admin user, call:
POST /api/v1/auth/admin/login
{"email": "admin@...", "password": "..."}

# Then setup TOTP:
POST /api/v1/auth/admin/totp/setup
# Returns QR code URL — scan with Google Authenticator

# Verify TOTP:
POST /api/v1/auth/admin/totp/verify
{"code": "123456"}
```

### 6. Upload Expert Transcripts

```bash
# Create expert
POST /api/v1/admin/experts
{
    "name": "Arpit Bhiyani — System Design",
    "slug": "arpit-system-design",
    "domain": "system_design",
    "description": "Expert in distributed systems and system design"
}

# Upload transcript
curl -X POST /api/v1/admin/experts/{id}/ingest \
    -H "Authorization: Bearer {admin_token}" \
    -F "transcript=@/path/to/transcript.md"

# Check ingestion status
GET /api/v1/admin/experts/{id}/jobs
```

## API Reference

See `AI_AVENGERS_SYSTEM_ARCHITECTURE.md` for complete API documentation.

## Monitoring

```bash
# View logs
docker-compose logs -f api
docker-compose logs -f ml-sidecar

# Check health
curl http://localhost:8080/health

# Admin stats (requires admin token)
curl -H "Authorization: Bearer {token}" http://localhost:8080/api/v1/admin/stats
```

## Troubleshooting

**ML Sidecar slow to start:**
First startup downloads ~800MB of models. Wait 2-3 minutes.

**Database connection failed:**
Check `DATABASE_URL` in `.env`. Ensure PostgreSQL is running.

**JWT_SECRET too short:**
Must be minimum 32 characters.

**ENCRYPTION_KEY wrong length:**
Must be exactly 32 characters for AES-256.


**Implementation phase fails with `mkdir /workspaces/<id>: permission denied`:**

The `api` container runs as a non-root user and writes the Aider workspaces
into the shared `workspaces` Docker volume. The image now declares
`/workspaces` with that user as owner, but Docker only applies an image
directory's ownership when it creates a **new** empty named volume. A volume
created before this change is still root-owned, so it has to be removed once:

```bash
docker-compose down
docker volume rm ai_avengers_workspaces   # name = <project>_workspaces
docker-compose up -d --build api aider-service
```

This deletes generated workspaces only — Postgres and Redis data live in
separate volumes and are untouched.

Both `api` and `aider-service` deliberately run as UID 1000 because they share
this volume: the Go side creates the workspace and runs `git` in it, and
aider-service edits and commits files in the same directory. If you change the
user in one Dockerfile, change it in the other too.
