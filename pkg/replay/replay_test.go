package replay

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewReplayRecorder(t *testing.T) {
	recorder := NewReplayRecorder(12345, 4)

	if recorder.seed != 12345 {
		t.Errorf("NewReplayRecorder() seed = %v, want 12345", recorder.seed)
	}
	if recorder.playerCount != 4 {
		t.Errorf("NewReplayRecorder() playerCount = %v, want 4", recorder.playerCount)
	}
	if recorder.inputs == nil {
		t.Error("NewReplayRecorder() inputs should be initialized")
	}
}

func TestRecordInput(t *testing.T) {
	recorder := NewReplayRecorder(999, 2)

	// Record some inputs
	recorder.RecordInput(0, InputMoveUp|InputFire, 10, -5)
	recorder.RecordInput(1, InputMoveLeft, -3, 2)

	if recorder.InputCount() != 2 {
		t.Errorf("InputCount() = %v, want 2", recorder.InputCount())
	}

	// Verify first frame
	frame := recorder.inputs[0]
	if frame.PlayerID != 0 {
		t.Errorf("inputs[0].PlayerID = %v, want 0", frame.PlayerID)
	}
	if frame.Flags != (InputMoveUp | InputFire) {
		t.Errorf("inputs[0].Flags = %v, want %v", frame.Flags, InputMoveUp|InputFire)
	}
	if frame.MouseDeltaX != 10 {
		t.Errorf("inputs[0].MouseDeltaX = %v, want 10", frame.MouseDeltaX)
	}
	if frame.MouseDeltaY != -5 {
		t.Errorf("inputs[0].MouseDeltaY = %v, want -5", frame.MouseDeltaY)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	replayPath := filepath.Join(tmpDir, "test.vrep")

	// Record a replay
	recorder := NewReplayRecorder(42, 2)
	recorder.RecordInput(0, InputMoveUp, 5, -3)
	recorder.RecordInput(1, InputFire, 0, 0)
	recorder.RecordInput(0, InputMoveLeft|InputSprint, -2, 1)

	// Wait a bit to get a measurable duration
	time.Sleep(10 * time.Millisecond)

	// Save replay
	if err := recorder.Save(replayPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(replayPath); os.IsNotExist(err) {
		t.Fatalf("replay file not created: %v", err)
	}

	// Load replay
	player, err := LoadReplay(replayPath)
	if err != nil {
		t.Fatalf("LoadReplay() error = %v", err)
	}

	// Verify metadata
	if player.GetSeed() != 42 {
		t.Errorf("GetSeed() = %v, want 42", player.GetSeed())
	}
	if player.GetPlayerCount() != 2 {
		t.Errorf("GetPlayerCount() = %v, want 2", player.GetPlayerCount())
	}
	if player.InputCount() != 3 {
		t.Errorf("InputCount() = %v, want 3", player.InputCount())
	}
	if player.GetDuration() == 0 {
		t.Error("GetDuration() should be > 0")
	}

	// Verify frames
	frame1, ok1 := player.Step()
	if !ok1 {
		t.Fatal("Step() should return first frame")
	}
	if frame1.PlayerID != 0 || frame1.Flags != InputMoveUp {
		t.Errorf("frame1 = {PlayerID:%v, Flags:%v}, want {PlayerID:0, Flags:%v}",
			frame1.PlayerID, frame1.Flags, InputMoveUp)
	}

	frame2, ok2 := player.Step()
	if !ok2 {
		t.Fatal("Step() should return second frame")
	}
	if frame2.PlayerID != 1 || frame2.Flags != InputFire {
		t.Errorf("frame2 = {PlayerID:%v, Flags:%v}, want {PlayerID:1, Flags:%v}",
			frame2.PlayerID, frame2.Flags, InputFire)
	}

	frame3, ok3 := player.Step()
	if !ok3 {
		t.Fatal("Step() should return third frame")
	}
	if frame3.PlayerID != 0 || frame3.Flags != (InputMoveLeft|InputSprint) {
		t.Errorf("frame3 = {PlayerID:%v, Flags:%v}, want {PlayerID:0, Flags:%v}",
			frame3.PlayerID, frame3.Flags, InputMoveLeft|InputSprint)
	}

	// Verify end of replay
	_, ok4 := player.Step()
	if ok4 {
		t.Error("Step() should return false at end of replay")
	}
}

func TestReplayPlayerReset(t *testing.T) {
	tmpDir := t.TempDir()
	replayPath := filepath.Join(tmpDir, "reset_test.vrep")

	recorder := NewReplayRecorder(100, 1)
	recorder.RecordInput(0, InputMoveUp, 0, 0)
	recorder.RecordInput(0, InputMoveDown, 0, 0)

	if err := recorder.Save(replayPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	player, err := LoadReplay(replayPath)
	if err != nil {
		t.Fatalf("LoadReplay() error = %v", err)
	}

	// Step through
	player.Step()
	player.Step()

	if player.Position() != 2 {
		t.Errorf("Position() = %v, want 2", player.Position())
	}

	// Reset
	player.Reset()

	if player.Position() != 0 {
		t.Errorf("Position() after Reset() = %v, want 0", player.Position())
	}

	// Should be able to step through again
	frame, ok := player.Step()
	if !ok {
		t.Fatal("Step() should work after Reset()")
	}
	if frame.Flags != InputMoveUp {
		t.Errorf("frame.Flags = %v, want %v", frame.Flags, InputMoveUp)
	}
}

func TestReplayPlayerSeek(t *testing.T) {
	tmpDir := t.TempDir()
	replayPath := filepath.Join(tmpDir, "seek_test.vrep")

	recorder := NewReplayRecorder(200, 1)

	// Record inputs at specific times
	baseTime := recorder.startTime
	recorder.startTime = baseTime.Add(-100 * time.Millisecond)
	recorder.RecordInput(0, InputMoveUp, 0, 0) // ~0ms

	recorder.startTime = baseTime.Add(-50 * time.Millisecond)
	recorder.RecordInput(0, InputMoveDown, 0, 0) // ~50ms

	recorder.startTime = baseTime
	recorder.RecordInput(0, InputMoveLeft, 0, 0) // ~100ms

	if err := recorder.Save(replayPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	player, err := LoadReplay(replayPath)
	if err != nil {
		t.Fatalf("LoadReplay() error = %v", err)
	}

	// Seek to 75ms (should land on or before frame at ~100ms)
	player.Seek(75)

	frame, ok := player.Step()
	if !ok {
		t.Fatal("Step() should return frame after Seek()")
	}

	// Should get the frame at ~100ms or the one before
	if frame.Flags != InputMoveDown && frame.Flags != InputMoveLeft {
		t.Errorf("Seek(75) returned unexpected frame with flags %v", frame.Flags)
	}
}

func TestInputFlags(t *testing.T) {
	tests := []struct {
		name  string
		flags InputFlags
		has   InputFlags
		want  bool
	}{
		{
			name:  "single flag set",
			flags: InputMoveUp,
			has:   InputMoveUp,
			want:  true,
		},
		{
			name:  "multiple flags set",
			flags: InputMoveUp | InputFire,
			has:   InputFire,
			want:  true,
		},
		{
			name:  "flag not set",
			flags: InputMoveUp,
			has:   InputFire,
			want:  false,
		},
		{
			name:  "no flags",
			flags: InputNone,
			has:   InputMoveUp,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := (tt.flags & tt.has) != 0
			if got != tt.want {
				t.Errorf("flags check = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadInvalidReplay(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		content []byte
		wantErr bool
	}{
		{
			name:    "empty file",
			content: []byte{},
			wantErr: true,
		},
		{
			name:    "invalid magic bytes",
			content: []byte("WRNG" + string(make([]byte, 28))),
			wantErr: true,
		},
		{
			name:    "truncated header",
			content: []byte("VREP" + string(make([]byte, 10))),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.name+".vrep")
			if err := os.WriteFile(path, tt.content, 0o644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			_, err := LoadReplay(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadReplay() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadNonexistentReplay(t *testing.T) {
	_, err := LoadReplay("/nonexistent/path/replay.vrep")
	if err == nil {
		t.Error("LoadReplay() should error on nonexistent file")
	}
}

func TestRecorderDuration(t *testing.T) {
	recorder := NewReplayRecorder(123, 1)

	duration1 := recorder.Duration()
	if duration1 < 0 {
		t.Error("Duration() should be non-negative")
	}

	time.Sleep(10 * time.Millisecond)

	duration2 := recorder.Duration()
	if duration2 <= duration1 {
		t.Errorf("Duration() should increase over time: %v <= %v", duration2, duration1)
	}
}

func TestReplayConstants(t *testing.T) {
	if MagicBytes != "VREP" {
		t.Errorf("MagicBytes = %q, want %q", MagicBytes, "VREP")
	}
	if CurrentVersion != 1 {
		t.Errorf("CurrentVersion = %v, want 1", CurrentVersion)
	}
	if HeaderSize != 32 {
		t.Errorf("HeaderSize = %v, want 32", HeaderSize)
	}
}

func TestLargeReplay(t *testing.T) {
	tmpDir := t.TempDir()
	replayPath := filepath.Join(tmpDir, "large.vrep")

	recorder := NewReplayRecorder(9999, 8)

	// Record 1000 inputs
	for i := 0; i < 1000; i++ {
		playerID := uint8(i % 8)
		flags := InputFlags(i % 16)
		recorder.RecordInput(playerID, flags, int16(i), int16(-i))
	}

	if err := recorder.Save(replayPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	player, err := LoadReplay(replayPath)
	if err != nil {
		t.Fatalf("LoadReplay() error = %v", err)
	}

	if player.InputCount() != 1000 {
		t.Errorf("InputCount() = %v, want 1000", player.InputCount())
	}

	// Verify we can step through all frames
	count := 0
	for {
		_, ok := player.Step()
		if !ok {
			break
		}
		count++
	}

	if count != 1000 {
		t.Errorf("stepped through %v frames, want 1000", count)
	}
}

func TestSaveInvalidPath(t *testing.T) {
	recorder := NewReplayRecorder(123, 1)
	recorder.RecordInput(0, InputMoveUp, 0, 0)

	// Try to save to invalid path (directory that doesn't exist)
	err := recorder.Save("/nonexistent/directory/replay.vrep")
	if err == nil {
		t.Error("Save() should error on invalid path")
	}
}

func TestSaveToReadOnlyDirectory(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping test when running as root")
	}

	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0o555); err != nil {
		t.Fatalf("failed to create read-only dir: %v", err)
	}
	defer os.Chmod(readOnlyDir, 0o755) // cleanup

	recorder := NewReplayRecorder(456, 1)
	recorder.RecordInput(0, InputFire, 0, 0)

	err := recorder.Save(filepath.Join(readOnlyDir, "test.vrep"))
	if err == nil {
		t.Error("Save() should error when writing to read-only directory")
	}
}

// BenchmarkRecordInput measures input recording performance
func BenchmarkRecordInput(b *testing.B) {
	recorder := NewReplayRecorder(123, 4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder.RecordInput(0, InputMoveUp|InputFire, 10, -5)
	}
}

// BenchmarkSave measures replay save performance
func BenchmarkSave(b *testing.B) {
	tmpDir := b.TempDir()

	recorder := NewReplayRecorder(456, 2)
	for i := 0; i < 1000; i++ {
		recorder.RecordInput(uint8(i%2), InputFlags(i%16), int16(i), int16(-i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := filepath.Join(tmpDir, "bench.vrep")
		if err := recorder.Save(path); err != nil {
			b.Fatalf("Save() error = %v", err)
		}
	}
}

// BenchmarkLoad measures replay load performance
func BenchmarkLoad(b *testing.B) {
	tmpDir := b.TempDir()
	replayPath := filepath.Join(tmpDir, "bench.vrep")

	recorder := NewReplayRecorder(789, 4)
	for i := 0; i < 1000; i++ {
		recorder.RecordInput(uint8(i%4), InputFlags(i%16), int16(i), int16(-i))
	}
	if err := recorder.Save(replayPath); err != nil {
		b.Fatalf("Save() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := LoadReplay(replayPath); err != nil {
			b.Fatalf("LoadReplay() error = %v", err)
		}
	}
}
