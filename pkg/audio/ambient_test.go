package audio

import (
	"bytes"
	"testing"
)

func TestNewAmbientSoundscape(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		seed    uint64
	}{
		{"fantasy soundscape", "fantasy", 12345},
		{"scifi soundscape", "scifi", 67890},
		{"horror soundscape", "horror", 11111},
		{"cyberpunk soundscape", "cyberpunk", 22222},
		{"postapoc soundscape", "postapoc", 33333},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ambient := NewAmbientSoundscape(tt.genreID, tt.seed)

			if ambient == nil {
				t.Fatal("NewAmbientSoundscape returned nil")
			}
			if ambient.genreID != tt.genreID {
				t.Errorf("genreID = %v, want %v", ambient.genreID, tt.genreID)
			}
			if ambient.seed != tt.seed {
				t.Errorf("seed = %v, want %v", ambient.seed, tt.seed)
			}
			if ambient.duration != sampleRate*60 {
				t.Errorf("duration = %v, want %v", ambient.duration, sampleRate*60)
			}
		})
	}
}

func TestAmbientSoundscape_Generate(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		seed    uint64
	}{
		{"fantasy dungeon echo", "fantasy", 1111},
		{"scifi station hum", "scifi", 2222},
		{"horror hospital silence", "horror", 3333},
		{"cyberpunk server drone", "cyberpunk", 4444},
		{"postapoc wind", "postapoc", 5555},
		{"unknown genre", "unknown", 6666},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ambient := NewAmbientSoundscape(tt.genreID, tt.seed)
			ambient.Generate()

			if ambient.loopData == nil {
				t.Fatal("Generate did not create loop data")
			}

			// Check WAV header is present
			if len(ambient.loopData) < 44 {
				t.Errorf("loop data too short: %v bytes", len(ambient.loopData))
			}

			// Verify WAV header
			if !bytes.Equal(ambient.loopData[0:4], []byte("RIFF")) {
				t.Error("missing RIFF header")
			}
			if !bytes.Equal(ambient.loopData[8:12], []byte("WAVE")) {
				t.Error("missing WAVE header")
			}

			// Expected data size: 44 byte header + (duration * 4 bytes per stereo sample)
			expectedSize := 44 + ambient.duration*4
			if len(ambient.loopData) != expectedSize {
				t.Errorf("loop data size = %v, want %v", len(ambient.loopData), expectedSize)
			}
		})
	}
}

func TestAmbientSoundscape_GetLoopData(t *testing.T) {
	t.Run("generates on first call", func(t *testing.T) {
		ambient := NewAmbientSoundscape("fantasy", 7777)

		data := ambient.GetLoopData()
		if data == nil {
			t.Fatal("GetLoopData returned nil")
		}
		if len(data) < 44 {
			t.Error("GetLoopData returned data too short")
		}
	})

	t.Run("returns cached data on subsequent calls", func(t *testing.T) {
		ambient := NewAmbientSoundscape("scifi", 8888)

		data1 := ambient.GetLoopData()
		data2 := ambient.GetLoopData()

		if len(data1) != len(data2) {
			t.Error("GetLoopData returned different sized data")
		}

		// Should be the same underlying array
		if &data1[0] != &data2[0] {
			t.Error("GetLoopData did not return cached data")
		}
	})
}

func TestAmbientSoundscape_SetGenre(t *testing.T) {
	t.Run("changes genre and clears cache", func(t *testing.T) {
		ambient := NewAmbientSoundscape("fantasy", 9999)
		ambient.Generate()

		if ambient.loopData == nil {
			t.Fatal("initial generation failed")
		}

		ambient.SetGenre("scifi")

		if ambient.genreID != "scifi" {
			t.Errorf("genreID = %v, want scifi", ambient.genreID)
		}
		if ambient.loopData != nil {
			t.Error("loopData was not cleared after genre change")
		}
	})

	t.Run("same genre does not clear cache", func(t *testing.T) {
		ambient := NewAmbientSoundscape("horror", 1234)
		ambient.Generate()

		originalData := ambient.loopData
		ambient.SetGenre("horror")

		if ambient.loopData == nil {
			t.Error("loopData was cleared even though genre did not change")
		}
		if len(ambient.loopData) != len(originalData) {
			t.Error("loopData was regenerated even though genre did not change")
		}
	})
}

func TestAmbientSoundscape_Determinism(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		seed    uint64
	}{
		{"fantasy deterministic", "fantasy", 11111},
		{"scifi deterministic", "scifi", 22222},
		{"horror deterministic", "horror", 33333},
		{"cyberpunk deterministic", "cyberpunk", 44444},
		{"postapoc deterministic", "postapoc", 55555},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ambient1 := NewAmbientSoundscape(tt.genreID, tt.seed)
			ambient2 := NewAmbientSoundscape(tt.genreID, tt.seed)

			data1 := ambient1.GetLoopData()
			data2 := ambient2.GetLoopData()

			if len(data1) != len(data2) {
				t.Fatalf("data lengths differ: %v vs %v", len(data1), len(data2))
			}

			if !bytes.Equal(data1, data2) {
				t.Error("same seed produced different audio data (non-deterministic)")
			}
		})
	}
}

func TestAmbientSoundscape_GenreUniqueness(t *testing.T) {
	seed := uint64(99999)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	genreData := make(map[string][]byte)

	for _, genre := range genres {
		ambient := NewAmbientSoundscape(genre, seed)
		genreData[genre] = ambient.GetLoopData()
	}

	// Verify each genre produces unique output
	for i, genre1 := range genres {
		for j, genre2 := range genres {
			if i >= j {
				continue
			}

			data1 := genreData[genre1]
			data2 := genreData[genre2]

			if bytes.Equal(data1, data2) {
				t.Errorf("%s and %s produced identical audio (should be genre-specific)", genre1, genre2)
			}
		}
	}
}

func TestAmbientSoundscape_DungeonEcho(t *testing.T) {
	ambient := NewAmbientSoundscape("fantasy", 12345)
	data := ambient.GetLoopData()

	if len(data) < 44 {
		t.Fatal("data too short")
	}

	// Verify audio contains non-zero samples (not silent)
	nonZeroCount := 0
	for i := 44; i < len(data); i += 2 {
		if data[i] != 0 || data[i+1] != 0 {
			nonZeroCount++
		}
	}

	if nonZeroCount == 0 {
		t.Error("dungeon echo generated silent audio")
	}
}

func TestAmbientSoundscape_StationHum(t *testing.T) {
	ambient := NewAmbientSoundscape("scifi", 23456)
	data := ambient.GetLoopData()

	if len(data) < 44 {
		t.Fatal("data too short")
	}

	nonZeroCount := 0
	for i := 44; i < len(data); i += 2 {
		if data[i] != 0 || data[i+1] != 0 {
			nonZeroCount++
		}
	}

	if nonZeroCount == 0 {
		t.Error("station hum generated silent audio")
	}
}

func TestAmbientSoundscape_HospitalSilence(t *testing.T) {
	ambient := NewAmbientSoundscape("horror", 34567)
	data := ambient.GetLoopData()

	if len(data) < 44 {
		t.Fatal("data too short")
	}

	nonZeroCount := 0
	for i := 44; i < len(data); i += 2 {
		if data[i] != 0 || data[i+1] != 0 {
			nonZeroCount++
		}
	}

	if nonZeroCount == 0 {
		t.Error("hospital silence generated silent audio")
	}
}

func TestAmbientSoundscape_ServerDrone(t *testing.T) {
	ambient := NewAmbientSoundscape("cyberpunk", 45678)
	data := ambient.GetLoopData()

	if len(data) < 44 {
		t.Fatal("data too short")
	}

	nonZeroCount := 0
	for i := 44; i < len(data); i += 2 {
		if data[i] != 0 || data[i+1] != 0 {
			nonZeroCount++
		}
	}

	if nonZeroCount == 0 {
		t.Error("server drone generated silent audio")
	}
}

func TestAmbientSoundscape_Wind(t *testing.T) {
	ambient := NewAmbientSoundscape("postapoc", 56789)
	data := ambient.GetLoopData()

	if len(data) < 44 {
		t.Fatal("data too short")
	}

	nonZeroCount := 0
	for i := 44; i < len(data); i += 2 {
		if data[i] != 0 || data[i+1] != 0 {
			nonZeroCount++
		}
	}

	if nonZeroCount == 0 {
		t.Error("wind generated silent audio")
	}
}

func BenchmarkAmbientSoundscape_Generate(b *testing.B) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		b.Run(genre, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ambient := NewAmbientSoundscape(genre, uint64(i))
				ambient.Generate()
			}
		})
	}
}

func BenchmarkAmbientSoundscape_GetLoopData(b *testing.B) {
	ambient := NewAmbientSoundscape("fantasy", 12345)
	ambient.Generate() // Pre-generate to test cached retrieval

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ambient.GetLoopData()
	}
}
