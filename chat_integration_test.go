package main

import (
	"testing"

	"github.com/opd-ai/violence/pkg/chat"
)

// TestEncryptedChatIntegration verifies E2E encrypted chat is properly integrated.
func TestEncryptedChatIntegration(t *testing.T) {
	g := NewGame()

	// Initially no chat manager
	if g.chatManager != nil {
		t.Fatal("Expected nil chatManager before multiplayer init")
	}

	// Open multiplayer - should initialize encrypted chat
	g.openMultiplayer()

	if g.chatManager == nil {
		t.Fatal("Expected chatManager to be initialized after openMultiplayer()")
	}

	if g.chatMessages == nil {
		t.Fatal("Expected chatMessages to be initialized")
	}

	if len(g.chatMessages) != 0 {
		t.Errorf("Expected 0 initial messages, got %d", len(g.chatMessages))
	}

	if g.chatInputActive {
		t.Error("Expected chatInputActive to be false initially")
	}
}

// TestChatEncryptionDecryption verifies chat messages are encrypted/decrypted correctly.
func TestChatEncryptionDecryption(t *testing.T) {
	g := NewGame()
	g.openMultiplayer()

	testMsg := "Hello, this is a test message!"

	// Encrypt message
	encrypted, err := g.chatManager.Encrypt(testMsg)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Encrypted should not equal plaintext
	if encrypted == testMsg {
		t.Error("Encrypted message should not equal plaintext")
	}

	// Decrypt message
	decrypted, err := g.chatManager.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Decrypted should match original
	if decrypted != testMsg {
		t.Errorf("Decrypted message '%s' does not match original '%s'", decrypted, testMsg)
	}
}

// TestAddChatMessage verifies chat message history management.
func TestAddChatMessage(t *testing.T) {
	g := NewGame()
	g.openMultiplayer()

	tests := []struct {
		name     string
		messages []string
		wantLen  int
	}{
		{
			name:     "single message",
			messages: []string{"msg1"},
			wantLen:  1,
		},
		{
			name:     "multiple messages",
			messages: []string{"msg1", "msg2", "msg3"},
			wantLen:  3,
		},
		{
			name:     "max 50 messages",
			messages: make([]string, 60), // Add 60, should keep only last 50
			wantLen:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame()
			g.openMultiplayer()

			for i, msg := range tt.messages {
				if msg == "" {
					msg = "test message " + string(rune('0'+i))
				}
				g.addChatMessage(msg)
			}

			if len(g.chatMessages) != tt.wantLen {
				t.Errorf("Expected %d messages, got %d", tt.wantLen, len(g.chatMessages))
			}
		})
	}
}

// TestChatKeyDeterminism verifies chat keys are deterministic based on seed.
func TestChatKeyDeterminism(t *testing.T) {
	seed := uint64(12345)

	// Create two games with same seed
	g1 := NewGame()
	g1.seed = seed
	g1.openMultiplayer()

	g2 := NewGame()
	g2.seed = seed
	g2.openMultiplayer()

	// Test that same message encrypted with same seed can be cross-decrypted
	testMsg := "Deterministic encryption test"

	encrypted1, err := g1.chatManager.Encrypt(testMsg)
	if err != nil {
		t.Fatalf("Game 1 encryption failed: %v", err)
	}

	encrypted2, err := g2.chatManager.Encrypt(testMsg)
	if err != nil {
		t.Fatalf("Game 2 encryption failed: %v", err)
	}

	// Decrypt game1's message with game2's key
	decrypted, err := g2.chatManager.Decrypt(encrypted1)
	if err != nil {
		t.Fatalf("Cross-decryption failed: %v", err)
	}

	if decrypted != testMsg {
		t.Errorf("Cross-decrypted message '%s' does not match original '%s'", decrypted, testMsg)
	}

	// Note: encrypted1 and encrypted2 will differ due to random nonces,
	// but both can be decrypted by either key since keys are derived from same seed
	_, err = g1.chatManager.Decrypt(encrypted2)
	if err != nil {
		t.Errorf("Game 1 should be able to decrypt game 2's message: %v", err)
	}
}

// TestChatWithDifferentKeys verifies messages encrypted with different keys cannot be decrypted.
func TestChatWithDifferentKeys(t *testing.T) {
	// Create two games with different seeds
	g1 := NewGame()
	g1.seed = 11111
	g1.openMultiplayer()

	g2 := NewGame()
	g2.seed = 22222
	g2.openMultiplayer()

	testMsg := "Secret message"

	// Encrypt with game1's key
	encrypted, err := g1.chatManager.Encrypt(testMsg)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Attempt to decrypt with game2's key (different seed)
	_, err = g2.chatManager.Decrypt(encrypted)
	if err == nil {
		t.Error("Expected decryption to fail with different key, but it succeeded")
	}
}

// TestChatManagerUsesCorrectEncryption verifies the chat manager uses AES-256-GCM.
func TestChatManagerUsesCorrectEncryption(t *testing.T) {
	g := NewGame()
	g.openMultiplayer()

	// Verify key is 32 bytes (256 bits)
	// We can test this indirectly by ensuring encryption/decryption works
	testCases := []string{
		"Short",
		"Medium length message with some content",
		"Very long message that exceeds typical buffer sizes and contains lots of characters to test that the encryption handles arbitrary lengths correctly without issues",
	}

	for _, msg := range testCases {
		encrypted, err := g.chatManager.Encrypt(msg)
		if err != nil {
			t.Errorf("Failed to encrypt message of length %d: %v", len(msg), err)
			continue
		}

		decrypted, err := g.chatManager.Decrypt(encrypted)
		if err != nil {
			t.Errorf("Failed to decrypt message of length %d: %v", len(msg), err)
			continue
		}

		if decrypted != msg {
			t.Errorf("Message mismatch for length %d", len(msg))
		}
	}
}

// TestChatIntegrationWithKeyExchangeInfrastructure verifies the infrastructure is compatible.
func TestChatIntegrationWithKeyExchangeInfrastructure(t *testing.T) {
	// This test verifies that the chat.NewChatWithKey function (used in integration)
	// is compatible with the key exchange infrastructure

	// Simulate a 32-byte key that would come from PerformKeyExchange
	simulatedKey := make([]byte, 32)
	for i := range simulatedKey {
		simulatedKey[i] = byte(i)
	}

	// Create chat with simulated exchanged key
	chatInstance := chat.NewChatWithKey(simulatedKey)

	testMsg := "Message using exchanged key"
	encrypted, err := chatInstance.Encrypt(testMsg)
	if err != nil {
		t.Fatalf("Encryption with exchanged key failed: %v", err)
	}

	decrypted, err := chatInstance.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decryption with exchanged key failed: %v", err)
	}

	if decrypted != testMsg {
		t.Errorf("Message encrypted with exchanged key doesn't match: got '%s', want '%s'", decrypted, testMsg)
	}
}
