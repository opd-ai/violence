// Package event provides a world event and trigger system.
package event

import (
	"math/rand"
	"sync"
)

// EventType defines the type of world event.
type EventType int

const (
	EventAlarm EventType = iota
	EventLockdown
	EventBossArena
)

// Event represents a world event.
type Event struct {
	ID          string
	Name        string
	Type        EventType
	Description string
	Active      bool
}

// Trigger defines a condition that fires an event.
type Trigger struct {
	EventID   string
	Condition string
}

// AlarmTrigger puts all enemies in alert state and locks doors temporarily.
type AlarmTrigger struct {
	ID       string
	Duration float64 // seconds
	Elapsed  float64
	Active   bool
	mu       sync.RWMutex
}

// NewAlarmTrigger creates a new alarm trigger with given duration.
func NewAlarmTrigger(id string, duration float64) *AlarmTrigger {
	return &AlarmTrigger{
		ID:       id,
		Duration: duration,
		Active:   false,
	}
}

// Activate starts the alarm.
func (a *AlarmTrigger) Activate() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Active = true
	a.Elapsed = 0
}

// Update advances the alarm timer.
func (a *AlarmTrigger) Update(deltaTime float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.Active {
		return
	}
	a.Elapsed += deltaTime
	if a.Elapsed >= a.Duration {
		a.Active = false
		a.Elapsed = 0
	}
}

// IsActive returns whether the alarm is currently active.
func (a *AlarmTrigger) IsActive() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Active
}

// GetProgress returns completion percentage (0.0 to 1.0).
func (a *AlarmTrigger) GetProgress() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.Duration == 0 {
		return 1.0
	}
	return a.Elapsed / a.Duration
}

// TimedLockdown implements a countdown timer with escape objective.
type TimedLockdown struct {
	ID           string
	CountdownS   float64 // seconds
	Remaining    float64
	Active       bool
	wasActivated bool
	mu           sync.RWMutex
}

// NewTimedLockdown creates a new timed lockdown event.
func NewTimedLockdown(id string, countdownS float64) *TimedLockdown {
	return &TimedLockdown{
		ID:         id,
		CountdownS: countdownS,
		Active:     false,
	}
}

// Activate starts the lockdown countdown.
func (t *TimedLockdown) Activate() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Active = true
	t.wasActivated = true
	t.Remaining = t.CountdownS
}

// Update decrements the countdown timer.
func (t *TimedLockdown) Update(deltaTime float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.Active {
		return
	}
	t.Remaining -= deltaTime
	if t.Remaining <= 0 {
		t.Active = false
		t.Remaining = 0
	}
}

// IsActive returns whether the lockdown is in progress.
func (t *TimedLockdown) IsActive() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Active
}

// GetRemaining returns remaining time in seconds.
func (t *TimedLockdown) GetRemaining() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Remaining
}

// IsExpired returns true if countdown reached zero after being activated.
func (t *TimedLockdown) IsExpired() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.wasActivated && t.Remaining <= 0 && !t.Active
}

// BossArenaEvent spawns enemy wave on room entry.
type BossArenaEvent struct {
	ID         string
	RoomID     string
	WaveCount  int
	Triggered  bool
	SpawnDelay float64 // seconds between waves
	mu         sync.RWMutex
}

// NewBossArenaEvent creates a new boss arena event.
func NewBossArenaEvent(id, roomID string, waveCount int, spawnDelay float64) *BossArenaEvent {
	return &BossArenaEvent{
		ID:         id,
		RoomID:     roomID,
		WaveCount:  waveCount,
		SpawnDelay: spawnDelay,
		Triggered:  false,
	}
}

// Trigger activates the boss arena event.
func (b *BossArenaEvent) Trigger() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Triggered = true
}

// IsTriggered returns whether the event has been triggered.
func (b *BossArenaEvent) IsTriggered() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Triggered
}

// GetWaveCount returns the number of enemy waves.
func (b *BossArenaEvent) GetWaveCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.WaveCount
}

// GetSpawnDelay returns delay between waves.
func (b *BossArenaEvent) GetSpawnDelay() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.SpawnDelay
}

// Fire dispatches an event by ID.
func Fire(eventID string) {}

var (
	currentGenre = "scifi"
	genreMu      sync.RWMutex
)

// SetGenre configures world events for a genre.
func SetGenre(genreID string) {
	genreMu.Lock()
	defer genreMu.Unlock()
	currentGenre = genreID
}

// GetGenre returns the current genre.
func GetGenre() string {
	genreMu.RLock()
	defer genreMu.RUnlock()
	return currentGenre
}

// GenerateEventText generates genre-flavored event text procedurally.
func GenerateEventText(seed uint64, eventType EventType) string {
	rng := rand.New(rand.NewSource(int64(seed)))
	genre := GetGenre()

	templates := getEventTemplates(genre, eventType)
	if len(templates) == 0 {
		return "Event triggered"
	}

	return templates[rng.Intn(len(templates))]
}

func getEventTemplates(genre string, eventType EventType) []string {
	switch genre {
	case "fantasy":
		switch eventType {
		case EventAlarm:
			return []string{
				"The bell tower rings! Guards have been alerted!",
				"Warning bells echo through the halls!",
				"The alarm gong sounds—guards converge!",
			}
		case EventLockdown:
			return []string{
				"The gates are sealing! Escape before they close!",
				"Portcullis descending! Find another way out!",
				"Magical wards activate—you must flee!",
			}
		case EventBossArena:
			return []string{
				"The chamber door slams shut! Prepare for battle!",
				"A dark presence emerges from the shadows!",
				"Ancient guardians awaken to defend this sanctum!",
			}
		}
	case "scifi":
		switch eventType {
		case EventAlarm:
			return []string{
				"ALERT: Security breach detected. All units respond.",
				"WARNING: Intruder alert. Lockdown protocol initiated.",
				"EMERGENCY: Hostile presence confirmed. Engaging defenses.",
			}
		case EventLockdown:
			return []string{
				"LOCKDOWN: All blast doors sealing in 60 seconds.",
				"CONTAINMENT: Station lockdown in progress. Evacuate immediately.",
				"EMERGENCY: Sector isolation in 60 seconds. Find exit now.",
			}
		case EventBossArena:
			return []string{
				"SYSTEM: Combat arena activated. Prepare for engagement.",
				"WARNING: High-threat entity detected. Recommend tactical withdrawal.",
				"ALERT: Boss protocol engaged. All exits sealed.",
			}
		}
	case "horror":
		switch eventType {
		case EventAlarm:
			return []string{
				"Distant screams echo... something knows you're here.",
				"The lights flicker. They're coming.",
				"A wailing siren pierces the darkness. RUN.",
			}
		case EventLockdown:
			return []string{
				"The walls are closing in. Find a way out before it's too late.",
				"Doors slam shut throughout the facility. You're being trapped.",
				"The building itself is alive... and it wants you inside.",
			}
		case EventBossArena:
			return []string{
				"The door vanishes behind you. Something massive stirs ahead.",
				"You hear breathing... everywhere. The room seals itself.",
				"A grotesque shadow fills the chamber. There's no escape.",
			}
		}
	case "cyberpunk":
		switch eventType {
		case EventAlarm:
			return []string{
				"ICE detected! Corp security is tracing your signal!",
				"ALERT: NetWatch dispatched to your location!",
				"WARNING: Security spiders deployed. Evade immediately.",
			}
		case EventLockdown:
			return []string{
				"LOCKDOWN: Building AI sealing all exits in 60 seconds!",
				"SYSTEM: Emergency partition in progress. Escape now!",
				"WARNING: Megacorp lockdown protocol active. Find exit fast!",
			}
		case EventBossArena:
			return []string{
				"SYSTEM: Combat daemon initialized. Good luck, choom.",
				"WARNING: High-level ICE detected. This is gonna hurt.",
				"ALERT: Corporate enforcer inbound. Time to flatline or jack out.",
			}
		}
	case "postapoc":
		switch eventType {
		case EventAlarm:
			return []string{
				"Raiders spotted! They know you're here!",
				"Warning: Hostile scavengers converging on your position!",
				"The bandit alarm sounds—expect company!",
			}
		case EventLockdown:
			return []string{
				"The bunker is sealing! Get out before the blast doors close!",
				"Emergency lockdown! 60 seconds to escape!",
				"Vault entrance closing! Move or be trapped!",
			}
		case EventBossArena:
			return []string{
				"The warlord appears! This is a fight to the death!",
				"A massive mutant blocks your path. Kill or be killed.",
				"The raider boss emerges. No retreat, no surrender.",
			}
		}
	}

	// Default fallback
	switch eventType {
	case EventAlarm:
		return []string{"Alarm triggered!"}
	case EventLockdown:
		return []string{"Lockdown initiated!"}
	case EventBossArena:
		return []string{"Boss encounter!"}
	}

	return []string{"Event triggered"}
}

// GenerateEventAudioSting generates a procedural audio description for event.
// Returns audio parameters for procedural generation.
func GenerateEventAudioSting(seed uint64, eventType EventType) AudioSting {
	rng := rand.New(rand.NewSource(int64(seed)))
	genre := GetGenre()

	var sting AudioSting
	sting.Seed = seed

	switch eventType {
	case EventAlarm:
		sting.Type = "alarm"
		sting.Frequency = 440.0 + float64(rng.Intn(220)) // 440-660 Hz
		sting.Duration = 0.3 + rng.Float64()*0.2         // 0.3-0.5 seconds
		sting.Pattern = getAlarmPattern(genre, rng)
	case EventLockdown:
		sting.Type = "lockdown"
		sting.Frequency = 220.0 + float64(rng.Intn(110)) // 220-330 Hz (lower, ominous)
		sting.Duration = 1.0 + rng.Float64()*0.5         // 1.0-1.5 seconds
		sting.Pattern = getLockdownPattern(genre, rng)
	case EventBossArena:
		sting.Type = "boss"
		sting.Frequency = 110.0 + float64(rng.Intn(55)) // 110-165 Hz (very low, dramatic)
		sting.Duration = 2.0 + rng.Float64()            // 2.0-3.0 seconds
		sting.Pattern = getBossPattern(genre, rng)
	}

	return sting
}

// AudioSting contains procedural audio parameters.
type AudioSting struct {
	Seed      uint64
	Type      string
	Frequency float64
	Duration  float64
	Pattern   string
}

func getAlarmPattern(genre string, rng *rand.Rand) string {
	patterns := map[string][]string{
		"fantasy":   {"bell", "gong", "horn"},
		"scifi":     {"siren", "pulse", "beep"},
		"horror":    {"screech", "wail", "static"},
		"cyberpunk": {"glitch", "digital", "synth"},
		"postapoc":  {"klaxon", "clanging", "rattle"},
	}

	if p, ok := patterns[genre]; ok {
		return p[rng.Intn(len(p))]
	}
	return "siren"
}

func getLockdownPattern(genre string, rng *rand.Rand) string {
	patterns := map[string][]string{
		"fantasy":   {"deep_gong", "grinding_stone", "heavy_chains"},
		"scifi":     {"hydraulic", "mechanical", "airlock"},
		"horror":    {"grinding_metal", "distant_slam", "echoing_boom"},
		"cyberpunk": {"electronic_lock", "system_shutdown", "power_down"},
		"postapoc":  {"rusty_door", "metal_slam", "vault_seal"},
	}

	if p, ok := patterns[genre]; ok {
		return p[rng.Intn(len(p))]
	}
	return "mechanical"
}

func getBossPattern(genre string, rng *rand.Rand) string {
	patterns := map[string][]string{
		"fantasy":   {"orchestral_hit", "thunder", "war_horn"},
		"scifi":     {"power_surge", "system_critical", "emergency_tone"},
		"horror":    {"crescendo_strings", "heartbeat", "breathing"},
		"cyberpunk": {"bass_drop", "synth_swell", "digital_roar"},
		"postapoc":  {"distant_explosion", "engine_roar", "metal_grind"},
	}

	if p, ok := patterns[genre]; ok {
		return p[rng.Intn(len(p))]
	}
	return "dramatic"
}
