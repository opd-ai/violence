package main

import (
	"net"
	"testing"

	"github.com/opd-ai/violence/pkg/config"
)

// TestInitializeEncryptedChat_LocalMode tests seed-based key for local/single-player.
func TestInitializeEncryptedChat_LocalMode(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		name        string
		networkMode bool
		networkConn net.Conn
		seed        uint64
		wantSeedKey bool
	}{
		{
			name:        "single-player mode",
			networkMode: false,
			networkConn: nil,
			seed:        12345,
			wantSeedKey: true,
		},
		{
			name:        "network mode without connection",
			networkMode: true,
			networkConn: nil,
			seed:        67890,
			wantSeedKey: true,
		},
		{
			name:        "local mode with seed",
			networkMode: false,
			networkConn: nil,
			seed:        0xDEADBEEF,
			wantSeedKey: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame()
			g.networkMode = tt.networkMode
			g.networkConn = tt.networkConn
			g.seed = tt.seed

			g.initializeEncryptedChat()

			if g.chatManager == nil {
				t.Fatal("Chat manager not initialized")
			}

			// Verify deterministic behavior: same seed = same key
			if tt.wantSeedKey {
				g2 := NewGame()
				g2.seed = tt.seed
				g2.networkMode = false
				g2.initializeEncryptedChat()

				// Both should encrypt/decrypt consistently
				msg := "Test message"
				enc1, _ := g.chatManager.Encrypt(msg)
				dec2, err := g2.chatManager.Decrypt(enc1)
				if err != nil {
					t.Errorf("Failed to decrypt across same seed: %v", err)
				}
				if dec2 != msg {
					t.Errorf("Decrypted = %q, want %q", dec2, msg)
				}
			}
		})
	}
}

// TestInitializeEncryptedChat_KeyExchangeFailure tests fallback on network failure.
func TestInitializeEncryptedChat_KeyExchangeFailure(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create a connection that will fail during key exchange
	server, client := net.Pipe()
	server.Close() // Close server immediately to cause failure
	defer client.Close()

	g := NewGame()
	g.networkMode = true
	g.networkConn = client
	g.seed = 0xCAFEBABE

	// Should fallback to seed-based key
	g.initializeEncryptedChat()

	if g.chatManager == nil {
		t.Fatal("Chat manager not initialized after key exchange failure")
	}

	// Verify it's using seed-based key by checking determinism
	g2 := NewGame()
	g2.seed = 0xCAFEBABE
	g2.networkMode = false
	g2.initializeEncryptedChat()

	msg := "Fallback test"
	enc, _ := g.chatManager.Encrypt(msg)
	dec, err := g2.chatManager.Decrypt(enc)
	if err != nil {
		t.Errorf("Seed-based fallback failed: %v", err)
	}
	if dec != msg {
		t.Errorf("Decrypted = %q, want %q", dec, msg)
	}
}

// TestDeriveSeedKey verifies seed key derivation.
func TestDeriveSeedKey(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		name string
		seed uint64
	}{
		{"zero seed", 0},
		{"small seed", 42},
		{"large seed", 0xFFFFFFFFFFFFFFFF},
		{"specific seed", 0x123456789ABCDEF0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame()
			g.seed = tt.seed

			key := g.deriveSeedKey()

			// Verify key length
			if len(key) != 32 {
				t.Errorf("Key length = %d, want 32", len(key))
			}

			// Verify determinism
			key2 := g.deriveSeedKey()
			for i := 0; i < 32; i++ {
				if key[i] != key2[i] {
					t.Errorf("Key not deterministic at byte %d", i)
				}
			}

			// Verify different seeds produce different keys
			g.seed = tt.seed + 1
			key3 := g.deriveSeedKey()
			same := true
			for i := 0; i < 32; i++ {
				if key[i] != key3[i] {
					same = false
					break
				}
			}
			if same && tt.seed != 0xFFFFFFFFFFFFFFFF {
				t.Error("Different seeds produced identical keys")
			}
		})
	}
}

// TestChatEncryption_EndToEnd tests complete encryption workflow.
func TestChatEncryption_EndToEnd(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	g := NewGame()
	g.seed = 0xDEADBEEF
	g.initializeEncryptedChat()

	messages := []string{
		"Hello, world!",
		"Special chars: @#$%^&*()",
		"Unicode: こんにちは 🎮",
		"Long message: " + string(make([]byte, 500)),
		"",
	}

	for _, msg := range messages {
		if msg == "" {
			// Empty messages should fail encryption
			_, err := g.chatManager.Encrypt(msg)
			if err == nil {
				t.Error("Expected error for empty message")
			}
			continue
		}

		encrypted, err := g.chatManager.Encrypt(msg)
		if err != nil {
			t.Errorf("Failed to encrypt %q: %v", msg[:min(len(msg), 20)], err)
			continue
		}

		if encrypted == msg {
			t.Error("Encrypted message equals plaintext")
		}

		decrypted, err := g.chatManager.Decrypt(encrypted)
		if err != nil {
			t.Errorf("Failed to decrypt: %v", err)
			continue
		}

		if decrypted != msg {
			t.Errorf("Decrypted = %q, want %q", decrypted, msg)
		}
	}
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
