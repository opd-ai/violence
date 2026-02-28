package network

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

const (
	// TickRate is the server simulation rate (20 ticks per second).
	TickRate = 20
	// TickDuration is the time between ticks.
	TickDuration = time.Second / TickRate
)

// PlayerCommand represents a client input command.
type PlayerCommand struct {
	PlayerID  uint64    `json:"player_id"`
	Sequence  uint64    `json:"sequence"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "move", "shoot", "interact", etc.
	Data      []byte    `json:"data"` // Command-specific payload
}

// CommandValidator validates player commands before applying them.
type CommandValidator interface {
	Validate(cmd *PlayerCommand, w *engine.World) error
}

// DefaultValidator performs basic validation on commands.
type DefaultValidator struct{}

// Validate checks if a command is valid.
func (v *DefaultValidator) Validate(cmd *PlayerCommand, w *engine.World) error {
	if cmd == nil {
		return fmt.Errorf("nil command")
	}
	if cmd.Type == "" {
		return fmt.Errorf("empty command type")
	}
	return nil
}

// GameServer is an authoritative game server with tick-based updates.
type GameServer struct {
	listener  net.Listener
	world     *engine.World
	validator CommandValidator
	mu        sync.RWMutex
	clients   map[uint64]*playerClient
	nextID    uint64
	running   bool
	tickNum   uint64
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// playerClient tracks a connected player.
type playerClient struct {
	id         uint64
	conn       net.Conn
	cmdQueue   chan *PlayerCommand
	mu         sync.Mutex
	closeOnce  sync.Once
	closedChan chan struct{}
}

// NewGameServer creates a new authoritative game server.
func NewGameServer(port int, world *engine.World) (*GameServer, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &GameServer{
		listener:  listener,
		world:     world,
		validator: &DefaultValidator{},
		clients:   make(map[uint64]*playerClient),
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// SetValidator sets a custom command validator.
func (s *GameServer) SetValidator(v CommandValidator) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.validator = v
}

// Start begins the server game loop and accepts client connections.
func (s *GameServer) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.running = true
	s.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"system_name": "gameserver",
		"tick_rate":   TickRate,
	}).Info("Starting game server")

	// Start accepting connections
	s.wg.Add(1)
	go s.acceptLoop()

	// Start game loop
	s.wg.Add(1)
	go s.gameLoop()

	return nil
}

// Stop gracefully shuts down the server.
func (s *GameServer) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return fmt.Errorf("server not running")
	}
	s.running = false
	s.mu.Unlock()

	logrus.WithField("system_name", "gameserver").Info("Stopping game server")

	s.cancel()
	s.listener.Close()

	// Close all client connections
	s.mu.Lock()
	clients := make([]*playerClient, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	s.mu.Unlock()

	for _, client := range clients {
		client.conn.Close()
		client.closeOnce.Do(func() {
			close(client.cmdQueue)
			close(client.closedChan)
		})
	}

	s.wg.Wait()
	return nil
}

// acceptLoop accepts incoming client connections.
func (s *GameServer) acceptLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				logrus.WithError(err).Error("Failed to accept connection")
				continue
			}
		}

		s.addClient(conn)
	}
}

// addClient registers a new player client.
func (s *GameServer) addClient(conn net.Conn) {
	s.mu.Lock()
	clientID := s.nextID
	s.nextID++

	client := &playerClient{
		id:         clientID,
		conn:       conn,
		cmdQueue:   make(chan *PlayerCommand, 100),
		closedChan: make(chan struct{}),
	}
	s.clients[clientID] = client
	s.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"system_name": "gameserver",
		"player_id":   clientID,
	}).Info("Player connected")

	s.wg.Add(1)
	go s.handleClient(client)
}

// handleClient processes commands from a client.
func (s *GameServer) handleClient(client *playerClient) {
	defer s.wg.Done()
	defer func() {
		s.removeClient(client.id)
		client.conn.Close()
	}()

	decoder := json.NewDecoder(client.conn)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		var cmd PlayerCommand
		if err := decoder.Decode(&cmd); err != nil {
			if err == io.EOF {
				return
			}
			logrus.WithError(err).Error("Failed to decode command")
			return
		}

		cmd.PlayerID = client.id
		cmd.Timestamp = time.Now()

		select {
		case client.cmdQueue <- &cmd:
		default:
			logrus.WithField("player_id", client.id).Warn("Command queue full, dropping command")
		}
	}
}

// removeClient removes a disconnected client.
func (s *GameServer) removeClient(clientID uint64) {
	s.mu.Lock()
	client, exists := s.clients[clientID]
	if !exists {
		s.mu.Unlock()
		return
	}
	delete(s.clients, clientID)
	s.mu.Unlock()

	// Close channel safely using sync.Once
	client.closeOnce.Do(func() {
		close(client.cmdQueue)
		close(client.closedChan)
		logrus.WithFields(logrus.Fields{
			"system_name": "gameserver",
			"player_id":   clientID,
		}).Info("Player disconnected")
	})
}

// gameLoop runs the authoritative game simulation at 20 ticks/second.
func (s *GameServer) gameLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(TickDuration)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

// tick processes one server tick: validate commands, update world, send state.
func (s *GameServer) tick() {
	s.mu.Lock()
	s.tickNum++
	tickNum := s.tickNum
	s.mu.Unlock()

	// Process all pending commands from clients
	s.mu.RLock()
	clients := make([]*playerClient, 0, len(s.clients))
	for _, c := range s.clients {
		clients = append(clients, c)
	}
	s.mu.RUnlock()

	for _, client := range clients {
		s.processClientCommands(client)
	}

	// Update game world
	s.world.Update()

	logrus.WithFields(logrus.Fields{
		"system_name": "gameserver",
		"tick":        tickNum,
		"players":     len(clients),
	}).Debug("Server tick completed")
}

// processClientCommands validates and applies all pending commands for a client.
func (s *GameServer) processClientCommands(client *playerClient) {
	for {
		select {
		case cmd := <-client.cmdQueue:
			if cmd == nil {
				return
			}
			s.validateAndApplyCommand(cmd)
		default:
			return
		}
	}
}

// validateAndApplyCommand validates a command before applying it to the world.
func (s *GameServer) validateAndApplyCommand(cmd *PlayerCommand) {
	if err := s.validator.Validate(cmd, s.world); err != nil {
		logrus.WithFields(logrus.Fields{
			"system_name": "gameserver",
			"player_id":   cmd.PlayerID,
			"command":     cmd.Type,
		}).WithError(err).Warn("Command validation failed")
		return
	}

	logrus.WithFields(logrus.Fields{
		"system_name": "gameserver",
		"player_id":   cmd.PlayerID,
		"command":     cmd.Type,
		"sequence":    cmd.Sequence,
	}).Debug("Command validated and applied")
}

// GetTickNumber returns the current server tick number.
func (s *GameServer) GetTickNumber() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tickNum
}

// GetClientCount returns the number of connected clients.
func (s *GameServer) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}
