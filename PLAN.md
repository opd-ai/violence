# Implementation Plan: v6.0 — Production Hardening & Feature Completion

## Phase Overview
- **Objective**: Complete remaining v5.0+ gaps, implement replay system, matchmaking algorithms, and anti-cheat foundations for production-ready multiplayer
- **Source Document**: `GAPS.md` (v5.0+ Remaining Work) and `ROADMAP.md` (Design Needed section)
- **Prerequisites**: v5.0 complete (networking, co-op, deathmatch, federation, chat encryption, WASM mods)
- **Estimated Scope**: Medium

## Implementation Steps

### 1. Matchmaking Algorithm — Skill-Based Team Balancing ✅ [2026-03-01]
- **Deliverable**: `pkg/network/matchmaking.go` with Elo-based skill rating and team balancing
- **Dependencies**: Existing `pkg/federation` matchmaking queue
- **Status**: Implemented with 100% test coverage on core functions

**Implementation Summary**:
- ✅ Elo rating system (starting rating: 1200, K-factor: 32)
- ✅ `CalculateEloChange()` - computes rating deltas after match results
- ✅ `BalanceTeams()` - creates balanced teams minimizing skill difference
- ✅ `MatchPlayersWithinSkillRange()` - filters players by skill tolerance
- ✅ Helper functions: `AverageElo()`, `TeamBalanceDifference()`
- ✅ Comprehensive unit tests with edge cases
- ✅ 100% coverage on CalculateEloChange, BalanceTeams, MatchPlayersWithinSkillRange, AverageElo
- ✅ 88.9% coverage on TeamBalanceDifference

**Technical Approach**:
- Implement Elo rating system (starting rating: 1200, K-factor: 32)
- Track wins/losses per player in persistent storage
- Balance teams by minimizing skill difference between team averages
- Support queue filters: mode, region, skill range tolerance (±200 Elo default)

**Implementation**:
```go
type MatchmakingConfig struct {
    MinPlayers      int     // Minimum players to start match
    MaxWaitTime     time.Duration
    SkillTolerance  int     // Elo points tolerance
    RegionPriority  []string
}

func BalanceTeams(players []Player) (teamA, teamB []Player)
func CalculateEloChange(winnerElo, loserElo int) (winnerDelta, loserDelta int)
```

### 2. Anti-Cheat Foundation — Server-Side Validation ✅ [2026-03-01]
- **Deliverable**: `pkg/network/anticheat.go` with input validation and anomaly detection
- **Dependencies**: Existing authoritative server in `pkg/network`
- **Status**: Implemented with >89% test coverage on all functions

**Implementation Summary**:
- ✅ `ValidateMovement()` - detects speed hacks (>2x normal speed = kick)
- ✅ `ValidateDamage()` - validates damage against weapon definitions
- ✅ `ValidateFireRate()` - prevents rapid-fire hacks
- ✅ `CheckStatisticalAnomaly()` - detects suspicious headshot ratios (>80% triggers review)
- ✅ Helper functions: `RecordShot()`, `RecordViolation()`
- ✅ Comprehensive unit tests with edge cases and error paths
- ✅ 100% coverage on ValidateMovement, CheckStatisticalAnomaly, RecordViolation
- ✅ 89-90% coverage on ValidateDamage, ValidateFireRate, RecordShot

**Technical Approach**:
- Validate movement speed against max allowed (prevent speed hacks)
- Validate damage output against weapon definitions (prevent damage hacks)
- Track statistical anomalies (headshot ratio >80% triggers review)
- Rate-limit actions (max 20 shots/second)
- Log suspicious activity for manual review

**Implementation**:
```go
type ValidationResult struct {
    Valid    bool
    Violation string
    Severity  int // 1=warning, 2=kick, 3=ban
}

func ValidateMovement(oldPos, newPos Vec2, deltaTime float64) ValidationResult
func ValidateDamage(weaponID int, damage int, distance float64) ValidationResult
func CheckStatisticalAnomaly(playerStats *PlayerStats) ValidationResult
```

### 3. Replay System — Deterministic Recording ✅ [2026-03-01]
- **Deliverable**: `pkg/replay/` package with recording and playback
- **Dependencies**: `pkg/rng` deterministic RNG, `pkg/network` delta states
- **Status**: Implemented with 89.5% test coverage

**Implementation Summary**:
- ✅ Binary file format with 32-byte header (magic "VREP", version, seed, duration, player count)
- ✅ `ReplayRecorder` - records seed + all player inputs with timestamps
- ✅ `ReplayPlayer` - loads and plays back recorded replays
- ✅ `Step()` - iterate through input frames sequentially
- ✅ `Seek()` - fast-forward/rewind to specific timestamp using binary search
- ✅ `Reset()` - rewind to beginning
- ✅ Input bitfield system (10 flags: WASD, fire, use, reload, sprint, crouch, jump)
- ✅ Error handling for invalid files, corrupted headers, missing files
- ✅ Comprehensive unit tests with edge cases and error paths (12 test functions)
- ✅ Benchmarks for recording, saving, and loading performance

**Technical Implementation**:
- Record seed + all player inputs with timestamps
- Replay by re-executing inputs against same seed
- Binary format with header (version, seed, duration) + input stream
- Support fast-forward (via Seek), rewind (via Reset), pause (external to package)

**File Format**:
```
Header (32 bytes):
  - Magic: "VREP" (4 bytes)
  - Version: uint16
  - Seed: int64
  - Duration: uint32 (ms)
  - PlayerCount: uint8
  - Reserved: 13 bytes

Input Stream:
  - Timestamp: uint32 (ms from start)
  - PlayerID: uint8
  - InputFlags: uint16 (bitfield: WASD, fire, use, etc.)
  - MouseDeltaX: int16
  - MouseDeltaY: int16
```

**Implementation**:
```go
type ReplayRecorder struct {
    seed        int64
    inputs      []InputFrame
    startTime   time.Time
    playerCount uint8
}

func (r *ReplayRecorder) RecordInput(playerID uint8, flags InputFlags, mouseDeltaX, mouseDeltaY int16)
func (r *ReplayRecorder) Save(path string) error

type ReplayPlayer struct {
    header  ReplayHeader
    inputs  []InputFrame
    cursor  int
}

func LoadReplay(path string) (*ReplayPlayer, error)
func (p *ReplayPlayer) Step() (InputFrame, bool)
func (p *ReplayPlayer) Seek(timestampMs uint32)
func (p *ReplayPlayer) Reset()
```

### 4. Leaderboards — Score Aggregation ✅ [2026-03-01]
- **Deliverable**: `pkg/leaderboard/` package with local and federated leaderboards
- **Dependencies**: `pkg/federation` for cross-server aggregation
- **Status**: Implemented with 86.8% test coverage

**Implementation Summary**:
- ✅ SQLite-based persistent storage with indexed queries
- ✅ `Leaderboard` - local score tracking with multi-period support (all_time, weekly, daily)
- ✅ `RecordScore()` - upsert player scores with automatic timestamp tracking
- ✅ `GetTop()` - retrieve top N entries for any stat/period combination
- ✅ `GetRank()` - query player's rank in leaderboard
- ✅ `IncrementScore()` - atomic score increments (thread-safe)
- ✅ `ClearPeriod()` - reset time-based leaderboards (e.g., weekly reset)
- ✅ `FederatedLeaderboard` - opt-in cross-server leaderboard aggregation
- ✅ `SyncToHub()` - upload local scores to federation hub
- ✅ `FetchGlobalTop()` - retrieve global leaderboard rankings
- ✅ `GetGlobalRank()` - query player's global rank across all servers
- ✅ Comprehensive unit tests with edge cases and error paths (29 test functions)
- ✅ Benchmarks for record, query, and federation performance
- ✅ Privacy-focused: federation is opt-in only

**Technical Approach**:
- Local SQLite database for persistent storage
- Track: kills, deaths, wins, play time, high scores per mode
- Federated leaderboards via hub aggregation (opt-in)
- Time-based leaderboards: all-time, weekly, daily

**Implementation**:
```go
type LeaderboardEntry struct {
    PlayerID   string
    PlayerName string
    Score      int64
    Stat       string // "kills", "wins", "high_score"
    Period     string // "all_time", "weekly", "daily"
    UpdatedAt  time.Time
}

func (lb *Leaderboard) RecordScore(playerID, playerName, stat, period string, value int64) error
func (lb *Leaderboard) GetTop(stat, period string, limit int) ([]LeaderboardEntry, error)
func (lb *Leaderboard) GetRank(playerID, stat, period string) (int, error)
func (lb *Leaderboard) IncrementScore(playerID, playerName, stat, period string, delta int64) error
func (lb *Leaderboard) ClearPeriod(period string) error

type FederatedLeaderboard struct {
    *Leaderboard
    config FederatedConfig
}

func (flb *FederatedLeaderboard) SyncToHub(stat string) error
func (flb *FederatedLeaderboard) FetchGlobalTop(stat string, limit int) ([]LeaderboardEntry, error)
func (flb *FederatedLeaderboard) GetGlobalRank(playerID, stat string) (int, error)
```

### 5. Achievements System — Local Tracking ✅ [2026-03-01]
- **Deliverable**: `pkg/achievements/` package with unlock conditions and persistence
- **Dependencies**: `pkg/save` for persistence
- **Status**: Implemented with 85.4% test coverage

**Implementation Summary**:
- ✅ Achievement system with 14 default achievements across 4 categories
- ✅ `AchievementManager` - manages definitions and unlock state with thread-safe operations
- ✅ `CheckUnlocks()` - evaluates player stats against achievement conditions
- ✅ `GetProgressWithStats()` - computes current progress toward achievements
- ✅ `IsUnlocked()` - checks if achievement has been unlocked
- ✅ `GetByCategory()` - filters achievements by category (Combat, Exploration, Survival, Social)
- ✅ `Save()/Load()` - JSON-based persistence of unlocked achievements
- ✅ `Reset()` - clears all unlocks for testing or game reset
- ✅ Thread-safe concurrent access with RWMutex
- ✅ Comprehensive unit tests with 85.4% coverage (37 test functions)
- ✅ Edge cases: nil stats, invalid IDs, concurrent access, persistence errors

**Achievement Categories Implemented**:
- Combat: First Blood, Centurion (100 kills), Pacifist (0 kills), Headhunter (50 headshots), Demolition Expert (25 explosive kills)
- Exploration: Cartographer (100% map), Secret Hunter (10 secrets), Explorer (100 doors)
- Survival: Iron Man (no deaths), Speed Demon (<5 min), Untouchable (no damage)
- Social: Team Player (10 co-op), Dominator (10 deathmatch wins), Social Butterfly (100 messages)

**Technical Approach**:
- Define achievements as condition functions
- Track progress counters (e.g., "Kill 100 enemies": progress 47/100)
- Persist unlocks in JSON save file
- Trigger toast notification on unlock (via logrus logging)

**Achievement Categories**:
- Combat: First Blood, Centurion (100 kills), Pacifist (complete level without killing)
- Exploration: Cartographer (reveal 100% of map), Secret Hunter (find 10 secrets)
- Survival: Iron Man (complete game without dying), Speed Demon (<30 min completion)
- Social: Team Player (10 co-op games), Dominator (win 10 deathmatches)

**Implementation**:
```go
type Achievement struct {
    ID          string
    Name        string
    Description string
    Category    string
    Condition   func(stats *PlayerStats) bool
    Progress    func(stats *PlayerStats) (current, target int)
}

func (am *AchievementManager) CheckUnlocks(stats *PlayerStats) []Achievement
func (am *AchievementManager) GetProgress(achievementID string) (int, int)
```

### 6. Profanity Filter Enhancement — Localized Word Lists ✅ [2026-03-01]
- **Deliverable**: Enhanced `pkg/chat/filter.go` with comprehensive procedural word generation
- **Dependencies**: Existing filter framework
- **Status**: Implemented with l33t speak variant generation

**Implementation Summary**:
- ✅ `generateLeetSpeakVariants()` - generates l33t speak substitutions (a→4/@, e→3, i→1/!, o→0, s→5/$, t→7)
- ✅ Enhanced wordlists for all 5 languages with expanded vocabulary (175-214 patterns per language)
- ✅ L33t speak detection working: "sh1t", "4ss", "fuk" all properly filtered
- ✅ Multi-character substitution support (e.g., "a→4, e→3" combined)
- ✅ Comprehensive test coverage for l33t speak variants
- ✅ Test functions: TestGenerateLeetSpeakVariants, TestLeetSpeakDetection, TestLeetSpeakSanitization, TestWordlistSizeIncrease
- ✅ All tests passing, code formatted and vetted

**Technical Approach**:
- Procedurally generate language-specific word patterns using seed-based phoneme combinations
- Generate variant patterns (l33t speak: a→4, e→3, i→1, o→0, s→5, t→7)
- Support 5 languages: English, Spanish, German, French, Portuguese
- Generate 175-214 patterns per language using deterministic algorithms

**Implementation**:
```go
func generateLeetSpeakVariants(word string) []string {
    // Single-character substitutions: a→4/@, e→3, i→1/!, o→0, s→5/$, t→7
    // Multi-character combinations: {a→4, e→3}, {i→1, o→0}, etc.
}

func (pf *ProfanityFilter) LoadAllLanguages() error
```

## Technical Specifications

- **Database**: SQLite for local leaderboards/achievements (embedded, no external dependency)
- **Replay Format**: Custom binary format, ~100KB per 10 minutes of gameplay
- **Anti-Cheat**: Server-side only, no client-side components (privacy-preserving)
- **Matchmaking**: Async queue with WebSocket notifications when match found
- **Elo System**: Standard formula `newRating = oldRating + K * (actual - expected)`

## Validation Criteria

- [x] Matchmaking balances teams within 10% average Elo difference ✅ (achieved <10% in tests)
- [x] Anti-cheat detects speed >2x normal and rejects invalid movement ✅ (SpeedHackThreshold = 2x MaxSprintSpeed)
- [x] Replay files save and load correctly with identical input frames ✅ (89.5% test coverage)
- [ ] Replay playback produces deterministic results when re-executing with same seed (integration test needed)
- [x] Leaderboards persist across game restarts ✅ (TestPersistence validates DB persistence)
- [x] Achievements unlock correctly when conditions met ✅ (14 default achievements with 85.4% test coverage)
- [x] Profanity filter detects l33t speak variants (e.g., "sh1t" matches "shit", "4ss" matches "ass") ✅
- [x] All new code has >82% test coverage ✅ (matchmaking: 88-100%, anticheat: 89-100%, replay: 89.5%, leaderboard: 86.8%, achievements: 85.4%)
- [ ] Integration tests verify 4-player matchmaking queue

## Known Gaps

- **DHT Federation**: Distributed hash table discovery deferred to v6.1 (HTTP hubs sufficient for launch)
- **Mobile Store Publishing**: iOS App Store / Google Play submission process not documented (requires developer accounts)
- **Mod Marketplace**: Centralized mod distribution deferred to v7.0 (local mods work now)
