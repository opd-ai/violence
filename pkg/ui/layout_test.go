package ui

import (
	"testing"
)

func TestLayoutManager_NoOverlap(t *testing.T) {
	lm := NewLayoutManager(800, 600)

	// Reserve two non-overlapping elements
	x1, y1, ok1 := lm.Reserve("elem1", 100, 100, 50, 20, PriorityImportant, true)
	if !ok1 {
		t.Fatal("First element should reserve successfully")
	}
	if x1 != 100 || y1 != 100 {
		t.Errorf("First element position changed: got (%f, %f), want (100, 100)", x1, y1)
	}

	x2, y2, ok2 := lm.Reserve("elem2", 200, 200, 50, 20, PriorityImportant, true)
	if !ok2 {
		t.Fatal("Second non-overlapping element should reserve successfully")
	}
	if x2 != 200 || y2 != 200 {
		t.Errorf("Second element position changed: got (%f, %f), want (200, 200)", x2, y2)
	}

	if lm.GetElementCount() != 2 {
		t.Errorf("Expected 2 elements, got %d", lm.GetElementCount())
	}
}

func TestLayoutManager_OverlapPriority(t *testing.T) {
	lm := NewLayoutManager(800, 600)

	// Reserve high-priority element
	x1, y1, ok1 := lm.Reserve("critical", 100, 100, 50, 20, PriorityCritical, false)
	if !ok1 || x1 != 100 || y1 != 100 {
		t.Fatal("Critical element should reserve at requested position")
	}

	// Try to reserve overlapping low-priority element
	x2, y2, ok2 := lm.Reserve("ambient", 110, 105, 50, 20, PriorityAmbient, true)

	// Should either be repositioned or rejected
	if ok2 && x2 == 110 && y2 == 105 {
		t.Error("Low-priority element should not overlap high-priority element")
	}
}

func TestLayoutManager_DamageNumberStacking(t *testing.T) {
	lm := NewLayoutManager(800, 600)

	// Reserve multiple damage numbers at similar X positions
	_, y1, ok1 := lm.ReserveDamageNumber("dmg1", 100, 200, 30, 15, PriorityImportant)
	if !ok1 {
		t.Fatal("First damage number should reserve successfully")
	}

	_, y2, ok2 := lm.ReserveDamageNumber("dmg2", 102, 200, 30, 15, PriorityImportant)
	if !ok2 {
		t.Fatal("Second damage number should reserve successfully")
	}

	// Second damage number should be stacked above first
	if y2 >= y1 {
		t.Errorf("Damage numbers should stack vertically: y1=%f, y2=%f", y1, y2)
	}

	_, y3, ok3 := lm.ReserveDamageNumber("dmg3", 98, 200, 30, 15, PriorityImportant)
	if !ok3 {
		t.Fatal("Third damage number should reserve successfully")
	}

	// Third should stack above second
	if y3 >= y2 {
		t.Errorf("Third damage number should stack above second: y2=%f, y3=%f", y2, y3)
	}
}

func TestLayoutManager_FixedElement(t *testing.T) {
	lm := NewLayoutManager(800, 600)

	// Reserve fixed (immovable) element
	x1, y1, ok1 := lm.Reserve("fixed", 100, 100, 100, 50, PriorityImportant, false)
	if !ok1 || x1 != 100 || y1 != 100 {
		t.Fatal("Fixed element should reserve at requested position")
	}

	// Try to reserve movable element at same location
	x2, y2, ok2 := lm.Reserve("movable", 100, 100, 50, 30, PriorityImportant, true)

	// Movable element should be repositioned or rejected
	if ok2 {
		// If accepted, must be at different position
		if x2 == 100 && y2 == 100 {
			t.Error("Movable element should not overlap fixed element at same position")
		}
	}
	// If rejected (ok2 == false), that's also acceptable behavior
}

func TestLayoutManager_Clear(t *testing.T) {
	lm := NewLayoutManager(800, 600)

	lm.Reserve("elem1", 100, 100, 50, 20, PriorityImportant, true)
	lm.Reserve("elem2", 200, 200, 50, 20, PriorityImportant, true)

	if lm.GetElementCount() != 2 {
		t.Fatalf("Expected 2 elements before clear, got %d", lm.GetElementCount())
	}

	lm.Clear()

	if lm.GetElementCount() != 0 {
		t.Errorf("Expected 0 elements after clear, got %d", lm.GetElementCount())
	}
}

func TestLayoutManager_ScreenBounds(t *testing.T) {
	lm := NewLayoutManager(800, 600)

	// Try to reserve element outside screen bounds
	x, y, ok := lm.Reserve("offscreen", -50, -50, 50, 20, PriorityImportant, true)

	// Should either reposition inside bounds or reject
	if ok && (x < 0 || y < 0 || x+50 > 800 || y+20 > 600) {
		t.Error("Element should not be placed outside screen bounds")
	}
	// It's also acceptable to reject the element (ok == false)
}

func TestRect_Overlaps(t *testing.T) {
	tests := []struct {
		name     string
		r1       Rect
		r2       Rect
		expected bool
	}{
		{
			name:     "no overlap - separated horizontally",
			r1:       Rect{X: 0, Y: 0, Width: 10, Height: 10},
			r2:       Rect{X: 20, Y: 0, Width: 10, Height: 10},
			expected: false,
		},
		{
			name:     "no overlap - separated vertically",
			r1:       Rect{X: 0, Y: 0, Width: 10, Height: 10},
			r2:       Rect{X: 0, Y: 20, Width: 10, Height: 10},
			expected: false,
		},
		{
			name:     "overlap - partial",
			r1:       Rect{X: 0, Y: 0, Width: 10, Height: 10},
			r2:       Rect{X: 5, Y: 5, Width: 10, Height: 10},
			expected: true,
		},
		{
			name:     "overlap - contained",
			r1:       Rect{X: 0, Y: 0, Width: 20, Height: 20},
			r2:       Rect{X: 5, Y: 5, Width: 5, Height: 5},
			expected: true,
		},
		{
			name:     "overlap - edge touching",
			r1:       Rect{X: 0, Y: 0, Width: 10, Height: 10},
			r2:       Rect{X: 10, Y: 0, Width: 10, Height: 10},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.r1.Overlaps(tt.r2)
			if result != tt.expected {
				t.Errorf("Overlaps() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLayoutManager_AlternativePosition(t *testing.T) {
	lm := NewLayoutManager(800, 600)

	// Reserve element at center
	lm.Reserve("center", 100, 100, 50, 20, PriorityCritical, false)

	// Try to reserve another element at same position but with lower priority
	x, y, ok := lm.Reserve("nearby", 100, 100, 50, 20, PrioritySecondary, true)

	if !ok {
		// Acceptable: element was suppressed
		return
	}

	// If accepted, must be at different position
	if x == 100 && y == 100 {
		t.Error("Lower priority element should find alternative position")
	}

	// Should be near original position
	dx := x - 100
	dy := y - 100
	distance := dx*dx + dy*dy
	if distance > 2500 { // More than 50 pixels away
		t.Errorf("Alternative position too far from original: (%f, %f)", x, y)
	}
}

func BenchmarkLayoutManager_Reserve(b *testing.B) {
	lm := NewLayoutManager(1920, 1080)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lm.Clear()
		// Simulate typical frame with many UI elements
		for j := 0; j < 20; j++ {
			x := float32((j * 50) % 800)
			y := float32((j * 30) % 600)
			lm.Reserve("elem", x, y, 40, 15, PriorityImportant, true)
		}
	}
}

func BenchmarkLayoutManager_DamageNumberStacking(b *testing.B) {
	lm := NewLayoutManager(1920, 1080)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lm.Clear()
		// Simulate burst of damage numbers at same location
		for j := 0; j < 10; j++ {
			lm.ReserveDamageNumber("dmg", 400, 300, 30, 15, PriorityImportant)
		}
	}
}
