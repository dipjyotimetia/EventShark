# Build stage
FROM golang:1.24-bullseye as builder

LABEL org.opencontainers.image.authors="Dipjyoti Metia"
LABEL org.opencontainers.image.version="1.0"
LABEL org.opencontainers.image.description="Event Stream Service"

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

# Copy only files needed for dependency download to leverage caching
COPY go.* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the source code
COPY . .

# Build with specific security flags
RUN --mount=type=cache,target=/go/pkg/mod \
    go build -ldflags="-s -w" -a -o ./server ./cmd

# Final stage - using distroless for minimal attack surface
FROM gcr.io/distroless/static-debian11

WORKDIR /app

# Copy only the compiled binary
COPY --from=builder /app/server /app/server
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Use non-root user with specific UID/GID
USER nonroot:nonroot

EXPOSE 8083

# Add health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 CMD ["/app/server", "health"] || exit 1

# Use ENTRYPOINT for fixed command
ENTRYPOINT ["/app/server"]