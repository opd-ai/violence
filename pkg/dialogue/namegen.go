package dialogue

import (
	"fmt"
	"math/rand"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// NameGenerator procedurally generates NPC names from phonetic patterns.
type NameGenerator struct {
	syllables map[string][][]string
	titles    map[string][]string
}

// NewNameGenerator creates a name generator with genre-specific phonetic patterns.
func NewNameGenerator() *NameGenerator {
	return &NameGenerator{
		syllables: makeGenreSyllables(),
		titles:    makeGenreTitles(),
	}
}

// Generate creates a procedurally generated name for a speaker type.
func (ng *NameGenerator) Generate(genre string, speakerType SpeakerType, seed int64) string {
	rng := rand.New(rand.NewSource(seed))

	// Get phonetic syllables for genre
	syllableSet := ng.getSyllables(genre)
	if len(syllableSet) == 0 {
		syllableSet = ng.getSyllables("fantasy")
	}

	// Generate base name from syllables (2-3 syllables)
	syllableCount := 2 + rng.Intn(2)
	nameParts := make([]string, syllableCount)
	for i := 0; i < syllableCount; i++ {
		nameParts[i] = syllableSet[rng.Intn(len(syllableSet))]
	}
	baseName := strings.Join(nameParts, "")

	// Capitalize first letter
	caser := cases.Title(language.English)
	baseName = caser.String(strings.ToLower(baseName))

	// Add title/rank based on speaker type
	title := ng.getTitle(genre, speakerType, rng)
	if title != "" {
		return fmt.Sprintf("%s %s", title, baseName)
	}

	return baseName
}

// getSyllables returns phonetic syllables for a genre.
func (ng *NameGenerator) getSyllables(genre string) []string {
	syllables, ok := ng.syllables[genre]
	if !ok {
		return ng.syllables["fantasy"]
	}

	// Flatten syllable patterns
	var flat []string
	for _, pattern := range syllables {
		flat = append(flat, pattern...)
	}
	return flat
}

// getTitle returns a title/rank for a speaker type.
func (ng *NameGenerator) getTitle(genre string, speakerType SpeakerType, rng *rand.Rand) string {
	titles, ok := ng.titles[genre]
	if !ok {
		titles = ng.titles["fantasy"]
	}

	// Map speaker type to title category
	var titleCandidates []string
	switch speakerType {
	case SpeakerGuard:
		titleCandidates = []string{titles[0], titles[1], titles[2]}
	case SpeakerMerchant:
		titleCandidates = []string{titles[3], titles[4]}
	case SpeakerCommander:
		titleCandidates = []string{titles[5], titles[6], titles[7]}
	case SpeakerTechnician:
		titleCandidates = []string{titles[8], titles[9]}
	case SpeakerMystic:
		titleCandidates = []string{titles[10], titles[11]}
	default:
		return "" // No title for civilians and others
	}

	if len(titleCandidates) == 0 {
		return ""
	}
	return titleCandidates[rng.Intn(len(titleCandidates))]
}

// makeGenreSyllables creates phonetic syllable patterns for each genre.
func makeGenreSyllables() map[string][][]string {
	return map[string][][]string{
		"fantasy": {
			{"al", "ar", "el", "er", "or", "ul", "ur"},
			{"bor", "dor", "gar", "mar", "ral", "tar", "var"},
			{"dar", "gor", "kar", "lor", "nor", "sor", "tor"},
			{"an", "en", "in", "on", "un"},
			{"dric", "dwin", "ren", "ric", "ron", "wyn"},
			{"eth", "ith", "oth", "uth"},
			{"len", "lyn", "mir", "ril", "thel", "wen"},
		},
		"scifi": {
			{"ax", "ex", "ix", "ox", "ux"},
			{"cor", "dex", "kex", "lex", "nex", "rex", "vex"},
			{"ai", "ei", "oi", "ui"},
			{"tron", "dron", "kron", "nar", "zar"},
			{"ko", "ro", "zo", "vo"},
			{"tan", "kan", "zan", "van"},
		},
		"horror": {
			{"ash", "black", "grim", "grey", "pale"},
			{"ford", "wood", "stone", "cross", "grave"},
			{"mor", "ven", "bal", "cal", "mal"},
			{"ris", "vis", "dis", "lis"},
		},
		"cyberpunk": {
			{"byte", "chip", "code", "data", "net"},
			{"chrome", "nyx", "glitch", "spark", "blade"},
			{"zero", "cipher", "ghost", "link"},
		},
		"postapoc": {
			{"rust", "dust", "ash", "red", "grey"},
			{"stone", "wolf", "crow", "fox", "hawk"},
			{"max", "cruz", "reyes", "doc", "sal"},
		},
	}
}

// makeGenreTitles creates rank/title patterns for each genre.
func makeGenreTitles() map[string][]string {
	return map[string][]string{
		"fantasy": {
			"Guard", "Captain", "Watchman", // Guard titles
			"Merchant", "Trader", // Merchant titles
			"Commander", "General", "Lord", // Commander titles
			"Alchemist", "Blacksmith", // Technician titles
			"Oracle", "Sage", // Mystic titles
		},
		"scifi": {
			"Officer", "Sentinel", "Security", // Guard titles
			"Trader", "Vendor", // Merchant titles
			"Commander", "Admiral", "Captain", // Commander titles
			"Engineer", "Technician", // Technician titles
			"Analyst", "Specialist", // Mystic titles
		},
		"horror": {
			"Guard", "Officer", "Watchman", // Guard titles
			"Vendor", "Trader", // Merchant titles
			"Director", "Chief", "Warden", // Commander titles
			"Doctor", "Analyst", // Technician titles
			"Medium", "Priest", // Mystic titles
		},
		"cyberpunk": {
			"Enforcer", "Officer", "Guard", // Guard titles
			"Fixer", "Dealer", // Merchant titles
			"Boss", "Director", "Exec", // Commander titles
			"Netrunner", "Tech", // Technician titles
			"Prophet", "Oracle", // Mystic titles
		},
		"postapoc": {
			"Guard", "Sentry", "Watchman", // Guard titles
			"Trader", "Merchant", // Merchant titles
			"Chief", "Boss", "Leader", // Commander titles
			"Mechanic", "Engineer", // Technician titles
			"Prophet", "Seer", // Mystic titles
		},
	}
}
