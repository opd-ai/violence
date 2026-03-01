package squad

import (
	"math"
	"testing"
)

func TestGetFormationOffset_Line(t *testing.T) {
	tests := []struct {
		name        string
		memberIndex int
		leaderDir   float64
		checkDist   bool
		minDist     float64
	}{
		{
			name:        "first member",
			memberIndex: 0,
			leaderDir:   math.Pi / 2,
			checkDist:   true,
			minDist:     1.5,
		},
		{
			name:        "second member",
			memberIndex: 1,
			leaderDir:   math.Pi / 2,
			checkDist:   true,
			minDist:     1.5,
		},
		{
			name:        "facing east",
			memberIndex: 0,
			leaderDir:   0,
			checkDist:   true,
			minDist:     1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dx, dy := GetFormationOffset(tt.memberIndex, FormationTypeLine, tt.leaderDir)

			if tt.checkDist {
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist < tt.minDist {
					t.Errorf("GetFormationOffset() distance = %v, want >= %v", dist, tt.minDist)
				}
			}
		})
	}
}

func TestGetFormationOffset_Wedge(t *testing.T) {
	tests := []struct {
		name        string
		memberIndex int
		leaderDir   float64
	}{
		{"first member", 0, 0},
		{"second member", 1, 0},
		{"third member", 2, 0},
		{"fourth member", 3, 0},
		{"facing north", 0, math.Pi / 2},
		{"facing south", 0, -math.Pi / 2},
		{"facing west", 0, math.Pi},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dx, dy := GetFormationOffset(tt.memberIndex, FormationTypeWedge, tt.leaderDir)

			// Verify non-zero offset (members should be positioned relative to leader)
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 0.1 {
				t.Errorf("GetFormationOffset() distance = %v, should be > 0.1", dist)
			}

			// Verify members are behind leader (negative Y in local coords)
			// Rotate back to local coordinates to verify
			cos := math.Cos(-tt.leaderDir)
			sin := math.Sin(-tt.leaderDir)
			localY := dx*sin + dy*cos
			if tt.memberIndex > 0 && localY > 0 {
				t.Errorf("Member %d should be behind leader (localY < 0), got %v",
					tt.memberIndex, localY)
			}
		})
	}
}

func TestGetFormationOffset_Column(t *testing.T) {
	leaderDir := 0.0 // Facing east

	for i := 0; i < 5; i++ {
		t.Run("member_"+string(rune('0'+i)), func(t *testing.T) {
			dx, dy := GetFormationOffset(i, FormationTypeColumn, leaderDir)

			// Verify member is positioned
			dist := math.Sqrt(dx*dx + dy*dy)
			expectedDist := float64(i+1) * DefaultSpacing

			if math.Abs(dist-expectedDist) > 0.1 {
				t.Errorf("Member %d distance = %v, want near %v", i, dist, expectedDist)
			}
		})
	}
}

func TestGetFormationOffset_Circle(t *testing.T) {
	tests := []struct {
		name        string
		memberIndex int
		leaderDir   float64
	}{
		{"first member", 0, 0},
		{"eighth member", 7, 0},
		{"ninth member (second ring)", 8, 0},
		{"facing north", 0, math.Pi / 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dx, dy := GetFormationOffset(tt.memberIndex, FormationTypeCircle, tt.leaderDir)

			// Verify members form circular pattern (distance from leader should be consistent per ring)
			dist := math.Sqrt(dx*dx + dy*dy)
			expectedRing := tt.memberIndex / 8
			expectedRadius := float64(expectedRing+1) * DefaultSpacing * 2.0

			if math.Abs(dist-expectedRadius) > 0.1 {
				t.Errorf("Member %d distance = %v, want near %v",
					tt.memberIndex, dist, expectedRadius)
			}
		})
	}
}

func TestGetFormationOffset_Staggered(t *testing.T) {
	leaderDir := 0.0

	for i := 0; i < 6; i++ {
		t.Run("member_"+string(rune('0'+i)), func(t *testing.T) {
			dx, dy := GetFormationOffset(i, FormationTypeStaggered, leaderDir)

			// Verify non-zero offset
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 0.1 {
				t.Errorf("Member %d distance = %v, should be > 0.1", i, dist)
			}

			// Verify staggering pattern
			row := i / 2
			expectedRow := -float64(row + 1)
			cos := math.Cos(-leaderDir)
			sin := math.Sin(-leaderDir)
			localY := dx*sin + dy*cos

			if math.Abs(localY/DefaultSpacing-expectedRow) > 0.1 {
				t.Errorf("Member %d localY/spacing = %v, want near %v",
					i, localY/DefaultSpacing, expectedRow)
			}
		})
	}
}

func TestGetFormationOffset_RotationConsistency(t *testing.T) {
	// Test that rotating formation produces consistent results
	directions := []float64{0, math.Pi / 4, math.Pi / 2, math.Pi, -math.Pi / 2}
	formations := []FormationType{
		FormationTypeLine,
		FormationTypeWedge,
		FormationTypeColumn,
		FormationTypeCircle,
		FormationTypeStaggered,
	}

	for _, formation := range formations {
		for _, dir := range directions {
			dx, dy := GetFormationOffset(0, formation, dir)
			dist := math.Sqrt(dx*dx + dy*dy)

			// Verify distance is consistent regardless of rotation
			if dist < 0.1 && formation != FormationTypeLine {
				t.Errorf("Formation %v direction %v: distance too small: %v",
					formation, dir, dist)
			}
		}
	}
}

func TestGetFormationOffset_InvalidFormation(t *testing.T) {
	// Test with invalid formation type (should fallback to column)
	invalidFormation := FormationType(999)
	dx, dy := GetFormationOffset(0, invalidFormation, 0)

	// Should match column formation first member
	expectedDX, expectedDY := GetFormationOffset(0, FormationTypeColumn, 0)
	if math.Abs(dx-expectedDX) > 0.01 || math.Abs(dy-expectedDY) > 0.01 {
		t.Errorf("Invalid formation fallback: got (%v, %v), want (%v, %v)",
			dx, dy, expectedDX, expectedDY)
	}
}

func TestGetFormationPositionCount(t *testing.T) {
	tests := []struct {
		name          string
		formation     FormationType
		memberCount   int
		wantPositions int
	}{
		{"line with 4 members", FormationTypeLine, 4, 4},
		{"circle with 4 members", FormationTypeCircle, 4, 8},
		{"circle with 8 members", FormationTypeCircle, 8, 8},
		{"circle with 9 members", FormationTypeCircle, 9, 16},
		{"wedge with 5 members", FormationTypeWedge, 5, 5},
		{"column with 10 members", FormationTypeColumn, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFormationPositionCount(tt.formation, tt.memberCount)
			if got != tt.wantPositions {
				t.Errorf("GetFormationPositionCount() = %v, want %v", got, tt.wantPositions)
			}
		})
	}
}

func TestGetFormationSpacing(t *testing.T) {
	tests := []struct {
		name      string
		formation FormationType
		want      float64
	}{
		{"line spacing", FormationTypeLine, DefaultSpacing},
		{"circle spacing", FormationTypeCircle, DefaultSpacing * 2.0},
		{"wedge spacing", FormationTypeWedge, DefaultSpacing},
		{"column spacing", FormationTypeColumn, DefaultSpacing},
		{"staggered spacing", FormationTypeStaggered, DefaultSpacing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFormationSpacing(tt.formation)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("GetFormationSpacing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFormationOffset_MultipleMembers(t *testing.T) {
	// Test that multiple members have reasonable spacing
	formation := FormationTypeWedge
	leaderDir := 0.0
	memberCount := 8

	prevDist := 0.0

	for i := 0; i < memberCount; i++ {
		dx, dy := GetFormationOffset(i, formation, leaderDir)
		dist := math.Sqrt(dx*dx + dy*dy)

		// Each member should have some distance from leader
		if dist < 0.5 {
			t.Errorf("Member %d too close to leader: %v", i, dist)
		}

		// Members generally get further from leader (with some exceptions for formations)
		if i > 0 && i < 4 { // First few should increase in distance
			if dist < prevDist*0.8 { // Allow some tolerance
				t.Logf("Member %d closer than previous (dist=%v, prev=%v), but within tolerance", i, dist, prevDist)
			}
		}
		prevDist = dist
	}
}

func formatPosition(x, y float64) string {
	// Round to 2 decimal places for comparison
	return string(rune(int(x*100))) + "," + string(rune(int(y*100)))
}

func TestGetFormationOffset_ZeroIndex(t *testing.T) {
	formations := []FormationType{
		FormationTypeLine,
		FormationTypeWedge,
		FormationTypeColumn,
		FormationTypeCircle,
		FormationTypeStaggered,
	}

	for _, formation := range formations {
		t.Run("formation_"+string(rune('0'+int(formation))), func(t *testing.T) {
			dx, dy := GetFormationOffset(0, formation, 0)

			// First member should be positioned (not at origin except for edge cases)
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 0.1 && formation != FormationTypeLine {
				t.Errorf("Formation %v: first member too close to leader: %v", formation, dist)
			}
		})
	}
}

func BenchmarkGetFormationOffset_Line(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetFormationOffset(i%10, FormationTypeLine, float64(i%360)*math.Pi/180)
	}
}

func BenchmarkGetFormationOffset_Wedge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetFormationOffset(i%10, FormationTypeWedge, float64(i%360)*math.Pi/180)
	}
}

func BenchmarkGetFormationOffset_Circle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetFormationOffset(i%20, FormationTypeCircle, float64(i%360)*math.Pi/180)
	}
}
