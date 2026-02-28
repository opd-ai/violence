// Package federation provides server federation and matchmaking.
package federation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestFederationIntegration_MultiServerWithLookupAndMatchmaking tests the full
// federation system: multiple servers announcing, player lookup, and matchmaking.
func TestFederationIntegration_MultiServerWithLookupAndMatchmaking(t *testing.T) {
	// Start federation hub
	hub := NewFederationHub()
	err := hub.Start("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}
	defer hub.Stop()

	hubAddr := hub.httpServer.Addr
	hubURL := fmt.Sprintf("ws://%s/announce", hubAddr)
	queryURL := fmt.Sprintf("http://%s/query", hubAddr)
	lookupURL := fmt.Sprintf("http://%s/lookup", hubAddr)

	// Give hub time to start
	time.Sleep(100 * time.Millisecond)

	// Create 3 game servers with different configurations
	server1 := NewServerAnnouncer(hubURL, ServerAnnouncement{
		Name:       "server1",
		Address:    "game1.example.com:7777",
		Region:     RegionUSEast,
		Genre:      "fantasy",
		Players:    0,
		MaxPlayers: 16,
		PlayerList: []string{},
	})
	server1.SetInterval(200 * time.Millisecond) // Fast interval for testing

	server2 := NewServerAnnouncer(hubURL, ServerAnnouncement{
		Name:       "server2",
		Address:    "game2.example.com:7777",
		Region:     RegionUSWest,
		Genre:      "scifi",
		Players:    0,
		MaxPlayers: 16,
		PlayerList: []string{},
	})
	server2.SetInterval(200 * time.Millisecond)

	server3 := NewServerAnnouncer(hubURL, ServerAnnouncement{
		Name:       "server3",
		Address:    "game3.example.com:7777",
		Region:     RegionUSEast,
		Genre:      "fantasy",
		Players:    0,
		MaxPlayers: 16,
		PlayerList: []string{},
	})
	server3.SetInterval(200 * time.Millisecond)

	// Start all servers
	if err := server1.Start(); err != nil {
		t.Fatalf("failed to start server1 announcer: %v", err)
	}
	defer server1.Stop()

	if err := server2.Start(); err != nil {
		t.Fatalf("failed to start server2 announcer: %v", err)
	}
	defer server2.Stop()

	if err := server3.Start(); err != nil {
		t.Fatalf("failed to start server3 announcer: %v", err)
	}
	defer server3.Stop()

	// Wait for announcements to propagate
	time.Sleep(200 * time.Millisecond)

	// Verify all servers are registered
	if hub.GetServerCount() != 3 {
		t.Errorf("expected 3 servers, got %d", hub.GetServerCount())
	}

	// Test 1: Query servers by region
	t.Run("QueryByRegion", func(t *testing.T) {
		region := RegionUSEast
		query := ServerQuery{Region: &region}
		servers := queryServers(t, queryURL, query)

		if len(servers) != 2 {
			t.Errorf("expected 2 US-East servers, got %d", len(servers))
		}

		// Verify both US-East servers are returned
		found := make(map[string]bool)
		for _, s := range servers {
			found[s.Name] = true
		}
		if !found["server1"] || !found["server3"] {
			t.Errorf("expected server1 and server3, got: %v", found)
		}
	})

	// Test 2: Query servers by genre
	t.Run("QueryByGenre", func(t *testing.T) {
		genre := "fantasy"
		query := ServerQuery{Genre: &genre}
		servers := queryServers(t, queryURL, query)

		if len(servers) != 2 {
			t.Errorf("expected 2 fantasy servers, got %d", len(servers))
		}
	})

	// Test 3: Query servers by region AND genre
	t.Run("QueryByRegionAndGenre", func(t *testing.T) {
		region := RegionUSEast
		genre := "fantasy"
		query := ServerQuery{Region: &region, Genre: &genre}
		servers := queryServers(t, queryURL, query)

		if len(servers) != 2 {
			t.Errorf("expected 2 servers (US-East fantasy), got %d", len(servers))
		}
	})

	// Test 4: Add players to servers and test player lookup
	t.Run("PlayerLookup", func(t *testing.T) {
		// Update server1 with player list
		server1.UpdatePlayerList([]string{"player1", "player2"})
		// Wait for next announcement cycle
		time.Sleep(400 * time.Millisecond)

		// Update server2 with different players
		server2.UpdatePlayerList([]string{"player3", "player4"})
		time.Sleep(400 * time.Millisecond)

		// Lookup player1 (should be on server1)
		response := lookupPlayer(t, lookupURL, "player1")
		if !response.Online {
			t.Errorf("expected player1 to be online")
		}
		if response.ServerName != "server1" {
			t.Errorf("expected player1 on server1, got %s", response.ServerName)
		}
		if response.ServerAddress != "game1.example.com:7777" {
			t.Errorf("expected address game1.example.com:7777, got %s", response.ServerAddress)
		}

		// Lookup player3 (should be on server2)
		response = lookupPlayer(t, lookupURL, "player3")
		if !response.Online {
			t.Errorf("expected player3 to be online")
		}
		if response.ServerName != "server2" {
			t.Errorf("expected player3 on server2, got %s", response.ServerName)
		}

		// Lookup non-existent player
		response = lookupPlayer(t, lookupURL, "player999")
		if response.Online {
			t.Errorf("expected player999 to be offline")
		}
		if response.ServerAddress != "" {
			t.Errorf("expected empty server address for offline player")
		}
	})

	// Test 5: Matchmaking integration
	t.Run("Matchmaking", func(t *testing.T) {
		// Create matchmaker
		matchmaker := NewMatchmaker(hub)
		matchmaker.Start()
		defer matchmaker.Stop()

		// Enqueue 4 players for FFA fantasy in US-East
		for i := 1; i <= 4; i++ {
			err := matchmaker.Enqueue(
				fmt.Sprintf("ffa_player%d", i),
				ModeFFA,
				"fantasy",
				RegionUSEast,
			)
			if err != nil {
				t.Fatalf("failed to enqueue player: %v", err)
			}
		}

		// Wait for matchmaking to process
		time.Sleep(3 * time.Second)

		// Queue should be empty (players matched)
		queueSize := matchmaker.GetQueueSize(ModeFFA)
		if queueSize != 0 {
			t.Errorf("expected queue to be empty after match, got %d players", queueSize)
		}
	})

	// Test 6: Cross-server player migration
	t.Run("PlayerMigration", func(t *testing.T) {
		// Player moves from server1 to server2
		server1.UpdatePlayerList([]string{"migrating_player"})
		time.Sleep(400 * time.Millisecond)

		// Verify player on server1
		response := lookupPlayer(t, lookupURL, "migrating_player")
		if response.ServerName != "server1" {
			t.Errorf("expected player on server1, got %s", response.ServerName)
		}

		// Player moves to server2
		server1.UpdatePlayerList([]string{}) // Remove from server1
		server2.UpdatePlayerList([]string{"migrating_player"})
		time.Sleep(400 * time.Millisecond)

		// Verify player on server2
		response = lookupPlayer(t, lookupURL, "migrating_player")
		if response.ServerName != "server2" {
			t.Errorf("expected player on server2 after migration, got %s", response.ServerName)
		}
	})

	// Test 7: Server capacity filtering
	t.Run("ServerCapacity", func(t *testing.T) {
		// Fill server1 to capacity
		fullPlayerList := make([]string, 16)
		for i := 0; i < 16; i++ {
			fullPlayerList[i] = fmt.Sprintf("capacity_player%d", i)
		}
		server1.UpdatePlayerList(fullPlayerList)
		time.Sleep(400 * time.Millisecond)

		// Create matchmaker and try to match players
		matchmaker := NewMatchmaker(hub)
		matchmaker.Start()
		defer matchmaker.Stop()

		// Enqueue 4 players for fantasy US-East (server1 is full, should use server3)
		for i := 1; i <= 4; i++ {
			matchmaker.Enqueue(
				fmt.Sprintf("capacity_test_player%d", i),
				ModeCoop,
				"fantasy",
				RegionUSEast,
			)
		}

		// Wait for matchmaking
		time.Sleep(3 * time.Second)

		// Queue should be empty (matched to server3)
		queueSize := matchmaker.GetQueueSize(ModeCoop)
		if queueSize != 0 {
			t.Errorf("expected queue to be empty (matched to server3), got %d players", queueSize)
		}
	})

	// Test 8: Stale server cleanup
	t.Run("StaleServerCleanup", func(t *testing.T) {
		// Stop server1 announcements (simulate server going offline)
		server1.Stop()

		// Set stale timeout to a shorter value
		hub.SetStaleTimeout(500 * time.Millisecond)

		// Wait for server to become stale + cleanup to trigger
		// Cleanup runs every 10 seconds by default, so we wait for one cleanup cycle
		time.Sleep(11 * time.Second)

		// Server1 should be removed
		serverCount := hub.GetServerCount()
		if serverCount != 2 {
			t.Errorf("expected 2 servers after cleanup, got %d", serverCount)
		}

		// Player lookup should fail for players on removed server
		response := lookupPlayer(t, lookupURL, "capacity_player0")
		if response.Online {
			t.Errorf("expected player to be offline after server cleanup")
		}
	})
}

// TestFederationIntegration_MatchmakingModes tests matchmaking across all game modes.
func TestFederationIntegration_MatchmakingModes(t *testing.T) {
	hub := NewFederationHub()
	err := hub.Start("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}
	defer hub.Stop()

	time.Sleep(100 * time.Millisecond)

	// Register a server for each mode
	hubAddr := hub.httpServer.Addr
	hubURL := fmt.Sprintf("ws://%s/announce", hubAddr)

	modes := []struct {
		name    string
		mode    GameMode
		minSize int
		maxSize int
	}{
		{"co-op", ModeCoop, 2, 4},
		{"ffa", ModeFFA, 2, 8},
		{"team-dm", ModeTeamDM, 2, 16},
		{"territory", ModeTerritory, 2, 16},
	}

	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			// Create server for this mode
			server := NewServerAnnouncer(hubURL, ServerAnnouncement{
				Name:       fmt.Sprintf("%s-server", m.name),
				Address:    fmt.Sprintf("%s.example.com:7777", m.name),
				Region:     RegionUSEast,
				Genre:      "scifi",
				Players:    0,
				MaxPlayers: m.maxSize,
			})

			if err := server.Start(); err != nil {
				t.Fatalf("failed to start server: %v", err)
			}
			defer server.Stop()

			time.Sleep(150 * time.Millisecond)

			// Create matchmaker
			matchmaker := NewMatchmaker(hub)
			matchmaker.Start()
			defer matchmaker.Stop()

			// Enqueue minimum players
			for i := 0; i < m.minSize; i++ {
				err := matchmaker.Enqueue(
					fmt.Sprintf("%s_player%d", m.name, i),
					m.mode,
					"scifi",
					RegionUSEast,
				)
				if err != nil {
					t.Fatalf("failed to enqueue player: %v", err)
				}
			}

			// Wait for matchmaking
			time.Sleep(3 * time.Second)

			// Verify match was created (queue empty)
			queueSize := matchmaker.GetQueueSize(m.mode)
			if queueSize != 0 {
				t.Errorf("expected queue to be empty after match, got %d players", queueSize)
			}
		})
	}
}

// TestFederationIntegration_ConcurrentAnnouncements tests hub stability under
// concurrent server announcements.
func TestFederationIntegration_ConcurrentAnnouncements(t *testing.T) {
	hub := NewFederationHub()
	err := hub.Start("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}
	defer hub.Stop()

	time.Sleep(100 * time.Millisecond)

	hubAddr := hub.httpServer.Addr
	hubURL := fmt.Sprintf("ws://%s/announce", hubAddr)

	// Create 10 servers announcing concurrently
	servers := make([]*ServerAnnouncer, 10)
	for i := 0; i < 10; i++ {
		servers[i] = NewServerAnnouncer(hubURL, ServerAnnouncement{
			Name:       fmt.Sprintf("concurrent-server%d", i),
			Address:    fmt.Sprintf("server%d.example.com:7777", i),
			Region:     RegionUSEast,
			Genre:      "fantasy",
			Players:    i,
			MaxPlayers: 16,
		})

		if err := servers[i].Start(); err != nil {
			t.Fatalf("failed to start server %d: %v", i, err)
		}
		defer servers[i].Stop()
	}

	// Wait for announcements
	time.Sleep(300 * time.Millisecond)

	// Verify all servers registered
	if hub.GetServerCount() != 10 {
		t.Errorf("expected 10 servers, got %d", hub.GetServerCount())
	}

	// Update all servers concurrently with player lists
	for i := 0; i < 10; i++ {
		go func(idx int) {
			players := make([]string, idx)
			for j := 0; j < idx; j++ {
				players[j] = fmt.Sprintf("server%d_player%d", idx, j)
			}
			servers[idx].UpdatePlayerList(players)
		}(i)
	}

	// Wait for updates
	time.Sleep(300 * time.Millisecond)

	// Verify no data corruption
	if hub.GetServerCount() != 10 {
		t.Errorf("expected 10 servers after concurrent updates, got %d", hub.GetServerCount())
	}
}

// TestFederationIntegration_RegionBasedMatching tests that matchmaking respects
// region preferences.
func TestFederationIntegration_RegionBasedMatching(t *testing.T) {
	hub := NewFederationHub()
	err := hub.Start("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}
	defer hub.Stop()

	time.Sleep(100 * time.Millisecond)

	hubAddr := hub.httpServer.Addr
	hubURL := fmt.Sprintf("ws://%s/announce", hubAddr)

	// Create servers in different regions
	regions := []Region{RegionUSEast, RegionUSWest, RegionEUWest, RegionAsiaPac}
	for _, region := range regions {
		server := NewServerAnnouncer(hubURL, ServerAnnouncement{
			Name:       fmt.Sprintf("region-server-%s", region),
			Address:    fmt.Sprintf("%s.example.com:7777", region),
			Region:     region,
			Genre:      "fantasy",
			Players:    0,
			MaxPlayers: 16,
		})

		if err := server.Start(); err != nil {
			t.Fatalf("failed to start server: %v", err)
		}
		defer server.Stop()
	}

	time.Sleep(200 * time.Millisecond)

	// Create matchmaker
	matchmaker := NewMatchmaker(hub)
	matchmaker.Start()
	defer matchmaker.Stop()

	// Enqueue players for different regions
	for _, region := range regions {
		for j := 0; j < 2; j++ {
			err := matchmaker.Enqueue(
				fmt.Sprintf("region_%s_player%d", region, j),
				ModeFFA,
				"fantasy",
				region,
			)
			if err != nil {
				t.Fatalf("failed to enqueue player %d for region %s: %v", j, region, err)
			}
		}
	}

	// Wait for matchmaking
	time.Sleep(3 * time.Second)

	// All queues should be empty (players matched in their regions)
	queueSize := matchmaker.GetQueueSize(ModeFFA)
	if queueSize != 0 {
		t.Errorf("expected all queues to be empty, got %d players remaining", queueSize)
	}
}

// Helper function to query servers via HTTP
func queryServers(t *testing.T, url string, query ServerQuery) []*ServerAnnouncement {
	data, err := json.Marshal(query)
	if err != nil {
		t.Fatalf("failed to marshal query: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("failed to query servers: %v", err)
	}
	defer resp.Body.Close()

	var servers []*ServerAnnouncement
	if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return servers
}

// Helper function to lookup player via HTTP
func lookupPlayer(t *testing.T, url, playerID string) PlayerLookupResponse {
	req := PlayerLookupRequest{PlayerID: playerID}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal lookup request: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("failed to lookup player: %v", err)
	}
	defer resp.Body.Close()

	var response PlayerLookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return response
}
