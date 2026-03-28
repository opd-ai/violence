package hitmarker

import (
	"reflect"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewSystem(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre)

			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.genreID != genre {
				t.Errorf("genreID = %q, want %q", sys.genreID, genre)
			}
			if sys.markerCache == nil {
				t.Error("markerCache should be initialized")
			}
			if sys.logger == nil {
				t.Error("logger should be initialized")
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy")

	// Cache a marker
	_ = sys.getMarkerImage(HitNormal)
	if len(sys.markerCache) == 0 {
		t.Error("markerCache should have entry after getMarkerImage")
	}

	// Change genre should clear cache
	sys.SetGenre("cyberpunk")

	if sys.genreID != "cyberpunk" {
		t.Errorf("genreID = %q, want %q", sys.genreID, "cyberpunk")
	}
	if len(sys.markerCache) != 0 {
		t.Error("markerCache should be cleared after SetGenre")
	}
}

func TestSetScreenSize(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(640, 480)

	if sys.screenWidth != 640 {
		t.Errorf("screenWidth = %d, want 640", sys.screenWidth)
	}
	if sys.screenHeight != 480 {
		t.Errorf("screenHeight = %d, want 480", sys.screenHeight)
	}
}

func TestGenreColorThemes(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre)

			// All colors should be non-zero (not black)
			if sys.normalColor.R == 0 && sys.normalColor.G == 0 && sys.normalColor.B == 0 {
				t.Error("normalColor should not be black")
			}
			if sys.criticalColor.R == 0 && sys.criticalColor.G == 0 && sys.criticalColor.B == 0 {
				t.Error("criticalColor should not be black")
			}
			if sys.killColor.R == 0 && sys.killColor.G == 0 && sys.killColor.B == 0 {
				t.Error("killColor should not be black")
			}
			if sys.headshotColor.R == 0 && sys.headshotColor.G == 0 && sys.headshotColor.B == 0 {
				t.Error("headshotColor should not be black")
			}
			if sys.weakpointColor.R == 0 && sys.weakpointColor.G == 0 && sys.weakpointColor.B == 0 {
				t.Error("weakpointColor should not be black")
			}

			// Colors should have full alpha
			if sys.normalColor.A != 255 {
				t.Errorf("normalColor.A = %d, want 255", sys.normalColor.A)
			}
		})
	}
}

func TestGetColorForHitType(t *testing.T) {
	sys := NewSystem("fantasy")

	tests := []struct {
		hitType  HitType
		wantName string
	}{
		{HitNormal, "normal"},
		{HitCritical, "critical"},
		{HitKill, "kill"},
		{HitHeadshot, "headshot"},
		{HitWeakpoint, "weakpoint"},
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			color := sys.getColorForHitType(tt.hitType)
			if color.A == 0 {
				t.Error("Color alpha should not be 0")
			}
		})
	}
}

func TestGenerateMarkerImage(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)

	hitTypes := []HitType{HitNormal, HitCritical, HitKill, HitHeadshot, HitWeakpoint}

	for _, ht := range hitTypes {
		t.Run(hitTypeName(ht), func(t *testing.T) {
			img := sys.generateMarkerImage(ht)
			if img == nil {
				t.Fatal("generateMarkerImage returned nil")
			}

			bounds := img.Bounds()
			if bounds.Dx() < 20 || bounds.Dy() < 20 {
				t.Errorf("Marker too small: %dx%d", bounds.Dx(), bounds.Dy())
			}
			if bounds.Dx() > 40 || bounds.Dy() > 40 {
				t.Errorf("Marker too large: %dx%d", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestMarkerCaching(t *testing.T) {
	sys := NewSystem("fantasy")

	// First call should generate
	img1 := sys.getMarkerImage(HitNormal)
	if img1 == nil {
		t.Fatal("First getMarkerImage returned nil")
	}

	// Second call should return cached
	img2 := sys.getMarkerImage(HitNormal)
	if img2 != img1 {
		t.Error("Second call should return same cached image")
	}

	// Different type should generate new
	img3 := sys.getMarkerImage(HitKill)
	if img3 == img1 {
		t.Error("Different hit type should generate different image")
	}
}

func TestUpdateAnimation(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)

	world := engine.NewWorld()
	ent := world.AddEntity()
	comp := NewComponent()
	comp.Trigger(HitNormal, 50, 160, 100)
	world.AddComponent(ent, comp)

	// Initial state
	initialScale := comp.Scale
	initialAlpha := comp.Alpha

	// Update once
	sys.Update(world)

	// Age should increase
	if comp.Age <= 0 {
		t.Error("Age should increase after update")
	}

	// Scale should change during pop animation
	if comp.Scale == initialScale && comp.Age > 0 {
		t.Error("Scale should change during animation")
	}

	// Alpha should still be full early in animation
	if comp.Alpha != initialAlpha && comp.Age < comp.Duration*0.6 {
		t.Error("Alpha should not fade early in animation")
	}

	// Run until complete
	for i := 0; i < 60; i++ {
		sys.Update(world)
	}

	// Should be reset after duration
	if comp.Active {
		t.Error("Component should be inactive after duration")
	}
}

func TestUpdateFadeOut(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()
	ent := world.AddEntity()
	comp := NewComponent()
	comp.Trigger(HitNormal, 50, 160, 100)
	world.AddComponent(ent, comp)

	// Advance to fade phase (after 60% of duration)
	comp.Age = comp.Duration * 0.7
	sys.Update(world)

	// Alpha should be less than 1.0 in fade phase
	if comp.Alpha >= 1.0 {
		t.Errorf("Alpha = %f, should be < 1.0 in fade phase", comp.Alpha)
	}
}

func TestSpawnHitMarker(t *testing.T) {
	world := engine.NewWorld()
	ent := SpawnHitMarker(world)

	// Entity ID 0 is valid (first entity in empty world)
	// Check component was added
	compType := reflect.TypeOf((*Component)(nil))
	comp, found := world.GetComponent(ent, compType)
	if !found {
		t.Error("Entity should have hitmarker component")
	}

	hm, ok := comp.(*Component)
	if !ok {
		t.Error("Component should be *Component type")
	}
	if hm.Active {
		t.Error("Spawned marker should not be active until triggered")
	}
}

func TestTriggerHit(t *testing.T) {
	world := engine.NewWorld()
	ent := SpawnHitMarker(world)

	TriggerHit(world, ent, HitCritical, 100, 160, 100)

	compType := reflect.TypeOf((*Component)(nil))
	comp, found := world.GetComponent(ent, compType)
	if !found {
		t.Fatal("Component not found")
	}

	hm := comp.(*Component)
	if !hm.Active {
		t.Error("Component should be active after TriggerHit")
	}
	if hm.HitType != HitCritical {
		t.Errorf("HitType = %v, want HitCritical", hm.HitType)
	}
	if hm.DamageValue != 100 {
		t.Errorf("DamageValue = %d, want 100", hm.DamageValue)
	}
}

func TestKillMarkerRotation(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()
	ent := world.AddEntity()
	comp := NewComponent()
	comp.Trigger(HitKill, 200, 160, 100)
	world.AddComponent(ent, comp)

	// Update a few times
	for i := 0; i < 5; i++ {
		sys.Update(world)
	}

	// Kill markers should have rotation
	if comp.Rotation == 0 && comp.Active {
		t.Error("Kill marker should have non-zero rotation during animation")
	}
}

// hitTypeName returns a string name for a hit type (for test naming).
func hitTypeName(ht HitType) string {
	switch ht {
	case HitNormal:
		return "normal"
	case HitCritical:
		return "critical"
	case HitKill:
		return "kill"
	case HitHeadshot:
		return "headshot"
	case HitWeakpoint:
		return "weakpoint"
	default:
		return "unknown"
	}
}

func BenchmarkUpdate(b *testing.B) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()

	// Create 10 hit markers
	for i := 0; i < 10; i++ {
		ent := world.AddEntity()
		comp := NewComponent()
		comp.Trigger(HitType(i%5), 50, float64(i*32), float64(i*20))
		world.AddComponent(ent, comp)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(world)
	}
}

func BenchmarkGenerateMarker(b *testing.B) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)

	// Clear cache each iteration to test generation
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.markerCache = make(map[HitType]*ebiten.Image)
		sys.generateMarkerImage(HitType(i % 5))
	}
}
