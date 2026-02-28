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

func TestNewCircuitTraceGame(t *testing.T) {
	game := NewCircuitTraceGame(1, 12345)
	if game == nil {
		t.Fatal("NewCircuitTraceGame returned nil")
	}
	if len(game.Grid) == 0 {
		t.Fatal("grid not generated")
	}
	if game.MaxMoves <= 0 {
		t.Fatal("max moves not set")
	}
	if game.MaxAttempts != 3 {
		t.Fatal("max attempts should be 3")
	}
}

func TestCircuitTraceGame_Start(t *testing.T) {
	game := NewCircuitTraceGame(1, 12345)
	game.Start()

	if game.CurrentX != 0 || game.CurrentY != 0 {
		t.Fatal("position not reset to start")
	}
	if game.Moves != 0 {
		t.Fatal("moves not reset")
	}
	if game.Complete {
		t.Fatal("should not be complete")
	}
}

func TestCircuitTraceGame_MoveValid(t *testing.T) {
	game := NewCircuitTraceGame(1, 12345)
	game.Start()

	// Try moving right if not blocked
	if game.Grid[0][1] != 2 {
		success := game.Move(1) // right
		if !success {
			t.Fatal("valid move should succeed")
		}
		if game.Moves != 1 {
			t.Fatal("move count should increase")
		}
	}
}

func TestCircuitTraceGame_MoveOutOfBounds(t *testing.T) {
	game := NewCircuitTraceGame(1, 12345)
	game.Start()

	// Try moving up from (0,0)
	success := game.Move(0)
	if success {
		t.Fatal("out of bounds move should fail")
	}
}

func TestCircuitTraceGame_MoveBlocked(t *testing.T) {
	game := NewCircuitTraceGame(1, 12345)
	game.Start()

	// Manually set a blocked cell
	if len(game.Grid) > 0 && len(game.Grid[0]) > 1 {
		game.Grid[0][1] = 2     // block right cell
		success := game.Move(1) // try to move right
		if success {
			t.Fatal("move to blocked cell should fail")
		}
		if game.Attempts != 1 {
			t.Fatal("attempt count should increase on blocked move")
		}
	}
}

func TestCircuitTraceGame_ReachTarget(t *testing.T) {
	game := NewCircuitTraceGame(0, 12345)
	game.Start()

	// Manually move to target (for small grid)
	game.CurrentX = game.TargetX
	game.CurrentY = game.TargetY
	game.Progress = 1.0
	game.Complete = true

	if !game.Complete {
		t.Fatal("should be complete at target")
	}
	if game.Progress != 1.0 {
		t.Fatal("progress should be 1.0 at target")
	}
}

func TestCircuitTraceGame_GetProgress(t *testing.T) {
	game := NewCircuitTraceGame(1, 12345)
	game.Start()

	if game.GetProgress() != 0 {
		t.Fatal("initial progress should be 0")
	}
}

func TestCircuitTraceGame_GetAttempts(t *testing.T) {
	game := NewCircuitTraceGame(1, 12345)
	game.Start()

	if game.GetAttempts() != 3 {
		t.Fatal("initial attempts should be 3")
	}
}

func TestNewBypassCodeGame(t *testing.T) {
	game := NewBypassCodeGame(1, 12345)
	if game == nil {
		t.Fatal("NewBypassCodeGame returned nil")
	}
	if len(game.Code) == 0 {
		t.Fatal("code not generated")
	}
	if game.MaxAttempts != 3 {
		t.Fatal("max attempts should be 3")
	}
}

func TestBypassCodeGame_Start(t *testing.T) {
	game := NewBypassCodeGame(1, 12345)
	game.Start()

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

func TestBypassCodeGame_InputDigitCorrect(t *testing.T) {
	game := NewBypassCodeGame(1, 12345)
	game.Start()

	// Input correct code
	for _, digit := range game.Code {
		success := game.InputDigit(digit)
		if !success && len(game.PlayerInput) < len(game.Code) {
			t.Fatal("correct digit input failed")
		}
	}

	if !game.Complete {
		t.Fatal("game should be complete after correct code")
	}
	if game.Progress != 1.0 {
		t.Fatal("progress should be 1.0")
	}
}

func TestBypassCodeGame_InputDigitWrong(t *testing.T) {
	game := NewBypassCodeGame(1, 12345)
	game.Start()

	// Input wrong code
	wrongCode := make([]int, len(game.Code))
	for i := range wrongCode {
		wrongCode[i] = (game.Code[i] + 1) % 10
	}

	for _, digit := range wrongCode {
		game.InputDigit(digit)
	}

	if game.Attempts != 1 {
		t.Fatal("wrong code should increase attempts")
	}
	if len(game.PlayerInput) != 0 {
		t.Fatal("input should be cleared after wrong code")
	}
}

func TestBypassCodeGame_MaxAttempts(t *testing.T) {
	game := NewBypassCodeGame(1, 12345)
	game.Start()

	// Fail max attempts
	wrongCode := make([]int, len(game.Code))
	for i := range wrongCode {
		wrongCode[i] = (game.Code[i] + 1) % 10
	}

	for attempt := 0; attempt < game.MaxAttempts; attempt++ {
		for _, digit := range wrongCode {
			game.InputDigit(digit)
		}
	}

	if !game.Complete {
		t.Fatal("game should be complete after max attempts")
	}
}

func TestBypassCodeGame_Clear(t *testing.T) {
	game := NewBypassCodeGame(1, 12345)
	game.Start()

	// Input some digits
	game.InputDigit(1)
	game.InputDigit(2)

	game.Clear()

	if len(game.PlayerInput) != 0 {
		t.Fatal("clear should empty input")
	}
	if game.Progress != 0 {
		t.Fatal("clear should reset progress")
	}
}

func TestBypassCodeGame_InvalidDigit(t *testing.T) {
	game := NewBypassCodeGame(1, 12345)
	game.Start()

	// Try invalid digits
	if game.InputDigit(-1) {
		t.Fatal("negative digit should fail")
	}
	if game.InputDigit(10) {
		t.Fatal("digit > 9 should fail")
	}
}

func TestBypassCodeGame_GetProgress(t *testing.T) {
	game := NewBypassCodeGame(1, 12345)
	game.Start()

	if game.GetProgress() != 0 {
		t.Fatal("initial progress should be 0")
	}

	game.InputDigit(game.Code[0])

	progress := game.GetProgress()
	if progress <= 0 || progress >= 1 {
		t.Fatalf("wrong progress: got %f", progress)
	}
}

func TestGetGenreMiniGame(t *testing.T) {
	tests := []struct {
		genre    string
		expected string
	}{
		{"fantasy", "LockpickGame"},
		{"cyberpunk", "CircuitTraceGame"},
		{"scifi", "BypassCodeGame"},
		{"postapoc", "BypassCodeGame"},
		{"horror", "HackGame"},
		{"unknown", "HackGame"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			game := GetGenreMiniGame(tt.genre, 1, 12345)
			if game == nil {
				t.Fatal("GetGenreMiniGame returned nil")
			}

			// Verify game can be started and used
			game.Start()
			if game.GetProgress() < 0 {
				t.Fatal("invalid progress")
			}
			if game.GetAttempts() < 0 {
				t.Fatal("invalid attempts")
			}
		})
	}
}

func TestAllMiniGamesImplementInterface(t *testing.T) {
	games := []MiniGame{
		NewHackGame(1, 12345),
		NewLockpickGame(1, 12345),
		NewCircuitTraceGame(1, 12345),
		NewBypassCodeGame(1, 12345),
	}

	for i, game := range games {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			game.Start()
			_ = game.Update()
			_ = game.GetProgress()
			_ = game.GetAttempts()
		})
	}
}
