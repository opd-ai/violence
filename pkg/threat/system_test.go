package threat

import (
	"image/color"
	"math"
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewSystem(t *testing.T) {
	s := NewSystem("fantasy")
	if s == nil {
		t.Fatal("NewSystem returned nil")
	}
	if s.genreID != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got %q", s.genreID)
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name        string
		genre       string
		wantPrimary color.RGBA
	}{
		{"fantasy", "fantasy", color.RGBA{R: 255, G: 180, B: 0, A: 255}},
		{"cyberpunk", "cyberpunk", color.RGBA{R: 255, G: 0, B: 180, A: 255}},
		{"horror", "horror", color.RGBA{R: 180, G: 0, B: 0, A: 255}},
		{"scifi", "scifi", color.RGBA{R: 0, G: 200, B: 255, A: 255}},
		{"postapoc", "postapoc", color.RGBA{R: 255, G: 120, B: 0, A: 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSystem("fantasy")
			s.SetGenre(tt.genre)
			if s.style.PrimaryColor != tt.wantPrimary {
				t.Errorf("Primary color mismatch for %s: got %v, want %v",
					tt.genre, s.style.PrimaryColor, tt.wantPrimary)
			}
		})
	}
}

func TestSetScreenSize(t *testing.T) {
	s := NewSystem("fantasy")
	s.SetScreenSize(640, 480)
	if s.screenWidth != 640 || s.screenHeight != 480 {
		t.Errorf("Screen size not set correctly: got %dx%d, want 640x480",
			s.screenWidth, s.screenHeight)
	}
}

func TestSetPlayerPosition(t *testing.T) {
	s := NewSystem("fantasy")
	s.SetPlayerPosition(100.5, 200.5)
	if s.playerX != 100.5 || s.playerY != 200.5 {
		t.Errorf("Player position not set correctly: got (%v, %v), want (100.5, 200.5)",
			s.playerX, s.playerY)
	}
}

func TestNewComponent(t *testing.T) {
	c := NewComponent()
	if c == nil {
		t.Fatal("NewComponent returned nil")
	}
	if c.ThreatLevel != ThreatNone {
		t.Errorf("Default threat level should be ThreatNone, got %v", c.ThreatLevel)
	}
	if c.Type() != "threat" {
		t.Errorf("Component type should be 'threat', got %q", c.Type())
	}
}

func TestMarkThreat(t *testing.T) {
	s := NewSystem("fantasy")
	w := engine.NewWorld()
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 10, Y: 20})

	// Mark threat - should add component
	s.MarkThreat(w, entity, ThreatHigh, 3.0)

	// Verify component was added
	compType := reflect.TypeOf(&Component{})
	comp, ok := w.GetComponent(entity, compType)
	if !ok {
		t.Fatal("Component was not added after MarkThreat")
	}

	tc := comp.(*Component)
	if tc.ThreatLevel != ThreatHigh {
		t.Errorf("Threat level should be ThreatHigh, got %v", tc.ThreatLevel)
	}
	if tc.ThreatDecay != 3.0 {
		t.Errorf("Threat decay should be 3.0, got %v", tc.ThreatDecay)
	}
}

func TestMarkThreatExistingComponent(t *testing.T) {
	s := NewSystem("fantasy")
	w := engine.NewWorld()
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 10, Y: 20})

	// Add initial component
	tc := NewComponent()
	tc.ThreatLevel = ThreatLow
	tc.ThreatDecay = 1.0
	w.AddComponent(entity, tc)

	// Mark higher threat - should upgrade
	s.MarkThreat(w, entity, ThreatHigh, 5.0)

	compType := reflect.TypeOf(&Component{})
	comp, _ := w.GetComponent(entity, compType)
	updated := comp.(*Component)

	if updated.ThreatLevel != ThreatHigh {
		t.Errorf("Threat level should upgrade to ThreatHigh, got %v", updated.ThreatLevel)
	}
	if updated.ThreatDecay != 5.0 {
		t.Errorf("Threat decay should upgrade to 5.0, got %v", updated.ThreatDecay)
	}
}

func TestMarkThreatNoDowngrade(t *testing.T) {
	s := NewSystem("fantasy")
	w := engine.NewWorld()
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 10, Y: 20})

	// Add high threat component
	tc := NewComponent()
	tc.ThreatLevel = ThreatCritical
	tc.ThreatDecay = 10.0
	w.AddComponent(entity, tc)

	// Mark lower threat - should not downgrade
	s.MarkThreat(w, entity, ThreatLow, 2.0)

	compType := reflect.TypeOf(&Component{})
	comp, _ := w.GetComponent(entity, compType)
	updated := comp.(*Component)

	if updated.ThreatLevel != ThreatCritical {
		t.Errorf("Threat level should remain ThreatCritical, got %v", updated.ThreatLevel)
	}
	// Decay should still be 10.0 since 2.0 < 10.0
	if updated.ThreatDecay != 10.0 {
		t.Errorf("Threat decay should remain 10.0, got %v", updated.ThreatDecay)
	}
}

func TestMarkAttackWindup(t *testing.T) {
	s := NewSystem("fantasy")
	w := engine.NewWorld()
	entity := w.AddEntity()

	tc := NewComponent()
	w.AddComponent(entity, tc)

	// Mark windup start
	s.MarkAttackWindup(w, entity, true)
	if !tc.AttackWindup {
		t.Error("AttackWindup should be true")
	}

	// Mark windup end
	s.MarkAttackWindup(w, entity, false)
	if tc.AttackWindup {
		t.Error("AttackWindup should be false")
	}
	if tc.WindupProgress != 0 {
		t.Errorf("WindupProgress should be reset to 0, got %v", tc.WindupProgress)
	}
}

func TestAddOffscreenThreat(t *testing.T) {
	s := NewSystem("fantasy")

	// Add first indicator
	s.AddOffscreenThreat(0, 10.0, ThreatHigh)
	if s.GetOffscreenIndicatorCount() != 1 {
		t.Errorf("Expected 1 indicator, got %d", s.GetOffscreenIndicatorCount())
	}

	// Add more
	for i := 0; i < 10; i++ {
		s.AddOffscreenThreat(float64(i)*0.5, 5.0, ThreatMedium)
	}

	// Should be capped at maxOffscreen (8)
	if s.GetOffscreenIndicatorCount() > s.maxOffscreen {
		t.Errorf("Indicators exceeded max: got %d, max %d",
			s.GetOffscreenIndicatorCount(), s.maxOffscreen)
	}
}

func TestClearOffscreenIndicators(t *testing.T) {
	s := NewSystem("fantasy")
	s.AddOffscreenThreat(0, 10.0, ThreatHigh)
	s.AddOffscreenThreat(math.Pi, 15.0, ThreatMedium)

	s.ClearOffscreenIndicators()

	if s.GetOffscreenIndicatorCount() != 0 {
		t.Errorf("Expected 0 indicators after clear, got %d", s.GetOffscreenIndicatorCount())
	}
}

func TestThreatLevelToAlpha(t *testing.T) {
	s := NewSystem("fantasy")

	tests := []struct {
		level     ThreatLevel
		wantAlpha float64
	}{
		{ThreatNone, 0.0},
		{ThreatLow, 0.35},
		{ThreatMedium, 0.6},
		{ThreatHigh, 0.85},
		{ThreatCritical, 1.0},
	}

	for _, tt := range tests {
		got := s.threatLevelToAlpha(tt.level)
		if got != tt.wantAlpha {
			t.Errorf("threatLevelToAlpha(%v) = %v, want %v", tt.level, got, tt.wantAlpha)
		}
	}
}

func TestDecreaseThreatLevel(t *testing.T) {
	s := NewSystem("fantasy")

	// Normal entity should decrease
	tc := &Component{ThreatLevel: ThreatHigh}
	s.decreaseThreatLevel(tc)
	if tc.ThreatLevel != ThreatMedium {
		t.Errorf("Expected ThreatMedium after decrease, got %v", tc.ThreatLevel)
	}

	// Should stop at ThreatNone
	tc.ThreatLevel = ThreatLow
	s.decreaseThreatLevel(tc)
	if tc.ThreatLevel != ThreatNone {
		t.Errorf("Expected ThreatNone, got %v", tc.ThreatLevel)
	}
	s.decreaseThreatLevel(tc)
	if tc.ThreatLevel != ThreatNone {
		t.Errorf("ThreatNone should not decrease further, got %v", tc.ThreatLevel)
	}
}

func TestDecreaseThreatLevelBoss(t *testing.T) {
	s := NewSystem("fantasy")

	// Boss should not go below ThreatMedium
	tc := &Component{ThreatLevel: ThreatMedium, IsBoss: true}
	s.decreaseThreatLevel(tc)
	if tc.ThreatLevel != ThreatMedium {
		t.Errorf("Boss should stay at ThreatMedium, got %v", tc.ThreatLevel)
	}

	// Boss at high can decrease to medium
	tc.ThreatLevel = ThreatHigh
	s.decreaseThreatLevel(tc)
	if tc.ThreatLevel != ThreatMedium {
		t.Errorf("Boss should decrease to ThreatMedium, got %v", tc.ThreatLevel)
	}
}

func TestLerp(t *testing.T) {
	tests := []struct {
		a, b, t, want float64
	}{
		{0, 10, 0, 0},
		{0, 10, 1, 10},
		{0, 10, 0.5, 5},
		{0, 10, -0.5, 0}, // Clamped
		{0, 10, 1.5, 10}, // Clamped
		{5, 15, 0.25, 7.5},
	}

	for _, tt := range tests {
		got := lerp(tt.a, tt.b, tt.t)
		if math.Abs(got-tt.want) > 0.0001 {
			t.Errorf("lerp(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.t, got, tt.want)
		}
	}
}

func TestGetStyle(t *testing.T) {
	s := NewSystem("cyberpunk")
	style := s.GetStyle()
	if style.PrimaryColor.R != 255 || style.PrimaryColor.G != 0 {
		t.Errorf("Unexpected style for cyberpunk: %v", style.PrimaryColor)
	}
}

func TestWorldToScreen(t *testing.T) {
	s := NewSystem("fantasy")
	s.SetScreenSize(320, 200)

	// Entity at camera position should be at center
	x, y, visible := s.worldToScreen(0, 0, 0, 0)
	if !visible {
		t.Error("Entity at camera should be visible")
	}
	if x != 160 || y != 100 {
		t.Errorf("Expected center (160, 100), got (%v, %v)", x, y)
	}

	// Entity too far should not be visible
	_, _, visible = s.worldToScreen(100, 100, 0, 0)
	if visible {
		t.Error("Distant entity should not be visible")
	}
}

// BenchmarkUpdate benchmarks the Update method.
func BenchmarkUpdate(b *testing.B) {
	s := NewSystem("fantasy")
	w := engine.NewWorld()

	// Add 100 entities with threat components
	for i := 0; i < 100; i++ {
		entity := w.AddEntity()
		w.AddComponent(entity, &engine.Position{X: float64(i), Y: float64(i)})
		tc := NewComponent()
		tc.ThreatLevel = ThreatLevel(i % 5)
		tc.ThreatDecay = float64(i % 10)
		w.AddComponent(entity, tc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Update(w)
	}
}

// BenchmarkAddOffscreenThreat benchmarks adding off-screen indicators.
func BenchmarkAddOffscreenThreat(b *testing.B) {
	s := NewSystem("fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ClearOffscreenIndicators()
		for j := 0; j < 8; j++ {
			s.AddOffscreenThreat(float64(j)*0.5, 10.0, ThreatHigh)
		}
	}
}
