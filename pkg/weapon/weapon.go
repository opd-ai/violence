// Package weapon implements the weapon and firing system.
package weapon

// Weapon represents a player weapon.
type Weapon struct {
	Name     string
	Damage   float64
	FireRate float64
	Ammo     string
}

// Arsenal manages the player's collection of weapons.
type Arsenal struct {
	Weapons []Weapon
	Current int
}

// NewArsenal creates an empty arsenal.
func NewArsenal() *Arsenal {
	return &Arsenal{}
}

// Fire discharges the current weapon.
func (a *Arsenal) Fire() {}

// Reload reloads the current weapon.
func (a *Arsenal) Reload() {}

// SetGenre configures the weapon set for a genre.
func SetGenre(genreID string) {}
