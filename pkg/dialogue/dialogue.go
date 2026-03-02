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
	genre      string
	rng        *rand.Rand
	nameGen    *NameGenerator
	grammarGen *GrammarGenerator
	choiceGen  *ChoiceGenerator
}

// NewGenerator creates a dialogue generator with a seed.
func NewGenerator(seed int64) *Generator {
	return &Generator{
		genre:      "fantasy",
		rng:        rand.New(rand.NewSource(seed)),
		nameGen:    NewNameGenerator(),
		grammarGen: NewGrammarGenerator(),
		choiceGen:  NewChoiceGenerator(),
	}
}

// SetGenre configures genre-specific dialogue themes.
func (g *Generator) SetGenre(genreID string) {
	g.genre = genreID
}

// Generate creates a procedural dialogue exchange.
func (g *Generator) Generate(id string, speakerType SpeakerType, dialogueType DialogueType) Dialogue {
	hash := hashString(id)

	// Use hash for deterministic generation
	speakerName := g.nameGen.Generate(g.genre, speakerType, hash)
	lines := g.grammarGen.Generate(g.genre, speakerType, dialogueType, hash+1)
	choices := g.choiceGen.Generate(g.genre, dialogueType, hash+2)

	return Dialogue{
		ID:          id,
		SpeakerName: speakerName,
		SpeakerType: speakerType,
		Type:        dialogueType,
		Lines:       lines,
		Choices:     choices,
	}
}

// GenerateMissionBriefing creates a mission briefing dialogue.
func (g *Generator) GenerateMissionBriefing(missionID, objectiveDesc string) Dialogue {
	hash := hashString(missionID)

	// Generate commander name procedurally
	commanderName := g.nameGen.Generate(g.genre, SpeakerCommander, hash)

	// Generate intro and conclusion procedurally
	intro := g.grammarGen.Generate(g.genre, SpeakerCommander, DialogueMissionBriefing, hash+1)
	conclusion := g.grammarGen.Generate(g.genre, SpeakerCommander, DialogueMissionComplete, hash+2)

	// Build briefing lines
	lines := []string{}
	if len(intro) > 0 {
		lines = append(lines, intro[0])
	}
	lines = append(lines, objectiveDesc)
	if len(conclusion) > 0 {
		lines = append(lines, conclusion[0])
	}

	// Generate choices procedurally
	choices := g.choiceGen.Generate(g.genre, DialogueMissionBriefing, hash+3)

	return Dialogue{
		ID:          missionID,
		SpeakerName: commanderName,
		SpeakerType: SpeakerCommander,
		Type:        DialogueMissionBriefing,
		Lines:       lines,
		Choices:     choices,
	}
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
