// Package leaderboard provides local and federated score tracking.
package leaderboard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// FederatedConfig holds configuration for federated leaderboard sync.
type FederatedConfig struct {
	HubURL      string        // Federation hub URL
	ServerID    string        // Unique server identifier
	SyncPeriod  time.Duration // How often to sync (0 = disabled)
	OptIn       bool          // Whether to participate in federated leaderboards
	HTTPTimeout time.Duration // HTTP request timeout
}

// FederatedLeaderboard extends Leaderboard with federation support.
type FederatedLeaderboard struct {
	*Leaderboard
	config FederatedConfig
	client *http.Client
}

// NewFederated creates a federated leaderboard instance.
func NewFederated(dbPath string, config FederatedConfig) (*FederatedLeaderboard, error) {
	lb, err := New(dbPath)
	if err != nil {
		return nil, err
	}

	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = 10 * time.Second
	}

	return &FederatedLeaderboard{
		Leaderboard: lb,
		config:      config,
		client: &http.Client{
			Timeout: config.HTTPTimeout,
		},
	}, nil
}

// SyncToHub uploads local leaderboard data to the federation hub.
// Only syncs all_time period for global rankings.
func (fl *FederatedLeaderboard) SyncToHub(stat string) error {
	if !fl.config.OptIn {
		return fmt.Errorf("federated sync disabled (opt-in required)")
	}

	if fl.config.HubURL == "" {
		return fmt.Errorf("hub URL not configured")
	}

	// Get top 100 for this stat
	entries, err := fl.GetTop(stat, "all_time", 100)
	if err != nil {
		return fmt.Errorf("failed to get local scores: %w", err)
	}

	// Prepare sync payload
	payload := struct {
		ServerID string             `json:"server_id"`
		Stat     string             `json:"stat"`
		Entries  []LeaderboardEntry `json:"entries"`
		SyncTime time.Time          `json:"sync_time"`
	}{
		ServerID: fl.config.ServerID,
		Stat:     stat,
		Entries:  entries,
		SyncTime: time.Now(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Send to hub
	url := fmt.Sprintf("%s/api/leaderboard/sync", fl.config.HubURL)
	resp, err := fl.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to sync to hub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("hub sync failed: status %d", resp.StatusCode)
	}

	logrus.WithFields(logrus.Fields{
		"server_id":   fl.config.ServerID,
		"stat":        stat,
		"entry_count": len(entries),
		"hub_url":     fl.config.HubURL,
	}).Debug("leaderboard synced to hub")

	return nil
}

// FetchGlobalTop retrieves global leaderboard rankings from the federation hub.
func (fl *FederatedLeaderboard) FetchGlobalTop(stat string, limit int) ([]LeaderboardEntry, error) {
	if fl.config.HubURL == "" {
		return nil, fmt.Errorf("hub URL not configured")
	}

	url := fmt.Sprintf("%s/api/leaderboard/global?stat=%s&limit=%d", fl.config.HubURL, stat, limit)
	resp, err := fl.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch global leaderboard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hub fetch failed: status %d", resp.StatusCode)
	}

	var entries []LeaderboardEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return entries, nil
}

// GetGlobalRank fetches a player's global rank from the federation hub.
func (fl *FederatedLeaderboard) GetGlobalRank(playerID, stat string) (int, error) {
	if fl.config.HubURL == "" {
		return 0, fmt.Errorf("hub URL not configured")
	}

	url := fmt.Sprintf("%s/api/leaderboard/rank?player_id=%s&stat=%s", fl.config.HubURL, playerID, stat)
	resp, err := fl.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch global rank: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("hub rank fetch failed: status %d", resp.StatusCode)
	}

	var result struct {
		Rank int `json:"rank"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode rank response: %w", err)
	}

	return result.Rank, nil
}
