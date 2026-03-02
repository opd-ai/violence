package stats

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestSystemUpdate(t *testing.T) {
	sys := NewSystem()
	w := engine.NewWorld()

	entity := w.AddEntity()

	attrs := NewAttributes()
	attrs.AddPoints(20)

	// Allocate 20 points to vitality for 100% health increase
	for i := 0; i < 20; i++ {
		attrs.Allocate(StatVitality)
	}

	w.AddComponent(entity, &StatAllocationComponent{
		Attributes: attrs,
	})

	w.AddComponent(entity, &HealthComponent{
		Current:       100,
		MaxHealth:     100,
		BaseMaxHealth: 100,
	})

	sys.Update(w)

	healthType := reflect.TypeOf((*HealthComponent)(nil))
	healthComp, ok := w.GetComponent(entity, healthType)
	if !ok {
		t.Fatal("Health component not found")
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		t.Fatal("Health component wrong type")
	}

	// With 20 points in vitality: 1.0 + 20 * 0.05 = 2.0x multiplier
	expectedMax := 200
	if health.MaxHealth != expectedMax {
		t.Errorf("MaxHealth = %d, want %d", health.MaxHealth, expectedMax)
	}
}

func TestSystemUpdateWithoutStatsComponent(t *testing.T) {
	sys := NewSystem()
	w := engine.NewWorld()

	entity := w.AddEntity()

	w.AddComponent(entity, &HealthComponent{
		Current:       100,
		MaxHealth:     100,
		BaseMaxHealth: 100,
	})

	sys.Update(w)

	healthType := reflect.TypeOf((*HealthComponent)(nil))
	healthComp, ok := w.GetComponent(entity, healthType)
	if !ok {
		t.Fatal("Health component not found")
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		t.Fatal("Health component wrong type")
	}

	// Without stats, health should remain unchanged
	if health.MaxHealth != 100 {
		t.Errorf("MaxHealth = %d, want 100", health.MaxHealth)
	}
}

func TestApplyDamageBonus(t *testing.T) {
	tests := []struct {
		name        string
		stat        Stat
		allocate    int
		damageType  string
		baseDamage  float64
		expectedDmg float64
	}{
		{
			name:        "strength melee bonus",
			stat:        StatStrength,
			allocate:    10,
			damageType:  "melee",
			baseDamage:  100.0,
			expectedDmg: 120.0, // 100 * 1.20
		},
		{
			name:        "intelligence magic bonus",
			stat:        StatIntelligence,
			allocate:    8,
			damageType:  "magic",
			baseDamage:  100.0,
			expectedDmg: 120.0, // 100 * 1.20
		},
		{
			name:        "no bonus for unknown type",
			stat:        StatStrength,
			allocate:    10,
			damageType:  "unknown",
			baseDamage:  100.0,
			expectedDmg: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem()
			w := engine.NewWorld()
			entity := w.AddEntity()

			attrs := NewAttributes()
			attrs.AddPoints(tt.allocate)
			for i := 0; i < tt.allocate; i++ {
				attrs.Allocate(tt.stat)
			}

			w.AddComponent(entity, &StatAllocationComponent{
				Attributes: attrs,
			})

			got := sys.ApplyDamageBonus(w, entity, tt.baseDamage, tt.damageType)
			if diff := got - tt.expectedDmg; diff < -0.01 || diff > 0.01 {
				t.Errorf("ApplyDamageBonus() = %.2f, want %.2f", got, tt.expectedDmg)
			}
		})
	}
}

func TestApplyAccuracyBonus(t *testing.T) {
	sys := NewSystem()
	w := engine.NewWorld()
	entity := w.AddEntity()

	attrs := NewAttributes()
	attrs.AddPoints(20)
	for i := 0; i < 20; i++ {
		attrs.Allocate(StatDexterity)
	}

	w.AddComponent(entity, &StatAllocationComponent{
		Attributes: attrs,
	})

	baseAccuracy := 0.75
	got := sys.ApplyAccuracyBonus(w, entity, baseAccuracy)

	// 0.75 * (1.0 + 20 * 0.015) = 0.75 * 1.30 = 0.975
	expected := 0.975
	if diff := got - expected; diff < -0.001 || diff > 0.001 {
		t.Errorf("ApplyAccuracyBonus() = %.4f, want %.4f", got, expected)
	}
}

func TestRollCritical(t *testing.T) {
	sys := NewSystem()
	w := engine.NewWorld()
	entity := w.AddEntity()

	attrs := NewAttributes()
	attrs.AddPoints(20)
	for i := 0; i < 20; i++ {
		attrs.Allocate(StatLuck)
	}

	w.AddComponent(entity, &StatAllocationComponent{
		Attributes: attrs,
	})

	// Critical chance is 5% + 20 * 0.5% = 15%
	// Values below 0.15 should crit
	tests := []struct {
		randomValue float64
		wantCrit    bool
	}{
		{0.10, true},
		{0.14, true},
		{0.1500001, false},
		{0.20, false},
	}

	for _, tt := range tests {
		got := sys.RollCritical(w, entity, tt.randomValue)
		if got != tt.wantCrit {
			t.Errorf("RollCritical(%f) = %v, want %v", tt.randomValue, got, tt.wantCrit)
		}
	}
}

func TestRollDodge(t *testing.T) {
	sys := NewSystem()
	w := engine.NewWorld()
	entity := w.AddEntity()

	attrs := NewAttributes()
	attrs.AddPoints(30)
	for i := 0; i < 30; i++ {
		attrs.Allocate(StatDexterity)
	}

	w.AddComponent(entity, &StatAllocationComponent{
		Attributes: attrs,
	})

	// Dodge chance is 30 * 0.3% = 9%
	tests := []struct {
		randomValue float64
		wantDodge   bool
	}{
		{0.05, true},
		{0.08, true},
		{0.09, false},
		{0.15, false},
	}

	for _, tt := range tests {
		got := sys.RollDodge(w, entity, tt.randomValue)
		if got != tt.wantDodge {
			t.Errorf("RollDodge(%f) = %v, want %v", tt.randomValue, got, tt.wantDodge)
		}
	}
}

func TestGetAttackSpeed(t *testing.T) {
	sys := NewSystem()
	w := engine.NewWorld()
	entity := w.AddEntity()

	attrs := NewAttributes()
	attrs.AddPoints(15)
	for i := 0; i < 15; i++ {
		attrs.Allocate(StatDexterity)
	}

	w.AddComponent(entity, &StatAllocationComponent{
		Attributes: attrs,
	})

	got := sys.GetAttackSpeed(w, entity)
	expected := 1.15 // 1.0 + 15 * 0.01

	if diff := got - expected; diff < -0.001 || diff > 0.001 {
		t.Errorf("GetAttackSpeed() = %.4f, want %.4f", got, expected)
	}
}

func TestGetLootQualityModifier(t *testing.T) {
	sys := NewSystem()
	w := engine.NewWorld()
	entity := w.AddEntity()

	attrs := NewAttributes()
	attrs.AddPoints(25)
	for i := 0; i < 25; i++ {
		attrs.Allocate(StatLuck)
	}

	w.AddComponent(entity, &StatAllocationComponent{
		Attributes: attrs,
	})

	got := sys.GetLootQualityModifier(w, entity)
	expected := 0.25 // 25 * 0.01

	if diff := got - expected; diff < -0.001 || diff > 0.001 {
		t.Errorf("GetLootQualityModifier() = %.4f, want %.4f", got, expected)
	}
}

func TestSystemWithoutEntity(t *testing.T) {
	sys := NewSystem()
	w := engine.NewWorld()

	// Test all methods with entity that doesn't have stat component
	entity := w.AddEntity()

	if dmg := sys.ApplyDamageBonus(w, entity, 100.0, "melee"); dmg != 100.0 {
		t.Errorf("ApplyDamageBonus without stats = %.2f, want 100.0", dmg)
	}

	if acc := sys.ApplyAccuracyBonus(w, entity, 0.75); acc != 0.75 {
		t.Errorf("ApplyAccuracyBonus without stats = %.4f, want 0.75", acc)
	}

	if crit := sys.RollCritical(w, entity, 0.5); crit {
		t.Error("RollCritical without stats = true, want false")
	}

	if dodge := sys.RollDodge(w, entity, 0.5); dodge {
		t.Error("RollDodge without stats = true, want false")
	}

	if speed := sys.GetAttackSpeed(w, entity); speed != 1.0 {
		t.Errorf("GetAttackSpeed without stats = %.2f, want 1.0", speed)
	}

	if loot := sys.GetLootQualityModifier(w, entity); loot != 0.0 {
		t.Errorf("GetLootQualityModifier without stats = %.2f, want 0.0", loot)
	}
}
