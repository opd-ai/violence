# Mobile Touch Controls Design

## Overview

Violence supports mobile devices (iOS, Android) through Ebitengine's touch input API. This document defines the virtual control layout optimized for touchscreen FPS gameplay while maintaining the game's retro aesthetic.

## Design Philosophy

- **Minimal Screen Clutter**: Controls fade when not in use
- **Thumb Zone Optimization**: Primary controls within comfortable reach
- **Visual Clarity**: Semi-transparent overlays don't obscure gameplay
- **Accessibility**: Adjustable opacity, size, and position
- **Genre Consistency**: Control aesthetics match genre themes

## Control Layout

### Screen Regions

```
┌─────────────────────────────────────────────┐
│  [HP] [Armor]          [Weapon]  [Ammo]     │ ← Status bar (top 10%)
│                                              │
│                                     ┌────┐  │
│                                     │Fire│  │ ← Fire button (right 20%)
│            Gameplay Area            │ RT │  │
│          (center 60%)               └────┘  │
│                                     ┌────┐  │
│                                     │Alt │  │ ← Alt fire (right 20%)
│  ┌────┐                             │ LT │  │
│  │ ⊕  │                             └────┘  │
│  │Move│                                      │ ← Movement stick (left 20%)
│  └────┘                                      │
│ [Map][Inv][Pause]  [Jump][Reload][Interact] │ ← Action bar (bottom 10%)
└─────────────────────────────────────────────┘
```

### Dimensions (percentage of screen)

- **Movement Joystick**: 15% width, left edge + 5% margin, bottom edge + 15% margin
- **Look Area**: Center 60% (full height) - swipe to look around
- **Fire Button**: 12% width, right edge + 5% margin, 40% from top
- **Alt Fire Button**: 12% width, right edge + 5% margin, 55% from top
- **Action Bar**: Full width, bottom 8%, 6 evenly spaced buttons

## Virtual Controls

### Movement Joystick (Left Side)

**Type**: Floating virtual joystick

**Behavior**:
- **Touch Down**: Joystick appears at touch location (within left 25% of screen)
- **Drag**: Player moves in direction of drag, speed based on distance from center
- **Release**: Player stops, joystick fades out after 0.5s
- **Dead Zone**: 10% radius (no movement)
- **Max Speed**: 100% at 80% stick radius

**Visual**:
- Outer ring: 120x120px, 40% opacity, genre-themed color
- Inner knob: 60x60px, 80% opacity, contrasting color
- Directional indicators: Faint cardinal direction marks

**Code Mapping**:
```go
// pkg/input/touch.go
type VirtualJoystick struct {
    CenterX, CenterY float64
    KnobX, KnobY     float64
    MaxRadius        float64
    DeadZone         float64
    Active           bool
}

func (vj *VirtualJoystick) GetAxis() (x, y float64)
```

### Look Control (Center Area)

**Type**: Direct touch-to-look

**Behavior**:
- **Touch Down**: Lock current touch position as reference
- **Drag**: Camera rotates based on delta from reference point
- **Sensitivity**: Configurable (default: 0.15 degrees per pixel)
- **Vertical Clamping**: ±60° pitch limit
- **Multi-touch**: Only first touch controls look (others ignored)

**Visual**: No overlay (transparent area)

**Code Mapping**:
```go
type TouchLookController struct {
    RefX, RefY       int
    Sensitivity      float64
    YawDelta         float64
    PitchDelta       float64
    TouchID          ebiten.TouchID
}

func (tlc *TouchLookController) Update(touches []ebiten.Touch)
```

### Fire Button (Right Side, Upper)

**Type**: Tap-to-shoot button

**Behavior**:
- **Tap**: Single shot (semi-auto weapons)
- **Hold**: Continuous fire (auto weapons)
- **Visual Feedback**: Glow pulse on press, crosshair flash on shot

**Visual**:
- Circle: 100x100px, 50% opacity
- Icon: "RT" text or trigger symbol
- Press effect: Scale 1.0 → 0.9, opacity 50% → 80%

**Genre Variants**:
- Fantasy: Rune circle, orange glow
- SciFi: Hexagon, cyan glow
- Horror: Jagged circle, red pulse
- Cyberpunk: Neon octagon, magenta glow
- Postapoc: Scrap metal circle, yellow spark

### Alt Fire Button (Right Side, Lower)

**Type**: Tap-to-activate button

**Behavior**:
- **Tap**: Trigger alternate weapon mode (scope, grenade, melee bash)
- **Cooldown**: Visual indicator shows reload/cooldown time

**Visual**: Same as Fire button but with "LT" label or alt icon

### Action Bar (Bottom)

Six buttons spanning bottom 8% of screen:

1. **Automap Toggle** (`[Map]`)
   - Tap: Show/hide minimap overlay
   - Icon: Map symbol

2. **Inventory** (`[Inv]`)
   - Tap: Open inventory screen (pauses game)
   - Icon: Backpack

3. **Pause Menu** (`[Pause]`)
   - Tap: Open pause menu
   - Icon: Three horizontal lines

4. **Jump** (`[Jump]`)
   - Tap: Jump (if enabled in genre/mode)
   - Icon: Upward arrow

5. **Reload** (`[Reload]`)
   - Tap: Manual reload
   - Icon: Circular arrows

6. **Interact** (`[Interact]`)
   - Tap: Use object, open door, pick up item
   - Icon: Hand symbol

**Visual**: 
- Each button: 60x60px, 40% opacity, rounded rectangle
- Active state: Opacity → 80%, scale → 1.1
- Disabled state: Opacity → 20%, grayscale

## Advanced Features

### Customization Settings

**Settings Menu** (`Settings > Controls > Touch`):

```
[x] Enable Touch Controls
[ ] Show Movement Joystick Always
Joystick Size:     [====|----] (50%-150%)
Button Size:       [======|--] (75%-125%)
Button Opacity:    [====|----] (30%-90%)
Look Sensitivity:  [===|-----] (0.05-0.50)
[x] Haptic Feedback
[ ] Show Touch Debug Overlay
```

**Preset Layouts**:
- Default: As described above
- Left-Handed: Mirror layout (fire buttons on left, joystick on right)
- Claw Grip: Compact buttons for index finger use
- Tablet: Larger spacing, bigger buttons

### Haptic Feedback

**Events with vibration**:
- Fire weapon: 20ms pulse (intensity based on weapon)
- Take damage: 100ms medium pulse
- Pickup item: 15ms light pulse
- Door open: 30ms light pulse
- Low health: 500ms repeating gentle pulse

**Implementation**:
```go
// pkg/input/haptic.go
func TriggerHaptic(pattern HapticPattern) {
    if !Settings.TouchControls.HapticEnabled {
        return
    }
    ebiten.Vibrate(pattern.Duration, pattern.Intensity)
}
```

### Multi-Touch Gestures

**Two-Finger Gestures**:
- **Pinch**: Zoom in/out (if scope weapon equipped)
- **Two-Finger Swipe Up**: Quick weapon switch to next
- **Two-Finger Swipe Down**: Quick weapon switch to previous
- **Two-Finger Tap**: Quick use healthpack

**Three-Finger Gestures**:
- **Three-Finger Swipe Down**: Screenshot
- **Three-Finger Tap**: Toggle HUD visibility

### Orientation Support

**Portrait Mode**:
- Not recommended but functional
- Joystick moves to bottom-left corner
- Fire buttons stack vertically on right
- Look area reduced to 40% width center
- Action bar uses 2 rows of 3 buttons

**Landscape Mode** (default):
- As described in main layout

**Auto-Rotation**:
- Controls reposition smoothly during orientation change
- Game pauses briefly (0.2s) to allow repositioning

## Platform-Specific Considerations

### iOS

- **Safe Area Insets**: Controls avoid notch and home indicator
- **Touch ID/Face ID**: Pause game when authentication prompt appears
- **App Switching**: Save game state on background transition

### Android

- **Navigation Bar**: Bottom margin adjusted for gesture navigation or on-screen buttons
- **Notch/Cutout**: Query cutout area and avoid placing controls
- **Back Button**: Maps to Pause Menu (overrides system back)

### Tablets

- **Larger Screens**: Scale controls proportionally but cap at 150% phone size
- **Split-Screen**: Reduce control size if app window <600dp wide
- **Stylus Support**: Treat as precise touch (no palm rejection needed)

## Accessibility Options

### Visual Accessibility

- **High Contrast Mode**: Increase button opacity to 90%, use bold outlines
- **Color Blind Modes**: Adjust genre color schemes (protanopia, deuteranopia, tritanopia)
- **Large Touch Targets**: Increase button size to 125% by default

### Motor Accessibility

- **Sticky Fire**: Toggle fire button (tap to start shooting, tap to stop)
- **Auto-Fire**: Weapon fires automatically when enemy in crosshair
- **Single-Handed Mode**: All controls on one side (left or right), swipe to toggle modes
- **External Controller Support**: Bluetooth gamepad (see CONTROLS.md)

### Audio Accessibility

- **Haptic Feedback Intensity**: Adjust vibration strength (or disable)
- **Visual Fire Indicator**: Screen edge flash when firing (for deaf players)

## Testing and Validation

### Device Test Matrix

| Device Category | Example Devices | Screen Size | Priority |
|----------------|----------------|-------------|----------|
| Small Phone | iPhone SE, Galaxy A | 4.7"-5.5" | High |
| Standard Phone | iPhone 14, Pixel 7 | 6.0"-6.5" | Critical |
| Large Phone | iPhone 14 Pro Max, Galaxy S23 Ultra | 6.7"-7.0" | High |
| Small Tablet | iPad Mini, Galaxy Tab A | 7.9"-8.4" | Medium |
| Large Tablet | iPad Pro, Galaxy Tab S | 10"-13" | Medium |

### Usability Metrics

**Target Performance**:
- **Touch Latency**: <50ms from touch to action
- **Joystick Response**: <16ms (1 frame at 60 FPS)
- **Look Sensitivity Range**: Accommodate 95th percentile users
- **Button Miss Rate**: <5% for primary actions (fire, move)
- **Session Duration**: Comfortable gameplay for 30+ minutes

### Playtesting Checklist

- [ ] Movement joystick responds to touch in all left-quadrant positions
- [ ] Look control allows 360° horizontal rotation and ±60° vertical
- [ ] Fire button detects tap vs hold correctly
- [ ] Multi-touch works (move + look + fire simultaneously)
- [ ] Action bar buttons are reachable without repositioning grip
- [ ] Controls don't obscure critical UI (health, ammo, crosshair)
- [ ] Haptic feedback intensity feels appropriate
- [ ] Custom layout saves and persists across sessions
- [ ] Orientation change doesn't break control positions
- [ ] Works on smallest supported screen size (4.7")

## Implementation Roadmap

### Phase 1: Core Controls (v1.0)
- Virtual joystick with dead zone
- Touch-to-look camera control
- Fire and alt-fire buttons
- Basic action bar (pause, interact)

### Phase 2: Polish (v1.1)
- Customization settings (size, opacity, sensitivity)
- Haptic feedback
- Genre-themed button styles

### Phase 3: Advanced (v1.2)
- Multi-touch gestures
- Accessibility options
- Tablet optimizations

### Phase 4: Refinement (v1.3)
- Platform-specific safe area handling
- External controller integration
- User feedback iteration

## Code Structure

```
pkg/input/
├── touch.go              // Main touch input handling
├── virtual_joystick.go   // Joystick implementation
├── touch_button.go       // Button widget
├── haptic.go             // Vibration/haptic feedback
├── touch_renderer.go     // Draw virtual controls
└── touch_test.go         // Unit tests

pkg/ui/mobile/
├── layout.go             // Control positioning
├── customization.go      // Settings UI
└── presets.go            // Layout presets
```

### Sample Implementation

```go
// pkg/input/touch.go
package input

import "github.com/hajimehoshi/ebiten/v2"

type TouchInputManager struct {
    Joystick     *VirtualJoystick
    LookControl  *TouchLookController
    FireButton   *TouchButton
    AltButton    *TouchButton
    ActionBar    []*TouchButton
    Settings     *TouchSettings
}

func NewTouchInputManager() *TouchInputManager {
    return &TouchInputManager{
        Joystick:    NewVirtualJoystick(0.15, 0.85, 0.1),
        LookControl: NewTouchLookController(0.15),
        FireButton:  NewTouchButton(0.85, 0.4, "Fire"),
        AltButton:   NewTouchButton(0.85, 0.55, "Alt"),
        ActionBar:   createActionBar(),
        Settings:    LoadTouchSettings(),
    }
}

func (tim *TouchInputManager) Update() {
    touches := ebiten.AppendTouchIDs(nil)
    
    for _, id := range touches {
        x, y := ebiten.TouchPosition(id)
        
        // Route touch to appropriate control
        if tim.Joystick.HandleTouch(id, x, y) {
            continue
        }
        if tim.FireButton.HandleTouch(id, x, y) {
            continue
        }
        if tim.AltButton.HandleTouch(id, x, y) {
            continue
        }
        // Default: Look control
        tim.LookControl.HandleTouch(id, x, y)
    }
}

func (tim *TouchInputManager) Draw(screen *ebiten.Image) {
    tim.Joystick.Draw(screen)
    tim.FireButton.Draw(screen)
    tim.AltButton.Draw(screen)
    for _, btn := range tim.ActionBar {
        btn.Draw(screen)
    }
}
```

## Future Enhancements

- **Cloud Save Sync**: Save custom layouts to cloud (iCloud, Google Drive)
- **Gyroscope Aiming**: Tilt device to fine-tune aim (optional)
- **Adaptive Layout**: ML-based layout optimization from usage patterns
- **Tutorial Overlay**: First-launch interactive tutorial
- **Touch Recording**: Record and replay touch input for bug reports

## Conclusion

This mobile control scheme balances accessibility with precision FPS gameplay. The virtual joystick + touch-to-look combination has proven effective in titles like PUBG Mobile and Call of Duty Mobile. Customization ensures players can adapt to their preferred grip styles, while haptic feedback provides essential kinesthetic feedback on touchscreens.

**Next Steps**:
1. Implement `pkg/input/touch.go` stub with basic structure
2. Create prototype UI in Figma for visual validation
3. Conduct initial usability testing on 3-5 devices
4. Iterate based on feedback
5. Full implementation targeting v1.0 mobile release
