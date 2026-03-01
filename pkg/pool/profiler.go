// Package pool provides memory profiling utilities for monitoring allocation hot paths.
package pool

import (
	"runtime"
	"sync"
	"time"
)

// Profiler tracks memory allocations and pool efficiency.
type Profiler struct {
	mu sync.RWMutex

	// Pool statistics
	entitySliceHits   uint64
	entitySliceMisses uint64
	imageHits         uint64
	imageMisses       uint64
	polygonHits       uint64
	polygonMisses     uint64

	// Memory statistics
	lastAllocBytes  uint64
	lastTotalAlloc  uint64
	lastNumGC       uint32
	lastSampleTime  time.Time
	allocRateBytesS float64
}

// NewProfiler creates a memory profiler.
func NewProfiler() *Profiler {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return &Profiler{
		lastAllocBytes: m.Alloc,
		lastTotalAlloc: m.TotalAlloc,
		lastNumGC:      m.NumGC,
		lastSampleTime: time.Now(),
	}
}

// Sample takes a memory snapshot and updates statistics.
func (p *Profiler) Sample() {
	p.mu.Lock()
	defer p.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	now := time.Now()
	elapsed := now.Sub(p.lastSampleTime).Seconds()
	if elapsed > 0 {
		allocDiff := m.TotalAlloc - p.lastTotalAlloc
		p.allocRateBytesS = float64(allocDiff) / elapsed
	}

	p.lastAllocBytes = m.Alloc
	p.lastTotalAlloc = m.TotalAlloc
	p.lastNumGC = m.NumGC
	p.lastSampleTime = now
}

// RecordEntitySliceHit records a pool hit for entity slices.
func (p *Profiler) RecordEntitySliceHit() {
	p.mu.Lock()
	p.entitySliceHits++
	p.mu.Unlock()
}

// RecordEntitySliceMiss records a pool miss for entity slices.
func (p *Profiler) RecordEntitySliceMiss() {
	p.mu.Lock()
	p.entitySliceMisses++
	p.mu.Unlock()
}

// RecordImageHit records a pool hit for images.
func (p *Profiler) RecordImageHit() {
	p.mu.Lock()
	p.imageHits++
	p.mu.Unlock()
}

// RecordImageMiss records a pool miss for images.
func (p *Profiler) RecordImageMiss() {
	p.mu.Lock()
	p.imageMisses++
	p.mu.Unlock()
}

// RecordPolygonHit records a pool hit for polygons.
func (p *Profiler) RecordPolygonHit() {
	p.mu.Lock()
	p.polygonHits++
	p.mu.Unlock()
}

// RecordPolygonMiss records a pool miss for polygons.
func (p *Profiler) RecordPolygonMiss() {
	p.mu.Lock()
	p.polygonMisses++
	p.mu.Unlock()
}

// Stats returns current profiling statistics.
type Stats struct {
	EntitySliceHitRate float64
	ImageHitRate       float64
	PolygonHitRate     float64
	AllocBytes         uint64
	AllocRateBytesS    float64
	NumGC              uint32
}

// GetStats returns current profiling statistics.
func (p *Profiler) GetStats() Stats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	entityTotal := p.entitySliceHits + p.entitySliceMisses
	imageTotal := p.imageHits + p.imageMisses
	polygonTotal := p.polygonHits + p.polygonMisses

	entityHitRate := 0.0
	if entityTotal > 0 {
		entityHitRate = float64(p.entitySliceHits) / float64(entityTotal)
	}

	imageHitRate := 0.0
	if imageTotal > 0 {
		imageHitRate = float64(p.imageHits) / float64(imageTotal)
	}

	polygonHitRate := 0.0
	if polygonTotal > 0 {
		polygonHitRate = float64(p.polygonHits) / float64(polygonTotal)
	}

	return Stats{
		EntitySliceHitRate: entityHitRate,
		ImageHitRate:       imageHitRate,
		PolygonHitRate:     polygonHitRate,
		AllocBytes:         p.lastAllocBytes,
		AllocRateBytesS:    p.allocRateBytesS,
		NumGC:              p.lastNumGC,
	}
}

// Reset clears all profiling counters.
func (p *Profiler) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.entitySliceHits = 0
	p.entitySliceMisses = 0
	p.imageHits = 0
	p.imageMisses = 0
	p.polygonHits = 0
	p.polygonMisses = 0
}

// GlobalProfiler provides singleton access to profiling.
var GlobalProfiler = NewProfiler()
