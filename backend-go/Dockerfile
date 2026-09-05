FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git — needed by go mod tidy to fetch some dependencies
RUN apk --no-cache add git

# Copy go.mod first (go.sum may not exist yet)
# WHY go mod tidy not go mod download:
# go mod download requires go.sum to already exist.
# go mod tidy generates go.sum from go.mod — works even when
# go.sum is missing. Safe to run when go.sum exists too.
COPY go.mod ./
COPY go.sum* ./
RUN go mod tidy

# Copy source
COPY . .

# Build both binaries in one layer
# CGO_ENABLED=0: Static binary, no C dependencies
# -ldflags "-s -w": Strip debug info, reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o server ./cmd/server/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o migrate ./cmd/migrate/main.go

# ============================================================
# Final image — minimal
# ============================================================
FROM alpine:3.19

# Install CA certificates for HTTPS calls to OpenRouter
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy both binaries
COPY --from=builder /app/server .
COPY --from=builder /app/migrate .

# Copy migrations directory so migrate binary can find them
COPY migrations/ ./migrations/

# Non-root user for security
RUN addgroup -S avengers && adduser -S avengers -G avengers
USER avengers

EXPOSE 8080

CMD ["./server"]
