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
- Matchmaker queues are protected by `sync.Mutex`

## Matchmaking System

The matchmaking system automatically groups players into matches and assigns them to available servers based on game mode, genre, and region preferences.

### Game Modes

- **Co-op** (`ModeCoop`): 2-4 players cooperative gameplay
- **Free-for-All** (`ModeFFA`): 2-8 players deathmatch
- **Team Deathmatch** (`ModeTeamDM`): 2-16 players, team-based
- **Territory Control** (`ModeTerritory`): 2-16 players, capture points

### Using the Matchmaker

```go
// Create matchmaker with federation hub
hub := federation.NewFederationHub()
hub.Start("0.0.0.0:9000")

mm := federation.NewMatchmaker(hub)
mm.Start()
defer mm.Stop()

// Enqueue player for matchmaking
err := mm.Enqueue("player123", federation.ModeFFA, "scifi", federation.RegionUSEast)
if err != nil {
    log.Printf("failed to enqueue: %v", err)
}

// Check if player is in queue
if mm.IsQueued("player123") {
    log.Println("player is waiting for match")
}

// Remove player from queue (if they cancel)
mm.Dequeue("player123")

// Get current queue size
size := mm.GetQueueSize(federation.ModeFFA)
log.Printf("FFA queue: %d players", size)
```

### Matchmaking Process

1. **Player Enqueues**: Client calls `Enqueue()` with player ID, game mode, genre, and region
2. **Grouping**: Matchmaker groups players by mode, genre, and region
3. **Matching**: Every 2 seconds, matchmaker attempts to create matches:
   - Check if enough players are queued (minimum per mode)
   - Find available server with sufficient capacity
   - Create match and remove players from queue
4. **Timeout**: Players in queue for 60+ seconds are automatically removed

### Match Configuration

Each game mode has configured player limits:

| Mode | Min Players | Max Players |
|------|-------------|-------------|
| Co-op | 2 | 4 |
| FFA | 2 | 8 |
| Team DM | 2 | 16 |
| Territory | 2 | 16 |

### Server Assignment

The matchmaker assigns players to servers based on:
1. **Genre match**: Server must support the requested genre
2. **Region match**: Server must be in the requested region
3. **Capacity**: Server must have enough available slots for all matched players

### Queue Management

The matchmaker provides several management functions:

- `Enqueue(playerID, mode, genre, region)`: Add player to queue
- `Dequeue(playerID)`: Remove player from all queues
- `IsQueued(playerID)`: Check if player is in any queue
- `GetQueueSize(mode)`: Get number of players waiting for a mode
- `GetQueuedPlayers(mode)`: Get list of all player IDs in queue for mode

### Player Lookup

The federation hub maintains a player index for cross-server lookups:

```go
// Looking up a player
req := federation.PlayerLookupRequest{PlayerID: "player123"}
response := hub.lookupPlayer(req.PlayerID)

if response.Online {
    log.Printf("Player is on server: %s (%s)", response.ServerName, response.ServerAddress)
} else {
    log.Println("Player is offline")
}
```

Servers update their player lists via `ServerAnnouncer`:

```go
// Update player list on server
playerIDs := []string{"player1", "player2", "player3"}
announcer.UpdatePlayerList(playerIDs)
```

The hub automatically maintains the player index and cleans up stale entries when servers disconnect.

### Resource Cleanup

Both `FederationHub` and `ServerAnnouncer` support graceful shutdown:
- `hub.Stop()` shuts down the HTTP server with 5-second timeout
- `announcer.Stop()` closes the WebSocket connection and stops the announcement loop

### Error Handling

- Connection failures are logged but do not crash the server
- Invalid announcements are logged and discarded
- Stale servers are cleaned up automatically

## Testing

The federation system has comprehensive test coverage (93.9%):
- Unit tests for query filtering logic
- Unit tests for matchmaking queue management
- Integration tests for hub/announcer communication
- Integration tests for matchmaking process
- HTTP endpoint tests
- Concurrent access tests
- Stale server cleanup tests
- Queue timeout tests

Run tests with:
```bash
go test ./pkg/federation/... -v
```

## Future Enhancements

Completed features (see PLAN.md for details):
- ✅ Cross-server player lookup (Task 27)
- ✅ Matchmaking queue system (Task 28)

Planned:
- Multi-server federation integration tests (Task 29)
