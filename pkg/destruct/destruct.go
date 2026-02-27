// Package destruct implements destructible environment objects.
package destruct

// Destructible represents a destructible object in the world.
type Destructible struct {
	ID     string
	Health float64
}

// Damage applies damage to a destructible object.
func (d *Destructible) Damage(amount float64) {
	d.Health -= amount
}

// Destroy removes the destructible object from the world.
func (d *Destructible) Destroy() {
	d.Health = 0
}

// SetGenre configures destructible types for a genre.
func SetGenre(genreID string) {}
