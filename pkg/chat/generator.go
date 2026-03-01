package chat

import (
	"math/rand"
	"strings"
)

// GenerateProfanityWordlist generates a deterministic profanity wordlist for a given language
// Uses linguistic patterns and phonetic rules to create offensive-sounding words
// Deterministic: same seed + language always produces same wordlist
func GenerateProfanityWordlist(language string, seed int64) []string {
	rng := rand.New(rand.NewSource(seed))

	switch language {
	case "en":
		return generateEnglishWordlist(rng)
	case "es":
		return generateSpanishWordlist(rng)
	case "de":
		return generateGermanWordlist(rng)
	case "fr":
		return generateFrenchWordlist(rng)
	case "pt":
		return generatePortugueseWordlist(rng)
	default:
		return generateEnglishWordlist(rng)
	}
}

// generateEnglishWordlist creates English profanity patterns
func generateEnglishWordlist(rng *rand.Rand) []string {
	// Base patterns for common English profanity
	bases := []string{
		"sh", "f", "d", "b", "a", "c", "p", "w", "m",
	}

	// Common endings that create offensive words
	endings := [][]string{
		{"it", "ite", "itt"},       // sh-it
		{"uck", "ucker", "ucking"}, // f-uck
		{"amn", "ammed"},           // d-amn
		{"itch", "itches"},         // b-itch
		{"ss", "sshole", "sses"},   // a-ss
		{"rap", "rappy"},           // c-rap
		{"iss", "isser"},           // p-iss
		{"hore", "horing"},         // w-hore
		{"oron", "oronic"},         // m-oron
	}

	// Additional whole words commonly filtered
	wholeWords := []string{
		"bastard", "bitch", "bollocks", "bugger",
		"bullshit", "cock", "cunt", "dick",
		"fuck", "piss", "shit",
		"slut", "twat", "wanker", "whore",
	}

	var words []string
	words = append(words, wholeWords...)

	// Generate combinations
	for i, base := range bases {
		if i < len(endings) {
			for _, end := range endings[i] {
				words = append(words, base+end)
			}
		}
	}

	// Shuffle deterministically for variety
	rng.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})

	return deduplicateAndNormalize(words)
}

// generateSpanishWordlist creates Spanish profanity patterns
func generateSpanishWordlist(rng *rand.Rand) []string {
	words := []string{
		"mierda", "puta", "puto", "coño", "joder",
		"cabron", "cabrón", "pendejo", "chingar",
		"verga", "pinche", "culero", "idiota",
		"estupido", "estúpido", "imbecil", "imbécil",
		"gilipollas", "hijo de puta", "marica",
	}

	rng.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})

	return deduplicateAndNormalize(words)
}

// generateGermanWordlist creates German profanity patterns
func generateGermanWordlist(rng *rand.Rand) []string {
	words := []string{
		"scheiße", "scheisse", "scheiss", "arsch",
		"arschloch", "fick", "ficken", "hure",
		"fotze", "wichser", "verdammt", "blöd",
		"blödsinn", "dummkopf", "idiot", "depp",
		"trottel", "schwanz", "sau", "saubande",
	}

	rng.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})

	return deduplicateAndNormalize(words)
}

// generateFrenchWordlist creates French profanity patterns
func generateFrenchWordlist(rng *rand.Rand) []string {
	words := []string{
		"merde", "putain", "con", "connard",
		"salope", "enculé", "enculer", "chier",
		"bordel", "foutre", "bite", "couille",
		"couilles", "pute", "imbécile", "crétin",
		"idiot", "merdique", "va te faire",
	}

	rng.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})

	return deduplicateAndNormalize(words)
}

// generatePortugueseWordlist creates Portuguese profanity patterns
func generatePortugueseWordlist(rng *rand.Rand) []string {
	words := []string{
		"merda", "puta", "puto", "caralho",
		"foder", "fodido", "cu", "cuzão",
		"filho da puta", "idiota", "imbecil",
		"burro", "estúpido", "estupido", "bosta",
		"cacete", "babaca", "otário",
	}

	rng.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})

	return deduplicateAndNormalize(words)
}

// deduplicateAndNormalize removes duplicates and normalizes to lowercase
func deduplicateAndNormalize(words []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(words))

	for _, word := range words {
		normalized := strings.ToLower(strings.TrimSpace(word))
		if normalized == "" {
			continue
		}
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}

	return result
}
