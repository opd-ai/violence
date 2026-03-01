package tutorial

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewTutorial(t *testing.T) {
	tutorial := NewTutorial()

	if tutorial == nil {
		t.Fatal("NewTutorial returned nil")
	}

	if tutorial.Active {
		t.Error("expected tutorial to be inactive on creation")
	}

	if tutorial.completed == nil {
		t.Error("expected completed map to be initialized")
	}

	if tutorial.savePath == "" {
		t.Error("expected savePath to be set")
	}
}

func TestShowPrompt_FirstTime(t *testing.T) {
	tutorial := &Tutorial{
		completed: make(map[PromptType]bool),
	}

	shown := tutorial.ShowPrompt(PromptMovement, "Test message")

	if !shown {
		t.Error("expected prompt to be shown first time")
	}

	if !tutorial.Active {
		t.Error("expected tutorial to be active after showing prompt")
	}

	if tutorial.Current != "Test message" {
		t.Errorf("expected message 'Test message', got '%s'", tutorial.Current)
	}

	if tutorial.Type != PromptMovement {
		t.Errorf("expected type PromptMovement, got %v", tutorial.Type)
	}
}

func TestShowPrompt_Suppression(t *testing.T) {
	tests := []struct {
		name       string
		promptType PromptType
		message    string
		completed  map[PromptType]bool
		wantShown  bool
	}{
		{
			name:       "not completed",
			promptType: PromptMovement,
			message:    "Move around",
			completed:  map[PromptType]bool{},
			wantShown:  true,
		},
		{
			name:       "already completed",
			promptType: PromptMovement,
			message:    "Move around",
			completed:  map[PromptType]bool{PromptMovement: true},
			wantShown:  false,
		},
		{
			name:       "different prompt completed",
			promptType: PromptShoot,
			message:    "Shoot",
			completed:  map[PromptType]bool{PromptMovement: true},
			wantShown:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tutorial := &Tutorial{
				completed: tt.completed,
			}

			shown := tutorial.ShowPrompt(tt.promptType, tt.message)

			if shown != tt.wantShown {
				t.Errorf("ShowPrompt() = %v, want %v", shown, tt.wantShown)
			}

			if tt.wantShown {
				if !tutorial.Active {
					t.Error("expected tutorial to be active")
				}
				if tutorial.Current != tt.message {
					t.Errorf("expected message '%s', got '%s'", tt.message, tutorial.Current)
				}
			} else {
				if tutorial.Active {
					t.Error("expected tutorial to remain inactive for suppressed prompt")
				}
			}
		})
	}
}

func TestComplete(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "tutorial_state.json")

	tutorial := &Tutorial{
		completed: make(map[PromptType]bool),
		savePath:  savePath,
		Active:    true,
		Type:      PromptMovement,
		Current:   "Test message",
	}

	tutorial.Complete()

	if tutorial.Active {
		t.Error("expected tutorial to be inactive after Complete()")
	}

	if tutorial.Current != "" {
		t.Error("expected Current to be cleared")
	}

	if tutorial.Type != "" {
		t.Error("expected Type to be cleared")
	}

	if !tutorial.completed[PromptMovement] {
		t.Error("expected PromptMovement to be marked completed")
	}

	// Verify persistence
	data, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatalf("failed to read save file: %v", err)
	}

	var state map[string]bool
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("failed to unmarshal state: %v", err)
	}

	if !state["movement"] {
		t.Error("expected movement to be persisted as completed")
	}
}

func TestDismiss(t *testing.T) {
	tutorial := &Tutorial{
		completed: make(map[PromptType]bool),
		Active:    true,
		Type:      PromptShoot,
		Current:   "Test message",
	}

	tutorial.Dismiss()

	if tutorial.Active {
		t.Error("expected tutorial to be inactive after Dismiss()")
	}

	if tutorial.Current != "" {
		t.Error("expected Current to be cleared")
	}

	if tutorial.Type != "" {
		t.Error("expected Type to be cleared")
	}

	if tutorial.completed[PromptShoot] {
		t.Error("expected PromptShoot to NOT be marked completed after Dismiss")
	}
}

func TestIsCompleted(t *testing.T) {
	tests := []struct {
		name       string
		promptType PromptType
		completed  map[PromptType]bool
		want       bool
	}{
		{
			name:       "completed",
			promptType: PromptMovement,
			completed:  map[PromptType]bool{PromptMovement: true},
			want:       true,
		},
		{
			name:       "not completed",
			promptType: PromptMovement,
			completed:  map[PromptType]bool{},
			want:       false,
		},
		{
			name:       "different prompt completed",
			promptType: PromptShoot,
			completed:  map[PromptType]bool{PromptMovement: true},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tutorial := &Tutorial{
				completed: tt.completed,
			}

			got := tutorial.IsCompleted(tt.promptType)
			if got != tt.want {
				t.Errorf("IsCompleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReset(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "tutorial_state.json")

	tutorial := &Tutorial{
		completed: map[PromptType]bool{
			PromptMovement: true,
			PromptShoot:    true,
		},
		savePath: savePath,
		Active:   true,
		Type:     PromptMovement,
		Current:  "Test",
	}

	tutorial.Reset()

	if tutorial.Active {
		t.Error("expected tutorial to be inactive after Reset()")
	}

	if tutorial.Current != "" {
		t.Error("expected Current to be cleared")
	}

	if tutorial.Type != "" {
		t.Error("expected Type to be cleared")
	}

	if len(tutorial.completed) != 0 {
		t.Errorf("expected completed map to be empty, got %d items", len(tutorial.completed))
	}

	// Verify persistence
	data, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatalf("failed to read save file: %v", err)
	}

	var state map[string]bool
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("failed to unmarshal state: %v", err)
	}

	if len(state) != 0 {
		t.Errorf("expected persisted state to be empty, got %d items", len(state))
	}
}

func TestSaveAndLoadState(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "tutorial_state.json")

	// Create and save state
	tutorial1 := &Tutorial{
		completed: map[PromptType]bool{
			PromptMovement: true,
			PromptShoot:    true,
			PromptPickup:   false,
		},
		savePath: savePath,
	}

	tutorial1.saveState()

	// Create new tutorial and load state
	tutorial2 := &Tutorial{
		completed: make(map[PromptType]bool),
		savePath:  savePath,
	}

	tutorial2.loadState()

	if !tutorial2.completed[PromptMovement] {
		t.Error("expected PromptMovement to be loaded as completed")
	}

	if !tutorial2.completed[PromptShoot] {
		t.Error("expected PromptShoot to be loaded as completed")
	}

	if tutorial2.completed[PromptPickup] {
		t.Error("expected PromptPickup to be loaded as not completed")
	}
}

func TestLoadState_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "nonexistent.json")

	tutorial := &Tutorial{
		completed: make(map[PromptType]bool),
		savePath:  savePath,
	}

	// Should not panic or error
	tutorial.loadState()

	if len(tutorial.completed) != 0 {
		t.Error("expected completed map to remain empty when file missing")
	}
}

func TestLoadState_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	os.WriteFile(savePath, []byte("not valid json {{{"), 0o644)

	tutorial := &Tutorial{
		completed: make(map[PromptType]bool),
		savePath:  savePath,
	}

	// Should not panic
	tutorial.loadState()

	if len(tutorial.completed) != 0 {
		t.Error("expected completed map to remain empty when JSON invalid")
	}
}

func TestGetMessage(t *testing.T) {
	tests := []struct {
		name       string
		promptType PromptType
		want       string
	}{
		{
			name:       "movement",
			promptType: PromptMovement,
			want:       "Use WASD to move, mouse to look around",
		},
		{
			name:       "shoot",
			promptType: PromptShoot,
			want:       "Left-click or Left Trigger to shoot",
		},
		{
			name:       "pickup",
			promptType: PromptPickup,
			want:       "Press E to pick up items",
		},
		{
			name:       "door",
			promptType: PromptDoor,
			want:       "Press E to open doors",
		},
		{
			name:       "automap",
			promptType: PromptAutomap,
			want:       "Press TAB to toggle the automap",
		},
		{
			name:       "weapon",
			promptType: PromptWeapon,
			want:       "Press 1-5 to switch weapons",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMessage(tt.promptType)
			if got != tt.want {
				t.Errorf("GetMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConcurrency(t *testing.T) {
	tutorial := &Tutorial{
		completed: make(map[PromptType]bool),
	}

	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(idx int) {
			promptTypes := []PromptType{
				PromptMovement, PromptShoot, PromptPickup,
				PromptDoor, PromptAutomap, PromptWeapon,
			}
			pt := promptTypes[idx%len(promptTypes)]
			tutorial.ShowPrompt(pt, "test")
			tutorial.Complete()
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func(idx int) {
			promptTypes := []PromptType{
				PromptMovement, PromptShoot, PromptPickup,
				PromptDoor, PromptAutomap, PromptWeapon,
			}
			pt := promptTypes[idx%len(promptTypes)]
			tutorial.IsCompleted(pt)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// No assertion needed - test passes if no race condition
}

func TestPromptType_AllTypes(t *testing.T) {
	types := []PromptType{
		PromptMovement,
		PromptShoot,
		PromptPickup,
		PromptDoor,
		PromptAutomap,
		PromptWeapon,
	}

	tutorial := &Tutorial{
		completed: make(map[PromptType]bool),
	}

	for _, pt := range types {
		shown := tutorial.ShowPrompt(pt, GetMessage(pt))
		if !shown {
			t.Errorf("expected %v to be shown first time", pt)
		}

		tutorial.Complete()

		if !tutorial.IsCompleted(pt) {
			t.Errorf("expected %v to be completed", pt)
		}

		// Should be suppressed now
		shown = tutorial.ShowPrompt(pt, GetMessage(pt))
		if shown {
			t.Errorf("expected %v to be suppressed after completion", pt)
		}
	}
}

func BenchmarkShowPrompt(b *testing.B) {
	tutorial := &Tutorial{
		completed: make(map[PromptType]bool),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tutorial.ShowPrompt(PromptMovement, "Test message")
	}
}

func BenchmarkComplete(b *testing.B) {
	tmpDir := b.TempDir()
	savePath := filepath.Join(tmpDir, "tutorial_state.json")

	tutorial := &Tutorial{
		completed: make(map[PromptType]bool),
		savePath:  savePath,
		Active:    true,
		Type:      PromptMovement,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tutorial.Complete()
		tutorial.Active = true
		tutorial.Type = PromptMovement
	}
}

func BenchmarkIsCompleted(b *testing.B) {
	tutorial := &Tutorial{
		completed: map[PromptType]bool{
			PromptMovement: true,
			PromptShoot:    true,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tutorial.IsCompleted(PromptMovement)
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name     string
		genreID  string
		expected string
	}{
		{"fantasy", "fantasy", "fantasy"},
		{"scifi", "scifi", "scifi"},
		{"horror", "horror", "horror"},
		{"cyberpunk", "cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetGenre(tt.genreID)
			if got := GetCurrentGenre(); got != tt.expected {
				t.Errorf("GetCurrentGenre() = %v, want %v", got, tt.expected)
			}
		})
	}
}
