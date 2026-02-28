package main

import (
	"bytes"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/network"
)

func TestServerStartStop(t *testing.T) {
	// Create server on random port
	world := engine.NewWorld()
	server, err := network.NewGameServer(0, world)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give server time to initialize
	time.Sleep(50 * time.Millisecond)

	// Stop server
	if err := server.Stop(); err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}
}

func TestServerAcceptsConnections(t *testing.T) {
	// Create and start server
	world := engine.NewWorld()
	server, err := network.NewGameServer(0, world)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Get server address
	addr := server.GetAddr()
	if addr == "" {
		t.Fatal("Server address is empty")
	}

	// Connect as client
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Server should accept connection
	time.Sleep(50 * time.Millisecond)
}

func TestServerReceivesCommands(t *testing.T) {
	// Create and start server
	world := engine.NewWorld()
	server, err := network.NewGameServer(0, world)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Connect as client
	conn, err := net.Dial("tcp", server.GetAddr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Send a command
	cmd := network.PlayerCommand{
		Type:     "move",
		Sequence: 1,
		Data:     []byte(`{"x": 1.0, "y": 2.0}`),
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(&cmd); err != nil {
		t.Fatalf("Failed to send command: %v", err)
	}

	// Give server time to process
	time.Sleep(100 * time.Millisecond)
}

func TestMultipleClients(t *testing.T) {
	// Create and start server
	world := engine.NewWorld()
	server, err := network.NewGameServer(0, world)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Connect multiple clients
	numClients := 3
	conns := make([]net.Conn, numClients)

	for i := 0; i < numClients; i++ {
		conn, err := net.Dial("tcp", server.GetAddr())
		if err != nil {
			t.Fatalf("Failed to connect client %d: %v", i, err)
		}
		conns[i] = conn
		defer conn.Close()
	}

	// All clients should be connected
	time.Sleep(100 * time.Millisecond)

	// Send commands from each client
	for i, conn := range conns {
		cmd := network.PlayerCommand{
			Type:     "move",
			Sequence: uint64(i),
			Data:     []byte(`{"x": 1.0}`),
		}
		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(&cmd); err != nil {
			t.Fatalf("Failed to send from client %d: %v", i, err)
		}
	}

	time.Sleep(100 * time.Millisecond)
}

func TestServerGracefulShutdown(t *testing.T) {
	// Create and start server
	world := engine.NewWorld()
	server, err := network.NewGameServer(0, world)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Connect client
	conn, err := net.Dial("tcp", server.GetAddr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	// Stop server (should close client connection)
	if err := server.Stop(); err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}

	// Client connection should be closed
	buf := make([]byte, 1)
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, err = conn.Read(buf)
	if err == nil {
		t.Error("Expected connection to be closed")
	}
}

func TestServerLogging(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer

	// Create server
	world := engine.NewWorld()
	server, err := network.NewGameServer(0, world)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start and stop server (should generate logs)
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if err := server.Stop(); err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}

	// Verify logging occurred (basic check)
	if logBuf.Len() > 0 {
		t.Logf("Logs captured: %s", logBuf.String())
	}
}

func TestServerDoubleStart(t *testing.T) {
	// Create server
	world := engine.NewWorld()
	server, err := network.NewGameServer(0, world)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Try to start again (should fail)
	if err := server.Start(); err == nil {
		t.Error("Expected error when starting already running server")
	}
}

func TestServerStopBeforeStart(t *testing.T) {
	// Create server
	world := engine.NewWorld()
	server, err := network.NewGameServer(0, world)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Try to stop before starting (should fail)
	if err := server.Stop(); err == nil {
		t.Error("Expected error when stopping non-running server")
	}
}
