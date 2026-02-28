# VIOLENCE

Raycasting first-person shooter built with Go and Ebitengine.

## Directory Structure

```
main.go                  Entry point — Ebitengine game loop
config.toml              Default configuration file
pkg/
  config/                Configuration loading (Viper)
  engine/                ECS framework (entities, components, systems)
  procgen/genre/         Genre registry and SetGenre interface
  rng/                   Seed-based deterministic RNG
  raycaster/             DDA raycasting engine
  bsp/                   BSP procedural level generator
  render/                Rendering pipeline (raycaster → framebuffer → screen)
  camera/                First-person camera (FOV, pitch, head-bob)
  input/                 Input manager (keyboard, mouse, gamepad)
  audio/                 Audio engine (procedurally generated music, SFX, positional audio)
  ui/                    HUD, menus, and settings screens
  tutorial/              Context-sensitive tutorial prompts
  weapon/                Weapon definitions, firing, and mastery progression
  upgrade/               Weapon upgrade token system
  ammo/                  Ammo types and pools
  door/                  Keycard and door system
  automap/               Fog-of-war automap
  ai/                    Enemy behavior trees
  combat/                Damage model and hit feedback
  status/                Status effects (poison, burn, bleed, radiation)
  loot/                  Loot tables and drops
  progression/           XP and leveling
  class/                 Character class definitions
  texture/               Procedural texture atlas
  lighting/              Sector-based dynamic lighting
  particle/              Particle emitters and effects
  event/                 World events and timed triggers
  inventory/             Item inventory
  crafting/              Scrap-to-ammo crafting
  quest/                 Procedurally generated level objectives and tracking
  shop/                  Between-level armory shop
  squad/                 Squad companion AI
  lore/                  Procedurally generated collectible lore and codex
  minigame/              Hacking and lockpicking mini-games
  secret/                Push-wall secret discovery
  destruct/              Destructible environments
  props/                 Decorative prop placement
  skills/                Skill and talent trees
  save/                  Save and load (cross-platform)
  network/               Client/server netcode
  federation/            Cross-server matchmaking
  chat/                  E2E encrypted in-game chat
  mod/                   Mod loader and plugin API
```

## Build and Run

Requires Go 1.24+ and a C compiler (for CGo, used by Ebitengine).

```sh
go build -o violence .
./violence
```

Or run directly:

```sh
go run .
```

## Configuration

Configuration is loaded from `config.toml` in the working directory or `$HOME/.violence/config.toml`.

Settings include window size, internal resolution, FOV, mouse sensitivity, audio volumes, default genre, VSync, and fullscreen mode. See `config.toml` for all options.

## Procedural Generation Policy

**100% of gameplay assets are procedurally generated at runtime using deterministic algorithms.** This includes all audio (music, SFX, ambient), all visuals (textures, sprites, particles, UI elements), and all narrative content (dialogue, lore, quests, world-building text, plot progression, character backstories). No pre-rendered, embedded, or bundled asset files (e.g., `.mp3`, `.wav`, `.ogg`, `.png`, `.jpg`, `.svg`, `.gif`) or static narrative content (e.g., hardcoded dialogue, pre-written cutscene scripts, fixed story arcs, embedded text assets) are permitted in the project. All procedural generation is deterministic: identical inputs (seeds) produce identical outputs across all platforms.

**Note on encoding formats:** While bundled audio/image asset files are prohibited, runtime use of standard encoding formats (e.g., WAV for in-memory PCM audio buffers, PNG for screenshot exports) is permitted when necessary for interfacing with system libraries or export features. The policy prohibits pre-authored assets, not the technical use of common container formats for runtime-generated data.

## Dependencies

- [Ebitengine v2](https://ebitengine.org/) — 2D game engine
- [Viper](https://github.com/spf13/viper) — configuration management
