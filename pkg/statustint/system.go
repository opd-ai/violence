package statustint

import (
	"image/color"
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/status"
	"github.com/sirupsen/logrus"
)

// System computes visual tinting parameters based on active status effects.
type System struct {
	genreID  string
	profiles map[string]EffectTintProfile
	logger   *logrus.Entry
	time     float64
}

// NewSystem creates a new status tinting system for the given genre.
func NewSystem(genreID string) *System {
	return &System{
		genreID:  genreID,
		profiles: DefaultEffectProfiles(genreID),
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "statustint",
			"package":     "statustint",
		}),
		time: 0,
	}
}

// SetGenre updates the genre and reloads effect profiles.
func (s *System) SetGenre(genreID string) {
	if s.genreID != genreID {
		s.genreID = genreID
		s.profiles = DefaultEffectProfiles(genreID)
		s.logger.WithField("genre", genreID).Debug("Genre updated, profiles reloaded")
	}
}

// Update processes all entities with StatusComponent and computes TintComponent values.
func (s *System) Update(w *engine.World) {
	if w == nil {
		return
	}

	const deltaTime = 1.0 / 60.0
	s.time += deltaTime

	statusType := reflect.TypeOf(&status.StatusComponent{})
	tintType := reflect.TypeOf(&TintComponent{})

	entities := w.Query(statusType)

	for _, entity := range entities {
		statusComp, hasStatus := w.GetComponent(entity, statusType)
		if !hasStatus {
			s.removeTintComponent(w, entity, tintType)
			continue
		}

		sc := statusComp.(*status.StatusComponent)
		if len(sc.ActiveEffects) == 0 {
			s.removeTintComponent(w, entity, tintType)
			continue
		}

		s.updateTintComponent(w, entity, sc, tintType)
	}

	s.cleanOrphanedTints(w, statusType, tintType)
}

// removeTintComponent removes the tint component if present.
func (s *System) removeTintComponent(w *engine.World, entity engine.Entity, tintType reflect.Type) {
	if w.HasComponent(entity, tintType) {
		w.RemoveComponent(entity, tintType)
	}
}

// updateTintComponent computes aggregate tint from all active effects.
func (s *System) updateTintComponent(w *engine.World, entity engine.Entity, sc *status.StatusComponent, tintType reflect.Type) {
	tintComp, hasTint := w.GetComponent(entity, tintType)
	var tc *TintComponent

	if !hasTint {
		tc = &TintComponent{}
		tc.Reset()
		w.AddComponent(entity, tc)
	} else {
		tc = tintComp.(*TintComponent)
	}

	s.computeAggregateTint(tc, sc.ActiveEffects)
}

// computeAggregateTint blends tint parameters from all active effects.
func (s *System) computeAggregateTint(tc *TintComponent, effects []status.ActiveEffect) {
	tc.Reset()

	if len(effects) == 0 {
		return
	}

	// Track totals for weighted averaging
	var totalWeight float64
	var totalR, totalG, totalB float64
	var totalEdgeR, totalEdgeG, totalEdgeB float64
	var dominantEffect string
	var dominantStrength float64

	for _, effect := range effects {
		profile, exists := s.profiles[effect.EffectName]
		if !exists {
			continue
		}

		// Calculate effect intensity based on time remaining
		intensity := s.calculateIntensity(effect.TimeRemaining.Seconds())

		// Weight by tint strength and intensity
		weight := profile.TintStrength * intensity
		if weight <= 0.01 {
			continue
		}

		// Track dominant effect
		if weight > dominantStrength {
			dominantStrength = weight
			dominantEffect = effect.EffectName
		}

		// Accumulate weighted tint colors
		totalR += float64(profile.TintColor.R) * weight
		totalG += float64(profile.TintColor.G) * weight
		totalB += float64(profile.TintColor.B) * weight

		// Accumulate weighted edge glow
		edgeWeight := profile.EdgeGlowStrength * intensity
		totalEdgeR += float64(profile.EdgeGlowColor.R) * edgeWeight
		totalEdgeG += float64(profile.EdgeGlowColor.G) * edgeWeight
		totalEdgeB += float64(profile.EdgeGlowColor.B) * edgeWeight

		// Accumulate modifiers (additive)
		tc.Saturation += profile.SaturationShift * intensity
		tc.Brightness += profile.BrightnessShift * intensity
		tc.EdgeGlowIntensity += profile.EdgeGlowStrength * intensity
		tc.NoiseIntensity += profile.NoiseStrength * intensity

		// Use highest pulse amplitude and frequency from active effects
		if profile.PulseAmplitude > tc.PulseAmplitude {
			tc.PulseAmplitude = profile.PulseAmplitude
			tc.PulseFrequency = profile.PulseFrequency
		}

		totalWeight += weight
	}

	// Normalize weighted averages
	if totalWeight > 0.01 {
		tc.TintColor = color.RGBA{
			R: uint8(clamp(totalR/totalWeight, 0, 255)),
			G: uint8(clamp(totalG/totalWeight, 0, 255)),
			B: uint8(clamp(totalB/totalWeight, 0, 255)),
			A: 255,
		}
		tc.Intensity = clamp(totalWeight, 0, 1)
	}

	// Normalize edge glow
	if tc.EdgeGlowIntensity > 0.01 {
		tc.EdgeGlowColor = color.RGBA{
			R: uint8(clamp(totalEdgeR/tc.EdgeGlowIntensity, 0, 255)),
			G: uint8(clamp(totalEdgeG/tc.EdgeGlowIntensity, 0, 255)),
			B: uint8(clamp(totalEdgeB/tc.EdgeGlowIntensity, 0, 255)),
			A: uint8(clamp(tc.EdgeGlowIntensity*255, 0, 255)),
		}
		tc.EdgeGlowIntensity = clamp(tc.EdgeGlowIntensity, 0, 1)
	}

	// Clamp cumulative modifiers
	tc.Saturation = clamp(tc.Saturation, -1, 1)
	tc.Brightness = clamp(tc.Brightness, -1, 1)
	tc.NoiseIntensity = clamp(tc.NoiseIntensity, 0, 1)

	// Update pulse phase
	tc.PulsePhase = math.Mod(s.time*tc.PulseFrequency*2*math.Pi, 2*math.Pi)

	// Update noise phase
	tc.NoisePhase = s.time

	tc.DominantEffect = dominantEffect
	tc.Dirty = false
}

// calculateIntensity computes tint intensity based on remaining effect duration.
func (s *System) calculateIntensity(timeRemaining float64) float64 {
	// Full intensity for most of duration, fade in last 2 seconds
	if timeRemaining > 2.0 {
		return 1.0
	}
	if timeRemaining <= 0 {
		return 0
	}
	return timeRemaining / 2.0
}

// cleanOrphanedTints removes tint components from entities without status effects.
func (s *System) cleanOrphanedTints(w *engine.World, statusType, tintType reflect.Type) {
	tintEntities := w.Query(tintType)
	for _, entity := range tintEntities {
		if !w.HasComponent(entity, statusType) {
			w.RemoveComponent(entity, tintType)
		}
	}
}

// ApplyTintToColor applies the tint parameters to a source color.
// This is the core function used by sprite rendering to modify pixel colors.
func ApplyTintToColor(src color.RGBA, tint *TintComponent) color.RGBA {
	if tint == nil || !tint.HasTint() {
		return src
	}

	// Convert to float for calculations
	r := float64(src.R)
	g := float64(src.G)
	b := float64(src.B)
	a := float64(src.A)

	// Apply saturation modifier
	if tint.Saturation != 0 {
		gray := 0.299*r + 0.587*g + 0.114*b
		satFactor := 1.0 + tint.Saturation
		r = gray + (r-gray)*satFactor
		g = gray + (g-gray)*satFactor
		b = gray + (b-gray)*satFactor
	}

	// Apply brightness modifier
	if tint.Brightness != 0 {
		brightAdjust := tint.Brightness * 255
		r += brightAdjust
		g += brightAdjust
		b += brightAdjust
	}

	// Apply contrast modifier
	if tint.Contrast != 1.0 {
		r = ((r/255.0-0.5)*tint.Contrast + 0.5) * 255
		g = ((g/255.0-0.5)*tint.Contrast + 0.5) * 255
		b = ((b/255.0-0.5)*tint.Contrast + 0.5) * 255
	}

	// Apply tint color blend
	if tint.Intensity > 0.01 {
		// Calculate pulsing intensity
		pulseIntensity := tint.Intensity
		if tint.PulseAmplitude > 0 {
			pulse := math.Sin(tint.PulsePhase) * tint.PulseAmplitude
			pulseIntensity = clamp(tint.Intensity*(1.0+pulse), 0, 1)
		}

		// Blend with tint color using overlay blending
		r = blendOverlay(r, float64(tint.TintColor.R), pulseIntensity)
		g = blendOverlay(g, float64(tint.TintColor.G), pulseIntensity)
		b = blendOverlay(b, float64(tint.TintColor.B), pulseIntensity)
	}

	// Clamp final values
	r = clamp(r, 0, 255)
	g = clamp(g, 0, 255)
	b = clamp(b, 0, 255)

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: uint8(a),
	}
}

// blendOverlay performs overlay blending between base and blend colors.
// This preserves highlights and shadows while applying the tint.
func blendOverlay(base, blend, amount float64) float64 {
	// Normalize to 0-1 range
	baseN := base / 255.0
	blendN := blend / 255.0

	var result float64
	if baseN < 0.5 {
		result = 2 * baseN * blendN
	} else {
		result = 1 - 2*(1-baseN)*(1-blendN)
	}

	// Lerp between original and blended based on amount
	result = baseN*(1-amount) + result*amount

	return result * 255.0
}

// GetEdgeGlowPixel calculates edge glow contribution for a pixel.
// distFromEdge: 0.0 = at edge, 1.0 = center of sprite
func GetEdgeGlowPixel(tint *TintComponent, distFromEdge float64) color.RGBA {
	if tint == nil || tint.EdgeGlowIntensity < 0.01 {
		return color.RGBA{0, 0, 0, 0}
	}

	// Edge glow fades from edge (full) to center (none)
	edgeFalloff := 1.0 - distFromEdge
	edgeFalloff = edgeFalloff * edgeFalloff // Quadratic falloff

	// Apply pulse to edge glow
	pulse := 1.0
	if tint.PulseAmplitude > 0 {
		pulse = 1.0 + math.Sin(tint.PulsePhase)*tint.PulseAmplitude*0.5
	}

	alpha := tint.EdgeGlowIntensity * edgeFalloff * pulse
	alpha = clamp(alpha, 0, 1)

	return color.RGBA{
		R: tint.EdgeGlowColor.R,
		G: tint.EdgeGlowColor.G,
		B: tint.EdgeGlowColor.B,
		A: uint8(alpha * 255),
	}
}

// GetNoiseOffset returns a noise-based offset for glitch/radiation effects.
// Returns offset in range [-1, 1] for x and y.
func GetNoiseOffset(tint *TintComponent, x, y int) (float64, float64) {
	if tint == nil || tint.NoiseIntensity < 0.01 {
		return 0, 0
	}

	// Simple deterministic noise based on position and time
	seed := float64(x*73 + y*97)
	noiseX := math.Sin(seed+tint.NoisePhase*10) * tint.NoiseIntensity
	noiseY := math.Cos(seed*1.3+tint.NoisePhase*8) * tint.NoiseIntensity

	return noiseX, noiseY
}

// ShouldApplyNoiseColor returns true if a pixel should have noise color applied.
func ShouldApplyNoiseColor(tint *TintComponent, x, y int) bool {
	if tint == nil || tint.NoiseIntensity < 0.1 {
		return false
	}

	// Deterministic but time-varying noise pattern
	seed := float64(x*37+y*59) + tint.NoisePhase*5
	threshold := math.Sin(seed) * 0.5 * tint.NoiseIntensity
	return threshold > 0.4
}

// clamp restricts a value to the given range.
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
