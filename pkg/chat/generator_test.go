package chat

import (
	"math/rand"
	"strings"
	"testing"
)

func TestGenerateProfanityWordlist(t *testing.T) {
	tests := []struct {
		language string
		seed     int64
	}{
		{"en", 42},
		{"es", 42},
		{"de", 42},
		{"fr", 42},
		{"pt", 42},
		{"invalid", 42}, // Should fall back to English
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			words := GenerateProfanityWordlist(tt.language, tt.seed)

			if len(words) == 0 {
				t.Errorf("GenerateProfanityWordlist(%q, %d) returned empty list", tt.language, tt.seed)
			}

			// Verify all words are lowercase and non-empty
			for i, word := range words {
				if word == "" {
					t.Errorf("word[%d] is empty", i)
				}
				if word != strings.ToLower(word) {
					t.Errorf("word[%d] = %q is not lowercase", i, word)
				}
			}
		})
	}
}

func TestGenerateProfanityWordlistDeterminism(t *testing.T) {
	// Same seed should produce identical results
	languages := []string{"en", "es", "de", "fr", "pt"}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			seed := int64(12345)

			words1 := GenerateProfanityWordlist(lang, seed)
			words2 := GenerateProfanityWordlist(lang, seed)

			if len(words1) != len(words2) {
				t.Errorf("wordlist lengths differ: %d vs %d", len(words1), len(words2))
			}

			for i := range words1 {
				if words1[i] != words2[i] {
					t.Errorf("word[%d] differs: %q vs %q", i, words1[i], words2[i])
				}
			}
		})
	}
}

func TestGenerateProfanityWordlistDifferentSeeds(t *testing.T) {
	// Different seeds should produce different word orders (due to shuffle)
	lang := "en"
	words1 := GenerateProfanityWordlist(lang, 42)
	words2 := GenerateProfanityWordlist(lang, 99)

	// Words should have same content but potentially different order
	if len(words1) != len(words2) {
		t.Errorf("different seeds produced different wordlist lengths: %d vs %d", len(words1), len(words2))
	}

	// Create maps to verify same content
	map1 := make(map[string]bool)
	map2 := make(map[string]bool)

	for _, w := range words1 {
		map1[w] = true
	}
	for _, w := range words2 {
		map2[w] = true
	}

	if len(map1) != len(map2) {
		t.Errorf("different seeds produced different word sets: %d vs %d unique words", len(map1), len(map2))
	}

	for word := range map1 {
		if !map2[word] {
			t.Errorf("word %q in wordlist1 but not in wordlist2", word)
		}
	}
}

func TestGenerateEnglishWordlist(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	words := generateEnglishWordlist(rng)

	if len(words) == 0 {
		t.Fatal("generateEnglishWordlist returned empty list")
	}

	// Check for expected common words
	expectedWords := []string{"shit", "fuck", "damn"}
	for _, expected := range expectedWords {
		found := false
		for _, word := range words {
			if word == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("English wordlist missing common word: %q", expected)
		}
	}
}

func TestGenerateSpanishWordlist(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	words := generateSpanishWordlist(rng)

	if len(words) == 0 {
		t.Fatal("generateSpanishWordlist returned empty list")
	}

	// Check for expected Spanish profanity
	expectedWords := []string{"mierda", "puta"}
	for _, expected := range expectedWords {
		found := false
		for _, word := range words {
			if word == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Spanish wordlist missing common word: %q", expected)
		}
	}
}

func TestGenerateGermanWordlist(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	words := generateGermanWordlist(rng)

	if len(words) == 0 {
		t.Fatal("generateGermanWordlist returned empty list")
	}

	// Check for expected German profanity
	expectedWords := []string{"scheiße", "arsch"}
	for _, expected := range expectedWords {
		found := false
		for _, word := range words {
			if word == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("German wordlist missing common word: %q", expected)
		}
	}
}

func TestGenerateFrenchWordlist(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	words := generateFrenchWordlist(rng)

	if len(words) == 0 {
		t.Fatal("generateFrenchWordlist returned empty list")
	}

	// Check for expected French profanity
	expectedWords := []string{"merde", "putain"}
	for _, expected := range expectedWords {
		found := false
		for _, word := range words {
			if word == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("French wordlist missing common word: %q", expected)
		}
	}
}

func TestGeneratePortugueseWordlist(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	words := generatePortugueseWordlist(rng)

	if len(words) == 0 {
		t.Fatal("generatePortugueseWordlist returned empty list")
	}

	// Check for expected Portuguese profanity
	expectedWords := []string{"merda", "puta"}
	for _, expected := range expectedWords {
		found := false
		for _, word := range words {
			if word == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Portuguese wordlist missing common word: %q", expected)
		}
	}
}

func TestDeduplicateAndNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  int // expected length after deduplication
	}{
		{
			name:  "no duplicates",
			input: []string{"word1", "word2", "word3"},
			want:  3,
		},
		{
			name:  "with duplicates",
			input: []string{"word1", "word2", "word1", "word3", "word2"},
			want:  3,
		},
		{
			name:  "mixed case duplicates",
			input: []string{"Word1", "word1", "WORD1", "word2"},
			want:  2,
		},
		{
			name:  "with whitespace",
			input: []string{"  word1  ", "word2", "word1"},
			want:  2,
		},
		{
			name:  "empty strings",
			input: []string{"word1", "", "word2", "   ", "word3"},
			want:  3,
		},
		{
			name:  "all empty",
			input: []string{"", "  ", ""},
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicateAndNormalize(tt.input)

			if len(result) != tt.want {
				t.Errorf("deduplicateAndNormalize() returned %d words, want %d", len(result), tt.want)
			}

			// Verify all results are lowercase and trimmed
			for i, word := range result {
				if word != strings.ToLower(word) {
					t.Errorf("result[%d] = %q is not lowercase", i, word)
				}
				if word != strings.TrimSpace(word) {
					t.Errorf("result[%d] = %q is not trimmed", i, word)
				}
				if word == "" {
					t.Errorf("result[%d] is empty string", i)
				}
			}

			// Verify no duplicates
			seen := make(map[string]bool)
			for _, word := range result {
				if seen[word] {
					t.Errorf("duplicate word found: %q", word)
				}
				seen[word] = true
			}
		})
	}
}

func BenchmarkGenerateProfanityWordlist(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateProfanityWordlist("en", 42)
	}
}

func BenchmarkGenerateProfanityWordlistAllLanguages(b *testing.B) {
	languages := []string{"en", "es", "de", "fr", "pt"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, lang := range languages {
			GenerateProfanityWordlist(lang, 42)
		}
	}
}
