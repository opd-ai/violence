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
