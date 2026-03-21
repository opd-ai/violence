package proximityui

import (
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/common"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// Config holds distance thresholds for detail levels.
type Config struct {
	// AdjacentDistance is the max distance for full detail (0 to this value).
	AdjacentDistance float64

	// NearDistance is the max distance for moderate detail.
	NearDistance float64

	// MidDistance is the max distance for minimal detail.
	MidDistance float64

	// FarDistance is the max distance for any UI (beyond this = none).
	FarDistance float64

	// TransitionSpeed is how fast detail levels fade (units per second).
	TransitionSpeed float64

	// DistanceFadeStart is where opacity starts decreasing within a tier.
	DistanceFadeStart float64

	// DistanceFadeEnd is where opacity reaches minimum within a tier.
	DistanceFadeEnd float64
}

// System manages proximity-based UI detail levels for all entities.
type System struct {
	genre  string
	config Config
	logger *logrus.Entry

	// Camera position for distance calculations
	cameraX float64
	cameraY float64

	// Targeted entity (if any)
	targetedEntity engine.Entity
	hasTarget      bool
}

// NewSystem creates a proximity UI system with genre-specific defaults.
func NewSystem(genre string) *System {
	s := &System{
		genre: genre,
		logger: logrus.WithFields(logrus.Fields{
			"system":  "proximityui",
			"package": "proximityui",
		}),
	}
	s.applyGenreConfig()
	return s
}

// SetGenre updates genre-specific thresholds.
func (s *System) SetGenre(genre string) {
	s.genre = genre
	s.applyGenreConfig()
}

// applyGenreConfig sets distance thresholds based on genre.
func (s *System) applyGenreConfig() {
	switch s.genre {
	case "horror":
		// Horror: short visibility ranges to maintain tension
		s.config = Config{
			AdjacentDistance:  2.0,
			NearDistance:      5.0,
			MidDistance:       10.0,
			FarDistance:       15.0,
			TransitionSpeed:   3.0,
			DistanceFadeStart: 0.6,
			DistanceFadeEnd:   0.9,
		}
	case "scifi":
		// Sci-fi: longer ranges for open environments
		s.config = Config{
			AdjacentDistance:  4.0,
			NearDistance:      10.0,
			MidDistance:       18.0,
			FarDistance:       25.0,
			TransitionSpeed:   4.0,
			DistanceFadeStart: 0.7,
			DistanceFadeEnd:   0.95,
		}
	case "cyberpunk":
		// Cyberpunk: moderate ranges, neon visibility
		s.config = Config{
			AdjacentDistance:  3.5,
			NearDistance:      9.0,
			MidDistance:       16.0,
			FarDistance:       22.0,
			TransitionSpeed:   4.5,
			DistanceFadeStart: 0.65,
			DistanceFadeEnd:   0.9,
		}
	case "postapoc":
		// Post-apocalyptic: moderate ranges, dust/debris
		s.config = Config{
			AdjacentDistance:  3.0,
			NearDistance:      8.0,
			MidDistance:       14.0,
			FarDistance:       20.0,
			TransitionSpeed:   3.5,
			DistanceFadeStart: 0.6,
			DistanceFadeEnd:   0.85,
		}
	default: // fantasy
		// Fantasy: standard ranges
		s.config = Config{
			AdjacentDistance:  3.0,
			NearDistance:      8.0,
			MidDistance:       15.0,
			FarDistance:       20.0,
			TransitionSpeed:   4.0,
			DistanceFadeStart: 0.65,
			DistanceFadeEnd:   0.9,
		}
	}
}

// SetCameraPosition updates the camera position for distance calculations.
func (s *System) SetCameraPosition(x, y float64) {
	s.cameraX = x
	s.cameraY = y
}

// SetTargetedEntity marks an entity as targeted (always full detail).
func (s *System) SetTargetedEntity(entity engine.Entity) {
	s.targetedEntity = entity
	s.hasTarget = true
}

// ClearTargetedEntity removes the current target.
func (s *System) ClearTargetedEntity() {
	s.hasTarget = false
}

// Update processes all entities with proximity UI components.
func (s *System) Update(w *engine.World) {
	deltaTime := common.DeltaTime

	compType := reflect.TypeOf(&Component{})
	posType := reflect.TypeOf(&engine.Position{})

	entities := w.Query(compType)
	for _, eid := range entities {
		comp, ok := w.GetComponent(eid, compType)
		if !ok {
			continue
		}
		proxComp := comp.(*Component)

		// Get entity position
		posComp, hasPos := w.GetComponent(eid, posType)
		if !hasPos {
			continue
		}
		pos := posComp.(*engine.Position)

		// Calculate distance from camera
		dx := pos.X - s.cameraX
		dy := pos.Y - s.cameraY
		distance := math.Sqrt(dx*dx + dy*dy)
		proxComp.LastDistance = distance

		// Check if this entity is currently targeted
		if s.hasTarget && eid == s.targetedEntity {
			proxComp.IsTargeted = true
		} else {
			proxComp.IsTargeted = false
		}

		// Calculate target detail level
		targetLevel := s.calculateDetailLevel(proxComp, distance)
		proxComp.TargetDetailLevel = targetLevel

		// Smooth transition to target level
		s.updateTransition(proxComp, deltaTime)

		// Calculate fade alpha based on distance within tier
		proxComp.FadeAlpha = s.calculateFadeAlpha(proxComp, distance)
	}
}

// calculateDetailLevel determines the appropriate detail level for an entity.
func (s *System) calculateDetailLevel(comp *Component, distance float64) DetailLevel {
	// Priority overrides
	if comp.IsTargeted {
		return DetailFull
	}
	if comp.IsBoss {
		minLevel := DetailModerate
		distLevel := s.distanceToDetailLevel(distance)
		if distLevel > minLevel {
			return distLevel
		}
		return minLevel
	}
	if comp.IsQuestNPC {
		minLevel := DetailModerate
		distLevel := s.distanceToDetailLevel(distance)
		if distLevel > minLevel {
			return distLevel
		}
		return minLevel
	}
	if comp.IsPlayer {
		minLevel := DetailMinimal
		distLevel := s.distanceToDetailLevel(distance)
		if distLevel > minLevel {
			return distLevel
		}
		return minLevel
	}
	if comp.PriorityOverride >= 0 {
		minLevel := comp.PriorityOverride
		distLevel := s.distanceToDetailLevel(distance)
		if distLevel > minLevel {
			return distLevel
		}
		return minLevel
	}

	// Distance-based calculation
	return s.distanceToDetailLevel(distance)
}

// distanceToDetailLevel converts raw distance to detail level.
func (s *System) distanceToDetailLevel(distance float64) DetailLevel {
	if distance <= s.config.AdjacentDistance {
		return DetailFull
	}
	if distance <= s.config.NearDistance {
		return DetailModerate
	}
	if distance <= s.config.MidDistance {
		return DetailMinimal
	}
	return DetailNone
}

// updateTransition smoothly transitions between detail levels.
func (s *System) updateTransition(comp *Component, deltaTime float64) {
	if comp.CurrentDetailLevel == comp.TargetDetailLevel {
		comp.TransitionProgress = 1.0
		return
	}

	// Moving toward target
	comp.TransitionProgress += deltaTime * s.config.TransitionSpeed
	if comp.TransitionProgress >= 1.0 {
		comp.TransitionProgress = 1.0
		comp.CurrentDetailLevel = comp.TargetDetailLevel
	}
}

// calculateFadeAlpha determines opacity based on distance within the current tier.
func (s *System) calculateFadeAlpha(comp *Component, distance float64) float64 {
	// Get the distance range for current detail level
	var tierStart, tierEnd float64
	switch comp.TargetDetailLevel {
	case DetailFull:
		tierStart = 0
		tierEnd = s.config.AdjacentDistance
	case DetailModerate:
		tierStart = s.config.AdjacentDistance
		tierEnd = s.config.NearDistance
	case DetailMinimal:
		tierStart = s.config.NearDistance
		tierEnd = s.config.MidDistance
	default:
		// DetailNone - fade out completely
		tierStart = s.config.MidDistance
		tierEnd = s.config.FarDistance
		if distance >= tierEnd {
			return 0.0
		}
		progress := (distance - tierStart) / (tierEnd - tierStart)
		return 1.0 - progress
	}

	// Calculate position within tier (0 to 1)
	tierRange := tierEnd - tierStart
	if tierRange <= 0 {
		return 1.0
	}

	progress := (distance - tierStart) / tierRange
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	// Apply fade curve (ease-out for smooth transition)
	fadeStart := s.config.DistanceFadeStart
	fadeEnd := s.config.DistanceFadeEnd

	if progress < fadeStart {
		return 1.0
	}
	if progress > fadeEnd {
		fadeProgress := (progress - fadeStart) / (fadeEnd - fadeStart)
		return 1.0 - fadeProgress*0.5 // Max fade to 50% within tier
	}

	fadeProgress := (progress - fadeStart) / (fadeEnd - fadeStart)
	return 1.0 - fadeProgress*0.3 // Partial fade
}

// GetDetailLevel returns the current detail level for an entity.
// This is the primary query method for other UI systems.
func (s *System) GetDetailLevel(w *engine.World, entity engine.Entity) DetailLevel {
	compType := reflect.TypeOf(&Component{})
	comp, ok := w.GetComponent(entity, compType)
	if !ok {
		// No component = default to distance-based
		return s.getDetailLevelByDistance(w, entity)
	}

	proxComp := comp.(*Component)
	return proxComp.CurrentDetailLevel
}

// getDetailLevelByDistance calculates detail level for entities without component.
func (s *System) getDetailLevelByDistance(w *engine.World, entity engine.Entity) DetailLevel {
	posType := reflect.TypeOf(&engine.Position{})
	posComp, ok := w.GetComponent(entity, posType)
	if !ok {
		return DetailNone
	}

	pos := posComp.(*engine.Position)
	dx := pos.X - s.cameraX
	dy := pos.Y - s.cameraY
	distance := math.Sqrt(dx*dx + dy*dy)

	return s.distanceToDetailLevel(distance)
}

// GetFadeAlpha returns the current fade alpha for an entity.
func (s *System) GetFadeAlpha(w *engine.World, entity engine.Entity) float64 {
	compType := reflect.TypeOf(&Component{})
	comp, ok := w.GetComponent(entity, compType)
	if !ok {
		return 1.0
	}

	proxComp := comp.(*Component)
	return proxComp.GetEffectiveAlpha()
}

// ShouldRenderHealthBar checks if health bar should render for entity.
func (s *System) ShouldRenderHealthBar(w *engine.World, entity engine.Entity) bool {
	level := s.GetDetailLevel(w, entity)
	return level >= DetailMinimal
}

// ShouldRenderName checks if name label should render for entity.
func (s *System) ShouldRenderName(w *engine.World, entity engine.Entity) bool {
	level := s.GetDetailLevel(w, entity)
	return level >= DetailModerate
}

// ShouldRenderStatusIcons checks if status icons should render for entity.
func (s *System) ShouldRenderStatusIcons(w *engine.World, entity engine.Entity) bool {
	level := s.GetDetailLevel(w, entity)
	return level >= DetailFull
}

// GetConfig returns the current configuration (for testing/debugging).
func (s *System) GetConfig() Config {
	return s.config
}

// SetConfig allows runtime configuration changes (for settings menu).
func (s *System) SetConfig(cfg Config) {
	s.config = cfg
}
