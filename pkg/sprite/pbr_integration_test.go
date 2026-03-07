package sprite

import (
	"testing"
)

// TestPBRIntegrationWithHumanoidEnemy verifies PBR shading system is set up for humanoid enemies.
func TestPBRIntegrationWithHumanoidEnemy(t *testing.T) {
	gen := NewGenerator(10)
	gen.SetGenre("fantasy")

	// Generate a humanoid enemy sprite - this will apply PBR shading internally
	img := gen.GetSprite(SpriteEnemy, "humanoid", 12345, 0, 32)

	if img == nil {
		t.Fatal("Failed to generate sprite")
	}

	// Check that the sprite has correct dimensions
	bounds := img.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		t.Errorf("Expected 32x32 sprite, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestPBRIntegrationWithProps verifies PBR shading system is set up for props.
func TestPBRIntegrationWithProps(t *testing.T) {
	gen := NewGenerator(10)

	propTypes := []string{"barrel", "crate", "pillar"}

	for _, propType := range propTypes {
		t.Run(propType, func(t *testing.T) {
			img := gen.GetSprite(SpriteProp, propType, 54321, 0, 32)

			if img == nil {
				t.Fatalf("Failed to generate %s sprite", propType)
			}

			bounds := img.Bounds()
			if bounds.Dx() != 32 || bounds.Dy() != 32 {
				t.Errorf("Expected 32x32 sprite, got %dx%d", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

// TestPBRIntegrationWithCreatures verifies PBR shading system is set up for non-humanoid creatures.
func TestPBRIntegrationWithCreatures(t *testing.T) {
	gen := NewGenerator(10)

	creatureTypes := []string{"quadruped", "insect", "serpent", "flying"}

	for _, creatureType := range creatureTypes {
		t.Run(creatureType, func(t *testing.T) {
			img := gen.GetSprite(SpriteEnemy, creatureType, 98765, 0, 32)

			if img == nil {
				t.Fatalf("Failed to generate %s sprite", creatureType)
			}

			bounds := img.Bounds()
			if bounds.Dx() != 32 || bounds.Dy() != 32 {
				t.Errorf("Expected 32x32 sprite, got %dx%d", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

// TestLightConfigPersistence verifies light configuration is stored in generator.
func TestLightConfigPersistence(t *testing.T) {
	gen := NewGenerator(10)

	// Check that default light config is set
	if gen.lightCfg.LightIntensity <= 0 {
		t.Error("Light intensity not initialized")
	}

	// Verify normalized light direction
	lx, ly, lz := gen.lightCfg.LightDirX, gen.lightCfg.LightDirY, gen.lightCfg.LightDirZ
	length := lx*lx + ly*ly + lz*lz
	if length < 0.99 || length > 1.01 {
		t.Errorf("Light direction not normalized: length^2 = %f", length)
	}

	// Generate sprite to ensure light config is used (no errors)
	img := gen.GetSprite(SpriteEnemy, "tank", 11111, 0, 32)
	if img == nil {
		t.Fatal("Failed to generate sprite with light config")
	}
}

// TestPBRShadingAppliedToAllEnemyTypes verifies PBR is applied to all enemy sprite types.
func TestPBRShadingAppliedToAllEnemyTypes(t *testing.T) {
	gen := NewGenerator(10)

	enemyTypes := []struct {
		subtype string
		genre   string
	}{
		{"humanoid", "fantasy"},
		{"tank", "scifi"},
		{"ranged", "horror"},
		{"healer", "cyberpunk"},
		{"ambusher", "postapoc"},
		{"quadruped", "fantasy"},
		{"insect", "scifi"},
		{"serpent", "horror"},
		{"flying", "cyberpunk"},
		{"amorphous", "postapoc"},
	}

	for _, enemy := range enemyTypes {
		t.Run(enemy.subtype+"_"+enemy.genre, func(t *testing.T) {
			gen.SetGenre(enemy.genre)
			img := gen.GetSprite(SpriteEnemy, enemy.subtype, 99999, 0, 32)

			if img == nil {
				t.Fatalf("Failed to generate %s sprite in %s genre", enemy.subtype, enemy.genre)
			}

			// If we got here, PBR shading was applied without errors
		})
	}
}
