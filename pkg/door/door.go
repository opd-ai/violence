// Package door implements keycard-locked doors with state-based animation.
package door

// DoorState represents the current state of a door.
type DoorState int

const (
	StateClosed  DoorState = iota // Door is fully closed
	StateOpening                  // Door is animating open
	StateOpen                     // Door is fully open
	StateClosing                  // Door is animating closed
)

// DoorType defines the visual style of a door.
type DoorType int

const (
	TypeSwing        DoorType = iota // Swinging door (fantasy)
	TypeSliding                      // Sliding bulkhead (scifi)
	TypePortcullis                   // Vertical gate (fantasy)
	TypeShutter                      // Mechanical shutter (cyberpunk)
	TypeLaserBarrier                 // Energy barrier (scifi)
)

// Door represents a door entity in the game world.
type Door struct {
	ID              string
	X               float64 // World position X
	Y               float64 // World position Y
	Type            DoorType
	State           DoorState
	Locked          bool
	RequiredKeycard string // Color of required keycard (e.g., "red", "blue", "yellow")
	AnimationFrame  int    // Current animation frame
	AnimationSpeed  int    // Frames per animation step (60 TPS)
	MaxFrames       int    // Total frames in open/close animation
}

// KeycardInventory tracks collected keycards.
type KeycardInventory struct {
	keycards map[string]bool // Color -> collected
}

// NewKeycardInventory creates a new empty keycard inventory.
func NewKeycardInventory() *KeycardInventory {
	return &KeycardInventory{
		keycards: make(map[string]bool),
	}
}

// AddKeycard adds a keycard to the inventory.
func (inv *KeycardInventory) AddKeycard(color string) {
	inv.keycards[color] = true
}

// HasKeycard checks if a specific color keycard is in inventory.
func (inv *KeycardInventory) HasKeycard(color string) bool {
	return inv.keycards[color]
}

// GetAll returns all collected keycard colors.
func (inv *KeycardInventory) GetAll() []string {
	colors := make([]string, 0, len(inv.keycards))
	for color := range inv.keycards {
		colors = append(colors, color)
	}
	return colors
}

// DoorSystem manages door state transitions and animation.
type DoorSystem struct {
	doors []*Door
}

// NewDoorSystem creates a new door management system.
func NewDoorSystem() *DoorSystem {
	return &DoorSystem{
		doors: make([]*Door, 0),
	}
}

// AddDoor registers a door with the system.
func (ds *DoorSystem) AddDoor(door *Door) {
	ds.doors = append(ds.doors, door)
}

// Update advances all door animations (call every frame at 60 TPS).
func (ds *DoorSystem) Update() {
	for _, door := range ds.doors {
		switch door.State {
		case StateOpening:
			door.AnimationFrame++
			if door.AnimationFrame >= door.MaxFrames {
				door.State = StateOpen
				door.AnimationFrame = door.MaxFrames
			}
		case StateClosing:
			door.AnimationFrame--
			if door.AnimationFrame <= 0 {
				door.State = StateClosed
				door.AnimationFrame = 0
			}
		}
	}
}

// TryOpen attempts to open a door using the keycard inventory.
// Returns (success, message) where message explains why door didn't open.
func (ds *DoorSystem) TryOpen(door *Door, inventory *KeycardInventory) (bool, string) {
	// Already open or opening
	if door.State == StateOpen || door.State == StateOpening {
		return false, "Door is already open"
	}

	// Check if locked
	if door.Locked {
		if door.RequiredKeycard == "" {
			return false, "Door is locked"
		}
		if !inventory.HasKeycard(door.RequiredKeycard) {
			return false, "Need " + door.RequiredKeycard + " keycard"
		}
		// Has keycard - unlock and open
		door.Locked = false
	}

	// Open the door
	door.State = StateOpening
	return true, ""
}

// Close closes an open door.
func (ds *DoorSystem) Close(door *Door) bool {
	if door.State == StateClosed || door.State == StateClosing {
		return false
	}
	door.State = StateClosing
	return true
}

// NewDoor creates a door with default animation parameters.
func NewDoor(id string, x, y float64, doorType DoorType, locked bool, requiredColor string) *Door {
	return &Door{
		ID:              id,
		X:               x,
		Y:               y,
		Type:            doorType,
		State:           StateClosed,
		Locked:          locked,
		RequiredKeycard: requiredColor,
		AnimationFrame:  0,
		AnimationSpeed:  2,  // 2 frames per animation step
		MaxFrames:       30, // 30 frames for full open (0.5 seconds at 60 TPS)
	}
}

// Genre-specific door/keycard name mappings
var genreKeycardNames = map[string]map[string]string{
	"fantasy": {
		"red":    "Crimson Rune",
		"blue":   "Sapphire Key",
		"yellow": "Golden Sigil",
	},
	"scifi": {
		"red":    "Red Access Card",
		"blue":   "Blue Clearance",
		"yellow": "Yellow Authorization",
	},
	"horror": {
		"red":    "Blood Key",
		"blue":   "Frozen Key",
		"yellow": "Rusty Key",
	},
	"cyberpunk": {
		"red":    "Red Biometric",
		"blue":   "Blue Data Chip",
		"yellow": "Yellow Override",
	},
	"postapoc": {
		"red":    "Red Tag",
		"blue":   "Blue Badge",
		"yellow": "Yellow Pass",
	},
}

var genreDoorTypes = map[string]DoorType{
	"fantasy":   TypePortcullis,
	"scifi":     TypeSliding,
	"horror":    TypeSwing,
	"cyberpunk": TypeShutter,
	"postapoc":  TypeSwing,
}

var currentGenre = "fantasy"

// SetGenre configures door/key themes for a genre.
func SetGenre(genreID string) {
	if _, ok := genreKeycardNames[genreID]; ok {
		currentGenre = genreID
	}
}

// GetKeycardDisplayName returns the genre-specific name for a keycard color.
func GetKeycardDisplayName(color string) string {
	if names, ok := genreKeycardNames[currentGenre]; ok {
		if name, found := names[color]; found {
			return name
		}
	}
	return color + " keycard" // Fallback
}

// GetGenreDoorType returns the default door type for the current genre.
func GetGenreDoorType() DoorType {
	if doorType, ok := genreDoorTypes[currentGenre]; ok {
		return doorType
	}
	return TypeSwing // Fallback
}

// Keycard represents a keycard that can open doors (legacy compatibility).
type Keycard struct {
	Color string
}

// TryOpen attempts to open a door with the given keycard (legacy function).
func TryOpen(d *Door, k *Keycard) bool {
	if !d.Locked {
		return true
	}
	return d.RequiredKeycard == k.Color
}
