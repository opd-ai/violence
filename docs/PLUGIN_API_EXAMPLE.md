# Plugin API Example

This document demonstrates how to create and use plugins with the Violence mod system.

## Basic Plugin Example

```go
package main

import (
	"fmt"
	"github.com/opd-ai/violence/pkg/mod"
)

// ExamplePlugin demonstrates basic plugin implementation.
type ExamplePlugin struct {
	name    string
	version string
}

// NewExamplePlugin creates a new example plugin.
func NewExamplePlugin() *ExamplePlugin {
	return &ExamplePlugin{
		name:    "ExampleMod",
		version: "1.0.0",
	}
}

// Load is called when the plugin is loaded.
func (p *ExamplePlugin) Load() error {
	fmt.Println("Example plugin loaded!")
	return nil
}

// Unload is called when the plugin is unloaded.
func (p *ExamplePlugin) Unload() error {
	fmt.Println("Example plugin unloaded!")
	return nil
}

// Name returns the plugin name.
func (p *ExamplePlugin) Name() string {
	return p.name
}

// Version returns the plugin version.
func (p *ExamplePlugin) Version() string {
	return p.version
}
```

## Using Hooks

Plugins can register callbacks for game events:

```go
func main() {
	// Create mod loader
	loader := mod.NewLoader()
	
	// Load plugin
	plugin := NewExamplePlugin()
	if err := loader.RegisterPlugin(plugin); err != nil {
		panic(err)
	}
	
	// Register hook for weapon fire events
	loader.PluginManager().Hooks().Register(mod.HookTypeWeaponFire, func(data interface{}) error {
		weaponData := data.(map[string]interface{})
		fmt.Printf("Weapon fired: %v\n", weaponData)
		
		// Modify weapon behavior (e.g., double damage)
		if damage, ok := weaponData["damage"].(int); ok {
			weaponData["damage"] = damage * 2
		}
		
		return nil
	})
	
	// Register hook for enemy spawn events
	loader.PluginManager().Hooks().Register(mod.HookTypeEnemySpawn, func(data interface{}) error {
		enemyData := data.(map[string]interface{})
		fmt.Printf("Enemy spawned: %v\n", enemyData)
		
		// Modify enemy stats (e.g., increase health)
		if health, ok := enemyData["health"].(int); ok {
			enemyData["health"] = health + 50
		}
		
		return nil
	})
}
```

## Available Hook Types

| Hook Type                  | When Triggered                | Example Data                                    |
|---------------------------|-------------------------------|------------------------------------------------|
| `HookTypeWeaponFire`      | When a weapon fires           | `{"weapon": "pistol", "damage": 10}`           |
| `HookTypeEnemySpawn`      | When an enemy spawns          | `{"type": "guard", "health": 100}`             |
| `HookTypePlayerDamage`    | When player takes damage      | `{"damage": 25, "source": "enemy"}`            |
| `HookTypeLevelGenerate`   | During level generation       | `{"seed": 12345, "size": 50}`                  |
| `HookTypeItemPickup`      | When player picks up item     | `{"item": "health_pack", "value": 25}`         |
| `HookTypeDoorOpen`        | When a door opens             | `{"type": "standard", "locked": false}`        |
| `HookTypeGenreSet`        | When genre changes            | `{"genre": "scifi"}`                           |

## Custom Generators

Plugins can provide custom procedural content generators:

```go
// CustomWeaponGenerator generates custom weapons.
type CustomWeaponGenerator struct{}

// Type returns the generator type.
func (g *CustomWeaponGenerator) Type() string {
	return "weapon"
}

// Generate creates a custom weapon using the given seed and parameters.
func (g *CustomWeaponGenerator) Generate(seed int64, params map[string]interface{}) (interface{}, error) {
	// Use seed for deterministic generation
	rng := rand.New(rand.NewSource(seed))
	
	weapon := map[string]interface{}{
		"name":     "Custom Laser Rifle",
		"damage":   20 + rng.Intn(10),
		"fireRate": 0.5,
		"ammoType": "energy",
	}
	
	// Apply parameters from caller
	if genreName, ok := params["genre"].(string); ok {
		weapon["genre"] = genreName
	}
	
	return weapon, nil
}

// Register the generator
func main() {
	loader := mod.NewLoader()
	
	gen := &CustomWeaponGenerator{}
	if err := loader.PluginManager().Generators().Register(gen); err != nil {
		panic(err)
	}
	
	// Use the generator
	result, err := gen.Generate(12345, map[string]interface{}{
		"genre": "scifi",
	})
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("Generated weapon: %v\n", result)
}
```

## Complete Plugin Example

```go
package main

import (
	"fmt"
	"github.com/opd-ai/violence/pkg/mod"
	"math/rand"
)

// DoubleDamagePlugin doubles all weapon damage.
type DoubleDamagePlugin struct{}

func (p *DoubleDamagePlugin) Load() error {
	fmt.Println("DoubleDamage plugin loaded")
	return nil
}

func (p *DoubleDamagePlugin) Unload() error {
	fmt.Println("DoubleDamage plugin unloaded")
	return nil
}

func (p *DoubleDamagePlugin) Name() string {
	return "DoubleDamage"
}

func (p *DoubleDamagePlugin) Version() string {
	return "1.0.0"
}

func main() {
	// Initialize loader
	loader := mod.NewLoader()
	
	// Load plugin
	plugin := &DoubleDamagePlugin{}
	if err := loader.RegisterPlugin(plugin); err != nil {
		panic(err)
	}
	
	// Register weapon fire hook
	loader.PluginManager().Hooks().Register(mod.HookTypeWeaponFire, func(data interface{}) error {
		weaponData := data.(map[string]interface{})
		if damage, ok := weaponData["damage"].(int); ok {
			weaponData["damage"] = damage * 2
			fmt.Printf("Doubled damage: %d -> %d\n", damage, weaponData["damage"])
		}
		return nil
	})
	
	// Simulate weapon fire event
	testData := map[string]interface{}{
		"weapon": "pistol",
		"damage": 10,
	}
	
	fmt.Println("Before hook:", testData)
	loader.PluginManager().Hooks().Trigger(mod.HookTypeWeaponFire, testData)
	fmt.Println("After hook:", testData)
}
```

## Thread Safety

All plugin API components are thread-safe:

- `HookRegistry` uses mutex locks for concurrent hook registration and triggering
- `GeneratorRegistry` uses mutex locks for concurrent generator registration
- `PluginManager` uses mutex locks for concurrent plugin loading/unloading

Plugins can safely register hooks and generators from multiple goroutines.

## Best Practices

1. **Error Handling**: Always return meaningful errors from hook functions
2. **Type Safety**: Validate data types before casting in hook functions
3. **Determinism**: Use the provided seed parameter in generators for reproducible results
4. **Cleanup**: Properly cleanup resources in the Unload() method
5. **Documentation**: Document your plugin's hooks and generators clearly
6. **Testing**: Write tests for your plugin using the mock implementations in `plugin_test.go`
