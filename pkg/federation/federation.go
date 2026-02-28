// Package federation provides server federation and matchmaking.
package federation

import (
	"fmt"
	"sync"

	"github.com/opd-ai/violence/pkg/rng"
)

// ServerInfo holds server metadata for matchmaking.
type ServerInfo struct {
	Name       string
	Address    string
	Players    int
	MaxPlayers int
	Genre      string
}

// Federation manages a network of game servers.
type Federation struct {
	servers map[string]*ServerInfo
	mu      sync.RWMutex
	rng     *rng.RNG
}

// NewFederation creates a new federation instance.
func NewFederation() *Federation {
	return &Federation{
		servers: make(map[string]*ServerInfo),
		rng:     rng.NewRNG(0xfed3a710),
	}
}

// Register adds a server to the federation.
func (f *Federation) Register(name, address string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.servers[name] = &ServerInfo{
		Name:       name,
		Address:    address,
		Players:    0,
		MaxPlayers: 16,
	}
}

// RegisterWithInfo adds a server with full metadata.
func (f *Federation) RegisterWithInfo(info *ServerInfo) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.servers[info.Name] = info
}

// Lookup finds a server by name.
func (f *Federation) Lookup(name string) (string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	info, ok := f.servers[name]
	if !ok {
		return "", false
	}
	return info.Address, true
}

// Match finds a suitable server for matchmaking.
// Returns the address of a server with available slots, or error if none found.
func (f *Federation) Match() (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var available []*ServerInfo
	for _, info := range f.servers {
		if info.Players < info.MaxPlayers {
			available = append(available, info)
		}
	}

	if len(available) == 0 {
		return "", fmt.Errorf("no available servers")
	}

	// Random selection among available servers
	selected := available[f.rng.Intn(len(available))]
	return selected.Address, nil
}

// MatchGenre finds a server matching the requested genre.
func (f *Federation) MatchGenre(genreID string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var available []*ServerInfo
	for _, info := range f.servers {
		if info.Players < info.MaxPlayers && info.Genre == genreID {
			available = append(available, info)
		}
	}

	if len(available) == 0 {
		return "", fmt.Errorf("no available servers for genre %s", genreID)
	}

	selected := available[f.rng.Intn(len(available))]
	return selected.Address, nil
}

// List returns all registered servers.
func (f *Federation) List() []*ServerInfo {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make([]*ServerInfo, 0, len(f.servers))
	for _, info := range f.servers {
		result = append(result, info)
	}
	return result
}

// Unregister removes a server from the federation.
func (f *Federation) Unregister(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.servers, name)
}

// SetGenre configures the federation system for a genre.
func SetGenre(genreID string) {}
