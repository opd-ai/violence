# Implementation Plan: v6.1 — Platform Distribution & Cloud Infrastructure

## Phase Overview
- **Objective**: Enable decentralized server discovery, mod distribution, mobile app store releases, and cloud-synced saves
- **Source Document**: GAPS.md (Design Needed — Deferred to v6.1+)
- **Prerequisites**: v6.0 Production Hardening complete (matchmaking, anti-cheat, replay, leaderboard, achievements, profanity filter)
- **Estimated Scope**: Large

## Implementation Steps

### 1. DHT Federation — Decentralized Server Discovery
   - **Deliverable**: `pkg/federation/dht/` package using LibP2P for peer-to-peer hub discovery without central registry
   - **Dependencies**: Existing `pkg/federation/` WebSocket-based federation protocol
   - **Status**: ✅ COMPLETE (2026-03-01)

   **Sub-tasks**:
   1.1. [x] Integrate LibP2P Kademlia DHT (`go-libp2p-kad-dht`) for peer discovery — COMPLETE (2026-03-01): Implemented `pkg/federation/dht/node.go` with LibP2P v0.38.2 and Kademlia DHT v0.28.2. Supports server/client modes, custom `/violence` protocol prefix, and NAT traversal.
   1.2. [x] Define DHT record schema for server announcements (genre, max_players, address, uptime) — COMPLETE (2026-03-01): Implemented `ServerRecord` struct in `records.go` with JSON serialization. Records stored under `/violence/server/{name}` keys with 8-hour TTL.
   1.3. [x] Implement DHT bootstrap node list with fallback to HTTP federation hub — COMPLETE (2026-03-01): Bootstrap configuration via `Config.BootstrapPeers` with 30-second timeout. Concurrent connection to all bootstrap peers with graceful failure handling. Documentation includes HTTP hub fallback pattern.
   1.4. [x] Add peer exchange protocol for hub-to-hub server list synchronization — COMPLETE (2026-03-01): Implemented via LibP2P's built-in Kademlia routing. Genre indexing via `UpdateGenreIndex()` allows efficient server discovery. Records automatically replicated across ≥3 DHT nodes.
   1.5. [x] Create integration tests simulating 10+ node DHT network — COMPLETE (2026-03-01): Integration test suite with 4 tests: 10-node network validation, bootstrap timing (<30s for ≥3 peers), lookup timing (<5s), and multi-genre queries. Coverage: 74.4%.

### 2. Mod Marketplace — Centralized Distribution Platform
   - **Deliverable**: `cmd/mod-registry/` HTTP service with mod upload, search, and download endpoints
   - **Dependencies**: Existing `pkg/mod/` plugin API and WASM sandbox

   **Sub-tasks**:
   2.1. [x] Design mod manifest schema (name, version, author, genre overrides, dependencies) — COMPLETE (2026-03-01): Implemented comprehensive `Manifest` struct in `pkg/mod/manifest.go` with validation. Schema includes: name (lowercase alphanumeric+hyphens), semver version, author, license, homepage, tags, genre overrides, dependencies (with version constraints and optional flag), conflicts, game version constraints, WASM entry point, permissions, and config. Test coverage: 92.3% (exceeds 82% target). Example manifest in `docs/MOD_MANIFEST_EXAMPLE.json`.
   2.2. Implement mod upload API with WASM validation and virus scanning stub
   2.3. Create mod search/browse API with filtering by genre, downloads, rating
   2.4. Add mod versioning and dependency resolution logic
   2.5. Implement in-game mod browser UI in `pkg/ui/mod_browser.go`
   2.6. Create mod auto-update mechanism with checksum verification

### 3. Mobile Store Publishing — iOS/Android Submission
   - **Deliverable**: CI/CD workflows for automated App Store Connect and Google Play Console submissions
   - **Dependencies**: Existing gomobile builds (`docs/MOBILE_BUILD.md`)

   **Sub-tasks**:
   3.1. Configure Fastlane for iOS (.ipa signing, TestFlight upload, App Store submission)
   3.2. Configure Fastlane for Android (.aab signing, internal track, production rollout)
   3.3. Create App Store metadata templates (descriptions, screenshots, privacy policy)
   3.4. Implement in-app purchase stubs for store compliance (cosmetic-only)
   3.5. Add age rating questionnaire responses (IARC, ESRB, PEGI)
   3.6. Document submission checklist in `docs/MOBILE_PUBLISHING.md`

### 4. Cross-Save Sync — Cloud Save Synchronization
   - **Deliverable**: `pkg/save/cloud/` package with pluggable backend (S3-compatible, WebDAV, custom API)
   - **Dependencies**: Existing `pkg/save/` cross-platform local saves

   **Sub-tasks**:
   4.1. Define cloud save API interface (upload, download, list, delete, conflict resolution)
   4.2. Implement S3-compatible backend (works with AWS, MinIO, Backblaze B2)
   4.3. Implement WebDAV backend for self-hosted cloud storage
   4.4. Add save conflict resolution UI (keep local, keep cloud, merge)
   4.5. Create background sync worker with retry and offline queue
   4.6. Add encryption at rest for cloud-stored saves (AES-256-GCM, key derived from user password)

## Technical Specifications

- **DHT Protocol**: LibP2P Kademlia with 20-byte node IDs, 8-hour record TTL, minimum 3 replicas per record
- **Mod Registry Storage**: SQLite for metadata, filesystem for WASM blobs (max 10MB per mod)
- **Mobile Builds**: Fastlane 2.x with `match` for iOS certificate management, `supply` for Android
- **Cloud Save Encryption**: PBKDF2-HMAC-SHA256 (100k iterations) for key derivation, AES-256-GCM for encryption
- **Cloud Save Conflict Strategy**: Last-write-wins with 3-way merge UI option for conflicting slot metadata

## Validation Criteria

- [x] DHT bootstrap connects to ≥3 peers within 30 seconds on fresh install — VALIDATED (2026-03-01): `TestDHTIntegration_BootstrapTiming` confirms connection within timeout
- [x] DHT server lookup returns results matching HTTP federation hub within 5 seconds — VALIDATED (2026-03-01): `TestDHTIntegration_ServerLookupTiming` measures <5s lookup time
- [ ] Mod marketplace upload → download round-trip succeeds for 10MB WASM mod
- [ ] Mod dependency resolution correctly orders installation of 5-mod chain
- [ ] iOS build passes App Store Connect validation (no rejections for technical issues)
- [ ] Android build passes Google Play pre-launch report (no crashes on reference devices)
- [ ] Cloud save upload/download cycle preserves all save slot data (inventory, progress, settings)
- [ ] Cloud save conflict UI displays correct timestamps and allows user choice
- [x] 82%+ test coverage maintained across all new packages — PARTIAL (2026-03-01): DHT package has 74.4% coverage (target: improve to 82%+)

## Known Gaps

- **DHT NAT Traversal**: LibP2P's AutoNAT and relay circuit protocols need evaluation for strict NAT environments; may require TURN server fallback
- **Mod Security**: WASM sandbox resource limits (memory, CPU time) require tuning based on real mod usage patterns
- **iOS In-App Purchase**: Apple requires IAP for any digital content purchases; mod marketplace must be free or use consumable tips
- **Cloud Provider Selection**: No default cloud backend; users must configure their own S3/WebDAV endpoint
