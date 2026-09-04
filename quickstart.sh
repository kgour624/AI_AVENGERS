#!/bin/bash
# AI Avengers — Quick Start Script
# Usage: ./quickstart.sh

set -e

echo "=================================================="
echo "  AI Avengers — Quick Start"
echo "=================================================="

# Check prerequisites
check_command() {
    if ! command -v "$1" &> /dev/null; then
        echo "ERROR: $1 is required but not installed"
        exit 1
    fi
}

check_command docker
check_command docker-compose

# Create .env if not exists
if [ ! -f .env ]; then
    echo "Creating .env from .env.prod.example..."
    cp .env.prod.example .env
    echo ""
    echo "IMPORTANT: Edit .env and set required values:"
    echo "  - DATABASE_URL"
    echo "  - POSTGRES_PASSWORD"
    echo "  - REDIS_PASSWORD"
    echo "  - JWT_SECRET (min 32 chars)"
    echo "  - OPENROUTER_API_KEY"
    echo "  - ENCRYPTION_KEY (exactly 32 chars)"
    echo ""
    echo "Then run this script again."
    exit 0
fi

# Validate required env vars
required_vars=("JWT_SECRET" "OPENROUTER_API_KEY" "ENCRYPTION_KEY" "POSTGRES_PASSWORD")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "ERROR: $var is not set in .env"
        exit 1
    fi
done

echo "Starting services..."
docker-compose up -d

echo "Waiting for PostgreSQL to be ready..."
until docker-compose exec -T postgres pg_isready -U avengers -d ai_avengers 2>/dev/null; do
    echo -n "."
    sleep 2
done
echo " Ready!"

echo "Running database migrations..."
docker-compose exec -T api sh -c "
    cd /app && \
    DATABASE_URL=\$DATABASE_URL \
    go run cmd/migrate/main.go up
" 2>/dev/null || echo "Note: Run migrations manually if needed"

echo ""
echo "=================================================="
echo "  AI Avengers is running!"
echo "=================================================="
echo "  API:        http://localhost:8080"
echo "  Health:     http://localhost:8080/health"
echo "  API Docs:   See AI_AVENGERS_SYSTEM_ARCHITECTURE.md"
echo ""
echo "  Next steps:"
echo "  1. Create admin user (see SETUP.md)"
echo "  2. Upload expert transcripts via admin panel"
echo "  3. Create your first project"
echo "=================================================="
