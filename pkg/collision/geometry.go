// Package collision provides collision geometry extraction from sprites.
package collision

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// GeometryExtractor extracts collision geometry from sprite data.
type GeometryExtractor struct {
	alphaThreshold  uint8
	simplifyEpsilon float64
}

// NewGeometryExtractor creates a geometry extractor with default settings.
func NewGeometryExtractor() *GeometryExtractor {
	return &GeometryExtractor{
		alphaThreshold:  128, // 50% alpha threshold
		simplifyEpsilon: 2.0, // Simplification tolerance in pixels
	}
}

// SetAlphaThreshold sets the alpha threshold for considering pixels solid (0-255).
func (ge *GeometryExtractor) SetAlphaThreshold(threshold uint8) {
	ge.alphaThreshold = threshold
}

// SetSimplifyEpsilon sets the Douglas-Peucker simplification tolerance.
func (ge *GeometryExtractor) SetSimplifyEpsilon(epsilon float64) {
	ge.simplifyEpsilon = epsilon
}

// ExtractConvexHull generates a convex hull polygon from a sprite's alpha channel.
// Returns vertices in local space (relative to sprite center).
func (ge *GeometryExtractor) ExtractConvexHull(sprite *ebiten.Image) []Point {
	if sprite == nil {
		return nil
	}

	bounds := sprite.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width == 0 || height == 0 {
		return nil
	}

	// Find all solid pixels
	var points []Point
	centerX := float64(width) / 2
	centerY := float64(height) / 2

	// Sample sprite pixels
	img := sprite
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			// RGBA returns premultiplied alpha, convert to 0-255
			alpha := uint8(a >> 8)

			// Skip transparent pixels
			if alpha < ge.alphaThreshold {
				continue
			}

			// Also check if pixel has color (not just alpha)
			if r == 0 && g == 0 && b == 0 && alpha > 0 {
				// Pure transparent black, skip
				continue
			}

			// Add point relative to center
			points = append(points, Point{
				X: float64(x) - centerX,
				Y: float64(y) - centerY,
			})
		}
	}

	if len(points) < 3 {
		// Not enough points for a polygon, return a simple square
		halfW := centerX * 0.6
		halfH := centerY * 0.6
		return []Point{
			{X: -halfW, Y: -halfH},
			{X: halfW, Y: -halfH},
			{X: halfW, Y: halfH},
			{X: -halfW, Y: halfH},
		}
	}

	// Compute convex hull
	hull := convexHull(points)

	// Simplify if too many points
	if len(hull) > 12 {
		hull = douglasPeucker(hull, ge.simplifyEpsilon)
	}

	return hull
}

// ExtractBoundingBox returns a tight AABB around solid pixels.
// Returns width, height in local space.
func (ge *GeometryExtractor) ExtractBoundingBox(sprite *ebiten.Image) (width, height float64) {
	if sprite == nil {
		return 0, 0
	}

	bounds := sprite.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return 0, 0
	}

	minX, minY := w, h
	maxX, maxY := 0, 0

	img := sprite
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			alpha := uint8(a >> 8)
			if alpha >= ge.alphaThreshold {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	if maxX < minX || maxY < minY {
		// No solid pixels found, use full bounds
		return float64(w), float64(h)
	}

	return float64(maxX - minX + 1), float64(maxY - minY + 1)
}

// GenerateAttackArc creates a polygon representing a weapon swing arc.
// startAngle and endAngle are in radians, radius is the attack reach.
// segments determines arc smoothness (recommended: 8-12).
func GenerateAttackArc(centerX, centerY, radius, startAngle, endAngle float64, segments int) []Point {
	if segments < 3 {
		segments = 8
	}

	// Create arc polygon: center point + arc points + back to center
	points := make([]Point, 0, segments+2)
	points = append(points, Point{X: centerX, Y: centerY})

	angleStep := (endAngle - startAngle) / float64(segments-1)
	for i := 0; i < segments; i++ {
		angle := startAngle + float64(i)*angleStep
		px := centerX + radius*math.Cos(angle)
		py := centerY + radius*math.Sin(angle)
		points = append(points, Point{X: px, Y: py})
	}

	return points
}

// GenerateConeShape creates a cone/triangle for directional attacks (thrusts, beams).
// direction is in radians, spread is the cone half-angle in radians.
func GenerateConeShape(originX, originY, length, direction, spread float64) []Point {
	// Three-point cone: origin, left edge, right edge
	leftAngle := direction - spread
	rightAngle := direction + spread

	return []Point{
		{X: originX, Y: originY},
		{X: originX + length*math.Cos(leftAngle), Y: originY + length*math.Sin(leftAngle)},
		{X: originX + length*math.Cos(rightAngle), Y: originY + length*math.Sin(rightAngle)},
	}
}

// GenerateRectangleShape creates a rotated rectangle for beam/slash attacks.
func GenerateRectangleShape(centerX, centerY, width, height, rotation float64) []Point {
	// Half dimensions
	hw := width / 2
	hh := height / 2

	// Four corners (unrotated)
	corners := []Point{
		{X: -hw, Y: -hh},
		{X: hw, Y: -hh},
		{X: hw, Y: hh},
		{X: -hw, Y: hh},
	}

	// Rotate and translate
	cos := math.Cos(rotation)
	sin := math.Sin(rotation)

	for i := range corners {
		// Rotate
		rx := corners[i].X*cos - corners[i].Y*sin
		ry := corners[i].X*sin + corners[i].Y*cos
		// Translate
		corners[i].X = rx + centerX
		corners[i].Y = ry + centerY
	}

	return corners
}

// convexHull computes the convex hull of a set of points using Graham scan.
func convexHull(points []Point) []Point {
	if len(points) < 3 {
		return points
	}

	// Find the bottommost point (or leftmost if tie)
	start := 0
	for i := 1; i < len(points); i++ {
		if points[i].Y < points[start].Y ||
			(points[i].Y == points[start].Y && points[i].X < points[start].X) {
			start = i
		}
	}

	// Swap start to front
	points[0], points[start] = points[start], points[0]
	p0 := points[0]

	// Sort by polar angle from p0
	sortByAngle(points[1:], p0)

	// Build hull
	hull := []Point{points[0], points[1]}

	for i := 2; i < len(points); i++ {
		// Remove points that make right turn
		for len(hull) > 1 {
			p1 := hull[len(hull)-2]
			p2 := hull[len(hull)-1]
			p3 := points[i]

			// Cross product to determine turn direction
			cross := (p2.X-p1.X)*(p3.Y-p1.Y) - (p2.Y-p1.Y)*(p3.X-p1.X)
			if cross > 0 {
				break // Left turn, keep p2
			}
			hull = hull[:len(hull)-1] // Right turn, remove p2
		}
		hull = append(hull, points[i])
	}

	return hull
}

// sortByAngle sorts points by polar angle relative to origin.
func sortByAngle(points []Point, origin Point) {
	// Simple insertion sort (good for small arrays)
	for i := 1; i < len(points); i++ {
		key := points[i]
		j := i - 1

		for j >= 0 && polarAngle(points[j], origin) > polarAngle(key, origin) {
			points[j+1] = points[j]
			j--
		}
		points[j+1] = key
	}
}

// polarAngle computes the polar angle of p relative to origin.
func polarAngle(p, origin Point) float64 {
	return math.Atan2(p.Y-origin.Y, p.X-origin.X)
}

// douglasPeucker simplifies a polygon using the Douglas-Peucker algorithm.
func douglasPeucker(points []Point, epsilon float64) []Point {
	if len(points) < 3 {
		return points
	}

	// Find the point with maximum distance
	dmax := 0.0
	index := 0
	end := len(points) - 1

	for i := 1; i < end; i++ {
		d := perpendicularDistance(points[i], points[0], points[end])
		if d > dmax {
			index = i
			dmax = d
		}
	}

	// If max distance is greater than epsilon, recursively simplify
	if dmax > epsilon {
		// Recursive call
		rec1 := douglasPeucker(points[:index+1], epsilon)
		rec2 := douglasPeucker(points[index:], epsilon)

		// Build result list
		result := make([]Point, 0, len(rec1)+len(rec2)-1)
		result = append(result, rec1[:len(rec1)-1]...)
		result = append(result, rec2...)
		return result
	}

	// Max distance <= epsilon, return endpoints
	return []Point{points[0], points[end]}
}

// perpendicularDistance computes the perpendicular distance from point to line segment.
func perpendicularDistance(point, lineStart, lineEnd Point) float64 {
	dx := lineEnd.X - lineStart.X
	dy := lineEnd.Y - lineStart.Y

	if dx == 0 && dy == 0 {
		// Line segment is a point
		return distance(point, lineStart)
	}

	// Compute the perpendicular distance
	num := math.Abs(dy*point.X - dx*point.Y + lineEnd.X*lineStart.Y - lineEnd.Y*lineStart.X)
	den := math.Sqrt(dx*dx + dy*dy)

	return num / den
}

// distance computes Euclidean distance between two points.
func distance(a, b Point) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}
