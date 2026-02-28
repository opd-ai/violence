// Package camera manages the first-person camera.
package camera

import (
	"math"

	"github.com/opd-ai/violence/pkg/raycaster"
)

const (
	MaxPitch         = 30.0
	MinPitch         = -30.0
	HeadBobFrequency = 8.0
	HeadBobAmplitude = 0.05
)

// Camera represents the player's viewpoint.
type Camera struct {
	X, Y          float64
	DirX          float64
	DirY          float64
	FOV           float64
	Pitch         float64
	HeadBob       float64
	headBobPhase  float64
	movementSpeed float64
}

// NewCamera creates a camera with default settings.
func NewCamera(fov float64) *Camera {
	return &Camera{FOV: fov, DirX: 1}
}

// Update applies movement deltas and updates head-bob.
// deltaX, deltaY: position deltas in world space
// deltaDirX, deltaDirY: direction vector changes
// deltaPitch: pitch angle change in degrees
func (c *Camera) Update(deltaX, deltaY, deltaDirX, deltaDirY, deltaPitch float64) {
	c.X += deltaX
	c.Y += deltaY
	c.DirX += deltaDirX
	c.DirY += deltaDirY

	dirLen := math.Sqrt(c.DirX*c.DirX + c.DirY*c.DirY)
	if dirLen > 0.0001 {
		c.DirX /= dirLen
		c.DirY /= dirLen
	}

	c.Pitch += deltaPitch
	if c.Pitch > MaxPitch {
		c.Pitch = MaxPitch
	}
	if c.Pitch < MinPitch {
		c.Pitch = MinPitch
	}

	movementMagnitude := math.Sqrt(deltaX*deltaX + deltaY*deltaY)
	c.movementSpeed = movementMagnitude

	if movementMagnitude > 0.001 {
		c.headBobPhase += movementMagnitude * HeadBobFrequency
		c.HeadBob = math.Sin(c.headBobPhase) * HeadBobAmplitude
	} else {
		c.headBobPhase = 0
		c.HeadBob = 0
	}
}

// Rotate rotates the camera direction by the given angle in radians.
func (c *Camera) Rotate(angleRadians float64) {
	oldDirX := c.DirX
	sinAngle := raycaster.Sin(angleRadians)
	cosAngle := raycaster.Cos(angleRadians)
	c.DirX = c.DirX*cosAngle - c.DirY*sinAngle
	c.DirY = oldDirX*sinAngle + c.DirY*cosAngle
}

// SetGenre configures camera behavior for a genre.
func SetGenre(genreID string) {}
