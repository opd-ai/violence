package projectile

import (
	"testing"
)

func TestCalculateDamage(t *testing.T) {
	tests := []struct {
		name        string
		baseDamage  float64
		damageType  DamageType
		resistances map[DamageType]float64
		expected    float64
	}{
		{
			name:        "no resistance",
			baseDamage:  100.0,
			damageType:  DamageFire,
			resistances: map[DamageType]float64{},
			expected:    100.0,
		},
		{
			name:        "50% resistance",
			baseDamage:  100.0,
			damageType:  DamageFire,
			resistances: map[DamageType]float64{DamageFire: 0.5},
			expected:    50.0,
		},
		{
			name:        "immune (100% resistance)",
			baseDamage:  100.0,
			damageType:  DamagePoison,
			resistances: map[DamageType]float64{DamagePoison: 1.0},
			expected:    0.0,
		},
		{
			name:        "weakness (negative resistance)",
			baseDamage:  100.0,
			damageType:  DamageHoly,
			resistances: map[DamageType]float64{DamageHoly: -0.5},
			expected:    150.0,
		},
		{
			name:        "double damage weakness",
			baseDamage:  100.0,
			damageType:  DamageHoly,
			resistances: map[DamageType]float64{DamageHoly: -1.0},
			expected:    200.0,
		},
		{
			name:        "nil resistances",
			baseDamage:  100.0,
			damageType:  DamageFire,
			resistances: nil,
			expected:    100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateDamage(tt.baseDamage, tt.damageType, tt.resistances)
			if result != tt.expected {
				t.Errorf("CalculateDamage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNewResistanceComponent(t *testing.T) {
	rc := NewResistanceComponent()

	if rc == nil {
		t.Fatal("NewResistanceComponent() returned nil")
	}

	if rc.Resistances == nil {
		t.Fatal("Resistances map is nil")
	}

	if rc.Type() != "ResistanceComponent" {
		t.Errorf("Type() = %v, want ResistanceComponent", rc.Type())
	}
}

func TestDamageTypeNames(t *testing.T) {
	expectedNames := map[DamageType]string{
		DamagePhysical:  "Physical",
		DamageFire:      "Fire",
		DamageIce:       "Ice",
		DamageLightning: "Lightning",
		DamagePoison:    "Poison",
		DamageHoly:      "Holy",
		DamageShadow:    "Shadow",
		DamageArcane:    "Arcane",
	}

	for dt, expectedName := range expectedNames {
		if name, exists := DamageTypeNames[dt]; !exists {
			t.Errorf("DamageType %v missing from DamageTypeNames", dt)
		} else if name != expectedName {
			t.Errorf("DamageTypeNames[%v] = %v, want %v", dt, name, expectedName)
		}
	}
}

func TestResistanceComponent_Coverage(t *testing.T) {
	rc := NewResistanceComponent()

	// Set some resistances
	rc.Resistances[DamageFire] = 0.5
	rc.Resistances[DamageIce] = -0.3

	// Verify retrieval
	if rc.Resistances[DamageFire] != 0.5 {
		t.Errorf("Fire resistance = %v, want 0.5", rc.Resistances[DamageFire])
	}

	if rc.Resistances[DamageIce] != -0.3 {
		t.Errorf("Ice resistance = %v, want -0.3", rc.Resistances[DamageIce])
	}

	// Unset resistance should return zero value
	if rc.Resistances[DamagePoison] != 0.0 {
		t.Errorf("Poison resistance = %v, want 0.0", rc.Resistances[DamagePoison])
	}
}

func BenchmarkCalculateDamage(b *testing.B) {
	resistances := map[DamageType]float64{
		DamageFire:      0.5,
		DamageIce:       0.3,
		DamageLightning: -0.2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateDamage(100.0, DamageFire, resistances)
	}
}
