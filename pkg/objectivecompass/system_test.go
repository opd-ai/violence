package objectivecompass

import (
	"math"
	"testing"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem("fantasy")

	if sys.genreID != "fantasy" {
		t.Errorf("Expected genreID 'fantasy', got %q", sys.genreID)
	}
	if sys.screenWidth != 320 {
		t.Errorf("Expected screenWidth 320, got %d", sys.screenWidth)
	}
	if sys.screenHeight != 200 {
		t.Errorf("Expected screenHeight 200, got %d", sys.screenHeight)
	}
	if len(sys.objectives) != 0 {
		t.Errorf("Expected empty objectives map, got %d entries", len(sys.objectives))
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy")

	tests := []struct {
		genre string
	}{
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
		{"fantasy"},
	}

	for _, tt := range tests {
		sys.SetGenre(tt.genre)
		if sys.genreID != tt.genre {
			t.Errorf("Expected genreID %q, got %q", tt.genre, sys.genreID)
		}
	}
}

func TestSetGenreInvalid(t *testing.T) {
	sys := NewSystem("fantasy")
	originalStyle := sys.style

	sys.SetGenre("nonexistent_genre")

	// Should fall back to fantasy
	if sys.style.MainColor != originalStyle.MainColor {
		t.Error("Invalid genre should fall back to fantasy style")
	}
}

func TestSetScreenSize(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(640, 480)

	if sys.screenWidth != 640 {
		t.Errorf("Expected screenWidth 640, got %d", sys.screenWidth)
	}
	if sys.screenHeight != 480 {
		t.Errorf("Expected screenHeight 480, got %d", sys.screenHeight)
	}
}

func TestSetPlayerPosition(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetPlayerPosition(100.5, 200.5, math.Pi/4)

	if sys.playerX != 100.5 {
		t.Errorf("Expected playerX 100.5, got %f", sys.playerX)
	}
	if sys.playerY != 200.5 {
		t.Errorf("Expected playerY 200.5, got %f", sys.playerY)
	}
	if sys.playerAngle != math.Pi/4 {
		t.Errorf("Expected playerAngle π/4, got %f", sys.playerAngle)
	}
}

func TestAddObjective(t *testing.T) {
	sys := NewSystem("fantasy")

	sys.AddObjective("obj1", TypeMain, 100, 200)
	sys.AddObjective("obj2", TypeBonus, 300, 400)

	if sys.GetObjectiveCount() != 2 {
		t.Errorf("Expected 2 objectives, got %d", sys.GetObjectiveCount())
	}

	if sys.objectives["obj1"] == nil {
		t.Error("Expected obj1 to exist")
	}
	if sys.objectives["obj1"].ObjType != TypeMain {
		t.Errorf("Expected obj1 to be TypeMain, got %d", sys.objectives["obj1"].ObjType)
	}
}

func TestRemoveObjective(t *testing.T) {
	sys := NewSystem("fantasy")

	sys.AddObjective("obj1", TypeMain, 100, 200)
	sys.AddObjective("obj2", TypeBonus, 300, 400)

	sys.RemoveObjective("obj1")

	if sys.GetObjectiveCount() != 1 {
		t.Errorf("Expected 1 objective after removal, got %d", sys.GetObjectiveCount())
	}
	if sys.objectives["obj1"] != nil {
		t.Error("Expected obj1 to be removed")
	}
}

func TestCompleteObjective(t *testing.T) {
	sys := NewSystem("fantasy")

	sys.AddObjective("obj1", TypeMain, 100, 200)
	sys.CompleteObjective("obj1")

	if !sys.objectives["obj1"].Completed {
		t.Error("Expected obj1 to be completed")
	}
}

func TestUpdateObjectivePosition(t *testing.T) {
	sys := NewSystem("fantasy")

	sys.AddObjective("obj1", TypeMain, 100, 200)
	sys.UpdateObjectivePosition("obj1", 500, 600)

	obj := sys.objectives["obj1"]
	if obj.WorldX != 500 || obj.WorldY != 600 {
		t.Errorf("Expected position (500, 600), got (%f, %f)", obj.WorldX, obj.WorldY)
	}
}

func TestClearObjectives(t *testing.T) {
	sys := NewSystem("fantasy")

	sys.AddObjective("obj1", TypeMain, 100, 200)
	sys.AddObjective("obj2", TypeBonus, 300, 400)
	sys.ClearObjectives()

	if sys.GetObjectiveCount() != 0 {
		t.Errorf("Expected 0 objectives after clear, got %d", sys.GetObjectiveCount())
	}
}

func TestUpdate(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)
	sys.SetPlayerPosition(0, 0, 0)

	// Add objective far to the right
	sys.AddObjective("obj1", TypeMain, 100, 0)

	// Run several update cycles
	for i := 0; i < 10; i++ {
		sys.Update(0.016) // ~60fps
	}

	obj := sys.objectives["obj1"]
	if obj.Distance == 0 {
		t.Error("Expected non-zero distance after update")
	}
	if obj.PulsePhase == 0 {
		t.Error("Expected pulse animation to advance")
	}
}

func TestUpdateCompletedObjectiveFades(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.AddObjective("obj1", TypeMain, 100, 0)
	sys.CompleteObjective("obj1")

	// Update multiple times to allow fade
	for i := 0; i < 60; i++ {
		sys.Update(0.016)
	}

	obj := sys.objectives["obj1"]
	if obj.Alpha > 0.5 {
		t.Errorf("Expected completed objective to fade, alpha: %f", obj.Alpha)
	}
}

func TestGetVisibleCount(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)
	sys.SetPlayerPosition(0, 0, 0)

	// Add far-away objective (should be visible indicator)
	sys.AddObjective("obj1", TypeMain, 100, 0)

	// Initial alpha is 1.0, and off-screen
	sys.Update(0.016)

	visible := sys.GetVisibleCount()
	if visible < 1 {
		t.Errorf("Expected at least 1 visible indicator, got %d", visible)
	}
}

func TestGetObjectiveColor(t *testing.T) {
	sys := NewSystem("fantasy")

	tests := []struct {
		objType ObjectiveType
		name    string
	}{
		{TypeMain, "main"},
		{TypeBonus, "bonus"},
		{TypePOI, "poi"},
		{TypeExit, "exit"},
	}

	for _, tt := range tests {
		col := sys.getObjectiveColor(tt.objType)
		if col.A == 0 {
			t.Errorf("Expected non-zero alpha for %s color", tt.name)
		}
	}
}

func TestNormalizeAngle(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0, 0},
		{math.Pi, math.Pi},
		{-math.Pi, -math.Pi},
		{2 * math.Pi, 0},
		{-2 * math.Pi, 0},
		{3 * math.Pi, math.Pi},
		{-3 * math.Pi, -math.Pi},
	}

	for _, tt := range tests {
		result := normalizeAngle(tt.input)
		// Allow small floating point tolerance
		if math.Abs(result-tt.expected) > 0.0001 {
			t.Errorf("normalizeAngle(%f) = %f, expected %f", tt.input, result, tt.expected)
		}
	}
}

func TestLerpf(t *testing.T) {
	tests := []struct {
		a, b, t  float64
		expected float64
	}{
		{0, 10, 0.5, 5},
		{0, 10, 0, 0},
		{0, 10, 1, 10},
		{0, 10, -0.5, 0}, // Clamped
		{0, 10, 1.5, 10}, // Clamped
		{5, 15, 0.25, 7.5},
	}

	for _, tt := range tests {
		result := lerpf(tt.a, tt.b, tt.t)
		if math.Abs(result-tt.expected) > 0.0001 {
			t.Errorf("lerpf(%f, %f, %f) = %f, expected %f", tt.a, tt.b, tt.t, result, tt.expected)
		}
	}
}

func TestCalculateEdgePosition(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)

	halfW := 160.0
	halfH := 100.0
	padding := float64(sys.style.EdgePadding)

	// Test right edge (angle 0)
	x, y := sys.calculateEdgePosition(0, halfW, halfH, padding)
	expectedX := float32(halfW + halfW - padding)
	if x < expectedX-10 {
		t.Errorf("Angle 0: Expected x near %f, got %f", expectedX, x)
	}

	// Test left edge (angle π)
	x, _ = sys.calculateEdgePosition(math.Pi, halfW, halfH, padding)
	expectedX = float32(padding)
	if x > expectedX+10 {
		t.Errorf("Angle π: Expected x near %f, got %f", expectedX, x)
	}

	// Test top edge (angle π/2)
	_, y = sys.calculateEdgePosition(math.Pi/2, halfW, halfH, padding)
	expectedY := float32(padding)
	if y > expectedY+10 {
		t.Errorf("Angle π/2: Expected y near %f, got %f", expectedY, y)
	}
}

func TestGetStyle(t *testing.T) {
	sys := NewSystem("cyberpunk")
	style := sys.GetStyle()

	// Cyberpunk should have distinct magenta color
	if style.MainColor.R < 200 || style.MainColor.B < 100 {
		t.Error("Expected cyberpunk magenta main color")
	}
}

func BenchmarkUpdate(b *testing.B) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)
	sys.SetPlayerPosition(0, 0, 0)

	// Add multiple objectives
	for i := 0; i < 10; i++ {
		sys.AddObjective(
			string(rune('a'+i)),
			ObjectiveType(i%4),
			float64(i*50),
			float64(i*30),
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(0.016)
	}
}

func BenchmarkResolveOverlaps(b *testing.B) {
	sys := NewSystem("fantasy")
	sys.SetScreenSize(320, 200)
	sys.SetPlayerPosition(0, 0, 0)

	// Add overlapping objectives
	for i := 0; i < 8; i++ {
		sys.AddObjective(
			string(rune('a'+i)),
			TypeMain,
			float64(i*5)+100,
			0,
		)
	}

	// Warm up to set positions
	sys.Update(0.016)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.resolveOverlaps()
	}
}
