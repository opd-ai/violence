package ai

import (
	"image"
	"testing"
)

func TestGetBodyPlan(t *testing.T) {
	tests := []struct {
		name     string
		ctype    CreatureType
		expected BodyPlan
	}{
		{"Wolf is quadruped", CreatureWolf, BodyPlanQuadruped},
		{"Bear is quadruped", CreatureBear, BodyPlanQuadruped},
		{"Lion is quadruped", CreatureLion, BodyPlanQuadruped},
		{"Hound is quadruped", CreatureHound, BodyPlanQuadruped},
		{"Raptor is quadruped", CreatureRaptor, BodyPlanQuadruped},
		{"Spider is insect", CreatureSpider, BodyPlanInsect},
		{"Beetle is insect", CreatureBeetle, BodyPlanInsect},
		{"Mantis is insect", CreatureMantis, BodyPlanInsect},
		{"Scorpion is insect", CreatureScorpion, BodyPlanInsect},
		{"Ant is insect", CreatureAnt, BodyPlanInsect},
		{"Snake is serpent", CreatureSnake, BodyPlanSerpent},
		{"Worm is serpent", CreatureWorm, BodyPlanSerpent},
		{"Serpent is serpent", CreatureSerpent, BodyPlanSerpent},
		{"Lamia is serpent", CreatureLamia, BodyPlanSerpent},
		{"Bat is flying", CreatureBat, BodyPlanFlying},
		{"Drake is flying", CreatureDrake, BodyPlanFlying},
		{"Harpy is flying", CreatureHarpy, BodyPlanFlying},
		{"Wasp is flying", CreatureWasp, BodyPlanFlying},
		{"Slime is amorphous", CreatureSlime, BodyPlanAmorphous},
		{"Ooze is amorphous", CreatureOoze, BodyPlanAmorphous},
		{"Elemental is amorphous", CreatureElemental, BodyPlanAmorphous},
		{"Wraith is amorphous", CreatureWraith, BodyPlanAmorphous},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBodyPlan(tt.ctype)
			if result != tt.expected {
				t.Errorf("GetBodyPlan(%s) = %v, want %v", tt.ctype, result, tt.expected)
			}
		})
	}
}

func TestGenerateCreatureSprite_Deterministic(t *testing.T) {
	creatures := []CreatureType{
		CreatureWolf, CreatureBear, CreatureSpider, CreatureMantis,
		CreatureSnake, CreatureLamia, CreatureBat, CreatureDrake,
		CreatureSlime, CreatureElemental,
	}

	for _, ctype := range creatures {
		t.Run(string(ctype), func(t *testing.T) {
			seed := int64(12345)
			frame := AnimFrameIdle

			img1 := GenerateCreatureSprite(seed, ctype, frame)
			img2 := GenerateCreatureSprite(seed, ctype, frame)

			if !imagesEqual(img1, img2) {
				t.Errorf("GenerateCreatureSprite(%s) not deterministic", ctype)
			}
		})
	}
}

func TestGenerateCreatureSprite_DifferentSeeds(t *testing.T) {
	ctype := CreatureWolf
	frame := AnimFrameIdle

	img1 := GenerateCreatureSprite(100, ctype, frame)
	img2 := GenerateCreatureSprite(200, ctype, frame)

	// Different seeds should produce variations (color, size, etc.)
	// But both should be valid 64x64 images
	if img1.Bounds() != img2.Bounds() {
		t.Error("Images with different seeds have different bounds")
	}

	bounds := image.Rect(0, 0, 64, 64)
	if img1.Bounds() != bounds {
		t.Errorf("Expected bounds %v, got %v", bounds, img1.Bounds())
	}
}

func TestGenerateCreatureSprite_AllFrames(t *testing.T) {
	frames := []AnimFrame{
		AnimFrameIdle,
		AnimFrameWalk1,
		AnimFrameWalk2,
		AnimFrameAttack,
		AnimFrameDeath,
	}

	creatures := []CreatureType{
		CreatureWolf, CreatureSpider, CreatureSnake,
		CreatureBat, CreatureSlime,
	}

	for _, ctype := range creatures {
		for _, frame := range frames {
			t.Run(string(ctype)+"_"+frameToString(frame), func(t *testing.T) {
				img := GenerateCreatureSprite(12345, ctype, frame)

				if img == nil {
					t.Fatal("Generated image is nil")
				}

				expected := image.Rect(0, 0, 64, 64)
				if img.Bounds() != expected {
					t.Errorf("Expected bounds %v, got %v", expected, img.Bounds())
				}

				// Check that image has some non-zero pixels
				hasPixels := false
				for y := 0; y < 64; y++ {
					for x := 0; x < 64; x++ {
						r, g, b, a := img.At(x, y).RGBA()
						if a > 0 {
							hasPixels = true
							break
						}
						_, _, _ = r, g, b
					}
					if hasPixels {
						break
					}
				}

				if !hasPixels {
					t.Error("Generated sprite has no visible pixels")
				}
			})
		}
	}
}

func TestGenerateCreatureSprite_QuadrupedVariety(t *testing.T) {
	quadrupeds := []CreatureType{
		CreatureWolf, CreatureBear, CreatureLion, CreatureHound, CreatureRaptor,
	}

	// All quadrupeds should generate valid sprites
	for _, ctype := range quadrupeds {
		t.Run(string(ctype), func(t *testing.T) {
			img := GenerateCreatureSprite(54321, ctype, AnimFrameIdle)

			if img == nil {
				t.Fatal("Generated image is nil")
			}

			if img.Bounds() != image.Rect(0, 0, 64, 64) {
				t.Error("Invalid image bounds")
			}
		})
	}
}

func TestGenerateCreatureSprite_InsectVariety(t *testing.T) {
	insects := []CreatureType{
		CreatureSpider, CreatureBeetle, CreatureMantis, CreatureScorpion, CreatureAnt,
	}

	for _, ctype := range insects {
		t.Run(string(ctype), func(t *testing.T) {
			img := GenerateCreatureSprite(99999, ctype, AnimFrameIdle)

			if img == nil {
				t.Fatal("Generated image is nil")
			}

			if img.Bounds() != image.Rect(0, 0, 64, 64) {
				t.Error("Invalid image bounds")
			}
		})
	}
}

func TestGenerateCreatureSprite_SerpentVariety(t *testing.T) {
	serpents := []CreatureType{
		CreatureSnake, CreatureWorm, CreatureSerpent, CreatureLamia,
	}

	for _, ctype := range serpents {
		t.Run(string(ctype), func(t *testing.T) {
			img := GenerateCreatureSprite(11111, ctype, AnimFrameIdle)

			if img == nil {
				t.Fatal("Generated image is nil")
			}

			if img.Bounds() != image.Rect(0, 0, 64, 64) {
				t.Error("Invalid image bounds")
			}
		})
	}
}

func TestGenerateCreatureSprite_FlyingVariety(t *testing.T) {
	flying := []CreatureType{
		CreatureBat, CreatureDrake, CreatureHarpy, CreatureWasp,
	}

	for _, ctype := range flying {
		t.Run(string(ctype), func(t *testing.T) {
			img := GenerateCreatureSprite(77777, ctype, AnimFrameIdle)

			if img == nil {
				t.Fatal("Generated image is nil")
			}

			if img.Bounds() != image.Rect(0, 0, 64, 64) {
				t.Error("Invalid image bounds")
			}
		})
	}
}

func TestGenerateCreatureSprite_AmorphousVariety(t *testing.T) {
	amorphous := []CreatureType{
		CreatureSlime, CreatureOoze, CreatureElemental, CreatureWraith,
	}

	for _, ctype := range amorphous {
		t.Run(string(ctype), func(t *testing.T) {
			img := GenerateCreatureSprite(33333, ctype, AnimFrameIdle)

			if img == nil {
				t.Fatal("Generated image is nil")
			}

			if img.Bounds() != image.Rect(0, 0, 64, 64) {
				t.Error("Invalid image bounds")
			}
		})
	}
}

func TestGenerateCreatureSprite_AnimationFrames(t *testing.T) {
	// Test that different frames produce different images (for animated creatures)
	ctype := CreatureWolf
	seed := int64(42)

	idle := GenerateCreatureSprite(seed, ctype, AnimFrameIdle)
	walk1 := GenerateCreatureSprite(seed, ctype, AnimFrameWalk1)
	walk2 := GenerateCreatureSprite(seed, ctype, AnimFrameWalk2)
	attack := GenerateCreatureSprite(seed, ctype, AnimFrameAttack)

	// Walk frames should differ from idle
	if imagesEqual(idle, walk1) {
		t.Error("Walk1 frame should differ from Idle")
	}

	if imagesEqual(idle, walk2) {
		t.Error("Walk2 frame should differ from Idle")
	}

	// Attack should differ from idle
	if imagesEqual(idle, attack) {
		t.Error("Attack frame should differ from Idle")
	}
}

// Helper functions

func imagesEqual(img1, img2 *image.RGBA) bool {
	if img1.Bounds() != img2.Bounds() {
		return false
	}

	for y := img1.Bounds().Min.Y; y < img1.Bounds().Max.Y; y++ {
		for x := img1.Bounds().Min.X; x < img1.Bounds().Max.X; x++ {
			if img1.At(x, y) != img2.At(x, y) {
				return false
			}
		}
	}

	return true
}

func frameToString(frame AnimFrame) string {
	switch frame {
	case AnimFrameIdle:
		return "idle"
	case AnimFrameWalk1:
		return "walk1"
	case AnimFrameWalk2:
		return "walk2"
	case AnimFrameAttack:
		return "attack"
	case AnimFrameDeath:
		return "death"
	default:
		return "unknown"
	}
}
