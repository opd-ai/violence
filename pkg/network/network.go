// Package network provides client-server networking primitives.
package network

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// Client represents a network client connection.
type Client struct {
	Address string
	conn    net.Conn
	mu      sync.Mutex
}

// Server represents a network server.
type Server struct {
	Port     int
	listener net.Listener
	mu       sync.Mutex
	clients  []net.Conn
}

// Connect establishes a client connection to the given address.
func (c *Client) Connect(address string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Address = address
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	c.conn = conn
	return nil
}

// Listen starts the server on its configured port.
func (s *Server) Listen() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	addr := fmt.Sprintf(":%d", s.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.Port, err)
	}
	s.listener = listener
	return nil
}

// Accept accepts a new client connection.
func (s *Server) Accept() (net.Conn, error) {
	if s.listener == nil {
		return nil, fmt.Errorf("server not listening")
	}
	conn, err := s.listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("failed to accept connection: %w", err)
	}
	s.mu.Lock()
	s.clients = append(s.clients, conn)
	s.mu.Unlock()
	return conn, nil
}

// Send transmits data over the connection.
func (c *Client) Send(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	_, err := c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}
	return nil
}

// Receive reads incoming data from the connection.
func (c *Client) Receive() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}
	buf := make([]byte, 4096)
	n, err := c.conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to receive data: %w", err)
	}
	return buf[:n], nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// Close closes the server and all client connections.
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var firstErr error
	for _, conn := range s.clients {
		if err := conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	s.clients = nil

	if s.listener != nil {
		if err := s.listener.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		s.listener = nil
	}
	return firstErr
}

// SetGenre configures the network system for a genre.
func SetGenre(genreID string) {}
