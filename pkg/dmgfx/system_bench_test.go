package dmgfx

import (
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func BenchmarkApplyDamageVisual(b *testing.B) {
	world := engine.NewWorld()
	sys := NewSystem()

	mockParticles := &mockParticleSpawner{}
	mockFeedback := &mockFeedbackProvider{}
	sys.SetParticleSpawner(mockParticles)
	sys.SetFeedbackProvider(mockFeedback)

	// Create 100 entities
	entities := make([]engine.Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.AddEntity()
		world.AddComponent(entity, &engine.Position{X: float64(i * 10), Y: float64(i * 10)})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entity := entities[i%100]
		damageType := []string{"Fire", "Ice", "Lightning", "Poison"}[i%4]
		sys.ApplyDamageVisual(world, entity, damageType, 50.0, 10.0, 20.0)
	}
}

func BenchmarkUpdate(b *testing.B) {
	world := engine.NewWorld()
	sys := NewSystem()

	mockParticles := &mockParticleSpawner{}
	sys.SetParticleSpawner(mockParticles)

	// Create 100 entities with active effects
	for i := 0; i < 100; i++ {
		entity := world.AddEntity()
		world.AddComponent(entity, &engine.Position{X: float64(i * 10), Y: float64(i * 10)})

		dv := &DamageVisualComponent{
			ActiveEffects: []ActiveEffect{
				{
					DamageTypeName: "Fire",
					Intensity:      1.0,
					Duration:       2.0,
					MaxDuration:    2.0,
				},
			},
		}
		world.AddComponent(entity, dv)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(world)
	}
}

func BenchmarkGetDamageProfile(b *testing.B) {
	damageTypes := []string{"Fire", "Ice", "Lightning", "Poison", "Holy", "Shadow", "Arcane", "Physical"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getDamageProfile(damageTypes[i%8])
	}
}
