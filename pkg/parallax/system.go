package parallax

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System manages parallax background rendering.
type System struct {
	logger *logrus.Entry
}

// NewSystem creates a new parallax system.
func NewSystem() *System {
	return &System{
		logger: logrus.WithFields(logrus.Fields{
			"system": "parallax",
		}),
	}
}

// Update processes parallax components (currently no per-frame update needed).
func (s *System) Update(w *engine.World) {
	// No per-frame updates needed - layers are static once generated
}

// Render draws parallax layers behind the main scene.
func (s *System) Render(screen *ebiten.Image, comp *Component) {
	if !comp.Enabled || len(comp.Layers) == 0 {
		return
	}

	// Sort layers by Z-index (render back to front)
	sortedLayers := make([]*Layer, len(comp.Layers))
	copy(sortedLayers, comp.Layers)
	sort.Slice(sortedLayers, func(i, j int) bool {
		return sortedLayers[i].ZIndex < sortedLayers[j].ZIndex
	})

	// Render each layer
	for _, layer := range sortedLayers {
		s.renderLayer(screen, layer, comp.CameraX, comp.CameraY, comp.ViewWidth, comp.ViewHeight)
	}
}

// renderLayer draws a single parallax layer.
func (s *System) renderLayer(screen *ebiten.Image, layer *Layer, cameraX, cameraY float64, viewWidth, viewHeight int) {
	if layer.Image == nil {
		return
	}

	// Calculate parallax offset
	offsetX := -cameraX * layer.ScrollSpeed
	offsetY := -cameraY * layer.ScrollSpeed

	// Add layer-specific offset
	offsetX += layer.OffsetX
	offsetY += layer.OffsetY

	// Apply tiling if enabled
	if layer.RepeatX {
		// Wrap offset to layer width
		for offsetX < 0 {
			offsetX += float64(layer.Width)
		}
		for offsetX >= float64(layer.Width) {
			offsetX -= float64(layer.Width)
		}

		// Render multiple tiles to fill screen
		startX := int(offsetX) - layer.Width
		for x := startX; x < viewWidth+layer.Width; x += layer.Width {
			s.drawLayerTile(screen, layer, float64(x), offsetY)
		}
	} else {
		s.drawLayerTile(screen, layer, offsetX, offsetY)
	}
}

// drawLayerTile draws a single tile of a parallax layer.
func (s *System) drawLayerTile(screen *ebiten.Image, layer *Layer, x, y float64) {
	opts := &ebiten.DrawImageOptions{}

	// Position
	opts.GeoM.Translate(x, y)

	// Apply opacity and tint
	opts.ColorScale.Scale(
		float32(layer.Tint[0]),
		float32(layer.Tint[1]),
		float32(layer.Tint[2]),
		float32(layer.Opacity*layer.Tint[3]),
	)

	screen.DrawImage(layer.Image, opts)
}

// InitializeForWorld creates parallax layers for a world/level.
func (s *System) InitializeForWorld(comp *Component, viewWidth, viewHeight int) {
	if comp == nil {
		return
	}

	// Clear existing layers
	comp.Layers = make([]*Layer, 0, 4)

	// Generate layers based on genre and biome
	layers := GenerateLayers(comp.GenreID, comp.BiomeID, comp.Seed, viewWidth*2, viewHeight)

	// Add generated layers to component
	for _, layer := range layers {
		comp.AddLayer(layer)
	}

	s.logger.WithFields(logrus.Fields{
		"genre":       comp.GenreID,
		"biome":       comp.BiomeID,
		"layer_count": len(comp.Layers),
	}).Debug("Initialized parallax layers")
}
