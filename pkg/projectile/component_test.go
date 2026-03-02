package projectile

import (
	"image/color"
	"testing"
)

func TestNewProjectileComponent(t *testing.T) {
	proj := NewProjectileComponent(5.0, 3.0, 50.0, DamageFire, 123)
	
	if proj == nil {
		t.Fatal("NewProjectileComponent() returned nil")
	}
	
	if proj.VelX != 5.0 || proj.VelY != 3.0 {
		t.Errorf("Velocity = (%v, %v), want (5.0, 3.0)", proj.VelX, proj.VelY)
	}
	
	if proj.Damage != 50.0 {
		t.Errorf("Damage = %v, want 50.0", proj.Damage)
	}
	
	if proj.DamageType != DamageFire {
		t.Errorf("DamageType = %v, want DamageFire", proj.DamageType)
	}
	
	if proj.OwnerID != 123 {
		t.Errorf("OwnerID = %v, want 123", proj.OwnerID)
	}
	
	if proj.Shape != ShapeCircle {
		t.Errorf("Shape = %v, want ShapeCircle", proj.Shape)
	}
	
	if proj.Lifetime != 5.0 || proj.MaxLifetime != 5.0 {
		t.Errorf("Lifetime/MaxLifetime = %v/%v, want 5.0/5.0", proj.Lifetime, proj.MaxLifetime)
	}
	
	if !proj.TrailParticles {
		t.Error("TrailParticles should be true by default")
	}
	
	if proj.HitEntities == nil {
		t.Error("HitEntities map should be initialized")
	}
	
	if proj.Type() != "ProjectileComponent" {
		t.Errorf("Type() = %v, want ProjectileComponent", proj.Type())
	}
}

func TestGetDamageTypeColor(t *testing.T) {
	tests := []struct {
		damageType    DamageType
		expectedColor color.RGBA
	}{
		{DamagePhysical, color.RGBA{R: 192, G: 192, B: 192, A: 255}},
		{DamageFire, color.RGBA{R: 255, G: 80, B: 20, A: 255}},
		{DamageIce, color.RGBA{R: 100, G: 200, B: 255, A: 255}},
		{DamageLightning, color.RGBA{R: 255, G: 255, B: 100, A: 255}},
		{DamagePoison, color.RGBA{R: 100, G: 255, B: 100, A: 255}},
		{DamageHoly, color.RGBA{R: 255, G: 255, B: 200, A: 255}},
		{DamageShadow, color.RGBA{R: 80, G: 60, B: 120, A: 255}},
		{DamageArcane, color.RGBA{R: 200, G: 100, B: 255, A: 255}},
	}
	
	for _, tt := range tests {
		result := getDamageTypeColor(tt.damageType)
		if result != tt.expectedColor {
			t.Errorf("getDamageTypeColor(%v) = %v, want %v", tt.damageType, result, tt.expectedColor)
		}
	}
}

func TestProjectileShapes(t *testing.T) {
	// Verify shape constants exist and are distinct
	shapes := []ProjectileShape{ShapeCircle, ShapeBeam, ShapeAOE}
	seen := make(map[ProjectileShape]bool)
	
	for _, shape := range shapes {
		if seen[shape] {
			t.Errorf("Duplicate shape value: %v", shape)
		}
		seen[shape] = true
	}
}

func TestProjectileComponent_Pierce(t *testing.T) {
	proj := NewProjectileComponent(5.0, 0.0, 25.0, DamageLightning, 1)
	proj.PierceCount = 3
	
	// Track some hits
	proj.HitEntities[10] = true
	proj.HitEntities[11] = true
	
	if len(proj.HitEntities) != 2 {
		t.Errorf("HitEntities count = %v, want 2", len(proj.HitEntities))
	}
	
	if !proj.HitEntities[10] {
		t.Error("Entity 10 should be marked as hit")
	}
}

func TestProjectileComponent_Explosion(t *testing.T) {
	proj := NewProjectileComponent(3.0, 4.0, 100.0, DamageFire, 5)
	proj.ExplodeOnDeath = true
	proj.ExplosionRadius = 2.5
	
	if !proj.ExplodeOnDeath {
		t.Error("ExplodeOnDeath should be true")
	}
	
	if proj.ExplosionRadius != 2.5 {
		t.Errorf("ExplosionRadius = %v, want 2.5", proj.ExplosionRadius)
	}
}

func BenchmarkNewProjectileComponent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewProjectileComponent(5.0, 5.0, 50.0, DamageFire, 1)
	}
}
