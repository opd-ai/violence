// Package main provides a standalone mod registry HTTP server.
//
// The mod registry allows users to upload, download, and manage mods for the
// VIOLENCE game. It provides authentication, versioning, and dependency resolution.
//
// Usage:
//
//	go build -o mod-registry ./cmd/mod-registry
//	./mod-registry -addr :8081 -db mod-registry.db
//
// Server flags:
//   - -addr: HTTP server address (default: :8081)
//   - -db: SQLite database path (default: mod-registry.db)
//   - -storage: Mod storage directory (default: mod-storage)
//   - -log-level: Logging verbosity: debug, info, warn, error (default: info)
//   - -max-mod-size: Maximum mod size in bytes (default: 10MB)
package main
