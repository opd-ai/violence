package toast

import (
	"image/color"
	"math"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font/basicfont"

	"github.com/opd-ai/violence/pkg/common"
	"github.com/opd-ai/violence/pkg/engine"
)

const (
	defaultDuration     = 3.0  // seconds
	maxQueueSize        = 20   // max pending notifications
	maxVisible          = 4    // max visible at once
	toastHeight         = 24.0 // pixels
	toastPadding        = 4.0  // vertical spacing
	toastMargin         = 8.0  // screen edge margin
	toastWidth          = 180  // max width
	enterDuration       = 0.25 // seconds
	exitDuration        = 0.3  // seconds
	iconSize            = 16.0 // icon dimensions
	slideDistance       = 200  // pixels to slide from
	pulseFrequency      = 4.0  // Hz for pulse animation
	pulseAmplitude      = 0.05 // scale variation
	criticalFlashPeriod = 0.5  // seconds
)

// System manages toast notifications with queue and animation.
type System struct {
	queue      []*Notification
	active     []*Notification
	genre      string
	nextID     uint64
	logger     *logrus.Entry
	screenW    int
	screenH    int
	position   Position // screen corner
	iconImages map[NotificationType]*ebiten.Image
}

// Position defines where toasts appear on screen.
type Position int

const (
	// PositionTopRight places toasts in top-right corner.
	PositionTopRight Position = iota
	// PositionTopLeft places toasts in top-left corner.
	PositionTopLeft
	// PositionBottomRight places toasts in bottom-right corner.
	PositionBottomRight
	// PositionBottomLeft places toasts in bottom-left corner.
	PositionBottomLeft
)

// NewSystem creates a toast notification system.
func NewSystem(genreID string) *System {
	s := &System{
		queue:    make([]*Notification, 0, maxQueueSize),
		active:   make([]*Notification, 0, maxVisible),
		genre:    genreID,
		nextID:   1,
		position: PositionTopRight, // Default position
		screenW:  320,
		screenH:  200,
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "toast",
			"package":     "toast",
		}),
		iconImages: make(map[NotificationType]*ebiten.Image),
	}
	s.generateIcons()
	return s
}

// SetGenre configures genre-specific styling.
func (s *System) SetGenre(genreID string) {
	s.genre = genreID
	s.generateIcons()
}

// SetPosition sets where toasts appear on screen.
func (s *System) SetPosition(pos Position) {
	s.position = pos
}

// SetScreenSize updates the screen dimensions for positioning.
func (s *System) SetScreenSize(width, height int) {
	s.screenW = width
	s.screenH = height
}

// Update processes the notification queue and animations.
func (s *System) Update(w *engine.World) {
	dt := common.DeltaTime

	// Update active notifications
	s.updateActive(dt)

	// Promote from queue if we have room
	s.promoteFromQueue()
}

// UpdateWithDelta processes notifications with a specified delta time (for testing).
func (s *System) UpdateWithDelta(w *engine.World, dt float64) {
	// Update active notifications
	s.updateActive(dt)

	// Promote from queue if we have room
	s.promoteFromQueue()
}

// updateActive processes animation states for active notifications.
func (s *System) updateActive(dt float64) {
	remaining := make([]*Notification, 0, len(s.active))

	for _, n := range s.active {
		n.Elapsed += dt
		n.StateTime += dt

		switch n.State {
		case StateEntering:
			s.updateEntering(n)
		case StateVisible:
			s.updateVisible(n, dt)
		case StateExiting:
			s.updateExiting(n)
		}

		if n.State != StateRemoved {
			remaining = append(remaining, n)
		}
	}

	s.active = remaining
	s.repositionActive()
}

// updateEntering handles slide-in animation.
func (s *System) updateEntering(n *Notification) {
	progress := n.GetProgress()
	eased := easeOutCubic(progress)

	// Slide from off-screen
	n.ScreenX = n.TargetX + slideDistance*(1.0-eased)
	if s.position == PositionTopLeft || s.position == PositionBottomLeft {
		n.ScreenX = n.TargetX - slideDistance*(1.0-eased)
	}

	n.Alpha = eased

	if progress >= 1.0 {
		n.State = StateVisible
		n.StateTime = 0
		n.ScreenX = n.TargetX
		n.Alpha = 1.0
	}
}

// updateVisible handles the main display state.
func (s *System) updateVisible(n *Notification, dt float64) {
	n.ScreenX = n.TargetX
	n.Alpha = 1.0

	// Pulse effect for critical
	if n.Priority == PriorityCritical {
		n.Scale = 1.0 + pulseAmplitude*math.Sin(n.StateTime*pulseFrequency*2*math.Pi)
	} else {
		n.Scale = 1.0
	}

	// Check if display time has elapsed
	if n.Elapsed >= n.Duration {
		n.State = StateExiting
		n.StateTime = 0
	}
}

// updateExiting handles fade-out animation.
func (s *System) updateExiting(n *Notification) {
	progress := n.GetProgress()
	eased := easeOutCubic(progress)

	n.Alpha = 1.0 - eased

	// Slide out
	if s.position == PositionTopLeft || s.position == PositionBottomLeft {
		n.ScreenX = n.TargetX - slideDistance*eased
	} else {
		n.ScreenX = n.TargetX + slideDistance*eased
	}

	if progress >= 1.0 {
		n.State = StateRemoved
	}
}

// repositionActive updates Y positions for stacking.
func (s *System) repositionActive() {
	y := toastMargin
	if s.position == PositionBottomLeft || s.position == PositionBottomRight {
		y = float64(s.screenH) - toastMargin - toastHeight
	}

	for i, n := range s.active {
		n.TargetY = y
		if n.State == StateEntering || n.State == StateVisible {
			n.ScreenY = y
		}

		if s.position == PositionBottomLeft || s.position == PositionBottomRight {
			y -= toastHeight + toastPadding
		} else {
			y += toastHeight + toastPadding
		}

		// Limit visible count
		if i >= maxVisible-1 {
			break
		}
	}
}

// promoteFromQueue moves notifications from queue to active.
func (s *System) promoteFromQueue() {
	for len(s.active) < maxVisible && len(s.queue) > 0 {
		// Pop highest priority from queue
		idx := s.findHighestPriority()
		n := s.queue[idx]
		s.queue = append(s.queue[:idx], s.queue[idx+1:]...)

		// Initialize animation
		n.State = StateEntering
		n.StateTime = 0
		n.Alpha = 0

		// Set target position
		x := float64(s.screenW) - toastWidth - toastMargin
		if s.position == PositionTopLeft || s.position == PositionBottomLeft {
			x = toastMargin
		}
		n.TargetX = x

		s.active = append(s.active, n)
	}
}

// findHighestPriority returns the index of the highest priority notification in queue.
func (s *System) findHighestPriority() int {
	best := 0
	for i, n := range s.queue {
		if n.Priority > s.queue[best].Priority {
			best = i
		}
	}
	return best
}

// Queue adds a notification to the queue.
func (s *System) Queue(ntype NotificationType, message string, priority Priority) {
	if len(s.queue) >= maxQueueSize {
		// Discard lowest priority oldest notification
		s.discardLowest()
	}

	colors := s.getColors(ntype, priority)

	n := &Notification{
		ID:          atomic.AddUint64(&s.nextID, 1),
		Type:        ntype,
		Message:     message,
		Priority:    priority,
		Duration:    s.getDuration(priority),
		Elapsed:     0,
		State:       StateEntering,
		Scale:       1.0,
		IconColor:   colors.icon,
		TextColor:   colors.text,
		BorderColor: colors.border,
		BGColor:     colors.bg,
	}

	s.queue = append(s.queue, n)

	s.logger.WithFields(logrus.Fields{
		"type":     ntype,
		"priority": priority,
		"message":  message,
	}).Debug("Queued toast notification")
}

type colorSet struct {
	icon, text, border, bg color.RGBA
}

// getColors returns genre-appropriate colors for a notification.
func (s *System) getColors(ntype NotificationType, priority Priority) colorSet {
	cs := colorSet{
		text:   color.RGBA{R: 255, G: 255, B: 255, A: 255},
		border: color.RGBA{R: 100, G: 100, B: 100, A: 200},
		bg:     color.RGBA{R: 20, G: 20, B: 30, A: 220},
	}

	// Type-specific icon colors
	switch ntype {
	case TypeLevelUp:
		cs.icon = color.RGBA{R: 255, G: 215, B: 0, A: 255} // Gold
		cs.border = color.RGBA{R: 200, G: 180, B: 50, A: 200}
	case TypeAchievement:
		cs.icon = color.RGBA{R: 180, G: 130, B: 255, A: 255} // Purple
		cs.border = color.RGBA{R: 150, G: 100, B: 200, A: 200}
	case TypeQuest:
		cs.icon = color.RGBA{R: 100, G: 200, B: 255, A: 255} // Blue
		cs.border = color.RGBA{R: 80, G: 150, B: 200, A: 200}
	case TypeLoot:
		cs.icon = color.RGBA{R: 100, G: 255, B: 100, A: 255} // Green
		cs.border = color.RGBA{R: 60, G: 180, B: 60, A: 200}
	case TypeSkill:
		cs.icon = color.RGBA{R: 0, G: 200, B: 255, A: 255} // Cyan
	case TypeCurrency:
		cs.icon = color.RGBA{R: 255, G: 200, B: 50, A: 255} // Gold coin
	case TypeDeath:
		cs.icon = color.RGBA{R: 200, G: 50, B: 50, A: 255} // Red
		cs.border = color.RGBA{R: 180, G: 30, B: 30, A: 200}
	case TypeWarning:
		cs.icon = color.RGBA{R: 255, G: 150, B: 0, A: 255} // Orange
		cs.border = color.RGBA{R: 200, G: 120, B: 0, A: 200}
	case TypeItem:
		cs.icon = color.RGBA{R: 200, G: 200, B: 200, A: 255} // Silver
	default:
		cs.icon = color.RGBA{R: 150, G: 150, B: 150, A: 255} // Gray
	}

	// Genre adjustments
	switch s.genre {
	case "cyberpunk":
		cs.bg = color.RGBA{R: 10, G: 20, B: 30, A: 230}
		if ntype == TypeLevelUp {
			cs.icon = color.RGBA{R: 0, G: 255, B: 255, A: 255}
		}
	case "horror":
		cs.bg = color.RGBA{R: 30, G: 15, B: 15, A: 230}
		cs.text = color.RGBA{R: 200, G: 180, B: 180, A: 255}
	case "scifi":
		cs.bg = color.RGBA{R: 15, G: 20, B: 35, A: 230}
		if ntype == TypeLevelUp {
			cs.icon = color.RGBA{R: 100, G: 200, B: 255, A: 255}
		}
	case "postapoc":
		cs.bg = color.RGBA{R: 35, G: 30, B: 20, A: 230}
		cs.text = color.RGBA{R: 230, G: 220, B: 180, A: 255}
	}

	// Priority adjustments
	if priority == PriorityCritical {
		cs.border.R = min(cs.border.R+50, 255)
		cs.border.A = 255
	}

	return cs
}

// getDuration returns display duration based on priority.
func (s *System) getDuration(priority Priority) float64 {
	switch priority {
	case PriorityLow:
		return 2.0
	case PriorityNormal:
		return 3.0
	case PriorityHigh:
		return 4.0
	case PriorityCritical:
		return 5.0
	default:
		return defaultDuration
	}
}

// discardLowest removes the lowest priority oldest notification from queue.
func (s *System) discardLowest() {
	if len(s.queue) == 0 {
		return
	}

	// Find lowest priority
	lowestIdx := 0
	for i, n := range s.queue {
		if n.Priority < s.queue[lowestIdx].Priority {
			lowestIdx = i
		}
	}

	s.queue = append(s.queue[:lowestIdx], s.queue[lowestIdx+1:]...)
}

// Render draws active notifications to screen.
func (s *System) Render(screen *ebiten.Image) {
	for _, n := range s.active {
		if !n.IsVisible() {
			continue
		}
		s.renderNotification(screen, n)
	}
}

// renderNotification draws a single toast notification.
func (s *System) renderNotification(screen *ebiten.Image, n *Notification) {
	x := float32(n.ScreenX)
	y := float32(n.ScreenY)
	w := float32(toastWidth)
	h := float32(toastHeight)
	alpha := n.GetAlpha()

	// Apply scale for critical pulse
	if n.Scale != 1.0 {
		scaleDiff := float32((n.Scale - 1.0) * toastWidth / 2)
		x -= scaleDiff
		w *= float32(n.Scale)
		h *= float32(n.Scale)
	}

	// Background with alpha
	bgColor := n.BGColor
	bgColor.A = uint8(float64(bgColor.A) * float64(alpha) / 255.0)
	vector.DrawFilledRect(screen, x, y, w, h, bgColor, false)

	// Border
	borderColor := n.BorderColor
	borderColor.A = uint8(float64(borderColor.A) * float64(alpha) / 255.0)
	vector.StrokeRect(screen, x, y, w, h, 1, borderColor, false)

	// Icon
	iconX := x + 4
	iconY := y + (h-iconSize)/2
	iconImg := s.iconImages[n.Type]
	if iconImg != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(iconX), float64(iconY))
		opts.ColorScale.ScaleAlpha(float32(alpha) / 255.0)
		// Tint icon with type color
		r := float32(n.IconColor.R) / 255.0
		g := float32(n.IconColor.G) / 255.0
		b := float32(n.IconColor.B) / 255.0
		opts.ColorScale.Scale(r, g, b, 1.0)
		screen.DrawImage(iconImg, opts)
	}

	// Text
	textX := int(iconX + iconSize + 4)
	textY := int(y + h/2 + 4)
	textColor := n.TextColor
	textColor.A = alpha

	// Truncate message to fit
	msg := n.Message
	maxChars := int((w - iconSize - 12) / 7) // ~7px per char
	if len(msg) > maxChars && maxChars > 3 {
		msg = msg[:maxChars-3] + "..."
	}

	text.Draw(screen, msg, basicfont.Face7x13, textX, textY, textColor)
}

// generateIcons creates procedural icons for each notification type.
func (s *System) generateIcons() {
	size := int(iconSize)

	// Item icon (box/chest shape)
	s.iconImages[TypeItem] = s.genIconItem(size)

	// Level up icon (arrow up)
	s.iconImages[TypeLevelUp] = s.genIconLevelUp(size)

	// Achievement icon (star)
	s.iconImages[TypeAchievement] = s.genIconStar(size)

	// Quest icon (exclamation)
	s.iconImages[TypeQuest] = s.genIconExclamation(size)

	// Loot icon (gem)
	s.iconImages[TypeLoot] = s.genIconGem(size)

	// Skill icon (lightning bolt)
	s.iconImages[TypeSkill] = s.genIconBolt(size)

	// Currency icon (coin)
	s.iconImages[TypeCurrency] = s.genIconCoin(size)

	// Death icon (skull)
	s.iconImages[TypeDeath] = s.genIconSkull(size)

	// Warning icon (triangle)
	s.iconImages[TypeWarning] = s.genIconWarning(size)

	// Info icon (circle i)
	s.iconImages[TypeInfo] = s.genIconInfo(size)
}

// Icon generation helpers - procedural pixel art icons.

func (s *System) genIconItem(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	// Simple box shape
	for x := 2; x < size-2; x++ {
		img.Set(x, 4, white)      // top
		img.Set(x, size-4, white) // bottom
	}
	for y := 4; y < size-4; y++ {
		img.Set(2, y, white)      // left
		img.Set(size-3, y, white) // right
	}
	// Lid
	img.Set(size/2, 3, white)
	img.Set(size/2-1, 3, white)
	img.Set(size/2+1, 3, white)
	return img
}

func (s *System) genIconLevelUp(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	mid := size / 2
	// Upward arrow
	for i := 0; i < size/2; i++ {
		img.Set(mid-i, size/2-2+i, white)
		img.Set(mid+i, size/2-2+i, white)
	}
	// Stem
	for y := size/2 - 1; y < size-3; y++ {
		img.Set(mid, y, white)
	}
	return img
}

func (s *System) genIconStar(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	mid := size / 2
	// 5-point star (simplified)
	img.Set(mid, 2, white)
	img.Set(mid-1, 4, white)
	img.Set(mid+1, 4, white)
	img.Set(mid-3, 5, white)
	img.Set(mid+3, 5, white)
	for x := mid - 2; x <= mid+2; x++ {
		img.Set(x, 6, white)
	}
	img.Set(mid-2, 7, white)
	img.Set(mid+2, 7, white)
	img.Set(mid-3, 9, white)
	img.Set(mid+3, 9, white)
	img.Set(mid, 8, white)
	return img
}

func (s *System) genIconExclamation(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	mid := size / 2
	// Exclamation mark
	for y := 3; y < size-5; y++ {
		img.Set(mid, y, white)
	}
	img.Set(mid, size-4, white)
	return img
}

func (s *System) genIconGem(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	mid := size / 2
	// Diamond shape
	for i := 0; i < 4; i++ {
		img.Set(mid-i, 3+i, white)
		img.Set(mid+i, 3+i, white)
		img.Set(mid-i, size-4-i, white)
		img.Set(mid+i, size-4-i, white)
	}
	return img
}

func (s *System) genIconBolt(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	// Lightning bolt
	img.Set(size/2+2, 2, white)
	img.Set(size/2+1, 3, white)
	img.Set(size/2, 4, white)
	img.Set(size/2, 5, white)
	img.Set(size/2+1, 5, white)
	img.Set(size/2+2, 5, white)
	img.Set(size/2, 6, white)
	img.Set(size/2-1, 7, white)
	img.Set(size/2-1, 8, white)
	img.Set(size/2-2, 9, white)
	return img
}

func (s *System) genIconCoin(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	mid := size / 2
	// Circle
	for x := mid - 3; x <= mid+3; x++ {
		img.Set(x, 3, white)
		img.Set(x, size-4, white)
	}
	for y := 4; y < size-4; y++ {
		img.Set(mid-4, y, white)
		img.Set(mid+4, y, white)
	}
	// $ symbol
	img.Set(mid, 4, white)
	img.Set(mid, size-5, white)
	img.Set(mid-1, 5, white)
	img.Set(mid+1, 6, white)
	img.Set(mid-1, 7, white)
	img.Set(mid+1, 8, white)
	return img
}

func (s *System) genIconSkull(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	mid := size / 2
	// Skull outline
	for x := mid - 3; x <= mid+3; x++ {
		img.Set(x, 2, white)
	}
	for y := 3; y < 8; y++ {
		img.Set(mid-4, y, white)
		img.Set(mid+4, y, white)
	}
	// Eyes
	img.Set(mid-2, 5, white)
	img.Set(mid+2, 5, white)
	// Jaw
	for x := mid - 2; x <= mid+2; x++ {
		img.Set(x, 9, white)
	}
	img.Set(mid-3, 8, white)
	img.Set(mid+3, 8, white)
	return img
}

func (s *System) genIconWarning(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	mid := size / 2
	// Triangle
	img.Set(mid, 2, white)
	for i := 0; i < size/2; i++ {
		img.Set(mid-i, size-4, white)
		img.Set(mid+i, size-4, white)
	}
	for y := 3; y < size-4; y++ {
		offset := (y - 2) / 2
		img.Set(mid-offset, y, white)
		img.Set(mid+offset, y, white)
	}
	// Exclamation
	img.Set(mid, 5, white)
	img.Set(mid, 6, white)
	img.Set(mid, 8, white)
	return img
}

func (s *System) genIconInfo(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	mid := size / 2
	// Circle
	for x := mid - 3; x <= mid+3; x++ {
		img.Set(x, 2, white)
		img.Set(x, size-3, white)
	}
	for y := 3; y < size-3; y++ {
		img.Set(mid-4, y, white)
		img.Set(mid+4, y, white)
	}
	// i
	img.Set(mid, 4, white)
	for y := 6; y < size-4; y++ {
		img.Set(mid, y, white)
	}
	return img
}

// easeOutCubic provides smooth deceleration.
func easeOutCubic(t float64) float64 {
	t = t - 1
	return t*t*t + 1
}

// GetActiveCount returns the number of visible notifications.
func (s *System) GetActiveCount() int {
	return len(s.active)
}

// GetQueueCount returns the number of pending notifications.
func (s *System) GetQueueCount() int {
	return len(s.queue)
}

// Clear removes all notifications.
func (s *System) Clear() {
	s.queue = s.queue[:0]
	s.active = s.active[:0]
}

// Helper function to spawn toast notifications from anywhere.

// Queue adds a notification to the world's toast system.
// This is a convenience function for spawning toasts.
func Queue(w *engine.World, ntype NotificationType, message string, priority Priority) {
	// The actual queuing is handled by the system's Queue method
	// This function would need access to the system instance
	// For now, this serves as documentation of the intended API
}
