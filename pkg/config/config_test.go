package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()

	tests := []struct {
		name     string
		field    string
		expected interface{}
	}{
		{"WindowWidth", "WindowWidth", 960},
		{"WindowHeight", "WindowHeight", 600},
		{"InternalWidth", "InternalWidth", 320},
		{"InternalHeight", "InternalHeight", 200},
		{"FOV", "FOV", 66.0},
		{"MouseSensitivity", "MouseSensitivity", 1.0},
		{"MasterVolume", "MasterVolume", 0.8},
		{"MusicVolume", "MusicVolume", 0.7},
		{"SFXVolume", "SFXVolume", 0.8},
		{"DefaultGenre", "DefaultGenre", "fantasy"},
		{"VSync", "VSync", true},
		{"FullScreen", "FullScreen", false},
		{"MaxTPS", "MaxTPS", 60},
	}

	if err := Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Get()
			var actual interface{}
			switch tt.field {
			case "WindowWidth":
				actual = cfg.WindowWidth
			case "WindowHeight":
				actual = cfg.WindowHeight
			case "InternalWidth":
				actual = cfg.InternalWidth
			case "InternalHeight":
				actual = cfg.InternalHeight
			case "FOV":
				actual = cfg.FOV
			case "MouseSensitivity":
				actual = cfg.MouseSensitivity
			case "MasterVolume":
				actual = cfg.MasterVolume
			case "MusicVolume":
				actual = cfg.MusicVolume
			case "SFXVolume":
				actual = cfg.SFXVolume
			case "DefaultGenre":
				actual = cfg.DefaultGenre
			case "VSync":
				actual = cfg.VSync
			case "FullScreen":
				actual = cfg.FullScreen
			case "MaxTPS":
				actual = cfg.MaxTPS
			}
			if actual != tt.expected {
				t.Errorf("Config.%s = %v, want %v", tt.field, actual, tt.expected)
			}
		})
	}
}

func TestLoad_TOMLParsing(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configData := `
WindowWidth = 1920
WindowHeight = 1080
InternalWidth = 640
InternalHeight = 400
FOV = 90.0
MouseSensitivity = 1.5
MasterVolume = 0.9
MusicVolume = 0.6
SFXVolume = 0.7
DefaultGenre = "scifi"
VSync = false
FullScreen = true

[KeyBindings]
Forward = 87
Back = 83
`

	if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Reset viper and set config path
	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(tmpDir)

	// Set defaults before loading
	viper.SetDefault("WindowWidth", 960)
	viper.SetDefault("WindowHeight", 600)
	viper.SetDefault("InternalWidth", 320)
	viper.SetDefault("InternalHeight", 200)
	viper.SetDefault("FOV", 66.0)
	viper.SetDefault("MouseSensitivity", 1.0)
	viper.SetDefault("MasterVolume", 0.8)
	viper.SetDefault("MusicVolume", 0.7)
	viper.SetDefault("SFXVolume", 0.8)
	viper.SetDefault("DefaultGenre", "fantasy")
	viper.SetDefault("VSync", true)
	viper.SetDefault("FullScreen", false)
	viper.SetDefault("MaxTPS", 60)
	viper.SetDefault("KeyBindings", map[string]int{})

	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("viper.ReadInConfig() failed: %v", err)
	}

	if err := viper.Unmarshal(&C); err != nil {
		t.Fatalf("viper.Unmarshal() failed: %v", err)
	}

	cfg := Get()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"WindowWidth", cfg.WindowWidth, 1920},
		{"WindowHeight", cfg.WindowHeight, 1080},
		{"InternalWidth", cfg.InternalWidth, 640},
		{"InternalHeight", cfg.InternalHeight, 400},
		{"FOV", cfg.FOV, 90.0},
		{"MouseSensitivity", cfg.MouseSensitivity, 1.5},
		{"MasterVolume", cfg.MasterVolume, 0.9},
		{"MusicVolume", cfg.MusicVolume, 0.6},
		{"SFXVolume", cfg.SFXVolume, 0.7},
		{"DefaultGenre", cfg.DefaultGenre, "scifi"},
		{"VSync", cfg.VSync, false},
		{"FullScreen", cfg.FullScreen, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Config.%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	// Check key bindings (note: TOML lowercases keys)
	t.Logf("KeyBindings: %+v (length: %d)", cfg.KeyBindings, len(cfg.KeyBindings))

	if len(cfg.KeyBindings) != 2 {
		t.Errorf("len(KeyBindings) = %d, want 2", len(cfg.KeyBindings))
	}
	// TOML lowercases the keys
	if cfg.KeyBindings["forward"] != 87 {
		t.Errorf("KeyBindings[forward] = %d, want 87", cfg.KeyBindings["forward"])
	}
	if cfg.KeyBindings["back"] != 83 {
		t.Errorf("KeyBindings[back] = %d, want 83", cfg.KeyBindings["back"])
	}
}

func TestLoad_MissingFileFallback(t *testing.T) {
	// Reset viper with a non-existent path
	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("/nonexistent/path")

	// Should not error, just use defaults
	if err := Load(); err != nil {
		t.Errorf("Load() with missing file should not error, got: %v", err)
	}

	cfg := Get()
	if cfg.WindowWidth != 960 {
		t.Errorf("Default WindowWidth = %d, want 960", cfg.WindowWidth)
	}
}

func TestSave_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Reset viper
	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(tmpDir)

	// Load defaults
	if err := Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Modify config
	cfg := Config{
		WindowWidth:      1280,
		WindowHeight:     720,
		InternalWidth:    320,
		InternalHeight:   200,
		FOV:              75.0,
		MouseSensitivity: 2.0,
		MasterVolume:     0.5,
		MusicVolume:      0.4,
		SFXVolume:        0.6,
		DefaultGenre:     "horror",
		VSync:            false,
		FullScreen:       true,
		MaxTPS:           120,
		KeyBindings: map[string]int{
			"Jump": 32,
		},
	}
	Set(cfg)

	// Save config manually via viper
	viper.Set("WindowWidth", cfg.WindowWidth)
	viper.Set("WindowHeight", cfg.WindowHeight)
	viper.Set("InternalWidth", cfg.InternalWidth)
	viper.Set("InternalHeight", cfg.InternalHeight)
	viper.Set("FOV", cfg.FOV)
	viper.Set("MouseSensitivity", cfg.MouseSensitivity)
	viper.Set("MasterVolume", cfg.MasterVolume)
	viper.Set("MusicVolume", cfg.MusicVolume)
	viper.Set("SFXVolume", cfg.SFXVolume)
	viper.Set("DefaultGenre", cfg.DefaultGenre)
	viper.Set("VSync", cfg.VSync)
	viper.Set("FullScreen", cfg.FullScreen)
	viper.Set("MaxTPS", cfg.MaxTPS)
	viper.Set("KeyBindings", cfg.KeyBindings)

	if err := viper.WriteConfigAs(configPath); err != nil {
		t.Fatalf("viper.WriteConfigAs() failed: %v", err)
	}

	// Reset and reload
	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(tmpDir)

	if err := Load(); err != nil {
		t.Fatalf("Load() after save failed: %v", err)
	}

	newCfg := Get()
	if newCfg.WindowWidth != 1280 {
		t.Errorf("WindowWidth = %d, want 1280", newCfg.WindowWidth)
	}
	if newCfg.DefaultGenre != "horror" {
		t.Errorf("DefaultGenre = %s, want horror", newCfg.DefaultGenre)
	}
	if newCfg.MouseSensitivity != 2.0 {
		t.Errorf("MouseSensitivity = %f, want 2.0", newCfg.MouseSensitivity)
	}
}

func TestWatch_HotReload(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create initial config
	initialData := `
WindowWidth = 960
WindowHeight = 600
FOV = 66.0
DefaultGenre = "fantasy"
`
	if err := os.WriteFile(configPath, []byte(initialData), 0o644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Reset viper completely - critical for test isolation
	viper.Reset()

	// Reset global C to zero state to avoid pollution from other tests
	mu.Lock()
	C = Config{}
	mu.Unlock()

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(tmpDir)

	// Set defaults
	viper.SetDefault("WindowWidth", 960)
	viper.SetDefault("WindowHeight", 600)
	viper.SetDefault("FOV", 66.0)
	viper.SetDefault("DefaultGenre", "fantasy")

	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("viper.ReadInConfig() failed: %v", err)
	}

	mu.Lock()
	if err := viper.Unmarshal(&C); err != nil {
		mu.Unlock()
		t.Fatalf("viper.Unmarshal() failed: %v", err)
	}
	mu.Unlock()

	// Verify initial state
	initialCfg := Get()
	if initialCfg.WindowWidth != 960 {
		t.Fatalf("Initial WindowWidth = %d, want 960", initialCfg.WindowWidth)
	}

	// Track callback invocations
	var callbackCalled bool
	var newCfg Config
	var cbMu sync.Mutex

	callback := func(old, new Config) {
		cbMu.Lock()
		callbackCalled = true
		newCfg = new
		cbMu.Unlock()
		t.Logf("Hot-reload callback invoked: old.Width=%d, new.Width=%d", old.WindowWidth, new.WindowWidth)
	}

	// Start watching
	stop, err := Watch(callback)
	if err != nil {
		t.Fatalf("Watch() failed: %v", err)
	}
	defer stop()

	// Give fsnotify time to set up
	time.Sleep(100 * time.Millisecond)

	// Modify config file
	modifiedData := `
WindowWidth = 1920
WindowHeight = 1080
FOV = 90.0
DefaultGenre = "scifi"
`
	if err := os.WriteFile(configPath, []byte(modifiedData), 0o644); err != nil {
		t.Fatalf("Failed to write modified config: %v", err)
	}

	// Wait for fsnotify to detect change
	time.Sleep(500 * time.Millisecond)

	cbMu.Lock()
	called := callbackCalled
	cbMu.Unlock()

	if !called {
		t.Error("Callback was not called after config change")
		return
	}

	// Check that new config passed to callback has the updated values
	cbMu.Lock()
	if newCfg.WindowWidth != 1920 {
		t.Errorf("Callback new.WindowWidth = %d, want 1920", newCfg.WindowWidth)
	}
	if newCfg.DefaultGenre != "scifi" {
		t.Errorf("Callback new.DefaultGenre = %s, want scifi", newCfg.DefaultGenre)
	}
	cbMu.Unlock()

	// Check global config was updated to new values
	cfg := Get()
	if cfg.WindowWidth != 1920 {
		t.Errorf("Global WindowWidth = %d, want 1920", cfg.WindowWidth)
	}
	if cfg.FOV != 90.0 {
		t.Errorf("Global FOV = %f, want 90.0", cfg.FOV)
	}
	if cfg.DefaultGenre != "scifi" {
		t.Errorf("Global DefaultGenre = %s, want scifi", cfg.DefaultGenre)
	}
}

func TestWatch_NilCallback(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	initialData := `WindowWidth = 960`
	if err := os.WriteFile(configPath, []byte(initialData), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(tmpDir)

	if err := Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Watch with nil callback should not panic
	stop, err := Watch(nil)
	if err != nil {
		t.Fatalf("Watch(nil) failed: %v", err)
	}
	defer stop()

	time.Sleep(100 * time.Millisecond)

	// Modify config
	modifiedData := `WindowWidth = 1920`
	if err := os.WriteFile(configPath, []byte(modifiedData), 0o644); err != nil {
		t.Fatalf("Failed to write modified config: %v", err)
	}

	// Wait for change to be processed
	time.Sleep(500 * time.Millisecond)

	// Config should still be updated
	cfg := Get()
	if cfg.WindowWidth != 1920 {
		t.Errorf("WindowWidth = %d, want 1920", cfg.WindowWidth)
	}
}

func TestGetSet_Concurrency(t *testing.T) {
	// Reset viper
	viper.Reset()
	if err := Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = Get()
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				cfg := Get()
				cfg.WindowWidth = 800 + id
				Set(cfg)
			}
		}(i)
	}

	wg.Wait()

	// Should not panic or race
	cfg := Get()
	if cfg.WindowWidth < 800 || cfg.WindowWidth >= 810 {
		t.Logf("Final WindowWidth = %d (expected in range [800, 810))", cfg.WindowWidth)
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Write invalid TOML
	invalidData := `
WindowWidth = "not a number"
[[[invalid structure
`
	if err := os.WriteFile(configPath, []byte(invalidData), 0o644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(tmpDir)

	// Should return error for invalid TOML
	err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid TOML")
	}
}

func BenchmarkGet(b *testing.B) {
	viper.Reset()
	if err := Load(); err != nil {
		b.Fatalf("Load() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Get()
	}
}

func BenchmarkSet(b *testing.B) {
	viper.Reset()
	if err := Load(); err != nil {
		b.Fatalf("Load() failed: %v", err)
	}

	cfg := Get()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Set(cfg)
	}
}

func BenchmarkGetSet_Concurrent(b *testing.B) {
	viper.Reset()
	if err := Load(); err != nil {
		b.Fatalf("Load() failed: %v", err)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cfg := Get()
			cfg.WindowWidth = 1024
			Set(cfg)
		}
	})
}
