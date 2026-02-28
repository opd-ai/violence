package automap

import "testing"

func TestNewMap(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"small_map", 10, 10},
		{"rectangular", 20, 15},
		{"large_map", 100, 100},
		{"minimal", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMap(tt.width, tt.height)
			if m == nil {
				t.Fatal("NewMap returned nil")
			}
			if m.Width != tt.width {
				t.Errorf("Width: expected %d, got %d", tt.width, m.Width)
			}
			if m.Height != tt.height {
				t.Errorf("Height: expected %d, got %d", tt.height, m.Height)
			}
			if m.Revealed == nil {
				t.Fatal("Revealed array not initialized")
			}
			if len(m.Revealed) != tt.height {
				t.Errorf("Revealed height: expected %d, got %d", tt.height, len(m.Revealed))
			}
			for y := 0; y < tt.height; y++ {
				if len(m.Revealed[y]) != tt.width {
					t.Errorf("Revealed width at row %d: expected %d, got %d", y, tt.width, len(m.Revealed[y]))
				}
			}
			if m.Annotations == nil {
				t.Fatal("Annotations array not initialized")
			}
			if len(m.Annotations) != 0 {
				t.Errorf("Annotations should be initially empty, got %d", len(m.Annotations))
			}
		})
	}
}

func TestRevealTile(t *testing.T) {
	m := NewMap(10, 10)

	// Initially all tiles should be unrevealed
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if m.Revealed[y][x] {
				t.Errorf("Tile (%d,%d) should be initially unrevealed", x, y)
			}
		}
	}

	// Reveal single tile
	m.Reveal(5, 5)
	if !m.Revealed[5][5] {
		t.Error("Tile (5,5) should be revealed")
	}

	// Verify adjacent tiles still unrevealed
	if m.Revealed[4][5] || m.Revealed[6][5] || m.Revealed[5][4] || m.Revealed[5][6] {
		t.Error("Adjacent tiles should remain unrevealed")
	}
}

func TestRevealMultipleTiles(t *testing.T) {
	m := NewMap(10, 10)

	tiles := []struct{ x, y int }{
		{0, 0}, {9, 9}, {5, 5}, {2, 7}, {8, 3},
	}

	for _, tile := range tiles {
		m.Reveal(tile.x, tile.y)
	}

	// Verify all revealed tiles
	for _, tile := range tiles {
		if !m.Revealed[tile.y][tile.x] {
			t.Errorf("Tile (%d,%d) should be revealed", tile.x, tile.y)
		}
	}
}

func TestRevealOutOfBounds(t *testing.T) {
	m := NewMap(10, 10)

	tests := []struct {
		name string
		x, y int
	}{
		{"negative_x", -1, 5},
		{"negative_y", 5, -1},
		{"negative_both", -1, -1},
		{"x_too_large", 10, 5},
		{"y_too_large", 5, 10},
		{"both_too_large", 10, 10},
		{"far_negative", -100, -100},
		{"far_positive", 1000, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Reveal(%d,%d) panicked: %v", tt.x, tt.y, r)
				}
			}()
			m.Reveal(tt.x, tt.y)
		})
	}

	// Verify no tiles were revealed
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if m.Revealed[y][x] {
				t.Errorf("No tiles should be revealed after out-of-bounds calls, but (%d,%d) is revealed", x, y)
			}
		}
	}
}

func TestRevealSameTileMultipleTimes(t *testing.T) {
	m := NewMap(10, 10)

	// Reveal same tile multiple times
	for i := 0; i < 5; i++ {
		m.Reveal(5, 5)
	}

	// Should still be revealed (idempotent)
	if !m.Revealed[5][5] {
		t.Error("Tile (5,5) should be revealed")
	}
}

func TestRevealCornerCases(t *testing.T) {
	m := NewMap(10, 10)

	// Test all corners
	corners := []struct{ x, y int }{
		{0, 0}, // top-left
		{9, 0}, // top-right
		{0, 9}, // bottom-left
		{9, 9}, // bottom-right
	}

	for _, corner := range corners {
		m.Reveal(corner.x, corner.y)
		if !m.Revealed[corner.y][corner.x] {
			t.Errorf("Corner (%d,%d) should be revealed", corner.x, corner.y)
		}
	}
}

func TestRevealEdges(t *testing.T) {
	m := NewMap(10, 10)

	// Reveal top edge
	for x := 0; x < m.Width; x++ {
		m.Reveal(x, 0)
	}

	// Reveal left edge
	for y := 0; y < m.Height; y++ {
		m.Reveal(0, y)
	}

	// Verify edges revealed
	for x := 0; x < m.Width; x++ {
		if !m.Revealed[0][x] {
			t.Errorf("Top edge at x=%d should be revealed", x)
		}
	}

	for y := 0; y < m.Height; y++ {
		if !m.Revealed[y][0] {
			t.Errorf("Left edge at y=%d should be revealed", y)
		}
	}
}

func TestRenderDoesNotPanic(t *testing.T) {
	m := NewMap(10, 10)

	// Render on empty map
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Render panicked on empty map: %v", r)
		}
	}()
	m.Render()

	// Render on partially revealed map
	m.Reveal(5, 5)
	m.Render()

	// Render on fully revealed map
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			m.Reveal(x, y)
		}
	}
	m.Render()
}

func TestSetGenre(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetGenre(%q) panicked: %v", genre, r)
				}
			}()
			SetGenre(genre)
		})
	}
}

func TestMapDimensions(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"wide", 100, 10},
		{"tall", 10, 100},
		{"square_small", 5, 5},
		{"square_large", 50, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMap(tt.width, tt.height)

			// Test revealing in each dimension
			m.Reveal(tt.width-1, tt.height-1) // Last valid cell
			if !m.Revealed[tt.height-1][tt.width-1] {
				t.Error("Last valid cell should be revealed")
			}

			// Test boundary
			m.Reveal(tt.width, tt.height) // Should not panic (out of bounds)
		})
	}
}

func TestAddAnnotation(t *testing.T) {
	m := NewMap(10, 10)

	// Add secret annotation
	m.AddAnnotation(5, 5, AnnotationSecret)
	if len(m.Annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(m.Annotations))
	}

	ann := m.Annotations[0]
	if ann.X != 5 || ann.Y != 5 {
		t.Errorf("Annotation position = (%d, %d), want (5, 5)", ann.X, ann.Y)
	}
	if ann.Type != AnnotationSecret {
		t.Errorf("Annotation type = %d, want %d (AnnotationSecret)", ann.Type, AnnotationSecret)
	}
}

func TestAddMultipleAnnotations(t *testing.T) {
	m := NewMap(10, 10)

	m.AddAnnotation(1, 1, AnnotationSecret)
	m.AddAnnotation(2, 2, AnnotationObjective)
	m.AddAnnotation(3, 3, AnnotationItem)

	if len(m.Annotations) != 3 {
		t.Errorf("Expected 3 annotations, got %d", len(m.Annotations))
	}

	expectedTypes := []AnnotationType{AnnotationSecret, AnnotationObjective, AnnotationItem}
	for i, ann := range m.Annotations {
		if ann.Type != expectedTypes[i] {
			t.Errorf("Annotation %d type = %d, want %d", i, ann.Type, expectedTypes[i])
		}
	}
}

func TestAddAnnotationDuplicate(t *testing.T) {
	m := NewMap(10, 10)

	// Add same annotation twice
	m.AddAnnotation(5, 5, AnnotationSecret)
	m.AddAnnotation(5, 5, AnnotationSecret)

	if len(m.Annotations) != 1 {
		t.Errorf("Duplicate annotation should not be added, got %d annotations", len(m.Annotations))
	}
}

func TestGetAnnotationsAt(t *testing.T) {
	m := NewMap(10, 10)

	m.AddAnnotation(5, 5, AnnotationSecret)
	m.AddAnnotation(3, 3, AnnotationObjective)

	// Get annotation at (5, 5)
	anns := m.GetAnnotationsAt(5, 5)
	if len(anns) != 1 {
		t.Errorf("Expected 1 annotation at (5, 5), got %d", len(anns))
	}
	if len(anns) > 0 && anns[0].Type != AnnotationSecret {
		t.Errorf("Annotation type = %d, want %d", anns[0].Type, AnnotationSecret)
	}

	// Get annotation at empty position
	anns = m.GetAnnotationsAt(7, 7)
	if len(anns) != 0 {
		t.Errorf("Expected 0 annotations at (7, 7), got %d", len(anns))
	}
}

func TestGetAnnotationsAtMultiple(t *testing.T) {
	m := NewMap(10, 10)

	// Note: AddAnnotation prevents duplicates at same position,
	// so this tests the query function with single annotation
	m.AddAnnotation(5, 5, AnnotationSecret)

	anns := m.GetAnnotationsAt(5, 5)
	if len(anns) != 1 {
		t.Errorf("Expected 1 annotation at (5, 5), got %d", len(anns))
	}
}

func TestAnnotationTypes(t *testing.T) {
	m := NewMap(10, 10)

	types := []AnnotationType{
		AnnotationNone,
		AnnotationSecret,
		AnnotationObjective,
		AnnotationItem,
	}

	for i, annType := range types {
		m.AddAnnotation(i, i, annType)
	}

	// Verify all types stored correctly
	for i, annType := range types {
		anns := m.GetAnnotationsAt(i, i)
		if len(anns) != 1 {
			t.Errorf("Expected 1 annotation at (%d, %d), got %d", i, i, len(anns))
			continue
		}
		if anns[0].Type != annType {
			t.Errorf("Annotation at (%d, %d) type = %d, want %d", i, i, anns[0].Type, annType)
		}
	}
}
