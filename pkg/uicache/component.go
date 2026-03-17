package uicache

// CacheEntry represents a single cached UI element.
type CacheEntry struct {
	// ID uniquely identifies this cached element.
	ID string

	// Width is the pixel width of the cached image.
	Width int

	// Height is the pixel height of the cached image.
	Height int

	// Dirty indicates whether this entry needs re-rendering.
	Dirty bool

	// LastAccess is the frame number when this entry was last used.
	LastAccess uint64

	// Hash is a content hash for change detection.
	Hash uint64

	// PositionX is the cached screen X position.
	PositionX float32

	// PositionY is the cached screen Y position.
	PositionY float32
}

// Type returns the component type identifier.
func (c *CacheEntry) Type() string {
	return "uicache.CacheEntry"
}

// CacheStats tracks cache performance metrics.
type CacheStats struct {
	// Hits is the number of cache hits this frame.
	Hits int

	// Misses is the number of cache misses this frame.
	Misses int

	// Evictions is the number of LRU evictions this frame.
	Evictions int

	// TotalEntries is the current cache entry count.
	TotalEntries int

	// TotalHits is the cumulative hit count.
	TotalHits uint64

	// TotalMisses is the cumulative miss count.
	TotalMisses uint64

	// PoolAllocations is the number of images allocated from pool.
	PoolAllocations int

	// PoolReturns is the number of images returned to pool.
	PoolReturns int
}

// Type returns the component type identifier.
func (c *CacheStats) Type() string {
	return "uicache.CacheStats"
}

// HitRate returns the cache hit rate as a percentage (0-100).
func (c *CacheStats) HitRate() float64 {
	total := c.TotalHits + c.TotalMisses
	if total == 0 {
		return 0
	}
	return float64(c.TotalHits) / float64(total) * 100
}

// FrameHitRate returns the hit rate for the current frame.
func (c *CacheStats) FrameHitRate() float64 {
	total := c.Hits + c.Misses
	if total == 0 {
		return 0
	}
	return float64(c.Hits) / float64(total) * 100
}

// SizeBucket represents a pooled image size category.
type SizeBucket int

const (
	// Bucket16 is for 16x16 and smaller images.
	Bucket16 SizeBucket = 16
	// Bucket32 is for 17x17 to 32x32 images.
	Bucket32 SizeBucket = 32
	// Bucket64 is for 33x33 to 64x64 images.
	Bucket64 SizeBucket = 64
	// Bucket128 is for 65x65 to 128x128 images.
	Bucket128 SizeBucket = 128
	// Bucket256 is for 129x129 to 256x256 images.
	Bucket256 SizeBucket = 256
	// BucketLarge is for images larger than 256x256.
	BucketLarge SizeBucket = 512
)

// GetBucket returns the appropriate size bucket for given dimensions.
func GetBucket(width, height int) SizeBucket {
	maxDim := width
	if height > maxDim {
		maxDim = height
	}

	switch {
	case maxDim <= 16:
		return Bucket16
	case maxDim <= 32:
		return Bucket32
	case maxDim <= 64:
		return Bucket64
	case maxDim <= 128:
		return Bucket128
	case maxDim <= 256:
		return Bucket256
	default:
		return BucketLarge
	}
}
