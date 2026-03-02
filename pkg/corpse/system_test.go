package corpse

import (
	"testing"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem(100, "fantasy", 12345)

	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}

	if sys.maxCorpses != 100 {
		t.Errorf("maxCorpses = %d, want 100", sys.maxCorpses)
	}

	if sys.genreID != "fantasy" {
		t.Errorf("genreID = %s, want fantasy", sys.genreID)
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem(100, "fantasy", 12345)

	genres := []string{"scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		sys.SetGenre(genre)
		if sys.genreID != genre {
			t.Errorf("After SetGenre(%s), genreID = %s", genre, sys.genreID)
		}
	}
}

func TestSpawnCorpse(t *testing.T) {
	sys := NewSystem(10, "fantasy", 12345)
	corpses := make([]Corpse, 0)

	sys.SpawnCorpse(&corpses, 100, 200, 42, "enemy", "humanoid", DeathNormal, 64, true)

	if len(corpses) != 1 {
		t.Fatalf("len(corpses) = %d, want 1", len(corpses))
	}

	c := corpses[0]
	if c.X != 100 || c.Y != 200 {
		t.Errorf("corpse position = (%f, %f), want (100, 200)", c.X, c.Y)
	}

	if c.Seed != 42 {
		t.Errorf("corpse seed = %d, want 42", c.Seed)
	}

	if c.EntityType != "enemy" {
		t.Errorf("corpse entityType = %s, want enemy", c.EntityType)
	}

	if c.DeathType != DeathNormal {
		t.Errorf("corpse deathType = %d, want %d", c.DeathType, DeathNormal)
	}

	if c.Size != 64 {
		t.Errorf("corpse size = %d, want 64", c.Size)
	}

	if !c.HasLoot {
		t.Error("corpse hasLoot = false, want true")
	}
}

func TestCorpseLimit(t *testing.T) {
	sys := NewSystem(5, "fantasy", 12345)
	corpses := make([]Corpse, 0)

	for i := 0; i < 10; i++ {
		sys.SpawnCorpse(&corpses, float64(i*10), float64(i*10), int64(i), "enemy", "humanoid", DeathNormal, 64, false)
	}

	if len(corpses) > 5 {
		t.Errorf("len(corpses) = %d, want <= 5", len(corpses))
	}
}

func TestUpdateCorpses(t *testing.T) {
	sys := NewSystem(10, "fantasy", 12345)
	corpses := make([]Corpse, 0)

	sys.SpawnCorpse(&corpses, 100, 200, 42, "enemy", "humanoid", DeathNormal, 64, false)

	initialOpacity := corpses[0].Opacity
	if initialOpacity != 1.0 {
		t.Errorf("initial opacity = %f, want 1.0", initialOpacity)
	}

	sys.UpdateCorpses(&corpses, 1.0)

	if corpses[0].Age != 1.0 {
		t.Errorf("corpse age = %f, want 1.0", corpses[0].Age)
	}

	if corpses[0].Opacity >= initialOpacity {
		t.Errorf("corpse opacity should decrease, got %f", corpses[0].Opacity)
	}
}

func TestCorpseFading(t *testing.T) {
	sys := NewSystem(10, "fantasy", 12345)
	corpses := make([]Corpse, 0)

	sys.SpawnCorpse(&corpses, 100, 200, 42, "enemy", "humanoid", DeathNormal, 64, false)
	maxAge := corpses[0].MaxAge

	sys.UpdateCorpses(&corpses, maxAge+1.0)

	if len(corpses) != 0 {
		t.Errorf("corpse should be removed after maxAge, len = %d", len(corpses))
	}
}

func TestDeathTypes(t *testing.T) {
	sys := NewSystem(10, "fantasy", 12345)
	corpses := make([]Corpse, 0)

	deathTypes := []DeathType{
		DeathNormal,
		DeathBurn,
		DeathFreeze,
		DeathElectric,
		DeathAcid,
		DeathExplosion,
		DeathSlash,
		DeathCrush,
		DeathDisintegrate,
	}

	for _, dt := range deathTypes {
		sys.SpawnCorpse(&corpses, 100, 200, 42, "enemy", "humanoid", dt, 64, false)
	}

	if len(corpses) != len(deathTypes) {
		t.Errorf("len(corpses) = %d, want %d", len(corpses), len(deathTypes))
	}

	for i, c := range corpses {
		if c.DeathType != deathTypes[i] {
			t.Errorf("corpse[%d].deathType = %d, want %d", i, c.DeathType, deathTypes[i])
		}
	}
}

func TestGetCorpseAt(t *testing.T) {
	sys := NewSystem(10, "fantasy", 12345)
	corpses := make([]Corpse, 0)

	sys.SpawnCorpse(&corpses, 100, 200, 42, "enemy", "humanoid", DeathNormal, 64, true)
	sys.SpawnCorpse(&corpses, 500, 600, 43, "enemy", "humanoid", DeathNormal, 64, false)

	found := sys.GetCorpseAt(corpses, 100, 200, 10.0)
	if found == nil {
		t.Error("GetCorpseAt should find corpse with loot")
	}

	notFound := sys.GetCorpseAt(corpses, 500, 600, 10.0)
	if notFound != nil {
		t.Error("GetCorpseAt should not find corpse without loot")
	}

	farAway := sys.GetCorpseAt(corpses, 1000, 1000, 10.0)
	if farAway != nil {
		t.Error("GetCorpseAt should not find distant corpse")
	}
}

func TestDetermineDeathType(t *testing.T) {
	tests := []struct {
		damageType string
		want       DeathType
	}{
		{"fire", DeathBurn},
		{"burn", DeathBurn},
		{"ice", DeathFreeze},
		{"freeze", DeathFreeze},
		{"electric", DeathElectric},
		{"lightning", DeathElectric},
		{"acid", DeathAcid},
		{"poison", DeathAcid},
		{"explosion", DeathExplosion},
		{"slash", DeathSlash},
		{"crush", DeathCrush},
		{"disintegrate", DeathDisintegrate},
		{"unknown", DeathNormal},
		{"", DeathNormal},
	}

	for _, tt := range tests {
		t.Run(tt.damageType, func(t *testing.T) {
			got := DetermineDeathType(tt.damageType)
			if got != tt.want {
				t.Errorf("DetermineDeathType(%s) = %d, want %d", tt.damageType, got, tt.want)
			}
		})
	}
}

func TestCorpseComponent(t *testing.T) {
	comp := &CorpseComponent{
		Corpses: make([]Corpse, 0),
	}

	if comp.Type() != "CorpseComponent" {
		t.Errorf("Type() = %s, want CorpseComponent", comp.Type())
	}
}

func TestGeneratorGenres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping generator test in short mode (requires display)")
	}

	gen := NewGenerator(50)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			gen.SetGenre(genre)

			img := gen.GetCorpseImage(12345, "enemy", DeathNormal, 0, 64)
			if img == nil {
				t.Error("GetCorpseImage returned nil")
			}

			bounds := img.Bounds()
			if bounds.Dx() != 64 || bounds.Dy() != 64 {
				t.Errorf("image size = %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestGeneratorAllDeathTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping generator test in short mode (requires display)")
	}

	gen := NewGenerator(50)
	gen.SetGenre("fantasy")

	deathTypes := []DeathType{
		DeathNormal,
		DeathBurn,
		DeathFreeze,
		DeathElectric,
		DeathAcid,
		DeathExplosion,
		DeathSlash,
		DeathCrush,
		DeathDisintegrate,
	}

	for _, dt := range deathTypes {
		t.Run(dt.String(), func(t *testing.T) {
			img := gen.GetCorpseImage(12345, "enemy", dt, 0, 64)
			if img == nil {
				t.Errorf("GetCorpseImage for %v returned nil", dt)
			}
		})
	}
}

func (dt DeathType) String() string {
	names := []string{
		"Normal", "Burn", "Freeze", "Electric", "Acid",
		"Explosion", "Slash", "Crush", "Disintegrate",
	}
	if int(dt) < len(names) {
		return names[dt]
	}
	return "Unknown"
}

func TestGeneratorCaching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping generator test in short mode (requires display)")
	}

	gen := NewGenerator(50)
	gen.SetGenre("fantasy")

	img1 := gen.GetCorpseImage(12345, "enemy", DeathNormal, 0, 64)
	img2 := gen.GetCorpseImage(12345, "enemy", DeathNormal, 0, 64)

	if img1 != img2 {
		t.Error("Generator should return cached image for same parameters")
	}

	img3 := gen.GetCorpseImage(54321, "enemy", DeathNormal, 0, 64)
	if img1 == img3 {
		t.Error("Generator should return different image for different seed")
	}
}

func TestGeneratorLRUEviction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping generator test in short mode (requires display)")
	}

	gen := NewGenerator(5)
	gen.SetGenre("fantasy")

	for i := 0; i < 10; i++ {
		gen.GetCorpseImage(int64(i), "enemy", DeathNormal, 0, 64)
	}

	gen.mu.RLock()
	cacheSize := len(gen.cache)
	gen.mu.RUnlock()

	if cacheSize > 5 {
		t.Errorf("cache size = %d, want <= 5", cacheSize)
	}
}
