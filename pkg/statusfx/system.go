package statusfx

import (
	"image/color"
	"math"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/config"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/particle"
	"github.com/opd-ai/violence/pkg/status"
	"github.com/sirupsen/logrus"
)

// System manages visual effects for status conditions on entities.
type System struct {
	genreID         string
	particleSystem  *particle.ParticleSystem
	logger          *logrus.Entry
	pulseTimer      float64
	particleEmitAge float64
}

// NewSystem creates a new status visual effects system.
func NewSystem(genreID string, particleSys *particle.ParticleSystem) *System {
	return &System{
		genreID:        genreID,
		particleSystem: particleSys,
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "statusfx",
		}),
		pulseTimer:      0,
		particleEmitAge: 0,
	}
}

// SetGenre configures genre-specific visual parameters.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update synchronizes visual components with active status effects.
func (s *System) Update(w *engine.World) {
	if w == nil {
		return
	}

	s.pulseTimer += 1.0 / 60.0
	s.particleEmitAge += 1.0 / 60.0

	statusType := reflect.TypeOf(&status.StatusComponent{})
	visualType := reflect.TypeOf(&VisualComponent{})

	entities := w.Query(statusType)
	for _, entity := range entities {
		statusComp, hasStatus := w.GetComponent(entity, statusType)
		if !hasStatus {
			s.removeVisualComponent(w, entity, visualType)
			continue
		}

		sc := statusComp.(*status.StatusComponent)
		if len(sc.ActiveEffects) == 0 {
			s.removeVisualComponent(w, entity, visualType)
			continue
		}

		s.updateOrCreateVisualComponent(w, entity, sc, visualType)
	}

	s.cleanOrphanedVisuals(w, statusType, visualType)
}

// removeVisualComponent removes the visual component if present.
func (s *System) removeVisualComponent(w *engine.World, entity engine.Entity, visualType reflect.Type) {
	if w.HasComponent(entity, visualType) {
		w.RemoveComponent(entity, visualType)
	}
}

// updateOrCreateVisualComponent synchronizes the visual component with active effects.
func (s *System) updateOrCreateVisualComponent(w *engine.World, entity engine.Entity, sc *status.StatusComponent, visualType reflect.Type) {
	visualComp, hasVisual := w.GetComponent(entity, visualType)
	var vc *VisualComponent

	if !hasVisual {
		vc = &VisualComponent{Effects: make([]EffectVisual, 0, len(sc.ActiveEffects))}
		w.AddComponent(entity, vc)
	} else {
		vc = visualComp.(*VisualComponent)
	}

	vc.Effects = vc.Effects[:0]
	for i := range sc.ActiveEffects {
		effect := &sc.ActiveEffects[i]
		intensity := s.calculateIntensity(effect.TimeRemaining.Seconds())
		vc.Effects = append(vc.Effects, EffectVisual{
			Name:        effect.EffectName,
			Color:       effect.VisualColor,
			Intensity:   intensity,
			ParticleAge: s.particleEmitAge,
		})
	}
}

// cleanOrphanedVisuals removes visual components from entities without status effects.
func (s *System) cleanOrphanedVisuals(w *engine.World, statusType, visualType reflect.Type) {
	visualEntities := w.Query(visualType)
	for _, entity := range visualEntities {
		if !w.HasComponent(entity, statusType) {
			w.RemoveComponent(entity, visualType)
		}
	}
}

// calculateIntensity computes pulsating intensity based on remaining time.
func (s *System) calculateIntensity(timeRemaining float64) float64 {
	base := 0.5 + 0.5*math.Sin(s.pulseTimer*3.0)
	fade := math.Min(1.0, timeRemaining/2.0)
	return base * fade
}

// Render draws visual effects for all entities with status effects.
func (s *System) Render(screen *ebiten.Image, w *engine.World, camX, camY float64) {
	if w == nil {
		return
	}

	visualType := reflect.TypeOf(&VisualComponent{})
	posType := reflect.TypeOf(&engine.Position{})

	entities := w.Query(visualType, posType)

	for _, entity := range entities {
		visualComp, hasVisual := w.GetComponent(entity, visualType)
		if !hasVisual {
			continue
		}

		posComp, hasPos := w.GetComponent(entity, posType)
		if !hasPos {
			continue
		}

		vc := visualComp.(*VisualComponent)
		pc := posComp.(*engine.Position)

		s.renderEntityEffects(screen, vc, pc, camX, camY)
	}
}

// renderEntityEffects draws all active visual effects for an entity.
func (s *System) renderEntityEffects(screen *ebiten.Image, vc *VisualComponent, pc *engine.Position, camX, camY float64) {
	dx := pc.X - camX
	dy := pc.Y - camY
	dist := dx*dx + dy*dy

	if dist > 400 {
		return
	}

	screenX := int(float64(config.C.InternalWidth)/2 + dx*64)
	screenY := int(float64(config.C.InternalHeight)/2 + dy*64)

	if screenX < -100 || screenX > config.C.InternalWidth+100 ||
		screenY < -100 || screenY > config.C.InternalHeight+100 {
		return
	}

	for i := range vc.Effects {
		effect := &vc.Effects[i]
		s.renderEffect(screen, effect, screenX, screenY, dist, pc.X, pc.Y)
	}
}

// renderEffect draws a single status effect visual at the entity's position.
func (s *System) renderEffect(screen *ebiten.Image, effect *EffectVisual, screenX, screenY int, dist, worldX, worldY float64) {
	r := uint8((effect.Color >> 24) & 0xFF)
	g := uint8((effect.Color >> 16) & 0xFF)
	b := uint8((effect.Color >> 8) & 0xFF)
	a := uint8(effect.Color & 0xFF)

	intensity := effect.Intensity
	if dist > 250 {
		intensity *= 1.0 - (dist-250)/150
	}

	alpha := float32(a) / 255.0 * float32(intensity)
	if alpha < 0.05 {
		return
	}

	s.renderAura(screen, screenX, screenY, r, g, b, alpha)
	s.emitParticles(effect, worldX, worldY, r, g, b)
}

// renderAura draws a pulsating glow around the entity.
func (s *System) renderAura(screen *ebiten.Image, x, y int, r, g, b uint8, alpha float32) {
	radius := 20.0 + 5.0*math.Sin(s.pulseTimer*4.0)

	for ring := 0; ring < 3; ring++ {
		ringRadius := float32(radius) * (1.0 - float32(ring)*0.3)
		ringAlpha := alpha * (1.0 - float32(ring)*0.4)
		if ringAlpha < 0.05 {
			break
		}

		ringCol := color.RGBA{r, g, b, uint8(float32(80) * ringAlpha)}
		vector.DrawFilledCircle(screen, float32(x), float32(y), ringRadius, ringCol, false)
	}
}

// emitParticles generates status effect particles based on effect type and time.
func (s *System) emitParticles(effect *EffectVisual, x, y float64, r, g, b uint8) {
	if s.particleSystem == nil {
		return
	}

	emitInterval := s.getEmitInterval(effect.Name)
	if math.Mod(effect.ParticleAge, emitInterval) < emitInterval-0.016 {
		return
	}

	count := s.getParticleCount(effect.Name)
	for i := 0; i < count; i++ {
		angle := float64(i) * (2 * math.Pi / float64(count))
		vx := math.Cos(angle) * 0.3
		vy := math.Sin(angle) * 0.3
		vz := 0.1

		s.particleSystem.Spawn(
			x, y, 0.5,
			vx, vy, vz,
			0.5, 0.15,
			color.RGBA{r, g, b, 150},
		)
	}
}

// getEmitInterval returns particle emission interval based on effect type.
func (s *System) getEmitInterval(effectName string) float64 {
	switch effectName {
	case "burning":
		return 0.1
	case "poisoned", "bleeding", "irradiated", "infected":
		return 0.3
	case "stunned", "emp_stunned":
		return 0.2
	case "regeneration", "nanoheal", "blessed":
		return 0.25
	default:
		return 0.4
	}
}

// getParticleCount returns number of particles to emit per cycle.
func (s *System) getParticleCount(effectName string) int {
	switch effectName {
	case "burning":
		return 3
	case "poisoned", "bleeding", "irradiated", "infected":
		return 2
	case "stunned", "emp_stunned":
		return 4
	case "regeneration", "nanoheal", "blessed":
		return 2
	default:
		return 1
	}
}
