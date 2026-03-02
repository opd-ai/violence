package attacktrail

import (
	"image/color"
	"math"
	"math/rand"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
)

// AttachTrailToAttack creates and attaches a trail to an entity based on weapon type.
// This is a convenience function for integrating attack trails with the combat system.
func AttachTrailToAttack(world *engine.World, entityID engine.Entity, x, y, dirX, dirY, range_ float64, weaponType string, rng *rand.Rand, genreColors func(string, *rand.Rand) color.RGBA) {
	// Get or create trail component
	var trailComp *TrailComponent

	comp, found := world.GetComponent(entityID, reflect.TypeOf(&TrailComponent{}))
	if found {
		if tc, ok := comp.(*TrailComponent); ok {
			trailComp = tc
		}
	}

	if trailComp == nil {
		trailComp = NewTrailComponent(3) // Max 3 simultaneous trails per entity
		world.AddComponent(entityID, trailComp)
	}

	// Calculate attack angle
	angle := math.Atan2(dirY, dirX)

	// Get weapon-appropriate color
	weaponColor := genreColors(weaponType, rng)

	// Create trail based on weapon type
	var trail *Trail

	switch weaponType {
	case "sword", "katana", "scimitar":
		// Slashing weapons - arc trail
		arc := math.Pi / 2.5 // ~72 degrees
		trail = CreateSlashTrail(x, y, angle, range_*1.2, arc, 3.0, weaponColor)

	case "greatsword", "claymore", "battleaxe":
		// Heavy weapons - wide cleave
		arc := math.Pi / 2.0 // 90 degrees
		trail = CreateCleaveTrail(x, y, angle, range_*1.3, arc, 5.0, weaponColor)

	case "spear", "lance", "rapier":
		// Stabbing weapons - thrust trail
		trail = CreateThrustTrail(x, y, angle, range_*1.4, 2.5, weaponColor)

	case "hammer", "mace", "club":
		// Blunt weapons - smash impact
		trail = CreateSmashTrail(x, y, range_*0.8, 4.0, weaponColor)

	case "staff", "quarterstaff":
		// Spinning weapons
		trail = CreateSpinTrail(x, y, range_*1.1, 3.5, weaponColor)

	case "dagger", "knife":
		// Quick slash with smaller arc
		arc := math.Pi / 4.0 // 45 degrees
		trail = CreateSlashTrail(x, y, angle, range_*0.9, arc, 2.0, weaponColor)

	case "whip", "chain":
		// Extended slash
		arc := math.Pi / 1.5 // ~120 degrees
		trail = CreateSlashTrail(x, y, angle, range_*1.6, arc, 2.5, weaponColor)

	case "fist", "unarmed":
		// Short punch trail
		trail = CreateThrustTrail(x, y, angle, range_*0.6, 3.0, weaponColor)

	default:
		// Default to slash
		arc := math.Pi / 3.0
		trail = CreateSlashTrail(x, y, angle, range_, arc, 3.0, weaponColor)
	}

	if trail != nil {
		trailComp.AddTrail(trail)
	}
}

// WeaponTypeFromName attempts to classify a weapon by name.
func WeaponTypeFromName(name string) string {
	// Simple classification based on name keywords
	nameLower := name

	switch {
	case contains(nameLower, "sword") && !contains(nameLower, "great"):
		return "sword"
	case contains(nameLower, "great") || contains(nameLower, "claymore"):
		return "greatsword"
	case contains(nameLower, "spear") || contains(nameLower, "lance"):
		return "spear"
	case contains(nameLower, "rapier"):
		return "rapier"
	case contains(nameLower, "hammer") || contains(nameLower, "mace"):
		return "hammer"
	case contains(nameLower, "axe"):
		if contains(nameLower, "battle") {
			return "battleaxe"
		}
		return "sword" // Treat as slashing
	case contains(nameLower, "dagger") || contains(nameLower, "knife"):
		return "dagger"
	case contains(nameLower, "staff"):
		return "staff"
	case contains(nameLower, "whip") || contains(nameLower, "chain"):
		return "whip"
	case contains(nameLower, "fist") || contains(nameLower, "unarmed"):
		return "fist"
	default:
		return "sword"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}
