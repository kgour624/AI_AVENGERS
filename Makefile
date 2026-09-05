# AI Avengers — Build Automation
# Usage:
#   make generate          — regenerate types from openapi.yaml (run after any spec change)
#   make verify-contract   — check generated files are up to date (run in CI)
#   make dev               — start all services
#   make build             — build all Docker images

.PHONY: generate verify-contract dev build help

OPENAPI_SPEC := backend-go/api/openapi.yaml
GO_GENERATED := backend-go/internal/api/generated/types.go
TS_GENERATED := frontend/src/types/generated.ts

# ============================================================
# generate — Single command to sync all types from openapi.yaml
#
# WHY both generators in one command:
# If you run only one, the other side drifts.
# One command = both sides always in sync.
# ============================================================
generate: generate-go generate-ts
	@echo ""
	@echo "Contract generation complete."
	@echo "Both Go and TypeScript types are now in sync with openapi.yaml."

generate-go:
	@echo "Generating Go types from openapi.yaml..."
	@mkdir -p backend-go/internal/api/generated
	@cd backend-go && go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.3.0 \
		--config api/oapi-codegen.yaml \
		api/openapi.yaml \
		> internal/api/generated/types.go
	@echo "Go types generated: $(GO_GENERATED)"

generate-ts:
	@echo "Generating TypeScript types from openapi.yaml..."
	@cd frontend && npx openapi-typescript ../$(OPENAPI_SPEC) \
		--output src/types/generated.ts
	@echo "TypeScript types generated: $(TS_GENERATED)"

# ============================================================
# verify-contract — CI check: generated files must be up to date
#
# How it works:
# 1. Run generate into temp files
# 2. Diff temp vs committed
# 3. If diff exists -> someone changed openapi.yaml without running generate
# 4. Fail CI -> force developer to run 'make generate' and commit
# ============================================================
verify-contract:
	@echo "Verifying contract is up to date..."
	@cp $(GO_GENERATED) /tmp/types_go_current.go
	@cp $(TS_GENERATED) /tmp/types_ts_current.ts
	@$(MAKE) generate > /dev/null 2>&1
	@if ! diff -q $(GO_GENERATED) /tmp/types_go_current.go > /dev/null 2>&1; then \
		echo "ERROR: Go types are out of date. Run 'make generate' and commit."; \
		cp /tmp/types_go_current.go $(GO_GENERATED); \
		exit 1; \
	fi
	@if ! diff -q $(TS_GENERATED) /tmp/types_ts_current.ts > /dev/null 2>&1; then \
		echo "ERROR: TypeScript types are out of date. Run 'make generate' and commit."; \
		cp /tmp/types_ts_current.ts $(TS_GENERATED); \
		exit 1; \
	fi
	@echo "Contract is up to date."

dev:
	docker-compose up -d

build:
	docker-compose build

help:
	@echo "Available commands:"
	@echo "  make generate        — Regenerate types from openapi.yaml"
	@echo "  make verify-contract — Check generated files are up to date (CI)"
	@echo "  make dev             — Start all services"
	@echo "  make build           — Build Docker images"
