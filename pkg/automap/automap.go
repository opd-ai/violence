// Package automap provides the in-game auto-mapping system.
package automap

// Map tracks explored areas of the current level.
type Map struct {
	Width, Height int
	Revealed      [][]bool
}

// NewMap creates a map for the given dimensions.
func NewMap(width, height int) *Map {
	revealed := make([][]bool, height)
	for i := range revealed {
		revealed[i] = make([]bool, width)
	}
	return &Map{Width: width, Height: height, Revealed: revealed}
}

// Reveal marks a cell as explored.
func (m *Map) Reveal(x, y int) {
	if x >= 0 && x < m.Width && y >= 0 && y < m.Height {
		m.Revealed[y][x] = true
	}
}

// Render draws the automap (stub).
func (m *Map) Render() {}

// SetGenre configures automap visuals for a genre.
func SetGenre(genreID string) {}
