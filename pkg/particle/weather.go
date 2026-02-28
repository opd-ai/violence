// Package particle implements genre-specific indoor weather and atmospheric particle effects.
package particle

import (
	"image/color"
	"math"
)

// WeatherEmitter generates continuous atmospheric particles for indoor environments.
type WeatherEmitter struct {
	system       *ParticleSystem
	genreID      string
	x, y         float64 // Center position
	width        float64 // Effect area width
	height       float64 // Effect area height
	emitTimer    float64 // Time accumulator for emission
	emitInterval float64 // Seconds between particle spawns
}

// NewWeatherEmitter creates a weather emitter for a specific area.
func NewWeatherEmitter(system *ParticleSystem, genreID string, x, y, width, height float64) *WeatherEmitter {
	interval := 0.2 // Default: 5 particles per second
	switch genreID {
	case "fantasy":
		interval = 0.3 // Slower dripping
	case "scifi":
		interval = 0.15 // Faster steam bursts
	case "horror":
		interval = 0.5 // Slow, ominous
	case "cyberpunk":
		interval = 0.1 // Rapid glitch/static
	case "postapoc":
		interval = 0.25 // Dust settling
	}

	return &WeatherEmitter{
		system:       system,
		genreID:      genreID,
		x:            x,
		y:            y,
		width:        width,
		height:       height,
		emitInterval: interval,
	}
}

// Update advances the weather emitter by deltaTime and spawns particles.
func (w *WeatherEmitter) Update(deltaTime float64) {
	w.emitTimer += deltaTime

	for w.emitTimer >= w.emitInterval {
		w.emitTimer -= w.emitInterval
		w.emit()
	}
}

// emit spawns genre-specific weather particles.
func (w *WeatherEmitter) emit() {
	switch w.genreID {
	case "fantasy":
		w.emitDrippingWater()
	case "scifi":
		w.emitVentSteam()
	case "horror":
		w.emitFlickeringDust()
	case "cyberpunk":
		w.emitHolographicStatic()
	case "postapoc":
		w.emitFallingDust()
	}
}

// emitDrippingWater spawns water droplets for fantasy dungeons.
func (w *WeatherEmitter) emitDrippingWater() {
	// Random position along ceiling
	x := w.x + (w.system.rng.Float64()-0.5)*w.width
	y := w.y + (w.system.rng.Float64()-0.5)*w.height

	// Water drops fall straight down with slight randomness
	vx := (w.system.rng.Float64()*2 - 1) * 0.5
	vy := (w.system.rng.Float64()*2 - 1) * 0.5
	vz := -8.0 - w.system.rng.Float64()*2.0 // Fall speed

	c := color.RGBA{R: 100, G: 150, B: 200, A: 180}
	w.system.Spawn(x, y, 3.0, vx, vy, vz, 1.5, 0.3, c)
}

// emitVentSteam spawns steam particles for sci-fi facilities.
func (w *WeatherEmitter) emitVentSteam() {
	// Random vent position
	x := w.x + (w.system.rng.Float64()-0.5)*w.width
	y := w.y + (w.system.rng.Float64()-0.5)*w.height

	// Steam rises and spreads
	angle := w.system.rng.Float64() * 2 * math.Pi
	speed := 2.0 + w.system.rng.Float64()*3.0

	vx := math.Cos(angle) * speed * 0.5
	vy := math.Sin(angle) * speed * 0.5
	vz := 3.0 + w.system.rng.Float64()*2.0 // Rise speed

	c := color.RGBA{R: 200, G: 210, B: 220, A: 120}
	w.system.Spawn(x, y, 0.5, vx, vy, vz, 2.5, 1.5, c)
}

// emitFlickeringDust spawns ominous particles for horror settings.
func (w *WeatherEmitter) emitFlickeringDust() {
	// Random position
	x := w.x + (w.system.rng.Float64()-0.5)*w.width
	y := w.y + (w.system.rng.Float64()-0.5)*w.height

	// Dust drifts slowly
	vx := (w.system.rng.Float64()*2 - 1) * 0.8
	vy := (w.system.rng.Float64()*2 - 1) * 0.8
	vz := -0.5 - w.system.rng.Float64()*0.5 // Slow fall

	c := color.RGBA{R: 80, G: 80, B: 70, A: 100}
	w.system.Spawn(x, y, 2.0, vx, vy, vz, 3.0, 0.6, c)
}

// emitHolographicStatic spawns glitchy particles for cyberpunk environments.
func (w *WeatherEmitter) emitHolographicStatic() {
	// Random position
	x := w.x + (w.system.rng.Float64()-0.5)*w.width
	y := w.y + (w.system.rng.Float64()-0.5)*w.height

	// Static particles have erratic movement
	vx := (w.system.rng.Float64()*2 - 1) * 5.0
	vy := (w.system.rng.Float64()*2 - 1) * 5.0
	vz := (w.system.rng.Float64()*2 - 1) * 3.0

	// Alternate between magenta and cyan
	var c color.RGBA
	if w.system.rng.Float64() > 0.5 {
		c = color.RGBA{R: 255, G: 0, B: 255, A: 150}
	} else {
		c = color.RGBA{R: 0, G: 255, B: 255, A: 150}
	}

	w.system.Spawn(x, y, 1.0, vx, vy, vz, 0.3, 0.5, c)
}

// emitFallingDust spawns dust particles for post-apocalyptic ruins.
func (w *WeatherEmitter) emitFallingDust() {
	// Random position
	x := w.x + (w.system.rng.Float64()-0.5)*w.width
	y := w.y + (w.system.rng.Float64()-0.5)*w.height

	// Dust falls slowly and drifts
	vx := (w.system.rng.Float64()*2 - 1) * 1.0
	vy := (w.system.rng.Float64()*2 - 1) * 1.0
	vz := -1.5 - w.system.rng.Float64()*1.0

	c := color.RGBA{R: 120, G: 100, B: 70, A: 130}
	w.system.Spawn(x, y, 2.5, vx, vy, vz, 4.0, 0.8, c)
}

// FlickeringLightController manages light-flickering particle effects for horror ambience.
type FlickeringLightController struct {
	system      *ParticleSystem
	x, y        float64
	flickerRate float64 // Flickers per second
	nextFlicker float64 // Time until next flicker
}

// NewFlickeringLightController creates a light flicker controller.
func NewFlickeringLightController(system *ParticleSystem, x, y, flickerRate float64) *FlickeringLightController {
	return &FlickeringLightController{
		system:      system,
		x:           x,
		y:           y,
		flickerRate: flickerRate,
		nextFlicker: 1.0 / flickerRate,
	}
}

// Update advances the flicker controller and spawns flicker particles.
func (f *FlickeringLightController) Update(deltaTime float64) {
	f.nextFlicker -= deltaTime

	if f.nextFlicker <= 0 {
		// Reset timer with slight randomness
		f.nextFlicker = (1.0 / f.flickerRate) * (0.8 + f.system.rng.Float64()*0.4)

		// Spawn brief flash particle at light position
		c := color.RGBA{R: 255, G: 240, B: 200, A: 200}
		f.system.Spawn(f.x, f.y, 0, 0, 0, 0, 0.08, 2.0, c)
	}
}

// DustParticleEmitter creates ambient dust for multiple genres.
type DustParticleEmitter struct {
	system    *ParticleSystem
	genreID   string
	bounds    [4]float64 // minX, maxX, minY, maxY
	density   float64    // Particles per second per unit area
	timer     float64
	spawnRate float64
}

// NewDustParticleEmitter creates a dust emitter for a bounded area.
func NewDustParticleEmitter(system *ParticleSystem, genreID string, minX, maxX, minY, maxY, density float64) *DustParticleEmitter {
	area := (maxX - minX) * (maxY - minY)
	spawnRate := 1.0 / (density * area)
	if spawnRate < 0.05 {
		spawnRate = 0.05 // Cap at 20 particles/second
	}

	return &DustParticleEmitter{
		system:    system,
		genreID:   genreID,
		bounds:    [4]float64{minX, maxX, minY, maxY},
		density:   density,
		spawnRate: spawnRate,
	}
}

// Update advances the emitter and spawns dust particles.
func (d *DustParticleEmitter) Update(deltaTime float64) {
	d.timer += deltaTime

	for d.timer >= d.spawnRate {
		d.timer -= d.spawnRate
		d.spawnDust()
	}
}

// spawnDust creates a single dust particle.
func (d *DustParticleEmitter) spawnDust() {
	x := d.bounds[0] + d.system.rng.Float64()*(d.bounds[1]-d.bounds[0])
	y := d.bounds[2] + d.system.rng.Float64()*(d.bounds[3]-d.bounds[2])

	vx := (d.system.rng.Float64()*2 - 1) * 0.5
	vy := (d.system.rng.Float64()*2 - 1) * 0.5
	vz := -0.8 - d.system.rng.Float64()*0.4

	var c color.RGBA
	switch d.genreID {
	case "fantasy":
		c = color.RGBA{R: 100, G: 90, B: 70, A: 80}
	case "scifi":
		c = color.RGBA{R: 150, G: 150, B: 160, A: 70}
	case "horror":
		c = color.RGBA{R: 70, G: 65, B: 60, A: 90}
	case "cyberpunk":
		c = color.RGBA{R: 140, G: 120, B: 150, A: 75}
	case "postapoc":
		c = color.RGBA{R: 110, G: 90, B: 60, A: 100}
	default:
		c = color.RGBA{R: 120, G: 120, B: 120, A: 80}
	}

	d.system.Spawn(x, y, 2.0, vx, vy, vz, 3.5, 0.5, c)
}

// HolographicStaticEmitter creates glitchy hologram particles (cyberpunk).
type HolographicStaticEmitter struct {
	system    *ParticleSystem
	x, y      float64
	radius    float64
	burstMin  float64 // Min time between bursts
	burstMax  float64 // Max time between bursts
	nextBurst float64
}

// NewHolographicStaticEmitter creates a holographic static emitter.
func NewHolographicStaticEmitter(system *ParticleSystem, x, y, radius, burstMin, burstMax float64) *HolographicStaticEmitter {
	return &HolographicStaticEmitter{
		system:    system,
		x:         x,
		y:         y,
		radius:    radius,
		burstMin:  burstMin,
		burstMax:  burstMax,
		nextBurst: burstMin + system.rng.Float64()*(burstMax-burstMin),
	}
}

// Update advances the emitter and spawns static bursts.
func (h *HolographicStaticEmitter) Update(deltaTime float64) {
	h.nextBurst -= deltaTime

	if h.nextBurst <= 0 {
		// Reset timer
		h.nextBurst = h.burstMin + h.system.rng.Float64()*(h.burstMax-h.burstMin)

		// Spawn burst of static
		count := 5 + h.system.rng.Intn(10)
		for i := 0; i < count; i++ {
			angle := h.system.rng.Float64() * 2 * math.Pi
			dist := h.system.rng.Float64() * h.radius

			x := h.x + math.Cos(angle)*dist
			y := h.y + math.Sin(angle)*dist

			vx := (h.system.rng.Float64()*2 - 1) * 8.0
			vy := (h.system.rng.Float64()*2 - 1) * 8.0
			vz := (h.system.rng.Float64()*2 - 1) * 5.0

			// Magenta or cyan
			var c color.RGBA
			if h.system.rng.Float64() > 0.5 {
				c = color.RGBA{R: 255, G: 0, B: 255, A: 200}
			} else {
				c = color.RGBA{R: 0, G: 255, B: 255, A: 200}
			}

			h.system.Spawn(x, y, 0.5, vx, vy, vz, 0.2, 0.4, c)
		}
	}
}

// VentSteamEmitter creates steam emission from vents (sci-fi).
type VentSteamEmitter struct {
	system   *ParticleSystem
	x, y     float64
	dirX     float64 // Direction X component
	dirY     float64 // Direction Y component
	interval float64 // Seconds between emissions
	timer    float64
}

// NewVentSteamEmitter creates a vent steam emitter.
func NewVentSteamEmitter(system *ParticleSystem, x, y, dirX, dirY, interval float64) *VentSteamEmitter {
	return &VentSteamEmitter{
		system:   system,
		x:        x,
		y:        y,
		dirX:     dirX,
		dirY:     dirY,
		interval: interval,
	}
}

// Update advances the emitter and spawns steam.
func (v *VentSteamEmitter) Update(deltaTime float64) {
	v.timer += deltaTime

	for v.timer >= v.interval {
		v.timer -= v.interval

		// Spawn steam puff
		count := 3 + v.system.rng.Intn(5)
		for i := 0; i < count; i++ {
			speed := 3.0 + v.system.rng.Float64()*2.0
			spread := 0.3

			angle := math.Atan2(v.dirY, v.dirX) + (v.system.rng.Float64()*2-1)*spread
			vx := math.Cos(angle) * speed
			vy := math.Sin(angle) * speed
			vz := 1.0 + v.system.rng.Float64()*2.0

			c := color.RGBA{R: 220, G: 225, B: 230, A: 140}
			v.system.Spawn(v.x, v.y, 0, vx, vy, vz, 2.0, 1.2, c)
		}
	}
}

// DrippingWaterEmitter creates periodic water drips (fantasy).
type DrippingWaterEmitter struct {
	system   *ParticleSystem
	x, y     float64
	interval float64 // Seconds between drips
	timer    float64
}

// NewDrippingWaterEmitter creates a dripping water emitter.
func NewDrippingWaterEmitter(system *ParticleSystem, x, y, interval float64) *DrippingWaterEmitter {
	return &DrippingWaterEmitter{
		system:   system,
		x:        x,
		y:        y,
		interval: interval,
	}
}

// Update advances the emitter and spawns water drops.
func (d *DrippingWaterEmitter) Update(deltaTime float64) {
	d.timer += deltaTime

	for d.timer >= d.interval {
		d.timer -= d.interval

		// Spawn single water droplet
		vx := (d.system.rng.Float64()*2 - 1) * 0.3
		vy := (d.system.rng.Float64()*2 - 1) * 0.3
		vz := -10.0 - d.system.rng.Float64()*2.0

		c := color.RGBA{R: 100, G: 160, B: 220, A: 200}
		d.system.Spawn(d.x, d.y, 3.0, vx, vy, vz, 1.0, 0.25, c)
	}
}
