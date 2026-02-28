package federation

import (
	"fmt"
	"testing"
	"time"
)

func TestNewMatchmaker(t *testing.T) {
	hub := NewFederationHub()
	mm := NewMatchmaker(hub)

	if mm == nil {
		t.Fatal("expected non-nil matchmaker")
	}
	if mm.hub != hub {
		t.Error("hub not set correctly")
	}
	if mm.matchInterval != 2*time.Second {
		t.Errorf("expected match interval 2s, got %v", mm.matchInterval)
	}
	if mm.queueTimeout != 60*time.Second {
		t.Errorf("expected queue timeout 60s, got %v", mm.queueTimeout)
	}

	// Verify min/max players per mode
	tests := []struct {
		mode       GameMode
		minPlayers int
		maxPlayers int
	}{
		{ModeCoop, 2, 4},
		{ModeFFA, 2, 8},
		{ModeTeamDM, 2, 16},
		{ModeTerritory, 2, 16},
	}

	for _, tt := range tests {
		if min := mm.minPlayersPerMode[tt.mode]; min != tt.minPlayers {
			t.Errorf("mode %s: expected min players %d, got %d", tt.mode, tt.minPlayers, min)
		}
		if max := mm.maxPlayersPerMode[tt.mode]; max != tt.maxPlayers {
			t.Errorf("mode %s: expected max players %d, got %d", tt.mode, tt.maxPlayers, max)
		}
	}
}

func TestMatchmaker_Enqueue(t *testing.T) {
	hub := NewFederationHub()
	mm := NewMatchmaker(hub)

	tests := []struct {
		name      string
		playerID  string
		mode      GameMode
		genre     string
		region    Region
		wantError bool
	}{
		{"valid enqueue", "player1", ModeFFA, "scifi", RegionUSEast, false},
		{"empty player ID", "", ModeFFA, "scifi", RegionUSEast, true},
		{"empty mode", "player2", "", "scifi", RegionUSEast, true},
		{"valid with empty genre", "player3", ModeCoop, "", RegionUSWest, false},
		{"duplicate player", "player1", ModeFFA, "scifi", RegionUSEast, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mm.Enqueue(tt.playerID, tt.mode, tt.genre, tt.region)
			if (err != nil) != tt.wantError {
				t.Errorf("Enqueue() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}

	// Verify queue sizes
	if size := mm.GetQueueSize(ModeFFA); size != 1 {
		t.Errorf("expected FFA queue size 1, got %d", size)
	}
	if size := mm.GetQueueSize(ModeCoop); size != 1 {
		t.Errorf("expected Coop queue size 1, got %d", size)
	}
}

func TestMatchmaker_Dequeue(t *testing.T) {
	hub := NewFederationHub()
	mm := NewMatchmaker(hub)

	// Enqueue players
	mm.Enqueue("player1", ModeFFA, "scifi", RegionUSEast)
	mm.Enqueue("player2", ModeFFA, "scifi", RegionUSEast)
	mm.Enqueue("player3", ModeCoop, "fantasy", RegionEUWest)

	if size := mm.GetQueueSize(ModeFFA); size != 2 {
		t.Fatalf("expected FFA queue size 2, got %d", size)
	}

	// Dequeue player1
	mm.Dequeue("player1")

	if size := mm.GetQueueSize(ModeFFA); size != 1 {
		t.Errorf("expected FFA queue size 1 after dequeue, got %d", size)
	}

	// Verify player2 still in queue
	if !mm.IsQueued("player2") {
		t.Error("player2 should still be in queue")
	}

	// Verify player1 removed
	if mm.IsQueued("player1") {
		t.Error("player1 should be removed from queue")
	}

	// Dequeue non-existent player (should not error)
	mm.Dequeue("nonexistent")
}

func TestMatchmaker_GetQueuedPlayers(t *testing.T) {
	hub := NewFederationHub()
	mm := NewMatchmaker(hub)

	// Enqueue players
	mm.Enqueue("player1", ModeFFA, "scifi", RegionUSEast)
	mm.Enqueue("player2", ModeFFA, "scifi", RegionUSEast)
	mm.Enqueue("player3", ModeCoop, "fantasy", RegionEUWest)

	players := mm.GetQueuedPlayers(ModeFFA)
	if len(players) != 2 {
		t.Errorf("expected 2 players in FFA queue, got %d", len(players))
	}

	// Verify both players are in the list
	found := make(map[string]bool)
	for _, pid := range players {
		found[pid] = true
	}
	if !found["player1"] || !found["player2"] {
		t.Error("expected player1 and player2 in FFA queue")
	}

	coopPlayers := mm.GetQueuedPlayers(ModeCoop)
	if len(coopPlayers) != 1 {
		t.Errorf("expected 1 player in Coop queue, got %d", len(coopPlayers))
	}
	if coopPlayers[0] != "player3" {
		t.Errorf("expected player3 in Coop queue, got %s", coopPlayers[0])
	}
}

func TestMatchmaker_IsQueued(t *testing.T) {
	hub := NewFederationHub()
	mm := NewMatchmaker(hub)

	mm.Enqueue("player1", ModeFFA, "scifi", RegionUSEast)
	mm.Enqueue("player2", ModeCoop, "fantasy", RegionEUWest)

	tests := []struct {
		playerID string
		want     bool
	}{
		{"player1", true},
		{"player2", true},
		{"player3", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.playerID, func(t *testing.T) {
			if got := mm.IsQueued(tt.playerID); got != tt.want {
				t.Errorf("IsQueued(%s) = %v, want %v", tt.playerID, got, tt.want)
			}
		})
	}
}

func TestMatchmaker_GroupPlayers(t *testing.T) {
	hub := NewFederationHub()
	mm := NewMatchmaker(hub)

	queue := []*QueueEntry{
		{PlayerID: "p1", Genre: "scifi", Region: RegionUSEast},
		{PlayerID: "p2", Genre: "scifi", Region: RegionUSEast},
		{PlayerID: "p3", Genre: "fantasy", Region: RegionUSEast},
		{PlayerID: "p4", Genre: "scifi", Region: RegionEUWest},
	}

	groups := mm.groupPlayers(queue)

	// Should have 3 groups: scifi/us-east, fantasy/us-east, scifi/eu-west
	if len(groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(groups))
	}

	// Check scifi/us-east group has 2 players
	key := groupKey{genre: "scifi", region: RegionUSEast}
	if len(groups[key]) != 2 {
		t.Errorf("expected 2 players in scifi/us-east group, got %d", len(groups[key]))
	}
}

func TestMatchmaker_CleanupExpiredEntries(t *testing.T) {
	hub := NewFederationHub()
	mm := NewMatchmaker(hub)
	mm.queueTimeout = 100 * time.Millisecond // Short timeout for testing

	// Enqueue players
	mm.Enqueue("player1", ModeFFA, "scifi", RegionUSEast)
	mm.Enqueue("player2", ModeFFA, "scifi", RegionUSEast)

	if size := mm.GetQueueSize(ModeFFA); size != 2 {
		t.Fatalf("expected initial queue size 2, got %d", size)
	}

	// Wait for entries to expire
	time.Sleep(150 * time.Millisecond)

	// Run cleanup
	mm.cleanupExpiredEntries()

	if size := mm.GetQueueSize(ModeFFA); size != 0 {
		t.Errorf("expected queue size 0 after cleanup, got %d", size)
	}
}

func TestMatchmaker_ProcessMatches_InsufficientPlayers(t *testing.T) {
	hub := NewFederationHub()
	defer hub.Stop()

	// Start hub
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}

	mm := NewMatchmaker(hub)

	// Enqueue only 1 player (need 2 minimum for FFA)
	mm.Enqueue("player1", ModeFFA, "scifi", RegionUSEast)

	// Process matches
	mm.processMatches()

	// Player should still be in queue
	if !mm.IsQueued("player1") {
		t.Error("player1 should still be in queue (insufficient players)")
	}
	if size := mm.GetQueueSize(ModeFFA); size != 1 {
		t.Errorf("expected queue size 1, got %d", size)
	}
}

func TestMatchmaker_ProcessMatches_NoAvailableServer(t *testing.T) {
	hub := NewFederationHub()
	defer hub.Stop()

	// Start hub (no servers registered)
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}

	mm := NewMatchmaker(hub)

	// Enqueue enough players
	mm.Enqueue("player1", ModeFFA, "scifi", RegionUSEast)
	mm.Enqueue("player2", ModeFFA, "scifi", RegionUSEast)

	// Process matches (no servers available)
	mm.processMatches()

	// Players should still be in queue
	if size := mm.GetQueueSize(ModeFFA); size != 2 {
		t.Errorf("expected queue size 2 (no server available), got %d", size)
	}
}

func TestMatchmaker_ProcessMatches_SuccessfulMatch(t *testing.T) {
	hub := NewFederationHub()
	defer hub.Stop()

	// Start hub
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}

	// Register a server
	genre := "scifi"
	region := RegionUSEast
	hub.registerServer(&ServerAnnouncement{
		Name:       "test-server",
		Address:    "127.0.0.1:8080",
		Region:     region,
		Genre:      genre,
		Players:    0,
		MaxPlayers: 8,
	})

	mm := NewMatchmaker(hub)

	// Enqueue players
	mm.Enqueue("player1", ModeFFA, genre, region)
	mm.Enqueue("player2", ModeFFA, genre, region)

	if size := mm.GetQueueSize(ModeFFA); size != 2 {
		t.Fatalf("expected initial queue size 2, got %d", size)
	}

	// Process matches
	mm.processMatches()

	// Players should be matched and removed from queue
	if size := mm.GetQueueSize(ModeFFA); size != 0 {
		t.Errorf("expected queue size 0 after match, got %d", size)
	}
}

func TestMatchmaker_ProcessMatches_MaxPlayersLimit(t *testing.T) {
	hub := NewFederationHub()
	defer hub.Stop()

	// Start hub
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}

	// Register a server
	genre := "scifi"
	region := RegionUSEast
	hub.registerServer(&ServerAnnouncement{
		Name:       "test-server",
		Address:    "127.0.0.1:8080",
		Region:     region,
		Genre:      genre,
		Players:    0,
		MaxPlayers: 16,
	})

	mm := NewMatchmaker(hub)

	// Enqueue more than max players for FFA (max 8)
	for i := 1; i <= 10; i++ {
		mm.Enqueue(fmt.Sprintf("player%d", i), ModeFFA, genre, region)
	}

	if size := mm.GetQueueSize(ModeFFA); size != 10 {
		t.Fatalf("expected initial queue size 10, got %d", size)
	}

	// Process matches
	mm.processMatches()

	// First 8 should be matched, 2 should remain in queue
	if size := mm.GetQueueSize(ModeFFA); size != 2 {
		t.Errorf("expected queue size 2 after match (max 8 per FFA), got %d", size)
	}
}

func TestMatchmaker_ProcessMatches_MultipleGroups(t *testing.T) {
	hub := NewFederationHub()
	defer hub.Stop()

	// Start hub
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}

	// Register servers for different genres/regions
	hub.registerServer(&ServerAnnouncement{
		Name:       "scifi-us",
		Address:    "127.0.0.1:8080",
		Region:     RegionUSEast,
		Genre:      "scifi",
		Players:    0,
		MaxPlayers: 8,
	})
	hub.registerServer(&ServerAnnouncement{
		Name:       "fantasy-eu",
		Address:    "127.0.0.1:8081",
		Region:     RegionEUWest,
		Genre:      "fantasy",
		Players:    0,
		MaxPlayers: 8,
	})

	mm := NewMatchmaker(hub)

	// Enqueue players for different groups
	mm.Enqueue("p1", ModeFFA, "scifi", RegionUSEast)
	mm.Enqueue("p2", ModeFFA, "scifi", RegionUSEast)
	mm.Enqueue("p3", ModeFFA, "fantasy", RegionEUWest)
	mm.Enqueue("p4", ModeFFA, "fantasy", RegionEUWest)

	if size := mm.GetQueueSize(ModeFFA); size != 4 {
		t.Fatalf("expected initial queue size 4, got %d", size)
	}

	// Process matches
	mm.processMatches()

	// Both groups should match and queue should be empty
	if size := mm.GetQueueSize(ModeFFA); size != 0 {
		t.Errorf("expected queue size 0 after matches, got %d", size)
	}
}

func TestMatchmaker_ProcessMatches_InsufficientServerCapacity(t *testing.T) {
	hub := NewFederationHub()
	defer hub.Stop()

	// Start hub
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}

	// Register a server with limited capacity
	genre := "scifi"
	region := RegionUSEast
	hub.registerServer(&ServerAnnouncement{
		Name:       "test-server",
		Address:    "127.0.0.1:8080",
		Region:     region,
		Genre:      genre,
		Players:    7, // Only 1 slot available
		MaxPlayers: 8,
	})

	mm := NewMatchmaker(hub)

	// Enqueue 2 players (need at least 2 for FFA)
	mm.Enqueue("player1", ModeFFA, genre, region)
	mm.Enqueue("player2", ModeFFA, genre, region)

	// Process matches (server doesn't have enough capacity for 2 players)
	mm.processMatches()

	// Players should still be in queue
	if size := mm.GetQueueSize(ModeFFA); size != 2 {
		t.Errorf("expected queue size 2 (insufficient server capacity), got %d", size)
	}
}

func TestMatchmaker_StartStop(t *testing.T) {
	hub := NewFederationHub()
	mm := NewMatchmaker(hub)

	// Start matchmaker
	mm.Start()

	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)

	// Stop matchmaker
	mm.Stop()

	// Verify context is canceled
	select {
	case <-mm.ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("matchmaker context not canceled after Stop()")
	}
}

func TestMatchmaker_Integration(t *testing.T) {
	hub := NewFederationHub()
	defer hub.Stop()

	// Start hub
	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}

	// Register servers
	hub.registerServer(&ServerAnnouncement{
		Name:       "coop-server",
		Address:    "127.0.0.1:8080",
		Region:     RegionUSEast,
		Genre:      "scifi",
		Players:    0,
		MaxPlayers: 4,
	})
	hub.registerServer(&ServerAnnouncement{
		Name:       "ffa-server",
		Address:    "127.0.0.1:8081",
		Region:     RegionUSEast,
		Genre:      "scifi",
		Players:    0,
		MaxPlayers: 8,
	})

	mm := NewMatchmaker(hub)
	mm.matchInterval = 50 * time.Millisecond // Fast interval for testing
	mm.Start()
	defer mm.Stop()

	// Enqueue players for co-op
	mm.Enqueue("coop1", ModeCoop, "scifi", RegionUSEast)
	mm.Enqueue("coop2", ModeCoop, "scifi", RegionUSEast)

	// Enqueue players for FFA
	mm.Enqueue("ffa1", ModeFFA, "scifi", RegionUSEast)
	mm.Enqueue("ffa2", ModeFFA, "scifi", RegionUSEast)
	mm.Enqueue("ffa3", ModeFFA, "scifi", RegionUSEast)

	// Wait for matchmaker to process
	time.Sleep(150 * time.Millisecond)

	// Verify queues are empty (all matched)
	if size := mm.GetQueueSize(ModeCoop); size != 0 {
		t.Errorf("expected Coop queue size 0, got %d", size)
	}
	if size := mm.GetQueueSize(ModeFFA); size != 0 {
		t.Errorf("expected FFA queue size 0, got %d", size)
	}
}

func TestMatchmaker_QueueTimeout_Integration(t *testing.T) {
	hub := NewFederationHub()
	defer hub.Stop()

	if err := hub.Start("127.0.0.1:0"); err != nil {
		t.Fatalf("failed to start hub: %v", err)
	}

	mm := NewMatchmaker(hub)
	mm.queueTimeout = 100 * time.Millisecond
	mm.matchInterval = 50 * time.Millisecond
	mm.cleanupExpiredEntries() // Initial cleanup
	mm.Start()
	defer mm.Stop()

	// Enqueue single player (won't match)
	mm.Enqueue("lonely", ModeFFA, "scifi", RegionUSEast)

	if !mm.IsQueued("lonely") {
		t.Fatal("player should be queued initially")
	}

	// Wait for timeout + cleanup
	time.Sleep(200 * time.Millisecond)

	// Player should be removed due to timeout
	if mm.IsQueued("lonely") {
		t.Error("player should be removed after timeout")
	}
}
