// Package decal implements persistent combat decal rendering and management.
package decal

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// System manages persistent combat decals.
type System struct {
	generator *Generator
	maxDecals int
	genreID   string
	logger    *logrus.Entry
	rng       *rand.Rand
}

// NewSystem creates a decal management system.
func NewSystem(maxDecals int, genreID string, seed int64) *System {
	return &System{
		generator: NewGenerator(200), // Cache up to 200 unique decal images
		maxDecals: maxDecals,
		genreID:   genreID,
		logger: logrus.WithFields(logrus.Fields{
			"system": "decal",
		}),
		rng: rand.New(rand.NewSource(seed)),
	}
}

// SetGenre updates the genre for decal generation.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	s.generator.SetGenre(genreID)
}

// Update fades and removes old decals.
func (s *System) Update(entities interface{}, deltaTime float64) {
	// Type assert to expected entity slice type
	// For now, we'll handle decals at game level, not ECS entities
	// This is a stub to satisfy the system interface
}

// UpdateDecals fades decals over time and removes fully faded ones.
func (s *System) UpdateDecals(decals *[]Decal, deltaTime float64) {
	if decals == nil {
		return
	}

	remaining := (*decals)[:0]
	for i := range *decals {
		d := &(*decals)[i]
		d.Age += deltaTime

		// Update opacity based on age
		if d.Age < d.MaxAge {
			d.Opacity = 1.0 - (d.Age / d.MaxAge)
			if d.Opacity > 0.05 { // Keep if still visible
				remaining = append(remaining, *d)
			}
		}
	}
	*decals = remaining
}

// SpawnDecal adds a new decal to the list.
func (s *System) SpawnDecal(decals *[]Decal, x, y float64, decalType DecalType, angle float64, layer int) {
	if decals == nil {
		return
	}

	// Limit total decals
	if len(*decals) >= s.maxDecals {
		// Remove oldest
		*decals = (*decals)[1:]
	}

	// Determine max age based on type
	maxAge := 30.0 // Default 30 seconds
	switch decalType {
	case DecalBlood:
		maxAge = 60.0 // Blood lasts longer
	case DecalScorch:
		maxAge = 45.0
	case DecalSlash:
		maxAge = 40.0
	case DecalBulletHole:
		maxAge = 50.0
	case DecalMagicBurn:
		maxAge = 35.0
	case DecalAcid:
		maxAge = 25.0
	case DecalFreeze:
		maxAge = 20.0
	case DecalElectric:
		maxAge = 15.0
	}

	// Create decal
	decal := Decal{
		X:       x,
		Y:       y,
		Type:    decalType,
		Subtype: int(s.rng.Int63n(4)), // 4 variations per type
		Seed:    s.rng.Int63(),
		Angle:   angle,
		Scale:   0.8 + s.rng.Float64()*0.4, // 0.8-1.2
		Opacity: 1.0,
		Age:     0,
		MaxAge:  maxAge,
		Layer:   layer,
		GenreID: s.genreID,
	}

	*decals = append(*decals, decal)
}

// RenderDecals draws all decals in screen space.
func (s *System) RenderDecals(screen *ebiten.Image, decals []Decal, cameraX, cameraY float64) {
	const tileSize = 64.0
	screenW := float64(screen.Bounds().Dx())
	screenH := float64(screen.Bounds().Dy())

	for i := range decals {
		d := &decals[i]

		// Only render floor decals for now (layer 0)
		if d.Layer != 0 {
			continue
		}

		// World to screen projection (top-down)
		dx := d.X - cameraX
		dy := d.Y - cameraY

		screenX := screenW/2 + dx*tileSize
		screenY := screenH/2 + dy*tileSize

		// Frustum culling
		const margin = 100.0
		if screenX < -margin || screenX > screenW+margin || screenY < -margin || screenY > screenH+margin {
			continue
		}

		// Get decal image
		img := s.generator.GetDecal(d.Type, d.Subtype, d.Seed, 32)
		if img == nil {
			continue
		}

		// Draw with transformations
		opts := &ebiten.DrawImageOptions{}

		// Center pivot
		w, h := img.Size()
		opts.GeoM.Translate(-float64(w)/2, -float64(h)/2)

		// Scale
		opts.GeoM.Scale(d.Scale, d.Scale)

		// Rotate
		opts.GeoM.Rotate(d.Angle)

		// Translate to position
		opts.GeoM.Translate(screenX, screenY)

		// Apply opacity
		opts.ColorM.Scale(1, 1, 1, d.Opacity)

		screen.DrawImage(img, opts)
	}
}

// SpawnBloodSplatter creates blood splatter decals from combat.
func (s *System) SpawnBloodSplatter(decals *[]Decal, x, y, dirX, dirY float64) {
	if decals == nil {
		return
	}

	// Main splatter
	angle := math.Atan2(dirY, dirX)
	s.SpawnDecal(decals, x, y, DecalBlood, angle, 0)

	// Additional droplets
	numDrops := 2 + int(s.rng.Int63n(3))
	for i := 0; i < numDrops; i++ {
		offsetDist := 0.2 + s.rng.Float64()*0.3
		offsetAngle := angle + (s.rng.Float64()-0.5)*math.Pi*0.5
		dropX := x + math.Cos(offsetAngle)*offsetDist
		dropY := y + math.Sin(offsetAngle)*offsetDist
		s.SpawnDecal(decals, dropX, dropY, DecalBlood, s.rng.Float64()*2*math.Pi, 0)
	}
}

// SpawnImpactMark creates impact-specific decals.
func (s *System) SpawnImpactMark(decals *[]Decal, x, y, dirX, dirY float64, damageType string) {
	if decals == nil {
		return
	}

	angle := math.Atan2(dirY, dirX)

	var decalType DecalType
	switch damageType {
	case "fire", "explosion":
		decalType = DecalScorch
	case "slash", "pierce":
		decalType = DecalSlash
	case "projectile", "ballistic":
		decalType = DecalBulletHole
	case "magic", "arcane":
		decalType = DecalMagicBurn
	case "acid", "poison":
		decalType = DecalAcid
	case "ice", "frost":
		decalType = DecalFreeze
	case "lightning", "electric":
		decalType = DecalElectric
	default:
		decalType = DecalBlood
	}

	s.SpawnDecal(decals, x, y, decalType, angle, 0)
}
