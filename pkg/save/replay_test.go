package save

import (
	"os"
	"testing"
	"time"

	"github.com/opd-ai/violence/pkg/replay"
)

// TestGetReplayPath verifies replay path generation for different slots.
func TestGetReplayPath(t *testing.T) {
	testCases := []struct {
		slot    int
		wantErr bool
		wantExt string
	}{
		{0, false, ".vrep"},
		{1, false, ".vrep"},
		{9, false, ".vrep"},
		{-1, true, ""},
		{10, true, ""},
		{100, true, ""},
	}

	for _, tc := range testCases {
		path, err := GetReplayPath(tc.slot)

		if tc.wantErr {
			if err == nil {
				t.Errorf("slot %d: expected error, got nil", tc.slot)
			}
			if err != ErrInvalidSlot {
				t.Errorf("slot %d: expected ErrInvalidSlot, got %v", tc.slot, err)
			}
			continue
		}

		if err != nil {
			t.Errorf("slot %d: unexpected error: %v", tc.slot, err)
			continue
		}

		if len(path) == 0 {
			t.Errorf("slot %d: path should not be empty", tc.slot)
			continue
		}

		// Verify extension
		if len(path) < len(tc.wantExt) || path[len(path)-len(tc.wantExt):] != tc.wantExt {
			t.Errorf("slot %d: path should end with %s, got %s", tc.slot, tc.wantExt, path)
		}
	}
}

// TestReplayIntegration_SaveAndLoad verifies replay can be saved and file is created.
func TestReplayIntegration_SaveAndLoad(t *testing.T) {
	recorder := replay.NewReplayRecorder(12345, 1)

	// Record some inputs
	recorder.RecordInput(0, replay.InputMoveUp, 0, 0)
	time.Sleep(10 * time.Millisecond)
	recorder.RecordInput(0, replay.InputFire, 10, 5)
	time.Sleep(10 * time.Millisecond)
	recorder.RecordInput(0, replay.InputMoveDown, -5, -3)

	if recorder.InputCount() != 3 {
		t.Errorf("expected 3 inputs, got %d", recorder.InputCount())
	}

	// Get replay path for test slot
	slot := 5 // Use a non-autosave slot for testing
	replayPath, err := GetReplayPath(slot)
	if err != nil {
		t.Fatalf("failed to get replay path: %v", err)
	}

	// Save replay
	if err := recorder.Save(replayPath); err != nil {
		t.Fatalf("failed to save replay: %v", err)
	}

	// Verify file exists and has content
	stat, err := os.Stat(replayPath)
	if os.IsNotExist(err) {
		t.Fatalf("replay file should exist at %s", replayPath)
	}
	if stat.Size() == 0 {
		t.Error("replay file should not be empty")
	}

	// Verify we can load it back (basic validation)
	player, err := replay.LoadReplay(replayPath)
	if err != nil {
		t.Fatalf("failed to load replay: %v", err)
	}

	if player.GetSeed() != 12345 {
		t.Errorf("expected seed 12345, got %d", player.GetSeed())
	}
	if player.GetPlayerCount() != 1 {
		t.Errorf("expected 1 player, got %d", player.GetPlayerCount())
	}

	// Clean up
	os.Remove(replayPath)
}
