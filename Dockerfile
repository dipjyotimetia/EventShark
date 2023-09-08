FROM golang:1.21 as builder

ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /app

COPY go.* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -a -o ./server ./cmd

FROM debian:buster-slim
WORKDIR /app
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/server /app/server

EXPOSE 9050

ENTRYPOINT ["/app/server"]
