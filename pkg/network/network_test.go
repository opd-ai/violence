package network

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestClient_ConnectAndSend(t *testing.T) {
	// Start a server
	server := &Server{Port: 9999}
	if err := server.Listen(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Close()

	// Accept connections in background
	done := make(chan net.Conn, 1)
	go func() {
		conn, _ := server.Accept()
		done <- conn
	}()

	// Connect client
	client := &Client{}
	if err := client.Connect("localhost:9999"); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Wait for server to accept
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for server accept")
	}

	// Test send
	testData := []byte("test message")
	if err := client.Send(testData); err != nil {
		t.Fatalf("failed to send: %v", err)
	}
}

func TestClient_SendWithoutConnect(t *testing.T) {
	client := &Client{}
	err := client.Send([]byte("test"))
	if err == nil {
		t.Fatal("expected error when sending without connection")
	}
}

func TestClient_ReceiveWithoutConnect(t *testing.T) {
	client := &Client{}
	_, err := client.Receive()
	if err == nil {
		t.Fatal("expected error when receiving without connection")
	}
}

func TestClient_Close(t *testing.T) {
	client := &Client{}
	// Close without connection should not panic
	if err := client.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
}

func TestServer_ListenAndAccept(t *testing.T) {
	server := &Server{Port: 10000}
	if err := server.Listen(); err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer server.Close()

	// Try to accept in background with timeout
	done := make(chan error, 1)
	go func() {
		client := &Client{}
		done <- client.Connect("localhost:10000")
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("client connect failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for client")
	}

	conn, err := server.Accept()
	if err != nil {
		t.Fatalf("accept failed: %v", err)
	}
	if conn == nil {
		t.Fatal("accepted connection is nil")
	}
}

func TestServer_AcceptWithoutListen(t *testing.T) {
	server := &Server{Port: 10001}
	_, err := server.Accept()
	if err == nil {
		t.Fatal("expected error when accepting without listening")
	}
}

func TestServer_Close(t *testing.T) {
	server := &Server{Port: 10002}
	if err := server.Listen(); err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	if err := server.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	// Closing again should not panic
	if err := server.Close(); err != nil {
		t.Fatalf("second close failed: %v", err)
	}
}

func TestClientServer_RoundTrip(t *testing.T) {
	server := &Server{Port: 10003}
	if err := server.Listen(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Close()

	// Accept connections in background
	serverConn := make(chan net.Conn, 1)
	go func() {
		conn, _ := server.Accept()
		serverConn <- conn
	}()

	// Connect client
	client := &Client{}
	if err := client.Connect("localhost:10003"); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Wait for server connection
	var sConn net.Conn
	select {
	case sConn = <-serverConn:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for server accept")
	}

	// Client sends data
	testData := []byte("hello server")
	if err := client.Send(testData); err != nil {
		t.Fatalf("client send failed: %v", err)
	}

	// Server receives data
	buf := make([]byte, 4096)
	n, err := sConn.Read(buf)
	if err != nil {
		t.Fatalf("server read failed: %v", err)
	}

	received := buf[:n]
	if !bytes.Equal(received, testData) {
		t.Fatalf("data mismatch: got %q, want %q", received, testData)
	}
}

func TestSetGenre(t *testing.T) {
	// Should not panic
	SetGenre("fantasy")
	SetGenre("scifi")
	SetGenre("")
}

// TestServerUsesTCP verifies that the server uses TCP protocol as documented
func TestServerUsesTCP(t *testing.T) {
	server := &Server{Port: 10004}
	if err := server.Listen(); err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer server.Close()

	// Verify we can connect using TCP
	client := &Client{}
	if err := client.Connect("localhost:10004"); err != nil {
		t.Fatalf("TCP connection failed: %v", err)
	}
	defer client.Close()

	// Accept the connection
	conn, err := server.Accept()
	if err != nil {
		t.Fatalf("accept failed: %v", err)
	}

	// Verify connection type is TCP
	if conn.RemoteAddr().Network() != "tcp" {
		t.Errorf("expected TCP protocol, got %s", conn.RemoteAddr().Network())
	}
}
