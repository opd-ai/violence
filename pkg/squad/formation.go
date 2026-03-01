package squad

import "math"

// FormationType defines squad member positioning patterns.
type FormationType int

const (
	// FormationTypeLine positions squad members in a horizontal line.
	FormationTypeLine FormationType = iota
	// FormationTypeWedge positions squad members in a V formation.
	FormationTypeWedge
	// FormationTypeColumn positions squad members in a vertical line.
	FormationTypeColumn
	// FormationTypeCircle positions squad members in a circular pattern.
	FormationTypeCircle
	// FormationTypeStaggered positions squad members in alternating rows.
	FormationTypeStaggered
)

const (
	// DefaultSpacing is the default distance between squad members.
	DefaultSpacing = 1.5
)

// GetFormationOffset calculates the relative position offset for a squad member.
// memberIndex: zero-based index of the squad member (0 = first member, 1 = second, etc.)
// formationType: the formation pattern to use
// leaderDir: leader's facing direction in radians (0 = east, π/2 = north, π = west, 3π/2 = south)
// Returns (dx, dy) offset relative to leader position in world coordinates.
func GetFormationOffset(memberIndex int, formationType FormationType, leaderDir float64) (dx, dy float64) {
	// Calculate base offset in formation-local coordinates (leader facing north)
	var localX, localY float64

	switch formationType {
	case FormationTypeLine:
		// Horizontal line perpendicular to leader direction
		// Members spread left and right: ..., -2, -1, 0, 1, 2, ...
		centerOffset := float64(memberIndex) - float64(memberIndex)/2.0
		localX = centerOffset * DefaultSpacing
		localY = -2.0

	case FormationTypeWedge:
		// V formation: members form two diagonal lines behind leader
		// Pattern: 0=left-back, 1=right-back, 2=further-left-back, 3=further-right-back
		row := memberIndex / 2
		side := memberIndex % 2
		xOffset := float64(row+1) * DefaultSpacing
		if side == 0 {
			xOffset = -xOffset // Left side
		}
		localX = xOffset
		localY = -float64(row+1) * DefaultSpacing

	case FormationTypeColumn:
		// Single-file line behind leader
		localX = 0
		localY = -float64(memberIndex+1) * DefaultSpacing

	case FormationTypeCircle:
		// Members arranged in a circle around leader
		// Use 8 positions maximum before expanding radius
		ringSize := 8
		ring := memberIndex / ringSize
		posInRing := memberIndex % ringSize
		angle := 2.0 * math.Pi * float64(posInRing) / float64(ringSize)
		radius := float64(ring+1) * DefaultSpacing * 2.0
		localX = radius * math.Cos(angle)
		localY = radius * math.Sin(angle)

	case FormationTypeStaggered:
		// Alternating rows with offset
		row := memberIndex / 2
		side := memberIndex % 2
		localX = float64(side) * DefaultSpacing
		localY = -float64(row+1) * DefaultSpacing

	default:
		// Fallback to column formation
		localX = 0
		localY = -float64(memberIndex+1) * DefaultSpacing
	}

	// Rotate local offset by leader direction
	// Leader direction: 0 = facing east (+X), rotate to align with leader
	cos := math.Cos(leaderDir)
	sin := math.Sin(leaderDir)
	dx = localX*cos - localY*sin
	dy = localX*sin + localY*cos

	return dx, dy
}

// GetFormationPositionCount returns the optimal number of positions for a formation type.
// Some formations have natural limits (e.g., circle has discrete ring sizes).
func GetFormationPositionCount(formationType FormationType, memberCount int) int {
	switch formationType {
	case FormationTypeCircle:
		// Calculate number of full rings needed
		rings := (memberCount + 7) / 8
		return rings * 8
	default:
		return memberCount
	}
}

// GetFormationSpacing returns the spacing distance for a formation type.
func GetFormationSpacing(formationType FormationType) float64 {
	switch formationType {
	case FormationTypeCircle:
		return DefaultSpacing * 2.0
	default:
		return DefaultSpacing
	}
}
