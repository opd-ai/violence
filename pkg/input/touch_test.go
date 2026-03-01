package input

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewVirtualJoystick(t *testing.T) {
	tests := []struct {
		name      string
		centerX   float64
		centerY   float64
		maxRadius float64
		wantDZ    float64
	}{
		{"standard", 100, 200, 80, 8},
		{"large", 150, 300, 120, 12},
		{"small", 50, 100, 40, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vj := NewVirtualJoystick(tt.centerX, tt.centerY, tt.maxRadius)
			if vj.CenterX != tt.centerX {
				t.Errorf("CenterX = %v, want %v", vj.CenterX, tt.centerX)
			}
			if vj.CenterY != tt.centerY {
				t.Errorf("CenterY = %v, want %v", vj.CenterY, tt.centerY)
			}
			if vj.MaxRadius != tt.maxRadius {
				t.Errorf("MaxRadius = %v, want %v", vj.MaxRadius, tt.maxRadius)
			}
			if vj.DeadZone != tt.wantDZ {
				t.Errorf("DeadZone = %v, want %v", vj.DeadZone, tt.wantDZ)
			}
			if vj.Active {
				t.Error("joystick should not be active initially")
			}
		})
	}
}

func TestVirtualJoystickGetAxis(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*VirtualJoystick)
		wantX   float64
		wantY   float64
		epsilon float64
	}{
		{
			name: "inactive",
			setup: func(vj *VirtualJoystick) {
				vj.Active = false
			},
			wantX:   0,
			wantY:   0,
			epsilon: 0.001,
		},
		{
			name: "within_dead_zone",
			setup: func(vj *VirtualJoystick) {
				vj.Active = true
				vj.CenterX = 100
				vj.CenterY = 100
				vj.KnobX = 102
				vj.KnobY = 102
				vj.DeadZone = 10
			},
			wantX:   0,
			wantY:   0,
			epsilon: 0.001,
		},
		{
			name: "full_right",
			setup: func(vj *VirtualJoystick) {
				vj.Active = true
				vj.CenterX = 100
				vj.CenterY = 100
				vj.MaxRadius = 80
				vj.KnobX = 180
				vj.KnobY = 100
				vj.DeadZone = 8
			},
			wantX:   1.0,
			wantY:   0,
			epsilon: 0.1,
		},
		{
			name: "full_down",
			setup: func(vj *VirtualJoystick) {
				vj.Active = true
				vj.CenterX = 100
				vj.CenterY = 100
				vj.MaxRadius = 80
				vj.KnobX = 100
				vj.KnobY = 180
				vj.DeadZone = 8
			},
			wantX:   0,
			wantY:   1.0,
			epsilon: 0.1,
		},
		{
			name: "diagonal",
			setup: func(vj *VirtualJoystick) {
				vj.Active = true
				vj.CenterX = 100
				vj.CenterY = 100
				vj.MaxRadius = 80
				vj.KnobX = 140
				vj.KnobY = 140
				vj.DeadZone = 8
			},
			wantX:   0.5,
			wantY:   0.5,
			epsilon: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vj := NewVirtualJoystick(100, 100, 80)
			tt.setup(vj)
			x, y := vj.GetAxis()
			if math.Abs(x-tt.wantX) > tt.epsilon {
				t.Errorf("GetAxis() x = %v, want %v (±%v)", x, tt.wantX, tt.epsilon)
			}
			if math.Abs(y-tt.wantY) > tt.epsilon {
				t.Errorf("GetAxis() y = %v, want %v (±%v)", y, tt.wantY, tt.epsilon)
			}
		})
	}
}

func TestVirtualJoystickHandleTouch(t *testing.T) {
	tests := []struct {
		name         string
		initialState func(*VirtualJoystick)
		touchID      ebiten.TouchID
		x            int
		y            int
		wantHandled  bool
		wantActive   bool
	}{
		{
			name: "activate_left_region",
			initialState: func(vj *VirtualJoystick) {
				vj.Active = false
			},
			touchID:     1,
			x:           100,
			y:           200,
			wantHandled: true,
			wantActive:  true,
		},
		{
			name: "ignore_right_region",
			initialState: func(vj *VirtualJoystick) {
				vj.Active = false
			},
			touchID:     1,
			x:           400,
			y:           200,
			wantHandled: false,
			wantActive:  false,
		},
		{
			name: "drag_existing",
			initialState: func(vj *VirtualJoystick) {
				vj.Active = true
				vj.TouchID = 1
				vj.CenterX = 100
				vj.CenterY = 200
			},
			touchID:     1,
			x:           120,
			y:           220,
			wantHandled: true,
			wantActive:  true,
		},
		{
			name: "ignore_different_touch",
			initialState: func(vj *VirtualJoystick) {
				vj.Active = true
				vj.TouchID = 1
			},
			touchID:     2,
			x:           120,
			y:           220,
			wantHandled: false,
			wantActive:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vj := NewVirtualJoystick(100, 200, 80)
			tt.initialState(vj)
			handled := vj.HandleTouch(tt.touchID, tt.x, tt.y)
			if handled != tt.wantHandled {
				t.Errorf("HandleTouch() handled = %v, want %v", handled, tt.wantHandled)
			}
			if vj.Active != tt.wantActive {
				t.Errorf("Active = %v, want %v", vj.Active, tt.wantActive)
			}
		})
	}
}

func TestVirtualJoystickRelease(t *testing.T) {
	vj := NewVirtualJoystick(100, 200, 80)
	vj.Active = true
	vj.TouchID = 5
	vj.Release()
	if vj.Active {
		t.Error("joystick should be inactive after Release()")
	}
	if vj.TouchID != -1 {
		t.Errorf("TouchID = %v, want -1", vj.TouchID)
	}
}

func TestNewTouchLookController(t *testing.T) {
	tests := []struct {
		name        string
		sensitivity float64
	}{
		{"standard", 0.15},
		{"high", 0.30},
		{"low", 0.05},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlc := NewTouchLookController(tt.sensitivity)
			if tlc.Sensitivity != tt.sensitivity {
				t.Errorf("Sensitivity = %v, want %v", tlc.Sensitivity, tt.sensitivity)
			}
			if tlc.Active {
				t.Error("controller should not be active initially")
			}
			if tlc.TouchID != -1 {
				t.Errorf("TouchID = %v, want -1", tlc.TouchID)
			}
		})
	}
}

func TestTouchLookControllerHandleTouch(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*TouchLookController)
		touchID   ebiten.TouchID
		x         int
		y         int
		wantYaw   float64
		wantPitch float64
		epsilon   float64
	}{
		{
			name: "initial_touch",
			setup: func(tlc *TouchLookController) {
				tlc.Active = false
			},
			touchID:   1,
			x:         100,
			y:         200,
			wantYaw:   0,
			wantPitch: 0,
			epsilon:   0.001,
		},
		{
			name: "look_right",
			setup: func(tlc *TouchLookController) {
				tlc.Active = true
				tlc.TouchID = 1
				tlc.RefX = 100
				tlc.RefY = 200
				tlc.Sensitivity = 0.15
			},
			touchID:   1,
			x:         200,
			y:         200,
			wantYaw:   15.0,
			wantPitch: 0,
			epsilon:   0.1,
		},
		{
			name: "look_down",
			setup: func(tlc *TouchLookController) {
				tlc.Active = true
				tlc.TouchID = 1
				tlc.RefX = 100
				tlc.RefY = 200
				tlc.Sensitivity = 0.15
			},
			touchID:   1,
			x:         100,
			y:         300,
			wantYaw:   0,
			wantPitch: 15.0,
			epsilon:   0.1,
		},
		{
			name: "pitch_clamp_positive",
			setup: func(tlc *TouchLookController) {
				tlc.Active = true
				tlc.TouchID = 1
				tlc.RefX = 100
				tlc.RefY = 200
				tlc.Sensitivity = 0.15
			},
			touchID:   1,
			x:         100,
			y:         700,
			wantYaw:   0,
			wantPitch: 60.0,
			epsilon:   0.1,
		},
		{
			name: "pitch_clamp_negative",
			setup: func(tlc *TouchLookController) {
				tlc.Active = true
				tlc.TouchID = 1
				tlc.RefX = 100
				tlc.RefY = 700
				tlc.Sensitivity = 0.15
			},
			touchID:   1,
			x:         100,
			y:         100,
			wantYaw:   0,
			wantPitch: -60.0,
			epsilon:   0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlc := NewTouchLookController(0.15)
			tt.setup(tlc)
			tlc.HandleTouch(tt.touchID, tt.x, tt.y)
			if math.Abs(tlc.YawDelta-tt.wantYaw) > tt.epsilon {
				t.Errorf("YawDelta = %v, want %v (±%v)", tlc.YawDelta, tt.wantYaw, tt.epsilon)
			}
			if math.Abs(tlc.PitchDelta-tt.wantPitch) > tt.epsilon {
				t.Errorf("PitchDelta = %v, want %v (±%v)", tlc.PitchDelta, tt.wantPitch, tt.epsilon)
			}
		})
	}
}

func TestTouchLookControllerGetDeltas(t *testing.T) {
	tlc := NewTouchLookController(0.15)
	tlc.YawDelta = 10.0
	tlc.PitchDelta = 5.0

	yaw, pitch := tlc.GetDeltas()
	if yaw != 10.0 {
		t.Errorf("GetDeltas() yaw = %v, want 10.0", yaw)
	}
	if pitch != 5.0 {
		t.Errorf("GetDeltas() pitch = %v, want 5.0", pitch)
	}

	// Deltas should be reset
	if tlc.YawDelta != 0 {
		t.Errorf("YawDelta = %v, want 0 after GetDeltas()", tlc.YawDelta)
	}
	if tlc.PitchDelta != 0 {
		t.Errorf("PitchDelta = %v, want 0 after GetDeltas()", tlc.PitchDelta)
	}
}

func TestTouchLookControllerRelease(t *testing.T) {
	tlc := NewTouchLookController(0.15)
	tlc.Active = true
	tlc.TouchID = 5
	tlc.Release()
	if tlc.Active {
		t.Error("controller should be inactive after Release()")
	}
	if tlc.TouchID != -1 {
		t.Errorf("TouchID = %v, want -1", tlc.TouchID)
	}
}

func TestNewTouchButton(t *testing.T) {
	tests := []struct {
		name  string
		x     float64
		y     float64
		label string
	}{
		{"fire", 0.85, 0.4, "Fire"},
		{"jump", 0.625, 0.92, "Jump"},
		{"pause", 0.5, 0.1, "Pause"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btn := NewTouchButton(tt.x, tt.y, tt.label)
			if btn.X != tt.x {
				t.Errorf("X = %v, want %v", btn.X, tt.x)
			}
			if btn.Y != tt.y {
				t.Errorf("Y = %v, want %v", btn.Y, tt.y)
			}
			if btn.Label != tt.label {
				t.Errorf("Label = %v, want %v", btn.Label, tt.label)
			}
			if btn.Active {
				t.Error("button should not be active initially")
			}
			if btn.Radius != 50 {
				t.Errorf("Radius = %v, want 50", btn.Radius)
			}
		})
	}
}

func TestTouchButtonHandleTouch(t *testing.T) {
	tests := []struct {
		name        string
		btnX        float64
		btnY        float64
		touchX      int
		touchY      int
		screenW     int
		screenH     int
		wantHandled bool
		wantActive  bool
	}{
		{
			name:        "inside_button",
			btnX:        0.5,
			btnY:        0.5,
			touchX:      400,
			touchY:      300,
			screenW:     800,
			screenH:     600,
			wantHandled: true,
			wantActive:  true,
		},
		{
			name:        "outside_button",
			btnX:        0.5,
			btnY:        0.5,
			touchX:      100,
			touchY:      100,
			screenW:     800,
			screenH:     600,
			wantHandled: false,
			wantActive:  false,
		},
		{
			name:        "edge_of_button",
			btnX:        0.5,
			btnY:        0.5,
			touchX:      450,
			touchY:      300,
			screenW:     800,
			screenH:     600,
			wantHandled: true,
			wantActive:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btn := NewTouchButton(tt.btnX, tt.btnY, "Test")
			handled := btn.HandleTouch(1, tt.touchX, tt.touchY, tt.screenW, tt.screenH)
			if handled != tt.wantHandled {
				t.Errorf("HandleTouch() handled = %v, want %v", handled, tt.wantHandled)
			}
			if btn.Active != tt.wantActive {
				t.Errorf("Active = %v, want %v", btn.Active, tt.wantActive)
			}
		})
	}
}

func TestTouchButtonIsPressed(t *testing.T) {
	btn := NewTouchButton(0.5, 0.5, "Test")
	if btn.IsPressed() {
		t.Error("button should not be pressed initially")
	}
	btn.Active = true
	if !btn.IsPressed() {
		t.Error("button should be pressed when Active is true")
	}
}

func TestTouchButtonRelease(t *testing.T) {
	btn := NewTouchButton(0.5, 0.5, "Test")
	btn.Active = true
	btn.TouchID = 5
	btn.Release()
	if btn.Active {
		t.Error("button should be inactive after Release()")
	}
	if btn.TouchID != -1 {
		t.Errorf("TouchID = %v, want -1", btn.TouchID)
	}
}

func TestNewTouchInputManager(t *testing.T) {
	tim := NewTouchInputManager()
	if tim.Joystick == nil {
		t.Error("Joystick should not be nil")
	}
	if tim.LookControl == nil {
		t.Error("LookControl should not be nil")
	}
	if tim.FireButton == nil {
		t.Error("FireButton should not be nil")
	}
	if tim.AltButton == nil {
		t.Error("AltButton should not be nil")
	}
	if len(tim.ActionBar) != 6 {
		t.Errorf("ActionBar length = %v, want 6", len(tim.ActionBar))
	}
}

func TestTouchInputManagerActionBarLabels(t *testing.T) {
	tim := NewTouchInputManager()
	expectedLabels := []string{"Map", "Inv", "Pause", "Jump", "Reload", "Interact"}
	for i, btn := range tim.ActionBar {
		if btn.Label != expectedLabels[i] {
			t.Errorf("ActionBar[%d].Label = %v, want %v", i, btn.Label, expectedLabels[i])
		}
	}
}

func TestTouchInputManagerScreenSizeIndependence(t *testing.T) {
	// Test that button positions scale with different screen sizes
	tests := []struct {
		name    string
		screenW int
		screenH int
	}{
		{"4.7inch", 750, 1334},        // iPhone SE
		{"6.1inch", 828, 1792},        // iPhone 11
		{"13inch_tablet", 2048, 2732}, // iPad Pro
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tim := NewTouchInputManager()

			// Fire button should be at 85% of screen width
			fireX := tim.FireButton.X * float64(tt.screenW)
			expectedFireX := 0.85 * float64(tt.screenW)
			if math.Abs(fireX-expectedFireX) > 1.0 {
				t.Errorf("Fire button X = %v, want %v", fireX, expectedFireX)
			}

			// Action bar should be at 92% of screen height
			for i, btn := range tim.ActionBar {
				btnY := btn.Y * float64(tt.screenH)
				expectedY := 0.92 * float64(tt.screenH)
				if math.Abs(btnY-expectedY) > 1.0 {
					t.Errorf("ActionBar[%d] Y = %v, want %v", i, btnY, expectedY)
				}
			}
		})
	}
}

func TestVirtualJoystickAllQuadrants(t *testing.T) {
	tests := []struct {
		name    string
		knobX   float64
		knobY   float64
		wantX   float64
		wantY   float64
		epsilon float64
	}{
		{"north_east", 140, 60, 0.5, -0.5, 0.2},
		{"north_west", 60, 60, -0.5, -0.5, 0.2},
		{"south_east", 140, 140, 0.5, 0.5, 0.2},
		{"south_west", 60, 140, -0.5, 0.5, 0.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vj := NewVirtualJoystick(100, 100, 80)
			vj.Active = true
			vj.KnobX = tt.knobX
			vj.KnobY = tt.knobY
			x, y := vj.GetAxis()
			if math.Abs(x-tt.wantX) > tt.epsilon {
				t.Errorf("GetAxis() x = %v, want %v (±%v)", x, tt.wantX, tt.epsilon)
			}
			if math.Abs(y-tt.wantY) > tt.epsilon {
				t.Errorf("GetAxis() y = %v, want %v (±%v)", y, tt.wantY, tt.epsilon)
			}
		})
	}
}

func TestTouchHapticFeedback(t *testing.T) {
	mgr := NewTouchHapticManager()
	if mgr == nil {
		t.Fatal("NewTouchHapticManager() returned nil")
	}

	// Test enabling/disabling
	mgr.SetEnabled(true)
	if !mgr.enabled {
		t.Error("enabled should be true after SetEnabled(true)")
	}
	mgr.SetEnabled(false)
	if mgr.enabled {
		t.Error("enabled should be false after SetEnabled(false)")
	}
}

func TestTouchHapticManagerEvents(t *testing.T) {
	tests := []struct {
		name        string
		event       HapticEvent
		wantPattern HapticPattern
	}{
		{"fire", HapticEventFire, HapticPatternShort},
		{"damage", HapticEventDamage, HapticPatternMedium},
		{"pickup", HapticEventPickup, HapticPatternLight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewTouchHapticManager()
			mgr.SetEnabled(true)
			mgr.TriggerEvent(tt.event)
			// Verify pattern was set (actual vibration would happen via ebiten.Vibrate)
			if mgr.lastPattern != tt.wantPattern {
				t.Errorf("lastPattern = %v, want %v", mgr.lastPattern, tt.wantPattern)
			}
		})
	}
}
