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
		"fuck", "piss", "shit", "ass",
		"slut", "twat", "wanker", "whore",
		"damn", "crap", "tits",
		"fag", "retard", "nigger", "nigga",
		"kike", "spic", "chink", "gook",
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

	// Generate l33t speak variants for all words
	var withVariants []string
	withVariants = append(withVariants, words...)
	for _, word := range words {
		variants := generateLeetSpeakVariants(word)
		withVariants = append(withVariants, variants...)
	}

	// Shuffle deterministically for variety
	rng.Shuffle(len(withVariants), func(i, j int) {
		withVariants[i], withVariants[j] = withVariants[j], withVariants[i]
	})

	return deduplicateAndNormalize(withVariants)
}

// generateSpanishWordlist creates Spanish profanity patterns
func generateSpanishWordlist(rng *rand.Rand) []string {
	words := []string{
		"mierda", "puta", "puto", "coño", "joder",
		"cabron", "cabrón", "pendejo", "chingar",
		"verga", "pinche", "culero", "idiota",
		"estupido", "estúpido", "imbecil", "imbécil",
		"gilipollas", "hijo de puta", "marica",
		"polla", "tonto", "mamon", "mamón",
		"maricon", "maricón", "boludo", "pelotudo",
		"huevon", "huevón", "concha", "maraca",
	}

	// Generate l33t speak variants
	var withVariants []string
	withVariants = append(withVariants, words...)
	for _, word := range words {
		variants := generateLeetSpeakVariants(word)
		withVariants = append(withVariants, variants...)
	}

	rng.Shuffle(len(withVariants), func(i, j int) {
		withVariants[i], withVariants[j] = withVariants[j], withVariants[i]
	})

	return deduplicateAndNormalize(withVariants)
}

// generateGermanWordlist creates German profanity patterns
func generateGermanWordlist(rng *rand.Rand) []string {
	words := []string{
		"scheiße", "scheisse", "scheiss", "arsch",
		"arschloch", "fick", "ficken", "hure",
		"fotze", "wichser", "verdammt", "blöd",
		"blödsinn", "dummkopf", "idiot", "depp",
		"trottel", "schwanz", "sau", "saubande",
		"schlampe", "hurensohn", "schwuchtel", "miststück",
		"pimmel", "arschgeige", "vollpfosten", "vollidiot",
	}

	// Generate l33t speak variants
	var withVariants []string
	withVariants = append(withVariants, words...)
	for _, word := range words {
		variants := generateLeetSpeakVariants(word)
		withVariants = append(withVariants, variants...)
	}

	rng.Shuffle(len(withVariants), func(i, j int) {
		withVariants[i], withVariants[j] = withVariants[j], withVariants[i]
	})

	return deduplicateAndNormalize(withVariants)
}

// generateFrenchWordlist creates French profanity patterns
func generateFrenchWordlist(rng *rand.Rand) []string {
	words := []string{
		"merde", "putain", "con", "connard",
		"salope", "enculé", "enculer", "chier",
		"bordel", "foutre", "bite", "couille",
		"couilles", "pute", "imbécile", "crétin",
		"idiot", "merdique", "va te faire",
		"connasse", "salaud", "bâtard", "batard",
		"pouffiasse", "branleur", "enfoiré", "enfoire",
	}

	// Generate l33t speak variants
	var withVariants []string
	withVariants = append(withVariants, words...)
	for _, word := range words {
		variants := generateLeetSpeakVariants(word)
		withVariants = append(withVariants, variants...)
	}

	rng.Shuffle(len(withVariants), func(i, j int) {
		withVariants[i], withVariants[j] = withVariants[j], withVariants[i]
	})

	return deduplicateAndNormalize(withVariants)
}

// generatePortugueseWordlist creates Portuguese profanity patterns
func generatePortugueseWordlist(rng *rand.Rand) []string {
	words := []string{
		"merda", "puta", "puto", "caralho",
		"foder", "fodido", "cu", "cuzão",
		"filho da puta", "idiota", "imbecil",
		"burro", "estúpido", "estupido", "bosta",
		"cacete", "babaca", "otário", "otario",
		"viado", "bicha", "arrombado", "fdp",
		"buceta", "porra", "merda", "corno",
	}

	// Generate l33t speak variants
	var withVariants []string
	withVariants = append(withVariants, words...)
	for _, word := range words {
		variants := generateLeetSpeakVariants(word)
		withVariants = append(withVariants, variants...)
	}

	rng.Shuffle(len(withVariants), func(i, j int) {
		withVariants[i], withVariants[j] = withVariants[j], withVariants[i]
	})

	return deduplicateAndNormalize(withVariants)
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

// generateLeetSpeakVariants generates l33t speak variations of a word
// Substitutions: a→4, e→3, i→1, o→0, s→5, t→7
func generateLeetSpeakVariants(word string) []string {
	variants := []string{word}

	// Map of character substitutions
	substitutions := map[rune][]rune{
		'a': {'4', '@'},
		'e': {'3'},
		'i': {'1', '!'},
		'o': {'0'},
		's': {'5', '$'},
		't': {'7'},
	}

	// Generate all single-character substitution variants
	for i, char := range word {
		if subs, exists := substitutions[char]; exists {
			for _, sub := range subs {
				variant := word[:i] + string(sub) + word[i+1:]
				variants = append(variants, variant)
			}
		}
	}

	// Generate common multi-character variants
	multiSubs := []map[rune]rune{
		{'a': '4', 'e': '3'},
		{'a': '4', 'o': '0'},
		{'i': '1', 'o': '0'},
		{'e': '3', 's': '5'},
		{'a': '4', 'e': '3', 'i': '1', 'o': '0'},
	}

	for _, subMap := range multiSubs {
		variant := []rune(word)
		for i, char := range variant {
			if replacement, exists := subMap[char]; exists {
				variant[i] = replacement
			}
		}
		variants = append(variants, string(variant))
	}

	return variants
}
