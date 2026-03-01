package dialogue

import (
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator(12345)
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if gen.genre != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got '%s'", gen.genre)
	}
}

func TestSetGenre(t *testing.T) {
	gen := NewGenerator(12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		gen.SetGenre(genre)
		if gen.genre != genre {
			t.Errorf("SetGenre(%s) failed, got %s", genre, gen.genre)
		}
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name         string
		genre        string
		speakerType  SpeakerType
		dialogueType DialogueType
	}{
		{"fantasy guard greeting", "fantasy", SpeakerGuard, DialogueGreeting},
		{"scifi merchant trade", "scifi", SpeakerMerchant, DialogueTrade},
		{"horror civilian warning", "horror", SpeakerCivilian, DialogueWarning},
		{"cyberpunk commander briefing", "cyberpunk", SpeakerCommander, DialogueMissionBriefing},
		{"postapoc mystic quest", "postapoc", SpeakerMystic, DialogueQuest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(12345)
			gen.SetGenre(tt.genre)

			dialogue := gen.Generate("test_id", tt.speakerType, tt.dialogueType)

			if dialogue.ID != "test_id" {
				t.Errorf("expected ID 'test_id', got '%s'", dialogue.ID)
			}
			if dialogue.SpeakerType != tt.speakerType {
				t.Errorf("expected SpeakerType %d, got %d", tt.speakerType, dialogue.SpeakerType)
			}
			if dialogue.Type != tt.dialogueType {
				t.Errorf("expected DialogueType %d, got %d", tt.dialogueType, dialogue.Type)
			}
			if dialogue.SpeakerName == "" {
				t.Error("SpeakerName should not be empty")
			}
			if len(dialogue.Lines) == 0 {
				t.Error("Lines should not be empty")
			}
		})
	}
}

func TestGenerateDeterministic(t *testing.T) {
	gen1 := NewGenerator(42)
	gen2 := NewGenerator(42)

	d1 := gen1.Generate("npc_001", SpeakerGuard, DialogueGreeting)
	d2 := gen2.Generate("npc_001", SpeakerGuard, DialogueGreeting)

	if d1.SpeakerName != d2.SpeakerName {
		t.Errorf("non-deterministic speaker name: %s vs %s", d1.SpeakerName, d2.SpeakerName)
	}
	if len(d1.Lines) != len(d2.Lines) {
		t.Errorf("non-deterministic line count: %d vs %d", len(d1.Lines), len(d2.Lines))
	}
	for i := range d1.Lines {
		if d1.Lines[i] != d2.Lines[i] {
			t.Errorf("non-deterministic line %d: %s vs %s", i, d1.Lines[i], d2.Lines[i])
		}
	}
}

func TestGenerateDifferentIDs(t *testing.T) {
	gen := NewGenerator(42)

	d1 := gen.Generate("npc_001", SpeakerGuard, DialogueGreeting)
	d2 := gen.Generate("npc_002", SpeakerGuard, DialogueGreeting)

	// Different IDs should produce different results
	different := false
	if d1.SpeakerName != d2.SpeakerName {
		different = true
	}
	if len(d1.Lines) != len(d2.Lines) {
		different = true
	} else {
		for i := range d1.Lines {
			if d1.Lines[i] != d2.Lines[i] {
				different = true
				break
			}
		}
	}

	if !different {
		t.Error("different IDs should produce different dialogue")
	}
}

func TestGetSpeakerNames(t *testing.T) {
	tests := []struct {
		genre       string
		speakerType SpeakerType
	}{
		{"fantasy", SpeakerGuard},
		{"scifi", SpeakerMerchant},
		{"horror", SpeakerCommander},
		{"cyberpunk", SpeakerCivilian},
		{"postapoc", SpeakerTechnician},
	}

	for _, tt := range tests {
		t.Run(tt.genre+"_"+string(rune('0'+tt.speakerType)), func(t *testing.T) {
			gen := NewGenerator(12345)
			gen.SetGenre(tt.genre)

			names := gen.getSpeakerNames(tt.speakerType)
			if len(names) == 0 {
				t.Errorf("no names returned for genre %s, speaker %d", tt.genre, tt.speakerType)
			}
		})
	}
}

func TestGenerateLines(t *testing.T) {
	gen := NewGenerator(12345)

	tests := []struct {
		name         string
		speakerType  SpeakerType
		dialogueType DialogueType
	}{
		{"guard greeting", SpeakerGuard, DialogueGreeting},
		{"merchant trade", SpeakerMerchant, DialogueTrade},
		{"commander briefing", SpeakerCommander, DialogueMissionBriefing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := gen.generateLines(tt.speakerType, tt.dialogueType, gen.rng)

			if len(lines) == 0 {
				t.Error("generateLines returned empty slice")
			}
			for i, line := range lines {
				if line == "" {
					t.Errorf("line %d is empty", i)
				}
			}
		})
	}
}

func TestGenerateChoices(t *testing.T) {
	gen := NewGenerator(12345)

	tests := []struct {
		dialogueType    DialogueType
		expectedChoices bool
	}{
		{DialogueMissionBriefing, true},
		{DialogueTrade, true},
		{DialogueQuest, true},
		{DialogueGreeting, true},
		{DialogueIdle, true},
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.dialogueType)), func(t *testing.T) {
			choices := gen.generateChoices(tt.dialogueType, gen.rng)

			if tt.expectedChoices && len(choices) == 0 {
				t.Errorf("expected choices for dialogue type %d, got none", tt.dialogueType)
			}

			for i, choice := range choices {
				if choice.Text == "" {
					t.Errorf("choice %d has empty text", i)
				}
				if choice.NextID == "" {
					t.Errorf("choice %d has empty NextID", i)
				}
				if choice.Outcome == "" {
					t.Errorf("choice %d has empty Outcome", i)
				}
			}
		})
	}
}

func TestGenerateMissionBriefing(t *testing.T) {
	gen := NewGenerator(12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			gen.SetGenre(genre)

			briefing := gen.GenerateMissionBriefing("mission_001", "Eliminate all hostiles in sector 7")

			if briefing.ID != "mission_001" {
				t.Errorf("expected ID 'mission_001', got '%s'", briefing.ID)
			}
			if briefing.SpeakerType != SpeakerCommander {
				t.Errorf("expected SpeakerCommander, got %d", briefing.SpeakerType)
			}
			if briefing.Type != DialogueMissionBriefing {
				t.Errorf("expected DialogueMissionBriefing, got %d", briefing.Type)
			}
			if len(briefing.Lines) < 2 {
				t.Errorf("expected at least 2 lines, got %d", len(briefing.Lines))
			}
			if len(briefing.Choices) == 0 {
				t.Error("expected choices for mission briefing")
			}

			// Check that objective description is included
			found := false
			for _, line := range briefing.Lines {
				if strings.Contains(line, "sector 7") {
					found = true
					break
				}
			}
			if !found {
				t.Error("mission briefing should include objective description")
			}
		})
	}
}

func TestGenerateContextualDialogue(t *testing.T) {
	gen := NewGenerator(12345)

	tests := []struct {
		context      string
		expectedType DialogueType
	}{
		{"mission_start", DialogueMissionBriefing},
		{"trade_interaction", DialogueTrade},
		{"warning_message", DialogueWarning},
		{"idle_chat", DialogueIdle},
	}

	for _, tt := range tests {
		t.Run(tt.context, func(t *testing.T) {
			dialogue := gen.GenerateContextualDialogue(tt.context, SpeakerGuard)

			if dialogue.Type != tt.expectedType {
				t.Errorf("expected type %d, got %d", tt.expectedType, dialogue.Type)
			}
			if len(dialogue.Lines) == 0 {
				t.Error("expected dialogue lines")
			}
		})
	}
}

func TestFormatDialogue(t *testing.T) {
	dialogue := Dialogue{
		ID:          "test",
		SpeakerName: "Captain Smith",
		SpeakerType: SpeakerCommander,
		Type:        DialogueMissionBriefing,
		Lines: []string{
			"Attention, soldier.",
			"Your mission is critical.",
		},
		Choices: []DialogueChoice{
			{Text: "Yes sir!", NextID: "accept", Outcome: "mission_started"},
			{Text: "I need details.", NextID: "details", Outcome: "show_details"},
		},
	}

	formatted := FormatDialogue(dialogue)

	if !strings.Contains(formatted, "Captain Smith") {
		t.Error("formatted dialogue should contain speaker name")
	}
	if !strings.Contains(formatted, "Attention, soldier") {
		t.Error("formatted dialogue should contain first line")
	}
	if !strings.Contains(formatted, "Your mission is critical") {
		t.Error("formatted dialogue should contain second line")
	}
	if !strings.Contains(formatted, "Yes sir!") {
		t.Error("formatted dialogue should contain first choice")
	}
	if !strings.Contains(formatted, "I need details") {
		t.Error("formatted dialogue should contain second choice")
	}
}

func TestHashString(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"test"},
		{"npc_001"},
		{"mission_briefing_alpha"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			hash1 := hashString(tt.input)
			hash2 := hashString(tt.input)

			if hash1 != hash2 {
				t.Error("hashString should be deterministic")
			}
		})
	}

	// Different strings should produce different hashes
	hash1 := hashString("test1")
	hash2 := hashString("test2")
	if hash1 == hash2 {
		t.Error("different strings should produce different hashes")
	}
}

func TestAllGenresHaveTemplates(t *testing.T) {
	gen := NewGenerator(12345)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			gen.SetGenre(genre)

			// Test each speaker type with each dialogue type
			speakerTypes := []SpeakerType{
				SpeakerGuard, SpeakerMerchant, SpeakerCommander,
				SpeakerCivilian, SpeakerTechnician, SpeakerMystic,
				SpeakerHostile, SpeakerAlly,
			}
			dialogueTypes := []DialogueType{
				DialogueGreeting, DialogueMissionBriefing, DialogueMissionComplete,
				DialogueIdle, DialogueWarning, DialogueTrade, DialogueRumor, DialogueQuest,
			}

			for _, st := range speakerTypes {
				for _, dt := range dialogueTypes {
					dialogue := gen.Generate("test", st, dt)
					if len(dialogue.Lines) == 0 {
						t.Errorf("no lines for genre=%s speaker=%d type=%d", genre, st, dt)
					}
				}
			}
		})
	}
}

func TestFillTemplate(t *testing.T) {
	gen := NewGenerator(42)

	tests := []struct {
		template string
		contains []string
	}{
		{
			"The {faction} controls {place}",
			[]string{"The ", " controls "},
		},
		{
			"Sector {number} has {adj} readings",
			[]string{"Sector ", " has ", " readings"},
		},
		{
			"Retrieve {artifact} from {place}",
			[]string{"Retrieve ", " from "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			result := gen.fillTemplate(tt.template, gen.rng)

			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("result should contain '%s', got: %s", substr, result)
				}
			}

			// Check that placeholders are replaced
			if strings.Contains(result, "{") || strings.Contains(result, "}") {
				t.Errorf("template should not contain placeholders after fill: %s", result)
			}
		})
	}
}

func TestDialogueChoicesStructure(t *testing.T) {
	gen := NewGenerator(12345)

	dialogue := gen.Generate("test", SpeakerCommander, DialogueMissionBriefing)

	for i, choice := range dialogue.Choices {
		if choice.Text == "" {
			t.Errorf("choice %d has empty Text", i)
		}
		if choice.NextID == "" {
			t.Errorf("choice %d has empty NextID", i)
		}
		if choice.Outcome == "" {
			t.Errorf("choice %d has empty Outcome", i)
		}
	}
}

func TestGenreSpecificNames(t *testing.T) {
	gen := NewGenerator(12345)

	// Fantasy names should not appear in scifi
	gen.SetGenre("scifi")
	scifiNames := gen.getSpeakerNames(SpeakerGuard)
	for _, name := range scifiNames {
		if strings.Contains(name, "Ser ") || strings.Contains(name, "Knight") {
			t.Errorf("scifi should not have fantasy names: %s", name)
		}
	}

	// Horror names should be appropriate
	gen.SetGenre("horror")
	horrorNames := gen.getSpeakerNames(SpeakerCivilian)
	if len(horrorNames) == 0 {
		t.Error("horror genre should have civilian names")
	}
}

func TestEmptyIDHandling(t *testing.T) {
	gen := NewGenerator(12345)

	dialogue := gen.Generate("", SpeakerGuard, DialogueGreeting)

	if dialogue.ID != "" {
		t.Errorf("expected empty ID preserved, got '%s'", dialogue.ID)
	}
	if len(dialogue.Lines) == 0 {
		t.Error("dialogue should still generate lines with empty ID")
	}
}

func TestSpeakerTypeConsistency(t *testing.T) {
	gen := NewGenerator(12345)

	allTypes := []SpeakerType{
		SpeakerGuard, SpeakerMerchant, SpeakerCommander,
		SpeakerCivilian, SpeakerTechnician, SpeakerMystic,
		SpeakerHostile, SpeakerAlly,
	}

	for _, st := range allTypes {
		dialogue := gen.Generate("test", st, DialogueGreeting)
		if dialogue.SpeakerType != st {
			t.Errorf("expected SpeakerType %d, got %d", st, dialogue.SpeakerType)
		}
	}
}

func TestMissionBriefingIntroConclusion(t *testing.T) {
	gen := NewGenerator(12345)
	gen.SetGenre("fantasy")

	briefing := gen.GenerateMissionBriefing("mission_001", "Test objective")

	if len(briefing.Lines) < 3 {
		t.Error("mission briefing should have intro, objective, and conclusion")
	}

	// Lines should be non-empty
	for i, line := range briefing.Lines {
		if line == "" {
			t.Errorf("line %d is empty", i)
		}
	}
}

func BenchmarkGenerate(b *testing.B) {
	gen := NewGenerator(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate("npc_001", SpeakerGuard, DialogueGreeting)
	}
}

func BenchmarkGenerateMissionBriefing(b *testing.B) {
	gen := NewGenerator(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GenerateMissionBriefing("mission_001", "Test objective")
	}
}

func BenchmarkFormatDialogue(b *testing.B) {
	gen := NewGenerator(12345)
	dialogue := gen.Generate("test", SpeakerCommander, DialogueMissionBriefing)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatDialogue(dialogue)
	}
}
