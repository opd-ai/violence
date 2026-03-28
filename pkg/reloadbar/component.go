package reloadbar

// Component stores reload progress state for an entity.
type Component struct {
	// IsReloading indicates whether a reload is in progress.
	IsReloading bool

	// Progress is the current reload progress (0.0 to 1.0).
	Progress float64

	// TotalDuration is the total reload time in seconds.
	TotalDuration float64

	// ElapsedTime is the time spent reloading in seconds.
	ElapsedTime float64

	// FadeAlpha controls the visibility fade (0.0 to 1.0).
	FadeAlpha float64

	// WeaponName is the name of the weapon being reloaded.
	WeaponName string

	// AmmoCount is the current ammo count after reload completes.
	AmmoCount int

	// ClipSize is the maximum clip capacity.
	ClipSize int
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "reloadbar"
}

// NewComponent creates a new reload bar component with default values.
func NewComponent() *Component {
	return &Component{
		IsReloading:   false,
		Progress:      0.0,
		TotalDuration: 1.0,
		ElapsedTime:   0.0,
		FadeAlpha:     0.0,
		WeaponName:    "",
		AmmoCount:     0,
		ClipSize:      0,
	}
}

// StartReload begins a reload sequence.
func (c *Component) StartReload(duration float64, weaponName string, clipSize int) {
	c.IsReloading = true
	c.Progress = 0.0
	c.TotalDuration = duration
	c.ElapsedTime = 0.0
	c.WeaponName = weaponName
	c.ClipSize = clipSize
}

// UpdateProgress advances the reload progress.
func (c *Component) UpdateProgress(deltaTime float64) {
	if !c.IsReloading {
		return
	}

	c.ElapsedTime += deltaTime
	if c.TotalDuration > 0 {
		c.Progress = c.ElapsedTime / c.TotalDuration
		if c.Progress >= 1.0 {
			c.Progress = 1.0
			c.CompleteReload()
		}
	}
}

// CompleteReload finishes the reload sequence.
func (c *Component) CompleteReload() {
	c.IsReloading = false
	c.Progress = 1.0
	c.AmmoCount = c.ClipSize
}

// CancelReload aborts the reload sequence.
func (c *Component) CancelReload() {
	c.IsReloading = false
	c.Progress = 0.0
	c.ElapsedTime = 0.0
}
