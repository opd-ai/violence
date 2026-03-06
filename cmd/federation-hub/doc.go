// Package main provides a standalone federation hub server for Violence game server discovery and matchmaking.
//
// # Overview
//
// The federation hub maintains a registry of active game servers with automatic TTL-based expiration,
// player location tracking, and hub-to-hub peering for distributed server discovery.
//
// # Features
//
//   - Server Registry: Maintains active game servers with 15-minute TTL
//   - Player Lookup: Find which server a player is currently on
//   - Rate Limiting: Configurable per-IP request throttling (default: 60 req/min)
//   - Optional Authentication: Bearer token protection for server registration
//   - Hub-to-Hub Peering: Sync server lists between multiple federation hubs
//   - Health Monitoring: Health check endpoint for load balancers
//
// # Usage
//
// Basic usage:
//
//	./federation-hub
//
// With configuration:
//
//	./federation-hub \
//	  -addr=:8080 \
//	  -auth-token=secret \
//	  -peers=http://hub1.example.com:8080,http://hub2.example.com:8080 \
//	  -log-level=info \
//	  -rate-limit=60
//
// # Command-Line Flags
//
//   - addr: HTTP server address (default: :8080)
//   - auth-token: Optional authentication token for server registration
//   - peers: Comma-separated list of peer hub URLs for syncing
//   - log-level: Log level: debug, info, warn, error (default: info)
//   - rate-limit: Rate limit per IP in requests per minute (default: 60)
//
// # API Endpoints
//
// POST /announce - Register or update a game server
//
// POST /query - Query for available servers
//
// POST /lookup - Find which server a player is on
//
// GET /health - Health check endpoint
//
// GET /peers - List configured peer hubs
//
// # Hub Peering
//
// Multiple hubs can be configured to sync their server registries every 5 minutes,
// providing redundancy and geographic distribution:
//
//	Hub 1: ./federation-hub -addr=:8080 -peers=http://hub2.example.com:8080
//	Hub 2: ./federation-hub -addr=:8080 -peers=http://hub1.example.com:8080
//
// # Deployment
//
// The server can be deployed as a systemd service, Docker container, or Kubernetes deployment.
// See the README.md file for detailed deployment instructions and examples.
//
// # Performance
//
//   - Handles 1000+ registered servers per hub
//   - Rate limiting prevents abuse
//   - Hub peering provides horizontal scalability
//   - 15-minute TTL with 60-second heartbeats balances freshness and network traffic
package main
