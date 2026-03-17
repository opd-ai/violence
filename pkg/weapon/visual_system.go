// Package weapon provides weapon visual enhancement system.
package weapon

import (
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// VisualSystem manages weapon visual rendering with materials and effects.
type VisualSystem struct {
	logger *logrus.Entry
}

// NewVisualSystem creates a weapon visual enhancement system.
func NewVisualSystem() *VisualSystem {
	return &VisualSystem{
		logger: logrus.WithFields(logrus.Fields{
			"system": "weapon_visual",
		}),
	}
}

// Update processes weapon visual components (ECS Update signature).
func (vs *VisualSystem) Update(w *engine.World) {
	compType := reflect.TypeOf((*VisualComponent)(nil))
	entities := w.Query(compType)

	for _, ent := range entities {
		comp, found := w.GetComponent(ent, compType)
		if !found {
			continue
		}

		visualComp, ok := comp.(*VisualComponent)
		if !ok {
			continue
		}

		// Ensure sprite is regenerated if needed
		if visualComp.NeedsRegen {
			visualComp.GetSprite()
		}
	}
}

// prepareWeaponDraw validates a visual component and returns its sprite with draw options.
// Returns nil if the component or sprite is invalid.
func (vs *VisualSystem) prepareWeaponDraw(visualComp *VisualComponent) (*ebiten.Image, *ebiten.DrawImageOptions) {
	if visualComp == nil {
		return nil, nil
	}

	sprite := visualComp.GetSprite()
	if sprite == nil {
		vs.logger.Warn("weapon visual component has no sprite")
		return nil, nil
	}

	opts := &ebiten.DrawImageOptions{}

	// Center the sprite
	bounds := sprite.Bounds()
	opts.GeoM.Translate(-float64(bounds.Dx())/2, -float64(bounds.Dy())/2)

	return sprite, opts
}

// RenderWeapon renders a weapon sprite at the given position.
func (vs *VisualSystem) RenderWeapon(
	screen *ebiten.Image,
	visualComp *VisualComponent,
	x, y float64,
	scale float64,
) {
	sprite, opts := vs.prepareWeaponDraw(visualComp)
	if sprite == nil {
		return
	}

	// Scale
	opts.GeoM.Scale(scale, scale)

	// Position
	opts.GeoM.Translate(x, y)

	screen.DrawImage(sprite, opts)
}

// RenderWeaponWithRotation renders a weapon sprite with rotation.
func (vs *VisualSystem) RenderWeaponWithRotation(
	screen *ebiten.Image,
	visualComp *VisualComponent,
	x, y float64,
	rotation float64,
	scale float64,
) {
	sprite, opts := vs.prepareWeaponDraw(visualComp)
	if sprite == nil {
		return
	}

	// Rotate
	opts.GeoM.Rotate(rotation)

	// Scale
	opts.GeoM.Scale(scale, scale)

	// Position
	opts.GeoM.Translate(x, y)

	screen.DrawImage(sprite, opts)
}

// UpdateWeaponDamage updates weapon damage state based on durability.
func (vs *VisualSystem) UpdateWeaponDamage(visualComp *VisualComponent, durability float64) {
	if visualComp == nil {
		return
	}

	var newState DamageState
	if durability >= 0.8 {
		newState = DamagePristine
	} else if durability >= 0.5 {
		newState = DamageScratched
	} else if durability >= 0.2 {
		newState = DamageWorn
	} else {
		newState = DamageBroken
	}

	visualComp.SetDamageState(newState)
}
