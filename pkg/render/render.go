// Package render provides the rendering pipeline for drawing frames.
package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/raycaster"
)

// TextureAtlas is an interface for texture retrieval.
// Allows testing with mocks while supporting the full texture.Atlas.
type TextureAtlas interface {
	Get(name string) (image.Image, bool)
	SetGenre(genreID string)
}

// LightMap is an interface for per-tile lighting data.
// Allows testing with mocks while supporting the full lighting.SectorLightMap.
type LightMap interface {
	GetLight(x, y int) float64
	Calculate()
}

// Renderer manages the rendering pipeline.
type Renderer struct {
	Width         int
	Height        int
	framebuffer   []byte
	raycaster     *raycaster.Raycaster
	palette       map[int]color.RGBA
	genreID       string
	atlas         TextureAtlas
	lightMap      LightMap
	postProcessor *PostProcessor
}

// NewRenderer creates a renderer with the given internal resolution.
func NewRenderer(width, height int, rc *raycaster.Raycaster) *Renderer {
	return &Renderer{
		Width:         width,
		Height:        height,
		framebuffer:   make([]byte, width*height*4),
		raycaster:     rc,
		palette:       getDefaultPalette(),
		genreID:       "fantasy",
		atlas:         nil, // Optional texture atlas
		lightMap:      nil, // Optional lighting map
		postProcessor: nil, // Optional post-processor
	}
}

// SetTextureAtlas assigns a texture atlas for textured rendering.
func (r *Renderer) SetTextureAtlas(atlas TextureAtlas) {
	r.atlas = atlas
}

// SetLightMap assigns a light map for dynamic lighting.
func (r *Renderer) SetLightMap(lightMap LightMap) {
	r.lightMap = lightMap
}

// SetPostProcessor assigns a post-processor for visual effects.
func (r *Renderer) SetPostProcessor(pp *PostProcessor) {
	r.postProcessor = pp
}

// Render draws a frame to the given screen image.
// Calls raycaster, writes column data to framebuffer, blits to screen.
func (r *Renderer) Render(screen *ebiten.Image, posX, posY, dirX, dirY, pitch float64) {
	hits := r.raycaster.CastRays(posX, posY, dirX, dirY)

	for y := 0; y < r.Height; y++ {
		for x := 0; x < r.Width; x++ {
			idx := (y*r.Width + x) * 4
			var c color.RGBA

			if y < r.Height/2 {
				c = r.renderCeiling(x, y, posX, posY, dirX, dirY, pitch)
			} else if y > r.Height/2 {
				c = r.renderFloor(x, y, posX, posY, dirX, dirY, pitch)
			} else {
				c = r.palette[0]
			}

			if x < len(hits) {
				wallColor := r.renderWall(x, y, hits[x])
				if wallColor.A > 0 {
					c = wallColor
				}
			}

			r.framebuffer[idx] = c.R
			r.framebuffer[idx+1] = c.G
			r.framebuffer[idx+2] = c.B
			r.framebuffer[idx+3] = c.A
		}
	}

	// Apply post-processing if enabled
	if r.postProcessor != nil {
		r.postProcessor.Apply(r.framebuffer)
	}

	img := ebiten.NewImageFromImageWithOptions(
		&frameImage{data: r.framebuffer, width: r.Width, height: r.Height},
		&ebiten.NewImageFromImageOptions{Unmanaged: true},
	)
	screen.DrawImage(img, nil)
}

// renderWall computes wall color for a given column and row.
func (r *Renderer) renderWall(x, y int, hit raycaster.RayHit) color.RGBA {
	if hit.Distance >= 1e30 || hit.WallType == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	lineHeight := int(float64(r.Height) / hit.Distance)
	drawStart := -lineHeight/2 + r.Height/2
	drawEnd := lineHeight/2 + r.Height/2

	if y < drawStart || y > drawEnd {
		return color.RGBA{0, 0, 0, 0}
	}

	baseColor := r.palette[hit.WallType]
	if hit.Side == 1 {
		baseColor.R = baseColor.R / 2
		baseColor.G = baseColor.G / 2
		baseColor.B = baseColor.B / 2
	}

	// Apply lighting if available
	lightMult := r.getLightMultiplier(hit.HitX, hit.HitY)

	foggedColor := r.raycaster.ApplyFog(
		[3]float64{
			float64(baseColor.R) / 255.0 * lightMult,
			float64(baseColor.G) / 255.0 * lightMult,
			float64(baseColor.B) / 255.0 * lightMult,
		},
		hit.Distance,
	)

	return color.RGBA{
		R: uint8(foggedColor[0] * 255),
		G: uint8(foggedColor[1] * 255),
		B: uint8(foggedColor[2] * 255),
		A: 255,
	}
}

// renderFloor computes floor color for a given pixel.
// If atlas is set, samples floor texture with perspective-correct coordinates.
func (r *Renderer) renderFloor(x, y int, posX, posY, dirX, dirY, pitch float64) color.RGBA {
	pixels := r.raycaster.CastFloorCeiling(y, posX, posY, dirX, dirY, pitch)
	if x >= len(pixels) {
		return r.palette[0]
	}

	var baseColor color.RGBA

	// Try to sample from texture atlas if available
	if r.atlas != nil {
		floorTex, hasFloor := r.atlas.Get("floor_main")
		if hasFloor {
			baseColor = r.sampleTexture(floorTex, pixels[x].WorldX, pixels[x].WorldY)
		} else {
			baseColor = r.palette[2]
		}
	} else {
		baseColor = r.palette[2]
	}

	// Apply lighting if available
	lightMult := r.getLightMultiplier(pixels[x].WorldX, pixels[x].WorldY)

	foggedColor := r.raycaster.ApplyFog(
		[3]float64{
			float64(baseColor.R) / 255.0 * lightMult,
			float64(baseColor.G) / 255.0 * lightMult,
			float64(baseColor.B) / 255.0 * lightMult,
		},
		pixels[x].Distance,
	)

	return color.RGBA{
		R: uint8(foggedColor[0] * 255),
		G: uint8(foggedColor[1] * 255),
		B: uint8(foggedColor[2] * 255),
		A: 255,
	}
}

// renderCeiling computes ceiling color for a given pixel.
// If atlas is set, samples ceiling texture with perspective-correct coordinates.
func (r *Renderer) renderCeiling(x, y int, posX, posY, dirX, dirY, pitch float64) color.RGBA {
	pixels := r.raycaster.CastFloorCeiling(r.Height-1-y, posX, posY, dirX, dirY, pitch)
	if x >= len(pixels) {
		return r.palette[0]
	}

	var baseColor color.RGBA

	// Try to sample from texture atlas if available
	if r.atlas != nil {
		ceilingTex, hasCeiling := r.atlas.Get("ceiling_main")
		if hasCeiling {
			baseColor = r.sampleTexture(ceilingTex, pixels[x].WorldX, pixels[x].WorldY)
		} else {
			baseColor = r.palette[3]
		}
	} else {
		baseColor = r.palette[3]
	}

	// Apply lighting if available
	lightMult := r.getLightMultiplier(pixels[x].WorldX, pixels[x].WorldY)

	foggedColor := r.raycaster.ApplyFog(
		[3]float64{
			float64(baseColor.R) / 255.0 * lightMult,
			float64(baseColor.G) / 255.0 * lightMult,
			float64(baseColor.B) / 255.0 * lightMult,
		},
		pixels[x].Distance,
	)

	return color.RGBA{
		R: uint8(foggedColor[0] * 255),
		G: uint8(foggedColor[1] * 255),
		B: uint8(foggedColor[2] * 255),
		A: 255,
	}
}

// sampleTexture samples a texture at world coordinates with wrapping.
// Uses perspective-correct texture mapping for floor/ceiling.
func (r *Renderer) sampleTexture(tex image.Image, worldX, worldY float64) color.RGBA {
	bounds := tex.Bounds()
	texWidth := bounds.Dx()
	texHeight := bounds.Dy()

	// Convert world coordinates to texture coordinates with wrapping
	// Each world tile maps to one texture repeat
	texX := int(worldX*float64(texWidth)) % texWidth
	texY := int(worldY*float64(texHeight)) % texHeight

	// Handle negative wrapping
	if texX < 0 {
		texX += texWidth
	}
	if texY < 0 {
		texY += texHeight
	}

	// Clamp to texture bounds for safety
	if texX >= texWidth {
		texX = texWidth - 1
	}
	if texY >= texHeight {
		texY = texHeight - 1
	}
	if texX < 0 {
		texX = 0
	}
	if texY < 0 {
		texY = 0
	}

	c := tex.At(texX, texY)
	cr, cg, cb, ca := c.RGBA()

	return color.RGBA{
		R: uint8(cr >> 8),
		G: uint8(cg >> 8),
		B: uint8(cb >> 8),
		A: uint8(ca >> 8),
	}
}

// getLightMultiplier returns the lighting multiplier at world coordinates.
// Returns 1.0 (full brightness) if no light map is set.
func (r *Renderer) getLightMultiplier(worldX, worldY float64) float64 {
	if r.lightMap == nil {
		return 1.0
	}

	// Convert world coordinates to tile coordinates
	tileX := int(worldX)
	tileY := int(worldY)

	return r.lightMap.GetLight(tileX, tileY)
}

// SetGenre configures the renderer for a genre.
func (r *Renderer) SetGenre(genreID string) {
	r.genreID = genreID
	r.palette = getPaletteForGenre(genreID)
	r.raycaster.SetGenre(genreID)
	if r.postProcessor != nil {
		r.postProcessor.SetGenre(genreID)
	}
}

// getDefaultPalette returns the default color palette.
func getDefaultPalette() map[int]color.RGBA {
	return getPaletteForGenre("fantasy")
}

// getPaletteForGenre returns genre-specific color palette.
func getPaletteForGenre(genreID string) map[int]color.RGBA {
	switch genreID {
	case "fantasy":
		return map[int]color.RGBA{
			0: {20, 15, 30, 255},   // Sky/background
			1: {100, 80, 60, 255},  // Stone wall
			2: {40, 35, 30, 255},   // Floor
			3: {30, 25, 35, 255},   // Ceiling
			4: {120, 100, 80, 255}, // Alternate wall
		}
	case "scifi":
		return map[int]color.RGBA{
			0: {10, 15, 25, 255},    // Sky/background
			1: {80, 90, 100, 255},   // Metal hull
			2: {30, 35, 40, 255},    // Floor
			3: {25, 30, 35, 255},    // Ceiling
			4: {100, 110, 120, 255}, // Alternate wall
		}
	case "horror":
		return map[int]color.RGBA{
			0: {15, 5, 5, 255},    // Sky/background
			1: {80, 60, 50, 255},  // Decayed plaster
			2: {30, 20, 15, 255},  // Floor
			3: {25, 15, 10, 255},  // Ceiling
			4: {100, 70, 60, 255}, // Alternate wall
		}
	case "cyberpunk":
		return map[int]color.RGBA{
			0: {20, 10, 25, 255},   // Sky/background
			1: {90, 70, 100, 255},  // Neon-lit concrete
			2: {35, 30, 40, 255},   // Floor
			3: {30, 25, 35, 255},   // Ceiling
			4: {110, 80, 120, 255}, // Alternate wall
		}
	case "postapoc":
		return map[int]color.RGBA{
			0: {25, 20, 15, 255},  // Sky/background
			1: {100, 80, 60, 255}, // Rusted metal
			2: {40, 30, 25, 255},  // Floor
			3: {35, 25, 20, 255},  // Ceiling
			4: {120, 90, 70, 255}, // Alternate wall
		}
	default:
		return map[int]color.RGBA{
			0: {0, 0, 0, 255},
			1: {128, 128, 128, 255},
			2: {64, 64, 64, 255},
			3: {96, 96, 96, 255},
			4: {160, 160, 160, 255},
		}
	}
}

// frameImage implements image.Image for framebuffer blitting.
type frameImage struct {
	data   []byte
	width  int
	height int
}

func (f *frameImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (f *frameImage) Bounds() image.Rectangle {
	return image.Rect(0, 0, f.width, f.height)
}

func (f *frameImage) At(x, y int) color.Color {
	if x < 0 || x >= f.width || y < 0 || y >= f.height {
		return color.RGBA{0, 0, 0, 255}
	}
	idx := (y*f.width + x) * 4
	return color.RGBA{
		R: f.data[idx],
		G: f.data[idx+1],
		B: f.data[idx+2],
		A: f.data[idx+3],
	}
}
