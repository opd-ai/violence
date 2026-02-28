package audio

import (
	"bytes"
	"testing"
)

func TestGenerateReloadSound(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		seed    uint64
	}{
		{"fantasy reload", "fantasy", 11111},
		{"scifi reload", "scifi", 22222},
		{"horror reload", "horror", 33333},
		{"cyberpunk reload", "cyberpunk", 44444},
		{"postapoc reload", "postapoc", 55555},
		{"unknown genre reload", "unknown", 66666},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := GenerateReloadSound(tt.genreID, tt.seed)

			if len(data) < 44 {
				t.Fatal("data too short")
			}

			// Verify WAV header
			if !bytes.Equal(data[0:4], []byte("RIFF")) {
				t.Error("missing RIFF header")
			}
			if !bytes.Equal(data[8:12], []byte("WAVE")) {
				t.Error("missing WAVE header")
			}

			// Verify audio has content
			nonZeroCount := 0
			for i := 44; i < len(data); i += 2 {
				if data[i] != 0 || data[i+1] != 0 {
					nonZeroCount++
				}
			}

			if nonZeroCount == 0 {
				t.Error("reload sound is silent")
			}
		})
	}
}

func TestGenerateReloadSound_Determinism(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			seed := uint64(12345)

			data1 := GenerateReloadSound(genre, seed)
			data2 := GenerateReloadSound(genre, seed)

			if !bytes.Equal(data1, data2) {
				t.Error("reload sound is non-deterministic")
			}
		})
	}
}

func TestGenerateReloadSound_GenreUniqueness(t *testing.T) {
	seed := uint64(99999)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	genreData := make(map[string][]byte)
	for _, genre := range genres {
		genreData[genre] = GenerateReloadSound(genre, seed)
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
				t.Errorf("%s and %s produced identical reload sounds", genre1, genre2)
			}
		}
	}
}

func TestGenerateEmptyClickSound(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		seed    uint64
	}{
		{"fantasy click", "fantasy", 11111},
		{"scifi click", "scifi", 22222},
		{"horror click", "horror", 33333},
		{"cyberpunk click", "cyberpunk", 44444},
		{"postapoc click", "postapoc", 55555},
		{"unknown genre click", "unknown", 66666},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := GenerateEmptyClickSound(tt.genreID, tt.seed)

			if len(data) < 44 {
				t.Fatal("data too short")
			}

			// Verify WAV header
			if !bytes.Equal(data[0:4], []byte("RIFF")) {
				t.Error("missing RIFF header")
			}

			// Verify audio has content
			nonZeroCount := 0
			for i := 44; i < len(data); i += 2 {
				if data[i] != 0 || data[i+1] != 0 {
					nonZeroCount++
				}
			}

			if nonZeroCount == 0 {
				t.Error("empty click sound is silent")
			}

			// Empty click should be shorter than reload
			reloadData := GenerateReloadSound(tt.genreID, tt.seed)
			if len(data) >= len(reloadData) {
				t.Error("empty click should be shorter than reload")
			}
		})
	}
}

func TestGenerateEmptyClickSound_Determinism(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			seed := uint64(12345)

			data1 := GenerateEmptyClickSound(genre, seed)
			data2 := GenerateEmptyClickSound(genre, seed)

			if !bytes.Equal(data1, data2) {
				t.Error("empty click sound is non-deterministic")
			}
		})
	}
}

func TestGenerateEmptyClickSound_GenreUniqueness(t *testing.T) {
	seed := uint64(99999)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	genreData := make(map[string][]byte)
	for _, genre := range genres {
		genreData[genre] = GenerateEmptyClickSound(genre, seed)
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
				t.Errorf("%s and %s produced identical empty click sounds", genre1, genre2)
			}
		}
	}
}

func TestGeneratePickupJingleSound(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		seed    uint64
	}{
		{"fantasy jingle", "fantasy", 11111},
		{"scifi jingle", "scifi", 22222},
		{"horror jingle", "horror", 33333},
		{"cyberpunk jingle", "cyberpunk", 44444},
		{"postapoc jingle", "postapoc", 55555},
		{"unknown genre jingle", "unknown", 66666},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := GeneratePickupJingleSound(tt.genreID, tt.seed)

			if len(data) < 44 {
				t.Fatal("data too short")
			}

			// Verify WAV header
			if !bytes.Equal(data[0:4], []byte("RIFF")) {
				t.Error("missing RIFF header")
			}

			// Verify audio has content
			nonZeroCount := 0
			for i := 44; i < len(data); i += 2 {
				if data[i] != 0 || data[i+1] != 0 {
					nonZeroCount++
				}
			}

			if nonZeroCount == 0 {
				t.Error("pickup jingle is silent")
			}
		})
	}
}

func TestGeneratePickupJingleSound_Determinism(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			seed := uint64(12345)

			data1 := GeneratePickupJingleSound(genre, seed)
			data2 := GeneratePickupJingleSound(genre, seed)

			if !bytes.Equal(data1, data2) {
				t.Error("pickup jingle is non-deterministic")
			}
		})
	}
}

func TestGeneratePickupJingleSound_GenreUniqueness(t *testing.T) {
	seed := uint64(99999)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	genreData := make(map[string][]byte)
	for _, genre := range genres {
		genreData[genre] = GeneratePickupJingleSound(genre, seed)
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
				t.Errorf("%s and %s produced identical pickup jingles", genre1, genre2)
			}
		}
	}
}

func BenchmarkGenerateReloadSound(b *testing.B) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		b.Run(genre, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = GenerateReloadSound(genre, uint64(i))
			}
		})
	}
}

func BenchmarkGenerateEmptyClickSound(b *testing.B) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		b.Run(genre, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = GenerateEmptyClickSound(genre, uint64(i))
			}
		})
	}
}

func BenchmarkGeneratePickupJingleSound(b *testing.B) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		b.Run(genre, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = GeneratePickupJingleSound(genre, uint64(i))
			}
		})
	}
}
