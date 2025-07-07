# Multi-stage build for production-ready GreenLync API Gateway
# Stage 1: Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build arguments for versioning
ARG VERSION=unknown
ARG COMMIT=unknown
ARG DATE=unknown

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o greenlync-api-gateway \
    cmd/*.go

# Stage 2: Final runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary from builder stage
COPY --from=builder /build/greenlync-api-gateway .

# Copy configuration and template files
COPY --chown=appuser:appgroup --from=builder /build/pkg/authz/model.conf ./pkg/authz/model.conf
COPY --chown=appuser:appgroup --from=builder /build/pkg/smtp/report_template.html ./pkg/smtp/report_template.html
COPY --chown=appuser:appgroup --from=builder /build/pkg/smtp/demo_account_template.html ./pkg/smtp/demo_account_template.html
COPY --chown=appuser:appgroup --from=builder /build/pkg/smtp/change_password_template.html ./pkg/smtp/change_password_template.html

# Create necessary directories
RUN mkdir -p /app/logs /app/reports && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8888

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8888/api/v1/system/monitor/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV TZ=UTC

# Run the application
CMD ["./greenlync-api-gateway"]