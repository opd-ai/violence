// Package mod provides the mod API for WASM modules.
package mod

import (
	"fmt"
)

// ModAPI defines the interface exposed to WASM mods.
// All mod interactions with the game go through this API.
type ModAPI struct {
	modName string

	// Permissions control what the mod can access
	permissions ModPermissions

	// Event handlers registered by the mod
	eventHandlers map[string][]EventHandler
}

// ModPermissions defines what capabilities a mod has.
type ModPermissions struct {
	// AllowFileRead permits reading files in allowed directories
	AllowFileRead bool

	// AllowFileWrite permits writing files in allowed directories
	AllowFileWrite bool

	// AllowEntitySpawn permits spawning game entities
	AllowEntitySpawn bool

	// AllowAssetLoad permits loading textures and sounds
	AllowAssetLoad bool

	// AllowUIModify permits modifying UI elements
	AllowUIModify bool
}

// DefaultPermissions returns a safe default permission set.
// By default, mods can only register event handlers and read files.
func DefaultPermissions() ModPermissions {
	return ModPermissions{
		AllowFileRead:    true,
		AllowFileWrite:   false,
		AllowEntitySpawn: false,
		AllowAssetLoad:   true,
		AllowUIModify:    false,
	}
}

// EventHandler is a callback for game events.
type EventHandler func(data EventData) error

// EventData contains event-specific information passed to handlers.
type EventData struct {
	Type   string
	Params map[string]interface{}
}

// NewModAPI creates a new mod API instance for a specific mod.
func NewModAPI(modName string, permissions ModPermissions) *ModAPI {
	return &ModAPI{
		modName:       modName,
		permissions:   permissions,
		eventHandlers: make(map[string][]EventHandler),
	}
}

// RegisterEventHandler registers a handler for a game event type.
// The handler will be called when the event is triggered.
func (api *ModAPI) RegisterEventHandler(eventType string, handler EventHandler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	api.eventHandlers[eventType] = append(api.eventHandlers[eventType], handler)
	return nil
}

// TriggerEvent invokes all registered handlers for an event type.
// Used by the game engine to notify mods of events.
func (api *ModAPI) TriggerEvent(eventType string, data EventData) error {
	handlers := api.eventHandlers[eventType]

	for i, handler := range handlers {
		if err := handler(data); err != nil {
			return fmt.Errorf("event handler %d for %s failed: %w", i, eventType, err)
		}
	}

	return nil
}

// SpawnEntity spawns a game entity at the specified position.
// Requires AllowEntitySpawn permission.
func (api *ModAPI) SpawnEntity(entityType string, x, y float64) (EntityID, error) {
	if !api.permissions.AllowEntitySpawn {
		return 0, fmt.Errorf("permission denied: entity spawn not allowed for mod %s", api.modName)
	}

	// Stub implementation - will be connected to game ECS system
	return EntityID(0), fmt.Errorf("not implemented")
}

// LoadTexture loads a texture from the mod's assets.
// Requires AllowAssetLoad permission.
// Returns a texture ID that can be used in rendering.
func (api *ModAPI) LoadTexture(path string) (TextureID, error) {
	if !api.permissions.AllowAssetLoad {
		return 0, fmt.Errorf("permission denied: asset load not allowed for mod %s", api.modName)
	}

	// Stub implementation - will be connected to render system
	return TextureID(0), fmt.Errorf("not implemented")
}

// PlaySound plays a sound effect.
// Requires AllowAssetLoad permission.
func (api *ModAPI) PlaySound(soundID SoundID) error {
	if !api.permissions.AllowAssetLoad {
		return fmt.Errorf("permission denied: asset load not allowed for mod %s", api.modName)
	}

	// Stub implementation - will be connected to audio system
	return fmt.Errorf("not implemented")
}

// ShowNotification displays a notification message to the player.
// Requires AllowUIModify permission.
func (api *ModAPI) ShowNotification(message string) error {
	if !api.permissions.AllowUIModify {
		return fmt.Errorf("permission denied: UI modify not allowed for mod %s", api.modName)
	}

	// Stub implementation - will be connected to UI system
	return fmt.Errorf("not implemented")
}

// GetModName returns the name of the mod using this API.
func (api *ModAPI) GetModName() string {
	return api.modName
}

// GetPermissions returns the current permissions for this mod.
func (api *ModAPI) GetPermissions() ModPermissions {
	return api.permissions
}

// EntityID is a unique identifier for a spawned entity.
type EntityID uint64

// TextureID is a handle to a loaded texture.
type TextureID uint32

// SoundID is a handle to a loaded sound.
type SoundID uint32

// EventType constants for common game events.
const (
	EventTypeWeaponFire    = "weapon.fire"
	EventTypeEnemySpawn    = "enemy.spawn"
	EventTypeEnemyKilled   = "enemy.killed"
	EventTypePlayerDamage  = "player.damage"
	EventTypePlayerHeal    = "player.heal"
	EventTypeLevelGenerate = "level.generate"
	EventTypeLevelComplete = "level.complete"
	EventTypeItemPickup    = "item.pickup"
	EventTypeDoorOpen      = "door.open"
	EventTypeDoorClose     = "door.close"
	EventTypeGenreSet      = "genre.set"
	EventTypeModLoad       = "mod.load"
	EventTypeModUnload     = "mod.unload"
)
