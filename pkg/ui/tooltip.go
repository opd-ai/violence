// Package ui provides a tooltip system with automatic screen-edge awareness.
//
// This system addresses the UI/UX problem: "Tooltips must never cover the element they describe."
// Tooltips automatically reposition to stay visible within screen bounds and avoid covering
// their target element.
package ui

import (
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font/basicfont"
)

// TooltipPosition defines where a tooltip appears relative to its target.
type TooltipPosition int

const (
	// TooltipAbove positions the tooltip above the target.
	TooltipAbove TooltipPosition = iota
	// TooltipBelow positions the tooltip below the target.
	TooltipBelow
	// TooltipLeft positions the tooltip to the left of the target.
	TooltipLeft
	// TooltipRight positions the tooltip to the right of the target.
	TooltipRight
	// TooltipAuto automatically chooses the best position.
	TooltipAuto
)

// TooltipConfig holds visual configuration for tooltips.
type TooltipConfig struct {
	// BackgroundColor is the fill color of the tooltip.
	BackgroundColor color.RGBA
	// BorderColor is the outline color.
	BorderColor color.RGBA
	// TextColor is the text color.
	TextColor color.RGBA
	// Padding is the space between text and border.
	Padding int
	// BorderWidth is the thickness of the border.
	BorderWidth float32
	// CornerRadius is the rounded corner radius.
	CornerRadius float32
	// ShowDelay is how long to hover before tooltip appears.
	ShowDelay time.Duration
	// FadeInDuration controls the fade-in animation time.
	FadeInDuration time.Duration
	// MaxWidth limits line length (wrapping at word boundaries).
	MaxWidth int
	// Offset is the distance from the target element.
	Offset int
}

// DefaultTooltipConfig returns sensible defaults for tooltip rendering.
func DefaultTooltipConfig() TooltipConfig {
	return TooltipConfig{
		BackgroundColor: color.RGBA{R: 30, G: 30, B: 35, A: 240},
		BorderColor:     color.RGBA{R: 80, G: 80, B: 90, A: 255},
		TextColor:       color.RGBA{R: 220, G: 220, B: 220, A: 255},
		Padding:         6,
		BorderWidth:     1.5,
		CornerRadius:    3.0,
		ShowDelay:       300 * time.Millisecond,
		FadeInDuration:  100 * time.Millisecond,
		MaxWidth:        200,
		Offset:          8,
	}
}

// Tooltip represents a single tooltip instance.
type Tooltip struct {
	// ID is a unique identifier for this tooltip.
	ID string
	// Text is the tooltip content (supports multiple lines with \n).
	Text string
	// TargetX, TargetY is the center of the element being described.
	TargetX, TargetY int
	// TargetW, TargetH is the size of the target element.
	TargetW, TargetH int
	// PreferredPosition is the preferred placement direction.
	PreferredPosition TooltipPosition
	// HoverStartTime is when the mouse started hovering.
	HoverStartTime time.Time
	// Visible indicates whether to render this tooltip.
	Visible bool
	// Opacity is the current fade-in progress (0.0-1.0).
	Opacity float64
	// FinalX, FinalY is the computed screen position (after repositioning).
	FinalX, FinalY int
	// FinalW, FinalH is the computed size.
	FinalW, FinalH int
}

// TooltipSystem manages tooltip display with screen-edge awareness.
type TooltipSystem struct {
	config       TooltipConfig
	tooltips     map[string]*Tooltip
	activeID     string
	screenWidth  int
	screenHeight int
	layoutMgr    *LayoutManager
	logger       *logrus.Entry
}

// NewTooltipSystem creates a tooltip management system.
func NewTooltipSystem(screenWidth, screenHeight int, config TooltipConfig) *TooltipSystem {
	return &TooltipSystem{
		config:       config,
		tooltips:     make(map[string]*Tooltip),
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		logger: logrus.WithFields(logrus.Fields{
			"system": "tooltip",
		}),
	}
}

// SetLayoutManager integrates with the UI layout manager for overlap prevention.
func (s *TooltipSystem) SetLayoutManager(lm *LayoutManager) {
	s.layoutMgr = lm
}

// SetScreenSize updates the screen bounds for repositioning calculations.
func (s *TooltipSystem) SetScreenSize(width, height int) {
	s.screenWidth = width
	s.screenHeight = height
}

// SetConfig updates the tooltip visual configuration.
func (s *TooltipSystem) SetConfig(config TooltipConfig) {
	s.config = config
}

// RegisterTooltip adds a new tooltip for an element.
func (s *TooltipSystem) RegisterTooltip(id, text string, targetX, targetY, targetW, targetH int, preferred TooltipPosition) {
	s.tooltips[id] = &Tooltip{
		ID:                id,
		Text:              text,
		TargetX:           targetX,
		TargetY:           targetY,
		TargetW:           targetW,
		TargetH:           targetH,
		PreferredPosition: preferred,
		Visible:           false,
		Opacity:           0.0,
	}
}

// UpdateTooltipTarget moves an existing tooltip's target.
func (s *TooltipSystem) UpdateTooltipTarget(id string, targetX, targetY, targetW, targetH int) {
	if tt, ok := s.tooltips[id]; ok {
		tt.TargetX = targetX
		tt.TargetY = targetY
		tt.TargetW = targetW
		tt.TargetH = targetH
	}
}

// UpdateTooltipText changes the tooltip content.
func (s *TooltipSystem) UpdateTooltipText(id, newText string) {
	if tt, ok := s.tooltips[id]; ok {
		tt.Text = newText
	}
}

// RemoveTooltip deletes a tooltip.
func (s *TooltipSystem) RemoveTooltip(id string) {
	delete(s.tooltips, id)
	if s.activeID == id {
		s.activeID = ""
	}
}

// OnHover should be called when the mouse hovers over a registered element.
func (s *TooltipSystem) OnHover(id string) {
	tt, ok := s.tooltips[id]
	if !ok {
		return
	}

	if s.activeID != id {
		// New hover target
		s.activeID = id
		tt.HoverStartTime = time.Now()
		tt.Visible = false
		tt.Opacity = 0.0
	}
}

// OnLeave should be called when the mouse leaves a registered element.
func (s *TooltipSystem) OnLeave(id string) {
	if tt, ok := s.tooltips[id]; ok {
		tt.Visible = false
		tt.Opacity = 0.0
	}
	if s.activeID == id {
		s.activeID = ""
	}
}

// Update processes tooltip timers and animations.
func (s *TooltipSystem) Update() {
	if s.activeID == "" {
		return
	}

	tt, ok := s.tooltips[s.activeID]
	if !ok {
		s.activeID = ""
		return
	}

	elapsed := time.Since(tt.HoverStartTime)

	// Check if delay has passed
	if elapsed >= s.config.ShowDelay {
		tt.Visible = true

		// Update fade-in opacity
		fadeElapsed := elapsed - s.config.ShowDelay
		if fadeElapsed < s.config.FadeInDuration {
			tt.Opacity = float64(fadeElapsed) / float64(s.config.FadeInDuration)
		} else {
			tt.Opacity = 1.0
		}

		// Compute position with screen-edge awareness
		s.computeTooltipPosition(tt)
	}
}

// computeTooltipPosition calculates the best screen position for a tooltip.
func (s *TooltipSystem) computeTooltipPosition(tt *Tooltip) {
	// Calculate text bounds
	lines := s.wrapText(tt.Text, s.config.MaxWidth)
	maxLineWidth := 0
	for _, line := range lines {
		bounds := text.BoundString(basicfont.Face7x13, line)
		if bounds.Dx() > maxLineWidth {
			maxLineWidth = bounds.Dx()
		}
	}

	lineHeight := 15 // basicfont.Face7x13 is approximately 13px + 2px spacing
	textHeight := len(lines) * lineHeight

	tt.FinalW = maxLineWidth + s.config.Padding*2
	tt.FinalH = textHeight + s.config.Padding*2

	// Determine position based on preference (or auto-select best)
	position := tt.PreferredPosition
	if position == TooltipAuto {
		position = s.selectBestPosition(tt)
	}

	// Calculate base position
	switch position {
	case TooltipAbove:
		tt.FinalX = tt.TargetX + tt.TargetW/2 - tt.FinalW/2
		tt.FinalY = tt.TargetY - tt.FinalH - s.config.Offset
	case TooltipBelow:
		tt.FinalX = tt.TargetX + tt.TargetW/2 - tt.FinalW/2
		tt.FinalY = tt.TargetY + tt.TargetH + s.config.Offset
	case TooltipLeft:
		tt.FinalX = tt.TargetX - tt.FinalW - s.config.Offset
		tt.FinalY = tt.TargetY + tt.TargetH/2 - tt.FinalH/2
	case TooltipRight:
		tt.FinalX = tt.TargetX + tt.TargetW + s.config.Offset
		tt.FinalY = tt.TargetY + tt.TargetH/2 - tt.FinalH/2
	}

	// Clamp to screen bounds
	tt.FinalX = s.clampX(tt.FinalX, tt.FinalW)
	tt.FinalY = s.clampY(tt.FinalY, tt.FinalH)

	// Verify tooltip doesn't cover target and reposition if needed
	s.ensureNoTargetCoverage(tt)
}

// selectBestPosition chooses the position with most available space.
func (s *TooltipSystem) selectBestPosition(tt *Tooltip) TooltipPosition {
	// Calculate available space in each direction
	spaceAbove := tt.TargetY
	spaceBelow := s.screenHeight - (tt.TargetY + tt.TargetH)
	spaceLeft := tt.TargetX
	spaceRight := s.screenWidth - (tt.TargetX + tt.TargetW)

	// Find the direction with most space
	maxSpace := spaceAbove
	bestPos := TooltipAbove

	if spaceBelow > maxSpace {
		maxSpace = spaceBelow
		bestPos = TooltipBelow
	}
	if spaceRight > maxSpace {
		maxSpace = spaceRight
		bestPos = TooltipRight
	}
	if spaceLeft > maxSpace {
		bestPos = TooltipLeft
	}

	return bestPos
}

// clampX clamps the X position to keep tooltip on screen.
func (s *TooltipSystem) clampX(x, width int) int {
	const margin = 4
	if x < margin {
		return margin
	}
	if x+width > s.screenWidth-margin {
		return s.screenWidth - width - margin
	}
	return x
}

// clampY clamps the Y position to keep tooltip on screen.
func (s *TooltipSystem) clampY(y, height int) int {
	const margin = 4
	if y < margin {
		return margin
	}
	if y+height > s.screenHeight-margin {
		return s.screenHeight - height - margin
	}
	return y
}

// ensureNoTargetCoverage verifies tooltip doesn't overlap its target and adjusts if needed.
func (s *TooltipSystem) ensureNoTargetCoverage(tt *Tooltip) {
	// Calculate overlap
	overlapX := s.rectOverlap(
		tt.FinalX, tt.FinalX+tt.FinalW,
		tt.TargetX, tt.TargetX+tt.TargetW,
	)
	overlapY := s.rectOverlap(
		tt.FinalY, tt.FinalY+tt.FinalH,
		tt.TargetY, tt.TargetY+tt.TargetH,
	)

	if overlapX > 0 && overlapY > 0 {
		// There's overlap - try to shift tooltip
		// Prefer shifting in the direction with most space
		shiftX := 0
		shiftY := 0

		if tt.FinalX+tt.FinalW/2 < tt.TargetX+tt.TargetW/2 {
			// Tooltip is mostly to the left, shift left
			shiftX = -overlapX - s.config.Offset
		} else {
			// Tooltip is mostly to the right, shift right
			shiftX = overlapX + s.config.Offset
		}

		if tt.FinalY+tt.FinalH/2 < tt.TargetY+tt.TargetH/2 {
			// Tooltip is mostly above, shift up
			shiftY = -overlapY - s.config.Offset
		} else {
			// Tooltip is mostly below, shift down
			shiftY = overlapY + s.config.Offset
		}

		// Apply smaller shift (prefer minimal movement)
		if math.Abs(float64(shiftX)) < math.Abs(float64(shiftY)) {
			tt.FinalX = s.clampX(tt.FinalX+shiftX, tt.FinalW)
		} else {
			tt.FinalY = s.clampY(tt.FinalY+shiftY, tt.FinalH)
		}
	}
}

// rectOverlap calculates the overlap between two 1D ranges.
func (s *TooltipSystem) rectOverlap(min1, max1, min2, max2 int) int {
	overlapStart := max(min1, min2)
	overlapEnd := min(max1, max2)
	if overlapStart < overlapEnd {
		return overlapEnd - overlapStart
	}
	return 0
}

// wrapText breaks text into lines that fit within maxWidth.
func (s *TooltipSystem) wrapText(content string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{content}
	}

	var lines []string
	var currentLine string

	words := splitWords(content)

	for _, word := range words {
		if word == "\n" {
			lines = append(lines, currentLine)
			currentLine = ""
			continue
		}

		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		bounds := text.BoundString(basicfont.Face7x13, testLine)
		if bounds.Dx() > maxWidth && currentLine != "" {
			// Start new line
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			currentLine = testLine
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	if len(lines) == 0 {
		lines = []string{""}
	}

	return lines
}

// splitWords splits text into words, preserving newlines as separate tokens.
func splitWords(s string) []string {
	var words []string
	var current string

	for _, r := range s {
		if r == '\n' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
			words = append(words, "\n")
		} else if r == ' ' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}

	if current != "" {
		words = append(words, current)
	}

	return words
}

// Render draws the active tooltip.
func (s *TooltipSystem) Render(screen *ebiten.Image) {
	if s.activeID == "" {
		return
	}

	tt, ok := s.tooltips[s.activeID]
	if !ok || !tt.Visible || tt.Opacity <= 0 {
		return
	}

	// Apply opacity
	alpha := uint8(tt.Opacity * 255)

	// Draw background
	bgColor := s.config.BackgroundColor
	bgColor.A = uint8(float64(bgColor.A) * tt.Opacity)

	x := float32(tt.FinalX)
	y := float32(tt.FinalY)
	w := float32(tt.FinalW)
	h := float32(tt.FinalH)

	// Draw rounded rectangle background
	s.drawRoundedRect(screen, x, y, w, h, s.config.CornerRadius, bgColor)

	// Draw border
	borderColor := s.config.BorderColor
	borderColor.A = uint8(float64(borderColor.A) * tt.Opacity)
	s.drawRoundedRectStroke(screen, x, y, w, h, s.config.CornerRadius, s.config.BorderWidth, borderColor)

	// Draw text
	textColor := s.config.TextColor
	textColor.A = alpha

	lines := s.wrapText(tt.Text, s.config.MaxWidth)
	lineHeight := 15
	textY := tt.FinalY + s.config.Padding + 12 // Baseline offset

	for _, line := range lines {
		text.Draw(screen, line, basicfont.Face7x13, tt.FinalX+s.config.Padding, textY, textColor)
		textY += lineHeight
	}
}

// drawRoundedRect draws a filled rounded rectangle.
func (s *TooltipSystem) drawRoundedRect(screen *ebiten.Image, x, y, w, h, r float32, clr color.RGBA) {
	if r <= 0 {
		vector.DrawFilledRect(screen, x, y, w, h, clr, false)
		return
	}

	// Draw the main body as a filled rectangle
	// Using simplified approach: draw inner rectangle and corner circles
	innerX := x + r
	innerY := y + r
	innerW := w - 2*r
	innerH := h - 2*r

	// Center rectangle
	vector.DrawFilledRect(screen, innerX, y, innerW, h, clr, false)
	// Left edge
	vector.DrawFilledRect(screen, x, innerY, r, innerH, clr, false)
	// Right edge
	vector.DrawFilledRect(screen, x+w-r, innerY, r, innerH, clr, false)

	// Corner circles
	vector.DrawFilledCircle(screen, innerX, innerY, r, clr, false)
	vector.DrawFilledCircle(screen, innerX+innerW, innerY, r, clr, false)
	vector.DrawFilledCircle(screen, innerX, innerY+innerH, r, clr, false)
	vector.DrawFilledCircle(screen, innerX+innerW, innerY+innerH, r, clr, false)
}

// drawRoundedRectStroke draws a stroked rounded rectangle outline.
func (s *TooltipSystem) drawRoundedRectStroke(screen *ebiten.Image, x, y, w, h, r, strokeWidth float32, clr color.RGBA) {
	if r <= 0 {
		vector.StrokeRect(screen, x, y, w, h, strokeWidth, clr, false)
		return
	}

	// Draw edges
	innerX := x + r
	innerY := y + r
	innerW := w - 2*r
	innerH := h - 2*r

	// Top edge
	vector.StrokeLine(screen, innerX, y, innerX+innerW, y, strokeWidth, clr, false)
	// Bottom edge
	vector.StrokeLine(screen, innerX, y+h, innerX+innerW, y+h, strokeWidth, clr, false)
	// Left edge
	vector.StrokeLine(screen, x, innerY, x, innerY+innerH, strokeWidth, clr, false)
	// Right edge
	vector.StrokeLine(screen, x+w, innerY, x+w, innerY+innerH, strokeWidth, clr, false)

	// Draw corner arcs (approximated as circles since StrokeArc doesn't exist)
	// For a clean look, we draw partial circles
	segments := 8
	for i := 0; i < segments; i++ {
		angle1 := float64(i) * math.Pi / 2 / float64(segments)
		angle2 := float64(i+1) * math.Pi / 2 / float64(segments)

		// Top-left corner (PI to 3PI/2)
		s.drawArcSegment(screen, innerX, innerY, r, math.Pi+angle1, math.Pi+angle2, strokeWidth, clr)
		// Top-right corner (3PI/2 to 2PI)
		s.drawArcSegment(screen, innerX+innerW, innerY, r, 3*math.Pi/2+angle1, 3*math.Pi/2+angle2, strokeWidth, clr)
		// Bottom-right corner (0 to PI/2)
		s.drawArcSegment(screen, innerX+innerW, innerY+innerH, r, angle1, angle2, strokeWidth, clr)
		// Bottom-left corner (PI/2 to PI)
		s.drawArcSegment(screen, innerX, innerY+innerH, r, math.Pi/2+angle1, math.Pi/2+angle2, strokeWidth, clr)
	}
}

// drawArcSegment draws a small line segment approximating an arc.
func (s *TooltipSystem) drawArcSegment(screen *ebiten.Image, cx, cy, r float32, angle1, angle2 float64, strokeWidth float32, clr color.RGBA) {
	x1 := cx + r*float32(math.Cos(angle1))
	y1 := cy + r*float32(math.Sin(angle1))
	x2 := cx + r*float32(math.Cos(angle2))
	y2 := cy + r*float32(math.Sin(angle2))

	vector.StrokeLine(screen, x1, y1, x2, y2, strokeWidth, clr, false)
}

// GetActiveTooltipID returns the currently active tooltip ID, if any.
func (s *TooltipSystem) GetActiveTooltipID() string {
	return s.activeID
}

// IsTooltipVisible returns whether a specific tooltip is currently visible.
func (s *TooltipSystem) IsTooltipVisible(id string) bool {
	if tt, ok := s.tooltips[id]; ok {
		return tt.Visible && tt.Opacity > 0
	}
	return false
}

// Clear removes all registered tooltips.
func (s *TooltipSystem) Clear() {
	s.tooltips = make(map[string]*Tooltip)
	s.activeID = ""
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
