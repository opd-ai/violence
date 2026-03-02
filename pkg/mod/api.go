// Package mod provides the mod API for WASM modules.
package mod

import (
	"fmt"

	"github.com/opd-ai/violence/pkg/engine"
)

// AudioEngine interface for playing sounds
type AudioEngine interface {
	PlaySFX(name string, x, y float64) error
}

// SpriteGenerator interface for procedural sprite generation
type SpriteGenerator interface {
	// Marker interface - actual sprite generation happens in game code
}

// ModAPI defines the interface exposed to WASM mods.
// All mod interactions with the game go through this API.
type ModAPI struct {
	modName string

	// Permissions control what the mod can access
	permissions ModPermissions

	// Event handlers registered by the mod
	eventHandlers map[string][]EventHandler

	// Game system references (set when API is bound to game)
	world          *engine.World
	audioEngine    AudioEngine
	spriteGen      SpriteGenerator
	hudMessage     *string
	hudMessageTime *int
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

// BindGameSystems attaches game system references to the ModAPI.
// This enables the mod to interact with the game world.
func (api *ModAPI) BindGameSystems(world *engine.World, audioEngine AudioEngine, spriteGen SpriteGenerator, hudMessage *string, hudMessageTime *int) {
	api.world = world
	api.audioEngine = audioEngine
	api.spriteGen = spriteGen
	api.hudMessage = hudMessage
	api.hudMessageTime = hudMessageTime
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
// entityType must be one of: "enemy", "prop", "pickup", "projectile"
func (api *ModAPI) SpawnEntity(entityType string, x, y float64) (EntityID, error) {
	if !api.permissions.AllowEntitySpawn {
		return 0, fmt.Errorf("permission denied: entity spawn not allowed for mod %s", api.modName)
	}

	if api.world == nil {
		return 0, fmt.Errorf("mod API not bound to game world")
	}

	// Create entity and add Position component
	e := api.world.AddEntity()
	api.world.AddComponent(e, &engine.Position{X: x, Y: y})

	// Add type-specific components
	switch entityType {
	case "enemy":
		api.world.AddComponent(e, &engine.Health{Current: 50, Max: 50})
		api.world.AddComponent(e, &engine.Velocity{DX: 0, DY: 0})
	case "prop":
		// Props are static decorations - no additional components needed
	case "pickup":
		// Pickups need no additional components, handled by game logic
	case "projectile":
		api.world.AddComponent(e, &engine.Velocity{DX: 0, DY: 0})
	default:
		api.world.RemoveEntity(e)
		return 0, fmt.Errorf("unknown entity type: %s (must be enemy, prop, pickup, or projectile)", entityType)
	}

	return EntityID(e), nil
}

// LoadTexture loads a texture from the mod's assets.
// Requires AllowAssetLoad permission.
// Returns a texture ID that can be used in rendering.
// path is a procedural generation key (e.g., "enemy:goblin:1234" for sprite generation)
func (api *ModAPI) LoadTexture(path string) (TextureID, error) {
	if !api.permissions.AllowAssetLoad {
		return 0, fmt.Errorf("permission denied: asset load not allowed for mod %s", api.modName)
	}

	if api.spriteGen == nil {
		return 0, fmt.Errorf("mod API not bound to sprite generator")
	}

	// Parse path format: "type:subtype:seed"
	// For now, return a hash of the path as the texture ID
	// The actual sprite will be generated on-demand by the game's sprite system
	var hash uint32
	for i := 0; i < len(path); i++ {
		hash = hash*31 + uint32(path[i])
	}

	return TextureID(hash), nil
}

// PlaySound plays a sound effect.
// Requires AllowAssetLoad permission.
// soundID should be generated using procedural sound synthesis params
func (api *ModAPI) PlaySound(soundID SoundID) error {
	if !api.permissions.AllowAssetLoad {
		return fmt.Errorf("permission denied: asset load not allowed for mod %s", api.modName)
	}

	if api.audioEngine == nil {
		return fmt.Errorf("mod API not bound to audio engine")
	}

	// Convert soundID to SFX name for procedural generation
	// The audio engine will generate the sound procedurally
	sfxName := fmt.Sprintf("mod_sfx_%d", soundID)
	return api.audioEngine.PlaySFX(sfxName, 0, 0)
}

// ShowNotification displays a notification message to the player.
// Requires AllowUIModify permission.
// Message is displayed on the HUD for 3 seconds (180 frames at 60 FPS)
func (api *ModAPI) ShowNotification(message string) error {
	if !api.permissions.AllowUIModify {
		return fmt.Errorf("permission denied: UI modify not allowed for mod %s", api.modName)
	}

	if api.hudMessage == nil || api.hudMessageTime == nil {
		return fmt.Errorf("mod API not bound to HUD system")
	}

	*api.hudMessage = message
	*api.hudMessageTime = 180 // 3 seconds at 60 FPS

	return nil
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
