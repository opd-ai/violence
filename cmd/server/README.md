# VIOLENCE Dedicated Server

Headless authoritative game server for VIOLENCE multiplayer sessions.

## Quick Start

```bash
# Build
go build -o violence-server .

# Run
./violence-server -port=7777 -log-level=info
```

## Flags

- `-port` - Server port (default: 7777)
- `-log-level` - Log level: debug, info, warn, error (default: info)

## Docker

See [docs/DOCKER_SERVER.md](../../docs/DOCKER_SERVER.md) for Docker deployment.

## Features

- Authoritative server with 20 tick/second update loop
- Client connection management
- Command validation
- JSON-formatted logs for monitoring
- Graceful shutdown on SIGINT/SIGTERM
- Minimal resource footprint (~4MB binary)

## Testing

```bash
go test -v
```

## Architecture

- Uses `pkg/network.GameServer` for core server logic
- Uses `pkg/engine.World` for game state management
- JSON over TCP for client communication
- Logrus for structured logging
