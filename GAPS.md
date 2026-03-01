# Implementation Gaps — v6.0 Production Hardening

**Last Updated**: 2026-03-01

## Status Summary

**v1.0 – v5.0**: ✓ ALL CORE FEATURES COMPLETE

All milestones through v5.0 are implemented and tested:
- ✅ Key Exchange Protocol (`pkg/chat/keyexchange.go` — ECDH P-256, HKDF-SHA3-256, AES-256-GCM)
- ✅ Mobile Touch Controls (`pkg/input/touch.go` — virtual joystick, touch-to-look, action buttons)
- ✅ Federation Hub (`cmd/federation-hub/` — HTTP API, peer sync, rate limiting)
- ✅ WASM Mod Sandboxing (`pkg/mod/wasm_loader.go` — Wasmer runtime, capability-based security)
- ✅ Profanity Filter Framework (`pkg/chat/filter.go` — procedural wordlist generation)

**v6.0**: Production hardening features in design/implementation phase.

---

## v6.0 Remaining Work

### Multiplayer — Competitive Features

- [ ] **Matchmaking Algorithm** — Elo-based skill rating and team balancing
  - **Status**: Queue infrastructure exists, balancing algorithm not implemented
  - **Acceptance**: Teams balanced within 10% average Elo; configurable tolerance
  - **Files needed**: `pkg/network/matchmaking.go`

- [ ] **Anti-Cheat Foundation** — Server-side input validation
  - **Status**: Authoritative server exists, validation rules not implemented
  - **Acceptance**: Detects speed/damage hacks; logs anomalies for review
  - **Files needed**: `pkg/network/anticheat.go`

### Production — Persistence

- [ ] **Replay System** — Deterministic recording and playback
  - **Status**: Deterministic RNG exists, replay format not defined
  - **Acceptance**: Binary format captures inputs; playback reproduces game exactly
  - **Files needed**: `pkg/replay/`

- [ ] **Leaderboards** — Score aggregation and ranking
  - **Status**: Not implemented
  - **Acceptance**: Local SQLite storage; federated aggregation optional
  - **Files needed**: `pkg/leaderboard/`

- [ ] **Achievements** — Progress tracking and unlocks
  - **Status**: Not implemented
  - **Acceptance**: Condition-based unlocks; persistent storage; toast notifications
  - **Files needed**: `pkg/achievements/`

### Production — Content Safety

- [ ] **Enhanced Profanity Patterns** — L33t speak and variant detection
  - **Status**: Basic filter works, variant detection limited
  - **Acceptance**: Detects common substitutions (a→4, e→3, etc.)
  - **Files needed**: `pkg/chat/filter.go` enhancement

---

## Design Needed (Deferred to v6.1+)

- **DHT Federation**: LibP2P-based decentralized hub discovery
- **Mod Marketplace**: Centralized mod distribution platform
- **Mobile Store Publishing**: iOS/Android submission workflows
- **Cross-Save Sync**: Cloud save synchronization

---

## Completion Criteria

v6.0 is considered feature-complete when:
1. ✓ Matchmaking balances teams fairly in <30 seconds queue time
2. ✓ Anti-cheat validates all movement and damage server-side
3. ✓ Replays play back identically on any platform
4. ✓ Leaderboards persist and display correctly
5. ✓ Achievements unlock based on defined conditions
6. ✓ 82%+ test coverage maintained

---

## Notes

- v5.0 work is **complete** — core multiplayer functional
- v6.0 focuses on competitive features and persistence
- DHT and marketplace deferred until player base established
