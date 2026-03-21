package statusbar

import (
	"image/color"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/status"
)

func TestNewComponent(t *testing.T) {
	c := NewComponent()

	if c == nil {
		t.Fatal("NewComponent returned nil")
	}

	if c.IconSize != 16 {
		t.Errorf("expected IconSize 16, got %d", c.IconSize)
	}

	if c.MaxIcons != 8 {
		t.Errorf("expected MaxIcons 8, got %d", c.MaxIcons)
	}

	if !c.Visible {
		t.Error("expected Visible to be true by default")
	}

	if len(c.Icons) != 0 {
		t.Errorf("expected empty Icons slice, got %d", len(c.Icons))
	}
}

func TestComponent_SetPosition(t *testing.T) {
	c := NewComponent()
	c.SetPosition(100.5, 200.5)

	if c.X != 100.5 {
		t.Errorf("expected X 100.5, got %f", c.X)
	}
	if c.Y != 200.5 {
		t.Errorf("expected Y 200.5, got %f", c.Y)
	}
}

func TestComponent_AddIcon(t *testing.T) {
	c := NewComponent()

	icon := IconState{
		EffectName:        "burning",
		DisplayName:       "Burning",
		IconType:          IconDamage,
		Color:             color.RGBA{255, 100, 50, 255},
		DurationRemaining: 5 * time.Second,
		TotalDuration:     10 * time.Second,
		StackCount:        1,
	}

	c.AddIcon(icon)

	if c.GetIconCount() != 1 {
		t.Errorf("expected 1 icon, got %d", c.GetIconCount())
	}

	if !c.HasEffect("burning") {
		t.Error("expected HasEffect to return true for 'burning'")
	}

	if c.HasEffect("poisoned") {
		t.Error("expected HasEffect to return false for 'poisoned'")
	}
}

func TestComponent_ClearIcons(t *testing.T) {
	c := NewComponent()

	c.AddIcon(IconState{EffectName: "test1"})
	c.AddIcon(IconState{EffectName: "test2"})

	if c.GetIconCount() != 2 {
		t.Errorf("expected 2 icons before clear, got %d", c.GetIconCount())
	}

	c.ClearIcons()

	if c.GetIconCount() != 0 {
		t.Errorf("expected 0 icons after clear, got %d", c.GetIconCount())
	}
}

func TestComponent_MaxIcons(t *testing.T) {
	c := NewComponent()
	c.MaxIcons = 3

	for i := 0; i < 5; i++ {
		c.AddIcon(IconState{EffectName: string(rune('a' + i))})
	}

	if c.GetIconCount() != 3 {
		t.Errorf("expected MaxIcons to limit to 3, got %d", c.GetIconCount())
	}
}

func TestNewSystem(t *testing.T) {
	s := NewSystem("fantasy")

	if s == nil {
		t.Fatal("NewSystem returned nil")
	}

	if s.genreID != "fantasy" {
		t.Errorf("expected genreID 'fantasy', got '%s'", s.genreID)
	}
}

func TestSystem_SetGenre(t *testing.T) {
	tests := []struct {
		genre      string
		expectBuff color.RGBA
	}{
		{"fantasy", color.RGBA{255, 215, 0, 255}},   // Gold
		{"scifi", color.RGBA{0, 255, 255, 255}},     // Cyan
		{"horror", color.RGBA{144, 238, 144, 255}},  // Sickly green
		{"cyberpunk", color.RGBA{255, 0, 255, 255}}, // Neon pink
		{"postapoc", color.RGBA{210, 105, 30, 255}}, // Rust orange
		{"unknown", color.RGBA{255, 215, 0, 255}},   // Default to fantasy
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			s := NewSystem(tt.genre)

			if s.buffColor != tt.expectBuff {
				t.Errorf("expected buffColor %v, got %v", tt.expectBuff, s.buffColor)
			}
		})
	}
}

func TestSystem_EffectTypeToIconType(t *testing.T) {
	s := NewSystem("fantasy")

	tests := []struct {
		effectName string
		expected   IconType
	}{
		{"burning", IconDamage},
		{"bleeding", IconDamage},
		{"poisoned", IconDamage},
		{"irradiated", IconDamage},
		{"regeneration", IconHeal},
		{"nanoheal", IconHeal},
		{"blessed", IconBuff},
		{"overcharged", IconBuff},
		{"cursed", IconDebuff},
		{"terrified", IconDebuff},
		{"stunned", IconStun},
		{"emp_stunned", IconStun},
		{"slowed", IconSlow},
		{"unknown_effect", IconDebuff}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.effectName, func(t *testing.T) {
			result := s.effectTypeToIconType(tt.effectName)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSystem_Uint32ToColor(t *testing.T) {
	s := NewSystem("fantasy")

	// Test RGBA packed as uint32: 0xRRGGBBAA
	input := uint32(0xFF8040C0) // R=255, G=128, B=64, A=192
	expected := color.RGBA{R: 255, G: 128, B: 64, A: 192}

	result := s.uint32ToColor(input)

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestFormatEffectName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"burning", "Burning"},
		{"bleeding", "Bleeding"},
		{"poisoned", "Poisoned"},
		{"emp_stunned", "EMP"},
		{"nanoheal", "Nanoheal"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatEffectName(tt.input)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSystem_Update_NoStatusComponent(t *testing.T) {
	s := NewSystem("fantasy")
	world := engine.NewWorld()

	entity := world.AddEntity()
	bar := NewComponent()
	bar.AddIcon(IconState{EffectName: "test"})
	world.AddComponent(entity, bar)

	s.Update(world)

	// Should clear icons when no status component
	if bar.GetIconCount() != 0 {
		t.Errorf("expected icons to be cleared, got %d", bar.GetIconCount())
	}
}

func TestSystem_Update_WithStatusComponent(t *testing.T) {
	s := NewSystem("fantasy")
	world := engine.NewWorld()

	entity := world.AddEntity()
	bar := NewComponent()
	world.AddComponent(entity, bar)

	// Add status component with active effects
	statusComp := &status.StatusComponent{
		ActiveEffects: []status.ActiveEffect{
			{
				EffectName:    "burning",
				TimeRemaining: 5 * time.Second,
				TickInterval:  time.Second,
				VisualColor:   0xFF880088,
			},
			{
				EffectName:    "slowed",
				TimeRemaining: 3 * time.Second,
				TickInterval:  time.Second,
				VisualColor:   0x0088FFAA,
			},
		},
	}
	world.AddComponent(entity, statusComp)

	s.Update(world)

	if bar.GetIconCount() != 2 {
		t.Errorf("expected 2 icons, got %d", bar.GetIconCount())
	}

	if !bar.HasEffect("burning") {
		t.Error("expected to have 'burning' effect")
	}

	if !bar.HasEffect("slowed") {
		t.Error("expected to have 'slowed' effect")
	}
}

func TestSystem_Update_StackingEffects(t *testing.T) {
	s := NewSystem("fantasy")
	world := engine.NewWorld()

	entity := world.AddEntity()
	bar := NewComponent()
	world.AddComponent(entity, bar)

	// Add status component with multiple stacks of same effect
	statusComp := &status.StatusComponent{
		ActiveEffects: []status.ActiveEffect{
			{
				EffectName:    "bleeding",
				TimeRemaining: 5 * time.Second,
				TickInterval:  time.Second,
				VisualColor:   0xAA000088,
			},
			{
				EffectName:    "bleeding",
				TimeRemaining: 3 * time.Second,
				TickInterval:  time.Second,
				VisualColor:   0xAA000088,
			},
			{
				EffectName:    "bleeding",
				TimeRemaining: 7 * time.Second, // Longest duration
				TickInterval:  time.Second,
				VisualColor:   0xAA000088,
			},
		},
	}
	world.AddComponent(entity, statusComp)

	s.Update(world)

	// Should only show 1 icon with stack count 3
	if bar.GetIconCount() != 1 {
		t.Errorf("expected 1 stacked icon, got %d", bar.GetIconCount())
	}

	if bar.Icons[0].StackCount != 3 {
		t.Errorf("expected stack count 3, got %d", bar.Icons[0].StackCount)
	}

	// Should use longest remaining duration
	if bar.Icons[0].DurationRemaining != 7*time.Second {
		t.Errorf("expected duration 7s, got %v", bar.Icons[0].DurationRemaining)
	}
}

func TestSystem_Render_NoComponent(t *testing.T) {
	s := NewSystem("fantasy")
	world := engine.NewWorld()
	entity := world.AddEntity()

	// Should not panic with missing component
	screen := ebiten.NewImage(320, 200)
	s.Render(screen, world, entity)
}

func TestSystem_Render_EmptyIcons(t *testing.T) {
	s := NewSystem("fantasy")
	world := engine.NewWorld()

	entity := world.AddEntity()
	bar := NewComponent()
	world.AddComponent(entity, bar)

	// Should not render anything with empty icons
	screen := ebiten.NewImage(320, 200)
	s.Render(screen, world, entity)
}

func TestSystem_Render_WithIcons(t *testing.T) {
	s := NewSystem("fantasy")
	world := engine.NewWorld()

	entity := world.AddEntity()
	bar := NewComponent()
	bar.AddIcon(IconState{
		EffectName:        "burning",
		IconType:          IconDamage,
		Color:             color.RGBA{255, 100, 50, 255},
		DurationRemaining: 5 * time.Second,
		TotalDuration:     10 * time.Second,
		StackCount:        1,
	})
	world.AddComponent(entity, bar)

	screen := ebiten.NewImage(320, 200)
	s.Render(screen, world, entity)

	// Visual verification would require image comparison
	// Here we just verify no panic
}

func TestSystem_RenderDirect(t *testing.T) {
	s := NewSystem("fantasy")

	bar := NewComponent()
	bar.AddIcon(IconState{
		EffectName:        "poisoned",
		IconType:          IconDamage,
		Color:             color.RGBA{100, 255, 100, 255},
		DurationRemaining: 8 * time.Second,
		TotalDuration:     10 * time.Second,
		StackCount:        2,
	})

	screen := ebiten.NewImage(320, 200)
	s.RenderDirect(screen, bar)

	// Verify no panic
}

func TestSystem_RenderDirect_NilBar(t *testing.T) {
	s := NewSystem("fantasy")
	screen := ebiten.NewImage(320, 200)

	// Should not panic with nil bar
	s.RenderDirect(screen, nil)
}

func TestSystem_RenderDirect_HiddenBar(t *testing.T) {
	s := NewSystem("fantasy")

	bar := NewComponent()
	bar.Visible = false
	bar.AddIcon(IconState{EffectName: "test"})

	screen := ebiten.NewImage(320, 200)
	s.RenderDirect(screen, bar)

	// Should not render when hidden
}

func TestSystem_GetBounds(t *testing.T) {
	s := NewSystem("fantasy")

	bar := NewComponent()
	bar.SetPosition(10, 20)
	bar.IconSize = 16
	bar.IconSpacing = 2

	bar.AddIcon(IconState{EffectName: "test1"})
	bar.AddIcon(IconState{EffectName: "test2"})
	bar.AddIcon(IconState{EffectName: "test3"})

	bounds := s.GetBounds(bar)

	// 3 icons * (16 + 2) - 2 = 52
	expectedWidth := 3*(16+2) - 2

	if bounds.Min.X != 10 {
		t.Errorf("expected Min.X 10, got %d", bounds.Min.X)
	}
	if bounds.Min.Y != 20 {
		t.Errorf("expected Min.Y 20, got %d", bounds.Min.Y)
	}
	if bounds.Dx() != expectedWidth {
		t.Errorf("expected width %d, got %d", expectedWidth, bounds.Dx())
	}
	if bounds.Dy() != 16 {
		t.Errorf("expected height 16, got %d", bounds.Dy())
	}
}

func TestSystem_GetBounds_Empty(t *testing.T) {
	s := NewSystem("fantasy")
	bar := NewComponent()

	bounds := s.GetBounds(bar)

	if !bounds.Empty() {
		t.Error("expected empty bounds for bar with no icons")
	}
}

func TestSystem_GetBounds_Nil(t *testing.T) {
	s := NewSystem("fantasy")

	bounds := s.GetBounds(nil)

	if !bounds.Empty() {
		t.Error("expected empty bounds for nil bar")
	}
}

func TestSystem_IconCache(t *testing.T) {
	s := NewSystem("fantasy")

	c := color.RGBA{255, 100, 50, 255}

	// First call should generate
	img1 := s.GetIcon(IconDamage, c)
	if img1 == nil {
		t.Fatal("GetIcon returned nil")
	}

	// Second call should return cached
	img2 := s.GetIcon(IconDamage, c)
	if img2 != img1 {
		t.Error("expected cached image to be returned")
	}
}

func TestSystem_IconCacheEviction(t *testing.T) {
	s := NewSystem("fantasy")
	s.maxCacheSize = 3

	// Fill cache
	for i := 0; i < 5; i++ {
		c := color.RGBA{uint8(i * 50), 100, 50, 255}
		s.GetIcon(IconDamage, c)
	}

	s.iconCacheMu.RLock()
	cacheSize := len(s.iconCache)
	s.iconCacheMu.RUnlock()

	if cacheSize > 3 {
		t.Errorf("expected cache size <= 3, got %d", cacheSize)
	}
}

func TestSystem_ExpiringEffectDetection(t *testing.T) {
	s := NewSystem("fantasy")
	world := engine.NewWorld()

	entity := world.AddEntity()
	bar := NewComponent()
	world.AddComponent(entity, bar)

	statusComp := &status.StatusComponent{
		ActiveEffects: []status.ActiveEffect{
			{
				EffectName:    "burning",
				TimeRemaining: 2 * time.Second, // Less than 3s threshold
				TickInterval:  time.Second,
				VisualColor:   0xFF880088,
			},
			{
				EffectName:    "slowed",
				TimeRemaining: 5 * time.Second, // More than 3s
				TickInterval:  time.Second,
				VisualColor:   0x0088FFAA,
			},
		},
	}
	world.AddComponent(entity, statusComp)

	s.Update(world)

	// Check expiring flag
	var burningIcon, slowedIcon *IconState
	for i := range bar.Icons {
		if bar.Icons[i].EffectName == "burning" {
			burningIcon = &bar.Icons[i]
		}
		if bar.Icons[i].EffectName == "slowed" {
			slowedIcon = &bar.Icons[i]
		}
	}

	if burningIcon == nil {
		t.Fatal("burning icon not found")
	}
	if slowedIcon == nil {
		t.Fatal("slowed icon not found")
	}

	if !burningIcon.IsExpiring {
		t.Error("burning should be marked as expiring")
	}
	if slowedIcon.IsExpiring {
		t.Error("slowed should not be marked as expiring")
	}
}

func TestComponent_Type(t *testing.T) {
	c := NewComponent()
	if c.Type() != "statusbar" {
		t.Errorf("expected type 'statusbar', got '%s'", c.Type())
	}
}

func BenchmarkSystem_Update(b *testing.B) {
	s := NewSystem("fantasy")
	world := engine.NewWorld()

	entity := world.AddEntity()
	bar := NewComponent()
	world.AddComponent(entity, bar)

	statusComp := &status.StatusComponent{
		ActiveEffects: []status.ActiveEffect{
			{EffectName: "burning", TimeRemaining: 5 * time.Second, VisualColor: 0xFF880088},
			{EffectName: "slowed", TimeRemaining: 3 * time.Second, VisualColor: 0x0088FFAA},
			{EffectName: "poisoned", TimeRemaining: 8 * time.Second, VisualColor: 0x00FF0088},
		},
	}
	world.AddComponent(entity, statusComp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Update(world)
	}
}

func BenchmarkSystem_Render(b *testing.B) {
	s := NewSystem("fantasy")
	world := engine.NewWorld()

	entity := world.AddEntity()
	bar := NewComponent()
	bar.AddIcon(IconState{EffectName: "burning", IconType: IconDamage, Color: color.RGBA{255, 100, 50, 255}})
	bar.AddIcon(IconState{EffectName: "slowed", IconType: IconSlow, Color: color.RGBA{100, 100, 255, 255}})
	bar.AddIcon(IconState{EffectName: "poisoned", IconType: IconDamage, Color: color.RGBA{100, 255, 100, 255}})
	world.AddComponent(entity, bar)

	screen := ebiten.NewImage(320, 200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Render(screen, world, entity)
	}
}

func BenchmarkSystem_RenderDirect(b *testing.B) {
	s := NewSystem("fantasy")

	bar := NewComponent()
	bar.AddIcon(IconState{
		EffectName:        "burning",
		IconType:          IconDamage,
		Color:             color.RGBA{255, 100, 50, 255},
		DurationRemaining: 5 * time.Second,
		TotalDuration:     10 * time.Second,
	})
	bar.AddIcon(IconState{
		EffectName:        "slowed",
		IconType:          IconSlow,
		Color:             color.RGBA{100, 100, 255, 255},
		DurationRemaining: 3 * time.Second,
		TotalDuration:     5 * time.Second,
	})

	screen := ebiten.NewImage(320, 200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.RenderDirect(screen, bar)
	}
}
