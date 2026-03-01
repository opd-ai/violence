# Audit: github.com/opd-ai/violence/pkg/chat
**Date**: 2026-03-01
**Status**: Needs Work

## Summary
The chat package provides encrypted in-game chat with profanity filtering, ECDH key exchange, and relay server functionality. While the cryptographic implementation is solid, the package suffers from race conditions in global state management, incomplete test coverage (50-55% vs 65% target), and several hardcoded configuration values that should be parameterized.

## Issues Found
- [ ] **high** Concurrency Safety — Global `profanityWords` slice modified without mutex protection (`chat.go:201-206`, `chat.go:214-219`)
- [ ] **high** Error Handling — Panic instead of error return in HKDF derivation (`keyexchange.go:122`)
- [ ] **med** Stub/Incomplete Code — Hardcoded timestamp value `Time: 0` with comment "Would use time.Now().Unix() in real impl" (`chat.go:66`)
- [ ] **med** Test Coverage — 50-55% coverage below 65% target (chat.go: 14.3%, filter.go: 24.9%, relay.go: 36.2%, generator.go: 0%, keyexchange.go: 0%)
- [ ] **med** API Design — Hardcoded 30-second timeout in relay server, should be configurable (`relay.go:119`)
- [ ] **med** API Design — Hardcoded 100ms timeout in ReceiveEncrypted, should be configurable (`relay.go:309`)
- [ ] **low** Error Handling — Weak fallback encryption key compromises security when crypto/rand fails (`chat.go:34-37`)
- [ ] **low** Error Handling — No validation of language codes in ProfanityFilter methods (`filter.go:31-45`)
- [ ] **low** Dependencies — Uses deprecated math/rand instead of crypto/rand for profanity generation (acceptable for deterministic wordlists) (`generator.go:12`)
- [ ] **low** Documentation — No `doc.go` file for package-level documentation
- [ ] **low** Documentation — Missing godoc comments on some exported functions (`generator.go:76`, `generator.go:92`, etc.)

## Test Coverage
50-55% (target: 65%)
- chat.go: 14.3%
- filter.go: 24.9%
- relay.go: 36.2%
- generator.go: 0%
- keyexchange.go: 0%

## Dependencies
**External:**
- `github.com/sirupsen/logrus` — Structured logging (justified for relay server)
- `golang.org/x/crypto/hkdf` — Key derivation function (standard)
- `golang.org/x/crypto/sha3` — SHA3 hashing (standard)

**Standard Library:**
- crypto/aes, crypto/cipher, crypto/ecdh, crypto/rand — Encryption
- net — Network relay server
- sync — Concurrency primitives

**Integration Points:**
- `pkg/federation` imports this package for squad chat

## Recommendations
1. **Fix race conditions** — Add mutex protection to global profanityWords functions or refactor to instance methods
2. **Increase test coverage** — Add tests for generator.go and keyexchange.go to reach 65% target
3. **Remove hardcoded values** — Make timeouts configurable via struct fields or functional options
4. **Complete stub code** — Implement real timestamp in Message.Time field
5. **Fix panic in keyexchange** — Replace panic with error return in deriveKey function
6. **Add doc.go** — Create package documentation file describing chat system architecture
