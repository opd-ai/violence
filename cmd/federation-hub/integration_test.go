package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/opd-ai/violence/pkg/federation"
)

// TestFederationHubIntegration tests the complete federation hub lifecycle:
// 1. Start hub server
// 2. Register servers via HTTP POST
// 3. Query servers
// 4. Player lookup
func TestFederationHubIntegration(t *testing.T) {
	// Start hub server
	hub := NewHubServer("", nil)
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}
	defer hub.Stop()

	time.Sleep(50 * time.Millisecond)

	hubAddr := hub.GetAddr()
	hubURL := "http://" + hubAddr

	// Test 1: Register servers via HTTP POST
	t.Run("server_registration", func(t *testing.T) {
		announcement := federation.ServerAnnouncement{
			Name:       "test-server-1",
			Address:    "game1.example.com:7777",
			Region:     federation.RegionUSEast,
			Genre:      "horror",
			Players:    5,
			MaxPlayers: 16,
			PlayerList: []string{"player1", "player2", "player3", "player4", "player5"},
		}

		data, _ := json.Marshal(announcement)
		resp, err := http.Post(hubURL+"/announce", "application/json", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("failed to register server: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status code = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		// Verify server count increased
		if hub.hub.GetServerCount() != 1 {
			t.Errorf("server count = %d, want 1", hub.hub.GetServerCount())
		}
	})

	// Test 2: Query servers
	t.Run("query_servers", func(t *testing.T) {
		// Register another server
		announcement := federation.ServerAnnouncement{
			Name:       "test-server-2",
			Address:    "game2.example.com:7777",
			Region:     federation.RegionEUWest,
			Genre:      "scifi",
			Players:    10,
			MaxPlayers: 16,
		}

		data, _ := json.Marshal(announcement)
		http.Post(hubURL+"/announce", "application/json", bytes.NewReader(data))

		// Query all servers
		query := federation.ServerQuery{}
		queryData, _ := json.Marshal(query)
		resp, err := http.Post(hubURL+"/query", "application/json", bytes.NewReader(queryData))
		if err != nil {
			t.Fatalf("failed to query servers: %v", err)
		}
		defer resp.Body.Close()

		var servers []*federation.ServerAnnouncement
		if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(servers) != 2 {
			t.Errorf("got %d servers, want 2", len(servers))
		}

		// Query by genre
		horrorGenre := "horror"
		query = federation.ServerQuery{Genre: &horrorGenre}
		queryData, _ = json.Marshal(query)
		resp, err = http.Post(hubURL+"/query", "application/json", bytes.NewReader(queryData))
		if err != nil {
			t.Fatalf("failed to query servers: %v", err)
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(servers) != 1 {
			t.Errorf("got %d horror servers, want 1", len(servers))
		}
	})

	// Test 3: Player lookup
	t.Run("player_lookup", func(t *testing.T) {
		lookupReq := federation.PlayerLookupRequest{PlayerID: "player1"}
		data, _ := json.Marshal(lookupReq)
		resp, err := http.Post(hubURL+"/lookup", "application/json", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("failed to lookup player: %v", err)
		}
		defer resp.Body.Close()

		var lookupResp federation.PlayerLookupResponse
		if err := json.NewDecoder(resp.Body).Decode(&lookupResp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !lookupResp.Online {
			t.Error("player1 should be online")
		}
		if lookupResp.ServerName != "test-server-1" {
			t.Errorf("server name = %q, want test-server-1", lookupResp.ServerName)
		}
	})

	// Test 4: Health check
	t.Run("health_check", func(t *testing.T) {
		resp, err := http.Get(hubURL + "/health")
		if err != nil {
			t.Fatalf("failed to get health: %v", err)
		}
		defer resp.Body.Close()

		var health HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if health.Status != "ok" {
			t.Errorf("status = %q, want ok", health.Status)
		}
		if health.ServerCount < 2 {
			t.Errorf("server count = %d, want >= 2", health.ServerCount)
		}
	})
}

// TestFederationHubPeering tests hub-to-hub synchronization.
func TestFederationHubPeering(t *testing.T) {
	// Start first hub
	hub1 := NewHubServer("", nil)
	if err := hub1.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start hub1: %v", err)
	}
	defer hub1.Stop()

	time.Sleep(50 * time.Millisecond)

	hub1URL := "http://" + hub1.GetAddr()

	// Register server on hub1
	announcement := federation.ServerAnnouncement{
		Name:       "hub1-server",
		Address:    "game1.example.com:7777",
		Region:     federation.RegionUSEast,
		Genre:      "horror",
		Players:    5,
		MaxPlayers: 16,
	}

	data, _ := json.Marshal(announcement)
	http.Post(hub1URL+"/announce", "application/json", bytes.NewReader(data))

	// Start second hub with hub1 as peer
	hub2 := NewHubServer("", []string{hub1URL})
	if err := hub2.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start hub2: %v", err)
	}
	defer hub2.Stop()

	time.Sleep(50 * time.Millisecond)

	// Manually trigger sync (normally happens every 5 minutes)
	hub2.syncWithPeer(hub1URL)

	// Verify hub2 has the server from hub1
	servers := hub2.hub.QueryServers(&federation.ServerQuery{})
	if len(servers) != 1 {
		t.Errorf("hub2 has %d servers, want 1", len(servers))
	}

	if len(servers) > 0 && servers[0].Name != "hub1-server" {
		t.Errorf("server name = %q, want hub1-server", servers[0].Name)
	}
}

// TestFederationHubWithAuth tests authentication for server registration.
func TestFederationHubWithAuth(t *testing.T) {
	hub := NewHubServer("secret-token-123", nil)
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}
	defer hub.Stop()

	time.Sleep(50 * time.Millisecond)

	hubURL := "http://" + hub.GetAddr()

	announcement := federation.ServerAnnouncement{
		Name:       "auth-server",
		Address:    "game.example.com:7777",
		Region:     federation.RegionUSEast,
		Genre:      "horror",
		Players:    5,
		MaxPlayers: 16,
	}

	data, _ := json.Marshal(announcement)

	// Test without auth - should fail
	t.Run("without_auth_fails", func(t *testing.T) {
		resp, err := http.Post(hubURL+"/announce", "application/json", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("failed to post: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status code = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	// Test with correct auth - should succeed
	t.Run("with_correct_auth_succeeds", func(t *testing.T) {
		req, _ := http.NewRequest("POST", hubURL+"/announce", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer secret-token-123")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to post: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status code = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})
}
