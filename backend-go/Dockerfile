FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git — needed by go mod tidy to fetch some dependencies
RUN apk --no-cache add git

# Copy go.mod first (layer cache: only re-download if go.mod changes)
COPY go.mod ./
COPY go.sum* ./

# Copy ALL source files before go mod tidy
# WHY: go mod tidy reads .go files to find actual imports.
# Without source, tidy resolves nothing and go.sum stays empty.
COPY . .

# Now tidy can see all imports and generate complete go.sum
RUN go mod tidy

# Build both binaries in one layer
# CGO_ENABLED=0: Static binary, no C dependencies
# -ldflags "-s -w": Strip debug info, reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o server ./cmd/server/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o migrate ./cmd/migrate/main.go

# ============================================================
# Final image — minimal
# ============================================================
FROM alpine:3.19

# Runtime tools:
#   ca-certificates — HTTPS calls to the LLM providers
#   tzdata          — timestamps
#   git             — AiderRunner.initWorkspace runs git init / config / commit,
#                     and WorkspaceMerger runs git diff / add / commit inside
#                     the per-expert workspaces. Without git in THIS image the
#                     implementation phase dies with
#                     "exec: git: executable file not found in $PATH".
#   rsync           — seedWorkspace copies main/ into an expert workspace and
#                     WorkspaceMerger merges expert workspaces back into main/,
#                     both via rsync.
#   go              — the implementation and QA phases verify generated code by
#                     building and testing it (internal/workflow/verify.go), and
#                     the validation pipeline runs "go vet" and "gofmt"
#                     (internal/validation/go_validator.go). Without the Go
#                     toolchain every one of those calls failed with
#                     "executable file not found", which the old code could not
#                     tell apart from a compile error: taskComplete could never
#                     become true, so every task burned all 5 iterations, and
#                     publishCodeArtifacts returned early on the "build failure"
#                     so no generated code ever reached the blackboard.
#   nodejs, npm     — same story for the frontend expert. tsc/npm are what
#                     verify a React workspace; "go build" in one is meaningless.
#
# COST: these two add roughly 300 MB to the image. Removing them does not break
# the build — verification degrades to "could not be verified", which is
# reported honestly and no longer blocks publishing — but then nothing the
# experts write is ever actually compiled or tested.
#
# RISK, stated plainly: "go test" and "npm run test" execute model-generated
# code inside this container as the non-root user. The validation sandbox
# (internal/validation/sandbox.go) applies ulimits and timeouts; verify.go
# currently relies only on the caller's context timeout. Moving these runs into
# a dedicated sandboxed service is the right long-term shape.
RUN apk --no-cache add ca-certificates tzdata git rsync go nodejs npm

# Caches for a non-root user.
#
# The Go toolchain refuses to run without a writable build cache and would fail
# with "failed to initialize build cache ... permission denied", because the
# default ($HOME/.cache/go-build) is not writable for the user below. npm has
# the same problem with its cache directory. /tmp is writable by everyone and a
# cache is meant to be throwaway, so it is the right home for both.
ENV GOCACHE=/tmp/gocache \
    GOPATH=/tmp/go \
    GOTMPDIR=/tmp \
    GOFLAGS=-mod=mod \
    npm_config_cache=/tmp/npm

WORKDIR /app

# Copy both binaries
COPY --from=builder /app/server .
COPY --from=builder /app/migrate .

# Copy migrations directory so migrate binary can find them
COPY migrations/ ./migrations/

# Non-root user for security.
#
# UID 1000 on purpose: aider-service's Dockerfile creates its user with
# "useradd -u 1000 aider", and both containers mount the SAME "workspaces"
# volume. The Go side creates /workspaces/{workflow}/{expert} and runs git
# there; aider-service then edits and commits files in that same directory.
# Different UIDs would make one of them unable to write the other's files.
RUN addgroup -S -g 1000 avengers && adduser -S -u 1000 -G avengers avengers

# Create the workspace root and hand it to that user BEFORE switching to it.
#
# WHY this is required: the api container runs as a non-root user, while a
# fresh Docker named volume is root-owned. Docker copies the image path's
# ownership into a NEW empty named volume, so declaring the directory here
# with the right owner is what makes the mounted volume writable. Without
# it the implementation phase failed on its very first step with
# "init workspace: mkdir /workspaces/<workflow-id>: permission denied".
# aider-service's Dockerfile already did exactly this; the api image did not.
#
# NOTE: this only fixes volumes created after this change. An existing
# root-owned workspaces volume must be removed once (see SETUP.md).
RUN mkdir -p /workspaces && chown avengers:avengers /workspaces

USER avengers

EXPOSE 8080

CMD ["./server"]
