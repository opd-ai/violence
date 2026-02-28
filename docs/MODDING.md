# Modding Guide

VIOLENCE provides a plugin API for extending game behavior through mods. Mods can hook into game events, register custom procedural content generators, and override generation parameters.

## Mod Structure

Each mod lives in its own directory under `mods/`:

```
mods/
  my-mod/
    mod.json        Mod manifest (required)
    plugin.so       Compiled Go plugin (Linux/macOS)
    plugin.dll      Compiled Go plugin (Windows)
```

### Mod Manifest (`mod.json`)

Every mod directory must contain a `mod.json` file:

```json
{
    "name": "my-mod",
    "version": "1.0.0",
    "description": "A custom mod that doubles weapon damage",
    "author": "YourName",
    "config": {
        "damage_multiplier": "2.0",
        "custom_option": "enabled"
    }
}
```

**Required fields:**
- `name` — Unique mod identifier (used for conflict detection)
- `version` — Semantic version string

**Optional fields:**
- `description` — Human-readable description
- `author` — Mod author
- `config` — Arbitrary key-value configuration map (string→string)

## Plugin Interface

All mods implement the `Plugin` interface from `pkg/mod`:

```go
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
```

### Example Plugin

```go
package main

import "fmt"

// MyPlugin implements the mod.Plugin interface.
type MyPlugin struct {
    loaded bool
}

func (p *MyPlugin) Load() error {
    p.loaded = true
    fmt.Println("MyPlugin loaded")
    return nil
}

func (p *MyPlugin) Unload() error {
    p.loaded = false
    fmt.Println("MyPlugin unloaded")
    return nil
}

func (p *MyPlugin) Name() string    { return "my-plugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }

// Plugin is the exported symbol that the mod loader looks for.
var Plugin MyPlugin
```

Build the plugin:

```sh
go build -buildmode=plugin -o mods/my-mod/plugin.so ./mods/my-mod/
```

## Hook System

Plugins register callbacks for game events through the `HookRegistry`. Hooks intercept and modify game behavior at defined extension points.

### Available Hook Types

| Hook Type | Constant | Description | Data Type |
| --------- | -------- | ----------- | --------- |
| Weapon Fire | `HookTypeWeaponFire` | Called when a weapon fires | `map[string]interface{}` with `"damage"`, `"weapon"` |
| Enemy Spawn | `HookTypeEnemySpawn` | Called when an enemy spawns | `map[string]interface{}` with `"enemy_type"`, `"position"` |
| Player Damage | `HookTypePlayerDamage` | Called when player takes damage | `map[string]interface{}` with `"damage"`, `"source"` |
| Level Generate | `HookTypeLevelGenerate` | Called during level generation | `map[string]interface{}` with `"seed"`, `"width"`, `"height"` |
| Item Pickup | `HookTypeItemPickup` | Called when player picks up an item | `map[string]interface{}` with `"item"`, `"quantity"` |
| Door Open | `HookTypeDoorOpen` | Called when a door opens | `map[string]interface{}` with `"door_id"`, `"locked"` |
| Genre Set | `HookTypeGenreSet` | Called when genre is changed | `map[string]interface{}` with `"genre_id"` |

### Registering Hooks

```go
func (p *MyPlugin) Load() error {
    // Get the plugin manager from the mod loader
    // Hooks are registered during Load()

    return nil
}

// Using the hook registry directly:
hookRegistry := loader.PluginManager().Hooks()

// Double all weapon damage
hookRegistry.Register(mod.HookTypeWeaponFire, func(data interface{}) error {
    weaponData := data.(map[string]interface{})
    if dmg, ok := weaponData["damage"].(int); ok {
        weaponData["damage"] = dmg * 2
    }
    return nil
})

// Log all enemy spawns
hookRegistry.Register(mod.HookTypeEnemySpawn, func(data interface{}) error {
    spawnData := data.(map[string]interface{})
    fmt.Printf("Enemy spawned: %v\n", spawnData["enemy_type"])
    return nil
})
```

### Hook Execution

- Hooks execute in registration order.
- If any hook returns an error, execution stops and the error propagates.
- Hook data is passed by reference — modifications are visible to subsequent hooks and the caller.

### Hook Management

```go
hooks := loader.PluginManager().Hooks()

// List all registered hook types
types := hooks.List()

// Count callbacks for a hook type
count := hooks.Count(mod.HookTypeWeaponFire)

// Remove all callbacks for a hook type
hooks.Unregister(mod.HookTypeWeaponFire)

// Remove all hooks
hooks.Clear()
```

## Generator System

Plugins can register custom procedural content generators that override or extend the default generation.

### Generator Interface

```go
type Generator interface {
    // Type returns the generator type (e.g., "weapon", "enemy", "texture").
    Type() string

    // Generate creates content using the given seed and parameters.
    // Returns generated data or error.
    Generate(seed int64, params map[string]interface{}) (interface{}, error)
}
```

### Example: Custom Weapon Generator

```go
type PlasmaWeaponGenerator struct{}

func (g *PlasmaWeaponGenerator) Type() string { return "weapon" }

func (g *PlasmaWeaponGenerator) Generate(seed int64, params map[string]interface{}) (interface{}, error) {
    rng := rand.New(rand.NewSource(seed))

    weapon := map[string]interface{}{
        "name":       fmt.Sprintf("Plasma Rifle Mk-%d", rng.Intn(100)),
        "damage":     30 + rng.Intn(20),
        "fire_rate":  0.5 + rng.Float64(),
        "projectile": "plasma",
    }
    return weapon, nil
}

// Register the generator
registry := loader.PluginManager().Generators()
err := registry.Register(&PlasmaWeaponGenerator{})
```

### Generator Management

```go
generators := loader.PluginManager().Generators()

// Register a generator (error if type already registered)
err := generators.Register(myGenerator)

// Retrieve a generator by type
gen := generators.Get("weapon")
if gen != nil {
    result, err := gen.Generate(seed, params)
}

// List all registered generator types
types := generators.List()

// Remove a specific generator
generators.Unregister("weapon")

// Remove all generators
generators.Clear()
```

## Mod Loader API

The `Loader` (`pkg/mod`) manages mod discovery, loading, and conflict detection.

### Loading Mods

```go
loader := mod.NewLoader()

// Load a mod from a directory (reads mod.json)
err := loader.LoadMod("mods/my-mod")

// Load with custom mods directory
loader = mod.NewLoaderWithDir("/path/to/mods")

// List all loaded mods
mods := loader.ListMods()

// Get a specific mod
m, err := loader.GetMod("my-mod")
```

### Enabling and Disabling

```go
// Disable a mod (remains loaded but inactive)
err := loader.DisableMod("my-mod")

// Re-enable a mod
err := loader.EnableMod("my-mod")

// Unload a mod entirely
err := loader.UnloadMod("my-mod")
```

### Conflict Detection

Mods can declare conflicts with other mods. Conflicting mods cannot be loaded simultaneously:

```go
// Register that mod-a and mod-b conflict
loader.AddConflict("mod-a", "mod-b")

// Now loading mod-b after mod-a (or vice versa) returns an error
loader.LoadMod("mods/mod-a") // succeeds
loader.LoadMod("mods/mod-b") // returns error: "mod mod-b conflicts with mod-a"
```

Conflicts are bidirectional — declaring a conflict between A and B prevents loading either when the other is active.

### Plugin Manager Access

```go
pm := loader.PluginManager()

// Load a plugin directly
err := pm.LoadPlugin(myPlugin)

// Unload a plugin
err := pm.UnloadPlugin("my-plugin")

// Access registries
hooks := pm.Hooks()
generators := pm.Generators()

// List loaded plugins
names := pm.ListPlugins()

// Unload everything
err := pm.UnloadAll()
```

## Determinism Requirements

All generators **must** be deterministic:
- Use `rand.New(rand.NewSource(seed))` with the provided seed parameter.
- Never use `time.Now()`, global `math/rand`, or other non-deterministic sources.
- Identical seed + parameters must produce identical output across all platforms.

This ensures multiplayer synchronization and reproducible save/load behavior.

## Limitations

- **Go plugins only**: Mods are compiled Go plugins (`.so` on Linux/macOS, `.dll` on Windows). WASM and cross-platform plugin support is not yet available.
- **No sandboxing**: Go plugins have full runtime access. Only load trusted mods.
- **One generator per type**: Only one generator can be registered per type string. Registering a duplicate returns an error.
- **Build compatibility**: Plugins must be compiled with the same Go version and module dependencies as the main binary.
