package corpse

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System manages persistent corpse rendering.
type System struct {
	generator  *Generator
	maxCorpses int
	genreID    string
	logger     *logrus.Entry
	rng        *rand.Rand
}

// NewSystem creates a corpse management system.
func NewSystem(maxCorpses int, genreID string, seed int64) *System {
	return &System{
		generator:  NewGenerator(150),
		maxCorpses: maxCorpses,
		genreID:    genreID,
		logger: logrus.WithFields(logrus.Fields{
			"system": "corpse",
		}),
		rng: rand.New(rand.NewSource(seed)),
	}
}

// SetGenre updates the genre for corpse generation.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	s.generator.SetGenre(genreID)
}

// Update fades and removes old corpses (ECS system interface).
func (s *System) Update(w *engine.World) {
}

// UpdateCorpses fades corpses over time and removes fully faded ones.
func (s *System) UpdateCorpses(corpses *[]Corpse, deltaTime float64) {
	if corpses == nil {
		return
	}

	remaining := (*corpses)[:0]
	for i := range *corpses {
		c := &(*corpses)[i]
		c.Age += deltaTime

		if c.Age < c.MaxAge {
			c.Opacity = 1.0 - (c.Age / c.MaxAge)
			if c.Opacity > 0.05 {
				remaining = append(remaining, *c)
			}
		}
	}
	*corpses = remaining
}

// SpawnCorpse adds a new corpse to the list.
func (s *System) SpawnCorpse(corpses *[]Corpse, x, y float64, seed int64, entityType, subtype string, deathType DeathType, size int, hasLoot bool) {
	if corpses == nil {
		return
	}

	if len(*corpses) >= s.maxCorpses {
		*corpses = (*corpses)[1:]
	}

	maxAge := 30.0
	if deathType == DeathDisintegrate {
		maxAge = 10.0
	} else if deathType == DeathBurn {
		maxAge = 20.0
	}

	corpse := Corpse{
		X:          x,
		Y:          y,
		Seed:       seed,
		EntityType: entityType,
		Subtype:    subtype,
		Angle:      s.rng.Float64() * 6.28,
		Opacity:    1.0,
		Age:        0,
		MaxAge:     maxAge,
		GenreID:    s.genreID,
		Size:       size,
		HasLoot:    hasLoot,
		DeathType:  deathType,
		BloodPool:  deathType == DeathNormal || deathType == DeathSlash,
		Frame:      0,
	}

	*corpses = append(*corpses, corpse)
	s.logger.WithFields(logrus.Fields{
		"x":          x,
		"y":          y,
		"deathType":  deathType,
		"entityType": entityType,
	}).Debug("Spawned corpse")
}

// RenderCorpse renders a single corpse with opacity.
func (s *System) RenderCorpse(screen *ebiten.Image, corpse *Corpse, cameraX, cameraY float64) {
	img := s.generator.GetCorpseImage(corpse.Seed, corpse.EntityType, corpse.DeathType, corpse.Frame, corpse.Size)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(-float64(corpse.Size)/2, -float64(corpse.Size)/2)
	opts.GeoM.Rotate(corpse.Angle)
	opts.GeoM.Translate(corpse.X-cameraX, corpse.Y-cameraY)

	opts.ColorScale.ScaleAlpha(float32(corpse.Opacity))

	screen.DrawImage(img, opts)
}

// RenderAllCorpses renders all corpses in the list.
func (s *System) RenderAllCorpses(screen *ebiten.Image, corpses []Corpse, cameraX, cameraY float64) {
	for i := range corpses {
		s.RenderCorpse(screen, &corpses[i], cameraX, cameraY)
	}
}

// GetCorpseAt checks if there's a corpse with loot at the given position.
func (s *System) GetCorpseAt(corpses []Corpse, x, y, radius float64) *Corpse {
	for i := range corpses {
		c := &corpses[i]
		if !c.HasLoot {
			continue
		}

		dx := c.X - x
		dy := c.Y - y
		distSq := dx*dx + dy*dy

		if distSq < radius*radius {
			return c
		}
	}
	return nil
}

// DetermineDeathType infers death type from damage type or attack type.
func DetermineDeathType(damageType string) DeathType {
	switch damageType {
	case "fire", "burn", "flame":
		return DeathBurn
	case "ice", "freeze", "frost", "cold":
		return DeathFreeze
	case "electric", "lightning", "shock":
		return DeathElectric
	case "acid", "poison", "toxic":
		return DeathAcid
	case "explosion", "explosive", "blast":
		return DeathExplosion
	case "slash", "cut", "blade":
		return DeathSlash
	case "crush", "bludgeon", "blunt":
		return DeathCrush
	case "disintegrate", "vaporize", "holy":
		return DeathDisintegrate
	default:
		return DeathNormal
	}
}
