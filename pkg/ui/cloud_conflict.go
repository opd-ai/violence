// Package ui provides cloud save conflict resolution UI.
package ui

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/input"
	"github.com/opd-ai/violence/pkg/save/cloud"
	"golang.org/x/image/font/basicfont"
)

// ConflictOption represents a user choice for conflict resolution.
type ConflictOption int

const (
	// ConflictKeepLocal keeps the local save version.
	ConflictKeepLocal ConflictOption = iota
	// ConflictKeepCloud keeps the cloud save version.
	ConflictKeepCloud
	// ConflictKeepBoth keeps both versions in separate slots.
	ConflictKeepBoth
	// ConflictCancel cancels the sync operation.
	ConflictCancel
)

// ConflictDialog manages cloud save conflict resolution UI.
type ConflictDialog struct {
	visible       bool
	localMeta     cloud.SaveMetadata
	cloudMeta     cloud.SaveMetadata
	selectedIndex int
	options       []string
	onResolve     func(ConflictOption) error
}

// NewConflictDialog creates a new conflict resolution dialog.
func NewConflictDialog() *ConflictDialog {
	return &ConflictDialog{
		visible: false,
		options: []string{
			"Keep Local Save",
			"Keep Cloud Save",
			"Keep Both (New Slot)",
			"Cancel",
		},
		selectedIndex: 0,
	}
}

// Show displays the conflict dialog with metadata.
func (c *ConflictDialog) Show(local, cloudData cloud.SaveMetadata, onResolve func(ConflictOption) error) {
	c.visible = true
	c.localMeta = local
	c.cloudMeta = cloudData
	c.selectedIndex = 0
	c.onResolve = onResolve
}

// Hide closes the conflict dialog.
func (c *ConflictDialog) Hide() {
	c.visible = false
	c.onResolve = nil
}

// IsVisible returns whether the dialog is shown.
func (c *ConflictDialog) IsVisible() bool {
	return c.visible
}

// Update handles user input for conflict resolution.
func (c *ConflictDialog) Update(mgr *input.Manager) error {
	if !c.visible {
		return nil
	}

	if mgr.IsJustPressed(input.ActionMoveForward) {
		c.navigateUp()
	}
	if mgr.IsJustPressed(input.ActionMoveBackward) {
		c.navigateDown()
	}

	if mgr.IsJustPressed(input.ActionInteract) || mgr.IsJustPressed(input.ActionFire) {
		return c.handleSelection()
	}

	return nil
}

// navigateUp moves selection to previous option.
func (c *ConflictDialog) navigateUp() {
	if !c.visible {
		return
	}
	c.selectedIndex = (c.selectedIndex - 1 + len(c.options)) % len(c.options)
}

// navigateDown moves selection to next option.
func (c *ConflictDialog) navigateDown() {
	if !c.visible {
		return
	}
	c.selectedIndex = (c.selectedIndex + 1) % len(c.options)
}

// handleSelection processes the user's conflict resolution choice.
func (c *ConflictDialog) handleSelection() error {
	if c.onResolve == nil {
		c.Hide()
		return nil
	}

	option := ConflictOption(c.selectedIndex)
	err := c.onResolve(option)
	if err != nil {
		return err
	}

	c.Hide()
	return nil
}

// Draw renders the conflict dialog.
func (c *ConflictDialog) Draw(screen *ebiten.Image) {
	if !c.visible {
		return
	}

	bounds := screen.Bounds()
	width, height := float32(bounds.Dx()), float32(bounds.Dy())

	bgColor := color.RGBA{0, 0, 0, 200}
	vector.DrawFilledRect(screen, 0, 0, width, height, bgColor, false)

	dialogW, dialogH := float32(500), float32(400)
	dialogX := (width - dialogW) / 2
	dialogY := (height - dialogH) / 2

	panelColor := color.RGBA{40, 40, 40, 255}
	borderColor := color.RGBA{100, 100, 100, 255}
	vector.DrawFilledRect(screen, dialogX, dialogY, dialogW, dialogH, panelColor, false)
	vector.StrokeRect(screen, dialogX, dialogY, dialogW, dialogH, 2, borderColor, false)

	drawTitle(screen, dialogX, dialogY, dialogW)
	drawMetadata(screen, dialogX, dialogY+60, c.localMeta, c.cloudMeta)
	drawOptions(screen, dialogX, dialogY+240, dialogW, c.options, c.selectedIndex)
}

// drawTitle renders the dialog title.
func drawTitle(screen *ebiten.Image, x, y, width float32) {
	titleColor := color.RGBA{255, 200, 0, 255}
	title := "Save Conflict Detected"
	titleX := int(x + width/2 - float32(len(title)*7)/2)
	text.Draw(screen, title, basicfont.Face7x13, titleX, int(y+30), titleColor)
}

// drawMetadata renders save metadata comparison.
func drawMetadata(screen *ebiten.Image, x, y float32, local, cloudData cloud.SaveMetadata) {
	textColor := color.RGBA{200, 200, 200, 255}
	highlightColor := color.RGBA{100, 200, 255, 255}

	labels := []string{
		"Local Save:",
		fmt.Sprintf("  Slot: %d", local.SlotID),
		fmt.Sprintf("  Time: %s", formatTime(local.Timestamp)),
		fmt.Sprintf("  Genre: %s", local.Genre),
		"Cloud Save:",
		fmt.Sprintf("  Slot: %d", cloudData.SlotID),
		fmt.Sprintf("  Time: %s", formatTime(cloudData.Timestamp)),
		fmt.Sprintf("  Genre: %s", cloudData.Genre),
	}

	offsetY := int(y)
	for i, label := range labels {
		col := textColor
		if i == 0 || i == 4 {
			col = highlightColor
		}
		text.Draw(screen, label, basicfont.Face7x13, int(x+20), offsetY, col)
		offsetY += 20
	}
}

// drawOptions renders selectable resolution options.
func drawOptions(screen *ebiten.Image, x, y, width float32, opts []string, selected int) {
	optionColor := color.RGBA{200, 200, 200, 255}
	selectedColor := color.RGBA{255, 255, 0, 255}

	offsetY := int(y)
	for i, opt := range opts {
		col := optionColor
		prefix := "  "
		if i == selected {
			col = selectedColor
			prefix = "> "
		}
		label := prefix + opt
		text.Draw(screen, label, basicfont.Face7x13, int(x+20), offsetY, col)
		offsetY += 25
	}
}

// formatTime formats timestamp for display.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return "Never"
	}
	return t.Format("2006-01-02 15:04:05")
}
