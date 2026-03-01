package lore

import (
	"strings"
	"testing"
)

func TestNewMarkovChain(t *testing.T) {
	corpus := []string{"the quick brown fox", "the lazy dog"}
	mc := NewMarkovChain(12345, corpus)

	if mc == nil {
		t.Fatal("NewMarkovChain returned nil")
	}
	if mc.chain == nil {
		t.Fatal("chain not initialized")
	}
	if mc.start == nil {
		t.Fatal("start words not initialized")
	}
	if mc.rng == nil {
		t.Fatal("rng not initialized")
	}
}

func TestMarkovChain_Train(t *testing.T) {
	mc := NewMarkovChain(12345, nil)
	mc.train("hello world hello")

	// Should have "hello" as start word
	found := false
	for _, word := range mc.start {
		if word == "hello" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("start word 'hello' not found")
	}

	// Should have "hello" -> "world" transition
	nextWords, ok := mc.chain["hello"]
	if !ok {
		t.Fatal("no transitions from 'hello'")
	}
	if len(nextWords) == 0 {
		t.Fatal("no next words for 'hello'")
	}
}

func TestMarkovChain_TrainEmpty(t *testing.T) {
	mc := NewMarkovChain(12345, nil)
	mc.train("")

	// Should handle empty string gracefully
	if len(mc.start) != 0 {
		t.Fatal("empty string should not add start words")
	}
}

func TestMarkovChain_TrainSingleWord(t *testing.T) {
	mc := NewMarkovChain(12345, nil)
	mc.train("word")

	// Single word should not build transitions
	if len(mc.chain) != 0 {
		t.Fatal("single word should not create transitions")
	}
}

func TestMarkovChain_Generate(t *testing.T) {
	corpus := []string{
		"the ancient wizard cast a powerful spell",
		"the brave knight wielded a legendary sword",
		"the dark forest held many secrets",
	}
	mc := NewMarkovChain(12345, corpus)

	text := mc.Generate(10)
	if text == "" {
		t.Fatal("generated text is empty")
	}

	words := strings.Fields(text)
	if len(words) > 10 {
		t.Fatalf("generated too many words: got %d, want max 10", len(words))
	}
}

func TestMarkovChain_GenerateZeroWords(t *testing.T) {
	corpus := []string{"test corpus"}
	mc := NewMarkovChain(12345, corpus)

	text := mc.Generate(0)
	if text != "" {
		t.Fatalf("expected empty string for 0 words, got %q", text)
	}
}

func TestMarkovChain_GenerateDeterministic(t *testing.T) {
	corpus := []string{"the quick brown fox jumps over the lazy dog"}

	mc1 := NewMarkovChain(12345, corpus)
	mc2 := NewMarkovChain(12345, corpus)

	text1 := mc1.Generate(5)
	text2 := mc2.Generate(5)

	if text1 != text2 {
		t.Fatalf("generation not deterministic:\n%q\nvs\n%q", text1, text2)
	}
}

func TestMarkovChain_GenerateEmpty(t *testing.T) {
	mc := NewMarkovChain(12345, nil)
	text := mc.Generate(10)

	if text != "" {
		t.Fatalf("expected empty string from empty chain, got %q", text)
	}
}

func TestMarkovChain_GenerateSentence(t *testing.T) {
	corpus := []string{
		"the wizard cast spell",
		"the knight fought dragon",
		"the hero found treasure",
	}
	mc := NewMarkovChain(12345, corpus)

	sentence := mc.GenerateSentence()
	if sentence == "" {
		t.Fatal("generated sentence is empty")
	}

	// Should start with capital letter
	if len(sentence) > 0 && sentence[0] < 'A' || sentence[0] > 'Z' {
		t.Fatalf("sentence doesn't start with capital: %q", sentence)
	}

	// Should end with punctuation
	if !strings.HasSuffix(sentence, ".") && !strings.HasSuffix(sentence, "!") && !strings.HasSuffix(sentence, "?") {
		t.Fatalf("sentence doesn't end with punctuation: %q", sentence)
	}
}

func TestMarkovChain_GenerateSentenceEmpty(t *testing.T) {
	mc := NewMarkovChain(12345, nil)
	sentence := mc.GenerateSentence()

	if sentence != "" {
		t.Fatalf("expected empty sentence from empty chain, got %q", sentence)
	}
}

func TestMarkovChain_MultipleCorpusTexts(t *testing.T) {
	corpus := []string{
		"alpha beta gamma",
		"beta gamma delta",
		"gamma delta epsilon",
	}
	mc := NewMarkovChain(12345, corpus)

	// Should have multiple transitions for shared words
	betaNext, ok := mc.chain["beta"]
	if !ok {
		t.Fatal("'beta' should have transitions")
	}
	if len(betaNext) < 2 {
		t.Fatalf("'beta' should have at least 2 transitions, got %d", len(betaNext))
	}
}

func TestGetGenreWordBank(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			bank := GetGenreWordBank(genre)
			if bank == nil {
				t.Fatal("word bank is nil")
			}
			if len(bank.Nouns) == 0 {
				t.Fatal("nouns are empty")
			}
			if len(bank.Adjectives) == 0 {
				t.Fatal("adjectives are empty")
			}
			if len(bank.Verbs) == 0 {
				t.Fatal("verbs are empty")
			}
			if len(bank.Places) == 0 {
				t.Fatal("places are empty")
			}
			if len(bank.Subjects) == 0 {
				t.Fatal("subjects are empty")
			}
		})
	}
}

func TestGetGenreWordBank_UnknownGenre(t *testing.T) {
	bank := GetGenreWordBank("unknown")
	if bank == nil {
		t.Fatal("should fallback to fantasy bank")
	}

	// Should return fantasy bank as fallback
	fantasyBank := GetGenreWordBank("fantasy")
	if len(bank.Nouns) != len(fantasyBank.Nouns) {
		t.Fatal("fallback should return fantasy bank")
	}
}

func TestGetGenreWordBank_Uniqueness(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			bank := GetGenreWordBank(genre)

			// Check for duplicate words in each category
			checkUnique := func(category string, words []string) {
				seen := make(map[string]bool)
				for _, word := range words {
					if seen[word] {
						t.Fatalf("%s: duplicate word %q", category, word)
					}
					seen[word] = true
				}
			}

			checkUnique("nouns", bank.Nouns)
			checkUnique("adjectives", bank.Adjectives)
			checkUnique("verbs", bank.Verbs)
			checkUnique("places", bank.Places)
			checkUnique("subjects", bank.Subjects)
		})
	}
}

func TestGetGenreWordBank_MinimumSize(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			bank := GetGenreWordBank(genre)

			// Each category should have at least 10 words for variety
			if len(bank.Nouns) < 10 {
				t.Fatalf("nouns: got %d, want at least 10", len(bank.Nouns))
			}
			if len(bank.Adjectives) < 10 {
				t.Fatalf("adjectives: got %d, want at least 10", len(bank.Adjectives))
			}
			if len(bank.Verbs) < 10 {
				t.Fatalf("verbs: got %d, want at least 10", len(bank.Verbs))
			}
			if len(bank.Places) < 10 {
				t.Fatalf("places: got %d, want at least 10", len(bank.Places))
			}
			if len(bank.Subjects) < 10 {
				t.Fatalf("subjects: got %d, want at least 10", len(bank.Subjects))
			}
		})
	}
}

func TestBuildGenreCorpus(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			corpus := BuildGenreCorpus(genre, 12345)

			if len(corpus) == 0 {
				t.Fatal("corpus is empty")
			}

			// Each corpus entry should be a non-empty sentence
			for i, sentence := range corpus {
				if sentence == "" {
					t.Fatalf("sentence %d is empty", i)
				}
				if len(strings.Fields(sentence)) < 3 {
					t.Fatalf("sentence %d too short: %q", i, sentence)
				}
			}
		})
	}
}

func TestBuildGenreCorpus_Deterministic(t *testing.T) {
	corpus1 := BuildGenreCorpus("fantasy", 12345)
	corpus2 := BuildGenreCorpus("fantasy", 12345)

	if len(corpus1) != len(corpus2) {
		t.Fatal("corpus lengths differ")
	}

	for i := range corpus1 {
		if corpus1[i] != corpus2[i] {
			t.Fatalf("sentence %d differs:\n%q\nvs\n%q", i, corpus1[i], corpus2[i])
		}
	}
}

func TestBuildGenreCorpus_Variety(t *testing.T) {
	corpus := BuildGenreCorpus("fantasy", 12345)

	// Check for variety - not all sentences should be identical
	if len(corpus) < 10 {
		t.Fatal("corpus too small for variety test")
	}

	seen := make(map[string]bool)
	for _, sentence := range corpus {
		seen[sentence] = true
	}

	// Should have at least 80% unique sentences
	uniqueRatio := float64(len(seen)) / float64(len(corpus))
	if uniqueRatio < 0.8 {
		t.Fatalf("insufficient variety: %.1f%% unique", uniqueRatio*100)
	}
}

func TestNewMarkovGenerator(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			mg := NewMarkovGenerator(12345, genre)

			if mg == nil {
				t.Fatal("NewMarkovGenerator returned nil")
			}
			if mg.chain == nil {
				t.Fatal("chain not initialized")
			}
			if mg.genre != genre {
				t.Fatalf("wrong genre: got %q, want %q", mg.genre, genre)
			}
		})
	}
}

func TestMarkovGenerator_GenerateText(t *testing.T) {
	mg := NewMarkovGenerator(12345, "fantasy")

	tests := []struct {
		name      string
		sentences int
	}{
		{"single", 1},
		{"few", 3},
		{"many", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := mg.GenerateText(tt.sentences)

			if text == "" {
				t.Fatal("generated text is empty")
			}

			// Count periods as rough sentence count
			periodCount := strings.Count(text, ".")
			if periodCount < tt.sentences/2 {
				t.Fatalf("expected ~%d sentences, got ~%d periods", tt.sentences, periodCount)
			}
		})
	}
}

func TestMarkovGenerator_GenerateTextZero(t *testing.T) {
	mg := NewMarkovGenerator(12345, "fantasy")
	text := mg.GenerateText(0)

	if text != "" {
		t.Fatalf("expected empty text for 0 sentences, got %q", text)
	}
}

func TestMarkovGenerator_GenerateTextDeterministic(t *testing.T) {
	mg1 := NewMarkovGenerator(12345, "fantasy")
	mg2 := NewMarkovGenerator(12345, "fantasy")

	text1 := mg1.GenerateText(3)
	text2 := mg2.GenerateText(3)

	if text1 != text2 {
		t.Fatalf("generation not deterministic:\n%q\nvs\n%q", text1, text2)
	}
}

func TestMarkovGenerator_GenerateLoreEntry(t *testing.T) {
	mg := NewMarkovGenerator(12345, "fantasy")
	entry := mg.GenerateLoreEntry("test_id", "magic")

	if entry.ID != "test_id" {
		t.Fatalf("wrong ID: got %q, want %q", entry.ID, "test_id")
	}
	if entry.Category != "magic" {
		t.Fatalf("wrong category: got %q, want %q", entry.Category, "magic")
	}
	if entry.Title == "" {
		t.Fatal("title is empty")
	}
	if entry.Text == "" {
		t.Fatal("text is empty")
	}
	if entry.Found {
		t.Fatal("entry should not be found by default")
	}
}

func TestMarkovGenerator_GenerateLoreEntryDeterministic(t *testing.T) {
	mg1 := NewMarkovGenerator(12345, "fantasy")
	mg2 := NewMarkovGenerator(12345, "fantasy")

	entry1 := mg1.GenerateLoreEntry("test", "category")
	entry2 := mg2.GenerateLoreEntry("test", "category")

	if entry1.Text != entry2.Text {
		t.Fatalf("entries not deterministic:\n%q\nvs\n%q", entry1.Text, entry2.Text)
	}
	if entry1.Title != entry2.Title {
		t.Fatalf("titles differ: %q vs %q", entry1.Title, entry2.Title)
	}
}

func TestMarkovGenerator_AllGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			mg := NewMarkovGenerator(12345, genre)
			entry := mg.GenerateLoreEntry("test", "category")

			if entry.Text == "" {
				t.Fatalf("genre %s: empty text", genre)
			}
			if entry.Title == "" {
				t.Fatalf("genre %s: empty title", genre)
			}
		})
	}
}

func TestMarkovChain_LongGeneration(t *testing.T) {
	corpus := []string{
		"the quick brown fox jumps over the lazy dog",
		"the lazy dog sleeps under the warm sun",
		"the warm sun shines on the green grass",
	}
	mc := NewMarkovChain(12345, corpus)

	text := mc.Generate(50)
	words := strings.Fields(text)

	// Should generate text but may be shorter if chain has dead ends
	if len(words) == 0 {
		t.Fatal("generated no words")
	}
	if len(words) > 50 {
		t.Fatalf("generated too many words: %d", len(words))
	}
}

func TestMarkovChain_RepeatedWords(t *testing.T) {
	// Test that repeated words in corpus increase their probability
	corpus := []string{
		"test test test test",
		"test test",
	}
	mc := NewMarkovChain(12345, corpus)

	// 'test' should map to itself multiple times
	nextWords := mc.chain["test"]
	if len(nextWords) < 4 {
		t.Fatalf("expected at least 4 'test' transitions, got %d", len(nextWords))
	}
}

func TestGenreWordBankContent(t *testing.T) {
	// Verify specific genre words are present
	tests := []struct {
		genre    string
		category string
		word     string
	}{
		{"fantasy", "nouns", "dragon"},
		{"scifi", "nouns", "ship"},
		{"horror", "adjectives", "terrifying"},
		{"cyberpunk", "verbs", "hacked"},
		{"postapoc", "places", "wasteland"},
	}

	for _, tt := range tests {
		t.Run(tt.genre+"_"+tt.category+"_"+tt.word, func(t *testing.T) {
			bank := GetGenreWordBank(tt.genre)

			var words []string
			switch tt.category {
			case "nouns":
				words = bank.Nouns
			case "adjectives":
				words = bank.Adjectives
			case "verbs":
				words = bank.Verbs
			case "places":
				words = bank.Places
			case "subjects":
				words = bank.Subjects
			}

			found := false
			for _, word := range words {
				if word == tt.word {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf("word %q not found in %s %s", tt.word, tt.genre, tt.category)
			}
		})
	}
}

func BenchmarkMarkovChain_Generate(b *testing.B) {
	corpus := BuildGenreCorpus("fantasy", 12345)
	mc := NewMarkovChain(12345, corpus)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.Generate(20)
	}
}

func BenchmarkMarkovGenerator_GenerateText(b *testing.B) {
	mg := NewMarkovGenerator(12345, "fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mg.GenerateText(3)
	}
}

func BenchmarkBuildGenreCorpus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildGenreCorpus("fantasy", 12345)
	}
}
