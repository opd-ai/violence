// Package raycaster implements the core raycasting engine.
package raycaster

import "math"

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
}

// CastRays casts all rays for a single frame using DDA algorithm.
// Returns per-column wall distances and hit information.
func (r *Raycaster) CastRays(posX, posY, dirX, dirY float64) []RayHit {
	hits := make([]RayHit, r.Width)

	// Camera plane perpendicular to direction vector
	planeX := -dirY * math.Tan(r.FOV*math.Pi/360.0)
	planeY := dirX * math.Tan(r.FOV*math.Pi/360.0)

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
	// Current map cell
	mapX := int(posX)
	mapY := int(posY)

	// Length of ray from one x or y-side to next x or y-side
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

	// Step direction and initial sideDist
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

	// DDA algorithm
	var side int // 0 = X-side, 1 = Y-side
	var hit bool
	const maxDepth = 100

	for depth := 0; depth < maxDepth && !hit; depth++ {
		// Jump to next map square
		if sideDistX < sideDistY {
			sideDistX += deltaDistX
			mapX += stepX
			side = 0
		} else {
			sideDistY += deltaDistY
			mapY += stepY
			side = 1
		}

		// Check if ray hit a wall
		if mapX < 0 || mapY < 0 || mapY >= len(r.Map) || mapX >= len(r.Map[0]) {
			// Out of bounds = wall
			return RayHit{Distance: 1e30, WallType: 1, Side: side}
		}

		if r.Map[mapY][mapX] > 0 {
			hit = true
		}
	}

	if !hit {
		return RayHit{Distance: 1e30, WallType: 0, Side: side}
	}

	// Calculate perpendicular distance to avoid fisheye effect
	var perpWallDist float64
	var hitX, hitY float64

	if side == 0 {
		// Hit on X-side (vertical wall)
		perpWallDist = (float64(mapX) - posX + (1.0-float64(stepX))/2.0) / rayDirX
		if stepX > 0 {
			hitX = float64(mapX)
		} else {
			hitX = float64(mapX + 1)
		}
		hitY = posY + perpWallDist*rayDirY
	} else {
		// Hit on Y-side (horizontal wall)
		perpWallDist = (float64(mapY) - posY + (1.0-float64(stepY))/2.0) / rayDirY
		if stepY > 0 {
			hitY = float64(mapY)
		} else {
			hitY = float64(mapY + 1)
		}
		hitX = posX + perpWallDist*rayDirX
	}

	return RayHit{
		Distance: math.Abs(perpWallDist),
		WallType: r.Map[mapY][mapX],
		Side:     side,
		HitX:     hitX,
		HitY:     hitY,
	}
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
	planeX := -dirY * math.Tan(r.FOV*math.Pi/360.0)
	planeY := dirX * math.Tan(r.FOV*math.Pi/360.0)

	// Determine if this row is floor or ceiling
	isFloor := row > r.Height/2

	// Distance from horizon (p must be non-zero)
	p := row - r.Height/2
	if p == 0 {
		// At horizon - return infinite distance
		for x := 0; x < r.Width; x++ {
			pixels[x] = FloorCeilPixel{
				WorldX:   posX,
				WorldY:   posY,
				Distance: 1e30,
				IsFloor:  isFloor,
			}
		}
		return pixels
	}

	// Camera height (1.0 = eye level, adjusted by pitch)
	cameraZ := 0.5 * float64(r.Height)
	pitchOffset := pitch * float64(r.Height) / 2.0

	// Vertical position of the row on screen
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

	// Camera plane perpendicular to direction vector
	planeX := -dirY * math.Tan(r.FOV*math.Pi/360.0)
	planeY := dirX * math.Tan(r.FOV*math.Pi/360.0)

	// Transform sprites to camera space and calculate distance
	type spriteData struct {
		sprite   Sprite
		distance float64
		index    int
	}

	spriteList := make([]spriteData, 0, len(sprites))
	for i, spr := range sprites {
		// Translate sprite position relative to camera
		dx := spr.X - posX
		dy := spr.Y - posY

		// Distance for sorting
		dist := math.Sqrt(dx*dx + dy*dy)
		spriteList = append(spriteList, spriteData{sprite: spr, distance: dist, index: i})
	}

	// Sort by distance (farthest first for painter's algorithm)
	for i := 0; i < len(spriteList)-1; i++ {
		for j := i + 1; j < len(spriteList); j++ {
			if spriteList[i].distance < spriteList[j].distance {
				spriteList[i], spriteList[j] = spriteList[j], spriteList[i]
			}
		}
	}

	// Project each sprite
	hits := make([]SpriteHit, 0, len(sprites))
	for _, sd := range spriteList {
		spr := sd.sprite

		// Sprite position relative to camera
		spriteX := spr.X - posX
		spriteY := spr.Y - posY

		// Inverse camera matrix for transformation
		invDet := 1.0 / (planeX*dirY - dirX*planeY)
		transformX := invDet * (dirY*spriteX - dirX*spriteY)
		transformY := invDet * (-planeY*spriteX + planeX*spriteY)

		// Skip sprites behind camera
		if transformY <= 0 {
			continue
		}

		// Screen X position
		spriteScreenX := int(float64(r.Width) / 2.0 * (1.0 + transformX/transformY))

		// Sprite dimensions in pixels
		spriteHeight := int(math.Abs(float64(r.Height) / transformY * spr.Height))
		spriteWidth := int(math.Abs(float64(r.Height) / transformY * spr.Width))

		// Calculate draw bounds
		drawStartY := -spriteHeight/2 + r.Height/2
		drawEndY := spriteHeight/2 + r.Height/2
		drawStartX := -spriteWidth/2 + spriteScreenX
		drawEndX := spriteWidth/2 + spriteScreenX

		// Clip to screen bounds
		if drawStartX < 0 {
			drawStartX = 0
		}
		if drawEndX >= r.Width {
			drawEndX = r.Width - 1
		}
		if drawStartY < 0 {
			drawStartY = 0
		}
		if drawEndY >= r.Height {
			drawEndY = r.Height - 1
		}

		// Skip if completely off-screen
		if drawStartX >= r.Width || drawEndX < 0 {
			continue
		}

		// Check occlusion against wall distances
		occluded := true
		for x := drawStartX; x <= drawEndX && x < len(wallDistances); x++ {
			if transformY < wallDistances[x].Distance {
				occluded = false
				break
			}
		}

		if occluded {
			continue
		}

		hits = append(hits, SpriteHit{
			ScreenX:      spriteScreenX,
			ScreenY:      r.Height / 2,
			ScreenWidth:  spriteWidth,
			ScreenHeight: spriteHeight,
			Distance:     transformY,
			Type:         spr.Type,
			DrawStartX:   drawStartX,
			DrawEndX:     drawEndX,
			DrawStartY:   drawStartY,
			DrawEndY:     drawEndY,
		})
	}

	return hits
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

// Package-level SetGenre for compatibility.
func SetGenre(genreID string) {}
