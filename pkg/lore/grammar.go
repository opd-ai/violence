// Package lore provides procedural text generation for lore, narrative, and backstory.
//
// Two text generation approaches are available:
//  1. Template-based generation (lore.go) - uses predefined sentence templates with word substitution
//  2. Markov chain generation (grammar.go) - uses bigram transitions trained on genre-specific word banks
//
// Example usage of Markov chain generation:
//
//	mg := lore.NewMarkovGenerator(seed, "fantasy")
//	entry := mg.GenerateLoreEntry("artifact_001", "artifacts")
//	fmt.Println(entry.Title) // "Tale of Artifacts"
//	fmt.Println(entry.Text)  // "The ancient wizard wielded powerful magic. The legendary sword was discovered..."
//
// Genre-specific word banks ensure thematically appropriate text for fantasy, scifi, horror, cyberpunk, and postapoc genres.
package lore

import (
	"math/rand"
	"strings"
)

// MarkovChain implements a simple Markov chain text generator.
// Uses bigram transitions (two consecutive words predict the next word).
type MarkovChain struct {
	chain map[string][]string
	start []string
	rng   *rand.Rand
}

// NewMarkovChain creates a Markov chain from a corpus of training text.
func NewMarkovChain(seed int64, corpus []string) *MarkovChain {
	mc := &MarkovChain{
		chain: make(map[string][]string),
		start: make([]string, 0),
		rng:   rand.New(rand.NewSource(seed)),
	}

	for _, text := range corpus {
		mc.train(text)
	}

	return mc
}

// train processes a single training text and builds the chain.
func (mc *MarkovChain) train(text string) {
	words := strings.Fields(text)
	if len(words) < 2 {
		return
	}

	// Add first word as potential start
	mc.start = append(mc.start, words[0])

	// Build bigram transitions
	for i := 0; i < len(words)-1; i++ {
		key := words[i]
		next := words[i+1]
		mc.chain[key] = append(mc.chain[key], next)
	}
}

// Generate creates a text of specified word count using the Markov chain.
func (mc *MarkovChain) Generate(wordCount int) string {
	if len(mc.start) == 0 || wordCount == 0 {
		return ""
	}

	result := make([]string, 0, wordCount)

	// Pick random starting word
	current := mc.start[mc.rng.Intn(len(mc.start))]
	result = append(result, current)

	for i := 1; i < wordCount; i++ {
		nextWords, ok := mc.chain[current]
		if !ok || len(nextWords) == 0 {
			// Dead end, restart from random start word
			if len(mc.start) > 0 {
				current = mc.start[mc.rng.Intn(len(mc.start))]
			} else {
				break
			}
		} else {
			current = nextWords[mc.rng.Intn(len(nextWords))]
		}
		result = append(result, current)
	}

	return strings.Join(result, " ")
}

// GenerateSentence creates a sentence-like text with punctuation.
func (mc *MarkovChain) GenerateSentence() string {
	// Generate 5-15 words for a sentence
	wordCount := 5 + mc.rng.Intn(11)
	text := mc.Generate(wordCount)

	if text == "" {
		return ""
	}

	// Capitalize first letter
	if len(text) > 0 {
		text = strings.ToUpper(text[:1]) + text[1:]
	}

	// Add period if not already punctuated
	if !strings.HasSuffix(text, ".") && !strings.HasSuffix(text, "!") && !strings.HasSuffix(text, "?") {
		text += "."
	}

	return text
}

// GenreWordBank contains genre-specific word banks for Markov chain training.
type GenreWordBank struct {
	Nouns      []string
	Adjectives []string
	Verbs      []string
	Places     []string
	Subjects   []string
}

// GetGenreWordBank returns word banks for a specific genre.
func GetGenreWordBank(genre string) *GenreWordBank {
	banks := map[string]*GenreWordBank{
		"fantasy": {
			Nouns:      []string{"sword", "magic", "dragon", "wizard", "spell", "quest", "kingdom", "treasure", "rune", "crystal", "armor", "shield", "staff", "tome", "prophecy", "curse", "beast", "knight", "mage", "warrior"},
			Adjectives: []string{"ancient", "mystical", "enchanted", "powerful", "legendary", "cursed", "sacred", "forgotten", "hidden", "dark", "bright", "eternal", "arcane", "divine", "demonic", "noble", "fierce", "wise", "brave", "lost"},
			Verbs:      []string{"discovered", "wielded", "cast", "summoned", "defeated", "sought", "guarded", "sealed", "destroyed", "created", "forged", "blessed", "cursed", "awakened", "vanquished", "conquered", "explored", "protected", "corrupted", "purified"},
			Places:     []string{"castle", "dungeon", "forest", "mountain", "temple", "tower", "cavern", "ruins", "citadel", "realm", "sanctuary", "vault", "chamber", "keep", "monastery", "shrine", "grove", "fortress", "crypt", "throne room"},
			Subjects:   []string{"the hero", "the sorcerer", "the guardian", "the king", "the oracle", "the champion", "the elder", "the paladin", "the ranger", "the druid", "the necromancer", "the alchemist", "the bard", "the thief", "the monk"},
		},
		"scifi": {
			Nouns:      []string{"station", "ship", "plasma", "android", "AI", "colony", "weapon", "reactor", "scanner", "shield", "warp", "laser", "data", "protocol", "specimen", "anomaly", "breach", "quarantine", "system", "network"},
			Adjectives: []string{"advanced", "alien", "experimental", "classified", "encrypted", "quantum", "neural", "synthetic", "orbital", "stellar", "cosmic", "automated", "hostile", "unknown", "critical", "emergency", "unauthorized", "corrupted", "unstable", "dimensional"},
			Verbs:      []string{"detected", "analyzed", "transmitted", "initiated", "terminated", "deployed", "extracted", "synthesized", "activated", "deactivated", "scanned", "encrypted", "decrypted", "jettisoned", "docked", "launched", "calibrated", "overloaded", "rerouted", "downloaded"},
			Places:     []string{"sector", "deck", "bay", "laboratory", "bridge", "quarters", "cargo hold", "engineering", "medbay", "airlock", "corridor", "terminal", "hangar", "reactor core", "control room", "observation deck", "escape pod", "docking port", "server room", "cryo chamber"},
			Subjects:   []string{"the captain", "the crew", "the scientist", "the android", "the pilot", "the engineer", "the medic", "the security officer", "the AI", "the commander", "the technician", "the colonist", "the agent", "the researcher", "the navigator"},
		},
		"horror": {
			Nouns:      []string{"shadow", "blood", "corpse", "ritual", "scream", "darkness", "nightmare", "entity", "fear", "madness", "flesh", "bone", "sacrifice", "whisper", "terror", "dread", "horror", "specter", "demon", "omen"},
			Adjectives: []string{"twisted", "rotting", "horrifying", "unspeakable", "eldritch", "nightmarish", "grotesque", "macabre", "sinister", "malevolent", "unholy", "cursed", "haunted", "forbidden", "terrifying", "gruesome", "ghastly", "eerie", "ominous", "dreadful"},
			Verbs:      []string{"haunted", "consumed", "mutilated", "summoned", "possessed", "corrupted", "sacrificed", "devoured", "tormented", "infected", "stalked", "whispered", "screamed", "manifested", "awakened", "emerged", "lurked", "crawled", "writhed", "decayed"},
			Places:     []string{"basement", "attic", "morgue", "cemetery", "asylum", "chapel", "chamber", "pit", "void", "abyss", "altar", "catacombs", "tomb", "cell", "ward", "laboratory", "crypt", "den", "lair", "sanctum"},
			Subjects:   []string{"the victim", "the cultist", "the investigator", "the patient", "the priest", "the thing", "the entity", "the witness", "the survivor", "the host", "the subject", "the chosen", "the damned", "the hunted", "the possessed"},
		},
		"cyberpunk": {
			Nouns:      []string{"chip", "implant", "code", "net", "corp", "hack", "data", "chrome", "augment", "virus", "ICE", "program", "subnet", "connection", "interface", "terminal", "cyberspace", "shard", "daemon", "firewall"},
			Adjectives: []string{"neon", "encrypted", "illegal", "corporate", "underground", "synthetic", "neural", "digital", "virtual", "black market", "upgraded", "glitched", "jacked", "rogue", "commercial", "proprietary", "open source", "counterfeit", "pirated", "bleeding edge"},
			Verbs:      []string{"hacked", "jacked", "downloaded", "uploaded", "encrypted", "cracked", "interfaced", "augmented", "patched", "corrupted", "traced", "spoofed", "routed", "breached", "extracted", "injected", "compiled", "executed", "terminated", "overclocked"},
			Places:     []string{"megacity", "district", "arcology", "subnet", "black market", "clinic", "nightclub", "netspace", "tower", "slum", "undercity", "hab block", "data fortress", "warehouse", "terminal", "node", "server farm", "junction", "grid sector", "combat zone"},
			Subjects:   []string{"the runner", "the netrunner", "the fixer", "the corpo", "the street samurai", "the techie", "the hacker", "the enforcer", "the dealer", "the medic", "the nomad", "the gang member", "the AI", "the suit", "the operator"},
		},
		"postapoc": {
			Nouns:      []string{"wasteland", "radiation", "bunker", "scrap", "survivor", "mutant", "ruin", "shelter", "supplies", "weapon", "water", "food", "medicine", "gas mask", "geiger counter", "fallout", "ash", "dust", "rubble", "salvage"},
			Adjectives: []string{"irradiated", "abandoned", "decayed", "rusted", "contaminated", "desolate", "scarce", "makeshift", "scavenged", "mutated", "barren", "toxic", "desperate", "hostile", "fortified", "crumbling", "radioactive", "scorched", "ravaged", "depleted"},
			Verbs:      []string{"scavenged", "survived", "mutated", "fortified", "rationed", "traded", "hunted", "defended", "looted", "barricaded", "filtered", "purified", "salvaged", "repaired", "hoarded", "fled", "migrated", "adapted", "endured", "perished"},
			Places:     []string{"vault", "bunker", "settlement", "ruins", "crater", "metro", "wasteland", "dead zone", "outpost", "camp", "shelter", "tunnels", "depot", "factory", "highway", "gas station", "mall", "hospital", "underground", "safehouse"},
			Subjects:   []string{"the survivor", "the scavenger", "the raider", "the vault dweller", "the trader", "the guard", "the wanderer", "the hunter", "the medic", "the mechanic", "the mutant", "the leader", "the exile", "the nomad", "the settler"},
		},
	}

	bank, ok := banks[genre]
	if !ok {
		return banks["fantasy"]
	}
	return bank
}

// BuildGenreCorpus creates training sentences from a genre word bank.
func BuildGenreCorpus(genre string, seed int64) []string {
	bank := GetGenreWordBank(genre)
	rng := rand.New(rand.NewSource(seed))

	corpus := make([]string, 0, 50)

	// Generate varied sentence structures
	templates := []string{
		"%s %s %s the %s %s.",
		"The %s %s %s in the %s.",
		"%s %s the %s %s with %s.",
		"In the %s, %s %s the %s.",
		"The %s %s was %s and %s.",
		"%s found %s %s in the %s.",
		"The %s held %s %s.",
		"%s %s through the %s %s.",
		"A %s %s emerged from the %s.",
		"%s discovered the %s was %s.",
	}

	for i := 0; i < 50; i++ {
		template := templates[rng.Intn(len(templates))]

		// Fill template with random words from banks
		words := []interface{}{
			bank.Subjects[rng.Intn(len(bank.Subjects))],
			bank.Verbs[rng.Intn(len(bank.Verbs))],
			bank.Adjectives[rng.Intn(len(bank.Adjectives))],
			bank.Nouns[rng.Intn(len(bank.Nouns))],
			bank.Places[rng.Intn(len(bank.Places))],
		}

		// Build sentence by replacing placeholders
		// Simple approach: use indices cyclically
		sentence := template
		for j, word := range words {
			if strings.Count(sentence, "%s") > 0 {
				sentence = strings.Replace(sentence, "%s", word.(string), 1)
			}
			if j >= 4 {
				break
			}
		}

		corpus = append(corpus, sentence)
	}

	return corpus
}

// MarkovGenerator wraps MarkovChain for lore generation.
type MarkovGenerator struct {
	chain *MarkovChain
	genre string
}

// NewMarkovGenerator creates a Markov-based lore generator.
func NewMarkovGenerator(seed int64, genre string) *MarkovGenerator {
	corpus := BuildGenreCorpus(genre, seed)
	chain := NewMarkovChain(seed, corpus)

	return &MarkovGenerator{
		chain: chain,
		genre: genre,
	}
}

// GenerateText creates procedural text with specified sentence count.
func (mg *MarkovGenerator) GenerateText(sentences int) string {
	result := make([]string, 0, sentences)

	for i := 0; i < sentences; i++ {
		sentence := mg.chain.GenerateSentence()
		if sentence != "" {
			result = append(result, sentence)
		}
	}

	return strings.Join(result, " ")
}

// GenerateLoreEntry creates a lore entry using Markov chain generation.
func (mg *MarkovGenerator) GenerateLoreEntry(id, category string) Entry {
	// Generate 2-4 sentences
	sentenceCount := 2 + (hashString(id) % 3)
	text := mg.GenerateText(int(sentenceCount))

	// Generate title from category
	title := "Tale of " + strings.Title(category)

	return Entry{
		ID:       id,
		Title:    title,
		Text:     text,
		Category: category,
		Found:    false,
	}
}
