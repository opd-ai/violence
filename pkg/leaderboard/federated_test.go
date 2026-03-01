package leaderboard

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestNewFederated(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := FederatedConfig{
		HubURL:     "http://hub.example.com",
		ServerID:   "server1",
		SyncPeriod: 5 * time.Minute,
		OptIn:      true,
	}

	flb, err := NewFederated(dbPath, config)
	if err != nil {
		t.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	if flb.config.HubURL != config.HubURL {
		t.Errorf("HubURL = %v, want %v", flb.config.HubURL, config.HubURL)
	}

	if flb.config.HTTPTimeout == 0 {
		t.Error("HTTPTimeout should have default value")
	}
}

func TestSyncToHub(t *testing.T) {
	tmpDir := t.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "test.db"), FederatedConfig{
		ServerID: "server1",
		OptIn:    true,
	})
	if err != nil {
		t.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	// Record some test scores
	flb.RecordScore("p1", "Alice", "kills", "all_time", 1000)
	flb.RecordScore("p2", "Bob", "kills", "all_time", 500)

	// Create mock hub server
	syncReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/leaderboard/sync" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var payload struct {
			ServerID string             `json:"server_id"`
			Stat     string             `json:"stat"`
			Entries  []LeaderboardEntry `json:"entries"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("failed to decode payload: %v", err)
		}

		if payload.ServerID != "server1" {
			t.Errorf("ServerID = %v, want server1", payload.ServerID)
		}

		if payload.Stat != "kills" {
			t.Errorf("Stat = %v, want kills", payload.Stat)
		}

		if len(payload.Entries) != 2 {
			t.Errorf("len(Entries) = %v, want 2", len(payload.Entries))
		}

		syncReceived = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	flb.config.HubURL = server.URL

	// Sync to hub
	if err := flb.SyncToHub("kills"); err != nil {
		t.Errorf("SyncToHub() error = %v", err)
	}

	if !syncReceived {
		t.Error("Hub did not receive sync request")
	}
}

func TestSyncToHubOptOut(t *testing.T) {
	tmpDir := t.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "test.db"), FederatedConfig{
		ServerID: "server1",
		OptIn:    false, // Opt-out
	})
	if err != nil {
		t.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	// Sync should fail when opted out
	err = flb.SyncToHub("kills")
	if err == nil {
		t.Error("SyncToHub() should error when opted out")
	}
}

func TestSyncToHubNoURL(t *testing.T) {
	tmpDir := t.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "test.db"), FederatedConfig{
		ServerID: "server1",
		OptIn:    true,
		HubURL:   "", // No hub URL
	})
	if err != nil {
		t.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	err = flb.SyncToHub("kills")
	if err == nil {
		t.Error("SyncToHub() should error when hub URL not configured")
	}
}

func TestSyncToHubServerError(t *testing.T) {
	tmpDir := t.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "test.db"), FederatedConfig{
		ServerID: "server1",
		OptIn:    true,
	})
	if err != nil {
		t.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	// Mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	flb.config.HubURL = server.URL

	err = flb.SyncToHub("kills")
	if err == nil {
		t.Error("SyncToHub() should error when hub returns 500")
	}
}

func TestFetchGlobalTop(t *testing.T) {
	tmpDir := t.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "test.db"), FederatedConfig{
		ServerID: "server1",
	})
	if err != nil {
		t.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	// Mock hub with global leaderboard data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/leaderboard/global" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		entries := []LeaderboardEntry{
			{PlayerID: "p1", PlayerName: "GlobalChamp", Score: 10000, Stat: "kills", Period: "all_time"},
			{PlayerID: "p2", PlayerName: "Runner-Up", Score: 8000, Stat: "kills", Period: "all_time"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	}))
	defer server.Close()

	flb.config.HubURL = server.URL

	// Fetch global top
	entries, err := flb.FetchGlobalTop("kills", 10)
	if err != nil {
		t.Errorf("FetchGlobalTop() error = %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("len(entries) = %v, want 2", len(entries))
	}

	if entries[0].PlayerName != "GlobalChamp" {
		t.Errorf("entries[0].PlayerName = %v, want GlobalChamp", entries[0].PlayerName)
	}

	if entries[0].Score != 10000 {
		t.Errorf("entries[0].Score = %v, want 10000", entries[0].Score)
	}
}

func TestFetchGlobalTopNoURL(t *testing.T) {
	tmpDir := t.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "test.db"), FederatedConfig{
		ServerID: "server1",
		HubURL:   "", // No hub URL
	})
	if err != nil {
		t.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	_, err = flb.FetchGlobalTop("kills", 10)
	if err == nil {
		t.Error("FetchGlobalTop() should error when hub URL not configured")
	}
}

func TestFetchGlobalTopServerError(t *testing.T) {
	tmpDir := t.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "test.db"), FederatedConfig{
		ServerID: "server1",
	})
	if err != nil {
		t.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	// Mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	flb.config.HubURL = server.URL

	_, err = flb.FetchGlobalTop("kills", 10)
	if err == nil {
		t.Error("FetchGlobalTop() should error when hub returns 404")
	}
}

func TestFetchGlobalTopInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "test.db"), FederatedConfig{
		ServerID: "server1",
	})
	if err != nil {
		t.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	// Mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	flb.config.HubURL = server.URL

	_, err = flb.FetchGlobalTop("kills", 10)
	if err == nil {
		t.Error("FetchGlobalTop() should error on invalid JSON")
	}
}

func TestGetGlobalRank(t *testing.T) {
	tmpDir := t.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "test.db"), FederatedConfig{
		ServerID: "server1",
	})
	if err != nil {
		t.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	// Mock hub with rank data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/leaderboard/rank" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		playerID := r.URL.Query().Get("player_id")
		if playerID != "p1" {
			t.Errorf("player_id = %v, want p1", playerID)
		}

		result := struct {
			Rank int `json:"rank"`
		}{
			Rank: 42,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	flb.config.HubURL = server.URL

	// Get global rank
	rank, err := flb.GetGlobalRank("p1", "kills")
	if err != nil {
		t.Errorf("GetGlobalRank() error = %v", err)
	}

	if rank != 42 {
		t.Errorf("rank = %v, want 42", rank)
	}
}

func TestGetGlobalRankNoURL(t *testing.T) {
	tmpDir := t.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "test.db"), FederatedConfig{
		ServerID: "server1",
		HubURL:   "", // No hub URL
	})
	if err != nil {
		t.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	_, err = flb.GetGlobalRank("p1", "kills")
	if err == nil {
		t.Error("GetGlobalRank() should error when hub URL not configured")
	}
}

func TestGetGlobalRankServerError(t *testing.T) {
	tmpDir := t.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "test.db"), FederatedConfig{
		ServerID: "server1",
	})
	if err != nil {
		t.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	// Mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	flb.config.HubURL = server.URL

	_, err = flb.GetGlobalRank("p1", "kills")
	if err == nil {
		t.Error("GetGlobalRank() should error when hub returns 503")
	}
}

func TestGetGlobalRankInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "test.db"), FederatedConfig{
		ServerID: "server1",
	})
	if err != nil {
		t.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	// Mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{invalid}"))
	}))
	defer server.Close()

	flb.config.HubURL = server.URL

	_, err = flb.GetGlobalRank("p1", "kills")
	if err == nil {
		t.Error("GetGlobalRank() should error on invalid JSON")
	}
}

func TestFederatedHTTPTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "test.db"), FederatedConfig{
		ServerID:    "server1",
		OptIn:       true,
		HTTPTimeout: 100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	// Mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Longer than timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	flb.config.HubURL = server.URL

	// Should timeout
	err = flb.SyncToHub("kills")
	if err == nil {
		t.Error("SyncToHub() should timeout on slow server")
	}
}

// BenchmarkSyncToHub measures federation sync performance.
func BenchmarkSyncToHub(b *testing.B) {
	tmpDir := b.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "bench.db"), FederatedConfig{
		ServerID: "server1",
		OptIn:    true,
	})
	if err != nil {
		b.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	// Record test data
	for i := 0; i < 100; i++ {
		flb.RecordScore("p1", "Alice", "kills", "all_time", int64(i))
	}

	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	flb.config.HubURL = server.URL

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flb.SyncToHub("kills")
	}
}

// BenchmarkFetchGlobalTop measures global leaderboard fetch performance.
func BenchmarkFetchGlobalTop(b *testing.B) {
	tmpDir := b.TempDir()
	flb, err := NewFederated(filepath.Join(tmpDir, "bench.db"), FederatedConfig{
		ServerID: "server1",
	})
	if err != nil {
		b.Fatalf("NewFederated() error = %v", err)
	}
	defer flb.Close()

	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entries := make([]LeaderboardEntry, 100)
		for i := 0; i < 100; i++ {
			entries[i] = LeaderboardEntry{
				PlayerID:   "p" + string(rune(i)),
				PlayerName: "Player",
				Score:      int64(1000 - i),
				Stat:       "kills",
				Period:     "all_time",
			}
		}
		json.NewEncoder(w).Encode(entries)
	}))
	defer server.Close()

	flb.config.HubURL = server.URL

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flb.FetchGlobalTop("kills", 100)
	}
}
