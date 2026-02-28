package network

import (
	"testing"
	"time"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestInterpolationBuffer_AddSnapshot(t *testing.T) {
	tests := []struct {
		name       string
		bufferSize int
		numSnaps   int
		wantSize   int
	}{
		{
			name:       "add single snapshot",
			bufferSize: 10,
			numSnaps:   1,
			wantSize:   1,
		},
		{
			name:       "add multiple snapshots within capacity",
			bufferSize: 10,
			numSnaps:   5,
			wantSize:   5,
		},
		{
			name:       "exceed buffer capacity",
			bufferSize: 5,
			numSnaps:   10,
			wantSize:   5,
		},
		{
			name:       "minimum buffer size enforced",
			bufferSize: 0,
			numSnaps:   5,
			wantSize:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := NewInterpolationBuffer(tt.bufferSize)

			for i := 0; i < tt.numSnaps; i++ {
				snapshot := &WorldSnapshot{
					TickNumber: uint64(i),
					Entities:   make(map[engine.Entity]*EntitySnapshot),
				}
				buffer.AddSnapshot(snapshot)
			}

			buffer.mu.RLock()
			actualSize := len(buffer.snapshots)
			buffer.mu.RUnlock()

			if actualSize != tt.wantSize {
				t.Errorf("buffer size = %d, want %d", actualSize, tt.wantSize)
			}
		})
	}
}

func TestInterpolationBuffer_GetInterpolatedState(t *testing.T) {
	tests := []struct {
		name        string
		snapshots   []uint64 // tick numbers
		currentTick uint64
		wantErr     bool
		wantTick    uint64 // Expected interpolated tick (or close to it)
	}{
		{
			name:        "no snapshots",
			snapshots:   []uint64{},
			currentTick: 10,
			wantErr:     true,
		},
		{
			name:        "single snapshot",
			snapshots:   []uint64{5},
			currentTick: 10,
			wantErr:     false,
			wantTick:    5,
		},
		{
			name:        "two snapshots, interpolate between",
			snapshots:   []uint64{5, 10},
			currentTick: 12, // 100ms = 2 ticks behind, so target is 10
			wantErr:     false,
			wantTick:    5, // Should return prev snapshot
		},
		{
			name:        "multiple snapshots",
			snapshots:   []uint64{0, 5, 10, 15, 20},
			currentTick: 22, // Target is 20
			wantErr:     false,
			wantTick:    15, // Should interpolate between 15 and 20
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := NewInterpolationBuffer(10)

			for _, tick := range tt.snapshots {
				snapshot := &WorldSnapshot{
					TickNumber: tick,
					Entities:   make(map[engine.Entity]*EntitySnapshot),
				}
				buffer.AddSnapshot(snapshot)
			}

			got, err := buffer.GetInterpolatedState(tt.currentTick)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetInterpolatedState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Errorf("GetInterpolatedState() returned nil, want snapshot")
				}
				// Note: We're doing simple checks here. Full interpolation would require more complex validation
			}
		})
	}
}

func TestLatencyMonitor_UpdateLatency(t *testing.T) {
	tests := []struct {
		name          string
		clientID      uint64
		latencies     []time.Duration
		wantSpectator bool
		wantReconnect bool
	}{
		{
			name:          "optimal latency",
			clientID:      1,
			latencies:     []time.Duration{100 * time.Millisecond},
			wantSpectator: false,
			wantReconnect: false,
		},
		{
			name:          "degraded latency",
			clientID:      2,
			latencies:     []time.Duration{400 * time.Millisecond},
			wantSpectator: false,
			wantReconnect: false,
		},
		{
			name:          "high latency triggers spectator mode",
			clientID:      3,
			latencies:     []time.Duration{6000 * time.Millisecond},
			wantSpectator: true,
			wantReconnect: true,
		},
		{
			name:     "latency recovery exits spectator mode",
			clientID: 4,
			latencies: []time.Duration{
				6000 * time.Millisecond, // Enter spectator
				300 * time.Millisecond,  // Recover
			},
			wantSpectator: false,
			wantReconnect: true, // Reconnect flag remains until acknowledged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewLatencyMonitor(tt.clientID)

			for _, latency := range tt.latencies {
				monitor.UpdateLatency(latency)
			}

			if got := monitor.IsSpectatorMode(); got != tt.wantSpectator {
				t.Errorf("IsSpectatorMode() = %v, want %v", got, tt.wantSpectator)
			}

			if got := monitor.ShouldReconnect(); got != tt.wantReconnect {
				t.Errorf("ShouldReconnect() = %v, want %v", got, tt.wantReconnect)
			}
		})
	}
}

func TestLatencyMonitor_GetLatency(t *testing.T) {
	monitor := NewLatencyMonitor(1)
	expected := 250 * time.Millisecond

	monitor.UpdateLatency(expected)

	got := monitor.GetLatency()
	if got != expected {
		t.Errorf("GetLatency() = %v, want %v", got, expected)
	}
}

func TestLatencyMonitor_AcknowledgeReconnectPrompt(t *testing.T) {
	monitor := NewLatencyMonitor(1)

	// Trigger spectator mode
	monitor.UpdateLatency(6000 * time.Millisecond)

	if !monitor.ShouldReconnect() {
		t.Error("Expected reconnect prompt to be set")
	}

	monitor.AcknowledgeReconnectPrompt()

	if monitor.ShouldReconnect() {
		t.Error("Expected reconnect prompt to be cleared")
	}
}

func TestIsInputStale(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		inputTime  time.Time
		serverTime time.Time
		wantStale  bool
	}{
		{
			name:       "fresh input",
			inputTime:  now.Add(-100 * time.Millisecond),
			serverTime: now,
			wantStale:  false,
		},
		{
			name:       "input at threshold",
			inputTime:  now.Add(-MaxStaleInput),
			serverTime: now,
			wantStale:  false,
		},
		{
			name:       "stale input",
			inputTime:  now.Add(-600 * time.Millisecond),
			serverTime: now,
			wantStale:  true,
		},
		{
			name:       "very stale input",
			inputTime:  now.Add(-2 * time.Second),
			serverTime: now,
			wantStale:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsInputStale(tt.inputTime, tt.serverTime)
			if got != tt.wantStale {
				t.Errorf("IsInputStale() = %v, want %v", got, tt.wantStale)
			}
		})
	}
}

func TestGetLatencyQuality(t *testing.T) {
	tests := []struct {
		name string
		rtt  time.Duration
		want string
	}{
		{
			name: "optimal",
			rtt:  100 * time.Millisecond,
			want: "optimal",
		},
		{
			name: "optimal at threshold",
			rtt:  200 * time.Millisecond,
			want: "optimal",
		},
		{
			name: "degraded",
			rtt:  400 * time.Millisecond,
			want: "degraded",
		},
		{
			name: "degraded at threshold",
			rtt:  500 * time.Millisecond,
			want: "degraded",
		},
		{
			name: "poor",
			rtt:  1000 * time.Millisecond,
			want: "poor",
		},
		{
			name: "poor at threshold",
			rtt:  5000 * time.Millisecond,
			want: "poor",
		},
		{
			name: "spectator",
			rtt:  6000 * time.Millisecond,
			want: "spectator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetLatencyQuality(tt.rtt)
			if got != tt.want {
				t.Errorf("GetLatencyQuality() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterpolationBuffer_Concurrent(t *testing.T) {
	buffer := NewInterpolationBuffer(10)
	done := make(chan bool)

	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			snapshot := &WorldSnapshot{
				TickNumber: uint64(i),
				Entities:   make(map[engine.Entity]*EntitySnapshot),
			}
			buffer.AddSnapshot(snapshot)
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			buffer.GetInterpolatedState(uint64(i))
		}
		done <- true
	}()

	<-done
	<-done
}

func TestLatencyMonitor_Concurrent(t *testing.T) {
	monitor := NewLatencyMonitor(1)
	done := make(chan bool)

	// Concurrent updates
	go func() {
		for i := 0; i < 100; i++ {
			monitor.UpdateLatency(time.Duration(i) * time.Millisecond)
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			monitor.GetLatency()
			monitor.IsSpectatorMode()
			monitor.ShouldReconnect()
		}
		done <- true
	}()

	<-done
	<-done
}
