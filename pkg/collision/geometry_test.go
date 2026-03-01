package collision

import (
	"math"
	"testing"
)

func TestGeometryExtractor_ExtractConvexHull(t *testing.T) {
	// Note: These tests cannot directly read pixels from ebiten.Image without a running game
	// Instead we test with nil sprites and edge cases that don't require pixel access
	tests := []struct {
		name    string
		wantMin int // Minimum expected vertices
		wantMax int // Maximum expected vertices
	}{
		{
			name:    "nil sprite",
			wantMin: 0,
			wantMax: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ge := NewGeometryExtractor()
			hull := ge.ExtractConvexHull(nil)

			if len(hull) < tt.wantMin || len(hull) > tt.wantMax {
				t.Errorf("ExtractConvexHull() got %d vertices, want between %d and %d",
					len(hull), tt.wantMin, tt.wantMax)
			}

			// Verify hull is convex if we have enough points
			if len(hull) >= 3 {
				if !isConvex(hull) {
					t.Errorf("ExtractConvexHull() result is not convex")
				}
			}
		})
	}
}

func TestGeometryExtractor_ExtractBoundingBox(t *testing.T) {
	// Note: Cannot read pixels without a running game
	tests := []struct {
		name       string
		wantWidth  float64
		wantHeight float64
		tolerance  float64
	}{
		{
			name:       "nil sprite",
			wantWidth:  0,
			wantHeight: 0,
			tolerance:  0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ge := NewGeometryExtractor()
			w, h := ge.ExtractBoundingBox(nil)

			if math.Abs(w-tt.wantWidth) > tt.tolerance {
				t.Errorf("ExtractBoundingBox() width = %v, want %v", w, tt.wantWidth)
			}
			if math.Abs(h-tt.wantHeight) > tt.tolerance {
				t.Errorf("ExtractBoundingBox() height = %v, want %v", h, tt.wantHeight)
			}
		})
	}
}

func TestGenerateAttackArc(t *testing.T) {
	tests := []struct {
		name       string
		centerX    float64
		centerY    float64
		radius     float64
		startAngle float64
		endAngle   float64
		segments   int
		wantPoints int
	}{
		{
			name:       "90 degree arc",
			centerX:    0,
			centerY:    0,
			radius:     10,
			startAngle: 0,
			endAngle:   math.Pi / 2,
			segments:   8,
			wantPoints: 9, // center + 8 arc points
		},
		{
			name:       "180 degree arc",
			centerX:    5,
			centerY:    5,
			radius:     15,
			startAngle: 0,
			endAngle:   math.Pi,
			segments:   12,
			wantPoints: 13,
		},
		{
			name:       "minimal segments",
			centerX:    0,
			centerY:    0,
			radius:     5,
			startAngle: 0,
			endAngle:   math.Pi / 4,
			segments:   2,
			wantPoints: 9, // Should clamp to minimum 8 segments
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arc := GenerateAttackArc(tt.centerX, tt.centerY, tt.radius,
				tt.startAngle, tt.endAngle, tt.segments)

			if len(arc) != tt.wantPoints {
				t.Errorf("GenerateAttackArc() got %d points, want %d",
					len(arc), tt.wantPoints)
			}

			// First point should be the center
			if math.Abs(arc[0].X-tt.centerX) > 0.01 ||
				math.Abs(arc[0].Y-tt.centerY) > 0.01 {
				t.Errorf("GenerateAttackArc() first point not at center: got (%v,%v), want (%v,%v)",
					arc[0].X, arc[0].Y, tt.centerX, tt.centerY)
			}

			// Check that arc points are approximately at the correct radius
			for i := 1; i < len(arc); i++ {
				dx := arc[i].X - tt.centerX
				dy := arc[i].Y - tt.centerY
				dist := math.Sqrt(dx*dx + dy*dy)
				if math.Abs(dist-tt.radius) > 0.01 {
					t.Errorf("GenerateAttackArc() point %d distance %v, want %v",
						i, dist, tt.radius)
				}
			}
		})
	}
}

func TestGenerateConeShape(t *testing.T) {
	tests := []struct {
		name      string
		originX   float64
		originY   float64
		length    float64
		direction float64
		spread    float64
	}{
		{
			name:      "cone pointing right",
			originX:   0,
			originY:   0,
			length:    10,
			direction: 0,
			spread:    math.Pi / 6, // 30 degrees
		},
		{
			name:      "cone pointing up",
			originX:   5,
			originY:   5,
			length:    15,
			direction: -math.Pi / 2,
			spread:    math.Pi / 4, // 45 degrees
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cone := GenerateConeShape(tt.originX, tt.originY, tt.length,
				tt.direction, tt.spread)

			if len(cone) != 3 {
				t.Errorf("GenerateConeShape() got %d points, want 3", len(cone))
			}

			// First point should be origin
			if math.Abs(cone[0].X-tt.originX) > 0.01 ||
				math.Abs(cone[0].Y-tt.originY) > 0.01 {
				t.Errorf("GenerateConeShape() origin not correct")
			}

			// Other two points should be at approximately the correct distance
			for i := 1; i < 3; i++ {
				dx := cone[i].X - tt.originX
				dy := cone[i].Y - tt.originY
				dist := math.Sqrt(dx*dx + dy*dy)
				if math.Abs(dist-tt.length) > 0.01 {
					t.Errorf("GenerateConeShape() point %d distance %v, want %v",
						i, dist, tt.length)
				}
			}
		})
	}
}

func TestGenerateRectangleShape(t *testing.T) {
	tests := []struct {
		name     string
		centerX  float64
		centerY  float64
		width    float64
		height   float64
		rotation float64
	}{
		{
			name:     "unrotated rectangle",
			centerX:  0,
			centerY:  0,
			width:    10,
			height:   5,
			rotation: 0,
		},
		{
			name:     "45 degree rotated",
			centerX:  5,
			centerY:  5,
			width:    8,
			height:   4,
			rotation: math.Pi / 4,
		},
		{
			name:     "90 degree rotated",
			centerX:  10,
			centerY:  10,
			width:    6,
			height:   12,
			rotation: math.Pi / 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rect := GenerateRectangleShape(tt.centerX, tt.centerY, tt.width,
				tt.height, tt.rotation)

			if len(rect) != 4 {
				t.Errorf("GenerateRectangleShape() got %d points, want 4", len(rect))
			}

			// Check that center of rectangle is correct
			avgX := 0.0
			avgY := 0.0
			for _, p := range rect {
				avgX += p.X
				avgY += p.Y
			}
			avgX /= 4
			avgY /= 4

			if math.Abs(avgX-tt.centerX) > 0.01 || math.Abs(avgY-tt.centerY) > 0.01 {
				t.Errorf("GenerateRectangleShape() center (%v,%v), want (%v,%v)",
					avgX, avgY, tt.centerX, tt.centerY)
			}
		})
	}
}

func TestConvexHull(t *testing.T) {
	tests := []struct {
		name   string
		points []Point
		want   int // Expected number of hull points
	}{
		{
			name: "square points",
			points: []Point{
				{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1},
			},
			want: 4,
		},
		{
			name: "square with interior point",
			points: []Point{
				{X: 0, Y: 0},
				{X: 1, Y: 0},
				{X: 1, Y: 1},
				{X: 0, Y: 1},
				{X: 0.5, Y: 0.5}, // Interior point
			},
			want: 4,
		},
		{
			name: "collinear points",
			points: []Point{
				{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0},
			},
			want: 2,
		},
		{
			name:   "insufficient points",
			points: []Point{{X: 0, Y: 0}},
			want:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hull := convexHull(tt.points)
			if len(hull) != tt.want {
				t.Errorf("convexHull() got %d points, want %d", len(hull), tt.want)
			}
		})
	}
}

func TestDouglasPeucker(t *testing.T) {
	tests := []struct {
		name    string
		points  []Point
		epsilon float64
		wantMax int // Maximum expected points after simplification
	}{
		{
			name: "simple line",
			points: []Point{
				{X: 0, Y: 0},
				{X: 1, Y: 0.1},
				{X: 2, Y: 0},
				{X: 3, Y: 0.1},
				{X: 4, Y: 0},
			},
			epsilon: 0.5,
			wantMax: 3,
		},
		{
			name: "complex polygon",
			points: []Point{
				{X: 0, Y: 0},
				{X: 1, Y: 0},
				{X: 1.5, Y: 0.1},
				{X: 2, Y: 0},
				{X: 2, Y: 1},
				{X: 0, Y: 1},
			},
			epsilon: 0.5,
			wantMax: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			simplified := douglasPeucker(tt.points, tt.epsilon)
			if len(simplified) > tt.wantMax {
				t.Errorf("douglasPeucker() got %d points, want max %d",
					len(simplified), tt.wantMax)
			}
			if len(simplified) < 2 {
				t.Errorf("douglasPeucker() got %d points, need at least 2", len(simplified))
			}
		})
	}
}

func TestGeometryExtractor_SetAlphaThreshold(t *testing.T) {
	ge := NewGeometryExtractor()
	ge.SetAlphaThreshold(200)

	if ge.alphaThreshold != 200 {
		t.Errorf("SetAlphaThreshold() got %d, want 200", ge.alphaThreshold)
	}
}

func TestGeometryExtractor_SetSimplifyEpsilon(t *testing.T) {
	ge := NewGeometryExtractor()
	ge.SetSimplifyEpsilon(5.0)

	if ge.simplifyEpsilon != 5.0 {
		t.Errorf("SetSimplifyEpsilon() got %v, want 5.0", ge.simplifyEpsilon)
	}
}

// Helper function to check if a polygon is convex
func isConvex(points []Point) bool {
	if len(points) < 3 {
		return true
	}

	n := len(points)
	sign := 0

	for i := 0; i < n; i++ {
		p1 := points[i]
		p2 := points[(i+1)%n]
		p3 := points[(i+2)%n]

		// Cross product
		cross := (p2.X-p1.X)*(p3.Y-p1.Y) - (p2.Y-p1.Y)*(p3.X-p1.X)

		if cross != 0 {
			if sign == 0 {
				sign = int(cross / math.Abs(cross))
			} else if int(cross/math.Abs(cross)) != sign {
				return false
			}
		}
	}

	return true
}

// Benchmark tests
func BenchmarkExtractConvexHull(b *testing.B) {
	// Cannot benchmark actual pixel extraction without running game
	// Test the convex hull algorithm with point data instead
	ge := NewGeometryExtractor()

	// Create test points
	points := make([]Point, 100)
	for i := 0; i < 100; i++ {
		angle := float64(i) * 2 * math.Pi / 100
		points[i] = Point{
			X: 16 * math.Cos(angle),
			Y: 16 * math.Sin(angle),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ge
		convexHull(points)
	}
}

func BenchmarkGenerateAttackArc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateAttackArc(0, 0, 10, 0, math.Pi/2, 12)
	}
}
