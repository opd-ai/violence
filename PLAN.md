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

### 3. Replay System — Deterministic Recording
- **Deliverable**: `pkg/replay/` package with recording and playback
- **Dependencies**: `pkg/rng` deterministic RNG, `pkg/network` delta states

**Technical Approach**:
- Record seed + all player inputs with timestamps
- Replay by re-executing inputs against same seed
- Binary format with header (version, seed, duration) + input stream
- Support fast-forward (2x, 4x), rewind, pause

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
    seed    int64
    inputs  []InputFrame
    started time.Time
}

func (r *ReplayRecorder) RecordInput(playerID int, input InputState)
func (r *ReplayRecorder) Save(path string) error

type ReplayPlayer struct {
    header  ReplayHeader
    inputs  []InputFrame
    cursor  int
}

func LoadReplay(path string) (*ReplayPlayer, error)
func (p *ReplayPlayer) Step() (InputFrame, bool)
```

### 4. Leaderboards — Score Aggregation
- **Deliverable**: `pkg/leaderboard/` package with local and federated leaderboards
- **Dependencies**: `pkg/federation` for cross-server aggregation

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

func (lb *Leaderboard) RecordScore(playerID, stat string, value int64)
func (lb *Leaderboard) GetTop(stat, period string, limit int) []LeaderboardEntry
func (lb *Leaderboard) GetRank(playerID, stat, period string) (int, error)
```

### 5. Achievements System — Local Tracking
- **Deliverable**: `pkg/achievements/` package with unlock conditions and persistence
- **Dependencies**: `pkg/save` for persistence

**Technical Approach**:
- Define achievements as condition functions
- Track progress counters (e.g., "Kill 100 enemies": progress 47/100)
- Persist unlocks in save file
- Trigger toast notification on unlock

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

### 6. Profanity Filter Enhancement — Localized Word Lists
- **Deliverable**: Enhanced `pkg/chat/filter.go` with comprehensive procedural word generation
- **Dependencies**: Existing filter framework

**Technical Approach**:
- Procedurally generate language-specific word patterns using seed-based phoneme combinations
- Generate variant patterns (l33t speak: a→4, e→3, i→1, o→0)
- Support 5 languages: English, Spanish, German, French, Portuguese
- Generate ~500 patterns per language using deterministic algorithms

**Implementation**:
```go
func GenerateProfanityPatterns(lang string, seed int64) []string {
    // Generate base patterns from phoneme rules
    // Generate l33t speak variants
    // Generate common misspellings
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
- [ ] Replay files play back identically to original game (deterministic)
- [ ] Leaderboards persist across game restarts
- [ ] Achievements unlock correctly when conditions met
- [ ] Profanity filter detects l33t speak variants (e.g., "b4d" matches "bad")
- [x] All new code has >82% test coverage ✅ (matchmaking: 88-100%, anticheat: 89-100%)
- [ ] Integration tests verify 4-player matchmaking queue

## Known Gaps

- **DHT Federation**: Distributed hash table discovery deferred to v6.1 (HTTP hubs sufficient for launch)
- **Mobile Store Publishing**: iOS App Store / Google Play submission process not documented (requires developer accounts)
- **Mod Marketplace**: Centralized mod distribution deferred to v7.0 (local mods work now)
