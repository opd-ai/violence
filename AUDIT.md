# VIOLENCE Codebase Audit Report

**Audit Date:** 2026-03-01  
**Auditor:** Comprehensive Functional Analysis  
**Codebase Version:** Latest (Go 1.24+)  
**Methodology:** Dependency-based analysis with systematic feature verification

---

## AUDIT SUMMARY

````
Total Issues Found: 7
- CRITICAL BUG: 0
- FUNCTIONAL MISMATCH: 3
- MISSING FEATURE: 2
- EDGE CASE BUG: 1
- PERFORMANCE ISSUE: 1

Overall Assessment: GOOD
Build Status: ✓ PASSING
Test Status: ✓ ALL TESTS PASSING
Procedural Generation Policy: ✓ COMPLIANT (with noted exception)
````

---

## DETAILED FINDINGS

---

### FUNCTIONAL MISMATCH: Profanity Filter Embedded Wordlists

**File:** pkg/chat/filter.go:10-11, pkg/chat/wordlists/*.txt  
**Severity:** Low  
**Description:** The chat profanity filter uses embedded static wordlist files (`en.txt`, `es.txt`, `de.txt`, `fr.txt`, `pt.txt`) via Go's `//go:embed` directive. While these are configuration files rather than narrative content, they are technically "embedded asset files" which the README's procedural generation policy appears to prohibit.

**Expected Behavior:** Per README line 92: "No pre-rendered, embedded, or bundled asset files (...) are permitted in the project."

**Actual Behavior:** Five .txt wordlist files are embedded into the binary using `//go:embed wordlists/*.txt`

**Impact:** Minor policy compliance issue. The wordlists are filter configuration (not gameplay assets or narrative), but they do constitute embedded text files. Does not affect gameplay or procedural generation of actual content.

**Reproduction:**
```bash
grep -r "//go:embed" pkg/chat/
cat pkg/chat/wordlists/en.txt
```

**Code Reference:**
```go
//go:embed wordlists/*.txt
var wordlistsFS embed.FS

// LoadLanguage loads a profanity word list for the given language code
func (pf *ProfanityFilter) LoadLanguage(lang string) error {
    filename := "wordlists/" + lang + ".txt"
    data, err := wordlistsFS.ReadFile(filename)
    // ...
}
```

---

### FUNCTIONAL MISMATCH: Chat E2E Encryption Not Integrated in Main Game Loop

**File:** main.go:1-3073, pkg/chat/chat.go, pkg/chat/keyexchange.go  
**Severity:** Medium  
**Description:** The README (line 52) advertises "E2E encrypted in-game chat" as a core feature. The encryption infrastructure exists and is fully functional (PerformKeyExchange, EncryptMessage, DecryptMessage), but there is **no integration** of the key exchange or encryption/decryption into the main game's multiplayer flow or UI.

**Expected Behavior:** When players connect to multiplayer (StateMultiplayer), they should perform ECDH key exchange and all chat messages should be encrypted/decrypted using the exchanged keys.

**Actual Behavior:** 
- Key exchange functions exist but are never called from main.go or network code
- Chat system in main game creates basic Chat instances but doesn't use encryption
- No UI exists for viewing/sending encrypted chat messages
- Tests verify encryption works, but it's not wired into actual gameplay

**Impact:** High marketing/feature mismatch. The feature is advertised but not actually usable by players. The infrastructure is complete and tested, just not connected.

**Reproduction:**
```bash
# Search for PerformKeyExchange calls outside tests
grep -r "PerformKeyExchange" . --include="*.go" | grep -v test
# Result: Only defined in keyexchange.go, never called

# Check if main.go uses encryption
grep -i "encrypt\|keyexchange" main.go
# Result: No matches
```

**Code Reference:**
```go
// pkg/chat/keyexchange.go - IMPLEMENTED BUT NOT USED
func PerformKeyExchange(conn net.Conn) ([]byte, error) {
    // Full ECDH implementation exists...
}

// main.go - No call to key exchange anywhere
func (g *Game) openMultiplayer() {
    g.mpSelectedMode = 0
    g.mpStatusMsg = ""
    // ... no encryption initialization
}
```

---

### MISSING FEATURE: No Visual Representation for Minigames

**File:** main.go:2817-2999, pkg/minigame/minigame.go  
**Severity:** Medium  
**Description:** Minigame logic (lockpicking, hacking, circuit tracing, bypass codes) is fully implemented with state management and input handling, but the rendering functions (`drawLockpickGame`, `drawHackGame`, `drawCircuitGame`, `drawCodeGame`) only draw primitive colored shapes without any actual UI elements, text, or visual feedback that would make the minigames playable/understandable.

**Expected Behavior:** Minigames should have visual representations that communicate game state to the player (e.g., showing the lockpick position indicator, displaying hack sequence numbers, showing circuit grid cells, displaying code entry digits).

**Actual Behavior:** Drawing functions create basic geometric shapes (rectangles, circles) with hardcoded colors but no text rendering or meaningful visual communication. Players would not understand what actions to take.

**Impact:** Minigames are technically functional but visually unplayable. This affects the door/lockpicking system mentioned in the README.

**Reproduction:**
1. Start a new game
2. Approach a locked door (no keycard)
3. Press interact to trigger minigame
4. Observe that only colored rectangles/circles are drawn with no labels, numbers, or instructions

**Code Reference:**
```go
// main.go:2867-2897
func (g *Game) drawLockpickGame(screen *ebiten.Image, centerX, centerY float32) {
    // ... draws rectangles and circles
    // NO TEXT RENDERING
    // NO VISUAL FEEDBACK for target zone
    // NO INDICATION of current pick position relative to target
}

// main.go:2899-2936 - Similar issues in all minigame drawing functions
```

---

### MISSING FEATURE: Federation Server Discovery Not Implemented

**File:** pkg/federation/discovery.go, main.go:2209-2229  
**Severity:** Low  
**Description:** The README documents cross-server federation/matchmaking as a feature (line 51). The federation package has discovery functions (`DiscoverServers`, `Announce`), but these functions have no actual implementation—they immediately return empty results.

**Expected Behavior:** `DiscoverServers()` should perform actual server discovery (UDP broadcast, multicast, or query a central registry). `refreshServerBrowser()` in main.go should populate with real servers.

**Actual Behavior:** 
- `DiscoverServers()` returns empty slice immediately
- `Announce()` is a no-op
- Server browser always shows "No servers found"

**Impact:** Federation/matchmaking feature advertised but non-functional. Players cannot discover or join federated servers.

**Reproduction:**
```bash
# Check implementation
cat pkg/federation/discovery.go
# See that DiscoverServers returns []ServerAnnouncement{} immediately
```

**Code Reference:**
```go
// pkg/federation/discovery.go
func DiscoverServers(timeout time.Duration) []ServerAnnouncement {
    // TODO: Implement actual server discovery
    // For now, return empty list
    return []ServerAnnouncement{}
}

func Announce(port int, info ServerInfo) error {
    // TODO: Implement announcement protocol
    return nil
}
```

---

### EDGE CASE BUG: Weapon Mastery XP Awarded Before Damage Validation

**File:** main.go:797-803  
**Severity:** Low  
**Description:** When a weapon hits an enemy, mastery XP is awarded immediately (line 798) before checking if the enemy is actually alive and can take damage (line 795 checks `agent.Health > 0` but XP is awarded inside that block before damage is applied). If the damage calculation or application fails, the player still receives mastery XP for a hit that didn't actually count.

**Expected Behavior:** Mastery XP should only be awarded after confirming damage was successfully applied to a living enemy.

**Actual Behavior:** XP is awarded as soon as `hitResult.Hit && hitResult.EntityID > 0` and the agent exists, regardless of whether damage is applied.

**Impact:** Minor progression inflation. Players can gain slightly more mastery XP than deserved if there are edge cases in damage application.

**Reproduction:**
1. Modify damage application to fail under specific conditions (e.g., invulnerable enemy state)
2. Shoot that enemy
3. Observe mastery XP increases despite no damage dealt

**Code Reference:**
```go
// main.go:791-806
for _, hitResult := range hitResults {
    if hitResult.Hit && hitResult.EntityID > 0 {
        agentIdx := int(hitResult.EntityID - 1)
        if agentIdx >= 0 && agentIdx < len(g.aiAgents) {
            agent := g.aiAgents[agentIdx]
            if agent.Health > 0 {
                // Award mastery XP for successful hit
                if g.masteryManager != nil {
                    g.masteryManager.AddMasteryXP(g.arsenal.CurrentSlot, 10) // BUG: Awarded before damage applied
                }
                // Apply damage with upgrades and mastery bonuses
                upgradedDamage := g.getUpgradedWeaponDamage(currentWeapon)
                agent.Health -= upgradedDamage // Damage applied AFTER XP awarded
```

---

### PERFORMANCE ISSUE: Particle System Renders All Particles Without Culling

**File:** main.go:2495-2516  
**Severity:** Medium  
**Description:** The `renderParticles` function iterates through ALL active particles and performs distance checks, but it uses a simple squared distance check (`dist < 400`) without spatial partitioning or proper view frustum culling. With a max capacity of 1024 particles, this performs ~1024 distance calculations per frame even if only 10 particles are visible.

**Expected Behavior:** Use spatial hashing or octree to only check particles in visible sectors. Alternatively, maintain a sorted list by distance and early-exit once particles are too far.

**Actual Behavior:** O(n) iteration over all particles every frame with naive distance checks.

**Impact:** Performance degradation when many particles are active (weather effects, explosions, combat). Can cause frame drops on lower-end hardware.

**Reproduction:**
1. Start new game in a genre with weather effects (postapoc radiation_glow, scifi blink_panel)
2. Trigger multiple explosions (destroy barrels)
3. Monitor FPS with many particles active
4. Profile shows renderParticles consuming significant CPU time

**Code Reference:**
```go
// main.go:2495-2516
func (g *Game) renderParticles(screen *ebiten.Image) {
    particles := g.particleSystem.GetActiveParticles() // Returns ALL particles
    for _, p := range particles { // O(n) iteration
        dx := p.X - g.camera.X
        dy := p.Y - g.camera.Y
        dist := dx*dx + dy*dy
        if dist < 400 { // Only culls by distance, not view frustum
            // Simplified screen position calculation
            screenX := config.C.InternalWidth/2 + int(dx*10)
            screenY := config.C.InternalHeight/2 + int(dy*10)
            // No check if particle is actually in camera view direction
            if screenX >= 0 && screenX < config.C.InternalWidth && screenY >= 0 && screenY < config.C.InternalHeight {
                // ...
            }
        }
    }
}
```

---

### FUNCTIONAL MISMATCH: Config Hot-Reload Documented But Watch Not Called

**File:** pkg/config/config.go:112-170, main.go  
**Severity:** Low  
**Description:** The config package implements a sophisticated hot-reload system using fsnotify with `Watch()` function and callback support, but this functionality is never invoked from main.go. The README mentions configuration loading but doesn't explicitly advertise hot-reload, so this is more of an unrealized feature than a bug.

**Expected Behavior:** Either:
1. README should not imply configuration changes are dynamically reloaded, OR
2. main.go should call `config.Watch()` to enable hot-reload during gameplay

**Actual Behavior:** 
- `config.Load()` is called once at startup (line 3053)
- `config.Watch()` is never called
- Configuration changes require game restart

**Impact:** Minor. Players cannot change settings like mouse sensitivity or volume by editing config.toml while game is running and have changes apply immediately. This is a "nice-to-have" feature that's implemented but dormant.

**Reproduction:**
```bash
# Search for Watch calls
grep -n "config.Watch" main.go
# Result: No matches

# Verify Watch function exists
grep -n "func Watch" pkg/config/config.go
# Result: Line 112 - function exists but is unused
```

**Code Reference:**
```go
// pkg/config/config.go:112-170
func Watch(callback ReloadCallback) (stop func(), err error) {
    // Full implementation with viper.WatchConfig()
    // But never called from main.go
}

// main.go:3052-3055
func main() {
    if err := config.Load(); err != nil {
        log.Fatal(err)
    }
    // No call to config.Watch() here
}
```

---

## POSITIVE FINDINGS

**Strengths Observed:**
1. ✓ **Procedural Generation Policy Compliance**: All textures, audio, lore, and world generation are truly procedural with deterministic seeds. No embedded .png, .wav, .mp3, etc. files found (except profanity filter wordlists which are configuration, not content).

2. ✓ **Server Implementation Complete**: Dedicated server builds, runs, and accepts all documented flags (-port, -log-level). UDP game server architecture is functional.

3. ✓ **Save/Load Cross-Platform**: Save system correctly implements platform-specific paths:
   - Windows: `%APPDATA%\violence\saves`
   - Unix/Linux/macOS: `~/.violence/saves`

4. ✓ **Genre System Comprehensive**: All 5 genres (fantasy, scifi, horror, cyberpunk, postapoc) are fully implemented with SetGenre cascade working correctly across all subsystems.

5. ✓ **Encryption Infrastructure Solid**: ECDH key exchange, AES-256-GCM encryption, and E2E chat are fully implemented and tested—just not integrated into gameplay UI.

6. ✓ **Test Coverage Excellent**: All tests passing, including integration tests for combat, shop, crafting, skills, federation, and all subsystems.

---

## RECOMMENDATIONS

**Priority 1 (Medium):**
1. Integrate chat encryption into multiplayer flow - wire up PerformKeyExchange when players connect
2. Implement proper minigame rendering with text/numbers/visual feedback
3. Optimize particle rendering with spatial partitioning or view frustum culling

**Priority 2 (Low):**
4. Implement actual server discovery for federation system or remove from README
5. Fix mastery XP award timing to occur after damage application validation
6. Either enable config.Watch() hot-reload or document that restart is required for config changes
7. Consider generating profanity filter wordlists procedurally or clarify policy exception for configuration files

**Priority 3 (Documentation):**
8. Clarify in README that E2E chat encryption exists but is not yet integrated into gameplay
9. Document that minigames are functional but need visual polish
10. Note that federation discovery is planned/stubbed but not yet implemented

---

## CONCLUSION

The VIOLENCE codebase is in **very good condition** with excellent adherence to its procedural generation policy and comprehensive feature implementation. The issues found are relatively minor:

- No critical bugs that would crash the game or corrupt data
- Most "missing" features are actually implemented but not yet wired into UI/gameplay
- Performance issues are isolated to particle rendering and can be easily optimized
- Code quality is high with good test coverage and clean separation of concerns

The main discrepancies between README documentation and implementation are:
1. E2E chat encryption exists but isn't integrated
2. Minigames work but have placeholder visuals
3. Federation discovery is stubbed out

These are reasonable gaps for a game in active development and do not represent fundamental architectural flaws.

**Overall Grade: B+** (Would be A- once chat encryption is integrated and minigame visuals are improved)
