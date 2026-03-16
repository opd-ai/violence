// Package main provides the dedicated multiplayer server for VIOLENCE.
//
// The server manages game state, handles client connections, and synchronizes
// player actions across the network. It runs headless without graphics.
//
// Usage:
//
//	go build -o violence-server ./cmd/server
//	./violence-server -port 7777 -log-level info
//
// Server flags:
//   - -port: TCP port to listen on (default: 7777)
//   - -log-level: Logging verbosity: debug, info, warn, error (default: info)
package main
