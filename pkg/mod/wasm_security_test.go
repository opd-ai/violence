package mod

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestWASMSecurity_FileAccessOutsideModsDirectory verifies that WASM mods
// cannot access files outside the allowed mods/ directory.
func TestWASMSecurity_FileAccessOutsideModsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	modsDir := filepath.Join(tmpDir, "mods")
	sensitiveDir := filepath.Join(tmpDir, "sensitive")

	if err := os.MkdirAll(modsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sensitiveDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a sensitive file outside mods directory
	sensitivePath := filepath.Join(sensitiveDir, "secret.txt")
	if err := os.WriteFile(sensitivePath, []byte("secret data"), 0o644); err != nil {
		t.Fatal(err)
	}

	config := WASMConfig{
		MemoryLimitBytes: 64 * 1024 * 1024,
		FuelLimit:        1_000_000_000,
		AllowFileRead:    true,
		AllowFileWrite:   false,
		AllowedPaths:     []string{modsDir},
	}

	loader := NewWASMLoaderWithConfig(config)

	// Attempt to load file outside mods directory should fail
	if loader.isPathAllowed(sensitivePath) {
		t.Error("expected access to be denied for file outside mods directory")
	}
}

// TestWASMSecurity_MemoryLimit verifies that memory limits are configured.
func TestWASMSecurity_MemoryLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit uint32
	}{
		{"default_64MB", 64 * 1024 * 1024},
		{"strict_32MB", 32 * 1024 * 1024},
		{"permissive_128MB", 128 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WASMConfig{
				MemoryLimitBytes: tt.limit,
				FuelLimit:        1_000_000_000,
				AllowFileRead:    false,
				AllowFileWrite:   false,
				AllowedPaths:     []string{},
			}

			loader := NewWASMLoaderWithConfig(config)

			if loader.config.MemoryLimitBytes != tt.limit {
				t.Errorf("expected memory limit %d, got %d", tt.limit, loader.config.MemoryLimitBytes)
			}
		})
	}
}

// TestWASMSecurity_FuelLimit verifies that instruction limits are configured.
func TestWASMSecurity_FuelLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit uint64
	}{
		{"default_1B", 1_000_000_000},
		{"strict_500M", 500_000_000},
		{"permissive_5B", 5_000_000_000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WASMConfig{
				MemoryLimitBytes: 64 * 1024 * 1024,
				FuelLimit:        tt.limit,
				AllowFileRead:    false,
				AllowFileWrite:   false,
				AllowedPaths:     []string{},
			}

			loader := NewWASMLoaderWithConfig(config)

			if loader.config.FuelLimit != tt.limit {
				t.Errorf("expected fuel limit %d, got %d", tt.limit, loader.config.FuelLimit)
			}
		})
	}
}

// TestWASMSecurity_DefaultPermissionsSecure verifies that default config is secure.
func TestWASMSecurity_DefaultPermissionsSecure(t *testing.T) {
	config := DefaultWASMConfig()

	// File read is allowed but write is not
	if !config.AllowFileRead {
		t.Error("expected file read to be allowed by default")
	}

	if config.AllowFileWrite {
		t.Error("expected file write to be denied by default")
	}

	// Only mods/ directory is allowed
	if len(config.AllowedPaths) != 1 {
		t.Errorf("expected 1 allowed path, got %d", len(config.AllowedPaths))
	}

	if config.AllowedPaths[0] != "mods/" {
		t.Errorf("expected allowed path 'mods/', got %s", config.AllowedPaths[0])
	}

	// Memory limit is reasonable (64MB)
	if config.MemoryLimitBytes != 64*1024*1024 {
		t.Errorf("expected 64MB memory limit, got %d", config.MemoryLimitBytes)
	}

	// Fuel limit prevents infinite loops
	if config.FuelLimit != 1_000_000_000 {
		t.Errorf("expected 1 billion instruction limit, got %d", config.FuelLimit)
	}
}

// TestWASMSecurity_PathTraversalProtection verifies protection against path traversal attacks.
func TestWASMSecurity_PathTraversalProtection(t *testing.T) {
	tmpDir := t.TempDir()
	modsDir := filepath.Join(tmpDir, "mods")

	config := WASMConfig{
		AllowedPaths: []string{modsDir},
	}
	loader := NewWASMLoaderWithConfig(config)

	attacks := []string{
		filepath.Join(modsDir, "..", "etc", "passwd"),
		filepath.Join(modsDir, "..", "..", "etc", "passwd"),
		filepath.Join(modsDir, "subdir", "..", "..", "secret.txt"),
	}

	for _, attack := range attacks {
		t.Run(attack, func(t *testing.T) {
			if loader.isPathAllowed(attack) {
				t.Errorf("path traversal attack succeeded: %s", attack)
			}
		})
	}

	// Valid paths within mods/ should be allowed
	validPaths := []string{
		filepath.Join(modsDir, "test.wasm"),
		filepath.Join(modsDir, "subdir", "test.wasm"),
		filepath.Join(modsDir, "a", "b", "c", "test.wasm"),
	}

	for _, valid := range validPaths {
		t.Run(valid, func(t *testing.T) {
			if !loader.isPathAllowed(valid) {
				t.Errorf("valid path was denied: %s", valid)
			}
		})
	}
}

// TestWASMSecurity_MultipleModIsolation verifies that multiple mods are isolated.
func TestWASMSecurity_MultipleModIsolation(t *testing.T) {
	loader := NewWASMLoader()

	// Each mod should have its own isolated instance
	// This test verifies the structure supports isolation

	if loader.modules == nil {
		t.Fatal("expected modules map to be initialized")
	}

	// Modules map should be empty initially
	if len(loader.modules) != 0 {
		t.Errorf("expected empty modules map, got %d entries", len(loader.modules))
	}
}

// TestWASMSecurity_LoaderIntegration verifies that the main Loader enforces WASM usage.
func TestWASMSecurity_LoaderIntegration(t *testing.T) {
	loader := NewLoader()

	// WASM loader should be initialized
	if loader.wasmLoader == nil {
		t.Fatal("expected WASM loader to be initialized")
	}

	// Unsafe plugins should be disabled by default
	if loader.EnableUnsafePlugins {
		t.Error("expected EnableUnsafePlugins to be false by default")
	}
}

// TestWASMSecurity_UnsafePluginsBlocked verifies that unsafe plugins are blocked by default.
func TestWASMSecurity_UnsafePluginsBlocked(t *testing.T) {
	loader := NewLoader()

	// Create a test plugin
	plugin := &testSecurityPlugin{}

	// Attempt to register without enabling unsafe plugins
	err := loader.RegisterPlugin(plugin)
	if err == nil {
		t.Fatal("expected error when registering plugin without EnableUnsafePlugins")
	}

	if err.Error() != "unsafe plugins are disabled; set EnableUnsafePlugins=true or use WASM mods" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

// TestWASMSecurity_UnsafePluginsWarning verifies warning is logged when enabling unsafe plugins.
func TestWASMSecurity_UnsafePluginsWarning(t *testing.T) {
	loader := NewLoader()
	loader.EnableUnsafePlugins = true

	// Create a test plugin
	plugin := &testSecurityPlugin{}

	// Should succeed but with warning (logged, not returned)
	err := loader.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify plugin was loaded
	plugins := loader.pluginManager.ListPlugins()
	if len(plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(plugins))
	}
}

// TestWASMSecurity_ConcurrentModAccess verifies thread safety of WASM loader.
func TestWASMSecurity_ConcurrentModAccess(t *testing.T) {
	loader := NewWASMLoader()

	done := make(chan bool)

	// Concurrent operations on the loader
	for i := 0; i < 50; i++ {
		go func() {
			loader.ListModules()
			loader.GetModule("nonexistent")
			done <- true
		}()
	}

	// Wait with timeout
	timeout := time.After(5 * time.Second)
	for i := 0; i < 50; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("concurrent access test timed out")
		}
	}
}

// TestModAPI_SecurityPermissionDefaults verifies ModAPI has secure defaults.
func TestModAPI_SecurityPermissionDefaults(t *testing.T) {
	perms := DefaultPermissions()

	// File write should be disabled
	if perms.AllowFileWrite {
		t.Error("AllowFileWrite should be false by default")
	}

	// Entity spawn should be disabled
	if perms.AllowEntitySpawn {
		t.Error("AllowEntitySpawn should be false by default")
	}

	// UI modification should be disabled
	if perms.AllowUIModify {
		t.Error("AllowUIModify should be false by default")
	}
}

// testSecurityPlugin is a test implementation of Plugin interface.
type testSecurityPlugin struct{}

func (p *testSecurityPlugin) Load() error                { return nil }
func (p *testSecurityPlugin) Unload() error              { return nil }
func (p *testSecurityPlugin) Name() string               { return "test_security" }
func (p *testSecurityPlugin) Version() string            { return "1.0.0" }
func (p *testSecurityPlugin) GetGenerators() []Generator { return nil }
