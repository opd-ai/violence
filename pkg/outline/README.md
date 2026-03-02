# Outline System

Per-pixel sprite outline rendering for improved visual clarity and entity distinction.

## Purpose

Addresses the **POOR READABILITY** visual problem by adding configurable sprite outlines that improve visual hierarchy during chaotic combat. Genre-specific color palettes ensure allies, enemies, and interactable objects are instantly recognizable.

## Features

- **Per-pixel outline generation** with distance-based alpha falloff
- **Genre-aware color palettes** (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)
- **Optional glow effect** with soft penumbra for magical/tech items
- **LRU caching** for performance (zero allocations on cache hits)
- **Memory-pooled buffers** to avoid GC pressure on hot paths

## Architecture

### Component

```go
type Component struct {
    Enabled   bool       // Whether outline rendering is active
    Color     color.RGBA // Override color (or use genre defaults)
    Thickness int        // Outline width in pixels (1-3 recommended)
    Glow      bool       // Enable soft glow falloff
}
```

### System

The `System` manages outline generation and caching:

- `NewSystem(genreID string)` — Create system with genre-specific defaults
- `SetGenre(genreID string)` — Update genre and clear cache
- `GenerateOutline(src, color, thickness, glow)` — Generate outlined sprite
- `Get{Player,Enemy,Ally,Neutral,Interact}Color()` — Genre-appropriate colors

## Usage

### Basic Integration

```go
outlineSystem := outline.NewSystem("fantasy")
world.AddSystem(outlineSystem)

// Add outline to entity
entity := world.AddEntity()
world.AddComponent(entity, &outline.Component{
    Enabled:   true,
    Color:     outlineSystem.GetEnemyColor(),
    Thickness: 2,
    Glow:      false,
})
```

### Manual Outline Generation

```go
outlined := outlineSystem.GenerateOutline(
    sprite, 
    color.RGBA{R: 255, G: 0, B: 0, A: 255}, 
    2,    // thickness
    false // glow
)
```

## Genre Color Schemes

| Genre      | Player   | Enemy    | Ally     | Neutral  | Interact |
|------------|----------|----------|----------|----------|----------|
| Fantasy    | Blue     | Red      | Green    | Gray     | Gold     |
| Sci-Fi     | Cyan     | Red      | Green    | Gray     | Yellow   |
| Horror     | Lt Blue  | Dk Red   | Lt Green | Gray     | Orange   |
| Cyberpunk  | Cyan     | Magenta  | Green    | Gray     | Yellow   |
| Post-Apoc  | Lt Blue  | Orange   | Green    | Tan      | Gold     |

## Performance

- **Cache hit**: O(1) lookup, zero allocations
- **Cache miss**: O(w×h×thickness²) per sprite
- **Cache limit**: 200 entries (auto-evicted on overflow)
- **Memory**: Pooled RGBA buffers prevent GC thrashing

Outline generation uses pooled image buffers and is only invoked once per unique (sprite, color, thickness, glow) tuple. Subsequent requests return the cached result.

## Integration Points

The system is registered in `main.go`:

1. System initialization: `outlineSystem: outline.NewSystem("fantasy")`
2. Genre propagation: `g.outlineSystem.SetGenre(g.genreID)` in `generateLevel()`
3. World registration: `g.world.AddSystem(g.outlineSystem)`

Rendering systems (sprite, equipment, particle) can call `GenerateOutline()` to add outlines to sprites before drawing.

## Testing

Run tests with:

```bash
go test ./pkg/outline/...
```

Note: Display-dependent tests require a graphics environment. Component and system initialization tests run in headless mode.

## Future Enhancements

- Integrate with animation system for per-frame outline updates
- Add outline width modulation for status effects (poison = pulsing green)
- Support custom outline shapes (star outline for quest objectives)
- GPU shader-based outline generation for higher performance
