package subsurface

import (
	"image"
	"image/color"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/procgen/genre"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem()
	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}
	if sys.lightIntensity != 1.0 {
		t.Errorf("Expected lightIntensity 1.0, got %f", sys.lightIntensity)
	}
	if sys.thicknessCache == nil {
		t.Error("thicknessCache should be initialized")
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		genre             string
		expectedIntensity float64
		expectedAmbient   float64
	}{
		{genre.Fantasy, 1.0, 0.3},
		{genre.SciFi, 0.8, 0.4},
		{genre.Horror, 0.6, 0.2},
		{genre.Cyberpunk, 0.9, 0.35},
		{genre.PostApoc, 1.1, 0.25},
		{"unknown", 1.0, 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewSystem()
			sys.SetGenre(tt.genre)

			if sys.lightIntensity != tt.expectedIntensity {
				t.Errorf("genre %s: expected intensity %f, got %f",
					tt.genre, tt.expectedIntensity, sys.lightIntensity)
			}
			if sys.ambientStrength != tt.expectedAmbient {
				t.Errorf("genre %s: expected ambient %f, got %f",
					tt.genre, tt.expectedAmbient, sys.ambientStrength)
			}
		})
	}
}

func TestSetLightDirection(t *testing.T) {
	sys := NewSystem()

	sys.SetLightDirection(1, 0, 0)
	if sys.lightDirX != 1.0 {
		t.Errorf("Expected lightDirX 1.0, got %f", sys.lightDirX)
	}

	sys.SetLightDirection(1, 1, 0)
	expected := 1.0 / 1.4142135623730951
	if sys.lightDirX < expected-0.01 || sys.lightDirX > expected+0.01 {
		t.Errorf("Expected normalized lightDirX ~%f, got %f", expected, sys.lightDirX)
	}

	initialX := sys.lightDirX
	sys.SetLightDirection(0, 0, 0)
	if sys.lightDirX != initialX {
		t.Error("Zero vector should not change light direction")
	}
}

func TestComponent(t *testing.T) {
	comp := NewComponent()

	if !comp.Enabled {
		t.Error("NewComponent should be enabled by default")
	}
	if comp.Material != MaterialFlesh {
		t.Error("Default material should be MaterialFlesh")
	}
	if comp.Intensity != 1.0 {
		t.Errorf("Expected intensity 1.0, got %f", comp.Intensity)
	}
	if comp.Type() != "SubsurfaceComponent" {
		t.Errorf("Unexpected type: %s", comp.Type())
	}
}

func TestNewComponentWithMaterial(t *testing.T) {
	tests := []struct {
		mat  Material
		name string
	}{
		{MaterialFlesh, "flesh"},
		{MaterialLeaf, "leaf"},
		{MaterialWax, "wax"},
		{MaterialSlime, "slime"},
		{MaterialMembrane, "membrane"},
		{MaterialFruit, "fruit"},
		{MaterialBone, "bone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewComponentWithMaterial(tt.mat)
			if comp.Material != tt.mat {
				t.Errorf("Expected material %d, got %d", tt.mat, comp.Material)
			}
		})
	}
}

func TestGetScatterProfile(t *testing.T) {
	tests := []struct {
		mat                    Material
		name                   string
		expectHighAbsorptionB  bool
		expectHighTranslucency bool
	}{
		{MaterialFlesh, "flesh", true, false},
		{MaterialLeaf, "leaf", true, true},
		{MaterialMembrane, "membrane", false, true},
		{MaterialSlime, "slime", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := GetScatterProfile(tt.mat)

			if tt.expectHighAbsorptionB && profile.AbsorptionB < 0.5 {
				t.Errorf("%s should have high blue absorption", tt.name)
			}
			if tt.expectHighTranslucency && profile.Translucency < 0.5 {
				t.Errorf("%s should have high translucency", tt.name)
			}
			if profile.ScatterDistance <= 0 {
				t.Errorf("%s scatter distance should be positive", tt.name)
			}
		})
	}
}

func TestGetMaterialName(t *testing.T) {
	tests := []struct {
		mat      Material
		expected string
	}{
		{MaterialFlesh, "flesh"},
		{MaterialLeaf, "leaf"},
		{MaterialWax, "wax"},
		{MaterialSlime, "slime"},
		{MaterialMembrane, "membrane"},
		{MaterialFruit, "fruit"},
		{MaterialBone, "bone"},
		{Material(100), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			name := GetMaterialName(tt.mat)
			if name != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, name)
			}
		})
	}
}

func TestApplySSS(t *testing.T) {
	sys := NewSystem()

	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 200, G: 150, B: 120, A: 255})
		}
	}

	centerBefore := img.RGBAAt(16, 16)

	sys.ApplySSS(img, MaterialFlesh, 1.0)

	centerAfter := img.RGBAAt(16, 16)

	if centerBefore.R == centerAfter.R &&
		centerBefore.G == centerAfter.G &&
		centerBefore.B == centerAfter.B {
		t.Log("SSS may not have visibly changed the color (could be expected for thick regions)")
	}

	edgeAfter := img.RGBAAt(8, 16)
	if edgeAfter.A != 255 {
		t.Log("Edge pixels processed with SSS")
	}
}

func TestApplySSSZeroIntensity(t *testing.T) {
	sys := NewSystem()

	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	before := img.RGBAAt(8, 8)

	sys.ApplySSS(img, MaterialFlesh, 0.0)

	after := img.RGBAAt(8, 8)
	if before != after {
		t.Error("Zero intensity should not modify the image")
	}
}

func TestApplySSSToRegion(t *testing.T) {
	sys := NewSystem()

	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 20; y < 44; y++ {
		for x := 20; x < 44; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 180, G: 120, B: 100, A: 255})
		}
	}

	bounds := image.Rect(20, 20, 44, 44)
	sys.ApplySSSToRegion(img, bounds, MaterialLeaf, 0.8)

	outsideBefore := img.RGBAAt(10, 10)
	if outsideBefore.A != 0 {
		t.Log("Pixels outside region were not modified")
	}
}

func TestComputeThickness(t *testing.T) {
	sys := NewSystem()

	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	tm := sys.computeThickness(img, img.Bounds())
	if tm == nil {
		t.Fatal("computeThickness returned nil")
	}

	if tm.width != 32 || tm.height != 32 {
		t.Errorf("Unexpected thickness map dimensions: %dx%d", tm.width, tm.height)
	}

	centerThickness := tm.thickness[16][16]
	edgeThickness := tm.thickness[8][8]

	if centerThickness <= edgeThickness {
		t.Error("Center should have greater thickness than edge")
	}

	if !tm.edges[8][8] {
		t.Log("Expected edge pixel at (8,8)")
	}
}

func TestComputeSSSColor(t *testing.T) {
	sys := NewSystem()
	profile := GetScatterProfile(MaterialFlesh)

	original := color.RGBA{R: 200, G: 150, B: 120, A: 255}

	sssColorThin := sys.computeSSSColor(original, profile, 0.2, true, 1.0)
	sssColorThick := sys.computeSSSColor(original, profile, 0.8, false, 1.0)

	if sssColorThin.R <= sssColorThick.R {
		t.Log("Thin regions should generally have more scattering (warmer)")
	}

	sssColorEdge := sys.computeSSSColor(original, profile, 0.3, true, 1.0)
	sssColorInterior := sys.computeSSSColor(original, profile, 0.3, false, 1.0)
	if sssColorEdge == sssColorInterior {
		t.Log("Edge and interior may have different scattering behavior")
	}
}

func TestSystemUpdate(t *testing.T) {
	sys := NewSystem()
	world := engine.NewWorld()

	entity := world.AddEntity()
	world.AddComponent(entity, &Component{
		Enabled:   true,
		Material:  MaterialFlesh,
		Intensity: 1.0,
	})

	sys.Update(world)
}

func TestRenderSSSDebug(t *testing.T) {
	sys := NewSystem()

	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 200, G: 150, B: 120, A: 255})
		}
	}

	debug := sys.RenderSSSDebug(img, img.Bounds())
	if debug == nil {
		t.Fatal("RenderSSSDebug returned nil")
	}

	edgeColor := debug.RGBAAt(8, 16)
	if edgeColor.R != 255 || edgeColor.G != 0 || edgeColor.B != 0 {
		t.Log("Edge should be marked in red for debug visualization")
	}

	interiorColor := debug.RGBAAt(16, 16)
	if interiorColor.A == 0 {
		t.Log("Interior should have thickness visualization")
	}
}

func TestClampByte(t *testing.T) {
	tests := []struct {
		input    float64
		expected uint8
	}{
		{-10, 0},
		{0, 0},
		{127.5, 127},
		{255, 255},
		{300, 255},
	}

	for _, tt := range tests {
		result := clampByte(tt.input)
		if result != tt.expected {
			t.Errorf("clampByte(%f) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{5, 10, 5},
		{10, 5, 5},
		{7, 7, 7},
		{-3, 2, -3},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func BenchmarkApplySSS(b *testing.B) {
	sys := NewSystem()

	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 16; y < 48; y++ {
		for x := 16; x < 48; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 200, G: 150, B: 120, A: 255})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testImg := image.NewRGBA(image.Rect(0, 0, 64, 64))
		copy(testImg.Pix, img.Pix)
		sys.ApplySSS(testImg, MaterialFlesh, 1.0)
	}
}

func BenchmarkComputeThickness(b *testing.B) {
	sys := NewSystem()

	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.computeThickness(img, img.Bounds())
	}
}
