package healthbar

import (
	"image/color"
	"reflect"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
)

func TestComponent_Type(t *testing.T) {
	comp := &Component{}
	if comp.Type() != "healthbar" {
		t.Errorf("expected type 'healthbar', got '%s'", comp.Type())
	}
}

func TestStatusIconsComponent_Type(t *testing.T) {
	comp := &StatusIconsComponent{}
	if comp.Type() != "statusicons" {
		t.Errorf("expected type 'statusicons', got '%s'", comp.Type())
	}
}

func TestNewSystem(t *testing.T) {
	sys := NewSystem("fantasy")
	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}
	if sys.genre != "fantasy" {
		t.Errorf("expected genre 'fantasy', got '%s'", sys.genre)
	}
	if sys.fadeDelay != 3.0 {
		t.Errorf("expected fadeDelay 3.0, got %f", sys.fadeDelay)
	}
}

func TestSystem_SetGenre(t *testing.T) {
	sys := NewSystem("fantasy")
	originalColor := sys.baseColor
	
	sys.SetGenre("cyberpunk")
	if sys.genre != "cyberpunk" {
		t.Errorf("expected genre 'cyberpunk', got '%s'", sys.genre)
	}
	
	if sys.baseColor == originalColor {
		t.Error("expected base color to change after genre change")
	}
}

func TestSystem_GenreThemes(t *testing.T) {
	genres := []string{"fantasy", "cyberpunk", "horror", "scifi", "postapoc"}
	
	for _, genre := range genres {
		sys := NewSystem(genre)
		if sys.baseColor.R == 0 && sys.baseColor.G == 0 && sys.baseColor.B == 0 {
			t.Errorf("genre %s has invalid base color", genre)
		}
	}
}

func TestSystem_Update(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem("fantasy")
	
	eid := world.AddEntity()
	world.AddComponent(eid, &engine.Health{Current: 100, Max: 100})
	world.AddComponent(eid, &Component{
		Visible:       true,
		Width:         40,
		Height:        4,
		OffsetY:       20,
		LastDamageAge: 0,
	})
	
	sys.Update(world)
	
	barType := reflect.TypeOf(&Component{})
	comp, ok := world.GetComponent(eid, barType)
	if !ok {
		t.Fatal("component not found after update")
	}
	
	bar := comp.(*Component)
	if bar.LastDamageAge == 0 {
		t.Error("LastDamageAge should have been updated")
	}
}

func TestSystem_UpdateStatusIconDurations(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem("fantasy")
	
	eid := world.AddEntity()
	world.AddComponent(eid, &engine.Health{Current: 50, Max: 100})
	
	statusIcons := &StatusIconsComponent{}
	statusIcons.AddIcon(IconPoison, 5.0, 1)
	statusIcons.AddIcon(IconBurn, 0.01, 1)
	world.AddComponent(eid, statusIcons)
	
	if len(statusIcons.Icons) != 2 {
		t.Fatalf("expected 2 icons, got %d", len(statusIcons.Icons))
	}
	
	for i := 0; i < 120; i++ {
		sys.Update(world)
	}
	
	statusType := reflect.TypeOf(&StatusIconsComponent{})
	comp, ok := world.GetComponent(eid, statusType)
	if !ok {
		t.Fatal("status icons component not found")
	}
	
	icons := comp.(*StatusIconsComponent)
	if len(icons.Icons) != 1 {
		t.Errorf("expected 1 icon after expiration, got %d", len(icons.Icons))
	}
	
	if len(icons.Icons) > 0 && icons.Icons[0].Type != IconPoison {
		t.Error("expected poison icon to remain")
	}
}

func TestStatusIconsComponent_AddIcon(t *testing.T) {
	comp := &StatusIconsComponent{}
	
	comp.AddIcon(IconPoison, 5.0, 1)
	if len(comp.Icons) != 1 {
		t.Fatalf("expected 1 icon, got %d", len(comp.Icons))
	}
	
	if comp.Icons[0].Type != IconPoison {
		t.Error("expected poison icon")
	}
	
	comp.AddIcon(IconPoison, 10.0, 2)
	if len(comp.Icons) != 1 {
		t.Errorf("expected icon to be updated, not added; got %d icons", len(comp.Icons))
	}
	
	if comp.Icons[0].Duration != 10.0 {
		t.Errorf("expected duration 10.0, got %f", comp.Icons[0].Duration)
	}
	
	if comp.Icons[0].Stacks != 2 {
		t.Errorf("expected stacks 2, got %d", comp.Icons[0].Stacks)
	}
}

func TestStatusIconsComponent_RemoveIcon(t *testing.T) {
	comp := &StatusIconsComponent{}
	
	comp.AddIcon(IconPoison, 5.0, 1)
	comp.AddIcon(IconBurn, 3.0, 1)
	
	if len(comp.Icons) != 2 {
		t.Fatalf("expected 2 icons, got %d", len(comp.Icons))
	}
	
	comp.RemoveIcon(IconPoison)
	
	if len(comp.Icons) != 1 {
		t.Errorf("expected 1 icon after removal, got %d", len(comp.Icons))
	}
	
	if len(comp.Icons) > 0 && comp.Icons[0].Type != IconBurn {
		t.Error("expected burn icon to remain")
	}
}

func TestStatusIconsComponent_UpdateDurations(t *testing.T) {
	comp := &StatusIconsComponent{}
	
	comp.AddIcon(IconPoison, 1.0, 1)
	comp.AddIcon(IconBurn, 2.0, 1)
	
	comp.UpdateDurations(0.5)
	
	if len(comp.Icons) != 2 {
		t.Errorf("expected 2 icons, got %d", len(comp.Icons))
	}
	
	comp.UpdateDurations(0.6)
	
	if len(comp.Icons) != 1 {
		t.Errorf("expected 1 icon after expiration, got %d", len(comp.Icons))
	}
}

func TestStatusIconsComponent_Clear(t *testing.T) {
	comp := &StatusIconsComponent{}
	
	comp.AddIcon(IconPoison, 5.0, 1)
	comp.AddIcon(IconBurn, 3.0, 1)
	
	comp.Clear()
	
	if len(comp.Icons) != 0 {
		t.Errorf("expected 0 icons after clear, got %d", len(comp.Icons))
	}
}

func TestSystem_GetHealthColor(t *testing.T) {
	sys := NewSystem("fantasy")
	bar := &Component{}
	
	tests := []struct {
		name      string
		healthPct float64
		expected  color.RGBA
	}{
		{"full health", 1.0, sys.baseColor},
		{"high health", 0.6, sys.baseColor},
		{"medium health", 0.4, sys.damageColor},
		{"low health", 0.2, sys.criticalColor},
		{"critical health", 0.1, sys.criticalColor},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sys.getHealthColor(tt.healthPct, bar)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSystem_GetHealthColor_Custom(t *testing.T) {
	sys := NewSystem("fantasy")
	customColor := color.RGBA{123, 45, 67, 255}
	bar := &Component{
		CustomColor: &customColor,
	}
	
	result := sys.getHealthColor(0.5, bar)
	if result != customColor {
		t.Errorf("expected custom color %v, got %v", customColor, result)
	}
}

func TestSystem_WorldToScreen(t *testing.T) {
	sys := NewSystem("fantasy")
	
	tests := []struct {
		name        string
		worldX      float64
		worldY      float64
		camX        float64
		camY        float64
		camDirX     float64
		camDirY     float64
		expectValid bool
	}{
		{"in front", 5.0, 5.0, 0.0, 0.0, 1.0, 0.0, true},
		{"behind camera", -5.0, 0.0, 0.0, 0.0, 1.0, 0.0, false},
		{"too far", 25.0, 0.0, 0.0, 0.0, 1.0, 0.0, false},
		{"too close", 0.2, 0.0, 0.0, 0.0, 1.0, 0.0, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, visible := sys.worldToScreen(tt.worldX, tt.worldY, tt.camX, tt.camY, tt.camDirX, tt.camDirY, 800, 600)
			if visible != tt.expectValid {
				t.Errorf("expected visible=%v, got %v", tt.expectValid, visible)
			}
		})
	}
}

func TestSystem_GetIconColor(t *testing.T) {
	sys := NewSystem("fantasy")
	
	iconTypes := []StatusIconType{
		IconPoison, IconBurn, IconFreeze, IconStun, IconBleed,
		IconRegen, IconShield, IconHaste, IconSlow, IconWeak,
		IconBerserk, IconInvisible,
	}
	
	for _, iconType := range iconTypes {
		color := sys.getIconColor(iconType)
		if color.R == 0 && color.G == 0 && color.B == 0 {
			t.Errorf("icon type %d has invalid color", iconType)
		}
	}
}

func TestSystem_GetStatusIconImage(t *testing.T) {
	sys := NewSystem("fantasy")
	
	img := sys.getStatusIconImage(IconPoison)
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	
	cached := sys.getStatusIconImage(IconPoison)
	if cached != img {
		t.Error("expected cached image to be returned")
	}
	
	if len(sys.statusIconCache) == 0 {
		t.Error("expected icon to be cached")
	}
}

func TestSystem_DrawIconShapes(t *testing.T) {
	sys := NewSystem("fantasy")
	
	iconTypes := []StatusIconType{
		IconPoison, IconBurn, IconFreeze, IconStun, IconBleed,
		IconRegen, IconShield, IconHaste, IconSlow, IconWeak,
		IconBerserk, IconInvisible,
	}
	
	for _, iconType := range iconTypes {
		img := ebiten.NewImage(12, 12)
		sys.drawIconShape(img, iconType, color.RGBA{255, 0, 0, 255}, 12)
	}
}

func TestSystem_RenderHealthBars(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem("fantasy")
	
	eid := world.AddEntity()
	world.AddComponent(eid, &engine.Health{Current: 50, Max: 100})
	world.AddComponent(eid, &engine.Position{X: 5.0, Y: 5.0})
	world.AddComponent(eid, &Component{
		Visible:       true,
		Width:         40,
		Height:        4,
		OffsetY:       20,
		ShowWhenFull:  false,
		LastDamageAge: 0,
	})
	
	screen := ebiten.NewImage(800, 600)
	sys.RenderHealthBars(screen, world, 0.0, 0.0, 1.0, 0.0, 800, 600)
}

func TestSystem_RenderHealthBars_WithIcons(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem("fantasy")
	
	eid := world.AddEntity()
	world.AddComponent(eid, &engine.Health{Current: 75, Max: 100})
	world.AddComponent(eid, &engine.Position{X: 5.0, Y: 5.0})
	world.AddComponent(eid, &Component{
		Visible:       true,
		Width:         40,
		Height:        4,
		OffsetY:       20,
		ShowWhenFull:  true,
		LastDamageAge: 0,
	})
	
	statusIcons := &StatusIconsComponent{}
	statusIcons.AddIcon(IconPoison, 5.0, 1)
	statusIcons.AddIcon(IconShield, 10.0, 2)
	world.AddComponent(eid, statusIcons)
	
	screen := ebiten.NewImage(800, 600)
	sys.RenderHealthBars(screen, world, 0.0, 0.0, 1.0, 0.0, 800, 600)
}

func TestSystem_RenderHealthBars_HiddenWhenFull(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem("fantasy")
	
	eid := world.AddEntity()
	world.AddComponent(eid, &engine.Health{Current: 100, Max: 100})
	world.AddComponent(eid, &engine.Position{X: 5.0, Y: 5.0})
	world.AddComponent(eid, &Component{
		Visible:       true,
		Width:         40,
		Height:        4,
		OffsetY:       20,
		ShowWhenFull:  false,
		LastDamageAge: 10.0,
	})
	
	screen := ebiten.NewImage(800, 600)
	sys.RenderHealthBars(screen, world, 0.0, 0.0, 1.0, 0.0, 800, 600)
}

func TestSystem_RenderHealthBars_FadeOut(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem("fantasy")
	
	eid := world.AddEntity()
	world.AddComponent(eid, &engine.Health{Current: 80, Max: 100})
	world.AddComponent(eid, &engine.Position{X: 5.0, Y: 5.0})
	world.AddComponent(eid, &Component{
		Visible:       true,
		Width:         40,
		Height:        4,
		OffsetY:       20,
		ShowWhenFull:  false,
		LastDamageAge: 4.0,
	})
	
	screen := ebiten.NewImage(800, 600)
	sys.RenderHealthBars(screen, world, 0.0, 0.0, 1.0, 0.0, 800, 600)
}

func TestComponent_DefaultValues(t *testing.T) {
	comp := &Component{
		Visible:  true,
		Width:    50,
		Height:   5,
		OffsetY:  25,
	}
	
	if comp.ShowWhenFull {
		t.Error("expected ShowWhenFull to default to false")
	}
	
	if comp.ThreatLevel != 0 {
		t.Errorf("expected ThreatLevel 0, got %d", comp.ThreatLevel)
	}
	
	if comp.LastDamageAge != 0 {
		t.Errorf("expected LastDamageAge 0, got %f", comp.LastDamageAge)
	}
	
	if comp.CustomColor != nil {
		t.Error("expected CustomColor to be nil")
	}
}

func TestStatusIcon_Properties(t *testing.T) {
	icon := StatusIcon{
		Type:     IconPoison,
		Duration: 5.0,
		Stacks:   3,
		Color:    color.RGBA{100, 200, 50, 255},
	}
	
	if icon.Type != IconPoison {
		t.Error("icon type mismatch")
	}
	
	if icon.Duration != 5.0 {
		t.Errorf("expected duration 5.0, got %f", icon.Duration)
	}
	
	if icon.Stacks != 3 {
		t.Errorf("expected stacks 3, got %d", icon.Stacks)
	}
}
