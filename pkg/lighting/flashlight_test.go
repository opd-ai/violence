package lighting

import (
	"math"
	"testing"

	"github.com/opd-ai/violence/pkg/procgen/genre"
)

func TestGetFlashlightPreset(t *testing.T) {
	tests := []struct {
		name         string
		genreID      string
		expectedName string
	}{
		{
			name:         "fantasy torch",
			genreID:      genre.Fantasy,
			expectedName: "torch",
		},
		{
			name:         "scifi headlamp",
			genreID:      genre.SciFi,
			expectedName: "headlamp",
		},
		{
			name:         "horror flashlight",
			genreID:      genre.Horror,
			expectedName: "flashlight",
		},
		{
			name:         "cyberpunk glow rod",
			genreID:      genre.Cyberpunk,
			expectedName: "glow_rod",
		},
		{
			name:         "postapoc salvaged lamp",
			genreID:      genre.PostApoc,
			expectedName: "salvaged_lamp",
		},
		{
			name:         "unknown genre",
			genreID:      "unknown",
			expectedName: "flashlight",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preset := GetFlashlightPreset(tt.genreID)
			if preset.Name != tt.expectedName {
				t.Errorf("Name = %s, want %s", preset.Name, tt.expectedName)
			}
			// Validate preset values
			if preset.Radius <= 0 {
				t.Errorf("Invalid radius: %f", preset.Radius)
			}
			if preset.Angle <= 0 || preset.Angle > math.Pi {
				t.Errorf("Invalid angle: %f", preset.Angle)
			}
			if preset.Intensity < 0 || preset.Intensity > 1 {
				t.Errorf("Invalid intensity: %f", preset.Intensity)
			}
		})
	}
}

func TestNewConeLight(t *testing.T) {
	preset := FlashlightPreset{
		Name:      "test",
		Radius:    10.0,
		Angle:     math.Pi / 6,
		Intensity: 0.8,
		R:         1.0,
		G:         0.9,
		B:         0.8,
	}

	cl := NewConeLight(5.0, 10.0, 1.0, 0.0, preset)

	if cl.X != 5.0 {
		t.Errorf("X = %f, want 5.0", cl.X)
	}
	if cl.Y != 10.0 {
		t.Errorf("Y = %f, want 10.0", cl.Y)
	}
	if cl.DirX != 1.0 || cl.DirY != 0.0 {
		t.Errorf("Direction = (%f, %f), want (1.0, 0.0)", cl.DirX, cl.DirY)
	}
	if cl.Radius != 10.0 {
		t.Errorf("Radius = %f, want 10.0", cl.Radius)
	}
	if cl.Intensity != 0.8 {
		t.Errorf("Intensity = %f, want 0.8", cl.Intensity)
	}
	if !cl.IsActive {
		t.Error("IsActive = false, want true")
	}
	if cl.FlashlightType != "test" {
		t.Errorf("FlashlightType = %s, want test", cl.FlashlightType)
	}
}

func TestNewConeLight_DirectionNormalization(t *testing.T) {
	tests := []struct {
		name               string
		dirX, dirY         float64
		wantDirX, wantDirY float64
	}{
		{
			name:     "already normalized",
			dirX:     1.0,
			dirY:     0.0,
			wantDirX: 1.0,
			wantDirY: 0.0,
		},
		{
			name:     "needs normalization",
			dirX:     3.0,
			dirY:     4.0,
			wantDirX: 0.6,
			wantDirY: 0.8,
		},
		{
			name:     "zero vector defaults to (1,0)",
			dirX:     0.0,
			dirY:     0.0,
			wantDirX: 1.0,
			wantDirY: 0.0,
		},
	}

	preset := FlashlightPreset{Name: "test", Radius: 10, Angle: math.Pi / 6, Intensity: 1.0, R: 1, G: 1, B: 1}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := NewConeLight(0, 0, tt.dirX, tt.dirY, preset)
			if math.Abs(cl.DirX-tt.wantDirX) > 0.01 {
				t.Errorf("DirX = %f, want %f", cl.DirX, tt.wantDirX)
			}
			if math.Abs(cl.DirY-tt.wantDirY) > 0.01 {
				t.Errorf("DirY = %f, want %f", cl.DirY, tt.wantDirY)
			}
			// Verify normalized
			length := math.Sqrt(cl.DirX*cl.DirX + cl.DirY*cl.DirY)
			if math.Abs(length-1.0) > 0.01 {
				t.Errorf("Direction not normalized: length = %f", length)
			}
		})
	}
}

func TestConeLight_SetDirection(t *testing.T) {
	preset := FlashlightPreset{Name: "test", Radius: 10, Angle: math.Pi / 6, Intensity: 1.0, R: 1, G: 1, B: 1}
	cl := NewConeLight(0, 0, 1, 0, preset)

	tests := []struct {
		name               string
		dirX, dirY         float64
		wantDirX, wantDirY float64
	}{
		{
			name:     "set new direction",
			dirX:     0.0,
			dirY:     1.0,
			wantDirX: 0.0,
			wantDirY: 1.0,
		},
		{
			name:     "set diagonal",
			dirX:     1.0,
			dirY:     1.0,
			wantDirX: 1.0 / math.Sqrt(2),
			wantDirY: 1.0 / math.Sqrt(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl.SetDirection(tt.dirX, tt.dirY)
			if math.Abs(cl.DirX-tt.wantDirX) > 0.01 {
				t.Errorf("DirX = %f, want %f", cl.DirX, tt.wantDirX)
			}
			if math.Abs(cl.DirY-tt.wantDirY) > 0.01 {
				t.Errorf("DirY = %f, want %f", cl.DirY, tt.wantDirY)
			}
		})
	}
}

func TestConeLight_SetPosition(t *testing.T) {
	preset := FlashlightPreset{Name: "test", Radius: 10, Angle: math.Pi / 6, Intensity: 1.0, R: 1, G: 1, B: 1}
	cl := NewConeLight(0, 0, 1, 0, preset)

	cl.SetPosition(15.5, 25.3)
	if cl.X != 15.5 {
		t.Errorf("X = %f, want 15.5", cl.X)
	}
	if cl.Y != 25.3 {
		t.Errorf("Y = %f, want 25.3", cl.Y)
	}
}

func TestConeLight_Toggle(t *testing.T) {
	preset := FlashlightPreset{Name: "test", Radius: 10, Angle: math.Pi / 6, Intensity: 1.0, R: 1, G: 1, B: 1}
	cl := NewConeLight(0, 0, 1, 0, preset)

	if !cl.IsActive {
		t.Error("Should start active")
	}

	cl.Toggle()
	if cl.IsActive {
		t.Error("Should be inactive after toggle")
	}

	cl.Toggle()
	if !cl.IsActive {
		t.Error("Should be active after second toggle")
	}
}

func TestConeLight_SetActive(t *testing.T) {
	preset := FlashlightPreset{Name: "test", Radius: 10, Angle: math.Pi / 6, Intensity: 1.0, R: 1, G: 1, B: 1}
	cl := NewConeLight(0, 0, 1, 0, preset)

	cl.SetActive(false)
	if cl.IsActive {
		t.Error("Should be inactive")
	}

	cl.SetActive(true)
	if !cl.IsActive {
		t.Error("Should be active")
	}
}

func TestApplyConeAttenuation(t *testing.T) {
	preset := FlashlightPreset{
		Name:      "test",
		Radius:    10.0,
		Angle:     math.Pi / 4, // 45 degrees
		Intensity: 1.0,
		R:         1.0,
		G:         1.0,
		B:         1.0,
	}
	// Light at (0,0) pointing in +X direction (1,0)
	cl := NewConeLight(0, 0, 1, 0, preset)

	tests := []struct {
		name     string
		targetX  float64
		targetY  float64
		wantZero bool
		desc     string
	}{
		{
			name:     "directly ahead close",
			targetX:  2.0,
			targetY:  0.0,
			wantZero: false,
			desc:     "should have strong light",
		},
		{
			name:     "directly ahead far",
			targetX:  8.0,
			targetY:  0.0,
			wantZero: false,
			desc:     "should have some light",
		},
		{
			name:     "within cone angle",
			targetX:  5.0,
			targetY:  3.0,
			wantZero: false,
			desc:     "should have light",
		},
		{
			name:     "outside cone angle",
			targetX:  2.0,
			targetY:  5.0,
			wantZero: true,
			desc:     "should have no light",
		},
		{
			name:     "behind light",
			targetX:  -2.0,
			targetY:  0.0,
			wantZero: true,
			desc:     "should have no light",
		},
		{
			name:     "outside radius",
			targetX:  15.0,
			targetY:  0.0,
			wantZero: true,
			desc:     "should have no light",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contrib := cl.ApplyConeAttenuation(tt.targetX, tt.targetY)
			if contrib < 0.0 || contrib > 1.0 {
				t.Errorf("Contribution out of range: %f", contrib)
			}
			if tt.wantZero && contrib > 0.001 {
				t.Errorf("%s: expected ~zero, got %f", tt.desc, contrib)
			}
			if !tt.wantZero && contrib < 0.001 {
				t.Errorf("%s: expected >0, got %f", tt.desc, contrib)
			}
		})
	}
}

func TestApplyConeAttenuation_WhenInactive(t *testing.T) {
	preset := FlashlightPreset{Name: "test", Radius: 10, Angle: math.Pi / 4, Intensity: 1.0, R: 1, G: 1, B: 1}
	cl := NewConeLight(0, 0, 1, 0, preset)
	cl.SetActive(false)

	// Should return 0 when inactive
	contrib := cl.ApplyConeAttenuation(2.0, 0.0)
	if contrib != 0.0 {
		t.Errorf("Inactive light should contribute 0, got %f", contrib)
	}
}

func TestApplyConeAttenuation_CenterBrighter(t *testing.T) {
	preset := FlashlightPreset{Name: "test", Radius: 10, Angle: math.Pi / 4, Intensity: 1.0, R: 1, G: 1, B: 1}
	cl := NewConeLight(0, 0, 1, 0, preset)

	// Center of cone should be brighter than edge
	centerContrib := cl.ApplyConeAttenuation(5.0, 0.0)
	edgeContrib := cl.ApplyConeAttenuation(5.0, 3.0) // Near cone edge

	if centerContrib <= edgeContrib {
		t.Errorf("Center (%f) should be brighter than edge (%f)", centerContrib, edgeContrib)
	}
}

func TestApplyConeAttenuation_DistanceFalloff(t *testing.T) {
	preset := FlashlightPreset{Name: "test", Radius: 10, Angle: math.Pi / 4, Intensity: 1.0, R: 1, G: 1, B: 1}
	cl := NewConeLight(0, 0, 1, 0, preset)

	// Closer points should be brighter
	close := cl.ApplyConeAttenuation(2.0, 0.0)
	mid := cl.ApplyConeAttenuation(5.0, 0.0)
	far := cl.ApplyConeAttenuation(8.0, 0.0)

	if close <= mid {
		t.Errorf("Close (%f) should be brighter than mid (%f)", close, mid)
	}
	if mid <= far {
		t.Errorf("Mid (%f) should be brighter than far (%f)", mid, far)
	}
}

func TestIsPointInCone(t *testing.T) {
	preset := FlashlightPreset{Name: "test", Radius: 10, Angle: math.Pi / 4, Intensity: 1.0, R: 1, G: 1, B: 1}
	cl := NewConeLight(0, 0, 1, 0, preset)

	tests := []struct {
		name    string
		targetX float64
		targetY float64
		want    bool
	}{
		{"at light position", 0.0, 0.0, true},
		{"directly ahead", 5.0, 0.0, true},
		{"within cone", 5.0, 2.0, true},
		{"outside angle", 5.0, 5.0, false},
		{"behind", -2.0, 0.0, false},
		{"outside radius", 15.0, 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cl.IsPointInCone(tt.targetX, tt.targetY)
			if got != tt.want {
				t.Errorf("IsPointInCone(%f, %f) = %v, want %v", tt.targetX, tt.targetY, got, tt.want)
			}
		})
	}
}

func TestIsPointInCone_WhenInactive(t *testing.T) {
	preset := FlashlightPreset{Name: "test", Radius: 10, Angle: math.Pi / 4, Intensity: 1.0, R: 1, G: 1, B: 1}
	cl := NewConeLight(0, 0, 1, 0, preset)
	cl.SetActive(false)

	// Should return false when inactive
	if cl.IsPointInCone(5.0, 0.0) {
		t.Error("Inactive light should not include any points in cone")
	}
}

func TestGetContributionAsPointLight(t *testing.T) {
	preset := FlashlightPreset{
		Name:      "test",
		Radius:    10.0,
		Angle:     math.Pi / 4,
		Intensity: 1.0,
		R:         0.8,
		G:         0.9,
		B:         1.0,
	}
	cl := NewConeLight(5.0, 10.0, 1.0, 0.0, preset)

	pointLight := cl.GetContributionAsPointLight()

	if pointLight.X != 5.0 {
		t.Errorf("X = %f, want 5.0", pointLight.X)
	}
	if pointLight.Y != 10.0 {
		t.Errorf("Y = %f, want 10.0", pointLight.Y)
	}
	if pointLight.Radius != 10.0 {
		t.Errorf("Radius = %f, want 10.0", pointLight.Radius)
	}
	// Intensity should be reduced for point approximation
	expectedIntensity := 0.7
	if pointLight.Intensity != expectedIntensity {
		t.Errorf("Intensity = %f, want %f", pointLight.Intensity, expectedIntensity)
	}
	if pointLight.R != 0.8 {
		t.Errorf("R = %f, want 0.8", pointLight.R)
	}
}

func TestFantasyFlashlightPreset(t *testing.T) {
	preset := GetFlashlightPreset(genre.Fantasy)

	if preset.Name != "torch" {
		t.Errorf("Fantasy flashlight name = %s, want torch", preset.Name)
	}
	// Should have warm orange color
	if preset.R < 0.8 {
		t.Error("Fantasy torch should have high red component")
	}
	if preset.B > 0.5 {
		t.Error("Fantasy torch should have low blue component")
	}
}

func TestSciFiFlashlightPreset(t *testing.T) {
	preset := GetFlashlightPreset(genre.SciFi)

	if preset.Name != "headlamp" {
		t.Errorf("SciFi flashlight name = %s, want headlamp", preset.Name)
	}
	// Should have long reach
	if preset.Radius < 10.0 {
		t.Error("SciFi headlamp should have long reach")
	}
	// Should be bright white
	if preset.Intensity < 0.9 {
		t.Error("SciFi headlamp should be very bright")
	}
}

func TestHorrorFlashlightPreset(t *testing.T) {
	preset := GetFlashlightPreset(genre.Horror)

	if preset.Name != "flashlight" {
		t.Errorf("Horror flashlight name = %s, want flashlight", preset.Name)
	}
	// Should be dim
	if preset.Intensity > 0.7 {
		t.Error("Horror flashlight should be dim")
	}
	// Should have narrow beam
	if preset.Angle > math.Pi/6 {
		t.Error("Horror flashlight should have narrow beam")
	}
}

func TestCyberpunkFlashlightPreset(t *testing.T) {
	preset := GetFlashlightPreset(genre.Cyberpunk)

	if preset.Name != "glow_rod" {
		t.Errorf("Cyberpunk flashlight name = %s, want glow_rod", preset.Name)
	}
	// Should have cyan color
	if preset.B < 0.8 {
		t.Error("Cyberpunk glow rod should have high blue component")
	}
	if preset.G < 0.5 {
		t.Error("Cyberpunk glow rod should have decent green component")
	}
}

func TestPostApocFlashlightPreset(t *testing.T) {
	preset := GetFlashlightPreset(genre.PostApoc)

	if preset.Name != "salvaged_lamp" {
		t.Errorf("PostApoc flashlight name = %s, want salvaged_lamp", preset.Name)
	}
	// Should have wide cone (unreliable salvaged equipment)
	if preset.Angle < math.Pi/4 {
		t.Error("PostApoc salvaged lamp should have wide cone")
	}
}

// BenchmarkApplyConeAttenuation measures cone attenuation calculation performance
func BenchmarkApplyConeAttenuation(b *testing.B) {
	preset := FlashlightPreset{Name: "test", Radius: 10, Angle: math.Pi / 4, Intensity: 1.0, R: 1, G: 1, B: 1}
	cl := NewConeLight(0, 0, 1, 0, preset)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cl.ApplyConeAttenuation(5.0, 2.0)
	}
}
