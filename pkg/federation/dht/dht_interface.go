package dht

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/multiformats/go-multiaddr"
)

// NodeInterface abstracts the DHT node so that callers can be tested without
// creating a live libp2p host.
type NodeInterface interface {
	// Close shuts down the node and releases all resources.
	Close() error
	// PeerID returns the local peer ID.
	PeerID() peer.ID
	// Addrs returns the listen addresses for this node.
	Addrs() []multiaddr.Multiaddr
	// Routing returns the underlying routing.Routing implementation.
	Routing() routing.Routing
	// PeerCount returns the number of currently connected peers.
	PeerCount() int
}

// Bootstrap connects to the default IPFS bootstrap peers so the node can join
// the global DHT network. Callers that only need local connectivity should pass
// an empty bootstrap list in Config instead.
type Bootstrapper interface {
	Bootstrap(ctx context.Context) error
}

// ensure *Node implements NodeInterface at compile time.
var _ NodeInterface = (*Node)(nil)
