package weather

// WeatherComponent attaches weather system to the game world.
// This is a singleton component (only one per world).
type WeatherComponent struct {
	System *WeatherSystem
}

// Type returns the component type identifier.
func (w *WeatherComponent) Type() string {
	return "Weather"
}
