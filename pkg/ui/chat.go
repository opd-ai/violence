package ui

import (
	"fmt"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	ChatMaxVisibleMessages = 10
	ChatInputMaxLength     = 200
	ChatHistoryMaxLength   = 100
)

// ChatMessage represents a single chat message with metadata.
type ChatMessage struct {
	Sender  string
	Content string
	Time    int64
}

// ChatOverlay displays in-game chat UI overlay.
type ChatOverlay struct {
	mu             sync.Mutex
	Visible        bool
	Messages       []ChatMessage
	InputBuffer    string
	CursorPosition int
	ScrollOffset   int
	X              int
	Y              int
	Width          int
	Height         int
}

// NewChatOverlay creates a new chat overlay.
func NewChatOverlay(x, y, width, height int) *ChatOverlay {
	return &ChatOverlay{
		Visible:        false,
		Messages:       make([]ChatMessage, 0, ChatHistoryMaxLength),
		InputBuffer:    "",
		CursorPosition: 0,
		ScrollOffset:   0,
		X:              x,
		Y:              y,
		Width:          width,
		Height:         height,
	}
}

// Show makes the chat overlay visible.
func (co *ChatOverlay) Show() {
	co.Visible = true
}

// Hide makes the chat overlay invisible.
func (co *ChatOverlay) Hide() {
	co.Visible = false
	co.InputBuffer = ""
	co.CursorPosition = 0
}

// Toggle toggles chat overlay visibility.
func (co *ChatOverlay) Toggle() {
	co.Visible = !co.Visible
	if !co.Visible {
		co.InputBuffer = ""
		co.CursorPosition = 0
	}
}

// IsVisible returns whether the chat overlay is currently visible.
func (co *ChatOverlay) IsVisible() bool {
	return co.Visible
}

// AddMessage adds a message to the chat history. Safe for concurrent use.
func (co *ChatOverlay) AddMessage(sender, content string, timestamp int64) {
	co.mu.Lock()
	defer co.mu.Unlock()

	msg := ChatMessage{
		Sender:  sender,
		Content: content,
		Time:    timestamp,
	}

	co.Messages = append(co.Messages, msg)

	// Trim old messages
	if len(co.Messages) > ChatHistoryMaxLength {
		co.Messages = co.Messages[1:]
	}

	// Auto-scroll to bottom on new message
	if len(co.Messages) > ChatMaxVisibleMessages {
		co.ScrollOffset = len(co.Messages) - ChatMaxVisibleMessages
	}
}

// GetInput returns the current input buffer.
func (co *ChatOverlay) GetInput() string {
	return co.InputBuffer
}

// ClearInput clears the input buffer.
func (co *ChatOverlay) ClearInput() {
	co.InputBuffer = ""
	co.CursorPosition = 0
}

// AppendToInput appends a character to the input buffer.
func (co *ChatOverlay) AppendToInput(char rune) {
	if len(co.InputBuffer) < ChatInputMaxLength {
		co.InputBuffer += string(char)
		co.CursorPosition = len(co.InputBuffer)
	}
}

// Backspace removes the last character from the input buffer.
func (co *ChatOverlay) Backspace() {
	if len(co.InputBuffer) > 0 {
		co.InputBuffer = co.InputBuffer[:len(co.InputBuffer)-1]
		co.CursorPosition = len(co.InputBuffer)
	}
}

// ScrollUp scrolls the message history up.
func (co *ChatOverlay) ScrollUp() {
	if co.ScrollOffset > 0 {
		co.ScrollOffset--
	}
}

// ScrollDown scrolls the message history down.
func (co *ChatOverlay) ScrollDown() {
	maxScroll := len(co.Messages) - ChatMaxVisibleMessages
	if maxScroll < 0 {
		maxScroll = 0
	}
	if co.ScrollOffset < maxScroll {
		co.ScrollOffset++
	}
}

// GetVisibleMessages returns the messages currently visible based on scroll offset.
func (co *ChatOverlay) GetVisibleMessages() []ChatMessage {
	totalMessages := len(co.Messages)
	if totalMessages == 0 {
		return []ChatMessage{}
	}

	startIdx := co.ScrollOffset
	endIdx := co.ScrollOffset + ChatMaxVisibleMessages

	if endIdx > totalMessages {
		endIdx = totalMessages
	}
	if startIdx >= totalMessages {
		startIdx = totalMessages - 1
		if startIdx < 0 {
			startIdx = 0
		}
	}

	return co.Messages[startIdx:endIdx]
}

// Draw renders the chat overlay to the screen.
func (co *ChatOverlay) Draw(screen *ebiten.Image) {
	if !co.Visible {
		return
	}

	x := float32(co.X)
	y := float32(co.Y)
	width := float32(co.Width)
	height := float32(co.Height)

	// Semi-transparent background
	bgColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}
	vector.DrawFilledRect(screen, x, y, width, height, bgColor, false)

	// Border
	borderColor := color.RGBA{R: 100, G: 150, B: 100, A: 255}
	vector.StrokeRect(screen, x, y, width, height, 2, borderColor, false)

	// Title
	titleColor := color.RGBA{R: 200, G: 255, B: 200, A: 255}
	text.Draw(screen, "CHAT", basicfont.Face7x13, co.X+10, co.Y+20, titleColor)

	// Message history area
	messageY := co.Y + 35
	messageColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	senderColor := color.RGBA{R: 100, G: 200, B: 255, A: 255}

	visibleMessages := co.GetVisibleMessages()
	for _, msg := range visibleMessages {
		// Draw sender name
		senderText := fmt.Sprintf("%s:", msg.Sender)
		text.Draw(screen, senderText, basicfont.Face7x13, co.X+10, messageY, senderColor)

		// Draw message content (offset after sender)
		contentX := co.X + 10 + len(senderText)*7 + 5
		text.Draw(screen, msg.Content, basicfont.Face7x13, contentX, messageY, messageColor)

		messageY += 15
	}

	// Input area separator
	inputY := co.Y + co.Height - 30
	separatorColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	vector.StrokeLine(screen, x, float32(inputY-5), x+width, float32(inputY-5), 1, separatorColor, false)

	// Input prompt
	promptColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	text.Draw(screen, ">", basicfont.Face7x13, co.X+10, inputY+10, promptColor)

	// Input buffer
	inputColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	text.Draw(screen, co.InputBuffer, basicfont.Face7x13, co.X+25, inputY+10, inputColor)

	// Cursor (blinking effect using frame counter)
	if co.Visible && ebiten.TPS() > 0 {
		cursorX := co.X + 25 + len(co.InputBuffer)*7
		cursorColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
		vector.StrokeLine(screen, float32(cursorX), float32(inputY), float32(cursorX), float32(inputY+12), 2, cursorColor, false)
	}

	// Scroll indicator
	if len(co.Messages) > ChatMaxVisibleMessages {
		scrollText := fmt.Sprintf("%d/%d", co.ScrollOffset+1, len(co.Messages)-ChatMaxVisibleMessages+1)
		scrollColor := color.RGBA{R: 150, G: 150, B: 150, A: 255}
		scrollX := co.X + co.Width - len(scrollText)*7 - 10
		text.Draw(screen, scrollText, basicfont.Face7x13, scrollX, co.Y+20, scrollColor)
	}

	// Help text
	helpColor := color.RGBA{R: 150, G: 150, B: 150, A: 255}
	helpText := "Enter: Send | Esc: Close | PgUp/PgDn: Scroll"
	helpX := co.X + 10
	helpY := co.Y + co.Height - 10
	text.Draw(screen, helpText, basicfont.Face7x13, helpX, helpY, helpColor)
}
