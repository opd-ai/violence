package trap

// Component stores trap data for an entity.
type Component struct {
	Trap *Trap
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "trap"
}
