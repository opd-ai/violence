// Package weather provides environmental particle effects for atmospheric immersion.
package weather

import (
	"image/color"
	"math"
	"math/rand"
)

// WeatherType represents different environmental particle effects.
type WeatherType int

const (
	WeatherNone       WeatherType = iota // WeatherNone is no weather effect.
	WeatherRain                          // WeatherRain is rain weather.
	WeatherSnow                          // WeatherSnow is snow weather.
	WeatherEmbers                        // WeatherEmbers is ember weather.
	WeatherDust                          // WeatherDust is dust weather.
	WeatherAsh                           // WeatherAsh is ash weather.
	WeatherFog                           // WeatherFog is fog weather.
	WeatherNeonGlitch                    // WeatherNeonGlitch is neon glitch weather.
)

// EnvironmentalParticle represents a single atmospheric particle.
type EnvironmentalParticle struct {
	X, Y          float64 // World position
	VX, VY        float64 // Velocity
	DepthLayer    float64 // Parallax depth (0.0 = background, 1.0 = foreground)
	Size          float64 // Visual size
	Alpha         uint8   // Transparency
	Color         color.RGBA
	Lifetime      float64 // Seconds remaining
	MaxLifetime   float64 // Total lifetime
	Active        bool
	RotationAngle float64 // For snowflakes
	RotationSpeed float64
	FlickerPhase  float64 // For embers/glitch
	FlickerSpeed  float64
}

// WeatherSystem manages environmental particles for atmospheric effects.
type WeatherSystem struct {
	particles             []EnvironmentalParticle
	weatherType           WeatherType
	intensity             float64 // 0.0-1.0
	windX, windY          float64
	genre                 string
	rng                   *rand.Rand
	spawnAccumulator      float64
	cameraX, cameraY      float64
	viewWidth, viewHeight float64
}

// NewWeatherSystem creates an environmental particle system.
func NewWeatherSystem(maxParticles int, seed int64) *WeatherSystem {
	return &WeatherSystem{
		particles:   make([]EnvironmentalParticle, maxParticles),
		weatherType: WeatherNone,
		intensity:   0.5,
		windX:       0.0,
		windY:       0.0,
		genre:       "fantasy",
		rng:         rand.New(rand.NewSource(seed)),
		viewWidth:   800,
		viewHeight:  600,
	}
}

// SetGenre configures weather for a specific genre.
func (w *WeatherSystem) SetGenre(genreID string) {
	w.genre = genreID
	w.applyGenreWeather(genreID)
}

// applyGenreWeather sets appropriate weather for each genre.
func (w *WeatherSystem) applyGenreWeather(genreID string) {
	switch genreID {
	case "fantasy":
		// Dungeons have dust and occasional embers from torches
		w.SetWeather(WeatherDust, 0.3)
		w.windX = w.rng.Float64()*10 - 5
		w.windY = 5.0
	case "scifi":
		// Space stations have minimal particles (clean environment)
		w.SetWeather(WeatherNone, 0.0)
	case "horror":
		// Fog and ash for creepy atmosphere
		w.SetWeather(WeatherFog, 0.6)
		w.windX = w.rng.Float64()*5 - 2.5
		w.windY = 2.0
	case "cyberpunk":
		// Rain with neon glitches
		w.SetWeather(WeatherRain, 0.7)
		w.windX = -20.0
		w.windY = 100.0
	case "postapoc":
		// Ash and dust from nuclear fallout
		w.SetWeather(WeatherAsh, 0.5)
		w.windX = w.rng.Float64()*15 - 7.5
		w.windY = 15.0
	default:
		w.SetWeather(WeatherNone, 0.0)
	}
}

// SetWeather changes the weather type and intensity.
func (w *WeatherSystem) SetWeather(weatherType WeatherType, intensity float64) {
	w.weatherType = weatherType
	w.intensity = clamp(intensity, 0.0, 1.0)
}

// SetWind updates wind direction and strength.
func (w *WeatherSystem) SetWind(x, y float64) {
	w.windX = x
	w.windY = y
}

// SetCamera updates the camera position for particle spawning bounds.
func (w *WeatherSystem) SetCamera(x, y, width, height float64) {
	w.cameraX = x
	w.cameraY = y
	w.viewWidth = width
	w.viewHeight = height
}

// Update advances all environmental particles.
func (w *WeatherSystem) Update(deltaTime float64) {
	if w.weatherType == WeatherNone {
		return
	}

	// Spawn new particles based on intensity
	w.spawnParticles(deltaTime)

	// Update existing particles
	for i := range w.particles {
		p := &w.particles[i]
		if !p.Active {
			continue
		}

		// Update lifetime
		p.Lifetime -= deltaTime
		if p.Lifetime <= 0 {
			p.Active = false
			continue
		}

		// Apply velocity and wind
		effectiveVX := p.VX + w.windX*deltaTime*(1.0-p.DepthLayer*0.5)
		effectiveVY := p.VY + w.windY*deltaTime
		p.X += effectiveVX * deltaTime
		p.Y += effectiveVY * deltaTime

		// Update rotation (snowflakes, embers)
		p.RotationAngle += p.RotationSpeed * deltaTime
		if p.RotationAngle > math.Pi*2 {
			p.RotationAngle -= math.Pi * 2
		}

		// Update flicker phase (embers, neon glitches)
		p.FlickerPhase += p.FlickerSpeed * deltaTime
		if p.FlickerPhase > math.Pi*2 {
			p.FlickerPhase -= math.Pi * 2
		}

		// Fade out near end of lifetime
		lifetimeRatio := p.Lifetime / p.MaxLifetime
		if lifetimeRatio < 0.2 {
			p.Alpha = uint8(float64(p.Alpha) * (lifetimeRatio / 0.2))
		}

		// Cull particles far outside view
		cullMargin := 200.0
		if p.X < w.cameraX-cullMargin || p.X > w.cameraX+w.viewWidth+cullMargin ||
			p.Y < w.cameraY-cullMargin || p.Y > w.cameraY+w.viewHeight+cullMargin {
			p.Active = false
		}
	}
}

// spawnParticles creates new particles based on weather type and intensity.
func (w *WeatherSystem) spawnParticles(deltaTime float64) {
	spawnRate := w.getSpawnRate()
	w.spawnAccumulator += spawnRate * deltaTime

	for w.spawnAccumulator >= 1.0 {
		w.spawnAccumulator -= 1.0
		w.spawnSingleParticle()
	}
}

// getSpawnRate returns particles per second for current weather.
func (w *WeatherSystem) getSpawnRate() float64 {
	baseRate := map[WeatherType]float64{
		WeatherRain:       200.0,
		WeatherSnow:       100.0,
		WeatherEmbers:     30.0,
		WeatherDust:       50.0,
		WeatherAsh:        80.0,
		WeatherFog:        40.0,
		WeatherNeonGlitch: 20.0,
	}

	rate, ok := baseRate[w.weatherType]
	if !ok {
		return 0.0
	}

	return rate * w.intensity
}

// spawnSingleParticle creates one particle matching current weather type.
func (w *WeatherSystem) spawnSingleParticle() {
	// Find inactive particle slot
	var p *EnvironmentalParticle
	for i := range w.particles {
		if !w.particles[i].Active {
			p = &w.particles[i]
			break
		}
	}
	if p == nil {
		return // Pool exhausted
	}

	// Spawn above and to sides of camera
	spawnMargin := 100.0
	p.X = w.cameraX + w.rng.Float64()*(w.viewWidth+spawnMargin*2) - spawnMargin
	p.Y = w.cameraY - spawnMargin - w.rng.Float64()*50
	p.DepthLayer = w.rng.Float64()
	p.Active = true

	switch w.weatherType {
	case WeatherRain:
		w.initRainParticle(p)
	case WeatherSnow:
		w.initSnowParticle(p)
	case WeatherEmbers:
		w.initEmberParticle(p)
	case WeatherDust:
		w.initDustParticle(p)
	case WeatherAsh:
		w.initAshParticle(p)
	case WeatherFog:
		w.initFogParticle(p)
	case WeatherNeonGlitch:
		w.initNeonGlitchParticle(p)
	}
}

// initRainParticle configures a raindrop particle.
func (w *WeatherSystem) initRainParticle(p *EnvironmentalParticle) {
	p.VX = w.windX * 0.5
	p.VY = 150.0 + w.rng.Float64()*50
	p.Size = 1.0 + w.rng.Float64()*0.5
	p.Alpha = uint8(100 + w.rng.Intn(100))
	p.Color = color.RGBA{R: 150, G: 180, B: 220, A: p.Alpha}
	p.Lifetime = 5.0
	p.MaxLifetime = 5.0
	p.RotationAngle = 0
	p.RotationSpeed = 0
	p.FlickerPhase = 0
	p.FlickerSpeed = 0
}

// initSnowParticle configures a snowflake particle.
func (w *WeatherSystem) initSnowParticle(p *EnvironmentalParticle) {
	p.VX = w.rng.Float64()*10 - 5
	p.VY = 20.0 + w.rng.Float64()*20
	p.Size = 2.0 + w.rng.Float64()*2.0
	p.Alpha = uint8(180 + w.rng.Intn(75))
	p.Color = color.RGBA{R: 240, G: 245, B: 255, A: p.Alpha}
	p.Lifetime = 8.0
	p.MaxLifetime = 8.0
	p.RotationAngle = w.rng.Float64() * math.Pi * 2
	p.RotationSpeed = (w.rng.Float64() - 0.5) * 2.0
	p.FlickerPhase = 0
	p.FlickerSpeed = 0
}

// initEmberParticle configures a glowing ember particle.
func (w *WeatherSystem) initEmberParticle(p *EnvironmentalParticle) {
	p.VX = (w.rng.Float64() - 0.5) * 20
	p.VY = -30.0 - w.rng.Float64()*20 // Rise upward
	p.Size = 1.5 + w.rng.Float64()*1.5
	p.Alpha = uint8(200 + w.rng.Intn(55))
	hue := w.rng.Float64()
	if hue < 0.7 {
		p.Color = color.RGBA{R: 255, G: uint8(100 + w.rng.Intn(100)), B: 0, A: p.Alpha}
	} else {
		p.Color = color.RGBA{R: 255, G: 220, B: uint8(50 + w.rng.Intn(100)), A: p.Alpha}
	}
	p.Lifetime = 3.0 + w.rng.Float64()*2.0
	p.MaxLifetime = p.Lifetime
	p.RotationAngle = 0
	p.RotationSpeed = 0
	p.FlickerPhase = w.rng.Float64() * math.Pi * 2
	p.FlickerSpeed = 5.0 + w.rng.Float64()*5.0
}

// initDustParticle configures a dust mote particle.
func (w *WeatherSystem) initDustParticle(p *EnvironmentalParticle) {
	p.VX = (w.rng.Float64() - 0.5) * 15
	p.VY = 10.0 + w.rng.Float64()*10
	p.Size = 1.0 + w.rng.Float64()*1.0
	p.Alpha = uint8(40 + w.rng.Intn(80))
	gray := uint8(120 + w.rng.Intn(80))
	p.Color = color.RGBA{R: gray, G: gray, B: gray, A: p.Alpha}
	p.Lifetime = 6.0 + w.rng.Float64()*4.0
	p.MaxLifetime = p.Lifetime
	p.RotationAngle = 0
	p.RotationSpeed = 0
	p.FlickerPhase = 0
	p.FlickerSpeed = 0
}

// initAshParticle configures an ash particle.
func (w *WeatherSystem) initAshParticle(p *EnvironmentalParticle) {
	p.VX = (w.rng.Float64() - 0.5) * 20
	p.VY = 25.0 + w.rng.Float64()*15
	p.Size = 1.5 + w.rng.Float64()*1.5
	p.Alpha = uint8(80 + w.rng.Intn(100))
	gray := uint8(40 + w.rng.Intn(60))
	p.Color = color.RGBA{R: gray, G: gray - 10, B: gray - 15, A: p.Alpha}
	p.Lifetime = 7.0 + w.rng.Float64()*5.0
	p.MaxLifetime = p.Lifetime
	p.RotationAngle = w.rng.Float64() * math.Pi * 2
	p.RotationSpeed = (w.rng.Float64() - 0.5) * 1.0
	p.FlickerPhase = 0
	p.FlickerSpeed = 0
}

// initFogParticle configures a fog wisp particle.
func (w *WeatherSystem) initFogParticle(p *EnvironmentalParticle) {
	p.VX = (w.rng.Float64() - 0.5) * 8
	p.VY = 5.0 + w.rng.Float64()*5
	p.Size = 8.0 + w.rng.Float64()*12.0
	p.Alpha = uint8(20 + w.rng.Intn(40))
	p.Color = color.RGBA{R: 180, G: 190, B: 180, A: p.Alpha}
	p.Lifetime = 10.0 + w.rng.Float64()*8.0
	p.MaxLifetime = p.Lifetime
	p.RotationAngle = 0
	p.RotationSpeed = 0
	p.FlickerPhase = 0
	p.FlickerSpeed = 0
}

// initNeonGlitchParticle configures a cyberpunk glitch particle.
func (w *WeatherSystem) initNeonGlitchParticle(p *EnvironmentalParticle) {
	p.VX = (w.rng.Float64() - 0.5) * 40
	p.VY = 80.0 + w.rng.Float64()*40
	p.Size = 2.0 + w.rng.Float64()*3.0
	p.Alpha = uint8(150 + w.rng.Intn(105))
	colors := []color.RGBA{
		{R: 255, G: 0, B: 255, A: p.Alpha}, // Magenta
		{R: 0, G: 255, B: 255, A: p.Alpha}, // Cyan
		{R: 255, G: 255, B: 0, A: p.Alpha}, // Yellow
		{R: 0, G: 255, B: 100, A: p.Alpha}, // Neon green
	}
	p.Color = colors[w.rng.Intn(len(colors))]
	p.Lifetime = 2.0 + w.rng.Float64()*2.0
	p.MaxLifetime = p.Lifetime
	p.RotationAngle = 0
	p.RotationSpeed = 0
	p.FlickerPhase = w.rng.Float64() * math.Pi * 2
	p.FlickerSpeed = 15.0 + w.rng.Float64()*10.0
}

// GetActiveParticles returns all active environmental particles.
func (w *WeatherSystem) GetActiveParticles() []EnvironmentalParticle {
	active := make([]EnvironmentalParticle, 0, len(w.particles))
	for i := range w.particles {
		if w.particles[i].Active {
			active = append(active, w.particles[i])
		}
	}
	return active
}

// Clear removes all active particles.
func (w *WeatherSystem) Clear() {
	for i := range w.particles {
		w.particles[i].Active = false
	}
	w.spawnAccumulator = 0
}

// GetWeatherType returns the current weather type.
func (w *WeatherSystem) GetWeatherType() WeatherType {
	return w.weatherType
}

// GetIntensity returns the current weather intensity (0.0-1.0).
func (w *WeatherSystem) GetIntensity() float64 {
	return w.intensity
}

// clamp restricts a value to [min, max].
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
