package damagenumber

import (
	"image/color"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestComponent_Type(t *testing.T) {
	comp := &Component{}
	if got := comp.Type(); got != "DamageNumber" {
		t.Errorf("Component.Type() = %q, want %q", got, "DamageNumber")
	}
}

func TestSpawn(t *testing.T) {
	tests := []struct {
		name       string
		value      int
		x          float64
		y          float64
		damageType string
		isCritical bool
		isHeal     bool
	}{
		{"normal damage", 42, 100.0, 200.0, "physical", false, false},
		{"critical damage", 99, 150.0, 250.0, "fire", true, false},
		{"heal", 25, 120.0, 180.0, "", false, true},
		{"poison damage", 15, 110.0, 220.0, "poison", false, false},
		{"lightning crit", 150, 200.0, 300.0, "lightning", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := engine.NewWorld()
			ent := Spawn(w, tt.value, tt.x, tt.y, tt.damageType, tt.isCritical, tt.isHeal)

			compType := reflect.TypeOf((*Component)(nil))
			comp, found := w.GetComponent(ent, compType)
			if !found {
				t.Fatal("Spawn() did not add component to entity")
			}

			dmg, ok := comp.(*Component)
			if !ok {
				t.Fatalf("Component is not *Component, got %T", comp)
			}

			if dmg.Value != tt.value {
				t.Errorf("Value = %d, want %d", dmg.Value, tt.value)
			}
			if dmg.X != tt.x {
				t.Errorf("X = %f, want %f", dmg.X, tt.x)
			}
			if dmg.Y != tt.y {
				t.Errorf("Y = %f, want %f", dmg.Y, tt.y)
			}
			if dmg.DamageType != tt.damageType {
				t.Errorf("DamageType = %q, want %q", dmg.DamageType, tt.damageType)
			}
			if dmg.IsCritical != tt.isCritical {
				t.Errorf("IsCritical = %v, want %v", dmg.IsCritical, tt.isCritical)
			}
			if dmg.IsHeal != tt.isHeal {
				t.Errorf("IsHeal = %v, want %v", dmg.IsHeal, tt.isHeal)
			}

			if dmg.Age != 0 {
				t.Errorf("Age = %f, want 0", dmg.Age)
			}
			if dmg.Lifetime <= 0 {
				t.Errorf("Lifetime = %f, want > 0", dmg.Lifetime)
			}
			if dmg.VelocityY <= 0 {
				t.Errorf("VelocityY = %f, want > 0", dmg.VelocityY)
			}

			if tt.isCritical {
				if dmg.Lifetime <= 1.5 {
					t.Errorf("Critical should have longer lifetime, got %f", dmg.Lifetime)
				}
				if dmg.VelocityY <= 40.0 {
					t.Errorf("Critical should have higher velocity, got %f", dmg.VelocityY)
				}
			}
		})
	}
}

func TestGetDamageColor(t *testing.T) {
	tests := []struct {
		damageType string
		isHeal     bool
		wantGreen  bool
		wantRed    bool
		wantBlue   bool
	}{
		{"physical", false, false, true, false},
		{"fire", false, false, true, false},
		{"ice", false, false, false, true},
		{"lightning", false, false, false, false},
		{"poison", false, true, false, false},
		{"", true, true, false, false},
		{"anything", true, true, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.damageType, func(t *testing.T) {
			got := getDamageColor(tt.damageType, tt.isHeal)

			if tt.wantGreen && got.G < 200 {
				t.Errorf("Expected green-ish color for %q (heal=%v), got %+v", tt.damageType, tt.isHeal, got)
			}
			if tt.wantRed && got.R < 200 {
				t.Errorf("Expected red-ish color for %q, got %+v", tt.damageType, got)
			}
			if tt.wantBlue && got.B < 200 {
				t.Errorf("Expected blue-ish color for %q, got %+v", tt.damageType, got)
			}

			if got.A != 255 {
				t.Errorf("Alpha should be 255, got %d", got.A)
			}
		})
	}
}

func TestSystem_Update(t *testing.T) {
	w := engine.NewWorld()
	sys := NewSystem("fantasy")

	ent1 := Spawn(w, 50, 100.0, 100.0, "physical", false, false)
	ent2 := Spawn(w, 100, 150.0, 150.0, "fire", true, false)

	comp1, _ := w.GetComponent(ent1, nil)
	dmg1 := comp1.(*Component)
	initialY1 := dmg1.Y

	comp2, _ := w.GetComponent(ent2, nil)
	dmg2 := comp2.(*Component)
	initialY2 := dmg2.Y

	for i := 0; i < 10; i++ {
		sys.Update(w)
	}

	comp1, found1 := w.GetComponent(ent1, nil)
	if !found1 {
		t.Fatal("Entity 1 was removed prematurely")
	}
	dmg1 = comp1.(*Component)

	if dmg1.Age <= 0 {
		t.Errorf("Age should increase, got %f", dmg1.Age)
	}
	if dmg1.Y >= initialY1 {
		t.Errorf("Y should decrease (rise up), was %f now %f", initialY1, dmg1.Y)
	}

	comp2, found2 := w.GetComponent(ent2, nil)
	if !found2 {
		t.Fatal("Entity 2 was removed prematurely")
	}
	dmg2 = comp2.(*Component)

	if dmg2.Y >= initialY2 {
		t.Errorf("Critical Y should decrease (rise up), was %f now %f", initialY2, dmg2.Y)
	}
}

func TestSystem_Update_Expiration(t *testing.T) {
	w := engine.NewWorld()
	sys := NewSystem("fantasy")

	ent := Spawn(w, 10, 100.0, 100.0, "physical", false, false)
	comp, _ := w.GetComponent(ent, nil)
	dmg := comp.(*Component)
	dmg.Lifetime = 0.1

	for i := 0; i < 10; i++ {
		sys.Update(w)
	}

	_, found := w.GetComponent(ent, nil)
	if found {
		t.Error("Entity should be removed after lifetime expires")
	}
}

func TestSystem_Update_ScaleAnimation(t *testing.T) {
	w := engine.NewWorld()
	sys := NewSystem("fantasy")

	ent := Spawn(w, 50, 100.0, 100.0, "physical", false, false)

	comp, _ := w.GetComponent(ent, nil)
	dmg := comp.(*Component)
	initialScale := dmg.Scale

	for i := 0; i < 5; i++ {
		sys.Update(w)
	}

	comp, _ = w.GetComponent(ent, nil)
	dmg = comp.(*Component)

	if dmg.Scale <= initialScale {
		t.Errorf("Scale should grow during spawn-in, was %f now %f", initialScale, dmg.Scale)
	}
}

func TestSystem_Update_FadeOut(t *testing.T) {
	w := engine.NewWorld()
	sys := NewSystem("fantasy")

	ent := Spawn(w, 50, 100.0, 100.0, "physical", false, false)
	comp, _ := w.GetComponent(ent, nil)
	dmg := comp.(*Component)
	dmg.Age = dmg.Lifetime * 0.75

	sys.Update(w)

	comp, _ = w.GetComponent(ent, nil)
	dmg = comp.(*Component)

	if dmg.Alpha >= 1.0 {
		t.Errorf("Alpha should decrease during fade-out, got %f", dmg.Alpha)
	}
}

func TestNewSystem(t *testing.T) {
	tests := []string{"fantasy", "scifi", "cyberpunk", "horror"}

	for _, genreID := range tests {
		t.Run(genreID, func(t *testing.T) {
			sys := NewSystem(genreID)
			if sys == nil {
				t.Fatal("NewSystem() returned nil")
			}
			if sys.genreID != genreID {
				t.Errorf("genreID = %q, want %q", sys.genreID, genreID)
			}
			if sys.logger == nil {
				t.Error("logger is nil")
			}
			if sys.font == nil {
				t.Error("font is nil")
			}
		})
	}
}

func TestComponent_Fields(t *testing.T) {
	comp := &Component{
		Value:      123,
		DamageType: "test",
		IsCritical: true,
		IsHeal:     false,
		X:          1.0,
		Y:          2.0,
		VelocityY:  3.0,
		Lifetime:   4.0,
		Age:        5.0,
		Scale:      6.0,
		Alpha:      0.5,
		Color:      color.RGBA{1, 2, 3, 4},
	}

	if comp.Value != 123 {
		t.Errorf("Value = %d, want 123", comp.Value)
	}
	if comp.DamageType != "test" {
		t.Errorf("DamageType = %q, want %q", comp.DamageType, "test")
	}
	if !comp.IsCritical {
		t.Error("IsCritical should be true")
	}
	if comp.IsHeal {
		t.Error("IsHeal should be false")
	}
	if comp.Color.R != 1 || comp.Color.G != 2 || comp.Color.B != 3 || comp.Color.A != 4 {
		t.Errorf("Color = %+v, want {1 2 3 4}", comp.Color)
	}
}

func BenchmarkSpawn(b *testing.B) {
	w := engine.NewWorld()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Spawn(w, 50, 100.0, 100.0, "physical", false, false)
	}
}

func BenchmarkUpdate(b *testing.B) {
	w := engine.NewWorld()
	sys := NewSystem("fantasy")

	for i := 0; i < 100; i++ {
		Spawn(w, 50, float64(i*10), float64(i*10), "physical", i%5 == 0, false)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(w)
	}
}
