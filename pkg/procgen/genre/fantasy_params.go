package genre

// FantasyParams defines all procedural asset generation parameters for the Fantasy genre.
type FantasyParams struct {
	Fog     FogParams
	Palette PaletteParams
	Texture TextureParams
	SFX     SFXParams
	Music   MusicParams
}

// FogParams defines fog rendering parameters.
type FogParams struct {
	R uint8
	G uint8
	B uint8
}

// PaletteParams defines color palette generation parameters.
type PaletteParams struct {
	PrimaryHue   float64
	SecondaryHue float64
	Saturation   float64
	Brightness   float64
	NumColors    int
}

// TextureParams defines procedural texture generation parameters.
type TextureParams struct {
	SeedOffset   int64
	NoiseScale   float64
	OctaveCount  int
	Persistence  float64
	TileSize     int
	WallVariance float64
}

// SFXParams defines audio synthesis parameters for sound effects.
type SFXParams struct {
	Waveform  WaveformType
	Frequency float64
	Envelope  EnvelopeParams
}

// WaveformType defines audio waveform types.
type WaveformType int

const (
	WaveformSine WaveformType = iota
	WaveformSquare
	WaveformSawtooth
	WaveformTriangle
	WaveformNoise
)

// EnvelopeParams defines ADSR envelope parameters (seconds).
type EnvelopeParams struct {
	Attack  float64
	Decay   float64
	Sustain float64
	Release float64
}

// MusicParams defines procedural music generation parameters.
type MusicParams struct {
	Scale       ScaleType
	Tempo       int
	TimeSignNum int
	TimeSignDen int
	ChordProg   []int
}

// ScaleType defines musical scale types.
type ScaleType int

const (
	ScaleMinor ScaleType = iota
	ScaleMajor
	ScaleDorian
	ScalePhrygian
	ScaleLydian
	ScaleMixolydian
	ScaleAeolian
	ScaleLocrian
)

// DefaultFantasyParams returns the default parameter set for Fantasy genre.
func DefaultFantasyParams() FantasyParams {
	return FantasyParams{
		Fog: FogParams{
			R: 120,
			G: 130,
			B: 150,
		},
		Palette: PaletteParams{
			PrimaryHue:   240.0,
			SecondaryHue: 30.0,
			Saturation:   0.6,
			Brightness:   0.7,
			NumColors:    16,
		},
		Texture: TextureParams{
			SeedOffset:   1000,
			NoiseScale:   0.15,
			OctaveCount:  4,
			Persistence:  0.5,
			TileSize:     64,
			WallVariance: 0.3,
		},
		SFX: SFXParams{
			Waveform:  WaveformTriangle,
			Frequency: 440.0,
			Envelope: EnvelopeParams{
				Attack:  0.05,
				Decay:   0.1,
				Sustain: 0.3,
				Release: 0.2,
			},
		},
		Music: MusicParams{
			Scale:       ScaleDorian,
			Tempo:       90,
			TimeSignNum: 3,
			TimeSignDen: 4,
			ChordProg:   []int{1, 6, 4, 5},
		},
	}
}
