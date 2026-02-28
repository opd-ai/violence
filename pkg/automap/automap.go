// Package automap provides the in-game auto-mapping system.
package automap

// AnnotationType represents special markers on the automap.
type AnnotationType int

const (
	AnnotationNone AnnotationType = iota
	AnnotationSecret
	AnnotationObjective
	AnnotationItem
)

// Annotation represents a special marker on the automap.
type Annotation struct {
	X, Y int
	Type AnnotationType
}

// Map tracks explored areas of the current level.
type Map struct {
	Width, Height int
	Revealed      [][]bool
	Annotations   []Annotation
}

// NewMap creates a map for the given dimensions.
func NewMap(width, height int) *Map {
	revealed := make([][]bool, height)
	for i := range revealed {
		revealed[i] = make([]bool, width)
	}
	return &Map{
		Width:       width,
		Height:      height,
		Revealed:    revealed,
		Annotations: make([]Annotation, 0),
	}
}

// Reveal marks a cell as explored.
func (m *Map) Reveal(x, y int) {
	if x >= 0 && x < m.Width && y >= 0 && y < m.Height {
		m.Revealed[y][x] = true
	}
}

// AddAnnotation adds a special marker to the automap.
func (m *Map) AddAnnotation(x, y int, annotationType AnnotationType) {
	// Check if annotation already exists at this position
	for _, ann := range m.Annotations {
		if ann.X == x && ann.Y == y {
			return
		}
	}
	m.Annotations = append(m.Annotations, Annotation{
		X:    x,
		Y:    y,
		Type: annotationType,
	})
}

// GetAnnotationsAt returns all annotations at the given position.
func (m *Map) GetAnnotationsAt(x, y int) []Annotation {
	var result []Annotation
	for _, ann := range m.Annotations {
		if ann.X == x && ann.Y == y {
			result = append(result, ann)
		}
	}
	return result
}

// Render draws the automap (stub).
func (m *Map) Render() {}

// SetGenre configures automap visuals for a genre.
func SetGenre(genreID string) {}
