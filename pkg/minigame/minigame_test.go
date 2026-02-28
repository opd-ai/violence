package minigame

import (
	"testing"
)

func TestNewHackGame(t *testing.T) {
	game := NewHackGame(1, 12345)
	if game == nil {
		t.Fatal("NewHackGame returned nil")
	}
	if len(game.Sequence) == 0 {
		t.Fatal("sequence not generated")
	}
	if game.MaxAttempts != 3 {
		t.Fatalf("wrong max attempts: got %d", game.MaxAttempts)
	}
}

func TestHackGame_Start(t *testing.T) {
	game := NewHackGame(1, 12345)
	game.Start()

	if game.Progress != 0 {
		t.Fatal("progress not reset")
	}
	if len(game.PlayerInput) != 0 {
		t.Fatal("player input not cleared")
	}
	if game.Attempts != 0 {
		t.Fatal("attempts not reset")
	}
	if game.Complete {
		t.Fatal("should not be complete")
	}
}

func TestHackGame_InputCorrect(t *testing.T) {
	game := NewHackGame(1, 12345)
	game.Start()

	// Input correct sequence
	for _, node := range game.Sequence {
		success := game.Input(node)
		if !success {
			t.Fatal("correct input failed")
		}
	}

	if !game.Complete {
		t.Fatal("game should be complete")
	}
	if game.Progress != 1.0 {
		t.Fatalf("wrong progress: got %f", game.Progress)
	}
}

func TestHackGame_InputWrong(t *testing.T) {
	game := NewHackGame(1, 12345)
	game.Start()

	// Input wrong node
	wrongNode := (game.Sequence[0] + 1) % 6
	success := game.Input(wrongNode)
	if success {
		t.Fatal("wrong input should fail")
	}
	if game.Attempts != 1 {
		t.Fatalf("wrong attempts: got %d", game.Attempts)
	}
	if game.Progress != 0 {
		t.Fatal("progress should reset on wrong input")
	}
}

func TestHackGame_MaxAttempts(t *testing.T) {
	game := NewHackGame(1, 12345)
	game.Start()

	// Fail max attempts
	wrongNode := (game.Sequence[0] + 1) % 6
	for i := 0; i < game.MaxAttempts; i++ {
		game.Input(wrongNode)
	}

	if !game.Complete {
		t.Fatal("game should be complete after max attempts")
	}
}

func TestHackGame_GetProgress(t *testing.T) {
	game := NewHackGame(1, 12345)
	game.Start()

	if game.GetProgress() != 0 {
		t.Fatal("initial progress should be 0")
	}

	// Input one correct node
	game.Input(game.Sequence[0])
	progress := game.GetProgress()
	if progress <= 0 || progress >= 1 {
		t.Fatalf("wrong progress: got %f", progress)
	}
}

func TestHackGame_GetAttempts(t *testing.T) {
	game := NewHackGame(1, 12345)
	game.Start()

	if game.GetAttempts() != 3 {
		t.Fatalf("wrong initial attempts: got %d", game.GetAttempts())
	}

	wrongNode := (game.Sequence[0] + 1) % 6
	game.Input(wrongNode)

	if game.GetAttempts() != 2 {
		t.Fatalf("wrong attempts after fail: got %d", game.GetAttempts())
	}
}

func TestNewLockpickGame(t *testing.T) {
	game := NewLockpickGame(1, 12345)
	if game == nil {
		t.Fatal("NewLockpickGame returned nil")
	}
	if game.Pins <= 0 {
		t.Fatal("pins not set")
	}
	if game.Speed <= 0 {
		t.Fatal("speed not set")
	}
	if game.Tolerance <= 0 {
		t.Fatal("tolerance not set")
	}
}

func TestLockpickGame_Start(t *testing.T) {
	game := NewLockpickGame(1, 12345)
	game.Start()

	if game.Position != 0 {
		t.Fatal("position not reset")
	}
	if game.UnlockedPins != 0 {
		t.Fatal("unlocked pins not reset")
	}
	if game.Complete {
		t.Fatal("should not be complete")
	}
}

func TestLockpickGame_Advance(t *testing.T) {
	game := NewLockpickGame(1, 12345)
	game.Start()

	initialPos := game.Position
	game.Advance()

	if game.Position <= initialPos {
		t.Fatal("position should increase")
	}
}

func TestLockpickGame_AdvanceWrap(t *testing.T) {
	game := NewLockpickGame(1, 12345)
	game.Start()
	game.Position = 0.99

	game.Advance()

	if game.Position >= 1.0 {
		t.Fatal("position should wrap at 1.0")
	}
}

func TestLockpickGame_AttemptSuccess(t *testing.T) {
	game := NewLockpickGame(1, 12345)
	game.Start()

	// Set position at target
	game.Position = game.Target

	success := game.Attempt()
	if !success {
		t.Fatal("attempt at target should succeed")
	}
	if game.UnlockedPins != 1 {
		t.Fatal("pin should be unlocked")
	}
}

func TestLockpickGame_AttemptFailure(t *testing.T) {
	game := NewLockpickGame(1, 12345)
	game.Start()

	// Set position far from target
	game.Position = (game.Target + 0.5)
	if game.Position > 1.0 {
		game.Position -= 1.0
	}

	success := game.Attempt()
	if success {
		t.Fatal("attempt far from target should fail")
	}
	if game.UnlockedPins != 0 {
		t.Fatal("no pins should be unlocked")
	}
}

func TestLockpickGame_CompleteAllPins(t *testing.T) {
	game := NewLockpickGame(1, 12345)
	game.Start()

	// Unlock all pins
	for i := 0; i < game.Pins; i++ {
		game.Position = game.Target
		game.Attempt()
	}

	if !game.Complete {
		t.Fatal("game should be complete after unlocking all pins")
	}
	if game.Progress != 1.0 {
		t.Fatalf("wrong progress: got %f", game.Progress)
	}
}

func TestLockpickGame_GetProgress(t *testing.T) {
	game := NewLockpickGame(1, 12345)
	game.Start()

	if game.GetProgress() != 0 {
		t.Fatal("initial progress should be 0")
	}

	game.Position = game.Target
	game.Attempt()

	progress := game.GetProgress()
	if progress <= 0 || progress >= 1 {
		t.Fatalf("wrong progress: got %f", progress)
	}
}

func TestSetGenre(t *testing.T) {
	// Should not panic
	SetGenre("fantasy")
	SetGenre("scifi")
	SetGenre("")
}
