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
  audio/                 Audio engine (music, SFX, positional audio)
  ui/                    HUD, menus, and settings screens
  tutorial/              Context-sensitive tutorial prompts
  weapon/                Weapon definitions and firing
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
  quest/                 Level objectives and tracking
  shop/                  Between-level armory shop
  squad/                 Squad companion AI
  lore/                  Collectible lore and codex
  minigame/              Hacking and lockpicking mini-games
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

## Dependencies

- [Ebitengine v2](https://ebitengine.org/) — 2D game engine
- [Viper](https://github.com/spf13/viper) — configuration management
