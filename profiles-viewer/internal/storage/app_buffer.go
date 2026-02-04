package storage

import (
	"sync"
	"time"

	"github.com/odigos-io/odigos/profiles-viewer/internal/models"
)

// AppBuffer manages a circular buffer of time buckets for a single app.
type AppBuffer struct {
	mu         sync.RWMutex
	appID      models.AppID
	buckets    []*models.TimeBucket
	memoryUsed int64
	maxBuckets int
}

// NewAppBuffer creates a new app buffer.
func NewAppBuffer(appID models.AppID) *AppBuffer {
	return &AppBuffer{
		appID:      appID,
		buckets:    make([]*models.TimeBucket, 0, MaxBucketsPerApp),
		memoryUsed: 0,
		maxBuckets: MaxBucketsPerApp,
	}
}

// AppID returns the app identifier.
func (b *AppBuffer) AppID() models.AppID {
	return b.appID
}

// MemoryUsed returns the current memory usage.
func (b *AppBuffer) MemoryUsed() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.memoryUsed
}

// BucketCount returns the number of buckets.
func (b *AppBuffer) BucketCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.buckets)
}

// GetOrCreateBucket gets the bucket for the given timestamp, creating it if necessary.
func (b *AppBuffer) GetOrCreateBucket(ts time.Time) *models.TimeBucket {
	b.mu.Lock()
	defer b.mu.Unlock()

	bucketStart := ts.Truncate(BucketDurationSeconds * time.Second)
	bucketEnd := bucketStart.Add(BucketDurationSeconds * time.Second)

	// Search for existing bucket
	for _, bucket := range b.buckets {
		if bucket.StartTime.Equal(bucketStart) {
			return bucket
		}
	}

	// Create new bucket
	bucket := models.NewTimeBucket(bucketStart, bucketEnd)
	b.buckets = append(b.buckets, bucket)

	// Sort buckets by start time
	b.sortBuckets()

	// Enforce max buckets
	for len(b.buckets) > b.maxBuckets {
		b.evictOldest()
	}

	return bucket
}

// sortBuckets sorts buckets by start time (oldest first).
func (b *AppBuffer) sortBuckets() {
	for i := len(b.buckets) - 1; i > 0; i-- {
		if b.buckets[i].StartTime.Before(b.buckets[i-1].StartTime) {
			b.buckets[i], b.buckets[i-1] = b.buckets[i-1], b.buckets[i]
		} else {
			break
		}
	}
}

// evictOldest removes the oldest bucket and returns its memory size.
func (b *AppBuffer) evictOldest() int64 {
	if len(b.buckets) == 0 {
		return 0
	}

	evicted := b.buckets[0]
	b.buckets = b.buckets[1:]
	b.memoryUsed -= evicted.MemoryBytes
	return evicted.MemoryBytes
}

// EvictOldestBucket removes the oldest bucket (thread-safe).
func (b *AppBuffer) EvictOldestBucket() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.evictOldest()
}

// UpdateMemory updates the memory usage for a bucket.
func (b *AppBuffer) UpdateMemory(bucket *models.TimeBucket, newSize int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	diff := newSize - bucket.MemoryBytes
	bucket.MemoryBytes = newSize
	b.memoryUsed += diff
}

// GetBuckets returns a copy of all buckets.
func (b *AppBuffer) GetBuckets() []*models.TimeBucket {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]*models.TimeBucket, len(b.buckets))
	copy(result, b.buckets)
	return result
}

// GetBucketsInRange returns buckets within the given time range.
func (b *AppBuffer) GetBucketsInRange(start, end time.Time) []*models.TimeBucket {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var result []*models.TimeBucket
	for _, bucket := range b.buckets {
		if !bucket.EndTime.Before(start) && !bucket.StartTime.After(end) {
			result = append(result, bucket)
		}
	}
	return result
}

// GetInfo returns summary information about the buffer.
func (b *AppBuffer) GetInfo() models.AppInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	info := models.AppInfo{
		AppID:       b.appID,
		BucketCount: len(b.buckets),
	}

	if len(b.buckets) > 0 {
		info.OldestData = b.buckets[0].StartTime
		info.NewestData = b.buckets[len(b.buckets)-1].EndTime
		for _, bucket := range b.buckets {
			info.TotalSamples += bucket.SampleCount
		}
	}

	return info
}

// CanEvict returns true if the buffer has more than the minimum required buckets.
func (b *AppBuffer) CanEvict() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.buckets) > MinBucketsPerApp
}
