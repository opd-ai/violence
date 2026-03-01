package chat

import (
	"bytes"
	"crypto/rand"
	"io"
	"net"
	"testing"
	"time"
)

func TestPerformKeyExchange(t *testing.T) {
	// Create a pair of connected net.Pipe connections
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	// Channels to collect results
	clientKeyChan := make(chan []byte, 1)
	serverKeyChan := make(chan []byte, 1)
	clientErrChan := make(chan error, 1)
	serverErrChan := make(chan error, 1)

	// Perform key exchange from client side
	go func() {
		key, err := PerformKeyExchange(client)
		clientKeyChan <- key
		clientErrChan <- err
	}()

	// Perform key exchange from server side
	go func() {
		key, err := PerformKeyExchange(server)
		serverKeyChan <- key
		serverErrChan <- err
	}()

	// Wait for both sides to complete
	clientKey := <-clientKeyChan
	serverKey := <-serverKeyChan
	clientErr := <-clientErrChan
	serverErr := <-serverErrChan

	// Check for errors
	if clientErr != nil {
		t.Fatalf("client key exchange failed: %v", clientErr)
	}
	if serverErr != nil {
		t.Fatalf("server key exchange failed: %v", serverErr)
	}

	// Verify keys are non-nil and correct length
	if clientKey == nil {
		t.Fatal("client key is nil")
	}
	if serverKey == nil {
		t.Fatal("server key is nil")
	}
	if len(clientKey) != 32 {
		t.Errorf("client key length = %d, want 32", len(clientKey))
	}
	if len(serverKey) != 32 {
		t.Errorf("server key length = %d, want 32", len(serverKey))
	}

	// Verify both sides derived the same key
	if !bytes.Equal(clientKey, serverKey) {
		t.Error("client and server keys do not match")
	}
}

func TestPerformKeyExchangeNilConnection(t *testing.T) {
	_, err := PerformKeyExchange(nil)
	if err == nil {
		t.Error("expected error for nil connection, got nil")
	}
}

func TestPerformKeyExchangeBrokenConnection(t *testing.T) {
	client, server := net.Pipe()
	// Close server immediately to simulate broken connection
	server.Close()

	_, err := PerformKeyExchange(client)
	if err == nil {
		t.Error("expected error for broken connection, got nil")
	}
	client.Close()
}

func TestEncryptDecryptMessage(t *testing.T) {
	// Generate a random AES key
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{"empty", ""},
		{"short", "Hi"},
		{"normal", "Hello, this is a test message!"},
		{"long", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."},
		{"unicode", "Hello 世界! 🎮"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plaintext := []byte(tt.plaintext)

			// Encrypt
			ciphertext, err := EncryptMessage(key, plaintext)
			if err != nil {
				t.Fatalf("EncryptMessage failed: %v", err)
			}

			// Verify ciphertext is different from plaintext
			if len(plaintext) > 0 && bytes.Equal(ciphertext, plaintext) {
				t.Error("ciphertext equals plaintext")
			}

			// Decrypt
			decrypted, err := DecryptMessage(key, ciphertext)
			if err != nil {
				t.Fatalf("DecryptMessage failed: %v", err)
			}

			// Verify decrypted matches original
			if !bytes.Equal(decrypted, plaintext) {
				t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
			}
		})
	}
}

func TestDecryptMessageWrongKey(t *testing.T) {
	// Generate two different keys
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	io.ReadFull(rand.Reader, key1)
	io.ReadFull(rand.Reader, key2)

	plaintext := []byte("secret message")

	// Encrypt with key1
	ciphertext, err := EncryptMessage(key1, plaintext)
	if err != nil {
		t.Fatalf("EncryptMessage failed: %v", err)
	}

	// Try to decrypt with key2 (should fail)
	_, err = DecryptMessage(key2, ciphertext)
	if err == nil {
		t.Error("expected error when decrypting with wrong key, got nil")
	}
}

func TestDecryptMessageTamperedCiphertext(t *testing.T) {
	key := make([]byte, 32)
	io.ReadFull(rand.Reader, key)

	plaintext := []byte("authentic message")

	// Encrypt
	ciphertext, err := EncryptMessage(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptMessage failed: %v", err)
	}

	// Tamper with ciphertext
	if len(ciphertext) > 20 {
		ciphertext[20] ^= 0xFF
	}

	// Try to decrypt (should fail authentication)
	_, err = DecryptMessage(key, ciphertext)
	if err == nil {
		t.Error("expected error when decrypting tampered ciphertext, got nil")
	}
}

func TestDecryptMessageShortCiphertext(t *testing.T) {
	key := make([]byte, 32)
	io.ReadFull(rand.Reader, key)

	// Try to decrypt too-short ciphertext
	shortCiphertext := []byte{0x01, 0x02}
	_, err := DecryptMessage(key, shortCiphertext)
	if err == nil {
		t.Error("expected error for short ciphertext, got nil")
	}
}

func TestEncryptMessageInvalidKey(t *testing.T) {
	invalidKey := []byte{0x01, 0x02, 0x03} // Too short

	_, err := EncryptMessage(invalidKey, []byte("test"))
	if err == nil {
		t.Error("expected error for invalid key, got nil")
	}
}

func TestDeriveKey(t *testing.T) {
	secret := make([]byte, 32)
	io.ReadFull(rand.Reader, secret)

	// Derive key
	key := deriveKey(secret, 32)

	// Verify key length
	if len(key) != 32 {
		t.Errorf("key length = %d, want 32", len(key))
	}

	// Verify determinism (same secret produces same key)
	key2 := deriveKey(secret, 32)
	if !bytes.Equal(key, key2) {
		t.Error("deriveKey is not deterministic")
	}

	// Verify different secrets produce different keys
	secret2 := make([]byte, 32)
	io.ReadFull(rand.Reader, secret2)
	key3 := deriveKey(secret2, 32)
	if bytes.Equal(key, key3) {
		t.Error("different secrets produced same key")
	}
}

func TestEndToEndEncryption(t *testing.T) {
	// Create a pair of connected connections
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	// Perform key exchange
	clientKeyChan := make(chan []byte, 1)
	serverKeyChan := make(chan []byte, 1)

	go func() {
		key, _ := PerformKeyExchange(client)
		clientKeyChan <- key
	}()

	go func() {
		key, _ := PerformKeyExchange(server)
		serverKeyChan <- key
	}()

	clientKey := <-clientKeyChan
	serverKey := <-serverKeyChan

	// Test message encryption between peers
	message := []byte("Hello from client!")

	// Client encrypts
	encrypted, err := EncryptMessage(clientKey, message)
	if err != nil {
		t.Fatalf("client encrypt failed: %v", err)
	}

	// Server decrypts (using same derived key)
	decrypted, err := DecryptMessage(serverKey, encrypted)
	if err != nil {
		t.Fatalf("server decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted, message) {
		t.Errorf("decrypted = %q, want %q", decrypted, message)
	}

	// Test reverse direction
	response := []byte("Hello from server!")
	encrypted2, err := EncryptMessage(serverKey, response)
	if err != nil {
		t.Fatalf("server encrypt failed: %v", err)
	}

	decrypted2, err := DecryptMessage(clientKey, encrypted2)
	if err != nil {
		t.Fatalf("client decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted2, response) {
		t.Errorf("decrypted = %q, want %q", decrypted2, response)
	}
}

func TestSendReceivePublicKey(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	testKey := make([]byte, 65) // P-256 uncompressed public key
	io.ReadFull(rand.Reader, testKey)

	// Send from client
	go func() {
		sendPublicKey(client, testKey)
	}()

	// Receive on server
	receivedKey, err := receivePublicKey(server)
	if err != nil {
		t.Fatalf("receivePublicKey failed: %v", err)
	}

	if !bytes.Equal(receivedKey, testKey) {
		t.Error("received key does not match sent key")
	}
}

func TestReceivePublicKeyInvalidLength(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	// Send invalid length (0)
	go func() {
		lengthBuf := []byte{0x00, 0x00}
		client.Write(lengthBuf)
	}()

	_, err := receivePublicKey(server)
	if err == nil {
		t.Error("expected error for invalid length, got nil")
	}
}

func TestReceivePublicKeyTooLarge(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	// Send too large length
	go func() {
		lengthBuf := []byte{0xFF, 0xFF} // 65535 bytes
		client.Write(lengthBuf)
	}()

	_, err := receivePublicKey(server)
	if err == nil {
		t.Error("expected error for too large length, got nil")
	}
}

func TestKeyExchangeTimeout(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	// Set read deadline
	client.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	// Don't send anything from server (will timeout)
	_, err := PerformKeyExchange(client)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}
