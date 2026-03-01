// Package save provides save data schema and persistence functionality.
package save

// SaveState represents the canonical save data schema matching player entity components.
// This schema is designed to persist all critical player state for save/load functionality.
type SaveState struct {
	// LevelSeed is the procedural generation seed for the current level
	LevelSeed int64 `json:"level_seed"`

	// PlayerPosition stores the player's X, Y coordinates in world space
	PlayerPosition Position `json:"player_position"`

	// Health stores current and maximum health values
	Health HealthData `json:"health"`

	// Armor stores armor value for damage reduction
	Armor int `json:"armor"`

	// Inventory contains items and credits
	Inventory InventoryData `json:"inventory"`

	// DiscoveredTiles is a set of tile coordinates the player has explored
	// Key format: "x,y" for efficient lookup and serialization
	DiscoveredTiles map[string]bool `json:"discovered_tiles"`

	// CurrentObjective stores the active quest/objective identifier
	CurrentObjective string `json:"current_objective"`

	// CameraDirection stores the player's view direction and FOV
	CameraDirection CameraData `json:"camera_direction"`
}

// Position represents a 2D coordinate in world space.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// HealthData stores health point data.
type HealthData struct {
	Current int `json:"current"`
	Max     int `json:"max"`
}

// InventoryData stores inventory items and currency.
type InventoryData struct {
	Items   []string `json:"items"`
	Credits int      `json:"credits"`
}

// CameraData stores camera/view state.
type CameraData struct {
	DirX         float64 `json:"dir_x"`
	DirY         float64 `json:"dir_y"`
	PlaneX       float64 `json:"plane_x"`
	PlaneY       float64 `json:"plane_y"`
	FOV          float64 `json:"fov"`
	PitchRadians float64 `json:"pitch_radians"`
}
