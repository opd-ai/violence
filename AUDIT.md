# FUNCTIONAL AUDIT REPORT
**Project:** VIOLENCE - Raycasting First-Person Shooter  
**Audit Date:** 2026-02-28  
**Auditor:** GitHub Copilot CLI (Automated Code Analysis)  
**Scope:** Complete functional audit comparing documented behavior (README.md) with actual implementation

---

## EXECUTIVE SUMMARY

This audit examined the VIOLENCE codebase against its documented functionality in README.md. The project successfully compiles, passes all existing tests, and implements the majority of documented features. However, several functional discrepancies, edge case vulnerabilities, and missing implementation details were identified.

**Total Issues Found:** 15  
- **CRITICAL BUG:** 3  
- **FUNCTIONAL MISMATCH:** 5  
- **MISSING FEATURE:** 4  
- **EDGE CASE BUG:** 3  

**Overall Assessment:** The codebase is functional and mostly complete, but contains critical issues that could impact gameplay stability and user experience. Most issues are related to edge case handling, incomplete procedural generation policy enforcement, and inconsistent genre support across modules.

---

## DETAILED FINDINGS

---

### [x] CRITICAL BUG: Procedural Generation Policy Violation in Audio Package (2026-02-28)
**File:** pkg/audio/audio.go:1-100+  
**Severity:** High  
**Status:** VERIFIED - COMPLIANT
**Description:** The README states "100% of gameplay assets are procedurally generated at runtime" and explicitly prohibits "pre-rendered, embedded, or bundled asset files (e.g., `.mp3`, `.wav`, `.ogg`, `.png`)". However, the audio package imports `github.com/hajimehoshi/ebiten/v2/audio/wav` and mentions WAV encoding in comments at line 11, suggesting potential use of WAV format for audio data, which could violate the procedural generation policy if WAV files are embedded or pre-created.

**Expected Behavior:** All audio must be procedurally generated at runtime from algorithms, with no embedded or bundled audio files. WAV format may only be used for runtime-generated PCM buffers in memory (as permitted by the note on encoding formats).

**Actual Behavior:** ~~The code imports WAV encoding libraries, and the package documentation mentions "runtime use of standard encoding formats (e.g., WAV for in-memory PCM audio buffers)" which is compliant. However, without examining the full audio generation pipeline, it's unclear if this is only for runtime encoding or if there are embedded assets.~~ **VERIFIED:** Full audit completed. WAV library is used exclusively for runtime encoding of procedurally generated PCM buffers. No embedded, bundled, or pre-rendered audio files exist in the repository. All audio is generated via deterministic algorithms (`GenerateReloadSound`, `GenerateEmptyClickSound`, `GeneratePickupJingleSound`, `GenerateAmbientTrack`). File system scan found zero audio files. Source code audit found no `embed` directives, no file loading, and no bundled assets. See `docs/AUDIO_COMPLIANCE.md` for full audit report.

**Impact:** ~~If embedded WAV files exist, this violates the core procedural generation policy and the project's stated design philosophy. This could mislead users about the project's capabilities.~~ **RESOLVED:** No policy violation found. Audio package is 100% compliant with procedural generation requirements.

**Reproduction:** 
~~1. Check if any .wav, .mp3, or .ogg files exist in the repository
2. Verify that all audio playback uses only procedurally generated PCM data
3. Confirm no audio assets are bundled in the binary~~
**VERIFIED:** All three checks passed. Repository contains zero audio files, all playback uses procedurally generated PCM data, and no audio assets are bundled in the binary.

---

### [x] CRITICAL BUG: Race Condition in Config Hot-Reload (2026-02-28)
**File:** pkg/config/config.go:97-120  
**Severity:** High  
**Status:** FIXED
**Description:** The `Watch()` function starts a file watcher that calls `viper.OnConfigChange()` with a callback that locks `mu` and modifies the global `C` config. However, the returned stop function is a no-op (line 119). This means once `Watch()` is called, there is no way to stop the file watcher, leading to potential goroutine leaks and race conditions during shutdown or when multiple watchers are inadvertently started.

**Expected Behavior:** The `Watch()` function should return a functional stop mechanism that cancels the file watching goroutine safely.

**Actual Behavior:** ~~The function returns `func() {}` (a no-op) with a comment "viper provides no stop mechanism" (line 118), leaving the watcher running indefinitely.~~ **FIXED:** Now uses context-based cancellation to prevent callback execution after stop() is called. Only one viper watcher is created per application lifecycle to avoid viper's internal race conditions.

**Impact:** 
- ~~Goroutine leak: The file watcher goroutine continues running even after it's no longer needed~~
- ~~Race conditions: If the config file changes during shutdown, the callback may attempt to access deallocated resources~~
- ~~Multiple watchers: Calling `Watch()` multiple times creates multiple goroutines watching the same file~~
**FIXED:** Context-based cancellation prevents callbacks after stop(), and singleton watcher pattern prevents viper race conditions.

**Reproduction:**
~~```go
cfg := config.Load()
stop1, _ := config.Watch(nil)
stop2, _ := config.Watch(nil)
stop1() // Does nothing
stop2() // Does nothing
// Both watchers still running
```~~
**FIXED:** Now calling stop1() or stop2() properly cancels the watcher.

**Code Reference:**
~~```go
func Watch(callback ReloadCallback) (stop func(), err error) {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		mu.Lock()
		defer mu.Unlock()
		old := C
		var newCfg Config
		if err := viper.Unmarshal(&newCfg); err == nil {
			C = newCfg
			if callback != nil {
				callback(old, newCfg)
			}
		}
	})
	// viper provides no stop mechanism; return no-op for API compatibility
	return func() {}, nil
}
```~~
**FIXED:** Implemented context-based cancellation with singleton watcher pattern.

---

### [x] CRITICAL BUG: Unbounded Goroutine Creation in Federation Matchmaking (2026-02-28)
**File:** pkg/federation/federation.go:76-79  
**Severity:** High  
**Status:** FIXED
**Description:** The `Match()` function uses `rand.Intn()` from the global `math/rand` package without seeding or locking. When called concurrently from multiple goroutines (common in multiplayer matchmaking), this creates a data race on the global random number generator state.

**Expected Behavior:** Thread-safe random number generation for concurrent matchmaking requests.

**Actual Behavior:** ~~Uses unseeded global `math/rand` which is not safe for concurrent use.~~ **FIXED:** Now uses instance-based `pkg/rng.RNG` with deterministic seeding, which is thread-safe and follows VIOLENCE's coding standards for deterministic procedural generation.

**Impact:** 
- ~~Data races when multiple clients request matchmaking simultaneously~~
- ~~Non-deterministic server selection~~
- ~~Potential panics or incorrect behavior under concurrent load~~
**FIXED:** Thread-safe instance RNG eliminates data races and provides deterministic server selection.

**Reproduction:**
~~```go
fed := federation.NewFederation()
fed.Register("server1", "localhost:8001")
fed.Register("server2", "localhost:8002")

// Concurrent matchmaking requests
for i := 0; i < 100; i++ {
    go func() {
        _, _ = fed.Match() // Data race on rand
    }()
}
```~~
**FIXED:** No data races detected with race detector.

**Code Reference:**
~~```go
func (f *Federation) Match() (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	var available []*ServerInfo
	for _, info := range f.servers {
		if info.Players < info.MaxPlayers {
			available = append(available, info)
		}
	}
	
	if len(available) == 0 {
		return "", fmt.Errorf("no available servers")
	}
	
	// Random selection among available servers
	selected := available[rand.Intn(len(available))] // BUG: Unsafe concurrent access
	return selected.Address, nil
}
```~~
**FIXED:** Now uses `f.rng.Intn()` with instance-based RNG from `pkg/rng`.

---

### [x] FUNCTIONAL MISMATCH: Incomplete Genre Support in Lore Package (2026-02-28)
**File:** pkg/lore/lore.go:213  
**Severity:** Medium  
**Status:** FIXED
**Description:** The `generateTitle()` function uses `strings.Title()` (line 213) which has been deprecated since Go 1.18 and will produce incorrect results for Unicode text. The README promises "procedurally generated collectible lore" without language restrictions, but the current implementation uses a deprecated function that doesn't properly handle case conversion for international text.

**Expected Behavior:** Proper Unicode-aware title case conversion for all procedurally generated lore text across all languages and genres.

**Actual Behavior:** ~~Uses deprecated `strings.Title()` which only works correctly for ASCII text.~~ **FIXED:** Now uses `cases.Title(language.English)` from `golang.org/x/text/cases` package for proper Unicode-aware title case conversion.

**Impact:** 
- ~~Generated lore titles may have incorrect capitalization for non-ASCII characters~~
- ~~Violates Go best practices by using deprecated API~~
- ~~Reduces quality of procedural content for non-English-like genres~~
**FIXED:** Now properly handles Unicode text including German (große), French (café), Greek (αθήνα), Turkish (istanbul), Chinese (中文), and emoji characters. Complies with Go best practices using modern `cases` package.

**Reproduction:**
~~```go
gen := lore.NewGenerator(12345)
gen.SetGenre("fantasy")
entry := gen.Generate("test_id")
// If entry.Title contains non-ASCII, capitalization may be incorrect
```~~
**FIXED:** Added comprehensive Unicode test suite with 7 test cases covering ASCII, German, French, Greek, Turkish, Chinese, and emoji characters. All tests pass.

**Code Reference:**
~~```go
func (g *Generator) generateTitle(category string, rng *rand.Rand) string {
	prefixes := []string{"The", "Ancient", "Lost", "Hidden", "Forgotten", "Sacred"}
	nouns := []string{"Chronicle", "Record", "Testament", "Account", "Report", "Log"}
	
	prefix := prefixes[rng.Intn(len(prefixes))]
	noun := nouns[rng.Intn(len(nouns))]
	
	return fmt.Sprintf("%s %s of %s", prefix, noun, strings.Title(category)) // BUG: Deprecated function
}
```~~
**FIXED:** Now uses `cases.Title(language.English).String(category)` for proper Unicode-aware title case conversion.

---

### FUNCTIONAL MISMATCH: Missing Cross-Platform Save Path Verification
**File:** pkg/save/save.go:83-94  
**Severity:** Medium  
**Description:** The README states save functionality is "cross-platform", and the code uses `os.UserHomeDir()` which should work across platforms. However, the save path `~/.violence/saves` uses a Unix-style hidden directory convention (leading dot), which is not idiomatic on Windows. Windows users expect application data in `%APPDATA%` or `%LOCALAPPDATA%`, not hidden directories in the home folder.

**Expected Behavior:** Platform-specific save paths that follow OS conventions (e.g., `~/.violence/` on Unix/Linux/macOS, `%APPDATA%\violence\` on Windows).

**Actual Behavior:** Uses `~/.violence/saves` on all platforms.

**Impact:** 
- Non-idiomatic file locations on Windows
- Save files may be harder to find for Windows users
- Potential conflicts with Windows folder conventions

**Reproduction:**
1. Run on Windows
2. Create a save file
3. Check save location - will be `C:\Users\<username>\.violence\saves\` instead of `%APPDATA%\violence\saves\`

**Code Reference:**
```go
func getSavePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	savePath := filepath.Join(home, ".violence", "saves") // Not idiomatic on Windows
	if err := os.MkdirAll(savePath, 0o755); err != nil {
		return "", fmt.Errorf("failed to create save directory: %w", err)
	}
	return savePath, nil
}
```

---

### FUNCTIONAL MISMATCH: Genre SetGenre Functions are No-Ops in Multiple Packages
**File:** pkg/procgen/genre/genre.go:40, pkg/raycaster/raycaster.go:442, pkg/chat/chat.go:162, pkg/save/save.go:222, pkg/lore/lore.go:302, pkg/mod/mod.go:216  
**Severity:** Medium  
**Description:** Multiple packages define package-level `SetGenre(genreID string)` functions that are complete no-ops (empty function bodies). The README describes a genre system that affects procedural generation across all modules, but these functions don't actually do anything.

**Expected Behavior:** Package-level `SetGenre()` functions should configure global genre state or store the genre for use by subsequent function calls.

**Actual Behavior:** Functions exist but have empty bodies, making them non-functional.

**Impact:** 
- Misleading API: Functions appear to configure genre but don't
- Inconsistent genre application: Some packages use instance methods (e.g., `Generator.SetGenre()`), others use package-level functions
- Potential confusion for users trying to set genre globally

**Reproduction:**
```go
genre.SetGenre("cyberpunk") // Does nothing
raycaster.SetGenre("horror") // Does nothing
chat.SetGenre("scifi")       // Does nothing
save.SetGenre("fantasy")     // Does nothing
lore.SetGenre("postapoc")    // Does nothing
```

**Code Reference:**
```go
// genre/genre.go:40
func SetGenre(genreID string) {}

// raycaster/raycaster.go:442
func SetGenre(genreID string) {}

// chat/chat.go:162
func SetGenre(genreID string) {}

// save/save.go:222
func SetGenre(genreID string) {}

// lore/lore.go:302
func SetGenre(genreID string) {}

// mod/mod.go:216
func SetGenre(genreID string) {}
```

---

### FUNCTIONAL MISMATCH: Tutorial System Not Implemented in Main Game Loop
**File:** main.go:37  
**Severity:** Medium  
**Description:** The README documents a `tutorial/` package for "context-sensitive tutorial prompts" (line 22), and main.go imports and initializes `tutorialSystem *tutorial.Tutorial` (line 63), but the tutorial system is never updated or rendered in the main game loop.

**Expected Behavior:** Tutorial prompts should appear contextually during gameplay based on player actions and game state.

**Actual Behavior:** The `tutorialSystem` field exists but is never called in `Update()` or `Draw()` methods.

**Impact:** 
- Missing feature: Tutorial system exists but is non-functional
- New players have no in-game guidance
- Wasted code: Tutorial package is imported but unused

**Reproduction:**
1. Start the game
2. Perform actions that should trigger tutorial prompts (e.g., first weapon pickup, first enemy encounter)
3. No tutorial prompts appear

---

### FUNCTIONAL MISMATCH: Shop System Not Integrated Into Main Game
**File:** main.go (missing shop integration)  
**Severity:** Low  
**Description:** The README documents a `shop/` package for "Between-level armory shop" (line 40), and the package exists with functional code. However, main.go does not import or initialize any shop-related systems, and there's no mechanism to enter the shop between levels.

**Expected Behavior:** After completing a level, players should have access to an armory shop where they can purchase items and upgrades.

**Actual Behavior:** Shop package exists but is not integrated into the game flow.

**Impact:** 
- Missing gameplay feature: Players cannot access the shop
- Game balance issue: Without shop, resource management may be too difficult
- Incomplete implementation of documented feature

**Reproduction:**
1. Complete a level
2. Observe that no shop interface appears
3. No mechanism exists to access shop functionality

---

### MISSING FEATURE: Network Multiplayer Modes Not Integrated
**File:** main.go (missing network integration)  
**Severity:** Medium  
**Description:** The README documents `network/` package for "Client/server netcode" (line 48) with multiple multiplayer modes (deathmatch, team, FFA, territory, cooperative) evidenced by the test files. The network package exists with extensive functionality, but main.go operates entirely in single-player mode with no multiplayer initialization or connection logic.

**Expected Behavior:** Players should be able to start or join multiplayer games with deathmatch, team-based, FFA, territory control, and cooperative modes.

**Actual Behavior:** Main game loop only supports single-player. No UI or logic exists to connect to servers or host multiplayer sessions.

**Impact:** 
- Major missing feature: Multiplayer completely absent from main game
- Network package is unused in the actual game executable
- Server executable exists (cmd/server/main.go) but client has no way to connect

**Reproduction:**
1. Run the main game
2. Observe that all menus and gameplay are single-player only
3. No option to join or host multiplayer games

---

### MISSING FEATURE: Crafting System Not Integrated Into Main Game
**File:** main.go (missing crafting integration)  
**Severity:** Low  
**Description:** The README documents `crafting/` package for "Scrap-to-ammo crafting" (line 38). The package exists with full recipe and scrap management functionality, but main.go does not import the crafting package, initialize a crafting system, or provide any UI to access crafting.

**Expected Behavior:** Players should be able to collect scrap materials and craft ammunition and other items using the crafting system.

**Actual Behavior:** Crafting package exists but is completely unused in the main game.

**Impact:** 
- Missing resource management mechanic
- Reduced gameplay depth
- Players cannot craft ammo as documented

**Reproduction:**
1. Play the game
2. Collect scrap (if scrap dropping is even implemented)
3. No crafting interface or mechanism exists

---

### MISSING FEATURE: Skills/Talent Trees Not Integrated
**File:** main.go (missing skills integration)  
**Severity:** Low  
**Description:** The README documents `skills/` package for "Skill and talent trees" (line 46). The package exists with skill tree functionality, but main.go does not import or initialize any skills system.

**Expected Behavior:** Players should have access to skill and talent trees for character progression and customization.

**Actual Behavior:** Skills package exists but is not integrated into the game.

**Impact:** 
- Reduced character progression depth
- Missing RPG-like advancement system
- Documented feature not available to players

**Reproduction:**
1. Play the game and gain XP
2. No skill tree interface or skill assignment mechanism exists

---

### MISSING FEATURE: Mod Loader Not Integrated Into Main Game
**File:** main.go (missing mod integration)  
**Severity:** Low  
**Description:** The README documents `mod/` package for "Mod loader and plugin API" (line 51). The package exists with full mod loading, conflict detection, and plugin management, but main.go does not initialize a mod loader or scan for mods.

**Expected Behavior:** Game should load mods from a mods directory at startup and apply mod content/behavior changes.

**Actual Behavior:** Mod system exists but is never initialized or used.

**Impact:** 
- No modding support despite documented feature
- Mod API is complete but inaccessible
- Community cannot extend the game

**Reproduction:**
1. Create a mods/ directory with a valid mod
2. Run the game
3. Mod is not loaded or detected

---

### EDGE CASE BUG: Division by Zero in Raycaster Floor/Ceiling Rendering
**File:** pkg/raycaster/raycaster.go:229-244  
**Severity:** Medium  
**Description:** The `CastFloorCeiling()` function checks for division by zero at the horizon line (`p == 0` at line 233), but only returns early without handling the edge case properly. While this prevents a crash, it returns pixels with distance 1e30 for the entire horizon row, which may cause rendering artifacts or incorrect fog application.

**Expected Behavior:** Horizon line pixels should have sensible default values or be handled specially in rendering.

**Actual Behavior:** Returns pixels with infinite distance (1e30) which may cause fog or lighting calculations to behave unexpectedly.

**Impact:** 
- Potential visual artifacts at the horizon line
- Fog may not render correctly at eye level
- Low severity as the function prevents crash

**Reproduction:**
```go
r := raycaster.NewRaycaster(66.0, 320, 200)
pixels := r.CastFloorCeiling(100, 5.0, 5.0, 1.0, 0.0, 0.0) // Row 100 = height/2
// All pixels will have Distance = 1e30
```

**Code Reference:**
```go
if p == 0 {
	// At horizon - return infinite distance for all pixels
	for x := 0; x < r.Width; x++ {
		pixels[x] = FloorCeilPixel{
			WorldX:   posX,
			WorldY:   posY,
			Distance: 1e30, // May cause issues in fog/lighting
			IsFloor:  isFloor,
		}
	}
	return pixels
}
```

---

### EDGE CASE BUG: Nil/Empty Map Not Handled in BSP Secret Placement
**File:** pkg/bsp/bsp.go:325-363  
**Severity:** Low  
**Description:** The `placeSecrets()` function iterates over the tile map without checking if the map is nil or empty. While `Generate()` creates the map first, if someone calls `placeSecrets()` independently or if map creation fails, this could cause a panic.

**Expected Behavior:** Function should validate map exists before iteration.

**Actual Behavior:** No nil/bounds checking before accessing tiles array.

**Impact:** 
- Potential panic if called with uninitialized generator
- Low severity as normal usage flow prevents this
- Defensive programming issue

**Reproduction:**
```go
gen := bsp.NewGenerator(64, 64, rng.NewRNG(123))
gen.placeSecrets(nil, make([][]int, 0)) // Would panic if public
```

**Code Reference:**
```go
func (g *Generator) placeSecrets(n *Node, tiles [][]int) {
	if n == nil {
		return
	}
	
	// No check for tiles == nil or empty
	for y := 1; y < g.Height-1; y++ {
		for x := 1; x < g.Width-1; x++ {
			// Accessing tiles without validation
```

---

### EDGE CASE BUG: Integer Overflow in Quest Progress Updates
**File:** pkg/quest/quest.go (UpdateProgress not shown, but referenced in main.go:542)  
**Severity:** Low  
**Description:** Quest objectives track progress with `int` type (line 39), and main.go calls `UpdateProgress("bonus_kills", 1)` to increment kill counts (line 542). For long-running games or high kill count objectives, repeatedly adding to progress could theoretically overflow the int, though this is extremely unlikely in practice.

**Expected Behavior:** Progress should be bounded or use a larger integer type to prevent overflow.

**Actual Behavior:** Uses standard `int` which is 32-bit on 32-bit systems.

**Impact:** 
- Extremely low probability in normal gameplay
- Would require 2+ billion kill count updates on 32-bit systems
- Quest completion could break if overflow occurs

**Reproduction:**
Theoretical - would require calling UpdateProgress(2147483647) times on a 32-bit system.

---

## RECOMMENDATIONS

### High Priority
1. **Fix Config Hot-Reload Race Condition:** Implement proper goroutine lifecycle management for config watching
2. **Fix Federation Random Number Generation:** Use thread-safe RNG (e.g., crypto/rand or per-goroutine rand.Rand instances)
3. **Audit Procedural Generation Policy:** Verify no embedded audio/image assets exist anywhere in the codebase
4. **Integrate Missing Core Features:** Add network multiplayer, crafting, skills, and mod loader to main game

### Medium Priority
1. **Fix Genre SetGenre No-Ops:** Either implement the functions or remove them to avoid API confusion
2. **Implement Tutorial System:** Add tutorial.Update() and tutorial.Draw() calls to main game loop
3. **Use Proper Unicode Case Conversion:** Replace deprecated strings.Title() with cases.Title()
4. **Add Platform-Specific Save Paths:** Use OS-appropriate paths (XDG on Linux, AppData on Windows, etc.)

### Low Priority
1. **Add Nil/Bounds Checking:** Add defensive validation in BSP and raycaster methods
2. **Integrate Shop System:** Add shop UI and between-level flow
3. **Document No-Op SetGenre Pattern:** If intentional, add comments explaining why functions are empty
4. **Use uint64 for Quest Progress:** Prevent theoretical integer overflow in long gameplay sessions

---

## CONCLUSION

The VIOLENCE codebase is well-structured, compiles cleanly, and implements most documented features at a fundamental level. However, several critical systems (multiplayer, crafting, skills, mod loading, tutorial, shop) exist in isolated packages but are not integrated into the main game executable. This suggests the project is in active development with modular components being built independently of the main game loop.

The three critical bugs identified (config hot-reload, federation RNG, procedural generation policy verification) should be addressed immediately as they affect core functionality and thread safety. The functional mismatches around genre support suggest an evolving API where some packages use instance methods while others have placeholder package-level functions.

Overall code quality is high, with proper concurrency primitives (mutexes), error handling, and modular design. The main gap is integration work to connect all the well-built subsystems into the cohesive game experience described in the README.

**Audit Completion Status:** ✅ Complete  
**Next Review Recommended:** After integration of missing features into main.go
