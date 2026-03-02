package dialogue

import (
	"math/rand"
)

// ChoiceGenerator procedurally generates dialogue choices.
type ChoiceGenerator struct{}

// NewChoiceGenerator creates a choice generator.
func NewChoiceGenerator() *ChoiceGenerator {
	return &ChoiceGenerator{}
}

// Generate creates dialogue choices based on dialogue type.
func (cg *ChoiceGenerator) Generate(genre string, dialogueType DialogueType, seed int64) []DialogueChoice {
	rng := rand.New(rand.NewSource(seed))

	// Different dialogue types have different choice structures
	switch dialogueType {
	case DialogueGreeting:
		return cg.generateGreetingChoices(genre, rng)
	case DialogueMissionBriefing:
		return cg.generateMissionChoices(genre, rng)
	case DialogueMissionComplete:
		return cg.generateCompleteChoices(genre, rng)
	case DialogueTrade:
		return cg.generateTradeChoices(genre, rng)
	case DialogueQuest:
		return cg.generateQuestChoices(genre, rng)
	default:
		return cg.generateGenericChoices(genre, rng)
	}
}

// generateGreetingChoices creates choices for greeting dialogues.
func (cg *ChoiceGenerator) generateGreetingChoices(genre string, rng *rand.Rand) []DialogueChoice {
	greetings := map[string][][]string{
		"fantasy":   {{"Greetings.", "greet_01", "friendly"}, {"What do you want?", "greet_02", "neutral"}, {"Leave me be.", "greet_03", "hostile"}},
		"scifi":     {{"Report.", "greet_01", "formal"}, {"Status?", "greet_02", "neutral"}, {"Access denied.", "greet_03", "hostile"}},
		"horror":    {{"Please help!", "greet_01", "desperate"}, {"Who are you?", "greet_02", "cautious"}, {"Stay away!", "greet_03", "afraid"}},
		"cyberpunk": {{"What's the job?", "greet_01", "business"}, {"I'm listening.", "greet_02", "neutral"}, {"Not interested.", "greet_03", "dismissive"}},
		"postapoc":  {{"Any supplies?", "greet_01", "trade"}, {"Who are you?", "greet_02", "cautious"}, {"Move along.", "greet_03", "hostile"}},
	}

	options, ok := greetings[genre]
	if !ok {
		options = greetings["fantasy"]
	}

	// Select 2-3 choices
	choiceCount := 2 + rng.Intn(2)
	if choiceCount > len(options) {
		choiceCount = len(options)
	}

	// Shuffle and select
	rng.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	choices := make([]DialogueChoice, choiceCount)
	for i := 0; i < choiceCount; i++ {
		choices[i] = DialogueChoice{
			Text:    options[i][0],
			NextID:  options[i][1],
			Outcome: options[i][2],
		}
	}

	return choices
}

// generateMissionChoices creates choices for mission briefing dialogues.
func (cg *ChoiceGenerator) generateMissionChoices(genre string, rng *rand.Rand) []DialogueChoice {
	missions := map[string][][]string{
		"fantasy":   {{"I accept.", "mission_accept", "start"}, {"Tell me more.", "mission_details", "info"}, {"I refuse.", "mission_decline", "end"}},
		"scifi":     {{"Acknowledged.", "mission_accept", "start"}, {"Parameters?", "mission_details", "info"}, {"Declined.", "mission_decline", "end"}},
		"horror":    {{"I'll help.", "mission_accept", "start"}, {"What happened?", "mission_details", "info"}, {"I can't.", "mission_decline", "end"}},
		"cyberpunk": {{"I'm in.", "mission_accept", "start"}, {"What's the pay?", "mission_details", "info"}, {"Pass.", "mission_decline", "end"}},
		"postapoc":  {{"I'll do it.", "mission_accept", "start"}, {"What's out there?", "mission_details", "info"}, {"Too dangerous.", "mission_decline", "end"}},
	}

	options, ok := missions[genre]
	if !ok {
		options = missions["fantasy"]
	}

	choices := make([]DialogueChoice, len(options))
	for i, opt := range options {
		choices[i] = DialogueChoice{
			Text:    opt[0],
			NextID:  opt[1],
			Outcome: opt[2],
		}
	}

	return choices
}

// generateCompleteChoices creates choices for mission complete dialogues.
func (cg *ChoiceGenerator) generateCompleteChoices(genre string, rng *rand.Rand) []DialogueChoice {
	completes := map[string][][]string{
		"fantasy":   {{"My pleasure.", "complete_01", "reward"}, {"What's next?", "complete_02", "continue"}},
		"scifi":     {{"Mission complete.", "complete_01", "reward"}, {"Next objective?", "complete_02", "continue"}},
		"horror":    {{"Thank God.", "complete_01", "relief"}, {"Is it over?", "complete_02", "uncertain"}},
		"cyberpunk": {{"Transfer credits.", "complete_01", "reward"}, {"Got another job?", "complete_02", "continue"}},
		"postapoc":  {{"Done.", "complete_01", "reward"}, {"Any more work?", "complete_02", "continue"}},
	}

	options, ok := completes[genre]
	if !ok {
		options = completes["fantasy"]
	}

	choices := make([]DialogueChoice, len(options))
	for i, opt := range options {
		choices[i] = DialogueChoice{
			Text:    opt[0],
			NextID:  opt[1],
			Outcome: opt[2],
		}
	}

	return choices
}

// generateTradeChoices creates choices for trade dialogues.
func (cg *ChoiceGenerator) generateTradeChoices(genre string, rng *rand.Rand) []DialogueChoice {
	trades := map[string][][]string{
		"fantasy":   {{"Show me your wares.", "trade_browse", "shop"}, {"I'm selling.", "trade_sell", "sell"}, {"Goodbye.", "trade_exit", "end"}},
		"scifi":     {{"Browse inventory.", "trade_browse", "shop"}, {"Sell items.", "trade_sell", "sell"}, {"Exit.", "trade_exit", "end"}},
		"horror":    {{"What do you have?", "trade_browse", "shop"}, {"I have something.", "trade_sell", "sell"}, {"Nevermind.", "trade_exit", "end"}},
		"cyberpunk": {{"Show me the chrome.", "trade_browse", "shop"}, {"Selling hardware.", "trade_sell", "sell"}, {"Later.", "trade_exit", "end"}},
		"postapoc":  {{"What's for trade?", "trade_browse", "shop"}, {"I have scrap.", "trade_sell", "sell"}, {"Not now.", "trade_exit", "end"}},
	}

	options, ok := trades[genre]
	if !ok {
		options = trades["fantasy"]
	}

	choices := make([]DialogueChoice, len(options))
	for i, opt := range options {
		choices[i] = DialogueChoice{
			Text:    opt[0],
			NextID:  opt[1],
			Outcome: opt[2],
		}
	}

	return choices
}

// generateQuestChoices creates choices for quest dialogues.
func (cg *ChoiceGenerator) generateQuestChoices(genre string, rng *rand.Rand) []DialogueChoice {
	quests := map[string][][]string{
		"fantasy":   {{"I'll help.", "quest_accept", "start"}, {"Why me?", "quest_why", "info"}, {"Find someone else.", "quest_decline", "end"}},
		"scifi":     {{"Accepted.", "quest_accept", "start"}, {"Clarify objective.", "quest_why", "info"}, {"Declined.", "quest_decline", "end"}},
		"horror":    {{"Tell me what to do.", "quest_accept", "start"}, {"What's happening?", "quest_why", "info"}, {"I'm leaving.", "quest_decline", "end"}},
		"cyberpunk": {{"I'm in.", "quest_accept", "start"}, {"What's the catch?", "quest_why", "info"}, {"Not my problem.", "quest_decline", "end"}},
		"postapoc":  {{"I'll do it.", "quest_accept", "start"}, {"Why is this important?", "quest_why", "info"}, {"Too risky.", "quest_decline", "end"}},
	}

	options, ok := quests[genre]
	if !ok {
		options = quests["fantasy"]
	}

	choices := make([]DialogueChoice, len(options))
	for i, opt := range options {
		choices[i] = DialogueChoice{
			Text:    opt[0],
			NextID:  opt[1],
			Outcome: opt[2],
		}
	}

	return choices
}

// generateGenericChoices creates generic dialogue choices.
func (cg *ChoiceGenerator) generateGenericChoices(genre string, rng *rand.Rand) []DialogueChoice {
	generics := [][]string{
		{"Continue.", "continue", "next"},
		{"Goodbye.", "exit", "end"},
	}

	choices := make([]DialogueChoice, len(generics))
	for i, opt := range generics {
		choices[i] = DialogueChoice{
			Text:    opt[0],
			NextID:  opt[1],
			Outcome: opt[2],
		}
	}

	return choices
}
