// Package weapon implements the weapon and firing system.
package weapon

import (
	"math"
	"math/rand"
)

// WeaponType defines the weapon firing mechanic.
type WeaponType int

const (
	TypeHitscan    WeaponType = iota // Instant ray-cast hit
	TypeProjectile                   // In-world projectile simulation
	TypeMelee                        // Short-range melee attack
)

// AnimState represents weapon animation state.
type AnimState int

const (
	AnimIdle   AnimState = iota // Weapon ready, idle bobbing
	AnimRaise                   // Weapon being raised
	AnimLower                   // Weapon being lowered
	AnimFire                    // Weapon firing
	AnimReload                  // Weapon reloading
)

// Weapon represents a player weapon.
type Weapon struct {
	Name        string
	Type        WeaponType
	Damage      float64
	FireRate    float64 // Frames between shots (60 TPS)
	AmmoType    string  // bullets, shells, cells, rockets
	ClipSize    int
	SpreadAngle float64 // Degrees; for shotgun multi-ray spread
	RayCount    int     // Number of rays per shot (shotgun = 7, others = 1)
	Range       float64 // Max distance; melee = 1.5, hitscan = 100
	Projectile  bool    // True if spawns projectile entity
}

// AnimFrame represents a single animation frame with procedural parameters.
type AnimFrame struct {
	OffsetX    float64 // Horizontal offset
	OffsetY    float64 // Vertical offset
	Scale      float64 // Size scale
	Rotation   float64 // Rotation in radians
	Brightness float64 // Brightness multiplier (0-1)
}

// Animation holds frames for a specific animation state.
type Animation struct {
	Frames        []AnimFrame
	FrameDuration int // Frames per animation frame at 60 TPS
	Loop          bool
}

// WeaponAnimator manages weapon animation state.
type WeaponAnimator struct {
	CurrentState AnimState
	CurrentFrame int
	FrameCounter int
	Animations   map[AnimState]Animation
	Seed         int64
}

// Arsenal manages the player's collection of weapons.
type Arsenal struct {
	Weapons         []Weapon
	CurrentSlot     int
	Ammo            map[string]int // AmmoType -> count
	Clips           map[int]int    // Weapon slot -> ammo in clip
	FramesSinceFire map[int]int    // Weapon slot -> cooldown counter
	genre           string
	Animator        *WeaponAnimator
}

// NewArsenal creates an empty arsenal with default weapons.
func NewArsenal() *Arsenal {
	a := &Arsenal{
		Weapons:         make([]Weapon, 7),
		CurrentSlot:     1,
		Ammo:            make(map[string]int),
		Clips:           make(map[int]int),
		FramesSinceFire: make(map[int]int),
		genre:           "fantasy",
		Animator:        NewWeaponAnimator(42),
	}
	a.loadDefaultWeapons()
	// Initialize cooldowns to allow immediate fire
	for i := range a.Weapons {
		a.FramesSinceFire[i] = 1000
	}
	return a
}

// loadDefaultWeapons initializes the 7-weapon loadout.
func (a *Arsenal) loadDefaultWeapons() {
	a.Weapons[0] = Weapon{Name: "Fist", Type: TypeMelee, Damage: 10, FireRate: 20, Range: 1.2, RayCount: 1}
	a.Weapons[1] = Weapon{Name: "Pistol", Type: TypeHitscan, Damage: 15, FireRate: 15, AmmoType: "bullets", ClipSize: 12, Range: 100, RayCount: 1}
	a.Weapons[2] = Weapon{Name: "Shotgun", Type: TypeHitscan, Damage: 10, FireRate: 30, AmmoType: "shells", ClipSize: 8, SpreadAngle: 10, RayCount: 7, Range: 30}
	a.Weapons[3] = Weapon{Name: "Chaingun", Type: TypeHitscan, Damage: 12, FireRate: 5, AmmoType: "bullets", ClipSize: 100, Range: 100, RayCount: 1}
	a.Weapons[4] = Weapon{Name: "Rocket Launcher", Type: TypeProjectile, Damage: 100, FireRate: 45, AmmoType: "rockets", ClipSize: 5, Range: 200, RayCount: 1, Projectile: true}
	a.Weapons[5] = Weapon{Name: "Plasma Gun", Type: TypeProjectile, Damage: 40, FireRate: 10, AmmoType: "cells", ClipSize: 40, Range: 150, RayCount: 1, Projectile: true}
	a.Weapons[6] = Weapon{Name: "Knife", Type: TypeMelee, Damage: 25, FireRate: 18, Range: 1.5, RayCount: 1}

	// Initialize ammo pools
	a.Ammo["bullets"] = 50
	a.Ammo["shells"] = 8
	a.Ammo["cells"] = 40
	a.Ammo["rockets"] = 5

	// Initialize clips
	for i := range a.Weapons {
		a.Clips[i] = a.Weapons[i].ClipSize
	}
}

// HitResult contains the result of a weapon firing.
type HitResult struct {
	Hit      bool
	Distance float64
	Damage   float64
	HitX     float64
	HitY     float64
	EntityID uint64 // 0 if no entity hit, otherwise entity ID
}

// Fire discharges the current weapon.
// Returns hit results for each ray cast (shotgun = 7, others = 1).
// posX, posY: shooter position; dirX, dirY: aim direction normalized.
// raycast: function that casts a ray and returns (hit, distance, hitX, hitY, entityID).
func (a *Arsenal) Fire(posX, posY, dirX, dirY float64, raycast func(x, y, dx, dy, maxDist float64) (bool, float64, float64, float64, uint64)) []HitResult {
	weapon := a.Weapons[a.CurrentSlot]

	// Check cooldown
	if a.FramesSinceFire[a.CurrentSlot] < int(weapon.FireRate) {
		return nil
	}

	// Check ammo for non-melee
	if weapon.Type != TypeMelee {
		if a.Clips[a.CurrentSlot] <= 0 {
			return nil // Out of ammo
		}
		a.Clips[a.CurrentSlot]--
	}

	// Reset cooldown
	a.FramesSinceFire[a.CurrentSlot] = 0

	// Trigger fire animation
	if a.Animator != nil {
		a.Animator.SetState(AnimFire)
	}

	results := make([]HitResult, 0, weapon.RayCount)

	for i := 0; i < weapon.RayCount; i++ {
		// Calculate spread offset for shotgun
		spreadOffset := 0.0
		if weapon.RayCount > 1 {
			// Distribute rays across spread angle
			spreadOffset = weapon.SpreadAngle * (float64(i)/float64(weapon.RayCount-1) - 0.5) * math.Pi / 180.0
		}

		// Rotate direction by spread offset
		cos := math.Cos(spreadOffset)
		sin := math.Sin(spreadOffset)
		rayDirX := dirX*cos - dirY*sin
		rayDirY := dirX*sin + dirY*cos

		// Cast ray
		hit, dist, hitX, hitY, entityID := raycast(posX, posY, rayDirX, rayDirY, weapon.Range)

		result := HitResult{
			Hit:      hit,
			Distance: dist,
			HitX:     hitX,
			HitY:     hitY,
			EntityID: entityID,
		}

		if hit && dist <= weapon.Range {
			result.Damage = weapon.Damage
		}

		results = append(results, result)
	}

	return results
}

// Reload reloads the current weapon from the ammo pool.
func (a *Arsenal) Reload() bool {
	weapon := a.Weapons[a.CurrentSlot]

	// Melee weapons don't reload
	if weapon.Type == TypeMelee {
		return false
	}

	// Already full
	if a.Clips[a.CurrentSlot] >= weapon.ClipSize {
		return false
	}

	// Check ammo pool
	available := a.Ammo[weapon.AmmoType]
	if available <= 0 {
		return false
	}

	// Calculate ammo needed
	needed := weapon.ClipSize - a.Clips[a.CurrentSlot]
	toReload := needed
	if available < needed {
		toReload = available
	}

	// Transfer ammo from pool to clip
	a.Clips[a.CurrentSlot] += toReload
	a.Ammo[weapon.AmmoType] -= toReload

	// Trigger reload animation
	if a.Animator != nil {
		a.Animator.SetState(AnimReload)
	}

	return true
}

// SwitchTo changes the active weapon slot (0-6).
func (a *Arsenal) SwitchTo(slot int) bool {
	if slot < 0 || slot >= len(a.Weapons) {
		return false
	}

	// Trigger lower animation for current weapon, then raise for new
	if a.Animator != nil && slot != a.CurrentSlot {
		a.Animator.SetState(AnimLower)
	}

	a.CurrentSlot = slot

	// After switching, trigger raise animation
	if a.Animator != nil {
		a.Animator.SetState(AnimRaise)
	}

	return true
}

// Update increments frame counters for cooldown tracking and animations.
func (a *Arsenal) Update() {
	for i := range a.FramesSinceFire {
		a.FramesSinceFire[i]++
	}

	// Update weapon animation
	if a.Animator != nil {
		a.Animator.UpdateAnimation()
	}
}

// GetCurrentWeapon returns the active weapon.
func (a *Arsenal) GetCurrentWeapon() Weapon {
	return a.Weapons[a.CurrentSlot]
}

// AddAmmo adds ammo to the pool.
func (a *Arsenal) AddAmmo(ammoType string, amount int) {
	a.Ammo[ammoType] += amount
}

// SetGenre configures weapon names and visuals for a genre.
func (a *Arsenal) SetGenre(genreID string) {
	a.genre = genreID
	a.applyGenreNames()
}

// applyGenreNames remaps weapon names per genre.
func (a *Arsenal) applyGenreNames() {
	switch a.genre {
	case "scifi":
		a.Weapons[1].Name = "Blaster"
		a.Weapons[2].Name = "Scatter Cannon"
		a.Weapons[3].Name = "Pulse Rifle"
		a.Weapons[4].Name = "Missile Launcher"
		a.Weapons[5].Name = "Plasma Gun"
		a.Weapons[6].Name = "Combat Knife"
	case "horror":
		a.Weapons[1].Name = "Revolver"
		a.Weapons[2].Name = "Sawed-off Shotgun"
		a.Weapons[3].Name = "Submachine Gun"
		a.Weapons[4].Name = "Grenade Launcher"
		a.Weapons[5].Name = "Flamethrower"
		a.Weapons[6].Name = "Rusty Blade"
	case "cyberpunk":
		a.Weapons[1].Name = "Smart Pistol"
		a.Weapons[2].Name = "Auto-Shotgun"
		a.Weapons[3].Name = "Minigun"
		a.Weapons[4].Name = "Rocket Pod"
		a.Weapons[5].Name = "Energy Rifle"
		a.Weapons[6].Name = "Mono-Blade"
	case "postapoc":
		a.Weapons[1].Name = "Makeshift Pistol"
		a.Weapons[2].Name = "Pipe Shotgun"
		a.Weapons[3].Name = "Scrap Rifle"
		a.Weapons[4].Name = "Improvised Launcher"
		a.Weapons[5].Name = "Jury-Rigged Laser"
		a.Weapons[6].Name = "Sharpened Rebar"
	default: // fantasy
		a.Weapons[1].Name = "Crossbow"
		a.Weapons[2].Name = "Blunderbuss"
		a.Weapons[3].Name = "Repeating Crossbow"
		a.Weapons[4].Name = "Explosive Orb"
		a.Weapons[5].Name = "Arcane Staff"
		a.Weapons[6].Name = "Dagger"
	}
}

// NewWeaponAnimator creates an animator with procedurally generated animations.
func NewWeaponAnimator(seed int64) *WeaponAnimator {
	wa := &WeaponAnimator{
		CurrentState: AnimIdle,
		CurrentFrame: 0,
		FrameCounter: 0,
		Animations:   make(map[AnimState]Animation),
		Seed:         seed,
	}
	wa.generateAnimations()
	return wa
}

// generateAnimations procedurally generates all animation frames from seed.
func (wa *WeaponAnimator) generateAnimations() {
	rng := rand.New(rand.NewSource(wa.Seed))

	wa.Animations[AnimIdle] = wa.generateIdleAnimation(rng)
	wa.Animations[AnimRaise] = wa.generateRaiseAnimation(rng)
	wa.Animations[AnimLower] = wa.generateLowerAnimation(rng)
	wa.Animations[AnimFire] = wa.generateFireAnimation(rng)
	wa.Animations[AnimReload] = wa.generateReloadAnimation(rng)
}

// generateIdleAnimation creates idle bobbing animation.
func (wa *WeaponAnimator) generateIdleAnimation(rng *rand.Rand) Animation {
	frames := make([]AnimFrame, 30)
	for i := range frames {
		t := float64(i) / float64(len(frames))
		bobY := math.Sin(t*2*math.Pi) * 0.02
		bobX := math.Cos(t*2*math.Pi) * 0.01
		frames[i] = AnimFrame{
			OffsetX:    bobX,
			OffsetY:    bobY,
			Scale:      1.0,
			Rotation:   0,
			Brightness: 1.0,
		}
	}
	return Animation{Frames: frames, FrameDuration: 2, Loop: true}
}

// generateRaiseAnimation creates weapon raise animation.
func (wa *WeaponAnimator) generateRaiseAnimation(rng *rand.Rand) Animation {
	frames := make([]AnimFrame, 8)
	for i := range frames {
		t := float64(i) / float64(len(frames)-1)
		offsetY := (1.0 - t) * 0.5
		scale := 0.5 + t*0.5
		frames[i] = AnimFrame{
			OffsetX:    0,
			OffsetY:    offsetY,
			Scale:      scale,
			Rotation:   0,
			Brightness: 0.6 + t*0.4,
		}
	}
	return Animation{Frames: frames, FrameDuration: 2, Loop: false}
}

// generateLowerAnimation creates weapon lower animation.
func (wa *WeaponAnimator) generateLowerAnimation(rng *rand.Rand) Animation {
	frames := make([]AnimFrame, 8)
	for i := range frames {
		t := float64(i) / float64(len(frames)-1)
		offsetY := t * 0.5
		scale := 1.0 - t*0.5
		frames[i] = AnimFrame{
			OffsetX:    0,
			OffsetY:    offsetY,
			Scale:      scale,
			Rotation:   0,
			Brightness: 1.0 - t*0.4,
		}
	}
	return Animation{Frames: frames, FrameDuration: 2, Loop: false}
}

// generateFireAnimation creates weapon fire animation.
func (wa *WeaponAnimator) generateFireAnimation(rng *rand.Rand) Animation {
	frames := make([]AnimFrame, 6)
	for i := range frames {
		recoil := 0.0
		if i < 2 {
			recoil = -0.1 * (1.0 - float64(i)/2.0)
		}
		flash := 0.0
		if i < 2 {
			flash = 0.3 * (1.0 - float64(i)/2.0)
		}
		frames[i] = AnimFrame{
			OffsetX:    rng.Float64()*0.01 - 0.005,
			OffsetY:    recoil,
			Scale:      1.0,
			Rotation:   (rng.Float64()*0.04 - 0.02),
			Brightness: 1.0 + flash,
		}
	}
	return Animation{Frames: frames, FrameDuration: 1, Loop: false}
}

// generateReloadAnimation creates weapon reload animation.
func (wa *WeaponAnimator) generateReloadAnimation(rng *rand.Rand) Animation {
	frames := make([]AnimFrame, 20)
	for i := range frames {
		t := float64(i) / float64(len(frames)-1)
		offsetY := math.Sin(t*math.Pi) * 0.15
		rotation := math.Sin(t*math.Pi) * 0.2
		frames[i] = AnimFrame{
			OffsetX:    0,
			OffsetY:    offsetY,
			Scale:      1.0,
			Rotation:   rotation,
			Brightness: 1.0,
		}
	}
	return Animation{Frames: frames, FrameDuration: 2, Loop: false}
}

// SetState transitions to a new animation state.
func (wa *WeaponAnimator) SetState(state AnimState) {
	if wa.CurrentState == state {
		return
	}
	wa.CurrentState = state
	wa.CurrentFrame = 0
	wa.FrameCounter = 0
}

// Update advances the animation state machine.
func (wa *WeaponAnimator) UpdateAnimation() {
	anim, ok := wa.Animations[wa.CurrentState]
	if !ok {
		return
	}

	wa.FrameCounter++
	if wa.FrameCounter >= anim.FrameDuration {
		wa.FrameCounter = 0
		wa.CurrentFrame++

		if wa.CurrentFrame >= len(anim.Frames) {
			if anim.Loop {
				wa.CurrentFrame = 0
			} else {
				wa.CurrentFrame = len(anim.Frames) - 1
				if wa.CurrentState != AnimIdle {
					wa.SetState(AnimIdle)
				}
			}
		}
	}
}

// GetCurrentFrame returns the current animation frame.
func (wa *WeaponAnimator) GetCurrentFrame() AnimFrame {
	anim, ok := wa.Animations[wa.CurrentState]
	if !ok || len(anim.Frames) == 0 {
		return AnimFrame{Scale: 1.0, Brightness: 1.0}
	}
	if wa.CurrentFrame >= len(anim.Frames) {
		return anim.Frames[len(anim.Frames)-1]
	}
	return anim.Frames[wa.CurrentFrame]
}

// FireProjectile spawns a projectile entity for projectile weapons.
// Returns projectile initial velocity and true if projectile spawned.
func (a *Arsenal) FireProjectile(posX, posY, dirX, dirY float64, rng *rand.Rand) (velX, velY float64, spawned bool) {
	weapon := a.Weapons[a.CurrentSlot]

	if weapon.Type != TypeProjectile {
		return 0, 0, false
	}

	// Check cooldown
	if a.FramesSinceFire[a.CurrentSlot] < int(weapon.FireRate) {
		return 0, 0, false
	}

	// Check ammo
	if a.Clips[a.CurrentSlot] <= 0 {
		return 0, 0, false
	}

	a.Clips[a.CurrentSlot]--
	a.FramesSinceFire[a.CurrentSlot] = 0

	// Projectile velocity based on weapon
	speed := 0.3 // units per frame at 60 TPS
	if weapon.Name == "Rocket Launcher" || weapon.Name == "Missile Launcher" || weapon.Name == "Improvised Launcher" {
		speed = 0.25
	}

	return dirX * speed, dirY * speed, true
}

// SetGenre configures the weapon system for a genre (global).
func SetGenre(genreID string) {
	// Global genre setting would affect default weapon templates
	// For now, genre is applied per-arsenal instance
}
