// Package texture manages texture loading and atlas mapping.
package texture

import "image"

// Atlas stores a collection of textures.
type Atlas struct {
	textures map[string]image.Image
}

// NewAtlas creates an empty texture atlas.
func NewAtlas() *Atlas {
	return &Atlas{textures: make(map[string]image.Image)}
}

// Load loads a texture from the given path into the atlas.
func (a *Atlas) Load(name, path string) error {
	return nil
}

// Get retrieves a texture by name.
func (a *Atlas) Get(name string) (image.Image, bool) {
	img, ok := a.textures[name]
	return img, ok
}

// SetGenre configures texture sets for a genre.
func SetGenre(genreID string) {}
