package dmgfx

import (
	"image/color"
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

// Mock particle spawner for testing.
type mockParticleSpawner struct {
	burstCalls        int
	lastColor         color.RGBA
	lastCount         int
	lastParticleSpeed float64
}

func (m *mockParticleSpawner) SpawnBurst(x, y, z float64, count int, speed, spread, life, size float64, col color.RGBA) {
	m.burstCalls++
	m.lastColor = col
	m.lastCount = count
	m.lastParticleSpeed = speed
}

// Mock feedback provider for testing.
type mockFeedbackProvider struct {
	shakeCalls      int
	flashCalls      int
	colorFlashCalls int
	lastShake       float64
	lastColorFlash  color.RGBA
}

func (m *mockFeedbackProvider) AddScreenShake(intensity float64) {
	m.shakeCalls++
	m.lastShake = intensity
}

func (m *mockFeedbackProvider) AddHitFlash(intensity float64) {
	m.flashCalls++
}

func (m *mockFeedbackProvider) AddColorFlash(col color.RGBA, intensity float64) {
	m.colorFlashCalls++
	m.lastColorFlash = col
}

func TestNewSystem(t *testing.T) {
	sys := NewSystem()
	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}
	if sys.logger == nil {
		t.Error("System logger not initialized")
	}
}

func TestComponentType(t *testing.T) {
	comp := &DamageVisualComponent{}
	if comp.Type() != "DamageVisualComponent" {
		t.Errorf("Expected type 'DamageVisualComponent', got '%s'", comp.Type())
	}
}

func TestApplyDamageVisual_CreatesComponent(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem()

	mockParticles := &mockParticleSpawner{}
	mockFeedback := &mockFeedbackProvider{}

	sys.SetParticleSpawner(mockParticles)
	sys.SetFeedbackProvider(mockFeedback)

	entity := world.AddEntity()
	world.AddComponent(entity, &engine.Position{X: 10.0, Y: 20.0})

	sys.ApplyDamageVisual(world, entity, "Fire", 50.0, 10.0, 20.0)

	dvType := reflect.TypeOf((*DamageVisualComponent)(nil))
	comp, ok := world.GetComponent(entity, dvType)
	if !ok {
		t.Fatal("DamageVisualComponent not added to entity")
	}

	dv, ok := comp.(*DamageVisualComponent)
	if !ok {
		t.Fatal("Component is not DamageVisualComponent")
	}

	if len(dv.ActiveEffects) != 1 {
		t.Errorf("Expected 1 active effect, got %d", len(dv.ActiveEffects))
	}

	if dv.ActiveEffects[0].DamageTypeName != "Fire" {
		t.Errorf("Expected damage type 'Fire', got '%s'", dv.ActiveEffects[0].DamageTypeName)
	}
}

func TestApplyDamageVisual_SpawnsParticles(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem()

	mockParticles := &mockParticleSpawner{}
	sys.SetParticleSpawner(mockParticles)

	entity := world.AddEntity()
	world.AddComponent(entity, &engine.Position{X: 10.0, Y: 20.0})

	sys.ApplyDamageVisual(world, entity, "Ice", 30.0, 10.0, 20.0)

	if mockParticles.burstCalls != 1 {
		t.Errorf("Expected 1 particle burst, got %d", mockParticles.burstCalls)
	}

	expectedColor := color.RGBA{R: 100, G: 200, B: 255, A: 255}
	if mockParticles.lastColor != expectedColor {
		t.Errorf("Expected ice color %+v, got %+v", expectedColor, mockParticles.lastColor)
	}
}

func TestApplyDamageVisual_ScreenEffects(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem()

	mockFeedback := &mockFeedbackProvider{}
	sys.SetFeedbackProvider(mockFeedback)

	entity := world.AddEntity()
	world.AddComponent(entity, &engine.Position{X: 10.0, Y: 20.0})

	sys.ApplyDamageVisual(world, entity, "Lightning", 60.0, 10.0, 20.0)

	if mockFeedback.shakeCalls != 1 {
		t.Errorf("Expected 1 screen shake, got %d", mockFeedback.shakeCalls)
	}

	if mockFeedback.colorFlashCalls != 1 {
		t.Errorf("Expected 1 color flash for lightning, got %d", mockFeedback.colorFlashCalls)
	}

	if mockFeedback.lastShake <= 0 {
		t.Errorf("Expected positive shake intensity, got %f", mockFeedback.lastShake)
	}
}

func TestDamageProfiles_AllTypes(t *testing.T) {
	damageTypes := []string{
		"Fire", "Ice", "Lightning", "Poison", "Holy", "Shadow", "Arcane", "Physical",
	}

	for _, dt := range damageTypes {
		t.Run(dt, func(t *testing.T) {
			profile := getDamageProfile(dt)

			if profile.Color.A == 0 {
				t.Errorf("Damage type %s has zero alpha", dt)
			}

			if profile.ParticleSpeed <= 0 {
				t.Errorf("Damage type %s has invalid particle speed %f", dt, profile.ParticleSpeed)
			}

			if profile.ParticleLifetime <= 0 {
				t.Errorf("Damage type %s has invalid particle lifetime %f", dt, profile.ParticleLifetime)
			}
		})
	}
}

func TestDamageProfiles_UniqueColors(t *testing.T) {
	fire := getDamageProfile("Fire")
	ice := getDamageProfile("Ice")
	poison := getDamageProfile("Poison")

	if fire.Color == ice.Color {
		t.Error("Fire and Ice should have different colors")
	}

	if fire.Color == poison.Color {
		t.Error("Fire and Poison should have different colors")
	}

	if ice.Color == poison.Color {
		t.Error("Ice and Poison should have different colors")
	}
}

func TestDamageProfiles_LingeringEffects(t *testing.T) {
	tests := []struct {
		damageType   string
		shouldLinger bool
	}{
		{"Fire", true},
		{"Ice", true},
		{"Poison", true},
		{"Physical", false},
	}

	for _, tt := range tests {
		t.Run(tt.damageType, func(t *testing.T) {
			profile := getDamageProfile(tt.damageType)
			hasLinger := profile.LingeringDuration > 0

			if hasLinger != tt.shouldLinger {
				t.Errorf("Damage type %s lingering=%v, expected %v", tt.damageType, hasLinger, tt.shouldLinger)
			}
		})
	}
}

func TestUpdate_DecaysEffects(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem()

	entity := world.AddEntity()
	world.AddComponent(entity, &engine.Position{X: 10.0, Y: 20.0})

	dv := &DamageVisualComponent{
		ActiveEffects: []ActiveEffect{
			{
				DamageTypeName: "Fire",
				Intensity:      1.0,
				Duration:       1.0,
				MaxDuration:    2.0,
			},
		},
	}
	world.AddComponent(entity, dv)

	// Run update 60 times (1 second at 60 FPS)
	for i := 0; i < 60; i++ {
		sys.Update(world)
	}

	dvType := reflect.TypeOf((*DamageVisualComponent)(nil))
	comp, _ := world.GetComponent(entity, dvType)
	dv, _ = comp.(*DamageVisualComponent)

	if len(dv.ActiveEffects) != 0 {
		t.Errorf("Expected effect to expire, but %d effects remain", len(dv.ActiveEffects))
	}
}

func TestUpdate_SpawnsLingeringParticles(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem()

	mockParticles := &mockParticleSpawner{}
	sys.SetParticleSpawner(mockParticles)

	entity := world.AddEntity()
	world.AddComponent(entity, &engine.Position{X: 10.0, Y: 20.0})

	dv := &DamageVisualComponent{
		ActiveEffects: []ActiveEffect{
			{
				DamageTypeName: "Fire",
				Intensity:      1.0,
				Duration:       0.5,
				MaxDuration:    1.0,
			},
		},
	}
	world.AddComponent(entity, dv)

	initialBursts := mockParticles.burstCalls

	// Run several update frames
	for i := 0; i < 10; i++ {
		sys.Update(world)
	}

	if mockParticles.burstCalls <= initialBursts {
		t.Error("Expected lingering particles to be spawned during update")
	}
}

func TestApplyDamageVisual_DamageScaling(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem()

	mockParticles := &mockParticleSpawner{}
	mockFeedback := &mockFeedbackProvider{}
	sys.SetParticleSpawner(mockParticles)
	sys.SetFeedbackProvider(mockFeedback)

	entity1 := world.AddEntity()
	world.AddComponent(entity1, &engine.Position{X: 10.0, Y: 20.0})

	entity2 := world.AddEntity()
	world.AddComponent(entity2, &engine.Position{X: 30.0, Y: 40.0})

	// Low damage
	sys.ApplyDamageVisual(world, entity1, "Fire", 10.0, 10.0, 20.0)
	lowShake := mockFeedback.lastShake
	lowCount := mockParticles.lastCount

	// High damage
	sys.ApplyDamageVisual(world, entity2, "Fire", 100.0, 30.0, 40.0)
	highShake := mockFeedback.lastShake
	highCount := mockParticles.lastCount

	if highShake <= lowShake {
		t.Errorf("High damage shake %f should be greater than low damage shake %f", highShake, lowShake)
	}

	if highCount <= lowCount {
		t.Errorf("High damage particle count %d should be greater than low damage count %d", highCount, lowCount)
	}
}
