# Build stage
FROM golang:1.22.1-bullseye as builder

LABEL author="Dipjyoti Metia"
LABEL version="1.0"

ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /app

COPY go.* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -a -o ./server ./cmd

# Final stage
FROM debian:buster-slim

WORKDIR /app

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Create a group and user
RUN groupadd -r app && useradd -r -g app app

# Change the ownership of the /app directory to our app user
RUN chown -R app:app /app

# Switch to 'app' user
USER app

COPY --from=builder /app/server /app/server

EXPOSE 8083

ENTRYPOINT ["/app/server"]