package chat

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestRelayServerCreation tests relay server instantiation.
func TestRelayServerCreation(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "valid address",
			addr:    "127.0.0.1:0",
			wantErr: false,
		},
		{
			name:    "localhost",
			addr:    "localhost:0",
			wantErr: false,
		},
		{
			name:    "invalid address",
			addr:    "invalid:99999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs, err := NewRelayServer(tt.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRelayServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if rs != nil {
				rs.Stop()
			}
		})
	}
}

// TestRelayServerStartStop tests server lifecycle.
func TestRelayServerStartStop(t *testing.T) {
	rs, err := NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("NewRelayServer() failed: %v", err)
	}

	if err := rs.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Verify server is listening
	if rs.listener == nil {
		t.Error("listener is nil after Start()")
	}

	// Stop server
	if err := rs.Stop(); err != nil {
		t.Errorf("Stop() failed: %v", err)
	}

	// Verify client count is zero
	if count := rs.GetClientCount(); count != 0 {
		t.Errorf("GetClientCount() = %d, want 0 after Stop()", count)
	}
}

// TestRelayClientConnection tests client connection and disconnection.
func TestRelayClientConnection(t *testing.T) {
	rs, err := NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("NewRelayServer() failed: %v", err)
	}
	defer rs.Stop()

	if err := rs.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	addr := rs.listener.Addr().String()

	// Connect client
	client, err := NewRelayClient(addr, "player1")
	if err != nil {
		t.Fatalf("NewRelayClient() failed: %v", err)
	}
	defer client.Close()

	// Wait for connection to register
	time.Sleep(50 * time.Millisecond)

	// Verify client count
	if count := rs.GetClientCount(); count != 1 {
		t.Errorf("GetClientCount() = %d, want 1", count)
	}

	// Disconnect client
	client.Close()
	time.Sleep(50 * time.Millisecond)

	// Verify client count is zero
	if count := rs.GetClientCount(); count != 0 {
		t.Errorf("GetClientCount() = %d, want 0 after disconnect", count)
	}
}

// TestRelayEncryptedMessage tests encrypted message relay without decryption.
func TestRelayEncryptedMessage(t *testing.T) {
	rs, err := NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("NewRelayServer() failed: %v", err)
	}
	defer rs.Stop()

	if err := rs.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	addr := rs.listener.Addr().String()

	// Create two clients
	client1, err := NewRelayClient(addr, "player1")
	if err != nil {
		t.Fatalf("NewRelayClient(player1) failed: %v", err)
	}
	defer client1.Close()

	client2, err := NewRelayClient(addr, "player2")
	if err != nil {
		t.Fatalf("NewRelayClient(player2) failed: %v", err)
	}
	defer client2.Close()

	time.Sleep(50 * time.Millisecond)

	// Client 1 sends encrypted message to client 2
	// Server should relay without decryption
	ciphertext := "encrypted_blob_base64"
	if err := client1.SendEncrypted("player2", ciphertext); err != nil {
		t.Fatalf("SendEncrypted() failed: %v", err)
	}

	// Client 2 receives encrypted message
	msg, err := client2.ReceiveEncrypted()
	if err != nil {
		t.Fatalf("ReceiveEncrypted() failed: %v", err)
	}

	if msg == nil {
		t.Fatal("ReceiveEncrypted() returned nil message")
	}

	if msg.From != "player1" {
		t.Errorf("From = %s, want player1", msg.From)
	}

	if msg.Ciphertext != ciphertext {
		t.Errorf("Ciphertext = %s, want %s", msg.Ciphertext, ciphertext)
	}
}

// TestRelayBroadcastMessage tests broadcast to all clients.
func TestRelayBroadcastMessage(t *testing.T) {
	rs, err := NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("NewRelayServer() failed: %v", err)
	}
	defer rs.Stop()

	if err := rs.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	addr := rs.listener.Addr().String()

	// Create three clients
	clients := make([]*RelayClient, 3)
	for i := 0; i < 3; i++ {
		playerID := fmt.Sprintf("player%d", i+1)
		client, err := NewRelayClient(addr, playerID)
		if err != nil {
			t.Fatalf("NewRelayClient(%s) failed: %v", playerID, err)
		}
		defer client.Close()
		clients[i] = client
	}

	time.Sleep(50 * time.Millisecond)

	// Client 1 broadcasts message
	ciphertext := "broadcast_encrypted_blob"
	if err := clients[0].SendEncrypted("all", ciphertext); err != nil {
		t.Fatalf("SendEncrypted() failed: %v", err)
	}

	// Clients 2 and 3 should receive message
	for i := 1; i < 3; i++ {
		msg, err := clients[i].ReceiveEncrypted()
		if err != nil {
			t.Fatalf("ReceiveEncrypted() for client %d failed: %v", i+1, err)
		}

		if msg == nil {
			t.Fatalf("ReceiveEncrypted() for client %d returned nil", i+1)
		}

		if msg.From != "player1" {
			t.Errorf("From = %s, want player1", msg.From)
		}

		if msg.Ciphertext != ciphertext {
			t.Errorf("Ciphertext = %s, want %s", msg.Ciphertext, ciphertext)
		}
	}

	// Client 1 should NOT receive their own broadcast
	msg, err := clients[0].ReceiveEncrypted()
	if err != nil {
		t.Fatalf("ReceiveEncrypted() for sender failed: %v", err)
	}
	if msg != nil {
		t.Error("Sender received their own broadcast message")
	}
}

// TestRelayServerNoPlaintextStorage tests that server never stores plaintext.
func TestRelayServerNoPlaintextStorage(t *testing.T) {
	rs, err := NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("NewRelayServer() failed: %v", err)
	}
	defer rs.Stop()

	if err := rs.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	addr := rs.listener.Addr().String()

	// Create encryption chat instances for two clients
	key := make([]byte, 32)
	copy(key, []byte("shared-encryption-key-32-bytes!"))

	chat1 := NewChatWithKey(key)
	chat2 := NewChatWithKey(key)

	// Create relay clients
	client1, err := NewRelayClient(addr, "player1")
	if err != nil {
		t.Fatalf("NewRelayClient(player1) failed: %v", err)
	}
	defer client1.Close()

	client2, err := NewRelayClient(addr, "player2")
	if err != nil {
		t.Fatalf("NewRelayClient(player2) failed: %v", err)
	}
	defer client2.Close()

	time.Sleep(50 * time.Millisecond)

	// Client 1 encrypts and sends message
	plaintext := "secret message"
	ciphertext, err := chat1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() failed: %v", err)
	}

	if err := client1.SendEncrypted("player2", ciphertext); err != nil {
		t.Fatalf("SendEncrypted() failed: %v", err)
	}

	// Client 2 receives encrypted blob
	msg, err := client2.ReceiveEncrypted()
	if err != nil {
		t.Fatalf("ReceiveEncrypted() failed: %v", err)
	}

	if msg == nil {
		t.Fatal("ReceiveEncrypted() returned nil")
	}

	// Client 2 decrypts the message
	decrypted, err := chat2.Decrypt(msg.Ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypted message = %s, want %s", decrypted, plaintext)
	}

	// Verify server never had plaintext
	// The ciphertext should not contain the plaintext
	if strings.Contains(msg.Ciphertext, plaintext) {
		t.Error("Server relayed plaintext; encryption failed")
	}
}

// TestRelayParseMessage tests message parsing.
func TestRelayParseMessage(t *testing.T) {
	rs, err := NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("NewRelayServer() failed: %v", err)
	}
	defer rs.Stop()

	tests := []struct {
		name       string
		from       string
		data       string
		wantTo     string
		wantCipher string
		wantNil    bool
	}{
		{
			name:       "valid message",
			from:       "player1",
			data:       "player2|encrypted_blob",
			wantTo:     "player2",
			wantCipher: "encrypted_blob",
			wantNil:    false,
		},
		{
			name:       "broadcast message",
			from:       "player1",
			data:       "all|encrypted_blob",
			wantTo:     "all",
			wantCipher: "encrypted_blob",
			wantNil:    false,
		},
		{
			name:    "invalid format",
			from:    "player1",
			data:    "invalid",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := rs.parseMessage(tt.from, tt.data)

			if tt.wantNil {
				if msg != nil {
					t.Error("parseMessage() should return nil for invalid format")
				}
				return
			}

			if msg == nil {
				t.Fatal("parseMessage() returned nil for valid format")
			}

			if msg.From != tt.from {
				t.Errorf("From = %s, want %s", msg.From, tt.from)
			}

			if msg.To != tt.wantTo {
				t.Errorf("To = %s, want %s", msg.To, tt.wantTo)
			}

			if msg.Ciphertext != tt.wantCipher {
				t.Errorf("Ciphertext = %s, want %s", msg.Ciphertext, tt.wantCipher)
			}

			if msg.Timestamp == 0 {
				t.Error("Timestamp should be set")
			}
		})
	}
}

// TestRelayMultipleMessages tests sending multiple messages.
func TestRelayMultipleMessages(t *testing.T) {
	rs, err := NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("NewRelayServer() failed: %v", err)
	}
	defer rs.Stop()

	if err := rs.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	addr := rs.listener.Addr().String()

	// Create two clients
	client1, err := NewRelayClient(addr, "player1")
	if err != nil {
		t.Fatalf("NewRelayClient(player1) failed: %v", err)
	}
	defer client1.Close()

	client2, err := NewRelayClient(addr, "player2")
	if err != nil {
		t.Fatalf("NewRelayClient(player2) failed: %v", err)
	}
	defer client2.Close()

	time.Sleep(50 * time.Millisecond)

	// Send multiple messages
	messageCount := 5
	for i := 0; i < messageCount; i++ {
		ciphertext := fmt.Sprintf("message_%d", i)
		if err := client1.SendEncrypted("player2", ciphertext); err != nil {
			t.Fatalf("SendEncrypted() message %d failed: %v", i, err)
		}
	}

	// Receive all messages
	for i := 0; i < messageCount; i++ {
		msg, err := client2.ReceiveEncrypted()
		if err != nil {
			t.Fatalf("ReceiveEncrypted() message %d failed: %v", i, err)
		}

		if msg == nil {
			t.Fatalf("ReceiveEncrypted() message %d returned nil", i)
		}

		expectedCipher := fmt.Sprintf("message_%d", i)
		if msg.Ciphertext != expectedCipher {
			t.Errorf("Message %d ciphertext = %s, want %s", i, msg.Ciphertext, expectedCipher)
		}
	}
}

// TestRelayClientReceiveTimeout tests receive timeout behavior.
func TestRelayClientReceiveTimeout(t *testing.T) {
	rs, err := NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("NewRelayServer() failed: %v", err)
	}
	defer rs.Stop()

	if err := rs.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	addr := rs.listener.Addr().String()

	client, err := NewRelayClient(addr, "player1")
	if err != nil {
		t.Fatalf("NewRelayClient() failed: %v", err)
	}
	defer client.Close()

	time.Sleep(50 * time.Millisecond)

	// Receive with no messages should timeout and return nil
	msg, err := client.ReceiveEncrypted()
	if err != nil {
		t.Fatalf("ReceiveEncrypted() failed: %v", err)
	}

	if msg != nil {
		t.Error("ReceiveEncrypted() should return nil when no messages available")
	}
}

// TestRelayEndToEndWithEncryption tests full E2E encrypted chat flow.
func TestRelayEndToEndWithEncryption(t *testing.T) {
	rs, err := NewRelayServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("NewRelayServer() failed: %v", err)
	}
	defer rs.Stop()

	if err := rs.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	addr := rs.listener.Addr().String()

	// Shared encryption key (in real impl, would use key exchange)
	key := make([]byte, 32)
	copy(key, []byte("violence-shared-key-32-bytes!!!"))

	// Create chat instances with shared key
	chat1 := NewChatWithKey(key)
	chat2 := NewChatWithKey(key)

	// Create relay clients
	client1, err := NewRelayClient(addr, "alice")
	if err != nil {
		t.Fatalf("NewRelayClient(alice) failed: %v", err)
	}
	defer client1.Close()

	client2, err := NewRelayClient(addr, "bob")
	if err != nil {
		t.Fatalf("NewRelayClient(bob) failed: %v", err)
	}
	defer client2.Close()

	time.Sleep(50 * time.Millisecond)

	// Alice sends encrypted message to Bob
	plaintext := "Hello Bob, this is a secret!"
	ciphertext, err := chat1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() failed: %v", err)
	}

	if err := client1.SendEncrypted("bob", ciphertext); err != nil {
		t.Fatalf("SendEncrypted() failed: %v", err)
	}

	// Bob receives and decrypts
	msg, err := client2.ReceiveEncrypted()
	if err != nil {
		t.Fatalf("ReceiveEncrypted() failed: %v", err)
	}

	if msg == nil {
		t.Fatal("ReceiveEncrypted() returned nil")
	}

	decrypted, err := chat2.Decrypt(msg.Ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypted = %s, want %s", decrypted, plaintext)
	}

	// Bob sends reply to Alice
	replyPlaintext := "Hi Alice, received your secret!"
	replyCiphertext, err := chat2.Encrypt(replyPlaintext)
	if err != nil {
		t.Fatalf("Encrypt() reply failed: %v", err)
	}

	if err := client2.SendEncrypted("alice", replyCiphertext); err != nil {
		t.Fatalf("SendEncrypted() reply failed: %v", err)
	}

	// Alice receives and decrypts reply
	replyMsg, err := client1.ReceiveEncrypted()
	if err != nil {
		t.Fatalf("ReceiveEncrypted() reply failed: %v", err)
	}

	if replyMsg == nil {
		t.Fatal("ReceiveEncrypted() reply returned nil")
	}

	decryptedReply, err := chat1.Decrypt(replyMsg.Ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() reply failed: %v", err)
	}

	if decryptedReply != replyPlaintext {
		t.Errorf("Decrypted reply = %s, want %s", decryptedReply, replyPlaintext)
	}
}
