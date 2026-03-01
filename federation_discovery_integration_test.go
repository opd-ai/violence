package main

import (
	"testing"
	"time"

	"github.com/opd-ai/violence/pkg/config"
	"github.com/opd-ai/violence/pkg/federation"
)

// TestFederationDiscoveryIntegration tests the end-to-end flow of:
// 1. Starting a federation hub
// 2. Announcing servers to the hub
// 3. Discovering servers from a client using the new DiscoverServers function
func TestFederationDiscoveryIntegration(t *testing.T) {
	// Start a federation hub
	hub := federation.NewFederationHub()
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}
	defer hub.Stop()

	time.Sleep(100 * time.Millisecond) // Give server time to start

	hubURL := "http://" + hub.GetAddr()

	// Register some test servers directly (simulating server announcements)
	hub.RegisterServer(&federation.ServerAnnouncement{
		Name:       "fantasy-server-1",
		Address:    "game1.example.com:7777",
		Region:     federation.RegionUSEast,
		Genre:      "fantasy",
		Players:    5,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	})

	hub.RegisterServer(&federation.ServerAnnouncement{
		Name:       "scifi-server-1",
		Address:    "game2.example.com:7777",
		Region:     federation.RegionEUWest,
		Genre:      "scifi",
		Players:    10,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	})

	hub.RegisterServer(&federation.ServerAnnouncement{
		Name:       "fantasy-server-2",
		Address:    "game3.example.com:7777",
		Region:     federation.RegionUSWest,
		Genre:      "fantasy",
		Players:    2,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	})

	// Test discovery with various filters
	tests := []struct {
		name    string
		query   *federation.ServerQuery
		wantLen int
		wantErr bool
	}{
		{
			name:    "discover all servers",
			query:   &federation.ServerQuery{},
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "discover fantasy servers only",
			query: &federation.ServerQuery{
				Genre: ptrString("fantasy"),
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "discover scifi servers only",
			query: &federation.ServerQuery{
				Genre: ptrString("scifi"),
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "discover servers in US East",
			query: &federation.ServerQuery{
				Region: ptrRegion(federation.RegionUSEast),
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "discover servers with min 8 players",
			query: &federation.ServerQuery{
				MinPlayers: ptrInt(8),
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "no horror servers available",
			query: &federation.ServerQuery{
				Genre: ptrString("horror"),
			},
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			servers, err := federation.DiscoverServers(hubURL, tt.query, 5*time.Second)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiscoverServers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(servers) != tt.wantLen {
				t.Errorf("got %d servers, want %d", len(servers), tt.wantLen)
			}
		})
	}
}

// TestRefreshServerBrowserIntegration tests the refreshServerBrowser function
// with a real federation hub configured via config.
func TestRefreshServerBrowserIntegration(t *testing.T) {
	// Start a federation hub
	hub := federation.NewFederationHub()
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}
	defer hub.Stop()

	time.Sleep(100 * time.Millisecond)

	hubURL := "http://" + hub.GetAddr()

	// Register test servers
	hub.RegisterServer(&federation.ServerAnnouncement{
		Name:       "test-server-1",
		Address:    "server1.example.com:7777",
		Region:     federation.RegionUSEast,
		Genre:      "fantasy",
		Players:    5,
		MaxPlayers: 16,
		Timestamp:  time.Now(),
	})

	// Configure the game to use this hub
	originalURL := config.C.FederationHubURL
	config.C.FederationHubURL = hubURL
	defer func() { config.C.FederationHubURL = originalURL }()

	// Create a game instance
	game := NewGame()

	// Set genre to fantasy
	game.genreID = "fantasy"

	// Refresh server browser
	game.refreshServerBrowser()

	// Verify servers were discovered
	if len(game.serverBrowser) != 1 {
		t.Errorf("got %d servers in browser, want 1", len(game.serverBrowser))
	}

	if len(game.serverBrowser) > 0 && game.serverBrowser[0].Name != "test-server-1" {
		t.Errorf("server name = %s, want test-server-1", game.serverBrowser[0].Name)
	}

	if game.mpStatusMsg == "No servers found. Press R to refresh." {
		t.Error("status message indicates no servers found, but we registered one")
	}
}

// Helper functions
func ptrString(s string) *string {
	return &s
}

func ptrInt(i int) *int {
	return &i
}

func ptrRegion(r federation.Region) *federation.Region {
	return &r
}
