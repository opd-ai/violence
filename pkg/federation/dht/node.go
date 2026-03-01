// Package dht provides decentralized server discovery using LibP2P Kademlia DHT.
package dht

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

const (
	// RecordTTL is the time-to-live for DHT records (8 hours)
	RecordTTL = 8 * time.Hour
	// MinReplicas is the minimum number of replicas per DHT record
	MinReplicas = 3
	// BootstrapTimeout is the max time to wait for bootstrap
	BootstrapTimeout = 30 * time.Second
)

// ServerRecord represents a server announcement stored in the DHT.
type ServerRecord struct {
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	Genre      string    `json:"genre"`
	MaxPlayers int       `json:"maxPlayers"`
	Uptime     int64     `json:"uptime"` // seconds
	Timestamp  time.Time `json:"timestamp"`
}

// Node represents a DHT node for peer-to-peer server discovery.
type Node struct {
	host      host.Host
	dht       *dht.IpfsDHT
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	bootstrap []string // bootstrap node multiaddrs
}

// Config holds configuration for DHT node creation.
type Config struct {
	// ListenAddrs are the multiaddrs to listen on (e.g., "/ip4/0.0.0.0/tcp/0")
	ListenAddrs []string
	// BootstrapPeers are known DHT bootstrap nodes
	BootstrapPeers []string
	// Mode is either "server" or "client" (affects DHT mode)
	Mode string
}

// NewNode creates a new DHT node with the given configuration.
func NewNode(ctx context.Context, cfg Config) (*Node, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	nodeCtx, cancel := context.WithCancel(ctx)

	// Parse listen addresses
	var listenAddrs []multiaddr.Multiaddr
	for _, addr := range cfg.ListenAddrs {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("invalid listen address %s: %w", addr, err)
		}
		listenAddrs = append(listenAddrs, maddr)
	}

	// Create libp2p host
	h, err := libp2p.New(
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.DefaultSecurity,
		libp2p.DefaultTransports,
		libp2p.NATPortMap(),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// Determine DHT mode
	var dhtMode dht.ModeOpt
	if cfg.Mode == "server" {
		dhtMode = dht.ModeServer
	} else {
		dhtMode = dht.ModeClient
	}

	// Create DHT with namespaced validator and custom protocol prefix
	kdht, err := dht.New(nodeCtx, h,
		dht.Mode(dhtMode),
		dht.ProtocolPrefix("/violence"),
		dht.NamespacedValidator(ViolenceNamespace, ViolenceValidator{}),
	)
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	// Bootstrap the DHT
	if err := kdht.Bootstrap(nodeCtx); err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	node := &Node{
		host:      h,
		dht:       kdht,
		ctx:       nodeCtx,
		cancel:    cancel,
		bootstrap: cfg.BootstrapPeers,
	}

	// Connect to bootstrap peers
	if len(cfg.BootstrapPeers) > 0 {
		go node.connectBootstrap()
	}

	logrus.WithFields(logrus.Fields{
		"peer_id":   h.ID().String(),
		"addrs":     h.Addrs(),
		"mode":      cfg.Mode,
		"bootstrap": len(cfg.BootstrapPeers),
	}).Info("DHT node started")

	return node, nil
}

// connectBootstrap connects to bootstrap peers.
func (n *Node) connectBootstrap() {
	ctx, cancel := context.WithTimeout(n.ctx, BootstrapTimeout)
	defer cancel()

	var wg sync.WaitGroup
	connected := 0
	var connMu sync.Mutex

	for _, addrStr := range n.bootstrap {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()

			maddr, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				logrus.WithError(err).WithField("addr", addr).Warn("invalid bootstrap address")
				return
			}

			peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				logrus.WithError(err).WithField("addr", addr).Warn("failed to parse peer info")
				return
			}

			if err := n.host.Connect(ctx, *peerInfo); err != nil {
				logrus.WithError(err).WithField("peer", peerInfo.ID).Debug("failed to connect to bootstrap peer")
				return
			}

			connMu.Lock()
			connected++
			connMu.Unlock()

			logrus.WithField("peer", peerInfo.ID).Info("connected to bootstrap peer")
		}(addrStr)
	}

	wg.Wait()

	logrus.WithFields(logrus.Fields{
		"connected": connected,
		"total":     len(n.bootstrap),
	}).Info("bootstrap connection attempt complete")
}

// Close shuts down the DHT node.
func (n *Node) Close() error {
	n.cancel()
	if err := n.dht.Close(); err != nil {
		logrus.WithError(err).Warn("error closing DHT")
	}
	return n.host.Close()
}

// PeerID returns the node's peer ID.
func (n *Node) PeerID() peer.ID {
	return n.host.ID()
}

// Addrs returns the node's listen addresses.
func (n *Node) Addrs() []multiaddr.Multiaddr {
	return n.host.Addrs()
}

// Routing returns the DHT routing interface.
func (n *Node) Routing() routing.Routing {
	return n.dht
}

// PeerCount returns the number of connected peers.
func (n *Node) PeerCount() int {
	return len(n.host.Network().Peers())
}
