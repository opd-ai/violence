// Package sprite provides procedural sprite generation with shading and detail.
package sprite

import (
	"container/list"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/common"
	"github.com/opd-ai/violence/pkg/dither"
	"github.com/opd-ai/violence/pkg/pool"
	"github.com/opd-ai/violence/pkg/rimlight"
)

// SpriteType identifies the category of sprite to generate.
type SpriteType int

const (
	SpriteEnemy SpriteType = iota
	SpriteProp
	SpriteLoreItem
	SpriteDestructible
	SpritePickup
	SpriteProjectile
)

// SpriteKey uniquely identifies a cached sprite.
type SpriteKey struct {
	Type    SpriteType
	Subtype string
	Seed    int64
	Frame   int
	Size    int
}

// CachedSprite stores a generated sprite with metadata.
type CachedSprite struct {
	Image     *ebiten.Image
	Key       SpriteKey
	AccessCnt int
}

// Generator creates procedural sprites with caching.
type Generator struct {
	cache       map[SpriteKey]*list.Element
	lruList     *list.List
	maxEntries  int
	mu          sync.RWMutex
	genreID     string
	lightCfg    LightConfig
	ditherSys   *dither.System
	rimlightSys *rimlight.System
}

// NewGenerator creates a sprite generator with LRU cache.
func NewGenerator(maxCacheEntries int) *Generator {
	return &Generator{
		cache:       make(map[SpriteKey]*list.Element),
		lruList:     list.New(),
		maxEntries:  maxCacheEntries,
		genreID:     "fantasy",
		lightCfg:    DefaultLightConfig(),
		ditherSys:   dither.NewSystem(12345),
		rimlightSys: rimlight.NewSystem("fantasy"),
	}
}

// SetGenre configures genre-specific sprite generation.
func (g *Generator) SetGenre(genreID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.genreID = genreID
	g.cache = make(map[SpriteKey]*list.Element)
	g.lruList = list.New()
	g.rimlightSys.SetGenre(genreID)
}

// GetSprite retrieves or generates a sprite.
func (g *Generator) GetSprite(spriteType SpriteType, subtype string, seed int64, frame, size int) *ebiten.Image {
	key := SpriteKey{
		Type:    spriteType,
		Subtype: subtype,
		Seed:    seed,
		Frame:   frame,
		Size:    size,
	}

	g.mu.Lock()
	if elem, found := g.cache[key]; found {
		g.lruList.MoveToFront(elem)
		cached := elem.Value.(*CachedSprite)
		cached.AccessCnt++
		g.mu.Unlock()
		return cached.Image
	}
	g.mu.Unlock()

	img := g.generateSprite(spriteType, subtype, seed, frame, size)

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.lruList.Len() >= g.maxEntries {
		oldest := g.lruList.Back()
		if oldest != nil {
			g.lruList.Remove(oldest)
			delete(g.cache, oldest.Value.(*CachedSprite).Key)
		}
	}

	cached := &CachedSprite{
		Image:     img,
		Key:       key,
		AccessCnt: 1,
	}
	elem := g.lruList.PushFront(cached)
	g.cache[key] = elem

	return img
}

// generateSprite creates a new procedural sprite.
func (g *Generator) generateSprite(spriteType SpriteType, subtype string, seed int64, frame, size int) *ebiten.Image {
	rgba := pool.GlobalPools.Images.Get(size, size)
	rng := rand.New(rand.NewSource(seed))

	// Determine dominant material for dithering, rim lighting, and normal perturbation
	var dominantMaterial dither.Material
	var rimMaterial rimlight.Material
	var pbrMaterial MaterialDetail

	switch spriteType {
	case SpriteEnemy:
		g.generateEnemySprite(rgba, subtype, rng, frame)
		dominantMaterial = g.getDominantMaterialForEnemy(subtype)
		rimMaterial = g.getRimMaterialForEnemy(subtype)
		pbrMaterial = g.getPBRMaterialForEnemy(subtype)
	case SpriteProp:
		g.generatePropSprite(rgba, subtype, rng, frame)
		dominantMaterial = g.getDominantMaterialForProp(subtype)
		rimMaterial = g.getRimMaterialForProp(subtype)
		pbrMaterial = g.getPBRMaterialForProp(subtype)
	case SpriteLoreItem:
		g.generateLoreSprite(rgba, subtype, rng, frame)
		dominantMaterial = dither.MaterialCrystal
		rimMaterial = rimlight.MaterialMagic
		pbrMaterial = MaterialCrystal
	case SpriteDestructible:
		g.generateDestructibleSprite(rgba, subtype, rng, frame)
		dominantMaterial = dither.MaterialLeather
		rimMaterial = rimlight.MaterialDefault
		pbrMaterial = MaterialLeather
	case SpritePickup:
		g.generatePickupSprite(rgba, subtype, rng, frame)
		dominantMaterial = dither.MaterialMetal
		rimMaterial = rimlight.MaterialMetal
		pbrMaterial = MaterialMetal
	case SpriteProjectile:
		g.generateProjectileSprite(rgba, subtype, rng, frame)
		dominantMaterial = dither.MaterialCrystal
		rimMaterial = rimlight.MaterialMagic
		pbrMaterial = MaterialCrystal
	default:
		g.generateDefaultSprite(rgba, rng)
		dominantMaterial = dither.MaterialDefault
		rimMaterial = rimlight.MaterialDefault
		pbrMaterial = MaterialLeather
	}

	// Apply normal perturbation pass for micro-surface detail
	// This adds material-specific surface texture to lighting calculations
	g.applyNormalPerturbationPass(rgba, pbrMaterial, seed)

	// Apply dithering as final pass for smoother tonal transitions
	g.applyDitheringPass(rgba, dominantMaterial, seed)

	// Apply rim lighting for directional edge highlights
	g.applyRimLightPass(rgba, rimMaterial, seed)

	result := ebiten.NewImageFromImage(rgba)
	pool.GlobalPools.Images.Put(rgba)
	return result
}

// applyDitheringPass applies material-appropriate dithering to enhance visual depth.
func (g *Generator) applyDitheringPass(img *image.RGBA, material dither.Material, seed int64) {
	// Use seed-based intensity variation for organic-looking results
	rng := rand.New(rand.NewSource(seed))
	intensity := 0.3 + rng.Float64()*0.2 // 0.3-0.5 intensity

	// Apply material-specific dithering
	g.ditherSys.ApplyDithering(img, material, intensity)

	// Apply edge-aware dithering for smooth shading transitions
	g.ditherSys.ApplyEdgeDithering(img, img.Bounds(), intensity*0.5)
}

// applyNormalPerturbationPass applies material-specific normal perturbation shading.
// This creates micro-surface detail that makes materials appear more realistic by
// perturbing surface normals to add texture-appropriate lighting variation.
func (g *Generator) applyNormalPerturbationPass(img *image.RGBA, material MaterialDetail, seed int64) {
	bounds := img.Bounds()
	cx := bounds.Dx() / 2
	cy := bounds.Dy() / 2
	radius := float64(bounds.Dx()) / 2

	perturbCfg := NormalPerturbForMaterial(material)

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			// Skip transparent pixels
			existingColor := img.At(x, y)
			_, _, _, ea := existingColor.RGBA()
			if ea == 0 {
				continue
			}

			// Get current color
			er, eg, eb, _ := existingColor.RGBA()
			r := float64(er >> 8)
			gr := float64(eg >> 8)
			b := float64(eb >> 8)

			// Calculate base normal from position (assuming spherical)
			dx := float64(x - cx)
			dy := float64(y - cy)
			nx, ny, nz := CalculateSurfaceNormal(dx, dy, radius)

			// Perturb the normal based on material
			pnx, pny, pnz := PerturbNormal(nx, ny, nz, x, y, material, seed, perturbCfg)

			// Calculate lighting adjustment from perturbed normal
			// Use default light direction (-0.5773, -0.5773, 0.5773) normalized
			lightDirX, lightDirY, lightDirZ := -0.5773, -0.5773, 0.5773

			// Original NdotL
			origNdotL := math.Max(0, nx*lightDirX+ny*lightDirY+nz*lightDirZ)
			// Perturbed NdotL
			perturbedNdotL := math.Max(0, pnx*lightDirX+pny*lightDirY+pnz*lightDirZ)

			// Calculate lighting adjustment factor
			// We don't want drastic changes, just subtle surface detail
			lightAdjust := 1.0 + (perturbedNdotL-origNdotL)*0.4

			// Apply adjustment
			r = math.Max(0, math.Min(255, r*lightAdjust))
			gr = math.Max(0, math.Min(255, gr*lightAdjust))
			b = math.Max(0, math.Min(255, b*lightAdjust))

			img.Set(x, y, color.RGBA{
				R: uint8(r),
				G: uint8(gr),
				B: uint8(b),
				A: uint8(ea >> 8),
			})
		}
	}
}

// getPBRMaterialForEnemy returns the PBR material type for an enemy subtype.
func (g *Generator) getPBRMaterialForEnemy(subtype string) MaterialDetail {
	switch subtype {
	case "humanoid", "skeleton", "zombie":
		return MaterialLeather
	case "quadruped", "wolf", "bear":
		return MaterialFur
	case "insect", "spider", "scorpion":
		return MaterialChitin
	case "serpent", "snake", "dragon":
		return MaterialScales
	case "flying", "bat", "bird":
		return MaterialFur
	case "amorphous", "slime", "ooze":
		return MaterialSlime
	case "golem", "robot", "mech":
		return MaterialMetal
	case "elemental", "spirit", "ghost":
		return MaterialCrystal
	default:
		return MaterialLeather
	}
}

// getPBRMaterialForProp returns the PBR material type for a prop subtype.
func (g *Generator) getPBRMaterialForProp(subtype string) MaterialDetail {
	switch subtype {
	case "barrel", "crate", "table":
		return MaterialLeather // Wood-like
	case "terminal", "container":
		return MaterialMetal
	case "bones", "debris":
		return MaterialLeather
	case "plant":
		return MaterialCloth
	case "pillar":
		return MaterialLeather
	case "torch":
		return MaterialCrystal
	default:
		return MaterialLeather
	}
}

// getDominantMaterialForEnemy returns the dither material for an enemy subtype.
func (g *Generator) getDominantMaterialForEnemy(subtype string) dither.Material {
	switch subtype {
	case "humanoid", "skeleton", "zombie":
		return dither.MaterialFlesh
	case "quadruped", "wolf", "bear":
		return dither.MaterialFur
	case "insect", "spider", "scorpion":
		return dither.MaterialScales
	case "serpent", "snake", "dragon":
		return dither.MaterialScales
	case "flying", "bat", "bird":
		return dither.MaterialFur
	case "amorphous", "slime", "ooze":
		return dither.MaterialSlime
	case "golem", "robot", "mech":
		return dither.MaterialMetal
	case "elemental", "spirit", "ghost":
		return dither.MaterialCrystal
	default:
		return dither.MaterialDefault
	}
}

// getDominantMaterialForProp returns the dither material for a prop subtype.
func (g *Generator) getDominantMaterialForProp(subtype string) dither.Material {
	switch subtype {
	case "barrel", "crate", "table":
		return dither.MaterialLeather // Wood-like
	case "terminal", "container":
		return dither.MaterialMetal
	case "bones", "debris":
		return dither.MaterialFlesh
	case "plant":
		return dither.MaterialCloth // Soft organic
	case "pillar":
		return dither.MaterialLeather // Stone-like
	case "torch":
		return dither.MaterialCrystal // Glowing
	default:
		return dither.MaterialDefault
	}
}

// applyRimLightPass applies rim lighting for directional edge highlights.
func (g *Generator) applyRimLightPass(img *image.RGBA, material rimlight.Material, seed int64) {
	// Use seed-based intensity variation
	rng := rand.New(rand.NewSource(seed))
	intensity := 0.8 + rng.Float64()*0.4 // 0.8-1.2 intensity

	g.rimlightSys.ApplyRimLightToImage(img, material, intensity)
}

// getRimMaterialForEnemy returns the rim light material for an enemy subtype.
func (g *Generator) getRimMaterialForEnemy(subtype string) rimlight.Material {
	switch subtype {
	case "humanoid", "skeleton", "zombie":
		return rimlight.MaterialOrganic
	case "quadruped", "wolf", "bear":
		return rimlight.MaterialOrganic
	case "insect", "spider", "scorpion":
		return rimlight.MaterialLeather
	case "serpent", "snake", "dragon":
		return rimlight.MaterialLeather
	case "flying", "bat", "bird":
		return rimlight.MaterialCloth
	case "amorphous", "slime", "ooze":
		return rimlight.MaterialCrystal
	case "golem", "robot", "mech":
		return rimlight.MaterialMetal
	case "elemental", "spirit", "ghost":
		return rimlight.MaterialMagic
	default:
		return rimlight.MaterialDefault
	}
}

// getRimMaterialForProp returns the rim light material for a prop subtype.
func (g *Generator) getRimMaterialForProp(subtype string) rimlight.Material {
	switch subtype {
	case "barrel", "crate", "table":
		return rimlight.MaterialLeather
	case "terminal", "container":
		return rimlight.MaterialMetal
	case "bones", "debris":
		return rimlight.MaterialOrganic
	case "plant":
		return rimlight.MaterialCloth
	case "pillar":
		return rimlight.MaterialDefault
	case "torch":
		return rimlight.MaterialMagic
	default:
		return rimlight.MaterialDefault
	}
}

// generatePropSprite creates prop sprites with genre-specific styling.
func (g *Generator) generatePropSprite(img *image.RGBA, subtype string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	switch subtype {
	case "barrel":
		g.drawBarrel(img, cx, cy, size, rng)
	case "crate":
		g.drawCrate(img, cx, cy, size, rng)
	case "table":
		g.drawTable(img, cx, cy, size, rng)
	case "terminal":
		g.drawTerminal(img, cx, cy, size, rng)
	case "bones":
		g.drawBones(img, cx, cy, size, rng)
	case "plant":
		g.drawPlant(img, cx, cy, size, rng)
	case "pillar":
		g.drawPillar(img, cx, cy, size, rng)
	case "torch":
		g.drawTorch(img, cx, cy, size, rng, frame)
	case "debris":
		g.drawDebris(img, cx, cy, size, rng)
	case "container":
		g.drawContainer(img, cx, cy, size, rng)
	default:
		g.drawCrate(img, cx, cy, size, rng)
	}
}

// drawBarrel renders a cylindrical barrel with wood grain and metal bands.
func (g *Generator) drawBarrel(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	baseColor := g.getGenreWoodColor()
	metalColor := color.RGBA{R: 80, G: 80, B: 90, A: 255}

	radius := size / 3
	height := size * 2 / 3

	for y := cy - height/2; y < cy+height/2; y++ {
		bulge := int(float64(radius) * (1.0 + 0.2*math.Sin(float64(y-cy+height/2)/float64(height)*math.Pi)))
		for x := cx - bulge; x < cx+bulge; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}

			dist := math.Abs(float64(x - cx))
			shade := 1.0 - dist/float64(bulge)*0.5

			woodGrain := math.Sin(float64(y)*0.3+rng.Float64()*0.5) * 0.1
			shade += woodGrain

			r := uint8(float64(baseColor.R) * shade)
			g := uint8(float64(baseColor.G) * shade)
			b := uint8(float64(baseColor.B) * shade)

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	bandY1 := cy - height/3
	bandY2 := cy + height/3
	for _, by := range []int{bandY1, bandY2} {
		for y := by - 2; y < by+2; y++ {
			for x := cx - radius; x < cx+radius; x++ {
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					img.Set(x, y, metalColor)
				}
			}
		}
	}

	// Apply PBR shading for realistic wood and metal rendering
	barrelBounds := image.Rect(cx-radius, cy-height/2, cx+radius, cy+height/2)
	g.ApplyPBRShadingToRegion(img, barrelBounds, MaterialLeather, "cylindrical", g.lightCfg)
}

// drawCrate renders a wooden crate with planks and corner reinforcements.
func (g *Generator) drawCrate(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	baseColor := g.getGenreWoodColor()
	darkColor := color.RGBA{
		R: baseColor.R / 2,
		G: baseColor.G / 2,
		B: baseColor.B / 2,
		A: 255,
	}

	boxSize := size * 2 / 3
	x1, y1 := cx-boxSize/2, cy-boxSize/2
	x2, y2 := cx+boxSize/2, cy+boxSize/2

	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}

			dx := float64(x - cx)
			dy := float64(y - cy)
			distFromEdge := math.Min(math.Abs(dx), math.Abs(dy))
			shade := 0.7 + 0.3*distFromEdge/float64(boxSize/2)

			plankNoise := rng.Float64()*0.1 - 0.05
			shade += plankNoise

			r := uint8(math.Min(255, float64(baseColor.R)*shade))
			g := uint8(math.Min(255, float64(baseColor.G)*shade))
			b := uint8(math.Min(255, float64(baseColor.B)*shade))

			plankPhase := int(math.Floor(float64(y-y1) / 8))
			if plankPhase%2 == 0 {
				r = uint8(float64(r) * 0.9)
				g = uint8(float64(g) * 0.9)
				b = uint8(float64(b) * 0.9)
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	for i := 0; i < 4; i++ {
		cornerX := x1
		cornerY := y1
		if i%2 == 1 {
			cornerX = x2 - 4
		}
		if i/2 == 1 {
			cornerY = y2 - 4
		}
		common.FillRect(img, cornerX, cornerY, cornerX+4, cornerY+4, darkColor)
	}

	for x := x1; x < x2; x++ {
		img.Set(x, y1, darkColor)
		img.Set(x, y2-1, darkColor)
	}
	for y := y1; y < y2; y++ {
		img.Set(x1, y, darkColor)
		img.Set(x2-1, y, darkColor)
	}

	// Apply PBR shading for realistic wood rendering
	crateBounds := image.Rect(x1, y1, x2, y2)
	g.ApplyPBRShadingToRegion(img, crateBounds, MaterialLeather, "planar", g.lightCfg)
}

// drawTable renders a table sprite with perspective.
func (g *Generator) drawTable(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	woodColor := g.getGenreWoodColor()

	tableW := size * 3 / 4
	tableH := size / 4
	tableY := cy - size/6

	for y := tableY; y < tableY+tableH; y++ {
		for x := cx - tableW/2; x < cx+tableW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				shade := 0.8 + 0.2*float64(y-tableY)/float64(tableH)
				r := uint8(float64(woodColor.R) * shade)
				g := uint8(float64(woodColor.G) * shade)
				b := uint8(float64(woodColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	legColor := color.RGBA{R: woodColor.R / 2, G: woodColor.G / 2, B: woodColor.B / 2, A: 255}
	legPositions := [][2]int{
		{cx - tableW/2 + 4, tableY + tableH},
		{cx + tableW/2 - 8, tableY + tableH},
	}

	for _, legPos := range legPositions {
		lx, ly := legPos[0], legPos[1]
		common.FillRect(img, lx, ly, lx+4, ly+size/3, legColor)
	}
}

// drawTerminal renders a sci-fi terminal with screen and panel.
func (g *Generator) drawTerminal(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	panelColor := color.RGBA{R: 40, G: 45, B: 50, A: 255}
	screenColor := color.RGBA{R: 0, G: 100, B: 120, A: 255}
	glowColor := color.RGBA{R: 0, G: 200, B: 255, A: 255}

	termW := size * 2 / 3
	termH := size * 3 / 4
	x1, y1 := cx-termW/2, cy-termH/2

	common.FillRect(img, x1, y1, x1+termW, y1+termH, panelColor)

	screenX := x1 + termW/6
	screenY := y1 + termH/6
	screenW := termW * 2 / 3
	screenH := termH / 2

	for y := screenY; y < screenY+screenH; y++ {
		for x := screenX; x < screenX+screenW; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				scanline := math.Sin(float64(y)*0.5) * 0.15
				r := uint8(float64(screenColor.R) * (1.0 + scanline))
				g := uint8(float64(screenColor.G) * (1.0 + scanline))
				b := uint8(float64(screenColor.B) * (1.0 + scanline))
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	for i := 0; i < 3; i++ {
		lx := screenX + i*screenW/4 + 4
		ly := screenY + screenH/4
		common.FillCircle(img, lx, ly, 2, glowColor)
	}
}

// drawBones renders skeletal remains.
func (g *Generator) drawBones(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	boneColor := color.RGBA{R: 220, G: 210, B: 190, A: 255}

	for i := 0; i < 3; i++ {
		angle := rng.Float64() * 2 * math.Pi
		length := size/3 + rng.Intn(size/4)
		x1 := cx + int(math.Cos(angle)*float64(length/4))
		y1 := cy + int(math.Sin(angle)*float64(length/4))
		x2 := cx + int(math.Cos(angle)*float64(length))
		y2 := cy + int(math.Sin(angle)*float64(length))

		common.DrawThickLine(img, x1, y1, x2, y2, 2, boneColor)
	}

	common.FillCircle(img, cx, cy, size/8, boneColor)
}

// drawPlant renders foliage with leaves.
func (g *Generator) drawPlant(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	leafColor := g.getGenreLeafColor()
	stemColor := color.RGBA{R: 60, G: 80, B: 40, A: 255}

	stemX := cx
	stemBot := cy + size/3
	stemTop := cy - size/3

	common.DrawThickLine(img, stemX, stemBot, stemX, stemTop, 2, stemColor)

	for i := 0; i < 5; i++ {
		leafY := stemBot - i*size/7
		leafSize := 6 - i
		if leafSize < 2 {
			leafSize = 2
		}

		common.FillCircle(img, stemX-leafSize, leafY, leafSize, leafColor)
		common.FillCircle(img, stemX+leafSize, leafY, leafSize, leafColor)
	}
}

// drawPillar renders a stone pillar with weathering.
func (g *Generator) drawPillar(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	stoneColor := g.getGenreStoneColor()

	pillarW := size / 3
	pillarH := size * 3 / 4
	x1 := cx - pillarW/2
	y1 := cy - pillarH/2

	for y := y1; y < y1+pillarH; y++ {
		for x := x1; x < x1+pillarW; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}

			dx := float64(x - cx)
			shade := 1.0 - math.Abs(dx)/float64(pillarW)*0.5

			noise := (rng.Float64() - 0.5) * 0.2
			shade += noise

			r := uint8(math.Max(0, math.Min(255, float64(stoneColor.R)*shade)))
			g := uint8(math.Max(0, math.Min(255, float64(stoneColor.G)*shade)))
			b := uint8(math.Max(0, math.Min(255, float64(stoneColor.B)*shade)))

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	capH := pillarH / 10
	for y := y1; y < y1+capH; y++ {
		for x := x1 - 2; x < x1+pillarW+2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, stoneColor)
			}
		}
	}
}

// drawTorch renders an animated torch with flame.
func (g *Generator) drawTorch(img *image.RGBA, cx, cy, size int, rng *rand.Rand, frame int) {
	handleColor := color.RGBA{R: 80, G: 60, B: 40, A: 255}
	fireColor := color.RGBA{R: 255, G: 140, B: 0, A: 255}
	glowColor := color.RGBA{R: 255, G: 200, B: 50, A: 255}

	handleW := 4
	handleH := size / 2
	handleX := cx - handleW/2
	handleY := cy

	common.FillRect(img, handleX, handleY, handleX+handleW, handleY+handleH, handleColor)

	flameY := cy - size/8
	flameFlicker := int(math.Sin(float64(frame)*0.2) * 3)
	flameH := size/4 + flameFlicker

	for i := flameH; i > 0; i-- {
		flameSize := 3 + i/3
		y := flameY - i
		c := fireColor
		if i > flameH/2 {
			c = glowColor
		}
		common.FillCircle(img, cx, y, flameSize, c)
	}
}

// drawDebris renders scattered rubble.
func (g *Generator) drawDebris(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	debrisColor := color.RGBA{R: 100, G: 95, B: 90, A: 255}

	for i := 0; i < 6; i++ {
		dx := rng.Intn(size/2) - size/4
		dy := rng.Intn(size/2) - size/4
		dsize := 3 + rng.Intn(5)

		shade := 0.6 + rng.Float64()*0.4
		r := uint8(float64(debrisColor.R) * shade)
		g := uint8(float64(debrisColor.G) * shade)
		b := uint8(float64(debrisColor.B) * shade)

		common.FillCircle(img, cx+dx, cy+dy, dsize, color.RGBA{R: r, G: g, B: b, A: 255})
	}
}

// drawContainer renders a futuristic storage container.
func (g *Generator) drawContainer(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	baseColor := color.RGBA{R: 180, G: 180, B: 190, A: 255}
	accentColor := color.RGBA{R: 255, G: 200, B: 0, A: 255}

	boxSize := size * 2 / 3
	x1, y1 := cx-boxSize/2, cy-boxSize/2

	common.FillRect(img, x1, y1, x1+boxSize, y1+boxSize, baseColor)

	for i := 0; i < 3; i++ {
		stripeY := y1 + i*boxSize/4 + boxSize/8
		common.FillRect(img, x1, stripeY, x1+boxSize, stripeY+2, accentColor)
	}

	handleW := boxSize / 6
	handleY := y1 + boxSize/3
	common.FillRect(img, cx-handleW/2, handleY, cx+handleW/2, handleY+4, color.RGBA{R: 100, G: 100, B: 110, A: 255})
}

// generateLoreSprite creates lore item sprites.
func (g *Generator) generateLoreSprite(img *image.RGBA, subtype string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	switch subtype {
	case "note":
		g.drawNote(img, cx, cy, size, rng)
	case "audiolog":
		g.drawAudioLog(img, cx, cy, size, rng)
	case "graffiti":
		g.drawGraffiti(img, cx, cy, size, rng)
	case "body":
		g.drawBodyArrangement(img, cx, cy, size, rng)
	default:
		g.drawNote(img, cx, cy, size, rng)
	}
}

// drawNote renders a paper document.
func (g *Generator) drawNote(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	paperColor := color.RGBA{R: 240, G: 235, B: 210, A: 255}
	textColor := color.RGBA{R: 40, G: 40, B: 50, A: 255}

	noteW := size / 2
	noteH := size * 2 / 3
	x1, y1 := cx-noteW/2, cy-noteH/2

	common.FillRect(img, x1, y1, x1+noteW, y1+noteH, paperColor)

	for i := 0; i < 5; i++ {
		lineY := y1 + noteH/6 + i*noteH/8
		common.FillRect(img, x1+4, lineY, x1+noteW-4, lineY+1, textColor)
	}
}

// drawAudioLog renders a recording device.
func (g *Generator) drawAudioLog(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	deviceColor := color.RGBA{R: 60, G: 70, B: 80, A: 255}
	ledColor := color.RGBA{R: 0, G: 255, B: 100, A: 255}

	deviceSize := size / 2
	x1, y1 := cx-deviceSize/2, cy-deviceSize/2

	common.FillRect(img, x1, y1, x1+deviceSize, y1+deviceSize, deviceColor)

	common.FillCircle(img, cx, cy-deviceSize/4, 3, ledColor)
}

// drawGraffiti renders wall art.
func (g *Generator) drawGraffiti(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	graffitiColor := color.RGBA{R: 255, G: 50, B: 50, A: 255}

	for i := 0; i < 4; i++ {
		angle := float64(i) * math.Pi / 2
		x := cx + int(math.Cos(angle)*float64(size/4))
		y := cy + int(math.Sin(angle)*float64(size/4))
		common.FillCircle(img, x, y, 3, graffitiColor)
	}

	common.FillCircle(img, cx, cy, size/6, graffitiColor)
}

// drawBodyArrangement renders skeletal arrangement.
func (g *Generator) drawBodyArrangement(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	g.drawBones(img, cx, cy, size, rng)
}

// generateDestructibleSprite creates destructible object sprites.
func (g *Generator) generateDestructibleSprite(img *image.RGBA, subtype string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	switch subtype {
	case "barrel":
		g.drawBarrel(img, cx, cy, size, rng)
	case "crate":
		g.drawCrate(img, cx, cy, size, rng)
	default:
		g.drawCrate(img, cx, cy, size, rng)
	}
}

// generatePickupSprite creates item pickup sprites.
func (g *Generator) generatePickupSprite(img *image.RGBA, subtype string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	pulse := 1.0 + 0.1*math.Sin(float64(frame)*0.1)

	switch subtype {
	case "health":
		g.drawHealthPickup(img, cx, cy, size, pulse)
	case "ammo":
		g.drawAmmoPickup(img, cx, cy, size, pulse)
	case "armor":
		g.drawArmorPickup(img, cx, cy, size, pulse)
	default:
		g.drawHealthPickup(img, cx, cy, size, pulse)
	}
}

// drawHealthPickup renders a health pack with cross.
func (g *Generator) drawHealthPickup(img *image.RGBA, cx, cy, size int, pulse float64) {
	healthColor := color.RGBA{R: uint8(float64(255) * pulse), G: 50, B: 50, A: 255}

	crossW := size / 3
	crossH := size / 2

	common.FillRect(img, cx-crossW/2, cy-crossH/2, cx+crossW/2, cy+crossH/2, healthColor)
	common.FillRect(img, cx-crossH/2, cy-crossW/2, cx+crossH/2, cy+crossW/2, healthColor)
}

// drawAmmoPickup renders ammunition box.
func (g *Generator) drawAmmoPickup(img *image.RGBA, cx, cy, size int, pulse float64) {
	ammoColor := color.RGBA{R: uint8(float64(200) * pulse), G: uint8(float64(150) * pulse), B: 0, A: 255}

	boxSize := size / 2
	common.FillRect(img, cx-boxSize/2, cy-boxSize/2, cx+boxSize/2, cy+boxSize/2, ammoColor)
}

// drawArmorPickup renders armor plate.
func (g *Generator) drawArmorPickup(img *image.RGBA, cx, cy, size int, pulse float64) {
	armorColor := color.RGBA{R: uint8(float64(100) * pulse), G: uint8(float64(150) * pulse), B: uint8(float64(200) * pulse), A: 255}

	shieldRadius := size / 3
	common.FillCircle(img, cx, cy, shieldRadius, armorColor)
}

// generateProjectileSprite creates projectile sprites.
func (g *Generator) generateProjectileSprite(img *image.RGBA, subtype string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	bulletColor := color.RGBA{R: 255, G: 220, B: 100, A: 255}
	common.FillCircle(img, cx, cy, size/4, bulletColor)
}

// generateEnemySprite creates enemy sprites with body plan variety.
func (g *Generator) generateEnemySprite(img *image.RGBA, subtype string, rng *rand.Rand, frame int) {
	switch subtype {
	case "humanoid", "tank", "ranged", "healer", "ambusher", "scout":
		g.generateHumanoidEnemy(img, subtype, rng, frame)
	case "quadruped":
		g.generateQuadrupedEnemy(img, rng, frame)
	case "insect":
		g.generateInsectEnemy(img, rng, frame)
	case "serpent":
		g.generateSerpentEnemy(img, rng, frame)
	case "flying":
		g.generateFlyingEnemy(img, rng, frame)
	case "amorphous":
		g.generateAmorphousEnemy(img, rng, frame)
	default:
		g.generateHumanoidEnemy(img, subtype, rng, frame)
	}
}

// generateHumanoidEnemy creates genre-aware humanoid enemy sprites.
func (g *Generator) generateHumanoidEnemy(img *image.RGBA, role string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	armorColor, accentColor, skinColor, weaponType := g.selectGenreColors()
	bodyParts := g.calculateBodyPartPositions(size, cx, cy, role, frame)

	g.drawLegs(img, size, cx, armorColor, bodyParts)
	g.drawTorso(img, cx, armorColor, accentColor, role, bodyParts)
	g.drawArms(img, size, cx, armorColor, frame, bodyParts)
	g.drawHead(img, size, cx, skinColor, bodyParts)
	g.drawWeapon(img, size, cx, weaponType, armorColor, frame, bodyParts)
	g.drawRoleSpecificDecorations(img, cx, role, accentColor, bodyParts)
	g.applyMaterialTextures(img, size, cx, armorColor, skinColor, rng, bodyParts)

	// Apply PBR shading to each body part for realistic lighting
	// Legs (cylindrical geometry)
	leftLegBounds := image.Rect(cx-size/8-bodyParts.legW/2, bodyParts.leftLegY, cx-size/8+bodyParts.legW/2, bodyParts.leftLegY+bodyParts.legH)
	g.ApplyPBRShadingToRegion(img, leftLegBounds, MaterialMetal, "cylindrical", g.lightCfg)
	rightLegBounds := image.Rect(cx+size/8-bodyParts.legW/2, bodyParts.rightLegY, cx+size/8+bodyParts.legW/2, bodyParts.rightLegY+bodyParts.legH)
	g.ApplyPBRShadingToRegion(img, rightLegBounds, MaterialMetal, "cylindrical", g.lightCfg)

	// Torso (cylindrical geometry)
	torsoBounds := image.Rect(cx-bodyParts.torsoW/2, bodyParts.bodyY, cx+bodyParts.torsoW/2, bodyParts.bodyY+bodyParts.torsoH)
	g.ApplyPBRShadingToRegion(img, torsoBounds, MaterialMetal, "cylindrical", g.lightCfg)

	// Arms (cylindrical geometry)
	leftArmBounds := image.Rect(cx-bodyParts.torsoW/2-bodyParts.armW, bodyParts.leftArmY, cx-bodyParts.torsoW/2, bodyParts.leftArmY+bodyParts.armH)
	g.ApplyPBRShadingToRegion(img, leftArmBounds, MaterialMetal, "cylindrical", g.lightCfg)
	rightArmBounds := image.Rect(cx+bodyParts.torsoW/2, bodyParts.rightArmY+bodyParts.attackOffset, cx+bodyParts.torsoW/2+bodyParts.armW, bodyParts.rightArmY+bodyParts.armH+bodyParts.attackOffset)
	g.ApplyPBRShadingToRegion(img, rightArmBounds, MaterialMetal, "cylindrical", g.lightCfg)

	// Head (spherical geometry)
	headBounds := image.Rect(cx-bodyParts.headRadius, bodyParts.bodyY-size/8-bodyParts.headRadius, cx+bodyParts.headRadius, bodyParts.bodyY-size/8+bodyParts.headRadius)
	g.ApplyPBRShadingToRegion(img, headBounds, MaterialLeather, "spherical", g.lightCfg)
}

type bodyPartPositions struct {
	leftLegY, rightLegY int
	leftArmY, rightArmY int
	bodyY               int
	torsoW, torsoH      int
	legW, legH          int
	armW, armH          int
	headRadius          int
	attackOffset        int
}

// selectGenreColors returns color scheme and weapon type based on genre.
func (g *Generator) selectGenreColors() (armorColor, accentColor, skinColor color.RGBA, weaponType int) {
	switch g.genreID {
	case "scifi":
		armorColor = color.RGBA{R: 40, G: 60, B: 80, A: 255}
		accentColor = color.RGBA{R: 80, G: 200, B: 240, A: 255}
		skinColor = color.RGBA{R: 100, G: 180, B: 255, A: 255}
		weaponType = 2
	case "horror":
		armorColor = color.RGBA{R: 60, G: 20, B: 20, A: 255}
		accentColor = color.RGBA{R: 255, G: 50, B: 50, A: 255}
		skinColor = color.RGBA{R: 180, G: 170, B: 160, A: 255}
		weaponType = 3
	case "cyberpunk":
		armorColor = color.RGBA{R: 30, G: 30, B: 35, A: 255}
		accentColor = color.RGBA{R: 255, G: 0, B: 128, A: 255}
		skinColor = color.RGBA{R: 0, G: 200, B: 255, A: 255}
		weaponType = 2
	case "postapoc":
		armorColor = color.RGBA{R: 100, G: 80, B: 60, A: 255}
		accentColor = color.RGBA{R: 110, G: 90, B: 70, A: 255}
		skinColor = color.RGBA{R: 190, G: 160, B: 140, A: 255}
		weaponType = 3
	default:
		armorColor = color.RGBA{R: 120, G: 120, B: 130, A: 255}
		accentColor = color.RGBA{R: 140, G: 140, B: 150, A: 255}
		skinColor = color.RGBA{R: 210, G: 180, B: 160, A: 255}
		weaponType = 1
	}
	return armorColor, accentColor, skinColor, weaponType
}

// calculateBodyPartPositions computes positions and dimensions for all body parts with animation offsets.
func (g *Generator) calculateBodyPartPositions(size, cx, cy int, role string, frame int) bodyPartPositions {
	positions := bodyPartPositions{
		leftLegY:   cy + size/6,
		rightLegY:  cy + size/6,
		leftArmY:   cy - size/10,
		rightArmY:  cy - size/10,
		bodyY:      cy - size/8,
		torsoW:     size / 3,
		torsoH:     size / 3,
		legW:       size / 12,
		legH:       size / 4,
		headRadius: size / 10,
	}

	positions.armW = positions.legW - 1
	positions.armH = size / 4

	if frame%3 == 1 {
		positions.leftLegY += 2
		positions.rightLegY -= 2
	} else if frame%3 == 2 {
		positions.leftLegY -= 2
		positions.rightLegY += 2
	}

	if role == "ambusher" && frame == 0 {
		positions.bodyY += 2
		positions.leftArmY += 2
		positions.rightArmY += 2
	}

	if frame%4 == 3 {
		positions.attackOffset = -3
	}

	return positions
}

// drawLegs renders both legs with shading.
func (g *Generator) drawLegs(img *image.RGBA, size, cx int, armorColor color.RGBA, pos bodyPartPositions) {
	for y := pos.leftLegY; y < pos.leftLegY+pos.legH; y++ {
		for x := cx - size/8 - pos.legW/2; x < cx-size/8+pos.legW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				shade := 0.7 + 0.3*float64(y-pos.leftLegY)/float64(pos.legH)
				r := uint8(float64(armorColor.R) * shade)
				g := uint8(float64(armorColor.G) * shade)
				b := uint8(float64(armorColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	for y := pos.rightLegY; y < pos.rightLegY+pos.legH; y++ {
		for x := cx + size/8 - pos.legW/2; x < cx+size/8+pos.legW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				shade := 0.7 + 0.3*float64(y-pos.rightLegY)/float64(pos.legH)
				r := uint8(float64(armorColor.R) * shade)
				g := uint8(float64(armorColor.G) * shade)
				b := uint8(float64(armorColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}
}

// drawTorso renders the body with optional role-specific accent stripes.
func (g *Generator) drawTorso(img *image.RGBA, cx int, armorColor, accentColor color.RGBA, role string, pos bodyPartPositions) {
	for y := pos.bodyY; y < pos.bodyY+pos.torsoH; y++ {
		for x := cx - pos.torsoW/2; x < cx+pos.torsoW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				dx := float64(x - cx)
				shade := 1.0 - math.Abs(dx)/float64(pos.torsoW)*0.4
				r := uint8(float64(armorColor.R) * shade)
				g := uint8(float64(armorColor.G) * shade)
				b := uint8(float64(armorColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	if role == "tank" || role == "healer" {
		accentY := pos.bodyY + pos.torsoH/3
		accentH := 2
		common.FillRect(img, cx-pos.torsoW/3, accentY, cx+pos.torsoW/3, accentY+accentH, accentColor)
	}
}

// drawArms renders both arms with animation offset for attack frames.
func (g *Generator) drawArms(img *image.RGBA, size, cx int, armorColor color.RGBA, frame int, pos bodyPartPositions) {
	for y := pos.leftArmY; y < pos.leftArmY+pos.armH; y++ {
		for x := cx - pos.torsoW/2 - pos.armW; x < cx-pos.torsoW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				shade := 0.8 + 0.2*float64(y-pos.leftArmY)/float64(pos.armH)
				r := uint8(float64(armorColor.R) * shade)
				g := uint8(float64(armorColor.G) * shade)
				b := uint8(float64(armorColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	for y := pos.rightArmY + pos.attackOffset; y < pos.rightArmY+pos.armH+pos.attackOffset; y++ {
		for x := cx + pos.torsoW/2; x < cx+pos.torsoW/2+pos.armW; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				shade := 0.8 + 0.2*float64(y-pos.rightArmY-pos.attackOffset)/float64(pos.armH)
				r := uint8(float64(armorColor.R) * shade)
				g := uint8(float64(armorColor.G) * shade)
				b := uint8(float64(armorColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}
}

// drawHead renders a circular head with gradient shading.
func (g *Generator) drawHead(img *image.RGBA, size, cx int, skinColor color.RGBA, pos bodyPartPositions) {
	for y := -pos.headRadius; y <= pos.headRadius; y++ {
		for x := -pos.headRadius; x <= pos.headRadius; x++ {
			if x*x+y*y <= pos.headRadius*pos.headRadius {
				px := cx + x
				py := pos.bodyY - size/16 + y
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					dist := math.Sqrt(float64(x*x + y*y))
					shade := 1.0 - (dist / float64(pos.headRadius) * 0.3)
					r := uint8(math.Min(255, float64(skinColor.R)*shade))
					g := uint8(math.Min(255, float64(skinColor.G)*shade))
					b := uint8(math.Min(255, float64(skinColor.B)*shade))
					img.Set(px, py, color.RGBA{R: r, G: g, B: b, A: 255})
				}
			}
		}
	}
}

// drawWeapon renders weapon based on type with animation offset.
func (g *Generator) drawWeapon(img *image.RGBA, size, cx, weaponType int, armorColor color.RGBA, frame int, pos bodyPartPositions) {
	weaponColor := color.RGBA{R: 200, G: 200, B: 220, A: 255}
	if weaponType == 2 {
		weaponColor = color.RGBA{R: 60, G: 60, B: 70, A: 255}
	} else if weaponType == 3 {
		weaponColor = color.RGBA{R: 120, G: 100, B: 80, A: 255}
	}

	weaponX := cx + pos.torsoW/2 + pos.armW
	weaponY := pos.rightArmY + pos.armH/2 + pos.attackOffset
	weaponLen := size / 5

	if weaponType == 2 {
		common.FillRect(img, weaponX, weaponY-1, weaponX+weaponLen, weaponY+1, weaponColor)
		common.FillCircle(img, weaponX+weaponLen, weaponY, 2, color.RGBA{R: 100, G: 100, B: 110, A: 255})
	} else {
		common.FillRect(img, weaponX, weaponY-1, weaponX+weaponLen, weaponY+1, weaponColor)
		if weaponType == 1 {
			common.FillCircle(img, weaponX+weaponLen, weaponY, 3, color.RGBA{R: 180, G: 180, B: 200, A: 255})
		}
	}
}

// drawRoleSpecificDecorations adds visual indicators for tank and healer roles.
func (g *Generator) drawRoleSpecificDecorations(img *image.RGBA, cx int, role string, accentColor color.RGBA, pos bodyPartPositions) {
	if role == "tank" {
		shieldX := cx - pos.torsoW/2 - pos.armW - 2
		shieldY := pos.leftArmY + pos.armH/4
		shieldH := pos.armH / 2
		for y := shieldY; y < shieldY+shieldH; y++ {
			for x := shieldX - 4; x < shieldX; x++ {
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					img.Set(x, y, color.RGBA{R: 160, G: 140, B: 100, A: 255})
				}
			}
		}
	}

	if role == "healer" {
		symbolX := cx - 2
		symbolY := pos.bodyY + pos.torsoH/2 - 2
		common.FillRect(img, symbolX, symbolY-4, symbolX+4, symbolY+4, accentColor)
		common.FillRect(img, symbolX-4, symbolY, symbolX+8, symbolY+4, accentColor)
	}
}

// applyMaterialTextures applies material detail to armor and skin.
func (g *Generator) applyMaterialTextures(img *image.RGBA, size, cx int, armorColor, skinColor color.RGBA, rng *rand.Rand, pos bodyPartPositions) {
	armorBounds := image.Rect(cx-pos.torsoW/2, pos.bodyY, cx+pos.torsoW/2, pos.bodyY+pos.torsoH)
	g.applyMaterialDetail(img, armorBounds, MaterialMetal, rng.Int63(), 1.0, armorColor)

	headBounds := image.Rect(cx-pos.headRadius, pos.bodyY-size/8-pos.headRadius, cx+pos.headRadius, pos.bodyY-size/8+pos.headRadius)
	g.applyMaterialDetail(img, headBounds, MaterialLeather, rng.Int63(), 0.5, skinColor)
}

// generateQuadrupedEnemy creates four-legged creature sprites.
func (g *Generator) generateQuadrupedEnemy(img *image.RGBA, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	bodyColor := g.getCreatureColor("quadruped", rng)
	darkColor := calculateDarkerColor(bodyColor)

	bodyW := size / 2
	bodyH := size / 5
	bodyY := cy - size/10

	drawQuadrupedBody(img, cx, bodyY, bodyW, bodyH, bodyColor)
	drawQuadrupedLegs(img, cx, bodyY, bodyW, bodyH, size, frame, darkColor)
	drawQuadrupedHead(img, cx, bodyY, bodyW, bodyH, size, bodyColor)
	drawQuadrupedEye(img, cx, bodyY, bodyW, size)
	drawQuadrupedTail(img, cx, bodyY, bodyW, bodyH, size, frame, darkColor)

	bodyBounds := image.Rect(cx-bodyW/2, bodyY, cx+bodyW/2, bodyY+bodyH)
	g.applyMaterialDetail(img, bodyBounds, MaterialFur, rng.Int63(), 0.8, bodyColor)

	// Apply PBR shading for realistic fur rendering
	g.ApplyPBRShadingToRegion(img, bodyBounds, MaterialFur, "cylindrical", g.lightCfg)

	// Head with spherical shading
	headW := bodyW / 3
	headH := size / 5
	headX := cx + bodyW/2
	headY := bodyY - headH/2
	headBounds := image.Rect(headX, headY, headX+headW, headY+headH)
	g.ApplyPBRShadingToRegion(img, headBounds, MaterialFur, "spherical", g.lightCfg)
}

// calculateDarkerColor creates a darker version of a color.
func calculateDarkerColor(base color.RGBA) color.RGBA {
	return color.RGBA{
		R: base.R / 2,
		G: base.G / 2,
		B: base.B / 2,
		A: 255,
	}
}

// drawQuadrupedBody draws the main body with vertical shading.
func drawQuadrupedBody(img *image.RGBA, cx, bodyY, bodyW, bodyH int, bodyColor color.RGBA) {
	for y := bodyY; y < bodyY+bodyH; y++ {
		for x := cx - bodyW/2; x < cx+bodyW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				dy := float64(y - bodyY)
				shade := 0.7 + 0.3*(1.0-dy/float64(bodyH))
				r := uint8(float64(bodyColor.R) * shade)
				g := uint8(float64(bodyColor.G) * shade)
				b := uint8(float64(bodyColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}
}

// drawQuadrupedLegs draws four animated legs.
func drawQuadrupedLegs(img *image.RGBA, cx, bodyY, bodyW, bodyH, size, frame int, darkColor color.RGBA) {
	legOffsets := calculateLegOffsets(frame)

	legPositions := [][2]int{
		{cx - bodyW/3, bodyY + bodyH},
		{cx - bodyW/6, bodyY + bodyH},
		{cx + bodyW/6, bodyY + bodyH},
		{cx + bodyW/3, bodyY + bodyH},
	}

	for i, pos := range legPositions {
		drawSingleLeg(img, pos[0], pos[1]+legOffsets[i], size, darkColor)
	}
}

// calculateLegOffsets returns leg animation offsets for the current frame.
func calculateLegOffsets(frame int) []int {
	if frame%3 == 1 {
		return []int{2, -2, -2, 2}
	} else if frame%3 == 2 {
		return []int{-2, 2, 2, -2}
	}
	return []int{0, 0, 0, 0}
}

// drawSingleLeg draws one leg at the specified position.
func drawSingleLeg(img *image.RGBA, legX, legY, size int, darkColor color.RGBA) {
	legW := size / 16
	legH := size / 4

	for y := legY; y < legY+legH; y++ {
		for x := legX - legW/2; x < legX+legW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, darkColor)
			}
		}
	}
}

// drawQuadrupedHead draws the creature's head with horizontal shading.
func drawQuadrupedHead(img *image.RGBA, cx, bodyY, bodyW, bodyH, size int, bodyColor color.RGBA) {
	headW := bodyW / 3
	headH := size / 5
	headX := cx + bodyW/2
	headY := bodyY - headH/2

	for y := headY; y < headY+headH; y++ {
		for x := headX; x < headX+headW; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				dx := float64(x - headX)
				shade := 1.0 - dx/float64(headW)*0.3
				r := uint8(float64(bodyColor.R) * shade)
				g := uint8(float64(bodyColor.G) * shade)
				b := uint8(float64(bodyColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}
}

// drawQuadrupedEye draws a single eye on the head.
func drawQuadrupedEye(img *image.RGBA, cx, bodyY, bodyW, size int) {
	headW := bodyW / 3
	headH := size / 5
	headX := cx + bodyW/2
	headY := bodyY - headH/2

	eyeColor := color.RGBA{R: 255, G: 200, B: 50, A: 255}
	common.FillCircle(img, headX+headW-3, headY+headH/3, 2, eyeColor)
}

// drawQuadrupedTail draws an animated tail.
func drawQuadrupedTail(img *image.RGBA, cx, bodyY, bodyW, bodyH, size, frame int, darkColor color.RGBA) {
	tailX := cx - bodyW/2
	tailY := bodyY + bodyH/2
	tailAngle := float64(frame%8) * math.Pi / 16
	tailLen := size / 4
	tailEndX := tailX - int(float64(tailLen)*math.Cos(tailAngle))
	tailEndY := tailY + int(float64(tailLen)*math.Sin(tailAngle))
	common.DrawThickLine(img, tailX, tailY, tailEndX, tailEndY, 2, darkColor)
}

// generateInsectEnemy creates multi-legged insect sprites.
func (g *Generator) generateInsectEnemy(img *image.RGBA, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	bodyColor := g.getCreatureColor("insect", rng)
	darkColor := createDarkColor(bodyColor)

	g.drawInsectBody(img, cx, cy, size, bodyColor)
	g.drawInsectLegs(img, cx, cy, size, frame, bodyColor, darkColor)
	g.drawInsectHead(img, cx, cy, size, frame, bodyColor, darkColor)
	g.applyInsectTexture(img, cx, cy, size, bodyColor, rng)

	// Apply PBR shading for realistic chitin rendering
	bodyBounds := image.Rect(cx-size/4, cy-size/6, cx+size/4, cy+size/3)
	g.ApplyPBRShadingToRegion(img, bodyBounds, MaterialChitin, "cylindrical", g.lightCfg)
}

// createDarkColor generates a darker variant of the given color.
func createDarkColor(base color.RGBA) color.RGBA {
	return color.RGBA{
		R: base.R / 3,
		G: base.G / 3,
		B: base.B / 3,
		A: 255,
	}
}

// drawInsectBody renders the segmented body of the insect.
func (g *Generator) drawInsectBody(img *image.RGBA, cx, cy, size int, bodyColor color.RGBA) {
	segmentCount := 3
	segmentW := size / 4
	segmentH := size / 6

	for i := 0; i < segmentCount; i++ {
		drawBodySegment(img, cx, cy, size, i, segmentW, segmentH, bodyColor)
	}
}

// drawBodySegment renders a single body segment with shading.
func drawBodySegment(img *image.RGBA, cx, cy, size, segIndex, segmentW, segmentH int, bodyColor color.RGBA) {
	segY := cy - size/6 + segIndex*segmentH
	segW := segmentW - segIndex*2
	if segW < segmentW/2 {
		segW = segmentW / 2
	}

	for y := segY; y < segY+segmentH-2; y++ {
		for x := cx - segW/2; x < cx+segW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				shadedColor := calculateShading(x, cx, segW, bodyColor)
				img.Set(x, y, shadedColor)
			}
		}
	}
}

// calculateShading applies horizontal shading based on distance from center.
func calculateShading(x, cx, segW int, baseColor color.RGBA) color.RGBA {
	dx := float64(x - cx)
	shade := 1.0 - math.Abs(dx)/float64(segW)*0.5
	return color.RGBA{
		R: uint8(float64(baseColor.R) * shade),
		G: uint8(float64(baseColor.G) * shade),
		B: uint8(float64(baseColor.B) * shade),
		A: 255,
	}
}

// drawInsectLegs renders all legs with animation.
func (g *Generator) drawInsectLegs(img *image.RGBA, cx, cy, size, frame int, bodyColor, darkColor color.RGBA) {
	legCount := 6
	segmentW := size / 4
	segmentH := size / 6

	for i := 0; i < legCount; i++ {
		drawAnimatedLeg(img, cx, cy, size, i, frame, segmentW, segmentH, darkColor)
	}
}

// drawAnimatedLeg renders a single animated insect leg.
func drawAnimatedLeg(img *image.RGBA, cx, cy, size, legIndex, frame, segmentW, segmentH int, darkColor color.RGBA) {
	side := 1
	if legIndex%2 == 0 {
		side = -1
	}
	segIdx := legIndex / 2
	legY := cy - size/6 + segIdx*segmentH + segmentH/2

	legOffset := calculateLegOffset(frame, legIndex)

	legStartX := cx + side*segmentW/2
	legMidX := legStartX + side*(size/6+legOffset)
	legMidY := legY - size/12
	legEndX := legMidX + side*size/12
	legEndY := legY + size/8

	common.DrawThickLine(img, legStartX, legY, legMidX, legMidY, 1, darkColor)
	common.DrawThickLine(img, legMidX, legMidY, legEndX, legEndY, 1, darkColor)
}

// calculateLegOffset determines leg animation offset based on frame and index.
func calculateLegOffset(frame, legIndex int) int {
	if frame%3 == 1 && legIndex%2 == 0 {
		return 2
	}
	if frame%3 == 2 && legIndex%2 == 1 {
		return 2
	}
	return 0
}

// drawInsectHead renders the head with eyes and antennae.
func (g *Generator) drawInsectHead(img *image.RGBA, cx, cy, size, frame int, bodyColor, darkColor color.RGBA) {
	headRadius := size / 8
	headY := cy - size/6 - 2
	common.FillCircle(img, cx, headY, headRadius, bodyColor)

	drawInsectEyes(img, cx, headY, headRadius)
	drawInsectAntennae(img, cx, headY, headRadius, size, frame, darkColor)
}

// drawInsectEyes renders the insect's eyes.
func drawInsectEyes(img *image.RGBA, cx, headY, headRadius int) {
	eyeColor := color.RGBA{R: 255, G: 50, B: 50, A: 255}
	common.FillCircle(img, cx-headRadius/2, headY, 2, eyeColor)
	common.FillCircle(img, cx+headRadius/2, headY, 2, eyeColor)
}

// drawInsectAntennae renders animated antennae.
func drawInsectAntennae(img *image.RGBA, cx, headY, headRadius, size, frame int, darkColor color.RGBA) {
	antennaLen := size / 5
	antennaAngle := float64(frame%8) * math.Pi / 32

	leftAntennaX := cx - headRadius/2 - int(float64(antennaLen)*math.Sin(antennaAngle))
	leftAntennaY := headY - headRadius - int(float64(antennaLen)*math.Cos(antennaAngle))
	rightAntennaX := cx + headRadius/2 + int(float64(antennaLen)*math.Sin(antennaAngle))
	rightAntennaY := headY - headRadius - int(float64(antennaLen)*math.Cos(antennaAngle))

	common.DrawThickLine(img, cx-headRadius/2, headY-headRadius, leftAntennaX, leftAntennaY, 1, darkColor)
	common.DrawThickLine(img, cx+headRadius/2, headY-headRadius, rightAntennaX, rightAntennaY, 1, darkColor)
}

// applyInsectTexture applies chitin texture to all body segments.
func (g *Generator) applyInsectTexture(img *image.RGBA, cx, cy, size int, bodyColor color.RGBA, rng *rand.Rand) {
	segmentCount := 3
	segmentW := size / 4
	segmentH := size / 6

	for i := 0; i < segmentCount; i++ {
		segY := cy - size/6 + i*segmentH
		segW := segmentW - i*2
		if segW < segmentW/2 {
			segW = segmentW / 2
		}
		segBounds := image.Rect(cx-segW/2, segY, cx+segW/2, segY+segmentH-2)
		g.applyMaterialDetail(img, segBounds, MaterialChitin, rng.Int63()+int64(i), 0.9, bodyColor)
	}
}

// generateSerpentEnemy creates snake-like creature sprites.
func (g *Generator) generateSerpentEnemy(img *image.RGBA, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	bodyColor := g.getCreatureColor("serpent", rng)
	scaleColor := calculateLighterColor(bodyColor, 1.2)

	segments := 8
	baseRadius := size / 8
	wavePhase := float64(frame) * 0.3

	drawSerpentBody(img, cx, cy, size, segments, baseRadius, wavePhase, bodyColor, scaleColor)
	drawSerpentHead(img, cx, cy, size, baseRadius, wavePhase, bodyColor, frame)

	fullBounds := image.Rect(cx-size/4, cy-size/4, cx+size/4, cy+size/2)
	g.applyMaterialDetail(img, fullBounds, MaterialScales, rng.Int63(), 1.0, bodyColor)
}

// calculateLighterColor creates a lighter version of a color.
func calculateLighterColor(base color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(math.Min(255, float64(base.R)*factor)),
		G: uint8(math.Min(255, float64(base.G)*factor)),
		B: uint8(math.Min(255, float64(base.B)*factor)),
		A: 255,
	}
}

// drawSerpentBody draws the segmented body of a serpent with wave motion.
func drawSerpentBody(img *image.RGBA, cx, cy, size, segments, baseRadius int, wavePhase float64, bodyColor, scaleColor color.RGBA) {
	for i := 0; i < segments; i++ {
		t := float64(i) / float64(segments)
		segY := cy - size/4 + int(t*float64(size)*0.8)
		waveOffset := int(math.Sin(t*math.Pi*2+wavePhase) * float64(size) / 6)
		segX := cx + waveOffset

		radius := baseRadius - i
		if radius < 2 {
			radius = 2
		}

		drawSerpentSegment(img, segX, segY, radius, bodyColor)

		if i%2 == 0 && i < segments-1 {
			common.FillCircle(img, segX-radius/2, segY, 1, scaleColor)
			common.FillCircle(img, segX+radius/2, segY, 1, scaleColor)
		}
	}
}

// drawSerpentSegment draws a single body segment with shading.
func drawSerpentSegment(img *image.RGBA, segX, segY, radius int, bodyColor color.RGBA) {
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if x*x+y*y <= radius*radius {
				px := segX + x
				py := segY + y
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					dist := math.Sqrt(float64(x*x + y*y))
					shade := 1.0 - (dist / float64(radius) * 0.4)
					r := uint8(math.Min(255, float64(bodyColor.R)*shade))
					g := uint8(math.Min(255, float64(bodyColor.G)*shade))
					b := uint8(math.Min(255, float64(bodyColor.B)*shade))
					img.Set(px, py, color.RGBA{R: r, G: g, B: b, A: 255})
				}
			}
		}
	}
}

// drawSerpentHead draws the head with eyes and animated tongue.
func drawSerpentHead(img *image.RGBA, cx, cy, size, baseRadius int, wavePhase float64, bodyColor color.RGBA, frame int) {
	headRadius := baseRadius + 2
	headY := cy - size/4
	waveOffset := int(math.Sin(wavePhase) * float64(size) / 6)
	headX := cx + waveOffset

	drawSerpentHeadShape(img, headX, headY, headRadius, bodyColor)
	drawSerpentEyes(img, headX, headY, headRadius)
	drawSerpentTongue(img, headX, headY, headRadius, size, frame)
}

// drawSerpentHeadShape renders the head as a shaded circle.
func drawSerpentHeadShape(img *image.RGBA, headX, headY, headRadius int, bodyColor color.RGBA) {
	for y := -headRadius; y <= headRadius; y++ {
		for x := -headRadius; x <= headRadius; x++ {
			if x*x+y*y <= headRadius*headRadius {
				px := headX + x
				py := headY + y
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					dist := math.Sqrt(float64(x*x + y*y))
					shade := 1.2 - (dist / float64(headRadius) * 0.5)
					r := uint8(math.Min(255, float64(bodyColor.R)*shade))
					g := uint8(math.Min(255, float64(bodyColor.G)*shade))
					b := uint8(math.Min(255, float64(bodyColor.B)*shade))
					img.Set(px, py, color.RGBA{R: r, G: g, B: b, A: 255})
				}
			}
		}
	}
}

// drawSerpentEyes renders yellow eyes on the serpent head.
func drawSerpentEyes(img *image.RGBA, headX, headY, headRadius int) {
	eyeColor := color.RGBA{R: 255, G: 200, B: 0, A: 255}
	common.FillCircle(img, headX-headRadius/2, headY, 2, eyeColor)
	common.FillCircle(img, headX+headRadius/2, headY, 2, eyeColor)
}

// drawSerpentTongue renders an animated forked tongue.
func drawSerpentTongue(img *image.RGBA, headX, headY, headRadius, size, frame int) {
	tongueLen := size / 8
	if frame%4 < 2 {
		tongueLen = size / 12
	}
	tongueColor := color.RGBA{R: 200, G: 50, B: 50, A: 255}
	common.DrawThickLine(img, headX, headY+headRadius/2, headX, headY+headRadius/2+tongueLen, 1, tongueColor)
}

// generateFlyingEnemy creates winged creature sprites.
func (g *Generator) generateFlyingEnemy(img *image.RGBA, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	bodyColor := g.getCreatureColor("flying", rng)
	wingColor := deriveWingColor(bodyColor)
	hoverOffset := calcHoverOffset(frame)

	wingH := size / 4
	if wingH == 0 {
		return
	}
	wingY := cy + hoverOffset - wingH/2
	wingAngle := calcWingAngle(frame)

	g.drawFlyingWings(img, cx, wingY, size, wingH, wingAngle, wingColor)
	g.drawFlyingBody(img, cx, cy, size, hoverOffset, bodyColor)
	g.drawFlyingHead(img, cx, cy, size, hoverOffset, bodyColor)
	g.drawFlyingTail(img, cx, cy, size, hoverOffset, frame, bodyColor)
	g.applyMaterialDetail(img, image.Rect(cx-size/2, wingY, cx+size/2, wingY+wingH), MaterialMembrane, rng.Int63(), 0.6, wingColor)
}

// deriveWingColor creates a darker wing color from body color.
func deriveWingColor(bodyColor color.RGBA) color.RGBA {
	return color.RGBA{R: bodyColor.R / 2, G: bodyColor.G / 2, B: bodyColor.B / 2, A: 200}
}

// calcHoverOffset calculates vertical hover offset based on animation frame.
func calcHoverOffset(frame int) int {
	if frame%4 < 2 {
		return -2
	}
	return 2
}

// calcWingAngle calculates wing flap angle based on animation frame.
func calcWingAngle(frame int) float64 {
	if frame%4 < 2 {
		return math.Pi / 6
	}
	return -math.Pi / 6
}

// drawFlyingWings draws both wings of a flying creature.
func (g *Generator) drawFlyingWings(img *image.RGBA, cx, wingY, size, wingH int, wingAngle float64, wingColor color.RGBA) {
	wingSpan := size / 2
	for side := -1; side <= 1; side += 2 {
		g.drawSingleWing(img, cx, wingY, size, wingH, wingSpan, wingAngle, side, wingColor)
	}
}

// drawSingleWing draws one wing of a flying creature.
func (g *Generator) drawSingleWing(img *image.RGBA, cx, wingY, size, wingH, wingSpan int, wingAngle float64, side int, wingColor color.RGBA) {
	wingCenterX := cx + side*size/8
	for y := 0; y < wingH; y++ {
		yOffset := float64(y) - float64(wingH)/2
		rotatedY := int(yOffset*math.Cos(wingAngle)) + wingY + wingH/2
		wingWidth := int(float64(wingSpan) * (1.0 - float64(y)/float64(wingH)))

		if wingWidth <= 0 {
			continue
		}
		for x := 0; x < wingWidth; x++ {
			px, py := wingCenterX+side*x, rotatedY
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				alpha := uint8(200 - uint8(float64(x)/float64(wingWidth)*150))
				img.Set(px, py, color.RGBA{R: wingColor.R, G: wingColor.G, B: wingColor.B, A: alpha})
			}
		}
	}
}

// drawFlyingBody draws the body of a flying creature.
func (g *Generator) drawFlyingBody(img *image.RGBA, cx, cy, size, hoverOffset int, bodyColor color.RGBA) {
	bodyW := size / 5
	bodyH := size / 3
	bodyY := cy + hoverOffset - bodyH/2

	for y := bodyY; y < bodyY+bodyH; y++ {
		for x := cx - bodyW/2; x < cx+bodyW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				dy := float64(y - bodyY)
				shade := 0.8 + 0.4*(1.0-math.Abs(dy-float64(bodyH)/2)/float64(bodyH))
				r := uint8(float64(bodyColor.R) * shade)
				gr := uint8(float64(bodyColor.G) * shade)
				b := uint8(float64(bodyColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: gr, B: b, A: 255})
			}
		}
	}
}

// drawFlyingHead draws the head and eyes of a flying creature.
func (g *Generator) drawFlyingHead(img *image.RGBA, cx, cy, size, hoverOffset int, bodyColor color.RGBA) {
	bodyH := size / 3
	bodyY := cy + hoverOffset - bodyH/2
	headRadius := size / 9
	headY := bodyY - 2

	common.FillCircle(img, cx, headY, headRadius, bodyColor)
	eyeColor := color.RGBA{R: 255, G: 100, B: 0, A: 255}
	common.FillCircle(img, cx-headRadius/2, headY, 2, eyeColor)
	common.FillCircle(img, cx+headRadius/2, headY, 2, eyeColor)
}

// drawFlyingTail draws the tail of a flying creature.
func (g *Generator) drawFlyingTail(img *image.RGBA, cx, cy, size, hoverOffset, frame int, bodyColor color.RGBA) {
	bodyH := size / 3
	bodyY := cy + hoverOffset - bodyH/2
	tailY := bodyY + bodyH
	tailLen := size / 5
	tailEndY := tailY + tailLen
	tailSway := int(math.Sin(float64(frame)*0.4) * 3)
	common.DrawThickLine(img, cx, tailY, cx+tailSway, tailEndY, 2, bodyColor)
}

// generateAmorphousEnemy creates blob/slime creature sprites.
func (g *Generator) generateAmorphousEnemy(img *image.RGBA, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	bodyColor := g.getCreatureColor("amorphous", rng)
	innerColor := calculateInnerColor(bodyColor)
	pulseAmount := calculatePulseAmount(frame)
	baseRadius := int(float64(size) / 3 * pulseAmount)

	radiusVariation := generateBlobShape(rng, 12)
	drawAmorphousBody(img, cx, cy, baseRadius, bodyColor, innerColor, radiusVariation)
	drawAmorphousEyes(img, cx, cy, baseRadius, rng)
	drawAmorphousHighlight(img, cx, cy, baseRadius, frame)
}

// calculateInnerColor computes a brighter inner color from the base body color.
func calculateInnerColor(bodyColor color.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8(math.Min(255, float64(bodyColor.R)*1.3)),
		G: uint8(math.Min(255, float64(bodyColor.G)*1.3)),
		B: uint8(math.Min(255, float64(bodyColor.B)*1.3)),
		A: 255,
	}
}

// calculatePulseAmount computes the pulsing animation scale factor.
func calculatePulseAmount(frame int) float64 {
	pulsePhase := float64(frame) * 0.2
	return 1.0 + math.Sin(pulsePhase)*0.15
}

// generateBlobShape creates random radius variations for blob points.
func generateBlobShape(rng *rand.Rand, blobPoints int) []float64 {
	radiusVariation := make([]float64, blobPoints)
	for i := 0; i < blobPoints; i++ {
		radiusVariation[i] = 0.8 + rng.Float64()*0.4
	}
	return radiusVariation
}

// drawAmorphousBody renders the amorphous blob body with shading.
func drawAmorphousBody(img *image.RGBA, cx, cy, baseRadius int, bodyColor, innerColor color.RGBA, radiusVariation []float64) {
	blobPoints := len(radiusVariation)

	for y := -baseRadius; y <= baseRadius; y++ {
		for x := -baseRadius; x <= baseRadius; x++ {
			angle := normalizeAngle(math.Atan2(float64(y), float64(x)))
			pointIdx := angleToPointIndex(angle, blobPoints)
			maxDist := float64(baseRadius) * radiusVariation[pointIdx]
			dist := math.Sqrt(float64(x*x + y*y))

			if dist <= maxDist {
				drawBlobPixel(img, cx+x, cy+y, dist, maxDist, bodyColor, innerColor)
			}
		}
	}
}

// normalizeAngle converts angle to 0-2π range.
func normalizeAngle(angle float64) float64 {
	if angle < 0 {
		return angle + 2*math.Pi
	}
	return angle
}

// angleToPointIndex maps an angle to a blob point index.
func angleToPointIndex(angle float64, blobPoints int) int {
	pointIdx := int(angle / (2 * math.Pi / float64(blobPoints)))
	if pointIdx >= blobPoints {
		pointIdx = blobPoints - 1
	}
	return pointIdx
}

// drawBlobPixel draws a single pixel of the blob with gradient shading.
func drawBlobPixel(img *image.RGBA, px, py int, dist, maxDist float64, bodyColor, innerColor color.RGBA) {
	if px < 0 || px >= img.Bounds().Dx() || py < 0 || py >= img.Bounds().Dy() {
		return
	}

	distRatio := dist / maxDist
	shade := 1.0 - distRatio*0.6
	c := blendInnerToBody(bodyColor, innerColor, distRatio)

	img.Set(px, py, color.RGBA{
		R: uint8(math.Min(255, float64(c.R)*shade)),
		G: uint8(math.Min(255, float64(c.G)*shade)),
		B: uint8(math.Min(255, float64(c.B)*shade)),
		A: 255,
	})
}

// blendInnerToBody blends inner color to body color based on distance ratio.
func blendInnerToBody(bodyColor, innerColor color.RGBA, distRatio float64) color.RGBA {
	if distRatio >= 0.3 {
		return bodyColor
	}

	blend := distRatio / 0.3
	return color.RGBA{
		R: uint8(float64(innerColor.R)*(1-blend) + float64(bodyColor.R)*blend),
		G: uint8(float64(innerColor.G)*(1-blend) + float64(bodyColor.G)*blend),
		B: uint8(float64(innerColor.B)*(1-blend) + float64(bodyColor.B)*blend),
		A: 255,
	}
}

// drawAmorphousEyes renders multiple eyes on the amorphous creature.
func drawAmorphousEyes(img *image.RGBA, cx, cy, baseRadius int, rng *rand.Rand) {
	eyeCount := 2 + rng.Intn(3)
	eyeColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	pupilColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	for i := 0; i < eyeCount; i++ {
		eyeAngle := float64(i) * 2 * math.Pi / float64(eyeCount)
		eyeDist := float64(baseRadius) / 3
		eyeX := cx + int(math.Cos(eyeAngle)*eyeDist)
		eyeY := cy + int(math.Sin(eyeAngle)*eyeDist) - baseRadius/4

		common.FillCircle(img, eyeX, eyeY, 4, eyeColor)
		common.FillCircle(img, eyeX, eyeY, 2, pupilColor)
	}
}

// drawAmorphousHighlight adds animated highlight to the amorphous creature.
func drawAmorphousHighlight(img *image.RGBA, cx, cy, baseRadius, frame int) {
	if frame%8 >= 4 {
		return
	}

	highlight1X := cx - baseRadius/3
	highlight1Y := cy - baseRadius/3
	common.FillCircle(img, highlight1X, highlight1Y, 3, color.RGBA{R: 255, G: 255, B: 255, A: 100})
}

// getCreatureColor returns genre and creature-type specific colors.
func (g *Generator) getCreatureColor(creatureType string, rng *rand.Rand) color.RGBA {
	switch g.genreID {
	case "scifi":
		switch creatureType {
		case "quadruped":
			return color.RGBA{R: 100, G: 120, B: 140, A: 255}
		case "insect":
			return color.RGBA{R: 80, G: 100, B: 120, A: 255}
		case "serpent":
			return color.RGBA{R: 60, G: 140, B: 160, A: 255}
		case "flying":
			return color.RGBA{R: 140, G: 120, B: 180, A: 255}
		case "amorphous":
			return color.RGBA{R: 0, G: 200, B: 150, A: 255}
		}
	case "horror":
		switch creatureType {
		case "quadruped":
			return color.RGBA{R: 80, G: 60, B: 60, A: 255}
		case "insect":
			return color.RGBA{R: 60, G: 40, B: 40, A: 255}
		case "serpent":
			return color.RGBA{R: 90, G: 80, B: 70, A: 255}
		case "flying":
			return color.RGBA{R: 70, G: 50, B: 80, A: 255}
		case "amorphous":
			return color.RGBA{R: 100, G: 50, B: 80, A: 255}
		}
	case "cyberpunk":
		switch creatureType {
		case "quadruped":
			return color.RGBA{R: 40, G: 40, B: 50, A: 255}
		case "insect":
			return color.RGBA{R: 255, G: 0, B: 128, A: 255}
		case "serpent":
			return color.RGBA{R: 0, G: 255, B: 200, A: 255}
		case "flying":
			return color.RGBA{R: 255, G: 100, B: 200, A: 255}
		case "amorphous":
			return color.RGBA{R: 128, G: 0, B: 255, A: 255}
		}
	case "postapoc":
		switch creatureType {
		case "quadruped":
			return color.RGBA{R: 120, G: 100, B: 80, A: 255}
		case "insect":
			return color.RGBA{R: 100, G: 80, B: 60, A: 255}
		case "serpent":
			return color.RGBA{R: 110, G: 90, B: 70, A: 255}
		case "flying":
			return color.RGBA{R: 90, G: 80, B: 70, A: 255}
		case "amorphous":
			return color.RGBA{R: 140, G: 100, B: 60, A: 255}
		}
	default:
		switch creatureType {
		case "quadruped":
			return color.RGBA{R: 140, G: 100, B: 60, A: 255}
		case "insect":
			return color.RGBA{R: 80, G: 120, B: 80, A: 255}
		case "serpent":
			return color.RGBA{R: 100, G: 140, B: 100, A: 255}
		case "flying":
			return color.RGBA{R: 120, G: 100, B: 140, A: 255}
		case "amorphous":
			return color.RGBA{R: 100, G: 200, B: 150, A: 255}
		}
	}
	return color.RGBA{R: 128, G: 128, B: 128, A: 255}
}

// generateDefaultSprite creates a fallback sprite.
func (g *Generator) generateDefaultSprite(img *image.RGBA, rng *rand.Rand) {
	size := img.Bounds().Dx()
	common.FillRect(img, 0, 0, size, size, color.RGBA{R: 128, G: 128, B: 128, A: 255})
}

// Genre-specific color helpers.

func (g *Generator) getGenreWoodColor() color.RGBA {
	switch g.genreID {
	case "scifi":
		return color.RGBA{R: 100, G: 100, B: 110, A: 255}
	case "horror":
		return color.RGBA{R: 80, G: 60, B: 50, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 40, G: 40, B: 50, A: 255}
	case "postapoc":
		return color.RGBA{R: 110, G: 90, B: 70, A: 255}
	default:
		return color.RGBA{R: 139, G: 90, B: 43, A: 255}
	}
}

func (g *Generator) getGenreLeafColor() color.RGBA {
	switch g.genreID {
	case "scifi":
		return color.RGBA{R: 0, G: 255, B: 150, A: 255}
	case "horror":
		return color.RGBA{R: 60, G: 80, B: 50, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 255, G: 0, B: 128, A: 255}
	default:
		return color.RGBA{R: 50, G: 150, B: 50, A: 255}
	}
}

func (g *Generator) getGenreStoneColor() color.RGBA {
	switch g.genreID {
	case "scifi":
		return color.RGBA{R: 140, G: 150, B: 160, A: 255}
	case "horror":
		return color.RGBA{R: 80, G: 80, B: 90, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 60, G: 70, B: 80, A: 255}
	default:
		return color.RGBA{R: 120, G: 120, B: 130, A: 255}
	}
}
