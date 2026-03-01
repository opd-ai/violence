# Federation Hub Server

A standalone server for Violence game server federation and matchmaking.

## Features

- **Server Registry**: Maintains a registry of active game servers with 15-minute TTL
- **Player Lookup**: Find which server a player is currently on across the federation
- **Rate Limiting**: 60 requests per minute per IP (configurable)
- **Optional Authentication**: Protect server registration with bearer tokens
- **Hub-to-Hub Peering**: Sync server lists between multiple federation hubs
- **Health Monitoring**: Health check endpoint for monitoring and load balancing

## Building

```bash
cd cmd/federation-hub
go build -o federation-hub
```

Or from the repository root:

```bash
go build -o federation-hub ./cmd/federation-hub
```

## Running

Basic usage:

```bash
./federation-hub
```

With options:

```bash
./federation-hub \
  -addr=:8080 \
  -auth-token=your-secret-token \
  -peers=http://hub1.example.com:8080,http://hub2.example.com:8080 \
  -log-level=info \
  -rate-limit=60
```

### Command-Line Options

- `-addr`: HTTP server address (default: `:8080`)
- `-auth-token`: Optional authentication token for server registration
- `-peers`: Comma-separated list of peer hub URLs for syncing
- `-log-level`: Log level: debug, info, warn, error (default: `info`)
- `-rate-limit`: Rate limit per IP in requests per minute (default: `60`)

## API Reference

### POST /announce

Register or update a game server.

**Headers:**
- `Authorization: Bearer <token>` (if auth is enabled)

**Request Body:**
```json
{
  "name": "my-server",
  "address": "game.example.com:7777",
  "region": "us-east",
  "genre": "horror",
  "players": 5,
  "maxPlayers": 16,
  "playerList": ["player-id-1", "player-id-2"]
}
```

**Response:**
```json
{
  "status": "registered"
}
```

### POST /query

Query for available servers.

**Request Body:**
```json
{
  "region": "us-east",
  "genre": "horror",
  "minPlayers": 1,
  "maxPlayers": 10
}
```

All fields are optional. Empty query returns all servers.

**Response:**
```json
[
  {
    "name": "server-1",
    "address": "game1.example.com:7777",
    "region": "us-east",
    "genre": "horror",
    "players": 5,
    "maxPlayers": 16,
    "playerList": ["player-1", "player-2"],
    "timestamp": "2026-03-01T12:00:00Z"
  }
]
```

### POST /lookup

Find which server a player is currently on.

**Request Body:**
```json
{
  "playerID": "player-id-123"
}
```

**Response:**
```json
{
  "online": true,
  "serverAddress": "game.example.com:7777",
  "serverName": "my-server"
}
```

### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "ok",
  "version": "6.0.0",
  "uptime": "2h15m30s",
  "serverCount": 42
}
```

### GET /peers

List configured peer hubs.

**Response:**
```json
{
  "peers": [
    "http://hub1.example.com:8080",
    "http://hub2.example.com:8080"
  ]
}
```

## Deployment

### Systemd Service

Create `/etc/systemd/system/federation-hub.service`:

```ini
[Unit]
Description=Violence Federation Hub
After=network.target

[Service]
Type=simple
User=federation
WorkingDirectory=/opt/federation-hub
ExecStart=/opt/federation-hub/federation-hub -addr=:8080 -log-level=info
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable federation-hub
sudo systemctl start federation-hub
```

### Docker

Build image:

```bash
docker build -t violence-federation-hub -f Dockerfile.federation-hub .
```

Run container:

```bash
docker run -d \
  -p 8080:8080 \
  --name federation-hub \
  violence-federation-hub
```

### Docker Compose

```yaml
version: '3.8'
services:
  federation-hub:
    image: ghcr.io/opd-ai/violence-federation-hub:latest
    ports:
      - "8080:8080"
    environment:
      - LOG_LEVEL=info
    command: ["-addr=:8080", "-log-level=info"]
    restart: unless-stopped
```

### Kubernetes

See `deploy/kubernetes/federation-hub.yaml` for a complete deployment configuration.

## Hub Peering

To create a federated network of hubs, configure each hub with peer URLs:

**Hub 1:**
```bash
./federation-hub -addr=:8080 -peers=http://hub2.example.com:8080,http://hub3.example.com:8080
```

**Hub 2:**
```bash
./federation-hub -addr=:8080 -peers=http://hub1.example.com:8080,http://hub3.example.com:8080
```

**Hub 3:**
```bash
./federation-hub -addr=:8080 -peers=http://hub1.example.com:8080,http://hub2.example.com:8080
```

Hubs sync their server registries every 5 minutes, providing redundancy and geographic distribution.

## Server Integration

Game servers should send heartbeats every 60 seconds to maintain their registration:

```go
import "github.com/opd-ai/violence/pkg/federation"

announcer := federation.NewServerAnnouncer(
    "ws://hub.example.com:8080/announce",
    federation.ServerAnnouncement{
        Name:       "my-server",
        Address:    "game.example.com:7777",
        Region:     federation.RegionUSEast,
        Genre:      "horror",
        Players:    0,
        MaxPlayers: 16,
    },
)

if err := announcer.Start(); err != nil {
    log.Fatal(err)
}
defer announcer.Stop()

// Update player count/list as needed
announcer.UpdatePlayerList([]string{"player-1", "player-2"})
```

## Monitoring

The `/health` endpoint can be used with monitoring tools:

**Prometheus:**
```yaml
scrape_configs:
  - job_name: 'federation-hub'
    metrics_path: '/health'
    static_configs:
      - targets: ['hub.example.com:8080']
```

**Uptime monitoring:**
```bash
curl -f http://hub.example.com:8080/health || alert
```

## Performance

- Handles 1000+ registered servers per hub
- Rate limiting prevents abuse
- Hub peering provides horizontal scalability
- 15-minute TTL with 60-second heartbeats balances freshness and network traffic

## License

Same as the Violence project (see root LICENSE file).
