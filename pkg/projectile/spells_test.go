package projectile

import (
	"math/rand"
	"testing"
)

func TestGenreSpellTemplates(t *testing.T) {
	genres := []string{"fantasy", "scifi", "cyberpunk", "horror"}
	
	for _, genre := range genres {
		templates, exists := GenreSpellTemplates[genre]
		if !exists {
			t.Errorf("Genre %s missing from GenreSpellTemplates", genre)
			continue
		}
		
		if len(templates) == 0 {
			t.Errorf("Genre %s has no spell templates", genre)
			continue
		}
		
		for _, template := range templates {
			if template.Name == "" {
				t.Errorf("Genre %s has template with empty name", genre)
			}
			if template.BaseDamage <= 0 {
				t.Errorf("Template %s has invalid damage: %v", template.Name, template.BaseDamage)
			}
			if template.Speed <= 0 {
				t.Errorf("Template %s has invalid speed: %v", template.Name, template.Speed)
			}
			if template.Lifetime <= 0 {
				t.Errorf("Template %s has invalid lifetime: %v", template.Name, template.Lifetime)
			}
		}
	}
}

func TestCreateSpellProjectile(t *testing.T) {
	template := SpellTemplate{
		Name:            "TestSpell",
		DamageType:      DamageFire,
		BaseDamage:      50.0,
		Speed:           10.0,
		Shape:           ShapeCircle,
		Radius:          0.3,
		PierceCount:     2,
		ExplodeOnDeath:  true,
		ExplosionRadius: 1.5,
		TrailParticles:  true,
		Lifetime:        3.0,
	}
	
	rng := rand.New(rand.NewSource(12345))
	proj := CreateSpellProjectile(template, 1.0, 0.0, 99, rng)
	
	if proj == nil {
		t.Fatal("CreateSpellProjectile() returned nil")
	}
	
	if proj.VelX != 10.0 || proj.VelY != 0.0 {
		t.Errorf("Velocity = (%v, %v), want (10.0, 0.0)", proj.VelX, proj.VelY)
	}
	
	// Damage should have variance (±10%)
	if proj.Damage < 45.0 || proj.Damage > 55.0 {
		t.Errorf("Damage = %v, want range [45.0, 55.0]", proj.Damage)
	}
	
	if proj.DamageType != DamageFire {
		t.Errorf("DamageType = %v, want DamageFire", proj.DamageType)
	}
	
	if proj.Shape != ShapeCircle {
		t.Errorf("Shape = %v, want ShapeCircle", proj.Shape)
	}
	
	if proj.PierceCount != 2 {
		t.Errorf("PierceCount = %v, want 2", proj.PierceCount)
	}
	
	if !proj.ExplodeOnDeath {
		t.Error("ExplodeOnDeath should be true")
	}
	
	if proj.ExplosionRadius != 1.5 {
		t.Errorf("ExplosionRadius = %v, want 1.5", proj.ExplosionRadius)
	}
	
	if proj.OwnerID != 99 {
		t.Errorf("OwnerID = %v, want 99", proj.OwnerID)
	}
}

func TestGetRandomSpellForGenre(t *testing.T) {
	rng := rand.New(rand.NewSource(67890))
	
	// Test valid genre
	spell := GetRandomSpellForGenre("fantasy", rng)
	if spell.Name == "" {
		t.Error("GetRandomSpellForGenre() returned empty spell name")
	}
	
	// Test invalid genre (should fallback to fantasy)
	spell = GetRandomSpellForGenre("invalid_genre", rng)
	if spell.Name == "" {
		t.Error("GetRandomSpellForGenre() with invalid genre returned empty spell name")
	}
	
	// Test nil rng (should return first spell)
	spell = GetRandomSpellForGenre("scifi", nil)
	if spell.Name == "" {
		t.Error("GetRandomSpellForGenre() with nil rng returned empty spell name")
	}
}

func TestCreateResistanceProfile(t *testing.T) {
	rng := rand.New(rand.NewSource(11111))
	
	tests := []struct {
		genre      string
		entityType string
		checkType  DamageType
		minValue   float64
		maxValue   float64
	}{
		{"fantasy", "fire_elemental", DamageFire, 0.7, 0.8},      // Should have fire resistance
		{"fantasy", "fire_elemental", DamageIce, -0.6, -0.4},     // Should have ice weakness
		{"fantasy", "undead", DamagePoison, 1.0, 1.0},            // Should be immune to poison
		{"fantasy", "undead", DamageHoly, -1.0, -0.5},            // Should be weak to holy
		{"scifi", "robot", DamageLightning, -0.6, -0.4},          // Should be weak to lightning
		{"scifi", "robot", DamagePoison, 1.0, 1.0},               // Should be immune to poison
		{"horror", "ghost", DamagePhysical, 0.7, 0.9},            // Should resist physical
	}
	
	for _, tt := range tests {
		rc := CreateResistanceProfile(tt.genre, tt.entityType, rng)
		if rc == nil {
			t.Errorf("CreateResistanceProfile(%s, %s) returned nil", tt.genre, tt.entityType)
			continue
		}
		
		resistance, exists := rc.Resistances[tt.checkType]
		if !exists {
			t.Errorf("Resistance profile for %s/%s missing %v", tt.genre, tt.entityType, DamageTypeNames[tt.checkType])
			continue
		}
		
		if resistance < tt.minValue || resistance > tt.maxValue {
			t.Errorf("Resistance for %s/%s %v = %v, want range [%v, %v]",
				tt.genre, tt.entityType, DamageTypeNames[tt.checkType],
				resistance, tt.minValue, tt.maxValue)
		}
	}
}

func TestCreateResistanceProfile_UnknownType(t *testing.T) {
	rng := rand.New(rand.NewSource(22222))
	rc := CreateResistanceProfile("fantasy", "unknown_type", rng)
	
	if rc == nil {
		t.Fatal("CreateResistanceProfile() returned nil")
	}
	
	// Unknown type should have empty resistances
	if len(rc.Resistances) != 0 {
		t.Errorf("Unknown entity type should have 0 resistances, got %v", len(rc.Resistances))
	}
}

func TestGetDamageTypeColor_Export(t *testing.T) {
	// Test the exported function
	col := GetDamageTypeColor(DamageFire)
	if col.R != 255 || col.G != 80 || col.B != 20 {
		t.Errorf("GetDamageTypeColor(DamageFire) = %v, want (255, 80, 20)", col)
	}
}

func BenchmarkCreateSpellProjectile(b *testing.B) {
	template := GenreSpellTemplates["fantasy"][0]
	rng := rand.New(rand.NewSource(99999))
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CreateSpellProjectile(template, 1.0, 0.0, 1, rng)
	}
}

func BenchmarkGetRandomSpellForGenre(b *testing.B) {
	rng := rand.New(rand.NewSource(88888))
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetRandomSpellForGenre("fantasy", rng)
	}
}

func BenchmarkCreateResistanceProfile(b *testing.B) {
	rng := rand.New(rand.NewSource(77777))
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CreateResistanceProfile("fantasy", "fire_elemental", rng)
	}
}
