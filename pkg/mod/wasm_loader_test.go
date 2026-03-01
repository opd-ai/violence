package mod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultWASMConfig(t *testing.T) {
	config := DefaultWASMConfig()

	if config.MemoryLimitBytes != 64*1024*1024 {
		t.Errorf("expected memory limit 64MB, got %d", config.MemoryLimitBytes)
	}

	if config.FuelLimit != 1_000_000_000 {
		t.Errorf("expected fuel limit 1 billion, got %d", config.FuelLimit)
	}

	if !config.AllowFileRead {
		t.Error("expected AllowFileRead to be true")
	}

	if config.AllowFileWrite {
		t.Error("expected AllowFileWrite to be false")
	}

	if len(config.AllowedPaths) != 1 || config.AllowedPaths[0] != "mods/" {
		t.Errorf("expected AllowedPaths to be ['mods/'], got %v", config.AllowedPaths)
	}
}

func TestNewWASMLoader(t *testing.T) {
	loader := NewWASMLoader()

	if loader == nil {
		t.Fatal("expected non-nil loader")
	}

	if loader.config.MemoryLimitBytes != 64*1024*1024 {
		t.Error("expected default config")
	}

	if len(loader.modules) != 0 {
		t.Error("expected empty modules map")
	}
}

func TestNewWASMLoaderWithConfig(t *testing.T) {
	config := WASMConfig{
		MemoryLimitBytes: 128 * 1024 * 1024,
		FuelLimit:        2_000_000_000,
		AllowFileRead:    false,
		AllowFileWrite:   true,
		AllowedPaths:     []string{"custom/"},
	}

	loader := NewWASMLoaderWithConfig(config)

	if loader.config.MemoryLimitBytes != 128*1024*1024 {
		t.Error("expected custom memory limit")
	}

	if loader.config.FuelLimit != 2_000_000_000 {
		t.Error("expected custom fuel limit")
	}

	if loader.config.AllowFileRead {
		t.Error("expected AllowFileRead to be false")
	}

	if !loader.config.AllowFileWrite {
		t.Error("expected AllowFileWrite to be true")
	}
}

func TestWASMLoader_LoadWASM_FileNotFound(t *testing.T) {
	loader := NewWASMLoader()

	_, err := loader.LoadWASM("nonexistent.wasm")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestWASMLoader_LoadWASM_PathValidation(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	modsDir := filepath.Join(tmpDir, "mods")
	if err := os.MkdirAll(modsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a simple WASM file (minimal valid WASM)
	wasmPath := filepath.Join(modsDir, "test.wasm")
	// This is a minimal valid WASM module: magic number + version
	wasmBytes := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	if err := os.WriteFile(wasmPath, wasmBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	// Test with allowed path
	config := WASMConfig{
		MemoryLimitBytes: 64 * 1024 * 1024,
		FuelLimit:        1_000_000_000,
		AllowFileRead:    true,
		AllowFileWrite:   false,
		AllowedPaths:     []string{modsDir},
	}
	loader := NewWASMLoaderWithConfig(config)

	// This will fail to compile (invalid WASM), but path validation should pass
	_, err := loader.LoadWASM(wasmPath)
	if err != nil && err.Error() == "access denied: path outside allowed directories" {
		t.Errorf("path validation failed for allowed path: %v", err)
	}
}

func TestWASMLoader_UnloadWASM_NotLoaded(t *testing.T) {
	loader := NewWASMLoader()

	err := loader.UnloadWASM("nonexistent")
	if err == nil {
		t.Fatal("expected error for unloading nonexistent module")
	}

	if err.Error() != "module nonexistent not loaded" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestWASMLoader_GetModule_NotFound(t *testing.T) {
	loader := NewWASMLoader()

	_, err := loader.GetModule("nonexistent")
	if err == nil {
		t.Fatal("expected error for getting nonexistent module")
	}

	if err.Error() != "module nonexistent not found" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestWASMLoader_ListModules_Empty(t *testing.T) {
	loader := NewWASMLoader()

	modules := loader.ListModules()
	if len(modules) != 0 {
		t.Errorf("expected empty list, got %d modules", len(modules))
	}
}

func TestWASMLoader_isPathAllowed(t *testing.T) {
	tmpDir := t.TempDir()
	modsDir := filepath.Join(tmpDir, "mods")
	otherDir := filepath.Join(tmpDir, "other")

	config := WASMConfig{
		AllowedPaths: []string{modsDir},
	}
	loader := NewWASMLoaderWithConfig(config)

	tests := []struct {
		name    string
		path    string
		allowed bool
	}{
		{
			name:    "allowed_direct",
			path:    filepath.Join(modsDir, "test.wasm"),
			allowed: true,
		},
		{
			name:    "allowed_subdirectory",
			path:    filepath.Join(modsDir, "subdir", "test.wasm"),
			allowed: true,
		},
		{
			name:    "denied_outside",
			path:    filepath.Join(otherDir, "test.wasm"),
			allowed: false,
		},
		{
			name:    "denied_parent",
			path:    filepath.Join(tmpDir, "test.wasm"),
			allowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loader.isPathAllowed(tt.path)
			if result != tt.allowed {
				t.Errorf("expected %v, got %v for path %s", tt.allowed, result, tt.path)
			}
		})
	}
}

func TestWASMModule_HasExport(t *testing.T) {
	// This test requires a valid WASM module with exports
	// For now, we test the interface structure
	mod := &WASMModule{
		Name: "test",
		Path: "/test/path.wasm",
	}

	if mod.Name != "test" {
		t.Errorf("expected name 'test', got %s", mod.Name)
	}

	if mod.Path != "/test/path.wasm" {
		t.Errorf("expected path '/test/path.wasm', got %s", mod.Path)
	}
}

func TestWASMLoader_SecurityConstraints(t *testing.T) {
	tests := []struct {
		name   string
		config WASMConfig
	}{
		{
			name: "strict_security",
			config: WASMConfig{
				MemoryLimitBytes: 32 * 1024 * 1024, // 32MB
				FuelLimit:        500_000_000,      // 500M instructions
				AllowFileRead:    false,
				AllowFileWrite:   false,
				AllowedPaths:     []string{},
			},
		},
		{
			name: "permissive_security",
			config: WASMConfig{
				MemoryLimitBytes: 256 * 1024 * 1024, // 256MB
				FuelLimit:        5_000_000_000,     // 5B instructions
				AllowFileRead:    true,
				AllowFileWrite:   true,
				AllowedPaths:     []string{"mods/", "config/"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewWASMLoaderWithConfig(tt.config)

			if loader.config.MemoryLimitBytes != tt.config.MemoryLimitBytes {
				t.Errorf("memory limit mismatch")
			}

			if loader.config.FuelLimit != tt.config.FuelLimit {
				t.Errorf("fuel limit mismatch")
			}

			if loader.config.AllowFileRead != tt.config.AllowFileRead {
				t.Errorf("file read permission mismatch")
			}

			if loader.config.AllowFileWrite != tt.config.AllowFileWrite {
				t.Errorf("file write permission mismatch")
			}
		})
	}
}

// TestWASMLoader_ConcurrentAccess tests thread safety of the loader
func TestWASMLoader_ConcurrentAccess(t *testing.T) {
	loader := NewWASMLoader()

	done := make(chan bool, 10)

	// Concurrent list operations
	for i := 0; i < 10; i++ {
		go func() {
			loader.ListModules()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
