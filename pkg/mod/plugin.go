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

	// GetGenerators returns custom procedural content generators provided by this plugin.
	// Return nil or empty slice if no custom generators are provided.
	GetGenerators() []Generator
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
	ownership  map[string]string // generator type -> plugin name that registered it
	mu         sync.RWMutex
}

// NewGeneratorRegistry creates a new generator registry.
func NewGeneratorRegistry() *GeneratorRegistry {
	return &GeneratorRegistry{
		generators: make(map[string]Generator),
		ownership:  make(map[string]string),
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

// RegisterFrom adds a generator with ownership tracking. Returns error if type
// already registered by a different plugin. Used during plugin loading.
func (r *GeneratorRegistry) RegisterFrom(gen Generator, pluginName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	genType := gen.Type()
	if owner, exists := r.ownership[genType]; exists {
		return fmt.Errorf("generator type %q conflicts: already registered by %q", genType, owner)
	}

	r.generators[genType] = gen
	r.ownership[genType] = pluginName
	return nil
}

// ForceRegister overrides an existing generator, returning the previous owner
// name if one existed (empty string if none). This is used when a higher-priority
// mod overrides a generator.
func (r *GeneratorRegistry) ForceRegister(gen Generator, pluginName string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	genType := gen.Type()
	prevOwner := r.ownership[genType]

	r.generators[genType] = gen
	r.ownership[genType] = pluginName
	return prevOwner
}

// Owner returns the plugin name that registered a generator type.
func (r *GeneratorRegistry) Owner(genType string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ownership[genType]
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
	r.ownership = make(map[string]string)
}

// PluginManager coordinates plugin lifecycle and provides access to registries.
type PluginManager struct {
	plugins    map[string]Plugin
	hooks      *HookRegistry
	generators *GeneratorRegistry
	overrides  *ParamRegistry
	warnings   []string
	mu         sync.RWMutex
}

// NewPluginManager creates a new plugin manager.
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins:    make(map[string]Plugin),
		hooks:      NewHookRegistry(),
		generators: NewGeneratorRegistry(),
		overrides:  NewParamRegistry(),
		warnings:   make([]string, 0),
	}
}

// LoadPlugin loads and initializes a plugin.
// If the plugin provides generators via GetGenerators(), they are automatically
// registered. Generator type conflicts produce a warning but do not prevent loading.
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

	// Auto-register generators from plugin
	gens := p.GetGenerators()
	for _, gen := range gens {
		if err := m.generators.RegisterFrom(gen, name); err != nil {
			m.warnings = append(m.warnings, fmt.Sprintf("plugin %s: %s", name, err.Error()))
		}
	}

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

// Overrides returns the parameter override registry.
func (m *PluginManager) Overrides() *ParamRegistry {
	return m.overrides
}

// Warnings returns all accumulated warning messages.
func (m *PluginManager) Warnings() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.warnings))
	copy(result, m.warnings)
	return result
}

// ClearWarnings removes all warnings.
func (m *PluginManager) ClearWarnings() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnings = m.warnings[:0]
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
	m.overrides.Clear()
	m.warnings = m.warnings[:0]

	return firstErr
}

// ParamOverride represents a mod's override for a generation parameter.
// When multiple mods override the same parameter, the highest priority wins.
type ParamOverride struct {
	// GeneratorType is the type of generator this override applies to (e.g., "weapon", "enemy").
	GeneratorType string

	// Key is the specific parameter to override (e.g., "damage_multiplier", "health").
	Key string

	// Value is the override value.
	Value interface{}

	// Priority determines precedence when multiple mods override the same parameter.
	// Higher values take priority. Default is 0.
	Priority int

	// ModName is the name of the mod that registered this override.
	ModName string
}

// ParamRegistry manages generation parameter overrides from mods.
type ParamRegistry struct {
	// overrides maps generatorType -> key -> []ParamOverride (sorted by priority)
	overrides map[string]map[string][]ParamOverride
	mu        sync.RWMutex
}

// NewParamRegistry creates a new parameter override registry.
func NewParamRegistry() *ParamRegistry {
	return &ParamRegistry{
		overrides: make(map[string]map[string][]ParamOverride),
	}
}

// Register adds a parameter override. If another override exists for the same
// generator type and key from a different mod, both are stored and resolved by priority.
func (r *ParamRegistry) Register(override ParamOverride) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.overrides[override.GeneratorType]; !ok {
		r.overrides[override.GeneratorType] = make(map[string][]ParamOverride)
	}

	key := override.Key
	overrides := r.overrides[override.GeneratorType][key]

	// Insert sorted by priority descending
	inserted := false
	for i, existing := range overrides {
		if override.Priority > existing.Priority {
			overrides = append(overrides[:i+1], overrides[i:]...)
			overrides[i] = override
			inserted = true
			break
		}
	}
	if !inserted {
		overrides = append(overrides, override)
	}

	r.overrides[override.GeneratorType][key] = overrides
}

// Get returns the highest-priority override value for a generator type and key.
// Returns the value and true if found, or nil and false if no override exists.
func (r *ParamRegistry) Get(generatorType, key string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if keys, ok := r.overrides[generatorType]; ok {
		if overrides, ok := keys[key]; ok && len(overrides) > 0 {
			return overrides[0].Value, true
		}
	}
	return nil, false
}

// GetAll returns all overrides for a generator type, keyed by parameter name.
// Each key maps to the highest-priority override value.
func (r *ParamRegistry) GetAll(generatorType string) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]interface{})
	if keys, ok := r.overrides[generatorType]; ok {
		for key, overrides := range keys {
			if len(overrides) > 0 {
				result[key] = overrides[0].Value
			}
		}
	}
	return result
}

// GetOverrides returns all override entries for a generator type and key,
// sorted by priority (highest first).
func (r *ParamRegistry) GetOverrides(generatorType, key string) []ParamOverride {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if keys, ok := r.overrides[generatorType]; ok {
		if overrides, ok := keys[key]; ok {
			result := make([]ParamOverride, len(overrides))
			copy(result, overrides)
			return result
		}
	}
	return nil
}

// RemoveByMod removes all overrides registered by a specific mod.
func (r *ParamRegistry) RemoveByMod(modName string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	removed := 0
	for genType, keys := range r.overrides {
		for key, overrides := range keys {
			filtered := overrides[:0]
			for _, o := range overrides {
				if o.ModName != modName {
					filtered = append(filtered, o)
				} else {
					removed++
				}
			}
			if len(filtered) == 0 {
				delete(keys, key)
			} else {
				r.overrides[genType][key] = filtered
			}
		}
		if len(keys) == 0 {
			delete(r.overrides, genType)
		}
	}
	return removed
}

// ListTypes returns all generator types that have overrides.
func (r *ParamRegistry) ListTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.overrides))
	for genType := range r.overrides {
		types = append(types, genType)
	}
	return types
}

// Clear removes all parameter overrides.
func (r *ParamRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.overrides = make(map[string]map[string][]ParamOverride)
}

// ApplyOverrides merges parameter overrides into a params map for a given generator type.
// Existing params are overwritten by override values. Returns the merged map.
func (r *ParamRegistry) ApplyOverrides(generatorType string, params map[string]interface{}) map[string]interface{} {
	overrides := r.GetAll(generatorType)
	if len(overrides) == 0 {
		return params
	}

	if params == nil {
		params = make(map[string]interface{})
	}

	result := make(map[string]interface{}, len(params)+len(overrides))
	for k, v := range params {
		result[k] = v
	}
	for k, v := range overrides {
		result[k] = v
	}
	return result
}
