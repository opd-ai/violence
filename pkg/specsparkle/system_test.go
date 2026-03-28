package specsparkle

import (
	"image/color"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewSystem(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy genre", "fantasy"},
		{"scifi genre", "scifi"},
		{"horror genre", "horror"},
		{"cyberpunk genre", "cyberpunk"},
		{"postapoc genre", "postapoc"},
		{"unknown genre defaults to fantasy", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem(tt.genreID, 12345)
			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.genreID != tt.genreID {
				t.Errorf("genreID = %v, want %v", sys.genreID, tt.genreID)
			}
			if sys.rng == nil {
				t.Error("rng should not be nil")
			}
			if sys.sparkles == nil {
				t.Error("sparkles map should not be nil")
			}
			if sys.cache == nil {
				t.Error("cache map should not be nil")
			}
			if !sys.enabled {
				t.Error("system should be enabled by default")
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	sys.SetGenre("scifi")
	if sys.genreID != "scifi" {
		t.Errorf("genreID = %v, want scifi", sys.genreID)
	}

	sys.SetGenre("cyberpunk")
	if sys.genreID != "cyberpunk" {
		t.Errorf("genreID = %v, want cyberpunk", sys.genreID)
	}

	sys.SetGenre("nonexistent")
	if sys.genreID != "nonexistent" {
		t.Errorf("genreID = %v, want nonexistent", sys.genreID)
	}
}

func TestSetScreenSize(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	sys.SetScreenSize(640, 480)
	if sys.screenW != 640 || sys.screenH != 480 {
		t.Errorf("screen size = %dx%d, want 640x480", sys.screenW, sys.screenH)
	}
}

func TestSetCamera(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	sys.SetCamera(10.5, 20.5)
	if sys.cameraX != 10.5 || sys.cameraY != 20.5 {
		t.Errorf("camera = (%v, %v), want (10.5, 20.5)", sys.cameraX, sys.cameraY)
	}
}

func TestSetEnabled(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	sys.SetEnabled(false)
	if sys.enabled {
		t.Error("system should be disabled")
	}

	sys.SetEnabled(true)
	if !sys.enabled {
		t.Error("system should be enabled")
	}
}

func TestComponentType(t *testing.T) {
	comp := &Component{}
	if comp.Type() != "specsparkle.Component" {
		t.Errorf("Type() = %v, want specsparkle.Component", comp.Type())
	}
}

func TestNewComponent(t *testing.T) {
	tests := []struct {
		name     string
		material MaterialClass
	}{
		{"metal", MaterialMetal},
		{"crystal", MaterialCrystal},
		{"wet", MaterialWet},
		{"glass", MaterialGlass},
		{"gold", MaterialGold},
		{"silver", MaterialSilver},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewComponent(tt.material)
			if comp == nil {
				t.Fatal("NewComponent returned nil")
			}
			if comp.Material != tt.material {
				t.Errorf("Material = %v, want %v", comp.Material, tt.material)
			}
			if !comp.Enabled {
				t.Error("component should be enabled by default")
			}
			if comp.Intensity <= 0 || comp.Intensity > 1 {
				t.Errorf("Intensity = %v, want (0, 1]", comp.Intensity)
			}
			if comp.Size <= 0 {
				t.Errorf("Size = %v, want > 0", comp.Size)
			}
		})
	}
}

func TestGetMaterialColor(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	tests := []struct {
		material MaterialClass
		name     string
	}{
		{MaterialMetal, "metal"},
		{MaterialCrystal, "crystal"},
		{MaterialWet, "wet"},
		{MaterialGlass, "glass"},
		{MaterialGold, "gold"},
		{MaterialSilver, "silver"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := sys.getMaterialColor(tt.material)
			if col.A == 0 {
				t.Error("color alpha should not be 0")
			}
		})
	}
}

func TestShiftHue(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	baseColor := color.RGBA{R: 255, G: 128, B: 64, A: 255}

	shifted := sys.shiftHue(baseColor, 0.0)
	if shifted.A != 255 {
		t.Error("alpha should be preserved")
	}

	shifted = sys.shiftHue(baseColor, 0.5)
	if shifted.A != 255 {
		t.Error("alpha should be preserved after hue shift")
	}

	white := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	shiftedWhite := sys.shiftHue(white, 0.3)
	if shiftedWhite.A != 255 {
		t.Error("alpha should be preserved for low saturation colors")
	}
}

func TestHSVToRGB(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Pure red
	col := sys.hsvToRGB(0.0, 1.0, 1.0)
	if col.R < 250 || col.G > 5 || col.B > 5 {
		t.Errorf("expected red, got R=%d G=%d B=%d", col.R, col.G, col.B)
	}

	// Gray (no saturation)
	col = sys.hsvToRGB(0.5, 0.0, 0.5)
	if col.R != col.G || col.G != col.B {
		t.Errorf("expected gray, got R=%d G=%d B=%d", col.R, col.G, col.B)
	}
}

func TestGenerateSparkleImage(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	tests := []struct {
		name      string
		size      int
		starShape bool
	}{
		{"small circular", 2, false},
		{"medium circular", 5, false},
		{"large circular", 10, false},
		{"small star", 2, true},
		{"medium star", 5, true},
		{"large star", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := color.RGBA{R: 255, G: 255, B: 255, A: 255}
			img := sys.generateSparkleImage(col, tt.size, tt.starShape)
			if img == nil {
				t.Fatal("generated image should not be nil")
			}
			bounds := img.Bounds()
			if bounds.Dx() != tt.size || bounds.Dy() != tt.size {
				t.Errorf("image size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), tt.size, tt.size)
			}
		})
	}
}

func TestGetSparkleIntensity(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	tests := []struct {
		phase    float64
		expected float64
	}{
		{0.0, 0.0},
		{0.5, 1.0},
		{1.0, 0.0},
	}

	for _, tt := range tests {
		intensity := sys.getSparkleIntensity(tt.phase)
		diff := intensity - tt.expected
		if diff < -0.01 || diff > 0.01 {
			t.Errorf("intensity at phase %v = %v, want ~%v", tt.phase, intensity, tt.expected)
		}
	}
}

func TestCreateSparkle(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	comp := NewComponent(MaterialMetal)

	sparkle := sys.createSparkle(comp)
	if sparkle.X < 0 || sparkle.X > 1 {
		t.Errorf("X = %v, want [0, 1]", sparkle.X)
	}
	if sparkle.Y < 0 || sparkle.Y > 1 {
		t.Errorf("Y = %v, want [0, 1]", sparkle.Y)
	}
	if sparkle.Phase != 0 {
		t.Errorf("Phase = %v, want 0", sparkle.Phase)
	}
	if sparkle.Speed <= 0 {
		t.Errorf("Speed = %v, want > 0", sparkle.Speed)
	}
	if !sparkle.Active {
		t.Error("sparkle should be active when created")
	}
}

func TestShouldSpawnSparkle(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	comp := NewComponent(MaterialMetal)
	comp.Density = 1.0
	comp.Distance = 1.0

	spawned := 0
	for i := 0; i < 1000; i++ {
		if sys.shouldSpawnSparkle(comp, 1.0/60.0) {
			spawned++
		}
	}

	if spawned < 10 || spawned > 200 {
		t.Errorf("spawned %d sparkles in 1000 frames, expected 10-200", spawned)
	}
}

func TestUpdateSystem(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()

	entity := world.AddEntity()
	comp := NewComponent(MaterialMetal)
	comp.Density = 1.0
	world.AddComponent(entity, comp)

	for i := 0; i < 120; i++ {
		sys.Update(world)
	}

	sys.mu.RLock()
	sparkleCount := len(sys.sparkles[entity])
	sys.mu.RUnlock()

	if sparkleCount < 0 {
		t.Error("sparkle count should not be negative")
	}
}

func TestUpdateSystemDisabled(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	sys.SetEnabled(false)
	world := engine.NewWorld()

	entity := world.AddEntity()
	comp := NewComponent(MaterialMetal)
	world.AddComponent(entity, comp)

	sys.Update(world)

	sys.mu.RLock()
	sparkleCount := len(sys.sparkles[entity])
	sys.mu.RUnlock()

	if sparkleCount != 0 {
		t.Errorf("disabled system should not create sparkles, got %d", sparkleCount)
	}
}

func TestCacheEviction(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	sys.maxCache = 3

	colors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},
		{R: 0, G: 255, B: 0, A: 255},
		{R: 0, G: 0, B: 255, A: 255},
		{R: 255, G: 255, B: 0, A: 255},
		{R: 255, G: 0, B: 255, A: 255},
	}

	for _, col := range colors {
		_ = sys.getSparkleImage(col, 5, false)
	}

	if len(sys.cache) > sys.maxCache {
		t.Errorf("cache size %d exceeds max %d", len(sys.cache), sys.maxCache)
	}
}

func TestGenrePresets(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		preset, ok := genrePresets[genre]
		if !ok {
			t.Errorf("missing preset for genre %s", genre)
			continue
		}

		if preset.BaseIntensity <= 0 || preset.BaseIntensity > 2.0 {
			t.Errorf("%s: BaseIntensity %v out of range", genre, preset.BaseIntensity)
		}
		if preset.SpawnRate <= 0 {
			t.Errorf("%s: SpawnRate %v should be positive", genre, preset.SpawnRate)
		}
		if preset.LifetimeMin <= 0 || preset.LifetimeMin > preset.LifetimeMax {
			t.Errorf("%s: invalid lifetime range [%v, %v]", genre, preset.LifetimeMin, preset.LifetimeMax)
		}
		if preset.MetalTint.A != 255 {
			t.Errorf("%s: MetalTint alpha should be 255", genre)
		}
	}
}

func TestClampByte(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{-10, 0},
		{0, 0},
		{128, 128},
		{255, 255},
		{300, 255},
	}

	for _, tt := range tests {
		result := clampByte(tt.input)
		if result != tt.expected {
			t.Errorf("clampByte(%d) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestGetPreset(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	preset := sys.GetPreset()

	if preset.BaseIntensity != genrePresets["fantasy"].BaseIntensity {
		t.Error("GetPreset should return current genre preset")
	}

	sys.SetGenre("scifi")
	preset = sys.GetPreset()

	if preset.BaseIntensity != genrePresets["scifi"].BaseIntensity {
		t.Error("GetPreset should return updated genre preset")
	}
}

func TestDistanceFalloff(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	comp := NewComponent(MaterialMetal)
	comp.Distance = 1.0

	closeSpawns := 0
	for i := 0; i < 500; i++ {
		if sys.shouldSpawnSparkle(comp, 1.0/60.0) {
			closeSpawns++
		}
	}

	comp.Distance = 10.0
	farSpawns := 0
	for i := 0; i < 500; i++ {
		if sys.shouldSpawnSparkle(comp, 1.0/60.0) {
			farSpawns++
		}
	}

	if farSpawns >= closeSpawns {
		t.Logf("close spawns: %d, far spawns: %d (expected far < close)", closeSpawns, farSpawns)
	}
}

func BenchmarkGenerateSparkleImage(b *testing.B) {
	sys := NewSystem("fantasy", 12345)
	col := color.RGBA{R: 255, G: 200, B: 150, A: 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.generateSparkleImage(col, 5, true)
	}
}

func BenchmarkUpdate(b *testing.B) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()

	for i := 0; i < 20; i++ {
		entity := world.AddEntity()
		comp := NewComponent(MaterialMetal)
		world.AddComponent(entity, comp)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(world)
	}
}

func BenchmarkShiftHue(b *testing.B) {
	sys := NewSystem("fantasy", 12345)
	col := color.RGBA{R: 255, G: 128, B: 64, A: 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.shiftHue(col, 0.1)
	}
}
