package dialogue

import (
	"fmt"
	"math/rand"
	"strings"
)

// GrammarGenerator procedurally generates dialogue using context-free grammars.
type GrammarGenerator struct {
	rules map[string]map[string][][]string
}

// NewGrammarGenerator creates a grammar-based dialogue generator.
func NewGrammarGenerator() *GrammarGenerator {
	return &GrammarGenerator{
		rules: makeGrammarRules(),
	}
}

// Generate creates dialogue from grammar rules.
func (gg *GrammarGenerator) Generate(genre string, speakerType SpeakerType, dialogueType DialogueType, seed int64) []string {
	rng := rand.New(rand.NewSource(seed))

	// Get grammar rules for genre
	genreRules, ok := gg.rules[genre]
	if !ok {
		genreRules = gg.rules["fantasy"]
	}

	// Select grammar pattern based on speaker and dialogue type
	key := fmt.Sprintf("%d_%d", speakerType, dialogueType)
	patterns, ok := genreRules[key]
	if !ok {
		// Fallback to generic pattern
		patterns = gg.getFallbackPattern(dialogueType)
	}

	if len(patterns) == 0 {
		return []string{"..."}
	}

	// Generate 1-3 lines
	lineCount := 1 + rng.Intn(3)
	lines := make([]string, lineCount)

	for i := 0; i < lineCount; i++ {
		pattern := patterns[rng.Intn(len(patterns))]
		lines[i] = gg.expandPattern(pattern, genre, rng)
	}

	return lines
}

// expandPattern recursively expands grammar tokens.
func (gg *GrammarGenerator) expandPattern(pattern []string, genre string, rng *rand.Rand) string {
	var parts []string
	for _, token := range pattern {
		// Check if token contains {placeholders}
		if strings.Contains(token, "{") && strings.Contains(token, "}") {
			// Expand all placeholders in the token
			expanded := gg.expandPlaceholders(token, genre, rng)
			parts = append(parts, expanded)
		} else {
			// Terminal symbol without placeholders
			parts = append(parts, token)
		}
	}
	return strings.Join(parts, " ")
}

// expandPlaceholders replaces all {placeholder} tokens in a string.
func (gg *GrammarGenerator) expandPlaceholders(text, genre string, rng *rand.Rand) string {
	result := text
	vocab := gg.getVocabulary(genre)

	// Find and replace all {token} patterns
	for tokenName, options := range vocab {
		placeholder := "{" + tokenName + "}"
		if strings.Contains(result, placeholder) && len(options) > 0 {
			replacement := options[rng.Intn(len(options))]
			result = strings.ReplaceAll(result, placeholder, replacement)
		}
	}

	return result
}

// getVocabulary returns genre-specific vocabulary for token expansion.
func (gg *GrammarGenerator) getVocabulary(genre string) map[string][]string {
	vocabMap := map[string]map[string][]string{
		"fantasy": {
			"place":    {"castle", "dungeon", "forest", "ruins", "temple", "crypt", "tower"},
			"faction":  {"kingdom", "guild", "order", "cult", "clan", "tribe"},
			"adj":      {"ancient", "dark", "noble", "cursed", "sacred", "forbidden"},
			"artifact": {"sword", "staff", "amulet", "tome", "crown", "gem"},
			"goal":     {"defeat the enemy", "retrieve the artifact", "protect the village", "seal the portal"},
		},
		"scifi": {
			"place":    {"station", "colony", "ship", "facility", "sector", "outpost"},
			"faction":  {"corporation", "fleet", "alliance", "syndicate"},
			"adj":      {"anomalous", "critical", "classified", "tactical", "hostile"},
			"artifact": {"data core", "weapon", "module", "device", "sample"},
			"goal":     {"secure the area", "extract the data", "eliminate the threat", "repair the system"},
			"number":   {"7", "12", "Alpha", "Beta", "Gamma"},
		},
		"horror": {
			"place":   {"asylum", "basement", "morgue", "chapel", "attic", "cellar"},
			"faction": {"cult", "entity", "presence", "force"},
			"adj":     {"terrible", "unspeakable", "dreadful", "horrific", "sinister"},
			"goal":    {"escape this place", "survive the night", "find the truth", "stop the ritual"},
		},
		"cyberpunk": {
			"place":    {"tower", "net", "district", "server", "subnet", "node"},
			"faction":  {"corporation", "gang", "syndicate", "network"},
			"adj":      {"encrypted", "lethal", "corporate", "black-market", "military-grade"},
			"artifact": {"data", "implant", "chip", "program", "hardware"},
			"goal":     {"hack the system", "extract the data", "neutralize security", "upload the virus"},
			"number":   {"2077", "451", "XIII", "Omega"},
		},
		"postapoc": {
			"place":    {"settlement", "ruins", "wasteland", "bunker", "outpost", "scrapyard"},
			"faction":  {"raiders", "mutants", "survivors", "scavengers"},
			"adj":      {"desperate", "dangerous", "scarce", "valuable", "contaminated"},
			"artifact": {"supplies", "water", "ammunition", "fuel", "medicine"},
			"goal":     {"scavenge for supplies", "clear the area", "defend the settlement", "find survivors"},
		},
	}

	vocab, ok := vocabMap[genre]
	if !ok {
		return vocabMap["fantasy"]
	}
	return vocab
}

// makeGrammarRules creates grammar patterns for all genre/speaker/dialogue combinations.
func makeGrammarRules() map[string]map[string][][]string {
	return map[string]map[string][][]string{
		"fantasy": {
			"0_0": { // Guard, Greeting
				{"Halt,", "traveler!"},
				{"Greetings,", "stranger."},
				{"State", "your", "business."},
			},
			"0_1": { // Guard, Mission Briefing
				{"We", "need", "aid", "with", "a", "{adj}", "threat."},
				{"Venture", "to", "the", "{place}", "and", "{goal}."},
			},
			"0_2": { // Guard, Mission Complete
				{"Well", "done!", "The", "{place}", "is", "safe."},
				{"Your", "deeds", "are", "legendary."},
			},
			"1_0": { // Merchant, Greeting
				{"Welcome,", "friend.", "What", "do", "you", "seek?"},
				{"I", "have", "{adj}", "wares."},
			},
			"2_1": { // Commander, Mission Briefing
				{"Your", "objective:", "{goal}."},
				{"The", "{faction}", "must", "be", "stopped."},
			},
			"3_0": { // Civilian, Greeting
				{"Please", "help", "us!"},
				{"The", "{faction}", "attacked", "us."},
			},
			"5_6": { // Mystic, Rumor
				{"The", "{artifact}", "holds", "{adj}", "power."},
				{"I", "sense", "{adj}", "forces", "in", "the", "{place}."},
			},
			"6_4": { // Hostile, Warning
				{"Surrender", "or", "die!"},
				{"Leave", "the", "{place}", "now!"},
			},
		},
		"scifi": {
			"0_0": { // Guard, Greeting
				{"Halt!", "Identify", "yourself."},
				{"Sector", "{number}.", "State", "your", "purpose."},
			},
			"0_1": { // Guard, Mission Briefing
				{"Mission:", "{goal}", "in", "sector", "{number}."},
				{"{adj}", "anomaly", "detected", "at", "{place}."},
			},
			"0_2": { // Guard, Mission Complete
				{"Mission", "complete.", "Return", "to", "base."},
				{"Excellent", "work,", "operative."},
			},
			"1_0": { // Merchant, Greeting
				{"Looking", "to", "trade?"},
				{"I", "have", "{adj}", "tech", "for", "sale."},
			},
			"2_1": { // Commander, Mission Briefing
				{"Infiltrate", "{place}", "and", "{goal}."},
				{"Eliminate", "all", "{adj}", "contacts."},
			},
		},
		"horror": {
			"0_0": { // Guard/Survivor, Greeting
				{"Thank", "God!", "Another", "person!"},
				{"Don't", "go", "in", "there..."},
			},
			"0_1": { // Guard/Survivor, Mission Briefing
				{"We", "must", "{goal}."},
				{"Something", "{adj}", "is", "hunting", "us."},
			},
			"3_0": { // Civilian, Greeting
				{"They", "took", "everything..."},
				{"The", "{place}", "is", "not", "safe."},
			},
			"5_6": { // Mystic, Rumor
				{"Dark", "forces", "gather", "at", "the", "{place}."},
				{"The", "ritual", "must", "be", "stopped."},
			},
			"6_4": { // Hostile, Warning
				{"Join", "us..."},
				{"Your", "flesh", "will", "serve", "us."},
			},
		},
		"cyberpunk": {
			"0_0": { // Guard, Greeting
				{"Credentials", "check", "out.", "Move", "along."},
				{"Corp", "territory.", "Got", "clearance?"},
			},
			"0_1": { // Guard, Mission Briefing
				{"Hack", "into", "{place}", "and", "{goal}."},
				{"Extract", "{adj}", "data", "from", "sector", "{number}."},
			},
			"1_0": { // Merchant, Greeting
				{"Selling", "{adj}", "chrome.", "Interested?"},
				{"Got", "hardware,", "got", "programs."},
			},
			"2_1": { // Commander, Mission Briefing
				{"Jack", "into", "{place}", "and", "{goal}."},
				{"Stealth", "recommended.", "ICE", "is", "{adj}."},
			},
		},
		"postapoc": {
			"0_0": { // Guard, Greeting
				{"Stop!", "What's", "your", "business?"},
				{"Another", "survivor.", "Rare", "sight."},
			},
			"0_1": { // Guard, Mission Briefing
				{"Scout", "the", "{place}.", "Report", "back."},
				{"Raiders", "at", "the", "{place}.", "Clear", "them", "out."},
			},
			"1_0": { // Merchant, Greeting
				{"Trading", "{adj}", "supplies."},
				{"Scrap", "for", "food."},
			},
			"2_1": { // Commander, Mission Briefing
				{"Clear", "the", "{place}", "of", "{adj}", "threats."},
				{"Find", "{artifact}", "and", "return."},
			},
		},
	}
}

// getFallbackPattern returns generic patterns for unknown combinations.
func (gg *GrammarGenerator) getFallbackPattern(dialogueType DialogueType) [][]string {
	fallbacks := map[DialogueType][][]string{
		DialogueGreeting:        {{"Hello."}, {"Greetings."}},
		DialogueMissionBriefing: {{"Listen", "carefully."}, {"Your", "mission:"}},
		DialogueMissionComplete: {{"Well", "done."}, {"Mission", "complete."}},
		DialogueIdle:            {{"..."}, {"Nothing", "to", "say."}},
		DialogueWarning:         {{"Be", "careful!"}, {"Danger", "ahead!"}},
		DialogueTrade:           {{"Want", "to", "trade?"}, {"I", "have", "goods."}},
		DialogueRumor:           {{"I", "heard", "something."}, {"Rumor", "has", "it..."}},
	}

	patterns, ok := fallbacks[dialogueType]
	if !ok {
		return [][]string{{"..."}}
	}
	return patterns
}
