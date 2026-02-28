// Package genre provides a registry of game genre definitions.
package genre

const (
	Fantasy   = "fantasy"
	SciFi     = "scifi"
	Horror    = "horror"
	Cyberpunk = "cyberpunk"
	PostApoc  = "postapoc"
)

// Genre describes a game genre and its associated asset/behavior set.
type Genre struct {
	ID   string
	Name string
}

// Registry holds all registered genres.
type Registry struct {
	genres map[string]Genre
}

// NewRegistry creates an empty genre registry.
func NewRegistry() *Registry {
	return &Registry{genres: make(map[string]Genre)}
}

// Register adds a genre to the registry.
func (r *Registry) Register(g Genre) {
	r.genres[g.ID] = g
}

// Get retrieves a genre by ID.
func (r *Registry) Get(id string) (Genre, bool) {
	g, ok := r.genres[id]
	return g, ok
}
