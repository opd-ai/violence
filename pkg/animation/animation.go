// Package animation provides state-based sprite animation for entities.
package animation

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"reflect"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// AnimationState represents the current animation state of an entity.
type AnimationState int

const (
	StateIdle AnimationState = iota
	StateWalk
	StateRun
	StateAttack
	StateHurt
	StateDeath
	StateCast
	StateBlock
)

// Direction represents the cardinal direction an entity is facing.
type Direction int

const (
	DirNorth Direction = iota
	DirNorthEast
	DirEast
	DirSouthEast
	DirSouth
	DirSouthWest
	DirWest
	DirNorthWest
)

// AnimationComponent holds animation state for an entity.
type AnimationComponent struct {
	State            AnimationState
	Frame            int
	FrameTime        float64
	Direction        Direction
	Seed             int64
	PrevState        AnimationState
	TransitionPct    float64
	FrameRate        float64 // Target 12 FPS
	DistanceToCamera float64 // For LOD
	Archetype        string  // Entity archetype for sprite generation
}

// Type implements Component interface.
func (a *AnimationComponent) Type() string {
	return "animation"
}

// SpriteCache stores generated sprite frames with LRU eviction.
type SpriteCache struct {
	mu      sync.RWMutex
	entries map[cacheKey]*ebiten.Image
	lru     []cacheKey
	maxSize int
}

type cacheKey struct {
	seed  int64
	state AnimationState
	frame int
	dir   Direction
	genre string
}

// NewSpriteCache creates a cache with maximum size.
func NewSpriteCache(maxSize int) *SpriteCache {
	return &SpriteCache{
		entries: make(map[cacheKey]*ebiten.Image),
		lru:     make([]cacheKey, 0, maxSize),
		maxSize: maxSize,
	}
}

// Get retrieves a cached sprite or returns nil.
func (sc *SpriteCache) Get(seed int64, state AnimationState, frame int, dir Direction, genre string) *ebiten.Image {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	key := cacheKey{seed, state, frame, dir, genre}
	return sc.entries[key]
}

// Put stores a sprite in the cache with LRU eviction.
func (sc *SpriteCache) Put(seed int64, state AnimationState, frame int, dir Direction, genre string, img *ebiten.Image) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	key := cacheKey{seed, state, frame, dir, genre}

	// Check if already exists
	if _, exists := sc.entries[key]; exists {
		return
	}

	// Evict if at capacity
	if len(sc.entries) >= sc.maxSize {
		// Remove oldest entry
		if len(sc.lru) > 0 {
			oldest := sc.lru[0]
			delete(sc.entries, oldest)
			sc.lru = sc.lru[1:]
		}
	}

	sc.entries[key] = img
	sc.lru = append(sc.lru, key)
}

// Clear removes all cached sprites.
func (sc *SpriteCache) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.entries = make(map[cacheKey]*ebiten.Image)
	sc.lru = sc.lru[:0]
}

// AnimationSystem updates animation states and generates sprite frames.
type AnimationSystem struct {
	cache      *SpriteCache
	genre      string
	poolBySize map[int][]*image.RGBA
	poolMu     sync.Mutex
}

// NewAnimationSystem creates the animation system with caching.
func NewAnimationSystem(genre string) *AnimationSystem {
	return &AnimationSystem{
		cache:      NewSpriteCache(100),
		genre:      genre,
		poolBySize: make(map[int][]*image.RGBA),
	}
}

// Update processes all entities with AnimationComponent.
func (sys *AnimationSystem) Update(w *engine.World) {
	deltaTime := 1.0 / 60.0 // 60 FPS

	animType := reflect.TypeOf(&AnimationComponent{})
	entities := w.Query(animType)

	for _, entity := range entities {
		compRaw, ok := w.GetComponent(entity, animType)
		if !ok {
			continue
		}
		anim, ok := compRaw.(*AnimationComponent)
		if !ok {
			continue
		}

		// Update frame time
		anim.FrameTime += deltaTime

		// Determine frame rate based on LOD (distance-based)
		frameInterval := sys.getFrameInterval(anim.DistanceToCamera)

		// Advance frame when time threshold is reached
		if anim.FrameTime >= frameInterval {
			anim.FrameTime = 0
			anim.Frame++

			// Get frame count for current state
			frameCount := sys.getFrameCount(anim.State)
			if anim.Frame >= frameCount {
				// Check for one-shot animations
				if sys.isOneShotAnimation(anim.State) {
					// Transition back to idle
					anim.PrevState = anim.State
					anim.State = StateIdle
					anim.Frame = 0
					anim.TransitionPct = 0
				} else {
					// Loop animation
					anim.Frame = 0
				}
			}
		}

		// Update transition
		if anim.TransitionPct < 1.0 {
			anim.TransitionPct += deltaTime * 8.0 // Fast transition
			if anim.TransitionPct > 1.0 {
				anim.TransitionPct = 1.0
			}
		}
	}
}

// getFrameInterval returns seconds per frame based on distance for LOD.
func (sys *AnimationSystem) getFrameInterval(distance float64) float64 {
	if distance <= 200 {
		return 1.0 / 12.0 // 12 FPS full rate
	} else if distance <= 400 {
		return 1.0 / 6.0 // 6 FPS half rate
	} else {
		return 1.0 / 4.0 // 4 FPS minimal
	}
}

// getFrameCount returns frame count for animation state.
func (sys *AnimationSystem) getFrameCount(state AnimationState) int {
	switch state {
	case StateIdle:
		return 4
	case StateWalk, StateRun:
		return 8
	case StateAttack, StateCast:
		return 6
	case StateHurt:
		return 3
	case StateDeath:
		return 8
	case StateBlock:
		return 2
	default:
		return 1
	}
}

// isOneShotAnimation returns true if animation should not loop.
func (sys *AnimationSystem) isOneShotAnimation(state AnimationState) bool {
	return state == StateAttack || state == StateHurt || state == StateDeath || state == StateCast
}

// SetState changes animation state with transition.
func (sys *AnimationSystem) SetState(anim *AnimationComponent, state AnimationState) {
	if anim.State != state {
		anim.PrevState = anim.State
		anim.State = state
		anim.Frame = 0
		anim.FrameTime = 0
		anim.TransitionPct = 0
	}
}

// SetDirection sets facing direction based on velocity or target.
func (sys *AnimationSystem) SetDirection(anim *AnimationComponent, dx, dy float64) {
	if dx == 0 && dy == 0 {
		return // Keep current direction
	}

	angle := math.Atan2(dy, dx)
	// Convert angle to 8-direction
	octant := int((angle + math.Pi + math.Pi/8) / (math.Pi / 4))
	anim.Direction = Direction(octant % 8)
}

// GenerateSprite creates a sprite for the current animation state.
func (sys *AnimationSystem) GenerateSprite(anim *AnimationComponent) *ebiten.Image {
	// Check cache
	cached := sys.cache.Get(anim.Seed, anim.State, anim.Frame, anim.Direction, sys.genre)
	if cached != nil {
		return cached
	}

	// Generate new sprite
	img := sys.generateSpriteFrame(anim.Seed, anim.Archetype, anim.State, anim.Frame, anim.Direction)
	ebitenImg := ebiten.NewImageFromImage(img)

	// Cache result
	sys.cache.Put(anim.Seed, anim.State, anim.Frame, anim.Direction, sys.genre, ebitenImg)

	return ebitenImg
}

// generateSpriteFrame creates a sprite frame with proper shading and detail.
func (sys *AnimationSystem) generateSpriteFrame(seed int64, archetype string, state AnimationState, frame int, dir Direction) *image.RGBA {
	size := 32
	img := sys.acquireImageBuffer(size * size * 4)

	rng := rand.New(rand.NewSource(seed))

	// Base colors from genre and archetype
	baseColor := sys.getArchetypeColor(archetype, rng)
	accentColor := sys.getAccentColor(archetype, rng)

	// Apply animation offsets
	centerX, centerY := 16, 16
	offsetX, offsetY := sys.getAnimationOffset(state, frame, dir)

	// Generate sprite body with perspective
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - centerX + int(offsetX))
			dy := float64(y - centerY + int(offsetY))
			dist := math.Sqrt(dx*dx + dy*dy)

			// Draw character shape with shading
			if dist < 10 {
				// Body core
				shading := 1.0 - (dist / 15.0)
				shading = math.Max(0.3, shading)

				// Height-based gradient (top-down perspective)
				heightGradient := float64(y) / float64(size)
				shading *= (0.8 + heightGradient*0.4)

				// Apply state-based coloring
				c := sys.applyStateColor(baseColor, accentColor, state, frame, shading)
				img.Set(x, y, c)
			} else if dist < 12 {
				// Outline for readability
				img.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}

	// Add directional indicators and equipment
	sys.addDirectionalDetails(img, dir, baseColor, accentColor, archetype)

	// Add state-specific effects
	sys.addStateEffects(img, state, frame, accentColor)

	return img
}

// getArchetypeColor returns base color for entity archetype.
func (sys *AnimationSystem) getArchetypeColor(archetype string, rng *rand.Rand) color.RGBA {
	switch sys.genre {
	case "fantasy":
		switch archetype {
		case "warrior":
			return color.RGBA{180, 100, 60, 255} // Brown armor
		case "mage":
			return color.RGBA{80, 80, 180, 255} // Blue robes
		case "rogue":
			return color.RGBA{60, 60, 60, 255} // Dark leather
		default:
			return color.RGBA{100, 150, 100, 255} // Green default
		}
	case "scifi":
		return color.RGBA{120 + uint8(rng.Intn(60)), 140 + uint8(rng.Intn(60)), 200, 255}
	case "horror":
		return color.RGBA{80 + uint8(rng.Intn(40)), 60 + uint8(rng.Intn(40)), 70 + uint8(rng.Intn(40)), 255}
	case "cyberpunk":
		return color.RGBA{200, 50 + uint8(rng.Intn(100)), 220, 255}
	default:
		return color.RGBA{128, 128, 128, 255}
	}
}

// getAccentColor returns accent/highlight color.
func (sys *AnimationSystem) getAccentColor(archetype string, rng *rand.Rand) color.RGBA {
	switch sys.genre {
	case "fantasy":
		return color.RGBA{220, 180, 80, 255} // Gold
	case "scifi":
		return color.RGBA{100, 220, 255, 255} // Cyan
	case "horror":
		return color.RGBA{180, 20, 20, 255} // Blood red
	case "cyberpunk":
		return color.RGBA{255, 0, 180, 255} // Neon pink
	default:
		return color.RGBA{255, 255, 255, 255}
	}
}

// getAnimationOffset returns sprite offset for animation frame.
func (sys *AnimationSystem) getAnimationOffset(state AnimationState, frame int, dir Direction) (float64, float64) {
	var offsetX, offsetY float64

	switch state {
	case StateWalk, StateRun:
		// Bob up and down
		bobCycle := float64(frame) / 8.0 * 2.0 * math.Pi
		offsetY = math.Sin(bobCycle) * 2.0
		offsetX = math.Sin(bobCycle*2) * 1.0
	case StateAttack:
		// Lunge forward
		lunge := float64(frame) / 6.0
		if lunge < 0.5 {
			offsetY = lunge * 4.0
		} else {
			offsetY = (1.0 - lunge) * 4.0
		}
	case StateHurt:
		// Recoil
		offsetY = float64(3-frame) * 1.5
	case StateDeath:
		// Fall down
		offsetY = float64(frame) * 0.5
	}

	return offsetX, offsetY
}

// applyStateColor modifies color based on animation state.
func (sys *AnimationSystem) applyStateColor(base, accent color.RGBA, state AnimationState, frame int, shading float64) color.RGBA {
	r, g, b, a := base.R, base.G, base.B, base.A

	switch state {
	case StateAttack, StateCast:
		// Flash with accent color
		flashPct := float64(frame) / 6.0
		if flashPct < 0.3 {
			mix := flashPct / 0.3
			r = uint8(float64(r)*(1-mix) + float64(accent.R)*mix)
			g = uint8(float64(g)*(1-mix) + float64(accent.G)*mix)
			b = uint8(float64(b)*(1-mix) + float64(accent.B)*mix)
		}
	case StateHurt:
		// Red flash
		r = uint8(math.Min(255, float64(r)+100))
	case StateDeath:
		// Fade out
		fadePct := float64(frame) / 8.0
		darken := 1.0 - fadePct*0.7
		r = uint8(float64(r) * darken)
		g = uint8(float64(g) * darken)
		b = uint8(float64(b) * darken)
	}

	// Apply shading
	r = uint8(float64(r) * shading)
	g = uint8(float64(g) * shading)
	b = uint8(float64(b) * shading)

	return color.RGBA{r, g, b, a}
}

// addDirectionalDetails adds facing-direction indicators.
func (sys *AnimationSystem) addDirectionalDetails(img *image.RGBA, dir Direction, base, accent color.RGBA, archetype string) {
	// Add simple directional marker (e.g., weapon, shield side)
	markerX, markerY := 16, 16

	switch dir {
	case DirEast, DirNorthEast, DirSouthEast:
		markerX += 8
	case DirWest, DirNorthWest, DirSouthWest:
		markerX -= 8
	}

	// Draw small accent marker
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			x, y := markerX+dx, markerY+dy
			if x >= 0 && x < 32 && y >= 0 && y < 32 {
				img.Set(x, y, accent)
			}
		}
	}
}

// addStateEffects adds visual effects for animation states.
func (sys *AnimationSystem) addStateEffects(img *image.RGBA, state AnimationState, frame int, accent color.RGBA) {
	if state == StateAttack || state == StateCast {
		// Add motion blur lines
		if frame < 3 {
			for i := 0; i < 3; i++ {
				x := 20 + i*2
				for y := 12; y < 20; y++ {
					if x < 32 {
						current := img.At(x, y)
						if _, _, _, a := current.RGBA(); a == 0 {
							alpha := uint8(100 - i*30)
							img.Set(x, y, color.RGBA{accent.R, accent.G, accent.B, alpha})
						}
					}
				}
			}
		}
	}
}

// acquireImageBuffer gets a pooled image buffer.
func (sys *AnimationSystem) acquireImageBuffer(sizeKey int) *image.RGBA {
	sys.poolMu.Lock()
	defer sys.poolMu.Unlock()

	pool := sys.poolBySize[sizeKey]
	if len(pool) > 0 {
		img := pool[len(pool)-1]
		sys.poolBySize[sizeKey] = pool[:len(pool)-1]
		// Clear image
		for i := range img.Pix {
			img.Pix[i] = 0
		}
		return img
	}

	// Create new buffer
	return image.NewRGBA(image.Rect(0, 0, 32, 32))
}

// releaseImageBuffer returns buffer to pool.
func (sys *AnimationSystem) releaseImageBuffer(img *image.RGBA) {
	sys.poolMu.Lock()
	defer sys.poolMu.Unlock()

	sizeKey := len(img.Pix)
	sys.poolBySize[sizeKey] = append(sys.poolBySize[sizeKey], img)
}

// SetGenre updates the genre for sprite generation.
func (sys *AnimationSystem) SetGenre(genre string) {
	if sys.genre != genre {
		sys.genre = genre
		sys.cache.Clear()
		logrus.WithFields(logrus.Fields{
			"system": "animation",
			"genre":  genre,
		}).Info("Genre changed, sprite cache cleared")
	}
}
