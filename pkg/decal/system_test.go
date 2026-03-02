package decal

import (
	"testing"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem(500, "fantasy", 12345)
	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}
	if sys.maxDecals != 500 {
		t.Errorf("maxDecals = %d, want 500", sys.maxDecals)
	}
	if sys.genreID != "fantasy" {
		t.Errorf("genreID = %s, want fantasy", sys.genreID)
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem(100, "fantasy", 12345)
	sys.SetGenre("scifi")
	if sys.genreID != "scifi" {
		t.Errorf("genreID = %s, want scifi", sys.genreID)
	}
}

func TestSpawnDecal(t *testing.T) {
	sys := NewSystem(10, "fantasy", 12345)
	decals := make([]Decal, 0)

	sys.SpawnDecal(&decals, 5.0, 10.0, DecalBlood, 0, 0)

	if len(decals) != 1 {
		t.Fatalf("len(decals) = %d, want 1", len(decals))
	}

	d := decals[0]
	if d.X != 5.0 || d.Y != 10.0 {
		t.Errorf("position = (%.1f, %.1f), want (5.0, 10.0)", d.X, d.Y)
	}
	if d.Type != DecalBlood {
		t.Errorf("type = %v, want DecalBlood", d.Type)
	}
	if d.Opacity != 1.0 {
		t.Errorf("opacity = %.2f, want 1.0", d.Opacity)
	}
}

func TestSpawnDecalMaxLimit(t *testing.T) {
	sys := NewSystem(5, "fantasy", 12345)
	decals := make([]Decal, 0)

	// Spawn 10 decals but limit is 5
	for i := 0; i < 10; i++ {
		sys.SpawnDecal(&decals, float64(i), float64(i), DecalBlood, 0, 0)
	}

	if len(decals) != 5 {
		t.Errorf("len(decals) = %d, want 5 (max limit)", len(decals))
	}

	// Check that oldest were removed (should have indices 5-9)
	if decals[0].X != 5.0 {
		t.Errorf("oldest decal X = %.1f, want 5.0", decals[0].X)
	}
}

func TestUpdateDecals(t *testing.T) {
	sys := NewSystem(100, "fantasy", 12345)
	decals := make([]Decal, 0)

	sys.SpawnDecal(&decals, 0, 0, DecalBlood, 0, 0)

	if decals[0].Age != 0 {
		t.Errorf("initial age = %.2f, want 0", decals[0].Age)
	}

	// Advance time
	sys.UpdateDecals(&decals, 1.0)

	if len(decals) != 1 {
		t.Fatal("decal should still exist after 1 second")
	}
	if decals[0].Age != 1.0 {
		t.Errorf("age after 1s = %.2f, want 1.0", decals[0].Age)
	}
	if decals[0].Opacity >= 1.0 {
		t.Error("opacity should have decreased")
	}
}

func TestUpdateDecalsRemovesFaded(t *testing.T) {
	sys := NewSystem(100, "fantasy", 12345)
	decals := make([]Decal, 0)

	sys.SpawnDecal(&decals, 0, 0, DecalBlood, 0, 0)
	maxAge := decals[0].MaxAge

	// Age beyond max
	sys.UpdateDecals(&decals, maxAge+10.0)

	if len(decals) != 0 {
		t.Errorf("len(decals) = %d, want 0 (should be removed)", len(decals))
	}
}

func TestSpawnBloodSplatter(t *testing.T) {
	sys := NewSystem(100, "fantasy", 12345)
	decals := make([]Decal, 0)

	sys.SpawnBloodSplatter(&decals, 5.0, 5.0, 1.0, 0.0)

	// Should create main splatter + droplets
	if len(decals) < 2 {
		t.Errorf("len(decals) = %d, want at least 2 (main + droplets)", len(decals))
	}

	// All should be blood type
	for i, d := range decals {
		if d.Type != DecalBlood {
			t.Errorf("decal[%d].Type = %v, want DecalBlood", i, d.Type)
		}
	}
}

func TestSpawnImpactMark(t *testing.T) {
	tests := []struct {
		damageType string
		wantType   DecalType
	}{
		{"fire", DecalScorch},
		{"explosion", DecalScorch},
		{"slash", DecalSlash},
		{"pierce", DecalSlash},
		{"projectile", DecalBulletHole},
		{"ballistic", DecalBulletHole},
		{"magic", DecalMagicBurn},
		{"arcane", DecalMagicBurn},
		{"acid", DecalAcid},
		{"poison", DecalAcid},
		{"ice", DecalFreeze},
		{"frost", DecalFreeze},
		{"lightning", DecalElectric},
		{"electric", DecalElectric},
		{"unknown", DecalBlood},
	}

	for _, tt := range tests {
		t.Run(tt.damageType, func(t *testing.T) {
			sys := NewSystem(100, "fantasy", 12345)
			decals := make([]Decal, 0)

			sys.SpawnImpactMark(&decals, 0, 0, 1, 0, tt.damageType)

			if len(decals) != 1 {
				t.Fatalf("len(decals) = %d, want 1", len(decals))
			}
			if decals[0].Type != tt.wantType {
				t.Errorf("decal type = %v, want %v", decals[0].Type, tt.wantType)
			}
		})
	}
}

func TestDecalVariation(t *testing.T) {
	sys := NewSystem(100, "fantasy", 12345)
	decals := make([]Decal, 0)

	// Spawn multiple decals, should have variation
	for i := 0; i < 10; i++ {
		sys.SpawnDecal(&decals, 0, 0, DecalBlood, 0, 0)
	}

	// Check that seeds differ (variation)
	seedMap := make(map[int64]bool)
	for _, d := range decals {
		seedMap[d.Seed] = true
	}

	if len(seedMap) < 5 {
		t.Errorf("only %d unique seeds in 10 decals, want more variation", len(seedMap))
	}
}

func TestUpdateNilDecals(t *testing.T) {
	sys := NewSystem(100, "fantasy", 12345)

	// Should not panic
	sys.UpdateDecals(nil, 1.0)
	sys.SpawnDecal(nil, 0, 0, DecalBlood, 0, 0)
	sys.SpawnBloodSplatter(nil, 0, 0, 0, 0)
	sys.SpawnImpactMark(nil, 0, 0, 0, 0, "fire")
}
