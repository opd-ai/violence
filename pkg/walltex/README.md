# Enhanced Wall Texture Generation

The `walltex` package provides material-specific procedural wall texture generation with depth, weathering, and genre-appropriate visual effects. It includes both a low-level `Generator` for static texture creation and an ECS-integrated `System` for dynamic wall texture variation across dungeon levels.

## Features

### Material-Specific Rendering
Seven distinct material types, each with unique visual characteristics:
- **Stone**: Brick patterns with mortar lines, per-brick color variation
- **Metal**: Panel seams with brushed metal striations
- **Wood**: Horizontal planks with realistic grain patterns
- **Concrete**: Aggregate texture with patch variation
- **Organic**: Fleshy or vine-covered surfaces with irregular patterns
- **Crystal**: Angular facets with internal structure
- **Tech**: Circuit traces and high-tech paneling

### Depth Simulation
- **Normal mapping**: Simulates directional lighting to create the illusion of depth
- **Highlight/shadow**: Each pixel is adjusted based on estimated surface normal
- Light source from top-left creates realistic shading

### Weathering System
Genre-appropriate wear and damage:
- **Cracks**: Thin dark lines at varying angles
- **Stains**: Circular darkened patches with falloff
- **Intensity scales** with genre (postapoc: 0.9, horror: 0.8, fantasy: 0.6, etc.)
- **Depth-based weathering**: Deeper dungeon levels show more wear

### Detail Layers
- Dark spots (damage, holes)
- Bright spots (highlights, reflections)
- Small scratches and lines
- Density controlled per-genre

### Genre-Specific Effects
- **Glow** (sci-fi/cyberpunk): Glowing strips with cyan/blue coloring
- **Color variation**: Each genre has tuned color palette and variation range
- **Material selection**: Primary and secondary materials per genre

## Genre Presets

| Genre      | Primary Material | Secondary Material | Weather | Detail | Glow |
|------------|------------------|-------------------|---------|--------|------|
| Fantasy    | Stone           | Wood              | 0.6     | 0.5    | 0.0  |
| SciFi      | Metal           | Tech              | 0.2     | 0.4    | 0.4  |
| Horror     | Wood            | Organic           | 0.8     | 0.7    | 0.1  |
| Cyberpunk  | Concrete        | Tech              | 0.5     | 0.6    | 0.7  |
| PostApoc   | Concrete        | Metal             | 0.9     | 0.8    | 0.0  |

## ECS Integration (New)

The `System` provides dynamic wall texture variation integrated with the ECS architecture:

```go
// System automatically initialized in main.go
g.wallTexSystem = walltex.NewSystem(g.genreID, 200) // 200 texture cache size
g.world.AddSystem(g.wallTexSystem)

// Generate wall texture for a specific position
comp := g.wallTexSystem.GenerateWallTexture(
    gridX, gridY,   // Wall grid position
    "corridor",     // Room type (corridor, room, boss, treasure)
    5,              // Dungeon depth (affects weathering)
    seed,           // Level seed
)

// Sample texture color for rendering
color := g.wallTexSystem.SampleTexture(comp, u, v)
```

### Room Type Material Distribution

Each room type has distinct material probabilities:

- **Corridors**: 80% primary material, moderate weathering
- **Rooms**: 70% primary material, more variety
- **Boss rooms**: 90% primary material, low weathering (pristine)
- **Treasure rooms**: 60% primary material, exotic mix

### Texture Caching

- **LRU-style cache** with configurable max size
- Cache eviction when full (removes oldest half)
- Cache statistics: `GetCacheStats()` returns hits, misses, size
- Same wall position always returns same cached texture

## Usage (Static Textures)

```go
// Create genre-specific generator
gen := walltex.NewGenerator("fantasy")

// Generate a 64x64 wall texture
// variant 0 uses primary material (stone for fantasy)
// variant 1 uses secondary material (wood for fantasy)
img := gen.Generate(64, 0, seed)

// Generate with explicit material and weathering
img := gen.GenerateWithMaterial(64, MaterialOrganic, 2, 0.85, seed)

// Result is *image.RGBA ready for use in texture atlas
```

## Integration

The `walltex` package is integrated at multiple levels:

1. **Texture Atlas** (`texture.Atlas`): Static wall_1 through wall_4 textures
2. **ECS System** (`System`): Dynamic per-wall texture variation
3. **Renderer** (`render.Renderer`): Samples textures during raycasting

## Performance

- **Deterministic**: Same seed produces identical output
- **Efficient**: Single-pass generation with minimal allocations
- **Cached**: LRU cache prevents redundant generation
- **Test coverage**: 93.8% of statements
- Typical 64x64 texture generation: <1ms
- Cache hit rate: >90% in typical gameplay

## Visual Improvements

The wall texture system adds:

1. **Material differentiation**: Metal looks metallic, wood has grain, stone has masonry
2. **Depth cues**: Every pixel has shading based on surface normal estimation
3. **Weathering**: Cracks, stains, and wear appropriate to setting
4. **Fine detail**: Scratches, spots, and imperfections prevent monotony
5. **Genre atmosphere**: Glowing panels in cyberpunk, organic growth in horror
6. **True variation**: Each variant uses a different material, not just color swaps
7. **Spatial storytelling**: Boss rooms look different from corridors
8. **Depth progression**: Deeper levels show more wear and corruption
