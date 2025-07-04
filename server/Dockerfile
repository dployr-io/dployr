# Build stage for Go application
FROM golang:1.23.9-alpine AS builder

# Install build dependencies including C compiler for CGO
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install Riverqueue cli
RUN go install github.com/riverqueue/river/cmd/river@latest

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# TODO: REMOBE THIS LINE, Sqlite3 is no longer needed
# Run River migrations during build
RUN mkdir -p /app/data && \
    export DATABASE_URL="sqlite:///app/data/data.sqlite3" && \
    /go/bin/river migrate-up --line main --database-url "$DATABASE_URL"

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata sqlite

# Create app user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy database migrations
COPY --from=builder /app/db ./db

# Copy frontend assets (public directory, not dashboard)
COPY --from=builder /app/public ./public

# Copy .env file 
COPY --from=builder /app/.env* ./

# Copy the migrated database from builder stage
COPY --from=builder /app/data ./data

# Create data directory for SQLite and set permissions
RUN mkdir -p /app/data && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 7879

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:7879/health || exit 1

# Run the application
CMD ["./main"]
