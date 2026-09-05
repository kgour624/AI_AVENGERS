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
