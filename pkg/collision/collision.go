// Package collision provides collision detection with layer masking and shape precision.
package collision

import (
	"math"
)

// Layer defines collision layer bitflags.
type Layer uint32

const (
	LayerNone        Layer = 0
	LayerPlayer      Layer = 1 << 0 // Player entities
	LayerEnemy       Layer = 1 << 1 // Enemy entities
	LayerProjectile  Layer = 1 << 2 // Projectile entities
	LayerTerrain     Layer = 1 << 3 // Static terrain (walls, obstacles)
	LayerEnvironment Layer = 1 << 4 // Props and decorations
	LayerEthereal    Layer = 1 << 5 // Ghost/ethereal entities (pass through most)
	LayerInteractive Layer = 1 << 6 // Doors, chests, switches
	LayerTrigger     Layer = 1 << 7 // Trigger zones
	LayerAll         Layer = 0xFFFFFFFF
)

// Shape defines collision shape types.
type Shape int

const (
	ShapeCircle  Shape = iota // Circle (x, y, radius)
	ShapeCapsule              // Capsule (x1, y1, x2, y2, radius)
	ShapeAABB                 // Axis-aligned bounding box (x, y, w, h)
	ShapePolygon              // Convex polygon (vertices)
)

// Collider stores collision geometry and layer information.
type Collider struct {
	Layer Layer   // Which layer this collider belongs to
	Mask  Layer   // Which layers this collider can interact with
	Shape Shape   // Collision shape type
	X, Y  float64 // Position (center for circle/AABB, start for capsule)

	// Shape-specific data
	Radius  float64 // Circle/capsule radius
	X2, Y2  float64 // Capsule end point
	W, H    float64 // AABB dimensions
	Polygon []Point // Polygon vertices (local space)
	Enabled bool    // Whether collision is active
}

// Point represents a 2D point.
type Point struct {
	X, Y float64
}

// NewCircleCollider creates a circular collider.
func NewCircleCollider(x, y, radius float64, layer, mask Layer) *Collider {
	return &Collider{
		Layer:   layer,
		Mask:    mask,
		Shape:   ShapeCircle,
		X:       x,
		Y:       y,
		Radius:  radius,
		Enabled: true,
	}
}

// NewCapsuleCollider creates a capsule (line segment with radius) collider.
func NewCapsuleCollider(x1, y1, x2, y2, radius float64, layer, mask Layer) *Collider {
	return &Collider{
		Layer:   layer,
		Mask:    mask,
		Shape:   ShapeCapsule,
		X:       x1,
		Y:       y1,
		X2:      x2,
		Y2:      y2,
		Radius:  radius,
		Enabled: true,
	}
}

// NewAABBCollider creates an axis-aligned bounding box collider.
func NewAABBCollider(x, y, w, h float64, layer, mask Layer) *Collider {
	return &Collider{
		Layer:   layer,
		Mask:    mask,
		Shape:   ShapeAABB,
		X:       x,
		Y:       y,
		W:       w,
		H:       h,
		Enabled: true,
	}
}

// NewPolygonCollider creates a convex polygon collider.
func NewPolygonCollider(x, y float64, vertices []Point, layer, mask Layer) *Collider {
	return &Collider{
		Layer:   layer,
		Mask:    mask,
		Shape:   ShapePolygon,
		X:       x,
		Y:       y,
		Polygon: vertices,
		Enabled: true,
	}
}

// CanCollide checks if two colliders can interact based on layer masks.
func CanCollide(a, b *Collider) bool {
	if !a.Enabled || !b.Enabled {
		return false
	}
	return (a.Layer&b.Mask) != 0 || (b.Layer&a.Mask) != 0
}

// TestCollision checks if two colliders intersect.
func TestCollision(a, b *Collider) bool {
	if !CanCollide(a, b) {
		return false
	}

	switch {
	case a.Shape == ShapeCircle && b.Shape == ShapeCircle:
		return circleCircle(a, b)
	case a.Shape == ShapeCircle && b.Shape == ShapeCapsule:
		return circleCapsule(a, b)
	case a.Shape == ShapeCapsule && b.Shape == ShapeCircle:
		return circleCapsule(b, a)
	case a.Shape == ShapeCircle && b.Shape == ShapeAABB:
		return circleAABB(a, b)
	case a.Shape == ShapeAABB && b.Shape == ShapeCircle:
		return circleAABB(b, a)
	case a.Shape == ShapeCapsule && b.Shape == ShapeCapsule:
		return capsuleCapsule(a, b)
	case a.Shape == ShapeAABB && b.Shape == ShapeAABB:
		return aabbAABB(a, b)
	case a.Shape == ShapeCircle && b.Shape == ShapePolygon:
		return circlePolygon(a, b)
	case a.Shape == ShapePolygon && b.Shape == ShapeCircle:
		return circlePolygon(b, a)
	case a.Shape == ShapePolygon && b.Shape == ShapePolygon:
		return polygonPolygon(a, b)
	default:
		// Fallback to circle-circle with bounding circles
		return circleCircleFallback(a, b)
	}
}

// GetBoundingCircle returns a bounding circle for any collider shape.
func GetBoundingCircle(c *Collider) (x, y, r float64) {
	switch c.Shape {
	case ShapeCircle:
		return c.X, c.Y, c.Radius
	case ShapeCapsule:
		dx := c.X2 - c.X
		dy := c.Y2 - c.Y
		length := math.Sqrt(dx*dx + dy*dy)
		return (c.X + c.X2) / 2, (c.Y + c.Y2) / 2, length/2 + c.Radius
	case ShapeAABB:
		return c.X + c.W/2, c.Y + c.H/2, math.Sqrt(c.W*c.W+c.H*c.H) / 2
	case ShapePolygon:
		if len(c.Polygon) == 0 {
			return c.X, c.Y, 0
		}
		maxDist := 0.0
		for _, p := range c.Polygon {
			dist := math.Sqrt(p.X*p.X + p.Y*p.Y)
			if dist > maxDist {
				maxDist = dist
			}
		}
		return c.X, c.Y, maxDist
	default:
		return c.X, c.Y, 1.0
	}
}

// Circle-Circle collision
func circleCircle(a, b *Collider) bool {
	dx := b.X - a.X
	dy := b.Y - a.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	return dist <= (a.Radius + b.Radius)
}

// Circle-AABB collision
func circleAABB(circle, aabb *Collider) bool {
	closestX := clamp(circle.X, aabb.X, aabb.X+aabb.W)
	closestY := clamp(circle.Y, aabb.Y, aabb.Y+aabb.H)
	dx := circle.X - closestX
	dy := circle.Y - closestY
	return (dx*dx + dy*dy) <= (circle.Radius * circle.Radius)
}

// Circle-Capsule collision
func circleCapsule(circle, capsule *Collider) bool {
	// Find closest point on line segment to circle center
	px, py := closestPointOnSegment(circle.X, circle.Y, capsule.X, capsule.Y, capsule.X2, capsule.Y2)
	dx := circle.X - px
	dy := circle.Y - py
	dist := math.Sqrt(dx*dx + dy*dy)
	return dist <= (circle.Radius + capsule.Radius)
}

// Capsule-Capsule collision
func capsuleCapsule(a, b *Collider) bool {
	// Simplified: check distance between line segments
	dist := segmentDistance(a.X, a.Y, a.X2, a.Y2, b.X, b.Y, b.X2, b.Y2)
	return dist <= (a.Radius + b.Radius)
}

// AABB-AABB collision
func aabbAABB(a, b *Collider) bool {
	return a.X < b.X+b.W &&
		a.X+a.W > b.X &&
		a.Y < b.Y+b.H &&
		a.Y+a.H > b.Y
}

// Circle-Polygon collision (SAT-based)
func circlePolygon(circle, polygon *Collider) bool {
	if len(polygon.Polygon) < 3 {
		return false
	}

	// Transform polygon vertices to world space
	worldVerts := make([]Point, len(polygon.Polygon))
	for i, v := range polygon.Polygon {
		worldVerts[i] = Point{X: polygon.X + v.X, Y: polygon.Y + v.Y}
	}

	// Check if circle center is inside polygon
	if pointInPolygon(circle.X, circle.Y, worldVerts) {
		return true
	}

	// Check distance to each edge
	for i := 0; i < len(worldVerts); i++ {
		j := (i + 1) % len(worldVerts)
		px, py := closestPointOnSegment(circle.X, circle.Y, worldVerts[i].X, worldVerts[i].Y, worldVerts[j].X, worldVerts[j].Y)
		dx := circle.X - px
		dy := circle.Y - py
		if dx*dx+dy*dy <= circle.Radius*circle.Radius {
			return true
		}
	}

	return false
}

// Polygon-Polygon collision (SAT-based simplified)
func polygonPolygon(a, b *Collider) bool {
	// Transform to world space
	aVerts := make([]Point, len(a.Polygon))
	for i, v := range a.Polygon {
		aVerts[i] = Point{X: a.X + v.X, Y: a.Y + v.Y}
	}
	bVerts := make([]Point, len(b.Polygon))
	for i, v := range b.Polygon {
		bVerts[i] = Point{X: b.X + v.X, Y: b.Y + v.Y}
	}

	// Simplified SAT: check if any vertex of one polygon is inside the other
	for _, v := range aVerts {
		if pointInPolygon(v.X, v.Y, bVerts) {
			return true
		}
	}
	for _, v := range bVerts {
		if pointInPolygon(v.X, v.Y, aVerts) {
			return true
		}
	}

	return false
}

// Fallback collision using bounding circles
func circleCircleFallback(a, b *Collider) bool {
	ax, ay, ar := GetBoundingCircle(a)
	bx, by, br := GetBoundingCircle(b)
	dx := bx - ax
	dy := by - ay
	dist := math.Sqrt(dx*dx + dy*dy)
	return dist <= (ar + br)
}

// closestPointOnSegment finds the closest point on line segment (x1,y1)-(x2,y2) to point (px,py).
func closestPointOnSegment(px, py, x1, y1, x2, y2 float64) (float64, float64) {
	dx := x2 - x1
	dy := y2 - y1
	if dx == 0 && dy == 0 {
		return x1, y1
	}

	t := ((px-x1)*dx + (py-y1)*dy) / (dx*dx + dy*dy)
	t = clamp(t, 0, 1)

	return x1 + t*dx, y1 + t*dy
}

// segmentDistance returns the minimum distance between two line segments.
func segmentDistance(ax1, ay1, ax2, ay2, bx1, by1, bx2, by2 float64) float64 {
	// Test four endpoint-to-segment distances
	d1x, d1y := closestPointOnSegment(ax1, ay1, bx1, by1, bx2, by2)
	d1 := math.Sqrt((ax1-d1x)*(ax1-d1x) + (ay1-d1y)*(ay1-d1y))

	d2x, d2y := closestPointOnSegment(ax2, ay2, bx1, by1, bx2, by2)
	d2 := math.Sqrt((ax2-d2x)*(ax2-d2x) + (ay2-d2y)*(ay2-d2y))

	d3x, d3y := closestPointOnSegment(bx1, by1, ax1, ay1, ax2, ay2)
	d3 := math.Sqrt((bx1-d3x)*(bx1-d3x) + (by1-d3y)*(by1-d3y))

	d4x, d4y := closestPointOnSegment(bx2, by2, ax1, ay1, ax2, ay2)
	d4 := math.Sqrt((bx2-d4x)*(bx2-d4x) + (by2-d4y)*(by2-d4y))

	return math.Min(math.Min(d1, d2), math.Min(d3, d4))
}

// pointInPolygon checks if a point is inside a convex polygon using winding number.
func pointInPolygon(px, py float64, vertices []Point) bool {
	n := len(vertices)
	if n < 3 {
		return false
	}

	// Ray casting algorithm
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		xi, yi := vertices[i].X, vertices[i].Y
		xj, yj := vertices[j].X, vertices[j].Y

		if ((yi > py) != (yj > py)) && (px < (xj-xi)*(py-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}

	return inside
}

// clamp restricts a value to a range.
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// SlideVector computes a sliding vector along a surface given a collision normal.
// Used for smooth wall sliding instead of stopping at collision.
func SlideVector(velocityX, velocityY, normalX, normalY float64) (slideX, slideY float64) {
	// Project velocity onto tangent of normal
	dot := velocityX*normalX + velocityY*normalY
	slideX = velocityX - dot*normalX
	slideY = velocityY - dot*normalY
	return slideX, slideY
}

// GetCollisionNormal computes the collision normal from a to b.
// Returns zero vector if no collision or unable to compute.
func GetCollisionNormal(a, b *Collider) (nx, ny float64) {
	switch {
	case a.Shape == ShapeCircle && b.Shape == ShapeCircle:
		dx := a.X - b.X
		dy := a.Y - b.Y
		length := math.Sqrt(dx*dx + dy*dy)
		if length > 0 {
			return dx / length, dy / length
		}
	case a.Shape == ShapeCircle && b.Shape == ShapeAABB:
		closestX := clamp(a.X, b.X, b.X+b.W)
		closestY := clamp(a.Y, b.Y, b.Y+b.H)
		dx := a.X - closestX
		dy := a.Y - closestY
		length := math.Sqrt(dx*dx + dy*dy)
		if length > 0 {
			return dx / length, dy / length
		}
		// If circle center is inside AABB, push out from nearest edge
		if a.X >= b.X && a.X <= b.X+b.W && a.Y >= b.Y && a.Y <= b.Y+b.H {
			// Distance to each edge
			distLeft := a.X - b.X
			distRight := (b.X + b.W) - a.X
			distTop := a.Y - b.Y
			distBottom := (b.Y + b.H) - a.Y
			minDist := math.Min(math.Min(distLeft, distRight), math.Min(distTop, distBottom))
			if minDist == distLeft {
				return -1, 0
			} else if minDist == distRight {
				return 1, 0
			} else if minDist == distTop {
				return 0, -1
			} else {
				return 0, 1
			}
		}
	}
	return 0, 0
}
