# DHT Federation Package

Decentralized server discovery using LibP2P Kademlia DHT for the Violence game.

## Overview

This package provides peer-to-peer server discovery without requiring a central federation hub. Game servers can announce themselves to a distributed hash table (DHT), and clients can query for available servers by genre, region, or other criteria.

## Features

- **Decentralized Discovery**: No single point of failure - servers are discovered through a distributed network
- **8-Hour TTL**: Server records automatically expire after 8 hours
- **Minimum 3 Replicas**: Each record is replicated across at least 3 DHT nodes
- **Genre Indexing**: Efficient server lookups by game genre
- **Bootstrap Support**: Connect to known bootstrap peers for network entry
- **Custom Protocol**: Uses `/violence` protocol prefix to avoid IPFS network conflicts

## Usage

### Creating a DHT Node

```go
import (
    "context"
    "github.com/opd-ai/violence/pkg/federation/dht"
)

// Server mode (participates in DHT routing)
node, err := dht.NewNode(context.Background(), dht.Config{
    ListenAddrs:    []string{"/ip4/0.0.0.0/tcp/4001"},
    BootstrapPeers: []string{
        "/ip4/bootstrap.example.com/tcp/4001/p2p/12D3Koo...",
    },
    Mode: "server",
})
if err != nil {
    log.Fatal(err)
}
defer node.Close()
```

### Announcing a Server

```go
record := dht.ServerRecord{
    Name:       "my-scifi-server",
    Address:    "203.0.113.5:7777",
    Genre:      "scifi",
    MaxPlayers: 32,
    Uptime:     3600, // seconds
}

ctx := context.Background()
if err := node.AnnounceServer(ctx, record); err != nil {
    log.Fatalf("Failed to announce: %v", err)
}

// Update genre index for efficient searches
if err := node.UpdateGenreIndex(ctx, "scifi", "my-scifi-server", true); err != nil {
    log.Fatalf("Failed to update index: %v", err)
}
```

### Looking Up a Server

```go
ctx := context.Background()
server, err := node.LookupServer(ctx, "my-scifi-server")
if err != nil {
    log.Fatalf("Server not found: %v", err)
}

fmt.Printf("Server: %s at %s\n", server.Name, server.Address)
```

### Querying Servers by Genre

```go
ctx := context.Background()
servers, err := node.QueryServers(ctx, "scifi", 10)
if err != nil {
    log.Fatalf("Query failed: %v", err)
}

for _, srv := range servers {
    fmt.Printf("%s: %s (%d/%d players)\n",
        srv.Name, srv.Address, srv.Players, srv.MaxPlayers)
}
```

## Architecture

### DHT Structure

- **Keys**: `/violence/{namespace}/{value}` format
  - Server records: `/violence/server/{serverName}`
  - Genre indexes: `/violence/genre/{genreName}`

### Record Format

Server records are stored as JSON:

```json
{
    "name": "my-server",
    "address": "203.0.113.5:7777",
    "genre": "scifi",
    "maxPlayers": 32,
    "uptime": 3600,
    "timestamp": "2026-03-01T15:00:00Z"
}
```

### Bootstrap Process

1. Node connects to configured bootstrap peers (timeout: 30s)
2. Node joins the DHT network
3. Node begins participating in routing (server mode) or queries (client mode)

### Record Expiration

- Records have an 8-hour TTL
- Servers should re-announce every 4 hours to stay visible
- Expired records are rejected on lookup

## Configuration

### Node Modes

- **server**: Participates in DHT routing, stores records
- **client**: Queries DHT but doesn't store records

### Bootstrap Peers

Bootstrap peers should be stable, publicly-accessible DHT nodes. For production:

- Use multiple bootstrap peers for redundancy
- Include geographically distributed peers
- Update peer list regularly

Example bootstrap configuration:

```go
BootstrapPeers: []string{
    "/dns4/dht1.violence.example.com/tcp/4001/p2p/12D3Koo...",
    "/dns4/dht2.violence.example.com/tcp/4001/p2p/12D3Koo...",
    "/dns4/dht3.violence.example.com/tcp/4001/p2p/12D3Koo...",
}
```

## Testing

Run tests:

```bash
# Unit tests only (fast)
go test ./pkg/federation/dht/... -short

# All tests including integration (slower)
go test ./pkg/federation/dht/... -v

# With coverage
go test ./pkg/federation/dht/... -cover
```

Integration tests verify:
- 10-node DHT network operation
- Bootstrap connection within 30 seconds to ≥3 peers
- Server lookup within 5 seconds
- Multi-genre server queries

## Performance

### Expected Latency

- **Bootstrap**: <30 seconds to connect to 3+ peers
- **Server Lookup**: <5 seconds (validated in tests)
- **Server Announce**: <2 seconds to propagate to network

### Network Requirements

- **Bandwidth**: ~1 Kbps per DHT node for routing overhead
- **Ports**: Configurable, default 4001/tcp
- **NAT**: Supports NAT traversal via LibP2P AutoNAT

## Fallback to HTTP Hub

For maximum availability, combine DHT with HTTP federation hub:

```go
// Try DHT first
servers, err := dhtNode.QueryServers(ctx, genre, 10)
if err != nil || len(servers) == 0 {
    // Fall back to HTTP hub
    servers, err = httpClient.QueryHub(hubURL, genre)
}
```

## License

See LICENSE file in repository root.
