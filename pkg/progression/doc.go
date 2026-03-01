// Package progression manages player experience points (XP) and automatic leveling.
//
// # Overview
//
// The progression system tracks a player's XP accumulation and automatically
// handles level-ups when XP thresholds are reached. XP requirements scale linearly
// with level (100 XP for level 2, 200 XP for level 3, etc.).
//
// # Key Features
//
//   - Automatic level-up when XP threshold is reached
//   - Thread-safe concurrent access to XP and level data
//   - Genre-specific configuration support for future XP scaling
//   - Maximum level cap of 99
//   - XP validation (prevents negative XP)
//
// # Basic Usage
//
//	p := progression.NewProgression()
//
//	// Grant XP from kills
//	if err := p.AddXP(100); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Automatically leveled to 2 if threshold reached
//	fmt.Printf("Level: %d, XP: %d\n", p.GetLevel(), p.GetXP())
//
//	// Check XP needed for next level
//	fmt.Printf("XP for next level: %d\n", p.XPForNextLevel())
//
// # Genre Support
//
// The package supports genre-specific configuration for future XP curve tuning:
//
//	if err := p.SetGenre("horror"); err != nil {
//	    log.Fatal(err)
//	}
//
// Valid genres: fantasy, scifi, horror, cyberpunk, postapoc
//
// # Thread Safety
//
// All methods are safe for concurrent access from multiple goroutines.
// The Progression struct uses internal locking to protect XP and level state.
//
// # Design Rationale
//
// XP and level fields are private to enforce encapsulation and ensure all
// mutations go through validated methods. Auto-leveling on AddXP removes
// manual level-up logic from calling code and prevents state desync.
package progression
