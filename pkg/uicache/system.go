package uicache

import (
	"hash/fnv"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultMaxEntries is the default maximum cache entries.
	DefaultMaxEntries = 200

	// DefaultPoolSize is the default number of images per pool bucket.
	DefaultPoolSize = 20
)

// System manages UI render caching with dirty-flag tracking.
type System struct {
	logger *logrus.Entry

	// Cache entries indexed by element ID
	entries map[string]*cacheItem

	// Cached images indexed by element ID
	images map[string]*ebiten.Image

	// Image pools by size bucket
	pools map[SizeBucket]*imagePool

	// Screen dimensions for resize detection
	screenWidth  int
	screenHeight int

	// Frame counter for LRU tracking
	frameNumber uint64

	// Configuration
	maxEntries int

	// Statistics
	stats CacheStats

	// Mutex for thread safety
	mu sync.RWMutex
}

// cacheItem holds internal cache state.
type cacheItem struct {
	entry CacheEntry
	image *ebiten.Image
}

// imagePool manages pooled images of a specific size.
type imagePool struct {
	size   SizeBucket
	images []*ebiten.Image
	mu     sync.Mutex
}

// NewSystem creates a UI render caching system.
func NewSystem(screenWidth, screenHeight int) *System {
	s := &System{
		logger: logrus.WithFields(logrus.Fields{
			"system":  "uicache",
			"package": "uicache",
		}),
		entries:      make(map[string]*cacheItem, DefaultMaxEntries),
		images:       make(map[string]*ebiten.Image, DefaultMaxEntries),
		pools:        make(map[SizeBucket]*imagePool),
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		maxEntries:   DefaultMaxEntries,
	}

	// Initialize pools for each size bucket
	buckets := []SizeBucket{Bucket16, Bucket32, Bucket64, Bucket128, Bucket256, BucketLarge}
	for _, bucket := range buckets {
		s.pools[bucket] = &imagePool{
			size:   bucket,
			images: make([]*ebiten.Image, 0, DefaultPoolSize),
		}
	}

	return s
}

// SetScreenSize updates screen dimensions and marks all entries dirty if changed.
func (s *System) SetScreenSize(width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.screenWidth != width || s.screenHeight != height {
		s.screenWidth = width
		s.screenHeight = height

		// Mark all entries dirty on resize
		for _, item := range s.entries {
			item.entry.Dirty = true
		}

		s.logger.WithFields(logrus.Fields{
			"width":    width,
			"height":   height,
			"affected": len(s.entries),
		}).Debug("Screen resize - invalidating cache")
	}
}

// BeginFrame prepares the cache for a new render frame.
func (s *System) BeginFrame() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.frameNumber++

	// Reset per-frame stats
	s.stats.Hits = 0
	s.stats.Misses = 0
	s.stats.Evictions = 0
	s.stats.PoolAllocations = 0
	s.stats.PoolReturns = 0
}

// EndFrame finalizes the frame and performs LRU eviction if needed.
func (s *System) EndFrame() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.TotalEntries = len(s.entries)

	// LRU eviction if over capacity
	if len(s.entries) > s.maxEntries {
		s.evictLRU(len(s.entries) - s.maxEntries)
	}
}

// Get retrieves a cached image if available and clean.
// Returns (image, true) if cache hit, (nil, false) if miss or dirty.
func (s *System) Get(id string) (*ebiten.Image, bool) {
	s.mu.RLock()
	item, exists := s.entries[id]
	s.mu.RUnlock()

	if !exists || item.entry.Dirty || item.image == nil {
		s.mu.Lock()
		s.stats.Misses++
		s.stats.TotalMisses++
		s.mu.Unlock()
		return nil, false
	}

	s.mu.Lock()
	item.entry.LastAccess = s.frameNumber
	s.stats.Hits++
	s.stats.TotalHits++
	s.mu.Unlock()

	return item.image, true
}

// GetOrCreate retrieves a cached image or allocates one for rendering.
// If clean is true, caller can use the cached image directly.
// If clean is false, caller should render to the returned image and call MarkClean.
func (s *System) GetOrCreate(id string, width, height int) (*ebiten.Image, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.entries[id]

	// Cache hit - return existing clean entry
	if exists && !item.entry.Dirty && item.image != nil {
		bounds := item.image.Bounds()
		if bounds.Dx() >= width && bounds.Dy() >= height {
			item.entry.LastAccess = s.frameNumber
			s.stats.Hits++
			s.stats.TotalHits++
			return item.image, true
		}
		// Size mismatch - need to reallocate
		s.returnToPool(item.image)
	}

	// Cache miss or dirty - allocate/reuse image
	s.stats.Misses++
	s.stats.TotalMisses++

	img := s.getFromPool(width, height)

	if exists {
		item.image = img
		item.entry.Width = width
		item.entry.Height = height
		item.entry.Dirty = true // Mark as needing render
		item.entry.LastAccess = s.frameNumber
	} else {
		s.entries[id] = &cacheItem{
			entry: CacheEntry{
				ID:         id,
				Width:      width,
				Height:     height,
				Dirty:      true,
				LastAccess: s.frameNumber,
			},
			image: img,
		}
	}

	return img, false
}

// Put stores a rendered image in the cache.
func (s *System) Put(id string, img *ebiten.Image, width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.entries[id]
	if exists {
		// Return old image to pool if different
		if item.image != nil && item.image != img {
			s.returnToPool(item.image)
		}
		item.image = img
		item.entry.Width = width
		item.entry.Height = height
		item.entry.Dirty = false
		item.entry.LastAccess = s.frameNumber
	} else {
		s.entries[id] = &cacheItem{
			entry: CacheEntry{
				ID:         id,
				Width:      width,
				Height:     height,
				Dirty:      false,
				LastAccess: s.frameNumber,
			},
			image: img,
		}
	}
}

// MarkDirty invalidates a cache entry, requiring re-render.
func (s *System) MarkDirty(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item, exists := s.entries[id]; exists {
		item.entry.Dirty = true
	}
}

// MarkDirtyIfChanged invalidates entry if content hash differs.
func (s *System) MarkDirtyIfChanged(id string, newHash uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.entries[id]
	if !exists {
		return true // Will be treated as miss
	}

	if item.entry.Hash != newHash {
		item.entry.Hash = newHash
		item.entry.Dirty = true
		return true
	}

	return false
}

// MarkClean marks an entry as clean after rendering.
func (s *System) MarkClean(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item, exists := s.entries[id]; exists {
		item.entry.Dirty = false
	}
}

// Invalidate removes an entry from the cache.
func (s *System) Invalidate(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item, exists := s.entries[id]; exists {
		if item.image != nil {
			s.returnToPool(item.image)
		}
		delete(s.entries, id)
	}
}

// InvalidateAll clears all cache entries.
func (s *System) InvalidateAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.entries {
		if item.image != nil {
			s.returnToPool(item.image)
		}
	}

	s.entries = make(map[string]*cacheItem, DefaultMaxEntries)
	s.images = make(map[string]*ebiten.Image, DefaultMaxEntries)

	s.logger.Debug("Cache invalidated - all entries cleared")
}

// InvalidateByPrefix removes entries matching a prefix (e.g., "healthbar_").
func (s *System) InvalidateByPrefix(prefix string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, item := range s.entries {
		if len(id) >= len(prefix) && id[:len(prefix)] == prefix {
			if item.image != nil {
				s.returnToPool(item.image)
			}
			delete(s.entries, id)
			count++
		}
	}

	return count
}

// GetStats returns current cache statistics.
func (s *System) GetStats() CacheStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}

// SetMaxEntries configures the maximum cache size.
func (s *System) SetMaxEntries(max int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxEntries = max
}

// HashString computes a content hash for dirty detection.
func HashString(values ...interface{}) uint64 {
	h := fnv.New64a()
	for _, v := range values {
		switch val := v.(type) {
		case string:
			h.Write([]byte(val))
		case int:
			b := make([]byte, 8)
			b[0] = byte(val)
			b[1] = byte(val >> 8)
			b[2] = byte(val >> 16)
			b[3] = byte(val >> 24)
			b[4] = byte(val >> 32)
			b[5] = byte(val >> 40)
			b[6] = byte(val >> 48)
			b[7] = byte(val >> 56)
			h.Write(b)
		case float64:
			// Convert to bits for consistent hashing
			bits := uint64(val * 1000000) // 6 decimal precision
			b := make([]byte, 8)
			b[0] = byte(bits)
			b[1] = byte(bits >> 8)
			b[2] = byte(bits >> 16)
			b[3] = byte(bits >> 24)
			b[4] = byte(bits >> 32)
			b[5] = byte(bits >> 40)
			b[6] = byte(bits >> 48)
			b[7] = byte(bits >> 56)
			h.Write(b)
		case float32:
			bits := uint64(float64(val) * 1000000)
			b := make([]byte, 8)
			b[0] = byte(bits)
			b[1] = byte(bits >> 8)
			b[2] = byte(bits >> 16)
			b[3] = byte(bits >> 24)
			h.Write(b[:4])
		case bool:
			if val {
				h.Write([]byte{1})
			} else {
				h.Write([]byte{0})
			}
		case uint64:
			b := make([]byte, 8)
			b[0] = byte(val)
			b[1] = byte(val >> 8)
			b[2] = byte(val >> 16)
			b[3] = byte(val >> 24)
			b[4] = byte(val >> 32)
			b[5] = byte(val >> 40)
			b[6] = byte(val >> 48)
			b[7] = byte(val >> 56)
			h.Write(b)
		}
	}
	return h.Sum64()
}

// evictLRU removes the least recently used entries.
func (s *System) evictLRU(count int) {
	if count <= 0 || len(s.entries) == 0 {
		return
	}

	// Find oldest entries by last access time
	type evictCandidate struct {
		id         string
		lastAccess uint64
	}

	candidates := make([]evictCandidate, 0, len(s.entries))
	for id, item := range s.entries {
		candidates = append(candidates, evictCandidate{
			id:         id,
			lastAccess: item.entry.LastAccess,
		})
	}

	// Sort by last access (oldest first)
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].lastAccess < candidates[i].lastAccess {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Evict oldest entries
	evicted := 0
	for i := 0; i < count && i < len(candidates); i++ {
		id := candidates[i].id
		if item, exists := s.entries[id]; exists {
			if item.image != nil {
				s.returnToPool(item.image)
			}
			delete(s.entries, id)
			evicted++
		}
	}

	s.stats.Evictions = evicted

	s.logger.WithFields(logrus.Fields{
		"evicted":   evicted,
		"remaining": len(s.entries),
	}).Debug("LRU eviction completed")
}

// getFromPool retrieves an image from the pool or allocates a new one.
func (s *System) getFromPool(width, height int) *ebiten.Image {
	bucket := GetBucket(width, height)
	pool := s.pools[bucket]

	pool.mu.Lock()
	defer pool.mu.Unlock()

	// Try to reuse from pool
	if len(pool.images) > 0 {
		img := pool.images[len(pool.images)-1]
		pool.images = pool.images[:len(pool.images)-1]

		// Clear the image for reuse
		img.Clear()
		s.stats.PoolAllocations++
		return img
	}

	// Allocate new image at bucket size
	size := int(bucket)
	if width > size {
		size = width
	}
	if height > size {
		size = height
	}

	s.stats.PoolAllocations++
	return ebiten.NewImage(size, size)
}

// returnToPool returns an image to the appropriate pool.
func (s *System) returnToPool(img *ebiten.Image) {
	if img == nil {
		return
	}

	bounds := img.Bounds()
	bucket := GetBucket(bounds.Dx(), bounds.Dy())
	pool := s.pools[bucket]

	pool.mu.Lock()
	defer pool.mu.Unlock()

	// Only keep up to pool size
	if len(pool.images) < DefaultPoolSize {
		pool.images = append(pool.images, img)
		s.stats.PoolReturns++
	}
	// If pool is full, image is GC'd
}

// Update implements system interface for ECS integration.
func (s *System) Update(world interface{}) {
	// UI cache is render-driven, no update logic needed
}

// IsDirty checks if an entry needs re-rendering.
func (s *System) IsDirty(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.entries[id]
	if !exists {
		return true // Treat missing as dirty
	}
	return item.entry.Dirty
}

// HasEntry checks if an entry exists in the cache.
func (s *System) HasEntry(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.entries[id]
	return exists
}

// GetEntryCount returns the number of cached entries.
func (s *System) GetEntryCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}
