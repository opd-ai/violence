package audio

import (
	"bytes"
	"testing"

	"github.com/opd-ai/violence/pkg/bsp"
)

func TestNewReverbCalculator(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"small room", 10, 10},
		{"medium room", 25, 25},
		{"large room", 50, 50},
		{"rectangular narrow", 10, 50},
		{"rectangular wide", 50, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewReverbCalculator(tt.width, tt.height)

			if calc == nil {
				t.Fatal("NewReverbCalculator returned nil")
			}
			if calc.roomWidth != tt.width {
				t.Errorf("roomWidth = %v, want %v", calc.roomWidth, tt.width)
			}
			if calc.roomHeight != tt.height {
				t.Errorf("roomHeight = %v, want %v", calc.roomHeight, tt.height)
			}
			if calc.decay < 0.0 || calc.decay > 1.0 {
				t.Errorf("decay out of range: %v", calc.decay)
			}
			if calc.wetMix < 0.0 || calc.wetMix > 1.0 {
				t.Errorf("wetMix out of range: %v", calc.wetMix)
			}
			if calc.dryMix < 0.0 || calc.dryMix > 1.0 {
				t.Errorf("dryMix out of range: %v", calc.dryMix)
			}
		})
	}
}

func TestReverbCalculator_SetRoomSize(t *testing.T) {
	calc := NewReverbCalculator(10, 10)

	originalDecay := calc.decay
	originalWet := calc.wetMix

	// Change to larger room
	calc.SetRoomSize(50, 50)

	if calc.roomWidth != 50 {
		t.Errorf("roomWidth = %v, want 50", calc.roomWidth)
	}
	if calc.roomHeight != 50 {
		t.Errorf("roomHeight = %v, want 50", calc.roomHeight)
	}

	// Larger room should have more reverb
	if calc.decay <= originalDecay {
		t.Errorf("decay did not increase: %v -> %v", originalDecay, calc.decay)
	}
	if calc.wetMix <= originalWet {
		t.Errorf("wetMix did not increase: %v -> %v", originalWet, calc.wetMix)
	}
}

func TestReverbCalculator_GetDecay(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		minDecay float64
		maxDecay float64
	}{
		{"small room low decay", 10, 10, 0.1, 0.3},
		{"medium room mid decay", 25, 25, 0.2, 0.5},
		{"large room high decay", 50, 50, 0.5, 0.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewReverbCalculator(tt.width, tt.height)
			decay := calc.GetDecay()

			if decay < tt.minDecay || decay > tt.maxDecay {
				t.Errorf("decay = %v, want between %v and %v", decay, tt.minDecay, tt.maxDecay)
			}
		})
	}
}

func TestReverbCalculator_GetWetMix(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		minWet float64
		maxWet float64
	}{
		{"small room low wet", 10, 10, 0.0, 0.1},
		{"medium room mid wet", 25, 25, 0.1, 0.3},
		{"large room high wet", 50, 50, 0.3, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewReverbCalculator(tt.width, tt.height)
			wet := calc.GetWetMix()

			if wet < tt.minWet || wet > tt.maxWet {
				t.Errorf("wetMix = %v, want between %v and %v", wet, tt.minWet, tt.maxWet)
			}
		})
	}
}

func TestReverbCalculator_GetDryMix(t *testing.T) {
	calc := NewReverbCalculator(25, 25)
	dry := calc.GetDryMix()

	if dry < 0.0 || dry > 1.0 {
		t.Errorf("dryMix = %v, want between 0.0 and 1.0", dry)
	}

	// Dry mix should always be fairly strong
	if dry < 0.5 {
		t.Errorf("dryMix = %v, want >= 0.5", dry)
	}
}

func TestReverbCalculator_RoomSizeProgression(t *testing.T) {
	// Test that reverb increases with room size
	sizes := []struct {
		width  int
		height int
	}{
		{10, 10},
		{20, 20},
		{30, 30},
		{40, 40},
		{50, 50},
	}

	var prevDecay, prevWet float64

	for i, size := range sizes {
		calc := NewReverbCalculator(size.width, size.height)

		if i > 0 {
			if calc.decay <= prevDecay {
				t.Errorf("decay not increasing: room %dx%d decay=%v, prev=%v",
					size.width, size.height, calc.decay, prevDecay)
			}
			if calc.wetMix <= prevWet {
				t.Errorf("wetMix not increasing: room %dx%d wet=%v, prev=%v",
					size.width, size.height, calc.wetMix, prevWet)
			}
		}

		prevDecay = calc.decay
		prevWet = calc.wetMix
	}
}

func TestReverbCalculator_Apply(t *testing.T) {
	t.Run("applies reverb to audio data", func(t *testing.T) {
		calc := NewReverbCalculator(30, 30)

		// Generate test audio
		original := generateBlip(sampleRate / 10)

		// Apply reverb
		processed := calc.Apply(original)

		if len(processed) != len(original) {
			t.Errorf("processed length = %v, want %v", len(processed), len(original))
		}

		// Should have WAV header
		if !bytes.Equal(processed[0:4], []byte("RIFF")) {
			t.Error("missing RIFF header in processed audio")
		}
	})

	t.Run("handles short data", func(t *testing.T) {
		calc := NewReverbCalculator(20, 20)
		short := []byte{1, 2, 3}

		result := calc.Apply(short)

		if !bytes.Equal(result, short) {
			t.Error("short data was modified")
		}
	})

	t.Run("handles empty data", func(t *testing.T) {
		calc := NewReverbCalculator(20, 20)
		empty := []byte{}

		result := calc.Apply(empty)

		if len(result) != 0 {
			t.Error("empty data was not handled correctly")
		}
	})
}

func TestReverbCalculator_ApplyDeterminism(t *testing.T) {
	calc := NewReverbCalculator(25, 25)
	original := generateBlip(sampleRate / 20)

	result1 := calc.Apply(original)
	result2 := calc.Apply(original)

	if !bytes.Equal(result1, result2) {
		t.Error("reverb application is non-deterministic")
	}
}

func TestReverbCalculator_ApplyEffectiveness(t *testing.T) {
	// Test that larger rooms produce more audible reverb
	smallRoom := NewReverbCalculator(10, 10)
	largeRoom := NewReverbCalculator(50, 50)

	original := generateBlip(sampleRate / 10)

	smallResult := smallRoom.Apply(original)
	largeResult := largeRoom.Apply(original)

	// Check that reverb parameters are different
	if smallRoom.decay >= largeRoom.decay {
		t.Errorf("large room should have more decay: small=%v, large=%v",
			smallRoom.decay, largeRoom.decay)
	}
	if smallRoom.wetMix >= largeRoom.wetMix {
		t.Errorf("large room should have more wet mix: small=%v, large=%v",
			smallRoom.wetMix, largeRoom.wetMix)
	}

	// Results should have same length
	if len(smallResult) != len(largeResult) {
		t.Errorf("result lengths differ: small=%v, large=%v",
			len(smallResult), len(largeResult))
	}
}

func TestReverbCalculator_Calculate(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"minimum room", 5, 5},
		{"small room", 10, 10},
		{"medium room", 25, 25},
		{"large room", 50, 50},
		{"huge room", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := &ReverbCalculator{
				roomWidth:  tt.width,
				roomHeight: tt.height,
			}
			calc.calculate()

			// All values should be in valid ranges
			if calc.decay < 0.1 || calc.decay > 0.8 {
				t.Errorf("decay = %v, want between 0.1 and 0.8", calc.decay)
			}
			if calc.wetMix < 0.0 || calc.wetMix > 0.5 {
				t.Errorf("wetMix = %v, want between 0.0 and 0.5", calc.wetMix)
			}
			if calc.dryMix < 0.8 || calc.dryMix > 1.0 {
				t.Errorf("dryMix = %v, want between 0.8 and 1.0", calc.dryMix)
			}
		})
	}
}

func TestReverbCalculator_ApplyReverb(t *testing.T) {
	calc := NewReverbCalculator(30, 30)

	// Create simple test signal - impulse
	samples := make([]int16, 10000)
	for i := 0; i < 100; i++ {
		samples[i] = 10000 // Impulse at start
	}

	output := calc.applyReverb(samples)

	if len(output) != len(samples) {
		t.Errorf("output length = %v, want %v", len(output), len(samples))
	}

	// Check that the impulse is preserved (dry signal)
	hasInitialSignal := false
	for i := 0; i < 100; i++ {
		if output[i] != 0 {
			hasInitialSignal = true
			break
		}
	}

	if !hasInitialSignal {
		t.Error("reverb removed initial signal")
	}

	// With decay and feedback, we should see some delayed echo
	// The wet signal adds delayed samples back in
	hasDelayedSignal := false
	delayTime := 0.02 + calc.decay*0.08
	delaySamples := int(delayTime * float64(sampleRate))
	if delaySamples < 100 {
		delaySamples = 100
	}

	// Check for signal after the delay period
	checkStart := delaySamples + 100
	checkEnd := checkStart + 500
	if checkEnd > len(output) {
		checkEnd = len(output)
	}

	for i := checkStart; i < checkEnd; i++ {
		if output[i] != 0 {
			hasDelayedSignal = true
			break
		}
	}

	if !hasDelayedSignal {
		t.Error("reverb did not produce delayed echo")
	}
}

func TestReverbCalculator_ApplyReverbEmptySamples(t *testing.T) {
	calc := NewReverbCalculator(20, 20)

	empty := []int16{}
	result := calc.applyReverb(empty)

	if len(result) != 0 {
		t.Error("empty samples produced output")
	}
}

func BenchmarkReverbCalculator_Calculate(b *testing.B) {
	calc := &ReverbCalculator{roomWidth: 30, roomHeight: 30}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calc.calculate()
	}
}

func BenchmarkReverbCalculator_Apply(b *testing.B) {
	calc := NewReverbCalculator(30, 30)
	audio := generateBlip(sampleRate / 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calc.Apply(audio)
	}
}

func BenchmarkReverbCalculator_ApplyReverb(b *testing.B) {
	calc := NewReverbCalculator(30, 30)
	samples := make([]int16, sampleRate/10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calc.applyReverb(samples)
	}
}

func TestReverbCalculator_SetRoomFromBSP(t *testing.T) {
	calc := NewReverbCalculator(10, 10)
	initialDecay := calc.GetDecay()

	t.Run("sets room dimensions from BSP room", func(t *testing.T) {
		room := &bsp.Room{X: 5, Y: 5, W: 30, H: 25}
		calc.SetRoomFromBSP(room)

		if calc.roomWidth != 30 {
			t.Errorf("roomWidth = %v, want 30", calc.roomWidth)
		}
		if calc.roomHeight != 25 {
			t.Errorf("roomHeight = %v, want 25", calc.roomHeight)
		}

		// Larger room should have more decay
		if calc.GetDecay() <= initialDecay {
			t.Errorf("decay did not increase: %v -> %v", initialDecay, calc.GetDecay())
		}
	})

	t.Run("handles nil room gracefully", func(t *testing.T) {
		calc := NewReverbCalculator(10, 10)
		origWidth := calc.roomWidth
		origHeight := calc.roomHeight

		calc.SetRoomFromBSP(nil)

		if calc.roomWidth != origWidth || calc.roomHeight != origHeight {
			t.Error("nil room changed dimensions")
		}
	})

	t.Run("recalculates reverb parameters", func(t *testing.T) {
		calc := NewReverbCalculator(10, 10)
		smallDecay := calc.GetDecay()
		smallWet := calc.GetWetMix()

		largeRoom := &bsp.Room{X: 0, Y: 0, W: 50, H: 50}
		calc.SetRoomFromBSP(largeRoom)

		if calc.GetDecay() <= smallDecay {
			t.Errorf("decay not increased for large room")
		}
		if calc.GetWetMix() <= smallWet {
			t.Errorf("wet mix not increased for large room")
		}
	})
}

func TestReverbCalculator_BSPRoomSizeMapping(t *testing.T) {
	tests := []struct {
		name     string
		room     *bsp.Room
		minDecay float64
		maxDecay float64
	}{
		{"small BSP room", &bsp.Room{X: 0, Y: 0, W: 8, H: 8}, 0.1, 0.2},
		{"medium BSP room", &bsp.Room{X: 0, Y: 0, W: 20, H: 20}, 0.2, 0.4},
		{"large BSP room", &bsp.Room{X: 0, Y: 0, W: 40, H: 40}, 0.4, 0.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewReverbCalculator(10, 10)
			calc.SetRoomFromBSP(tt.room)

			decay := calc.GetDecay()
			if decay < tt.minDecay || decay > tt.maxDecay {
				t.Errorf("decay = %v, want between %v and %v", decay, tt.minDecay, tt.maxDecay)
			}
		})
	}
}
