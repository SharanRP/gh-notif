# Use a minimal base image for the final stage
FROM alpine:latest AS base

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN adduser -D -s /bin/sh appuser

# Final stage
FROM scratch

# Import from base
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the pre-built binary (provided by GoReleaser)
COPY gh-notif /usr/local/bin/gh-notif

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/gh-notif"]

# Add labels
LABEL org.opencontainers.image.title="gh-notif"
LABEL org.opencontainers.image.description="A high-performance CLI tool for managing GitHub notifications"
LABEL org.opencontainers.image.vendor="gh-notif Contributors"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/SharanRP/gh-notif"
LABEL org.opencontainers.image.documentation="https://github.com/SharanRP/gh-notif/blob/main/README.md"
