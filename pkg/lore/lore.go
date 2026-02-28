// Package lore manages the in-game lore codex with procedural generation.
package lore

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Entry represents a single lore codex entry.
type Entry struct {
	ID       string
	Title    string
	Text     string
	Category string
	Found    bool
}

// LoreItemType represents different environmental storytelling formats.
type LoreItemType int

const (
	LoreItemNote LoreItemType = iota
	LoreItemAudioLog
	LoreItemGraffiti
	LoreItemBodyArrangement
)

// LoreItem represents an environmental storytelling element placed in the world.
type LoreItem struct {
	ID        string
	Type      LoreItemType
	PosX      float64
	PosY      float64
	Text      string
	Context   string
	CodexID   string
	Activated bool
}

// ContextType represents the context for lore generation.
type ContextType string

const (
	ContextCombat   ContextType = "combat"
	ContextLab      ContextType = "lab"
	ContextQuarters ContextType = "quarters"
	ContextStorage  ContextType = "storage"
	ContextEscape   ContextType = "escape"
	ContextGeneral  ContextType = "general"
)

// BackstoryType represents different types of world backstory entries.
type BackstoryType string

const (
	BackstoryWorld     BackstoryType = "world"
	BackstoryFaction   BackstoryType = "faction"
	BackstoryCharacter BackstoryType = "character"
	BackstoryLocation  BackstoryType = "location"
	BackstoryEvent     BackstoryType = "event"
	BackstoryArtifact  BackstoryType = "artifact"
)

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

	caser := cases.Title(language.English)
	return fmt.Sprintf("%s %s of %s", prefix, noun, caser.String(category))
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

// GenerateLoreText creates procedural narrative text based on genre and context.
func (g *Generator) GenerateLoreText(seed int64, context ContextType) string {
	localRng := rand.New(rand.NewSource(seed))

	templates := g.getContextTemplates(context)
	if len(templates) == 0 {
		templates = g.getTemplates()
	}

	// Generate 1-3 sentences for environmental lore
	sentenceCount := 1 + localRng.Intn(3)
	sentences := make([]string, sentenceCount)

	for i := 0; i < sentenceCount; i++ {
		template := templates[localRng.Intn(len(templates))]
		sentences[i] = g.fillTemplate(template, localRng)
	}

	return strings.Join(sentences, " ")
}

// getContextTemplates returns genre-specific templates for a given context.
func (g *Generator) getContextTemplates(context ContextType) []string {
	contextMap := map[string]map[ContextType][]string{
		"fantasy": {
			ContextCombat:   {"The {faction} fell defending {place}.", "Blood stains mark where {adj} warriors made their stand."},
			ContextLab:      {"Alchemical notes describe {adj} {experiment}.", "Tome fragments reveal {adj} incantations."},
			ContextQuarters: {"A diary entry: {adj} dreams haunt me.", "Personal effects of {faction} members."},
			ContextStorage:  {"Inventory log: {adj} supplies depleted.", "Quartermaster note: {resource} critically low."},
			ContextEscape:   {"Hurried scrawl: They're coming, must flee to {place}.", "The {faction} abandoned this route."},
		},
		"scifi": {
			ContextCombat:   {"Hull breach logged at sector {number}.", "Combat log: {faction} overwhelmed by {threat}."},
			ContextLab:      {"Research terminal {number}: {adj} anomaly detected.", "Experiment log: {adj} containment failure."},
			ContextQuarters: {"Crew manifest: {number} personnel missing.", "Personal log: Morale at critical levels."},
			ContextStorage:  {"Cargo manifest: {resource} supplies at {number}%.", "Automated alert: {adj} inventory shortage."},
			ContextEscape:   {"Emergency broadcast: Evacuate to {place}.", "Escape pod {number} launched prematurely."},
		},
		"horror": {
			ContextCombat:   {"Bloodstains spell out {adj} warnings.", "The bodies were arranged in {adj} patterns."},
			ContextLab:      {"Dissection notes reveal {adj} anatomy.", "The specimens exhibited {adj} mutations."},
			ContextQuarters: {"Child's drawing depicts {adj} figures.", "Diary ends mid-sentence: They're here."},
			ContextStorage:  {"Supplies rotted {number} days ago.", "The {resource} has been contaminated."},
			ContextEscape:   {"Desperate note: No way out, {place} is sealed.", "The exits lead only to {adj} darkness."},
		},
		"cyberpunk": {
			ContextCombat:   {"Security log: Intruders neutralized at {place}.", "Tactical net: {faction} strike force eliminated."},
			ContextLab:      {"Neural scan results: {adj} implant rejection.", "R&D log: {technology} prototype unstable."},
			ContextQuarters: {"Employee file {number}: Terminated for insubordination.", "Personal shard: Corp owns us all."},
			ContextStorage:  {"Inventory audit: {resource} diverted to {faction}.", "Shipping manifest falsified."},
			ContextEscape:   {"Encrypted message: Extraction at {place} compromised.", "Exit routes monitored by {faction}."},
		},
		"postapoc": {
			ContextCombat:   {"Bullet casings litter {place}.", "Last stand: {faction} fought here {number} days ago."},
			ContextLab:      {"Medical notes: Infection rate at {number}%.", "Research abandoned: No cure found."},
			ContextQuarters: {"Rationing schedule: {number} calories per day.", "Family photo, faces scratched out."},
			ContextStorage:  {"Supply cache: {resource} expires in {number} days.", "Scavenger marks indicate {faction} presence."},
			ContextEscape:   {"Map annotation: Route to {place} blocked.", "Survivor note: Don't go north."},
		},
	}

	genreMap, ok := contextMap[g.genre]
	if !ok {
		return nil
	}

	templates, ok := genreMap[context]
	if !ok {
		return genreMap[ContextGeneral]
	}

	return templates
}

// GenerateLoreItem creates an environmental lore item at a position.
func (g *Generator) GenerateLoreItem(id string, itemType LoreItemType, x, y float64, context ContextType) LoreItem {
	hash := hashString(id)
	text := g.GenerateLoreText(hash, context)

	// Generate matching codex entry ID
	codexID := fmt.Sprintf("codex_%s", id)

	return LoreItem{
		ID:        id,
		Type:      itemType,
		PosX:      x,
		PosY:      y,
		Text:      text,
		Context:   string(context),
		CodexID:   codexID,
		Activated: false,
	}
}

// GetLoreItemTypeName returns display name for lore item type.
func GetLoreItemTypeName(itemType LoreItemType, genre string) string {
	names := map[string]map[LoreItemType]string{
		"fantasy": {
			LoreItemNote:            "Scroll",
			LoreItemAudioLog:        "Echo Stone",
			LoreItemGraffiti:        "Runes",
			LoreItemBodyArrangement: "Burial Site",
		},
		"scifi": {
			LoreItemNote:            "Data Pad",
			LoreItemAudioLog:        "Audio Log",
			LoreItemGraffiti:        "Graffiti",
			LoreItemBodyArrangement: "Corpse Cluster",
		},
		"horror": {
			LoreItemNote:            "Note",
			LoreItemAudioLog:        "Voice Recording",
			LoreItemGraffiti:        "Blood Writing",
			LoreItemBodyArrangement: "Ritual Site",
		},
		"cyberpunk": {
			LoreItemNote:            "Memory Shard",
			LoreItemAudioLog:        "Neural Recording",
			LoreItemGraffiti:        "Holo-Tag",
			LoreItemBodyArrangement: "Crime Scene",
		},
		"postapoc": {
			LoreItemNote:            "Journal Page",
			LoreItemAudioLog:        "Radio Message",
			LoreItemGraffiti:        "Spray Paint",
			LoreItemBodyArrangement: "Mass Grave",
		},
	}

	genreNames, ok := names[genre]
	if !ok {
		genreNames = names["fantasy"]
	}

	name, ok := genreNames[itemType]
	if !ok {
		return "Unknown"
	}

	return name
}

// GenerateBackstoryEntry creates a procedural codex entry for world backstory.
func (g *Generator) GenerateBackstoryEntry(seed int64, backstoryType BackstoryType, index int) Entry {
	localRng := rand.New(rand.NewSource(seed + int64(index)))

	id := fmt.Sprintf("%s_%s_%d", g.genre, backstoryType, index)
	category := string(backstoryType)

	title := g.generateBackstoryTitle(backstoryType, localRng)
	text := g.generateBackstoryText(backstoryType, localRng)

	return Entry{
		ID:       id,
		Title:    title,
		Text:     text,
		Category: category,
		Found:    false,
	}
}

// generateBackstoryTitle creates a title for a backstory entry.
func (g *Generator) generateBackstoryTitle(backstoryType BackstoryType, rng *rand.Rand) string {
	templates := g.getBackstoryTitleTemplates(backstoryType)
	if len(templates) == 0 {
		return "Untitled Entry"
	}

	template := templates[rng.Intn(len(templates))]
	return g.fillTemplate(template, rng)
}

// generateBackstoryText creates narrative text for a backstory entry.
func (g *Generator) generateBackstoryText(backstoryType BackstoryType, rng *rand.Rand) string {
	templates := g.getBackstoryTextTemplates(backstoryType)
	if len(templates) == 0 {
		templates = g.getTemplates()
	}

	// Generate 3-5 sentences for backstory entries
	sentenceCount := 3 + rng.Intn(3)
	sentences := make([]string, sentenceCount)

	for i := 0; i < sentenceCount; i++ {
		template := templates[rng.Intn(len(templates))]
		sentences[i] = g.fillTemplate(template, rng)
	}

	return strings.Join(sentences, " ")
}

// getBackstoryTitleTemplates returns genre-specific title templates for backstory types.
func (g *Generator) getBackstoryTitleTemplates(backstoryType BackstoryType) []string {
	titleMap := map[string]map[BackstoryType][]string{
		"fantasy": {
			BackstoryWorld:     {"The Age of {adj} Magic", "The {adj} Realms", "Chronicles of the {faction}"},
			BackstoryFaction:   {"The {faction}", "Order of the {adj} {noun}", "The {adj} Brotherhood"},
			BackstoryCharacter: {"{adj} Hero of {place}", "The {adj} Champion", "Legend of the {adj} Warrior"},
			BackstoryLocation:  {"The {adj} {place}", "{place} of {power}", "The {adj} Citadel"},
			BackstoryEvent:     {"The {adj} War", "Fall of {place}", "The {adj} Cataclysm"},
			BackstoryArtifact:  {"The {adj} {artifact}", "{artifact} of {power}", "The {adj} Relic"},
		},
		"scifi": {
			BackstoryWorld:     {"Sector {number} Report", "The {adj} Galaxy", "Galactic Era {number}"},
			BackstoryFaction:   {"{faction} Dossier", "The {adj} Corporation", "Alliance File {number}"},
			BackstoryCharacter: {"Commander {adj} Report", "Agent File {number}", "The {adj} Pilot"},
			BackstoryLocation:  {"Station {number}", "{place} Outpost", "Planet {adj}"},
			BackstoryEvent:     {"The {adj} Conflict", "Incident Report {number}", "The {adj} Crisis"},
			BackstoryArtifact:  {"{technology} Prototype", "Device {number}", "The {adj} Technology"},
		},
		"horror": {
			BackstoryWorld:     {"The {adj} Darkness", "Age of {adj} Terror", "The {adj} Times"},
			BackstoryFaction:   {"The {adj} Cult", "{faction} Members", "The {adj} Circle"},
			BackstoryCharacter: {"The {adj} Victim", "Case File: {adj} Subject", "The {adj} One"},
			BackstoryLocation:  {"{place} of Dread", "The {adj} {place}", "Abandoned {place}"},
			BackstoryEvent:     {"The {adj} Incident", "Night of {adj} Horror", "The {adj} Ritual"},
			BackstoryArtifact:  {"The {adj} {artifact}", "Cursed {artifact}", "The {adj} Object"},
		},
		"cyberpunk": {
			BackstoryWorld:     {"The {adj} Net", "Corporate Era {number}", "The {adj} Grid"},
			BackstoryFaction:   {"{faction} Corp Profile", "The {adj} Syndicate", "Gang File {number}"},
			BackstoryCharacter: {"Runner ID {number}", "The {adj} Hacker", "Subject {number}"},
			BackstoryLocation:  {"District {number}", "The {adj} Zone", "{place} Sector"},
			BackstoryEvent:     {"The {adj} Breach", "Incident {number}", "The {adj} Blackout"},
			BackstoryArtifact:  {"{technology} Implant", "Chip {number}", "The {adj} Program"},
		},
		"postapoc": {
			BackstoryWorld:     {"The {adj} Collapse", "After the Fall", "Year {number} Post-Impact"},
			BackstoryFaction:   {"{faction} Survivors", "The {adj} Tribe", "Settlement {number}"},
			BackstoryCharacter: {"Survivor {number}", "The {adj} Wanderer", "Vault Dweller {number}"},
			BackstoryLocation:  {"The {adj} {place}", "Ruins of {place}", "Bunker {number}"},
			BackstoryEvent:     {"The {adj} War", "Day {number}", "The {adj} Exodus"},
			BackstoryArtifact:  {"Pre-War {artifact}", "Relic {number}", "The {adj} Cache"},
		},
	}

	genreMap, ok := titleMap[g.genre]
	if !ok {
		return nil
	}

	templates, ok := genreMap[backstoryType]
	if !ok {
		return nil
	}

	return templates
}

// getBackstoryTextTemplates returns genre-specific text templates for backstory types.
func (g *Generator) getBackstoryTextTemplates(backstoryType BackstoryType) []string {
	textMap := map[string]map[BackstoryType][]string{
		"fantasy": {
			BackstoryWorld:     {"The realm was forged in {adj} fires.", "Magic flowed through {place} like water.", "The {faction} ruled with {adj} wisdom."},
			BackstoryFaction:   {"Founded in the age of {noun}, the {faction} sought {goal}.", "Members wielded {power} against darkness.", "Their stronghold stood in {place}."},
			BackstoryCharacter: {"Born in {place}, they mastered {adj} arts.", "Their legend grew through {adj} deeds.", "They carried {artifact} into battle."},
			BackstoryLocation:  {"Built by the {faction}, {place} housed {adj} secrets.", "Its walls witnessed {adj} events.", "Magic saturated every stone."},
			BackstoryEvent:     {"The conflict began when {faction} discovered {artifact}.", "{adj} forces clashed at {place}.", "The aftermath reshaped the realm."},
			BackstoryArtifact:  {"Crafted from {adj} materials, it contained {power}.", "Legends spoke of its {adj} abilities.", "The {faction} guarded it zealously."},
		},
		"scifi": {
			BackstoryWorld:     {"Colonization began in sector {number}.", "The {faction} controlled jump routes through {place}.", "Technology advanced beyond {adj} limits."},
			BackstoryFaction:   {"Formed during the {adj} expansion, they monopolized {resource}.", "Their fleet numbered {number} vessels.", "Corporate headquarters on {place}."},
			BackstoryCharacter: {"Service record shows {number} missions.", "Augmented with {technology}.", "Last known coordinates: {place}."},
			BackstoryLocation:  {"Established year {number}, population {number}.", "Primary exports: {resource}.", "Strategic value: {adj}."},
			BackstoryEvent:     {"Initial reports indicate {adj} anomaly at {place}.", "Casualties estimated at {number}.", "Investigation ongoing."},
			BackstoryArtifact:  {"Prototype designation {number}.", "Research log indicates {adj} results.", "Current status: {adj}."},
		},
		"horror": {
			BackstoryWorld:     {"The darkness spread from {place}.", "{adj} entities emerged from shadows.", "Reality fractured in {adj} ways."},
			BackstoryFaction:   {"The cult worshipped {adj} entities.", "Rituals performed in {place}.", "Members numbered in the {number}s."},
			BackstoryCharacter: {"Last seen near {place}.", "Exhibited {adj} behavior.", "Found possessing {artifact}."},
			BackstoryLocation:  {"Site of {adj} occurrences.", "Locals avoid after dark.", "Strange sounds echo from {place}."},
			BackstoryEvent:     {"Witnesses describe {adj} phenomena.", "Occurred at {place} on {time}.", "No survivors."},
			BackstoryArtifact:  {"Origin unknown, properties {adj}.", "Proximity causes {adj} effects.", "Currently sealed in {place}."},
		},
		"cyberpunk": {
			BackstoryWorld:     {"The net evolved beyond {adj} control.", "Corporations divided {place} into districts.", "Data became the new {resource}."},
			BackstoryFaction:   {"Corporate charter {number} grants {adj} authority.", "Assets totaling {number} credits.", "Territory: {place}."},
			BackstoryCharacter: {"Neural profile: {adj}.", "Implant count: {number}.", "Last jack-in location: {place}."},
			BackstoryLocation:  {"Security rating: {adj}.", "Primary industry: {technology}.", "Population density: {number} per block."},
			BackstoryEvent:     {"Network breach detected at {place}.", "ICE response: {adj}.", "Damage estimate: {number} credits."},
			BackstoryArtifact:  {"Hardware specs: {adj}.", "Software version {number}.", "Black market value: {number} credits."},
		},
		"postapoc": {
			BackstoryWorld:     {"The bombs fell {number} years ago.", "Radiation levels remain {adj}.", "Survivors clustered in {place}."},
			BackstoryFaction:   {"Formed from vault {number} survivors.", "Controls {resource} supply to {place}.", "Numbers approximately {number}."},
			BackstoryCharacter: {"Vault ID {number}.", "Survived {number} days in wasteland.", "Last reported at {place}."},
			BackstoryLocation:  {"Pre-war population: {number}.", "Current status: {adj}.", "Scavenger reports: {adj} supplies."},
			BackstoryEvent:     {"Occurred {number} days after impact.", "Affected {place} region.", "Changed everything."},
			BackstoryArtifact:  {"Pre-war tech, condition {adj}.", "Functionality: {number}%.", "Located in {place}."},
		},
	}

	genreMap, ok := textMap[g.genre]
	if !ok {
		return nil
	}

	templates, ok := genreMap[backstoryType]
	if !ok {
		return nil
	}

	return templates
}

// GenerateWorldBackstory creates a complete set of backstory entries for a world seed.
func (g *Generator) GenerateWorldBackstory(worldSeed int64, entryCount int) []Entry {
	entries := make([]Entry, 0, entryCount)

	backstoryTypes := []BackstoryType{
		BackstoryWorld,
		BackstoryFaction,
		BackstoryCharacter,
		BackstoryLocation,
		BackstoryEvent,
		BackstoryArtifact,
	}

	for i := 0; i < entryCount; i++ {
		backstoryType := backstoryTypes[i%len(backstoryTypes)]
		entry := g.GenerateBackstoryEntry(worldSeed, backstoryType, i)
		entries = append(entries, entry)
	}

	return entries
}
