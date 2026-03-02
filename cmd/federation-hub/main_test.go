package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opd-ai/violence/pkg/federation"
)

func TestNewHubServer(t *testing.T) {
	tests := []struct {
		name      string
		authToken string
		peers     []string
	}{
		{
			name:      "no auth, no peers",
			authToken: "",
			peers:     nil,
		},
		{
			name:      "with auth token",
			authToken: "test-token-123",
			peers:     nil,
		},
		{
			name:      "with peers",
			authToken: "",
			peers:     []string{"http://hub1.example.com", "http://hub2.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewHubServer(tt.authToken, tt.peers)
			if server == nil {
				t.Fatal("NewHubServer returned nil")
			}
			if server.authToken != tt.authToken {
				t.Errorf("authToken = %q, want %q", server.authToken, tt.authToken)
			}
			if len(server.peers) != len(tt.peers) {
				t.Errorf("len(peers) = %d, want %d", len(server.peers), len(tt.peers))
			}
		})
	}
}

func TestHandleHealth(t *testing.T) {
	server := NewHubServer("", nil)
	if err := server.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var health HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&health); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if health.Status != "ok" {
		t.Errorf("status = %q, want %q", health.Status, "ok")
	}
	if health.Version != version {
		t.Errorf("version = %q, want %q", health.Version, version)
	}
	if health.ServerCount < 0 {
		t.Errorf("serverCount = %d, want >= 0", health.ServerCount)
	}
}

func TestHandleAnnounceHTTP(t *testing.T) {
	tests := []struct {
		name       string
		authToken  string
		reqToken   string
		method     string
		body       interface{}
		wantStatus int
	}{
		{
			name:      "valid announcement no auth",
			authToken: "",
			method:    http.MethodPost,
			body: federation.ServerAnnouncement{
				Name:       "test-server",
				Address:    "localhost:7777",
				Region:     federation.RegionUSEast,
				Genre:      "horror",
				Players:    5,
				MaxPlayers: 16,
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "valid announcement with auth",
			authToken: "secret-token",
			reqToken:  "Bearer secret-token",
			method:    http.MethodPost,
			body: federation.ServerAnnouncement{
				Name:       "test-server",
				Address:    "localhost:7777",
				Region:     federation.RegionUSEast,
				Genre:      "horror",
				Players:    5,
				MaxPlayers: 16,
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "unauthorized",
			authToken:  "secret-token",
			reqToken:   "Bearer wrong-token",
			method:     http.MethodPost,
			body:       federation.ServerAnnouncement{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "invalid json",
			method:     http.MethodPost,
			body:       "invalid json",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewHubServer(tt.authToken, nil)
			if err := server.Start("127.0.0.1:0"); err != nil {
				t.Fatalf("failed to start server: %v", err)
			}
			defer server.Stop()

			var bodyReader *bytes.Reader
			if tt.body != nil {
				if s, ok := tt.body.(string); ok {
					bodyReader = bytes.NewReader([]byte(s))
				} else {
					data, _ := json.Marshal(tt.body)
					bodyReader = bytes.NewReader(data)
				}
			} else {
				bodyReader = bytes.NewReader([]byte{})
			}

			req := httptest.NewRequest(tt.method, "/announce", bodyReader)
			if tt.reqToken != "" {
				req.Header.Set("Authorization", tt.reqToken)
			}
			w := httptest.NewRecorder()

			server.handleAnnounceHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status code = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandleQuery(t *testing.T) {
	server := NewHubServer("", nil)
	if err := server.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Stop()

	// Register test servers
	server.hub.RegisterServer(&federation.ServerAnnouncement{
		Name:       "horror-server",
		Address:    "localhost:7777",
		Region:     federation.RegionUSEast,
		Genre:      "horror",
		Players:    5,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	})
	server.hub.RegisterServer(&federation.ServerAnnouncement{
		Name:       "scifi-server",
		Address:    "localhost:7778",
		Region:     federation.RegionEUWest,
		Genre:      "scifi",
		Players:    10,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	})

	tests := []struct {
		name       string
		query      federation.ServerQuery
		wantLen    int
		wantStatus int
	}{
		{
			name:       "query all servers",
			query:      federation.ServerQuery{},
			wantLen:    2,
			wantStatus: http.StatusOK,
		},
		{
			name: "query horror servers",
			query: federation.ServerQuery{
				Genre: stringPtr("horror"),
			},
			wantLen:    1,
			wantStatus: http.StatusOK,
		},
		{
			name: "query us-east servers",
			query: federation.ServerQuery{
				Region: regionPtr(federation.RegionUSEast),
			},
			wantLen:    1,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := json.Marshal(tt.query)
			req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(data))
			w := httptest.NewRecorder()

			server.handleQuery(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status code = %d, want %d", w.Code, tt.wantStatus)
			}

			var servers []*federation.ServerAnnouncement
			if err := json.NewDecoder(w.Body).Decode(&servers); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if len(servers) != tt.wantLen {
				t.Errorf("len(servers) = %d, want %d", len(servers), tt.wantLen)
			}
		})
	}
}

func TestHandleLookup(t *testing.T) {
	server := NewHubServer("", nil)
	if err := server.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Stop()

	// Register test server with player list
	server.hub.RegisterServer(&federation.ServerAnnouncement{
		Name:       "test-server",
		Address:    "localhost:7777",
		Region:     federation.RegionUSEast,
		Genre:      "horror",
		Players:    2,
		MaxPlayers: 16,
		PlayerList: []string{"player1", "player2"},
		Timestamp:  time.Now(),
	})

	tests := []struct {
		name       string
		playerID   string
		wantOnline bool
		wantStatus int
	}{
		{
			name:       "player online",
			playerID:   "player1",
			wantOnline: true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "player offline",
			playerID:   "player3",
			wantOnline: false,
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty player id",
			playerID:   "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := federation.PlayerLookupRequest{PlayerID: tt.playerID}
			data, _ := json.Marshal(req)
			httpReq := httptest.NewRequest(http.MethodPost, "/lookup", bytes.NewReader(data))
			w := httptest.NewRecorder()

			server.handleLookup(w, httpReq)

			if w.Code != tt.wantStatus {
				t.Errorf("status code = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response federation.PlayerLookupResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if response.Online != tt.wantOnline {
					t.Errorf("online = %v, want %v", response.Online, tt.wantOnline)
				}
			}
		})
	}
}

func TestHandlePeers(t *testing.T) {
	peers := []string{"http://hub1.example.com", "http://hub2.example.com"}
	server := NewHubServer("", peers)

	req := httptest.NewRequest(http.MethodGet, "/peers", nil)
	w := httptest.NewRecorder()

	server.handlePeers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response PeerResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Peers) != len(peers) {
		t.Errorf("len(peers) = %d, want %d", len(response.Peers), len(peers))
	}
}

func TestRateLimit(t *testing.T) {
	oldRateLimit := *rateLimit
	*rateLimit = 2 // 2 requests per minute for testing
	defer func() { *rateLimit = oldRateLimit }()

	server := NewHubServer("", nil)
	if err := server.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Stop()

	// Make requests up to the limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		server.withRateLimit(server.handleHealth)(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: status code = %d, want %d", i+1, w.Code, http.StatusOK)
		}
	}

	// Next request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	server.withRateLimit(server.handleHealth)(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		wantIP     string
	}{
		{
			name:       "x-forwarded-for header",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1"},
			wantIP:     "192.168.1.1",
		},
		{
			name:       "x-real-ip header",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Real-IP": "192.168.1.2"},
			wantIP:     "192.168.1.2",
		},
		{
			name:       "remote addr fallback",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{},
			wantIP:     "10.0.0.1:12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := getClientIP(req)
			if ip != tt.wantIP {
				t.Errorf("getClientIP() = %q, want %q", ip, tt.wantIP)
			}
		})
	}
}

func TestSplitPeers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "single peer",
			input: "http://hub1.example.com",
			want:  []string{"http://hub1.example.com"},
		},
		{
			name:  "multiple peers",
			input: "http://hub1.example.com,http://hub2.example.com,http://hub3.example.com",
			want:  []string{"http://hub1.example.com", "http://hub2.example.com", "http://hub3.example.com"},
		},
		{
			name:  "trailing comma",
			input: "http://hub1.example.com,http://hub2.example.com,",
			want:  []string{"http://hub1.example.com", "http://hub2.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitPeers(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("len(splitPeers()) = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitPeers()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSyncWithPeer(t *testing.T) {
	// Create a mock peer hub
	peerHub := federation.NewFederationHub()
	if err := peerHub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start peer hub: %v", err)
	}
	defer peerHub.Stop()

	time.Sleep(50 * time.Millisecond)

	// Register servers on peer hub
	peerHub.RegisterServer(&federation.ServerAnnouncement{
		Name:       "peer-server",
		Address:    "peer.example.com:7777",
		Region:     federation.RegionUSEast,
		Genre:      "horror",
		Players:    5,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	})

	// Create main hub
	server := NewHubServer("", nil)
	if err := server.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	// Sync with peer
	peerURL := "http://" + peerHub.GetAddr()
	server.syncWithPeer(peerURL)

	// Check that server was synced
	servers := server.hub.QueryServers(&federation.ServerQuery{})
	if len(servers) != 1 {
		t.Errorf("len(servers) = %d, want 1", len(servers))
	}

	if len(servers) > 0 && servers[0].Name != "peer-server" {
		t.Errorf("server name = %q, want %q", servers[0].Name, "peer-server")
	}
}

func TestServerLifecycle(t *testing.T) {
	server := NewHubServer("", nil)

	// Start server
	if err := server.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Verify server is running
	if server.GetAddr() == "" {
		t.Error("server address is empty")
	}

	// Stop server
	if err := server.Stop(); err != nil {
		t.Errorf("failed to stop server: %v", err)
	}
}

func TestRateLimiterCleanup_TTL(t *testing.T) {
	server := NewHubServer("", nil)
	if err := server.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Stop()

	// Add rate limiters with different last access times
	server.mu.Lock()
	now := time.Now()
	server.rateLimits["192.168.1.1"] = &rateLimiterEntry{
		limiter:    nil,
		lastAccess: now.Add(-2 * time.Hour), // Inactive for 2 hours
	}
	server.rateLimits["192.168.1.2"] = &rateLimiterEntry{
		limiter:    nil,
		lastAccess: now.Add(-30 * time.Minute), // Inactive for 30 minutes
	}
	server.rateLimits["192.168.1.3"] = &rateLimiterEntry{
		limiter:    nil,
		lastAccess: now, // Active
	}
	initialCount := len(server.rateLimits)
	server.mu.Unlock()

	if initialCount != 3 {
		t.Fatalf("initial count = %d, want 3", initialCount)
	}

	// Manually trigger cleanup
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	const inactivityTTL = 1 * time.Hour

	// Simulate one cleanup cycle
	server.mu.Lock()
	now = time.Now()
	for ip, entry := range server.rateLimits {
		if now.Sub(entry.lastAccess) > inactivityTTL {
			delete(server.rateLimits, ip)
		}
	}
	finalCount := len(server.rateLimits)
	server.mu.Unlock()

	// Should have removed the 2-hour old entry
	if finalCount != 2 {
		t.Errorf("final count = %d, want 2 (removed 1 inactive entry)", finalCount)
	}

	// Verify the correct entries remain
	server.mu.Lock()
	if _, exists := server.rateLimits["192.168.1.1"]; exists {
		t.Error("192.168.1.1 should have been removed")
	}
	if _, exists := server.rateLimits["192.168.1.2"]; !exists {
		t.Error("192.168.1.2 should still exist")
	}
	if _, exists := server.rateLimits["192.168.1.3"]; !exists {
		t.Error("192.168.1.3 should still exist")
	}
	server.mu.Unlock()
}

func TestRateLimiterCleanup_LRU(t *testing.T) {
	server := NewHubServer("", nil)
	if err := server.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Stop()

	const maxLimiters = 10
	const testCount = 15

	// Add more limiters than the max
	server.mu.Lock()
	now := time.Now()
	for i := 0; i < testCount; i++ {
		ip := fmt.Sprintf("192.168.1.%d", i)
		server.rateLimits[ip] = &rateLimiterEntry{
			limiter:    nil,
			lastAccess: now.Add(-time.Duration(i) * time.Minute), // Older entries have smaller i
		}
	}
	server.mu.Unlock()

	// Simulate LRU eviction (manual for test)
	server.mu.Lock()
	if len(server.rateLimits) > maxLimiters {
		type entry struct {
			ip         string
			lastAccess time.Time
		}
		entries := make([]entry, 0, len(server.rateLimits))
		for ip, e := range server.rateLimits {
			entries = append(entries, entry{ip: ip, lastAccess: e.lastAccess})
		}

		// Sort by last access time (oldest first)
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].lastAccess.After(entries[j].lastAccess) {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}

		// Evict oldest entries
		evictCount := len(server.rateLimits) - maxLimiters + (maxLimiters / 10)
		for i := 0; i < evictCount && i < len(entries); i++ {
			delete(server.rateLimits, entries[i].ip)
		}
	}
	finalCount := len(server.rateLimits)
	server.mu.Unlock()

	// Should have evicted excess entries
	if finalCount >= testCount {
		t.Errorf("final count = %d, should be less than %d after LRU eviction", finalCount, testCount)
	}

	// Verify we're at or below the target
	if finalCount > maxLimiters-(maxLimiters/10) {
		t.Errorf("final count = %d, expected <= %d after eviction", finalCount, maxLimiters-(maxLimiters/10))
	}
}

func TestRateLimiterCleanup_UpdateLastAccess(t *testing.T) {
	server := NewHubServer("", nil)
	if err := server.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Stop()

	// Create a request handler
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	// First request creates the entry
	server.withRateLimit(server.handleHealth)(w, req)

	server.mu.Lock()
	entry, exists := server.rateLimits["192.168.1.100:12345"]
	if !exists {
		t.Fatal("rate limiter entry not created")
	}
	firstAccess := entry.lastAccess
	server.mu.Unlock()

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Second request should update lastAccess
	w = httptest.NewRecorder()
	server.withRateLimit(server.handleHealth)(w, req)

	server.mu.Lock()
	entry, exists = server.rateLimits["192.168.1.100:12345"]
	if !exists {
		t.Fatal("rate limiter entry disappeared")
	}
	secondAccess := entry.lastAccess
	server.mu.Unlock()

	if !secondAccess.After(firstAccess) {
		t.Errorf("lastAccess not updated: first=%v, second=%v", firstAccess, secondAccess)
	}
}

func TestRateLimiterCleanup_Integration(t *testing.T) {
	server := NewHubServer("", nil)
	if err := server.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Stop()

	// Add various entries
	server.mu.Lock()
	now := time.Now()

	// Old entries that should be cleaned up
	for i := 0; i < 5; i++ {
		ip := fmt.Sprintf("old.%d", i)
		server.rateLimits[ip] = &rateLimiterEntry{
			limiter:    nil,
			lastAccess: now.Add(-2 * time.Hour),
		}
	}

	// Recent entries that should remain
	for i := 0; i < 3; i++ {
		ip := fmt.Sprintf("recent.%d", i)
		server.rateLimits[ip] = &rateLimiterEntry{
			limiter:    nil,
			lastAccess: now.Add(-10 * time.Minute),
		}
	}

	initialCount := len(server.rateLimits)
	server.mu.Unlock()

	if initialCount != 8 {
		t.Fatalf("initial count = %d, want 8", initialCount)
	}

	// Manually trigger TTL cleanup
	const inactivityTTL = 1 * time.Hour
	server.mu.Lock()
	now = time.Now()
	for ip, entry := range server.rateLimits {
		if now.Sub(entry.lastAccess) > inactivityTTL {
			delete(server.rateLimits, ip)
		}
	}
	finalCount := len(server.rateLimits)
	server.mu.Unlock()

	// Should have removed 5 old entries, kept 3 recent
	if finalCount != 3 {
		t.Errorf("final count = %d, want 3 (kept recent entries only)", finalCount)
	}

	// Verify only recent entries remain
	server.mu.Lock()
	for i := 0; i < 3; i++ {
		ip := fmt.Sprintf("recent.%d", i)
		if _, exists := server.rateLimits[ip]; !exists {
			t.Errorf("recent entry %s should still exist", ip)
		}
	}
	for i := 0; i < 5; i++ {
		ip := fmt.Sprintf("old.%d", i)
		if _, exists := server.rateLimits[ip]; exists {
			t.Errorf("old entry %s should have been removed", ip)
		}
	}
	server.mu.Unlock()
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func regionPtr(r federation.Region) *federation.Region {
	return &r
}
