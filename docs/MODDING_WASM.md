# WASM Mod Development Guide

## Overview

Violence v6.0+ uses WebAssembly (WASM) for safe, sandboxed mod execution. This guide covers creating, testing, and deploying WASM mods.

## Why WASM?

- **Security**: Mods run in isolated sandbox with no direct file system, network, or memory access
- **Cross-Platform**: Single `.wasm` binary works on all platforms (Windows, Linux, macOS, mobile, web)
- **Performance**: Near-native speed with JIT compilation
- **Version Stability**: Mods don't break on game updates

## Quick Start

### Prerequisites

- [TinyGo](https://tinygo.org/) 0.28+ (Go subset that compiles to WASM)
- Basic Go knowledge
- Violence v6.0+

### Creating Your First Mod

1. **Create mod directory structure**:
   ```bash
   mkdir -p mods/my_mod
   cd mods/my_mod
   ```

2. **Create `mod.json` manifest**:
   ```json
   {
     "name": "my_mod",
     "version": "1.0.0",
     "description": "My awesome mod",
     "author": "YourName"
   }
   ```

3. **Create `mod.go` with WASM entry point**:
   ```go
   package main

   import "unsafe"

   //export init
   func init() {
       logMessage("My mod loaded!")
   }

   //export on_weapon_fire
   func onWeaponFire(weaponID int32, damage int32) {
       logMessage("Weapon fired!")
   }

   // Host function imports
   //go:wasm-module env
   //export log_message
   func logMessage(message string)

   func main() {}
   ```

4. **Compile to WASM**:
   ```bash
   tinygo build -o mod.wasm -target wasi mod.go
   ```

5. **Load in game**:
   - Place `mod.wasm` and `mod.json` in `mods/my_mod/`
   - Game will auto-load on startup
   - Check logs for "My mod loaded!" message

## Security Model

### Default Permissions

WASM mods have **minimal permissions** by default:

- ✅ Register event handlers
- ✅ Read files in `mods/` directory
- ✅ Load textures/sounds
- ❌ Write files
- ❌ Spawn entities
- ❌ Modify UI
- ❌ Network access
- ❌ Access outside `mods/` directory

### Resource Limits

- **Memory**: 64MB max per mod
- **CPU**: 1 billion instructions per function call (prevents infinite loops)
- **File Access**: Restricted to `mods/` directory only

### Requesting Additional Permissions

In `mod.json`, request permissions:

```json
{
  "name": "my_mod",
  "version": "1.0.0",
  "permissions": {
    "allow_entity_spawn": true,
    "allow_ui_modify": true
  }
}
```

## Mod API Reference

### Event Handlers

Register callbacks for game events:

```go
//export on_weapon_fire
func onWeaponFire(weaponID int32, damage int32) {
    // Called when player fires weapon
}

//export on_enemy_spawn
func onEnemySpawn(enemyID int32, x float32, y float32) {
    // Called when enemy spawns
}

//export on_player_damage
func onPlayerDamage(damage int32, sourceID int32) {
    // Called when player takes damage
}
```

### Available Events

- `on_weapon_fire` - Weapon fired
- `on_enemy_spawn` - Enemy spawned
- `on_enemy_killed` - Enemy killed
- `on_player_damage` - Player damaged
- `on_player_heal` - Player healed
- `on_level_generate` - Level generated
- `on_item_pickup` - Item picked up
- `on_door_open` - Door opened
- `on_genre_set` - Genre changed

### Host Functions (Stubs - v6.0)

These functions are defined but not yet fully implemented in v6.0:

```go
//go:wasm-module env
//export log_message
func logMessage(message string)

//go:wasm-module env
//export spawn_entity
func spawnEntity(entityType string, x float32, y float32) int32

//go:wasm-module env
//export load_texture
func loadTexture(path string) int32

//go:wasm-module env
//export play_sound
func playSound(soundID int32)
```

**Note**: Full implementation of entity spawning, texture loading, and sound playback will be available in v6.1.

## Testing

### Unit Tests

Write standard Go tests:

```go
package main

import "testing"

func TestModLogic(t *testing.T) {
    result := processWeaponDamage(100)
    if result != 200 {
        t.Errorf("expected 200, got %d", result)
    }
}
```

Run with regular Go:
```bash
go test
```

### Integration Tests

Test WASM compilation:

```bash
tinygo build -o mod.wasm -target wasi mod.go
ls -lh mod.wasm
```

Test loading in game:
```bash
violence-server --enable-wasm-mods --mods-dir ./mods
```

## Limitations (TinyGo)

TinyGo is a Go subset. **Not supported**:

- ❌ Goroutines / channels
- ❌ `reflect` package
- ❌ `net/http`, `database/sql`
- ❌ `cgo`
- ❌ Some standard library functions

**Supported**:
- ✅ `fmt`, `strings`, `math`, `sort`
- ✅ Structs, interfaces, methods
- ✅ Slices, maps, arrays
- ✅ Error handling

See [TinyGo documentation](https://tinygo.org/docs/reference/lang-support/) for full compatibility matrix.

## Legacy Go Plugins (DEPRECATED)

**⚠️ WARNING**: Go plugins are unsafe for untrusted mods and will be removed in v7.0.

If you must use plugins for internal/trusted development:

```bash
violence-server --enable-unsafe-plugins
```

We **strongly recommend** migrating to WASM mods.

## Debugging

### Common Issues

**1. `module verification failed`**
- Ensure you compiled with TinyGo, not standard Go
- Check target is `wasi`: `tinygo build -target wasi`

**2. `permission denied: file read`**
- File access restricted to `mods/` directory
- Use relative paths: `mods/my_mod/config.txt`

**3. `fuel exhausted`**
- Function took too many instructions (likely infinite loop)
- Optimize your algorithm or split into smaller functions

### Logging

Use `logMessage()` liberally:

```go
logMessage("DEBUG: weapon_id=" + string(weaponID))
```

View logs in `violence.log` or console output.

## Example Mods

See `examples/mods/` directory for reference implementations:

- `weapon_pack` - Adds 3 new weapons
- `difficulty_modifier` - Doubles enemy health
- `hud_custom` - Custom health bar design

## Community Resources

- [Violence Modding Forum](https://forum.violence.example.com/mods)
- [TinyGo Documentation](https://tinygo.org/docs/)
- [WASM Specification](https://webassembly.org/)

## Troubleshooting

For mod-related issues:

1. Check `violence.log` for errors
2. Verify `mod.json` syntax is valid JSON
3. Ensure WASM file is less than 10MB
4. Test with minimal mod first (just `init()` function)

## Version History

- **v6.0** (2026-03-01): Initial WASM support with basic API stubs
- **v6.1** (planned): Full entity spawning, texture loading, sound API
- **v7.0** (planned): Remove unsafe plugin support, WASM Component Model

## License

Mods are subject to the game's modding license (see `LICENSE.md`).
