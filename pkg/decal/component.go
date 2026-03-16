// Package decal implements persistent visual marks from combat and interactions.
package decal

// DecalComponent stores persistent visual marks (blood, scorch, slash, bullet holes).
type DecalComponent struct {
	Decals []Decal
}

// Type implements engine.Component.
func (d *DecalComponent) Type() string {
	return "DecalComponent"
}

// Decal represents a single persistent visual mark.
type Decal struct {
	X, Y    float64 // World position
	Type    DecalType
	Subtype int   // Variation index
	Seed    int64 // For procedural generation
	Angle   float64
	Scale   float64
	Opacity float64 // 0.0-1.0, fades over time
	Age     float64 // Seconds since creation
	MaxAge  float64 // Seconds until fully faded
	Layer   int     // 0=floor, 1=wall
	GenreID string  // Genre for visual style
}

// DecalType categorizes the visual mark.
type DecalType int

const (
	DecalBlood      DecalType = iota // DecalBlood is a blood splatter decal.
	DecalScorch                      // DecalScorch is a scorch mark decal.
	DecalSlash                       // DecalSlash is a slash mark decal.
	DecalBulletHole                  // DecalBulletHole is a bullet hole decal.
	DecalMagicBurn                   // DecalMagicBurn is a magic burn decal.
	DecalAcid                        // DecalAcid is an acid burn decal.
	DecalFreeze                      // DecalFreeze is a frost decal.
	DecalElectric                    // DecalElectric is an electric burn decal.
)
