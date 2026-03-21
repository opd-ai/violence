package weather

import (
	"reflect"

	"github.com/opd-ai/violence/pkg/common"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System provides ECS integration for the weather system.
type System struct {
	weather *WeatherSystem
	logger  *logrus.Entry
}

// NewSystem creates a weather system for ECS integration.
func NewSystem(maxParticles int, seed int64, genreID string) *System {
	ws := NewWeatherSystem(maxParticles, seed)
	ws.SetGenre(genreID)

	return &System{
		weather: ws,
		logger: logrus.WithFields(logrus.Fields{
			"system": "weather",
			"genre":  genreID,
		}),
	}
}

// Update processes the weather system each frame.
func (s *System) Update(w *engine.World) {
	deltaTime := common.DeltaTime

	// Update weather particles
	s.weather.Update(deltaTime)
}

// GetWeatherSystem returns the underlying weather system for rendering.
func (s *System) GetWeatherSystem() *WeatherSystem {
	return s.weather
}

// SetGenre updates the weather system for a new genre.
func (s *System) SetGenre(genreID string) {
	s.weather.SetGenre(genreID)
	s.logger = logrus.WithFields(logrus.Fields{
		"system": "weather",
		"genre":  genreID,
	})
}

// SetCamera updates camera bounds for particle spawning.
func (s *System) SetCamera(x, y, width, height float64) {
	s.weather.SetCamera(x, y, width, height)
}

// AddWeatherToWorld adds a weather component to the world entity.
func (s *System) AddWeatherToWorld(w *engine.World) {
	// Find or create world entity (entity 0 is typically the world)
	worldEntity := engine.Entity(0)

	comp := &WeatherComponent{
		System: s.weather,
	}

	w.AddComponent(worldEntity, comp)
}

// GetWeatherFromWorld retrieves the weather system from the world.
func GetWeatherFromWorld(w *engine.World) *WeatherSystem {
	weatherType := reflect.TypeOf(&WeatherComponent{})
	entities := w.Query(weatherType)

	if len(entities) == 0 {
		return nil
	}

	comp, ok := w.GetComponent(entities[0], weatherType)
	if !ok {
		return nil
	}

	weatherComp, ok := comp.(*WeatherComponent)
	if !ok {
		return nil
	}

	return weatherComp.System
}
