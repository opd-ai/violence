# Multi-stage build for minimal server image
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the dedicated server binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /violence-server \
    ./cmd/server

# Runtime stage - minimal distroless image
FROM gcr.io/distroless/static-debian12:latest

# Copy binary from builder
COPY --from=builder /violence-server /violence-server

# Expose default server port
EXPOSE 7777

# Run as non-root user
USER nonroot:nonroot

# Set entrypoint
ENTRYPOINT ["/violence-server"]

# Default flags (can be overridden)
CMD ["-port=7777", "-log-level=info"]
