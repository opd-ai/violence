package chat

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// EncryptedMessage represents an encrypted chat message blob.
// Server relays these without decryption capability.
type EncryptedMessage struct {
	From       string // Sender player ID
	To         string // Recipient player ID or "all" for broadcast
	Ciphertext string // Base64-encoded encrypted message
	Timestamp  int64  // Server timestamp
}

// RelayServer relays encrypted chat messages without plaintext storage.
// Messages are encrypted client-side; server has no decryption keys.
type RelayServer struct {
	listener       net.Listener
	clients        map[string]net.Conn // playerID -> connection
	messages       chan EncryptedMessage
	done           chan struct{}
	mu             sync.RWMutex
	logger         *logrus.Entry
	readTimeout    time.Duration
	messageTimeout time.Duration
}

// NewRelayServer creates a chat relay server.
func NewRelayServer(addr string) (*RelayServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	return &RelayServer{
		listener:       listener,
		clients:        make(map[string]net.Conn),
		messages:       make(chan EncryptedMessage, 100),
		done:           make(chan struct{}),
		readTimeout:    30 * time.Second,
		messageTimeout: 100 * time.Millisecond,
		logger: logrus.WithFields(logrus.Fields{
			"system": "chat_relay",
		}),
	}, nil
}

// Start begins accepting connections and relaying messages.
func (rs *RelayServer) Start() error {
	go rs.acceptConnections()
	go rs.relayMessages()
	return nil
}

// acceptConnections handles new client connections.
func (rs *RelayServer) acceptConnections() {
	for {
		select {
		case <-rs.done:
			return
		default:
		}

		conn, err := rs.listener.Accept()
		if err != nil {
			select {
			case <-rs.done:
				return
			default:
				rs.logger.WithError(err).Error("failed to accept connection")
				continue
			}
		}

		go rs.handleClient(conn)
	}
}

// handleClient processes messages from a connected client.
func (rs *RelayServer) handleClient(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Read player ID (first line)
	playerID, err := reader.ReadString('\n')
	if err != nil {
		rs.logger.WithError(err).Error("failed to read player ID")
		return
	}
	playerID = strings.TrimSpace(playerID)

	rs.mu.Lock()
	rs.clients[playerID] = conn
	rs.mu.Unlock()

	rs.logger.WithField("player_id", playerID).Info("client connected")

	defer func() {
		rs.mu.Lock()
		delete(rs.clients, playerID)
		rs.mu.Unlock()
		rs.logger.WithField("player_id", playerID).Info("client disconnected")
	}()

	// Read messages line by line
	for {
		select {
		case <-rs.done:
			return
		default:
		}

		// Set read deadline to detect disconnects
		conn.SetReadDeadline(time.Now().Add(rs.readTimeout))

		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		// Parse encrypted message
		data := strings.TrimSpace(line)
		msg := rs.parseMessage(playerID, data)
		if msg != nil {
			rs.messages <- *msg
		}
	}
}

// parseMessage parses incoming message data.
func (rs *RelayServer) parseMessage(from, data string) *EncryptedMessage {
	// Simple format: "to|ciphertext"
	parts := strings.SplitN(data, "|", 2)
	if len(parts) != 2 {
		rs.logger.WithField("data", data).Error("invalid message format")
		return nil
	}

	return &EncryptedMessage{
		From:       from,
		To:         parts[0],
		Ciphertext: parts[1],
		Timestamp:  time.Now().Unix(),
	}
}

// relayMessages forwards messages to recipients.
func (rs *RelayServer) relayMessages() {
	for {
		select {
		case <-rs.done:
			return
		case msg := <-rs.messages:
			rs.mu.RLock()
			if msg.To == "all" {
				// Broadcast to all clients except sender
				for playerID, conn := range rs.clients {
					if playerID != msg.From {
						rs.sendMessage(conn, msg)
					}
				}
			} else {
				// Send to specific recipient
				if conn, ok := rs.clients[msg.To]; ok {
					rs.sendMessage(conn, msg)
				}
			}
			rs.mu.RUnlock()
		}
	}
}

// sendMessage sends an encrypted message blob to a connection.
// Server never decrypts the message.
func (rs *RelayServer) sendMessage(conn net.Conn, msg EncryptedMessage) {
	// Format: "from|ciphertext\n"
	data := fmt.Sprintf("%s|%s\n", msg.From, msg.Ciphertext)
	if _, err := conn.Write([]byte(data)); err != nil {
		rs.logger.WithError(err).Error("failed to send message")
	}
}

// Stop gracefully shuts down the relay server.
func (rs *RelayServer) Stop() error {
	close(rs.done)

	rs.mu.Lock()
	for _, conn := range rs.clients {
		conn.Close()
	}
	rs.mu.Unlock()

	return rs.listener.Close()
}

// GetClientCount returns the number of connected clients.
func (rs *RelayServer) GetClientCount() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return len(rs.clients)
}

// GetAddr returns the server's listening address.
func (rs *RelayServer) GetAddr() string {
	if rs.listener == nil {
		return ""
	}
	return rs.listener.Addr().String()
}

// SetReadTimeout sets the timeout for client read operations.
func (rs *RelayServer) SetReadTimeout(timeout time.Duration) {
	rs.mu.Lock()
	rs.readTimeout = timeout
	rs.mu.Unlock()
}

// SetMessageTimeout sets the timeout for message receive operations.
func (rs *RelayServer) SetMessageTimeout(timeout time.Duration) {
	rs.mu.Lock()
	rs.messageTimeout = timeout
	rs.mu.Unlock()
}

// RelayClient connects to a chat relay server.
type RelayClient struct {
	conn           net.Conn
	playerID       string
	incoming       chan EncryptedMessage
	done           chan struct{}
	closed         bool
	mu             sync.Mutex
	logger         *logrus.Entry
	messageTimeout time.Duration
}

// NewRelayClient creates a chat relay client.
func NewRelayClient(addr, playerID string) (*RelayClient, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Send player ID with newline
	if _, err := conn.Write([]byte(playerID + "\n")); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send player ID: %w", err)
	}

	client := &RelayClient{
		conn:           conn,
		playerID:       playerID,
		incoming:       make(chan EncryptedMessage, 50),
		done:           make(chan struct{}),
		messageTimeout: 100 * time.Millisecond,
		logger: logrus.WithFields(logrus.Fields{
			"system":    "chat_relay_client",
			"player_id": playerID,
		}),
	}

	go client.receiveMessages()
	return client, nil
}

// SendEncrypted sends an encrypted message to the server.
func (rc *RelayClient) SendEncrypted(to, ciphertext string) error {
	data := fmt.Sprintf("%s|%s\n", to, ciphertext)
	if _, err := rc.conn.Write([]byte(data)); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// receiveMessages listens for incoming encrypted messages.
func (rc *RelayClient) receiveMessages() {
	defer close(rc.incoming)

	reader := bufio.NewReader(rc.conn)

	for {
		select {
		case <-rc.done:
			return
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			rc.logger.WithError(err).Error("failed to read message")
			return
		}

		// Parse message: "from|ciphertext"
		data := strings.TrimSpace(line)
		parts := strings.SplitN(data, "|", 2)
		if len(parts) != 2 {
			rc.logger.WithField("data", data).Error("invalid message format")
			continue
		}

		rc.incoming <- EncryptedMessage{
			From:       parts[0],
			Ciphertext: parts[1],
			Timestamp:  time.Now().Unix(),
		}
	}
}

// ReceiveEncrypted returns the next encrypted message from the server.
func (rc *RelayClient) ReceiveEncrypted() (*EncryptedMessage, error) {
	select {
	case msg, ok := <-rc.incoming:
		if !ok {
			return nil, fmt.Errorf("connection closed")
		}
		return &msg, nil
	case <-rc.done:
		return nil, fmt.Errorf("client stopped")
	case <-time.After(rc.messageTimeout):
		return nil, nil // No message available
	}
}

// SetMessageTimeout sets the timeout for message receive operations.
func (rc *RelayClient) SetMessageTimeout(timeout time.Duration) {
	rc.mu.Lock()
	rc.messageTimeout = timeout
	rc.mu.Unlock()
}

// Close disconnects from the relay server.
func (rc *RelayClient) Close() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.closed {
		return nil
	}
	rc.closed = true

	close(rc.done)
	return rc.conn.Close()
}
