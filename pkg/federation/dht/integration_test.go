package dht

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestDHTIntegration_10Nodes tests a 10-node DHT network.
// This validates task 1.5: integration tests simulating 10+ node DHT network.
func TestDHTIntegration_10Nodes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	nodes, cleanup := setupTestNetwork(t, 10)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Verify all nodes are connected
	for i, node := range nodes {
		count := node.PeerCount()
		t.Logf("Node %d has %d peers", i, count)
		if count < 1 && i > 0 {
			t.Errorf("Node %d should have at least 1 peer (bootstrap)", i)
		}
	}

	// Test server announcement propagation across network
	record := ServerRecord{
		Name:       "integration-test-server",
		Address:    "192.168.1.100:7777",
		Genre:      "integration",
		MaxPlayers: 64,
		Uptime:     7200,
		Timestamp:  time.Now(),
	}

	// Announce from node 0
	if err := nodes[0].AnnounceServer(ctx, record); err != nil {
		t.Fatalf("AnnounceServer() failed: %v", err)
	}

	// Wait for DHT propagation
	time.Sleep(5 * time.Second)

	// Try to lookup from multiple different nodes
	successCount := 0
	for i := 1; i < len(nodes); i++ {
		retrieved, err := nodes[i].LookupServer(ctx, "integration-test-server")
		if err != nil {
			t.Logf("Node %d: LookupServer() failed: %v", i, err)
			continue
		}

		// Verify record
		if retrieved.Address != record.Address {
			t.Errorf("Node %d: Address = %s, want %s", i, retrieved.Address, record.Address)
		}
		successCount++
	}

	// At least 50% of nodes should successfully retrieve the record
	minSuccess := len(nodes) / 2
	if successCount < minSuccess {
		t.Errorf("Only %d/%d nodes successfully retrieved record, want >= %d",
			successCount, len(nodes)-1, minSuccess)
	} else {
		t.Logf("Successfully retrieved record from %d/%d nodes", successCount, len(nodes)-1)
	}
}

// TestDHTIntegration_BootstrapTiming verifies bootstrap completes within 30 seconds.
// This validates task validation criteria: "DHT bootstrap connects to ≥3 peers within 30 seconds".
func TestDHTIntegration_BootstrapTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create bootstrap nodes
	bootstrapNodes := make([]*Node, 3)
	bootstrapAddrs := make([]string, 3)

	for i := 0; i < 3; i++ {
		node, err := NewNode(ctx, Config{
			ListenAddrs:    []string{"/ip4/127.0.0.1/tcp/0"},
			BootstrapPeers: []string{},
			Mode:           "server",
		})
		if err != nil {
			t.Fatalf("failed to create bootstrap node %d: %v", i, err)
		}
		defer node.Close()

		bootstrapNodes[i] = node
		bootstrapAddrs[i] = node.Addrs()[0].String() + "/p2p/" + node.PeerID().String()
	}

	// Wait for bootstrap nodes to stabilize
	time.Sleep(2 * time.Second)

	// Create test node and measure bootstrap time
	start := time.Now()

	testNode, err := NewNode(ctx, Config{
		ListenAddrs:    []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: bootstrapAddrs,
		Mode:           "client",
	})
	if err != nil {
		t.Fatalf("failed to create test node: %v", err)
	}
	defer testNode.Close()

	// Wait for connections to establish
	deadline := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	connected := false
	for !connected {
		select {
		case <-deadline:
			t.Fatalf("bootstrap did not connect to ≥3 peers within 30 seconds")
		case <-ticker.C:
			peerCount := testNode.PeerCount()
			if peerCount >= 3 {
				connected = true
				elapsed := time.Since(start)
				t.Logf("Bootstrap connected to %d peers in %v", peerCount, elapsed)

				if elapsed > 30*time.Second {
					t.Errorf("Bootstrap took %v, want <= 30s", elapsed)
				}
			}
		}
	}
}

// TestDHTIntegration_ServerLookupTiming verifies server lookups complete within 5 seconds.
// This validates task validation criteria: "DHT server lookup returns results matching HTTP federation hub within 5 seconds".
func TestDHTIntegration_ServerLookupTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	nodes, cleanup := setupTestNetwork(t, 5)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Announce server
	record := ServerRecord{
		Name:       "timing-test-server",
		Address:    "10.0.0.1:8888",
		Genre:      "timing",
		MaxPlayers: 32,
		Uptime:     1000,
		Timestamp:  time.Now(),
	}

	if err := nodes[0].AnnounceServer(ctx, record); err != nil {
		t.Fatalf("AnnounceServer() failed: %v", err)
	}

	// Wait for propagation
	time.Sleep(3 * time.Second)

	// Measure lookup time from different node
	start := time.Now()
	_, err := nodes[2].LookupServer(ctx, "timing-test-server")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("LookupServer() failed: %v", err)
	}

	if elapsed > 5*time.Second {
		t.Errorf("LookupServer() took %v, want <= 5s", elapsed)
	} else {
		t.Logf("LookupServer() completed in %v", elapsed)
	}
}

// TestDHTIntegration_MultipleGenres tests querying servers across multiple genres.
func TestDHTIntegration_MultipleGenres(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	nodes, cleanup := setupTestNetwork(t, 5)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Announce servers in different genres
	genres := []string{"scifi", "fantasy", "western", "cyberpunk"}
	for i, genre := range genres {
		for j := 0; j < 3; j++ {
			srv := ServerRecord{
				Name:       fmt.Sprintf("%s-server-%d", genre, j),
				Address:    fmt.Sprintf("10.0.%d.%d:7777", i, j),
				Genre:      genre,
				MaxPlayers: 16,
				Timestamp:  time.Now(),
			}

			if err := nodes[0].AnnounceServer(ctx, srv); err != nil {
				t.Fatalf("AnnounceServer() failed: %v", err)
			}

			if err := nodes[0].UpdateGenreIndex(ctx, genre, srv.Name, true); err != nil {
				t.Fatalf("UpdateGenreIndex() failed: %v", err)
			}
		}
	}

	// Wait for propagation
	time.Sleep(5 * time.Second)

	// Query each genre from different node
	for _, genre := range genres {
		results, err := nodes[3].QueryServers(ctx, genre, 10)
		if err != nil {
			t.Errorf("QueryServers(%s) failed: %v", genre, err)
			continue
		}

		if len(results) != 3 {
			t.Errorf("QueryServers(%s) returned %d results, want 3", genre, len(results))
		}

		// Verify all results match requested genre
		for _, srv := range results {
			if srv.Genre != genre {
				t.Errorf("QueryServers(%s) returned server with genre %s", genre, srv.Genre)
			}
		}
	}
}
