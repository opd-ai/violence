# Weather System

Environmental particle system for atmospheric immersion in the Violence action-RPG.

## Features

- **Genre-specific weather effects** (rain, snow, embers, dust, ash, fog, neon glitches)
- **Parallax depth layers** for visual depth (background particles move slower)
- **Wind simulation** affecting particle movement
- **Performance-optimized** with pre-allocated particle pool and culling
- **Automatic genre adaptation** based on game setting

## Weather Types

| Type | Genre | Description |
|------|-------|-------------|
| Dust | Fantasy | Floating dust motes in dungeons |
| Rain | Cyberpunk | Heavy rain with neon reflections |
| Fog | Horror | Creeping fog wisps |
| Ash | Post-apocalyptic | Falling ash from nuclear fallout |
| Snow | N/A | Falling snowflakes (can be manually set) |
| Embers | N/A | Rising embers from fires |
| Neon Glitch | N/A | Digital artifacts in cyberpunk settings |

## Integration

The weather system is automatically initialized and registered with the ECS:

```go
weatherSystem := weather.NewSystem(2000, seed, genreID)
world.AddSystem(weatherSystem)
weatherSystem.AddWeatherToWorld(world)
```

Weather is rendered with depth-based parallax for visual depth cues.

## Performance

- Pre-allocated particle pool (2000 particles)
- Spatial culling outside camera view
- Depth-based LOD for rendering
- Target: 60+ FPS maintained
- Benchmark: ~1.2ms per frame for 2000 active particles

## Configuration

Weather intensity and type can be changed programmatically:

```go
weatherSys.GetWeatherSystem().SetWeather(weather.WeatherRain, 0.8)
weatherSys.GetWeatherSystem().SetWind(10.0, 50.0)
```
