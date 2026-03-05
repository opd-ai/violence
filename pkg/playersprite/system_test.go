package playersprite

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestComponent_Type(t *testing.T) {
	c := &Component{}
	if got := c.Type(); got != "PlayerSprite" {
		t.Errorf("Type() = %v, want %v", got, "PlayerSprite")
	}
}

func TestAnimationStates(t *testing.T) {
	tests := []struct {
		name  string
		state AnimationState
		want  int
	}{
		{"Idle", AnimIdle, 0},
		{"Walk", AnimWalk, 1},
		{"Attack", AnimAttack, 2},
		{"Hurt", AnimHurt, 3},
		{"Death", AnimDeath, 4},
		{"Dodge", AnimDodge, 5},
		{"Cast", AnimCast, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.state) != tt.want {
				t.Errorf("AnimationState %s = %d, want %d", tt.name, tt.state, tt.want)
			}
		})
	}
}

func TestComponent_Fields(t *testing.T) {
	c := &Component{
		Class:          "warrior",
		EquippedWeapon: "sword",
		EquippedArmor:  "heavy",
		CurrentFrame:   5,
		AnimState:      AnimAttack,
		Seed:           12345,
		Facing:         1,
		DirtyFlag:      true,
	}

	if c.Class != "warrior" {
		t.Errorf("Class = %v, want warrior", c.Class)
	}
	if c.EquippedWeapon != "sword" {
		t.Errorf("EquippedWeapon = %v, want sword", c.EquippedWeapon)
	}
	if c.EquippedArmor != "heavy" {
		t.Errorf("EquippedArmor = %v, want heavy", c.EquippedArmor)
	}
	if c.CurrentFrame != 5 {
		t.Errorf("CurrentFrame = %v, want 5", c.CurrentFrame)
	}
	if c.AnimState != AnimAttack {
		t.Errorf("AnimState = %v, want AnimAttack", c.AnimState)
	}
	if c.Seed != 12345 {
		t.Errorf("Seed = %v, want 12345", c.Seed)
	}
	if c.Facing != 1 {
		t.Errorf("Facing = %v, want 1", c.Facing)
	}
	if !c.DirtyFlag {
		t.Errorf("DirtyFlag = %v, want true", c.DirtyFlag)
	}
}

func TestSystem_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires graphics context")
	}

	sys := NewSystem("fantasy")
	world := engine.NewWorld()

	// Create entity with PlayerSprite component
	eid := world.AddEntity()
	comp := &Component{
		Class:          "warrior",
		EquippedWeapon: "sword",
		EquippedArmor:  "heavy",
		AnimState:      AnimIdle,
		CurrentFrame:   0,
		Seed:           42,
		DirtyFlag:      true,
	}
	world.AddComponent(eid, comp)

	// Update should generate sprite
	sys.Update(world)

	// Verify sprite was generated
	componentType := reflect.TypeOf(&Component{})
	updatedComp, _ := world.GetComponent(eid, componentType)
	playerSprite := updatedComp.(*Component)

	if playerSprite.CachedSprite == nil {
		t.Error("Expected CachedSprite to be generated")
	}
	if playerSprite.DirtyFlag {
		t.Error("Expected DirtyFlag to be cleared after generation")
	}
}

func TestSystem_SetGenre(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetGenre("scifi")

	if sys.genreID != "scifi" {
		t.Errorf("genreID = %v, want scifi", sys.genreID)
	}
	if sys.generator.genreID != "scifi" {
		t.Errorf("generator.genreID = %v, want scifi", sys.generator.genreID)
	}
}

func TestSystem_Type(t *testing.T) {
	sys := NewSystem("fantasy")
	if got := sys.Type(); got != "PlayerSpriteSystem" {
		t.Errorf("Type() = %v, want PlayerSpriteSystem", got)
	}
}

func TestSystem_UpdateWithoutComponent(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()

	// Entity without PlayerSprite component
	world.AddEntity()

	// Should not panic
	sys.Update(world)
}

func TestSystem_UpdateCachedSprite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires graphics context")
	}

	sys := NewSystem("fantasy")
	world := engine.NewWorld()

	eid := world.AddEntity()
	comp := &Component{
		Class:          "mage",
		EquippedWeapon: "staff",
		EquippedArmor:  "light",
		AnimState:      AnimWalk,
		CurrentFrame:   3,
		Seed:           99,
		DirtyFlag:      true,
	}
	world.AddComponent(eid, comp)

	// First update generates sprite
	sys.Update(world)

	componentType := reflect.TypeOf(&Component{})
	updatedComp, _ := world.GetComponent(eid, componentType)
	playerSprite := updatedComp.(*Component)
	firstSprite := playerSprite.CachedSprite

	// Second update without dirty flag should not regenerate
	playerSprite.DirtyFlag = false
	sys.Update(world)

	if playerSprite.CachedSprite != firstSprite {
		t.Error("Expected sprite to be cached when DirtyFlag is false")
	}

	// Third update with dirty flag should regenerate
	playerSprite.DirtyFlag = true
	playerSprite.EquippedWeapon = "wand"
	sys.Update(world)

	if playerSprite.CachedSprite == nil {
		t.Error("Expected new sprite to be generated")
	}
}

