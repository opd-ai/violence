package sprite

import (
	"testing"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator(100)
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if gen.maxEntries != 100 {
		t.Errorf("Expected maxEntries=100, got %d", gen.maxEntries)
	}
	if gen.genreID != "fantasy" {
		t.Errorf("Expected default genre='fantasy', got %s", gen.genreID)
	}
}

func TestSetGenre(t *testing.T) {
	gen := NewGenerator(50)
	gen.SetGenre("scifi")

	gen.mu.RLock()
	defer gen.mu.RUnlock()
	if gen.genreID != "scifi" {
		t.Errorf("Expected genre='scifi', got %s", gen.genreID)
	}
}

func TestGetSprite_PropBarrel(t *testing.T) {
	gen := NewGenerator(10)
	sprite := gen.GetSprite(SpriteProp, "barrel", 12345, 0, 64)

	if sprite == nil {
		t.Fatal("GetSprite returned nil for barrel")
	}

	bounds := sprite.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		t.Errorf("Expected 64x64 sprite, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestGetSprite_Caching(t *testing.T) {
	gen := NewGenerator(5)

	sprite1 := gen.GetSprite(SpriteProp, "crate", 9999, 0, 32)
	sprite2 := gen.GetSprite(SpriteProp, "crate", 9999, 0, 32)

	if sprite1 != sprite2 {
		t.Error("Expected cached sprite to be returned (same pointer)")
	}

	gen.mu.RLock()
	cacheSize := gen.lruList.Len()
	gen.mu.RUnlock()

	if cacheSize != 1 {
		t.Errorf("Expected cache size=1, got %d", cacheSize)
	}
}

func TestGetSprite_LRUEviction(t *testing.T) {
	gen := NewGenerator(2)

	gen.GetSprite(SpriteProp, "barrel", 1, 0, 64)
	gen.GetSprite(SpriteProp, "crate", 2, 0, 64)
	gen.GetSprite(SpriteProp, "table", 3, 0, 64)

	gen.mu.RLock()
	cacheSize := gen.lruList.Len()
	gen.mu.RUnlock()

	if cacheSize != 2 {
		t.Errorf("Expected LRU cache to maintain size=2, got %d", cacheSize)
	}
}

func TestGetSprite_DifferentTypes(t *testing.T) {
	gen := NewGenerator(20)

	testCases := []struct {
		spriteType SpriteType
		subtype    string
		seed       int64
	}{
		{SpriteProp, "barrel", 111},
		{SpriteProp, "crate", 222},
		{SpriteProp, "table", 333},
		{SpriteProp, "terminal", 444},
		{SpriteProp, "bones", 555},
		{SpriteProp, "plant", 666},
		{SpriteProp, "pillar", 777},
		{SpriteProp, "torch", 888},
		{SpriteProp, "debris", 999},
		{SpriteProp, "container", 1010},
		{SpriteLoreItem, "note", 1111},
		{SpriteLoreItem, "audiolog", 1212},
		{SpriteLoreItem, "graffiti", 1313},
		{SpriteLoreItem, "body", 1414},
		{SpriteDestructible, "barrel", 1515},
		{SpriteDestructible, "crate", 1616},
		{SpritePickup, "health", 1717},
		{SpritePickup, "ammo", 1818},
		{SpritePickup, "armor", 1919},
		{SpriteProjectile, "bullet", 2020},
	}

	for _, tc := range testCases {
		sprite := gen.GetSprite(tc.spriteType, tc.subtype, tc.seed, 0, 64)
		if sprite == nil {
			t.Errorf("GetSprite(%v, %s, %d) returned nil", tc.spriteType, tc.subtype, tc.seed)
		}
	}
}

func TestGetSprite_GenreColors(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		gen := NewGenerator(10)
		gen.SetGenre(genre)

		sprite := gen.GetSprite(SpriteProp, "crate", 42, 0, 64)
		if sprite == nil {
			t.Errorf("Failed to generate crate sprite for genre=%s", genre)
		}

		woodColor := gen.getGenreWoodColor()
		if woodColor.R == 0 && woodColor.G == 0 && woodColor.B == 0 {
			t.Errorf("Genre %s returned invalid wood color (all zeros)", genre)
		}

		stoneColor := gen.getGenreStoneColor()
		if stoneColor.R == 0 && stoneColor.G == 0 && stoneColor.B == 0 {
			t.Errorf("Genre %s returned invalid stone color (all zeros)", genre)
		}

		leafColor := gen.getGenreLeafColor()
		if leafColor.R == 0 && leafColor.G == 0 && leafColor.B == 0 {
			t.Errorf("Genre %s returned invalid leaf color (all zeros)", genre)
		}
	}
}

func TestGetSprite_AnimationFrames(t *testing.T) {
	gen := NewGenerator(20)

	frame0 := gen.GetSprite(SpriteProp, "torch", 123, 0, 64)
	frame5 := gen.GetSprite(SpriteProp, "torch", 123, 5, 64)
	frame10 := gen.GetSprite(SpriteProp, "torch", 123, 10, 64)

	if frame0 == nil || frame5 == nil || frame10 == nil {
		t.Fatal("Animation frames returned nil sprites")
	}

	if frame0 == frame5 {
		t.Error("Expected different cache entries for different frames")
	}

	if frame5 == frame10 {
		t.Error("Expected different cache entries for different frames")
	}
}

func BenchmarkGetSprite_Uncached(b *testing.B) {
	gen := NewGenerator(1)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gen.GetSprite(SpriteProp, "barrel", int64(i), 0, 64)
	}
}

func BenchmarkGetSprite_Cached(b *testing.B) {
	gen := NewGenerator(100)
	gen.GetSprite(SpriteProp, "barrel", 42, 0, 64)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gen.GetSprite(SpriteProp, "barrel", 42, 0, 64)
	}
}

func BenchmarkGeneration_AllPropTypes(b *testing.B) {
	subtypes := []string{"barrel", "crate", "table", "terminal", "bones", "plant", "pillar", "torch", "debris", "container"}
	gen := NewGenerator(100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		subtype := subtypes[i%len(subtypes)]
		gen.GetSprite(SpriteProp, subtype, int64(i), 0, 64)
	}
}
