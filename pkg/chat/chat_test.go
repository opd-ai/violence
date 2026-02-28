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

func TestFilterProfanity_Disabled(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{"simple", "fuck this shit", "fuck this shit"},
		{"mixed", "what the damn hell", "what the damn hell"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterProfanity(tt.message, false)
			if result != tt.expected {
				t.Errorf("FilterProfanity(%q, false) = %q, want %q", tt.message, result, tt.expected)
			}
		})
	}
}

func TestFilterProfanity_Enabled(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{"single_word", "fuck", "****"},
		{"multiple_words", "fuck this shit", "**** this ****"},
		{"mixed_case", "FUCK this SHIT", "**** this ****"},
		{"at_start", "shit happens", "**** happens"},
		{"at_end", "holy shit", "holy ****"},
		{"in_middle", "what the fuck man", "what the **** man"},
		{"multiple_same", "fuck fuck fuck", "**** **** ****"},
		{"partial_match", "shifts", "shifts"}, // Should not match "shit"
		{"empty", "", ""},
		{"clean_message", "hello world", "hello world"},
		{"all_profanity", "fuck shit damn", "**** **** ****"},
		{"with_punctuation", "fuck! shit?", "****! ****?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterProfanity(tt.message, true)
			if result != tt.expected {
				t.Errorf("FilterProfanity(%q, true) = %q, want %q", tt.message, result, tt.expected)
			}
		})
	}
}

func TestAddProfanityWord(t *testing.T) {
	// Save original list
	original := make([]string, len(profanityWords))
	copy(original, profanityWords)
	defer func() {
		SetProfanityWords(original)
	}()

	// Add custom word
	AddProfanityWord("badword")

	result := FilterProfanity("this is badword test", true)
	expected := "this is ******* test"
	if result != expected {
		t.Errorf("FilterProfanity after AddProfanityWord = %q, want %q", result, expected)
	}
}

func TestAddProfanityWord_Empty(t *testing.T) {
	original := len(profanityWords)
	AddProfanityWord("")
	if len(profanityWords) != original {
		t.Error("AddProfanityWord should not add empty string")
	}
}

func TestClearProfanityWords(t *testing.T) {
	// Save original
	original := make([]string, len(profanityWords))
	copy(original, profanityWords)
	defer func() {
		SetProfanityWords(original)
	}()

	ClearProfanityWords()
	if len(profanityWords) != 0 {
		t.Errorf("ClearProfanityWords: got %d words, want 0", len(profanityWords))
	}

	// No words should be filtered
	result := FilterProfanity("fuck shit damn", true)
	if result != "fuck shit damn" {
		t.Errorf("FilterProfanity after clear = %q, want unchanged", result)
	}
}

func TestSetProfanityWords(t *testing.T) {
	// Save original
	original := make([]string, len(profanityWords))
	copy(original, profanityWords)
	defer func() {
		SetProfanityWords(original)
	}()

	// Set custom list
	custom := []string{"badword", "wrongword"}
	SetProfanityWords(custom)

	if len(profanityWords) != 2 {
		t.Errorf("SetProfanityWords: got %d words, want 2", len(profanityWords))
	}

	// Only custom words should be filtered
	result := FilterProfanity("fuck badword wrongword", true)
	expected := "fuck ******* *********"
	if result != expected {
		t.Errorf("FilterProfanity with custom words = %q, want %q", result, expected)
	}
}

func TestFilterProfanity_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{"uppercase", "FUCK", "****"},
		{"lowercase", "fuck", "****"},
		{"mixed_case", "FuCk", "****"},
		{"sentence_case", "Fuck this", "**** this"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterProfanity(tt.message, true)
			if result != tt.expected {
				t.Errorf("FilterProfanity(%q, true) = %q, want %q", tt.message, result, tt.expected)
			}
		})
	}
}

func TestFilterProfanity_PreservesLength(t *testing.T) {
	tests := []string{
		"fuck",
		"shit",
		"damn",
		"bastard",
	}

	for _, word := range tests {
		t.Run(word, func(t *testing.T) {
			result := FilterProfanity(word, true)
			if len(result) != len(word) {
				t.Errorf("FilterProfanity(%q) length = %d, want %d", word, len(result), len(word))
			}
			if !strings.Contains(result, "*") {
				t.Errorf("FilterProfanity(%q) = %q, should contain asterisks", word, result)
			}
		})
	}
}

func TestFilterProfanity_EncryptionRoundTrip(t *testing.T) {
	// Test that filtering works correctly with encryption/decryption
	key := make([]byte, 32)
	copy(key, []byte("test-filter-key-32-bytes-long!"))
	chat := NewChatWithKey(key)

	// Send profane message, encrypt it
	originalMsg := "this is some fuck shit"
	encrypted, err := chat.Encrypt(originalMsg)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Decrypt message
	decrypted, err := chat.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	// Apply filter client-side after decryption
	filtered := FilterProfanity(decrypted, true)
	expected := "this is some **** ****"
	if filtered != expected {
		t.Errorf("FilterProfanity after decryption = %q, want %q", filtered, expected)
	}
}

func TestFilterProfanity_RelayIntegration(t *testing.T) {
	// Test that profanity filter works with relay server
	// Server relays encrypted blobs; client-side filtering happens after decryption
	key := make([]byte, 32)
	copy(key, []byte("shared-key-32-bytes-long-here!"))

	// Simulate sender encrypting profane message
	chat1 := NewChatWithKey(key)
	profaneMsg := "what the fuck is this shit"
	encrypted, err := chat1.Encrypt(profaneMsg)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Simulate receiver decrypting and filtering
	chat2 := NewChatWithKey(key)
	decrypted, err := chat2.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	// Receiver applies filter
	filtered := FilterProfanity(decrypted, true)
	expected := "what the **** is this ****"
	if filtered != expected {
		t.Errorf("FilterProfanity = %q, want %q", filtered, expected)
	}

	// Verify original message unchanged (server sees encrypted blob only)
	if decrypted != profaneMsg {
		t.Errorf("decrypted message changed: got %q, want %q", decrypted, profaneMsg)
	}
}

func TestFilterProfanity_MultipleOccurrences(t *testing.T) {
	// Test that all occurrences are filtered
	msg := "fuck you and fuck that and fuck this"
	result := FilterProfanity(msg, true)
	expected := "**** you and **** that and **** this"
	if result != expected {
		t.Errorf("FilterProfanity(%q) = %q, want %q", msg, result, expected)
	}

	// Count asterisks
	asteriskCount := strings.Count(result, "*")
	if asteriskCount != 12 { // 3 occurrences of "fuck" (4 chars each)
		t.Errorf("expected 12 asterisks, got %d", asteriskCount)
	}
}

func TestFilterProfanity_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		enabled  bool
		expected string
	}{
		{"word_boundaries", "shift", true, "shift"},               // Should not match "shit"
		{"substring_ass_match", "assessment", true, "***essment"}, // Matches "ass"
		{"prefix_ass_match", "bassoon", true, "b***oon"},          // Matches "ass"
		{"unicode", "hello 世界 fuck", true, "hello 世界 ****"},
		{"numbers_attached", "fuck123", true, "****123"},  // Matches "fuck"
		{"special_chars", "f*ck", true, "f*ck"},           // Only exact match
		{"repeated_filter", "fuckfuck", true, "********"}, // Two occurrences
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterProfanity(tt.message, tt.enabled)
			if result != tt.expected {
				t.Errorf("FilterProfanity(%q, %v) = %q, want %q", tt.message, tt.enabled, result, tt.expected)
			}
		})
	}
}
