package event

import (
	"sync"
	"testing"
)

// TestAlarmTrigger_Lifecycle verifies alarm activation, update, and deactivation.
func TestAlarmTrigger_Lifecycle(t *testing.T) {
	alarm := NewAlarmTrigger("alarm1", 5.0)

	if alarm.IsActive() {
		t.Error("New alarm should not be active")
	}

	alarm.Activate()
	if !alarm.IsActive() {
		t.Error("Alarm should be active after Activate()")
	}
	if alarm.GetProgress() != 0.0 {
		t.Errorf("Progress should be 0.0, got %f", alarm.GetProgress())
	}

	// Update halfway
	alarm.Update(2.5)
	if !alarm.IsActive() {
		t.Error("Alarm should still be active")
	}
	if p := alarm.GetProgress(); p < 0.49 || p > 0.51 {
		t.Errorf("Progress should be ~0.5, got %f", p)
	}

	// Update to completion
	alarm.Update(2.5)
	if alarm.IsActive() {
		t.Error("Alarm should deactivate after duration")
	}
	if alarm.Elapsed != 0 {
		t.Errorf("Elapsed should reset to 0, got %f", alarm.Elapsed)
	}
}

// TestAlarmTrigger_ZeroDuration handles edge case of zero duration.
func TestAlarmTrigger_ZeroDuration(t *testing.T) {
	alarm := NewAlarmTrigger("alarm1", 0.0)
	alarm.Activate()
	if p := alarm.GetProgress(); p != 1.0 {
		t.Errorf("Zero duration should return progress 1.0, got %f", p)
	}
}

// TestAlarmTrigger_UpdateInactive verifies no-op when updating inactive alarm.
func TestAlarmTrigger_UpdateInactive(t *testing.T) {
	alarm := NewAlarmTrigger("alarm1", 5.0)
	alarm.Update(10.0)
	if alarm.Elapsed != 0 {
		t.Errorf("Inactive alarm should not update elapsed time, got %f", alarm.Elapsed)
	}
}

// TestAlarmTrigger_Concurrent verifies thread-safe operations.
func TestAlarmTrigger_Concurrent(t *testing.T) {
	alarm := NewAlarmTrigger("alarm1", 10.0)
	alarm.Activate()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				alarm.Update(0.01)
				alarm.IsActive()
				alarm.GetProgress()
			}
		}()
	}
	wg.Wait()
}

// TestTimedLockdown_Lifecycle verifies lockdown countdown.
func TestTimedLockdown_Lifecycle(t *testing.T) {
	lockdown := NewTimedLockdown("lockdown1", 60.0)

	if lockdown.IsActive() {
		t.Error("New lockdown should not be active")
	}
	if lockdown.IsExpired() {
		t.Error("New lockdown should not be expired")
	}

	lockdown.Activate()
	if !lockdown.IsActive() {
		t.Error("Lockdown should be active after Activate()")
	}
	if r := lockdown.GetRemaining(); r != 60.0 {
		t.Errorf("Remaining should be 60.0, got %f", r)
	}

	// Update halfway
	lockdown.Update(30.0)
	if !lockdown.IsActive() {
		t.Error("Lockdown should still be active")
	}
	if r := lockdown.GetRemaining(); r != 30.0 {
		t.Errorf("Remaining should be 30.0, got %f", r)
	}

	// Update to expiration
	lockdown.Update(30.0)
	if lockdown.IsActive() {
		t.Error("Lockdown should deactivate when expired")
	}
	if r := lockdown.GetRemaining(); r != 0.0 {
		t.Errorf("Remaining should be 0.0, got %f", r)
	}
	if !lockdown.IsExpired() {
		t.Error("Lockdown should be expired")
	}
}

// TestTimedLockdown_Overtime verifies no negative remaining time.
func TestTimedLockdown_Overtime(t *testing.T) {
	lockdown := NewTimedLockdown("lockdown1", 5.0)
	lockdown.Activate()
	lockdown.Update(10.0) // Overshoot

	if r := lockdown.GetRemaining(); r != 0.0 {
		t.Errorf("Remaining should be clamped at 0.0, got %f", r)
	}
	if lockdown.IsActive() {
		t.Error("Lockdown should deactivate")
	}
}

// TestTimedLockdown_UpdateInactive verifies no-op when updating inactive lockdown.
func TestTimedLockdown_UpdateInactive(t *testing.T) {
	lockdown := NewTimedLockdown("lockdown1", 60.0)
	lockdown.Update(10.0)
	if r := lockdown.GetRemaining(); r != 0.0 {
		t.Errorf("Inactive lockdown should not decrement, got %f", r)
	}
}

// TestTimedLockdown_Concurrent verifies thread-safe operations.
func TestTimedLockdown_Concurrent(t *testing.T) {
	lockdown := NewTimedLockdown("lockdown1", 10.0)
	lockdown.Activate()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				lockdown.Update(0.001)
				lockdown.IsActive()
				lockdown.GetRemaining()
				lockdown.IsExpired()
			}
		}()
	}
	wg.Wait()
}

// TestBossArenaEvent_Trigger verifies boss arena event triggering.
func TestBossArenaEvent_Trigger(t *testing.T) {
	boss := NewBossArenaEvent("boss1", "room_5", 3, 2.5)

	if boss.IsTriggered() {
		t.Error("New boss event should not be triggered")
	}

	boss.Trigger()
	if !boss.IsTriggered() {
		t.Error("Boss event should be triggered after Trigger()")
	}

	if w := boss.GetWaveCount(); w != 3 {
		t.Errorf("Wave count should be 3, got %d", w)
	}
	if d := boss.GetSpawnDelay(); d != 2.5 {
		t.Errorf("Spawn delay should be 2.5, got %f", d)
	}
}

// TestBossArenaEvent_Concurrent verifies thread-safe operations.
func TestBossArenaEvent_Concurrent(t *testing.T) {
	boss := NewBossArenaEvent("boss1", "room_5", 3, 2.5)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				boss.Trigger()
				boss.IsTriggered()
				boss.GetWaveCount()
				boss.GetSpawnDelay()
			}
		}()
	}
	wg.Wait()

	if !boss.IsTriggered() {
		t.Error("Boss should be triggered after concurrent triggers")
	}
}

// TestSetGenre_GetGenre verifies genre management.
func TestSetGenre_GetGenre(t *testing.T) {
	tests := []struct {
		name  string
		genre string
	}{
		{"Fantasy", "fantasy"},
		{"Scifi", "scifi"},
		{"Horror", "horror"},
		{"Cyberpunk", "cyberpunk"},
		{"Postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetGenre(tt.genre)
			if g := GetGenre(); g != tt.genre {
				t.Errorf("GetGenre() = %q, want %q", g, tt.genre)
			}
		})
	}
}

// TestGenerateEventText_AllGenres verifies text generation for all genres and event types.
func TestGenerateEventText_AllGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	eventTypes := []EventType{EventAlarm, EventLockdown, EventBossArena}

	for _, genre := range genres {
		SetGenre(genre)
		for _, eventType := range eventTypes {
			t.Run(genre+"_"+eventType.String(), func(t *testing.T) {
				text := GenerateEventText(12345, eventType)
				if text == "" {
					t.Error("Generated text should not be empty")
				}
			})
		}
	}
}

// TestGenerateEventText_Deterministic verifies same seed produces same text.
func TestGenerateEventText_Deterministic(t *testing.T) {
	SetGenre("scifi")
	seed := uint64(42)

	text1 := GenerateEventText(seed, EventAlarm)
	text2 := GenerateEventText(seed, EventAlarm)

	if text1 != text2 {
		t.Errorf("Same seed should produce same text: %q != %q", text1, text2)
	}
}

// TestGenerateEventText_UnknownGenre verifies fallback for unknown genre.
func TestGenerateEventText_UnknownGenre(t *testing.T) {
	SetGenre("unknown")
	text := GenerateEventText(12345, EventAlarm)
	if text != "Alarm triggered!" {
		t.Errorf("Unknown genre should use fallback text, got %q", text)
	}
}

// TestGenerateEventAudioSting_AllTypes verifies audio sting generation.
func TestGenerateEventAudioSting_AllTypes(t *testing.T) {
	eventTypes := []EventType{EventAlarm, EventLockdown, EventBossArena}

	for _, eventType := range eventTypes {
		t.Run(eventType.String(), func(t *testing.T) {
			sting := GenerateEventAudioSting(12345, eventType)
			if sting.Seed != 12345 {
				t.Errorf("Seed should be 12345, got %d", sting.Seed)
			}
			if sting.Type == "" {
				t.Error("Type should not be empty")
			}
			if sting.Frequency <= 0 {
				t.Errorf("Frequency should be positive, got %f", sting.Frequency)
			}
			if sting.Duration <= 0 {
				t.Errorf("Duration should be positive, got %f", sting.Duration)
			}
			if sting.Pattern == "" {
				t.Error("Pattern should not be empty")
			}
		})
	}
}

// TestGenerateEventAudioSting_Deterministic verifies same seed produces same audio.
func TestGenerateEventAudioSting_Deterministic(t *testing.T) {
	SetGenre("scifi")
	seed := uint64(42)

	sting1 := GenerateEventAudioSting(seed, EventAlarm)
	sting2 := GenerateEventAudioSting(seed, EventAlarm)

	if sting1.Frequency != sting2.Frequency {
		t.Errorf("Frequency should match: %f != %f", sting1.Frequency, sting2.Frequency)
	}
	if sting1.Duration != sting2.Duration {
		t.Errorf("Duration should match: %f != %f", sting1.Duration, sting2.Duration)
	}
	if sting1.Pattern != sting2.Pattern {
		t.Errorf("Pattern should match: %q != %q", sting1.Pattern, sting2.Pattern)
	}
}

// TestGenerateEventAudioSting_AllGenres verifies audio patterns for all genres.
func TestGenerateEventAudioSting_AllGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		SetGenre(genre)
		t.Run(genre, func(t *testing.T) {
			sting := GenerateEventAudioSting(12345, EventAlarm)
			if sting.Pattern == "" {
				t.Errorf("Pattern should not be empty for genre %s", genre)
			}
		})
	}
}

// TestGenerateEventAudioSting_FrequencyRanges verifies frequency ranges per event type.
func TestGenerateEventAudioSting_FrequencyRanges(t *testing.T) {
	tests := []struct {
		eventType EventType
		minFreq   float64
		maxFreq   float64
	}{
		{EventAlarm, 440.0, 660.0},
		{EventLockdown, 220.0, 330.0},
		{EventBossArena, 110.0, 165.0},
	}

	for _, tt := range tests {
		t.Run(tt.eventType.String(), func(t *testing.T) {
			// Test multiple seeds to verify range
			for seed := uint64(0); seed < 10; seed++ {
				sting := GenerateEventAudioSting(seed, tt.eventType)
				if sting.Frequency < tt.minFreq || sting.Frequency > tt.maxFreq {
					t.Errorf("Frequency %f out of range [%f, %f]", sting.Frequency, tt.minFreq, tt.maxFreq)
				}
			}
		})
	}
}

// Helper method for EventType string representation (for testing).
func (e EventType) String() string {
	switch e {
	case EventAlarm:
		return "EventAlarm"
	case EventLockdown:
		return "EventLockdown"
	case EventBossArena:
		return "EventBossArena"
	default:
		return "Unknown"
	}
}

// BenchmarkAlarmTrigger_Update benchmarks alarm update performance.
func BenchmarkAlarmTrigger_Update(b *testing.B) {
	alarm := NewAlarmTrigger("alarm1", 100.0)
	alarm.Activate()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alarm.Update(0.016) // ~60 FPS
	}
}

// BenchmarkTimedLockdown_Update benchmarks lockdown update performance.
func BenchmarkTimedLockdown_Update(b *testing.B) {
	lockdown := NewTimedLockdown("lockdown1", 100.0)
	lockdown.Activate()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lockdown.Update(0.016) // ~60 FPS
	}
}

// BenchmarkGenerateEventText benchmarks text generation.
func BenchmarkGenerateEventText(b *testing.B) {
	SetGenre("scifi")
	seed := uint64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateEventText(seed, EventAlarm)
	}
}

// BenchmarkGenerateEventAudioSting benchmarks audio sting generation.
func BenchmarkGenerateEventAudioSting(b *testing.B) {
	SetGenre("scifi")
	seed := uint64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateEventAudioSting(seed, EventAlarm)
	}
}

// TestAlarmTrigger_ReactivationResets verifies reactivation resets state.
func TestAlarmTrigger_ReactivationResets(t *testing.T) {
	alarm := NewAlarmTrigger("alarm1", 5.0)
	alarm.Activate()
	alarm.Update(3.0)

	// Reactivate
	alarm.Activate()
	if alarm.Elapsed != 0 {
		t.Errorf("Reactivation should reset elapsed to 0, got %f", alarm.Elapsed)
	}
	if !alarm.IsActive() {
		t.Error("Alarm should be active after reactivation")
	}
}

// TestTimedLockdown_ReactivationResets verifies reactivation resets countdown.
func TestTimedLockdown_ReactivationResets(t *testing.T) {
	lockdown := NewTimedLockdown("lockdown1", 60.0)
	lockdown.Activate()
	lockdown.Update(30.0)

	// Reactivate
	lockdown.Activate()
	if r := lockdown.GetRemaining(); r != 60.0 {
		t.Errorf("Reactivation should reset remaining to 60.0, got %f", r)
	}
	if !lockdown.IsActive() {
		t.Error("Lockdown should be active after reactivation")
	}
}

// TestConcurrentGenreAccess verifies thread-safe genre operations.
func TestConcurrentGenreAccess(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				SetGenre(genres[idx%len(genres)])
				GetGenre()
				GenerateEventText(uint64(j), EventAlarm)
			}
		}(i)
	}
	wg.Wait()
}

// TestEventType_Coverage ensures all event types are tested.
func TestEventType_Coverage(t *testing.T) {
	// This test ensures we don't forget to test new EventType values
	eventTypes := []EventType{EventAlarm, EventLockdown, EventBossArena}

	for _, et := range eventTypes {
		// Verify text generation works
		text := GenerateEventText(12345, et)
		if text == "" {
			t.Errorf("EventType %v should generate non-empty text", et)
		}

		// Verify audio generation works
		audio := GenerateEventAudioSting(12345, et)
		if audio.Type == "" {
			t.Errorf("EventType %v should generate non-empty audio type", et)
		}
	}
}
