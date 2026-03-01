package feedback

import (
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewFeedbackSystem(t *testing.T) {
	fs := NewFeedbackSystem(12345)

	if fs == nil {
		t.Fatal("NewFeedbackSystem returned nil")
	}

	if fs.screenShake == nil {
		t.Error("screenShake not initialized")
	}

	if fs.hitFlash == nil {
		t.Error("hitFlash not initialized")
	}

	if fs.damageNumbers == nil {
		t.Error("damageNumbers not initialized")
	}

	if fs.impactEffects == nil {
		t.Error("impactEffects not initialized")
	}

	if fs.genre != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %s", fs.genre)
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name      string
		genre     string
		expectedR uint8
		expectedG uint8
		expectedB uint8
	}{
		{"fantasy", "fantasy", 255, 0, 0},
		{"cyberpunk", "cyberpunk", 0, 255, 255},
		{"horror", "horror", 180, 0, 0},
		{"scifi", "scifi", 100, 200, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFeedbackSystem(12345)
			fs.SetGenre(tt.genre)

			if fs.genre != tt.genre {
				t.Errorf("genre not set correctly: got %s, want %s", fs.genre, tt.genre)
			}

			if fs.hitFlash.color.R != tt.expectedR {
				t.Errorf("hitFlash.color.R = %d, want %d", fs.hitFlash.color.R, tt.expectedR)
			}
			if fs.hitFlash.color.G != tt.expectedG {
				t.Errorf("hitFlash.color.G = %d, want %d", fs.hitFlash.color.G, tt.expectedG)
			}
			if fs.hitFlash.color.B != tt.expectedB {
				t.Errorf("hitFlash.color.B = %d, want %d", fs.hitFlash.color.B, tt.expectedB)
			}
		})
	}
}

func TestAddScreenShake(t *testing.T) {
	fs := NewFeedbackSystem(12345)

	if fs.screenShake.intensity != 0 {
		t.Errorf("initial intensity should be 0, got %f", fs.screenShake.intensity)
	}

	fs.AddScreenShake(5.0)
	if fs.screenShake.intensity != 5.0 {
		t.Errorf("intensity should be 5.0, got %f", fs.screenShake.intensity)
	}

	fs.AddScreenShake(10.0)
	if fs.screenShake.intensity != 15.0 {
		t.Errorf("intensity should be 15.0, got %f", fs.screenShake.intensity)
	}

	// Test clamping
	fs.AddScreenShake(20.0)
	if fs.screenShake.intensity > 20.0 {
		t.Errorf("intensity should be clamped to 20.0, got %f", fs.screenShake.intensity)
	}
}

func TestAddHitFlash(t *testing.T) {
	fs := NewFeedbackSystem(12345)

	if fs.hitFlash.intensity != 0 {
		t.Errorf("initial intensity should be 0, got %f", fs.hitFlash.intensity)
	}

	fs.AddHitFlash(0.5)
	if fs.hitFlash.intensity != 0.5 {
		t.Errorf("intensity should be 0.5, got %f", fs.hitFlash.intensity)
	}

	// Test clamping
	fs.AddHitFlash(0.8)
	if fs.hitFlash.intensity > 1.0 {
		t.Errorf("intensity should be clamped to 1.0, got %f", fs.hitFlash.intensity)
	}
}

func TestSpawnDamageNumber(t *testing.T) {
	fs := NewFeedbackSystem(12345)

	if len(fs.damageNumbers) != 0 {
		t.Errorf("initial damage numbers count should be 0, got %d", len(fs.damageNumbers))
	}

	fs.SpawnDamageNumber(10.0, 20.0, 50, false)
	if len(fs.damageNumbers) != 1 {
		t.Fatalf("damage numbers count should be 1, got %d", len(fs.damageNumbers))
	}

	dn := fs.damageNumbers[0]
	if dn.damage != 50 {
		t.Errorf("damage should be 50, got %d", dn.damage)
	}
	if dn.critical {
		t.Error("critical should be false")
	}
	if dn.x != 10.0 {
		t.Errorf("x should be 10.0, got %f", dn.x)
	}
	if dn.y != 20.0 {
		t.Errorf("y should be 20.0, got %f", dn.y)
	}
}

func TestSpawnDamageNumberCritical(t *testing.T) {
	fs := NewFeedbackSystem(12345)
	fs.SpawnDamageNumber(10.0, 20.0, 100, true)

	if len(fs.damageNumbers) != 1 {
		t.Fatalf("damage numbers count should be 1, got %d", len(fs.damageNumbers))
	}

	dn := fs.damageNumbers[0]
	if !dn.critical {
		t.Error("critical should be true")
	}
	if dn.damage != 100 {
		t.Errorf("damage should be 100, got %d", dn.damage)
	}
}

func TestSpawnImpactEffect(t *testing.T) {
	fs := NewFeedbackSystem(12345)

	if len(fs.impactEffects) != 0 {
		t.Errorf("initial impact effects count should be 0, got %d", len(fs.impactEffects))
	}

	fs.SpawnImpactEffect(15.0, 25.0, ImpactHit)
	if len(fs.impactEffects) != 1 {
		t.Fatalf("impact effects count should be 1, got %d", len(fs.impactEffects))
	}

	ie := fs.impactEffects[0]
	if ie.x != 15.0 {
		t.Errorf("x should be 15.0, got %f", ie.x)
	}
	if ie.y != 25.0 {
		t.Errorf("y should be 25.0, got %f", ie.y)
	}
	if ie.itype != ImpactHit {
		t.Errorf("type should be ImpactHit, got %v", ie.itype)
	}
}

func TestUpdate(t *testing.T) {
	fs := NewFeedbackSystem(12345)
	w := engine.NewWorld()

	fs.AddScreenShake(10.0)
	fs.AddHitFlash(0.8)
	fs.SpawnDamageNumber(10.0, 20.0, 50, false)
	fs.SpawnImpactEffect(15.0, 25.0, ImpactHit)

	// Run updates
	for i := 0; i < 10; i++ {
		fs.Update(w)
	}

	// Screen shake should decay
	if fs.screenShake.intensity >= 10.0 {
		t.Error("screen shake should decay after updates")
	}

	// Hit flash should decay
	if fs.hitFlash.intensity >= 0.8 {
		t.Error("hit flash should decay after updates")
	}

	// Damage numbers should move
	if len(fs.damageNumbers) > 0 {
		dn := fs.damageNumbers[0]
		if dn.y == 20.0 {
			t.Error("damage number should have moved")
		}
	}
}

func TestDamageNumberExpiration(t *testing.T) {
	fs := NewFeedbackSystem(12345)
	w := engine.NewWorld()

	fs.SpawnDamageNumber(10.0, 20.0, 50, false)

	// Run enough updates to expire the damage number (maxLife = 1.5s at 60fps = 90 frames)
	for i := 0; i < 100; i++ {
		fs.Update(w)
	}

	if len(fs.damageNumbers) != 0 {
		t.Errorf("damage number should be expired, got %d active", len(fs.damageNumbers))
	}
}

func TestImpactEffectExpiration(t *testing.T) {
	fs := NewFeedbackSystem(12345)
	w := engine.NewWorld()

	fs.SpawnImpactEffect(15.0, 25.0, ImpactHit)

	// Run enough updates to expire the impact effect (maxLife = 0.3s at 60fps = 18 frames)
	for i := 0; i < 25; i++ {
		fs.Update(w)
	}

	if len(fs.impactEffects) != 0 {
		t.Errorf("impact effect should be expired, got %d active", len(fs.impactEffects))
	}
}

func TestGetScreenShakeOffset(t *testing.T) {
	fs := NewFeedbackSystem(12345)

	x, y := fs.GetScreenShakeOffset()
	if x != 0 || y != 0 {
		t.Errorf("initial offset should be (0, 0), got (%f, %f)", x, y)
	}

	fs.AddScreenShake(5.0)
	w := engine.NewWorld()
	fs.Update(w)

	x, y = fs.GetScreenShakeOffset()
	if x == 0 && y == 0 {
		t.Error("offset should be non-zero after adding shake")
	}
}

func TestGetHitFlashIntensity(t *testing.T) {
	fs := NewFeedbackSystem(12345)

	intensity := fs.GetHitFlashIntensity()
	if intensity != 0 {
		t.Errorf("initial intensity should be 0, got %f", intensity)
	}

	fs.AddHitFlash(0.7)
	intensity = fs.GetHitFlashIntensity()
	if intensity != 0.7 {
		t.Errorf("intensity should be 0.7, got %f", intensity)
	}
}

func TestDamageNumberFormatting(t *testing.T) {
	tests := []struct {
		name     string
		damage   int
		critical bool
		expected string
	}{
		{"normal", 50, false, "-50"},
		{"critical", 100, true, "-100!"},
		{"low damage", 5, false, "-5"},
		{"high damage critical", 999, true, "-999!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFeedbackSystem(12345)
			fs.SpawnDamageNumber(10.0, 20.0, tt.damage, tt.critical)

			dn := fs.damageNumbers[0]
			formatted := dn.FormatDamageNumber()
			if formatted != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, formatted)
			}
		})
	}
}

func TestDamageNumberAlpha(t *testing.T) {
	fs := NewFeedbackSystem(12345)
	fs.SpawnDamageNumber(10.0, 20.0, 50, false)

	dn := fs.damageNumbers[0]

	// Initially should be fully opaque
	alpha := dn.GetAlpha()
	if alpha != 255 {
		t.Errorf("initial alpha should be 255, got %d", alpha)
	}

	// Advance lifetime
	dn.lifetime = dn.maxLife / 2
	alpha = dn.GetAlpha()
	if alpha < 100 || alpha > 155 {
		t.Errorf("mid-life alpha should be around 127, got %d", alpha)
	}

	// At end of life, should be transparent
	dn.lifetime = dn.maxLife
	alpha = dn.GetAlpha()
	if alpha != 0 {
		t.Errorf("end-of-life alpha should be 0, got %d", alpha)
	}
}

func TestDamageNumberScale(t *testing.T) {
	fs := NewFeedbackSystem(12345)

	fs.SpawnDamageNumber(10.0, 20.0, 50, false)
	normalDN := fs.damageNumbers[0]
	normalScale := normalDN.GetScale()
	if normalScale != 1.0 {
		t.Errorf("normal damage number scale should be 1.0, got %f", normalScale)
	}

	fs.SpawnDamageNumber(10.0, 20.0, 100, true)
	criticalDN := fs.damageNumbers[1]
	criticalScale := criticalDN.GetScale()
	if criticalScale < 1.0 || criticalScale > 2.0 {
		t.Errorf("critical damage number scale should be between 1.0 and 2.0, got %f", criticalScale)
	}
}

func TestImpactEffectColors(t *testing.T) {
	tests := []struct {
		name  string
		itype ImpactType
	}{
		{"hit", ImpactHit},
		{"critical", ImpactCritical},
		{"block", ImpactBlock},
		{"miss", ImpactMiss},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFeedbackSystem(12345)
			fs.SpawnImpactEffect(10.0, 20.0, tt.itype)

			ie := fs.impactEffects[0]
			color := ie.GetColor()

			// Verify color is not nil and has some alpha
			if color.A == 0 {
				t.Error("impact effect color should have some alpha initially")
			}

			// Verify type is set correctly
			if ie.GetType() != tt.itype {
				t.Errorf("expected type %v, got %v", tt.itype, ie.GetType())
			}
		})
	}
}

func TestMaxDamageNumbers(t *testing.T) {
	fs := NewFeedbackSystem(12345)
	fs.maxDamageNums = 10

	// Spawn more than max
	for i := 0; i < 15; i++ {
		fs.SpawnDamageNumber(float64(i), float64(i), 10, false)
	}

	// Should be capped at max
	if len(fs.damageNumbers) > 10 {
		t.Errorf("damage numbers should be capped at 10, got %d", len(fs.damageNumbers))
	}
}

func TestMaxImpactEffects(t *testing.T) {
	fs := NewFeedbackSystem(12345)
	fs.maxImpacts = 10

	// Spawn more than max
	for i := 0; i < 15; i++ {
		fs.SpawnImpactEffect(float64(i), float64(i), ImpactHit)
	}

	// Should be capped at max
	if len(fs.impactEffects) > 10 {
		t.Errorf("impact effects should be capped at 10, got %d", len(fs.impactEffects))
	}
}
