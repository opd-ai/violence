# Federation Hub Protocol

## Overview

Violence supports decentralized multiplayer through **self-hosted federation hubs**. These hubs enable server discovery, matchmaking, and lobby management without relying on centralized infrastructure. This document specifies the federation hub protocol, self-hosting instructions, and optional DHT discovery.

## Architecture

### Federation Model

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│   Player    │◄───────►│ Game Server │◄───────►│ Federation  │
│   Client    │         │  (Self-Host)│         │     Hub     │
└─────────────┘         └─────────────┘         └─────────────┘
                               │                       │
                               │                       │
                               ▼                       ▼
                        ┌─────────────┐         ┌─────────────┐
                        │  Other Game │         │ Other Hubs  │
                        │   Servers   │         │ (Peering)   │
                        └─────────────┘         └─────────────┘
```

**Key Principles**:
- **No Single Point of Failure**: Hubs are peer-to-peer; any hub can discover others
- **Self-Sovereignty**: Anyone can run a hub; no central authority
- **Privacy-First**: Players control data exposure; opt-in telemetry only
- **Lightweight**: Hub runs on minimal resources (Raspberry Pi, VPS, homelab)

### Hub Responsibilities

1. **Server Registry**: Maintain list of active game servers
2. **Health Checks**: Periodically ping servers to remove stale entries
3. **Matchmaking**: Return server list matching player query (region, mode, skill)
4. **Hub Peering**: Discover and sync with other federation hubs
5. **Telemetry** (Optional): Aggregate anonymous stats (player count, regions)

### Hub Does NOT Handle

- **Game State**: All gameplay logic stays on game servers
- **Chat Relay**: Direct peer-to-peer or server-relayed chat
- **Authentication**: Optional; servers may enforce their own auth
- **Hosting Games**: Hub only tracks servers; doesn't host matches

## Hub API Specification

### Protocol: HTTP/JSON over TCP

**Base Endpoint**: `http://<hub-address>:<port>/api/v1/`

**Authentication**: Optional Bearer token for write operations (server registration)

### Endpoints

#### 1. Register Game Server

**Request**:
```http
POST /api/v1/servers
Content-Type: application/json
Authorization: Bearer <optional-token>

{
  "server_id": "uuid-or-generated-hash",
  "address": "192.168.1.100:7654",
  "public_address": "203.0.113.42:7654",
  "name": "Alice's Cyberpunk Deathmatch",
  "genre": "cyberpunk",
  "mode": "deathmatch",
  "max_players": 8,
  "current_players": 3,
  "region": "us-west",
  "password_protected": false,
  "version": "5.0.0",
  "tags": ["modded", "hardcore"],
  "metadata": {
    "map_seed": 42,
    "difficulty": "nightmare"
  }
}
```

**Response**:
```http
HTTP/1.1 201 Created
Content-Type: application/json

{
  "server_id": "uuid-or-generated-hash",
  "registered_at": "2026-03-01T00:52:17Z",
  "expires_at": "2026-03-01T01:07:17Z",
  "heartbeat_interval": 60
}
```

**Notes**:
- `address`: Private IP (for LAN discovery)
- `public_address`: Public IP + port (for internet play; NAT traversal required)
- `expires_at`: Server must send heartbeat before expiration (15 min default)
- `heartbeat_interval`: Recommended seconds between heartbeats (60s default)

#### 2. Heartbeat (Keep-Alive)

**Request**:
```http
PUT /api/v1/servers/{server_id}/heartbeat
Content-Type: application/json
Authorization: Bearer <optional-token>

{
  "current_players": 4,
  "status": "active"
}
```

**Response**:
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "expires_at": "2026-03-01T01:22:17Z"
}
```

**Notes**:
- Updates `current_players` and extends expiration
- Servers should send heartbeat every 60 seconds
- Status values: `active`, `full`, `starting`, `ending`

#### 3. Unregister Server

**Request**:
```http
DELETE /api/v1/servers/{server_id}
Authorization: Bearer <optional-token>
```

**Response**:
```http
HTTP/1.1 204 No Content
```

**Notes**:
- Graceful shutdown; removes server from registry immediately
- Optional: If server crashes, registry auto-removes after `expires_at`

#### 4. Query Servers

**Request**:
```http
GET /api/v1/servers?genre=cyberpunk&mode=deathmatch&region=us-west&available=true
```

**Query Parameters**:
- `genre`: Filter by genre (fantasy, scifi, horror, cyberpunk, postapoc)
- `mode`: Filter by mode (coop, deathmatch, team_dm, ctf, etc.)
- `region`: Filter by region (us-west, us-east, eu-west, ap-southeast, etc.)
- `available`: `true` = only servers with open slots
- `password_protected`: `false` = exclude password servers
- `version`: Filter by game version (exact match)
- `tags`: Comma-separated (e.g., `tags=modded,hardcore`)
- `max_ping`: Client-side hint (hub may ignore)
- `limit`: Max results (default: 50, max: 200)
- `offset`: Pagination offset

**Response**:
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "servers": [
    {
      "server_id": "uuid-1",
      "address": "192.168.1.100:7654",
      "public_address": "203.0.113.42:7654",
      "name": "Alice's Cyberpunk Deathmatch",
      "genre": "cyberpunk",
      "mode": "deathmatch",
      "max_players": 8,
      "current_players": 4,
      "region": "us-west",
      "password_protected": false,
      "version": "5.0.0",
      "tags": ["modded", "hardcore"],
      "registered_at": "2026-03-01T00:30:00Z",
      "last_heartbeat": "2026-03-01T00:51:45Z"
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

#### 5. Get Server Details

**Request**:
```http
GET /api/v1/servers/{server_id}
```

**Response**:
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "server_id": "uuid-1",
  "address": "192.168.1.100:7654",
  "public_address": "203.0.113.42:7654",
  "name": "Alice's Cyberpunk Deathmatch",
  "genre": "cyberpunk",
  "mode": "deathmatch",
  "max_players": 8,
  "current_players": 4,
  "region": "us-west",
  "password_protected": false,
  "version": "5.0.0",
  "tags": ["modded", "hardcore"],
  "metadata": {
    "map_seed": 42,
    "difficulty": "nightmare"
  },
  "registered_at": "2026-03-01T00:30:00Z",
  "last_heartbeat": "2026-03-01T00:51:45Z",
  "player_list": [
    {"name": "Alice", "score": 15},
    {"name": "Bob", "score": 12},
    {"name": "Charlie", "score": 8},
    {"name": "Diana", "score": 5}
  ]
}
```

**Notes**:
- `player_list` is optional; servers may opt out for privacy

#### 6. Hub Peering: Discover Other Hubs

**Request**:
```http
GET /api/v1/hubs
```

**Response**:
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "hubs": [
    {
      "hub_id": "hub-uuid-1",
      "address": "https://federation-hub-eu.example.com",
      "region": "eu-west",
      "server_count": 42,
      "last_seen": "2026-03-01T00:50:00Z"
    },
    {
      "hub_id": "hub-uuid-2",
      "address": "http://192.168.1.200:8080",
      "region": "us-west",
      "server_count": 18,
      "last_seen": "2026-03-01T00:51:30Z"
    }
  ]
}
```

**Notes**:
- Hubs sync peer lists every 5 minutes
- Clients can query multiple hubs for broader server discovery

#### 7. Hub Registration (Peer-to-Peer)

**Request**:
```http
POST /api/v1/hubs
Content-Type: application/json

{
  "hub_id": "hub-uuid-new",
  "address": "https://my-hub.example.com",
  "region": "ap-southeast",
  "public_key": "base64-encoded-ed25519-public-key"
}
```

**Response**:
```http
HTTP/1.1 201 Created
Content-Type: application/json

{
  "hub_id": "hub-uuid-new",
  "registered_at": "2026-03-01T00:52:17Z"
}
```

**Notes**:
- Hubs announce themselves to known peers
- `public_key` enables signature verification (optional security layer)

## Self-Hosting Instructions

### Prerequisites

- **Operating System**: Linux (Ubuntu 22.04+), macOS, Windows Server
- **Hardware**: 512MB RAM, 1 CPU core, 10GB disk (minimal); 2GB RAM + 2 cores recommended
- **Network**: Static IP or DDNS, port forwarding (default: 8080)
- **Software**: Go 1.21+ or Docker

### Installation Methods

#### Method 1: Docker (Recommended)

```bash
# Pull official hub image
docker pull ghcr.io/opd-ai/violence-federation-hub:latest

# Run hub on port 8080
docker run -d \
  --name violence-hub \
  -p 8080:8080 \
  -v /var/violence-hub:/data \
  -e HUB_REGION=us-west \
  -e HUB_AUTH_TOKEN=your-secret-token \
  ghcr.io/opd-ai/violence-federation-hub:latest

# View logs
docker logs -f violence-hub
```

#### Method 2: Binary Release

```bash
# Download latest release
wget https://github.com/opd-ai/violence/releases/download/v5.0.0/violence-hub-linux-amd64

# Make executable
chmod +x violence-hub-linux-amd64

# Run hub
./violence-hub-linux-amd64 \
  --port 8080 \
  --region us-west \
  --data-dir /var/violence-hub \
  --auth-token your-secret-token
```

#### Method 3: Build from Source

```bash
# Clone repository
git clone https://github.com/opd-ai/violence.git
cd violence

# Build hub binary
go build -o violence-hub ./cmd/federation-hub

# Run hub
./violence-hub --port 8080 --region us-west --data-dir ./hub-data
```

### Configuration

**Config File**: `config.toml` (optional; CLI flags override)

```toml
[hub]
port = 8080
region = "us-west"
data_dir = "/var/violence-hub"
auth_token = "your-secret-token"  # For server registration
max_servers = 1000
heartbeat_timeout = 900  # 15 minutes

[peering]
enabled = true
discovery_interval = 300  # 5 minutes
bootstrap_hubs = [
  "https://hub1.violence-game.org",
  "https://hub2.violence-game.org"
]

[security]
rate_limit_enabled = true
rate_limit_requests_per_minute = 60
require_auth = false  # Set true to require Bearer token

[telemetry]
enabled = false  # Opt-in anonymous stats
endpoint = "https://telemetry.violence-game.org"
```

### Network Setup

#### Port Forwarding

1. Log in to router admin panel
2. Forward external port 8080 → internal IP port 8080 (TCP)
3. Test with `curl http://<your-public-ip>:8080/api/v1/servers`

#### Firewall Rules

```bash
# Ubuntu/Debian (ufw)
sudo ufw allow 8080/tcp

# CentOS/RHEL (firewalld)
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload

# Windows Firewall (PowerShell)
New-NetFirewallRule -DisplayName "Violence Hub" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
```

#### HTTPS (Optional but Recommended)

```bash
# Use Let's Encrypt with Certbot
sudo certbot certonly --standalone -d your-hub-domain.com

# Run hub with TLS
./violence-hub \
  --port 443 \
  --tls-cert /etc/letsencrypt/live/your-hub-domain.com/fullchain.pem \
  --tls-key /etc/letsencrypt/live/your-hub-domain.com/privkey.pem
```

### Testing

```bash
# Check hub health
curl http://localhost:8080/api/v1/health

# Expected response:
# {"status": "ok", "version": "5.0.0", "uptime": 3600}

# Query servers (should be empty initially)
curl http://localhost:8080/api/v1/servers

# Expected response:
# {"servers": [], "total": 0, "limit": 50, "offset": 0}
```

### Maintenance

#### Backup

```bash
# Hub data is stored in data_dir (default: /var/violence-hub)
# Backup entire directory
tar -czf violence-hub-backup-$(date +%Y%m%d).tar.gz /var/violence-hub

# Restore
tar -xzf violence-hub-backup-20260301.tar.gz -C /
```

#### Updates

```bash
# Docker
docker pull ghcr.io/opd-ai/violence-federation-hub:latest
docker stop violence-hub
docker rm violence-hub
# Re-run docker run command from installation

# Binary
wget https://github.com/opd-ai/violence/releases/download/v5.1.0/violence-hub-linux-amd64
sudo systemctl restart violence-hub
```

#### Monitoring

```bash
# Systemd service (create /etc/systemd/system/violence-hub.service)
[Unit]
Description=Violence Federation Hub
After=network.target

[Service]
Type=simple
User=violence
ExecStart=/usr/local/bin/violence-hub --port 8080 --region us-west
Restart=on-failure

[Install]
WantedBy=multi-user.target

# Enable and start
sudo systemctl enable violence-hub
sudo systemctl start violence-hub

# Check status
sudo systemctl status violence-hub

# View logs
journalctl -u violence-hub -f
```

## DHT Discovery (Optional Alternative)

For truly decentralized operation, hubs can use **Distributed Hash Table (DHT)** discovery instead of bootstrap hub lists.

### DHT Protocol: Kademlia-Based

**Implementation**: Use `github.com/libp2p/go-libp2p-kad-dht`

**Concept**:
- Each hub generates a unique peer ID (derived from public key)
- Hubs join DHT by connecting to any known peer
- Game servers register with nearest hubs (by XOR distance in DHT keyspace)
- Clients query DHT for servers matching genre/mode/region

### DHT Bootstrap

**Public Bootstrap Nodes** (run by community):
```
/dnsaddr/bootstrap.violence-game.org/p2p/QmBootstrapPeerID1
/ip4/198.51.100.42/tcp/4001/p2p/QmBootstrapPeerID2
```

**Configuration**:
```toml
[dht]
enabled = true
bootstrap_peers = [
  "/dnsaddr/bootstrap.violence-game.org/p2p/QmBootstrapPeerID1",
  "/ip4/198.51.100.42/tcp/4001/p2p/QmBootstrapPeerID2"
]
listen_addr = "/ip4/0.0.0.0/tcp/4001"
```

### DHT Server Registration

Instead of `POST /api/v1/servers`, servers publish to DHT:

```go
// Pseudo-code
dhtKey := "/violence/servers/" + genre + "/" + mode + "/" + region
dhtValue := marshalJSON(serverInfo)
dht.PutValue(dhtKey, dhtValue, ttl=15*time.Minute)
```

### DHT Client Query

```go
// Client queries DHT
dhtKey := "/violence/servers/cyberpunk/deathmatch/us-west"
values := dht.GetValues(dhtKey, maxResults=50)
servers := unmarshallServers(values)
```

### DHT Advantages

- **No bootstrap hubs needed** after initial peer discovery
- **Censorship resistance**: No central hub to block
- **Global scale**: DHT automatically distributes load

### DHT Disadvantages

- **Complexity**: Harder to debug than HTTP API
- **NAT Traversal**: Requires STUN/TURN for residential routers
- **Bootstrap Dependency**: Still needs initial peers (can use DNS seeds)

### Recommendation

- **Default**: HTTP hub federation (easier to self-host)
- **Advanced**: Enable DHT as optional fallback for censorship-resistant regions

## Security Considerations

### Server Registration Abuse

**Problem**: Malicious actors spam fake servers

**Mitigation**:
1. **Rate Limiting**: Max 10 registrations per IP per hour
2. **Auth Tokens**: Require Bearer token for registration (hub operator issues)
3. **Proof of Work**: Server must solve computational puzzle before registration
4. **IP Reputation**: Block known VPN/proxy IPs (optional; may harm legitimate users)

### Hub Spoofing

**Problem**: Fake hub returns malicious server list

**Mitigation**:
1. **HTTPS**: Require TLS for hub communication
2. **Hub Signatures**: Hubs sign responses with Ed25519 keys
3. **Client Verification**: Clients maintain trusted hub list
4. **DHT Verification**: Use DHT content addressing (hash-based keys)

### DDoS Protection

**Problem**: Hub overwhelmed with queries

**Mitigation**:
1. **Cloudflare**: Put hub behind CDN with DDoS protection
2. **Rate Limiting**: Max 60 requests/min per IP
3. **Caching**: Cache server list for 30s, serve from memory
4. **Geo-Blocking**: Block regions with no servers (optional)

### Privacy

**Problem**: Hubs log player IPs and behavior

**Mitigation**:
1. **No Logging**: Hubs store minimal data (servers only, not queries)
2. **VPN-Friendly**: Don't block VPN/Tor IPs
3. **Anonymized Telemetry**: If enabled, aggregate by region, not IP
4. **GDPR Compliance**: Allow users to request data deletion

## Roadmap

### Phase 1: HTTP Federation (v5.0) ✅
- Basic hub API (register, heartbeat, query)
- Docker deployment
- Self-hosting documentation

### Phase 2: Hub Peering (v5.1)
- Hub discovery and sync
- Multi-hub queries from clients
- Hub health monitoring

### Phase 3: DHT Discovery (v5.2)
- LibP2P integration
- DHT bootstrap nodes
- Hybrid HTTP + DHT mode

### Phase 4: Advanced Features (v6.0)
- Mod distribution via hubs
- Server reputation system
- Matchmaking algorithms (skill-based)

## Code Structure

```
cmd/federation-hub/
├── main.go              // Hub server entry point
├── handlers.go          // HTTP endpoint handlers
├── registry.go          // Server registry logic
├── peering.go           // Hub-to-hub sync
└── dht.go               // Optional DHT integration

pkg/federation/
├── client.go            // Client SDK for querying hubs
├── server.go            // Server SDK for registration
├── types.go             // Shared types (ServerInfo, etc.)
└── federation_test.go   // Unit tests
```

## Conclusion

The Violence federation hub protocol enables decentralized multiplayer without corporate servers. Self-hosting is simple (Docker one-liner), and the HTTP API is human-readable. For advanced users, DHT provides censorship resistance. This architecture ensures the game remains playable even if official infrastructure disappears.

**Next Steps**:
1. Implement `cmd/federation-hub/main.go` HTTP server
2. Create Docker image and publish to GitHub Container Registry
3. Deploy official bootstrap hubs at `hub1.violence-game.org` and `hub2.violence-game.org`
4. Write integration tests for hub peering
5. Document community hub operators in `docs/COMMUNITY_HUBS.md`
