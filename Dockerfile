# ============================================================================
# Build Stage
# ============================================================================
FROM golang:1.25.1-alpine AS builder

# Build arguments for versioning
ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=unknown
ARG TARGETARCH

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    file \
    postgresql-client \
    ca-certificates \
    tzdata

WORKDIR /workspace

# Copy proto contracts first (from parent context)
COPY hub-proto-contracts ./hub-proto-contracts

# Copy all modules referenced in go.work
COPY hub-user-service/go.mod hub-user-service/go.sum ./hub-user-service/
COPY hub-api-gateway/go.mod hub-api-gateway/go.sum ./hub-api-gateway/

# Copy go.mod, go.sum and go.work from HubInvestmentsServer
COPY HubInvestmentsServer/go.mod HubInvestmentsServer/go.sum HubInvestmentsServer/go.work ./HubInvestmentsServer/

# Set working directory to service
WORKDIR /workspace/HubInvestmentsServer

# Download dependencies
RUN go mod download

# Copy source code
COPY HubInvestmentsServer/ .

# Build the main application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} \
    go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/hubinvestments \
    ./main.go

# Build the migration tool
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} \
    go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o /app/migrate \
    ./cmd/migrate/main.go

# Verify binaries
RUN ls -lh /app/hubinvestments /app/migrate && file /app/hubinvestments /app/migrate

# ============================================================================
# Runtime Stage
# ============================================================================
FROM alpine:latest

# OCI labels
LABEL org.opencontainers.image.title="Hub Investments Server"
LABEL org.opencontainers.image.description="Main monolith service for Hub Investments Platform"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.revision="${GIT_COMMIT}"

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    wget \
    postgresql-client \
    tzdata

# Create non-root user
RUN addgroup -g 1001 hubuser && \
    adduser -D -u 1001 -G hubuser hubuser

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/hubinvestments .
COPY --from=builder /app/migrate .

# Copy database migrations
COPY --from=builder /workspace/HubInvestmentsServer/database ./database
COPY --from=builder /workspace/HubInvestmentsServer/shared/infra/migration/sql ./shared/infra/migration/sql

# Create necessary directories
RUN mkdir -p /app/logs /app/tmp

# Change ownership
RUN chown -R hubuser:hubuser /app

# Switch to non-root user
USER hubuser

# Expose ports
EXPOSE 8080 50060

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=15s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./hubinvestments"]

