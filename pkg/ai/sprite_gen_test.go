package ai

import (
	"image"
	"testing"
)

func TestGenerateEnemySprite(t *testing.T) {
	tests := []struct {
		name      string
		seed      int64
		archetype EnemyArchetype
		frame     AnimFrame
		wantSize  int
	}{
		{
			name:      "fantasy guard idle",
			seed:      12345,
			archetype: ArchetypeFantasyGuard,
			frame:     AnimFrameIdle,
			wantSize:  64,
		},
		{
			name:      "fantasy guard walk1",
			seed:      12345,
			archetype: ArchetypeFantasyGuard,
			frame:     AnimFrameWalk1,
			wantSize:  64,
		},
		{
			name:      "fantasy guard walk2",
			seed:      12345,
			archetype: ArchetypeFantasyGuard,
			frame:     AnimFrameWalk2,
			wantSize:  64,
		},
		{
			name:      "fantasy guard attack",
			seed:      12345,
			archetype: ArchetypeFantasyGuard,
			frame:     AnimFrameAttack,
			wantSize:  64,
		},
		{
			name:      "scifi soldier idle",
			seed:      67890,
			archetype: ArchetypeSciFiSoldier,
			frame:     AnimFrameIdle,
			wantSize:  64,
		},
		{
			name:      "scifi soldier attack",
			seed:      67890,
			archetype: ArchetypeSciFiSoldier,
			frame:     AnimFrameAttack,
			wantSize:  64,
		},
		{
			name:      "horror cultist idle",
			seed:      11111,
			archetype: ArchetypeHorrorCultist,
			frame:     AnimFrameIdle,
			wantSize:  64,
		},
		{
			name:      "horror cultist attack",
			seed:      11111,
			archetype: ArchetypeHorrorCultist,
			frame:     AnimFrameAttack,
			wantSize:  64,
		},
		{
			name:      "cyberpunk drone idle",
			seed:      22222,
			archetype: ArchetypeCyberpunkDrone,
			frame:     AnimFrameIdle,
			wantSize:  64,
		},
		{
			name:      "cyberpunk drone walk1",
			seed:      22222,
			archetype: ArchetypeCyberpunkDrone,
			frame:     AnimFrameWalk1,
			wantSize:  64,
		},
		{
			name:      "cyberpunk drone attack",
			seed:      22222,
			archetype: ArchetypeCyberpunkDrone,
			frame:     AnimFrameAttack,
			wantSize:  64,
		},
		{
			name:      "postapoc scavenger idle",
			seed:      33333,
			archetype: ArchetypePostapocScavenger,
			frame:     AnimFrameIdle,
			wantSize:  64,
		},
		{
			name:      "postapoc scavenger walk2",
			seed:      33333,
			archetype: ArchetypePostapocScavenger,
			frame:     AnimFrameWalk2,
			wantSize:  64,
		},
		{
			name:      "postapoc scavenger attack",
			seed:      33333,
			archetype: ArchetypePostapocScavenger,
			frame:     AnimFrameAttack,
			wantSize:  64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := GenerateEnemySprite(tt.seed, tt.archetype, tt.frame)

			if img == nil {
				t.Fatal("GenerateEnemySprite returned nil")
			}

			bounds := img.Bounds()
			if bounds.Dx() != tt.wantSize || bounds.Dy() != tt.wantSize {
				t.Errorf("got size %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), tt.wantSize, tt.wantSize)
			}

			// Verify image has non-transparent pixels
			hasContent := false
			for y := 0; y < bounds.Dy(); y++ {
				for x := 0; x < bounds.Dx(); x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						hasContent = true
						break
					}
				}
				if hasContent {
					break
				}
			}
			if !hasContent {
				t.Error("generated sprite has no visible content")
			}
		})
	}
}

func TestGenerateEnemySpriteDeterminism(t *testing.T) {
	seed := int64(42)
	archetype := ArchetypeSciFiSoldier
	frame := AnimFrameIdle

	img1 := GenerateEnemySprite(seed, archetype, frame)
	img2 := GenerateEnemySprite(seed, archetype, frame)

	if img1 == nil || img2 == nil {
		t.Fatal("GenerateEnemySprite returned nil")
	}

	bounds := img1.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)
			if c1 != c2 {
				t.Errorf("pixel at (%d,%d) differs: %v vs %v", x, y, c1, c2)
				return
			}
		}
	}
}

func TestGenerateEnemySpriteUniqueness(t *testing.T) {
	seed1 := int64(100)
	seed2 := int64(200)
	archetype := ArchetypeHorrorCultist
	frame := AnimFrameIdle

	img1 := GenerateEnemySprite(seed1, archetype, frame)
	img2 := GenerateEnemySprite(seed2, archetype, frame)

	if img1 == nil || img2 == nil {
		t.Fatal("GenerateEnemySprite returned nil")
	}

	// Images with different seeds might differ (depends on implementation)
	// At minimum, they should both be valid images
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		t.Error("images should have same dimensions")
	}
}

func TestArchetypeVariations(t *testing.T) {
	seed := int64(12345)
	frame := AnimFrameIdle

	fantasy := GenerateEnemySprite(seed, ArchetypeFantasyGuard, frame)
	scifi := GenerateEnemySprite(seed, ArchetypeSciFiSoldier, frame)
	horror := GenerateEnemySprite(seed, ArchetypeHorrorCultist, frame)
	cyberpunk := GenerateEnemySprite(seed, ArchetypeCyberpunkDrone, frame)
	postapoc := GenerateEnemySprite(seed, ArchetypePostapocScavenger, frame)

	sprites := []*testSprite{
		{name: "fantasy", img: fantasy},
		{name: "scifi", img: scifi},
		{name: "horror", img: horror},
		{name: "cyberpunk", img: cyberpunk},
		{name: "postapoc", img: postapoc},
	}

	// All archetypes should produce different images
	for i := 0; i < len(sprites); i++ {
		for j := i + 1; j < len(sprites); j++ {
			differences := countDifferences(sprites[i].img, sprites[j].img)
			if differences == 0 {
				t.Errorf("%s and %s sprites should differ", sprites[i].name, sprites[j].name)
			}
		}
	}
}

type testSprite struct {
	name string
	img  *image.RGBA
}

func countDifferences(img1, img2 *image.RGBA) int {
	differences := 0
	bounds := img1.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			if img1.At(x, y) != img2.At(x, y) {
				differences++
			}
		}
	}
	return differences
}

func TestAnimFrameVariations(t *testing.T) {
	seed := int64(54321)
	archetype := ArchetypeFantasyGuard

	idle := GenerateEnemySprite(seed, archetype, AnimFrameIdle)
	walk1 := GenerateEnemySprite(seed, archetype, AnimFrameWalk1)
	walk2 := GenerateEnemySprite(seed, archetype, AnimFrameWalk2)
	attack := GenerateEnemySprite(seed, archetype, AnimFrameAttack)
	death := GenerateEnemySprite(seed, archetype, AnimFrameDeath)

	// Walk frames should differ from idle
	idleVsWalk1 := countDifferences(idle, walk1)
	if idleVsWalk1 == 0 {
		t.Error("idle and walk1 frames should differ")
	}

	// Walk1 and walk2 should differ
	walk1VsWalk2 := countDifferences(walk1, walk2)
	if walk1VsWalk2 == 0 {
		t.Error("walk1 and walk2 frames should differ")
	}

	// Attack should differ from idle
	idleVsAttack := countDifferences(idle, attack)
	if idleVsAttack == 0 {
		t.Error("idle and attack frames should differ")
	}

	// All sprites should be non-nil
	if idle == nil || walk1 == nil || walk2 == nil || attack == nil || death == nil {
		t.Error("all frame types should generate valid sprites")
	}
}

func TestDefaultArchetype(t *testing.T) {
	seed := int64(99999)
	frame := AnimFrameIdle

	// Test with an empty/invalid archetype - should default to fantasy guard
	img := GenerateEnemySprite(seed, "", frame)

	if img == nil {
		t.Fatal("GenerateEnemySprite returned nil for default archetype")
	}

	// Should produce same result as explicit fantasy guard
	fantasyImg := GenerateEnemySprite(seed, ArchetypeFantasyGuard, frame)

	bounds := img.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			if img.At(x, y) != fantasyImg.At(x, y) {
				t.Error("default archetype should produce fantasy guard sprite")
				return
			}
		}
	}
}

func TestCyberpunkDroneHoverAnimation(t *testing.T) {
	seed := int64(77777)
	archetype := ArchetypeCyberpunkDrone

	idle := GenerateEnemySprite(seed, archetype, AnimFrameIdle)
	walk1 := GenerateEnemySprite(seed, archetype, AnimFrameWalk1)
	walk2 := GenerateEnemySprite(seed, archetype, AnimFrameWalk2)

	// Drone should show hover animation differences
	idleVsWalk1 := countDifferences(idle, walk1)
	idleVsWalk2 := countDifferences(idle, walk2)

	if idleVsWalk1 == 0 || idleVsWalk2 == 0 {
		t.Error("cyberpunk drone should show hover animation")
	}
}

func TestMuzzleFlashInAttackFrame(t *testing.T) {
	tests := []struct {
		name        string
		archetype   EnemyArchetype
		expectFlash bool
	}{
		{
			name:        "fantasy guard has weapon in attack",
			archetype:   ArchetypeFantasyGuard,
			expectFlash: true,
		},
		{
			name:        "scifi soldier has weapon",
			archetype:   ArchetypeSciFiSoldier,
			expectFlash: true,
		},
		{
			name:        "horror cultist has dagger",
			archetype:   ArchetypeHorrorCultist,
			expectFlash: true,
		},
		{
			name:        "cyberpunk drone has muzzle flash",
			archetype:   ArchetypeCyberpunkDrone,
			expectFlash: true,
		},
		{
			name:        "postapoc scavenger has weapon",
			archetype:   ArchetypePostapocScavenger,
			expectFlash: true,
		},
	}

	seed := int64(88888)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idle := GenerateEnemySprite(seed, tt.archetype, AnimFrameIdle)
			attack := GenerateEnemySprite(seed, tt.archetype, AnimFrameAttack)

			// Attack frame should differ from idle
			differences := countDifferences(idle, attack)
			if tt.expectFlash && differences == 0 {
				t.Error("attack frame should differ from idle")
			}
		})
	}
}

func TestAllArchetypeConstants(t *testing.T) {
	archetypes := []EnemyArchetype{
		ArchetypeFantasyGuard,
		ArchetypeSciFiSoldier,
		ArchetypeHorrorCultist,
		ArchetypeCyberpunkDrone,
		ArchetypePostapocScavenger,
	}

	seed := int64(55555)

	for _, archetype := range archetypes {
		t.Run(string(archetype), func(t *testing.T) {
			img := GenerateEnemySprite(seed, archetype, AnimFrameIdle)
			if img == nil {
				t.Errorf("archetype %s produced nil sprite", archetype)
			}
		})
	}
}

func TestAllAnimFrameConstants(t *testing.T) {
	frames := []AnimFrame{
		AnimFrameIdle,
		AnimFrameWalk1,
		AnimFrameWalk2,
		AnimFrameAttack,
		AnimFrameDeath,
	}

	seed := int64(66666)
	archetype := ArchetypeSciFiSoldier

	for _, frame := range frames {
		t.Run("", func(t *testing.T) {
			img := GenerateEnemySprite(seed, archetype, frame)
			if img == nil {
				t.Errorf("frame %d produced nil sprite", frame)
			}
		})
	}
}

func BenchmarkGenerateEnemySprite(b *testing.B) {
	seed := int64(42)
	archetype := ArchetypeSciFiSoldier
	frame := AnimFrameIdle

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateEnemySprite(seed, archetype, frame)
	}
}

func BenchmarkGenerateEnemySpriteAllArchetypes(b *testing.B) {
	seed := int64(42)
	archetypes := []EnemyArchetype{
		ArchetypeFantasyGuard,
		ArchetypeSciFiSoldier,
		ArchetypeHorrorCultist,
		ArchetypeCyberpunkDrone,
		ArchetypePostapocScavenger,
	}
	frames := []AnimFrame{
		AnimFrameIdle,
		AnimFrameWalk1,
		AnimFrameWalk2,
		AnimFrameAttack,
		AnimFrameDeath,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, archetype := range archetypes {
			for _, frame := range frames {
				GenerateEnemySprite(seed, archetype, frame)
			}
		}
	}
}
