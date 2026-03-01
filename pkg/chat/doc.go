// Package chat provides encrypted in-game chat with profanity filtering and relay server functionality.
//
// # Architecture
//
// The chat package implements a complete encrypted chat system with three main components:
//
//  1. Chat client with AES-256-GCM encryption for local message management
//  2. ECDH key exchange protocol for establishing shared encryption keys over network
//  3. Relay server that forwards encrypted messages without decryption capability
//
// # Encryption
//
// Messages are encrypted client-side using AES-256-GCM. The relay server never has
// access to decryption keys, ensuring end-to-end encryption. Key exchange uses
// ephemeral ECDH P-256 keypairs with HKDF-SHA3-256 for key derivation.
//
// # Profanity Filtering
//
// The package includes deterministic profanity filtering for multiple languages (en, es, de, fr, pt).
// Word lists are procedurally generated from seeds to avoid embedding large static dictionaries.
// Filtering is performed client-side and can be enabled/disabled per user preference.
//
// # Usage Example
//
//	// Create chat client with key exchange
//	conn, _ := net.Dial("tcp", "server:8080")
//	key, _ := chat.PerformKeyExchange(conn)
//	chatClient := chat.NewChatWithKey(key)
//
//	// Send encrypted message
//	encrypted, _ := chatClient.Encrypt("Hello, world!")
//	conn.Write([]byte(encrypted))
//
//	// Receive and decrypt
//	ciphertext := receiveFromNetwork()
//	plaintext, _ := chatClient.Decrypt(ciphertext)
//
//	// Apply profanity filter
//	filtered := chat.FilterProfanity(plaintext, true)
//
// # Relay Server
//
// The relay server forwards encrypted message blobs without storing plaintext:
//
//	server, _ := chat.NewRelayServer(":8080")
//	server.Start()
//	defer server.Stop()
//
// Clients connect via RelayClient, which handles message routing and encryption:
//
//	client, _ := chat.NewRelayClient("server:8080", "player123")
//	client.SendEncrypted("player456", encryptedMessage)
//	msg, _ := client.ReceiveEncrypted()
//
// # Thread Safety
//
// All exported types are safe for concurrent use. The Chat type uses sync.RWMutex
// for message queue protection. Global profanity word lists are protected with
// dedicated mutex synchronization.
package chat
