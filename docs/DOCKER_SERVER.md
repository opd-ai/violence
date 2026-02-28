# VIOLENCE Dedicated Server - Docker Deployment

## Overview

The VIOLENCE dedicated server runs as a headless authoritative game server for multiplayer sessions. This document covers building and deploying the server using Docker.

## Quick Start

### Build the Docker Image

```bash
docker build -t violence-server:latest .
```

### Run the Server

```bash
docker run -d \
  --name violence-server \
  -p 7777:7777 \
  violence-server:latest
```

## Configuration

### Command-Line Flags

The server accepts the following flags:

- `-port` - Server port to listen on (default: 7777)
- `-log-level` - Logging level: debug, info, warn, error (default: info)

### Custom Configuration

Override default flags when running the container:

```bash
docker run -d \
  --name violence-server \
  -p 8080:8080 \
  violence-server:latest \
  -port=8080 -log-level=debug
```

## Docker Image Details

### Base Image

The server uses a multi-stage build:
- **Builder**: `golang:1.24-alpine` - Compiles the Go binary
- **Runtime**: `gcr.io/distroless/static-debian12` - Minimal runtime image for security

### Image Size

The final image is approximately 15-20 MB (minimal attack surface).

### Security Features

- Runs as non-root user (`nonroot:nonroot`)
- Distroless base image (no shell, package manager, or unnecessary tools)
- Static binary with no external dependencies
- Minimal attack surface

## Deployment Examples

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  violence-server:
    image: violence-server:latest
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "7777:7777"
    restart: unless-stopped
    command: ["-port=7777", "-log-level=info"]
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
```

Run with:

```bash
docker-compose up -d
```

### Kubernetes Deployment

Create `k8s-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: violence-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: violence-server
  template:
    metadata:
      labels:
        app: violence-server
    spec:
      containers:
      - name: violence-server
        image: violence-server:latest
        ports:
        - containerPort: 7777
          protocol: TCP
        args: ["-port=7777", "-log-level=info"]
        resources:
          limits:
            memory: "256Mi"
            cpu: "500m"
          requests:
            memory: "128Mi"
            cpu: "250m"
---
apiVersion: v1
kind: Service
metadata:
  name: violence-server
spec:
  type: LoadBalancer
  ports:
  - port: 7777
    targetPort: 7777
    protocol: TCP
  selector:
    app: violence-server
```

Deploy with:

```bash
kubectl apply -f k8s-deployment.yaml
```

## Publishing to GitHub Container Registry

### Prerequisites

1. GitHub Personal Access Token with `write:packages` permission
2. Docker configured for GHCR

### Build and Tag

```bash
docker build -t ghcr.io/opd-ai/violence-server:latest .
docker tag ghcr.io/opd-ai/violence-server:latest ghcr.io/opd-ai/violence-server:v5.0
```

### Login to GHCR

```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

### Push Image

```bash
docker push ghcr.io/opd-ai/violence-server:latest
docker push ghcr.io/opd-ai/violence-server:v5.0
```

### Pull and Run

```bash
docker pull ghcr.io/opd-ai/violence-server:latest
docker run -d -p 7777:7777 ghcr.io/opd-ai/violence-server:latest
```

## Monitoring and Logs

### View Logs

```bash
docker logs -f violence-server
```

Logs are in JSON format for easy parsing:

```json
{"level":"info","msg":"Starting VIOLENCE dedicated server","port":7777,"log_level":"info"}
{"level":"info","msg":"Server started successfully, waiting for connections..."}
{"level":"info","msg":"Player connected","player_id":1,"system_name":"gameserver"}
```

### Health Check

Connect to the server port to verify it's running:

```bash
nc -zv localhost 7777
```

### Resource Usage

Check container resource usage:

```bash
docker stats violence-server
```

## Troubleshooting

### Port Already in Use

If port 7777 is already bound, use a different port:

```bash
docker run -d -p 8080:7777 violence-server:latest
```

### Connection Refused

1. Check if server is running: `docker ps`
2. Check logs: `docker logs violence-server`
3. Verify port mapping: `docker port violence-server`
4. Check firewall rules

### High Memory Usage

Adjust container memory limits:

```bash
docker run -d \
  --memory=256m \
  --memory-swap=512m \
  -p 7777:7777 \
  violence-server:latest
```

## Performance Tuning

### Server Specs Recommendation

- **CPU**: 1-2 cores per 20 players
- **Memory**: 256 MB base + 16 MB per player
- **Network**: 1 Mbps upload per 10 players

### Tick Rate

Server runs at 20 ticks/second (50ms per tick). This is hardcoded in `pkg/network/gameserver.go`.

### Client Limits

Default max clients: Unlimited (limited by system resources)
Recommended max: 64 players per server

## CI/CD Integration

The Docker image is automatically built by GitHub Actions. See `.github/workflows/build.yml` for the full CI/CD pipeline.

### Automated Builds

Every push to `main` triggers:
1. Docker image build
2. Security scanning
3. Push to GHCR with `latest` tag

Every tagged release (e.g., `v5.0.0`) triggers:
1. Multi-platform build
2. Semantic version tagging
3. Release artifact publishing

## Security Best Practices

1. **Run behind a reverse proxy** (e.g., nginx) for DDoS protection
2. **Use TLS/SSL** for encrypted client connections (requires proxy)
3. **Enable rate limiting** to prevent abuse
4. **Monitor logs** for suspicious activity
5. **Keep image updated** with latest security patches
6. **Limit container resources** to prevent resource exhaustion

## Local Development

### Build Binary Directly

```bash
go build -o violence-server ./cmd/server
./violence-server -port=7777 -log-level=debug
```

### Run Tests

```bash
go test ./cmd/server -v
```

### Test Docker Build

```bash
docker build -t violence-server:dev .
docker run --rm -p 7777:7777 violence-server:dev
```

## See Also

- [BUILD_MATRIX.md](BUILD_MATRIX.md) - Multi-platform build instructions
- [FEDERATION.md](FEDERATION.md) - Federation hub setup
- [NETWORKING.md](NETWORKING.md) - Network protocol details
