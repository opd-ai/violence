package itemicon

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestIconGeneration(t *testing.T) {
	sys := NewSystem("fantasy", 100)

	tests := []struct {
		name string
		comp ItemIconComponent
	}{
		{
			name: "common sword",
			comp: ItemIconComponent{
				Seed:       12345,
				IconType:   "weapon",
				Rarity:     0,
				SubType:    "sword",
				IconSize:   48,
				BorderGlow: true,
				Durability: 1.0,
			},
		},
		{
			name: "legendary axe",
			comp: ItemIconComponent{
				Seed:         54321,
				IconType:     "weapon",
				Rarity:       4,
				SubType:      "axe",
				IconSize:     48,
				BorderGlow:   true,
				EnchantLevel: 3,
				Durability:   1.0,
			},
		},
		{
			name: "epic armor",
			comp: ItemIconComponent{
				Seed:       99999,
				IconType:   "armor",
				Rarity:     3,
				IconSize:   48,
				BorderGlow: true,
				Durability: 0.7,
			},
		},
		{
			name: "rare potion",
			comp: ItemIconComponent{
				Seed:       11111,
				IconType:   "consumable",
				Rarity:     2,
				SubType:    "potion",
				IconSize:   48,
				BorderGlow: true,
				Durability: 1.0,
			},
		},
		{
			name: "uncommon scroll",
			comp: ItemIconComponent{
				Seed:       22222,
				IconType:   "consumable",
				Rarity:     1,
				SubType:    "scroll",
				IconSize:   48,
				BorderGlow: true,
				Durability: 1.0,
			},
		},
		{
			name: "rare material",
			comp: ItemIconComponent{
				Seed:       33333,
				IconType:   "material",
				Rarity:     2,
				IconSize:   48,
				BorderGlow: true,
				Durability: 1.0,
			},
		},
		{
			name: "quest item",
			comp: ItemIconComponent{
				Seed:       44444,
				IconType:   "quest",
				Rarity:     3,
				IconSize:   48,
				BorderGlow: true,
				Durability: 1.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := sys.GenerateIcon(&tt.comp)
			if icon == nil {
				t.Errorf("GenerateIcon() returned nil")
				return
			}

			bounds := icon.Bounds()
			if bounds.Dx() != tt.comp.IconSize || bounds.Dy() != tt.comp.IconSize {
				t.Errorf("Icon size = %dx%d, want %dx%d",
					bounds.Dx(), bounds.Dy(), tt.comp.IconSize, tt.comp.IconSize)
			}
		})
	}
}

func TestCaching(t *testing.T) {
	sys := NewSystem("fantasy", 10)

	comp := ItemIconComponent{
		Seed:       12345,
		IconType:   "weapon",
		Rarity:     2,
		SubType:    "sword",
		IconSize:   48,
		BorderGlow: true,
		Durability: 1.0,
	}

	icon1 := sys.GenerateIcon(&comp)
	icon2 := sys.GenerateIcon(&comp)

	if icon1 != icon2 {
		t.Error("Expected same icon from cache")
	}
}

func TestGenreChange(t *testing.T) {
	sys := NewSystem("fantasy", 100)

	comp := ItemIconComponent{
		Seed:       12345,
		IconType:   "weapon",
		Rarity:     2,
		SubType:    "sword",
		IconSize:   48,
		BorderGlow: true,
		Durability: 1.0,
	}

	icon1 := sys.GenerateIcon(&comp)

	sys.SetGenre("scifi")

	icon2 := sys.GenerateIcon(&comp)

	if icon1 == icon2 {
		t.Error("Expected different icon after genre change")
	}
}

func TestRarityColors(t *testing.T) {
	sys := NewSystem("fantasy", 100)

	rarities := []int{0, 1, 2, 3, 4}

	for _, rarity := range rarities {
		comp := ItemIconComponent{
			Seed:       int64(rarity * 1000),
			IconType:   "weapon",
			Rarity:     rarity,
			SubType:    "sword",
			IconSize:   48,
			BorderGlow: true,
			Durability: 1.0,
		}

		icon := sys.GenerateIcon(&comp)
		if icon == nil {
			t.Errorf("Failed to generate icon for rarity %d", rarity)
		}
	}
}

func TestEnchantmentLevels(t *testing.T) {
	sys := NewSystem("fantasy", 100)

	for enchantLevel := 0; enchantLevel <= 5; enchantLevel++ {
		comp := ItemIconComponent{
			Seed:         int64(enchantLevel * 1000),
			IconType:     "weapon",
			Rarity:       3,
			SubType:      "sword",
			IconSize:     48,
			BorderGlow:   true,
			EnchantLevel: enchantLevel,
			Durability:   1.0,
		}

		icon := sys.GenerateIcon(&comp)
		if icon == nil {
			t.Errorf("Failed to generate icon for enchant level %d", enchantLevel)
		}
	}
}

func TestDurabilityWear(t *testing.T) {
	sys := NewSystem("fantasy", 100)

	durabilities := []float64{1.0, 0.75, 0.5, 0.25, 0.1}

	for _, durability := range durabilities {
		comp := ItemIconComponent{
			Seed:       12345,
			IconType:   "weapon",
			Rarity:     2,
			SubType:    "sword",
			IconSize:   48,
			BorderGlow: true,
			Durability: durability,
		}

		icon := sys.GenerateIcon(&comp)
		if icon == nil {
			t.Errorf("Failed to generate icon for durability %.2f", durability)
		}
	}
}

func TestAllIconTypes(t *testing.T) {
	sys := NewSystem("fantasy", 100)

	iconTypes := []struct {
		iconType string
		subType  string
	}{
		{"weapon", "sword"},
		{"weapon", "axe"},
		{"weapon", ""},
		{"armor", ""},
		{"consumable", "potion"},
		{"consumable", "scroll"},
		{"consumable", ""},
		{"material", ""},
		{"quest", ""},
		{"unknown", ""},
	}

	for _, tt := range iconTypes {
		comp := ItemIconComponent{
			Seed:       12345,
			IconType:   tt.iconType,
			Rarity:     2,
			SubType:    tt.subType,
			IconSize:   48,
			BorderGlow: true,
			Durability: 1.0,
		}

		icon := sys.GenerateIcon(&comp)
		if icon == nil {
			t.Errorf("Failed to generate icon for type=%s subtype=%s",
				tt.iconType, tt.subType)
		}
	}
}

func TestAllGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "cyberpunk", "horror", "postapoc"}

	for _, genre := range genres {
		sys := NewSystem(genre, 100)

		comp := ItemIconComponent{
			Seed:         12345,
			IconType:     "weapon",
			Rarity:       3,
			SubType:      "sword",
			IconSize:     48,
			BorderGlow:   true,
			EnchantLevel: 2,
			Durability:   1.0,
		}

		icon := sys.GenerateIcon(&comp)
		if icon == nil {
			t.Errorf("Failed to generate icon for genre %s", genre)
		}
	}
}

func TestIconSizes(t *testing.T) {
	sys := NewSystem("fantasy", 100)

	sizes := []int{32, 48, 64, 128}

	for _, size := range sizes {
		comp := ItemIconComponent{
			Seed:       12345,
			IconType:   "weapon",
			Rarity:     2,
			SubType:    "sword",
			IconSize:   size,
			BorderGlow: true,
			Durability: 1.0,
		}

		icon := sys.GenerateIcon(&comp)
		if icon == nil {
			t.Errorf("Failed to generate icon for size %d", size)
			continue
		}

		bounds := icon.Bounds()
		if bounds.Dx() != size || bounds.Dy() != size {
			t.Errorf("Icon size = %dx%d, want %dx%d",
				bounds.Dx(), bounds.Dy(), size, size)
		}
	}
}

func TestComponentType(t *testing.T) {
	comp := &ItemIconComponent{}
	if comp.Type() != "itemicon" {
		t.Errorf("Component.Type() = %s, want itemicon", comp.Type())
	}
}

func BenchmarkIconGeneration(b *testing.B) {
	sys := NewSystem("fantasy", 100)

	comp := ItemIconComponent{
		Seed:         12345,
		IconType:     "weapon",
		Rarity:       3,
		SubType:      "sword",
		IconSize:     48,
		BorderGlow:   true,
		EnchantLevel: 2,
		Durability:   1.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.Seed = int64(i)
		_ = sys.GenerateIcon(&comp)
	}
}

func BenchmarkIconGenerationCached(b *testing.B) {
	sys := NewSystem("fantasy", 100)

	comp := ItemIconComponent{
		Seed:       12345,
		IconType:   "weapon",
		Rarity:     3,
		SubType:    "sword",
		IconSize:   48,
		BorderGlow: true,
		Durability: 1.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.GenerateIcon(&comp)
	}
}

func TestZeroSizeDefault(t *testing.T) {
	sys := NewSystem("fantasy", 100)

	comp := ItemIconComponent{
		Seed:       12345,
		IconType:   "weapon",
		Rarity:     2,
		SubType:    "sword",
		IconSize:   0,
		BorderGlow: true,
		Durability: 1.0,
	}

	icon := sys.GenerateIcon(&comp)
	if icon == nil {
		t.Error("GenerateIcon() returned nil")
		return
	}

	bounds := icon.Bounds()
	if bounds.Dx() != 48 || bounds.Dy() != 48 {
		t.Errorf("Icon size with zero input = %dx%d, want 48x48",
			bounds.Dx(), bounds.Dy())
	}
}

func TestCacheEviction(t *testing.T) {
	sys := NewSystem("fantasy", 5)

	for i := 0; i < 10; i++ {
		comp := ItemIconComponent{
			Seed:       int64(i),
			IconType:   "weapon",
			Rarity:     i % 5,
			SubType:    "sword",
			IconSize:   48,
			BorderGlow: true,
			Durability: 1.0,
		}

		icon := sys.GenerateIcon(&comp)
		if icon == nil {
			t.Errorf("Failed to generate icon %d", i)
		}
	}

	sys.mu.RLock()
	cacheSize := len(sys.cache)
	sys.mu.RUnlock()

	if cacheSize > sys.maxSize {
		t.Errorf("Cache size %d exceeds max size %d", cacheSize, sys.maxSize)
	}
}

func TestImagePooling(t *testing.T) {
	sys := NewSystem("fantasy", 100)

	comp := ItemIconComponent{
		Seed:       12345,
		IconType:   "weapon",
		Rarity:     2,
		SubType:    "sword",
		IconSize:   48,
		BorderGlow: true,
		Durability: 1.0,
	}

	for i := 0; i < 50; i++ {
		comp.Seed = int64(i)
		icon := sys.GenerateIcon(&comp)
		if icon == nil {
			t.Errorf("Failed to generate icon %d", i)
		}
	}
}

func TestNilImage(t *testing.T) {
	sys := NewSystem("fantasy", 100)

	comp := ItemIconComponent{
		Seed:       12345,
		IconType:   "weapon",
		Rarity:     2,
		SubType:    "sword",
		IconSize:   48,
		BorderGlow: true,
		Durability: 1.0,
	}

	icon := sys.GenerateIcon(&comp)

	if icon == nil {
		t.Fatal("GenerateIcon returned nil")
	}

	if icon == (*ebiten.Image)(nil) {
		t.Error("GenerateIcon returned typed nil")
	}
}
