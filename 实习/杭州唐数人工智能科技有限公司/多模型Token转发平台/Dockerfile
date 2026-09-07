# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.25-alpine AS builder

# Module proxy; override with --build-arg GOPROXY=https://proxy.golang.org,direct
# in network environments where goproxy.cn is unsuitable.
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

WORKDIR /app

# Install git and build tools for module downloads and static build
RUN apk add --no-cache git ca-certificates tzdata

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build both the server and the migration command.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/migrate ./cmd/migrate && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/admin ./cmd/admin

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -g '' appuser

WORKDIR /app

# Migrations are embedded in the image so the migrate binary can run without
# a host volume mount.
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/bin/migrate /app/migrate
COPY --from=builder /app/bin/admin /app/admin

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/server"]
