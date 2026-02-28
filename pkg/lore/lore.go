// Package lore manages the in-game lore codex with procedural generation.
package lore

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
)

// Entry represents a single lore codex entry.
type Entry struct {
	ID       string
	Title    string
	Text     string
	Category string
	Found    bool
}

// Codex holds discovered lore entries.
type Codex struct {
	Entries []Entry
	mu      sync.RWMutex
}

// Generator procedurally generates lore content from a seed.
type Generator struct {
	genre string
	rng   *rand.Rand
}

// NewCodex creates an empty codex.
func NewCodex() *Codex {
	return &Codex{
		Entries: make([]Entry, 0),
	}
}

// NewGenerator creates a lore generator with a seed.
func NewGenerator(seed int64) *Generator {
	return &Generator{
		genre: "fantasy",
		rng:   rand.New(rand.NewSource(seed)),
	}
}

// AddEntry adds a lore entry to the codex.
func (c *Codex) AddEntry(e Entry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if entry already exists
	for i, existing := range c.Entries {
		if existing.ID == e.ID {
			c.Entries[i] = e
			return
		}
	}
	c.Entries = append(c.Entries, e)
}

// GetEntry retrieves an entry by ID.
func (c *Codex) GetEntry(id string) (*Entry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for i := range c.Entries {
		if c.Entries[i].ID == id {
			return &c.Entries[i], true
		}
	}
	return nil, false
}

// GetEntries returns all lore entries.
func (c *Codex) GetEntries() []Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]Entry, len(c.Entries))
	copy(result, c.Entries)
	return result
}

// GetFoundEntries returns only discovered entries.
func (c *Codex) GetFoundEntries() []Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	found := make([]Entry, 0)
	for _, e := range c.Entries {
		if e.Found {
			found = append(found, e)
		}
	}
	return found
}

// MarkFound marks an entry as discovered.
func (c *Codex) MarkFound(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.Entries {
		if c.Entries[i].ID == id {
			c.Entries[i].Found = true
			return true
		}
	}
	return false
}

// SetGenreForGenerator sets the genre for lore generation.
func (g *Generator) SetGenre(genreID string) {
	g.genre = genreID
}

// Generate creates a procedural lore entry from the given ID.
// The ID determines the category and content via deterministic hashing.
func (g *Generator) Generate(id string) Entry {
	// Hash the ID to get deterministic random values
	hash := hashString(id)
	localRng := rand.New(rand.NewSource(hash))

	// Determine category based on ID prefix or hash
	category := g.getCategory(id, localRng)

	// Generate title
	title := g.generateTitle(category, localRng)

	// Generate text
	text := g.generateText(category, localRng)

	return Entry{
		ID:       id,
		Title:    title,
		Text:     text,
		Category: category,
		Found:    false,
	}
}

func (g *Generator) getCategory(id string, rng *rand.Rand) string {
	categories := map[string][]string{
		"fantasy":   {"history", "magic", "creatures", "artifacts", "prophecy"},
		"scifi":     {"technology", "factions", "ships", "planets", "research"},
		"horror":    {"events", "entities", "rituals", "victims", "curses"},
		"cyberpunk": {"corporations", "netspace", "augments", "gangs", "ai"},
		"postapoc":  {"collapse", "survival", "mutants", "ruins", "factions"},
	}

	genreCategories, ok := categories[g.genre]
	if !ok {
		genreCategories = categories["fantasy"]
	}

	return genreCategories[rng.Intn(len(genreCategories))]
}

func (g *Generator) generateTitle(category string, rng *rand.Rand) string {
	prefixes := []string{"The", "Ancient", "Lost", "Hidden", "Forgotten", "Sacred"}
	nouns := []string{"Chronicle", "Record", "Testament", "Account", "Report", "Log"}

	prefix := prefixes[rng.Intn(len(prefixes))]
	noun := nouns[rng.Intn(len(nouns))]

	return fmt.Sprintf("%s %s of %s", prefix, noun, strings.Title(category))
}

func (g *Generator) generateText(category string, rng *rand.Rand) string {
	// Generate 2-4 sentences
	sentenceCount := 2 + rng.Intn(3)
	sentences := make([]string, sentenceCount)

	templates := g.getTemplates()
	for i := 0; i < sentenceCount; i++ {
		template := templates[rng.Intn(len(templates))]
		sentences[i] = g.fillTemplate(template, rng)
	}

	return strings.Join(sentences, " ")
}

func (g *Generator) getTemplates() []string {
	genreTemplates := map[string][]string{
		"fantasy": {
			"In the age of {adj} {noun}, the {faction} sought {goal}.",
			"Legend speaks of {adj} {artifact} hidden in {place}.",
			"The {faction} wielded {power} against their enemies.",
		},
		"scifi": {
			"Sector {number} reported {adj} anomalies near {place}.",
			"The {faction} developed {technology} to combat {threat}.",
			"Research log {number}: {adj} results in {experiment}.",
		},
		"horror": {
			"The {adj} entity was sighted in {place} at {time}.",
			"Witnesses describe {adj} sounds emanating from {place}.",
			"The ritual required {adj} {artifact} under {condition}.",
		},
		"cyberpunk": {
			"Corp memo {number}: {adj} breach detected in {system}.",
			"The {faction} controls {resource} distribution in {place}.",
			"Neural scan reveals {adj} patterns in {subject}.",
		},
		"postapoc": {
			"Day {number}: {adj} supplies found in {place}.",
			"The {faction} controls access to {resource}.",
			"Radiation levels {adj} near {place}.",
		},
	}

	templates, ok := genreTemplates[g.genre]
	if !ok {
		templates = genreTemplates["fantasy"]
	}
	return templates
}

func (g *Generator) fillTemplate(template string, rng *rand.Rand) string {
	adjectives := []string{"strange", "ancient", "powerful", "mysterious", "dangerous", "forgotten"}
	nouns := []string{"power", "knowledge", "treasure", "secret", "weapon", "force"}
	factions := []string{"the Order", "the Collective", "the Council", "the Guild", "the Alliance"}
	places := []string{"the depths", "the ruins", "the wasteland", "the vault", "the citadel"}

	result := template
	result = strings.ReplaceAll(result, "{adj}", adjectives[rng.Intn(len(adjectives))])
	result = strings.ReplaceAll(result, "{noun}", nouns[rng.Intn(len(nouns))])
	result = strings.ReplaceAll(result, "{faction}", factions[rng.Intn(len(factions))])
	result = strings.ReplaceAll(result, "{place}", places[rng.Intn(len(places))])
	result = strings.ReplaceAll(result, "{number}", fmt.Sprintf("%d", 100+rng.Intn(900)))
	result = strings.ReplaceAll(result, "{goal}", "ultimate "+nouns[rng.Intn(len(nouns))])
	result = strings.ReplaceAll(result, "{artifact}", adjectives[rng.Intn(len(adjectives))]+" artifact")
	result = strings.ReplaceAll(result, "{power}", adjectives[rng.Intn(len(adjectives))]+" power")
	result = strings.ReplaceAll(result, "{threat}", adjectives[rng.Intn(len(adjectives))]+" threat")
	result = strings.ReplaceAll(result, "{technology}", adjectives[rng.Intn(len(adjectives))]+" technology")
	result = strings.ReplaceAll(result, "{experiment}", adjectives[rng.Intn(len(adjectives))]+" experiment")
	result = strings.ReplaceAll(result, "{time}", "midnight")
	result = strings.ReplaceAll(result, "{condition}", "full moon")
	result = strings.ReplaceAll(result, "{system}", "network")
	result = strings.ReplaceAll(result, "{resource}", "water")
	result = strings.ReplaceAll(result, "{subject}", "subject")

	return result
}

func hashString(s string) int64 {
	var hash int64
	for i := 0; i < len(s); i++ {
		hash = hash*31 + int64(s[i])
	}
	return hash
}

// SetGenre configures lore content for a genre.
func SetGenre(genreID string) {}
