package federation

import (
	"testing"
	"time"

	"github.com/opd-ai/violence/pkg/chat"
)

func TestNewSquadChatChannel(t *testing.T) {
	// Start a relay server for testing
	relayServer, err := chat.NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create relay server: %v", err)
	}
	defer relayServer.Stop()

	if err := relayServer.Start(); err != nil {
		t.Fatalf("failed to start relay server: %v", err)
	}

	addr := relayServer.GetAddr()

	tests := []struct {
		name      string
		squadID   string
		playerID  string
		wantError bool
	}{
		{
			name:      "create channel successfully",
			squadID:   "squad-1",
			playerID:  "player-1",
			wantError: false,
		},
		{
			name:      "create channel with different squad",
			squadID:   "squad-2",
			playerID:  "player-2",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel, err := NewSquadChatChannel(tt.squadID, addr, tt.playerID)
			if (err != nil) != tt.wantError {
				t.Errorf("NewSquadChatChannel() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if err == nil {
				defer channel.Close()

				if channel.squadID != tt.squadID {
					t.Errorf("squadID = %v, want %v", channel.squadID, tt.squadID)
				}
				if len(channel.encryptKey) != 32 {
					t.Errorf("encryptKey length = %d, want 32", len(channel.encryptKey))
				}
				if channel.relayClient == nil {
					t.Error("relayClient is nil")
				}
				if channel.chat == nil {
					t.Error("chat is nil")
				}
			}
		})
	}
}

func TestNewSquadChatChannelWithKey(t *testing.T) {
	relayServer, err := chat.NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create relay server: %v", err)
	}
	defer relayServer.Stop()

	if err := relayServer.Start(); err != nil {
		t.Fatalf("failed to start relay server: %v", err)
	}

	addr := relayServer.GetAddr()

	// Create a 32-byte key
	validKey := make([]byte, 32)
	for i := range validKey {
		validKey[i] = byte(i)
	}

	tests := []struct {
		name      string
		squadID   string
		key       []byte
		playerID  string
		wantError bool
	}{
		{
			name:      "valid key",
			squadID:   "squad-1",
			key:       validKey,
			playerID:  "player-1",
			wantError: false,
		},
		{
			name:      "invalid key length",
			squadID:   "squad-2",
			key:       make([]byte, 16), // Wrong length
			playerID:  "player-2",
			wantError: true,
		},
		{
			name:      "nil key",
			squadID:   "squad-3",
			key:       nil,
			playerID:  "player-3",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel, err := NewSquadChatChannelWithKey(tt.squadID, tt.key, addr, tt.playerID)
			if (err != nil) != tt.wantError {
				t.Errorf("NewSquadChatChannelWithKey() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if err == nil {
				defer channel.Close()

				if channel.squadID != tt.squadID {
					t.Errorf("squadID = %v, want %v", channel.squadID, tt.squadID)
				}
				if len(channel.encryptKey) != 32 {
					t.Errorf("encryptKey length = %d, want 32", len(channel.encryptKey))
				}
			}
		})
	}
}

func TestSquadChatChannel_SendAndReceive(t *testing.T) {
	// Start relay server
	relayServer, err := chat.NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create relay server: %v", err)
	}
	defer relayServer.Stop()

	if err := relayServer.Start(); err != nil {
		t.Fatalf("failed to start relay server: %v", err)
	}

	addr := relayServer.GetAddr()

	// Create two squad members with shared key
	sharedKey := make([]byte, 32)
	for i := range sharedKey {
		sharedKey[i] = byte(i)
	}

	channel1, err := NewSquadChatChannelWithKey("squad-1", sharedKey, addr, "player-1")
	if err != nil {
		t.Fatalf("failed to create channel1: %v", err)
	}
	defer channel1.Close()

	channel2, err := NewSquadChatChannelWithKey("squad-1", sharedKey, addr, "player-2")
	if err != nil {
		t.Fatalf("failed to create channel2: %v", err)
	}
	defer channel2.Close()

	// Wait for connections to establish
	time.Sleep(100 * time.Millisecond)

	// Player 1 sends a message
	testMessage := "Hello squad!"
	if err := channel1.SendMessage("player-1", testMessage); err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	// Wait for message propagation
	time.Sleep(200 * time.Millisecond)

	// Player 2 receives the message
	messages, err := channel2.ReceiveMessages()
	if err != nil {
		t.Fatalf("failed to receive messages: %v", err)
	}

	if len(messages) == 0 {
		t.Error("expected to receive at least one message")
	}
}

func TestSquadChatChannel_MultipleMessages(t *testing.T) {
	relayServer, err := chat.NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create relay server: %v", err)
	}
	defer relayServer.Stop()

	if err := relayServer.Start(); err != nil {
		t.Fatalf("failed to start relay server: %v", err)
	}

	addr := relayServer.GetAddr()

	sharedKey := make([]byte, 32)
	channel, err := NewSquadChatChannelWithKey("squad-1", sharedKey, addr, "player-1")
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}
	defer channel.Close()

	// Send multiple messages
	messages := []string{"Message 1", "Message 2", "Message 3"}
	for _, msg := range messages {
		if err := channel.SendMessage("player-1", msg); err != nil {
			t.Errorf("failed to send message '%s': %v", msg, err)
		}
	}

	// Check messages are stored
	stored := channel.GetMessages()
	if len(stored) != len(messages) {
		t.Errorf("stored messages count = %d, want %d", len(stored), len(messages))
	}

	for i, msg := range stored {
		if msg.Content != messages[i] {
			t.Errorf("message[%d] = %v, want %v", i, msg.Content, messages[i])
		}
	}
}

func TestSquadChatChannel_EmptyMessage(t *testing.T) {
	relayServer, err := chat.NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create relay server: %v", err)
	}
	defer relayServer.Stop()

	if err := relayServer.Start(); err != nil {
		t.Fatalf("failed to start relay server: %v", err)
	}

	addr := relayServer.GetAddr()

	sharedKey := make([]byte, 32)
	channel, err := NewSquadChatChannelWithKey("squad-1", sharedKey, addr, "player-1")
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}
	defer channel.Close()

	// Try to send empty message
	err = channel.SendMessage("player-1", "")
	if err == nil {
		t.Error("expected error for empty message, got nil")
	}
}

func TestSquadChatChannel_GetEncryptionKey(t *testing.T) {
	relayServer, err := chat.NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create relay server: %v", err)
	}
	defer relayServer.Stop()

	if err := relayServer.Start(); err != nil {
		t.Fatalf("failed to start relay server: %v", err)
	}

	addr := relayServer.GetAddr()

	originalKey := make([]byte, 32)
	for i := range originalKey {
		originalKey[i] = byte(i)
	}

	channel, err := NewSquadChatChannelWithKey("squad-1", originalKey, addr, "player-1")
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}
	defer channel.Close()

	// Get key copy
	keyCopy := channel.GetEncryptionKey()

	// Verify it's a copy, not the original
	if &keyCopy[0] == &channel.encryptKey[0] {
		t.Error("GetEncryptionKey returned reference to internal key, should be copy")
	}

	// Verify content matches
	if len(keyCopy) != len(originalKey) {
		t.Errorf("key length = %d, want %d", len(keyCopy), len(originalKey))
	}

	for i := range keyCopy {
		if keyCopy[i] != originalKey[i] {
			t.Errorf("key[%d] = %d, want %d", i, keyCopy[i], originalKey[i])
		}
	}
}

func TestSquadChatChannel_ClearMessages(t *testing.T) {
	relayServer, err := chat.NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create relay server: %v", err)
	}
	defer relayServer.Stop()

	if err := relayServer.Start(); err != nil {
		t.Fatalf("failed to start relay server: %v", err)
	}

	addr := relayServer.GetAddr()

	sharedKey := make([]byte, 32)
	channel, err := NewSquadChatChannelWithKey("squad-1", sharedKey, addr, "player-1")
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}
	defer channel.Close()

	// Send some messages
	channel.SendMessage("player-1", "Message 1")
	channel.SendMessage("player-1", "Message 2")

	if len(channel.GetMessages()) != 2 {
		t.Errorf("expected 2 messages, got %d", len(channel.GetMessages()))
	}

	// Clear messages
	channel.ClearMessages()

	if len(channel.GetMessages()) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(channel.GetMessages()))
	}
}

func TestSquadChatManager_CreateChannel(t *testing.T) {
	relayServer, err := chat.NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create relay server: %v", err)
	}
	defer relayServer.Stop()

	if err := relayServer.Start(); err != nil {
		t.Fatalf("failed to start relay server: %v", err)
	}

	addr := relayServer.GetAddr()

	manager := NewSquadChatManager()

	// Create first channel
	channel1, err := manager.CreateChannel("squad-1", addr, "player-1")
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}
	defer channel1.Close()

	if channel1.GetSquadID() != "squad-1" {
		t.Errorf("squadID = %v, want squad-1", channel1.GetSquadID())
	}

	// Try to create duplicate channel
	_, err = manager.CreateChannel("squad-1", addr, "player-2")
	if err == nil {
		t.Error("expected error for duplicate channel, got nil")
	}

	// Create different squad channel
	channel2, err := manager.CreateChannel("squad-2", addr, "player-3")
	if err != nil {
		t.Fatalf("failed to create second channel: %v", err)
	}
	defer channel2.Close()
}

func TestSquadChatManager_JoinChannel(t *testing.T) {
	relayServer, err := chat.NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create relay server: %v", err)
	}
	defer relayServer.Stop()

	if err := relayServer.Start(); err != nil {
		t.Fatalf("failed to start relay server: %v", err)
	}

	addr := relayServer.GetAddr()

	manager := NewSquadChatManager()

	// Create shared key
	sharedKey := make([]byte, 32)
	for i := range sharedKey {
		sharedKey[i] = byte(i)
	}

	// Join channel
	channel, err := manager.JoinChannel("squad-1", sharedKey, addr, "player-1")
	if err != nil {
		t.Fatalf("failed to join channel: %v", err)
	}
	defer channel.Close()

	// Try to join same channel again
	_, err = manager.JoinChannel("squad-1", sharedKey, addr, "player-1")
	if err == nil {
		t.Error("expected error for duplicate join, got nil")
	}

	// Join with invalid key
	invalidKey := make([]byte, 16)
	_, err = manager.JoinChannel("squad-2", invalidKey, addr, "player-2")
	if err == nil {
		t.Error("expected error for invalid key length, got nil")
	}
}

func TestSquadChatManager_GetChannel(t *testing.T) {
	relayServer, err := chat.NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create relay server: %v", err)
	}
	defer relayServer.Stop()

	if err := relayServer.Start(); err != nil {
		t.Fatalf("failed to start relay server: %v", err)
	}

	addr := relayServer.GetAddr()

	manager := NewSquadChatManager()

	// Create a channel
	created, err := manager.CreateChannel("squad-1", addr, "player-1")
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}
	defer created.Close()

	// Get existing channel
	retrieved, err := manager.GetChannel("squad-1")
	if err != nil {
		t.Fatalf("failed to get channel: %v", err)
	}

	if retrieved.GetSquadID() != "squad-1" {
		t.Errorf("squadID = %v, want squad-1", retrieved.GetSquadID())
	}

	// Try to get non-existent channel
	_, err = manager.GetChannel("squad-99")
	if err == nil {
		t.Error("expected error for non-existent channel, got nil")
	}
}

func TestSquadChatManager_RemoveChannel(t *testing.T) {
	relayServer, err := chat.NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create relay server: %v", err)
	}
	defer relayServer.Stop()

	if err := relayServer.Start(); err != nil {
		t.Fatalf("failed to start relay server: %v", err)
	}

	addr := relayServer.GetAddr()

	manager := NewSquadChatManager()

	// Create a channel
	_, err = manager.CreateChannel("squad-1", addr, "player-1")
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}

	// Remove channel
	if err := manager.RemoveChannel("squad-1"); err != nil {
		t.Fatalf("failed to remove channel: %v", err)
	}

	// Verify it's removed
	_, err = manager.GetChannel("squad-1")
	if err == nil {
		t.Error("expected error after removal, got nil")
	}

	// Try to remove non-existent channel
	err = manager.RemoveChannel("squad-99")
	if err == nil {
		t.Error("expected error for non-existent channel, got nil")
	}
}

func TestSquadChatManager_CloseAll(t *testing.T) {
	relayServer, err := chat.NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create relay server: %v", err)
	}
	defer relayServer.Stop()

	if err := relayServer.Start(); err != nil {
		t.Fatalf("failed to start relay server: %v", err)
	}

	addr := relayServer.GetAddr()

	manager := NewSquadChatManager()

	// Create multiple channels
	for i := 1; i <= 3; i++ {
		squadID := "squad-" + string(rune('0'+i))
		playerID := "player-" + string(rune('0'+i))
		_, err := manager.CreateChannel(squadID, addr, playerID)
		if err != nil {
			t.Fatalf("failed to create channel %d: %v", i, err)
		}
	}

	// Close all channels
	if err := manager.CloseAll(); err != nil {
		t.Fatalf("failed to close all channels: %v", err)
	}

	// Verify all are removed
	for i := 1; i <= 3; i++ {
		squadID := "squad-" + string(rune('0'+i))
		_, err := manager.GetChannel(squadID)
		if err == nil {
			t.Errorf("expected channel %s to be removed", squadID)
		}
	}
}

func TestSquadChatIntegration(t *testing.T) {
	// Integration test: Full squad chat workflow
	relayServer, err := chat.NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create relay server: %v", err)
	}
	defer relayServer.Stop()

	if err := relayServer.Start(); err != nil {
		t.Fatalf("failed to start relay server: %v", err)
	}

	addr := relayServer.GetAddr()

	// Create squad and chat manager
	manager := NewSquadChatManager()

	// Player 1 creates the squad and chat channel
	channel1, err := manager.CreateChannel("alpha-squad", addr, "player-1")
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}
	defer channel1.Close()

	// Get the shared encryption key to share with new members
	squadKey := channel1.GetEncryptionKey()

	// Player 2 joins using the shared key
	manager2 := NewSquadChatManager()
	channel2, err := manager2.JoinChannel("alpha-squad", squadKey, addr, "player-2")
	if err != nil {
		t.Fatalf("failed to join channel: %v", err)
	}
	defer channel2.Close()

	// Player 3 joins using the shared key
	manager3 := NewSquadChatManager()
	channel3, err := manager3.JoinChannel("alpha-squad", squadKey, addr, "player-3")
	if err != nil {
		t.Fatalf("failed to join channel: %v", err)
	}
	defer channel3.Close()

	// Wait for connections
	time.Sleep(100 * time.Millisecond)

	// Player 1 sends a message
	if err := channel1.SendMessage("player-1", "Squad, rally at point Alpha!"); err != nil {
		t.Fatalf("player 1 failed to send message: %v", err)
	}

	// Wait for message propagation
	time.Sleep(200 * time.Millisecond)

	// Player 2 and 3 should receive the message
	messages2, err := channel2.ReceiveMessages()
	if err != nil {
		t.Fatalf("player 2 failed to receive messages: %v", err)
	}

	messages3, err := channel3.ReceiveMessages()
	if err != nil {
		t.Fatalf("player 3 failed to receive messages: %v", err)
	}

	// Both should have received at least one message
	if len(messages2) == 0 {
		t.Error("player 2 received no messages")
	}
	if len(messages3) == 0 {
		t.Error("player 3 received no messages")
	}

	// Clear chat history
	channel1.ClearMessages()
	if len(channel1.GetMessages()) != 0 {
		t.Error("messages not cleared")
	}
}
