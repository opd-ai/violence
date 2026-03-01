package chat

import (
	"strings"
	"testing"
)

func TestNewProfanityFilter(t *testing.T) {
	filter := NewProfanityFilter()
	if filter == nil {
		t.Fatal("NewProfanityFilter returned nil")
	}
	if filter.wordlists == nil {
		t.Error("wordlists map is nil")
	}
}

func TestLoadLanguage(t *testing.T) {
	filter := NewProfanityFilter()

	tests := []struct {
		lang    string
		wantErr bool
	}{
		{"en", false},
		{"es", false},
		{"de", false},
		{"fr", false},
		{"pt", false},
		{"invalid", true}, // Now returns error instead of fallback
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			err := filter.LoadLanguage(tt.lang)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadLanguage(%q) error = %v, wantErr %v", tt.lang, err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify language was loaded
				langs := filter.GetLoadedLanguages()
				found := false
				for _, lang := range langs {
					if lang == tt.lang {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("language %q not in loaded languages: %v", tt.lang, langs)
				}

				// Verify word count is positive
				count := filter.WordCount(tt.lang)
				if count == 0 {
					t.Errorf("word count for %q is 0", tt.lang)
				}
			}
		})
	}
}

func TestLoadAllLanguages(t *testing.T) {
	filter := NewProfanityFilter()

	err := filter.LoadAllLanguages()
	if err != nil {
		t.Fatalf("LoadAllLanguages failed: %v", err)
	}

	expectedLangs := []string{"en", "es", "de", "fr", "pt"}
	loadedLangs := filter.GetLoadedLanguages()

	for _, expected := range expectedLangs {
		found := false
		for _, loaded := range loadedLangs {
			if loaded == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected language %q not loaded; loaded: %v", expected, loadedLangs)
		}
	}
}

func TestFilter(t *testing.T) {
	filter := NewProfanityFilter()
	filter.LoadLanguage("en")

	tests := []struct {
		name     string
		message  string
		language string
		want     bool
	}{
		{"clean message", "Hello, how are you?", "en", false},
		{"contains profanity", "This is shit", "en", true},
		{"profanity uppercase", "This is SHIT", "en", true},
		{"profanity mixed case", "This is ShIt", "en", true},
		{"clean gaming chat", "Good game, well played!", "en", false},
		{"empty message", "", "en", false},
		{"only spaces", "   ", "en", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.Filter(tt.message, tt.language)
			if got != tt.want {
				t.Errorf("Filter(%q, %q) = %v, want %v", tt.message, tt.language, got, tt.want)
			}
		})
	}
}

func TestFilterSpanish(t *testing.T) {
	filter := NewProfanityFilter()
	filter.LoadLanguage("es")

	tests := []struct {
		message string
		want    bool
	}{
		{"Hola amigo", false},
		{"Eres un idiota", true},
		{"mierda", true},
		{"Buenas noches", false},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			got := filter.Filter(tt.message, "es")
			if got != tt.want {
				t.Errorf("Filter(%q, 'es') = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestFilterGerman(t *testing.T) {
	filter := NewProfanityFilter()
	filter.LoadLanguage("de")

	tests := []struct {
		message string
		want    bool
	}{
		{"Guten Tag", false},
		{"Du bist ein arsch", true},
		{"scheisse", true},
		{"Wie geht es dir?", false},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			got := filter.Filter(tt.message, "de")
			if got != tt.want {
				t.Errorf("Filter(%q, 'de') = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestSanitize(t *testing.T) {
	filter := NewProfanityFilter()
	filter.LoadLanguage("en")

	tests := []struct {
		name     string
		message  string
		language string
		want     string
	}{
		{"clean message", "Hello there", "en", "Hello there"},
		{"single profanity", "This is shit", "en", "This is ****"},
		{"multiple profanity", "shit and fuck", "en", "**** and ****"},
		{"profanity mixed case", "This is ShIt", "en", "This is ****"},
		{"empty", "", "en", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.Sanitize(tt.message, tt.language)
			if got != tt.want {
				t.Errorf("Sanitize(%q, %q) = %q, want %q", tt.message, tt.language, got, tt.want)
			}
		})
	}
}

func TestSanitizePreservesLength(t *testing.T) {
	filter := NewProfanityFilter()
	filter.LoadLanguage("en")

	message := "This is shit"
	sanitized := filter.Sanitize(message, "en")

	if len(sanitized) != len(message) {
		t.Errorf("Sanitize changed message length: %d -> %d", len(message), len(sanitized))
	}
}

func TestFilterFallbackToEnglish(t *testing.T) {
	filter := NewProfanityFilter()
	filter.LoadLanguage("en")

	// Try filtering with unloaded language (should fall back to English)
	message := "This is shit"
	got := filter.Filter(message, "zh") // Chinese not loaded
	if !got {
		t.Error("Filter should fall back to English and detect profanity")
	}
}

func TestFilterNoWordlistsLoaded(t *testing.T) {
	filter := NewProfanityFilter()
	// Don't load any wordlists

	got := filter.Filter("shit", "en")
	if got {
		t.Error("Filter should return false when no wordlists loaded")
	}
}

func TestReplaceCaseInsensitive(t *testing.T) {
	tests := []struct {
		str  string
		old  string
		new  string
		want string
	}{
		{"Hello World", "hello", "Hi", "Hi World"},
		{"Test TEST test", "test", "***", "*** *** ***"},
		{"No match here", "xyz", "ABC", "No match here"},
		{"", "any", "thing", ""},
		{"UPPERCASE", "upper", "lower", "lowerCASE"},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got := replaceCaseInsensitive(tt.str, tt.old, tt.new)
			if got != tt.want {
				t.Errorf("replaceCaseInsensitive(%q, %q, %q) = %q, want %q",
					tt.str, tt.old, tt.new, got, tt.want)
			}
		})
	}
}

func TestGetLoadedLanguages(t *testing.T) {
	filter := NewProfanityFilter()

	// Initially empty
	langs := filter.GetLoadedLanguages()
	if len(langs) != 0 {
		t.Errorf("initially GetLoadedLanguages() = %v, want []", langs)
	}

	// Load some languages
	filter.LoadLanguage("en")
	filter.LoadLanguage("es")

	langs = filter.GetLoadedLanguages()
	if len(langs) != 2 {
		t.Errorf("GetLoadedLanguages() returned %d languages, want 2", len(langs))
	}
}

func TestWordCount(t *testing.T) {
	filter := NewProfanityFilter()

	// Unloaded language
	count := filter.WordCount("en")
	if count != 0 {
		t.Errorf("WordCount(unloaded) = %d, want 0", count)
	}

	// Load language
	filter.LoadLanguage("en")
	count = filter.WordCount("en")
	if count == 0 {
		t.Error("WordCount(en) = 0, want > 0")
	}
}

func TestConcurrentAccess(t *testing.T) {
	filter := NewProfanityFilter()
	filter.LoadLanguage("en")

	// Run multiple goroutines accessing the filter
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				filter.Filter("test message shit", "en")
				filter.Sanitize("test message", "en")
				filter.GetLoadedLanguages()
				filter.WordCount("en")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestLoadLanguageTwice(t *testing.T) {
	filter := NewProfanityFilter()

	// Load language once
	err := filter.LoadLanguage("en")
	if err != nil {
		t.Fatalf("first LoadLanguage failed: %v", err)
	}

	count1 := filter.WordCount("en")

	// Load again (should be no-op)
	err = filter.LoadLanguage("en")
	if err != nil {
		t.Fatalf("second LoadLanguage failed: %v", err)
	}

	count2 := filter.WordCount("en")

	if count1 != count2 {
		t.Errorf("word count changed after reload: %d -> %d", count1, count2)
	}
}

func BenchmarkFilter(b *testing.B) {
	filter := NewProfanityFilter()
	filter.LoadLanguage("en")

	message := "This is a test message with no profanity"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.Filter(message, "en")
	}
}

func BenchmarkFilterWithProfanity(b *testing.B) {
	filter := NewProfanityFilter()
	filter.LoadLanguage("en")

	message := "This message contains shit and other bad words"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.Filter(message, "en")
	}
}

func BenchmarkSanitize(b *testing.B) {
	filter := NewProfanityFilter()
	filter.LoadLanguage("en")

	message := "This is shit and damn it all"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.Sanitize(message, "en")
	}
}

func TestWordListContainsProfanity(t *testing.T) {
	// Verify each word list contains expected words
	tests := []struct {
		lang     string
		mustHave string
	}{
		{"en", "shit"},
		{"es", "mierda"},
		{"de", "scheiße"},
		{"fr", "merde"},
		{"pt", "merda"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			filter := NewProfanityFilter()
			filter.LoadLanguage(tt.lang)

			// Check that the must-have word is detected
			if !filter.Filter(tt.mustHave, tt.lang) {
				t.Errorf("word list %q missing expected word %q", tt.lang, tt.mustHave)
			}
		})
	}
}

func TestFilterLongMessage(t *testing.T) {
	filter := NewProfanityFilter()
	filter.LoadLanguage("en")

	// Test with long message
	longMessage := strings.Repeat("clean message ", 1000) + "shit" + strings.Repeat(" more clean", 1000)

	if !filter.Filter(longMessage, "en") {
		t.Error("Filter failed to detect profanity in long message")
	}
}
