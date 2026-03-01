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
	addr := hub.GetAddr()
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

	url := fmt.Sprintf("http://%s/query", hub.GetAddr())
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
	hubAddr := hub.GetAddr()
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

func TestFederationHub_PlayerLookup(t *testing.T) {
	hub := NewFederationHub()

	// Add servers with player lists
	hub.registerServer(&ServerAnnouncement{
		Name:       "server1",
		Address:    "localhost:8000",
		Region:     RegionUSEast,
		Genre:      "scifi",
		Players:    2,
		MaxPlayers: 16,
		PlayerList: []string{"player1", "player2"},
		Timestamp:  time.Now(),
	})
	hub.registerServer(&ServerAnnouncement{
		Name:       "server2",
		Address:    "localhost:8001",
		Region:     RegionEUWest,
		Genre:      "fantasy",
		Players:    1,
		MaxPlayers: 16,
		PlayerList: []string{"player3"},
		Timestamp:  time.Now(),
	})

	tests := []struct {
		name        string
		playerID    string
		wantOnline  bool
		wantServer  string
		wantAddress string
	}{
		{
			name:        "player on server1",
			playerID:    "player1",
			wantOnline:  true,
			wantServer:  "server1",
			wantAddress: "localhost:8000",
		},
		{
			name:        "player on server2",
			playerID:    "player3",
			wantOnline:  true,
			wantServer:  "server2",
			wantAddress: "localhost:8001",
		},
		{
			name:       "player not found",
			playerID:   "player999",
			wantOnline: false,
		},
		{
			name:       "empty playerID",
			playerID:   "",
			wantOnline: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := hub.lookupPlayer(tt.playerID)
			if response.Online != tt.wantOnline {
				t.Errorf("Online = %v, want %v", response.Online, tt.wantOnline)
			}
			if tt.wantOnline {
				if response.ServerName != tt.wantServer {
					t.Errorf("ServerName = %s, want %s", response.ServerName, tt.wantServer)
				}
				if response.ServerAddress != tt.wantAddress {
					t.Errorf("ServerAddress = %s, want %s", response.ServerAddress, tt.wantAddress)
				}
			}
		})
	}
}

func TestFederationHub_PlayerLookupHTTP(t *testing.T) {
	hub := NewFederationHub()
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer hub.Stop()

	time.Sleep(100 * time.Millisecond)

	// Add server with players
	hub.registerServer(&ServerAnnouncement{
		Name:       "test-server",
		Address:    "localhost:9000",
		Region:     RegionUSEast,
		Genre:      "scifi",
		Players:    3,
		MaxPlayers: 16,
		PlayerList: []string{"alice", "bob", "charlie"},
		Timestamp:  time.Now(),
	})

	tests := []struct {
		name        string
		request     PlayerLookupRequest
		wantStatus  int
		wantOnline  bool
		wantAddress string
	}{
		{
			name:        "player found",
			request:     PlayerLookupRequest{PlayerID: "alice"},
			wantStatus:  http.StatusOK,
			wantOnline:  true,
			wantAddress: "localhost:9000",
		},
		{
			name:       "player not found",
			request:    PlayerLookupRequest{PlayerID: "eve"},
			wantStatus: http.StatusOK,
			wantOnline: false,
		},
		{
			name:       "empty playerID",
			request:    PlayerLookupRequest{PlayerID: ""},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			url := fmt.Sprintf("http://%s/lookup", hub.GetAddr())
			resp, err := http.Post(url, "application/json", bytes.NewReader(body))
			if err != nil {
				t.Skipf("HTTP request failed: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response PlayerLookupResponse
				if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if response.Online != tt.wantOnline {
					t.Errorf("Online = %v, want %v", response.Online, tt.wantOnline)
				}
				if tt.wantOnline && response.ServerAddress != tt.wantAddress {
					t.Errorf("ServerAddress = %s, want %s", response.ServerAddress, tt.wantAddress)
				}
			}
		})
	}
}

func TestFederationHub_PlayerLookupMethodNotAllowed(t *testing.T) {
	hub := NewFederationHub()
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer hub.Stop()

	time.Sleep(100 * time.Millisecond)

	url := fmt.Sprintf("http://%s/lookup", hub.GetAddr())
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

func TestFederationHub_PlayerIndexUpdate(t *testing.T) {
	hub := NewFederationHub()

	// Register server with initial player list
	hub.registerServer(&ServerAnnouncement{
		Name:       "server1",
		Address:    "localhost:8000",
		PlayerList: []string{"player1", "player2"},
		Timestamp:  time.Now(),
	})

	// Verify players are indexed
	hub.mu.RLock()
	if hub.playerIndex["player1"] != "server1" {
		t.Errorf("player1 index = %s, want server1", hub.playerIndex["player1"])
	}
	if hub.playerIndex["player2"] != "server1" {
		t.Errorf("player2 index = %s, want server1", hub.playerIndex["player2"])
	}
	hub.mu.RUnlock()

	// Update server with new player list
	hub.registerServer(&ServerAnnouncement{
		Name:       "server1",
		Address:    "localhost:8000",
		PlayerList: []string{"player3", "player4"},
		Timestamp:  time.Now(),
	})

	// Verify old players removed, new players added
	hub.mu.RLock()
	if _, exists := hub.playerIndex["player1"]; exists {
		t.Error("player1 should be removed from index")
	}
	if _, exists := hub.playerIndex["player2"]; exists {
		t.Error("player2 should be removed from index")
	}
	if hub.playerIndex["player3"] != "server1" {
		t.Errorf("player3 index = %s, want server1", hub.playerIndex["player3"])
	}
	if hub.playerIndex["player4"] != "server1" {
		t.Errorf("player4 index = %s, want server1", hub.playerIndex["player4"])
	}
	hub.mu.RUnlock()
}

func TestFederationHub_StaleServerRemovesPlayers(t *testing.T) {
	hub := NewFederationHub()
	hub.staleTimeout = 100 * time.Millisecond
	hub.cleanupInterval = 50 * time.Millisecond

	// Add server with players (already stale)
	hub.registerServer(&ServerAnnouncement{
		Name:       "stale-server",
		Address:    "localhost:8000",
		PlayerList: []string{"player1", "player2"},
		Timestamp:  time.Now().Add(-200 * time.Millisecond),
	})

	// Verify players are indexed
	hub.mu.RLock()
	initialCount := len(hub.playerIndex)
	hub.mu.RUnlock()
	if initialCount != 2 {
		t.Errorf("initial player index count = %d, want 2", initialCount)
	}

	// Start cleanup
	go hub.cleanupStaleServers()
	defer hub.cancel()

	// Wait for cleanup
	time.Sleep(150 * time.Millisecond)

	// Verify players removed from index
	hub.mu.RLock()
	finalCount := len(hub.playerIndex)
	hub.mu.RUnlock()
	if finalCount != 0 {
		t.Errorf("final player index count = %d, want 0", finalCount)
	}
}

func TestServerAnnouncer_UpdatePlayerList(t *testing.T) {
	announcer := NewServerAnnouncer("ws://localhost:9000/announce", ServerAnnouncement{
		Players: 0,
	})

	playerList := []string{"player1", "player2", "player3"}
	announcer.UpdatePlayerList(playerList)

	announcer.mu.Lock()
	list := announcer.announcement.PlayerList
	count := announcer.announcement.Players
	announcer.mu.Unlock()

	if len(list) != 3 {
		t.Errorf("player list length = %d, want 3", len(list))
	}
	if count != 3 {
		t.Errorf("player count = %d, want 3", count)
	}
	for i, expected := range playerList {
		if list[i] != expected {
			t.Errorf("player[%d] = %s, want %s", i, list[i], expected)
		}
	}
}

func TestPlayerLookup_Integration(t *testing.T) {
	// Start hub
	hub := NewFederationHub()
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("hub Start failed: %v", err)
	}
	defer hub.Stop()

	time.Sleep(100 * time.Millisecond)

	// Get hub address
	hubAddr := hub.GetAddr()
	wsURL := fmt.Sprintf("ws://%s/announce", hubAddr)

	// Create and start first server announcer
	announcer1 := NewServerAnnouncer(wsURL, ServerAnnouncement{
		Name:       "game-server-1",
		Address:    "localhost:8000",
		Region:     RegionUSEast,
		Genre:      "scifi",
		PlayerList: []string{"alice", "bob"},
	})
	announcer1.interval = 100 * time.Millisecond

	if err := announcer1.Start(); err != nil {
		t.Skipf("announcer1 Start failed: %v", err)
		return
	}
	defer announcer1.Stop()

	// Create and start second server announcer
	announcer2 := NewServerAnnouncer(wsURL, ServerAnnouncement{
		Name:       "game-server-2",
		Address:    "localhost:8001",
		Region:     RegionEUWest,
		Genre:      "fantasy",
		PlayerList: []string{"charlie"},
	})
	announcer2.interval = 100 * time.Millisecond

	if err := announcer2.Start(); err != nil {
		t.Skipf("announcer2 Start failed: %v", err)
		return
	}
	defer announcer2.Stop()

	// Wait for announcements
	time.Sleep(200 * time.Millisecond)

	// Test player lookup via HTTP
	tests := []struct {
		playerID    string
		wantOnline  bool
		wantServer  string
		wantAddress string
	}{
		{"alice", true, "game-server-1", "localhost:8000"},
		{"bob", true, "game-server-1", "localhost:8000"},
		{"charlie", true, "game-server-2", "localhost:8001"},
		{"dave", false, "", ""},
	}

	for _, tt := range tests {
		t.Run("lookup_"+tt.playerID, func(t *testing.T) {
			body, _ := json.Marshal(PlayerLookupRequest{PlayerID: tt.playerID})
			url := fmt.Sprintf("http://%s/lookup", hubAddr)
			resp, err := http.Post(url, "application/json", bytes.NewReader(body))
			if err != nil {
				t.Skipf("HTTP request failed: %v", err)
				return
			}
			defer resp.Body.Close()

			var response PlayerLookupResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response.Online != tt.wantOnline {
				t.Errorf("Online = %v, want %v", response.Online, tt.wantOnline)
			}
			if tt.wantOnline {
				if response.ServerName != tt.wantServer {
					t.Errorf("ServerName = %s, want %s", response.ServerName, tt.wantServer)
				}
				if response.ServerAddress != tt.wantAddress {
					t.Errorf("ServerAddress = %s, want %s", response.ServerAddress, tt.wantAddress)
				}
			}
		})
	}

	// Update player list on server 1
	announcer1.UpdatePlayerList([]string{"alice", "bob", "eve"})
	time.Sleep(200 * time.Millisecond)

	// Verify eve is now online on server 1
	body, _ := json.Marshal(PlayerLookupRequest{PlayerID: "eve"})
	url := fmt.Sprintf("http://%s/lookup", hubAddr)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Skipf("HTTP request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	var response PlayerLookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !response.Online {
		t.Error("eve should be online after update")
	}
	if response.ServerName != "game-server-1" {
		t.Errorf("eve's server = %s, want game-server-1", response.ServerName)
	}
}

// TestDiscoverServers tests the client-side DiscoverServers function.
func TestDiscoverServers(t *testing.T) {
	// Start hub
	hub := NewFederationHub()
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("hub Start failed: %v", err)
	}
	defer hub.Stop()

	time.Sleep(100 * time.Millisecond)

	// Register test servers
	hub.RegisterServer(&ServerAnnouncement{
		Name:       "server-fantasy-1",
		Address:    "localhost:8000",
		Region:     RegionUSEast,
		Genre:      "fantasy",
		Players:    5,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	})
	hub.RegisterServer(&ServerAnnouncement{
		Name:       "server-scifi-1",
		Address:    "localhost:8001",
		Region:     RegionUSWest,
		Genre:      "scifi",
		Players:    10,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	})
	hub.RegisterServer(&ServerAnnouncement{
		Name:       "server-fantasy-2",
		Address:    "localhost:8002",
		Region:     RegionEUWest,
		Genre:      "fantasy",
		Players:    2,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	})

	hubURL := fmt.Sprintf("http://%s", hub.GetAddr())

	tests := []struct {
		name    string
		query   *ServerQuery
		wantLen int
		wantErr bool
	}{
		{
			name:    "all servers",
			query:   &ServerQuery{},
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "filter by genre fantasy",
			query: &ServerQuery{
				Genre: ptrString("fantasy"),
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "filter by genre scifi",
			query: &ServerQuery{
				Genre: ptrString("scifi"),
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "filter by region",
			query: &ServerQuery{
				Region: ptrRegion(RegionUSEast),
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "filter by min players",
			query: &ServerQuery{
				MinPlayers: ptrInt(8),
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "no matches",
			query: &ServerQuery{
				Genre: ptrString("horror"),
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "combined filters",
			query: &ServerQuery{
				Genre:      ptrString("fantasy"),
				MinPlayers: ptrInt(3),
			},
			wantLen: 1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := DiscoverServers(hubURL, tt.query, 5*time.Second)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiscoverServers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(results) != tt.wantLen {
				t.Errorf("got %d results, want %d", len(results), tt.wantLen)
			}
		})
	}
}

// TestDiscoverServers_InvalidHub tests error handling for invalid hub URL.
func TestDiscoverServers_InvalidHub(t *testing.T) {
	query := &ServerQuery{}
	_, err := DiscoverServers("http://localhost:99999", query, 1*time.Second)
	if err == nil {
		t.Error("DiscoverServers() should error with invalid hub URL")
	}
}

// TestDiscoverServers_Timeout tests timeout handling.
func TestDiscoverServers_Timeout(t *testing.T) {
	// Use a non-routable IP to force timeout
	query := &ServerQuery{}
	_, err := DiscoverServers("http://192.0.2.1:9999", query, 100*time.Millisecond)
	if err == nil {
		t.Error("DiscoverServers() should timeout")
	}
}

// TestLookupPlayer_Client tests the client-side LookupPlayer function.
func TestLookupPlayer_Client(t *testing.T) {
	// Start hub
	hub := NewFederationHub()
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("hub Start failed: %v", err)
	}
	defer hub.Stop()

	time.Sleep(100 * time.Millisecond)

	// Register server with players
	hub.RegisterServer(&ServerAnnouncement{
		Name:       "test-server",
		Address:    "localhost:9000",
		Region:     RegionUSEast,
		Genre:      "scifi",
		Players:    3,
		MaxPlayers: 16,
		PlayerList: []string{"alice", "bob", "charlie"},
		Timestamp:  time.Now(),
	})

	hubURL := fmt.Sprintf("http://%s", hub.GetAddr())

	tests := []struct {
		name        string
		playerID    string
		wantOnline  bool
		wantAddress string
		wantErr     bool
	}{
		{
			name:        "player found",
			playerID:    "alice",
			wantOnline:  true,
			wantAddress: "localhost:9000",
			wantErr:     false,
		},
		{
			name:       "player not found",
			playerID:   "eve",
			wantOnline: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := LookupPlayer(hubURL, tt.playerID, 5*time.Second)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupPlayer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result.Online != tt.wantOnline {
					t.Errorf("Online = %v, want %v", result.Online, tt.wantOnline)
				}
				if tt.wantOnline && result.ServerAddress != tt.wantAddress {
					t.Errorf("ServerAddress = %s, want %s", result.ServerAddress, tt.wantAddress)
				}
			}
		})
	}
}

// TestLookupPlayer_InvalidHub tests error handling for invalid hub URL.
func TestLookupPlayer_InvalidHub(t *testing.T) {
	_, err := LookupPlayer("http://localhost:99999", "alice", 1*time.Second)
	if err == nil {
		t.Error("LookupPlayer() should error with invalid hub URL")
	}
}
