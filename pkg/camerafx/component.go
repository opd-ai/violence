// Package camerafx provides camera effects like shake, zoom, and flash.
package camerafx

// Component is a pure data component for camera effects.
type Component struct {
	ShakeIntensity float64
	ShakeDecay     float64
	ShakeOffsetX   float64
	ShakeOffsetY   float64
	FlashAlpha     float64
	FlashDecay     float64
	FlashR         float64
	FlashG         float64
	FlashB         float64
	ZoomTarget     float64
	ZoomCurrent    float64
	ZoomSpeed      float64
	ChromaticAberr float64
	ChromaticDecay float64
	VignetteStr    float64
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "CameraFX"
}
