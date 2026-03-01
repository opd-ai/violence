package dht

import (
	"context"
	"testing"
	"time"
)

func TestNewNode(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "server mode with valid config",
			cfg: Config{
				ListenAddrs:    []string{"/ip4/127.0.0.1/tcp/0"},
				BootstrapPeers: []string{},
				Mode:           "server",
			},
			wantErr: false,
		},
		{
			name: "client mode with valid config",
			cfg: Config{
				ListenAddrs:    []string{"/ip4/127.0.0.1/tcp/0"},
				BootstrapPeers: []string{},
				Mode:           "client",
			},
			wantErr: false,
		},
		{
			name: "invalid listen address",
			cfg: Config{
				ListenAddrs:    []string{"invalid-address"},
				BootstrapPeers: []string{},
				Mode:           "server",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			node, err := NewNode(ctx, tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				defer node.Close()

				if node.PeerID() == "" {
					t.Error("NewNode() returned node with empty peer ID")
				}

				if len(node.Addrs()) == 0 {
					t.Error("NewNode() returned node with no addresses")
				}
			}
		})
	}
}

func TestNode_PeerCount(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := Config{
		ListenAddrs:    []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		Mode:           "server",
	}

	node, err := NewNode(ctx, cfg)
	if err != nil {
		t.Fatalf("NewNode() failed: %v", err)
	}
	defer node.Close()

	// Initially should have 0 peers
	if count := node.PeerCount(); count != 0 {
		t.Errorf("PeerCount() = %d, want 0", count)
	}
}

func TestNode_CloseIdempotent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := Config{
		ListenAddrs:    []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		Mode:           "server",
	}

	node, err := NewNode(ctx, cfg)
	if err != nil {
		t.Fatalf("NewNode() failed: %v", err)
	}

	// Close multiple times should not panic
	if err := node.Close(); err != nil {
		t.Errorf("first Close() failed: %v", err)
	}

	if err := node.Close(); err != nil {
		t.Errorf("second Close() failed: %v", err)
	}
}

func TestNode_MultipleNodes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Create first node
	node1, err := NewNode(ctx, Config{
		ListenAddrs:    []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		Mode:           "server",
	})
	if err != nil {
		t.Fatalf("failed to create node1: %v", err)
	}
	defer node1.Close()

	// Get node1's address for bootstrapping
	addr1 := node1.Addrs()[0].String() + "/p2p/" + node1.PeerID().String()

	// Create second node that bootstraps from first
	node2, err := NewNode(ctx, Config{
		ListenAddrs:    []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{addr1},
		Mode:           "client",
	})
	if err != nil {
		t.Fatalf("failed to create node2: %v", err)
	}
	defer node2.Close()

	// Wait for bootstrap connection
	time.Sleep(2 * time.Second)

	// Check peer counts
	if count := node1.PeerCount(); count < 1 {
		t.Errorf("node1 PeerCount() = %d, want >= 1", count)
	}

	if count := node2.PeerCount(); count < 1 {
		t.Errorf("node2 PeerCount() = %d, want >= 1", count)
	}
}

func TestNode_BootstrapTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create node with invalid bootstrap peer (will timeout)
	node, err := NewNode(ctx, Config{
		ListenAddrs:    []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{"/ip4/127.0.0.1/tcp/9999/p2p/12D3KooWInvalidPeerID"},
		Mode:           "client",
	})
	if err != nil {
		t.Fatalf("NewNode() failed: %v", err)
	}
	defer node.Close()

	// Should still create node successfully even if bootstrap fails
	if node.PeerID() == "" {
		t.Error("node should have valid peer ID even with failed bootstrap")
	}

	// Wait for bootstrap attempt
	time.Sleep(2 * time.Second)

	// Should have 0 peers due to invalid bootstrap
	if count := node.PeerCount(); count != 0 {
		t.Errorf("PeerCount() = %d, want 0 (bootstrap should have failed)", count)
	}
}
