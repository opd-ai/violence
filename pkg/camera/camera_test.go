package camera

import (
	"math"
	"testing"
)

func TestNewCamera(t *testing.T) {
	tests := []struct {
		name string
		fov  float64
	}{
		{"90 degree FOV", 90.0},
		{"60 degree FOV", 60.0},
		{"120 degree FOV", 120.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCamera(tt.fov)
			if c.FOV != tt.fov {
				t.Errorf("FOV = %v, want %v", c.FOV, tt.fov)
			}
			if c.DirX != 1.0 {
				t.Errorf("DirX = %v, want 1.0", c.DirX)
			}
			if c.DirY != 0.0 {
				t.Errorf("DirY = %v, want 0.0", c.DirY)
			}
		})
	}
}

func TestCamera_Update_Position(t *testing.T) {
	tests := []struct {
		name   string
		initX  float64
		initY  float64
		deltaX float64
		deltaY float64
		wantX  float64
		wantY  float64
	}{
		{"Forward movement", 5.0, 5.0, 1.0, 0.0, 6.0, 5.0},
		{"Backward movement", 5.0, 5.0, -1.0, 0.0, 4.0, 5.0},
		{"Strafe right", 5.0, 5.0, 0.0, 1.0, 5.0, 6.0},
		{"Strafe left", 5.0, 5.0, 0.0, -1.0, 5.0, 4.0},
		{"Diagonal movement", 5.0, 5.0, 0.5, 0.5, 5.5, 5.5},
		{"No movement", 5.0, 5.0, 0.0, 0.0, 5.0, 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCamera(90.0)
			c.X = tt.initX
			c.Y = tt.initY
			c.Update(tt.deltaX, tt.deltaY, 0, 0, 0)

			if math.Abs(c.X-tt.wantX) > 0.0001 {
				t.Errorf("X = %v, want %v", c.X, tt.wantX)
			}
			if math.Abs(c.Y-tt.wantY) > 0.0001 {
				t.Errorf("Y = %v, want %v", c.Y, tt.wantY)
			}
		})
	}
}

func TestCamera_Update_Direction(t *testing.T) {
	tests := []struct {
		name      string
		initDirX  float64
		initDirY  float64
		deltaDirX float64
		deltaDirY float64
		wantDirX  float64
		wantDirY  float64
	}{
		{"No rotation", 1.0, 0.0, 0.0, 0.0, 1.0, 0.0},
		{"Small change", 1.0, 0.0, 0.1, 0.0, 0.995, 0.0},       // Normalized
		{"Direction change", 1.0, 0.0, 0.0, 1.0, 0.707, 0.707}, // Normalized to (1,1) -> (0.707,0.707)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCamera(90.0)
			c.DirX = tt.initDirX
			c.DirY = tt.initDirY
			c.Update(0, 0, tt.deltaDirX, tt.deltaDirY, 0)

			if math.Abs(c.DirX-tt.wantDirX) > 0.01 {
				t.Errorf("DirX = %v, want %v", c.DirX, tt.wantDirX)
			}
			if math.Abs(c.DirY-tt.wantDirY) > 0.01 {
				t.Errorf("DirY = %v, want %v", c.DirY, tt.wantDirY)
			}

			dirLen := math.Sqrt(c.DirX*c.DirX + c.DirY*c.DirY)
			if math.Abs(dirLen-1.0) > 0.0001 {
				t.Errorf("Direction not normalized: length = %v", dirLen)
			}
		})
	}
}

func TestCamera_Update_PitchClamping(t *testing.T) {
	tests := []struct {
		name       string
		initPitch  float64
		deltaPitch float64
		wantPitch  float64
	}{
		{"Within bounds", 0.0, 10.0, 10.0},
		{"Clamp upper", 25.0, 10.0, MaxPitch},
		{"Clamp lower", -25.0, -10.0, MinPitch},
		{"At max", MaxPitch, 5.0, MaxPitch},
		{"At min", MinPitch, -5.0, MinPitch},
		{"Negative to positive", -10.0, 20.0, 10.0},
		{"Large positive delta", 0.0, 100.0, MaxPitch},
		{"Large negative delta", 0.0, -100.0, MinPitch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCamera(90.0)
			c.Pitch = tt.initPitch
			c.Update(0, 0, 0, 0, tt.deltaPitch)

			if math.Abs(c.Pitch-tt.wantPitch) > 0.0001 {
				t.Errorf("Pitch = %v, want %v", c.Pitch, tt.wantPitch)
			}
		})
	}
}

func TestCamera_Update_HeadBob(t *testing.T) {
	tests := []struct {
		name           string
		deltaX         float64
		deltaY         float64
		wantBobNonZero bool
	}{
		{"No movement no bob", 0.0, 0.0, false},
		{"Forward movement has bob", 0.1, 0.0, true},
		{"Backward movement has bob", -0.1, 0.0, true},
		{"Strafe movement has bob", 0.0, 0.1, true},
		{"Diagonal movement has bob", 0.1, 0.1, true},
		{"Tiny movement no bob", 0.0001, 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCamera(90.0)
			c.Update(tt.deltaX, tt.deltaY, 0, 0, 0)

			hasNonZeroBob := math.Abs(c.HeadBob) > 0.001
			if hasNonZeroBob != tt.wantBobNonZero {
				t.Errorf("HeadBob = %v, wantNonZero = %v", c.HeadBob, tt.wantBobNonZero)
			}
		})
	}
}

func TestCamera_Update_HeadBobOscillation(t *testing.T) {
	c := NewCamera(90.0)

	var bobValues []float64
	for i := 0; i < 100; i++ {
		c.Update(0.1, 0.0, 0, 0, 0)
		bobValues = append(bobValues, c.HeadBob)
	}

	foundPositive := false
	foundNegative := false
	for _, v := range bobValues {
		if v > 0.001 {
			foundPositive = true
		}
		if v < -0.001 {
			foundNegative = true
		}
	}

	if !foundPositive {
		t.Error("HeadBob never positive during movement")
	}
	if !foundNegative {
		t.Error("HeadBob never negative during movement")
	}

	for _, v := range bobValues {
		if math.Abs(v) > HeadBobAmplitude+0.001 {
			t.Errorf("HeadBob %v exceeds amplitude %v", v, HeadBobAmplitude)
		}
	}
}

func TestCamera_Update_HeadBobReset(t *testing.T) {
	c := NewCamera(90.0)

	c.Update(0.1, 0.0, 0, 0, 0)
	if math.Abs(c.HeadBob) < 0.001 {
		t.Error("Expected non-zero head bob after movement")
	}

	c.Update(0.0, 0.0, 0, 0, 0)
	if math.Abs(c.HeadBob) > 0.0001 {
		t.Errorf("HeadBob = %v, want 0 after stopping", c.HeadBob)
	}
	if math.Abs(c.headBobPhase) > 0.0001 {
		t.Errorf("headBobPhase = %v, want 0 after stopping", c.headBobPhase)
	}
}

func TestCamera_Rotate(t *testing.T) {
	tests := []struct {
		name         string
		initDirX     float64
		initDirY     float64
		angleRadians float64
		wantDirX     float64
		wantDirY     float64
	}{
		{"90 degrees right", 1.0, 0.0, math.Pi / 2, 0.0, 1.0},
		{"90 degrees left", 1.0, 0.0, -math.Pi / 2, 0.0, -1.0},
		{"180 degrees", 1.0, 0.0, math.Pi, -1.0, 0.0},
		{"360 degrees", 1.0, 0.0, 2 * math.Pi, 1.0, 0.0},
		{"45 degrees", 1.0, 0.0, math.Pi / 4, 0.707, 0.707},
		{"No rotation", 1.0, 0.0, 0.0, 1.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCamera(90.0)
			c.DirX = tt.initDirX
			c.DirY = tt.initDirY
			c.Rotate(tt.angleRadians)

			if math.Abs(c.DirX-tt.wantDirX) > 0.01 {
				t.Errorf("DirX = %v, want %v", c.DirX, tt.wantDirX)
			}
			if math.Abs(c.DirY-tt.wantDirY) > 0.01 {
				t.Errorf("DirY = %v, want %v", c.DirY, tt.wantDirY)
			}
		})
	}
}

func TestCamera_FOVAspectRatio(t *testing.T) {
	fovs := []float64{60.0, 75.0, 90.0, 105.0, 120.0}

	for _, fov := range fovs {
		c := NewCamera(fov)
		if c.FOV != fov {
			t.Errorf("FOV = %v, want %v", c.FOV, fov)
		}
	}
}

func BenchmarkCamera_Update(b *testing.B) {
	c := NewCamera(90.0)
	for i := 0; i < b.N; i++ {
		c.Update(0.1, 0.05, 0.01, 0.01, 0.5)
	}
}

func BenchmarkCamera_Rotate(b *testing.B) {
	c := NewCamera(90.0)
	angle := 0.01
	for i := 0; i < b.N; i++ {
		c.Rotate(angle)
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name     string
		genreID  string
		expected string
	}{
		{"fantasy", "fantasy", "fantasy"},
		{"scifi", "scifi", "scifi"},
		{"horror", "horror", "horror"},
		{"cyberpunk", "cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetGenre(tt.genreID)
			if got := GetCurrentGenre(); got != tt.expected {
				t.Errorf("GetCurrentGenre() = %v, want %v", got, tt.expected)
			}
		})
	}
}
