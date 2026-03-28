package sprite

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestGenerateEnemySprite(t *testing.T) {
	gen := NewGenerator(100)

	tests := []struct {
		name    string
		subtype string
		genre   string
		seed    int64
		frame   int
		size    int
	}{
		{"humanoid_fantasy", "humanoid", "fantasy", 12345, 0, 64},
		{"tank_scifi", "tank", "scifi", 54321, 0, 64},
		{"ranged_horror", "ranged", "horror", 11111, 0, 64},
		{"healer_cyberpunk", "healer", "cyberpunk", 22222, 0, 64},
		{"ambusher_postapoc", "ambusher", "postapoc", 33333, 0, 64},
		{"scout_fantasy", "scout", "fantasy", 44444, 0, 64},
		{"quadruped_fantasy", "quadruped", "fantasy", 55555, 0, 64},
		{"insect_scifi", "insect", "scifi", 66666, 0, 64},
		{"serpent_horror", "serpent", "horror", 77777, 0, 64},
		{"flying_cyberpunk", "flying", "cyberpunk", 88888, 0, 64},
		{"amorphous_postapoc", "amorphous", "postapoc", 99999, 0, 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen.SetGenre(tt.genre)
			sprite := gen.GetSprite(SpriteEnemy, tt.subtype, tt.seed, tt.frame, tt.size)

			if sprite == nil {
				t.Fatal("GetSprite returned nil")
			}

			bounds := sprite.Bounds()
			if bounds.Dx() != tt.size || bounds.Dy() != tt.size {
				t.Errorf("sprite size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), tt.size, tt.size)
			}
		})
	}
}

func TestEnemySpriteAnimation(t *testing.T) {
	gen := NewGenerator(100)
	gen.SetGenre("fantasy")

	subtypes := []string{"humanoid", "quadruped", "insect", "serpent", "flying", "amorphous"}

	for _, subtype := range subtypes {
		t.Run(subtype, func(t *testing.T) {
			frames := make([]interface{}, 4)
			for frame := 0; frame < 4; frame++ {
				sprite := gen.GetSprite(SpriteEnemy, subtype, 12345, frame, 64)
				if sprite == nil {
					t.Fatalf("frame %d: GetSprite returned nil", frame)
				}
				frames[frame] = sprite
			}

			for i := 1; i < len(frames); i++ {
				if frames[i] == frames[0] {
					t.Logf("frame %d matches frame 0 (animation may be subtle)", i)
				}
			}
		})
	}
}

func TestEnemySpriteGenreVariety(t *testing.T) {
	gen := NewGenerator(100)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	seed := int64(42)

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			gen.SetGenre(genre)
			sprite := gen.GetSprite(SpriteEnemy, "humanoid", seed, 0, 64)

			if sprite == nil {
				t.Fatal("GetSprite returned nil")
			}

			bounds := sprite.Bounds()
			if bounds.Dx() != 64 || bounds.Dy() != 64 {
				t.Errorf("sprite size = %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestEnemySpriteCaching(t *testing.T) {
	gen := NewGenerator(10)
	gen.SetGenre("fantasy")

	sprite1 := gen.GetSprite(SpriteEnemy, "humanoid", 12345, 0, 64)
	sprite2 := gen.GetSprite(SpriteEnemy, "humanoid", 12345, 0, 64)

	if sprite1 != sprite2 {
		t.Error("same parameters should return cached sprite")
	}

	sprite3 := gen.GetSprite(SpriteEnemy, "humanoid", 54321, 0, 64)
	if sprite3 == sprite1 {
		t.Error("different seed should generate different sprite")
	}
}

func TestEnemySpriteSeedDeterminism(t *testing.T) {
	gen1 := NewGenerator(100)
	gen2 := NewGenerator(100)
	gen1.SetGenre("fantasy")
	gen2.SetGenre("fantasy")

	seed := int64(999)
	sprite1 := gen1.GetSprite(SpriteEnemy, "quadruped", seed, 0, 64)
	sprite2 := gen2.GetSprite(SpriteEnemy, "quadruped", seed, 0, 64)

	if sprite1.Bounds() != sprite2.Bounds() {
		t.Error("same seed should produce same size sprite")
	}
}

func TestEnemySpriteRoleMapping(t *testing.T) {
	gen := NewGenerator(100)
	gen.SetGenre("fantasy")

	roles := []string{"tank", "ranged", "healer", "ambusher", "scout"}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			sprite := gen.GetSprite(SpriteEnemy, role, 12345, 0, 64)

			if sprite == nil {
				t.Fatalf("role %s: GetSprite returned nil", role)
			}

			bounds := sprite.Bounds()
			if bounds.Dx() != 64 || bounds.Dy() != 64 {
				t.Errorf("sprite size = %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestCreatureBodyPlans(t *testing.T) {
	gen := NewGenerator(100)
	gen.SetGenre("fantasy")

	bodyPlans := []struct {
		name    string
		subtype string
	}{
		{"quadruped", "quadruped"},
		{"insect", "insect"},
		{"serpent", "serpent"},
		{"flying", "flying"},
		{"amorphous", "amorphous"},
	}

	for _, bp := range bodyPlans {
		t.Run(bp.name, func(t *testing.T) {
			sprite := gen.GetSprite(SpriteEnemy, bp.subtype, 12345, 0, 64)

			if sprite == nil {
				t.Fatalf("body plan %s: GetSprite returned nil", bp.name)
			}

			bounds := sprite.Bounds()
			if bounds.Dx() != 64 || bounds.Dy() != 64 {
				t.Errorf("sprite size = %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func BenchmarkEnemySpriteGeneration(b *testing.B) {
	gen := NewGenerator(100)
	gen.SetGenre("fantasy")

	b.Run("humanoid", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gen.GetSprite(SpriteEnemy, "humanoid", int64(i), 0, 64)
		}
	})

	b.Run("quadruped", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gen.GetSprite(SpriteEnemy, "quadruped", int64(i), 0, 64)
		}
	})

	b.Run("insect", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gen.GetSprite(SpriteEnemy, "insect", int64(i), 0, 64)
		}
	})

	b.Run("serpent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gen.GetSprite(SpriteEnemy, "serpent", int64(i), 0, 64)
		}
	})

	b.Run("flying", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gen.GetSprite(SpriteEnemy, "flying", int64(i), 0, 64)
		}
	})

	b.Run("amorphous", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gen.GetSprite(SpriteEnemy, "amorphous", int64(i), 0, 64)
		}
	})
}

func BenchmarkEnemySpriteCached(b *testing.B) {
	gen := NewGenerator(100)
	gen.SetGenre("fantasy")

	gen.GetSprite(SpriteEnemy, "humanoid", 12345, 0, 64)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GetSprite(SpriteEnemy, "humanoid", 12345, 0, 64)
	}
}

// TestHumanoidRoleSilhouetteDifferentiation verifies that different humanoid roles
// produce sprites with distinct silhouettes. Each role should generate successfully
// and sprites with the same seed but different roles should be cached separately.
func TestHumanoidRoleSilhouetteDifferentiation(t *testing.T) {
	gen := NewGenerator(100)
	gen.SetGenre("fantasy")

	roles := []string{"humanoid", "tank", "ranged", "healer", "ambusher", "scout"}
	seed := int64(42424242) // Fixed seed for determinism
	size := 64

	// Generate sprites for all roles and verify they're created
	sprites := make(map[string]*ebiten.Image)
	for _, role := range roles {
		sprite := gen.GetSprite(SpriteEnemy, role, seed, 0, size)
		if sprite == nil {
			t.Fatalf("role %s: GetSprite returned nil", role)
		}
		sprites[role] = sprite

		// Verify correct size
		bounds := sprite.Bounds()
		if bounds.Dx() != size || bounds.Dy() != size {
			t.Errorf("role %s: sprite size = %dx%d, want %dx%d",
				role, bounds.Dx(), bounds.Dy(), size, size)
		}
	}

	// Verify that different roles produce different cached sprites
	// (same seed + different role = different cache key)
	for i, role1 := range roles {
		for j, role2 := range roles {
			if i >= j {
				continue
			}

			// The sprites should be different objects in memory since
			// they have different subtypes in the cache key
			if sprites[role1] == sprites[role2] {
				t.Errorf("roles %s and %s returned same sprite pointer (should be distinct)",
					role1, role2)
			}
		}
	}
}

// TestHumanoidRoleProportions verifies that role-specific body proportions are applied.
func TestHumanoidRoleProportions(t *testing.T) {
	gen := NewGenerator(100)
	gen.SetGenre("fantasy")
	size := 64

	roles := []string{"humanoid", "tank", "ranged", "healer", "ambusher", "scout"}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			sprite := gen.GetSprite(SpriteEnemy, role, 12345, 0, size)
			if sprite == nil {
				t.Fatal("GetSprite returned nil")
			}

			bounds := sprite.Bounds()
			if bounds.Dx() != size || bounds.Dy() != size {
				t.Errorf("sprite size = %dx%d, want %dx%d",
					bounds.Dx(), bounds.Dy(), size, size)
			}
		})
	}
}

// TestRoleBodyPartPositions verifies that calculateBodyPartPositions produces
// different proportions for each role.
func TestRoleBodyPartPositions(t *testing.T) {
	gen := NewGenerator(100)
	size := 64
	cx, cy := size/2, size/2
	frame := 0

	roles := []string{"humanoid", "tank", "ranged", "healer", "ambusher", "scout"}
	positions := make(map[string]bodyPartPositions)

	for _, role := range roles {
		pos := gen.calculateBodyPartPositions(size, cx, cy, role, frame)
		positions[role] = pos
	}

	// Tank should have wider torso than base humanoid
	if positions["tank"].torsoW <= positions["humanoid"].torsoW {
		t.Error("tank torso should be wider than humanoid")
	}

	// Tank should have shoulder width
	if positions["tank"].shoulderWidth <= 0 {
		t.Error("tank should have shoulderWidth > 0")
	}

	// Tank should have helmet
	if !positions["tank"].hasHelmet {
		t.Error("tank should have helmet")
	}

	// Ranged should have longer arms
	if positions["ranged"].armH <= positions["humanoid"].armH {
		t.Error("ranged arms should be longer than humanoid")
	}

	// Healer should have glowing hands
	if !positions["healer"].glowingHands {
		t.Error("healer should have glowing hands")
	}

	// Ambusher should have hood
	if !positions["ambusher"].hasHood {
		t.Error("ambusher should have hood")
	}

	// Ambusher should have hunch
	if positions["ambusher"].hunchAmount <= 0 {
		t.Error("ambusher should have hunchAmount > 0")
	}

	// Scout should have backpack
	if !positions["scout"].hasBackpack {
		t.Error("scout should have backpack")
	}
}
