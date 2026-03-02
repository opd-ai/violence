package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opd-ai/violence/pkg/mod"
	"github.com/opd-ai/violence/pkg/mod/registry"
)

func TestNewModBrowser(t *testing.T) {
	mb := NewModBrowser("http://localhost:8080")
	if mb == nil {
		t.Fatal("expected non-nil mod browser")
	}
	if mb.registryURL != "http://localhost:8080" {
		t.Errorf("expected registryURL %s, got %s", "http://localhost:8080", mb.registryURL)
	}
	if mb.visible {
		t.Error("expected browser to be hidden initially")
	}
	if mb.state != ModBrowserStateBrowse {
		t.Errorf("expected initial state %d, got %d", ModBrowserStateBrowse, mb.state)
	}
}

func TestModBrowserVisibility(t *testing.T) {
	mb := NewModBrowser("http://test")

	if mb.IsVisible() {
		t.Error("expected browser to be hidden initially")
	}

	mb.SetVisible(true)
	if !mb.IsVisible() {
		t.Error("expected browser to be visible after SetVisible(true)")
	}

	mb.SetVisible(false)
	if mb.IsVisible() {
		t.Error("expected browser to be hidden after SetVisible(false)")
	}
}

func TestRefreshModList(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"results": []registry.ModRecord{
				{
					Name:        "test-mod",
					Version:     "1.0.0",
					Author:      "testauthor",
					Description: "Test mod",
					Size:        1024,
					Downloads:   42,
				},
			},
			"count": 1,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mb := NewModBrowser(server.URL)
	if err := mb.RefreshModList(); err != nil {
		t.Fatalf("RefreshModList failed: %v", err)
	}

	mb.mu.RLock()
	modsCount := len(mb.mods)
	mb.mu.RUnlock()

	if modsCount != 1 {
		t.Errorf("expected 1 mod, got %d", modsCount)
	}

	mb.mu.RLock()
	firstMod := mb.mods[0]
	mb.mu.RUnlock()

	if firstMod.Name != "test-mod" {
		t.Errorf("expected mod name test-mod, got %s", firstMod.Name)
	}
	if firstMod.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", firstMod.Version)
	}
}

func TestRefreshModListServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	mb := NewModBrowser(server.URL)
	err := mb.RefreshModList()
	if err == nil {
		t.Error("expected error on server error, got nil")
	}
}

func TestRefreshModListInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	mb := NewModBrowser(server.URL)
	err := mb.RefreshModList()
	if err == nil {
		t.Error("expected error on invalid JSON, got nil")
	}
}

func TestCheckForUpdates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract name from query
		name := r.URL.Query().Get("name")

		version := "1.0.0"
		if name == "outdated-mod" {
			version = "2.0.0" // Newer version available
		}

		response := map[string]interface{}{
			"results": []registry.ModRecord{
				{
					Name:    name,
					Version: version,
					Author:  "testauthor",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mb := NewModBrowser(server.URL)

	// Mark mods as installed
	mb.SetInstalledMods(map[string]string{
		"outdated-mod": "1.0.0",
		"current-mod":  "1.0.0",
	})

	if err := mb.CheckForUpdates(); err != nil {
		t.Fatalf("CheckForUpdates failed: %v", err)
	}

	updateCount := mb.GetUpdateCount()
	if updateCount != 1 {
		t.Errorf("expected 1 update available, got %d", updateCount)
	}

	mb.mu.RLock()
	newVersion, hasUpdate := mb.updateAvailable["outdated-mod"]
	mb.mu.RUnlock()

	if !hasUpdate {
		t.Error("expected update for outdated-mod")
	}
	if newVersion != "2.0.0" {
		t.Errorf("expected new version 2.0.0, got %s", newVersion)
	}
}

func TestDownloadMod(t *testing.T) {
	wasmData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00} // WASM magic
	expectedChecksum := mod.ComputeSHA256(wasmData)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/download/test-mod/1.0.0" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("X-SHA256", expectedChecksum)
		w.Write(wasmData)
	}))
	defer server.Close()

	mb := NewModBrowser(server.URL)
	data, checksum, err := mb.DownloadMod("test-mod", "1.0.0")
	if err != nil {
		t.Fatalf("DownloadMod failed: %v", err)
	}

	if len(data) != len(wasmData) {
		t.Errorf("expected %d bytes, got %d", len(wasmData), len(data))
	}
	if checksum != expectedChecksum {
		t.Errorf("expected checksum %s, got %s", expectedChecksum, checksum)
	}
}

func TestDownloadModNoChecksum(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No X-SHA256 header
		w.Write([]byte("data"))
	}))
	defer server.Close()

	mb := NewModBrowser(server.URL)
	_, _, err := mb.DownloadMod("test-mod", "1.0.0")
	if err == nil {
		t.Error("expected error when checksum header is missing")
	}
}

func TestDownloadMod404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	mb := NewModBrowser(server.URL)
	_, _, err := mb.DownloadMod("nonexistent", "1.0.0")
	if err == nil {
		t.Error("expected error on 404 response")
	}
}

func TestInstallModChecksumMismatch(t *testing.T) {
	wasmData := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	wrongChecksum := "0000000000000000000000000000000000000000000000000000000000000000"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-SHA256", wrongChecksum)
		w.Write(wasmData)
	}))
	defer server.Close()

	mb := NewModBrowser(server.URL)
	err := mb.InstallMod("test-mod", "1.0.0")
	if err == nil {
		t.Error("expected error on checksum mismatch")
	}
}

func TestSetInstalledMods(t *testing.T) {
	mb := NewModBrowser("http://test")
	installed := map[string]string{
		"mod1": "1.0.0",
		"mod2": "2.0.0",
	}

	mb.SetInstalledMods(installed)

	mb.mu.RLock()
	count := len(mb.installedMods)
	mb.mu.RUnlock()

	if count != 2 {
		t.Errorf("expected 2 installed mods, got %d", count)
	}
}

func TestGetUpdateCount(t *testing.T) {
	mb := NewModBrowser("http://test")

	if mb.GetUpdateCount() != 0 {
		t.Error("expected 0 updates initially")
	}

	mb.mu.Lock()
	mb.updateAvailable["mod1"] = "2.0.0"
	mb.updateAvailable["mod2"] = "3.0.0"
	mb.mu.Unlock()

	if mb.GetUpdateCount() != 2 {
		t.Errorf("expected 2 updates, got %d", mb.GetUpdateCount())
	}
}

func TestModBrowserNavigation(t *testing.T) {
	mb := NewModBrowser("http://test")
	mb.visible = true

	// Add test mods
	mb.mods = []registry.ModRecord{
		{Name: "mod1", Version: "1.0.0"},
		{Name: "mod2", Version: "1.0.0"},
		{Name: "mod3", Version: "1.0.0"},
	}

	// Test down navigation
	mb.selectedIndex = 0
	mb.NavigateDown()
	if mb.selectedIndex != 1 {
		t.Errorf("expected selectedIndex 1, got %d", mb.selectedIndex)
	}

	// Test up navigation
	mb.NavigateUp()
	if mb.selectedIndex != 0 {
		t.Errorf("expected selectedIndex 0, got %d", mb.selectedIndex)
	}

	// Test bounds checking
	mb.selectedIndex = 2
	mb.NavigateDown() // Should not go beyond bounds
	if mb.selectedIndex != 2 {
		t.Errorf("expected selectedIndex 2 (at bounds), got %d", mb.selectedIndex)
	}
}

func TestModBrowserStateTransitions(t *testing.T) {
	mb := NewModBrowser("http://test")

	if mb.state != ModBrowserStateBrowse {
		t.Errorf("expected initial state Browse, got %d", mb.state)
	}

	mb.state = ModBrowserStateDetails
	if mb.state != ModBrowserStateDetails {
		t.Errorf("expected state Details, got %d", mb.state)
	}

	mb.state = ModBrowserStateInstalling
	if mb.state != ModBrowserStateInstalling {
		t.Errorf("expected state Installing, got %d", mb.state)
	}
}

func TestModBrowserConcurrentAccess(t *testing.T) {
	mb := NewModBrowser("http://test")

	// Concurrent reads and writes
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			mb.SetInstalledMods(map[string]string{"mod": "1.0.0"})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_ = mb.GetUpdateCount()
		}
		done <- true
	}()

	<-done
	<-done
}

func TestAutoUpdateCheck(t *testing.T) {
	mb := NewModBrowser("http://test")

	// Set auto-update check time to past
	mb.autoUpdateCheck = time.Now().Add(-31 * time.Minute)

	timeSince := time.Since(mb.autoUpdateCheck)
	if timeSince <= 30*time.Minute {
		t.Error("expected auto-update check to be overdue")
	}
}

func TestErrorMessageTimeout(t *testing.T) {
	mb := NewModBrowser("http://test")

	// Set error in the past
	mb.errorMessage = "test error"
	mb.errorTime = time.Now().Add(-6 * time.Second)

	// Error should be expired
	if time.Since(mb.errorTime) <= 5*time.Second {
		t.Error("error should have expired")
	}
}

func TestModBrowserSearchQuery(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// Check if query parameter is present
		query := r.URL.Query().Get("name")
		if requestCount == 2 && query != "search-term" {
			t.Errorf("expected search query 'search-term', got '%s'", query)
		}

		response := map[string]interface{}{
			"results": []registry.ModRecord{},
			"count":   0,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mb := NewModBrowser(server.URL)

	// First request without search
	mb.RefreshModList()

	// Second request with search
	mb.searchQuery = "search-term"
	mb.RefreshModList()

	if requestCount != 2 {
		t.Errorf("expected 2 requests, got %d", requestCount)
	}
}

func TestModBrowserInstallProgress(t *testing.T) {
	mb := NewModBrowser("http://test")

	mb.mu.Lock()
	mb.installProgress = "Downloading..."
	progress := mb.installProgress
	mb.mu.Unlock()

	if progress != "Downloading..." {
		t.Errorf("expected progress 'Downloading...', got '%s'", progress)
	}
}

func TestComputeSHA256(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "empty",
			data:     []byte{},
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "simple",
			data:     []byte("hello"),
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name:     "wasm magic",
			data:     []byte{0x00, 0x61, 0x73, 0x6d},
			expected: "4c910acba0aa1dd89b8cdd3ceb30165c84f2870c2e35e8ac82a7bda31a7ec69c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mod.ComputeSHA256(tt.data)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestLoadWASMModuleError(t *testing.T) {
	// Invalid WASM data should fail
	invalidData := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	err := mod.LoadWASMModule(invalidData)
	if err == nil {
		t.Error("expected error when loading invalid WASM data")
	}
}

func TestModBrowserHTTPClientTimeout(t *testing.T) {
	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(35 * time.Second) // Longer than client timeout
	}))
	defer server.Close()

	mb := NewModBrowser(server.URL)
	err := mb.RefreshModList()
	if err == nil {
		t.Error("expected timeout error")
	}
}

func BenchmarkRefreshModList(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mods := make([]registry.ModRecord, 50)
		for i := range mods {
			mods[i] = registry.ModRecord{
				Name:    fmt.Sprintf("mod-%d", i),
				Version: "1.0.0",
				Author:  "author",
			}
		}

		response := map[string]interface{}{
			"results": mods,
			"count":   len(mods),
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mb := NewModBrowser(server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mb.RefreshModList()
	}
}

func BenchmarkCheckForUpdates(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"results": []registry.ModRecord{
				{Name: "test-mod", Version: "2.0.0"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mb := NewModBrowser(server.URL)
	mb.SetInstalledMods(map[string]string{"test-mod": "1.0.0"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mb.CheckForUpdates()
	}
}
