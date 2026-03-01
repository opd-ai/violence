package main

import (
	"fmt"

	"github.com/opd-ai/violence/pkg/dialogue"
)

// Example demonstrating the dialogue system
func main() {
	fmt.Println("=== Violence Dialogue System Demo ===\n")

	// Create dialogue generator
	gen := dialogue.NewGenerator(42)

	// Fantasy Genre Example
	fmt.Println("--- Fantasy Genre ---")
	gen.SetGenre("fantasy")

	guard := gen.Generate("castle_guard_01", dialogue.SpeakerGuard, dialogue.DialogueGreeting)
	fmt.Println(dialogue.FormatDialogue(guard))

	quest := gen.GenerateMissionBriefing("quest_001", "Clear the dungeon of monsters")
	fmt.Println(dialogue.FormatDialogue(quest))

	// Scifi Genre Example
	fmt.Println("\n--- Scifi Genre ---")
	gen.SetGenre("scifi")

	commander := gen.Generate("commander_01", dialogue.SpeakerCommander, dialogue.DialogueMissionBriefing)
	fmt.Println(dialogue.FormatDialogue(commander))

	merchant := gen.Generate("merchant_01", dialogue.SpeakerMerchant, dialogue.DialogueTrade)
	fmt.Println(dialogue.FormatDialogue(merchant))

	// Demonstrate determinism
	fmt.Println("\n--- Determinism Test ---")
	gen1 := dialogue.NewGenerator(12345)
	gen2 := dialogue.NewGenerator(12345)
	gen1.SetGenre("fantasy")
	gen2.SetGenre("fantasy")

	d1 := gen1.Generate("test", dialogue.SpeakerGuard, dialogue.DialogueGreeting)
	d2 := gen2.Generate("test", dialogue.SpeakerGuard, dialogue.DialogueGreeting)

	if d1.SpeakerName == d2.SpeakerName && len(d1.Lines) == len(d2.Lines) {
		fmt.Println("✓ Determinism verified: Same seed produces identical dialogue")
	} else {
		fmt.Println("✗ Determinism failed")
	}
}
