package save

import (
	"encoding/json"
	"testing"
)

func TestSaveState_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name  string
		state SaveState
	}{
		{
			name: "minimal save state",
			state: SaveState{
				LevelSeed: 12345,
				PlayerPosition: Position{
					X: 10.5,
					Y: 20.3,
				},
				Health: HealthData{
					Current: 100,
					Max:     100,
				},
				Armor: 50,
				Inventory: InventoryData{
					Items:   []string{},
					Credits: 0,
				},
				DiscoveredTiles:  map[string]bool{},
				CurrentObjective: "",
				CameraDirection: CameraData{
					DirX:         0,
					DirY:         -1,
					PlaneX:       0.66,
					PlaneY:       0,
					FOV:          66.0,
					PitchRadians: 0,
				},
			},
		},
		{
			name: "populated save state",
			state: SaveState{
				LevelSeed: 98765,
				PlayerPosition: Position{
					X: 42.7,
					Y: 31.2,
				},
				Health: HealthData{
					Current: 75,
					Max:     100,
				},
				Armor: 25,
				Inventory: InventoryData{
					Items:   []string{"medkit", "keycard_red", "ammo_shells"},
					Credits: 150,
				},
				DiscoveredTiles: map[string]bool{
					"0,0":   true,
					"1,0":   true,
					"0,1":   true,
					"10,15": true,
				},
				CurrentObjective: "find_exit",
				CameraDirection: CameraData{
					DirX:         0.707,
					DirY:         -0.707,
					PlaneX:       0.47,
					PlaneY:       0.47,
					FOV:          66.0,
					PitchRadians: 0.1,
				},
			},
		},
		{
			name: "damaged player with full inventory",
			state: SaveState{
				LevelSeed: 55555,
				PlayerPosition: Position{
					X: 128.9,
					Y: 256.4,
				},
				Health: HealthData{
					Current: 23,
					Max:     150,
				},
				Armor: 100,
				Inventory: InventoryData{
					Items: []string{
						"weapon_pistol",
						"weapon_shotgun",
						"weapon_rifle",
						"medkit",
						"medkit",
						"keycard_blue",
						"keycard_red",
						"keycard_yellow",
						"ammo_bullets",
						"ammo_shells",
					},
					Credits: 9999,
				},
				DiscoveredTiles: map[string]bool{
					"5,5":   true,
					"5,6":   true,
					"6,5":   true,
					"6,6":   true,
					"100,0": true,
				},
				CurrentObjective: "defeat_boss",
				CameraDirection: CameraData{
					DirX:         -1,
					DirY:         0,
					PlaneX:       0,
					PlaneY:       0.66,
					FOV:          90.0,
					PitchRadians: -0.3,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.state)
			if err != nil {
				t.Fatalf("failed to marshal SaveState: %v", err)
			}

			// Unmarshal back to SaveState
			var unmarshaled SaveState
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("failed to unmarshal SaveState: %v", err)
			}

			// Verify LevelSeed
			if unmarshaled.LevelSeed != tt.state.LevelSeed {
				t.Errorf("LevelSeed mismatch: got %d, want %d", unmarshaled.LevelSeed, tt.state.LevelSeed)
			}

			// Verify PlayerPosition
			if unmarshaled.PlayerPosition.X != tt.state.PlayerPosition.X {
				t.Errorf("PlayerPosition.X mismatch: got %f, want %f", unmarshaled.PlayerPosition.X, tt.state.PlayerPosition.X)
			}
			if unmarshaled.PlayerPosition.Y != tt.state.PlayerPosition.Y {
				t.Errorf("PlayerPosition.Y mismatch: got %f, want %f", unmarshaled.PlayerPosition.Y, tt.state.PlayerPosition.Y)
			}

			// Verify Health
			if unmarshaled.Health.Current != tt.state.Health.Current {
				t.Errorf("Health.Current mismatch: got %d, want %d", unmarshaled.Health.Current, tt.state.Health.Current)
			}
			if unmarshaled.Health.Max != tt.state.Health.Max {
				t.Errorf("Health.Max mismatch: got %d, want %d", unmarshaled.Health.Max, tt.state.Health.Max)
			}

			// Verify Armor
			if unmarshaled.Armor != tt.state.Armor {
				t.Errorf("Armor mismatch: got %d, want %d", unmarshaled.Armor, tt.state.Armor)
			}

			// Verify Inventory
			if len(unmarshaled.Inventory.Items) != len(tt.state.Inventory.Items) {
				t.Errorf("Inventory.Items length mismatch: got %d, want %d", len(unmarshaled.Inventory.Items), len(tt.state.Inventory.Items))
			}
			for i := range tt.state.Inventory.Items {
				if i >= len(unmarshaled.Inventory.Items) {
					break
				}
				if unmarshaled.Inventory.Items[i] != tt.state.Inventory.Items[i] {
					t.Errorf("Inventory.Items[%d] mismatch: got %s, want %s", i, unmarshaled.Inventory.Items[i], tt.state.Inventory.Items[i])
				}
			}
			if unmarshaled.Inventory.Credits != tt.state.Inventory.Credits {
				t.Errorf("Inventory.Credits mismatch: got %d, want %d", unmarshaled.Inventory.Credits, tt.state.Inventory.Credits)
			}

			// Verify DiscoveredTiles
			if len(unmarshaled.DiscoveredTiles) != len(tt.state.DiscoveredTiles) {
				t.Errorf("DiscoveredTiles length mismatch: got %d, want %d", len(unmarshaled.DiscoveredTiles), len(tt.state.DiscoveredTiles))
			}
			for key, val := range tt.state.DiscoveredTiles {
				if unmarshaled.DiscoveredTiles[key] != val {
					t.Errorf("DiscoveredTiles[%s] mismatch: got %v, want %v", key, unmarshaled.DiscoveredTiles[key], val)
				}
			}

			// Verify CurrentObjective
			if unmarshaled.CurrentObjective != tt.state.CurrentObjective {
				t.Errorf("CurrentObjective mismatch: got %s, want %s", unmarshaled.CurrentObjective, tt.state.CurrentObjective)
			}

			// Verify CameraDirection
			if unmarshaled.CameraDirection.DirX != tt.state.CameraDirection.DirX {
				t.Errorf("CameraDirection.DirX mismatch: got %f, want %f", unmarshaled.CameraDirection.DirX, tt.state.CameraDirection.DirX)
			}
			if unmarshaled.CameraDirection.DirY != tt.state.CameraDirection.DirY {
				t.Errorf("CameraDirection.DirY mismatch: got %f, want %f", unmarshaled.CameraDirection.DirY, tt.state.CameraDirection.DirY)
			}
			if unmarshaled.CameraDirection.PlaneX != tt.state.CameraDirection.PlaneX {
				t.Errorf("CameraDirection.PlaneX mismatch: got %f, want %f", unmarshaled.CameraDirection.PlaneX, tt.state.CameraDirection.PlaneX)
			}
			if unmarshaled.CameraDirection.PlaneY != tt.state.CameraDirection.PlaneY {
				t.Errorf("CameraDirection.PlaneY mismatch: got %f, want %f", unmarshaled.CameraDirection.PlaneY, tt.state.CameraDirection.PlaneY)
			}
			if unmarshaled.CameraDirection.FOV != tt.state.CameraDirection.FOV {
				t.Errorf("CameraDirection.FOV mismatch: got %f, want %f", unmarshaled.CameraDirection.FOV, tt.state.CameraDirection.FOV)
			}
			if unmarshaled.CameraDirection.PitchRadians != tt.state.CameraDirection.PitchRadians {
				t.Errorf("CameraDirection.PitchRadians mismatch: got %f, want %f", unmarshaled.CameraDirection.PitchRadians, tt.state.CameraDirection.PitchRadians)
			}
		})
	}
}

func TestSaveState_ZeroValues(t *testing.T) {
	state := SaveState{}

	// Marshal zero-value state
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("failed to marshal zero SaveState: %v", err)
	}

	// Unmarshal back
	var unmarshaled SaveState
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal zero SaveState: %v", err)
	}

	// Verify zero values are preserved
	if unmarshaled.LevelSeed != 0 {
		t.Errorf("LevelSeed should be 0, got %d", unmarshaled.LevelSeed)
	}
	if unmarshaled.PlayerPosition.X != 0 || unmarshaled.PlayerPosition.Y != 0 {
		t.Errorf("PlayerPosition should be (0,0), got (%f,%f)", unmarshaled.PlayerPosition.X, unmarshaled.PlayerPosition.Y)
	}
	if unmarshaled.Health.Current != 0 || unmarshaled.Health.Max != 0 {
		t.Errorf("Health should be 0/0, got %d/%d", unmarshaled.Health.Current, unmarshaled.Health.Max)
	}
	if unmarshaled.Armor != 0 {
		t.Errorf("Armor should be 0, got %d", unmarshaled.Armor)
	}
	if unmarshaled.CurrentObjective != "" {
		t.Errorf("CurrentObjective should be empty, got %s", unmarshaled.CurrentObjective)
	}
}

func TestSaveState_NegativeValues(t *testing.T) {
	// Test that negative values are serialized correctly
	state := SaveState{
		LevelSeed: -12345,
		PlayerPosition: Position{
			X: -50.5,
			Y: -100.7,
		},
		Health: HealthData{
			Current: -10, // May occur in edge cases
			Max:     100,
		},
		Armor: -5,
		CameraDirection: CameraData{
			DirX:         -0.707,
			DirY:         -0.707,
			PitchRadians: -1.57,
		},
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("failed to marshal negative SaveState: %v", err)
	}

	var unmarshaled SaveState
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal negative SaveState: %v", err)
	}

	if unmarshaled.LevelSeed != state.LevelSeed {
		t.Errorf("LevelSeed mismatch: got %d, want %d", unmarshaled.LevelSeed, state.LevelSeed)
	}
	if unmarshaled.PlayerPosition.X != state.PlayerPosition.X {
		t.Errorf("PlayerPosition.X mismatch: got %f, want %f", unmarshaled.PlayerPosition.X, state.PlayerPosition.X)
	}
	if unmarshaled.Health.Current != state.Health.Current {
		t.Errorf("Health.Current mismatch: got %d, want %d", unmarshaled.Health.Current, state.Health.Current)
	}
}

func TestSaveState_LargeInventory(t *testing.T) {
	// Test with large inventory
	items := make([]string, 1000)
	for i := range items {
		items[i] = "item_" + string(rune(i))
	}

	state := SaveState{
		Inventory: InventoryData{
			Items:   items,
			Credits: 999999999,
		},
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("failed to marshal large inventory: %v", err)
	}

	var unmarshaled SaveState
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal large inventory: %v", err)
	}

	if len(unmarshaled.Inventory.Items) != len(state.Inventory.Items) {
		t.Errorf("Inventory.Items length mismatch: got %d, want %d", len(unmarshaled.Inventory.Items), len(state.Inventory.Items))
	}
	if unmarshaled.Inventory.Credits != state.Inventory.Credits {
		t.Errorf("Inventory.Credits mismatch: got %d, want %d", unmarshaled.Inventory.Credits, state.Inventory.Credits)
	}
}

func TestSaveState_LargeDiscoveredTiles(t *testing.T) {
	// Test with many discovered tiles
	tiles := make(map[string]bool)
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			key := string(rune(x)) + "," + string(rune(y))
			tiles[key] = true
		}
	}

	state := SaveState{
		DiscoveredTiles: tiles,
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("failed to marshal large DiscoveredTiles: %v", err)
	}

	var unmarshaled SaveState
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal large DiscoveredTiles: %v", err)
	}

	if len(unmarshaled.DiscoveredTiles) != len(state.DiscoveredTiles) {
		t.Errorf("DiscoveredTiles length mismatch: got %d, want %d", len(unmarshaled.DiscoveredTiles), len(state.DiscoveredTiles))
	}
}

func TestPosition_Struct(t *testing.T) {
	pos := Position{X: 123.456, Y: 789.012}

	data, err := json.Marshal(pos)
	if err != nil {
		t.Fatalf("failed to marshal Position: %v", err)
	}

	var unmarshaled Position
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal Position: %v", err)
	}

	if unmarshaled.X != pos.X || unmarshaled.Y != pos.Y {
		t.Errorf("Position mismatch: got (%f,%f), want (%f,%f)", unmarshaled.X, unmarshaled.Y, pos.X, pos.Y)
	}
}

func TestHealthData_Struct(t *testing.T) {
	health := HealthData{Current: 75, Max: 100}

	data, err := json.Marshal(health)
	if err != nil {
		t.Fatalf("failed to marshal HealthData: %v", err)
	}

	var unmarshaled HealthData
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal HealthData: %v", err)
	}

	if unmarshaled.Current != health.Current || unmarshaled.Max != health.Max {
		t.Errorf("HealthData mismatch: got %d/%d, want %d/%d", unmarshaled.Current, unmarshaled.Max, health.Current, health.Max)
	}
}

func TestInventoryData_Struct(t *testing.T) {
	inv := InventoryData{
		Items:   []string{"a", "b", "c"},
		Credits: 500,
	}

	data, err := json.Marshal(inv)
	if err != nil {
		t.Fatalf("failed to marshal InventoryData: %v", err)
	}

	var unmarshaled InventoryData
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal InventoryData: %v", err)
	}

	if len(unmarshaled.Items) != len(inv.Items) {
		t.Errorf("InventoryData.Items length mismatch: got %d, want %d", len(unmarshaled.Items), len(inv.Items))
	}
	if unmarshaled.Credits != inv.Credits {
		t.Errorf("InventoryData.Credits mismatch: got %d, want %d", unmarshaled.Credits, inv.Credits)
	}
}

func TestCameraData_Struct(t *testing.T) {
	cam := CameraData{
		DirX:         1.0,
		DirY:         0.0,
		PlaneX:       0.0,
		PlaneY:       0.66,
		FOV:          66.0,
		PitchRadians: 0.5,
	}

	data, err := json.Marshal(cam)
	if err != nil {
		t.Fatalf("failed to marshal CameraData: %v", err)
	}

	var unmarshaled CameraData
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal CameraData: %v", err)
	}

	if unmarshaled.DirX != cam.DirX || unmarshaled.DirY != cam.DirY {
		t.Errorf("CameraData Dir mismatch: got (%f,%f), want (%f,%f)", unmarshaled.DirX, unmarshaled.DirY, cam.DirX, cam.DirY)
	}
	if unmarshaled.PlaneX != cam.PlaneX || unmarshaled.PlaneY != cam.PlaneY {
		t.Errorf("CameraData Plane mismatch: got (%f,%f), want (%f,%f)", unmarshaled.PlaneX, unmarshaled.PlaneY, cam.PlaneX, cam.PlaneY)
	}
	if unmarshaled.FOV != cam.FOV {
		t.Errorf("CameraData.FOV mismatch: got %f, want %f", unmarshaled.FOV, cam.FOV)
	}
	if unmarshaled.PitchRadians != cam.PitchRadians {
		t.Errorf("CameraData.PitchRadians mismatch: got %f, want %f", unmarshaled.PitchRadians, cam.PitchRadians)
	}
}
