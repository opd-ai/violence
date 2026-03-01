// Package engine provides player entity schema and factory functions.
package engine

import "math"

// Position component represents an entity's location in 2D space.
type Position struct {
	X, Y float64
}

// Health component represents an entity's hit points.
type Health struct {
	Current int
	Max     int
}

// Armor component represents damage reduction.
type Armor struct {
	Value int
}

// Inventory component holds items and resources.
type Inventory struct {
	Items   []string
	Credits int
}

// Camera component tracks view direction and field of view.
type Camera struct {
	DirX, DirY   float64
	PlaneX       float64
	PlaneY       float64
	FOV          float64
	PitchRadians float64
}

// Input component stores player input state.
type Input struct {
	Forward   bool
	Backward  bool
	Left      bool
	Right     bool
	TurnLeft  bool
	TurnRight bool
	Fire      bool
	AltFire   bool
	Interact  bool
	Reload    bool
}

// NewPlayerEntity creates a player entity with canonical component set.
// The player entity includes Position, Health, Armor, Inventory, Camera, and Input.
// Position defaults to (0, 0), Health to 100/100, Armor to 0,
// Inventory to empty with 0 credits, Camera facing north with default FOV,
// and Input with all controls released.
func (w *World) NewPlayerEntity(x, y float64) Entity {
	e := w.AddEntity()

	// Position
	w.AddComponent(e, &Position{X: x, Y: y})
	w.AddArchetypeComponent(e, ComponentIDPosition)

	// Health
	w.AddComponent(e, &Health{Current: 100, Max: 100})
	w.AddArchetypeComponent(e, ComponentIDHealth)

	// Armor
	w.AddComponent(e, &Armor{Value: 0})
	w.AddArchetypeComponent(e, ComponentIDArmor)

	// Inventory
	w.AddComponent(e, &Inventory{Items: []string{}, Credits: 0})
	w.AddArchetypeComponent(e, ComponentIDInventory)

	// Camera (facing north, FOV 66 degrees)
	fovDegrees := 66.0
	fovRadians := fovDegrees * (math.Pi / 180.0)
	w.AddComponent(e, &Camera{
		DirX:         0,
		DirY:         -1,
		PlaneX:       math.Tan(fovRadians / 2.0),
		PlaneY:       0,
		FOV:          fovDegrees,
		PitchRadians: 0,
	})
	w.AddArchetypeComponent(e, ComponentIDCamera)

	// Input
	w.AddComponent(e, &Input{})
	w.AddArchetypeComponent(e, ComponentIDInput)

	return e
}

// IsPlayer checks if an entity has all canonical player components.
func (w *World) IsPlayer(e Entity) bool {
	mask := w.GetArchetype(e)
	requiredMask := uint64(1<<ComponentIDPosition |
		1<<ComponentIDHealth |
		1<<ComponentIDArmor |
		1<<ComponentIDInventory |
		1<<ComponentIDCamera |
		1<<ComponentIDInput)
	return mask&requiredMask == requiredMask
}
