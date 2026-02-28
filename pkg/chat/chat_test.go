package chat

import (
	"strings"
	"testing"
)

func TestNewChat(t *testing.T) {
	chat := NewChat()
	if chat == nil {
		t.Fatal("NewChat returned nil")
	}
	if chat.key == nil {
		t.Fatal("encryption key not initialized")
	}
	if len(chat.key) != 32 {
		t.Fatalf("wrong key length: got %d, want 32", len(chat.key))
	}
}

func TestNewChatWithKey(t *testing.T) {
	key := make([]byte, 32)
	copy(key, []byte("test-key-32-bytes-long-enough!"))

	chat := NewChatWithKey(key)
	if chat == nil {
		t.Fatal("NewChatWithKey returned nil")
	}
	if len(chat.key) != 32 {
		t.Fatalf("wrong key length: got %d, want 32", len(chat.key))
	}
}

func TestNewChatWithKey_InvalidLength(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid key length")
		}
	}()
	NewChatWithKey([]byte("short"))
}

func TestChat_SendAndReceive(t *testing.T) {
	chat := NewChat()

	// Send message
	msg := "hello world"
	if err := chat.Send(msg); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	// Receive message
	received, err := chat.Receive()
	if err != nil {
		t.Fatalf("receive failed: %v", err)
	}
	if received != msg {
		t.Fatalf("message mismatch: got %q, want %q", received, msg)
	}
}

func TestChat_SendEmpty(t *testing.T) {
	chat := NewChat()
	err := chat.Send("")
	if err == nil {
		t.Fatal("expected error for empty message")
	}
}

func TestChat_ReceiveEmpty(t *testing.T) {
	chat := NewChat()
	msg, err := chat.Receive()
	if err != nil {
		t.Fatalf("receive failed: %v", err)
	}
	if msg != "" {
		t.Fatalf("expected empty message, got %q", msg)
	}
}

func TestChat_MultipleMessages(t *testing.T) {
	chat := NewChat()

	messages := []string{"msg1", "msg2", "msg3"}
	for _, msg := range messages {
		if err := chat.Send(msg); err != nil {
			t.Fatalf("send failed: %v", err)
		}
	}

	for i, want := range messages {
		got, err := chat.Receive()
		if err != nil {
			t.Fatalf("receive %d failed: %v", i, err)
		}
		if got != want {
			t.Fatalf("message %d mismatch: got %q, want %q", i, got, want)
		}
	}
}

func TestChat_GetMessages(t *testing.T) {
	chat := NewChat()

	chat.Send("msg1")
	chat.Send("msg2")

	msgs := chat.GetMessages()
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}

	// GetMessages should not remove messages
	msgs2 := chat.GetMessages()
	if len(msgs2) != 2 {
		t.Fatalf("messages were removed by GetMessages")
	}
}

func TestChat_Clear(t *testing.T) {
	chat := NewChat()

	chat.Send("msg1")
	chat.Send("msg2")

	chat.Clear()

	msgs := chat.GetMessages()
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages after clear, got %d", len(msgs))
	}
}

func TestChat_EncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	copy(key, []byte("test-encryption-key-32-bytes!!"))
	chat := NewChatWithKey(key)

	plaintext := "secret message"
	ciphertext, err := chat.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	if ciphertext == plaintext {
		t.Fatal("ciphertext should not equal plaintext")
	}

	decrypted, err := chat.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("decryption mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestChat_EncryptEmpty(t *testing.T) {
	chat := NewChat()
	_, err := chat.Encrypt("")
	if err == nil {
		t.Fatal("expected error for empty plaintext")
	}
}

func TestChat_DecryptEmpty(t *testing.T) {
	chat := NewChat()
	_, err := chat.Decrypt("")
	if err == nil {
		t.Fatal("expected error for empty ciphertext")
	}
}

func TestChat_DecryptInvalid(t *testing.T) {
	chat := NewChat()

	// Invalid base64
	_, err := chat.Decrypt("not-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}

	// Valid base64 but invalid ciphertext
	_, err = chat.Decrypt("YWJjZGVm")
	if err == nil {
		t.Fatal("expected error for invalid ciphertext")
	}
}

func TestChat_DecryptWrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	copy(key1, []byte("key1-32-bytes-long-enough!!!!!!"))
	copy(key2, []byte("key2-32-bytes-long-enough!!!!!!"))

	chat1 := NewChatWithKey(key1)
	chat2 := NewChatWithKey(key2)

	plaintext := "secret"
	ciphertext, err := chat1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Try to decrypt with wrong key
	_, err = chat2.Decrypt(ciphertext)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
}

func TestChat_EncryptLongMessage(t *testing.T) {
	chat := NewChat()

	// Test with long message
	longMsg := strings.Repeat("a", 10000)
	ciphertext, err := chat.Encrypt(longMsg)
	if err != nil {
		t.Fatalf("encrypt long message failed: %v", err)
	}

	decrypted, err := chat.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt long message failed: %v", err)
	}

	if decrypted != longMsg {
		t.Fatal("long message decryption mismatch")
	}
}

func TestSetGenre(t *testing.T) {
	// Should not panic
	SetGenre("fantasy")
	SetGenre("scifi")
	SetGenre("")
}
