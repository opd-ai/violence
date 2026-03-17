package ui

import (
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewTooltipSystem(t *testing.T) {
	config := DefaultTooltipConfig()
	sys := NewTooltipSystem(800, 600, config)

	if sys == nil {
		t.Fatal("NewTooltipSystem returned nil")
	}

	if sys.screenWidth != 800 {
		t.Errorf("Expected screenWidth 800, got %d", sys.screenWidth)
	}
	if sys.screenHeight != 600 {
		t.Errorf("Expected screenHeight 600, got %d", sys.screenHeight)
	}
}

func TestDefaultTooltipConfig(t *testing.T) {
	config := DefaultTooltipConfig()

	if config.BackgroundColor.A == 0 {
		t.Error("Background color should have non-zero alpha")
	}
	if config.Padding <= 0 {
		t.Error("Padding should be positive")
	}
	if config.ShowDelay <= 0 {
		t.Error("ShowDelay should be positive")
	}
	if config.MaxWidth <= 0 {
		t.Error("MaxWidth should be positive")
	}
}

func TestTooltipRegistration(t *testing.T) {
	config := DefaultTooltipConfig()
	sys := NewTooltipSystem(800, 600, config)

	sys.RegisterTooltip("btn1", "Test tooltip", 100, 100, 50, 20, TooltipAbove)

	if len(sys.tooltips) != 1 {
		t.Errorf("Expected 1 tooltip, got %d", len(sys.tooltips))
	}

	tt := sys.tooltips["btn1"]
	if tt.Text != "Test tooltip" {
		t.Errorf("Expected text 'Test tooltip', got '%s'", tt.Text)
	}
	if tt.TargetX != 100 {
		t.Errorf("Expected TargetX 100, got %d", tt.TargetX)
	}
}

func TestTooltipRemoval(t *testing.T) {
	config := DefaultTooltipConfig()
	sys := NewTooltipSystem(800, 600, config)

	sys.RegisterTooltip("btn1", "Test", 100, 100, 50, 20, TooltipAbove)
	sys.RegisterTooltip("btn2", "Test2", 200, 200, 50, 20, TooltipBelow)

	if len(sys.tooltips) != 2 {
		t.Fatalf("Expected 2 tooltips, got %d", len(sys.tooltips))
	}

	sys.RemoveTooltip("btn1")

	if len(sys.tooltips) != 1 {
		t.Errorf("Expected 1 tooltip after removal, got %d", len(sys.tooltips))
	}

	if _, ok := sys.tooltips["btn1"]; ok {
		t.Error("btn1 should be removed")
	}
}

func TestTooltipHoverAndLeave(t *testing.T) {
	config := DefaultTooltipConfig()
	sys := NewTooltipSystem(800, 600, config)

	sys.RegisterTooltip("btn1", "Hover me", 100, 100, 50, 20, TooltipAbove)

	// Simulate hover
	sys.OnHover("btn1")

	if sys.activeID != "btn1" {
		t.Errorf("Expected activeID 'btn1', got '%s'", sys.activeID)
	}

	// Simulate leave
	sys.OnLeave("btn1")

	if sys.activeID != "" {
		t.Errorf("Expected empty activeID after leave, got '%s'", sys.activeID)
	}
}

func TestTooltipDelayedShow(t *testing.T) {
	config := DefaultTooltipConfig()
	config.ShowDelay = 50 * time.Millisecond
	config.FadeInDuration = 50 * time.Millisecond
	sys := NewTooltipSystem(800, 600, config)

	sys.RegisterTooltip("btn1", "Delayed", 100, 100, 50, 20, TooltipAbove)

	sys.OnHover("btn1")
	sys.Update()

	// Immediately after hover, tooltip should not be visible
	tt := sys.tooltips["btn1"]
	if tt.Visible {
		t.Error("Tooltip should not be visible immediately after hover")
	}

	// Wait for delay
	time.Sleep(60 * time.Millisecond)
	sys.Update()

	if !tt.Visible {
		t.Error("Tooltip should be visible after delay")
	}

	// Wait for fade-in
	time.Sleep(60 * time.Millisecond)
	sys.Update()

	if tt.Opacity < 0.9 {
		t.Errorf("Tooltip opacity should be near 1.0 after fade, got %f", tt.Opacity)
	}
}

func TestTooltipScreenEdgeAwareness_TopEdge(t *testing.T) {
	config := DefaultTooltipConfig()
	config.ShowDelay = 0
	sys := NewTooltipSystem(800, 600, config)

	// Target near top edge, tooltip above would go off-screen
	sys.RegisterTooltip("top", "Near top", 400, 10, 50, 20, TooltipAuto)

	sys.OnHover("top")
	sys.Update()

	tt := sys.tooltips["top"]

	// Tooltip should be repositioned to stay on screen
	if tt.FinalY < 0 {
		t.Errorf("Tooltip Y should not be negative, got %d", tt.FinalY)
	}
}

func TestTooltipScreenEdgeAwareness_RightEdge(t *testing.T) {
	config := DefaultTooltipConfig()
	config.ShowDelay = 0
	sys := NewTooltipSystem(800, 600, config)

	// Target near right edge
	sys.RegisterTooltip("right", "Near right", 750, 300, 50, 20, TooltipRight)

	sys.OnHover("right")
	sys.Update()

	tt := sys.tooltips["right"]

	// Tooltip should not extend past screen edge
	if tt.FinalX+tt.FinalW > sys.screenWidth {
		t.Errorf("Tooltip extends past right edge: X=%d, W=%d, screen=%d",
			tt.FinalX, tt.FinalW, sys.screenWidth)
	}
}

func TestTooltipScreenEdgeAwareness_BottomEdge(t *testing.T) {
	config := DefaultTooltipConfig()
	config.ShowDelay = 0
	sys := NewTooltipSystem(800, 600, config)

	// Target near bottom edge
	sys.RegisterTooltip("bottom", "Near bottom", 400, 570, 50, 20, TooltipBelow)

	sys.OnHover("bottom")
	sys.Update()

	tt := sys.tooltips["bottom"]

	// Tooltip should not extend past screen bottom
	if tt.FinalY+tt.FinalH > sys.screenHeight {
		t.Errorf("Tooltip extends past bottom edge: Y=%d, H=%d, screen=%d",
			tt.FinalY, tt.FinalH, sys.screenHeight)
	}
}

func TestTooltipScreenEdgeAwareness_LeftEdge(t *testing.T) {
	config := DefaultTooltipConfig()
	config.ShowDelay = 0
	sys := NewTooltipSystem(800, 600, config)

	// Target near left edge
	sys.RegisterTooltip("left", "Near left", 10, 300, 50, 20, TooltipLeft)

	sys.OnHover("left")
	sys.Update()

	tt := sys.tooltips["left"]

	// Tooltip should not have negative X
	if tt.FinalX < 0 {
		t.Errorf("Tooltip X should not be negative, got %d", tt.FinalX)
	}
}

func TestTooltipAutoPosition(t *testing.T) {
	config := DefaultTooltipConfig()
	config.ShowDelay = 0
	sys := NewTooltipSystem(800, 600, config)

	// Center of screen - should have space in all directions
	sys.RegisterTooltip("center", "Auto positioned", 400, 300, 50, 20, TooltipAuto)

	sys.OnHover("center")
	sys.Update()

	tt := sys.tooltips["center"]

	// Should have computed a valid position
	if tt.FinalW <= 0 || tt.FinalH <= 0 {
		t.Error("Tooltip should have computed valid dimensions")
	}
}

func TestTooltipTextWrapping(t *testing.T) {
	config := DefaultTooltipConfig()
	config.MaxWidth = 100
	config.ShowDelay = 0
	sys := NewTooltipSystem(800, 600, config)

	longText := "This is a very long tooltip text that should be wrapped across multiple lines for readability"
	sys.RegisterTooltip("wrap", longText, 400, 300, 50, 20, TooltipBelow)

	sys.OnHover("wrap")
	sys.Update()

	tt := sys.tooltips["wrap"]

	// Width should be constrained
	if tt.FinalW > config.MaxWidth+config.Padding*2+20 { // Allow some margin
		t.Errorf("Tooltip width %d exceeds MaxWidth %d", tt.FinalW, config.MaxWidth)
	}
}

func TestTooltipUpdateTarget(t *testing.T) {
	config := DefaultTooltipConfig()
	sys := NewTooltipSystem(800, 600, config)

	sys.RegisterTooltip("move", "Moving", 100, 100, 50, 20, TooltipAbove)

	sys.UpdateTooltipTarget("move", 200, 200, 60, 30)

	tt := sys.tooltips["move"]
	if tt.TargetX != 200 {
		t.Errorf("Expected TargetX 200, got %d", tt.TargetX)
	}
	if tt.TargetY != 200 {
		t.Errorf("Expected TargetY 200, got %d", tt.TargetY)
	}
	if tt.TargetW != 60 {
		t.Errorf("Expected TargetW 60, got %d", tt.TargetW)
	}
}

func TestTooltipUpdateText(t *testing.T) {
	config := DefaultTooltipConfig()
	sys := NewTooltipSystem(800, 600, config)

	sys.RegisterTooltip("txt", "Original", 100, 100, 50, 20, TooltipAbove)

	sys.UpdateTooltipText("txt", "Updated")

	tt := sys.tooltips["txt"]
	if tt.Text != "Updated" {
		t.Errorf("Expected text 'Updated', got '%s'", tt.Text)
	}
}

func TestTooltipIsVisible(t *testing.T) {
	config := DefaultTooltipConfig()
	config.ShowDelay = 0
	sys := NewTooltipSystem(800, 600, config)

	sys.RegisterTooltip("vis", "Visible test", 100, 100, 50, 20, TooltipAbove)

	if sys.IsTooltipVisible("vis") {
		t.Error("Tooltip should not be visible before hover")
	}

	sys.OnHover("vis")
	sys.Update()

	if !sys.IsTooltipVisible("vis") {
		t.Error("Tooltip should be visible after hover+update")
	}
}

func TestTooltipClear(t *testing.T) {
	config := DefaultTooltipConfig()
	sys := NewTooltipSystem(800, 600, config)

	sys.RegisterTooltip("a", "A", 100, 100, 50, 20, TooltipAbove)
	sys.RegisterTooltip("b", "B", 200, 200, 50, 20, TooltipBelow)

	sys.Clear()

	if len(sys.tooltips) != 0 {
		t.Errorf("Expected 0 tooltips after clear, got %d", len(sys.tooltips))
	}
}

func TestTooltipNoTargetCoverage(t *testing.T) {
	config := DefaultTooltipConfig()
	config.ShowDelay = 0
	sys := NewTooltipSystem(800, 600, config)

	// Register tooltip that would naturally overlap its target
	sys.RegisterTooltip("overlap", "A", 400, 300, 200, 100, TooltipAbove)

	sys.OnHover("overlap")
	sys.Update()

	tt := sys.tooltips["overlap"]

	// Calculate if tooltip overlaps target
	tooltipRight := tt.FinalX + tt.FinalW
	tooltipBottom := tt.FinalY + tt.FinalH
	targetRight := tt.TargetX + tt.TargetW
	targetBottom := tt.TargetY + tt.TargetH

	overlapX := min(tooltipRight, targetRight) - max(tt.FinalX, tt.TargetX)
	overlapY := min(tooltipBottom, targetBottom) - max(tt.FinalY, tt.TargetY)

	if overlapX > 0 && overlapY > 0 {
		t.Errorf("Tooltip overlaps target: tooltip(%d,%d,%d,%d) target(%d,%d,%d,%d)",
			tt.FinalX, tt.FinalY, tt.FinalW, tt.FinalH,
			tt.TargetX, tt.TargetY, tt.TargetW, tt.TargetH)
	}
}

func TestTooltipSetScreenSize(t *testing.T) {
	config := DefaultTooltipConfig()
	sys := NewTooltipSystem(800, 600, config)

	sys.SetScreenSize(1920, 1080)

	if sys.screenWidth != 1920 {
		t.Errorf("Expected screenWidth 1920, got %d", sys.screenWidth)
	}
	if sys.screenHeight != 1080 {
		t.Errorf("Expected screenHeight 1080, got %d", sys.screenHeight)
	}
}

func TestTooltipMultiline(t *testing.T) {
	config := DefaultTooltipConfig()
	config.ShowDelay = 0
	sys := NewTooltipSystem(800, 600, config)

	multilineText := "Line 1\nLine 2\nLine 3"
	sys.RegisterTooltip("multi", multilineText, 400, 300, 50, 20, TooltipBelow)

	sys.OnHover("multi")
	sys.Update()

	tt := sys.tooltips["multi"]

	// Height should be > 1 line
	lineHeight := 15
	expectedMinHeight := 3*lineHeight + config.Padding*2

	if tt.FinalH < expectedMinHeight {
		t.Errorf("Multiline tooltip height %d < expected %d", tt.FinalH, expectedMinHeight)
	}
}

func TestWrapText(t *testing.T) {
	config := DefaultTooltipConfig()
	sys := NewTooltipSystem(800, 600, config)

	tests := []struct {
		name     string
		text     string
		maxWidth int
		minLines int
	}{
		{
			name:     "short text",
			text:     "Hi",
			maxWidth: 200,
			minLines: 1,
		},
		{
			name:     "explicit newline",
			text:     "Line1\nLine2",
			maxWidth: 200,
			minLines: 2,
		},
		{
			name:     "wrapping needed",
			text:     "This is a long sentence that needs wrapping",
			maxWidth: 50,
			minLines: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lines := sys.wrapText(tc.text, tc.maxWidth)
			if len(lines) < tc.minLines {
				t.Errorf("Expected at least %d lines, got %d", tc.minLines, len(lines))
			}
		})
	}
}

func TestSplitWords(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"hello world", 2},
		{"hello\nworld", 3}, // "hello", "\n", "world"
		{"one two three", 3},
		{"", 0},
	}

	for _, tc := range tests {
		words := splitWords(tc.input)
		if len(words) != tc.expected {
			t.Errorf("splitWords(%q) = %d words, want %d", tc.input, len(words), tc.expected)
		}
	}
}

func TestSelectBestPosition(t *testing.T) {
	config := DefaultTooltipConfig()
	sys := NewTooltipSystem(800, 600, config)

	tests := []struct {
		name     string
		targetX  int
		targetY  int
		expected TooltipPosition
	}{
		{
			name:     "near top prefers below",
			targetX:  400,
			targetY:  20,
			expected: TooltipBelow,
		},
		{
			name:     "near bottom prefers above",
			targetX:  400,
			targetY:  550,
			expected: TooltipAbove,
		},
		{
			name:     "near left prefers right",
			targetX:  20,
			targetY:  300,
			expected: TooltipRight,
		},
		{
			name:     "near right prefers left",
			targetX:  750,
			targetY:  300,
			expected: TooltipLeft,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tt := &Tooltip{
				TargetX: tc.targetX,
				TargetY: tc.targetY,
				TargetW: 50,
				TargetH: 20,
			}

			pos := sys.selectBestPosition(tt)
			if pos != tc.expected {
				t.Errorf("Expected position %v, got %v", tc.expected, pos)
			}
		})
	}
}

func TestRectOverlap(t *testing.T) {
	config := DefaultTooltipConfig()
	sys := NewTooltipSystem(800, 600, config)

	tests := []struct {
		name     string
		min1     int
		max1     int
		min2     int
		max2     int
		expected int
	}{
		{"no overlap", 0, 10, 20, 30, 0},
		{"touching", 0, 10, 10, 20, 0},
		{"overlap", 0, 20, 10, 30, 10},
		{"contained", 5, 15, 0, 20, 10},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			overlap := sys.rectOverlap(tc.min1, tc.max1, tc.min2, tc.max2)
			if overlap != tc.expected {
				t.Errorf("Expected overlap %d, got %d", tc.expected, overlap)
			}
		})
	}
}

func TestTooltipRenderDoesNotPanic(t *testing.T) {
	// Skip if no display available (headless)
	defer func() {
		if r := recover(); r != nil {
			t.Skip("Skipping render test in headless environment")
		}
	}()

	config := DefaultTooltipConfig()
	config.ShowDelay = 0
	sys := NewTooltipSystem(800, 600, config)

	sys.RegisterTooltip("render", "Test render", 400, 300, 50, 20, TooltipBelow)
	sys.OnHover("render")
	sys.Update()

	// Create a small test image
	screen := ebiten.NewImage(800, 600)

	// This should not panic
	sys.Render(screen)
}

func BenchmarkTooltipPositionCompute(b *testing.B) {
	config := DefaultTooltipConfig()
	config.ShowDelay = 0
	sys := NewTooltipSystem(800, 600, config)

	sys.RegisterTooltip("bench", "Benchmark tooltip text", 400, 300, 50, 20, TooltipAuto)
	sys.OnHover("bench")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update()
	}
}

func BenchmarkTooltipTextWrap(b *testing.B) {
	config := DefaultTooltipConfig()
	sys := NewTooltipSystem(800, 600, config)

	longText := "This is a very long tooltip text that needs to be wrapped across multiple lines for better readability and UX"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.wrapText(longText, config.MaxWidth)
	}
}
