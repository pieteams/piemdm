# Multi-stage build for Go application
FROM golang:1.24.12-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory to root
WORKDIR /workspace

# Copy go.work and module files
COPY go.work go.work.sum* ./
COPY backend/go.mod backend/go.sum ./backend/
COPY packages/go/openapi/go.mod packages/go/openapi/go.work.sum* ./packages/go/openapi/

# Download dependencies
# Need to be in the respective module directories or use go work sync
RUN go mod download -C backend && go mod download -C packages/go/openapi

# Copy source code
COPY backend ./backend
COPY packages/go/openapi ./packages/go/openapi
COPY shared ./shared

# Build the application
WORKDIR /workspace/backend

# Build the application
# Use automatic architecture detection or provided ARG
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
    -ldflags="-s -w" \
    -o ./bin/server ./cmd/server/

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Set timezone to Shanghai
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# Create app user
RUN addgroup -g 1001 -S app && \
    adduser -u 1001 -S app -G app

# Set working directory
WORKDIR /app

# Copy binary and config from builder
COPY --from=builder /app/bin/server .
COPY --from=builder /app/config ./config

# Create necessary directories
RUN mkdir -p storage/logs export uploads && \
    chown -R app:app /app

# Switch to app user
USER app

# Expose port
EXPOSE 8787

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8787/health || exit 1

# Run the application
ENTRYPOINT ["./server"]
