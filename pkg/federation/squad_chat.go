package federation

import (
	"crypto/rand"
	"fmt"
	"sync"

	"github.com/opd-ai/violence/pkg/chat"
)

// SquadChatChannel represents a dedicated encrypted chat channel for squad members.
// Uses a shared squad encryption key visible only to squad members.
type SquadChatChannel struct {
	squadID     string
	encryptKey  []byte // Shared 32-byte AES-256 key for squad chat
	relayClient *chat.RelayClient
	chat        *chat.Chat
	messages    []chat.Message
	mu          sync.RWMutex
}

// NewSquadChatChannel creates a dedicated chat channel for a squad.
// Generates a shared encryption key that all squad members will use.
func NewSquadChatChannel(squadID, relayAddr, playerID string) (*SquadChatChannel, error) {
	// Generate shared squad encryption key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate squad encryption key: %w", err)
	}

	// Create relay client for this squad channel
	// Use squad ID as channel identifier prefix
	channelID := fmt.Sprintf("squad-%s-%s", squadID, playerID)
	relayClient, err := chat.NewRelayClient(relayAddr, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to create relay client: %w", err)
	}

	return &SquadChatChannel{
		squadID:     squadID,
		encryptKey:  key,
		relayClient: relayClient,
		chat:        chat.NewChatWithKey(key),
		messages:    make([]chat.Message, 0),
	}, nil
}

// NewSquadChatChannelWithKey creates a chat channel with an existing squad key.
// Used when a player joins an existing squad and receives the shared key.
func NewSquadChatChannelWithKey(squadID string, key []byte, relayAddr, playerID string) (*SquadChatChannel, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("squad encryption key must be 32 bytes, got %d", len(key))
	}

	// Create relay client
	channelID := fmt.Sprintf("squad-%s-%s", squadID, playerID)
	relayClient, err := chat.NewRelayClient(relayAddr, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to create relay client: %w", err)
	}

	return &SquadChatChannel{
		squadID:     squadID,
		encryptKey:  key,
		relayClient: relayClient,
		chat:        chat.NewChatWithKey(key),
		messages:    make([]chat.Message, 0),
	}, nil
}

// SendMessage encrypts and sends a message to all squad members.
func (scc *SquadChatChannel) SendMessage(from, message string) error {
	if message == "" {
		return fmt.Errorf("empty message")
	}

	// Encrypt message using shared squad key
	ciphertext, err := scc.chat.Encrypt(message)
	if err != nil {
		return fmt.Errorf("failed to encrypt message: %w", err)
	}

	// Broadcast to all clients - only squad members with the shared key can decrypt
	if err := scc.relayClient.SendEncrypted("all", ciphertext); err != nil {
		return fmt.Errorf("failed to send encrypted message: %w", err)
	}

	// Store message locally
	scc.mu.Lock()
	scc.messages = append(scc.messages, chat.Message{
		From:    from,
		To:      scc.squadID,
		Content: message,
		Time:    0,
	})
	scc.mu.Unlock()

	return nil
}

// ReceiveMessages polls for new encrypted messages and decrypts them.
// Returns all newly received messages since last call.
func (scc *SquadChatChannel) ReceiveMessages() ([]chat.Message, error) {
	newMessages := make([]chat.Message, 0)

	// Poll for encrypted messages
	for {
		encMsg, err := scc.relayClient.ReceiveEncrypted()
		if err != nil {
			return nil, fmt.Errorf("failed to receive encrypted message: %w", err)
		}
		if encMsg == nil {
			break // No more messages
		}

		// Decrypt using shared squad key
		plaintext, err := scc.chat.Decrypt(encMsg.Ciphertext)
		if err != nil {
			// Skip messages we can't decrypt (wrong key or corrupted)
			continue
		}

		msg := chat.Message{
			From:    encMsg.From,
			To:      scc.squadID,
			Content: plaintext,
			Time:    encMsg.Timestamp,
		}

		scc.mu.Lock()
		scc.messages = append(scc.messages, msg)
		scc.mu.Unlock()

		newMessages = append(newMessages, msg)
	}

	return newMessages, nil
}

// GetMessages returns all squad chat messages.
func (scc *SquadChatChannel) GetMessages() []chat.Message {
	scc.mu.RLock()
	defer scc.mu.RUnlock()

	result := make([]chat.Message, len(scc.messages))
	copy(result, scc.messages)
	return result
}

// GetEncryptionKey returns the shared squad encryption key.
// Used when inviting new members to share the key.
func (scc *SquadChatChannel) GetEncryptionKey() []byte {
	key := make([]byte, len(scc.encryptKey))
	copy(key, scc.encryptKey)
	return key
}

// GetSquadID returns the squad ID for this channel.
func (scc *SquadChatChannel) GetSquadID() string {
	return scc.squadID
}

// ClearMessages removes all stored messages.
func (scc *SquadChatChannel) ClearMessages() {
	scc.mu.Lock()
	defer scc.mu.Unlock()
	scc.messages = make([]chat.Message, 0)
}

// Close disconnects from the relay server.
func (scc *SquadChatChannel) Close() error {
	if scc.relayClient != nil {
		return scc.relayClient.Close()
	}
	return nil
}

// SquadChatManager manages chat channels for multiple squads.
type SquadChatManager struct {
	channels map[string]*SquadChatChannel // squadID -> channel
	mu       sync.RWMutex
}

// NewSquadChatManager creates a new squad chat manager.
func NewSquadChatManager() *SquadChatManager {
	return &SquadChatManager{
		channels: make(map[string]*SquadChatChannel),
	}
}

// CreateChannel creates a new chat channel for a squad.
func (scm *SquadChatManager) CreateChannel(squadID, relayAddr, playerID string) (*SquadChatChannel, error) {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	if _, exists := scm.channels[squadID]; exists {
		return nil, fmt.Errorf("chat channel already exists for squad %s", squadID)
	}

	channel, err := NewSquadChatChannel(squadID, relayAddr, playerID)
	if err != nil {
		return nil, err
	}

	scm.channels[squadID] = channel
	return channel, nil
}

// JoinChannel joins an existing squad chat channel with the shared key.
func (scm *SquadChatManager) JoinChannel(squadID string, key []byte, relayAddr, playerID string) (*SquadChatChannel, error) {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	if _, exists := scm.channels[squadID]; exists {
		return nil, fmt.Errorf("already joined chat channel for squad %s", squadID)
	}

	channel, err := NewSquadChatChannelWithKey(squadID, key, relayAddr, playerID)
	if err != nil {
		return nil, err
	}

	scm.channels[squadID] = channel
	return channel, nil
}

// GetChannel retrieves a squad's chat channel.
func (scm *SquadChatManager) GetChannel(squadID string) (*SquadChatChannel, error) {
	scm.mu.RLock()
	defer scm.mu.RUnlock()

	channel, exists := scm.channels[squadID]
	if !exists {
		return nil, fmt.Errorf("no chat channel for squad %s", squadID)
	}

	return channel, nil
}

// RemoveChannel removes a squad's chat channel and closes the connection.
func (scm *SquadChatManager) RemoveChannel(squadID string) error {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	channel, exists := scm.channels[squadID]
	if !exists {
		return fmt.Errorf("no chat channel for squad %s", squadID)
	}

	if err := channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}

	delete(scm.channels, squadID)
	return nil
}

// CloseAll closes all squad chat channels.
func (scm *SquadChatManager) CloseAll() error {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	for squadID, channel := range scm.channels {
		if err := channel.Close(); err != nil {
			return fmt.Errorf("failed to close channel for squad %s: %w", squadID, err)
		}
	}

	scm.channels = make(map[string]*SquadChatChannel)
	return nil
}
