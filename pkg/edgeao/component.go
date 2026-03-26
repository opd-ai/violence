package edgeao

// Component stores per-tile edge AO data.
// This is attached to floor/ground entities for per-tile queries.
type Component struct {
	// AOValue is the computed ambient occlusion factor [0.0-1.0].
	// 0.0 = no occlusion (fully lit), 1.0 = maximum occlusion (fully shadowed).
	AOValue float64

	// EdgeType indicates the type of geometric edge causing occlusion.
	EdgeType EdgeType

	// Dirty flags the component for AO recalculation.
	Dirty bool
}

// EdgeType categorizes geometric edge configurations.
type EdgeType int

const (
	// EdgeNone indicates no nearby geometric edges.
	EdgeNone EdgeType = iota
	// EdgeWallJunction is where floor meets wall base.
	EdgeWallJunction
	// EdgeInsideCorner is an L-shaped (concave) wall intersection.
	EdgeInsideCorner
	// EdgeOutsideCorner is a convex wall corner.
	EdgeOutsideCorner
	// EdgeNarrowPassage is a corridor with walls on both sides.
	EdgeNarrowPassage
	// EdgeAlcove is a recessed area with walls on three sides.
	EdgeAlcove
)

// Type returns the component type identifier for ECS.
func (c *Component) Type() string {
	return "edgeao.Component"
}

// NewComponent creates a default edge AO component.
func NewComponent() *Component {
	return &Component{
		AOValue:  0.0,
		EdgeType: EdgeNone,
		Dirty:    true,
	}
}

// String returns a human-readable edge type name.
func (e EdgeType) String() string {
	switch e {
	case EdgeWallJunction:
		return "wall_junction"
	case EdgeInsideCorner:
		return "inside_corner"
	case EdgeOutsideCorner:
		return "outside_corner"
	case EdgeNarrowPassage:
		return "narrow_passage"
	case EdgeAlcove:
		return "alcove"
	default:
		return "none"
	}
}
