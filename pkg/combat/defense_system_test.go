package combat

import (
	"math"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestDefenseSystem_Update(t *testing.T) {
	world := engine.NewWorld()
	system := NewDefenseSystem("fantasy")

	// Create entity with defense component
	e := world.AddEntity()
	pos := &engine.Position{X: 10, Y: 10}
	vel := &engine.Velocity{DX: 0, DY: 0}
	def := NewDefenseComponent("fantasy")

	world.AddComponent(e, pos)
	world.AddComponent(e, vel)
	world.AddComponent(e, def)

	// Test stamina regeneration
	def.StaminaCurrent = 50.0
	system.Update(world)

	if def.StaminaCurrent <= 50.0 {
		t.Error("stamina should regenerate during inactive state")
	}

	// Test cooldown countdown
	def.CooldownTimer = 1.0
	for i := 0; i < 60; i++ {
		system.Update(world)
	}
	if def.CooldownTimer > 0.1 {
		t.Errorf("cooldown should decrease, got %f", def.CooldownTimer)
	}
}

func TestDefenseSystem_InitiateDodge(t *testing.T) {
	system := NewDefenseSystem("fantasy")
	def := NewDefenseComponent("fantasy")

	// Should succeed with full stamina
	if !system.InitiateDodge(def, 1.0, 0.0, "fantasy") {
		t.Error("dodge should succeed with full stamina")
	}

	if def.Type != DefenseDodge {
		t.Error("defense type should be dodge")
	}
	if def.State != DefenseWindup {
		t.Error("defense state should be windup")
	}

	// Check dodge velocity is normalized
	mag := math.Sqrt(def.DodgeVelocityX*def.DodgeVelocityX + def.DodgeVelocityY*def.DodgeVelocityY)
	if math.Abs(mag-1.0) > 0.01 {
		t.Errorf("dodge velocity should be normalized, got magnitude %f", mag)
	}

	// Should fail on cooldown
	def.CooldownTimer = 0.5
	def.State = DefenseInactive
	def.Type = DefenseNone
	if system.InitiateDodge(def, 1.0, 0.0, "fantasy") {
		t.Error("dodge should fail on cooldown")
	}
}

func TestDefenseSystem_InitiateParry(t *testing.T) {
	system := NewDefenseSystem("fantasy")
	def := NewDefenseComponent("fantasy")

	// Should succeed with full stamina
	if !system.InitiateParry(def, "fantasy") {
		t.Error("parry should succeed with full stamina")
	}

	if def.Type != DefenseParry {
		t.Error("defense type should be parry")
	}
}

func TestDefenseSystem_InitiateBlock(t *testing.T) {
	system := NewDefenseSystem("fantasy")
	def := NewDefenseComponent("fantasy")

	// Should succeed with full stamina
	if !system.InitiateBlock(def, "fantasy") {
		t.Error("block should succeed with full stamina")
	}

	if def.Type != DefenseBlock {
		t.Error("defense type should be block")
	}
}

func TestDefenseSystem_CancelBlock(t *testing.T) {
	system := NewDefenseSystem("fantasy")
	def := NewDefenseComponent("fantasy")

	def.Type = DefenseBlock
	def.State = DefenseActive
	system.CancelBlock(def)

	if def.State != DefenseRecovery {
		t.Error("block cancel should enter recovery state")
	}
}

func TestDefenseSystem_StateMachine(t *testing.T) {
	world := engine.NewWorld()
	system := NewDefenseSystem("fantasy")

	e := world.AddEntity()
	pos := &engine.Position{X: 10, Y: 10}
	vel := &engine.Velocity{DX: 0, DY: 0}
	def := NewDefenseComponent("fantasy")

	world.AddComponent(e, pos)
	world.AddComponent(e, vel)
	world.AddComponent(e, def)

	// Initiate dodge
	system.InitiateDodge(def, 1.0, 0.0, "fantasy")

	// Run until windUp completes
	for i := 0; i < 10; i++ {
		system.Update(world)
		if def.State == DefenseActive {
			break
		}
	}

	if def.State != DefenseActive {
		t.Error("dodge should transition to active state")
	}

	// Run until active completes
	for i := 0; i < 30; i++ {
		system.Update(world)
		if def.State == DefenseRecovery {
			break
		}
	}

	if def.State != DefenseRecovery {
		t.Error("dodge should transition to recovery state")
	}

	// Run until recovery completes
	for i := 0; i < 20; i++ {
		system.Update(world)
		if def.State == DefenseInactive {
			break
		}
	}

	if def.State != DefenseInactive {
		t.Error("dodge should return to inactive state")
	}
}

func TestDefenseSystem_BlockStaminaDrain(t *testing.T) {
	world := engine.NewWorld()
	system := NewDefenseSystem("fantasy")

	e := world.AddEntity()
	pos := &engine.Position{X: 10, Y: 10}
	def := NewDefenseComponent("fantasy")

	world.AddComponent(e, pos)
	world.AddComponent(e, def)

	system.InitiateBlock(def, "fantasy")

	// Run until active
	for i := 0; i < 10; i++ {
		system.Update(world)
		if def.State == DefenseActive {
			break
		}
	}

	initialStamina := def.StaminaCurrent

	// Block should drain stamina
	for i := 0; i < 60; i++ {
		system.Update(world)
	}

	if def.StaminaCurrent >= initialStamina {
		t.Error("block should drain stamina")
	}
}

func TestDefenseSystem_Integration(t *testing.T) {
	world := engine.NewWorld()
	system := NewDefenseSystem("scifi")

	// Create player with defense
	player := world.AddEntity()
	playerPos := &engine.Position{X: 10, Y: 10}
	playerDef := NewDefenseComponent("scifi")
	world.AddComponent(player, playerPos)
	world.AddComponent(player, playerDef)

	// Initiate parry
	if !system.InitiateParry(playerDef, "scifi") {
		t.Fatal("failed to initiate parry")
	}

	// Advance to active state
	for i := 0; i < 10; i++ {
		system.Update(world)
	}

	// Check that we're in active parry state
	if !playerDef.IsParrying() {
		t.Error("should be in active parry state")
	}

	// Simulate incoming attack during parry window
	damage := 50.0
	modDamage, negated := ProcessIncomingDamage(playerDef, damage, 0, 0)

	if negated {
		t.Log("attack parried successfully")
	} else if modDamage < damage {
		t.Logf("attack partially parried: %f -> %f", damage, modDamage)
	} else {
		t.Error("parry should reduce damage")
	}
}

func TestDefensePresetGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			system := NewDefenseSystem(genre)
			if system.genreID != genre {
				t.Errorf("expected genre %s, got %s", genre, system.genreID)
			}

			def := NewDefenseComponent(genre)
			if def.StaminaMax <= 0 {
				t.Error("stamina max should be positive")
			}
		})
	}
}
