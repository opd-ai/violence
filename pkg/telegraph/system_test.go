package telegraph

import (
	"image/color"
	"testing"
)

func TestComponent(t *testing.T) {
	c := &Component{
		Active:         true,
		ChargeProgress: 0.5,
		TelegraphTime:  1.0,
		AttackType:     "melee",
		PrimaryColor:   color.RGBA{255, 0, 0, 255},
		SecondaryColor: color.RGBA{255, 128, 0, 255},
	}

	if c.Type() != "telegraph" {
		t.Errorf("expected type 'telegraph', got %s", c.Type())
	}

	if !c.Active {
		t.Error("component should be active")
	}

	if c.ChargeProgress != 0.5 {
		t.Errorf("expected charge progress 0.5, got %f", c.ChargeProgress)
	}
}

func TestSystem_Creation(t *testing.T) {
	tests := []struct {
		name  string
		genre string
		seed  int64
	}{
		{"fantasy system", "fantasy", 12345},
		{"scifi system", "scifi", 67890},
		{"horror system", "horror", 11111},
		{"cyberpunk system", "cyberpunk", 22222},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem(tt.genre, tt.seed)
			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.genre != tt.genre {
				t.Errorf("expected genre %s, got %s", tt.genre, sys.genre)
			}
		})
	}
}

func TestSystem_GetColorScheme(t *testing.T) {
	tests := []struct {
		genre      string
		attackType string
	}{
		{"fantasy", "melee"},
		{"fantasy", "ranged"},
		{"fantasy", "aoe"},
		{"fantasy", "charge"},
		{"scifi", "melee"},
		{"scifi", "ranged"},
		{"horror", "aoe"},
		{"cyberpunk", "charge"},
	}

	for _, tt := range tests {
		t.Run(tt.genre+"_"+tt.attackType, func(t *testing.T) {
			sys := NewSystem(tt.genre, 42)
			primary, secondary := sys.getColorScheme(tt.attackType)

			// Colors should not be zero
			if primary.R == 0 && primary.G == 0 && primary.B == 0 {
				t.Error("primary color should not be black")
			}
			if secondary.R == 0 && secondary.G == 0 && secondary.B == 0 {
				t.Error("secondary color should not be black")
			}
			// Alpha should be full
			if primary.A != 255 {
				t.Errorf("primary alpha should be 255, got %d", primary.A)
			}
			if secondary.A != 255 {
				t.Errorf("secondary alpha should be 255, got %d", secondary.A)
			}
		})
	}
}

func TestSystem_RenderMethods(t *testing.T) {
	// Skip if no display available (CI environment)
	t.Skip("Skipping render tests that require display")

	sys := NewSystem("fantasy", 42)

	comp := &Component{
		Active:          true,
		ChargeProgress:  0.5,
		TelegraphTime:   1.0,
		PrimaryColor:    color.RGBA{255, 0, 0, 255},
		SecondaryColor:  color.RGBA{255, 128, 0, 255},
		IndicatorRadius: 16.0,
		IndicatorAlpha:  0.7,
		X:               100,
		Y:               100,
	}

	// Test each telegraph type render method (should not panic)
	tests := []struct {
		name       string
		attackType string
	}{
		{"melee", "melee"},
		{"ranged", "ranged"},
		{"aoe", "aoe"},
		{"charge", "charge"},
		{"generic", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp.AttackType = tt.attackType

			// Call the render methods directly - should not panic
			switch tt.attackType {
			case "melee":
				sys.drawMeleeTelegraph(100, 100, comp)
			case "ranged":
				sys.drawRangedTelegraph(100, 100, comp)
			case "aoe":
				sys.drawAoETelegraph(100, 100, comp)
			case "charge":
				sys.drawChargeTelegraph(100, 100, comp)
			default:
				sys.drawGenericTelegraph(100, 100, comp)
			}
		})
	}
}

func TestComponent_ProgressionStates(t *testing.T) {
	tests := []struct {
		name     string
		progress float64
		wantMsg  string
	}{
		{"start", 0.0, "just started"},
		{"quarter", 0.25, "quarter charged"},
		{"half", 0.5, "half charged"},
		{"three-quarters", 0.75, "nearly charged"},
		{"complete", 1.0, "fully charged"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Component{
				Active:         true,
				ChargeProgress: tt.progress,
				TelegraphTime:  1.0,
			}

			if c.ChargeProgress != tt.progress {
				t.Errorf("expected progress %f, got %f", tt.progress, c.ChargeProgress)
			}

			// Verify progress is in valid range
			if c.ChargeProgress < 0.0 || c.ChargeProgress > 1.0 {
				t.Errorf("progress %f out of range [0, 1]", c.ChargeProgress)
			}
		})
	}
}

func TestComponent_AttackTypes(t *testing.T) {
	types := []string{"melee", "ranged", "aoe", "charge"}

	for _, attackType := range types {
		t.Run(attackType, func(t *testing.T) {
			c := &Component{
				Active:        true,
				AttackType:    attackType,
				TelegraphTime: 1.0,
			}

			if c.AttackType != attackType {
				t.Errorf("expected attack type %s, got %s", attackType, c.AttackType)
			}
		})
	}
}

func BenchmarkSystem_DrawMelee(b *testing.B) {
	b.Skip("Skipping benchmark that requires display")
	sys := NewSystem("fantasy", 42)
	comp := &Component{
		Active:          true,
		ChargeProgress:  0.5,
		PrimaryColor:    color.RGBA{255, 0, 0, 255},
		SecondaryColor:  color.RGBA{255, 128, 0, 255},
		IndicatorRadius: 16.0,
		IndicatorAlpha:  0.7,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.drawMeleeTelegraph(100, 100, comp)
	}
}

func BenchmarkSystem_DrawAoE(b *testing.B) {
	b.Skip("Skipping benchmark that requires display")
	sys := NewSystem("fantasy", 42)
	comp := &Component{
		Active:          true,
		ChargeProgress:  0.5,
		PrimaryColor:    color.RGBA{255, 0, 0, 255},
		SecondaryColor:  color.RGBA{255, 128, 0, 255},
		IndicatorRadius: 32.0,
		IndicatorAlpha:  0.7,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.drawAoETelegraph(200, 200, comp)
	}
}

func TestComponent_TypeString(t *testing.T) {
	c := &Component{}
	if c.Type() != "telegraph" {
		t.Errorf("Type() should return 'telegraph', got '%s'", c.Type())
	}
}

func TestSystem_GenreColorSchemes(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}
	attackTypes := []string{"melee", "ranged", "aoe", "charge"}

	for _, genre := range genres {
		for _, attackType := range attackTypes {
			t.Run(genre+"_"+attackType, func(t *testing.T) {
				sys := NewSystem(genre, 42)
				primary, secondary := sys.getColorScheme(attackType)

				// Verify we get valid colors
				if primary.A != 255 || secondary.A != 255 {
					t.Error("alpha channels should be 255")
				}

				// Colors should be distinct
				if primary.R == secondary.R && primary.G == secondary.G && primary.B == secondary.B {
					t.Error("primary and secondary should be different colors")
				}
			})
		}
	}
}

func TestComponent_FieldValidation(t *testing.T) {
	c := &Component{
		Active:          true,
		ChargeProgress:  0.75,
		TelegraphTime:   2.0,
		AttackType:      "aoe",
		PrimaryColor:    color.RGBA{255, 100, 50, 255},
		SecondaryColor:  color.RGBA{200, 150, 100, 255},
		IndicatorRadius: 24.0,
		IndicatorAlpha:  0.8,
		X:               150.5,
		Y:               200.3,
		EmitParticles:   true,
		ParticleCount:   5,
		ParticleSpread:  12.0,
		ScreenShake:     2.5,
		FlashIntensity:  0.6,
	}

	// Verify all fields are set correctly
	if !c.Active {
		t.Error("Active should be true")
	}
	if c.ChargeProgress != 0.75 {
		t.Errorf("ChargeProgress should be 0.75, got %f", c.ChargeProgress)
	}
	if c.TelegraphTime != 2.0 {
		t.Errorf("TelegraphTime should be 2.0, got %f", c.TelegraphTime)
	}
	if c.AttackType != "aoe" {
		t.Errorf("AttackType should be 'aoe', got '%s'", c.AttackType)
	}
	if c.IndicatorRadius != 24.0 {
		t.Errorf("IndicatorRadius should be 24.0, got %f", c.IndicatorRadius)
	}
	if c.X != 150.5 || c.Y != 200.3 {
		t.Errorf("Position should be (150.5, 200.3), got (%f, %f)", c.X, c.Y)
	}
	if !c.EmitParticles {
		t.Error("EmitParticles should be true")
	}
	if c.ParticleCount != 5 {
		t.Errorf("ParticleCount should be 5, got %d", c.ParticleCount)
	}
}

func TestPositionComponent(t *testing.T) {
	pos := &PositionComponent{X: 100.5, Y: 200.7}

	if pos.Type() != "Position" {
		t.Errorf("Expected Type() 'Position', got '%s'", pos.Type())
	}

	if pos.X != 100.5 || pos.Y != 200.7 {
		t.Errorf("Expected position (100.5, 200.7), got (%f, %f)", pos.X, pos.Y)
	}
}

func TestComponent_ZeroValues(t *testing.T) {
	c := &Component{}

	// Zero-valued component should be inactive
	if c.Active {
		t.Error("Default component should be inactive")
	}
	if c.ChargeProgress != 0.0 {
		t.Error("Default ChargeProgress should be 0.0")
	}
	if c.TelegraphTime != 0.0 {
		t.Error("Default TelegraphTime should be 0.0")
	}
	if c.AttackType != "" {
		t.Error("Default AttackType should be empty")
	}
}

func TestSystem_NilSafety(t *testing.T) {
	sys := NewSystem("fantasy", 123)

	if sys == nil {
		t.Fatal("NewSystem should never return nil")
	}

	if sys.logger == nil {
		t.Error("logger should be initialized")
	}

	if sys.rng == nil {
		t.Error("rng should be initialized")
	}

	// telegraphLayer is lazy initialized, so it should be nil initially
	if sys.telegraphLayer != nil {
		t.Error("telegraphLayer should be nil before first render")
	}
}
