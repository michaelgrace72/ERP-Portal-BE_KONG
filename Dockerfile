# ============================================
# BUILD STAGE
# ============================================
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Tidy dependencies
RUN go mod tidy

# Build binaries with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o main ./cmd/server

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o migrate ./cmd/migrate

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o seed ./cmd/seed

# ============================================
# PRODUCTION STAGE
# ============================================
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    wget \
    curl

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/seed .

# Copy required directories
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/assets ./assets

# Create directories for uploads (if using local storage)
RUN mkdir -p /app/assets/uploads && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Default command (can be overridden in docker-compose)
CMD ["./main"]
