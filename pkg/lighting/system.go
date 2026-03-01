package lighting

import (
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// LightingSystem manages dynamic lights attached to entities.
// Updates light positions, flicker, pulsing, and lifetime.
type LightingSystem struct {
	genre            string
	tick             int
	ambientR         float64
	ambientG         float64
	ambientB         float64
	ambientIntensity float64
	logger           *logrus.Entry
}

// NewLightingSystem creates a dynamic lighting system.
func NewLightingSystem(genreID string) *LightingSystem {
	s := &LightingSystem{
		genre: genreID,
		tick:  0,
		logger: logrus.WithFields(logrus.Fields{
			"system": "lighting",
			"genre":  genreID,
		}),
	}
	s.applyGenreAmbient(genreID)
	return s
}

// applyGenreAmbient sets genre-specific ambient lighting.
func (s *LightingSystem) applyGenreAmbient(genreID string) {
	switch genreID {
	case "fantasy":
		// Warm, dim dungeon ambient
		s.ambientR = 0.15
		s.ambientG = 0.12
		s.ambientB = 0.08
		s.ambientIntensity = 0.2
	case "scifi":
		// Cool blue ambient from monitors/screens
		s.ambientR = 0.08
		s.ambientG = 0.12
		s.ambientB = 0.18
		s.ambientIntensity = 0.25
	case "horror":
		// Very dim, slightly greenish (decay/mold)
		s.ambientR = 0.05
		s.ambientG = 0.08
		s.ambientB = 0.05
		s.ambientIntensity = 0.15
	case "cyberpunk":
		// Pink-purple neon ambient
		s.ambientR = 0.15
		s.ambientG = 0.08
		s.ambientB = 0.18
		s.ambientIntensity = 0.3
	case "postapoc":
		// Dusty, desaturated daylight
		s.ambientR = 0.18
		s.ambientG = 0.16
		s.ambientB = 0.14
		s.ambientIntensity = 0.28
	default:
		s.ambientR = 0.2
		s.ambientG = 0.2
		s.ambientB = 0.2
		s.ambientIntensity = 0.2
	}
}

// SetGenre updates the lighting system for a new genre.
func (s *LightingSystem) SetGenre(genreID string) {
	s.genre = genreID
	s.applyGenreAmbient(genreID)
}

// GetAmbient returns the current ambient light color and intensity.
func (s *LightingSystem) GetAmbient() (r, g, b, intensity float64) {
	return s.ambientR, s.ambientG, s.ambientB, s.ambientIntensity
}

// Update processes all light components.
// This is called by the ECS World each frame.
func (s *LightingSystem) Update(w *engine.World) {
	s.tick++
	deltaTime := 1.0 / 60.0 // Assume 60 FPS

	lightType := reflect.TypeOf(&LightComponent{})
	positionType := reflect.TypeOf(&PositionComponent{})

	// Query all entities with light components
	entities := w.Query(lightType)

	for _, entity := range entities {
		lightComp, ok := w.GetComponent(entity, lightType)
		if !ok {
			continue
		}

		light, ok := lightComp.(*LightComponent)
		if !ok || !light.Enabled {
			continue
		}

		// Update light age
		light.CurrentAge += deltaTime

		// Check lifetime expiration
		if light.Lifetime > 0 && light.CurrentAge >= light.Lifetime {
			light.Enabled = false
			continue
		}

		// Update position from entity if attached
		if light.AttachedToEntity {
			if posComp, found := w.GetComponent(entity, positionType); found {
				if pos, ok := posComp.(*PositionComponent); ok {
					light.X = pos.X + light.OffsetX
					light.Y = pos.Y + light.OffsetY
				}
			}
		}

		// Update flickering
		if light.IsFlickering {
			// Flicker is already handled by UpdateFlicker method
			// Just ensure it gets called during rendering
		}

		// Update pulsing
		if light.Pulsing {
			light.PulsePhase += deltaTime * light.PulseSpeed * 2.0 * math.Pi
			if light.PulsePhase > 2.0*math.Pi {
				light.PulsePhase -= 2.0 * math.Pi
			}
		}

		// Apply fade in/out based on lifetime
		if light.Lifetime > 0 {
			fadeMultiplier := 1.0

			// Fade in at start
			if light.FadeInDuration > 0 && light.CurrentAge < light.FadeInDuration {
				fadeMultiplier = light.CurrentAge / light.FadeInDuration
			}

			// Fade out at end
			timeRemaining := light.Lifetime - light.CurrentAge
			if light.FadeOutDuration > 0 && timeRemaining < light.FadeOutDuration {
				fadeMultiplier = math.Min(fadeMultiplier, timeRemaining/light.FadeOutDuration)
			}

			// Store base intensity if not already stored
			// Apply fade to current rendering (handled in GetEffectiveIntensity)
			light.Intensity = light.PointLight.Intensity * fadeMultiplier
		}
	}
}

// GetEffectiveIntensity returns the current intensity including flicker and pulse.
func GetEffectiveIntensity(light *LightComponent, tick int) float64 {
	intensity := light.Intensity

	// Apply flicker
	if light.IsFlickering {
		intensity = light.UpdateFlicker(tick)
	}

	// Apply pulse
	if light.Pulsing {
		pulseAmount := (math.Sin(light.PulsePhase) + 1.0) * 0.15 // ±15% variation
		intensity *= (1.0 + pulseAmount)
	}

	return clampF(intensity, 0.0, 1.0)
}

// CollectLights extracts all active lights from the world for rendering.
func (s *LightingSystem) CollectLights(w *engine.World) []Light {
	lightType := reflect.TypeOf(&LightComponent{})
	entities := w.Query(lightType)

	lights := make([]Light, 0, len(entities))

	for _, entity := range entities {
		lightComp, ok := w.GetComponent(entity, lightType)
		if !ok {
			continue
		}

		lc, ok := lightComp.(*LightComponent)
		if !ok || !lc.Enabled {
			continue
		}

		// Get effective intensity with flicker and pulse
		effectiveIntensity := GetEffectiveIntensity(lc, s.tick)

		// Add to render list
		lights = append(lights, Light{
			X:         lc.X,
			Y:         lc.Y,
			Radius:    lc.Radius,
			Intensity: effectiveIntensity,
			R:         lc.R,
			G:         lc.G,
			B:         lc.B,
		})
	}

	return lights
}

// AddLightToEntity is a convenience helper to add a light to an entity.
func (s *LightingSystem) AddLightToEntity(w *engine.World, entity engine.Entity, light *LightComponent) {
	w.AddComponent(entity, light)
}

// PositionComponent represents entity position (defined for reference).
type PositionComponent struct {
	X, Y float64
}

// Type returns the component type identifier.
func (p *PositionComponent) Type() string {
	return "Position"
}
