// Package raycaster implements the core raycasting engine.
package raycaster

import (
	"math"
	"sort"
)

// Raycaster performs raycasting against a 2D map.
type Raycaster struct {
	FOV        float64
	Width      int
	Height     int
	Map        [][]int    // 2D tile grid; 0 = empty, >0 = wall type
	FogColor   [3]float64 // RGB fog color (0.0-1.0)
	FogDensity float64    // Fog density for exponential falloff
}

// NewRaycaster creates a raycaster with the given field of view and resolution.
func NewRaycaster(fov float64, width, height int) *Raycaster {
	return &Raycaster{
		FOV:        fov,
		Width:      width,
		Height:     height,
		FogColor:   [3]float64{0.0, 0.0, 0.0}, // Default black fog
		FogDensity: 0.05,                      // Default density
	}
}

// SetMap assigns the tile grid for raycasting.
func (r *Raycaster) SetMap(tileMap [][]int) {
	r.Map = tileMap
}

// RayHit contains information about a ray-wall intersection.
type RayHit struct {
	Distance float64 // Perpendicular distance to wall
	WallType int     // Tile value from map
	Side     int     // 0 = horizontal, 1 = vertical wall face
	HitX     float64 // Exact X coordinate of wall hit
	HitY     float64 // Exact Y coordinate of wall hit
	TextureX float64 // Texture coordinate along wall (0.0-1.0)
}

// CastRays casts all rays for a single frame using DDA algorithm.
// Returns per-column wall distances and hit information.
func (r *Raycaster) CastRays(posX, posY, dirX, dirY float64) []RayHit {
	hits := make([]RayHit, r.Width)

	// Camera plane perpendicular to direction vector
	planeX := -dirY * Tan(r.FOV*math.Pi/360.0)
	planeY := dirX * Tan(r.FOV*math.Pi/360.0)

	for x := 0; x < r.Width; x++ {
		// Camera X coordinate in [-1, 1]
		cameraX := 2.0*float64(x)/float64(r.Width) - 1.0

		// Ray direction
		rayDirX := dirX + planeX*cameraX
		rayDirY := dirY + planeY*cameraX

		hits[x] = r.castRay(posX, posY, rayDirX, rayDirY)
	}

	return hits
}

// castRay performs DDA against the map grid for a single ray.
func (r *Raycaster) castRay(posX, posY, rayDirX, rayDirY float64) RayHit {
	if r.Map == nil || len(r.Map) == 0 || len(r.Map[0]) == 0 {
		return RayHit{Distance: 1e30, WallType: 1, Side: 0}
	}

	mapX := int(posX)
	mapY := int(posY)

	deltaDistX, deltaDistY := calculateDeltaDistances(rayDirX, rayDirY)
	stepX, stepY, sideDistX, sideDistY := initializeDDA(posX, posY, rayDirX, rayDirY, mapX, mapY, deltaDistX, deltaDistY)

	side, hit := performDDA(&mapX, &mapY, &sideDistX, &sideDistY, deltaDistX, deltaDistY, stepX, stepY, r.Map)
	if !hit {
		return RayHit{Distance: 1e30, WallType: 0, Side: side}
	}

	perpWallDist, hitX, hitY := calculateWallDistance(side, mapX, mapY, posX, posY, rayDirX, rayDirY, stepX, stepY)
	textureX := calculateTextureCoordinate(side, hitX, hitY)

	return RayHit{
		Distance: math.Abs(perpWallDist),
		WallType: r.Map[mapY][mapX],
		Side:     side,
		HitX:     hitX,
		HitY:     hitY,
		TextureX: textureX,
	}
}

// calculateDeltaDistances computes the distance the ray travels between grid lines.
func calculateDeltaDistances(rayDirX, rayDirY float64) (float64, float64) {
	var deltaDistX, deltaDistY float64
	if rayDirX == 0 {
		deltaDistX = 1e30
	} else {
		deltaDistX = math.Abs(1.0 / rayDirX)
	}
	if rayDirY == 0 {
		deltaDistY = 1e30
	} else {
		deltaDistY = math.Abs(1.0 / rayDirY)
	}
	return deltaDistX, deltaDistY
}

// initializeDDA sets up initial values for the DDA algorithm.
func initializeDDA(posX, posY, rayDirX, rayDirY float64, mapX, mapY int, deltaDistX, deltaDistY float64) (int, int, float64, float64) {
	var stepX, stepY int
	var sideDistX, sideDistY float64

	if rayDirX < 0 {
		stepX = -1
		sideDistX = (posX - float64(mapX)) * deltaDistX
	} else {
		stepX = 1
		sideDistX = (float64(mapX+1) - posX) * deltaDistX
	}

	if rayDirY < 0 {
		stepY = -1
		sideDistY = (posY - float64(mapY)) * deltaDistY
	} else {
		stepY = 1
		sideDistY = (float64(mapY+1) - posY) * deltaDistY
	}

	return stepX, stepY, sideDistX, sideDistY
}

// performDDA executes the DDA algorithm to find wall intersections.
func performDDA(mapX, mapY *int, sideDistX, sideDistY *float64, deltaDistX, deltaDistY float64, stepX, stepY int, tileMap [][]int) (int, bool) {
	var side int
	const maxDepth = 100

	for depth := 0; depth < maxDepth; depth++ {
		if *sideDistX < *sideDistY {
			*sideDistX += deltaDistX
			*mapX += stepX
			side = 0
		} else {
			*sideDistY += deltaDistY
			*mapY += stepY
			side = 1
		}

		if *mapX < 0 || *mapY < 0 || *mapY >= len(tileMap) || *mapX >= len(tileMap[0]) {
			return side, false
		}

		if IsWallTile(tileMap[*mapY][*mapX]) {
			return side, true
		}
	}

	return side, false
}

// IsWallTile returns true if a tile value represents a solid wall that should
// stop rays and block line of sight. Floor tiles (0, 2, 20-29) are not walls.
// Wall tiles (1, 3=door, 4=secret, 10-14=genre walls) are solid.
func IsWallTile(tile int) bool {
	if tile == 0 || tile == 2 {
		return false
	}
	if tile >= 20 && tile <= 29 {
		return false
	}
	return tile > 0
}

// calculateWallDistance computes the perpendicular distance to the wall and hit coordinates.
func calculateWallDistance(side, mapX, mapY int, posX, posY, rayDirX, rayDirY float64, stepX, stepY int) (float64, float64, float64) {
	var perpWallDist, hitX, hitY float64

	if side == 0 {
		perpWallDist = (float64(mapX) - posX + (1.0-float64(stepX))/2.0) / rayDirX
		if stepX > 0 {
			hitX = float64(mapX)
		} else {
			hitX = float64(mapX + 1)
		}
		hitY = posY + perpWallDist*rayDirY
	} else {
		perpWallDist = (float64(mapY) - posY + (1.0-float64(stepY))/2.0) / rayDirY
		if stepY > 0 {
			hitY = float64(mapY)
		} else {
			hitY = float64(mapY + 1)
		}
		hitX = posX + perpWallDist*rayDirX
	}

	return perpWallDist, hitX, hitY
}

// calculateTextureCoordinate computes the texture coordinate along the wall.
func calculateTextureCoordinate(side int, hitX, hitY float64) float64 {
	if side == 0 {
		return hitY - math.Floor(hitY)
	}
	return hitX - math.Floor(hitX)
}

// FloorCeilPixel contains floor/ceiling pixel information.
type FloorCeilPixel struct {
	WorldX   float64 // World X coordinate
	WorldY   float64 // World Y coordinate
	Distance float64 // Distance from camera
	IsFloor  bool    // true = floor, false = ceiling
}

// Sprite represents a billboard sprite in world space.
type Sprite struct {
	X      float64 // World X position
	Y      float64 // World Y position
	Type   int     // Sprite type/texture ID
	Width  float64 // Display width (in world units)
	Height float64 // Display height (in world units)
}

// SpriteHit contains information about a sprite's screen projection.
type SpriteHit struct {
	ScreenX      int     // Center X position on screen
	ScreenY      int     // Center Y position on screen
	ScreenWidth  int     // Width in pixels
	ScreenHeight int     // Height in pixels
	Distance     float64 // Distance from camera
	Type         int     // Sprite type
	DrawStartX   int     // Left edge clipping
	DrawEndX     int     // Right edge clipping
	DrawStartY   int     // Top edge clipping
	DrawEndY     int     // Bottom edge clipping
}

// CastFloorCeiling casts floor and ceiling for a single screen row.
// Returns per-column floor/ceiling pixel world coordinates.
func (r *Raycaster) CastFloorCeiling(row int, posX, posY, dirX, dirY, pitch float64) []FloorCeilPixel {
	pixels := make([]FloorCeilPixel, r.Width)

	// Camera plane perpendicular to direction vector
	planeX := -dirY * Tan(r.FOV*math.Pi/360.0)
	planeY := dirX * Tan(r.FOV*math.Pi/360.0)

	// Determine if this row is floor or ceiling
	isFloor := row > r.Height/2

	// Calculate distance from horizon (required for perspective division)
	p := row - r.Height/2
	// Guard against division by zero at horizon line
	if p == 0 {
		// At horizon - use far plane distance for consistent fog/lighting
		const farPlane = 1e10
		for x := 0; x < r.Width; x++ {
			pixels[x] = FloorCeilPixel{
				WorldX:   posX,
				WorldY:   posY,
				Distance: farPlane,
				IsFloor:  isFloor,
			}
		}
		return pixels
	}

	// Camera height (1.0 = eye level, adjusted by pitch)
	cameraZ := 0.5 * float64(r.Height)
	pitchOffset := pitch * float64(r.Height) / 2.0

	// Vertical position of the row on screen (safe: p != 0)
	rowDistance := (cameraZ + pitchOffset) / float64(p)
	if rowDistance < 0 {
		rowDistance = -rowDistance
	}

	// Calculate step in world space for each pixel
	floorStepX := rowDistance * (planeX * 2.0) / float64(r.Width)
	floorStepY := rowDistance * (planeY * 2.0) / float64(r.Width)

	// Starting position for leftmost pixel
	floorX := posX + rowDistance*dirX - rowDistance*planeX
	floorY := posY + rowDistance*dirY - rowDistance*planeY

	for x := 0; x < r.Width; x++ {
		pixels[x] = FloorCeilPixel{
			WorldX:   floorX,
			WorldY:   floorY,
			Distance: rowDistance,
			IsFloor:  isFloor,
		}

		floorX += floorStepX
		floorY += floorStepY
	}

	return pixels
}

// CastSprites projects sprites onto screen with depth sorting and occlusion.
// wallDistances is the output from CastRays() for depth comparison.
func (r *Raycaster) CastSprites(sprites []Sprite, posX, posY, dirX, dirY float64, wallDistances []RayHit) []SpriteHit {
	if len(sprites) == 0 {
		return nil
	}

	planeX, planeY := calculateCameraPlane(dirX, dirY, r.FOV)
	spriteList := prepareSpriteList(sprites, posX, posY)
	sortSpritesByDistance(spriteList)

	return projectSprites(spriteList, posX, posY, dirX, dirY, planeX, planeY, wallDistances, r.Width, r.Height)
}

// calculateCameraPlane computes the camera plane perpendicular to direction.
func calculateCameraPlane(dirX, dirY, fov float64) (float64, float64) {
	planeX := -dirY * Tan(fov*math.Pi/360.0)
	planeY := dirX * Tan(fov*math.Pi/360.0)
	return planeX, planeY
}

// prepareSpriteList transforms sprites to camera space with distances.
func prepareSpriteList(sprites []Sprite, posX, posY float64) []spriteData {
	spriteList := make([]spriteData, 0, len(sprites))
	for i, spr := range sprites {
		dx := spr.X - posX
		dy := spr.Y - posY
		dist := math.Sqrt(dx*dx + dy*dy)
		spriteList = append(spriteList, spriteData{sprite: spr, distance: dist, index: i})
	}
	return spriteList
}

// sortSpritesByDistance sorts sprites from farthest to nearest.
func sortSpritesByDistance(spriteList []spriteData) {
	sort.Slice(spriteList, func(i, j int) bool {
		return spriteList[i].distance > spriteList[j].distance
	})
}

// projectSprites projects sprites onto the screen.
func projectSprites(spriteList []spriteData, posX, posY, dirX, dirY, planeX, planeY float64, wallDistances []RayHit, width, height int) []SpriteHit {
	hits := make([]SpriteHit, 0, len(spriteList))

	for _, sd := range spriteList {
		hit := projectSingleSprite(sd.sprite, posX, posY, dirX, dirY, planeX, planeY, wallDistances, width, height)
		if hit != nil {
			hits = append(hits, *hit)
		}
	}

	return hits
}

// projectSingleSprite projects a single sprite onto screen space.
func projectSingleSprite(spr Sprite, posX, posY, dirX, dirY, planeX, planeY float64, wallDistances []RayHit, width, height int) *SpriteHit {
	spriteX := spr.X - posX
	spriteY := spr.Y - posY

	invDet := 1.0 / (planeX*dirY - dirX*planeY)
	transformX := invDet * (dirY*spriteX - dirX*spriteY)
	transformY := invDet * (-planeY*spriteX + planeX*spriteY)

	if transformY <= 0 {
		return nil
	}

	spriteScreenX := int(float64(width) / 2.0 * (1.0 + transformX/transformY))
	spriteHeight := int(math.Abs(float64(height) / transformY * spr.Height))
	spriteWidth := int(math.Abs(float64(height) / transformY * spr.Width))

	drawStartX, drawEndX, drawStartY, drawEndY := calculateSpriteBounds(spriteScreenX, spriteWidth, spriteHeight, width, height)

	if drawStartX >= width || drawEndX < 0 {
		return nil
	}

	if isSpriteOccluded(drawStartX, drawEndX, transformY, wallDistances) {
		return nil
	}

	return &SpriteHit{
		ScreenX:      spriteScreenX,
		ScreenY:      height / 2,
		ScreenWidth:  spriteWidth,
		ScreenHeight: spriteHeight,
		Distance:     transformY,
		Type:         spr.Type,
		DrawStartX:   drawStartX,
		DrawEndX:     drawEndX,
		DrawStartY:   drawStartY,
		DrawEndY:     drawEndY,
	}
}

// calculateSpriteBounds computes and clips sprite drawing bounds.
func calculateSpriteBounds(spriteScreenX, spriteWidth, spriteHeight, screenWidth, screenHeight int) (int, int, int, int) {
	drawStartY := -spriteHeight/2 + screenHeight/2
	drawEndY := spriteHeight/2 + screenHeight/2
	drawStartX := -spriteWidth/2 + spriteScreenX
	drawEndX := spriteWidth/2 + spriteScreenX

	if drawStartX < 0 {
		drawStartX = 0
	}
	if drawEndX >= screenWidth {
		drawEndX = screenWidth - 1
	}
	if drawStartY < 0 {
		drawStartY = 0
	}
	if drawEndY >= screenHeight {
		drawEndY = screenHeight - 1
	}

	return drawStartX, drawEndX, drawStartY, drawEndY
}

// isSpriteOccluded checks if a sprite is occluded by walls.
func isSpriteOccluded(drawStartX, drawEndX int, transformY float64, wallDistances []RayHit) bool {
	for x := drawStartX; x <= drawEndX && x < len(wallDistances); x++ {
		if transformY < wallDistances[x].Distance {
			return false
		}
	}
	return true
}

type spriteData struct {
	sprite   Sprite
	distance float64
	index    int
}

// ApplyFog applies exponential fog to a color based on distance.
// baseColor is RGB in range [0.0, 1.0], returns fogged RGB.
func (r *Raycaster) ApplyFog(baseColor [3]float64, distance float64) [3]float64 {
	// Exponential fog factor: e^(-density * distance)
	fogFactor := math.Exp(-r.FogDensity * distance)

	// Clamp fog factor to [0, 1]
	if fogFactor < 0 {
		fogFactor = 0
	} else if fogFactor > 1 {
		fogFactor = 1
	}

	// Linear interpolation between base color and fog color
	return [3]float64{
		baseColor[0]*fogFactor + r.FogColor[0]*(1-fogFactor),
		baseColor[1]*fogFactor + r.FogColor[1]*(1-fogFactor),
		baseColor[2]*fogFactor + r.FogColor[2]*(1-fogFactor),
	}
}

// SetGenre configures raycaster parameters for a genre.
func (r *Raycaster) SetGenre(genreID string) {
	// Genre-specific fog colors (RGB in 0.0-1.0 range)
	switch genreID {
	case "fantasy":
		r.FogColor = [3]float64{0.1, 0.05, 0.15} // Purple-ish
		r.FogDensity = 0.06
	case "scifi":
		r.FogColor = [3]float64{0.0, 0.1, 0.15} // Blue-ish
		r.FogDensity = 0.04
	case "horror":
		r.FogColor = [3]float64{0.05, 0.0, 0.0} // Dark red
		r.FogDensity = 0.08
	case "cyberpunk":
		r.FogColor = [3]float64{0.15, 0.0, 0.15} // Magenta
		r.FogDensity = 0.05
	case "postapoc":
		r.FogColor = [3]float64{0.1, 0.08, 0.05} // Brown-ish
		r.FogDensity = 0.07
	default:
		r.FogColor = [3]float64{0.0, 0.0, 0.0} // Black
		r.FogDensity = 0.05
	}
}
