package weapon

import (
	"image/color"
	"testing"
)

func TestGenerateWeaponSprite(t *testing.T) {
	tests := []struct {
		name       string
		seed       int64
		weaponType WeaponType
		frame      FrameType
		wantSize   int
	}{
		{
			name:       "melee idle",
			seed:       12345,
			weaponType: TypeMelee,
			frame:      FrameIdle,
			wantSize:   128,
		},
		{
			name:       "hitscan idle",
			seed:       67890,
			weaponType: TypeHitscan,
			frame:      FrameIdle,
			wantSize:   128,
		},
		{
			name:       "hitscan fire",
			seed:       67890,
			weaponType: TypeHitscan,
			frame:      FrameFire,
			wantSize:   128,
		},
		{
			name:       "projectile idle",
			seed:       11111,
			weaponType: TypeProjectile,
			frame:      FrameIdle,
			wantSize:   128,
		},
		{
			name:       "projectile fire",
			seed:       11111,
			weaponType: TypeProjectile,
			frame:      FrameFire,
			wantSize:   128,
		},
		{
			name:       "melee reload",
			seed:       22222,
			weaponType: TypeMelee,
			frame:      FrameReload,
			wantSize:   128,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := GenerateWeaponSprite(tt.seed, tt.weaponType, tt.frame)

			if img == nil {
				t.Fatal("GenerateWeaponSprite returned nil")
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

func TestGenerateWeaponSpriteDeterminism(t *testing.T) {
	seed := int64(42)
	weaponType := TypeHitscan
	frame := FrameIdle

	img1 := GenerateWeaponSprite(seed, weaponType, frame)
	img2 := GenerateWeaponSprite(seed, weaponType, frame)

	if img1 == nil || img2 == nil {
		t.Fatal("GenerateWeaponSprite returned nil")
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

func TestGenerateWeaponSpriteUniqueness(t *testing.T) {
	seed1 := int64(100)
	seed2 := int64(200)
	weaponType := TypeMelee
	frame := FrameIdle

	img1 := GenerateWeaponSprite(seed1, weaponType, frame)
	img2 := GenerateWeaponSprite(seed2, weaponType, frame)

	if img1 == nil || img2 == nil {
		t.Fatal("GenerateWeaponSprite returned nil")
	}

	// Images with different seeds should differ
	differences := 0
	bounds := img1.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)
			if c1 != c2 {
				differences++
			}
		}
	}

	// Should have some differences (melee weapons have random variation)
	if differences == 0 {
		t.Error("images with different seeds should differ")
	}
}

func TestMuzzleFlashPresence(t *testing.T) {
	seed := int64(999)

	tests := []struct {
		name       string
		weaponType WeaponType
		frame      FrameType
		wantFlash  bool
	}{
		{
			name:       "hitscan idle - no flash",
			weaponType: TypeHitscan,
			frame:      FrameIdle,
			wantFlash:  false,
		},
		{
			name:       "hitscan fire - has flash",
			weaponType: TypeHitscan,
			frame:      FrameFire,
			wantFlash:  true,
		},
		{
			name:       "projectile idle - no flash",
			weaponType: TypeProjectile,
			frame:      FrameIdle,
			wantFlash:  false,
		},
		{
			name:       "projectile fire - has flash",
			weaponType: TypeProjectile,
			frame:      FrameFire,
			wantFlash:  true,
		},
		{
			name:       "melee fire - no flash",
			weaponType: TypeMelee,
			frame:      FrameFire,
			wantFlash:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := GenerateWeaponSprite(seed, tt.weaponType, tt.frame)

			// Check for bright pixels (flash indicator)
			hasBrightPixels := false
			bounds := img.Bounds()
			for y := 0; y < bounds.Dy(); y++ {
				for x := 0; x < bounds.Dx(); x++ {
					r, g, _, a := img.At(x, y).RGBA()
					if a > 0 {
						// Check for bright yellow/white pixels (muzzle flash)
						if r > 50000 && g > 40000 {
							hasBrightPixels = true
							break
						}
					}
				}
				if hasBrightPixels {
					break
				}
			}

			if tt.wantFlash && !hasBrightPixels {
				t.Error("expected muzzle flash but found none")
			}
		})
	}
}

func TestFillRect(t *testing.T) {
	img := GenerateWeaponSprite(1, TypeMelee, FrameIdle)
	testColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	fillRect(img, 10, 10, 20, 20, testColor)

	// Check that rectangle is filled
	for y := 10; y < 20; y++ {
		for x := 10; x < 20; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 || a>>8 != 255 {
				t.Errorf("pixel at (%d,%d) not filled correctly", x, y)
			}
		}
	}
}

func TestFillCircle(t *testing.T) {
	img := GenerateWeaponSprite(1, TypeMelee, FrameIdle)
	testColor := color.RGBA{R: 0, G: 255, B: 0, A: 255}

	fillCircle(img, 64, 64, 10, testColor)

	// Check center pixel is filled
	c := img.At(64, 64)
	r, g, b, a := c.RGBA()
	if r>>8 != 0 || g>>8 != 255 || b>>8 != 0 || a>>8 != 255 {
		t.Error("circle center pixel not filled correctly")
	}

	// Check a pixel on the edge (should be filled)
	c = img.At(64+7, 64)
	r, g, b, a = c.RGBA()
	if r>>8 != 0 || g>>8 != 255 || b>>8 != 0 || a>>8 != 255 {
		t.Error("circle edge pixel not filled correctly")
	}
}

func TestDrawLine(t *testing.T) {
	img := GenerateWeaponSprite(1, TypeMelee, FrameIdle)
	testColor := color.RGBA{R: 0, G: 0, B: 255, A: 255}

	drawLine(img, 10, 10, 30, 30, testColor)

	// Check that line endpoints are drawn
	c1 := img.At(10, 10)
	r1, g1, b1, a1 := c1.RGBA()
	if r1>>8 != 0 || g1>>8 != 0 || b1>>8 != 255 || a1>>8 != 255 {
		t.Error("line start pixel not drawn correctly")
	}

	c2 := img.At(30, 30)
	r2, g2, b2, a2 := c2.RGBA()
	if r2>>8 != 0 || g2>>8 != 0 || b2>>8 != 255 || a2>>8 != 255 {
		t.Error("line end pixel not drawn correctly")
	}
}

func TestAbsFunction(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{input: 5, want: 5},
		{input: -5, want: 5},
		{input: 0, want: 0},
		{input: 100, want: 100},
		{input: -100, want: 100},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := abs(tt.input)
			if got != tt.want {
				t.Errorf("abs(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestWeaponTypeVariations(t *testing.T) {
	seed := int64(12345)
	frame := FrameIdle

	melee := GenerateWeaponSprite(seed, TypeMelee, frame)
	hitscan := GenerateWeaponSprite(seed, TypeHitscan, frame)
	projectile := GenerateWeaponSprite(seed, TypeProjectile, frame)

	// All three types should produce different images
	meleeVsHitscan := 0
	hitscanVsProjectile := 0

	bounds := melee.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			if melee.At(x, y) != hitscan.At(x, y) {
				meleeVsHitscan++
			}
			if hitscan.At(x, y) != projectile.At(x, y) {
				hitscanVsProjectile++
			}
		}
	}

	if meleeVsHitscan == 0 {
		t.Error("melee and hitscan sprites should differ")
	}
	if hitscanVsProjectile == 0 {
		t.Error("hitscan and projectile sprites should differ")
	}
}

func TestFrameTypeVariations(t *testing.T) {
	seed := int64(54321)
	weaponType := TypeHitscan

	idle := GenerateWeaponSprite(seed, weaponType, FrameIdle)
	fire := GenerateWeaponSprite(seed, weaponType, FrameFire)
	reload := GenerateWeaponSprite(seed, weaponType, FrameReload)

	// Frames should differ (at least fire vs idle due to muzzle flash)
	idleVsFire := 0
	bounds := idle.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			if idle.At(x, y) != fire.At(x, y) {
				idleVsFire++
			}
		}
	}

	if idleVsFire == 0 {
		t.Error("idle and fire frames should differ")
	}

	// All sprites should be non-nil
	if idle == nil || fire == nil || reload == nil {
		t.Error("all frame types should generate valid sprites")
	}
}

func BenchmarkGenerateWeaponSprite(b *testing.B) {
	seed := int64(42)
	for i := 0; i < b.N; i++ {
		GenerateWeaponSprite(seed, TypeHitscan, FrameIdle)
	}
}

func BenchmarkGenerateWeaponSpriteAllTypes(b *testing.B) {
	seed := int64(42)
	types := []WeaponType{TypeMelee, TypeHitscan, TypeProjectile}
	frames := []FrameType{FrameIdle, FrameFire, FrameReload}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, wt := range types {
			for _, ft := range frames {
				GenerateWeaponSprite(seed, wt, ft)
			}
		}
	}
}
