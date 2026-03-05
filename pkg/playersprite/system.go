package playersprite

import (
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System manages player sprite generation and caching.
type System struct {
	generator *Generator
	genreID   string
	logger    *logrus.Entry
}

// NewSystem creates a player sprite rendering system.
func NewSystem(genreID string) *System {
	return &System{
		generator: NewGenerator(genreID),
		genreID:   genreID,
		logger: logrus.WithFields(logrus.Fields{
			"system": "playersprite",
		}),
	}
}

// Update regenerates player sprites when equipment changes.
func (s *System) Update(w *engine.World) {
	componentType := reflect.TypeOf(&Component{})
	entities := w.Query(componentType)

	for _, eid := range entities {
		comp, ok := w.GetComponent(eid, componentType)
		if !ok {
			continue
		}

		playerSprite, ok := comp.(*Component)
		if !ok {
			continue
		}

		// Regenerate sprite if dirty or not cached
		if playerSprite.DirtyFlag || playerSprite.CachedSprite == nil {
			playerSprite.CachedSprite = s.generator.Generate(
				playerSprite.Class,
				playerSprite.Seed,
				playerSprite.AnimState,
				playerSprite.CurrentFrame,
				playerSprite.EquippedWeapon,
				playerSprite.EquippedArmor,
			)
			playerSprite.DirtyFlag = false
		}
	}
}

// SetGenre updates the system's genre configuration.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	s.generator.SetGenre(genreID)
}

// Type returns the system identifier for the engine.
func (s *System) Type() string {
	return "PlayerSpriteSystem"
}
