package chat

import (
	"fmt"
	"strings"
	"sync"

	"golang.org/x/text/unicode/norm"
)

// homoglyphMap maps common Unicode homoglyphs (Cyrillic, Greek, etc.) to their
// Latin ASCII equivalents for normalization before profanity matching.
var homoglyphMap = map[rune]rune{
	// Cyrillic lookalikes
	'а': 'a', 'е': 'e', 'о': 'o', 'р': 'p', 'с': 'c', 'х': 'x',
	'А': 'a', 'В': 'b', 'Е': 'e', 'К': 'k', 'М': 'm', 'Н': 'h',
	'О': 'o', 'Р': 'p', 'С': 'c', 'Т': 't', 'Х': 'x',
	// Greek lookalikes
	'α': 'a', 'β': 'b', 'ε': 'e', 'ο': 'o', 'ρ': 'p', 'υ': 'u',
	'ν': 'v', 'χ': 'x',
	// Full-width ASCII
	'ａ': 'a', 'ｂ': 'b', 'ｃ': 'c', 'ｄ': 'd', 'ｅ': 'e',
	'ｆ': 'f', 'ｇ': 'g', 'ｈ': 'h', 'ｉ': 'i', 'ｊ': 'j',
	'ｋ': 'k', 'ｌ': 'l', 'ｍ': 'm', 'ｎ': 'n', 'ｏ': 'o',
	'ｐ': 'p', 'ｑ': 'q', 'ｒ': 'r', 'ｓ': 's', 'ｔ': 't',
	'ｕ': 'u', 'ｖ': 'v', 'ｗ': 'w', 'ｘ': 'x', 'ｙ': 'y', 'ｚ': 'z',
	// Leet-speak reverse: normalize digits/symbols to letters so that
	// "а$$" (Cyrillic а + double-dollar) is detected as "ass".
	'4': 'a', '@': 'a',
	'3': 'e',
	'1': 'i', '!': 'i',
	'0': 'o',
	'5': 's', '$': 's',
	'7': 't',
}

// normalizeForFilter applies Unicode NFC normalization, homoglyph substitution,
// and lowercasing so that evasion attempts using lookalike characters are caught.
func normalizeForFilter(s string) string {
	// NFC normalization decomposes and recomposes combining characters
	s = norm.NFC.String(s)
	s = strings.ToLower(s)

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if latin, ok := homoglyphMap[r]; ok {
			b.WriteRune(latin)
		} else {
			b.WriteRune(r) // preserve unknown chars (ß, é, ü, etc.)
		}
	}
	return b.String()
}

// ProfanityFilter filters profane language from chat messages
type ProfanityFilter struct {
	wordlists map[string][]string // language -> words
	mu        sync.RWMutex
	loaded    bool
	seed      int64 // seed for deterministic wordlist generation
}

// NewProfanityFilter creates a new profanity filter with default seed
func NewProfanityFilter() *ProfanityFilter {
	return NewProfanityFilterWithSeed(42) // Default seed for consistent behavior
}

// NewProfanityFilterWithSeed creates a new profanity filter with custom seed
func NewProfanityFilterWithSeed(seed int64) *ProfanityFilter {
	return &ProfanityFilter{
		wordlists: make(map[string][]string),
		seed:      seed,
	}
}

// LoadLanguage loads a profanity word list for the given language code
// Supported: en, es, de, fr, pt
func (pf *ProfanityFilter) LoadLanguage(lang string) error {
	pf.mu.Lock()
	defer pf.mu.Unlock()

	// Validate language code
	validLangs := map[string]bool{
		"en": true, "es": true, "de": true, "fr": true, "pt": true,
	}
	if !validLangs[lang] {
		return fmt.Errorf("unsupported language code: %s", lang)
	}

	// Check if already loaded
	if _, exists := pf.wordlists[lang]; exists {
		return nil
	}

	// Generate wordlist procedurally
	words := GenerateProfanityWordlist(lang, pf.seed)
	pf.wordlists[lang] = words
	pf.loaded = true

	return nil
}

// LoadAllLanguages loads all available word lists
func (pf *ProfanityFilter) LoadAllLanguages() error {
	languages := []string{"en", "es", "de", "fr", "pt"}
	for _, lang := range languages {
		if err := pf.LoadLanguage(lang); err != nil {
			return err
		}
	}
	return nil
}

// Filter checks if a message contains profanity
// Returns true if profanity detected, false otherwise
func (pf *ProfanityFilter) Filter(message, language string) bool {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	// Ensure language is loaded
	words, exists := pf.wordlists[language]
	if !exists {
		// Fall back to English if language not loaded
		words, exists = pf.wordlists["en"]
		if !exists {
			return false // No wordlists loaded
		}
	}

	// Normalize message: NFC + homoglyph substitution + lowercase
	normalizedMsg := normalizeForFilter(message)

	// Check each word in the profanity list
	for _, word := range words {
		if strings.Contains(normalizedMsg, word) {
			return true // Profanity detected
		}
	}

	return false // Clean
}

// Sanitize replaces profanity with asterisks
func (pf *ProfanityFilter) Sanitize(message, language string) string {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	words, exists := pf.wordlists[language]
	if !exists {
		words, exists = pf.wordlists["en"]
		if !exists {
			return message // No wordlists loaded
		}
	}

	// sanitized starts as the original to preserve casing for non-leet text.
	sanitized := message
	normalizedMsg := normalizeForFilter(message)

	// Replace each profane word with asterisks
	for _, word := range words {
		if !strings.Contains(normalizedMsg, word) {
			continue
		}
		asterisks := strings.Repeat("*", len(word))
		// Try replacing in original first (preserves casing).
		replaced := replaceCaseInsensitive(sanitized, word, asterisks)
		if replaced != sanitized {
			sanitized = replaced
		} else {
			// Word exists only in normalized form (leet/homoglyph evasion).
			sanitized = replaceCaseInsensitive(normalizedMsg, word, asterisks)
			normalizedMsg = sanitized
		}
	}

	return sanitized
}

// replaceCaseInsensitive replaces old with new in str, case-insensitively
func replaceCaseInsensitive(str, old, new string) string {
	oldLower := strings.ToLower(old)
	strLower := strings.ToLower(str)

	idx := strings.Index(strLower, oldLower)
	for idx != -1 {
		// Replace the substring at idx
		str = str[:idx] + new + str[idx+len(old):]
		strLower = strings.ToLower(str)
		idx = strings.Index(strLower, oldLower)
	}

	return str
}

// GetLoadedLanguages returns list of loaded language codes
func (pf *ProfanityFilter) GetLoadedLanguages() []string {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	langs := make([]string, 0, len(pf.wordlists))
	for lang := range pf.wordlists {
		langs = append(langs, lang)
	}
	return langs
}

// WordCount returns the number of profane words in a language's list
func (pf *ProfanityFilter) WordCount(language string) int {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	if words, exists := pf.wordlists[language]; exists {
		return len(words)
	}
	return 0
}
