// Package replay provides deterministic game replay recording and playback.
package replay

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// MagicBytes identifies a Violence replay file.
	MagicBytes = "VREP"
	// CurrentVersion is the replay file format version.
	CurrentVersion = uint16(1)
	// HeaderSize is the size of the replay header in bytes.
	HeaderSize = 32
)

// InputFlags represents player input as a bitfield.
type InputFlags uint16

const (
	InputNone      InputFlags = 0
	InputMoveUp    InputFlags = 1 << 0
	InputMoveDown  InputFlags = 1 << 1
	InputMoveLeft  InputFlags = 1 << 2
	InputMoveRight InputFlags = 1 << 3
	InputFire      InputFlags = 1 << 4
	InputUse       InputFlags = 1 << 5
	InputReload    InputFlags = 1 << 6
	InputSprint    InputFlags = 1 << 7
	InputCrouch    InputFlags = 1 << 8
	InputJump      InputFlags = 1 << 9
)

// ReplayHeader contains metadata for a replay file.
type ReplayHeader struct {
	Magic       [4]byte  // "VREP"
	Version     uint16   // File format version
	Seed        int64    // RNG seed for deterministic replay
	Duration    uint32   // Total duration in milliseconds
	PlayerCount uint8    // Number of players
	Reserved    [13]byte // Reserved for future use
}

// InputFrame represents a single player input at a specific timestamp.
type InputFrame struct {
	Timestamp   uint32     // Milliseconds from replay start
	PlayerID    uint8      // Player identifier (0-255)
	Flags       InputFlags // Input bitfield
	MouseDeltaX int16      // Mouse X movement
	MouseDeltaY int16      // Mouse Y movement
}

// ReplayRecorder records game inputs for deterministic replay.
type ReplayRecorder struct {
	seed        int64
	inputs      []InputFrame
	startTime   time.Time
	playerCount uint8
}

// NewReplayRecorder creates a new replay recorder.
func NewReplayRecorder(seed int64, playerCount uint8) *ReplayRecorder {
	logrus.WithFields(logrus.Fields{
		"seed":         seed,
		"player_count": playerCount,
	}).Debug("replay recorder created")

	return &ReplayRecorder{
		seed:        seed,
		inputs:      make([]InputFrame, 0, 1000),
		startTime:   time.Now(),
		playerCount: playerCount,
	}
}

// RecordInput records a player input frame.
func (r *ReplayRecorder) RecordInput(playerID uint8, flags InputFlags, mouseDeltaX, mouseDeltaY int16) {
	elapsed := time.Since(r.startTime)
	timestamp := uint32(elapsed.Milliseconds())

	frame := InputFrame{
		Timestamp:   timestamp,
		PlayerID:    playerID,
		Flags:       flags,
		MouseDeltaX: mouseDeltaX,
		MouseDeltaY: mouseDeltaY,
	}

	r.inputs = append(r.inputs, frame)
}

// Save writes the replay to a file.
func (r *ReplayRecorder) Save(path string) error {
	duration := uint32(time.Since(r.startTime).Milliseconds())

	header := ReplayHeader{
		Version:     CurrentVersion,
		Seed:        r.seed,
		Duration:    duration,
		PlayerCount: r.playerCount,
	}
	copy(header.Magic[:], MagicBytes)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create replay file: %w", err)
	}
	defer file.Close()

	if err := r.writeHeader(file, header); err != nil {
		return err
	}

	if err := r.writeInputs(file); err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"path":        path,
		"seed":        r.seed,
		"duration_ms": duration,
		"input_count": len(r.inputs),
	}).Info("replay saved")

	return nil
}

// writeHeader writes the replay header to a writer.
func (r *ReplayRecorder) writeHeader(w io.Writer, header ReplayHeader) error {
	if err := binary.Write(w, binary.LittleEndian, header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	return nil
}

// writeInputs writes input frames to a writer.
func (r *ReplayRecorder) writeInputs(w io.Writer) error {
	for _, input := range r.inputs {
		if err := binary.Write(w, binary.LittleEndian, input); err != nil {
			return fmt.Errorf("failed to write input frame: %w", err)
		}
	}
	return nil
}

// InputCount returns the number of recorded input frames.
func (r *ReplayRecorder) InputCount() int {
	return len(r.inputs)
}

// Duration returns the current duration of the recording.
func (r *ReplayRecorder) Duration() time.Duration {
	return time.Since(r.startTime)
}

// ReplayPlayer plays back a recorded replay.
type ReplayPlayer struct {
	header ReplayHeader
	inputs []InputFrame
	cursor int
}

// LoadReplay loads a replay from a file.
func LoadReplay(path string) (*ReplayPlayer, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open replay file: %w", err)
	}
	defer file.Close()

	player := &ReplayPlayer{}

	if err := player.readHeader(file); err != nil {
		return nil, err
	}

	if err := player.readInputs(file); err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"path":        path,
		"seed":        player.header.Seed,
		"duration_ms": player.header.Duration,
		"input_count": len(player.inputs),
	}).Info("replay loaded")

	return player, nil
}

// readHeader reads the replay header from a reader.
func (p *ReplayPlayer) readHeader(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &p.header); err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Validate magic bytes
	if !bytes.Equal(p.header.Magic[:], []byte(MagicBytes)) {
		return fmt.Errorf("invalid magic bytes: expected %q, got %q", MagicBytes, p.header.Magic)
	}

	// Validate version
	if p.header.Version != CurrentVersion {
		return fmt.Errorf("unsupported version: %d (expected %d)", p.header.Version, CurrentVersion)
	}

	return nil
}

// readInputs reads all input frames from a reader.
func (p *ReplayPlayer) readInputs(r io.Reader) error {
	p.inputs = make([]InputFrame, 0, 1000)

	for {
		var frame InputFrame
		err := binary.Read(r, binary.LittleEndian, &frame)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read input frame: %w", err)
		}
		p.inputs = append(p.inputs, frame)
	}

	return nil
}

// Step returns the next input frame and whether more frames are available.
func (p *ReplayPlayer) Step() (InputFrame, bool) {
	if p.cursor >= len(p.inputs) {
		return InputFrame{}, false
	}

	frame := p.inputs[p.cursor]
	p.cursor++
	return frame, true
}

// Seek moves the cursor to a specific timestamp (in milliseconds).
func (p *ReplayPlayer) Seek(timestampMs uint32) {
	// Binary search for the frame closest to the timestamp
	left, right := 0, len(p.inputs)-1
	result := 0

	for left <= right {
		mid := (left + right) / 2
		if p.inputs[mid].Timestamp <= timestampMs {
			result = mid
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	p.cursor = result
}

// Reset resets the playback cursor to the beginning.
func (p *ReplayPlayer) Reset() {
	p.cursor = 0
}

// GetSeed returns the replay's RNG seed.
func (p *ReplayPlayer) GetSeed() int64 {
	return p.header.Seed
}

// GetDuration returns the total replay duration in milliseconds.
func (p *ReplayPlayer) GetDuration() uint32 {
	return p.header.Duration
}

// GetPlayerCount returns the number of players in the replay.
func (p *ReplayPlayer) GetPlayerCount() uint8 {
	return p.header.PlayerCount
}

// InputCount returns the total number of input frames.
func (p *ReplayPlayer) InputCount() int {
	return len(p.inputs)
}

// Position returns the current cursor position.
func (p *ReplayPlayer) Position() int {
	return p.cursor
}
