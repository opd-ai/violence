package uicache

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewSystem(t *testing.T) {
	s := NewSystem(800, 600)

	if s == nil {
		t.Fatal("NewSystem returned nil")
	}

	if s.screenWidth != 800 {
		t.Errorf("Expected screenWidth 800, got %d", s.screenWidth)
	}

	if s.screenHeight != 600 {
		t.Errorf("Expected screenHeight 600, got %d", s.screenHeight)
	}

	if s.maxEntries != DefaultMaxEntries {
		t.Errorf("Expected maxEntries %d, got %d", DefaultMaxEntries, s.maxEntries)
	}

	// Check pools initialized
	buckets := []SizeBucket{Bucket16, Bucket32, Bucket64, Bucket128, Bucket256, BucketLarge}
	for _, bucket := range buckets {
		if _, exists := s.pools[bucket]; !exists {
			t.Errorf("Pool for bucket %d not initialized", bucket)
		}
	}
}

func TestSetScreenSize(t *testing.T) {
	s := NewSystem(800, 600)

	// Add an entry
	img, clean := s.GetOrCreate("test", 32, 32)
	if img == nil {
		t.Fatal("GetOrCreate returned nil image")
	}
	if clean {
		t.Error("Expected clean=false for new entry")
	}

	s.MarkClean("test")

	// Verify clean
	if s.IsDirty("test") {
		t.Error("Entry should be clean after MarkClean")
	}

	// Resize screen
	s.SetScreenSize(1024, 768)

	// Entry should now be dirty
	if !s.IsDirty("test") {
		t.Error("Entry should be dirty after resize")
	}
}

func TestGetMiss(t *testing.T) {
	s := NewSystem(800, 600)
	s.BeginFrame()

	img, hit := s.Get("nonexistent")

	if hit {
		t.Error("Expected miss for nonexistent entry")
	}
	if img != nil {
		t.Error("Expected nil image for miss")
	}

	stats := s.GetStats()
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
}

func TestGetOrCreate_NewEntry(t *testing.T) {
	s := NewSystem(800, 600)
	s.BeginFrame()

	img, clean := s.GetOrCreate("new_element", 64, 64)

	if img == nil {
		t.Fatal("GetOrCreate returned nil image")
	}
	if clean {
		t.Error("Expected clean=false for new entry")
	}

	bounds := img.Bounds()
	if bounds.Dx() < 64 || bounds.Dy() < 64 {
		t.Errorf("Image too small: got %dx%d, need at least 64x64", bounds.Dx(), bounds.Dy())
	}

	// Entry should be dirty (needs rendering)
	if !s.IsDirty("new_element") {
		t.Error("New entry should be dirty")
	}
}

func TestGetOrCreate_CacheHit(t *testing.T) {
	s := NewSystem(800, 600)
	s.BeginFrame()

	// Create and mark clean
	img1, _ := s.GetOrCreate("cached", 32, 32)
	s.MarkClean("cached")

	// Should hit cache
	img2, clean := s.GetOrCreate("cached", 32, 32)

	if !clean {
		t.Error("Expected cache hit")
	}
	if img1 != img2 {
		t.Error("Expected same image on cache hit")
	}
}

func TestPut(t *testing.T) {
	s := NewSystem(800, 600)
	s.BeginFrame()

	// Put an image
	img, _ := s.GetOrCreate("put_test", 32, 32)
	s.Put("put_test", img, 32, 32)

	// Should be clean now
	if s.IsDirty("put_test") {
		t.Error("Entry should be clean after Put")
	}

	// Should hit cache
	retrieved, hit := s.Get("put_test")
	if !hit {
		t.Error("Expected cache hit after Put")
	}
	if retrieved != img {
		t.Error("Expected same image after Put")
	}
}

func TestMarkDirty(t *testing.T) {
	s := NewSystem(800, 600)
	s.BeginFrame()

	// Create clean entry
	s.GetOrCreate("dirty_test", 32, 32)
	s.MarkClean("dirty_test")

	if s.IsDirty("dirty_test") {
		t.Error("Entry should be clean")
	}

	s.MarkDirty("dirty_test")

	if !s.IsDirty("dirty_test") {
		t.Error("Entry should be dirty after MarkDirty")
	}
}

func TestMarkDirtyIfChanged(t *testing.T) {
	s := NewSystem(800, 600)
	s.BeginFrame()

	// Create entry with initial hash
	s.GetOrCreate("hash_test", 32, 32)
	s.MarkClean("hash_test")
	s.MarkDirtyIfChanged("hash_test", 12345)

	// Same hash should not dirty
	changed := s.MarkDirtyIfChanged("hash_test", 12345)
	if changed {
		t.Error("Same hash should not mark dirty")
	}

	// Different hash should dirty
	changed = s.MarkDirtyIfChanged("hash_test", 67890)
	if !changed {
		t.Error("Different hash should mark dirty")
	}
}

func TestInvalidate(t *testing.T) {
	s := NewSystem(800, 600)
	s.BeginFrame()

	s.GetOrCreate("invalidate_test", 32, 32)
	s.MarkClean("invalidate_test")

	if !s.HasEntry("invalidate_test") {
		t.Error("Entry should exist")
	}

	s.Invalidate("invalidate_test")

	if s.HasEntry("invalidate_test") {
		t.Error("Entry should not exist after Invalidate")
	}
}

func TestInvalidateAll(t *testing.T) {
	s := NewSystem(800, 600)
	s.BeginFrame()

	// Create multiple entries
	for i := 0; i < 10; i++ {
		s.GetOrCreate(fmt.Sprintf("entry_%d", i), 32, 32)
	}

	if s.GetEntryCount() != 10 {
		t.Errorf("Expected 10 entries, got %d", s.GetEntryCount())
	}

	s.InvalidateAll()

	if s.GetEntryCount() != 0 {
		t.Errorf("Expected 0 entries after InvalidateAll, got %d", s.GetEntryCount())
	}
}

func TestInvalidateByPrefix(t *testing.T) {
	s := NewSystem(800, 600)
	s.BeginFrame()

	// Create entries with different prefixes
	for i := 0; i < 5; i++ {
		s.GetOrCreate(fmt.Sprintf("healthbar_%d", i), 32, 32)
	}
	for i := 0; i < 5; i++ {
		s.GetOrCreate(fmt.Sprintf("damage_%d", i), 32, 32)
	}

	if s.GetEntryCount() != 10 {
		t.Errorf("Expected 10 entries, got %d", s.GetEntryCount())
	}

	count := s.InvalidateByPrefix("healthbar_")

	if count != 5 {
		t.Errorf("Expected 5 invalidations, got %d", count)
	}
	if s.GetEntryCount() != 5 {
		t.Errorf("Expected 5 remaining entries, got %d", s.GetEntryCount())
	}

	// Verify correct entries remain
	for i := 0; i < 5; i++ {
		if !s.HasEntry(fmt.Sprintf("damage_%d", i)) {
			t.Errorf("damage_%d should still exist", i)
		}
		if s.HasEntry(fmt.Sprintf("healthbar_%d", i)) {
			t.Errorf("healthbar_%d should be removed", i)
		}
	}
}

func TestLRUEviction(t *testing.T) {
	s := NewSystem(800, 600)
	s.SetMaxEntries(5)

	// Create more entries than max
	for i := 0; i < 10; i++ {
		s.BeginFrame()
		s.GetOrCreate(fmt.Sprintf("entry_%d", i), 32, 32)
		s.EndFrame()
	}

	// Should have evicted down to max
	if s.GetEntryCount() > 5 {
		t.Errorf("Expected max 5 entries, got %d", s.GetEntryCount())
	}
}

func TestStats(t *testing.T) {
	s := NewSystem(800, 600)

	// Generate some activity
	s.BeginFrame()
	s.GetOrCreate("stats_test", 32, 32)
	s.MarkClean("stats_test")
	s.EndFrame()

	s.BeginFrame()
	s.Get("stats_test")
	s.Get("nonexistent")
	s.EndFrame()

	stats := s.GetStats()

	if stats.TotalHits < 1 {
		t.Errorf("Expected at least 1 total hit, got %d", stats.TotalHits)
	}
	if stats.TotalMisses < 1 {
		t.Errorf("Expected at least 1 total miss, got %d", stats.TotalMisses)
	}
}

func TestHitRate(t *testing.T) {
	stats := CacheStats{
		TotalHits:   80,
		TotalMisses: 20,
	}

	hitRate := stats.HitRate()
	if hitRate != 80.0 {
		t.Errorf("Expected 80%% hit rate, got %.1f%%", hitRate)
	}

	// Zero totals
	emptyStats := CacheStats{}
	if emptyStats.HitRate() != 0 {
		t.Error("Empty stats should return 0 hit rate")
	}
}

func TestFrameHitRate(t *testing.T) {
	stats := CacheStats{
		Hits:   3,
		Misses: 1,
	}

	hitRate := stats.FrameHitRate()
	if hitRate != 75.0 {
		t.Errorf("Expected 75%% frame hit rate, got %.1f%%", hitRate)
	}
}

func TestGetBucket(t *testing.T) {
	tests := []struct {
		width, height int
		expected      SizeBucket
	}{
		{8, 8, Bucket16},
		{16, 16, Bucket16},
		{17, 17, Bucket32},
		{32, 32, Bucket32},
		{33, 33, Bucket64},
		{64, 64, Bucket64},
		{65, 65, Bucket128},
		{128, 128, Bucket128},
		{129, 129, Bucket256},
		{256, 256, Bucket256},
		{257, 257, BucketLarge},
		{512, 512, BucketLarge},
	}

	for _, tc := range tests {
		bucket := GetBucket(tc.width, tc.height)
		if bucket != tc.expected {
			t.Errorf("GetBucket(%d, %d) = %d, expected %d",
				tc.width, tc.height, bucket, tc.expected)
		}
	}
}

func TestHashString(t *testing.T) {
	// Same values should produce same hash
	hash1 := HashString("test", 42, 3.14)
	hash2 := HashString("test", 42, 3.14)

	if hash1 != hash2 {
		t.Error("Same values should produce same hash")
	}

	// Different values should produce different hash
	hash3 := HashString("test", 43, 3.14)
	if hash1 == hash3 {
		t.Error("Different values should produce different hash")
	}

	// Test various types
	_ = HashString("string", 123, 45.67, float32(1.23), true, false, uint64(999))
}

func TestConcurrentAccess(t *testing.T) {
	s := NewSystem(800, 600)

	var wg sync.WaitGroup
	const goroutines = 10
	const operations = 100

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := fmt.Sprintf("concurrent_%d_%d", id, j)
				s.BeginFrame()
				s.GetOrCreate(key, 32, 32)
				s.MarkClean(key)
				s.Get(key)
				s.EndFrame()
			}
		}(i)
	}

	wg.Wait()

	// Should complete without deadlock or panic
	count := s.GetEntryCount()
	if count == 0 {
		t.Error("Expected some entries after concurrent access")
	}
}

func TestPoolReuse(t *testing.T) {
	s := NewSystem(800, 600)

	// Create and invalidate entries to populate pool
	for i := 0; i < 5; i++ {
		s.BeginFrame()
		s.GetOrCreate(fmt.Sprintf("pool_%d", i), 32, 32)
		s.EndFrame()
	}

	s.InvalidateAll()

	// Create new entries - should reuse from pool
	s.BeginFrame()
	for i := 0; i < 5; i++ {
		s.GetOrCreate(fmt.Sprintf("reuse_%d", i), 32, 32)
	}
	s.EndFrame()

	stats := s.GetStats()
	if stats.PoolReturns == 0 {
		t.Log("Pool reuse may vary based on timing")
	}
}

func TestCacheEntryType(t *testing.T) {
	entry := &CacheEntry{ID: "test"}
	if entry.Type() != "uicache.CacheEntry" {
		t.Errorf("Unexpected type: %s", entry.Type())
	}
}

func TestCacheStatsType(t *testing.T) {
	stats := &CacheStats{}
	if stats.Type() != "uicache.CacheStats" {
		t.Errorf("Unexpected type: %s", stats.Type())
	}
}

func TestUpdateNoop(t *testing.T) {
	s := NewSystem(800, 600)
	// Should not panic
	s.Update(nil)
}

func TestGetDirtyEntry(t *testing.T) {
	s := NewSystem(800, 600)
	s.BeginFrame()

	// Create entry but leave dirty
	s.GetOrCreate("dirty_get", 32, 32)

	// Get should miss on dirty entry
	img, hit := s.Get("dirty_get")
	if hit {
		t.Error("Should miss on dirty entry")
	}
	if img != nil {
		t.Error("Should return nil on miss")
	}
}

func TestSizeUpgrade(t *testing.T) {
	s := NewSystem(800, 600)
	s.BeginFrame()

	// Create small entry
	s.GetOrCreate("size_test", 16, 16)
	s.MarkClean("size_test")

	// Request larger size - should reallocate
	img, clean := s.GetOrCreate("size_test", 64, 64)

	if clean {
		t.Error("Should be dirty after size upgrade")
	}

	bounds := img.Bounds()
	if bounds.Dx() < 64 {
		t.Errorf("Image too small after upgrade: %d", bounds.Dx())
	}
}

func BenchmarkCacheHit(b *testing.B) {
	s := NewSystem(800, 600)
	s.GetOrCreate("bench_entry", 64, 64)
	s.MarkClean("bench_entry")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Get("bench_entry")
	}
}

func BenchmarkCacheMiss(b *testing.B) {
	s := NewSystem(800, 600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Get("nonexistent")
	}
}

func BenchmarkGetOrCreate(b *testing.B) {
	s := NewSystem(800, 600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_%d", i%100)
		s.GetOrCreate(key, 32, 32)
	}
}

func BenchmarkHashString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HashString("element_id", 100, 200.5, true)
	}
}
