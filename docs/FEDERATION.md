# Federation System Documentation

## Overview

The federation system enables server discovery and cross-server matchmaking for Violence multiplayer games. Game servers announce their presence to a central federation hub, and clients query the hub to find available servers based on region, genre, and player count.

## Architecture

### Components

1. **FederationHub**: Central server that manages server announcements and client queries
2. **ServerAnnouncer**: Client-side component that sends periodic announcements to the hub
3. **ServerQuery**: Query structure for filtering available servers

### Communication Protocol

- **Server Announcements**: WebSocket-based, JSON-encoded messages
- **Client Queries**: REST API (HTTP POST) with JSON request/response
- **Heartbeat Interval**: 10 seconds (configurable)
- **Stale Timeout**: 30 seconds (servers not heard from are removed)

## Usage

### Starting a Federation Hub

```go
import "github.com/opd-ai/violence/pkg/federation"

hub := federation.NewFederationHub()
if err := hub.Start("0.0.0.0:9000"); err != nil {
    log.Fatal(err)
}
defer hub.Stop()
```

### Announcing a Game Server

```go
announcement := federation.ServerAnnouncement{
    Name:       "my-game-server",
    Address:    "game.example.com:8000",
    Region:     federation.RegionUSEast,
    Genre:      "scifi",
    Players:    5,
    MaxPlayers: 16,
}

announcer := federation.NewServerAnnouncer("ws://hub.example.com:9000/announce", announcement)
if err := announcer.Start(); err != nil {
    log.Fatal(err)
}
defer announcer.Stop()

// Update player count as players join/leave
announcer.UpdatePlayers(7)
```

### Querying for Servers

```go
// Query servers by region and genre
query := federation.ServerQuery{
    Region: &federation.RegionUSEast,
    Genre:  ptrString("scifi"),
}

body, _ := json.Marshal(query)
resp, err := http.Post("http://hub.example.com:9000/query", "application/json", bytes.NewReader(body))
// ... handle response

var servers []*federation.ServerAnnouncement
json.NewDecoder(resp.Body).Decode(&servers)
```

## Regions

The system supports the following regions:

- `RegionUSEast` - US East Coast
- `RegionUSWest` - US West Coast
- `RegionEUWest` - Western Europe
- `RegionEUEast` - Eastern Europe
- `RegionAsiaPac` - Asia-Pacific
- `RegionSouthAm` - South America
- `RegionUnknown` - Unknown/unspecified region

## Query Filters

Clients can filter servers using the following criteria:

- **Region**: Match servers in a specific geographical region
- **Genre**: Match servers running a specific game genre (e.g., "scifi", "fantasy", "horror")
- **MinPlayers**: Match servers with at least N players
- **MaxPlayers**: Match servers with at most N players

All filters are optional. An empty query returns all available servers.

## Server Lifecycle

1. Game server creates a `ServerAnnouncer` and calls `Start()`
2. Announcer sends initial announcement to hub
3. Announcer sends periodic heartbeat announcements every 10 seconds
4. Hub tracks last announcement timestamp for each server
5. Hub removes servers that haven't announced in 30+ seconds
6. When server shuts down, call `announcer.Stop()` to close connection

## REST API

### POST /query

Query for available servers.

**Request Body:**
```json
{
  "region": "us-east",
  "genre": "scifi",
  "minPlayers": 2,
  "maxPlayers": 8
}
```

**Response Body:**
```json
[
  {
    "name": "server-1",
    "address": "game1.example.com:8000",
    "region": "us-east",
    "genre": "scifi",
    "players": 5,
    "maxPlayers": 16,
    "timestamp": "2026-02-28T12:00:00Z"
  }
]
```

### WebSocket /announce

Server announcement endpoint. Accepts WebSocket connections and expects JSON-encoded announcements.

**Message Format:**
```json
{
  "name": "my-server",
  "address": "game.example.com:8000",
  "region": "us-east",
  "genre": "scifi",
  "players": 5,
  "maxPlayers": 16
}
```

## Configuration

### FederationHub Configuration

- `staleTimeout`: Duration before inactive servers are removed (default: 30s)
- `cleanupInterval`: How often to check for stale servers (default: 10s)

### ServerAnnouncer Configuration

- `interval`: How often to send announcements (default: 10s)

## Implementation Details

### Thread Safety

All components are thread-safe and use mutexes to protect shared state:
- Hub server list is protected by `sync.RWMutex`
- Announcer state is protected by `sync.Mutex`

### Resource Cleanup

Both `FederationHub` and `ServerAnnouncer` support graceful shutdown:
- `hub.Stop()` shuts down the HTTP server with 5-second timeout
- `announcer.Stop()` closes the WebSocket connection and stops the announcement loop

### Error Handling

- Connection failures are logged but do not crash the server
- Invalid announcements are logged and discarded
- Stale servers are cleaned up automatically

## Testing

The federation system has comprehensive test coverage (92.8%):
- Unit tests for query filtering logic
- Integration tests for hub/announcer communication
- HTTP endpoint tests
- Concurrent access tests
- Stale server cleanup tests

Run tests with:
```bash
go test ./pkg/federation/... -v
```

## Future Enhancements

See PLAN.md tasks 27-29 for planned features:
- Cross-server player lookup
- Matchmaking queue system
- Multi-server federation integration tests
