// Package props manages decorative props and level objects.
package props

import (
	"sync"

	"github.com/opd-ai/violence/pkg/procgen/genre"
	"github.com/opd-ai/violence/pkg/rng"
)

// PropType represents sprite type for rendering
type PropType int

const (
	PropBarrel PropType = iota
	PropCrate
	PropTable
	PropTerminal
	PropBones
	PropPlant
	PropPillar
	PropTorch
	PropDebris
	PropContainer
)

// Prop represents a decorative sprite in the game world.
type Prop struct {
	ID         string
	X, Y       float64
	SpriteType PropType
	Collision  bool // Whether prop blocks movement
	Name       string
}

// Room represents a rectangular room for prop placement.
type Room struct {
	X, Y, W, H int
}

// Manager handles prop placement and tracking.
type Manager struct {
	props      []*Prop
	genre      string
	propLists  map[string][]propTemplate
	mu         sync.RWMutex
	nextPropID int
}

// propTemplate defines spawn probabilities and properties for each prop type
type propTemplate struct {
	PropType  PropType
	Name      string
	Collision bool
	Weight    int // Spawn probability weight
}

// NewManager creates a new prop manager.
func NewManager() *Manager {
	m := &Manager{
		props:     make([]*Prop, 0),
		genre:     genre.Fantasy,
		propLists: make(map[string][]propTemplate),
	}
	m.initializePropLists()
	return m
}

// SetGenre configures prop sets for a genre.
func (m *Manager) SetGenre(genreID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.genre = genreID
}

// GetGenre returns current genre.
func (m *Manager) GetGenre() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.genre
}

// initializePropLists sets up genre-specific prop templates
func (m *Manager) initializePropLists() {
	// Fantasy: medieval/dungeon props
	m.propLists[genre.Fantasy] = []propTemplate{
		{PropBarrel, "Barrel", true, 20},
		{PropCrate, "Crate", true, 20},
		{PropTable, "Table", true, 15},
		{PropBones, "Bones", false, 10},
		{PropPillar, "Stone Pillar", true, 8},
		{PropTorch, "Torch", false, 12},
		{PropDebris, "Rubble", false, 15},
	}

	// SciFi: futuristic/tech props
	m.propLists[genre.SciFi] = []propTemplate{
		{PropCrate, "Supply Crate", true, 20},
		{PropTerminal, "Terminal", true, 18},
		{PropContainer, "Storage Pod", true, 15},
		{PropPillar, "Support Column", true, 10},
		{PropDebris, "Hull Fragment", false, 12},
		{PropTable, "Console", true, 15},
		{PropBarrel, "Fuel Drum", true, 10},
	}

	// Horror: creepy/abandoned props
	m.propLists[genre.Horror] = []propTemplate{
		{PropBones, "Corpse", false, 18},
		{PropDebris, "Debris", false, 20},
		{PropTable, "Gurney", true, 12},
		{PropBarrel, "Biohazard Barrel", true, 10},
		{PropCrate, "Medical Crate", true, 10},
		{PropContainer, "Body Bag", false, 15},
		{PropPillar, "Cracked Pillar", true, 8},
		{PropTorch, "Flickering Light", false, 7},
	}

	// Cyberpunk: neon/corporate props
	m.propLists[genre.Cyberpunk] = []propTemplate{
		{PropTerminal, "Data Terminal", true, 20},
		{PropCrate, "Corp Crate", true, 15},
		{PropContainer, "Vending Machine", true, 12},
		{PropTable, "Desk", true, 15},
		{PropDebris, "Glass Shards", false, 13},
		{PropPillar, "Concrete Pillar", true, 10},
		{PropBarrel, "Trash Can", true, 10},
		{PropTorch, "Neon Sign", false, 5},
	}

	// PostApoc: wasteland/scrap props
	m.propLists[genre.PostApoc] = []propTemplate{
		{PropBarrel, "Rusted Barrel", true, 20},
		{PropDebris, "Scrap Pile", false, 25},
		{PropCrate, "Salvage Crate", true, 15},
		{PropBones, "Skeleton", false, 12},
		{PropContainer, "Locker", true, 10},
		{PropPillar, "Rusted Support", true, 8},
		{PropTable, "Workbench", true, 10},
	}
}

// PlaceProps populates a room with props based on density.
// density: 0.0-1.0, higher = more props. Typical: 0.1-0.3
func (m *Manager) PlaceProps(room *Room, density float64, seed uint64) []*Prop {
	m.mu.Lock()
	defer m.mu.Unlock()

	r := rng.NewRNG(seed)
	placedProps := make([]*Prop, 0)

	// Calculate number of props based on room area and density
	area := room.W * room.H
	propCount := int(float64(area) * density * 0.05) // Scale factor for reasonable density
	if propCount < 1 && density > 0 {
		propCount = 1
	}

	// Get prop templates for current genre
	templates := m.propLists[m.genre]
	if len(templates) == 0 {
		return placedProps
	}

	// Calculate total weight for weighted random selection
	totalWeight := 0
	for _, t := range templates {
		totalWeight += t.Weight
	}

	// Place props in random positions within room
	for i := 0; i < propCount; i++ {
		// Weighted random selection of prop type
		roll := r.Intn(totalWeight)
		cumWeight := 0
		var selectedTemplate *propTemplate
		for j := range templates {
			cumWeight += templates[j].Weight
			if roll < cumWeight {
				selectedTemplate = &templates[j]
				break
			}
		}
		if selectedTemplate == nil {
			continue
		}

		// Random position within room (avoid edges)
		margin := 1
		if room.W <= 2 || room.H <= 2 {
			margin = 0
		}
		posX := float64(room.X+margin) + r.Float64()*float64(room.W-2*margin)
		posY := float64(room.Y+margin) + r.Float64()*float64(room.H-2*margin)

		// Create prop
		m.nextPropID++
		prop := &Prop{
			ID:         generatePropID(m.nextPropID),
			X:          posX,
			Y:          posY,
			SpriteType: selectedTemplate.PropType,
			Collision:  selectedTemplate.Collision,
			Name:       selectedTemplate.Name,
		}

		m.props = append(m.props, prop)
		placedProps = append(placedProps, prop)
	}

	return placedProps
}

// GetProps returns a copy of all props.
func (m *Manager) GetProps() []*Prop {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Prop, 0, len(m.props))
	for _, p := range m.props {
		// Create a copy of each prop
		propCopy := *p
		result = append(result, &propCopy)
	}
	return result
}

// Clear removes all props.
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.props = make([]*Prop, 0)
	m.nextPropID = 0
}

// AddProp manually adds a prop to the manager.
func (m *Manager) AddProp(prop *Prop) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if prop.ID == "" {
		m.nextPropID++
		prop.ID = generatePropID(m.nextPropID)
	}
	m.props = append(m.props, prop)
}

// RemoveProp removes a prop by ID.
func (m *Manager) RemoveProp(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, p := range m.props {
		if p.ID == id {
			m.props = append(m.props[:i], m.props[i+1:]...)
			return true
		}
	}
	return false
}

// GetPropsByType returns all props of a specific type.
func (m *Manager) GetPropsByType(propType PropType) []*Prop {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Prop, 0)
	for _, p := range m.props {
		if p.SpriteType == propType {
			result = append(result, p)
		}
	}
	return result
}

// generatePropID creates a unique prop ID.
func generatePropID(id int) string {
	return "prop_" + string(rune('0'+id/1000%10)) +
		string(rune('0'+id/100%10)) +
		string(rune('0'+id/10%10)) +
		string(rune('0'+id%10))
}

// Place creates and places a single prop (legacy compatibility).
func Place(name string, x, y float64) *Prop {
	return &Prop{
		Name: name,
		X:    x,
		Y:    y,
	}
}

// SetGenre is a package-level function for legacy compatibility.
func SetGenre(genreID string) {}
