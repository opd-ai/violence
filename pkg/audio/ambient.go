// Package audio manages sound effects and music playback with adaptive music layers.
package audio

import (
	"bytes"
	"math"
	"math/rand"
)

// AmbientSoundscape generates continuous background atmospheric audio.
// All generation is deterministic based on genre and seed.
type AmbientSoundscape struct {
	genreID   string
	seed      uint64
	loopData  []byte
	duration  int // samples per loop
	isPlaying bool
}

// NewAmbientSoundscape creates a new ambient soundscape generator.
func NewAmbientSoundscape(genreID string, seed uint64) *AmbientSoundscape {
	duration := sampleRate * 60 // 60 second loop
	return &AmbientSoundscape{
		genreID:  genreID,
		seed:     seed,
		duration: duration,
	}
}

// Generate creates the ambient audio loop data.
// Must be called before GetLoopData().
func (a *AmbientSoundscape) Generate() {
	a.loopData = a.generateLoop()
}

// GetLoopData returns the generated audio loop as WAV data.
func (a *AmbientSoundscape) GetLoopData() []byte {
	if a.loopData == nil {
		a.Generate()
	}
	return a.loopData
}

// SetGenre updates the genre and regenerates the soundscape.
func (a *AmbientSoundscape) SetGenre(genreID string) {
	if a.genreID != genreID {
		a.genreID = genreID
		a.loopData = nil // Force regeneration
	}
}

// generateLoop creates a genre-specific ambient loop.
func (a *AmbientSoundscape) generateLoop() []byte {
	rng := rand.New(rand.NewSource(int64(a.seed)))

	buf := &bytes.Buffer{}
	writeWAVHeader(buf, a.duration)

	pcmData := make([]int16, a.duration*2)

	switch a.genreID {
	case "fantasy":
		a.generateDungeonEcho(pcmData, rng)
	case "scifi":
		a.generateStationHum(pcmData, rng)
	case "horror":
		a.generateHospitalSilence(pcmData, rng)
	case "cyberpunk":
		a.generateServerDrone(pcmData, rng)
	case "postapoc":
		a.generateWind(pcmData, rng)
	default:
		a.generateGenericAmbient(pcmData, rng)
	}

	// Write PCM data
	for i := 0; i < len(pcmData); i++ {
		writeInt16(buf, pcmData[i])
	}

	return buf.Bytes()
}

// generateDungeonEcho creates fantasy dungeon atmosphere with water drips and distant echoes.
func (a *AmbientSoundscape) generateDungeonEcho(pcmData []int16, rng *rand.Rand) {
	// Low rumble base layer
	for i := 0; i < len(pcmData)/2; i++ {
		t := float64(i) / float64(sampleRate)

		// Deep rumble (30-40 Hz)
		rumble := math.Sin(2*math.Pi*35.0*t) * 0.15
		rumble += math.Sin(2*math.Pi*38.0*t) * 0.1

		val := int16(rumble * 2000.0)
		pcmData[i*2] += val
		pcmData[i*2+1] += val
	}

	// Random water drips throughout the loop
	numDrips := 15 + rng.Intn(10)
	for d := 0; d < numDrips; d++ {
		dripStart := rng.Intn(len(pcmData)/2 - sampleRate)
		dripLen := sampleRate / 20

		for i := 0; i < dripLen; i++ {
			env := math.Exp(-float64(i) / float64(dripLen/3))
			freq := 800.0 + float64(rng.Intn(400))
			val := math.Sin(2*math.Pi*freq*float64(i)/float64(sampleRate)) * env * 1500.0

			idx := (dripStart + i) * 2
			if idx+1 < len(pcmData) {
				pcmData[idx] += int16(val)
				pcmData[idx+1] += int16(val)
			}
		}
	}
}

// generateStationHum creates sci-fi station atmosphere with mechanical hum and electrical buzz.
func (a *AmbientSoundscape) generateStationHum(pcmData []int16, rng *rand.Rand) {
	// Electrical hum base (60 Hz and harmonics)
	for i := 0; i < len(pcmData)/2; i++ {
		t := float64(i) / float64(sampleRate)

		hum := math.Sin(2*math.Pi*60.0*t) * 0.3
		hum += math.Sin(2*math.Pi*120.0*t) * 0.15
		hum += math.Sin(2*math.Pi*180.0*t) * 0.08

		// Add slow modulation for ventilation system
		mod := 1.0 + math.Sin(2*math.Pi*0.5*t)*0.2

		val := int16(hum * mod * 2500.0)
		pcmData[i*2] += val
		pcmData[i*2+1] += val
	}

	// Random electrical sparks
	numSparks := 5 + rng.Intn(5)
	for s := 0; s < numSparks; s++ {
		sparkStart := rng.Intn(len(pcmData)/2 - sampleRate/10)
		sparkLen := sampleRate / 50

		for i := 0; i < sparkLen; i++ {
			env := math.Exp(-float64(i) / float64(sparkLen/5))
			noise := (rng.Float64()*2.0 - 1.0) * env * 3000.0

			idx := (sparkStart + i) * 2
			if idx+1 < len(pcmData) {
				pcmData[idx] += int16(noise)
				pcmData[idx+1] += int16(noise)
			}
		}
	}
}

// generateHospitalSilence creates horror atmosphere with unsettling silence and distant sounds.
func (a *AmbientSoundscape) generateHospitalSilence(pcmData []int16, rng *rand.Rand) {
	// Very low frequency drone for unease (20-30 Hz)
	for i := 0; i < len(pcmData)/2; i++ {
		t := float64(i) / float64(sampleRate)

		drone := math.Sin(2*math.Pi*22.0*t) * 0.08
		drone += math.Sin(2*math.Pi*28.0*t) * 0.05

		// Slow breathing-like modulation
		breathMod := 1.0 + math.Sin(2*math.Pi*0.15*t)*0.4

		val := int16(drone * breathMod * 1500.0)
		pcmData[i*2] += val
		pcmData[i*2+1] += val
	}

	// Rare distant metallic sounds
	numSounds := 3 + rng.Intn(3)
	for s := 0; s < numSounds; s++ {
		soundStart := rng.Intn(len(pcmData)/2 - sampleRate)
		soundLen := sampleRate / 4

		for i := 0; i < soundLen; i++ {
			t := float64(i) / float64(soundLen)
			env := math.Sin(t*math.Pi) * 0.3

			freq := 200.0 + float64(rng.Intn(300))
			val := math.Sin(2*math.Pi*freq*float64(i)/float64(sampleRate)) * env * 800.0

			idx := (soundStart + i) * 2
			if idx+1 < len(pcmData) {
				pcmData[idx] += int16(val)
				pcmData[idx+1] += int16(val)
			}
		}
	}
}

// generateServerDrone creates cyberpunk atmosphere with server hum and data processing sounds.
func (a *AmbientSoundscape) generateServerDrone(pcmData []int16, rng *rand.Rand) {
	// Multi-layered server hum
	for i := 0; i < len(pcmData)/2; i++ {
		t := float64(i) / float64(sampleRate)

		// Multiple frequencies for complex server room sound
		drone := math.Sin(2*math.Pi*85.0*t) * 0.2
		drone += math.Sin(2*math.Pi*92.0*t) * 0.15
		drone += math.Sin(2*math.Pi*110.0*t) * 0.12

		// Pulsing modulation
		pulse := 1.0 + math.Sin(2*math.Pi*2.0*t)*0.15

		val := int16(drone * pulse * 2800.0)
		pcmData[i*2] += val
		pcmData[i*2+1] += val
	}

	// Random hard drive seek sounds
	numSeeks := 20 + rng.Intn(15)
	for s := 0; s < numSeeks; s++ {
		seekStart := rng.Intn(len(pcmData)/2 - sampleRate/5)
		seekLen := sampleRate / 30

		for i := 0; i < seekLen; i++ {
			env := math.Exp(-float64(i) / float64(seekLen/4))
			freq := 400.0 + float64(i)*10.0
			val := math.Sin(2*math.Pi*freq*float64(i)/float64(sampleRate)) * env * 1200.0

			idx := (seekStart + i) * 2
			if idx+1 < len(pcmData) {
				pcmData[idx] += int16(val)
				pcmData[idx+1] += int16(val)
			}
		}
	}
}

// generateWind creates post-apocalyptic atmosphere with wind and debris.
func (a *AmbientSoundscape) generateWind(pcmData []int16, rng *rand.Rand) {
	// Wind base layer using filtered noise
	for i := 0; i < len(pcmData)/2; i++ {
		t := float64(i) / float64(sampleRate)

		// Low-pass filtered noise for wind
		noise := (rng.Float64()*2.0 - 1.0) * 0.4

		// Wind gusts modulation
		gustMod := 1.0 + math.Sin(2*math.Pi*0.1*t)*0.3 + math.Sin(2*math.Pi*0.05*t)*0.2

		val := int16(noise * gustMod * 2500.0)
		pcmData[i*2] += val
		pcmData[i*2+1] += val
	}

	// Random debris sounds (creaking metal, shifting rubble)
	numDebris := 10 + rng.Intn(8)
	for d := 0; d < numDebris; d++ {
		debrisStart := rng.Intn(len(pcmData)/2 - sampleRate)
		debrisLen := sampleRate / 8

		for i := 0; i < debrisLen; i++ {
			t := float64(i) / float64(debrisLen)
			env := math.Sin(t*math.Pi) * 0.5

			// Metallic creak sound
			freq := 80.0 + t*120.0
			creak := math.Sin(2 * math.Pi * freq * float64(i) / float64(sampleRate))
			noise := (rng.Float64()*2.0 - 1.0) * 0.3

			val := (creak*0.7 + noise) * env * 1800.0

			idx := (debrisStart + i) * 2
			if idx+1 < len(pcmData) {
				pcmData[idx] += int16(val)
				pcmData[idx+1] += int16(val)
			}
		}
	}
}

// generateGenericAmbient creates a generic ambient soundscape for unknown genres.
func (a *AmbientSoundscape) generateGenericAmbient(pcmData []int16, rng *rand.Rand) {
	// Simple low-frequency drone
	for i := 0; i < len(pcmData)/2; i++ {
		t := float64(i) / float64(sampleRate)

		drone := math.Sin(2*math.Pi*50.0*t) * 0.2
		drone += math.Sin(2*math.Pi*75.0*t) * 0.1

		val := int16(drone * 2000.0)
		pcmData[i*2] += val
		pcmData[i*2+1] += val
	}
}
