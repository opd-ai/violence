package input

import "github.com/hajimehoshi/ebiten/v2"

// VirtualJoystick represents a floating virtual joystick for touch input
type VirtualJoystick struct {
	CenterX, CenterY float64 // Joystick center position
	KnobX, KnobY     float64 // Current knob position
	MaxRadius        float64 // Maximum distance from center
	DeadZone         float64 // Dead zone radius (no movement)
	Active           bool    // Whether joystick is currently active
	TouchID          ebiten.TouchID
}

// NewVirtualJoystick creates a new virtual joystick at the given position
func NewVirtualJoystick(centerX, centerY, maxRadius float64) *VirtualJoystick {
	return &VirtualJoystick{
		CenterX:   centerX,
		CenterY:   centerY,
		MaxRadius: maxRadius,
		DeadZone:  maxRadius * 0.1, // 10% dead zone
		Active:    false,
		TouchID:   -1,
	}
}

// GetAxis returns normalized axis values (-1.0 to 1.0)
func (vj *VirtualJoystick) GetAxis() (x, y float64) {
	if !vj.Active {
		return 0, 0
	}

	dx := vj.KnobX - vj.CenterX
	dy := vj.KnobY - vj.CenterY
	distance := dx*dx + dy*dy

	if distance < vj.DeadZone*vj.DeadZone {
		return 0, 0
	}

	// Normalize to max radius
	x = dx / vj.MaxRadius
	y = dy / vj.MaxRadius

	// Clamp to unit circle
	magnitude := x*x + y*y
	if magnitude > 1.0 {
		scale := 1.0 / (magnitude * 0.5) // sqrt approximation
		x *= scale
		y *= scale
	}

	return x, y
}

// HandleTouch processes touch input for the joystick
// Returns true if touch was handled by this joystick
func (vj *VirtualJoystick) HandleTouch(id ebiten.TouchID, x, y int) bool {
	// Touch down: activate joystick if in left screen region
	if !vj.Active && x < 300 { // Left 25% of typical screen
		vj.Active = true
		vj.TouchID = id
		vj.CenterX = float64(x)
		vj.CenterY = float64(y)
		vj.KnobX = float64(x)
		vj.KnobY = float64(y)
		return true
	}

	// Touch drag: update knob position
	if vj.Active && vj.TouchID == id {
		vj.KnobX = float64(x)
		vj.KnobY = float64(y)
		return true
	}

	return false
}

// Release deactivates the joystick
func (vj *VirtualJoystick) Release() {
	vj.Active = false
	vj.TouchID = -1
}

// TouchLookController handles touch-to-look camera control
type TouchLookController struct {
	RefX, RefY  int     // Reference touch position
	Sensitivity float64 // Degrees per pixel
	YawDelta    float64 // Accumulated yaw change
	PitchDelta  float64 // Accumulated pitch change
	TouchID     ebiten.TouchID
	Active      bool
}

// NewTouchLookController creates a new touch look controller
func NewTouchLookController(sensitivity float64) *TouchLookController {
	return &TouchLookController{
		Sensitivity: sensitivity,
		TouchID:     -1,
	}
}

// HandleTouch processes touch input for look control
func (tlc *TouchLookController) HandleTouch(id ebiten.TouchID, x, y int) {
	if !tlc.Active {
		tlc.Active = true
		tlc.TouchID = id
		tlc.RefX = x
		tlc.RefY = y
		tlc.YawDelta = 0
		tlc.PitchDelta = 0
		return
	}

	if tlc.TouchID == id {
		dx := x - tlc.RefX
		dy := y - tlc.RefY

		tlc.YawDelta = float64(dx) * tlc.Sensitivity
		tlc.PitchDelta = float64(dy) * tlc.Sensitivity

		// Clamp pitch to ±60 degrees
		if tlc.PitchDelta > 60 {
			tlc.PitchDelta = 60
		} else if tlc.PitchDelta < -60 {
			tlc.PitchDelta = -60
		}
	}
}

// GetDeltas returns accumulated camera deltas and resets them
func (tlc *TouchLookController) GetDeltas() (yaw, pitch float64) {
	yaw = tlc.YawDelta
	pitch = tlc.PitchDelta
	tlc.YawDelta = 0
	tlc.PitchDelta = 0
	return yaw, pitch
}

// Release deactivates the controller
func (tlc *TouchLookController) Release() {
	tlc.Active = false
	tlc.TouchID = -1
}

// TouchButton represents a touch-activated button
type TouchButton struct {
	X, Y    float64 // Center position (normalized 0-1)
	Radius  float64 // Button radius in pixels
	Label   string  // Button label
	Active  bool    // Currently pressed
	TouchID ebiten.TouchID
}

// NewTouchButton creates a new touch button
func NewTouchButton(x, y float64, label string) *TouchButton {
	return &TouchButton{
		X:       x,
		Y:       y,
		Radius:  50, // Default 50px radius
		Label:   label,
		TouchID: -1,
	}
}

// HandleTouch processes touch input for the button
// Returns true if button was pressed/released
func (tb *TouchButton) HandleTouch(id ebiten.TouchID, x, y, screenW, screenH int) bool {
	// Convert normalized position to screen coords
	btnX := tb.X * float64(screenW)
	btnY := tb.Y * float64(screenH)

	dx := float64(x) - btnX
	dy := float64(y) - btnY
	distSq := dx*dx + dy*dy

	// Check if touch is within button radius
	if distSq <= tb.Radius*tb.Radius {
		if !tb.Active {
			tb.Active = true
			tb.TouchID = id
			return true
		}
	}

	return false
}

// IsPressed returns whether the button is currently pressed
func (tb *TouchButton) IsPressed() bool {
	return tb.Active
}

// Release deactivates the button
func (tb *TouchButton) Release() {
	tb.Active = false
	tb.TouchID = -1
}

// TouchInputManager coordinates all touch input controls
type TouchInputManager struct {
	Joystick    *VirtualJoystick
	LookControl *TouchLookController
	FireButton  *TouchButton
	AltButton   *TouchButton
	ActionBar   []*TouchButton
}

// NewTouchInputManager creates a new touch input manager
func NewTouchInputManager() *TouchInputManager {
	return &TouchInputManager{
		Joystick:    NewVirtualJoystick(100, 400, 80),
		LookControl: NewTouchLookController(0.15),
		FireButton:  NewTouchButton(0.85, 0.4, "Fire"),
		AltButton:   NewTouchButton(0.85, 0.55, "Alt"),
		ActionBar: []*TouchButton{
			NewTouchButton(0.125, 0.92, "Map"),
			NewTouchButton(0.292, 0.92, "Inv"),
			NewTouchButton(0.458, 0.92, "Pause"),
			NewTouchButton(0.625, 0.92, "Jump"),
			NewTouchButton(0.792, 0.92, "Reload"),
			NewTouchButton(0.958, 0.92, "Interact"),
		},
	}
}

// Update processes all touch input for the current frame
func (tim *TouchInputManager) Update() {
	touches := ebiten.AppendTouchIDs(nil)

	// If no touches, release all controls
	if len(touches) == 0 {
		tim.Joystick.Release()
		tim.LookControl.Release()
		tim.FireButton.Release()
		tim.AltButton.Release()
		for _, btn := range tim.ActionBar {
			btn.Release()
		}
		return
	}

	// Process each touch
	for _, id := range touches {
		x, y := ebiten.TouchPosition(id)
		screenW, screenH := ebiten.WindowSize()

		// Route touch to appropriate control
		if tim.Joystick.HandleTouch(id, x, y) {
			continue
		}
		if tim.FireButton.HandleTouch(id, x, y, screenW, screenH) {
			continue
		}
		if tim.AltButton.HandleTouch(id, x, y, screenW, screenH) {
			continue
		}

		// Check action bar
		handled := false
		for _, btn := range tim.ActionBar {
			if btn.HandleTouch(id, x, y, screenW, screenH) {
				handled = true
				break
			}
		}
		if handled {
			continue
		}

		// Default: look control (center area)
		tim.LookControl.HandleTouch(id, x, y)
	}
}
