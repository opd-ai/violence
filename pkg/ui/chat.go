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
	visible        bool
	messages       []ChatMessage
	inputBuffer    string
	cursorPosition int
	scrollOffset   int
	X              int
	Y              int
	Width          int
	Height         int
}

// NewChatOverlay creates a new chat overlay.
func NewChatOverlay(x, y, width, height int) *ChatOverlay {
	return &ChatOverlay{
		visible:        false,
		messages:       make([]ChatMessage, 0, ChatHistoryMaxLength),
		inputBuffer:    "",
		cursorPosition: 0,
		scrollOffset:   0,
		X:              x,
		Y:              y,
		Width:          width,
		Height:         height,
	}
}

// Show makes the chat overlay visible.
func (co *ChatOverlay) Show() {
	co.mu.Lock()
	defer co.mu.Unlock()
	co.visible = true
}

// Hide makes the chat overlay invisible.
func (co *ChatOverlay) Hide() {
	co.mu.Lock()
	defer co.mu.Unlock()
	co.visible = false
	co.inputBuffer = ""
	co.cursorPosition = 0
}

// Toggle toggles chat overlay visibility.
func (co *ChatOverlay) Toggle() {
	co.mu.Lock()
	defer co.mu.Unlock()
	co.visible = !co.visible
	if !co.visible {
		co.inputBuffer = ""
		co.cursorPosition = 0
	}
}

// IsVisible returns whether the chat overlay is currently visible.
func (co *ChatOverlay) IsVisible() bool {
	co.mu.Lock()
	defer co.mu.Unlock()
	return co.visible
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

	co.messages = append(co.messages, msg)

	// Trim old messages
	if len(co.messages) > ChatHistoryMaxLength {
		co.messages = co.messages[1:]
	}

	// Auto-scroll to bottom on new message
	if len(co.messages) > ChatMaxVisibleMessages {
		co.scrollOffset = len(co.messages) - ChatMaxVisibleMessages
	}
}

// GetInput returns the current input buffer.
func (co *ChatOverlay) GetInput() string {
	co.mu.Lock()
	defer co.mu.Unlock()
	return co.inputBuffer
}

// ClearInput clears the input buffer.
func (co *ChatOverlay) ClearInput() {
	co.mu.Lock()
	defer co.mu.Unlock()
	co.inputBuffer = ""
	co.cursorPosition = 0
}

// AppendToInput appends a character to the input buffer.
func (co *ChatOverlay) AppendToInput(char rune) {
	co.mu.Lock()
	defer co.mu.Unlock()
	if len(co.inputBuffer) < ChatInputMaxLength {
		co.inputBuffer += string(char)
		co.cursorPosition = len(co.inputBuffer)
	}
}

// Backspace removes the last character from the input buffer.
func (co *ChatOverlay) Backspace() {
	co.mu.Lock()
	defer co.mu.Unlock()
	if len(co.inputBuffer) > 0 {
		co.inputBuffer = co.inputBuffer[:len(co.inputBuffer)-1]
		co.cursorPosition = len(co.inputBuffer)
	}
}

// ScrollUp scrolls the message history up.
func (co *ChatOverlay) ScrollUp() {
	co.mu.Lock()
	defer co.mu.Unlock()
	if co.scrollOffset > 0 {
		co.scrollOffset--
	}
}

// ScrollDown scrolls the message history down.
func (co *ChatOverlay) ScrollDown() {
	co.mu.Lock()
	defer co.mu.Unlock()
	maxScroll := len(co.messages) - ChatMaxVisibleMessages
	if maxScroll < 0 {
		maxScroll = 0
	}
	if co.scrollOffset < maxScroll {
		co.scrollOffset++
	}
}

// GetVisibleMessages returns the messages currently visible based on scroll offset.
func (co *ChatOverlay) GetVisibleMessages() []ChatMessage {
	co.mu.Lock()
	defer co.mu.Unlock()
	totalMessages := len(co.messages)
	if totalMessages == 0 {
		return []ChatMessage{}
	}

	startIdx := co.scrollOffset
	endIdx := co.scrollOffset + ChatMaxVisibleMessages

	if endIdx > totalMessages {
		endIdx = totalMessages
	}
	if startIdx >= totalMessages {
		startIdx = totalMessages - 1
		if startIdx < 0 {
			startIdx = 0
		}
	}

	return co.messages[startIdx:endIdx]
}

// GetMessages returns a copy of all messages. Safe for concurrent use.
func (co *ChatOverlay) GetMessages() []ChatMessage {
	co.mu.Lock()
	defer co.mu.Unlock()
	result := make([]ChatMessage, len(co.messages))
	copy(result, co.messages)
	return result
}

// GetCursorPosition returns the current cursor position.
func (co *ChatOverlay) GetCursorPosition() int {
	co.mu.Lock()
	defer co.mu.Unlock()
	return co.cursorPosition
}

// GetScrollOffset returns the current scroll offset.
func (co *ChatOverlay) GetScrollOffset() int {
	co.mu.Lock()
	defer co.mu.Unlock()
	return co.scrollOffset
}

// SetScrollOffset sets the scroll offset directly (for testing).
func (co *ChatOverlay) SetScrollOffset(offset int) {
	co.mu.Lock()
	defer co.mu.Unlock()
	co.scrollOffset = offset
}

// SetInputBuffer sets the input buffer directly (for testing).
func (co *ChatOverlay) SetInputBuffer(input string) {
	co.mu.Lock()
	defer co.mu.Unlock()
	co.inputBuffer = input
	co.cursorPosition = len(input)
}

// SetVisible sets the visibility state directly (for testing).
func (co *ChatOverlay) SetVisible(visible bool) {
	co.mu.Lock()
	defer co.mu.Unlock()
	co.visible = visible
}

// ClearMessages removes all messages (for testing).
func (co *ChatOverlay) ClearMessages() {
	co.mu.Lock()
	defer co.mu.Unlock()
	co.messages = []ChatMessage{}
	co.scrollOffset = 0
}

// Draw renders the chat overlay to the screen.
func (co *ChatOverlay) Draw(screen *ebiten.Image) {
	co.mu.Lock()
	visible := co.visible
	inputBuffer := co.inputBuffer
	messages := co.messages
	scrollOffset := co.scrollOffset
	co.mu.Unlock()

	if !visible {
		return
	}

	x := float32(co.X)
	y := float32(co.Y)
	width := float32(co.Width)
	height := float32(co.Height)

	drawChatBackground(screen, x, y, width, height)
	drawChatTitle(screen, co.X, co.Y)

	messageY := co.Y + 35
	visibleMessages := calculateVisibleMessages(messages, scrollOffset)
	drawChatMessages(screen, visibleMessages, co.X, &messageY)

	inputY := co.Y + co.Height - 30
	drawInputArea(screen, x, y, width, inputY, inputBuffer, visible)

	if len(messages) > ChatMaxVisibleMessages {
		drawScrollIndicator(screen, co.X, co.Y, co.Width, scrollOffset, len(messages))
	}

	drawChatHelp(screen, co.X, co.Y+co.Height-10)
}

// drawChatBackground renders the semi-transparent background and border.
func drawChatBackground(screen *ebiten.Image, x, y, width, height float32) {
	bgColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}
	vector.DrawFilledRect(screen, x, y, width, height, bgColor, false)

	borderColor := color.RGBA{R: 100, G: 150, B: 100, A: 255}
	vector.StrokeRect(screen, x, y, width, height, 2, borderColor, false)
}

// drawChatTitle renders the chat overlay title.
func drawChatTitle(screen *ebiten.Image, x, y int) {
	titleColor := color.RGBA{R: 200, G: 255, B: 200, A: 255}
	text.Draw(screen, "CHAT", basicfont.Face7x13, x+10, y+20, titleColor)
}

// calculateVisibleMessages determines which messages to display based on scroll position.
func calculateVisibleMessages(messages []ChatMessage, scrollOffset int) []ChatMessage {
	var visibleMessages []ChatMessage
	totalMessages := len(messages)
	if totalMessages > 0 {
		startIdx := scrollOffset
		endIdx := scrollOffset + ChatMaxVisibleMessages

		if endIdx > totalMessages {
			endIdx = totalMessages
		}
		if startIdx >= totalMessages {
			startIdx = totalMessages - 1
			if startIdx < 0 {
				startIdx = 0
			}
		}
		visibleMessages = messages[startIdx:endIdx]
	}
	return visibleMessages
}

// drawChatMessages renders the visible chat message history.
func drawChatMessages(screen *ebiten.Image, visibleMessages []ChatMessage, x int, messageY *int) {
	messageColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	senderColor := color.RGBA{R: 100, G: 200, B: 255, A: 255}

	for _, msg := range visibleMessages {
		senderText := fmt.Sprintf("%s:", msg.Sender)
		text.Draw(screen, senderText, basicfont.Face7x13, x+10, *messageY, senderColor)

		contentX := x + 10 + len(senderText)*7 + 5
		text.Draw(screen, msg.Content, basicfont.Face7x13, contentX, *messageY, messageColor)

		*messageY += 15
	}
}

// drawInputArea renders the input field with separator, prompt, and cursor.
func drawInputArea(screen *ebiten.Image, x, y, width float32, inputY int, inputBuffer string, visible bool) {
	separatorColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	vector.StrokeLine(screen, x, float32(inputY-5), x+width, float32(inputY-5), 1, separatorColor, false)

	promptColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	text.Draw(screen, ">", basicfont.Face7x13, int(x)+10, inputY+10, promptColor)

	inputColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	text.Draw(screen, inputBuffer, basicfont.Face7x13, int(x)+25, inputY+10, inputColor)

	if visible && ebiten.TPS() > 0 {
		cursorX := int(x) + 25 + len(inputBuffer)*7
		cursorColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
		vector.StrokeLine(screen, float32(cursorX), float32(inputY), float32(cursorX), float32(inputY+12), 2, cursorColor, false)
	}
}

// drawScrollIndicator shows the current scroll position when there are many messages.
func drawScrollIndicator(screen *ebiten.Image, x, y, width, scrollOffset, totalMessages int) {
	scrollText := fmt.Sprintf("%d/%d", scrollOffset+1, totalMessages-ChatMaxVisibleMessages+1)
	scrollColor := color.RGBA{R: 150, G: 150, B: 150, A: 255}
	scrollX := x + width - len(scrollText)*7 - 10
	text.Draw(screen, scrollText, basicfont.Face7x13, scrollX, y+20, scrollColor)
}

// drawChatHelp renders the help text at the bottom of the chat overlay.
func drawChatHelp(screen *ebiten.Image, x, y int) {
	helpColor := color.RGBA{R: 150, G: 150, B: 150, A: 255}
	helpText := "Enter: Send | Esc: Close | PgUp/PgDn: Scroll"
	text.Draw(screen, helpText, basicfont.Face7x13, x+10, y, helpColor)
}
