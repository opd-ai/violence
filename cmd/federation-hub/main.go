// Package main provides a standalone federation hub server.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/opd-ai/violence/pkg/federation"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

const version = "6.0.0"

var (
	addr      = flag.String("addr", ":8080", "HTTP server address")
	authToken = flag.String("auth-token", "", "Optional auth token for server registration")
	peerURLs  = flag.String("peers", "", "Comma-separated list of peer hub URLs for syncing")
	logLevel  = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	rateLimit = flag.Int("rate-limit", 60, "Rate limit per IP (requests per minute)")
)

// HubServer wraps the federation hub with additional features.
type HubServer struct {
	hub        *federation.FederationHub
	authToken  string
	peers      []string
	startTime  time.Time
	rateLimits map[string]*rate.Limiter
	httpServer *http.Server
	addr       string
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.Mutex
}

// HealthResponse contains health check information.
type HealthResponse struct {
	Status      string `json:"status"`
	Version     string `json:"version"`
	Uptime      string `json:"uptime"`
	ServerCount int    `json:"serverCount"`
}

// NewHubServer creates a new hub server.
func NewHubServer(authToken string, peers []string) *HubServer {
	hub := federation.NewFederationHub()
	hub.SetStaleTimeout(15 * time.Minute)
	hub.SetCleanupInterval(1 * time.Minute)

	ctx, cancel := context.WithCancel(context.Background())

	return &HubServer{
		hub:        hub,
		authToken:  authToken,
		peers:      peers,
		startTime:  time.Now(),
		rateLimits: make(map[string]*rate.Limiter),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start begins the hub server.
func (s *HubServer) Start(addr string) error {
	mux := http.NewServeMux()

	// Wrap handlers with rate limiting
	mux.HandleFunc("/announce", s.withRateLimit(s.handleAnnounceHTTP))
	mux.HandleFunc("/query", s.withRateLimit(s.handleQuery))
	mux.HandleFunc("/lookup", s.withRateLimit(s.handleLookup))
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/peers", s.withRateLimit(s.handlePeers))

	// Create HTTP server
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Start listening
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.httpServer = server
	s.addr = listener.Addr().String()

	// Start serving in background
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Error("HTTP server error")
		}
	}()

	// Start hub-to-hub sync
	go s.syncWithPeers()

	// Start cleanup routine
	go s.runCleanup()

	logrus.WithFields(logrus.Fields{
		"addr":    listener.Addr().String(),
		"version": version,
		"peers":   len(s.peers),
	}).Info("federation hub started")

	return nil
}

// Stop gracefully shuts down the hub server.
func (s *HubServer) Stop() error {
	s.cancel()
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return err
		}
	}
	return s.hub.Stop()
}

// runCleanup periodically removes stale server registrations.
func (s *HubServer) runCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			// Query all servers and check timestamps
			servers := s.hub.QueryServers(&federation.ServerQuery{})
			now := time.Now()
			for _, server := range servers {
				if now.Sub(server.Timestamp) > 15*time.Minute {
					// Server is stale - we can't remove it directly,
					// so we'll rely on the internal cleanup
					logrus.WithField("server_name", server.Name).Debug("detected stale server")
				}
			}
		}
	}
}

// GetAddr returns the address the hub is listening on.
func (s *HubServer) GetAddr() string {
	return s.addr
}

// withRateLimit wraps a handler with rate limiting.
func (s *HubServer) withRateLimit(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		s.mu.Lock()
		limiter, exists := s.rateLimits[ip]
		if !exists {
			limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(*rateLimit)), *rateLimit)
			s.rateLimits[ip] = limiter
		}
		s.mu.Unlock()

		if !limiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			logrus.WithField("ip", ip).Warn("rate limit exceeded")
			return
		}

		handler(w, r)
	}
}

// handleAnnounceHTTP handles server registration via HTTP POST.
func (s *HubServer) handleAnnounceHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check auth token if configured
	if s.authToken != "" {
		token := r.Header.Get("Authorization")
		if token != "Bearer "+s.authToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	var announcement federation.ServerAnnouncement
	if err := json.NewDecoder(r.Body).Decode(&announcement); err != nil {
		http.Error(w, "invalid announcement", http.StatusBadRequest)
		return
	}

	announcement.Timestamp = time.Now()
	s.hub.RegisterServer(&announcement)

	logrus.WithFields(logrus.Fields{
		"server_name": announcement.Name,
		"region":      announcement.Region,
		"genre":       announcement.Genre,
		"players":     announcement.Players,
	}).Info("server registered")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}

// handleQuery proxies to the federation hub's query handler.
func (s *HubServer) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var query federation.ServerQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, "invalid query", http.StatusBadRequest)
		return
	}

	servers := s.hub.QueryServers(&query)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

// handleLookup handles player presence lookup requests.
func (s *HubServer) handleLookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req federation.PlayerLookupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.PlayerID == "" {
		http.Error(w, "playerID is required", http.StatusBadRequest)
		return
	}

	// Look up player by querying all servers
	servers := s.hub.QueryServers(&federation.ServerQuery{})
	var response federation.PlayerLookupResponse

	for _, server := range servers {
		for _, playerID := range server.PlayerList {
			if playerID == req.PlayerID {
				response.Online = true
				response.ServerAddress = server.Address
				response.ServerName = server.Name
				break
			}
		}
		if response.Online {
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth returns health check information.
func (s *HubServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(s.startTime)

	response := HealthResponse{
		Status:      "ok",
		Version:     version,
		Uptime:      uptime.String(),
		ServerCount: s.hub.GetServerCount(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PeerResponse contains information about peer hubs.
type PeerResponse struct {
	Peers []string `json:"peers"`
}

// handlePeers returns the list of peer hubs.
func (s *HubServer) handlePeers(w http.ResponseWriter, r *http.Request) {
	response := PeerResponse{
		Peers: s.peers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// syncWithPeers periodically syncs server registry with peer hubs.
func (s *HubServer) syncWithPeers() {
	if len(s.peers) == 0 {
		return
	}

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		for _, peerURL := range s.peers {
			s.syncWithPeer(peerURL)
		}
	}
}

// syncWithPeer syncs server registry with a single peer hub.
func (s *HubServer) syncWithPeer(peerURL string) {
	servers, err := federation.DiscoverServers(peerURL, &federation.ServerQuery{}, 10*time.Second)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"peer":  peerURL,
			"error": err,
		}).Warn("failed to sync with peer")
		return
	}

	for _, server := range servers {
		// Re-register servers from peer with current timestamp
		serverCopy := server
		serverCopy.Timestamp = time.Now()
		s.hub.RegisterServer(&serverCopy)
	}

	logrus.WithFields(logrus.Fields{
		"peer":    peerURL,
		"servers": len(servers),
	}).Debug("synced with peer")
}

// getClientIP extracts the client IP from the request.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

func main() {
	flag.Parse()

	// Configure logging
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logrus.Fatal("invalid log level")
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Parse peer URLs
	var peers []string
	if *peerURLs != "" {
		peers = splitPeers(*peerURLs)
	}

	// Create hub server
	server := NewHubServer(*authToken, peers)

	// Start server
	if err := server.Start(*addr); err != nil {
		logrus.WithError(err).Fatal("failed to start hub server")
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logrus.Info("shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan error)
	go func() {
		done <- server.Stop()
	}()

	select {
	case <-ctx.Done():
		logrus.Warn("shutdown timeout exceeded")
	case err := <-done:
		if err != nil {
			logrus.WithError(err).Error("shutdown error")
		} else {
			logrus.Info("shutdown complete")
		}
	}
}

// splitPeers splits a comma-separated string into a slice of peer URLs.
func splitPeers(s string) []string {
	var peers []string
	for i := 0; i < len(s); {
		// Find next comma
		end := i
		for end < len(s) && s[end] != ',' {
			end++
		}
		// Extract peer URL
		if peer := s[i:end]; peer != "" {
			peers = append(peers, peer)
		}
		i = end + 1
	}
	return peers
}
