package focusring

import (
	"image/color"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sirupsen/logrus"
)

// System manages focus ring rendering and keyboard navigation for UI elements.
type System struct {
	elements     []*FocusableElement
	elementMap   map[string]*FocusableElement
	state        FocusState
	config       FocusRingConfig
	presets      map[string]GenrePreset
	currentGenre string
	tabOrder     []*FocusableElement
	enabled      bool
}

// NewSystem creates a new focus ring system with default configuration.
func NewSystem() *System {
	return &System{
		elements:     make([]*FocusableElement, 0, 32),
		elementMap:   make(map[string]*FocusableElement),
		state:        FocusState{Visible: false},
		config:       DefaultConfig(),
		presets:      DefaultGenrePresets(),
		currentGenre: "fantasy",
		tabOrder:     make([]*FocusableElement, 0, 32),
		enabled:      true,
	}
}

// AddFocusable registers a UI element for keyboard focus navigation.
func (s *System) AddFocusable(elem *FocusableElement) {
	if elem == nil || elem.ID == "" {
		return
	}
	if elem.Enabled == false {
		// Default to enabled if not explicitly disabled
		elem.Enabled = true
	}
	s.elements = append(s.elements, elem)
	s.elementMap[elem.ID] = elem
	s.rebuildTabOrder()

	logrus.WithFields(logrus.Fields{
		"system":    "focusring",
		"element":   elem.ID,
		"tab_index": elem.TabIndex,
	}).Debug("Registered focusable element")
}

// RemoveFocusable unregisters a UI element from focus navigation.
func (s *System) RemoveFocusable(id string) {
	if _, exists := s.elementMap[id]; !exists {
		return
	}
	delete(s.elementMap, id)

	// Remove from slice
	for i, elem := range s.elements {
		if elem.ID == id {
			s.elements = append(s.elements[:i], s.elements[i+1:]...)
			break
		}
	}

	// Clear focus if this was the focused element
	if s.state.FocusedID == id {
		s.state.FocusedID = ""
		s.state.Visible = false
	}

	s.rebuildTabOrder()
}

// ClearFocusables removes all registered focusable elements.
func (s *System) ClearFocusables() {
	s.elements = s.elements[:0]
	s.elementMap = make(map[string]*FocusableElement)
	s.tabOrder = s.tabOrder[:0]
	s.state.FocusedID = ""
	s.state.Visible = false
}

// rebuildTabOrder sorts focusable elements by TabIndex.
func (s *System) rebuildTabOrder() {
	s.tabOrder = make([]*FocusableElement, 0, len(s.elements))
	for _, elem := range s.elements {
		if elem.Enabled {
			s.tabOrder = append(s.tabOrder, elem)
		}
	}
	sort.Slice(s.tabOrder, func(i, j int) bool {
		return s.tabOrder[i].TabIndex < s.tabOrder[j].TabIndex
	})
}

// SetFocus moves keyboard focus to the specified element ID.
func (s *System) SetFocus(id string) {
	elem, exists := s.elementMap[id]
	if !exists || !elem.Enabled {
		return
	}

	// Call blur on previous element
	if s.state.FocusedID != "" && s.state.FocusedID != id {
		if prev, ok := s.elementMap[s.state.FocusedID]; ok && prev.OnBlur != nil {
			prev.OnBlur()
		}
	}

	s.state.FocusedID = id
	s.state.TargetX = elem.X
	s.state.TargetY = elem.Y
	s.state.TargetW = elem.Width
	s.state.TargetH = elem.Height
	s.state.TransitionProgress = 0
	s.state.Visible = true

	// Initialize current position if this is first focus
	if s.state.CurrentW == 0 {
		s.state.CurrentX = elem.X
		s.state.CurrentY = elem.Y
		s.state.CurrentW = elem.Width
		s.state.CurrentH = elem.Height
		s.state.TransitionProgress = 1
	}

	// Call focus callback
	if elem.OnFocus != nil {
		elem.OnFocus()
	}

	logrus.WithFields(logrus.Fields{
		"system":  "focusring",
		"element": id,
	}).Debug("Focus changed")
}

// ClearFocus removes focus from all elements.
func (s *System) ClearFocus() {
	if s.state.FocusedID != "" {
		if prev, ok := s.elementMap[s.state.FocusedID]; ok && prev.OnBlur != nil {
			prev.OnBlur()
		}
	}
	s.state.FocusedID = ""
	s.state.Visible = false
}

// GetFocusedID returns the ID of the currently focused element.
func (s *System) GetFocusedID() string {
	return s.state.FocusedID
}

// SetGenre updates the focus ring visual style for the specified genre.
func (s *System) SetGenre(genreID string) {
	preset, exists := s.presets[genreID]
	if !exists {
		preset = s.presets["fantasy"]
	}
	s.currentGenre = genreID
	s.config.RingColor = preset.RingColor
	s.config.GlowColor = preset.GlowColor
	s.config.PulseSpeed = preset.PulseSpeed
	s.config.PulseIntensity = preset.PulseIntensity

	logrus.WithFields(logrus.Fields{
		"system": "focusring",
		"genre":  genreID,
	}).Debug("Genre updated")
}

// SetEnabled enables or disables the focus ring system.
func (s *System) SetEnabled(enabled bool) {
	s.enabled = enabled
	if !enabled {
		s.state.Visible = false
	}
}

// Update processes keyboard input and advances animations.
func (s *System) Update() {
	if !s.enabled || len(s.tabOrder) == 0 {
		return
	}

	// Process keyboard navigation
	s.processKeyboardInput()

	// Advance pulse animation
	s.state.PulsePhase += s.config.PulseSpeed
	if s.state.PulsePhase > 2*math.Pi {
		s.state.PulsePhase -= 2 * math.Pi
	}

	// Advance position transition
	if s.state.TransitionProgress < 1.0 {
		s.state.TransitionProgress += s.config.TransitionSpeed
		if s.state.TransitionProgress > 1.0 {
			s.state.TransitionProgress = 1.0
		}

		// Smooth interpolation using ease-out cubic
		t := easeOutCubic(s.state.TransitionProgress)
		s.state.CurrentX = lerp(s.state.CurrentX, s.state.TargetX, float32(t))
		s.state.CurrentY = lerp(s.state.CurrentY, s.state.TargetY, float32(t))
		s.state.CurrentW = lerp(s.state.CurrentW, s.state.TargetW, float32(t))
		s.state.CurrentH = lerp(s.state.CurrentH, s.state.TargetH, float32(t))
	}

	// Sync with focused element position (in case it moved)
	if s.state.FocusedID != "" {
		if elem, ok := s.elementMap[s.state.FocusedID]; ok {
			if s.state.TransitionProgress >= 1.0 {
				s.state.CurrentX = elem.X
				s.state.CurrentY = elem.Y
				s.state.CurrentW = elem.Width
				s.state.CurrentH = elem.Height
			}
			s.state.TargetX = elem.X
			s.state.TargetY = elem.Y
			s.state.TargetW = elem.Width
			s.state.TargetH = elem.Height
		}
	}
}

// processKeyboardInput handles Tab, Arrow keys, Enter, and Escape.
func (s *System) processKeyboardInput() {
	// Tab / Shift+Tab for linear navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			s.focusPrevious()
		} else {
			s.focusNext()
		}
	}

	// Arrow keys for spatial navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		s.focusSpatial(0, -1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		s.focusSpatial(0, 1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		s.focusSpatial(-1, 0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		s.focusSpatial(1, 0)
	}

	// Enter/Space to activate focused element
	if s.state.FocusedID != "" {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			if elem, ok := s.elementMap[s.state.FocusedID]; ok && elem.OnActivate != nil {
				elem.OnActivate()
			}
		}
	}

	// Escape clears focus
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.ClearFocus()
	}
}

// focusNext moves focus to the next element in tab order.
func (s *System) focusNext() {
	if len(s.tabOrder) == 0 {
		return
	}

	// If nothing focused, focus first element
	if s.state.FocusedID == "" {
		s.SetFocus(s.tabOrder[0].ID)
		return
	}

	// Find current index and move to next
	for i, elem := range s.tabOrder {
		if elem.ID == s.state.FocusedID {
			nextIdx := (i + 1) % len(s.tabOrder)
			s.SetFocus(s.tabOrder[nextIdx].ID)
			return
		}
	}

	// Fallback to first element
	s.SetFocus(s.tabOrder[0].ID)
}

// focusPrevious moves focus to the previous element in tab order.
func (s *System) focusPrevious() {
	if len(s.tabOrder) == 0 {
		return
	}

	// If nothing focused, focus last element
	if s.state.FocusedID == "" {
		s.SetFocus(s.tabOrder[len(s.tabOrder)-1].ID)
		return
	}

	// Find current index and move to previous
	for i, elem := range s.tabOrder {
		if elem.ID == s.state.FocusedID {
			prevIdx := i - 1
			if prevIdx < 0 {
				prevIdx = len(s.tabOrder) - 1
			}
			s.SetFocus(s.tabOrder[prevIdx].ID)
			return
		}
	}

	// Fallback to last element
	s.SetFocus(s.tabOrder[len(s.tabOrder)-1].ID)
}

// focusSpatial moves focus in the specified direction based on element positions.
func (s *System) focusSpatial(dx, dy int) {
	if s.state.FocusedID == "" || len(s.elements) == 0 {
		if len(s.tabOrder) > 0 {
			s.SetFocus(s.tabOrder[0].ID)
		}
		return
	}

	current, ok := s.elementMap[s.state.FocusedID]
	if !ok {
		return
	}

	// Find center of current element
	cx := current.X + current.Width/2
	cy := current.Y + current.Height/2

	var best *FocusableElement
	bestScore := float32(math.MaxFloat32)

	for _, elem := range s.elements {
		if elem.ID == current.ID || !elem.Enabled {
			continue
		}

		// Calculate center of candidate
		ex := elem.X + elem.Width/2
		ey := elem.Y + elem.Height/2

		// Check if candidate is in the correct direction
		deltaX := ex - cx
		deltaY := ey - cy

		inDirection := false
		if dx > 0 && deltaX > 0 {
			inDirection = true
		} else if dx < 0 && deltaX < 0 {
			inDirection = true
		} else if dy > 0 && deltaY > 0 {
			inDirection = true
		} else if dy < 0 && deltaY < 0 {
			inDirection = true
		}

		if !inDirection {
			continue
		}

		// Score: prefer elements more aligned in the primary direction
		var score float32
		if dx != 0 {
			// Moving horizontally: penalize vertical distance
			score = absf(deltaX) + absf(deltaY)*3
		} else {
			// Moving vertically: penalize horizontal distance
			score = absf(deltaY) + absf(deltaX)*3
		}

		if score < bestScore {
			bestScore = score
			best = elem
		}
	}

	if best != nil {
		s.SetFocus(best.ID)
	}
}

// Draw renders the focus ring around the currently focused element.
func (s *System) Draw(screen *ebiten.Image) {
	if !s.enabled || !s.state.Visible || s.state.FocusedID == "" {
		return
	}

	// Calculate pulse intensity
	pulseMultiplier := 1.0 + s.config.PulseIntensity*math.Sin(s.state.PulsePhase)

	// Draw outer glow layers (multiple passes for soft gradient)
	s.drawGlowLayers(screen, pulseMultiplier)

	// Draw inner ring stroke
	s.drawRingStroke(screen, pulseMultiplier)
}

// drawGlowLayers renders multiple semi-transparent layers for soft glow effect.
func (s *System) drawGlowLayers(screen *ebiten.Image, pulseMultiplier float64) {
	// Padding around element
	padding := s.config.GlowRadius

	// Multiple layers with decreasing opacity for soft falloff
	layers := 3
	for i := 0; i < layers; i++ {
		layerPadding := padding * float32(i+1) / float32(layers)
		opacity := uint8(float64(s.config.GlowColor.A) * pulseMultiplier * float64(layers-i) / float64(layers))

		glowColor := color.RGBA{
			R: s.config.GlowColor.R,
			G: s.config.GlowColor.G,
			B: s.config.GlowColor.B,
			A: opacity,
		}

		x := s.state.CurrentX - layerPadding
		y := s.state.CurrentY - layerPadding
		w := s.state.CurrentW + layerPadding*2
		h := s.state.CurrentH + layerPadding*2

		// Draw rounded rectangle for glow layer
		s.strokeRoundedRect(screen, x, y, w, h, s.config.CornerRadius+layerPadding, 2.0, glowColor)
	}
}

// drawRingStroke renders the main focus ring outline.
func (s *System) drawRingStroke(screen *ebiten.Image, pulseMultiplier float64) {
	// Inner ring with pulsing brightness
	brightness := uint8(math.Min(255, float64(s.config.RingColor.A)*pulseMultiplier))
	ringColor := color.RGBA{
		R: s.config.RingColor.R,
		G: s.config.RingColor.G,
		B: s.config.RingColor.B,
		A: brightness,
	}

	padding := float32(2.0) // Small padding from element edge
	x := s.state.CurrentX - padding
	y := s.state.CurrentY - padding
	w := s.state.CurrentW + padding*2
	h := s.state.CurrentH + padding*2

	s.strokeRoundedRect(screen, x, y, w, h, s.config.CornerRadius, s.config.RingThickness, ringColor)
}

// strokeRoundedRect draws a rounded rectangle outline.
func (s *System) strokeRoundedRect(screen *ebiten.Image, x, y, w, h, radius, thickness float32, clr color.RGBA) {
	if radius <= 0 {
		// Simple rectangle
		vector.StrokeRect(screen, x, y, w, h, thickness, clr, false)
		return
	}

	// Clamp radius to half of smallest dimension
	maxRadius := minf(w, h) / 2
	if radius > maxRadius {
		radius = maxRadius
	}

	// Draw four corner arcs and four edges
	// Top edge
	vector.StrokeLine(screen, x+radius, y, x+w-radius, y, thickness, clr, false)
	// Bottom edge
	vector.StrokeLine(screen, x+radius, y+h, x+w-radius, y+h, thickness, clr, false)
	// Left edge
	vector.StrokeLine(screen, x, y+radius, x, y+h-radius, thickness, clr, false)
	// Right edge
	vector.StrokeLine(screen, x+w, y+radius, x+w, y+h-radius, thickness, clr, false)

	// Draw corner arcs using short line segments
	segments := 8
	for i := 0; i < segments; i++ {
		a1 := float64(i) * (math.Pi / 2) / float64(segments)
		a2 := float64(i+1) * (math.Pi / 2) / float64(segments)

		// Top-left corner
		x1, y1 := x+radius-radius*float32(math.Cos(a1)), y+radius-radius*float32(math.Sin(a1))
		x2, y2 := x+radius-radius*float32(math.Cos(a2)), y+radius-radius*float32(math.Sin(a2))
		vector.StrokeLine(screen, x1, y1, x2, y2, thickness, clr, false)

		// Top-right corner
		x1, y1 = x+w-radius+radius*float32(math.Sin(a1)), y+radius-radius*float32(math.Cos(a1))
		x2, y2 = x+w-radius+radius*float32(math.Sin(a2)), y+radius-radius*float32(math.Cos(a2))
		vector.StrokeLine(screen, x1, y1, x2, y2, thickness, clr, false)

		// Bottom-right corner
		x1, y1 = x+w-radius+radius*float32(math.Cos(a1)), y+h-radius+radius*float32(math.Sin(a1))
		x2, y2 = x+w-radius+radius*float32(math.Cos(a2)), y+h-radius+radius*float32(math.Sin(a2))
		vector.StrokeLine(screen, x1, y1, x2, y2, thickness, clr, false)

		// Bottom-left corner
		x1, y1 = x+radius-radius*float32(math.Sin(a1)), y+h-radius+radius*float32(math.Cos(a1))
		x2, y2 = x+radius-radius*float32(math.Sin(a2)), y+h-radius+radius*float32(math.Cos(a2))
		vector.StrokeLine(screen, x1, y1, x2, y2, thickness, clr, false)
	}
}

// UpdateElementPosition updates the position of a registered focusable element.
func (s *System) UpdateElementPosition(id string, x, y, width, height float32) {
	elem, exists := s.elementMap[id]
	if !exists {
		return
	}
	elem.X = x
	elem.Y = y
	elem.Width = width
	elem.Height = height

	// Update target if this is the focused element
	if s.state.FocusedID == id {
		s.state.TargetX = x
		s.state.TargetY = y
		s.state.TargetW = width
		s.state.TargetH = height
	}
}

// SetElementEnabled enables or disables a focusable element.
func (s *System) SetElementEnabled(id string, enabled bool) {
	elem, exists := s.elementMap[id]
	if !exists {
		return
	}
	elem.Enabled = enabled
	s.rebuildTabOrder()

	// Clear focus if this element was disabled while focused
	if !enabled && s.state.FocusedID == id {
		s.focusNext()
	}
}

// Helper functions

func easeOutCubic(t float64) float64 {
	t = t - 1
	return t*t*t + 1
}

func lerp(a, b, t float32) float32 {
	return a + (b-a)*t
}

func absf(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func minf(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
