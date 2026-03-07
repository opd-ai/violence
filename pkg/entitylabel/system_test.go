package entitylabel

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/ui"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem("fantasy")

	if sys.genre != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got %q", sys.genre)
	}

	if sys.font == nil {
		t.Error("Font should be initialized")
	}

	if sys.fallbackFont == nil {
		t.Error("Fallback font should be initialized")
	}

	if sys.textCache == nil {
		t.Error("Text cache should be initialized")
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetGenre("cyberpunk")

	if sys.genre != "cyberpunk" {
		t.Errorf("Expected genre 'cyberpunk' after SetGenre, got %q", sys.genre)
	}
}

func TestUpdate(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()

	// Update should not panic with empty world
	sys.Update(world)
}

func TestRenderEmpty(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()
	screen := ebiten.NewImage(800, 600)

	// Render should not panic with no entities
	sys.Render(world, screen, 0, 0)
}

func TestRenderWithLabel(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()
	screen := ebiten.NewImage(800, 600)

	// Create entity with position and label
	ent := world.AddEntity()
	world.AddComponent(ent, &engine.Position{X: 5.0, Y: 5.0})
	world.AddComponent(ent, NewEnemyLabel("Test Enemy"))

	// Render should not panic
	sys.Render(world, screen, 0, 0)
}

func TestRenderDistanceFiltering(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()
	screen := ebiten.NewImage(800, 600)

	// Create entity far away (beyond MaxDistance)
	ent := world.AddEntity()
	world.AddComponent(ent, &engine.Position{X: 50.0, Y: 50.0})
	label := NewEnemyLabel("Far Enemy")
	label.MaxDistance = 5.0
	world.AddComponent(ent, label)

	// Should not render (too far)
	sys.Render(world, screen, 0, 0)
}

func TestRenderAlphaFading(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()
	screen := ebiten.NewImage(800, 600)

	// Create entity at medium distance
	ent := world.AddEntity()
	world.AddComponent(ent, &engine.Position{X: 8.0, Y: 8.0})
	label := NewEnemyLabel("Medium Enemy")
	label.MaxDistance = 12.0
	world.AddComponent(ent, label)

	// Should render with reduced alpha
	sys.Render(world, screen, 0, 0)
}

func TestClearCache(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.textCache["test"] = ebiten.NewImage(10, 10)

	if len(sys.textCache) != 1 {
		t.Errorf("Expected 1 cached entry, got %d", len(sys.textCache))
	}

	sys.ClearCache()

	if len(sys.textCache) != 0 {
		t.Errorf("Expected 0 cached entries after clear, got %d", len(sys.textCache))
	}
}

func TestDrawPlaceholder(t *testing.T) {
	sys := NewSystem("fantasy")
	img := ebiten.NewImage(100, 100)
	col := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	// Should not panic
	sys.drawPlaceholder(img, 10, 10, col)
}

func TestRenderWithLayoutManager(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()
	screen := ebiten.NewImage(800, 600)

	// Create entity with label
	ent := world.AddEntity()
	world.AddComponent(ent, &engine.Position{X: 5.0, Y: 5.0})
	world.AddComponent(ent, NewEnemyLabel("Test Enemy"))

	// Create real layout manager (not mock)
	layoutMgr := ui.NewLayoutManager(800, 600)

	// Should not panic with layout manager
	sys.RenderWithLayout(world, screen, 0, 0, layoutMgr)
}

func TestMultipleEntitiesRender(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()
	screen := ebiten.NewImage(800, 600)

	// Create multiple entities with different label types
	ent1 := world.AddEntity()
	world.AddComponent(ent1, &engine.Position{X: 2.0, Y: 2.0})
	world.AddComponent(ent1, NewEnemyLabel("Enemy 1"))

	ent2 := world.AddEntity()
	world.AddComponent(ent2, &engine.Position{X: 3.0, Y: 3.0})
	world.AddComponent(ent2, NewNPCLabel("NPC 1"))

	ent3 := world.AddEntity()
	world.AddComponent(ent3, &engine.Position{X: 4.0, Y: 4.0})
	world.AddComponent(ent3, NewLootLabel("Sword"))

	// Should render all without panic
	sys.Render(world, screen, 0, 0)
}

func TestScaledTextRendering(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()
	screen := ebiten.NewImage(800, 600)

	// Create boss with scaled label
	ent := world.AddEntity()
	world.AddComponent(ent, &engine.Position{X: 5.0, Y: 5.0})
	world.AddComponent(ent, NewBossLabel("Boss"))

	// Should render scaled text without panic
	sys.Render(world, screen, 0, 0)
}
