package spritebatch

import (
	"image"
	"sort"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultBatchCapacity is the initial capacity for batch items per layer.
	DefaultBatchCapacity = 256
	// MaxBatchSize is the maximum items before auto-flush.
	MaxBatchSize = 1024
)

// System provides batched sprite rendering for improved GPU efficiency.
type System struct {
	logger  *logrus.Entry
	batches [LayerUI + 1][]BatchItem
	active  bool
	mu      sync.Mutex

	// Statistics for performance monitoring
	stats BatchStats

	// Pool for DrawImageOptions to reduce allocations
	optsPool sync.Pool
}

// BatchStats tracks rendering performance metrics.
type BatchStats struct {
	// ItemsQueued is the number of items queued this frame.
	ItemsQueued int
	// DrawCalls is the number of draw calls issued this frame.
	DrawCalls int
	// BatchesProcessed is the number of batches processed.
	BatchesProcessed int
}

// NewSystem creates a sprite batch rendering system.
func NewSystem() *System {
	s := &System{
		logger: logrus.WithFields(logrus.Fields{
			"system":  "spritebatch",
			"package": "spritebatch",
		}),
		active: false,
	}

	// Initialize batch slices for each layer
	for i := range s.batches {
		s.batches[i] = make([]BatchItem, 0, DefaultBatchCapacity)
	}

	// Initialize options pool
	s.optsPool = sync.Pool{
		New: func() interface{} {
			return &ebiten.DrawImageOptions{}
		},
	}

	return s
}

// Begin starts a new batch frame. Call at the beginning of rendering.
func (s *System) Begin() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active {
		s.logger.Warn("Begin called while batch already active")
	}

	// Clear previous frame's batches
	for i := range s.batches {
		s.batches[i] = s.batches[i][:0]
	}

	// Reset statistics
	s.stats = BatchStats{}
	s.active = true
}

// Queue adds a sprite to the batch for deferred rendering.
// Thread-safe: can be called from multiple goroutines.
func (s *System) Queue(item BatchItem) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		s.logger.Warn("Queue called without active batch frame")
		return
	}

	if item.Source == nil {
		return
	}

	// Validate layer
	if item.Layer < LayerFloor || item.Layer > LayerUI {
		item.Layer = LayerEntity
	}

	s.batches[item.Layer] = append(s.batches[item.Layer], item)
	s.stats.ItemsQueued++
}

// QueueSimple adds a sprite with minimal parameters.
func (s *System) QueueSimple(source *ebiten.Image, x, y float64, layer Layer, zOrder float64) {
	s.Queue(NewBatchItem(source, x, y).WithLayer(layer).WithZOrder(zOrder))
}

// QueueWithTransform adds a sprite with transform parameters.
func (s *System) QueueWithTransform(source *ebiten.Image, x, y, scaleX, scaleY, rotation float64, layer Layer, zOrder float64) {
	item := NewBatchItem(source, x, y).
		WithScale(scaleX, scaleY).
		WithRotation(rotation).
		WithLayer(layer).
		WithZOrder(zOrder)
	s.Queue(item)
}

// QueueWithColor adds a sprite with color multiply.
func (s *System) QueueWithColor(source *ebiten.Image, x, y, r, g, b, alpha float64, layer Layer, zOrder float64) {
	item := NewBatchItem(source, x, y).
		WithColor(r, g, b).
		WithAlpha(alpha).
		WithLayer(layer).
		WithZOrder(zOrder)
	s.Queue(item)
}

// End flushes all batched sprites to the screen. Call at the end of rendering.
func (s *System) End(screen *ebiten.Image) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		s.logger.Warn("End called without active batch frame")
		return
	}

	// Process each layer in order
	for layer := LayerFloor; layer <= LayerUI; layer++ {
		batch := s.batches[layer]
		if len(batch) == 0 {
			continue
		}

		// Sort by z-order within layer
		sort.SliceStable(batch, func(i, j int) bool {
			return batch[i].ZOrder < batch[j].ZOrder
		})

		// Group by source texture for efficient batching
		s.renderBatch(screen, batch)
		s.stats.BatchesProcessed++
	}

	s.active = false
}

// renderBatch renders a slice of batch items, grouping by texture.
func (s *System) renderBatch(screen *ebiten.Image, items []BatchItem) {
	if len(items) == 0 {
		return
	}

	// Group items by source image for better batching
	groups := make(map[*ebiten.Image][]BatchItem)
	for _, item := range items {
		groups[item.Source] = append(groups[item.Source], item)
	}

	// Render each texture group
	for _, group := range groups {
		s.renderGroup(screen, group)
	}
}

// renderGroup renders a group of items sharing the same source texture.
func (s *System) renderGroup(screen *ebiten.Image, items []BatchItem) {
	for _, item := range items {
		s.renderItem(screen, item)
	}
}

// renderItem renders a single batch item.
func (s *System) renderItem(screen *ebiten.Image, item BatchItem) {
	opts := s.optsPool.Get().(*ebiten.DrawImageOptions)
	defer s.optsPool.Put(opts)

	// Reset options
	opts.GeoM.Reset()
	opts.ColorScale.Reset()

	// Get source bounds
	bounds := item.Source.Bounds()
	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())

	if item.SrcWidth > 0 && item.SrcHeight > 0 {
		srcW = float64(item.SrcWidth)
		srcH = float64(item.SrcHeight)
	}

	// Apply origin offset (move origin to center for rotation/scale)
	originX := srcW * item.OriginX
	originY := srcH * item.OriginY
	opts.GeoM.Translate(-originX, -originY)

	// Apply scale
	opts.GeoM.Scale(item.ScaleX, item.ScaleY)

	// Apply rotation
	if item.Rotation != 0 {
		opts.GeoM.Rotate(item.Rotation)
	}

	// Apply final position
	opts.GeoM.Translate(item.DstX, item.DstY)

	// Apply color multiply
	opts.ColorScale.Scale(float32(item.ColorR), float32(item.ColorG), float32(item.ColorB), float32(item.Alpha))

	// Draw with sub-image if source rect specified
	if item.SrcWidth > 0 && item.SrcHeight > 0 {
		subImg := item.Source.SubImage(image.Rect(
			item.SrcX, item.SrcY,
			item.SrcX+item.SrcWidth, item.SrcY+item.SrcHeight,
		)).(*ebiten.Image)
		screen.DrawImage(subImg, opts)
	} else {
		screen.DrawImage(item.Source, opts)
	}

	s.stats.DrawCalls++
}

// GetStats returns the current batch statistics.
func (s *System) GetStats() BatchStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stats
}

// IsActive returns whether a batch frame is currently active.
func (s *System) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

// GetQueuedCount returns the number of items queued in the current frame.
func (s *System) GetQueuedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	total := 0
	for _, batch := range s.batches {
		total += len(batch)
	}
	return total
}

// Update implements the System interface for ECS integration.
func (s *System) Update(world interface{}) {
	// SpriteBatch is render-only, no update logic needed
}
