package save

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// setupTestDir creates a temporary directory for testing.
func setupTestDir(t *testing.T) (string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "violence_save_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	cleanup := func() {
		os.Setenv("HOME", originalHome)
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestSave(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	tests := []struct {
		name    string
		slot    int
		state   *GameState
		wantErr bool
	}{
		{
			name: "valid save to slot 1",
			slot: 1,
			state: &GameState{
				Seed:  12345,
				Genre: "fantasy",
				Player: Player{
					X:      10.5,
					Y:      20.3,
					DirX:   1.0,
					DirY:   0.0,
					Pitch:  15.0,
					Health: 100,
					Armor:  50,
					Ammo:   200,
				},
				Map: Map{
					Width:  10,
					Height: 10,
					Tiles: [][]int{
						{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
						{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
						{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
						{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
					},
				},
				Inventory: Inventory{
					Items: []Item{
						{ID: "key1", Name: "Red Key", Qty: 1},
						{ID: "ammo", Name: "Bullets", Qty: 50},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid slot negative",
			slot:    -1,
			state:   &GameState{Seed: 123},
			wantErr: true,
		},
		{
			name:    "invalid slot too large",
			slot:    MaxSlots,
			state:   &GameState{Seed: 123},
			wantErr: true,
		},
		{
			name:    "nil state",
			slot:    1,
			state:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Save(tt.slot, tt.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify file exists for valid saves
			if !tt.wantErr && tt.state != nil {
				slotPath, _ := getSlotPath(tt.slot)
				if _, err := os.Stat(slotPath); os.IsNotExist(err) {
					t.Errorf("Save file was not created at %s", slotPath)
				}
			}
		})
	}
}

func TestLoad(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a test save
	testState := &GameState{
		Seed:  99999,
		Genre: "scifi",
		Player: Player{
			X:      5.5,
			Y:      8.2,
			DirX:   0.707,
			DirY:   0.707,
			Pitch:  -10.0,
			Health: 75,
			Armor:  25,
			Ammo:   150,
		},
		Map: Map{
			Width:  5,
			Height: 5,
			Tiles:  [][]int{{1, 1, 1, 1, 1}},
		},
		Inventory: Inventory{
			Items: []Item{
				{ID: "medkit", Name: "Medkit", Qty: 2},
			},
		},
	}

	if err := Save(3, testState); err != nil {
		t.Fatalf("failed to create test save: %v", err)
	}

	tests := []struct {
		name    string
		slot    int
		wantErr bool
		check   func(*testing.T, *GameState)
	}{
		{
			name:    "load existing save",
			slot:    3,
			wantErr: false,
			check: func(t *testing.T, state *GameState) {
				if state.Seed != 99999 {
					t.Errorf("Seed = %d, want 99999", state.Seed)
				}
				if state.Genre != "scifi" {
					t.Errorf("Genre = %s, want scifi", state.Genre)
				}
				if state.Player.X != 5.5 {
					t.Errorf("Player.X = %f, want 5.5", state.Player.X)
				}
				if state.Player.Health != 75 {
					t.Errorf("Player.Health = %d, want 75", state.Player.Health)
				}
				if len(state.Inventory.Items) != 1 {
					t.Errorf("Inventory items count = %d, want 1", len(state.Inventory.Items))
				}
				if state.Version != "1.0" {
					t.Errorf("Version = %s, want 1.0", state.Version)
				}
			},
		},
		{
			name:    "load non-existent slot",
			slot:    5,
			wantErr: true,
			check:   nil,
		},
		{
			name:    "invalid slot negative",
			slot:    -1,
			wantErr: true,
			check:   nil,
		},
		{
			name:    "invalid slot too large",
			slot:    MaxSlots,
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := Load(tt.slot)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, state)
			}
		})
	}
}

func TestAutoSave(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	state := &GameState{
		Seed:  77777,
		Genre: "horror",
		Player: Player{
			X:      1.0,
			Y:      2.0,
			Health: 90,
		},
	}

	err := AutoSave(state)
	if err != nil {
		t.Fatalf("AutoSave() error = %v", err)
	}

	// Verify it saved to slot 0
	loaded, err := Load(AutoSaveSlot)
	if err != nil {
		t.Fatalf("Load(AutoSaveSlot) error = %v", err)
	}

	if loaded.Seed != 77777 {
		t.Errorf("Seed = %d, want 77777", loaded.Seed)
	}
	if loaded.Genre != "horror" {
		t.Errorf("Genre = %s, want horror", loaded.Genre)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	tests := []struct {
		name  string
		slot  int
		state *GameState
	}{
		{
			name: "full state round trip",
			slot: 2,
			state: &GameState{
				Seed:  42424242,
				Genre: "cyberpunk",
				Player: Player{
					X:      100.5,
					Y:      200.3,
					DirX:   0.866,
					DirY:   0.5,
					Pitch:  22.5,
					Health: 100,
					Armor:  100,
					Ammo:   500,
				},
				Map: Map{
					Width:  20,
					Height: 20,
					Tiles: [][]int{
						{1, 1, 1, 1, 1},
						{1, 0, 2, 3, 1},
						{1, 0, 0, 0, 1},
						{1, 4, 0, 0, 1},
						{1, 1, 1, 1, 1},
					},
				},
				Inventory: Inventory{
					Items: []Item{
						{ID: "key_red", Name: "Red Keycard", Qty: 1},
						{ID: "key_blue", Name: "Blue Keycard", Qty: 1},
						{ID: "key_yellow", Name: "Yellow Keycard", Qty: 1},
						{ID: "ammo_pistol", Name: "Pistol Ammo", Qty: 100},
						{ID: "medkit_small", Name: "Small Medkit", Qty: 3},
					},
				},
			},
		},
		{
			name: "minimal state round trip",
			slot: 4,
			state: &GameState{
				Seed:  1,
				Genre: "fantasy",
				Player: Player{
					X:      0.0,
					Y:      0.0,
					DirX:   1.0,
					DirY:   0.0,
					Pitch:  0.0,
					Health: 100,
					Armor:  0,
					Ammo:   0,
				},
				Map: Map{
					Width:  1,
					Height: 1,
					Tiles:  [][]int{{0}},
				},
				Inventory: Inventory{
					Items: []Item{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save
			if err := Save(tt.slot, tt.state); err != nil {
				t.Fatalf("Save() error = %v", err)
			}

			// Load
			loaded, err := Load(tt.slot)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// Compare
			if loaded.Seed != tt.state.Seed {
				t.Errorf("Seed = %d, want %d", loaded.Seed, tt.state.Seed)
			}
			if loaded.Genre != tt.state.Genre {
				t.Errorf("Genre = %s, want %s", loaded.Genre, tt.state.Genre)
			}
			if loaded.Player.X != tt.state.Player.X {
				t.Errorf("Player.X = %f, want %f", loaded.Player.X, tt.state.Player.X)
			}
			if loaded.Player.Y != tt.state.Player.Y {
				t.Errorf("Player.Y = %f, want %f", loaded.Player.Y, tt.state.Player.Y)
			}
			if loaded.Player.DirX != tt.state.Player.DirX {
				t.Errorf("Player.DirX = %f, want %f", loaded.Player.DirX, tt.state.Player.DirX)
			}
			if loaded.Player.DirY != tt.state.Player.DirY {
				t.Errorf("Player.DirY = %f, want %f", loaded.Player.DirY, tt.state.Player.DirY)
			}
			if loaded.Player.Pitch != tt.state.Player.Pitch {
				t.Errorf("Player.Pitch = %f, want %f", loaded.Player.Pitch, tt.state.Player.Pitch)
			}
			if loaded.Player.Health != tt.state.Player.Health {
				t.Errorf("Player.Health = %d, want %d", loaded.Player.Health, tt.state.Player.Health)
			}
			if loaded.Player.Armor != tt.state.Player.Armor {
				t.Errorf("Player.Armor = %d, want %d", loaded.Player.Armor, tt.state.Player.Armor)
			}
			if loaded.Player.Ammo != tt.state.Player.Ammo {
				t.Errorf("Player.Ammo = %d, want %d", loaded.Player.Ammo, tt.state.Player.Ammo)
			}
			if loaded.Map.Width != tt.state.Map.Width {
				t.Errorf("Map.Width = %d, want %d", loaded.Map.Width, tt.state.Map.Width)
			}
			if loaded.Map.Height != tt.state.Map.Height {
				t.Errorf("Map.Height = %d, want %d", loaded.Map.Height, tt.state.Map.Height)
			}
			if len(loaded.Inventory.Items) != len(tt.state.Inventory.Items) {
				t.Errorf("Inventory items count = %d, want %d", len(loaded.Inventory.Items), len(tt.state.Inventory.Items))
			}
		})
	}
}

func TestListSlots(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	// Create some test saves
	saves := []struct {
		slot  int
		state *GameState
	}{
		{
			slot: 1,
			state: &GameState{
				Seed:   111,
				Genre:  "fantasy",
				Player: Player{Health: 100},
			},
		},
		{
			slot: 3,
			state: &GameState{
				Seed:   333,
				Genre:  "scifi",
				Player: Player{Health: 75},
			},
		},
		{
			slot: 7,
			state: &GameState{
				Seed:   777,
				Genre:  "horror",
				Player: Player{Health: 50},
			},
		},
	}

	for _, s := range saves {
		if err := Save(s.slot, s.state); err != nil {
			t.Fatalf("failed to create test save: %v", err)
		}
		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	slots, err := ListSlots()
	if err != nil {
		t.Fatalf("ListSlots() error = %v", err)
	}

	if len(slots) != MaxSlots {
		t.Errorf("ListSlots() returned %d slots, want %d", len(slots), MaxSlots)
	}

	// Check that expected slots exist
	for _, s := range saves {
		slot := slots[s.slot]
		if !slot.Exists {
			t.Errorf("Slot %d should exist but doesn't", s.slot)
		}
		if slot.Seed != s.state.Seed {
			t.Errorf("Slot %d Seed = %d, want %d", s.slot, slot.Seed, s.state.Seed)
		}
		if slot.Genre != s.state.Genre {
			t.Errorf("Slot %d Genre = %s, want %s", s.slot, slot.Genre, s.state.Genre)
		}
	}

	// Check that other slots don't exist
	for i := 0; i < MaxSlots; i++ {
		isUsed := false
		for _, s := range saves {
			if s.slot == i {
				isUsed = true
				break
			}
		}
		if !isUsed && slots[i].Exists {
			t.Errorf("Slot %d should not exist but does", i)
		}
	}
}

func TestDeleteSlot(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a test save
	state := &GameState{
		Seed:   12345,
		Genre:  "fantasy",
		Player: Player{Health: 100},
	}
	if err := Save(2, state); err != nil {
		t.Fatalf("failed to create test save: %v", err)
	}

	tests := []struct {
		name    string
		slot    int
		wantErr bool
	}{
		{
			name:    "delete existing slot",
			slot:    2,
			wantErr: false,
		},
		{
			name:    "delete non-existent slot",
			slot:    5,
			wantErr: true,
		},
		{
			name:    "delete invalid slot negative",
			slot:    -1,
			wantErr: true,
		},
		{
			name:    "delete invalid slot too large",
			slot:    MaxSlots,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DeleteSlot(tt.slot)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteSlot() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify file is gone for successful deletes
			if !tt.wantErr {
				slotPath, _ := getSlotPath(tt.slot)
				if _, err := os.Stat(slotPath); !os.IsNotExist(err) {
					t.Errorf("Save file still exists at %s after delete", slotPath)
				}
			}
		})
	}
}

func TestCrossPlatformPath(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	savePath, err := getSavePath()
	if err != nil {
		t.Fatalf("getSavePath() error = %v", err)
	}

	// Verify path contains .violence/saves
	if !filepath.IsAbs(savePath) {
		t.Errorf("getSavePath() returned non-absolute path: %s", savePath)
	}

	if filepath.Base(savePath) != "saves" {
		t.Errorf("getSavePath() doesn't end with 'saves': %s", savePath)
	}

	// Verify directory was created
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Errorf("getSavePath() didn't create directory: %s", savePath)
	}

	// Test that saves work with the path
	state := &GameState{
		Seed:   999,
		Genre:  "postapoc",
		Player: Player{Health: 50},
	}
	if err := Save(1, state); err != nil {
		t.Fatalf("Save() with cross-platform path error = %v", err)
	}

	slotPath, _ := getSlotPath(1)
	if !filepath.IsAbs(slotPath) {
		t.Errorf("getSlotPath() returned non-absolute path: %s", slotPath)
	}

	t.Logf("Cross-platform save path: %s", savePath)
	t.Logf("Cross-platform slot path: %s", slotPath)
}

func BenchmarkSave(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "violence_save_bench_*")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	state := &GameState{
		Seed:  12345,
		Genre: "fantasy",
		Player: Player{
			X: 10.5, Y: 20.3, DirX: 1.0, DirY: 0.0,
			Pitch: 15.0, Health: 100, Armor: 50, Ammo: 200,
		},
		Map: Map{
			Width:  64,
			Height: 64,
			Tiles:  make([][]int, 64),
		},
		Inventory: Inventory{
			Items: make([]Item, 20),
		},
	}

	for i := 0; i < 64; i++ {
		state.Map.Tiles[i] = make([]int, 64)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Save(1, state)
	}
}

func BenchmarkLoad(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "violence_save_bench_*")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	state := &GameState{
		Seed:  12345,
		Genre: "fantasy",
		Player: Player{
			X: 10.5, Y: 20.3, DirX: 1.0, DirY: 0.0,
			Pitch: 15.0, Health: 100, Armor: 50, Ammo: 200,
		},
		Map: Map{
			Width:  64,
			Height: 64,
			Tiles:  make([][]int, 64),
		},
		Inventory: Inventory{
			Items: make([]Item, 20),
		},
	}

	for i := 0; i < 64; i++ {
		state.Map.Tiles[i] = make([]int, 64)
	}

	Save(1, state)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Load(1)
	}
}

func TestGetSavePath_PlatformSpecific(t *testing.T) {
	_, cleanup := setupTestDir(t)
	defer cleanup()

	// Also set APPDATA for Windows testing
	originalAppData := os.Getenv("APPDATA")
	if runtime.GOOS == "windows" {
		tempAppData, err := os.MkdirTemp("", "appdata_test_*")
		if err != nil {
			t.Fatalf("failed to create temp appdata: %v", err)
		}
		defer os.RemoveAll(tempAppData)
		os.Setenv("APPDATA", tempAppData)
		defer os.Setenv("APPDATA", originalAppData)
	}

	savePath, err := getSavePath()
	if err != nil {
		t.Fatalf("getSavePath failed: %v", err)
	}

	// Verify path structure based on OS
	if runtime.GOOS == "windows" {
		// On Windows, should be %APPDATA%\violence\saves
		expectedSuffix := filepath.Join("violence", "saves")
		if !filepath.IsAbs(savePath) {
			t.Errorf("save path should be absolute, got: %s", savePath)
		}
		if !filepath.HasPrefix(savePath, os.Getenv("APPDATA")) && os.Getenv("APPDATA") != "" {
			t.Errorf("Windows save path should use APPDATA, got: %s", savePath)
		}
		if !filepath.Match("*"+expectedSuffix, savePath) {
			t.Logf("Windows save path: %s (expected suffix: %s)", savePath, expectedSuffix)
		}
	} else {
		// On Unix/Linux/macOS, should be ~/.violence/saves
		expectedSuffix := filepath.Join(".violence", "saves")
		if !filepath.IsAbs(savePath) {
			t.Errorf("save path should be absolute, got: %s", savePath)
		}
		home, _ := os.UserHomeDir()
		if !filepath.HasPrefix(savePath, home) {
			t.Errorf("Unix save path should use home directory, got: %s", savePath)
		}
		if !filepath.Match("*"+expectedSuffix, savePath) {
			t.Logf("Unix save path: %s (expected suffix: %s)", savePath, expectedSuffix)
		}
	}

	// Verify directory was created
	info, err := os.Stat(savePath)
	if err != nil {
		t.Fatalf("save path should exist: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("save path should be a directory")
	}
}
