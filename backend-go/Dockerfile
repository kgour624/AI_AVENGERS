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
RUN apk --no-cache add ca-certificates tzdata git rsync

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
