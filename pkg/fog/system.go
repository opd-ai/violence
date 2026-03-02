// Package fog - System implementation for atmospheric fog rendering.
package fog

import (
	"image/color"
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System manages atmospheric fog effects for depth cueing and genre atmosphere.
type System struct {
	genre       string
	enabled     bool
	fogStart    float64 // Distance where fog begins (tiles)
	fogEnd      float64 // Distance where fog is maximum (tiles)
	fogColor    color.RGBA
	fogDensity  float64 // Overall fog intensity [0.0-1.0]
	falloffType string  // "linear", "exponential", "exponential_squared"
	cameraX     float64
	cameraY     float64
	logger      *logrus.Entry
}

// NewSystem creates an atmospheric fog system.
func NewSystem(genreID string) *System {
	s := &System{
		genre:       genreID,
		enabled:     true,
		fogStart:    8.0,
		fogEnd:      20.0,
		fogColor:    color.RGBA{50, 50, 60, 255},
		fogDensity:  0.6,
		falloffType: "exponential",
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "fog",
			"genre":       genreID,
		}),
	}
	s.applyGenrePreset(genreID)
	return s
}

// applyGenrePreset configures fog based on genre aesthetics.
func (s *System) applyGenrePreset(genreID string) {
	switch genreID {
	case "fantasy":
		// Dungeon atmosphere - moderate fog with warm stone tint
		s.fogStart = 7.0
		s.fogEnd = 18.0
		s.fogColor = color.RGBA{40, 35, 50, 255} // Dark stone purple
		s.fogDensity = 0.65
		s.falloffType = "exponential"
		s.enabled = true

	case "scifi":
		// Clean space station - minimal fog, cool blue tint
		s.fogStart = 12.0
		s.fogEnd = 30.0
		s.fogColor = color.RGBA{20, 30, 50, 255} // Deep space blue
		s.fogDensity = 0.4
		s.falloffType = "linear"
		s.enabled = true

	case "horror":
		// Heavy atmospheric dread - thick fog, close range
		s.fogStart = 4.0
		s.fogEnd = 12.0
		s.fogColor = color.RGBA{25, 20, 30, 255} // Deep shadow purple-black
		s.fogDensity = 0.85
		s.falloffType = "exponential_squared"
		s.enabled = true

	case "cyberpunk":
		// Neon-lit smog - moderate fog with purple-pink tint
		s.fogStart = 8.0
		s.fogEnd = 22.0
		s.fogColor = color.RGBA{60, 30, 70, 255} // Neon purple haze
		s.fogDensity = 0.7
		s.falloffType = "exponential"
		s.enabled = true

	case "postapoc":
		// Nuclear fallout - thick radioactive haze
		s.fogStart = 6.0
		s.fogEnd = 16.0
		s.fogColor = color.RGBA{55, 50, 35, 255} // Sickly yellow-brown
		s.fogDensity = 0.75
		s.falloffType = "exponential"
		s.enabled = true

	default:
		s.logger.Warnf("unknown genre %s, using fantasy defaults", genreID)
		s.fogStart = 7.0
		s.fogEnd = 18.0
		s.fogColor = color.RGBA{40, 35, 50, 255}
		s.fogDensity = 0.65
		s.falloffType = "exponential"
		s.enabled = true
	}

	s.logger.Debugf("fog preset: start=%.1f, end=%.1f, density=%.2f, falloff=%s",
		s.fogStart, s.fogEnd, s.fogDensity, s.falloffType)
}

// SetGenre updates fog configuration for a new genre.
func (s *System) SetGenre(genreID string) {
	s.genre = genreID
	s.applyGenrePreset(genreID)
}

// SetCamera updates camera position for distance calculations.
func (s *System) SetCamera(x, y float64) {
	s.cameraX = x
	s.cameraY = y
}

// Update computes fog density for all entities based on distance from camera.
func (s *System) Update(world *engine.World) {
	if !s.enabled {
		return
	}

	positionType := reflect.TypeOf(&engine.Position{})
	fogType := reflect.TypeOf(&Component{})

	entities := world.Query(positionType)
	for _, entity := range entities {
		posComp, ok := world.GetComponent(entity, positionType)
		if !ok {
			continue
		}

		position, ok := posComp.(*engine.Position)
		if !ok {
			continue
		}

		// Calculate distance from camera
		dx := position.X - s.cameraX
		dy := position.Y - s.cameraY
		distance := math.Sqrt(dx*dx + dy*dy)

		// Get or create fog component
		fogComp, hasFog := world.GetComponent(entity, fogType)
		if !hasFog {
			fogComp = &Component{}
			world.AddComponent(entity, fogComp)
		}

		fog, ok := fogComp.(*Component)
		if !ok {
			continue
		}

		// Update distance and compute fog density
		fog.DistanceFromCamera = distance
		fog.FogDensity = s.computeFogDensity(distance)
		fog.Visible = fog.FogDensity < 0.95 // Hide if almost fully obscured

		// Compute color tint
		fog.Tint[0] = float64(s.fogColor.R) / 255.0
		fog.Tint[1] = float64(s.fogColor.G) / 255.0
		fog.Tint[2] = float64(s.fogColor.B) / 255.0
	}
}

// computeFogDensity calculates fog density [0.0-1.0] based on distance.
func (s *System) computeFogDensity(distance float64) float64 {
	if distance <= s.fogStart {
		return 0.0
	}
	if distance >= s.fogEnd {
		return s.fogDensity
	}

	// Normalized distance in fog range [0.0-1.0]
	t := (distance - s.fogStart) / (s.fogEnd - s.fogStart)

	var density float64
	switch s.falloffType {
	case "linear":
		density = t

	case "exponential":
		// Exponential falloff: density = 1 - e^(-k*t)
		// k controls curve steepness, using 3.0 for visible curve
		density = 1.0 - math.Exp(-3.0*t)

	case "exponential_squared":
		// Exponential squared: density = 1 - e^(-(k*t)^2)
		// Sharper falloff for heavy fog
		kt := 2.0 * t
		density = 1.0 - math.Exp(-kt*kt)

	default:
		density = t
	}

	return density * s.fogDensity
}

// GetFogColor returns the current fog color.
func (s *System) GetFogColor() color.RGBA {
	return s.fogColor
}

// SetEnabled toggles fog rendering.
func (s *System) SetEnabled(enabled bool) {
	s.enabled = enabled
}

// IsEnabled returns whether fog is currently active.
func (s *System) IsEnabled() bool {
	return s.enabled
}

// SetParameters allows runtime adjustment of fog parameters.
func (s *System) SetParameters(start, end, density float64, fogColor color.RGBA) {
	s.fogStart = start
	s.fogEnd = end
	s.fogDensity = density
	s.fogColor = fogColor
}
