// Package audio manages sound effects and music playback with adaptive music layers.
package audio

import (
	"bytes"
	"math"
)

// GenerateReloadSound creates a genre-specific weapon reload sound.
func GenerateReloadSound(genreID string, seed uint64) []byte {
	rng := newLocalRNG(seed)
	samples := sampleRate / 5
	buf := &bytes.Buffer{}
	writeWAVHeader(buf, samples)

	// Genre-specific parameters
	clickSharpness := 1.0
	metallic := 0.5
	mechanicalNoise := 1.0

	switch genreID {
	case "fantasy":
		// Wood and leather sounds - softer, less sharp
		clickSharpness = 0.6
		metallic = 0.3
		mechanicalNoise = 0.7
	case "scifi":
		// High-tech electronic sounds - crisp and precise
		clickSharpness = 1.5
		metallic = 0.8
		mechanicalNoise = 0.9
	case "horror":
		// Rusty, grinding sounds - harsh and disturbing
		clickSharpness = 0.8
		metallic = 1.2
		mechanicalNoise = 1.5
	case "cyberpunk":
		// Sleek electronic sounds with digital clicks
		clickSharpness = 1.3
		metallic = 0.7
		mechanicalNoise = 0.6
	case "postapoc":
		// Makeshift, rattling sounds - rough and unreliable
		clickSharpness = 0.9
		metallic = 1.0
		mechanicalNoise = 1.3
	}

	for i := 0; i < samples; i++ {
		env := 0.0

		// Main click at start
		if i < samples/10 {
			env = math.Exp(-float64(i) / float64(samples/50) * clickSharpness)
		}

		// Secondary mechanical sounds
		if i > samples/8 && i < samples/4 {
			t := float64(i-samples/8) / float64(samples/8)
			env += math.Exp(-t*3.0) * 0.3 * mechanicalNoise
		}

		// Generate noise with metallic character
		noise := (rng.Float64()*2.0 - 1.0) * env

		// Add metallic ringing for some genres
		if metallic > 0.5 {
			ringFreq := 2000.0 + float64(rng.Intn(1000))
			ring := math.Sin(2*math.Pi*ringFreq*float64(i)/float64(sampleRate)) * env * (metallic - 0.5)
			noise += ring * 0.3
		}

		val := int16(noise * 15000.0)
		writeInt16(buf, val)
		writeInt16(buf, val)
	}

	return buf.Bytes()
}

// GenerateEmptyClickSound creates a genre-specific empty weapon click sound.
func GenerateEmptyClickSound(genreID string, seed uint64) []byte {
	rng := newLocalRNG(seed)
	samples := sampleRate / 20 // Shorter than reload
	buf := &bytes.Buffer{}
	writeWAVHeader(buf, samples)

	// Genre-specific parameters
	clickPitch := 1.0
	dryness := 1.0

	switch genreID {
	case "fantasy":
		// Wooden click - lower pitch, dryer
		clickPitch = 0.7
		dryness = 1.2
	case "scifi":
		// Electronic beep - higher pitch, clean
		clickPitch = 1.5
		dryness = 0.8
	case "horror":
		// Disturbing mechanical failure - mid pitch, harsh
		clickPitch = 0.9
		dryness = 1.5
	case "cyberpunk":
		// Digital error sound - high pitch, crisp
		clickPitch = 1.4
		dryness = 0.7
	case "postapoc":
		// Rusty click - low pitch, rattling
		clickPitch = 0.8
		dryness = 1.3
	}

	for i := 0; i < samples; i++ {
		// Very short, sharp click
		env := math.Exp(-float64(i) / float64(samples/8) * dryness)

		// Brief burst of noise
		noise := (rng.Float64()*2.0 - 1.0) * env

		// Add a brief tone for clarity
		freq := 1200.0 * clickPitch
		tone := math.Sin(2*math.Pi*freq*float64(i)/float64(sampleRate)) * env * 0.4

		val := int16((noise*0.6 + tone*0.4) * 12000.0)
		writeInt16(buf, val)
		writeInt16(buf, val)
	}

	return buf.Bytes()
}

// GeneratePickupJingleSound creates a genre-specific item pickup sound.
func GeneratePickupJingleSound(genreID string, seed uint64) []byte {
	samples := sampleRate / 6
	buf := &bytes.Buffer{}
	writeWAVHeader(buf, samples)

	// Genre-specific parameters
	notes := []float64{440.0, 554.37, 659.25} // Default: A, C#, E (major chord)
	brightness := 1.0

	switch genreID {
	case "fantasy":
		// Magical chime - perfect fifth, bright
		notes = []float64{440.0, 587.33, 659.25} // A, D, E
		brightness = 1.3
	case "scifi":
		// Electronic beep sequence - chromatic, clean
		notes = []float64{523.25, 587.33, 698.46} // C, D, F
		brightness = 0.9
	case "horror":
		// Unsettling tone - dissonant, dark
		notes = []float64{415.30, 466.16, 493.88} // G#, A#, B (diminished)
		brightness = 0.6
	case "cyberpunk":
		// Digital confirmation - high, synthesized
		notes = []float64{659.25, 783.99, 880.0} // E, G, A
		brightness = 1.1
	case "postapoc":
		// Metallic clink - imperfect, dull
		notes = []float64{392.0, 440.0, 493.88} // G, A, B
		brightness = 0.7
	}

	for i := 0; i < samples; i++ {
		t := float64(i) / float64(samples)
		env := math.Exp(-t * 3.0) // Fast decay

		val := 0.0

		// Play notes in sequence with overlap
		for n, freq := range notes {
			noteStart := float64(n) * 0.15
			if t >= noteStart {
				noteT := t - noteStart
				noteEnv := math.Exp(-noteT * 4.0)

				// Fundamental + harmonic for brightness
				fundamental := math.Sin(2 * math.Pi * freq * float64(i) / float64(sampleRate))
				harmonic := math.Sin(2 * math.Pi * freq * 2.0 * float64(i) / float64(sampleRate))

				val += (fundamental + harmonic*brightness*0.3) * noteEnv * 0.4
			}
		}

		val *= env * 12000.0

		writeInt16(buf, int16(val))
		writeInt16(buf, int16(val))
	}

	return buf.Bytes()
}
