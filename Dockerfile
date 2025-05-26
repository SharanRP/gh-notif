# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go module files first
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o gh-notif \
    ./main.go

# Final stage
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/gh-notif /usr/local/bin/gh-notif

# Copy documentation
COPY --from=builder /app/docs /docs
COPY --from=builder /app/README.md /README.md
COPY --from=builder /app/LICENSE /LICENSE

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/gh-notif"]

# Add labels
LABEL org.opencontainers.image.title="gh-notif"
LABEL org.opencontainers.image.description="A high-performance CLI tool for managing GitHub notifications"
LABEL org.opencontainers.image.vendor="gh-notif Contributors"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/SharanRP/gh-notif"
LABEL org.opencontainers.image.documentation="https://github.com/SharanRP/gh-notif/blob/main/README.md"
