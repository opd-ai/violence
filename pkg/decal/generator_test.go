package decal

import (
	"testing"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator(100)
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if gen.maxEntries != 100 {
		t.Errorf("maxEntries = %d, want 100", gen.maxEntries)
	}
	if gen.imagePool == nil {
		t.Error("imagePool is nil")
	}
}

func TestGeneratorSetGenre(t *testing.T) {
	gen := NewGenerator(100)
	gen.SetGenre("scifi")
	if gen.genreID != "scifi" {
		t.Errorf("genreID = %s, want scifi", gen.genreID)
	}
}

func TestDecalKeyStructure(t *testing.T) {
	// Test that DecalKey can be used as map key
	m := make(map[DecalKey]int)
	k1 := DecalKey{Type: DecalBlood, Subtype: 0, Seed: 123, Size: 32}
	k2 := DecalKey{Type: DecalBlood, Subtype: 0, Seed: 123, Size: 32}
	k3 := DecalKey{Type: DecalBlood, Subtype: 1, Seed: 123, Size: 32}

	m[k1] = 1
	m[k2] = 2 // Should overwrite k1
	m[k3] = 3

	if m[k1] != 2 {
		t.Errorf("k1 value = %d, want 2 (should equal k2)", m[k1])
	}
	if len(m) != 2 {
		t.Errorf("map length = %d, want 2", len(m))
	}
}

func TestBloodColorByGenre(t *testing.T) {
	tests := []struct {
		genre string
		wantR uint8
	}{
		{"fantasy", 200},
		{"scifi", 100},
		{"horror", 120},
		{"cyberpunk", 255},
		{"unknown", 200}, // default
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			gen := NewGenerator(10)
			gen.SetGenre(tt.genre)
			c := gen.getBloodColor()
			if c.R != tt.wantR {
				t.Errorf("blood color R = %d, want %d for genre %s", c.R, tt.wantR, tt.genre)
			}
		})
	}
}

func TestDecalTypeString(t *testing.T) {
	// Helper to ensure DecalType can be used in string context
	dt := DecalBlood
	_ = dt // Just check it compiles
}

// Helper method for DecalType
func (d DecalType) String() string {
	names := []string{
		"Blood",
		"Scorch",
		"Slash",
		"BulletHole",
		"MagicBurn",
		"Acid",
		"Freeze",
		"Electric",
	}
	if int(d) < len(names) {
		return names[d]
	}
	return "Unknown"
}
