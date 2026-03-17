package flicker

import (
	"github.com/opd-ai/violence/pkg/lighting"
)

// FlickerResult contains the calculated flicker values for a light.
type FlickerResult struct {
	Intensity float64
	R, G, B   float64
}

// LightFlickerBridge provides integration between the flicker system and lighting package.
type LightFlickerBridge struct {
	sys    *System
	params map[int64]*FlickerParams // Cached params by seed
}

// NewLightFlickerBridge creates a bridge for integrating with lighting.LightComponent.
func NewLightFlickerBridge(genre string) *LightFlickerBridge {
	return &LightFlickerBridge{
		sys:    NewSystem(genre),
		params: make(map[int64]*FlickerParams),
	}
}

// SetGenre updates the flicker genre.
func (b *LightFlickerBridge) SetGenre(genre string) {
	b.sys.SetGenre(genre)
	// Clear cached params when genre changes
	b.params = make(map[int64]*FlickerParams)
}

// CalculateFlickerForLight computes realistic flicker for a lighting.PointLight.
// Returns intensity multiplier and RGB color modulation values.
func (b *LightFlickerBridge) CalculateFlickerForLight(light *lighting.PointLight, tick int) FlickerResult {
	// Get or create params for this light
	params, exists := b.params[light.FlickerSeed]
	if !exists {
		p := b.sys.GetFlickerParams(light.LightType, light.FlickerSeed, light.R, light.G, light.B)
		params = &p
		b.params[light.FlickerSeed] = params
	}

	// Calculate flicker
	intensity, rC, gC, bC := b.sys.CalculateFlicker(params, tick, light.Intensity)

	return FlickerResult{
		Intensity: intensity,
		R:         rC,
		G:         gC,
		B:         bC,
	}
}

// CalculateFlickerForLightComponent computes realistic flicker for a LightComponent.
func (b *LightFlickerBridge) CalculateFlickerForLightComponent(lc *lighting.LightComponent, tick int) FlickerResult {
	return b.CalculateFlickerForLight(&lc.PointLight, tick)
}

// GetParams retrieves cached parameters for a seed, or nil if not cached.
func (b *LightFlickerBridge) GetParams(seed int64) *FlickerParams {
	return b.params[seed]
}

// CalculateFlickerSimple computes flicker for a basic light with given seed and base values.
// This is useful for torch props that aren't full PointLight objects.
func (b *LightFlickerBridge) CalculateFlickerSimple(seed int64, tick int, baseIntensity, baseR, baseG, baseB float64) FlickerResult {
	// Get or create params for this seed
	params, exists := b.params[seed]
	if !exists {
		p := b.sys.GetFlickerParams("torch", seed, baseR, baseG, baseB)
		params = &p
		b.params[seed] = params
	}

	// Calculate flicker using the base intensity
	intensity, rC, gC, bC := b.sys.CalculateFlicker(params, tick, baseIntensity)

	return FlickerResult{
		Intensity: intensity,
		R:         rC,
		G:         gC,
		B:         bC,
	}
}

// ClearCache removes all cached parameters.
func (b *LightFlickerBridge) ClearCache() {
	b.params = make(map[int64]*FlickerParams)
}
