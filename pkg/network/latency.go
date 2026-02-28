// Package network provides latency tolerance for client-server gameplay.
//
// Latency Tolerance Strategy:
// - Optimal: 0-200ms RTT - full gameplay experience
// - Degraded: 200-500ms RTT - noticeable lag but playable
// - Poor: 500-5000ms RTT - significant lag, input acceptance continues
// - Spectator: >5000ms RTT - gameplay disabled, spectator mode with reconnect prompt
//
// Client-Side Interpolation:
// Clients render 100ms (2 ticks) behind server time by buffering snapshots.
// This smooths movement and masks minor network jitter.
//
// Server-Side Input Validation:
// Server rejects commands older than 500ms to prevent stale actions.
// Combined with lag compensation, this provides fair hit detection up to moderate latency.
package network

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// InterpolationDelay is the client-side buffer to smooth rendering (100ms).
	InterpolationDelay = 100 * time.Millisecond

	// MaxStaleInput is the maximum age of inputs the server will accept (500ms).
	MaxStaleInput = 500 * time.Millisecond

	// SpectatorThreshold is the latency at which clients enter spectator mode (5000ms).
	SpectatorThreshold = 5000 * time.Millisecond
)

// InterpolationBuffer stores snapshots for client-side interpolation.
type InterpolationBuffer struct {
	mu        sync.RWMutex
	snapshots []*WorldSnapshot
	maxSize   int
	targetAge time.Duration
}

// NewInterpolationBuffer creates a new interpolation buffer.
// bufferSize is the number of snapshots to retain.
func NewInterpolationBuffer(bufferSize int) *InterpolationBuffer {
	if bufferSize < 2 {
		bufferSize = 10 // Default: 500ms at 20 tick/s = 10 snapshots
	}
	return &InterpolationBuffer{
		snapshots: make([]*WorldSnapshot, 0, bufferSize),
		maxSize:   bufferSize,
		targetAge: InterpolationDelay,
	}
}

// AddSnapshot adds a new snapshot to the buffer.
func (b *InterpolationBuffer) AddSnapshot(snapshot *WorldSnapshot) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Append snapshot
	b.snapshots = append(b.snapshots, snapshot)

	// Trim oldest snapshots if buffer is full
	if len(b.snapshots) > b.maxSize {
		b.snapshots = b.snapshots[len(b.snapshots)-b.maxSize:]
	}

	logrus.WithFields(logrus.Fields{
		"system_name": "interpolation_buffer",
		"tick":        snapshot.TickNumber,
		"buffer_size": len(b.snapshots),
	}).Debug("Snapshot added to interpolation buffer")
}

// GetInterpolatedState returns the interpolated state at the target render time.
// renderTime is the current client render time, which should be InterpolationDelay behind server time.
func (b *InterpolationBuffer) GetInterpolatedState(currentTick uint64) (*WorldSnapshot, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.snapshots) < 2 {
		if len(b.snapshots) == 1 {
			// Only one snapshot available, return it
			return b.snapshots[0], nil
		}
		return nil, fmt.Errorf("insufficient snapshots for interpolation")
	}

	// Target tick is InterpolationDelay behind current server tick
	// At 20 tick/s, 100ms = 2 ticks
	ticksDelay := uint64(InterpolationDelay.Milliseconds() / TickDuration.Milliseconds())
	var targetTick uint64
	if currentTick > ticksDelay {
		targetTick = currentTick - ticksDelay
	} else {
		targetTick = 0
	}

	// Find two snapshots to interpolate between
	var prev, next *WorldSnapshot
	for i := 0; i < len(b.snapshots)-1; i++ {
		if b.snapshots[i].TickNumber <= targetTick && b.snapshots[i+1].TickNumber >= targetTick {
			prev = b.snapshots[i]
			next = b.snapshots[i+1]
			break
		}
	}

	// If no suitable pair found, use the most recent snapshot
	if prev == nil || next == nil {
		latest := b.snapshots[len(b.snapshots)-1]
		logrus.WithFields(logrus.Fields{
			"system_name": "interpolation_buffer",
			"target_tick": targetTick,
			"latest_tick": latest.TickNumber,
		}).Debug("Using latest snapshot (no interpolation pair found)")
		return latest, nil
	}

	// For simplicity, return prev snapshot (actual interpolation would blend prev and next)
	// Full interpolation requires lerping entity positions/rotations
	logrus.WithFields(logrus.Fields{
		"system_name": "interpolation_buffer",
		"target_tick": targetTick,
		"prev_tick":   prev.TickNumber,
		"next_tick":   next.TickNumber,
	}).Debug("Interpolating between snapshots")

	return prev, nil
}

// LatencyMonitor tracks client latency and manages degradation modes.
type LatencyMonitor struct {
	mu             sync.RWMutex
	clientID       uint64
	lastPingTime   time.Time
	roundTripTime  time.Duration
	spectatorMode  bool
	reconnectReady bool
}

// NewLatencyMonitor creates a new latency monitor for a client.
func NewLatencyMonitor(clientID uint64) *LatencyMonitor {
	return &LatencyMonitor{
		clientID:     clientID,
		lastPingTime: time.Now(),
	}
}

// UpdateLatency updates the measured round-trip time.
func (m *LatencyMonitor) UpdateLatency(rtt time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.roundTripTime = rtt
	m.lastPingTime = time.Now()

	logrus.WithFields(logrus.Fields{
		"system_name": "latency_monitor",
		"client_id":   m.clientID,
		"latency_ms":  rtt.Milliseconds(),
		"spectator":   m.spectatorMode,
	}).Debug("Latency updated")

	// Check thresholds
	if rtt > SpectatorThreshold && !m.spectatorMode {
		m.spectatorMode = true
		m.reconnectReady = true
		logrus.WithFields(logrus.Fields{
			"system_name": "latency_monitor",
			"client_id":   m.clientID,
			"latency_ms":  rtt.Milliseconds(),
		}).Warn("High latency detected, entering spectator mode")
	} else if rtt <= MaxStaleInput && m.spectatorMode {
		// Latency recovered below degraded threshold
		m.spectatorMode = false
		logrus.WithFields(logrus.Fields{
			"system_name": "latency_monitor",
			"client_id":   m.clientID,
			"latency_ms":  rtt.Milliseconds(),
		}).Info("Latency recovered, exiting spectator mode")
	}
}

// GetLatency returns the current round-trip time.
func (m *LatencyMonitor) GetLatency() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.roundTripTime
}

// IsSpectatorMode returns true if client is in spectator mode due to high latency.
func (m *LatencyMonitor) IsSpectatorMode() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.spectatorMode
}

// ShouldReconnect returns true if the client should show a reconnect prompt.
func (m *LatencyMonitor) ShouldReconnect() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.reconnectReady
}

// AcknowledgeReconnectPrompt clears the reconnect prompt flag.
func (m *LatencyMonitor) AcknowledgeReconnectPrompt() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reconnectReady = false
}

// IsInputStale checks if an input command is too old to be accepted.
func IsInputStale(inputTime, serverTime time.Time) bool {
	age := serverTime.Sub(inputTime)
	return age > MaxStaleInput
}

// GetLatencyQuality returns a quality classification based on RTT.
func GetLatencyQuality(rtt time.Duration) string {
	if rtt <= 200*time.Millisecond {
		return "optimal"
	} else if rtt <= MaxStaleInput {
		return "degraded"
	} else if rtt <= SpectatorThreshold {
		return "poor"
	}
	return "spectator"
}
