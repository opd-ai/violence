// Package audio manages sound effects and music playback with adaptive music layers.
package audio

import (
	"bytes"
	"math"
)

// ReverbCalculator computes reverb parameters based on room dimensions.
type ReverbCalculator struct {
	roomWidth  int
	roomHeight int
	decay      float64 // 0.0-1.0, how long reverb lasts
	wetMix     float64 // 0.0-1.0, how much reverb to add
	dryMix     float64 // 0.0-1.0, how much original signal
}

// NewReverbCalculator creates a reverb calculator for a room.
// Width and height are in grid tiles.
func NewReverbCalculator(width, height int) *ReverbCalculator {
	r := &ReverbCalculator{
		roomWidth:  width,
		roomHeight: height,
	}
	r.calculate()
	return r
}

// SetRoomSize updates room dimensions and recalculates parameters.
func (r *ReverbCalculator) SetRoomSize(width, height int) {
	r.roomWidth = width
	r.roomHeight = height
	r.calculate()
}

// GetDecay returns the reverb decay time (0.0-1.0).
func (r *ReverbCalculator) GetDecay() float64 {
	return r.decay
}

// GetWetMix returns the wet signal level (0.0-1.0).
func (r *ReverbCalculator) GetWetMix() float64 {
	return r.wetMix
}

// GetDryMix returns the dry signal level (0.0-1.0).
func (r *ReverbCalculator) GetDryMix() float64 {
	return r.dryMix
}

// calculate computes reverb parameters from room dimensions.
// Larger rooms = longer decay and more wet signal.
func (r *ReverbCalculator) calculate() {
	// Room area determines reverb characteristics
	area := float64(r.roomWidth * r.roomHeight)

	// Normalize area to reasonable reverb range
	// Small room (10x10=100) -> minimal reverb
	// Large room (50x50=2500) -> strong reverb
	normalizedArea := math.Min(area/2500.0, 1.0)

	// Decay increases with room size (0.1 to 0.8)
	r.decay = 0.1 + normalizedArea*0.7

	// Wet mix increases with room size (0.0 to 0.5)
	r.wetMix = normalizedArea * 0.5

	// Dry mix is always strong, but slightly reduced in large rooms
	r.dryMix = 1.0 - normalizedArea*0.2
}

// Apply applies reverb to audio data.
// Returns new audio data with reverb applied.
func (r *ReverbCalculator) Apply(audioData []byte) []byte {
	// Skip header, work on PCM data only
	if len(audioData) < 44 {
		return audioData
	}

	header := audioData[0:44]
	pcm := audioData[44:]

	// Convert bytes to int16 samples
	samples := make([]int16, len(pcm)/2)
	for i := 0; i < len(samples); i++ {
		low := uint16(pcm[i*2])
		high := uint16(pcm[i*2+1])
		samples[i] = int16(low | (high << 8))
	}

	// Apply simple reverb using comb filter
	output := r.applyReverb(samples)

	// Convert back to bytes
	buf := bytes.NewBuffer(header)
	for i := 0; i < len(output); i++ {
		writeInt16(buf, output[i])
	}

	return buf.Bytes()
}

// applyReverb applies a simple comb filter reverb to samples.
func (r *ReverbCalculator) applyReverb(samples []int16) []int16 {
	if len(samples) == 0 {
		return samples
	}

	// Calculate delay based on room size (in samples)
	// Larger rooms have longer delays
	delayTime := 0.02 + r.decay*0.08 // 20-100ms
	delaySamples := int(delayTime * float64(sampleRate))

	// Ensure stereo alignment (even number of samples)
	if delaySamples%2 != 0 {
		delaySamples++
	}

	// Minimum delay to ensure reverb effect
	if delaySamples < 100 {
		delaySamples = 100
	}

	output := make([]int16, len(samples))
	delayBuffer := make([]int16, delaySamples)
	delayIndex := 0

	for i := 0; i < len(samples); i++ {
		// Get delayed sample
		delayed := delayBuffer[delayIndex]

		// Mix dry and wet signals
		dry := float64(samples[i]) * r.dryMix
		wet := float64(delayed) * r.wetMix

		output[i] = int16(clamp(dry+wet, -32768.0, 32767.0))

		// Store in delay buffer with feedback
		feedback := float64(samples[i]) + float64(delayed)*r.decay
		delayBuffer[delayIndex] = int16(clamp(feedback, -32768.0, 32767.0))

		delayIndex = (delayIndex + 1) % delaySamples
	}

	return output
}
