// Package ui provides a spatial reservation system to prevent UI element overlap.
package ui

import (
	"math"
	"sort"

	"github.com/sirupsen/logrus"
)

// Priority levels for UI elements (higher = more important, less likely to move)
type Priority int

const (
	PriorityAmbient   Priority = 1 // Background info, distant entity status
	PrioritySecondary Priority = 2 // Mid-range entity info
	PriorityImportant Priority = 3 // Nearby entities, tooltips
	PriorityCritical  Priority = 4 // Player vitals, active threats
)

// Rect represents a screen-space rectangle.
type Rect struct {
	X      float32
	Y      float32
	Width  float32
	Height float32
}

// Contains checks if this rect fully contains another.
func (r Rect) Contains(other Rect) bool {
	return r.X <= other.X &&
		r.Y <= other.Y &&
		r.X+r.Width >= other.X+other.Width &&
		r.Y+r.Height >= other.Y+other.Height
}

// Overlaps checks if this rect overlaps another.
func (r Rect) Overlaps(other Rect) bool {
	return r.X < other.X+other.Width &&
		r.X+r.Width > other.X &&
		r.Y < other.Y+other.Height &&
		r.Y+r.Height > other.Y
}

// Center returns the center point of the rect.
func (r Rect) Center() (float32, float32) {
	return r.X + r.Width/2, r.Y + r.Height/2
}

// UIElement represents a positioned UI element with priority.
type UIElement struct {
	ID       string
	Bounds   Rect
	Priority Priority
	CanMove  bool // If false, this element is fixed and others must avoid it
}

// LayoutManager manages spatial reservation to prevent UI overlap.
type LayoutManager struct {
	screenWidth  int
	screenHeight int
	elements     []*UIElement
	logger       *logrus.Entry
	margin       float32 // Minimum spacing between elements
}

// NewLayoutManager creates a UI layout manager.
func NewLayoutManager(screenWidth, screenHeight int) *LayoutManager {
	return &LayoutManager{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		elements:     make([]*UIElement, 0, 100),
		margin:       4.0,
		logger: logrus.WithFields(logrus.Fields{
			"system":  "layout",
			"package": "ui",
		}),
	}
}

// SetScreenSize updates the screen dimensions.
func (lm *LayoutManager) SetScreenSize(width, height int) {
	lm.screenWidth = width
	lm.screenHeight = height
}

// Clear removes all reserved elements (call at start of each frame).
func (lm *LayoutManager) Clear() {
	lm.elements = lm.elements[:0]
}

// Reserve attempts to reserve screen space for a UI element.
// Returns adjusted position if moved to avoid overlap, or original if no conflict.
func (lm *LayoutManager) Reserve(id string, x, y, width, height float32, priority Priority, canMove bool) (float32, float32, bool) {
	// Clamp to screen bounds if movable
	if canMove {
		if x < 0 {
			x = 0
		}
		if y < 0 {
			y = 0
		}
		if x+width > float32(lm.screenWidth) {
			x = float32(lm.screenWidth) - width
		}
		if y+height > float32(lm.screenHeight) {
			y = float32(lm.screenHeight) - height
		}
	}

	elem := &UIElement{
		ID: id,
		Bounds: Rect{
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
		},
		Priority: priority,
		CanMove:  canMove,
	}

	// Check for overlaps with existing elements
	conflicting := lm.findConflicts(elem)

	if len(conflicting) == 0 {
		// No conflicts, reserve at requested position
		lm.elements = append(lm.elements, elem)
		return x, y, true
	}

	// Handle conflicts based on priority
	if !canMove {
		// Fixed element: force others to move
		lm.elements = append(lm.elements, elem)
		return x, y, true
	}

	// Find if we should yield to existing elements
	shouldYield := false
	for _, conflict := range conflicting {
		if conflict.Priority >= priority || !conflict.CanMove {
			shouldYield = true
			break
		}
	}

	if !shouldYield {
		// Higher priority: remove lower priority elements
		lm.removeConflicts(conflicting)
		lm.elements = append(lm.elements, elem)
		return x, y, true
	}

	// Lower priority: try to find alternative position
	newX, newY, found := lm.findAlternativePosition(elem, conflicting)
	if found {
		elem.Bounds.X = newX
		elem.Bounds.Y = newY
		lm.elements = append(lm.elements, elem)
		return newX, newY, true
	}

	// No space found: don't render this element
	lm.logger.WithFields(logrus.Fields{
		"element": id,
		"x":       x,
		"y":       y,
	}).Debug("UI element could not find space - suppressing")
	return x, y, false
}

// ReserveDamageNumber reserves space for a damage number with stacking.
// Returns adjusted Y position to stack above other damage numbers at same location.
func (lm *LayoutManager) ReserveDamageNumber(id string, x, y, width, height float32, priority Priority) (float32, float32, bool) {
	// Find damage numbers at similar X position
	const xTolerance = 30.0
	minY := y // Track the highest Y position we need to avoid

	for _, elem := range lm.elements {
		// Check if this element is at a similar X position
		if math.Abs(float64(elem.Bounds.X-x)) < xTolerance {
			// Check if this element overlaps our vertical range
			elemBottom := elem.Bounds.Y + elem.Bounds.Height
			proposedTop := minY

			// If the existing element would overlap with our current position
			if elem.Bounds.Y <= proposedTop && elemBottom+lm.margin >= proposedTop {
				// Move above this element
				minY = elem.Bounds.Y - height - lm.margin
			}
		}
	}

	adjustedY := minY

	// Clamp to screen bounds
	if adjustedY < 0 {
		adjustedY = 0
	}
	if adjustedY+height > float32(lm.screenHeight) {
		adjustedY = float32(lm.screenHeight) - height
	}

	elem := &UIElement{
		ID: id,
		Bounds: Rect{
			X:      x,
			Y:      adjustedY,
			Width:  width,
			Height: height,
		},
		Priority: priority,
		CanMove:  false,
	}

	lm.elements = append(lm.elements, elem)
	return x, adjustedY, true
}

// findConflicts returns elements that overlap with the given element.
func (lm *LayoutManager) findConflicts(elem *UIElement) []*UIElement {
	conflicts := make([]*UIElement, 0)

	// Expand bounds by margin
	checkBounds := Rect{
		X:      elem.Bounds.X - lm.margin,
		Y:      elem.Bounds.Y - lm.margin,
		Width:  elem.Bounds.Width + 2*lm.margin,
		Height: elem.Bounds.Height + 2*lm.margin,
	}

	for _, existing := range lm.elements {
		if checkBounds.Overlaps(existing.Bounds) {
			conflicts = append(conflicts, existing)
		}
	}

	return conflicts
}

// removeConflicts removes conflicting elements from reservation list.
func (lm *LayoutManager) removeConflicts(conflicts []*UIElement) {
	// Build set of conflicts for fast lookup
	conflictSet := make(map[*UIElement]bool)
	for _, c := range conflicts {
		conflictSet[c] = true
	}

	// Filter elements
	filtered := make([]*UIElement, 0, len(lm.elements))
	for _, elem := range lm.elements {
		if !conflictSet[elem] {
			filtered = append(filtered, elem)
		}
	}
	lm.elements = filtered
}

// findAlternativePosition attempts to find a nearby position that doesn't conflict.
func (lm *LayoutManager) findAlternativePosition(elem *UIElement, conflicts []*UIElement) (float32, float32, bool) {
	// Try offsets in expanding spiral pattern
	offsets := []struct{ dx, dy float32 }{
		{0, -20},
		{0, 20},
		{20, 0},
		{-20, 0},
		{20, -20},
		{-20, -20},
		{20, 20},
		{-20, 20},
		{0, -40},
		{0, 40},
		{40, 0},
		{-40, 0},
	}

	for _, offset := range offsets {
		testX := elem.Bounds.X + offset.dx
		testY := elem.Bounds.Y + offset.dy

		// Check screen bounds
		if testX < 0 || testY < 0 ||
			testX+elem.Bounds.Width > float32(lm.screenWidth) ||
			testY+elem.Bounds.Height > float32(lm.screenHeight) {
			continue
		}

		testBounds := Rect{
			X:      testX - lm.margin,
			Y:      testY - lm.margin,
			Width:  elem.Bounds.Width + 2*lm.margin,
			Height: elem.Bounds.Height + 2*lm.margin,
		}

		// Check conflicts at new position
		hasConflict := false
		for _, existing := range lm.elements {
			if testBounds.Overlaps(existing.Bounds) {
				hasConflict = true
				break
			}
		}

		if !hasConflict {
			return testX, testY, true
		}
	}

	return elem.Bounds.X, elem.Bounds.Y, false
}

// GetReservedBounds returns all currently reserved UI bounds (for debugging).
func (lm *LayoutManager) GetReservedBounds() []Rect {
	bounds := make([]Rect, len(lm.elements))
	for i, elem := range lm.elements {
		bounds[i] = elem.Bounds
	}
	return bounds
}

// SortElementsByPriority sorts elements by priority (higher priority first).
func (lm *LayoutManager) SortElementsByPriority() {
	sort.Slice(lm.elements, func(i, j int) bool {
		return lm.elements[i].Priority > lm.elements[j].Priority
	})
}

// GetElementCount returns the number of currently reserved elements.
func (lm *LayoutManager) GetElementCount() int {
	return len(lm.elements)
}
