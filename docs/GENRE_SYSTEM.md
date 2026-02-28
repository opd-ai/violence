# Genre System Guide

The genre system allows VIOLENCE to dynamically adapt visuals, audio, level generation, and narrative content to different thematic settings. All genre differences are cosmetic — gameplay mechanics remain consistent.

## Overview

Genre selection is a global setting that cascades through all procedural generation subsystems. Changing the genre changes:

- Wall and floor textures
- Post-processing visual effects
- Music composition parameters
- Sound effect characteristics
- Ambient audio atmosphere
- Enemy and item naming/descriptions
- Level architecture style
- Control point visual styles (multiplayer)
- Destructible environment debris types

## Built-in Genres

| Genre ID | Display Name | Setting |
| -------- | ------------ | ------- |
| `fantasy` | Fantasy | Stone dungeons, torches, medieval weapons |
| `scifi` | Sci-Fi | Metal hulls, tech terminals, energy weapons |
| `horror` | Horror | Cracked plaster, dim lighting, unsettling sounds |
| `cyberpunk` | Cyberpunk | Neon glass, server racks, high-tech gear |
| `postapoc` | Post-Apocalyptic | Rusted metal, debris, salvaged equipment |

Genre constants are defined in `pkg/procgen/genre`:

```go
import "github.com/opd-ai/violence/pkg/procgen/genre"

const (
    genre.Fantasy   = "fantasy"
    genre.SciFi     = "scifi"
    genre.Horror    = "horror"
    genre.Cyberpunk = "cyberpunk"
    genre.PostApoc  = "postapoc"
)
```

## Genre Registry (`pkg/procgen/genre`)

The `Registry` manages available genres:

```go
// Create a registry and register built-in genres
registry := genre.NewRegistry()
registry.Register(genre.Genre{ID: "fantasy", Name: "Fantasy"})
registry.Register(genre.Genre{ID: "scifi", Name: "Sci-Fi"})

// Look up a genre
g, ok := registry.Get("fantasy")
```

### Genre Type

```go
type Genre struct {
    ID   string // Unique identifier used in SetGenre() calls
    Name string // Human-readable display name
}
```

## The SetGenre Interface

There is no formal Go interface for `SetGenre` — it is a convention. Any package or type that adapts to genre implements a `SetGenre(genreID string)` method. The game's entry point (`main.go`) calls `setGenre()` to cascade the genre to all subsystems:

### Cascade Flow

```
Game.setGenre(genreID)
 ├── engine.World.SetGenre()         — ECS world genre tracking
 ├── engine.SetGenre()               — Global genre state
 ├── camera.SetGenre()               — Camera parameters
 ├── bsp.Generator.SetGenre()        — Level tile selection
 ├── render.Renderer.SetGenre()      — Visual pipeline
 │    └── PostProcessor.SetGenre()   — Post-processing effects
 │    └── TextureAtlas.SetGenre()    — Texture generation
 ├── audio.Engine.SetGenre()         — Music and SFX synthesis
 ├── combat.SetGenre()               — Damage model flavor
 ├── loot.SetGenre()                 — Item name/type generation
 ├── progression.SetGenre()          — XP curve labels
 ├── destruct.SetGenre()             — Debris types
 └── network.SetGenre()              — Network genre hint
```

## Implementing SetGenre in a Package

To make a package genre-aware, implement a `SetGenre` method that switches behavior based on the genre ID:

```go
package mypackage

import "github.com/opd-ai/violence/pkg/procgen/genre"

type MySystem struct {
    genreID string
    // ... other fields
}

func (s *MySystem) SetGenre(genreID string) {
    s.genreID = genreID

    switch genreID {
    case genre.Fantasy:
        // Configure for fantasy setting
    case genre.SciFi:
        // Configure for sci-fi setting
    case genre.Horror:
        // Configure for horror setting
    case genre.Cyberpunk:
        // Configure for cyberpunk setting
    case genre.PostApoc:
        // Configure for post-apocalyptic setting
    default:
        // Fallback to generic defaults
    }
}
```

## Genre Effects by Subsystem

### BSP Level Generation (`pkg/bsp`)

Each genre maps to specific wall and floor tile types:

| Genre | Wall Tile | Floor Tile |
| ----- | --------- | ---------- |
| Fantasy | `TileWallStone` (10) | `TileFloorStone` (20) |
| Sci-Fi | `TileWallHull` (11) | `TileFloorHull` (21) |
| Horror | `TileWallPlaster` (12) | `TileFloorWood` (22) |
| Cyberpunk | `TileWallConcrete` (13) | `TileFloorConcrete` (23) |
| Post-Apocalyptic | `TileWallRust` (14) | `TileFloorDirt` (24) |

### Rendering (`pkg/render`)

The post-processor applies genre-specific effect chains:

| Effect | Fantasy | Sci-Fi | Horror | Cyberpunk | Post-Apoc |
| ------ | ------- | ------ | ------ | --------- | --------- |
| Color Grade | Warm | Cool blue | Desaturated | Neon tint | Sepia |
| Vignette | Moderate | Light | Heavy | Moderate | Heavy |
| Film Grain | None | None | Heavy | Light | Moderate |
| Scanlines | None | Light | None | Heavy | None |
| Chromatic Aberration | None | None | Moderate | Heavy | None |
| Bloom | Light | Moderate | None | Heavy | None |
| Static Burst | None | None | Occasional | None | None |
| Film Scratches | None | None | Occasional | None | None |

### Audio (`pkg/audio`)

Genre determines:
- Base music composition style and instrument selection
- Ambient sound generation (dripping water vs. humming machinery vs. wind)
- Reverb characteristics (stone cathedral vs. metal corridor)
- SFX flavor (metallic weapons vs. energy weapons)

### Multiplayer Territory Control (`pkg/network`)

Control point visual styles per genre:

| Genre | Visual Style |
| ----- | ------------ |
| Fantasy | Altar |
| Sci-Fi | Terminal |
| Horror | Summoning Circle |
| Cyberpunk | Server Rack |
| Post-Apocalyptic | Scrap Pile |

## Adding a New Genre

To add a completely new genre:

1. **Register the genre ID** in `pkg/procgen/genre/genre.go`:
   ```go
   const MyGenre = "mygenre"
   ```

2. **Add a case to each SetGenre switch** in the packages listed in the cascade flow above.

3. **Define tile types** in `pkg/bsp/bsp.go` for wall and floor variants.

4. **Define post-processing parameters** in `pkg/render/postprocess.go`.

5. **Test with genre cascade test** in `test/genre_cascade_test.go`.

### Testing Genre Support

The `test/genre_cascade_test.go` file validates that all subsystems respond correctly to genre changes. When adding a new genre, add test cases for the new ID to verify it propagates through all systems.

## Configuration

The default genre is set in `config.toml`:

```toml
DefaultGenre = "fantasy"
```

This can be changed at any time — the genre cascade updates all systems immediately.
