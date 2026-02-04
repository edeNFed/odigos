package storage

import (
	"sync/atomic"
)

const (
	// MaxMemoryBytes is the total memory budget for profile storage.
	MaxMemoryBytes = 750 * 1024 * 1024 // 750MB
	// ReservedOverhead is memory reserved for runtime overhead.
	ReservedOverhead = 50 * 1024 * 1024 // 50MB
	// UsableMemory is the memory available for profile data.
	UsableMemory = MaxMemoryBytes - ReservedOverhead // 700MB
	// MinBucketsPerApp is the minimum number of buckets to keep per app.
	MinBucketsPerApp = 10
	// MaxBucketsPerApp is the maximum number of buckets per app (20 min at 30s intervals).
	MaxBucketsPerApp = 40
	// BucketDurationSeconds is the duration of each time bucket.
	BucketDurationSeconds = 30
)

// MemoryManager tracks and manages memory allocation across apps.
type MemoryManager struct {
	totalUsed atomic.Int64
	numApps   atomic.Int32
}

// NewMemoryManager creates a new memory manager.
func NewMemoryManager() *MemoryManager {
	return &MemoryManager{}
}

// TotalUsed returns the current total memory usage.
func (m *MemoryManager) TotalUsed() int64 {
	return m.totalUsed.Load()
}

// NumApps returns the current number of apps.
func (m *MemoryManager) NumApps() int32 {
	return m.numApps.Load()
}

// AddMemory adds to the total memory usage.
func (m *MemoryManager) AddMemory(bytes int64) {
	m.totalUsed.Add(bytes)
}

// RemoveMemory removes from the total memory usage.
func (m *MemoryManager) RemoveMemory(bytes int64) {
	m.totalUsed.Add(-bytes)
}

// RegisterApp increments the app count.
func (m *MemoryManager) RegisterApp() {
	m.numApps.Add(1)
}

// UnregisterApp decrements the app count.
func (m *MemoryManager) UnregisterApp() {
	m.numApps.Add(-1)
}

// BudgetPerApp returns the memory budget allocated per app.
func (m *MemoryManager) BudgetPerApp() int64 {
	numApps := m.numApps.Load()
	if numApps == 0 {
		return UsableMemory
	}
	return UsableMemory / int64(numApps)
}

// IsOverBudget returns true if total memory usage exceeds the limit.
func (m *MemoryManager) IsOverBudget() bool {
	return m.totalUsed.Load() > UsableMemory
}

// AvailableMemory returns the remaining available memory.
func (m *MemoryManager) AvailableMemory() int64 {
	available := UsableMemory - m.totalUsed.Load()
	if available < 0 {
		return 0
	}
	return available
}
