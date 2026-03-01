package collision

import (
	"testing"
)

func TestCircleCircleCollision(t *testing.T) {
	tests := []struct {
		name     string
		a        *Collider
		b        *Collider
		expected bool
	}{
		{
			name:     "overlapping circles",
			a:        NewCircleCollider(0, 0, 1, LayerPlayer, LayerEnemy),
			b:        NewCircleCollider(1, 0, 1, LayerEnemy, LayerPlayer),
			expected: true,
		},
		{
			name:     "separated circles",
			a:        NewCircleCollider(0, 0, 1, LayerPlayer, LayerEnemy),
			b:        NewCircleCollider(3, 0, 1, LayerEnemy, LayerPlayer),
			expected: false,
		},
		{
			name:     "touching circles",
			a:        NewCircleCollider(0, 0, 1, LayerPlayer, LayerEnemy),
			b:        NewCircleCollider(2, 0, 1, LayerEnemy, LayerPlayer),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TestCollision(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("TestCollision() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCircleAABBCollision(t *testing.T) {
	tests := []struct {
		name     string
		circle   *Collider
		aabb     *Collider
		expected bool
	}{
		{
			name:     "circle inside AABB",
			circle:   NewCircleCollider(5, 5, 1, LayerPlayer, LayerTerrain),
			aabb:     NewAABBCollider(0, 0, 10, 10, LayerTerrain, LayerPlayer),
			expected: true,
		},
		{
			name:     "circle outside AABB",
			circle:   NewCircleCollider(15, 15, 1, LayerPlayer, LayerTerrain),
			aabb:     NewAABBCollider(0, 0, 10, 10, LayerTerrain, LayerPlayer),
			expected: false,
		},
		{
			name:     "circle touching AABB edge",
			circle:   NewCircleCollider(11, 5, 1, LayerPlayer, LayerTerrain),
			aabb:     NewAABBCollider(0, 0, 10, 10, LayerTerrain, LayerPlayer),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TestCollision(tt.circle, tt.aabb)
			if result != tt.expected {
				t.Errorf("TestCollision() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLayerMasking(t *testing.T) {
	tests := []struct {
		name   string
		a      *Collider
		b      *Collider
		canHit bool
	}{
		{
			name:   "player can hit enemy",
			a:      NewCircleCollider(0, 0, 1, LayerPlayer, LayerEnemy),
			b:      NewCircleCollider(0, 0, 1, LayerEnemy, LayerPlayer),
			canHit: true,
		},
		{
			name:   "projectile ignores terrain",
			a:      NewCircleCollider(0, 0, 1, LayerProjectile, LayerEnemy),
			b:      NewCircleCollider(0, 0, 1, LayerTerrain, LayerPlayer|LayerEnemy),
			canHit: false,
		},
		{
			name:   "ethereal passes through walls",
			a:      NewCircleCollider(0, 0, 1, LayerEthereal, LayerNone),
			b:      NewCircleCollider(0, 0, 1, LayerTerrain, LayerPlayer|LayerEnemy),
			canHit: false,
		},
		{
			name:   "trigger detects player",
			a:      NewCircleCollider(0, 0, 1, LayerTrigger, LayerPlayer),
			b:      NewCircleCollider(0, 0, 1, LayerPlayer, LayerAll),
			canHit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCollide(tt.a, tt.b)
			if result != tt.canHit {
				t.Errorf("CanCollide() = %v, want %v", result, tt.canHit)
			}
		})
	}
}

func TestCapsuleCollision(t *testing.T) {
	tests := []struct {
		name     string
		a        *Collider
		b        *Collider
		expected bool
	}{
		{
			name:     "circle hits capsule middle",
			a:        NewCircleCollider(5, 5, 1, LayerPlayer, LayerEnemy),
			b:        NewCapsuleCollider(0, 5, 10, 5, 1, LayerEnemy, LayerPlayer),
			expected: true,
		},
		{
			name:     "circle misses capsule",
			a:        NewCircleCollider(5, 10, 1, LayerPlayer, LayerEnemy),
			b:        NewCapsuleCollider(0, 0, 10, 0, 1, LayerEnemy, LayerPlayer),
			expected: false,
		},
		{
			name:     "capsules cross",
			a:        NewCapsuleCollider(0, 0, 10, 0, 2, LayerPlayer, LayerEnemy),
			b:        NewCapsuleCollider(5, -3, 5, 3, 2, LayerEnemy, LayerPlayer),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TestCollision(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("TestCollision() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPolygonCollision(t *testing.T) {
	// Square polygon
	square := []Point{
		{X: -1, Y: -1},
		{X: 1, Y: -1},
		{X: 1, Y: 1},
		{X: -1, Y: 1},
	}

	tests := []struct {
		name     string
		a        *Collider
		b        *Collider
		expected bool
	}{
		{
			name:     "circle inside polygon",
			a:        NewCircleCollider(0, 0, 0.5, LayerPlayer, LayerEnvironment),
			b:        NewPolygonCollider(0, 0, square, LayerEnvironment, LayerPlayer),
			expected: true,
		},
		{
			name:     "circle outside polygon",
			a:        NewCircleCollider(5, 5, 0.5, LayerPlayer, LayerEnvironment),
			b:        NewPolygonCollider(0, 0, square, LayerEnvironment, LayerPlayer),
			expected: false,
		},
		{
			name:     "circle touches polygon edge",
			a:        NewCircleCollider(1.3, 0, 0.4, LayerPlayer, LayerEnvironment),
			b:        NewPolygonCollider(0, 0, square, LayerEnvironment, LayerPlayer),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TestCollision(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("TestCollision() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSlideVector(t *testing.T) {
	tests := []struct {
		name   string
		vx, vy float64
		nx, ny float64
		wantX  float64
		wantY  float64
	}{
		{
			name: "slide along horizontal wall",
			vx:   1, vy: 1,
			nx: 0, ny: 1,
			wantX: 1, wantY: 0,
		},
		{
			name: "slide along vertical wall",
			vx:   1, vy: 1,
			nx: 1, ny: 0,
			wantX: 0, wantY: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := SlideVector(tt.vx, tt.vy, tt.nx, tt.ny)
			tolerance := 0.001
			if abs(gotX-tt.wantX) > tolerance || abs(gotY-tt.wantY) > tolerance {
				t.Errorf("SlideVector() = (%v, %v), want (%v, %v)", gotX, gotY, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestGetCollisionNormal(t *testing.T) {
	tests := []struct {
		name     string
		a        *Collider
		b        *Collider
		wantZero bool
	}{
		{
			name:     "circle-circle normal",
			a:        NewCircleCollider(0, 0, 1, LayerPlayer, LayerEnemy),
			b:        NewCircleCollider(2, 0, 1, LayerEnemy, LayerPlayer),
			wantZero: false,
		},
		{
			name:     "circle-AABB normal",
			a:        NewCircleCollider(5, 5, 1, LayerPlayer, LayerTerrain),
			b:        NewAABBCollider(0, 0, 10, 10, LayerTerrain, LayerPlayer),
			wantZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nx, ny := GetCollisionNormal(tt.a, tt.b)
			isZero := (nx == 0 && ny == 0)
			if isZero != tt.wantZero {
				t.Errorf("GetCollisionNormal() returned zero=%v, want zero=%v", isZero, tt.wantZero)
			}
		})
	}
}

func TestDisabledCollider(t *testing.T) {
	a := NewCircleCollider(0, 0, 1, LayerPlayer, LayerEnemy)
	b := NewCircleCollider(0, 0, 1, LayerEnemy, LayerPlayer)

	// Should collide when enabled
	if !TestCollision(a, b) {
		t.Error("Expected collision when both enabled")
	}

	// Disable one
	a.Enabled = false
	if TestCollision(a, b) {
		t.Error("Expected no collision when one disabled")
	}

	// Re-enable
	a.Enabled = true
	if !TestCollision(a, b) {
		t.Error("Expected collision when re-enabled")
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func BenchmarkCircleCircle(b *testing.B) {
	c1 := NewCircleCollider(0, 0, 1, LayerPlayer, LayerEnemy)
	c2 := NewCircleCollider(1, 0, 1, LayerEnemy, LayerPlayer)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TestCollision(c1, c2)
	}
}

func BenchmarkCircleAABB(b *testing.B) {
	circle := NewCircleCollider(5, 5, 1, LayerPlayer, LayerTerrain)
	aabb := NewAABBCollider(0, 0, 10, 10, LayerTerrain, LayerPlayer)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TestCollision(circle, aabb)
	}
}

func BenchmarkCapsule(b *testing.B) {
	circle := NewCircleCollider(5, 5, 1, LayerPlayer, LayerEnemy)
	capsule := NewCapsuleCollider(0, 5, 10, 5, 1, LayerEnemy, LayerPlayer)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TestCollision(circle, capsule)
	}
}
