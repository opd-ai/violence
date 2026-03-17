// Package automap provides the in-game auto-mapping system with collapsible minimap functionality.
package automap

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sirupsen/logrus"
)

// MinimapState represents the current state of the collapsible minimap.
type MinimapState int

const (
	// StateExpanded shows the full minimap.
	StateExpanded MinimapState = iota
	// StateCompact shows a small minimap icon/indicator.
	StateCompact
	// StateHidden completely hides the minimap.
	StateHidden
)

// CollapsibleConfig holds configuration for the collapsible minimap behavior.
type CollapsibleConfig struct {
	// Expanded minimap dimensions
	ExpandedWidth  float32
	ExpandedHeight float32

	// Compact minimap dimensions (icon size when collapsed)
	CompactWidth  float32
	CompactHeight float32

	// Position (from top-right corner offset)
	MarginRight float32
	MarginTop   float32

	// Auto-hide behavior
	AutoHideEnabled bool    // Whether to auto-hide when idle
	AutoHideDelay   float64 // Seconds of no movement before auto-hide
	AutoShowOnMove  bool    // Show when player moves to new area

	// Animation
	TransitionSpeed float64 // How fast state transitions happen (0-1 per second)

	// Opacity settings
	ExpandedOpacity float32 // Opacity when expanded
	CompactOpacity  float32 // Opacity when compact
}

// DefaultCollapsibleConfig returns sensible defaults for the collapsible minimap.
func DefaultCollapsibleConfig() CollapsibleConfig {
	return CollapsibleConfig{
		ExpandedWidth:   200,
		ExpandedHeight:  200,
		CompactWidth:    50,
		CompactHeight:   50,
		MarginRight:     20,
		MarginTop:       20,
		AutoHideEnabled: true,
		AutoHideDelay:   5.0,
		AutoShowOnMove:  true,
		TransitionSpeed: 3.0,
		ExpandedOpacity: 0.85,
		CompactOpacity:  0.6,
	}
}

// CollapsibleMinimap wraps the standard Map with collapse/expand behavior.
type CollapsibleMinimap struct {
	baseMap *Map
	config  CollapsibleConfig
	logger  *logrus.Entry

	// Current state
	currentState MinimapState
	targetState  MinimapState

	// Animation progress (0 = compact, 1 = expanded)
	transition float64

	// Auto-hide tracking
	lastPlayerX     float64
	lastPlayerY     float64
	idleTime        float64
	lastRevealedX   int
	lastRevealedY   int
	newAreaRevealed bool

	// Screen dimensions cache
	screenWidth  int
	screenHeight int
}

// NewCollapsibleMinimap creates a collapsible wrapper around a base Map.
func NewCollapsibleMinimap(baseMap *Map, cfg CollapsibleConfig) *CollapsibleMinimap {
	return &CollapsibleMinimap{
		baseMap:      baseMap,
		config:       cfg,
		currentState: StateCompact, // Start compact by default
		targetState:  StateCompact,
		transition:   0.0,
		logger: logrus.WithFields(logrus.Fields{
			"system":  "automap",
			"feature": "collapsible",
		}),
	}
}

// SetState changes the target minimap state.
func (c *CollapsibleMinimap) SetState(state MinimapState) {
	if c.targetState != state {
		c.targetState = state
		c.logger.WithField("state", stateName(state)).Debug("minimap state change requested")
	}
}

// ToggleExpand toggles between expanded and compact states.
func (c *CollapsibleMinimap) ToggleExpand() {
	if c.targetState == StateExpanded {
		c.SetState(StateCompact)
	} else {
		c.SetState(StateExpanded)
	}
}

// GetState returns the current visual state.
func (c *CollapsibleMinimap) GetState() MinimapState {
	return c.currentState
}

// Update processes state transitions and auto-hide logic.
func (c *CollapsibleMinimap) Update(deltaTime, playerX, playerY float64) {
	// Track player movement for auto-hide
	if c.config.AutoHideEnabled {
		c.updateAutoHide(deltaTime, playerX, playerY)
	}

	// Process state transitions
	c.updateTransition(deltaTime)

	// Track revealed areas for auto-show
	if c.config.AutoShowOnMove && c.baseMap != nil {
		gridX, gridY := int(playerX), int(playerY)
		if gridX != c.lastRevealedX || gridY != c.lastRevealedY {
			if gridX >= 0 && gridX < c.baseMap.Width && gridY >= 0 && gridY < c.baseMap.Height {
				if !c.baseMap.Revealed[gridY][gridX] {
					c.newAreaRevealed = true
					// Flash to expanded briefly when entering new area
					if c.targetState == StateCompact || c.targetState == StateHidden {
						c.SetState(StateExpanded)
						c.idleTime = 0 // Reset idle timer to give player time to see the map
					}
				}
			}
			c.lastRevealedX = gridX
			c.lastRevealedY = gridY
		}
	}
}

// updateAutoHide handles automatic collapse after idle time.
func (c *CollapsibleMinimap) updateAutoHide(deltaTime, playerX, playerY float64) {
	const movementThreshold = 0.01

	dx := playerX - c.lastPlayerX
	dy := playerY - c.lastPlayerY
	moved := math.Abs(dx) > movementThreshold || math.Abs(dy) > movementThreshold

	if moved {
		c.idleTime = 0
		c.lastPlayerX = playerX
		c.lastPlayerY = playerY

		// Show compact at least when moving
		if c.targetState == StateHidden {
			c.SetState(StateCompact)
		}
	} else {
		c.idleTime += deltaTime

		// Auto-collapse to compact after delay
		if c.idleTime > c.config.AutoHideDelay && c.targetState == StateExpanded {
			c.SetState(StateCompact)
		}
	}
}

// updateTransition animates between states.
func (c *CollapsibleMinimap) updateTransition(deltaTime float64) {
	targetTransition := 0.0
	switch c.targetState {
	case StateExpanded:
		targetTransition = 1.0
	case StateCompact:
		targetTransition = 0.0
	case StateHidden:
		targetTransition = -0.5 // Below 0 for fade out
	}

	// Smooth transition using easing
	diff := targetTransition - c.transition
	if math.Abs(diff) < 0.001 {
		c.transition = targetTransition
		c.currentState = c.targetState
	} else {
		// Lerp towards target with clamped speed
		speed := c.config.TransitionSpeed * deltaTime
		if speed > 1.0 {
			speed = 1.0 // Cap max speed to prevent overshoot
		}
		c.transition += diff * speed * 3.0

		// Clamp transition to valid range
		if c.transition > 1.0 {
			c.transition = 1.0
		} else if c.transition < -0.5 {
			c.transition = -0.5
		}

		// Update current state based on transition progress
		if c.transition > 0.7 {
			c.currentState = StateExpanded
		} else if c.transition > -0.2 {
			c.currentState = StateCompact
		} else {
			c.currentState = StateHidden
		}
	}
}

// Render draws the collapsible minimap.
func (c *CollapsibleMinimap) Render(screen *ebiten.Image, cfg RenderConfig) {
	if c.baseMap == nil {
		return
	}

	bounds := screen.Bounds()
	c.screenWidth = bounds.Dx()
	c.screenHeight = bounds.Dy()

	// Skip rendering if fully hidden
	if c.currentState == StateHidden && c.transition <= -0.4 {
		return
	}

	// Calculate interpolated dimensions and position
	t := clampFloat64(c.transition, 0, 1)
	eased := easeOutCubic(t)

	width := lerp32(c.config.CompactWidth, c.config.ExpandedWidth, eased)
	height := lerp32(c.config.CompactHeight, c.config.ExpandedHeight, eased)
	opacity := lerp32(c.config.CompactOpacity, c.config.ExpandedOpacity, eased)

	// Handle hidden state fade
	if c.transition < 0 {
		hiddenFade := float32(1.0 + c.transition*2) // fade from 1 to 0 as transition goes -0.5 to 0
		if hiddenFade < 0 {
			hiddenFade = 0
		}
		opacity *= hiddenFade
	}

	x := float32(c.screenWidth) - width - c.config.MarginRight
	y := c.config.MarginTop

	// Prepare render config
	renderCfg := cfg
	renderCfg.X = x
	renderCfg.Y = y
	renderCfg.Width = width
	renderCfg.Height = height
	renderCfg.Opacity = opacity

	// Adjust cell size based on current size
	renderCfg.CellSize = cfg.CellSize * (width / c.config.ExpandedWidth)
	if renderCfg.CellSize < 1.5 {
		renderCfg.CellSize = 1.5 // Minimum readable cell size
	}

	// Render the minimap
	c.baseMap.RenderMinimap(screen, renderCfg)

	// Draw expand/collapse indicator when compact
	if c.currentState == StateCompact || (c.currentState == StateExpanded && c.transition < 0.9) {
		c.drawCollapseIndicator(screen, x, y, width, height, eased)
	}

	// Draw "new area" pulse effect
	if c.newAreaRevealed && c.currentState == StateExpanded {
		c.drawNewAreaPulse(screen, x, y, width, height)
		c.newAreaRevealed = false
	}
}

// drawCollapseIndicator renders a small indicator showing the minimap can be expanded.
func (c *CollapsibleMinimap) drawCollapseIndicator(screen *ebiten.Image, x, y, w, h float32, expandProgress float64) {
	// Draw expand icon in corner (small arrows pointing outward)
	iconSize := float32(8)
	iconX := x + w - iconSize - 4
	iconY := y + h - iconSize - 4

	// Fade icon as map expands
	iconAlpha := uint8(200 * (1.0 - expandProgress))
	if iconAlpha < 20 {
		return
	}

	iconColor := GetGenreTheme(GetCurrentGenre()).Player
	iconColor.A = iconAlpha

	// Draw expand arrows (top-left to bottom-right diagonal)
	// Top-left arrow
	vector.StrokeLine(screen, iconX, iconY+iconSize/2, iconX+iconSize/3, iconY+iconSize/3, 1.5, iconColor, false)
	vector.StrokeLine(screen, iconX+iconSize/3, iconY+iconSize/3, iconX+iconSize/2, iconY, 1.5, iconColor, false)

	// Bottom-right arrow
	vector.StrokeLine(screen, iconX+iconSize, iconY+iconSize/2, iconX+iconSize*2/3, iconY+iconSize*2/3, 1.5, iconColor, false)
	vector.StrokeLine(screen, iconX+iconSize*2/3, iconY+iconSize*2/3, iconX+iconSize/2, iconY+iconSize, 1.5, iconColor, false)
}

// drawNewAreaPulse draws a brief pulse effect when entering new map areas.
func (c *CollapsibleMinimap) drawNewAreaPulse(screen *ebiten.Image, x, y, w, h float32) {
	theme := GetGenreTheme(GetCurrentGenre())
	pulseColor := theme.Player
	pulseColor.A = 60

	// Draw glowing border
	vector.StrokeRect(screen, x-2, y-2, w+4, h+4, 2.5, pulseColor, false)
}

// GetBaseMap returns the underlying Map for direct manipulation.
func (c *CollapsibleMinimap) GetBaseMap() *Map {
	return c.baseMap
}

// SetBaseMap updates the underlying map (e.g., when changing levels).
func (c *CollapsibleMinimap) SetBaseMap(m *Map) {
	c.baseMap = m
	c.lastRevealedX = -1
	c.lastRevealedY = -1
	c.newAreaRevealed = false
}

// ForceExpand immediately expands the minimap (e.g., on key press).
func (c *CollapsibleMinimap) ForceExpand() {
	c.SetState(StateExpanded)
	c.idleTime = 0
}

// ForceCompact immediately collapses the minimap.
func (c *CollapsibleMinimap) ForceCompact() {
	c.SetState(StateCompact)
}

// IsExpanded returns true if the minimap is fully expanded.
func (c *CollapsibleMinimap) IsExpanded() bool {
	return c.currentState == StateExpanded && c.transition > 0.95
}

// Helper functions

func stateName(s MinimapState) string {
	switch s {
	case StateExpanded:
		return "expanded"
	case StateCompact:
		return "compact"
	case StateHidden:
		return "hidden"
	default:
		return "unknown"
	}
}

func lerp32(a, b float32, t float64) float32 {
	return a + (b-a)*float32(t)
}

func clampFloat64(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func easeOutCubic(t float64) float64 {
	t = t - 1
	return t*t*t + 1
}
