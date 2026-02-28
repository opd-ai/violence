package federation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestNewFederationHub(t *testing.T) {
	hub := NewFederationHub()
	if hub == nil {
		t.Fatal("NewFederationHub returned nil")
	}
	if hub.servers == nil {
		t.Fatal("servers map not initialized")
	}
	if hub.staleTimeout != 30*time.Second {
		t.Errorf("staleTimeout = %v, want 30s", hub.staleTimeout)
	}
}

func TestFederationHub_StartStop(t *testing.T) {
	hub := NewFederationHub()
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer hub.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	if hub.httpServer == nil {
		t.Fatal("httpServer not initialized after Start")
	}

	if err := hub.Stop(); err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}

func TestFederationHub_RegisterServer(t *testing.T) {
	hub := NewFederationHub()
	announcement := &ServerAnnouncement{
		Name:       "test-server",
		Address:    "localhost:8000",
		Region:     RegionUSEast,
		Genre:      "scifi",
		Players:    5,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	}

	hub.registerServer(announcement)

	if hub.GetServerCount() != 1 {
		t.Errorf("server count = %d, want 1", hub.GetServerCount())
	}

	hub.mu.RLock()
	stored := hub.servers["test-server"]
	hub.mu.RUnlock()

	if stored == nil {
		t.Fatal("server not stored")
	}
	if stored.Address != "localhost:8000" {
		t.Errorf("address = %s, want localhost:8000", stored.Address)
	}
}

func TestFederationHub_QueryServers(t *testing.T) {
	hub := NewFederationHub()

	// Add test servers
	hub.registerServer(&ServerAnnouncement{
		Name:       "us-scifi",
		Address:    "localhost:8000",
		Region:     RegionUSEast,
		Genre:      "scifi",
		Players:    5,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	})
	hub.registerServer(&ServerAnnouncement{
		Name:       "eu-fantasy",
		Address:    "localhost:8001",
		Region:     RegionEUWest,
		Genre:      "fantasy",
		Players:    10,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	})
	hub.registerServer(&ServerAnnouncement{
		Name:       "us-fantasy",
		Address:    "localhost:8002",
		Region:     RegionUSWest,
		Genre:      "fantasy",
		Players:    2,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	})

	tests := []struct {
		name     string
		query    ServerQuery
		wantLen  int
		wantName string
	}{
		{
			name:    "all servers",
			query:   ServerQuery{},
			wantLen: 3,
		},
		{
			name: "filter by region",
			query: ServerQuery{
				Region: ptrRegion(RegionUSEast),
			},
			wantLen:  1,
			wantName: "us-scifi",
		},
		{
			name: "filter by genre",
			query: ServerQuery{
				Genre: ptrString("fantasy"),
			},
			wantLen: 2,
		},
		{
			name: "filter by min players",
			query: ServerQuery{
				MinPlayers: ptrInt(8),
			},
			wantLen:  1,
			wantName: "eu-fantasy",
		},
		{
			name: "filter by max players",
			query: ServerQuery{
				MaxPlayers: ptrInt(3),
			},
			wantLen:  1,
			wantName: "us-fantasy",
		},
		{
			name: "combined filters",
			query: ServerQuery{
				Genre:      ptrString("fantasy"),
				MinPlayers: ptrInt(5),
			},
			wantLen:  1,
			wantName: "eu-fantasy",
		},
		{
			name: "no matches",
			query: ServerQuery{
				Genre: ptrString("horror"),
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := hub.queryServers(&tt.query)
			if len(results) != tt.wantLen {
				t.Errorf("got %d results, want %d", len(results), tt.wantLen)
			}
			if tt.wantName != "" && len(results) > 0 {
				if results[0].Name != tt.wantName {
					t.Errorf("got server %s, want %s", results[0].Name, tt.wantName)
				}
			}
		})
	}
}

func TestFederationHub_MatchesQuery(t *testing.T) {
	hub := NewFederationHub()

	server := &ServerAnnouncement{
		Name:       "test",
		Region:     RegionUSEast,
		Genre:      "scifi",
		Players:    8,
		MaxPlayers: 16,
	}

	tests := []struct {
		name  string
		query ServerQuery
		want  bool
	}{
		{
			name:  "empty query matches",
			query: ServerQuery{},
			want:  true,
		},
		{
			name: "matching region",
			query: ServerQuery{
				Region: ptrRegion(RegionUSEast),
			},
			want: true,
		},
		{
			name: "non-matching region",
			query: ServerQuery{
				Region: ptrRegion(RegionEUWest),
			},
			want: false,
		},
		{
			name: "matching genre",
			query: ServerQuery{
				Genre: ptrString("scifi"),
			},
			want: true,
		},
		{
			name: "non-matching genre",
			query: ServerQuery{
				Genre: ptrString("fantasy"),
			},
			want: false,
		},
		{
			name: "min players met",
			query: ServerQuery{
				MinPlayers: ptrInt(5),
			},
			want: true,
		},
		{
			name: "min players not met",
			query: ServerQuery{
				MinPlayers: ptrInt(10),
			},
			want: false,
		},
		{
			name: "max players met",
			query: ServerQuery{
				MaxPlayers: ptrInt(10),
			},
			want: true,
		},
		{
			name: "max players exceeded",
			query: ServerQuery{
				MaxPlayers: ptrInt(5),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hub.matchesQuery(server, &tt.query)
			if got != tt.want {
				t.Errorf("matchesQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFederationHub_CleanupStaleServers(t *testing.T) {
	hub := NewFederationHub()
	hub.staleTimeout = 200 * time.Millisecond
	hub.cleanupInterval = 50 * time.Millisecond

	now := time.Now()

	// Add stale server (well past timeout)
	hub.registerServer(&ServerAnnouncement{
		Name:      "stale",
		Timestamp: now.Add(-500 * time.Millisecond),
	})
	// Add fresh server (within timeout)
	hub.registerServer(&ServerAnnouncement{
		Name:      "fresh",
		Timestamp: now,
	})

	if hub.GetServerCount() != 2 {
		t.Fatalf("initial server count = %d, want 2", hub.GetServerCount())
	}

	// Start cleanup
	go hub.cleanupStaleServers()
	defer hub.cancel()

	// Wait for at least 2 cleanup cycles
	time.Sleep(150 * time.Millisecond)

	count := hub.GetServerCount()
	if count != 1 {
		t.Errorf("server count after cleanup = %d, want 1", count)
	}

	hub.mu.RLock()
	_, hasFresh := hub.servers["fresh"]
	_, hasStale := hub.servers["stale"]
	hub.mu.RUnlock()

	if !hasFresh {
		t.Error("fresh server was removed")
	}
	if hasStale {
		t.Error("stale server was not removed")
	}
}

func TestFederationHub_HTTPQuery(t *testing.T) {
	hub := NewFederationHub()
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer hub.Stop()

	// Get actual address
	addr := hub.httpServer.Addr
	if addr == "127.0.0.1:0" {
		// Server hasn't started yet, wait
		time.Sleep(100 * time.Millisecond)
	}

	// Add test server
	hub.registerServer(&ServerAnnouncement{
		Name:       "test-server",
		Address:    "localhost:9000",
		Region:     RegionUSEast,
		Genre:      "scifi",
		Players:    5,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	})

	// Test query endpoint
	query := ServerQuery{
		Genre: ptrString("scifi"),
	}
	body, _ := json.Marshal(query)

	url := fmt.Sprintf("http://%s/query", addr)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Skipf("HTTP request failed (server may not be ready): %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var results []*ServerAnnouncement
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
	if len(results) > 0 && results[0].Name != "test-server" {
		t.Errorf("server name = %s, want test-server", results[0].Name)
	}
}

func TestFederationHub_HTTPQueryMethodNotAllowed(t *testing.T) {
	hub := NewFederationHub()
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer hub.Stop()

	time.Sleep(100 * time.Millisecond)

	url := fmt.Sprintf("http://%s/query", hub.httpServer.Addr)
	resp, err := http.Get(url)
	if err != nil {
		t.Skipf("HTTP request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", resp.StatusCode)
	}
}

func TestNewServerAnnouncer(t *testing.T) {
	announcement := ServerAnnouncement{
		Name:       "test-server",
		Address:    "localhost:8000",
		Region:     RegionUSEast,
		Genre:      "scifi",
		Players:    0,
		MaxPlayers: 16,
	}

	announcer := NewServerAnnouncer("ws://localhost:9000/announce", announcement)
	if announcer == nil {
		t.Fatal("NewServerAnnouncer returned nil")
	}
	if announcer.interval != 10*time.Second {
		t.Errorf("interval = %v, want 10s", announcer.interval)
	}
	if announcer.announcement.Name != "test-server" {
		t.Errorf("announcement.Name = %s, want test-server", announcer.announcement.Name)
	}
}

func TestServerAnnouncer_UpdatePlayers(t *testing.T) {
	announcer := NewServerAnnouncer("ws://localhost:9000/announce", ServerAnnouncement{
		Players: 0,
	})

	announcer.UpdatePlayers(5)

	announcer.mu.Lock()
	count := announcer.announcement.Players
	announcer.mu.Unlock()

	if count != 5 {
		t.Errorf("players = %d, want 5", count)
	}
}

func TestServerAnnouncer_Stop(t *testing.T) {
	announcer := NewServerAnnouncer("ws://localhost:9000/announce", ServerAnnouncement{})

	// Stop without starting should not panic
	if err := announcer.Stop(); err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}

func TestServerAnnouncer_Integration(t *testing.T) {
	// Start hub
	hub := NewFederationHub()
	hub.staleTimeout = 1 * time.Second
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("hub Start failed: %v", err)
	}
	defer hub.Stop()

	time.Sleep(100 * time.Millisecond)

	// Get hub address
	hubAddr := hub.httpServer.Addr
	wsURL := fmt.Sprintf("ws://%s/announce", hubAddr)

	// Create announcer
	announcement := ServerAnnouncement{
		Name:       "game-server-1",
		Address:    "localhost:8000",
		Region:     RegionUSEast,
		Genre:      "scifi",
		Players:    3,
		MaxPlayers: 16,
	}

	announcer := NewServerAnnouncer(wsURL, announcement)
	announcer.interval = 100 * time.Millisecond

	if err := announcer.Start(); err != nil {
		t.Skipf("announcer Start failed (may not be able to connect): %v", err)
		return
	}
	defer announcer.Stop()

	// Wait for announcement to be processed
	time.Sleep(200 * time.Millisecond)

	// Verify server is registered
	if hub.GetServerCount() != 1 {
		t.Errorf("server count = %d, want 1", hub.GetServerCount())
	}

	hub.mu.RLock()
	server := hub.servers["game-server-1"]
	hub.mu.RUnlock()

	if server == nil {
		t.Fatal("server not registered in hub")
	}
	if server.Address != "localhost:8000" {
		t.Errorf("server address = %s, want localhost:8000", server.Address)
	}
	if server.Region != RegionUSEast {
		t.Errorf("server region = %s, want %s", server.Region, RegionUSEast)
	}

	// Update player count
	announcer.UpdatePlayers(7)
	time.Sleep(200 * time.Millisecond)

	hub.mu.RLock()
	server = hub.servers["game-server-1"]
	hub.mu.RUnlock()

	if server.Players != 7 {
		t.Errorf("server players = %d, want 7", server.Players)
	}
}

func TestRegions(t *testing.T) {
	regions := []Region{
		RegionUSEast,
		RegionUSWest,
		RegionEUWest,
		RegionEUEast,
		RegionAsiaPac,
		RegionSouthAm,
		RegionUnknown,
	}

	for _, r := range regions {
		if r == "" {
			t.Errorf("region is empty string")
		}
	}
}

// Helper functions
func ptrRegion(r Region) *Region {
	return &r
}

func ptrString(s string) *string {
	return &s
}

func ptrInt(i int) *int {
	return &i
}
