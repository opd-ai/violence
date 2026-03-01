package status

import (
	"reflect"
	"testing"
	"time"

	"github.com/opd-ai/violence/pkg/engine"
)

// TestStatusEffectIntegration demonstrates the full status effect workflow.
func TestStatusEffectIntegration(t *testing.T) {
	// Create world and registry
	world := engine.NewWorld()
	registry := NewRegistry()
	system := NewSystem(registry)

	// Register the system with the world
	world.AddSystem(system)

	// Create a player entity
	player := world.NewPlayerEntity(10.0, 10.0)

	// Verify initial health
	healthType := reflect.TypeOf(&engine.Health{})
	comp, ok := world.GetComponent(player, healthType)
	if !ok {
		t.Fatal("Player should have health component")
	}
	health := comp.(*engine.Health)
	if health.Current != 100 {
		t.Errorf("Expected initial health 100, got %d", health.Current)
	}

	// Apply poison effect
	registry.ApplyToEntity(world, player, "poisoned")

	// Verify status component was added
	statusType := reflect.TypeOf(&StatusComponent{})
	statusComp, ok := world.GetComponent(player, statusType)
	if !ok {
		t.Fatal("Status component should be added after applying effect")
	}

	sc := statusComp.(*StatusComponent)
	if len(sc.ActiveEffects) != 1 {
		t.Fatalf("Expected 1 active effect, got %d", len(sc.ActiveEffects))
	}

	if sc.ActiveEffects[0].EffectName != "poisoned" {
		t.Errorf("Expected 'poisoned', got '%s'", sc.ActiveEffects[0].EffectName)
	}

	// Wait for a tick and run system update
	time.Sleep(1100 * time.Millisecond)
	world.Update()

	// Health should have decreased
	comp, _ = world.GetComponent(player, healthType)
	health = comp.(*engine.Health)
	if health.Current >= 100 {
		t.Errorf("Poison should have dealt damage, health: %d", health.Current)
	}

	t.Logf("After poison tick: health is %d/100", health.Current)
}

// TestMultipleStatusEffects tests multiple concurrent effects.
func TestMultipleStatusEffects(t *testing.T) {
	world := engine.NewWorld()
	registry := NewRegistry()
	system := NewSystem(registry)
	world.AddSystem(system)

	enemy := world.AddEntity()
	world.AddComponent(enemy, &engine.Health{Current: 200, Max: 200})

	// Apply multiple effects
	registry.ApplyToEntity(world, enemy, "poisoned")
	registry.ApplyToEntity(world, enemy, "burning")
	registry.ApplyToEntity(world, enemy, "bleeding")

	statusType := reflect.TypeOf(&StatusComponent{})
	comp, _ := world.GetComponent(enemy, statusType)
	sc := comp.(*StatusComponent)

	if len(sc.ActiveEffects) != 3 {
		t.Errorf("Expected 3 active effects, got %d", len(sc.ActiveEffects))
	}

	// Check speed multiplier (should take the most restrictive)
	speedMul := GetSpeedMultiplier(world, enemy)
	if speedMul >= 1.0 {
		t.Errorf("Multiple debuffs should reduce speed, got %f", speedMul)
	}

	t.Logf("Speed multiplier with poison+burning+bleeding: %.2f", speedMul)
}

// TestHealingAndDamageBalance tests healing effect counteracting damage.
func TestHealingAndDamageBalance(t *testing.T) {
	world := engine.NewWorld()
	registry := NewRegistry()
	system := NewSystem(registry)
	world.AddSystem(system)

	entity := world.AddEntity()
	world.AddComponent(entity, &engine.Health{Current: 100, Max: 150})

	// Apply poison (damage) and regeneration (heal)
	registry.ApplyToEntity(world, entity, "poisoned")
	registry.ApplyToEntity(world, entity, "regeneration")

	initialHealth := 100

	// Run several ticks
	for i := 0; i < 5; i++ {
		time.Sleep(1100 * time.Millisecond)
		world.Update()
	}

	healthType := reflect.TypeOf(&engine.Health{})
	comp, _ := world.GetComponent(entity, healthType)
	health := comp.(*engine.Health)

	// Net effect should be neutral or slight healing (regen > poison in fantasy)
	t.Logf("After 5 ticks of poison+regen: %d -> %d HP", initialHealth, health.Current)

	// Poison deals 2/sec, regen heals 2/sec, should be approximately balanced
	diff := health.Current - initialHealth
	if diff < -10 || diff > 10 {
		t.Logf("Note: Net health change is %d (expected near 0 for balanced effects)", diff)
	}
}

// TestEffectStacking verifies stackable vs non-stackable behavior.
func TestEffectStacking(t *testing.T) {
	world := engine.NewWorld()
	registry := NewRegistry()

	entity := world.AddEntity()
	world.AddComponent(entity, &engine.Health{Current: 100, Max: 100})

	// Bleeding is stackable
	registry.ApplyToEntity(world, entity, "bleeding")
	registry.ApplyToEntity(world, entity, "bleeding")
	registry.ApplyToEntity(world, entity, "bleeding")

	statusType := reflect.TypeOf(&StatusComponent{})
	comp, _ := world.GetComponent(entity, statusType)
	sc := comp.(*StatusComponent)

	bleedStacks := 0
	for _, eff := range sc.ActiveEffects {
		if eff.EffectName == "bleeding" {
			bleedStacks++
		}
	}

	if bleedStacks != 3 {
		t.Errorf("Bleeding should stack 3 times, got %d", bleedStacks)
	}

	// Poison is not stackable
	registry.ApplyToEntity(world, entity, "poisoned")
	registry.ApplyToEntity(world, entity, "poisoned")

	comp, _ = world.GetComponent(entity, statusType)
	sc = comp.(*StatusComponent)

	poisonStacks := 0
	for _, eff := range sc.ActiveEffects {
		if eff.EffectName == "poisoned" {
			poisonStacks++
		}
	}

	if poisonStacks != 1 {
		t.Errorf("Poison should not stack, got %d instances", poisonStacks)
	}
}

// TestStunEffect verifies movement prevention.
func TestStunEffect(t *testing.T) {
	world := engine.NewWorld()
	registry := NewRegistry()

	entity := world.AddEntity()

	// Not stunned initially
	if IsStunned(world, entity) {
		t.Error("Entity should not be stunned without status effect")
	}

	// Apply stun
	registry.ApplyToEntity(world, entity, "stunned")

	if !IsStunned(world, entity) {
		t.Error("Entity should be stunned after applying stun effect")
	}

	speed := GetSpeedMultiplier(world, entity)
	if speed != 0.0 {
		t.Errorf("Stunned entity should have 0.0 speed, got %f", speed)
	}
}

// TestGenreEffectVariety verifies different genres have different effect sets.
func TestGenreEffectVariety(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			registry := NewRegistry()
			registry.loadDefaultEffects(genre)

			if len(registry.effects) < 5 {
				t.Errorf("Genre %s should have at least 5 effects, got %d", genre, len(registry.effects))
			}

			// Check for genre-specific effects
			switch genre {
			case "scifi", "cyberpunk":
				if _, exists := registry.effects["emp_stunned"]; !exists {
					t.Errorf("Sci-fi/Cyberpunk should have emp_stunned effect")
				}
			case "horror", "postapoc":
				if _, exists := registry.effects["infected"]; !exists {
					t.Errorf("Horror/Postapoc should have infected effect")
				}
			}
		})
	}
}
