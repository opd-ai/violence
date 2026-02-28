// Package mod provides plugin API for game mods.
//
// The mod package enables mods to extend game behavior through a plugin system with:
//
//  1. Plugin Interface: All mods implement the Plugin interface with Load/Unload lifecycle.
//  2. Hook System: Plugins register callbacks for game events (weapon fire, enemy spawn, etc).
//  3. Generator Registry: Plugins provide custom procedural content generators.
//
// Example Usage:
//
//	// Create a plugin
//	type MyPlugin struct{}
//	func (p *MyPlugin) Load() error { return nil }
//	func (p *MyPlugin) Unload() error { return nil }
//	func (p *MyPlugin) Name() string { return "MyPlugin" }
//	func (p *MyPlugin) Version() string { return "1.0.0" }
//
//	// Load plugin via mod loader
//	loader := mod.NewLoader()
//	loader.RegisterPlugin(&MyPlugin{})
//
//	// Register a hook to modify weapon fire behavior
//	loader.PluginManager().Hooks().Register(mod.HookTypeWeaponFire, func(data interface{}) error {
//		weaponData := data.(map[string]interface{})
//		weaponData["damage"] = weaponData["damage"].(int) * 2  // Double damage
//		return nil
//	})
//
//	// Register a custom generator
//	type MyWeaponGenerator struct{}
//	func (g *MyWeaponGenerator) Type() string { return "weapon" }
//	func (g *MyWeaponGenerator) Generate(seed int64, params map[string]interface{}) (interface{}, error) {
//		// Custom weapon generation logic
//		return customWeapon, nil
//	}
//	loader.PluginManager().Generators().Register(&MyWeaponGenerator{})
package mod

import (
	"fmt"
	"sync"
)

// Plugin defines the interface that all mods must implement.
// Plugins can hook into game systems, register custom content, and respond to events.
type Plugin interface {
	// Load is called when the plugin is loaded. Return error to abort loading.
	Load() error

	// Unload is called when the plugin is unloaded. Clean up resources here.
	Unload() error

	// Name returns the unique identifier for this plugin.
	Name() string

	// Version returns the plugin version string.
	Version() string
}

// HookType identifies different game extension points.
type HookType string

const (
	// HookTypeWeaponFire is called when a weapon fires.
	HookTypeWeaponFire HookType = "weapon.fire"

	// HookTypeEnemySpawn is called when an enemy spawns.
	HookTypeEnemySpawn HookType = "enemy.spawn"

	// HookTypePlayerDamage is called when player takes damage.
	HookTypePlayerDamage HookType = "player.damage"

	// HookTypeLevelGenerate is called during level generation.
	HookTypeLevelGenerate HookType = "level.generate"

	// HookTypeItemPickup is called when player picks up an item.
	HookTypeItemPickup HookType = "item.pickup"

	// HookTypeDoorOpen is called when a door opens.
	HookTypeDoorOpen HookType = "door.open"

	// HookTypeGenreSet is called when genre is changed.
	HookTypeGenreSet HookType = "genre.set"
)

// HookFunc is a callback function for game events.
// The data parameter contains event-specific data.
// Return error to indicate hook processing failed.
type HookFunc func(data interface{}) error

// HookRegistry manages plugin hooks and callbacks.
type HookRegistry struct {
	hooks map[HookType][]HookFunc
	mu    sync.RWMutex
}

// NewHookRegistry creates a new hook registry.
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		hooks: make(map[HookType][]HookFunc),
	}
}

// Register adds a hook callback for the given event type.
func (r *HookRegistry) Register(hookType HookType, fn HookFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[hookType] = append(r.hooks[hookType], fn)
}

// Unregister removes all hooks for the given event type.
func (r *HookRegistry) Unregister(hookType HookType) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.hooks, hookType)
}

// Trigger executes all registered hooks for the given event type.
// If any hook returns an error, execution stops and that error is returned.
func (r *HookRegistry) Trigger(hookType HookType, data interface{}) error {
	r.mu.RLock()
	hooks := r.hooks[hookType]
	r.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook(data); err != nil {
			return fmt.Errorf("hook %s failed: %w", hookType, err)
		}
	}
	return nil
}

// List returns all registered hook types.
func (r *HookRegistry) List() []HookType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]HookType, 0, len(r.hooks))
	for hookType := range r.hooks {
		types = append(types, hookType)
	}
	return types
}

// Count returns the number of callbacks registered for a hook type.
func (r *HookRegistry) Count(hookType HookType) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.hooks[hookType])
}

// Clear removes all registered hooks.
func (r *HookRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks = make(map[HookType][]HookFunc)
}

// Generator defines procedural content generation parameters.
// Plugins can register custom generators to override default behavior.
type Generator interface {
	// Type returns the generator type (e.g., "weapon", "enemy", "texture").
	Type() string

	// Generate creates content using the given seed and parameters.
	// Returns generated data or error.
	Generate(seed int64, params map[string]interface{}) (interface{}, error)
}

// GeneratorRegistry manages custom content generators.
type GeneratorRegistry struct {
	generators map[string]Generator
	mu         sync.RWMutex
}

// NewGeneratorRegistry creates a new generator registry.
func NewGeneratorRegistry() *GeneratorRegistry {
	return &GeneratorRegistry{
		generators: make(map[string]Generator),
	}
}

// Register adds a custom generator. Returns error if type already registered.
func (r *GeneratorRegistry) Register(gen Generator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	genType := gen.Type()
	if _, exists := r.generators[genType]; exists {
		return fmt.Errorf("generator type %s already registered", genType)
	}

	r.generators[genType] = gen
	return nil
}

// Unregister removes a generator by type.
func (r *GeneratorRegistry) Unregister(genType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.generators, genType)
}

// Get retrieves a generator by type. Returns nil if not found.
func (r *GeneratorRegistry) Get(genType string) Generator {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.generators[genType]
}

// List returns all registered generator types.
func (r *GeneratorRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.generators))
	for genType := range r.generators {
		types = append(types, genType)
	}
	return types
}

// Clear removes all registered generators.
func (r *GeneratorRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.generators = make(map[string]Generator)
}

// PluginManager coordinates plugin lifecycle and provides access to registries.
type PluginManager struct {
	plugins    map[string]Plugin
	hooks      *HookRegistry
	generators *GeneratorRegistry
	mu         sync.RWMutex
}

// NewPluginManager creates a new plugin manager.
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins:    make(map[string]Plugin),
		hooks:      NewHookRegistry(),
		generators: NewGeneratorRegistry(),
	}
}

// LoadPlugin loads and initializes a plugin.
func (m *PluginManager) LoadPlugin(p Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := p.Name()
	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin %s already loaded", name)
	}

	if err := p.Load(); err != nil {
		return fmt.Errorf("plugin %s load failed: %w", name, err)
	}

	m.plugins[name] = p
	return nil
}

// UnloadPlugin unloads a plugin by name.
func (m *PluginManager) UnloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not loaded", name)
	}

	if err := p.Unload(); err != nil {
		return fmt.Errorf("plugin %s unload failed: %w", name, err)
	}

	delete(m.plugins, name)
	return nil
}

// GetPlugin retrieves a plugin by name.
func (m *PluginManager) GetPlugin(name string) (Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, exists := m.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}
	return p, nil
}

// ListPlugins returns all loaded plugin names.
func (m *PluginManager) ListPlugins() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		names = append(names, name)
	}
	return names
}

// Hooks returns the hook registry.
func (m *PluginManager) Hooks() *HookRegistry {
	return m.hooks
}

// Generators returns the generator registry.
func (m *PluginManager) Generators() *GeneratorRegistry {
	return m.generators
}

// UnloadAll unloads all plugins.
func (m *PluginManager) UnloadAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	for name, p := range m.plugins {
		if err := p.Unload(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("plugin %s unload failed: %w", name, err)
		}
	}

	m.plugins = make(map[string]Plugin)
	m.hooks.Clear()
	m.generators.Clear()

	return firstErr
}
