package mod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader()
	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}
	if loader.mods == nil {
		t.Fatal("mods slice not initialized")
	}
	if loader.modsDir != "mods" {
		t.Fatalf("wrong default modsDir: got %q, want %q", loader.modsDir, "mods")
	}
}

func TestNewLoaderWithDir(t *testing.T) {
	loader := NewLoaderWithDir("/custom/mods")
	if loader.modsDir != "/custom/mods" {
		t.Fatalf("wrong modsDir: got %q, want %q", loader.modsDir, "/custom/mods")
	}
}

func TestLoader_LoadMod(t *testing.T) {
	// Create temp directory for test mod
	tmpDir := t.TempDir()
	modDir := filepath.Join(tmpDir, "testmod")
	if err := os.Mkdir(modDir, 0o755); err != nil {
		t.Fatalf("failed to create mod dir: %v", err)
	}

	// Write mod.json
	modJSON := `{
		"name": "TestMod",
		"version": "1.0.0",
		"description": "A test mod",
		"author": "Test Author",
		"config": {
			"key1": "value1"
		}
	}`
	manifestPath := filepath.Join(modDir, "mod.json")
	if err := os.WriteFile(manifestPath, []byte(modJSON), 0o644); err != nil {
		t.Fatalf("failed to write mod.json: %v", err)
	}

	// Load mod
	loader := NewLoader()
	if err := loader.LoadMod(modDir); err != nil {
		t.Fatalf("LoadMod failed: %v", err)
	}

	// Verify mod was loaded
	mods := loader.ListMods()
	if len(mods) != 1 {
		t.Fatalf("expected 1 mod, got %d", len(mods))
	}

	mod := mods[0]
	if mod.Name != "TestMod" {
		t.Fatalf("wrong name: got %q, want %q", mod.Name, "TestMod")
	}
	if mod.Version != "1.0.0" {
		t.Fatalf("wrong version: got %q, want %q", mod.Version, "1.0.0")
	}
	if mod.Description != "A test mod" {
		t.Fatalf("wrong description: got %q, want %q", mod.Description, "A test mod")
	}
	if mod.Author != "Test Author" {
		t.Fatalf("wrong author: got %q, want %q", mod.Author, "Test Author")
	}
	if mod.Path != modDir {
		t.Fatalf("wrong path: got %q, want %q", mod.Path, modDir)
	}
	if !mod.Enabled {
		t.Fatal("mod should be enabled by default")
	}
	if mod.Config["key1"] != "value1" {
		t.Fatalf("wrong config value: got %q, want %q", mod.Config["key1"], "value1")
	}
}

func TestLoader_LoadModNoManifest(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader()
	err := loader.LoadMod(tmpDir)
	if err == nil {
		t.Fatal("expected error when mod.json missing")
	}
}

func TestLoader_LoadModInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "mod.json")
	if err := os.WriteFile(manifestPath, []byte("invalid json"), 0o644); err != nil {
		t.Fatalf("failed to write invalid json: %v", err)
	}

	loader := NewLoader()
	err := loader.LoadMod(tmpDir)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoader_LoadModMissingName(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "mod.json")
	modJSON := `{"version": "1.0.0"}`
	if err := os.WriteFile(manifestPath, []byte(modJSON), 0o644); err != nil {
		t.Fatalf("failed to write mod.json: %v", err)
	}

	loader := NewLoader()
	err := loader.LoadMod(tmpDir)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestLoader_LoadModMissingVersion(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "mod.json")
	modJSON := `{"name": "TestMod"}`
	if err := os.WriteFile(manifestPath, []byte(modJSON), 0o644); err != nil {
		t.Fatalf("failed to write mod.json: %v", err)
	}

	loader := NewLoader()
	err := loader.LoadMod(tmpDir)
	if err == nil {
		t.Fatal("expected error for missing version")
	}
}

func TestLoader_LoadModAlreadyLoaded(t *testing.T) {
	tmpDir := t.TempDir()
	modDir := filepath.Join(tmpDir, "testmod")
	if err := os.Mkdir(modDir, 0o755); err != nil {
		t.Fatalf("failed to create mod dir: %v", err)
	}

	manifestPath := filepath.Join(modDir, "mod.json")
	modJSON := `{"name": "TestMod", "version": "1.0.0"}`
	if err := os.WriteFile(manifestPath, []byte(modJSON), 0o644); err != nil {
		t.Fatalf("failed to write mod.json: %v", err)
	}

	loader := NewLoader()
	if err := loader.LoadMod(modDir); err != nil {
		t.Fatalf("first LoadMod failed: %v", err)
	}

	// Try to load again
	err := loader.LoadMod(modDir)
	if err == nil {
		t.Fatal("expected error when loading same mod twice")
	}
}

func TestLoader_UnloadMod(t *testing.T) {
	tmpDir := t.TempDir()
	modDir := filepath.Join(tmpDir, "testmod")
	if err := os.Mkdir(modDir, 0o755); err != nil {
		t.Fatalf("failed to create mod dir: %v", err)
	}

	manifestPath := filepath.Join(modDir, "mod.json")
	modJSON := `{"name": "TestMod", "version": "1.0.0"}`
	if err := os.WriteFile(manifestPath, []byte(modJSON), 0o644); err != nil {
		t.Fatalf("failed to write mod.json: %v", err)
	}

	loader := NewLoader()
	if err := loader.LoadMod(modDir); err != nil {
		t.Fatalf("LoadMod failed: %v", err)
	}

	// Unload mod
	if err := loader.UnloadMod("TestMod"); err != nil {
		t.Fatalf("UnloadMod failed: %v", err)
	}

	mods := loader.ListMods()
	if len(mods) != 0 {
		t.Fatalf("expected 0 mods after unload, got %d", len(mods))
	}
}

func TestLoader_UnloadModNotFound(t *testing.T) {
	loader := NewLoader()
	err := loader.UnloadMod("NonexistentMod")
	if err == nil {
		t.Fatal("expected error when unloading nonexistent mod")
	}
}

func TestLoader_GetMod(t *testing.T) {
	tmpDir := t.TempDir()
	modDir := filepath.Join(tmpDir, "testmod")
	if err := os.Mkdir(modDir, 0o755); err != nil {
		t.Fatalf("failed to create mod dir: %v", err)
	}

	manifestPath := filepath.Join(modDir, "mod.json")
	modJSON := `{"name": "TestMod", "version": "1.0.0"}`
	if err := os.WriteFile(manifestPath, []byte(modJSON), 0o644); err != nil {
		t.Fatalf("failed to write mod.json: %v", err)
	}

	loader := NewLoader()
	if err := loader.LoadMod(modDir); err != nil {
		t.Fatalf("LoadMod failed: %v", err)
	}

	mod, err := loader.GetMod("TestMod")
	if err != nil {
		t.Fatalf("GetMod failed: %v", err)
	}
	if mod.Name != "TestMod" {
		t.Fatalf("wrong mod name: got %q, want %q", mod.Name, "TestMod")
	}
}

func TestLoader_GetModNotFound(t *testing.T) {
	loader := NewLoader()
	_, err := loader.GetMod("NonexistentMod")
	if err == nil {
		t.Fatal("expected error for nonexistent mod")
	}
}

func TestLoader_EnableDisableMod(t *testing.T) {
	tmpDir := t.TempDir()
	modDir := filepath.Join(tmpDir, "testmod")
	if err := os.Mkdir(modDir, 0o755); err != nil {
		t.Fatalf("failed to create mod dir: %v", err)
	}

	manifestPath := filepath.Join(modDir, "mod.json")
	modJSON := `{"name": "TestMod", "version": "1.0.0"}`
	if err := os.WriteFile(manifestPath, []byte(modJSON), 0o644); err != nil {
		t.Fatalf("failed to write mod.json: %v", err)
	}

	loader := NewLoader()
	if err := loader.LoadMod(modDir); err != nil {
		t.Fatalf("LoadMod failed: %v", err)
	}

	// Disable mod
	if err := loader.DisableMod("TestMod"); err != nil {
		t.Fatalf("DisableMod failed: %v", err)
	}

	mod, _ := loader.GetMod("TestMod")
	if mod.Enabled {
		t.Fatal("mod should be disabled")
	}

	// Enable mod
	if err := loader.EnableMod("TestMod"); err != nil {
		t.Fatalf("EnableMod failed: %v", err)
	}

	mod, _ = loader.GetMod("TestMod")
	if !mod.Enabled {
		t.Fatal("mod should be enabled")
	}
}

func TestLoader_EnableModNotFound(t *testing.T) {
	loader := NewLoader()
	err := loader.EnableMod("NonexistentMod")
	if err == nil {
		t.Fatal("expected error for nonexistent mod")
	}
}

func TestLoader_DisableModNotFound(t *testing.T) {
	loader := NewLoader()
	err := loader.DisableMod("NonexistentMod")
	if err == nil {
		t.Fatal("expected error for nonexistent mod")
	}
}

func TestLoader_AddConflict(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two conflicting mods
	mod1Dir := filepath.Join(tmpDir, "mod1")
	mod2Dir := filepath.Join(tmpDir, "mod2")
	if err := os.Mkdir(mod1Dir, 0o755); err != nil {
		t.Fatalf("failed to create mod1 dir: %v", err)
	}
	if err := os.Mkdir(mod2Dir, 0o755); err != nil {
		t.Fatalf("failed to create mod2 dir: %v", err)
	}

	mod1JSON := `{"name": "Mod1", "version": "1.0.0"}`
	mod2JSON := `{"name": "Mod2", "version": "1.0.0"}`
	if err := os.WriteFile(filepath.Join(mod1Dir, "mod.json"), []byte(mod1JSON), 0o644); err != nil {
		t.Fatalf("failed to write mod1.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mod2Dir, "mod.json"), []byte(mod2JSON), 0o644); err != nil {
		t.Fatalf("failed to write mod2.json: %v", err)
	}

	loader := NewLoader()
	loader.AddConflict("Mod1", "Mod2")

	// Load Mod1
	if err := loader.LoadMod(mod1Dir); err != nil {
		t.Fatalf("LoadMod Mod1 failed: %v", err)
	}

	// Try to load Mod2 - should fail due to conflict
	err := loader.LoadMod(mod2Dir)
	if err == nil {
		t.Fatal("expected error when loading conflicting mod")
	}
}

func TestLoader_SetGetModsDir(t *testing.T) {
	loader := NewLoader()
	loader.SetModsDir("/custom/mods")

	dir := loader.GetModsDir()
	if dir != "/custom/mods" {
		t.Fatalf("wrong modsDir: got %q, want %q", dir, "/custom/mods")
	}
}

func TestSetGenre(t *testing.T) {
	// Should not panic
	SetGenre("fantasy")
	SetGenre("scifi")
	SetGenre("")
}
