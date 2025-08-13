# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Ensure deterministic, small, static binary
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Install git for go modules that may need it
RUN apk add --no-cache git ca-certificates

# Cache module downloads
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the source
COPY . .

# Build the API binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags "-s -w" -o /app/bin/api ./cmd/api


# ---- Final stage ----
# Use a minimal image that includes CA certificates
FROM gcr.io/distroless/base-debian11:nonroot

WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /app/bin/api /app/api

# Default port
ENV PORT=8080
EXPOSE 8080

# Run as nonroot
USER nonroot:nonroot

ENTRYPOINT ["/app/api"]


