// Package dht provides decentralized server discovery using LibP2P Kademlia DHT.
package dht

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/sirupsen/logrus"
)

// AnnounceServer publishes a server record to the DHT.
// The record will be replicated across at least MinReplicas peers.
func (n *Node) AnnounceServer(ctx context.Context, record ServerRecord) error {
	if ctx == nil {
		ctx = n.ctx
	}

	// Set timestamp if not already set
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	// Serialize record to JSON
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	// Generate DHT key from server name
	key := makeKey("server", record.Name)

	// Put record in DHT
	if err := n.dht.PutValue(ctx, key, data); err != nil {
		return fmt.Errorf("failed to put DHT value: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"server_name": record.Name,
		"genre":       record.Genre,
		"key":         key,
	}).Info("announced server to DHT")

	return nil
}

// LookupServer retrieves a server record from the DHT by name.
func (n *Node) LookupServer(ctx context.Context, serverName string) (*ServerRecord, error) {
	if ctx == nil {
		ctx = n.ctx
	}

	key := makeKey("server", serverName)

	// Get value from DHT
	data, err := n.dht.GetValue(ctx, key)
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, fmt.Errorf("server not found: %s", serverName)
		}
		return nil, fmt.Errorf("failed to get DHT value: %w", err)
	}

	// Deserialize record
	var record ServerRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %w", err)
	}

	// Check if record is stale
	if time.Since(record.Timestamp) > RecordTTL {
		return nil, fmt.Errorf("server record expired: %s", serverName)
	}

	return &record, nil
}

// QueryServers searches for servers matching the given genre.
// Returns up to maxResults servers.
func (n *Node) QueryServers(ctx context.Context, genre string, maxResults int) ([]*ServerRecord, error) {
	if ctx == nil {
		ctx = n.ctx
	}
	if maxResults <= 0 {
		maxResults = 10
	}

	// For DHT-based queries, we need to search by genre prefix
	// This is a simplified implementation - in production, consider using
	// a secondary index or IPNS for more efficient genre-based lookups
	key := makeKey("genre", genre)

	// Get list of server names for this genre
	data, err := n.dht.GetValue(ctx, key)
	if err != nil {
		if err == datastore.ErrNotFound {
			return []*ServerRecord{}, nil
		}
		return nil, fmt.Errorf("failed to get genre index: %w", err)
	}

	// Deserialize server name list
	var serverNames []string
	if err := json.Unmarshal(data, &serverNames); err != nil {
		return nil, fmt.Errorf("failed to unmarshal genre index: %w", err)
	}

	// Fetch individual server records
	var results []*ServerRecord
	for i, name := range serverNames {
		if i >= maxResults {
			break
		}

		record, err := n.LookupServer(ctx, name)
		if err != nil {
			logrus.WithError(err).WithField("server", name).Debug("failed to fetch server record")
			continue
		}

		results = append(results, record)
	}

	return results, nil
}

// UpdateGenreIndex updates the genre index with a server name.
// This allows efficient lookups by genre.
func (n *Node) UpdateGenreIndex(ctx context.Context, genre, serverName string, add bool) error {
	if ctx == nil {
		ctx = n.ctx
	}

	key := makeKey("genre", genre)

	// Get current index
	var serverNames []string
	data, err := n.dht.GetValue(ctx, key)
	if err != nil && err != datastore.ErrNotFound {
		// Check for routing not found error (DHT-specific)
		if err.Error() != "routing: not found" {
			return fmt.Errorf("failed to get genre index: %w", err)
		}
		// Index doesn't exist yet, start with empty list
	}
	if err == nil {
		if err := json.Unmarshal(data, &serverNames); err != nil {
			return fmt.Errorf("failed to unmarshal genre index: %w", err)
		}
	}

	// Update index
	if add {
		// Add server name if not already present
		found := false
		for _, name := range serverNames {
			if name == serverName {
				found = true
				break
			}
		}
		if !found {
			serverNames = append(serverNames, serverName)
		}
	} else {
		// Remove server name
		filtered := make([]string, 0, len(serverNames))
		for _, name := range serverNames {
			if name != serverName {
				filtered = append(filtered, name)
			}
		}
		serverNames = filtered
	}

	// Serialize and store updated index
	data, err = json.Marshal(serverNames)
	if err != nil {
		return fmt.Errorf("failed to marshal genre index: %w", err)
	}

	if err := n.dht.PutValue(ctx, key, data); err != nil {
		return fmt.Errorf("failed to put genre index: %w", err)
	}

	return nil
}

// makeKey generates a DHT key with namespace prefix.
func makeKey(namespace, value string) string {
	return fmt.Sprintf("/violence/%s/%s", namespace, value)
}
