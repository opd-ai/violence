package toast

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem("fantasy")
	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}
	if sys.genre != "fantasy" {
		t.Errorf("expected genre 'fantasy', got '%s'", sys.genre)
	}
	if sys.GetActiveCount() != 0 {
		t.Errorf("expected 0 active notifications, got %d", sys.GetActiveCount())
	}
	if sys.GetQueueCount() != 0 {
		t.Errorf("expected 0 queued notifications, got %d", sys.GetQueueCount())
	}
}

func TestSetGenre(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem("fantasy")
			sys.SetGenre(genre)
			if sys.genre != genre {
				t.Errorf("expected genre '%s', got '%s'", genre, sys.genre)
			}
		})
	}
}

func TestQueue(t *testing.T) {
	sys := NewSystem("fantasy")

	sys.Queue(TypeItem, "Test item pickup", PriorityNormal)

	if sys.GetQueueCount() != 1 {
		t.Errorf("expected 1 queued notification, got %d", sys.GetQueueCount())
	}
}

func TestQueueOverflow(t *testing.T) {
	sys := NewSystem("fantasy")

	// Fill queue beyond max
	for i := 0; i < maxQueueSize+5; i++ {
		sys.Queue(TypeItem, "Test notification", PriorityNormal)
	}

	// Should cap at maxQueueSize
	if sys.GetQueueCount() > maxQueueSize {
		t.Errorf("queue overflow: expected max %d, got %d", maxQueueSize, sys.GetQueueCount())
	}
}

func TestQueuePriorityDiscard(t *testing.T) {
	sys := NewSystem("fantasy")

	// Fill with low priority
	for i := 0; i < maxQueueSize; i++ {
		sys.Queue(TypeInfo, "Low priority", PriorityLow)
	}

	// Add high priority - should discard a low priority
	sys.Queue(TypeLevelUp, "Level Up!", PriorityCritical)

	// Queue should have the critical notification
	found := false
	for i := 0; i < sys.GetQueueCount(); i++ {
		if sys.queue[i].Priority == PriorityCritical {
			found = true
			break
		}
	}
	if !found {
		t.Error("critical notification not found in queue after overflow")
	}
}

func TestPromoteFromQueue(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)

	sys.Queue(TypeItem, "Test", PriorityNormal)

	if sys.GetQueueCount() != 1 {
		t.Fatalf("expected 1 queued, got %d", sys.GetQueueCount())
	}

	// Simulate update to promote
	w := engine.NewWorld()

	sys.UpdateWithDelta(w, 0.1)

	if sys.GetActiveCount() != 1 {
		t.Errorf("expected 1 active notification after update, got %d", sys.GetActiveCount())
	}
	if sys.GetQueueCount() != 0 {
		t.Errorf("expected 0 queued after promotion, got %d", sys.GetQueueCount())
	}
}

func TestNotificationLifecycle(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)
	w := engine.NewWorld()

	sys.Queue(TypeLevelUp, "Level Up!", PriorityNormal)

	// Promote and start animation

	sys.UpdateWithDelta(w, 0.1)

	if sys.GetActiveCount() != 1 {
		t.Fatalf("expected 1 active, got %d", sys.GetActiveCount())
	}

	n := sys.active[0]
	if n.State != StateEntering {
		t.Errorf("expected StateEntering, got %d", n.State)
	}

	// Advance through entering state
	for i := 0; i < 10; i++ {
		sys.UpdateWithDelta(w, 0.05)
	}

	if sys.active[0].State != StateVisible {
		t.Errorf("expected StateVisible after enter animation, got %d", sys.active[0].State)
	}

	// Advance past duration
	for i := 0; i < 100; i++ {
		sys.UpdateWithDelta(w, 0.1)
	}

	// Should have exited and been removed
	if sys.GetActiveCount() != 0 {
		t.Errorf("expected 0 active after duration, got %d", sys.GetActiveCount())
	}
}

func TestNotificationTypes(t *testing.T) {
	types := []NotificationType{
		TypeItem,
		TypeLevelUp,
		TypeAchievement,
		TypeQuest,
		TypeLoot,
		TypeSkill,
		TypeCurrency,
		TypeDeath,
		TypeWarning,
		TypeInfo,
	}

	sys := NewSystem("fantasy")

	for _, ntype := range types {
		t.Run(string(ntype), func(t *testing.T) {
			sys.Queue(ntype, "Test message", PriorityNormal)

			n := sys.queue[len(sys.queue)-1]
			if n.Type != ntype {
				t.Errorf("expected type %s, got %s", ntype, n.Type)
			}

			// Verify colors are set
			if n.IconColor.A == 0 {
				t.Error("icon color alpha is 0")
			}
			if n.TextColor.A == 0 {
				t.Error("text color alpha is 0")
			}
		})
	}
}

func TestPriorityDurations(t *testing.T) {
	sys := NewSystem("fantasy")

	tests := []struct {
		priority Priority
		minDur   float64
	}{
		{PriorityLow, 1.5},
		{PriorityNormal, 2.5},
		{PriorityHigh, 3.5},
		{PriorityCritical, 4.5},
	}

	for _, tt := range tests {
		t.Run("priority", func(t *testing.T) {
			sys.Queue(TypeInfo, "Test", tt.priority)
			n := sys.queue[len(sys.queue)-1]
			if n.Duration < tt.minDur {
				t.Errorf("priority %d duration %f less than expected %f", tt.priority, n.Duration, tt.minDur)
			}
		})
	}
}

func TestClear(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)
	w := engine.NewWorld()

	for i := 0; i < 5; i++ {
		sys.Queue(TypeItem, "Test", PriorityNormal)
	}

	sys.UpdateWithDelta(w, 0.1)

	if sys.GetActiveCount() == 0 && sys.GetQueueCount() == 0 {
		t.Fatal("expected some notifications before clear")
	}

	sys.Clear()

	if sys.GetActiveCount() != 0 {
		t.Errorf("expected 0 active after clear, got %d", sys.GetActiveCount())
	}
	if sys.GetQueueCount() != 0 {
		t.Errorf("expected 0 queued after clear, got %d", sys.GetQueueCount())
	}
}

func TestRender(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)
	w := engine.NewWorld()

	sys.Queue(TypeLevelUp, "Level Up!", PriorityCritical)

	sys.UpdateWithDelta(w, 0.5)

	// Render should not panic
	screen := ebiten.NewImage(320, 200)
	sys.Render(screen)
}

func TestPositions(t *testing.T) {
	positions := []Position{
		PositionTopRight,
		PositionTopLeft,
		PositionBottomRight,
		PositionBottomLeft,
	}

	for _, pos := range positions {
		t.Run("position", func(t *testing.T) {
			sys := NewSystem("fantasy")
			sys.SetPosition(pos)
			sys.SetScreenSize(320, 200)

			sys.Queue(TypeItem, "Test", PriorityNormal)

			w := engine.NewWorld()
			sys.UpdateWithDelta(w, 0.5)

			if sys.GetActiveCount() != 1 {
				t.Fatal("expected 1 active notification")
			}

			n := sys.active[0]

			// Verify position is within screen bounds
			if n.ScreenX < 0 || n.ScreenX > float64(sys.screenW) {
				t.Errorf("X position %f out of bounds", n.ScreenX)
			}
			if n.ScreenY < 0 || n.ScreenY > float64(sys.screenH) {
				t.Errorf("Y position %f out of bounds", n.ScreenY)
			}
		})
	}
}

func TestGenreColors(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre)

			colors := sys.getColors(TypeLevelUp, PriorityCritical)

			// Verify colors are valid
			if colors.bg.A == 0 {
				t.Error("background alpha is 0")
			}
			if colors.text.A == 0 {
				t.Error("text alpha is 0")
			}
		})
	}
}

func TestComponent(t *testing.T) {
	comp := NewComponent()
	if comp == nil {
		t.Fatal("NewComponent returned nil")
	}
	if !comp.Enabled {
		t.Error("component should be enabled by default")
	}
	if comp.Type() != "toast" {
		t.Errorf("expected type 'toast', got '%s'", comp.Type())
	}
}

func TestNotificationMethods(t *testing.T) {
	n := &Notification{
		ID:        1,
		Type:      TypeLevelUp,
		Message:   "Test",
		Priority:  PriorityNormal,
		Duration:  3.0,
		State:     StateVisible,
		Alpha:     1.0,
		Scale:     1.0,
		IconColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		TextColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}

	if !n.IsActive() {
		t.Error("notification should be active")
	}
	if !n.IsVisible() {
		t.Error("notification should be visible")
	}
	if n.GetAlpha() != 255 {
		t.Errorf("expected alpha 255, got %d", n.GetAlpha())
	}

	// Test removed state
	n.State = StateRemoved
	if n.IsActive() {
		t.Error("removed notification should not be active")
	}
	if n.IsVisible() {
		t.Error("removed notification should not be visible")
	}
}

func TestIconGeneration(t *testing.T) {
	sys := NewSystem("fantasy")

	// All icon types should be generated
	types := []NotificationType{
		TypeItem,
		TypeLevelUp,
		TypeAchievement,
		TypeQuest,
		TypeLoot,
		TypeSkill,
		TypeCurrency,
		TypeDeath,
		TypeWarning,
		TypeInfo,
	}

	for _, ntype := range types {
		t.Run(string(ntype), func(t *testing.T) {
			img := sys.iconImages[ntype]
			if img == nil {
				t.Errorf("icon image for %s is nil", ntype)
			}
		})
	}
}

func TestMaxVisible(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)
	w := engine.NewWorld()

	// Queue more than max visible
	for i := 0; i < maxVisible+5; i++ {
		sys.Queue(TypeItem, "Test", PriorityNormal)
	}

	sys.UpdateWithDelta(w, 0.1)

	// Should only promote up to maxVisible
	if sys.GetActiveCount() > maxVisible {
		t.Errorf("active count %d exceeds max visible %d", sys.GetActiveCount(), maxVisible)
	}
}

func TestEasingFunction(t *testing.T) {
	// Test easing function properties
	tests := []struct {
		input    float64
		expected float64
	}{
		{0.0, 0.0},
		{1.0, 1.0},
	}

	for _, tt := range tests {
		result := easeOutCubic(tt.input)
		if result != tt.expected {
			t.Errorf("easeOutCubic(%f) = %f, expected %f", tt.input, result, tt.expected)
		}
	}

	// Verify monotonically increasing
	prev := 0.0
	for i := 0; i <= 10; i++ {
		x := float64(i) / 10.0
		y := easeOutCubic(x)
		if y < prev {
			t.Errorf("easeOutCubic not monotonically increasing at %f", x)
		}
		prev = y
	}
}

func TestHighPriorityPromotedFirst(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)

	// Queue low priority first
	sys.Queue(TypeInfo, "Low priority", PriorityLow)
	sys.Queue(TypeLevelUp, "Critical!", PriorityCritical)
	sys.Queue(TypeItem, "Normal", PriorityNormal)

	w := engine.NewWorld()

	sys.UpdateWithDelta(w, 0.1)

	// Critical should be promoted first
	if sys.GetActiveCount() < 1 {
		t.Fatal("expected at least 1 active notification")
	}

	first := sys.active[0]
	if first.Priority != PriorityCritical {
		t.Errorf("expected critical priority first, got %d", first.Priority)
	}
}

func BenchmarkQueue(b *testing.B) {
	sys := NewSystem("fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Queue(TypeItem, "Test notification", PriorityNormal)
		if sys.GetQueueCount() > maxQueueSize/2 {
			sys.Clear()
		}
	}
}

func BenchmarkUpdate(b *testing.B) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)
	w := engine.NewWorld()

	for i := 0; i < maxVisible; i++ {
		sys.Queue(TypeItem, "Test", PriorityNormal)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.UpdateWithDelta(w, 1.0/60.0)
		// Re-queue to keep active
		if sys.GetActiveCount() < maxVisible/2 {
			sys.Queue(TypeItem, "Test", PriorityNormal)
		}
	}
}

func BenchmarkRender(b *testing.B) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)
	w := engine.NewWorld()

	for i := 0; i < maxVisible; i++ {
		sys.Queue(TypeItem, "Test notification message", PriorityNormal)
	}

	sys.UpdateWithDelta(w, 0.5)

	screen := ebiten.NewImage(320, 200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Render(screen)
	}
}
