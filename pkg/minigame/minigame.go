// Package minigame provides interactive mini-game interfaces.
package minigame

import (
	"math/rand"
)

// MiniGame is the interface for all mini-games.
type MiniGame interface {
	Start()
	Update() bool
	GetProgress() float64
	GetAttempts() int
}

// HackGame is a hacking mini-game.
// Player must match a sequence of nodes within time limit.
type HackGame struct {
	Complete    bool
	Progress    float64
	Sequence    []int
	PlayerInput []int
	Attempts    int
	MaxAttempts int
	Difficulty  int
}

// NewHackGame creates a new hacking minigame.
func NewHackGame(difficulty int, seed int64) *HackGame {
	rng := rand.New(rand.NewSource(seed))
	sequenceLength := 3 + difficulty
	sequence := make([]int, sequenceLength)
	for i := range sequence {
		sequence[i] = rng.Intn(6) // 0-5 nodes
	}

	return &HackGame{
		Sequence:    sequence,
		PlayerInput: make([]int, 0),
		Attempts:    0,
		MaxAttempts: 3,
		Difficulty:  difficulty,
	}
}

// Start begins the hacking mini-game.
func (h *HackGame) Start() {
	h.Progress = 0
	h.PlayerInput = make([]int, 0)
	h.Attempts = 0
	h.Complete = false
}

// Input adds a player node selection.
func (h *HackGame) Input(node int) bool {
	if h.Complete {
		return false
	}

	h.PlayerInput = append(h.PlayerInput, node)

	// Check if input matches sequence so far
	idx := len(h.PlayerInput) - 1
	if idx >= len(h.Sequence) || h.PlayerInput[idx] != h.Sequence[idx] {
		// Wrong input
		h.Attempts++
		h.PlayerInput = make([]int, 0)
		h.Progress = 0
		if h.Attempts >= h.MaxAttempts {
			h.Complete = true // Failed
		}
		return false
	}

	// Correct input
	h.Progress = float64(len(h.PlayerInput)) / float64(len(h.Sequence))

	if len(h.PlayerInput) == len(h.Sequence) {
		h.Complete = true // Success
		return true
	}

	return true
}

// Update advances the hacking game; returns true when finished.
func (h *HackGame) Update() bool {
	return h.Complete
}

// GetProgress returns completion progress (0.0 to 1.0).
func (h *HackGame) GetProgress() float64 {
	return h.Progress
}

// GetAttempts returns remaining attempts.
func (h *HackGame) GetAttempts() int {
	return h.MaxAttempts - h.Attempts
}

// LockpickGame is a lockpicking mini-game.
// Player must stop a moving pin at the correct position.
type LockpickGame struct {
	Complete     bool
	Progress     float64
	Position     float64
	Target       float64
	Speed        float64
	Tolerance    float64
	Pins         int
	UnlockedPins int
	Attempts     int
	MaxAttempts  int
}

// NewLockpickGame creates a new lockpicking minigame.
func NewLockpickGame(difficulty int, seed int64) *LockpickGame {
	rng := rand.New(rand.NewSource(seed))
	pins := 2 + difficulty

	return &LockpickGame{
		Pins:         pins,
		UnlockedPins: 0,
		Speed:        0.05 + float64(difficulty)*0.02,
		Tolerance:    0.1 - float64(difficulty)*0.02,
		Target:       0.3 + rng.Float64()*0.4, // 0.3-0.7
		Position:     0,
		Attempts:     0,
		MaxAttempts:  pins * 2,
	}
}

// Start begins the lockpicking mini-game.
func (l *LockpickGame) Start() {
	l.Position = 0
	l.UnlockedPins = 0
	l.Attempts = 0
	l.Complete = false
	l.Progress = 0
}

// Advance moves the lockpick position.
func (l *LockpickGame) Advance() {
	if l.Complete {
		return
	}

	l.Position += l.Speed
	if l.Position > 1.0 {
		l.Position = 0
	}
}

// Attempt tries to unlock current pin at current position.
func (l *LockpickGame) Attempt() bool {
	if l.Complete {
		return false
	}

	l.Attempts++
	distance := l.Position - l.Target
	if distance < 0 {
		distance = -distance
	}

	if distance <= l.Tolerance {
		// Success
		l.UnlockedPins++
		l.Progress = float64(l.UnlockedPins) / float64(l.Pins)
		l.Position = 0
		if l.UnlockedPins >= l.Pins {
			l.Complete = true
			return true
		}
		return true
	}

	// Failure
	if l.Attempts >= l.MaxAttempts {
		l.Complete = true // Failed
	}
	return false
}

// Update advances the lockpicking game; returns true when finished.
func (l *LockpickGame) Update() bool {
	return l.Complete
}

// GetProgress returns completion progress (0.0 to 1.0).
func (l *LockpickGame) GetProgress() float64 {
	return l.Progress
}

// GetAttempts returns remaining attempts.
func (l *LockpickGame) GetAttempts() int {
	return l.MaxAttempts - l.Attempts
}

// SetGenre configures mini-game themes for a genre.
func SetGenre(genreID string) {}

// CircuitTraceGame is a hacking mini-game for cyberpunk genre.
// Player must trace a path through a circuit grid.
type CircuitTraceGame struct {
	Complete    bool
	Progress    float64
	Grid        [][]int // 0=empty, 1=path, 2=blocked
	CurrentX    int
	CurrentY    int
	TargetX     int
	TargetY     int
	Moves       int
	MaxMoves    int
	Attempts    int
	MaxAttempts int
	Difficulty  int
}

// NewCircuitTraceGame creates a new circuit trace hacking game.
func NewCircuitTraceGame(difficulty int, seed int64) *CircuitTraceGame {
	rng := rand.New(rand.NewSource(seed))
	gridSize := 4 + difficulty

	// Generate grid with path and blocks
	grid := make([][]int, gridSize)
	for i := range grid {
		grid[i] = make([]int, gridSize)
		for j := range grid[i] {
			if rng.Float64() < 0.2 {
				grid[i][j] = 2 // blocked
			} else {
				grid[i][j] = 0 // empty
			}
		}
	}

	// Ensure start and end are not blocked
	grid[0][0] = 0
	targetX := gridSize - 1
	targetY := gridSize - 1
	grid[targetY][targetX] = 0

	return &CircuitTraceGame{
		Grid:        grid,
		CurrentX:    0,
		CurrentY:    0,
		TargetX:     targetX,
		TargetY:     targetY,
		Moves:       0,
		MaxMoves:    gridSize * gridSize,
		Attempts:    0,
		MaxAttempts: 3,
		Difficulty:  difficulty,
	}
}

// Start begins the circuit trace game.
func (c *CircuitTraceGame) Start() {
	c.CurrentX = 0
	c.CurrentY = 0
	c.Moves = 0
	c.Attempts = 0
	c.Complete = false
	c.Progress = 0
}

// Move attempts to move in a direction (0=up, 1=right, 2=down, 3=left).
func (c *CircuitTraceGame) Move(direction int) bool {
	if c.Complete {
		return false
	}

	newX, newY := c.CurrentX, c.CurrentY
	switch direction {
	case 0: // up
		newY--
	case 1: // right
		newX++
	case 2: // down
		newY++
	case 3: // left
		newX--
	default:
		return false
	}

	// Check bounds
	if newY < 0 || newY >= len(c.Grid) || newX < 0 || newX >= len(c.Grid[0]) {
		return false
	}

	// Check if blocked
	if c.Grid[newY][newX] == 2 {
		c.Attempts++
		if c.Attempts >= c.MaxAttempts {
			c.Complete = true // Failed
		}
		return false
	}

	// Valid move
	c.CurrentX = newX
	c.CurrentY = newY
	c.Moves++

	// Calculate progress based on Manhattan distance to target
	dx := c.TargetX - c.CurrentX
	dy := c.TargetY - c.CurrentY
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	maxDist := c.TargetX + c.TargetY
	currentDist := dx + dy
	c.Progress = 1.0 - (float64(currentDist) / float64(maxDist))

	// Check if reached target
	if c.CurrentX == c.TargetX && c.CurrentY == c.TargetY {
		c.Complete = true
		c.Progress = 1.0
		return true
	}

	// Check if out of moves
	if c.Moves >= c.MaxMoves {
		c.Complete = true // Failed
		return false
	}

	return true
}

// Update advances the circuit trace game; returns true when finished.
func (c *CircuitTraceGame) Update() bool {
	return c.Complete
}

// GetProgress returns completion progress (0.0 to 1.0).
func (c *CircuitTraceGame) GetProgress() float64 {
	return c.Progress
}

// GetAttempts returns remaining attempts.
func (c *CircuitTraceGame) GetAttempts() int {
	return c.MaxAttempts - c.Attempts
}

// BypassCodeGame is a simple code entry mini-game for scifi/postapoc.
// Player must enter the correct code sequence.
type BypassCodeGame struct {
	Complete    bool
	Progress    float64
	Code        []int
	PlayerInput []int
	Attempts    int
	MaxAttempts int
	Difficulty  int
}

// NewBypassCodeGame creates a new bypass code entry game.
func NewBypassCodeGame(difficulty int, seed int64) *BypassCodeGame {
	rng := rand.New(rand.NewSource(seed))
	codeLength := 3 + difficulty
	code := make([]int, codeLength)
	for i := range code {
		code[i] = rng.Intn(10) // 0-9 digits
	}

	return &BypassCodeGame{
		Code:        code,
		PlayerInput: make([]int, 0),
		Attempts:    0,
		MaxAttempts: 3,
		Difficulty:  difficulty,
	}
}

// Start begins the bypass code game.
func (b *BypassCodeGame) Start() {
	b.PlayerInput = make([]int, 0)
	b.Attempts = 0
	b.Complete = false
	b.Progress = 0
}

// InputDigit adds a digit to the code entry.
func (b *BypassCodeGame) InputDigit(digit int) bool {
	if b.Complete || digit < 0 || digit > 9 {
		return false
	}

	b.PlayerInput = append(b.PlayerInput, digit)
	b.Progress = float64(len(b.PlayerInput)) / float64(len(b.Code))

	// Check if code is complete
	if len(b.PlayerInput) == len(b.Code) {
		// Verify code
		correct := true
		for i := range b.Code {
			if b.PlayerInput[i] != b.Code[i] {
				correct = false
				break
			}
		}

		if correct {
			b.Complete = true
			b.Progress = 1.0
			return true
		}

		// Wrong code
		b.Attempts++
		b.PlayerInput = make([]int, 0)
		b.Progress = 0

		if b.Attempts >= b.MaxAttempts {
			b.Complete = true // Failed
		}
		return false
	}

	return true
}

// Clear clears the current input.
func (b *BypassCodeGame) Clear() {
	if !b.Complete {
		b.PlayerInput = make([]int, 0)
		b.Progress = 0
	}
}

// Update advances the bypass code game; returns true when finished.
func (b *BypassCodeGame) Update() bool {
	return b.Complete
}

// GetProgress returns completion progress (0.0 to 1.0).
func (b *BypassCodeGame) GetProgress() float64 {
	return b.Progress
}

// GetAttempts returns remaining attempts.
func (b *BypassCodeGame) GetAttempts() int {
	return b.MaxAttempts - b.Attempts
}

// GetGenreMiniGame returns the appropriate mini-game type for a genre.
func GetGenreMiniGame(genre string, difficulty int, seed int64) MiniGame {
	switch genre {
	case "fantasy":
		return NewLockpickGame(difficulty, seed)
	case "cyberpunk":
		return NewCircuitTraceGame(difficulty, seed)
	case "scifi", "postapoc":
		return NewBypassCodeGame(difficulty, seed)
	case "horror":
		return NewHackGame(difficulty, seed)
	default:
		return NewHackGame(difficulty, seed)
	}
}
