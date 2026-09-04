# AI Avengers Backend — Go

Production-grade multi-agent AI platform backend.

## Quick Start

```bash
# Install dependencies
go mod download

# Run migrations
go run cmd/migrate/main.go up

# Start server
go run cmd/server/main.go
```

## Directory Structure

```
backend-go/
├── cmd/
│   ├── server/main.go      # HTTP server entry point
│   └── migrate/main.go     # Database migration runner
├── internal/
│   ├── auth/               # JWT + bcrypt + TOTP
│   ├── config/             # App configuration
│   ├── db/                 # Database connection pool
│   ├── middleware/         # HTTP middleware
│   ├── gateway/            # LLM model gateway
│   ├── ml/                 # ML sidecar client
│   ├── memory/             # L1/L2/L3 memory manager
│   ├── chinawall/          # 4-layer citation enforcement
│   ├── context/            # Context budget manager
│   ├── decision/           # 5-gate decision engine
│   ├── orchestrator/       # Expert orchestration
│   ├── training/           # Expert ingestion pipeline
│   ├── project/            # Project management
│   ├── chat/               # Chat windows + messages
│   ├── repo/               # GitHub/GitLab integration
│   ├── learning/           # Rating + self-learning
│   ├── admin/              # Admin panel handlers
│   └── response/           # Standardized API responses
├── migrations/             # SQL migration files
└── .env.example            # Environment variables template
```
