// Package destruct implements destructible environment objects.
package destruct

import (
	"sync"
)

// Destructible represents a destructible object in the world.
type Destructible struct {
	ID        string
	Health    float64
	MaxHealth float64
	Destroyed bool
	X, Y      float64
	Type      string
	DropItems []string
	mu        sync.RWMutex
}

// System manages destructible objects in a level.
type System struct {
	objects map[string]*Destructible
	mu      sync.RWMutex
}

// NewSystem creates a new destructible system.
func NewSystem() *System {
	return &System{
		objects: make(map[string]*Destructible),
	}
}

// NewDestructible creates a new destructible object.
func NewDestructible(id, objType string, health, x, y float64) *Destructible {
	return &Destructible{
		ID:        id,
		Health:    health,
		MaxHealth: health,
		Destroyed: false,
		X:         x,
		Y:         y,
		Type:      objType,
		DropItems: make([]string, 0),
	}
}

// Add adds a destructible object to the system.
func (s *System) Add(d *Destructible) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[d.ID] = d
}

// Remove removes a destructible object from the system.
func (s *System) Remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, id)
}

// Get retrieves a destructible object by ID.
func (s *System) Get(id string) (*Destructible, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.objects[id]
	return d, ok
}

// GetAll returns all destructible objects.
func (s *System) GetAll() []*Destructible {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Destructible, 0, len(s.objects))
	for _, d := range s.objects {
		result = append(result, d)
	}
	return result
}

// Damage applies damage to a destructible object.
func (d *Destructible) Damage(amount float64) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Destroyed {
		return false
	}

	d.Health -= amount
	if d.Health <= 0 {
		d.Health = 0
		d.Destroyed = true
		return true // Destroyed
	}
	return false
}

// Destroy immediately destroys the object.
func (d *Destructible) Destroy() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.Health = 0
	d.Destroyed = true
}

// Repair restores health to a destructible object.
func (d *Destructible) Repair(amount float64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Destroyed {
		return
	}

	d.Health += amount
	if d.Health > d.MaxHealth {
		d.Health = d.MaxHealth
	}
}

// IsDestroyed returns whether the object is destroyed.
func (d *Destructible) IsDestroyed() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Destroyed
}

// GetHealth returns current health.
func (d *Destructible) GetHealth() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Health
}

// AddDropItem adds an item ID to the drop list.
func (d *Destructible) AddDropItem(itemID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.DropItems = append(d.DropItems, itemID)
}

// GetDropItems returns the drop item list.
func (d *Destructible) GetDropItems() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]string, len(d.DropItems))
	copy(result, d.DropItems)
	return result
}

// SetGenre configures destructible types for a genre.
func SetGenre(genreID string) {}

// BreakableWall represents a wall that can be destroyed to reveal a passage.
type BreakableWall struct {
	ID           string
	X, Y         float64
	Health       float64
	MaxHealth    float64
	Destroyed    bool
	RevealsPath  bool
	PathX, PathY int
	mu           sync.RWMutex
}

// NewBreakableWall creates a new breakable wall.
func NewBreakableWall(id string, x, y, health float64, revealsPath bool) *BreakableWall {
	return &BreakableWall{
		ID:          id,
		X:           x,
		Y:           y,
		Health:      health,
		MaxHealth:   health,
		Destroyed:   false,
		RevealsPath: revealsPath,
	}
}

// Damage applies damage to the wall.
func (w *BreakableWall) Damage(amount float64) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.Destroyed {
		return false
	}

	w.Health -= amount
	if w.Health <= 0 {
		w.Health = 0
		w.Destroyed = true
		return true
	}
	return false
}

// IsDestroyed returns whether the wall is destroyed.
func (w *BreakableWall) IsDestroyed() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.Destroyed
}

// GetHealth returns current health.
func (w *BreakableWall) GetHealth() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.Health
}

// SetRevealedPath sets the tile coordinates revealed when destroyed.
func (w *BreakableWall) SetRevealedPath(x, y int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.PathX = x
	w.PathY = y
	w.RevealsPath = true
}

// GetRevealedPath returns the revealed path coordinates.
func (w *BreakableWall) GetRevealedPath() (int, int, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.PathX, w.PathY, w.RevealsPath
}

// DestructibleObject represents objects like barrels and crates with explosion chains.
type DestructibleObject struct {
	Destructible
	Explosive      bool
	ExplosionRange float64
	ChainReaction  bool
}

// NewDestructibleObject creates a new destructible object.
func NewDestructibleObject(id, objType string, health, x, y float64, explosive bool) *DestructibleObject {
	return &DestructibleObject{
		Destructible: Destructible{
			ID:        id,
			Health:    health,
			MaxHealth: health,
			Destroyed: false,
			X:         x,
			Y:         y,
			Type:      objType,
			DropItems: make([]string, 0),
		},
		Explosive:      explosive,
		ExplosionRange: 3.0,
		ChainReaction:  explosive,
	}
}

// GetExplosionTargets returns objects within explosion range.
func (o *DestructibleObject) GetExplosionTargets(allObjects []*DestructibleObject) []*DestructibleObject {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if !o.Explosive || !o.Destroyed {
		return nil
	}

	targets := make([]*DestructibleObject, 0)
	for _, other := range allObjects {
		if other.ID == o.ID || other.IsDestroyed() {
			continue
		}

		dx := other.X - o.X
		dy := other.Y - o.Y
		dist := dx*dx + dy*dy // squared distance

		if dist <= o.ExplosionRange*o.ExplosionRange {
			targets = append(targets, other)
		}
	}

	return targets
}

// Debris represents temporary debris from destroyed objects.
type Debris struct {
	ID            string
	X, Y          float64
	Material      string
	BlocksPath    bool
	TimeRemaining float64
	MaxTime       float64
	mu            sync.RWMutex
}

// NewDebris creates a new debris object.
func NewDebris(id string, x, y float64, material string, blocksPath bool, duration float64) *Debris {
	return &Debris{
		ID:            id,
		X:             x,
		Y:             y,
		Material:      material,
		BlocksPath:    blocksPath,
		TimeRemaining: duration,
		MaxTime:       duration,
	}
}

// Update advances debris timer; returns true when cleared.
func (d *Debris) Update(deltaTime float64) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.TimeRemaining -= deltaTime
	return d.TimeRemaining <= 0
}

// IsCleared returns whether debris has been cleared.
func (d *Debris) IsCleared() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.TimeRemaining <= 0
}

// GetProgress returns clear progress (0.0 = just created, 1.0 = cleared).
func (d *Debris) GetProgress() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.MaxTime <= 0 {
		return 1.0
	}
	return 1.0 - (d.TimeRemaining / d.MaxTime)
}

// GetDebrisMaterial returns genre-specific debris material name.
func GetDebrisMaterial(genre, objectType string) string {
	materials := map[string]map[string]string{
		"fantasy": {
			"wall":    "stone rubble",
			"barrel":  "wooden splinters",
			"crate":   "wooden fragments",
			"door":    "splintered wood",
			"default": "debris",
		},
		"scifi": {
			"wall":    "hull shards",
			"barrel":  "metal fragments",
			"crate":   "alloy pieces",
			"door":    "broken plating",
			"default": "wreckage",
		},
		"horror": {
			"wall":    "crumbling plaster",
			"barrel":  "rotted wood",
			"crate":   "decayed boards",
			"door":    "splintered planks",
			"default": "remains",
		},
		"cyberpunk": {
			"wall":    "shattered glass",
			"barrel":  "polymer shards",
			"crate":   "synthetic fragments",
			"door":    "broken circuits",
			"default": "scrap",
		},
		"postapoc": {
			"wall":    "concrete chunks",
			"barrel":  "rusted metal",
			"crate":   "salvaged parts",
			"door":    "warped metal",
			"default": "rubble",
		},
	}

	genreMaterials, ok := materials[genre]
	if !ok {
		genreMaterials = materials["fantasy"]
	}

	material, ok := genreMaterials[objectType]
	if !ok {
		material = genreMaterials["default"]
	}

	return material
}
