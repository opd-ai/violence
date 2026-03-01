// Package collision provides attack shape caching and frame-based hitbox management.
package collision

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// AttackShapeCache stores pre-computed attack shapes for weapons and spells.
type AttackShapeCache struct {
	shapes map[string]*AttackShape
	mu     sync.RWMutex
}

// AttackShape defines a cached attack hitbox pattern.
type AttackShape struct {
	Name     string
	Vertices []Point   // Polygon vertices in local space
	Collider *Collider // Pre-configured collider template
	Width    float64   // For capsule/AABB shapes
	Height   float64
}

// NewAttackShapeCache creates a new attack shape cache.
func NewAttackShapeCache() *AttackShapeCache {
	return &AttackShapeCache{
		shapes: make(map[string]*AttackShape),
	}
}

// RegisterShape adds a new attack shape to the cache.
func (c *AttackShapeCache) RegisterShape(name string, shape *AttackShape) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shapes[name] = shape
}

// GetShape retrieves a cached attack shape.
func (c *AttackShapeCache) GetShape(name string) *AttackShape {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.shapes[name]
}

// CreateColliderFromShape creates a positioned collider from a cached shape.
func (c *AttackShapeCache) CreateColliderFromShape(name string, x, y, rotation float64, layer, mask Layer) *Collider {
	c.mu.RLock()
	shape := c.shapes[name]
	c.mu.RUnlock()

	if shape == nil || len(shape.Vertices) == 0 {
		// Fallback to small circle
		return NewCircleCollider(x, y, 5, layer, mask)
	}

	// Rotate and translate vertices
	rotatedVerts := make([]Point, len(shape.Vertices))

	// For now, use vertices as-is (local space)
	// Direction-based rotation can be added later when needed
	for i, v := range shape.Vertices {
		rotatedVerts[i] = Point{X: v.X, Y: v.Y}
	}

	return NewPolygonCollider(x, y, rotatedVerts, layer, mask)
}

// WeaponShapeGenerator creates standard weapon attack shapes.
type WeaponShapeGenerator struct {
	cache *AttackShapeCache
}

// NewWeaponShapeGenerator creates a weapon shape generator.
func NewWeaponShapeGenerator(cache *AttackShapeCache) *WeaponShapeGenerator {
	return &WeaponShapeGenerator{cache: cache}
}

// GenerateWeaponShapes pre-generates common weapon attack patterns.
func (g *WeaponShapeGenerator) GenerateWeaponShapes() {
	// Sword horizontal slash (90 degree arc)
	g.cache.RegisterShape("sword_slash_h", &AttackShape{
		Name:     "sword_slash_h",
		Vertices: GenerateAttackArc(0, 0, 25, -0.785, 0.785, 10), // ~90 degrees
	})

	// Sword vertical slash (60 degree arc)
	g.cache.RegisterShape("sword_slash_v", &AttackShape{
		Name:     "sword_slash_v",
		Vertices: GenerateAttackArc(0, 0, 25, -0.524, 0.524, 8), // ~60 degrees
	})

	// Sword thrust (narrow cone)
	g.cache.RegisterShape("sword_thrust", &AttackShape{
		Name:     "sword_thrust",
		Vertices: GenerateConeShape(0, 0, 30, 0, 0.262), // ~15 degrees each side
	})

	// Axe wide swing (120 degree arc)
	g.cache.RegisterShape("axe_swing", &AttackShape{
		Name:     "axe_swing",
		Vertices: GenerateAttackArc(0, 0, 28, -1.047, 1.047, 12), // ~120 degrees
	})

	// Spear thrust (long narrow capsule)
	g.cache.RegisterShape("spear_thrust", &AttackShape{
		Name:     "spear_thrust",
		Vertices: GenerateRectangleShape(20, 0, 40, 4, 0), // Long thin rectangle
	})

	// Dagger quick stab (small cone)
	g.cache.RegisterShape("dagger_stab", &AttackShape{
		Name:     "dagger_stab",
		Vertices: GenerateConeShape(0, 0, 15, 0, 0.349), // ~20 degrees each side
	})

	// Hammer overhead smash (wide short arc)
	g.cache.RegisterShape("hammer_smash", &AttackShape{
		Name:     "hammer_smash",
		Vertices: GenerateAttackArc(0, 0, 22, -0.785, 0.785, 8),
	})

	// Whip long sweep (180 degree arc)
	g.cache.RegisterShape("whip_sweep", &AttackShape{
		Name:     "whip_sweep",
		Vertices: GenerateAttackArc(0, 0, 45, -1.571, 1.571, 16), // ~180 degrees
	})

	// Fireball (circle AoE on impact)
	g.cache.RegisterShape("fireball_impact", &AttackShape{
		Name: "fireball_impact",
		Vertices: []Point{
			{X: -10, Y: -10},
			{X: 10, Y: -10},
			{X: 10, Y: 10},
			{X: -10, Y: 10},
		}, // Approximate circle with square
	})

	// Lightning beam (long thin rectangle)
	g.cache.RegisterShape("lightning_beam", &AttackShape{
		Name:     "lightning_beam",
		Vertices: GenerateRectangleShape(25, 0, 50, 6, 0),
	})

	// Ice cone (60 degree spread)
	g.cache.RegisterShape("ice_cone", &AttackShape{
		Name:     "ice_cone",
		Vertices: GenerateConeShape(0, 0, 35, 0, 0.524), // ~30 degrees each side
	})

	// Shockwave ring (handled separately as it needs dual collider)
	// Players will use CreateRingCollider for this

	// Cleave (wide arc for 2H weapons)
	g.cache.RegisterShape("cleave", &AttackShape{
		Name:     "cleave",
		Vertices: GenerateAttackArc(0, 0, 32, -1.047, 1.047, 14), // ~120 degrees
	})

	// Backstab (small precise cone)
	g.cache.RegisterShape("backstab", &AttackShape{
		Name:     "backstab",
		Vertices: GenerateConeShape(0, 0, 12, 0, 0.175), // ~10 degrees each side
	})
}

// SpriteColliderComponent stores collision geometry extracted from sprite data.
type SpriteColliderComponent struct {
	Polygon      []Point   // Convex hull extracted from sprite
	BoundingBox  *Collider // Tight AABB for broadphase
	DetailedHull *Collider // Polygon collider for narrow phase
	LastSpriteID string    // Cache key to detect sprite changes
	Dirty        bool      // Whether geometry needs regeneration
}

// AttackFrameComponent stores attack hitbox data per animation frame.
type AttackFrameComponent struct {
	ActiveFrames map[int]*Collider // Frame index -> attack collider
	CurrentFrame int
	ShapeName    string // Name of cached attack shape
	Damage       float64
	Knockback    float64
}

// NewAttackFrameComponent creates an attack frame component.
func NewAttackFrameComponent(shapeName string, damage, knockback float64) *AttackFrameComponent {
	return &AttackFrameComponent{
		ActiveFrames: make(map[int]*Collider),
		ShapeName:    shapeName,
		Damage:       damage,
		Knockback:    knockback,
	}
}

// SetActiveFrame sets the attack collider for a specific animation frame.
func (a *AttackFrameComponent) SetActiveFrame(frame int, collider *Collider) {
	a.ActiveFrames[frame] = collider
}

// GetCurrentCollider returns the collider for the current animation frame, or nil.
func (a *AttackFrameComponent) GetCurrentCollider() *Collider {
	return a.ActiveFrames[a.CurrentFrame]
}

// UpdateFrame updates the current animation frame.
func (a *AttackFrameComponent) UpdateFrame(frame int) {
	a.CurrentFrame = frame
}

// CollisionGeometrySystem manages sprite-based collision geometry extraction.
type CollisionGeometrySystem struct {
	extractor *GeometryExtractor
	cache     *AttackShapeCache
	generator *WeaponShapeGenerator
}

// NewCollisionGeometrySystem creates a collision geometry system.
func NewCollisionGeometrySystem() *CollisionGeometrySystem {
	cache := NewAttackShapeCache()
	gen := NewWeaponShapeGenerator(cache)
	gen.GenerateWeaponShapes()

	return &CollisionGeometrySystem{
		extractor: NewGeometryExtractor(),
		cache:     cache,
		generator: gen,
	}
}

// ExtractSpriteCollider generates a collision polygon from a sprite image.
func (s *CollisionGeometrySystem) ExtractSpriteCollider(sprite *ebiten.Image, x, y float64, layer, mask Layer) *Collider {
	hull := s.extractor.ExtractConvexHull(sprite)
	if len(hull) == 0 {
		// Fallback to bounding box
		w, h := s.extractor.ExtractBoundingBox(sprite)
		return NewAABBCollider(x-w/2, y-h/2, w, h, layer, mask)
	}

	return NewPolygonCollider(x, y, hull, layer, mask)
}

// GetAttackShape retrieves a cached attack shape.
func (s *CollisionGeometrySystem) GetAttackShape(name string) *AttackShape {
	return s.cache.GetShape(name)
}

// CreateAttackCollider creates a positioned attack collider from a cached shape.
func (s *CollisionGeometrySystem) CreateAttackCollider(shapeName string, x, y, dirX, dirY float64, layer, mask Layer) *Collider {
	shape := s.cache.GetShape(shapeName)
	if shape == nil {
		// Fallback: small forward cone
		return CreateConeCollider(x, y, dirX, dirY, 20, 0.785, layer, mask)
	}

	// Use the cached shape vertices directly
	// Direction-based rotation can be applied when rendering
	// For now, just offset the shape to the attack position
	return NewPolygonCollider(x, y, shape.Vertices, layer, mask)
}

// UpdateSpriteCollider updates a sprite collider component if the sprite has changed.
func (s *CollisionGeometrySystem) UpdateSpriteCollider(comp *SpriteColliderComponent, sprite *ebiten.Image, spriteID string, x, y float64, layer, mask Layer) {
	if comp == nil {
		return
	}

	// Check if update is needed
	if !comp.Dirty && comp.LastSpriteID == spriteID {
		// Update position only
		if comp.DetailedHull != nil {
			comp.DetailedHull.X = x
			comp.DetailedHull.Y = y
		}
		if comp.BoundingBox != nil {
			comp.BoundingBox.X = x
			comp.BoundingBox.Y = y
		}
		return
	}

	// Extract new geometry
	hull := s.extractor.ExtractConvexHull(sprite)
	w, h := s.extractor.ExtractBoundingBox(sprite)

	comp.Polygon = hull
	comp.BoundingBox = NewAABBCollider(x-w/2, y-h/2, w, h, layer, mask)
	comp.DetailedHull = NewPolygonCollider(x, y, hull, layer, mask)
	comp.LastSpriteID = spriteID
	comp.Dirty = false
}

// GetExtractor returns the geometry extractor for direct use.
func (s *CollisionGeometrySystem) GetExtractor() *GeometryExtractor {
	return s.extractor
}

// GetShapeCache returns the attack shape cache for registration of custom shapes.
func (s *CollisionGeometrySystem) GetShapeCache() *AttackShapeCache {
	return s.cache
}
