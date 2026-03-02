package projectile

import (
	"image/color"
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

// Mock spatial grid
type mockSpatialGrid struct {
	entities []engine.Entity
}

func (m *mockSpatialGrid) QueryRadius(x, y, radius float64) []engine.Entity {
	return m.entities
}

// Mock particle spawner
type mockParticleSpawner struct {
	spawnCount int
}

func (m *mockParticleSpawner) SpawnBurst(x, y, z float64, count int, speed, lifetime, fadeTime, gravity float64, col color.RGBA) {
	m.spawnCount += count
}

// Mock feedback provider
type mockFeedbackProvider struct {
	shakeCount int
	flashCount int
}

func (m *mockFeedbackProvider) AddScreenShake(intensity float64) {
	m.shakeCount++
}

func (m *mockFeedbackProvider) AddHitFlash(intensity float64) {
	m.flashCount++
}

func TestSystem_Update_Movement(t *testing.T) {
	sys := NewSystem()
	w := engine.NewWorld()
	
	entity := w.AddEntity()
	proj := NewProjectileComponent(5.0, 3.0, 25.0, DamageFire, 99)
	pos := &engine.Position{X: 0.0, Y: 0.0}
	
	w.AddComponent(entity, proj)
	w.AddComponent(entity, pos)
	
	sys.Update(w)
	
	// Check that position was updated
	posComp, ok := w.GetComponent(entity, reflect.TypeOf((*engine.Position)(nil)))
	if !ok {
		t.Fatal("Position component missing after update")
	}
	
	newPos := posComp.(*engine.Position)
	deltaTime := 1.0 / 60.0
	expectedX := 5.0 * deltaTime
	expectedY := 3.0 * deltaTime
	
	if newPos.X != expectedX || newPos.Y != expectedY {
		t.Errorf("Position after update = (%v, %v), want (%v, %v)", newPos.X, newPos.Y, expectedX, expectedY)
	}
	
	if proj.Lifetime >= 5.0 {
		t.Errorf("Lifetime should decrease, got %v", proj.Lifetime)
	}
}

func TestSystem_Update_Lifetime(t *testing.T) {
	sys := NewSystem()
	w := engine.NewWorld()
	
	entity := w.AddEntity()
	proj := NewProjectileComponent(1.0, 1.0, 10.0, DamageIce, 1)
	proj.Lifetime = 0.01 // Very short lifetime
	pos := &engine.Position{X: 0.0, Y: 0.0}
	
	w.AddComponent(entity, proj)
	w.AddComponent(entity, pos)
	
	sys.Update(w)
	
	// Entity should be removed
	_, ok := w.GetComponent(entity, reflect.TypeOf((*ProjectileComponent)(nil)))
	if ok {
		t.Error("Entity should be removed after lifetime expires")
	}
}

func TestSystem_Update_Collision(t *testing.T) {
	sys := NewSystem()
	w := engine.NewWorld()
	
	// Create target
	target := w.AddEntity()
	targetPos := &engine.Position{X: 0.5, Y: 0.0}
	w.AddComponent(target, targetPos)
	
	// Setup spatial grid with target
	grid := &mockSpatialGrid{
		entities: []engine.Entity{target},
	}
	sys.SetSpatialGrid(grid)
	
	particles := &mockParticleSpawner{}
	sys.SetParticleSpawner(particles)
	
	// Create projectile
	projectileEntity := w.AddEntity()
	proj := NewProjectileComponent(10.0, 0.0, 50.0, DamageFire, 99)
	proj.Radius = 0.5
	projPos := &engine.Position{X: 0.4, Y: 0.0}
	
	w.AddComponent(projectileEntity, proj)
	w.AddComponent(projectileEntity, projPos)
	
	sys.Update(w)
	
	// Projectile should hit target
	if len(proj.HitEntities) == 0 {
		t.Error("Projectile should have hit the target")
	}
	
	targetID := int(target)
	if !proj.HitEntities[targetID] {
		t.Error("Target should be marked as hit")
	}
	
	// Particles should be spawned for impact
	if particles.spawnCount == 0 {
		t.Error("Impact particles should be spawned")
	}
}

func TestSystem_Update_OwnerIgnore(t *testing.T) {
	sys := NewSystem()
	w := engine.NewWorld()
	
	owner := w.AddEntity()
	ownerID := int(owner) // Use actual entity ID
	ownerPos := &engine.Position{X: 0.5, Y: 0.0}
	w.AddComponent(owner, ownerPos)
	
	grid := &mockSpatialGrid{
		entities: []engine.Entity{owner},
	}
	sys.SetSpatialGrid(grid)
	
	projectileEntity := w.AddEntity()
	proj := NewProjectileComponent(1.0, 0.0, 25.0, DamageFire, ownerID)
	projPos := &engine.Position{X: 0.4, Y: 0.0}
	
	w.AddComponent(projectileEntity, proj)
	w.AddComponent(projectileEntity, projPos)
	
	sys.Update(w)
	
	// Should not hit owner
	if len(proj.HitEntities) > 0 {
		t.Error("Projectile should not hit its owner")
	}
}

func TestSystem_Update_Explosion(t *testing.T) {
	sys := NewSystem()
	w := engine.NewWorld()
	
	target := w.AddEntity()
	targetPos := &engine.Position{X: 0.5, Y: 0.0}
	w.AddComponent(target, targetPos)
	
	grid := &mockSpatialGrid{
		entities: []engine.Entity{target},
	}
	sys.SetSpatialGrid(grid)
	
	particles := &mockParticleSpawner{}
	sys.SetParticleSpawner(particles)
	
	feedback := &mockFeedbackProvider{}
	sys.SetFeedbackProvider(feedback)
	
	projectileEntity := w.AddEntity()
	proj := NewProjectileComponent(5.0, 0.0, 50.0, DamageFire, 99)
	proj.Lifetime = 0.01 // Short lifetime to trigger death
	proj.ExplodeOnDeath = true
	proj.ExplosionRadius = 1.0
	projPos := &engine.Position{X: 0.5, Y: 0.0}
	
	w.AddComponent(projectileEntity, proj)
	w.AddComponent(projectileEntity, projPos)
	
	sys.Update(w)
	
	// Should be removed
	_, ok := w.GetComponent(projectileEntity, reflect.TypeOf((*ProjectileComponent)(nil)))
	if ok {
		t.Error("Projectile should be removed after lifetime expires")
	}
	
	// Explosion particles should spawn
	if particles.spawnCount == 0 {
		t.Error("Explosion should spawn particles")
	}
	
	// Screen shake should occur
	if feedback.shakeCount == 0 {
		t.Error("Explosion should trigger screen shake")
	}
}

func TestSystem_Update_Resistance(t *testing.T) {
	sys := NewSystem()
	w := engine.NewWorld()
	
	target := w.AddEntity()
	targetPos := &engine.Position{X: 0.5, Y: 0.0}
	rc := NewResistanceComponent()
	rc.Resistances[DamageFire] = 0.5 // 50% fire resistance
	w.AddComponent(target, targetPos)
	w.AddComponent(target, rc)
	
	grid := &mockSpatialGrid{
		entities: []engine.Entity{target},
	}
	sys.SetSpatialGrid(grid)
	
	projectileEntity := w.AddEntity()
	proj := NewProjectileComponent(5.0, 0.0, 100.0, DamageFire, 99)
	proj.Radius = 0.5
	projPos := &engine.Position{X: 0.4, Y: 0.0}
	
	w.AddComponent(projectileEntity, proj)
	w.AddComponent(projectileEntity, projPos)
	
	sys.Update(w)
	
	// Target should be hit (resistance calculation verified in other tests)
	if len(proj.HitEntities) == 0 {
		t.Error("Target with resistance should still be hit")
	}
}

func TestSystem_Type(t *testing.T) {
	sys := NewSystem()
	if sys.Type() != "ProjectileSystem" {
		t.Errorf("Type() = %v, want ProjectileSystem", sys.Type())
	}
}

func BenchmarkSystem_Update(b *testing.B) {
	sys := NewSystem()
	w := engine.NewWorld()
	
	grid := &mockSpatialGrid{entities: []engine.Entity{}}
	sys.SetSpatialGrid(grid)
	
	particles := &mockParticleSpawner{}
	sys.SetParticleSpawner(particles)
	
	for i := 0; i < 10; i++ {
		entity := w.AddEntity()
		proj := NewProjectileComponent(5.0, 0.0, 25.0, DamageFire, 999)
		pos := &engine.Position{X: float64(i), Y: 0.0}
		w.AddComponent(entity, proj)
		w.AddComponent(entity, pos)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(w)
	}
}
