package lore

import (
	"strings"
	"testing"
)

func TestNewCodex(t *testing.T) {
	codex := NewCodex()
	if codex == nil {
		t.Fatal("NewCodex returned nil")
	}
	if codex.Entries == nil {
		t.Fatal("Entries not initialized")
	}
}

func TestCodex_AddEntry(t *testing.T) {
	codex := NewCodex()
	entry := Entry{ID: "test1", Title: "Test", Text: "Content"}
	codex.AddEntry(entry)

	entries := codex.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ID != "test1" {
		t.Fatalf("wrong entry ID: got %q", entries[0].ID)
	}
}

func TestCodex_AddEntryDuplicate(t *testing.T) {
	codex := NewCodex()
	entry1 := Entry{ID: "test1", Title: "First", Text: "Content1"}
	entry2 := Entry{ID: "test1", Title: "Second", Text: "Content2"}

	codex.AddEntry(entry1)
	codex.AddEntry(entry2)

	entries := codex.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Title != "Second" {
		t.Fatal("duplicate entry should update existing")
	}
}

func TestCodex_GetEntry(t *testing.T) {
	codex := NewCodex()
	entry := Entry{ID: "test1", Title: "Test", Text: "Content"}
	codex.AddEntry(entry)

	found, ok := codex.GetEntry("test1")
	if !ok {
		t.Fatal("entry not found")
	}
	if found.ID != "test1" {
		t.Fatalf("wrong entry: got ID %q", found.ID)
	}
}

func TestCodex_GetEntryNotFound(t *testing.T) {
	codex := NewCodex()
	_, ok := codex.GetEntry("nonexistent")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestCodex_MarkFound(t *testing.T) {
	codex := NewCodex()
	entry := Entry{ID: "test1", Title: "Test", Text: "Content", Found: false}
	codex.AddEntry(entry)

	ok := codex.MarkFound("test1")
	if !ok {
		t.Fatal("MarkFound failed")
	}

	found, _ := codex.GetEntry("test1")
	if !found.Found {
		t.Fatal("entry not marked as found")
	}
}

func TestCodex_GetFoundEntries(t *testing.T) {
	codex := NewCodex()
	codex.AddEntry(Entry{ID: "1", Title: "A", Found: true})
	codex.AddEntry(Entry{ID: "2", Title: "B", Found: false})
	codex.AddEntry(Entry{ID: "3", Title: "C", Found: true})

	found := codex.GetFoundEntries()
	if len(found) != 2 {
		t.Fatalf("expected 2 found entries, got %d", len(found))
	}
}

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator(12345)
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if gen.genre != "fantasy" {
		t.Fatalf("wrong default genre: got %q", gen.genre)
	}
}

func TestGenerator_SetGenre(t *testing.T) {
	gen := NewGenerator(12345)
	gen.SetGenre("scifi")
	if gen.genre != "scifi" {
		t.Fatalf("genre not set: got %q", gen.genre)
	}
}

func TestGenerator_Generate(t *testing.T) {
	gen := NewGenerator(12345)
	entry := gen.Generate("test_entry_1")

	if entry.ID != "test_entry_1" {
		t.Fatalf("wrong ID: got %q", entry.ID)
	}
	if entry.Title == "" {
		t.Fatal("title is empty")
	}
	if entry.Text == "" {
		t.Fatal("text is empty")
	}
	if entry.Category == "" {
		t.Fatal("category is empty")
	}
	if entry.Found {
		t.Fatal("entry should not be marked found by default")
	}
}

func TestGenerator_GenerateDeterministic(t *testing.T) {
	gen1 := NewGenerator(12345)
	gen2 := NewGenerator(12345)

	entry1 := gen1.Generate("test_id")
	entry2 := gen2.Generate("test_id")

	if entry1.Title != entry2.Title {
		t.Fatalf("titles differ: %q vs %q", entry1.Title, entry2.Title)
	}
	if entry1.Text != entry2.Text {
		t.Fatalf("texts differ")
	}
	if entry1.Category != entry2.Category {
		t.Fatalf("categories differ")
	}
}

func TestGenerator_GenerateAllGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		gen := NewGenerator(12345)
		gen.SetGenre(genre)
		entry := gen.Generate("test")

		if entry.Title == "" {
			t.Fatalf("genre %s: empty title", genre)
		}
		if entry.Text == "" {
			t.Fatalf("genre %s: empty text", genre)
		}
		if entry.Category == "" {
			t.Fatalf("genre %s: empty category", genre)
		}

		// Check that text has multiple sentences
		if !strings.Contains(entry.Text, " ") {
			t.Fatalf("genre %s: text too short", genre)
		}
	}
}

func TestGenerator_UnknownGenre(t *testing.T) {
	gen := NewGenerator(12345)
	gen.SetGenre("unknown")
	entry := gen.Generate("test")

	// Should fall back to fantasy
	if entry.Title == "" || entry.Text == "" {
		t.Fatal("unknown genre should fall back to default")
	}
}

func TestSetGenre(t *testing.T) {
	// Should not panic
	SetGenre("fantasy")
	SetGenre("scifi")
	SetGenre("")
}

func TestHashString(t *testing.T) {
	// Same input produces same hash
	h1 := hashString("test")
	h2 := hashString("test")
	if h1 != h2 {
		t.Fatal("hash not deterministic")
	}

	// Different inputs produce different hashes
	h3 := hashString("different")
	if h1 == h3 {
		t.Fatal("different inputs produced same hash")
	}
}

func TestGenerateLoreText(t *testing.T) {
	gen := NewGenerator(12345)

	tests := []struct {
		genre   string
		context ContextType
	}{
		{"fantasy", ContextCombat},
		{"fantasy", ContextLab},
		{"fantasy", ContextQuarters},
		{"scifi", ContextLab},
		{"horror", ContextEscape},
		{"cyberpunk", ContextStorage},
		{"postapoc", ContextCombat},
	}

	for _, tt := range tests {
		t.Run(tt.genre+"_"+string(tt.context), func(t *testing.T) {
			gen.SetGenre(tt.genre)
			text := gen.GenerateLoreText(12345, tt.context)

			if text == "" {
				t.Fatal("generated text is empty")
			}

			// Text should have at least one complete thought
			if len(text) < 10 {
				t.Fatalf("text too short: %q", text)
			}
		})
	}
}

func TestGenerateLoreTextDeterministic(t *testing.T) {
	gen1 := NewGenerator(12345)
	gen2 := NewGenerator(12345)

	text1 := gen1.GenerateLoreText(99999, ContextCombat)
	text2 := gen2.GenerateLoreText(99999, ContextCombat)

	if text1 != text2 {
		t.Fatalf("lore text not deterministic:\n%q\nvs\n%q", text1, text2)
	}
}

func TestGenerateLoreTextUnknownContext(t *testing.T) {
	gen := NewGenerator(12345)
	text := gen.GenerateLoreText(12345, ContextType("unknown"))

	// Should fall back to general templates
	if text == "" {
		t.Fatal("should generate text for unknown context")
	}
}

func TestGenerateLoreItem(t *testing.T) {
	gen := NewGenerator(12345)

	tests := []struct {
		name     string
		itemType LoreItemType
		x, y     float64
		context  ContextType
	}{
		{"note_1", LoreItemNote, 10.5, 20.3, ContextLab},
		{"audio_1", LoreItemAudioLog, 5.0, 5.0, ContextCombat},
		{"graffiti_1", LoreItemGraffiti, 15.2, 8.9, ContextQuarters},
		{"body_1", LoreItemBodyArrangement, 3.3, 12.1, ContextEscape},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := gen.GenerateLoreItem(tt.name, tt.itemType, tt.x, tt.y, tt.context)

			if item.ID != tt.name {
				t.Fatalf("wrong ID: got %q, want %q", item.ID, tt.name)
			}
			if item.Type != tt.itemType {
				t.Fatalf("wrong type: got %v, want %v", item.Type, tt.itemType)
			}
			if item.PosX != tt.x || item.PosY != tt.y {
				t.Fatalf("wrong position: got (%.1f, %.1f), want (%.1f, %.1f)",
					item.PosX, item.PosY, tt.x, tt.y)
			}
			if item.Text == "" {
				t.Fatal("text is empty")
			}
			if item.Context != string(tt.context) {
				t.Fatalf("wrong context: got %q, want %q", item.Context, tt.context)
			}
			if item.CodexID == "" {
				t.Fatal("codex ID is empty")
			}
			if item.Activated {
				t.Fatal("item should not be activated by default")
			}
		})
	}
}

func TestGenerateLoreItemDeterministic(t *testing.T) {
	gen1 := NewGenerator(12345)
	gen2 := NewGenerator(12345)

	item1 := gen1.GenerateLoreItem("test_id", LoreItemNote, 5.0, 5.0, ContextCombat)
	item2 := gen2.GenerateLoreItem("test_id", LoreItemNote, 5.0, 5.0, ContextCombat)

	if item1.Text != item2.Text {
		t.Fatalf("lore items not deterministic:\n%q\nvs\n%q", item1.Text, item2.Text)
	}
	if item1.CodexID != item2.CodexID {
		t.Fatalf("codex IDs differ: %q vs %q", item1.CodexID, item2.CodexID)
	}
}

func TestGetLoreItemTypeName(t *testing.T) {
	tests := []struct {
		genre    string
		itemType LoreItemType
		want     string
	}{
		{"fantasy", LoreItemNote, "Scroll"},
		{"fantasy", LoreItemAudioLog, "Echo Stone"},
		{"scifi", LoreItemNote, "Data Pad"},
		{"scifi", LoreItemGraffiti, "Graffiti"},
		{"horror", LoreItemBodyArrangement, "Ritual Site"},
		{"cyberpunk", LoreItemAudioLog, "Neural Recording"},
		{"postapoc", LoreItemNote, "Journal Page"},
		{"unknown", LoreItemNote, "Scroll"}, // falls back to fantasy
	}

	for _, tt := range tests {
		t.Run(tt.genre+"_"+tt.want, func(t *testing.T) {
			got := GetLoreItemTypeName(tt.itemType, tt.genre)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetLoreItemTypeNameUnknownType(t *testing.T) {
	name := GetLoreItemTypeName(LoreItemType(999), "fantasy")
	if name != "Unknown" {
		t.Fatalf("expected 'Unknown' for invalid type, got %q", name)
	}
}

func TestLoreItemTypes(t *testing.T) {
	// Verify all item types are distinct
	types := []LoreItemType{
		LoreItemNote,
		LoreItemAudioLog,
		LoreItemGraffiti,
		LoreItemBodyArrangement,
	}

	seen := make(map[LoreItemType]bool)
	for _, itemType := range types {
		if seen[itemType] {
			t.Fatalf("duplicate item type: %v", itemType)
		}
		seen[itemType] = true
	}

	if len(seen) != 4 {
		t.Fatalf("expected 4 distinct item types, got %d", len(seen))
	}
}

func TestContextTypes(t *testing.T) {
	// Verify all context types are distinct
	contexts := []ContextType{
		ContextCombat,
		ContextLab,
		ContextQuarters,
		ContextStorage,
		ContextEscape,
		ContextGeneral,
	}

	seen := make(map[ContextType]bool)
	for _, ctx := range contexts {
		if seen[ctx] {
			t.Fatalf("duplicate context type: %v", ctx)
		}
		seen[ctx] = true
	}

	if len(seen) != 6 {
		t.Fatalf("expected 6 distinct context types, got %d", len(seen))
	}
}

func TestGenerateLoreTextAllGenresAllContexts(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	contexts := []ContextType{
		ContextCombat, ContextLab, ContextQuarters,
		ContextStorage, ContextEscape, ContextGeneral,
	}

	for _, genre := range genres {
		for _, context := range contexts {
			t.Run(genre+"_"+string(context), func(t *testing.T) {
				gen := NewGenerator(12345)
				gen.SetGenre(genre)
				text := gen.GenerateLoreText(12345, context)

				if text == "" {
					t.Fatalf("genre %s, context %s: empty text", genre, context)
				}
			})
		}
	}
}

func TestBackstoryTypes(t *testing.T) {
	types := []BackstoryType{
		BackstoryWorld,
		BackstoryFaction,
		BackstoryCharacter,
		BackstoryLocation,
		BackstoryEvent,
		BackstoryArtifact,
	}

	seen := make(map[BackstoryType]bool)
	for _, btype := range types {
		if seen[btype] {
			t.Fatalf("duplicate backstory type: %v", btype)
		}
		seen[btype] = true
	}

	if len(seen) != 6 {
		t.Fatalf("expected 6 distinct backstory types, got %d", len(seen))
	}
}

func TestGenerateBackstoryEntry(t *testing.T) {
	gen := NewGenerator(12345)

	tests := []struct {
		genre         string
		backstoryType BackstoryType
		index         int
	}{
		{"fantasy", BackstoryWorld, 0},
		{"fantasy", BackstoryFaction, 1},
		{"scifi", BackstoryCharacter, 2},
		{"horror", BackstoryLocation, 3},
		{"cyberpunk", BackstoryEvent, 4},
		{"postapoc", BackstoryArtifact, 5},
	}

	for _, tt := range tests {
		t.Run(tt.genre+"_"+string(tt.backstoryType), func(t *testing.T) {
			gen.SetGenre(tt.genre)
			entry := gen.GenerateBackstoryEntry(12345, tt.backstoryType, tt.index)

			if entry.ID == "" {
				t.Fatal("entry ID is empty")
			}
			if entry.Title == "" {
				t.Fatal("entry title is empty")
			}
			if entry.Text == "" {
				t.Fatal("entry text is empty")
			}
			if entry.Category != string(tt.backstoryType) {
				t.Fatalf("wrong category: got %q, want %q", entry.Category, tt.backstoryType)
			}
			if entry.Found {
				t.Fatal("entry should not be found by default")
			}

			// Backstory entries should have longer text (3-5 sentences)
			if len(entry.Text) < 30 {
				t.Fatalf("backstory text too short: %q", entry.Text)
			}
		})
	}
}

func TestGenerateBackstoryEntryDeterministic(t *testing.T) {
	gen1 := NewGenerator(12345)
	gen2 := NewGenerator(12345)

	entry1 := gen1.GenerateBackstoryEntry(99999, BackstoryWorld, 0)
	entry2 := gen2.GenerateBackstoryEntry(99999, BackstoryWorld, 0)

	if entry1.Title != entry2.Title {
		t.Fatalf("titles not deterministic:\n%q\nvs\n%q", entry1.Title, entry2.Title)
	}
	if entry1.Text != entry2.Text {
		t.Fatalf("text not deterministic")
	}
	if entry1.ID != entry2.ID {
		t.Fatalf("IDs differ: %q vs %q", entry1.ID, entry2.ID)
	}
}

func TestGenerateBackstoryEntryAllGenresAllTypes(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	backstoryTypes := []BackstoryType{
		BackstoryWorld, BackstoryFaction, BackstoryCharacter,
		BackstoryLocation, BackstoryEvent, BackstoryArtifact,
	}

	for _, genre := range genres {
		for _, btype := range backstoryTypes {
			t.Run(genre+"_"+string(btype), func(t *testing.T) {
				gen := NewGenerator(12345)
				gen.SetGenre(genre)
				entry := gen.GenerateBackstoryEntry(12345, btype, 0)

				if entry.Title == "" {
					t.Fatalf("genre %s, type %s: empty title", genre, btype)
				}
				if entry.Text == "" {
					t.Fatalf("genre %s, type %s: empty text", genre, btype)
				}
			})
		}
	}
}

func TestGenerateWorldBackstory(t *testing.T) {
	gen := NewGenerator(12345)

	tests := []struct {
		name       string
		genre      string
		entryCount int
	}{
		{"fantasy_12", "fantasy", 12},
		{"scifi_18", "scifi", 18},
		{"horror_6", "horror", 6},
		{"cyberpunk_24", "cyberpunk", 24},
		{"postapoc_10", "postapoc", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen.SetGenre(tt.genre)
			entries := gen.GenerateWorldBackstory(12345, tt.entryCount)

			if len(entries) != tt.entryCount {
				t.Fatalf("expected %d entries, got %d", tt.entryCount, len(entries))
			}

			// Verify all entries are valid
			for i, entry := range entries {
				if entry.ID == "" {
					t.Fatalf("entry %d: empty ID", i)
				}
				if entry.Title == "" {
					t.Fatalf("entry %d: empty title", i)
				}
				if entry.Text == "" {
					t.Fatalf("entry %d: empty text", i)
				}
				if entry.Found {
					t.Fatalf("entry %d: should not be found", i)
				}
			}

			// Verify distribution of backstory types
			typeCount := make(map[string]int)
			for _, entry := range entries {
				typeCount[entry.Category]++
			}

			// Should have variety (at least 3 different types for 6+ entries)
			if tt.entryCount >= 6 && len(typeCount) < 3 {
				t.Fatalf("insufficient variety: only %d types", len(typeCount))
			}
		})
	}
}

func TestGenerateWorldBackstoryDeterministic(t *testing.T) {
	gen1 := NewGenerator(12345)
	gen2 := NewGenerator(12345)

	entries1 := gen1.GenerateWorldBackstory(99999, 12)
	entries2 := gen2.GenerateWorldBackstory(99999, 12)

	if len(entries1) != len(entries2) {
		t.Fatal("entry counts differ")
	}

	for i := range entries1 {
		if entries1[i].Title != entries2[i].Title {
			t.Fatalf("entry %d titles differ", i)
		}
		if entries1[i].Text != entries2[i].Text {
			t.Fatalf("entry %d texts differ", i)
		}
		if entries1[i].ID != entries2[i].ID {
			t.Fatalf("entry %d IDs differ", i)
		}
	}
}

func TestGenerateWorldBackstoryUniqueness(t *testing.T) {
	gen := NewGenerator(12345)
	entries := gen.GenerateWorldBackstory(12345, 20)

	// All entries should have unique IDs
	seen := make(map[string]bool)
	for _, entry := range entries {
		if seen[entry.ID] {
			t.Fatalf("duplicate ID: %s", entry.ID)
		}
		seen[entry.ID] = true
	}
}

func TestGenerateWorldBackstoryZeroEntries(t *testing.T) {
	gen := NewGenerator(12345)
	entries := gen.GenerateWorldBackstory(12345, 0)

	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}
