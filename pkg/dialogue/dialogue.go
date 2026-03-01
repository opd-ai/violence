// Package dialogue provides procedurally generated NPC conversations and mission briefings.
package dialogue

import (
	"fmt"
	"math/rand"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// DialogueType represents the type of conversation.
type DialogueType int

const (
	DialogueGreeting DialogueType = iota
	DialogueMissionBriefing
	DialogueMissionComplete
	DialogueIdle
	DialogueWarning
	DialogueTrade
	DialogueRumor
	DialogueQuest
)

// SpeakerType represents the category of NPC.
type SpeakerType int

const (
	SpeakerGuard SpeakerType = iota
	SpeakerMerchant
	SpeakerCommander
	SpeakerCivilian
	SpeakerTechnician
	SpeakerMystic
	SpeakerHostile
	SpeakerAlly
)

// Dialogue represents a single dialogue exchange.
type Dialogue struct {
	ID          string
	SpeakerName string
	SpeakerType SpeakerType
	Type        DialogueType
	Lines       []string
	Choices     []DialogueChoice
}

// DialogueChoice represents a player response option.
type DialogueChoice struct {
	Text     string
	NextID   string
	Outcome  string
	Required string // Required item/condition
}

// Generator procedurally generates dialogue from seeds.
type Generator struct {
	genre string
	rng   *rand.Rand
}

// NewGenerator creates a dialogue generator with a seed.
func NewGenerator(seed int64) *Generator {
	return &Generator{
		genre: "fantasy",
		rng:   rand.New(rand.NewSource(seed)),
	}
}

// SetGenre configures genre-specific dialogue themes.
func (g *Generator) SetGenre(genreID string) {
	g.genre = genreID
}

// Generate creates a procedural dialogue exchange.
func (g *Generator) Generate(id string, speakerType SpeakerType, dialogueType DialogueType) Dialogue {
	hash := hashString(id)
	localRng := rand.New(rand.NewSource(hash))

	speakerName := g.generateSpeakerName(speakerType, localRng)
	lines := g.generateLines(speakerType, dialogueType, localRng)
	choices := g.generateChoices(dialogueType, localRng)

	return Dialogue{
		ID:          id,
		SpeakerName: speakerName,
		SpeakerType: speakerType,
		Type:        dialogueType,
		Lines:       lines,
		Choices:     choices,
	}
}

// generateSpeakerName creates a genre-appropriate NPC name.
func (g *Generator) generateSpeakerName(speakerType SpeakerType, rng *rand.Rand) string {
	names := g.getSpeakerNames(speakerType)
	if len(names) == 0 {
		return "Unknown"
	}
	return names[rng.Intn(len(names))]
}

// getSpeakerNames returns genre-specific NPC names by type.
func (g *Generator) getSpeakerNames(speakerType SpeakerType) []string {
	nameMap := map[string]map[SpeakerType][]string{
		"fantasy": {
			SpeakerGuard:      {"Ser Roland", "Captain Thorne", "Watchman Gareth", "Guard Captain Elena"},
			SpeakerMerchant:   {"Merchant Aldric", "Trader Mira", "Shopkeeper Tobias", "Vendor Lyra"},
			SpeakerCommander:  {"Commander Varius", "General Kara", "Lord Commander Marcus", "Captain Elara"},
			SpeakerCivilian:   {"Villager Tom", "Citizen Sarah", "Farmer John", "Townsfolk Anna"},
			SpeakerTechnician: {"Alchemist Magnus", "Blacksmith Grendel", "Enchanter Sylvie"},
			SpeakerMystic:     {"Oracle Zara", "Sage Aldwin", "Seer Morgana", "Mystic Raven"},
			SpeakerHostile:    {"Bandit Chief", "Cultist Leader", "Dark Mage", "Raider Boss"},
			SpeakerAlly:       {"Companion Rex", "Ally Vera", "Friend Marcus", "Comrade Talia"},
		},
		"scifi": {
			SpeakerGuard:      {"Security Officer Kane", "Guard Unit-7", "Sentinel Reyes", "Officer Park"},
			SpeakerMerchant:   {"Trader Vex", "Merchant Hiro", "Vendor AI-3", "Supplier Chen"},
			SpeakerCommander:  {"Commander Zhao", "Admiral Torres", "Captain Morrison", "Colonel Vasquez"},
			SpeakerCivilian:   {"Colonist Lee", "Resident Singh", "Citizen Nova", "Civilian Rivera"},
			SpeakerTechnician: {"Engineer Costa", "Technician AI-12", "Mechanic Brooks", "Systems Analyst Kim"},
			SpeakerMystic:     {"AI Oracle", "Psi-Analyst Dr. Wells", "Neural Specialist Tran"},
			SpeakerHostile:    {"Rogue AI", "Pirate Captain", "Hostile Commander", "Enemy Leader"},
			SpeakerAlly:       {"Squad Member Delta", "Ally Unit-5", "Companion Nexus", "Friend Cortez"},
		},
		"horror": {
			SpeakerGuard:      {"Security Guard Mike", "Night Watchman Bill", "Officer Jenkins", "Guard Stevens"},
			SpeakerMerchant:   {"Strange Vendor", "Pawn Shop Owner", "Old Trader", "Merchant Grim"},
			SpeakerCommander:  {"Director Blackwood", "Chief Investigator", "Head Researcher", "Warden Cross"},
			SpeakerCivilian:   {"Survivor Emma", "Witness David", "Patient Sarah", "Victim's Friend"},
			SpeakerTechnician: {"Lab Tech Morris", "Maintenance Worker", "Doctor Hayes", "Analyst Gray"},
			SpeakerMystic:     {"Medium Claire", "Priest Father Thomas", "Occultist Ash", "Psychic Vera"},
			SpeakerHostile:    {"Cult Leader", "Possessed One", "The Entity", "Dark Presence"},
			SpeakerAlly:       {"Fellow Survivor", "Investigator Partner", "Companion Alex", "Friend Jordan"},
		},
		"cyberpunk": {
			SpeakerGuard:      {"Corp Security Yamamoto", "Enforcer Brick", "Guard Unit 451", "Officer Diaz"},
			SpeakerMerchant:   {"Fixer Nyx", "Black Market Dealer", "Tech Vendor Cruz", "Merchant Glitch"},
			SpeakerCommander:  {"Corp Exec Morgan", "Gang Leader Razor", "Boss Santos", "Director Tanaka"},
			SpeakerCivilian:   {"Street Kid Miko", "Resident 2077", "Citizen Watts", "Civilian Nash"},
			SpeakerTechnician: {"Netrunner Ghost", "Ripper Doc", "Tech Specialist Link", "Mechanic Sparks"},
			SpeakerMystic:     {"AI Prophet", "Cyber-Shaman Zero", "Net Oracle", "Digital Medium"},
			SpeakerHostile:    {"Gang Boss", "Rogue AI Entity", "Corp Hitman", "Hostile Hacker"},
			SpeakerAlly:       {"Runner Ally", "Net Friend Cipher", "Companion Byte", "Partner Chrome"},
		},
		"postapoc": {
			SpeakerGuard:      {"Guard Jackson", "Sentry Miller", "Watchman Cruz", "Defender Reyes"},
			SpeakerMerchant:   {"Scrap Trader Pete", "Merchant Sal", "Vendor Red", "Supplier Doc"},
			SpeakerCommander:  {"Settlement Leader", "Chief Carter", "Boss Stone", "Commander Wolf"},
			SpeakerCivilian:   {"Survivor Jane", "Scavenger Tom", "Wanderer Max", "Settler Ruby"},
			SpeakerTechnician: {"Mechanic Rusty", "Engineer Jules", "Tech Salvager", "Fixer Quinn"},
			SpeakerMystic:     {"Wasteland Prophet", "Seer Ash", "Oracle Dust", "Shaman Crow"},
			SpeakerHostile:    {"Raider Chief", "Mutant Leader", "Bandit Boss", "Hostile Warlord"},
			SpeakerAlly:       {"Fellow Wanderer", "Companion Dog", "Ally Viper", "Friend Scout"},
		},
	}

	genreMap, ok := nameMap[g.genre]
	if !ok {
		genreMap = nameMap["fantasy"]
	}

	names, ok := genreMap[speakerType]
	if !ok {
		return []string{"Unknown"}
	}

	return names
}

// generateLines creates dialogue lines based on type and speaker.
func (g *Generator) generateLines(speakerType SpeakerType, dialogueType DialogueType, rng *rand.Rand) []string {
	templates := g.getDialogueTemplates(speakerType, dialogueType)
	if len(templates) == 0 {
		return []string{"..."}
	}

	// Generate 1-3 lines
	lineCount := 1 + rng.Intn(3)
	lines := make([]string, lineCount)

	for i := 0; i < lineCount; i++ {
		template := templates[rng.Intn(len(templates))]
		lines[i] = g.fillTemplate(template, rng)
	}

	return lines
}

// getDialogueTemplates returns templates for dialogue generation.
func (g *Generator) getDialogueTemplates(speakerType SpeakerType, dialogueType DialogueType) []string {
	// Genre-specific templates
	templateMap := g.getGenreTemplates()

	key := fmt.Sprintf("%d_%d", speakerType, dialogueType)
	templates, ok := templateMap[key]
	if !ok {
		// Fallback to generic greetings
		return g.getFallbackTemplates(dialogueType)
	}

	return templates
}

// getGenreTemplates returns all dialogue templates for current genre.
func (g *Generator) getGenreTemplates() map[string][]string {
	genreTemplates := map[string]map[string][]string{
		"fantasy":   g.getFantasyTemplates(),
		"scifi":     g.getScifiTemplates(),
		"horror":    g.getHorrorTemplates(),
		"cyberpunk": g.getCyberpunkTemplates(),
		"postapoc":  g.getPostapocTemplates(),
	}

	templates, ok := genreTemplates[g.genre]
	if !ok {
		return genreTemplates["fantasy"]
	}
	return templates
}

// getFantasyTemplates returns fantasy genre dialogue templates.
func (g *Generator) getFantasyTemplates() map[string][]string {
	return map[string][]string{
		"0_0": {"Halt, traveler! State your business.", "Greetings, stranger. Welcome to our {place}.", "Well met, adventurer."},
		"0_1": {"The {faction} has tasked me with a {adj} mission. Will you aid us?", "We need someone brave to venture into the {place}.", "A {adj} threat looms. We require your skills."},
		"0_2": {"You have done well, hero! The {place} is safe.", "Your deeds will be remembered in legend.", "Accept this reward for your {adj} service."},
		"1_0": {"Welcome to my shop, friend. What are you seeking?", "I have {adj} wares for sale. Take a look.", "Coin for goods, that's my trade."},
		"2_1": {"Listen carefully. Your objective is to {goal}.", "The enemy controls the {place}. Neutralize them.", "Retrieve the {artifact} and return safely."},
		"3_0": {"Please, you must help us! The {faction} took everything.", "I've seen {adj} things in the {place}.", "Be careful out there, stranger."},
		"5_6": {"The ancient scrolls speak of {adj} {artifact}.", "Magic flows strangely in the {place}.", "I sense {adj} forces at work."},
		"6_4": {"Surrender or face {adj} consequences!", "You dare intrude upon our {place}?", "Leave now or suffer the wrath of the {faction}!"},
	}
}

// getScifiTemplates returns scifi genre dialogue templates.
func (g *Generator) getScifiTemplates() map[string][]string {
	return map[string][]string{
		"0_0": {"Halt! Identify yourself, citizen.", "Welcome to sector {number}. State your purpose.", "Security checkpoint. ID please."},
		"0_1": {"Command has authorized a mission to {place}. Briefing follows.", "We've detected {adj} anomalies in sector {number}.", "Your target is in the {place}. Proceed with caution."},
		"0_2": {"Mission accomplished. Return to base for debrief.", "Excellent work, operative. Command is pleased.", "Bonus credits transferred to your account."},
		"1_0": {"Looking to buy or sell? I've got {adj} tech.", "Credits for goods. No questions asked.", "Check out my inventory, friend."},
		"2_1": {"Your mission parameters: infiltrate {place} and {goal}.", "Eliminate all hostile contacts in sector {number}.", "Retrieve the data core and exfiltrate."},
		"3_0": {"Help us! The station is compromised!", "I saw {adj} things in the {place}.", "Don't trust the AI systems."},
		"5_6": {"The data suggests {adj} patterns emerging.", "Neural scans indicate {adj} activity.", "My calculations point to the {place}."},
		"6_4": {"Terminate intruder! Security breach detected!", "You've violated corporate property. Surrender!", "Lethal force authorized!"},
	}
}

// getHorrorTemplates returns horror genre dialogue templates.
func (g *Generator) getHorrorTemplates() map[string][]string {
	return map[string][]string{
		"0_0": {"Oh thank God, another person! I thought I was alone.", "Don't go in there... please.", "You shouldn't be here."},
		"0_1": {"We need to get out of here. Help me find the exit.", "The {place} is not safe. We need to {goal}.", "Something {adj} is hunting us."},
		"0_2": {"We made it... I can't believe we survived.", "Never speak of this place again.", "At least it's over now."},
		"3_0": {"They took my family... those {adj} things.", "I heard screams from the {place}.", "Don't trust anyone. They could be infected."},
		"5_6": {"The spirits warn of {adj} danger ahead.", "I've performed the ritual. You must {goal}.", "Dark forces converge on the {place}."},
		"6_4": {"Join us... become one with the {faction}...", "Your flesh will serve our purpose.", "Resistance is futile. Embrace the darkness!"},
	}
}

// getCyberpunkTemplates returns cyberpunk genre dialogue templates.
func (g *Generator) getCyberpunkTemplates() map[string][]string {
	return map[string][]string{
		"0_0": {"Citizen {number}, your credentials check out. Move along.", "Corp territory ahead. You got clearance?", "Nice chrome, choom."},
		"0_1": {"I've got a job for you. Hack into {place} and {goal}.", "The corp needs {adj} data extracted from sector {number}.", "Interested in some side work? Pays well."},
		"0_2": {"Clean work. Credits deposited. Pleasure doing business.", "Corp's happy. You did good, runner.", "Nice job. Here's your cut."},
		"1_0": {"Selling {adj} implants and hardware. Interested?", "Got chrome, got programs, got everything. What you need?", "Black market specials today only."},
		"2_1": {"Jack into {place} and extract the {artifact}.", "Your target: neutralize {adj} security in sector {number}.", "Stealth recommended. ICE is {adj}."},
		"4_0": {"I can mod your chrome. Make you {adj}.", "Need upgrades? I'm your tech.", "Bring me {adj} hardware and I'll install it."},
		"5_6": {"The net shows {adj} patterns. Something big is coming.", "My implants detect {adj} signals from {place}.", "The code doesn't lie. Danger ahead."},
	}
}

// getPostapocTemplates returns postapoc genre dialogue templates.
func (g *Generator) getPostapocTemplates() map[string][]string {
	return map[string][]string{
		"0_0": {"Stop right there. What's your business in our settlement?", "Another survivor. Didn't think I'd see anyone today.", "You look like you've traveled far."},
		"0_1": {"We need someone to scout the {place}. It's dangerous.", "Raiders control the {place}. We need them gone.", "Find {adj} supplies and bring them back."},
		"0_2": {"Good work, wanderer. You've earned your keep.", "The settlement thanks you. Here's your share.", "You did what needed doing."},
		"1_0": {"Trading scrap and supplies. What you got?", "I'll trade {adj} gear for food or water.", "Ammunition's scarce but I've got some."},
		"2_1": {"Your mission: clear the {place} of hostiles.", "We need {adj} supplies from the ruins. Bring what you can.", "Scout ahead and report what you find."},
		"3_0": {"Please, we're starving. Do you have any food?", "The raiders took everything. We have nothing left.", "I saw {adj} mutants near the {place}."},
		"5_6": {"The wasteland speaks to those who listen.", "I've seen visions of {adj} times ahead.", "The radiation reveals truths."},
	}
}

// getFallbackTemplates returns generic templates for unknown combinations.
func (g *Generator) getFallbackTemplates(dialogueType DialogueType) []string {
	fallbacks := map[DialogueType][]string{
		DialogueGreeting:        {"Hello.", "Greetings.", "What do you want?"},
		DialogueMissionBriefing: {"Here's what you need to do.", "Listen carefully.", "Your mission is simple."},
		DialogueMissionComplete: {"Well done.", "Good job.", "Mission complete."},
		DialogueIdle:            {"...", "Nothing to say.", "Move along."},
		DialogueWarning:         {"Be careful!", "Watch out!", "Danger ahead!"},
		DialogueTrade:           {"Want to trade?", "I have goods.", "Looking to buy?"},
		DialogueRumor:           {"I heard something interesting.", "Rumor has it...", "Word on the street is..."},
		DialogueQuest:           {"I have a task for you.", "Will you help?", "I need assistance."},
	}

	templates, ok := fallbacks[dialogueType]
	if !ok {
		return []string{"..."}
	}
	return templates
}

// generateChoices creates player response options.
func (g *Generator) generateChoices(dialogueType DialogueType, rng *rand.Rand) []DialogueChoice {
	// Only certain dialogue types have choices
	switch dialogueType {
	case DialogueMissionBriefing:
		return []DialogueChoice{
			{Text: "I'll do it.", NextID: "accept", Outcome: "mission_accepted"},
			{Text: "Not interested.", NextID: "decline", Outcome: "mission_declined"},
			{Text: "What's the reward?", NextID: "reward", Outcome: "ask_reward"},
		}
	case DialogueTrade:
		return []DialogueChoice{
			{Text: "Show me your goods.", NextID: "shop", Outcome: "open_shop"},
			{Text: "Not now.", NextID: "leave", Outcome: "close_dialogue"},
		}
	case DialogueQuest:
		return []DialogueChoice{
			{Text: "Tell me more.", NextID: "details", Outcome: "quest_details"},
			{Text: "I accept.", NextID: "accept", Outcome: "quest_accepted"},
			{Text: "Maybe later.", NextID: "decline", Outcome: "quest_declined"},
		}
	default:
		return []DialogueChoice{
			{Text: "Goodbye.", NextID: "end", Outcome: "close_dialogue"},
		}
	}
}

// fillTemplate fills dialogue template placeholders with procedural values.
func (g *Generator) fillTemplate(template string, rng *rand.Rand) string {
	adjectives := []string{"dangerous", "mysterious", "urgent", "critical", "important", "secret"}
	places := []string{"the ruins", "the sector", "the facility", "the area", "the zone", "the depths"}
	factions := []string{"the Order", "the Collective", "command", "the Council", "the faction"}
	artifacts := []string{"data core", "artifact", "device", "relic", "item"}
	goals := []string{"secure the area", "eliminate threats", "retrieve the objective", "investigate"}

	result := template
	result = strings.ReplaceAll(result, "{adj}", adjectives[rng.Intn(len(adjectives))])
	result = strings.ReplaceAll(result, "{place}", places[rng.Intn(len(places))])
	result = strings.ReplaceAll(result, "{faction}", factions[rng.Intn(len(factions))])
	result = strings.ReplaceAll(result, "{artifact}", artifacts[rng.Intn(len(artifacts))])
	result = strings.ReplaceAll(result, "{goal}", goals[rng.Intn(len(goals))])
	result = strings.ReplaceAll(result, "{number}", fmt.Sprintf("%d", 100+rng.Intn(900)))

	return result
}

// GenerateMissionBriefing creates a mission briefing dialogue.
func (g *Generator) GenerateMissionBriefing(missionID, objectiveDesc string) Dialogue {
	hash := hashString(missionID)
	localRng := rand.New(rand.NewSource(hash))

	commanderName := g.generateSpeakerName(SpeakerCommander, localRng)

	// Build briefing lines
	lines := []string{
		g.getGenreSpecificIntro(localRng),
		objectiveDesc,
		g.getGenreSpecificConclusion(localRng),
	}

	return Dialogue{
		ID:          missionID,
		SpeakerName: commanderName,
		SpeakerType: SpeakerCommander,
		Type:        DialogueMissionBriefing,
		Lines:       lines,
		Choices: []DialogueChoice{
			{Text: "Understood.", NextID: "accept", Outcome: "mission_started"},
			{Text: "I need more details.", NextID: "details", Outcome: "show_details"},
		},
	}
}

// getGenreSpecificIntro returns genre-appropriate mission intro.
func (g *Generator) getGenreSpecificIntro(rng *rand.Rand) string {
	intros := map[string][]string{
		"fantasy":   {"Listen carefully, adventurer.", "The Council has issued new orders.", "Your next quest begins now."},
		"scifi":     {"Mission briefing initiated.", "Command has new orders for you.", "Attention, operative."},
		"horror":    {"Please, you have to listen to me.", "There's no time to explain everything.", "You need to know this."},
		"cyberpunk": {"Got a job for you, choom.", "Corp needs this done ASAP.", "Listen up, runner."},
		"postapoc":  {"Gather 'round, we've got a situation.", "Settlement needs your help again.", "Listen close."},
	}

	genreIntros, ok := intros[g.genre]
	if !ok {
		genreIntros = intros["fantasy"]
	}

	return genreIntros[rng.Intn(len(genreIntros))]
}

// getGenreSpecificConclusion returns genre-appropriate mission conclusion.
func (g *Generator) getGenreSpecificConclusion(rng *rand.Rand) string {
	conclusions := map[string][]string{
		"fantasy":   {"May fortune favor you.", "Return when you have completed this task.", "The realm depends on you."},
		"scifi":     {"Good luck out there. Command out.", "Mission parameters uploaded. Proceed.", "Rendezvous at extraction point."},
		"horror":    {"Please be careful. We've lost too many already.", "Don't end up like the others.", "Come back alive."},
		"cyberpunk": {"Don't flatline on me. I need results.", "Watch your back out there.", "Payment on completion."},
		"postapoc":  {"Stay sharp. The wasteland is unforgiving.", "Come back in one piece.", "We're counting on you."},
	}

	genreConclusions, ok := conclusions[g.genre]
	if !ok {
		genreConclusions = conclusions["fantasy"]
	}

	return genreConclusions[rng.Intn(len(genreConclusions))]
}

// GenerateContextualDialogue creates dialogue based on game state context.
func (g *Generator) GenerateContextualDialogue(context string, speakerType SpeakerType) Dialogue {
	dialogueType := DialogueIdle
	if strings.Contains(context, "mission") {
		dialogueType = DialogueMissionBriefing
	} else if strings.Contains(context, "trade") {
		dialogueType = DialogueTrade
	} else if strings.Contains(context, "warning") {
		dialogueType = DialogueWarning
	}

	return g.Generate(context, speakerType, dialogueType)
}

func hashString(s string) int64 {
	var hash int64
	for i := 0; i < len(s); i++ {
		hash = hash*31 + int64(s[i])
	}
	return hash
}

// FormatDialogue returns formatted dialogue text for display.
func FormatDialogue(d Dialogue) string {
	caser := cases.Title(language.English)
	var sb strings.Builder

	sb.WriteString(caser.String(d.SpeakerName))
	sb.WriteString(":\n")

	for _, line := range d.Lines {
		sb.WriteString("  ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	if len(d.Choices) > 0 {
		sb.WriteString("\nResponses:\n")
		for i, choice := range d.Choices {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, choice.Text))
		}
	}

	return sb.String()
}
