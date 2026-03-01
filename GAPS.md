# Implementation Gaps — v5.0+ Multiplayer & Production

**Last Updated**: 2026-03-01

## Status Summary

**v1.0 – v4.0**: ✓ ALL GAPS RESOLVED

All core engine features, weapons, AI, visual polish, and gameplay expansion gaps have been fully implemented and tested. See `docs/archive/GAPS_ORIGINAL_2026-03-01.md` for historical record.

**v5.0+**: Multiplayer and production features require design and implementation.

---

## v5.0+ Remaining Work

### Multiplayer — Core Networking

- [ ] **Key Exchange Protocol** — Design and implement ECDH key exchange for E2E encrypted chat
  - **Status**: Networking infrastructure exists (`pkg/network`), chat relay exists (`pkg/chat`), but key exchange not implemented
  - **Acceptance**: Two clients can establish shared AES key via ECDH without pre-shared secrets
  - **Files needed**: `pkg/chat/keyexchange.go`

### Multiplayer — Platform Support

- [ ] **Mobile Input Mapping** — Design touch control overlay for iOS/Android builds
  - **Status**: Gamepad and keyboard/mouse implemented, mobile touch controls not designed
  - **Acceptance**: Virtual joystick for movement, touch-to-aim, tap buttons for actions
  - **Files needed**: `pkg/input/touch.go`, `pkg/ui/touchoverlay.go`

### Multiplayer — Federation

- [ ] **Federation Hub Hosting** — Specify self-hosting architecture or DHT approach
  - **Status**: Federation protocol exists (`pkg/federation`), hub hosting strategy undefined
  - **Acceptance**: Documentation on running self-hosted federation hub or DHT node
  - **Files needed**: `docs/federation-hosting.md`, possible DHT implementation

### Production — Modding

- [ ] **Mod Sandboxing** — Define security model for Go plugins or evaluate WASM runtime
  - **Status**: Mod loader exists (`pkg/mod`), but no sandboxing or security restrictions
  - **Acceptance**: Mods cannot access filesystem/network outside permitted paths; documented security model
  - **Files needed**: `pkg/mod/sandbox.go`, `docs/mod-security.md`

### Production — Content Safety

- [ ] **Profanity Word List** — Compile and load localized word lists for chat filter
  - **Status**: Chat filter toggle exists, word lists not provided
  - **Acceptance**: English, Spanish, German, French, Portuguese profanity lists loaded from config
  - **Files needed**: `pkg/chat/wordlists/`, `pkg/chat/filter.go`

---

## Design Needed (No Active Tasks Yet)

The following features require design before implementation can begin:

- **Matchmaking Algorithm** — How to balance teams in deathmatch/co-op
- **Anti-Cheat Strategy** — Server-side validation approach
- **Replay System** — Deterministic replay recording format
- **Leaderboards** — Score aggregation and persistence
- **Achievements** — Local and server-synced achievement tracking

---

## Completion Criteria

v5.0 is considered feature-complete when:
1. ✓ All checkboxes above marked complete
2. ✓ Multiplayer integration tests pass (4-player co-op, deathmatch)
3. ✓ Mobile builds deploy to iOS/Android and accept touch input
4. ✓ Federation cross-server matchmaking functional
5. ✓ Mod security model documented and enforced
6. ✓ Chat profanity filter active for all supported languages

---

## Notes

- v1.0 – v4.0 work is **complete** and should not be reopened unless critical bugs discovered.
- Focus on v5.0 multiplayer features to unlock co-op and deathmatch modes.
- Mobile and federation features are lower priority until core multiplayer is validated.
