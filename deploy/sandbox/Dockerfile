# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application with security flags
# -ldflags="-s -w" to strip debug info
# CGO_ENABLED=0 for a static binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o refressh ./cmd/refressh/main.go

# Runtime Stage
FROM alpine:3.19

# Security: Add a non-root user
RUN addgroup -S refressh && adduser -S refressh -G refressh

# Install basic tools for a terminal environment
# We include bash, curl, and git as common tools for AI agents
RUN apk add --no-cache \
    bash \
    curl \
    git \
    openssh-client \
    ca-certificates

# Create refreSSH configuration directory and set ownership
RUN mkdir -p /home/refressh/.refreSSH && \
    chown -R refressh:refressh /home/refressh/.refreSSH

# Create home directory for refressh
WORKDIR /home/refressh

# Copy the binary from the builder stage
COPY --from=builder /app/refressh /usr/local/bin/refressh

# Set ownership
RUN chown refressh:refressh /usr/local/bin/refressh

# Use the non-root user
USER refressh

# Expose the API port (local-only by default)
EXPOSE 9000

# Set the entrypoint to start the daemon
# We use the 'daemon start' command as the default
ENTRYPOINT ["refressh", "daemon", "start"]
