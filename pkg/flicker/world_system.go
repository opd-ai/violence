package flicker

import (
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// WorldSystem processes flicker components in the game world.
type WorldSystem struct {
	flickerSys *System
	tick       int
	logger     *logrus.Entry
}

// NewWorldSystem creates a world-integrated flicker system.
func NewWorldSystem(genre string) *WorldSystem {
	return &WorldSystem{
		flickerSys: NewSystem(genre),
		tick:       0,
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "flicker",
			"package":     "flicker",
		}),
	}
}

// Update processes all entities with flicker components.
func (ws *WorldSystem) Update(w *engine.World) {
	ws.tick++

	compType := reflect.TypeOf(&Component{})
	entities := w.Query(compType)

	for _, entity := range entities {
		compRaw, found := w.GetComponent(entity, compType)
		if !found {
			continue
		}

		comp, ok := compRaw.(*Component)
		if !ok || !comp.Enabled {
			continue
		}

		// Calculate flicker values
		intensity, r, g, b := ws.flickerSys.CalculateFlicker(&comp.Params, ws.tick, 1.0)

		// Update component state
		comp.CurrentIntensity = intensity
		comp.CurrentR = r
		comp.CurrentG = g
		comp.CurrentB = b
	}
}

// SetGenre changes the flicker genre and reinitializes presets.
func (ws *WorldSystem) SetGenre(genre string) {
	ws.flickerSys.SetGenre(genre)
	ws.logger.WithField("genre", genre).Info("Flicker genre changed")
}

// GetFlickerSystem returns the underlying flicker calculation system.
func (ws *WorldSystem) GetFlickerSystem() *System {
	return ws.flickerSys
}

// GetTick returns the current tick count.
func (ws *WorldSystem) GetTick() int {
	return ws.tick
}

// InitializeComponent sets up a flicker component with proper parameters.
func (ws *WorldSystem) InitializeComponent(comp *Component, baseR, baseG, baseB float64) {
	comp.Initialize(ws.flickerSys, baseR, baseG, baseB)
}
