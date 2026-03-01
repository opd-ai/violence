# Dialogue Package

The `dialogue` package provides procedurally generated NPC conversations and mission briefings for the Violence game. All dialogue content is deterministically generated from seeds with full support for all 5 game genres.

## Features

- **8 Speaker Types**: Guard, Merchant, Commander, Civilian, Technician, Mystic, Hostile, Ally
- **8 Dialogue Types**: Greeting, Mission Briefing, Mission Complete, Idle, Warning, Trade, Rumor, Quest
- **5 Genre Themes**: Fantasy, Scifi, Horror, Cyberpunk, Postapoc
- **Deterministic Generation**: Identical seeds produce identical dialogue
- **Player Choices**: Interactive response options with outcomes
- **Template System**: Genre-specific dialogue templates with procedural word substitution

## Usage

### Basic Dialogue Generation

```go
import "github.com/opd-ai/violence/pkg/dialogue"

// Create generator with seed
gen := dialogue.NewGenerator(12345)
gen.SetGenre("scifi")

// Generate NPC dialogue
d := gen.Generate("npc_guard_001", dialogue.SpeakerGuard, dialogue.DialogueGreeting)

// Display dialogue
fmt.Println(dialogue.FormatDialogue(d))
// Output:
// Security Officer Kane:
//   Halt! Identify yourself, citizen.
//   Welcome to sector 543. State your purpose.
//
// Responses:
//   1. Goodbye.
```

### Mission Briefing

```go
gen := dialogue.NewGenerator(54321)
gen.SetGenre("fantasy")

briefing := gen.GenerateMissionBriefing(
    "mission_001",
    "Eliminate all hostiles in the dungeon",
)

fmt.Println(dialogue.FormatDialogue(briefing))
// Output:
// Commander Varius:
//   Listen carefully, adventurer.
//   Eliminate all hostiles in the dungeon
//   May fortune favor you.
//
// Responses:
//   1. Understood.
//   2. I need more details.
```

### Contextual Dialogue

```go
gen := dialogue.NewGenerator(99999)
gen.SetGenre("cyberpunk")

// Generate dialogue based on game context
d := gen.GenerateContextualDialogue("trade_shop_001", dialogue.SpeakerMerchant)

// Context keywords affect dialogue type:
// - "mission" -> DialogueMissionBriefing
// - "trade" -> DialogueTrade
// - "warning" -> DialogueWarning
// - default -> DialogueIdle
```

## Genre Integration

Each genre has unique speaker names and dialogue templates:

### Fantasy
- Guards: "Ser Roland", "Captain Thorne"
- Merchants: "Merchant Aldric", "Trader Mira"
- Templates: Medieval language, magic themes, quest terminology

### Scifi
- Guards: "Security Officer Kane", "Guard Unit-7"
- Merchants: "Trader Vex", "Vendor AI-3"
- Templates: Tech jargon, sector designations, corporate language

### Horror
- Guards: "Security Guard Mike", "Officer Jenkins"
- Civilians: "Survivor Emma", "Witness David"
- Templates: Survival themes, fear elements, desperate tone

### Cyberpunk
- Guards: "Corp Security Yamamoto", "Enforcer Brick"
- Merchants: "Fixer Nyx", "Black Market Dealer"
- Templates: Net slang, corporate terms, runner culture

### Postapoc
- Guards: "Guard Jackson", "Sentry Miller"
- Merchants: "Scrap Trader Pete", "Vendor Red"
- Templates: Wasteland survival, scavenging, settlement life

## Dialogue Types

### DialogueGreeting
Initial NPC contact, context-free greeting

### DialogueMissionBriefing
Quest/mission introduction with objectives
- Includes intro, objective, conclusion
- Offers "Accept" and "Details" choices

### DialogueMissionComplete
Quest completion acknowledgment and rewards

### DialogueTrade
Shop/merchant interaction
- Offers "Show goods" and "Leave" choices

### DialogueQuest
Quest offering with details
- Provides "Tell me more", "Accept", "Decline" options

### DialogueWarning
Urgent danger notification

### DialogueIdle
Generic ambient NPC chatter

### DialogueRumor
Worldbuilding hints and flavor text

## Speaker Types

- **SpeakerGuard**: Security, law enforcement, sentries
- **SpeakerMerchant**: Traders, vendors, shopkeepers
- **SpeakerCommander**: Leaders, quest givers, authority figures
- **SpeakerCivilian**: Common NPCs, bystanders, residents
- **SpeakerTechnician**: Engineers, mechanics, specialists
- **SpeakerMystic**: Oracles, sages, psychics, AI prophets
- **SpeakerHostile**: Enemies, antagonists, threats
- **SpeakerAlly**: Companions, squad members, friendly NPCs

## Determinism

Dialogue generation is deterministic based on:
1. Generator seed (controls RNG for template selection)
2. Dialogue ID (controls speaker name and line content via hash)
3. Speaker type (determines name pool and template set)
4. Dialogue type (selects appropriate templates and choices)
5. Genre (filters templates and names)

Same inputs always produce identical output.

## Integration Points

The dialogue system integrates with:
- **Quest System** (`pkg/quest`): Mission briefings use quest objectives
- **Shop System** (`pkg/shop`): Trade dialogues trigger shop UI
- **Genre System**: `SetGenre()` for thematic consistency
- **NPC AI** (`pkg/ai`): Speaker types map to enemy/ally archetypes

## Testing

Test coverage: **92.6%**

Run tests:
```bash
go test ./pkg/dialogue/... -v -cover
```

Benchmarks available for performance testing:
```bash
go test ./pkg/dialogue/... -bench=. -benchmem
```

## Performance

Dialogue generation is lightweight:
- ~10-50μs per dialogue generation
- ~1-5μs per format operation
- Zero allocations for template lookups
- Minimal memory footprint (~5KB per generator)

## Future Enhancements

Potential additions (not yet implemented):
- Markov chain generation (similar to `pkg/lore/grammar.go`)
- Emotion/mood modifiers affecting tone
- Relationship tracking affecting dialogue options
- Voice synthesis parameters (pitch, speed, timbre)
- Conversation tree branching with memory
- Dynamic placeholder expansion from game state

## See Also

- `pkg/lore` - Collectible lore entries and environmental storytelling
- `pkg/quest` - Objective tracking and mission generation
- `pkg/shop` - Economic system and trading mechanics
- `README.md` - Procedural generation policy and design goals
