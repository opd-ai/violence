package ui

import (
	"testing"
)

func TestNewChatOverlay(t *testing.T) {
	tests := []struct {
		name   string
		x      int
		y      int
		width  int
		height int
	}{
		{"default position", 10, 10, 400, 300},
		{"custom position", 50, 100, 600, 400},
		{"zero position", 0, 0, 320, 240},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			co := NewChatOverlay(tt.x, tt.y, tt.width, tt.height)

			if co == nil {
				t.Fatal("NewChatOverlay returned nil")
			}
			if co.Visible {
				t.Error("ChatOverlay should start hidden")
			}
			if co.X != tt.x {
				t.Errorf("X = %d, want %d", co.X, tt.x)
			}
			if co.Y != tt.y {
				t.Errorf("Y = %d, want %d", co.Y, tt.y)
			}
			if co.Width != tt.width {
				t.Errorf("Width = %d, want %d", co.Width, tt.width)
			}
			if co.Height != tt.height {
				t.Errorf("Height = %d, want %d", co.Height, tt.height)
			}
			if len(co.Messages) != 0 {
				t.Errorf("Messages should be empty initially, got %d", len(co.Messages))
			}
			if co.InputBuffer != "" {
				t.Error("InputBuffer should be empty initially")
			}
		})
	}
}

func TestChatOverlayVisibility(t *testing.T) {
	co := NewChatOverlay(10, 10, 400, 300)

	t.Run("initial state", func(t *testing.T) {
		if co.IsVisible() {
			t.Error("ChatOverlay should start invisible")
		}
	})

	t.Run("show", func(t *testing.T) {
		co.Show()
		if !co.IsVisible() {
			t.Error("ChatOverlay should be visible after Show()")
		}
	})

	t.Run("hide", func(t *testing.T) {
		co.InputBuffer = "test"
		co.Hide()
		if co.IsVisible() {
			t.Error("ChatOverlay should be invisible after Hide()")
		}
		if co.InputBuffer != "" {
			t.Error("InputBuffer should be cleared on Hide()")
		}
	})

	t.Run("toggle on", func(t *testing.T) {
		co.Visible = false
		co.Toggle()
		if !co.IsVisible() {
			t.Error("Toggle should make invisible overlay visible")
		}
	})

	t.Run("toggle off", func(t *testing.T) {
		co.Visible = true
		co.InputBuffer = "test"
		co.Toggle()
		if co.IsVisible() {
			t.Error("Toggle should make visible overlay invisible")
		}
		if co.InputBuffer != "" {
			t.Error("InputBuffer should be cleared when toggling off")
		}
	})
}

func TestAddMessage(t *testing.T) {
	co := NewChatOverlay(10, 10, 400, 300)

	tests := []struct {
		name      string
		sender    string
		content   string
		timestamp int64
	}{
		{"first message", "Player1", "Hello!", 1000},
		{"second message", "Player2", "Hi there", 1001},
		{"empty content", "Player3", "", 1002},
		{"long message", "Player4", "This is a very long message that should still be stored correctly in the chat history", 1003},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialCount := len(co.Messages)
			co.AddMessage(tt.sender, tt.content, tt.timestamp)

			if len(co.Messages) != initialCount+1 {
				t.Errorf("Message count = %d, want %d", len(co.Messages), initialCount+1)
			}

			lastMsg := co.Messages[len(co.Messages)-1]
			if lastMsg.Sender != tt.sender {
				t.Errorf("Sender = %q, want %q", lastMsg.Sender, tt.sender)
			}
			if lastMsg.Content != tt.content {
				t.Errorf("Content = %q, want %q", lastMsg.Content, tt.content)
			}
			if lastMsg.Time != tt.timestamp {
				t.Errorf("Time = %d, want %d", lastMsg.Time, tt.timestamp)
			}
		})
	}
}

func TestAddMessageHistoryLimit(t *testing.T) {
	co := NewChatOverlay(10, 10, 400, 300)

	// Add more than the history limit
	for i := 0; i < ChatHistoryMaxLength+50; i++ {
		co.AddMessage("Player", "Message", int64(i))
	}

	if len(co.Messages) > ChatHistoryMaxLength {
		t.Errorf("Message count = %d, should not exceed %d", len(co.Messages), ChatHistoryMaxLength)
	}

	// Verify oldest messages were removed
	if co.Messages[0].Time != 50 {
		t.Errorf("Oldest message time = %d, want 50", co.Messages[0].Time)
	}
}

func TestAddMessageAutoScroll(t *testing.T) {
	co := NewChatOverlay(10, 10, 400, 300)

	// Add enough messages to trigger auto-scroll
	for i := 0; i < ChatMaxVisibleMessages+5; i++ {
		co.AddMessage("Player", "Message", int64(i))
	}

	expectedScroll := len(co.Messages) - ChatMaxVisibleMessages
	if co.ScrollOffset != expectedScroll {
		t.Errorf("ScrollOffset = %d, want %d", co.ScrollOffset, expectedScroll)
	}
}

func TestInputBuffer(t *testing.T) {
	co := NewChatOverlay(10, 10, 400, 300)

	t.Run("get empty input", func(t *testing.T) {
		if co.GetInput() != "" {
			t.Error("GetInput should return empty string initially")
		}
	})

	t.Run("append to input", func(t *testing.T) {
		co.AppendToInput('H')
		co.AppendToInput('i')
		if co.GetInput() != "Hi" {
			t.Errorf("GetInput = %q, want %q", co.GetInput(), "Hi")
		}
	})

	t.Run("backspace", func(t *testing.T) {
		co.InputBuffer = "Hello"
		co.Backspace()
		if co.GetInput() != "Hell" {
			t.Errorf("GetInput = %q, want %q", co.GetInput(), "Hell")
		}
	})

	t.Run("backspace on empty", func(t *testing.T) {
		co.InputBuffer = ""
		co.Backspace()
		if co.GetInput() != "" {
			t.Error("Backspace on empty buffer should remain empty")
		}
	})

	t.Run("clear input", func(t *testing.T) {
		co.InputBuffer = "Test message"
		co.ClearInput()
		if co.GetInput() != "" {
			t.Error("ClearInput should empty the buffer")
		}
		if co.CursorPosition != 0 {
			t.Error("ClearInput should reset cursor position")
		}
	})

	t.Run("max length enforcement", func(t *testing.T) {
		co.InputBuffer = ""
		for i := 0; i < ChatInputMaxLength+10; i++ {
			co.AppendToInput('x')
		}
		if len(co.GetInput()) > ChatInputMaxLength {
			t.Errorf("Input length = %d, should not exceed %d", len(co.GetInput()), ChatInputMaxLength)
		}
	})
}

func TestScrolling(t *testing.T) {
	co := NewChatOverlay(10, 10, 400, 300)

	// Add more messages than can be visible
	for i := 0; i < ChatMaxVisibleMessages*2; i++ {
		co.AddMessage("Player", "Message", int64(i))
	}

	t.Run("scroll up", func(t *testing.T) {
		initialScroll := co.ScrollOffset
		co.ScrollUp()
		if co.ScrollOffset != initialScroll-1 {
			t.Errorf("ScrollOffset = %d, want %d", co.ScrollOffset, initialScroll-1)
		}
	})

	t.Run("scroll up at top", func(t *testing.T) {
		co.ScrollOffset = 0
		co.ScrollUp()
		if co.ScrollOffset != 0 {
			t.Error("ScrollUp at top should not go negative")
		}
	})

	t.Run("scroll down", func(t *testing.T) {
		co.ScrollOffset = 0
		co.ScrollDown()
		if co.ScrollOffset != 1 {
			t.Errorf("ScrollOffset = %d, want 1", co.ScrollOffset)
		}
	})

	t.Run("scroll down at bottom", func(t *testing.T) {
		maxScroll := len(co.Messages) - ChatMaxVisibleMessages
		co.ScrollOffset = maxScroll
		co.ScrollDown()
		if co.ScrollOffset > maxScroll {
			t.Error("ScrollDown at bottom should not exceed max")
		}
	})
}

func TestGetVisibleMessages(t *testing.T) {
	co := NewChatOverlay(10, 10, 400, 300)

	t.Run("no messages", func(t *testing.T) {
		visible := co.GetVisibleMessages()
		if len(visible) != 0 {
			t.Errorf("Visible messages = %d, want 0", len(visible))
		}
	})

	t.Run("less than max visible", func(t *testing.T) {
		co.Messages = []ChatMessage{}
		for i := 0; i < 5; i++ {
			co.AddMessage("Player", "Message", int64(i))
		}
		co.ScrollOffset = 0

		visible := co.GetVisibleMessages()
		if len(visible) != 5 {
			t.Errorf("Visible messages = %d, want 5", len(visible))
		}
	})

	t.Run("exactly max visible", func(t *testing.T) {
		co.Messages = []ChatMessage{}
		for i := 0; i < ChatMaxVisibleMessages; i++ {
			co.AddMessage("Player", "Message", int64(i))
		}
		co.ScrollOffset = 0

		visible := co.GetVisibleMessages()
		if len(visible) != ChatMaxVisibleMessages {
			t.Errorf("Visible messages = %d, want %d", len(visible), ChatMaxVisibleMessages)
		}
	})

	t.Run("more than max visible", func(t *testing.T) {
		co.Messages = []ChatMessage{}
		for i := 0; i < ChatMaxVisibleMessages+10; i++ {
			co.AddMessage("Player", "Message", int64(i))
		}
		co.ScrollOffset = 0

		visible := co.GetVisibleMessages()
		if len(visible) != ChatMaxVisibleMessages {
			t.Errorf("Visible messages = %d, want %d", len(visible), ChatMaxVisibleMessages)
		}

		// Verify we got the first messages
		if visible[0].Time != 0 {
			t.Errorf("First visible message time = %d, want 0", visible[0].Time)
		}
	})

	t.Run("scrolled position", func(t *testing.T) {
		co.Messages = []ChatMessage{}
		for i := 0; i < ChatMaxVisibleMessages+10; i++ {
			co.AddMessage("Player", "Message", int64(i))
		}
		co.ScrollOffset = 5

		visible := co.GetVisibleMessages()
		if len(visible) != ChatMaxVisibleMessages {
			t.Errorf("Visible messages = %d, want %d", len(visible), ChatMaxVisibleMessages)
		}

		// Verify we got messages starting from offset
		if visible[0].Time != 5 {
			t.Errorf("First visible message time = %d, want 5", visible[0].Time)
		}
	})

	t.Run("invalid scroll offset", func(t *testing.T) {
		co.Messages = []ChatMessage{}
		for i := 0; i < 5; i++ {
			co.AddMessage("Player", "Message", int64(i))
		}
		co.ScrollOffset = 100 // Invalid offset

		visible := co.GetVisibleMessages()
		if len(visible) == 0 {
			t.Error("Should return messages even with invalid offset")
		}
	})
}

func TestChatOverlayConcurrency(t *testing.T) {
	co := NewChatOverlay(10, 10, 400, 300)

	// Test concurrent message additions
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				co.AddMessage("Player", "Message", int64(id*10+j))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have received 100 messages (within history limit)
	expectedCount := 100
	if expectedCount > ChatHistoryMaxLength {
		expectedCount = ChatHistoryMaxLength
	}
	if len(co.Messages) != expectedCount {
		t.Errorf("Message count = %d, want %d", len(co.Messages), expectedCount)
	}
}

// TestChatOverlayRaceConditions verifies that all ChatOverlay methods are race-free.
// Run with: go test -race
func TestChatOverlayRaceConditions(t *testing.T) {
	co := NewChatOverlay(10, 10, 400, 300)

	// Seed with some messages
	for i := 0; i < 15; i++ {
		co.AddMessage("Player", "Message", int64(i))
	}

	done := make(chan bool)

	// Goroutine 1: Add messages
	go func() {
		for i := 0; i < 50; i++ {
			co.AddMessage("Player1", "Concurrent message", int64(i))
		}
		done <- true
	}()

	// Goroutine 2: Toggle visibility
	go func() {
		for i := 0; i < 50; i++ {
			co.Show()
			co.Hide()
			co.Toggle()
		}
		done <- true
	}()

	// Goroutine 3: Input operations
	go func() {
		for i := 0; i < 50; i++ {
			co.AppendToInput('a')
			_ = co.GetInput()
			co.Backspace()
			co.ClearInput()
		}
		done <- true
	}()

	// Goroutine 4: Scroll operations
	go func() {
		for i := 0; i < 50; i++ {
			co.ScrollUp()
			co.ScrollDown()
		}
		done <- true
	}()

	// Goroutine 5: Read operations
	go func() {
		for i := 0; i < 50; i++ {
			_ = co.IsVisible()
			_ = co.GetVisibleMessages()
			_ = co.GetInput()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify state is valid after concurrent access
	if co.ScrollOffset < 0 {
		t.Error("ScrollOffset became negative after concurrent access")
	}
	if co.CursorPosition < 0 {
		t.Error("CursorPosition became negative after concurrent access")
	}
}
