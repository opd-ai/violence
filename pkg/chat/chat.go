// Package chat provides in-game chat with encryption support.
package chat

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Message represents a chat message.
type Message struct {
	From    string
	To      string
	Content string
	Time    int64
}

// Chat manages chat messaging.
type Chat struct {
	key      []byte
	messages []Message
	mu       sync.RWMutex
}

// NewChat creates a new chat instance.
func NewChat() *Chat {
	// Generate a random 32-byte AES-256 key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		// Fallback to deterministic key if random fails
		copy(key, []byte("violence-chat-encryption-key"))
	}
	return &Chat{
		key:      key,
		messages: make([]Message, 0),
	}
}

// NewChatWithKey creates a chat instance with a specific encryption key.
func NewChatWithKey(key []byte) *Chat {
	if len(key) != 32 {
		panic("chat encryption key must be 32 bytes")
	}
	return &Chat{
		key:      key,
		messages: make([]Message, 0),
	}
}

// Send transmits a chat message.
func (c *Chat) Send(message string) error {
	if message == "" {
		return fmt.Errorf("empty message")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.messages = append(c.messages, Message{
		Content: message,
		Time:    time.Now().Unix(),
	})
	return nil
}

// Receive returns the next pending chat message.
func (c *Chat) Receive() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.messages) == 0 {
		return "", nil
	}

	msg := c.messages[0]
	c.messages = c.messages[1:]
	return msg.Content, nil
}

// GetMessages returns all pending messages without removing them.
func (c *Chat) GetMessages() []Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]Message, len(c.messages))
	copy(result, c.messages)
	return result
}

// Clear removes all pending messages.
func (c *Chat) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = make([]Message, 0)
}

// Encrypt encrypts a message for transmission using AES-256-GCM.
func (c *Chat) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", fmt.Errorf("empty plaintext")
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a received message using AES-256-GCM.
func (c *Chat) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", fmt.Errorf("empty ciphertext")
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(data) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, cipherBytes := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// profanityWords contains commonly filtered words.
// Note: This is a minimal list. Production would use localized word lists.
var profanityWords = []string{
	"fuck", "shit", "damn", "ass", "bitch", "bastard",
	"crap", "piss", "cock", "dick", "pussy", "cunt",
	"fag", "retard", "nigger", "nigga", "kike", "spic",
}
var profanityMu sync.RWMutex

// FilterProfanity applies profanity masking to a message if enabled.
// Flagged words are replaced with asterisks of equal length.
func FilterProfanity(message string, filterEnabled bool) string {
	if !filterEnabled || message == "" {
		return message
	}

	result := message
	lower := strings.ToLower(message)

	profanityMu.RLock()
	words := make([]string, len(profanityWords))
	copy(words, profanityWords)
	profanityMu.RUnlock()

	for _, word := range words {
		// Find all occurrences (case-insensitive)
		for {
			idx := strings.Index(lower, word)
			if idx == -1 {
				break
			}

			// Replace with asterisks, preserving original length
			wordLen := len(word)
			mask := strings.Repeat("*", wordLen)

			// Preserve case boundaries by replacing in both strings
			result = result[:idx] + mask + result[idx+wordLen:]
			lower = lower[:idx] + mask + lower[idx+wordLen:]
		}
	}

	return result
}

// AddProfanityWord adds a custom word to the profanity filter list.
func AddProfanityWord(word string) {
	if word == "" {
		return
	}
	profanityMu.Lock()
	profanityWords = append(profanityWords, strings.ToLower(word))
	profanityMu.Unlock()
}

// ClearProfanityWords clears all profanity words (useful for testing).
func ClearProfanityWords() {
	profanityMu.Lock()
	profanityWords = []string{}
	profanityMu.Unlock()
}

// SetProfanityWords replaces the profanity word list.
func SetProfanityWords(words []string) {
	profanityMu.Lock()
	profanityWords = make([]string, len(words))
	for i, word := range words {
		profanityWords[i] = strings.ToLower(word)
	}
	profanityMu.Unlock()
}
