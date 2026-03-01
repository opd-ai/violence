// Package feedback provides visual and kinesthetic feedback for player actions.
// Includes screen shake, hit flash, damage numbers, and enhanced combat juice.
package feedback

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/violence/pkg/engine"
)

// FeedbackSystem manages visual feedback for combat and player actions.
type FeedbackSystem struct {
	screenShake   *ScreenShake
	hitFlash      *HitFlash
	damageNumbers []*DamageNumber
	impactEffects []*ImpactEffect
	rng           *rand.Rand
	genre         string
	maxDamageNums int
	maxImpacts    int
}

// ScreenShake manages camera shake effects.
type ScreenShake struct {
	intensity float64
	decay     float64
	offsetX   float64
	offsetY   float64
	angle     float64
}

// HitFlash manages screen flash effects for damage feedback.
type HitFlash struct {
	intensity float64
	decay     float64
	color     color.RGBA
}

// DamageNumber is a floating damage indicator.
type DamageNumber struct {
	x        float64
	y        float64
	vx       float64
	vy       float64
	damage   int
	lifetime float64
	maxLife  float64
	critical bool
	color    color.RGBA
}

// ImpactEffect represents a brief visual effect at hit location.
type ImpactEffect struct {
	x        float64
	y        float64
	scale    float64
	rotation float64
	lifetime float64
	maxLife  float64
	itype    ImpactType
}

// ImpactType defines the visual style of an impact effect.
type ImpactType int

const (
	ImpactHit ImpactType = iota
	ImpactCritical
	ImpactBlock
	ImpactMiss
)

// NewFeedbackSystem creates a new feedback system.
func NewFeedbackSystem(seed int64) *FeedbackSystem {
	return &FeedbackSystem{
		screenShake: &ScreenShake{
			decay: 0.9,
		},
		hitFlash: &HitFlash{
			decay: 0.85,
			color: color.RGBA{R: 255, G: 0, B: 0, A: 128},
		},
		damageNumbers: make([]*DamageNumber, 0, 100),
		impactEffects: make([]*ImpactEffect, 0, 50),
		rng:           rand.New(rand.NewSource(seed)),
		genre:         "fantasy",
		maxDamageNums: 100,
		maxImpacts:    50,
	}
}

// SetGenre configures genre-specific feedback parameters.
func (f *FeedbackSystem) SetGenre(genreID string) {
	f.genre = genreID

	switch genreID {
	case "cyberpunk":
		f.hitFlash.color = color.RGBA{R: 0, G: 255, B: 255, A: 128}
	case "horror":
		f.hitFlash.color = color.RGBA{R: 180, G: 0, B: 0, A: 160}
	case "scifi":
		f.hitFlash.color = color.RGBA{R: 100, G: 200, B: 255, A: 128}
	default: // fantasy
		f.hitFlash.color = color.RGBA{R: 255, G: 0, B: 0, A: 128}
	}
}

// Update implements the System interface.
func (f *FeedbackSystem) Update(w *engine.World) {
	deltaTime := 1.0 / 60.0
	f.updateScreenShake(deltaTime)
	f.updateHitFlash(deltaTime)
	f.updateDamageNumbers(deltaTime)
	f.updateImpactEffects(deltaTime)
}

// updateScreenShake decays screen shake over time.
func (f *FeedbackSystem) updateScreenShake(deltaTime float64) {
	if f.screenShake.intensity > 0.01 {
		f.screenShake.intensity *= f.screenShake.decay
		f.screenShake.angle += 0.3

		f.screenShake.offsetX = math.Sin(f.screenShake.angle) * f.screenShake.intensity
		f.screenShake.offsetY = math.Cos(f.screenShake.angle*1.3) * f.screenShake.intensity
	} else {
		f.screenShake.intensity = 0
		f.screenShake.offsetX = 0
		f.screenShake.offsetY = 0
	}
}

// updateHitFlash decays hit flash intensity.
func (f *FeedbackSystem) updateHitFlash(deltaTime float64) {
	if f.hitFlash.intensity > 0.01 {
		f.hitFlash.intensity *= f.hitFlash.decay
	} else {
		f.hitFlash.intensity = 0
	}
}

// updateDamageNumbers updates all active damage numbers.
func (f *FeedbackSystem) updateDamageNumbers(deltaTime float64) {
	active := make([]*DamageNumber, 0, len(f.damageNumbers))

	for _, dn := range f.damageNumbers {
		dn.lifetime += deltaTime
		if dn.lifetime >= dn.maxLife {
			continue
		}

		dn.x += dn.vx * deltaTime
		dn.y += dn.vy * deltaTime
		dn.vy -= 3.0 * deltaTime // Gravity/slowdown

		active = append(active, dn)
	}

	f.damageNumbers = active
}

// updateImpactEffects updates all active impact effects.
func (f *FeedbackSystem) updateImpactEffects(deltaTime float64) {
	active := make([]*ImpactEffect, 0, len(f.impactEffects))

	for _, ie := range f.impactEffects {
		ie.lifetime += deltaTime
		if ie.lifetime >= ie.maxLife {
			continue
		}

		ie.rotation += 5.0 * deltaTime
		ie.scale = 1.0 + (ie.lifetime/ie.maxLife)*0.5

		active = append(active, ie)
	}

	f.impactEffects = active
}

// AddScreenShake adds camera shake with the specified intensity.
func (f *FeedbackSystem) AddScreenShake(intensity float64) {
	f.screenShake.intensity += intensity
	if f.screenShake.intensity > 20.0 {
		f.screenShake.intensity = 20.0
	}
}

// AddHitFlash adds a screen flash effect.
func (f *FeedbackSystem) AddHitFlash(intensity float64) {
	f.hitFlash.intensity += intensity
	if f.hitFlash.intensity > 1.0 {
		f.hitFlash.intensity = 1.0
	}
}

// SpawnDamageNumber creates a floating damage number at the specified position.
func (f *FeedbackSystem) SpawnDamageNumber(x, y float64, damage int, critical bool) {
	if len(f.damageNumbers) >= f.maxDamageNums {
		return
	}

	dnColor := f.getDamageNumberColor(critical)

	dn := &DamageNumber{
		x:        x,
		y:        y,
		vx:       (f.rng.Float64() - 0.5) * 2.0,
		vy:       2.0 + f.rng.Float64()*1.0,
		damage:   damage,
		lifetime: 0,
		maxLife:  1.5,
		critical: critical,
		color:    dnColor,
	}

	f.damageNumbers = append(f.damageNumbers, dn)
}

// getDamageNumberColor returns the color for a damage number based on genre and critical status.
func (f *FeedbackSystem) getDamageNumberColor(critical bool) color.RGBA {
	if critical {
		switch f.genre {
		case "cyberpunk":
			return color.RGBA{R: 0, G: 255, B: 255, A: 255}
		case "horror":
			return color.RGBA{R: 255, G: 50, B: 50, A: 255}
		case "scifi":
			return color.RGBA{R: 100, G: 200, B: 255, A: 255}
		default:
			return color.RGBA{R: 255, G: 200, B: 0, A: 255}
		}
	}

	switch f.genre {
	case "cyberpunk":
		return color.RGBA{R: 0, G: 220, B: 180, A: 255}
	case "horror":
		return color.RGBA{R: 200, G: 50, B: 50, A: 255}
	case "scifi":
		return color.RGBA{R: 150, G: 180, B: 255, A: 255}
	default:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
}

// SpawnImpactEffect creates an impact effect at the hit location.
func (f *FeedbackSystem) SpawnImpactEffect(x, y float64, itype ImpactType) {
	if len(f.impactEffects) >= f.maxImpacts {
		return
	}

	ie := &ImpactEffect{
		x:        x,
		y:        y,
		scale:    1.0,
		rotation: f.rng.Float64() * 6.28,
		lifetime: 0,
		maxLife:  0.3,
		itype:    itype,
	}

	f.impactEffects = append(f.impactEffects, ie)
}

// GetScreenShakeOffset returns the current camera shake offset.
func (f *FeedbackSystem) GetScreenShakeOffset() (float64, float64) {
	return f.screenShake.offsetX, f.screenShake.offsetY
}

// GetHitFlashIntensity returns the current hit flash intensity (0-1).
func (f *FeedbackSystem) GetHitFlashIntensity() float64 {
	return f.hitFlash.intensity
}

// GetHitFlashColor returns the current hit flash color.
func (f *FeedbackSystem) GetHitFlashColor() color.RGBA {
	c := f.hitFlash.color
	c.A = uint8(float64(c.A) * f.hitFlash.intensity)
	return c
}

// GetDamageNumbers returns all active damage numbers for rendering.
func (f *FeedbackSystem) GetDamageNumbers() []*DamageNumber {
	return f.damageNumbers
}

// GetImpactEffects returns all active impact effects for rendering.
func (f *FeedbackSystem) GetImpactEffects() []*ImpactEffect {
	return f.impactEffects
}

// FormatDamageNumber returns the display text for a damage number.
func (dn *DamageNumber) FormatDamageNumber() string {
	if dn.critical {
		return fmt.Sprintf("-%d!", dn.damage)
	}
	return fmt.Sprintf("-%d", dn.damage)
}

// GetAlpha returns the alpha value for a damage number based on lifetime.
func (dn *DamageNumber) GetAlpha() uint8 {
	progress := dn.lifetime / dn.maxLife
	alpha := 1.0 - progress
	if alpha < 0 {
		alpha = 0
	}
	return uint8(alpha * 255)
}

// GetScale returns the scale multiplier for a damage number.
func (dn *DamageNumber) GetScale() float64 {
	if dn.critical {
		return 1.5 + math.Sin(dn.lifetime*10.0)*0.2
	}
	return 1.0
}

// GetPosition returns the current position of the damage number.
func (dn *DamageNumber) GetPosition() (float64, float64) {
	return dn.x, dn.y
}

// GetColor returns the color of the damage number.
func (dn *DamageNumber) GetColor() color.RGBA {
	c := dn.color
	c.A = dn.GetAlpha()
	return c
}

// GetPosition returns the impact effect position.
func (ie *ImpactEffect) GetPosition() (float64, float64) {
	return ie.x, ie.y
}

// GetScale returns the impact effect scale.
func (ie *ImpactEffect) GetScale() float64 {
	return ie.scale
}

// GetRotation returns the impact effect rotation.
func (ie *ImpactEffect) GetRotation() float64 {
	return ie.rotation
}

// GetAlpha returns the alpha value for the impact effect.
func (ie *ImpactEffect) GetAlpha() uint8 {
	progress := ie.lifetime / ie.maxLife
	alpha := 1.0 - progress
	if alpha < 0 {
		alpha = 0
	}
	return uint8(alpha * 255)
}

// GetColor returns the color for the impact effect based on type.
func (ie *ImpactEffect) GetColor() color.RGBA {
	baseColor := color.RGBA{R: 255, G: 255, B: 255, A: 200}

	switch ie.itype {
	case ImpactCritical:
		baseColor = color.RGBA{R: 255, G: 200, B: 0, A: 255}
	case ImpactBlock:
		baseColor = color.RGBA{R: 100, G: 100, B: 255, A: 200}
	case ImpactMiss:
		baseColor = color.RGBA{R: 150, G: 150, B: 150, A: 150}
	default: // ImpactHit
		baseColor = color.RGBA{R: 255, G: 100, B: 100, A: 200}
	}

	baseColor.A = ie.GetAlpha()
	return baseColor
}

// GetType returns the impact effect type.
func (ie *ImpactEffect) GetType() ImpactType {
	return ie.itype
}
