// Package replay integration tests verify deterministic replay behavior.
package replay

import (
	"os"
	"testing"

	"github.com/opd-ai/violence/pkg/rng"
)

// SimpleGameState represents a minimal deterministic game simulation.
type SimpleGameState struct {
	seed    int64
	rng     *rng.RNG
	playerX []float64 // X positions for each player
	playerY []float64 // Y positions for each player
	scores  []int     // Scores for each player
	frame   int       // Current frame number
}

// NewSimpleGameState creates a deterministic game state.
func NewSimpleGameState(seed int64, playerCount int) *SimpleGameState {
	return &SimpleGameState{
		seed:    seed,
		rng:     rng.NewRNG(uint64(seed)),
		playerX: make([]float64, playerCount),
		playerY: make([]float64, playerCount),
		scores:  make([]int, playerCount),
		frame:   0,
	}
}

// ApplyInput processes a single input frame and updates game state.
func (gs *SimpleGameState) ApplyInput(input InputFrame) {
	playerID := int(input.PlayerID)
	if playerID >= len(gs.playerX) {
		return
	}

	// Update position based on input flags (deterministic)
	if input.Flags&InputMoveUp != 0 {
		gs.playerY[playerID] -= 1.0
	}
	if input.Flags&InputMoveDown != 0 {
		gs.playerY[playerID] += 1.0
	}
	if input.Flags&InputMoveLeft != 0 {
		gs.playerX[playerID] -= 1.0
	}
	if input.Flags&InputMoveRight != 0 {
		gs.playerX[playerID] += 1.0
	}

	// Fire action uses RNG (deterministic with seed)
	if input.Flags&InputFire != 0 {
		hitChance := gs.rng.Float64()
		if hitChance > 0.5 {
			gs.scores[playerID]++
		}
	}

	gs.frame++
}

// GetState returns current player positions and scores.
func (gs *SimpleGameState) GetState() ([]float64, []float64, []int, int) {
	return gs.playerX, gs.playerY, gs.scores, gs.frame
}

// TestReplayDeterminism verifies that replay playback produces identical results.
// This is the integration test required by PLAN.md task 3.
func TestReplayDeterminism(t *testing.T) {
	seed := int64(12345)
	playerCount := uint8(2)
	replayPath := "test_replay_determinism.vrep"
	defer os.Remove(replayPath)

	// Phase 1: Record a gameplay session
	recorder := NewReplayRecorder(seed, playerCount)

	// Simulate 100 frames of gameplay with varied inputs
	inputs := []struct {
		playerID uint8
		flags    InputFlags
		mouseX   int16
		mouseY   int16
	}{
		{0, InputMoveUp | InputFire, 10, -5},
		{1, InputMoveDown | InputMoveRight, -3, 8},
		{0, InputMoveLeft, 0, 0},
		{1, InputFire, 5, 5},
		{0, InputMoveRight | InputFire, 2, 2},
		{1, InputMoveUp | InputMoveLeft, -1, -1},
		{0, InputSprint | InputMoveUp, 0, 0},
		{1, InputCrouch | InputFire, 3, 3},
		{0, InputJump | InputMoveDown, 1, 1},
		{1, InputReload, 0, 0},
	}

	for i := 0; i < 10; i++ {
		for _, inp := range inputs {
			recorder.RecordInput(inp.playerID, inp.flags, inp.mouseX, inp.mouseY)
		}
	}

	// Save the replay
	if err := recorder.Save(replayPath); err != nil {
		t.Fatalf("failed to save replay: %v", err)
	}

	// Phase 2: First playback - execute all inputs
	player1, err := LoadReplay(replayPath)
	if err != nil {
		t.Fatalf("failed to load replay: %v", err)
	}

	if player1.GetSeed() != seed {
		t.Errorf("seed mismatch: got %d, want %d", player1.GetSeed(), seed)
	}

	game1 := NewSimpleGameState(player1.GetSeed(), int(playerCount))
	for {
		frame, ok := player1.Step()
		if !ok {
			break
		}
		game1.ApplyInput(frame)
	}

	x1, y1, scores1, frames1 := game1.GetState()

	// Phase 3: Second playback - must produce identical results
	player2, err := LoadReplay(replayPath)
	if err != nil {
		t.Fatalf("failed to load replay (second playback): %v", err)
	}

	game2 := NewSimpleGameState(player2.GetSeed(), int(playerCount))
	for {
		frame, ok := player2.Step()
		if !ok {
			break
		}
		game2.ApplyInput(frame)
	}

	x2, y2, scores2, frames2 := game2.GetState()

	// Phase 4: Verify determinism - all state must be identical
	if frames1 != frames2 {
		t.Errorf("frame count mismatch: playback1=%d, playback2=%d", frames1, frames2)
	}

	for i := 0; i < len(x1); i++ {
		if x1[i] != x2[i] {
			t.Errorf("player %d X position mismatch: playback1=%.2f, playback2=%.2f", i, x1[i], x2[i])
		}
		if y1[i] != y2[i] {
			t.Errorf("player %d Y position mismatch: playback1=%.2f, playback2=%.2f", i, y1[i], y2[i])
		}
		if scores1[i] != scores2[i] {
			t.Errorf("player %d score mismatch: playback1=%d, playback2=%d", i, scores1[i], scores2[i])
		}
	}

	// Phase 5: Test seeking and reset functionality
	player3, err := LoadReplay(replayPath)
	if err != nil {
		t.Fatalf("failed to load replay (third playback): %v", err)
	}

	// Seek to midpoint and verify we can read from there
	midpoint := player3.GetDuration() / 2
	player3.Seek(midpoint)

	// Count remaining frames after seek
	framesAfterSeek := 0
	for {
		_, ok := player3.Step()
		if !ok {
			break
		}
		framesAfterSeek++
	}

	// Reset and count all frames
	player3.Reset()
	totalFrames := 0
	for {
		_, ok := player3.Step()
		if !ok {
			break
		}
		totalFrames++
	}

	// Frames after seek should be less than total
	if framesAfterSeek >= totalFrames {
		t.Errorf("seek failed: framesAfterSeek=%d should be < totalFrames=%d", framesAfterSeek, totalFrames)
	}

	if totalFrames != 100 {
		t.Errorf("expected 100 total frames, got %d", totalFrames)
	}
}

// TestReplayDeterminismWithRNG verifies RNG-dependent gameplay is deterministic.
func TestReplayDeterminismWithRNG(t *testing.T) {
	seed := int64(99999)
	playerCount := uint8(4)
	replayPath := "test_replay_rng.vrep"
	defer os.Remove(replayPath)

	recorder := NewReplayRecorder(seed, playerCount)

	// Record 200 fire inputs to generate lots of RNG calls
	for i := 0; i < 200; i++ {
		playerID := uint8(i % 4)
		recorder.RecordInput(playerID, InputFire, 0, 0)
	}

	if err := recorder.Save(replayPath); err != nil {
		t.Fatalf("failed to save replay: %v", err)
	}

	// Run playback 3 times and verify all produce identical scores
	var allScores [][]int

	for run := 0; run < 3; run++ {
		player, err := LoadReplay(replayPath)
		if err != nil {
			t.Fatalf("run %d: failed to load replay: %v", run, err)
		}

		game := NewSimpleGameState(player.GetSeed(), int(playerCount))
		for {
			frame, ok := player.Step()
			if !ok {
				break
			}
			game.ApplyInput(frame)
		}

		_, _, scores, _ := game.GetState()
		allScores = append(allScores, scores)
	}

	// All runs must have identical scores
	for run := 1; run < len(allScores); run++ {
		for player := 0; player < len(allScores[0]); player++ {
			if allScores[0][player] != allScores[run][player] {
				t.Errorf("RNG non-determinism detected: run 0 player %d score=%d, run %d player %d score=%d",
					player, allScores[0][player], run, player, allScores[run][player])
			}
		}
	}

	// Verify that RNG was actually used (scores should vary between players)
	uniqueScores := make(map[int]bool)
	for _, score := range allScores[0] {
		uniqueScores[score] = true
	}

	if len(uniqueScores) < 2 {
		t.Errorf("RNG appears unused: all players have identical scores %v", allScores[0])
	}
}
