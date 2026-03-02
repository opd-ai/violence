package stats

import (
	"testing"
)

func TestNewAttributes(t *testing.T) {
	attrs := NewAttributes()

	tests := []struct {
		stat     Stat
		expected int
	}{
		{StatStrength, BaseStrength},
		{StatDexterity, BaseDexterity},
		{StatIntelligence, BaseIntelligence},
		{StatVitality, BaseVitality},
		{StatLuck, BaseLuck},
	}

	for _, tt := range tests {
		t.Run(string(tt.stat), func(t *testing.T) {
			if got := attrs.Get(tt.stat); got != tt.expected {
				t.Errorf("NewAttributes() %s = %d, want %d", tt.stat, got, tt.expected)
			}
		})
	}

	if attrs.GetUnallocatedPoints() != 0 {
		t.Errorf("NewAttributes() unallocated points = %d, want 0", attrs.GetUnallocatedPoints())
	}
}

func TestAllocate(t *testing.T) {
	tests := []struct {
		name          string
		initialPoints int
		stat          Stat
		wantValue     int
		wantPoints    int
		wantErr       bool
	}{
		{
			name:          "allocate strength with points",
			initialPoints: 1,
			stat:          StatStrength,
			wantValue:     BaseStrength + 1,
			wantPoints:    0,
			wantErr:       false,
		},
		{
			name:          "allocate without points",
			initialPoints: 0,
			stat:          StatDexterity,
			wantValue:     BaseDexterity,
			wantPoints:    0,
			wantErr:       true,
		},
		{
			name:          "allocate invalid stat",
			initialPoints: 1,
			stat:          Stat("invalid"),
			wantValue:     0,
			wantPoints:    1,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := NewAttributes()
			attrs.AddPoints(tt.initialPoints)

			err := attrs.Allocate(tt.stat)
			if (err != nil) != tt.wantErr {
				t.Errorf("Allocate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got := attrs.Get(tt.stat); got != tt.wantValue {
					t.Errorf("Allocate() stat value = %d, want %d", got, tt.wantValue)
				}
			}

			if got := attrs.GetUnallocatedPoints(); got != tt.wantPoints {
				t.Errorf("Allocate() unallocated points = %d, want %d", got, tt.wantPoints)
			}
		})
	}
}

func TestAddPoints(t *testing.T) {
	attrs := NewAttributes()

	attrs.AddPoints(5)
	if got := attrs.GetUnallocatedPoints(); got != 5 {
		t.Errorf("AddPoints(5) = %d, want 5", got)
	}

	attrs.AddPoints(3)
	if got := attrs.GetUnallocatedPoints(); got != 8 {
		t.Errorf("AddPoints(3) after 5 = %d, want 8", got)
	}
}

func TestBonusCalculations(t *testing.T) {
	tests := []struct {
		name          string
		stat          Stat
		allocate      int
		bonusFunc     func(*Attributes) float64
		expectedBonus float64
		bonusName     string
	}{
		{
			name:          "melee damage bonus",
			stat:          StatStrength,
			allocate:      10,
			bonusFunc:     func(a *Attributes) float64 { return a.GetMeleeDamageBonus() },
			expectedBonus: 1.20, // 1.0 + 10 * 0.02
			bonusName:     "melee damage",
		},
		{
			name:          "accuracy bonus",
			stat:          StatDexterity,
			allocate:      20,
			bonusFunc:     func(a *Attributes) float64 { return a.GetAccuracyBonus() },
			expectedBonus: 1.30, // 1.0 + 20 * 0.015
			bonusName:     "accuracy",
		},
		{
			name:          "skill power bonus",
			stat:          StatIntelligence,
			allocate:      8,
			bonusFunc:     func(a *Attributes) float64 { return a.GetSkillPowerBonus() },
			expectedBonus: 1.20, // 1.0 + 8 * 0.025
			bonusName:     "skill power",
		},
		{
			name:          "max health bonus",
			stat:          StatVitality,
			allocate:      10,
			bonusFunc:     func(a *Attributes) float64 { return a.GetMaxHealthBonus() },
			expectedBonus: 1.50, // 1.0 + 10 * 0.05
			bonusName:     "max health",
		},
		{
			name:          "critical chance",
			stat:          StatLuck,
			allocate:      20,
			bonusFunc:     func(a *Attributes) float64 { return a.GetCriticalChance() },
			expectedBonus: 0.15, // 0.05 + 20 * 0.005
			bonusName:     "critical chance",
		},
		{
			name:          "dodge chance",
			stat:          StatDexterity,
			allocate:      30,
			bonusFunc:     func(a *Attributes) float64 { return a.GetDodgeChance() },
			expectedBonus: 0.09, // 30 * 0.003
			bonusName:     "dodge chance",
		},
		{
			name:          "attack speed bonus",
			stat:          StatDexterity,
			allocate:      15,
			bonusFunc:     func(a *Attributes) float64 { return a.GetAttackSpeedBonus() },
			expectedBonus: 1.15, // 1.0 + 15 * 0.01
			bonusName:     "attack speed",
		},
		{
			name:          "loot quality bonus",
			stat:          StatLuck,
			allocate:      25,
			bonusFunc:     func(a *Attributes) float64 { return a.GetLootQualityBonus() },
			expectedBonus: 0.25, // 25 * 0.01
			bonusName:     "loot quality",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := NewAttributes()
			attrs.AddPoints(tt.allocate)

			for i := 0; i < tt.allocate; i++ {
				if err := attrs.Allocate(tt.stat); err != nil {
					t.Fatalf("Allocate() error = %v", err)
				}
			}

			got := tt.bonusFunc(attrs)
			if diff := got - tt.expectedBonus; diff < -0.001 || diff > 0.001 {
				t.Errorf("%s = %.4f, want %.4f", tt.bonusName, got, tt.expectedBonus)
			}
		})
	}
}

func TestReset(t *testing.T) {
	attrs := NewAttributes()
	attrs.AddPoints(15)

	// Allocate to various stats
	attrs.Allocate(StatStrength)
	attrs.Allocate(StatStrength)
	attrs.Allocate(StatStrength)
	attrs.Allocate(StatDexterity)
	attrs.Allocate(StatDexterity)
	attrs.Allocate(StatIntelligence)
	attrs.Allocate(StatVitality)
	attrs.Allocate(StatLuck)

	// Should have allocated 8 points, leaving 7
	if got := attrs.GetUnallocatedPoints(); got != 7 {
		t.Errorf("Before reset, unallocated = %d, want 7", got)
	}

	attrs.Reset()

	// All stats should be back to base
	tests := []struct {
		stat     Stat
		expected int
	}{
		{StatStrength, BaseStrength},
		{StatDexterity, BaseDexterity},
		{StatIntelligence, BaseIntelligence},
		{StatVitality, BaseVitality},
		{StatLuck, BaseLuck},
	}

	for _, tt := range tests {
		if got := attrs.Get(tt.stat); got != tt.expected {
			t.Errorf("After reset, %s = %d, want %d", tt.stat, got, tt.expected)
		}
	}

	// Should have all 15 points back
	if got := attrs.GetUnallocatedPoints(); got != 15 {
		t.Errorf("After reset, unallocated = %d, want 15", got)
	}
}

func TestGetAll(t *testing.T) {
	attrs := NewAttributes()
	attrs.AddPoints(5)
	attrs.Allocate(StatStrength)
	attrs.Allocate(StatLuck)

	all := attrs.GetAll()

	expected := map[string]int{
		"strength":     BaseStrength + 1,
		"dexterity":    BaseDexterity,
		"intelligence": BaseIntelligence,
		"vitality":     BaseVitality,
		"luck":         BaseLuck + 1,
		"unallocated":  3,
	}

	for key, want := range expected {
		if got, ok := all[key]; !ok {
			t.Errorf("GetAll() missing key %s", key)
		} else if got != want {
			t.Errorf("GetAll()[%s] = %d, want %d", key, got, want)
		}
	}
}

func TestStatAllocationComponent(t *testing.T) {
	comp := &StatAllocationComponent{
		Attributes: NewAttributes(),
	}

	if got := comp.Type(); got != "StatAllocation" {
		t.Errorf("StatAllocationComponent.Type() = %s, want StatAllocation", got)
	}
}

func TestConcurrentAccess(t *testing.T) {
	attrs := NewAttributes()
	attrs.AddPoints(100)

	done := make(chan bool)

	// Concurrent allocations
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				attrs.Allocate(StatStrength)
			}
			done <- true
		}()
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				attrs.Get(StatStrength)
				attrs.GetMeleeDamageBonus()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have allocated 50 points to strength
	if got := attrs.Get(StatStrength); got != BaseStrength+50 {
		t.Errorf("After concurrent allocations, strength = %d, want %d", got, BaseStrength+50)
	}

	if got := attrs.GetUnallocatedPoints(); got != 50 {
		t.Errorf("After concurrent allocations, unallocated = %d, want 50", got)
	}
}
