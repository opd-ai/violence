package motion

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// RenderHelper provides utilities for rendering with organic motion effects.
type RenderHelper struct {
	system *System
}

// NewRenderHelper creates a helper for applying motion effects during rendering.
func NewRenderHelper(system *System) *RenderHelper {
	return &RenderHelper{system: system}
}

// ApplyMotionTransform configures draw options to include squash/stretch and breath offset.
// Returns modified Y position (for breathing) and configured draw options.
func (rh *RenderHelper) ApplyMotionTransform(motion *Component, baseX, baseY float64, baseOpts *ebiten.DrawImageOptions) (renderY float64, opts *ebiten.DrawImageOptions) {
	if motion == nil {
		return baseY, baseOpts
	}

	opts = baseOpts
	if opts == nil {
		opts = &ebiten.DrawImageOptions{}
	}

	// Apply squash/stretch scaling
	scaleX, scaleY := rh.system.GetSquashStretch(motion)

	// Get breathing offset
	breathOffset := rh.system.GetBreathOffset(motion)

	// Apply scale transform centered on sprite
	opts.GeoM.Scale(scaleX, scaleY)

	// Return adjusted Y position with breath offset
	renderY = baseY + breathOffset

	return renderY, opts
}

// GetTrailPositions returns all trail segment positions for rendering.
func (rh *RenderHelper) GetTrailPositions(motion *Component) []struct{ X, Y float64 } {
	if motion == nil || motion.TrailLength == 0 {
		return nil
	}

	positions := make([]struct{ X, Y float64 }, 0, motion.TrailLength)
	for i := 0; i < motion.TrailLength; i++ {
		x, y, valid := rh.system.GetTrailSegment(motion, i)
		if valid {
			positions = append(positions, struct{ X, Y float64 }{X: x, Y: y})
		}
	}

	return positions
}
