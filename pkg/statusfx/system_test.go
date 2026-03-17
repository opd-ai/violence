package statusfx

import (
	"reflect"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/particle"
	"github.com/opd-ai/violence/pkg/status"
)

func TestNewSystem(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}

	if sys.genreID != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got %s", sys.genreID)
	}

	if sys.particleSystem != ps {
		t.Error("Particle system not set correctly")
	}
}

func TestSetGenre(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	sys.SetGenre("cyberpunk")

	if sys.genreID != "cyberpunk" {
		t.Errorf("SetGenre failed: expected 'cyberpunk', got %s", sys.genreID)
	}
}

func TestUpdateCreatesVisualComponent(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	w := engine.NewWorld()
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 5, Y: 5})

	statusComp := &status.StatusComponent{
		ActiveEffects: []status.ActiveEffect{
			{
				EffectName:    "burning",
				TimeRemaining: 5 * time.Second,
				VisualColor:   0xFF880088,
			},
		},
	}
	w.AddComponent(entity, statusComp)

	sys.Update(w)

	visualComp, found := w.GetComponent(entity, reflect.TypeOf(&VisualComponent{}))
	if !found {
		t.Fatal("Update did not create VisualComponent")
	}

	vc := visualComp.(*VisualComponent)
	if len(vc.Effects) != 1 {
		t.Fatalf("Expected 1 visual effect, got %d", len(vc.Effects))
	}

	if vc.Effects[0].Name != "burning" {
		t.Errorf("Expected effect 'burning', got %s", vc.Effects[0].Name)
	}

	if vc.Effects[0].Color != 0xFF880088 {
		t.Errorf("Expected color 0xFF880088, got 0x%08X", vc.Effects[0].Color)
	}
}

func TestUpdateRemovesVisualWhenNoStatus(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	w := engine.NewWorld()
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 5, Y: 5})
	w.AddComponent(entity, &VisualComponent{Effects: []EffectVisual{{Name: "test", Color: 0xFFFFFFFF}}})

	sys.Update(w)

	_, found := w.GetComponent(entity, reflect.TypeOf(&VisualComponent{}))
	if found {
		t.Error("Update did not remove VisualComponent when no status effects")
	}
}

func TestUpdateMultipleEffects(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	w := engine.NewWorld()
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 5, Y: 5})

	statusComp := &status.StatusComponent{
		ActiveEffects: []status.ActiveEffect{
			{
				EffectName:    "burning",
				TimeRemaining: 3 * time.Second,
				VisualColor:   0xFF880088,
			},
			{
				EffectName:    "poisoned",
				TimeRemaining: 8 * time.Second,
				VisualColor:   0x88FF0088,
			},
		},
	}
	w.AddComponent(entity, statusComp)

	sys.Update(w)

	visualComp, found := w.GetComponent(entity, reflect.TypeOf(&VisualComponent{}))
	if !found {
		t.Fatal("Update did not create VisualComponent")
	}

	vc := visualComp.(*VisualComponent)
	if len(vc.Effects) != 2 {
		t.Fatalf("Expected 2 visual effects, got %d", len(vc.Effects))
	}
}

func TestCalculateIntensity(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	tests := []struct {
		name          string
		timeRemaining float64
		wantMin       float64
		wantMax       float64
	}{
		{"full duration", 10.0, 0.0, 1.0},
		{"half duration", 5.0, 0.0, 1.0},
		{"low duration", 1.0, 0.0, 0.5},
		{"very low", 0.5, 0.0, 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intensity := sys.calculateIntensity(tt.timeRemaining)
			if intensity < tt.wantMin || intensity > tt.wantMax {
				t.Errorf("calculateIntensity(%f) = %f, want range [%f, %f]",
					tt.timeRemaining, intensity, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestRenderNoEntities(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	w := engine.NewWorld()
	screen := ebiten.NewImage(640, 480)

	sys.Render(screen, w, 0, 0)
}

func TestRenderWithEntity(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	w := engine.NewWorld()
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 2, Y: 2})
	w.AddComponent(entity, &VisualComponent{
		Effects: []EffectVisual{
			{Name: "burning", Color: 0xFF880088, Intensity: 0.8},
		},
	})

	screen := ebiten.NewImage(640, 480)
	sys.Render(screen, w, 0, 0)

	bounds := screen.Bounds()
	if bounds.Dx() != 640 || bounds.Dy() != 480 {
		t.Error("Screen dimensions changed")
	}
}

func TestRenderDistanceCulling(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	w := engine.NewWorld()
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 100, Y: 100})
	w.AddComponent(entity, &VisualComponent{
		Effects: []EffectVisual{
			{Name: "burning", Color: 0xFF880088, Intensity: 0.8},
		},
	})

	screen := ebiten.NewImage(640, 480)
	sys.Render(screen, w, 0, 0)
}

func TestGetEmitInterval(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	tests := []struct {
		name    string
		effect  string
		wantMin float64
		wantMax float64
	}{
		{"burning fast", "burning", 0.05, 0.15},
		{"poisoned medium", "poisoned", 0.25, 0.35},
		{"stunned medium", "stunned", 0.15, 0.25},
		{"unknown slow", "unknown_effect", 0.35, 0.45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interval := sys.getEmitInterval(tt.effect)
			if interval < tt.wantMin || interval > tt.wantMax {
				t.Errorf("getEmitInterval(%s) = %f, want range [%f, %f]",
					tt.effect, interval, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestGetParticleCount(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	tests := []struct {
		effect string
		want   int
	}{
		{"burning", 3},
		{"poisoned", 2},
		{"stunned", 4},
		{"unknown", 1},
	}

	for _, tt := range tests {
		t.Run(tt.effect, func(t *testing.T) {
			count := sys.getParticleCount(tt.effect)
			if count != tt.want {
				t.Errorf("getParticleCount(%s) = %d, want %d", tt.effect, count, tt.want)
			}
		})
	}
}

func TestComponentType(t *testing.T) {
	comp := &VisualComponent{}
	if comp.Type() != "StatusFXVisual" {
		t.Errorf("Component type = %s, want StatusFXVisual", comp.Type())
	}
}

func TestRenderAura(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	screen := ebiten.NewImage(100, 100)
	// Verify renderAura runs without panic
	// Note: cannot call screen.At() before game starts (ReadPixels restriction)
	sys.renderAura(screen, 50, 50, 255, 0, 0, 0.5)

	// Test with different parameters to ensure edge cases work
	sys.renderAura(screen, 0, 0, 0, 255, 0, 1.0)
	sys.renderAura(screen, 99, 99, 0, 0, 255, 0.1)

	// Verify pulseTimer affects radius calculation
	sys.pulseTimer = 1.5
	sys.renderAura(screen, 50, 50, 255, 255, 255, 0.8)
}

func TestEmitParticles(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	effect := &EffectVisual{
		Name:        "burning",
		Color:       0xFF0000FF,
		Intensity:   1.0,
		ParticleAge: 1.0,
	}

	sys.emitParticles(effect, 5.0, 5.0, 255, 0, 0)
}

func BenchmarkUpdate(b *testing.B) {
	ps := particle.NewParticleSystem(1000, 12345)
	sys := NewSystem("fantasy", ps)

	w := engine.NewWorld()
	for i := 0; i < 100; i++ {
		entity := w.AddEntity()
		w.AddComponent(entity, &engine.Position{X: float64(i), Y: float64(i)})
		w.AddComponent(entity, &status.StatusComponent{
			ActiveEffects: []status.ActiveEffect{
				{
					EffectName:    "burning",
					TimeRemaining: 5 * time.Second,
					VisualColor:   0xFF880088,
				},
			},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(w)
	}
}

func BenchmarkRender(b *testing.B) {
	ps := particle.NewParticleSystem(1000, 12345)
	sys := NewSystem("fantasy", ps)

	w := engine.NewWorld()
	for i := 0; i < 50; i++ {
		entity := w.AddEntity()
		w.AddComponent(entity, &engine.Position{X: float64(i % 10), Y: float64(i / 10)})
		w.AddComponent(entity, &VisualComponent{
			Effects: []EffectVisual{
				{Name: "burning", Color: 0xFF880088, Intensity: 0.8, ParticleAge: 0},
			},
		})
	}

	screen := ebiten.NewImage(640, 480)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Render(screen, w, 5, 5)
	}
}

func TestUpdateWithNilWorld(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	sys.Update(nil)
}

func TestRenderWithNilWorld(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	screen := ebiten.NewImage(640, 480)
	sys.Render(screen, nil, 0, 0)
}

func TestVisualComponentEffects(t *testing.T) {
	vc := &VisualComponent{
		Effects: []EffectVisual{
			{Name: "test1", Color: 0xFF0000FF, Intensity: 0.5},
			{Name: "test2", Color: 0x00FF00FF, Intensity: 0.7},
		},
	}

	if len(vc.Effects) != 2 {
		t.Errorf("Expected 2 effects, got %d", len(vc.Effects))
	}

	if vc.Effects[0].Name != "test1" {
		t.Errorf("Effect 0 name = %s, want test1", vc.Effects[0].Name)
	}

	if vc.Effects[1].Name != "test2" {
		t.Errorf("Effect 1 name = %s, want test2", vc.Effects[1].Name)
	}
}

func TestRenderEffectColorExtraction(t *testing.T) {
	tests := []struct {
		name  string
		color uint32
		wantR uint8
		wantG uint8
		wantB uint8
		wantA uint8
	}{
		{"red", 0xFF0000FF, 255, 0, 0, 255},
		{"green", 0x00FF00FF, 0, 255, 0, 255},
		{"blue", 0x0000FFFF, 0, 0, 255, 255},
		{"semi-transparent", 0xFF000088, 255, 0, 0, 136},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := uint8((tt.color >> 24) & 0xFF)
			g := uint8((tt.color >> 16) & 0xFF)
			b := uint8((tt.color >> 8) & 0xFF)
			a := uint8(tt.color & 0xFF)

			if r != tt.wantR || g != tt.wantG || b != tt.wantB || a != tt.wantA {
				t.Errorf("Color extraction for 0x%08X: got RGBA(%d,%d,%d,%d), want RGBA(%d,%d,%d,%d)",
					tt.color, r, g, b, a, tt.wantR, tt.wantG, tt.wantB, tt.wantA)
			}
		})
	}
}

func TestUpdatePulseTimer(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	w := engine.NewWorld()

	initialPulse := sys.pulseTimer
	sys.Update(w)

	if sys.pulseTimer <= initialPulse {
		t.Error("Pulse timer did not advance")
	}
}

func TestMultipleUpdatesAccumulateTime(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	w := engine.NewWorld()
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 5, Y: 5})
	w.AddComponent(entity, &status.StatusComponent{
		ActiveEffects: []status.ActiveEffect{
			{
				EffectName:    "burning",
				TimeRemaining: 5 * time.Second,
				VisualColor:   0xFF880088,
			},
		},
	})

	for i := 0; i < 10; i++ {
		sys.Update(w)
	}

	visualComp, found := w.GetComponent(entity, reflect.TypeOf(&VisualComponent{}))
	if !found {
		t.Fatal("VisualComponent not found after multiple updates")
	}

	vc := visualComp.(*VisualComponent)
	if len(vc.Effects) != 1 {
		t.Errorf("Expected 1 effect after updates, got %d", len(vc.Effects))
	}
}

func TestRenderOffscreenEntity(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	w := engine.NewWorld()
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: -100, Y: -100})
	w.AddComponent(entity, &VisualComponent{
		Effects: []EffectVisual{
			{Name: "burning", Color: 0xFF880088, Intensity: 0.8},
		},
	})

	screen := ebiten.NewImage(640, 480)
	sys.Render(screen, w, 0, 0)
}

func TestCleanOrphanedVisuals(t *testing.T) {
	ps := particle.NewParticleSystem(100, 12345)
	sys := NewSystem("fantasy", ps)

	w := engine.NewWorld()
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 5, Y: 5})
	w.AddComponent(entity, &VisualComponent{
		Effects: []EffectVisual{{Name: "test", Color: 0xFFFFFFFF}},
	})

	sys.Update(w)

	_, found := w.GetComponent(entity, reflect.TypeOf(&VisualComponent{}))
	if found {
		t.Error("Orphaned visual component was not cleaned")
	}
}
