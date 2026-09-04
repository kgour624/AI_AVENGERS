FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
# CGO_ENABLED=0: Static binary, no C dependencies
# -ldflags "-s -w": Strip debug info, reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o server ./cmd/server/main.go

# ============================================================
# Final image — minimal
# ============================================================
FROM alpine:3.19

# Install CA certificates for HTTPS calls to OpenRouter
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .

# Non-root user for security
RUN addgroup -S avengers && adduser -S avengers -G avengers
USER avengers

EXPOSE 8080

CMD ["./server"]
