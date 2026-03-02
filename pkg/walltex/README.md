# Enhanced Wall Texture Generation

The `walltex` package provides material-specific procedural wall texture generation with depth, weathering, and genre-appropriate visual effects.

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

## Usage

```go
// Create genre-specific generator
gen := walltex.NewGenerator("fantasy")

// Generate a 64x64 wall texture
// variant 0 uses primary material (stone for fantasy)
// variant 1 uses secondary material (wood for fantasy)
img := gen.Generate(64, 0, seed)

// Result is *image.RGBA ready for use in texture atlas
```

## Integration

The `walltex` package is integrated into the existing `texture.Atlas` through the `GenerateWallSet` method. When wall textures are generated, they automatically use the enhanced material-specific rendering with all visual improvements.

## Performance

- **Deterministic**: Same seed produces identical output
- **Efficient**: Single-pass generation with minimal allocations
- **Cacheable**: Output is static `*image.RGBA` suitable for LRU caching
- Typical 64x64 texture generation: <1ms

## Visual Improvements Over Previous System

The previous wall texture generator used simple Perlin noise with optional brick patterns. This enhancement adds:

1. **Material differentiation**: Metal looks metallic, wood has grain, stone has masonry
2. **Depth cues**: Every pixel has shading based on surface normal estimation
3. **Weathering**: Cracks, stains, and wear appropriate to setting
4. **Fine detail**: Scratches, spots, and imperfections prevent monotony
5. **Genre atmosphere**: Glowing panels in cyberpunk, organic growth in horror
6. **True variation**: Each variant uses a different material, not just color swaps
