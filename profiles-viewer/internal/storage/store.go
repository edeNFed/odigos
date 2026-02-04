package storage

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/odigos-io/odigos/profiles-viewer/internal/models"
	"go.opentelemetry.io/collector/pdata/pprofile"
)

// ProfileStore manages profile storage with circular buffers per app.
type ProfileStore struct {
	mu            sync.RWMutex
	apps          map[string]*AppBuffer // key: appID.String()
	memoryManager *MemoryManager
}

// NewProfileStore creates a new profile store.
func NewProfileStore() *ProfileStore {
	return &ProfileStore{
		apps:          make(map[string]*AppBuffer),
		memoryManager: NewMemoryManager(),
	}
}

// IngestProfiles ingests profile data for an app.
func (s *ProfileStore) IngestProfiles(ctx context.Context, appID models.AppID, pd pprofile.Profiles) error {
	if appID.IsEmpty() {
		log.Println("Skipping profiles with empty app ID")
		return nil
	}

	s.mu.Lock()
	buffer, exists := s.apps[appID.String()]
	if !exists {
		buffer = NewAppBuffer(appID)
		s.apps[appID.String()] = buffer
		s.memoryManager.RegisterApp()
		log.Printf("Registered new app: %s", appID.String())
	}
	s.mu.Unlock()

	// Get the shared dictionary
	dict := pd.ProfilesDictionary()

	// Process each resource profile
	for i := 0; i < pd.ResourceProfiles().Len(); i++ {
		rp := pd.ResourceProfiles().At(i)
		for j := 0; j < rp.ScopeProfiles().Len(); j++ {
			sp := rp.ScopeProfiles().At(j)
			for k := 0; k < sp.Profiles().Len(); k++ {
				profile := sp.Profiles().At(k)
				s.processProfile(buffer, profile, dict)
			}
		}
	}

	// Check memory budget and evict if needed
	s.enforceMemoryBudget()

	return nil
}

// processProfile processes a single profile and adds it to the buffer.
func (s *ProfileStore) processProfile(buffer *AppBuffer, profile pprofile.Profile, dict pprofile.ProfilesDictionary) {
	// Determine timestamp from the profile
	ts := time.Now()
	if profile.Time() != 0 {
		ts = profile.Time().AsTime()
	}

	bucket := buffer.GetOrCreateBucket(ts)

	// Extract and merge profile data
	s.mergeProfileIntoBucket(bucket, profile, dict)

	// Update memory estimate
	memSize := s.estimateProfileMemory(bucket.Profile)
	buffer.UpdateMemory(bucket, memSize)
	s.memoryManager.AddMemory(memSize - bucket.MemoryBytes)
}

// mergeProfileIntoBucket merges profile data into an existing bucket.
func (s *ProfileStore) mergeProfileIntoBucket(bucket *models.TimeBucket, profile pprofile.Profile, dict pprofile.ProfilesDictionary) {
	aggProfile := bucket.Profile

	// Get tables from the dictionary
	stringTable := dict.StringTable()
	funcTable := dict.FunctionTable()
	locTable := dict.LocationTable()

	// Add strings to our string table
	for i := 0; i < stringTable.Len(); i++ {
		aggProfile.StringTable = append(aggProfile.StringTable, stringTable.At(i))
	}

	// Process function table
	funcMapping := make(map[int64]int64) // old ID -> new ID
	for i := 0; i < funcTable.Len(); i++ {
		fn := funcTable.At(i)
		oldID := int64(i)
		newID := int64(len(aggProfile.Functions))

		name := ""
		if fn.NameStrindex() >= 0 && int(fn.NameStrindex()) < stringTable.Len() {
			name = stringTable.At(int(fn.NameStrindex()))
		}
		filename := ""
		if fn.FilenameStrindex() >= 0 && int(fn.FilenameStrindex()) < stringTable.Len() {
			filename = stringTable.At(int(fn.FilenameStrindex()))
		}

		aggProfile.Functions[newID] = &models.FunctionInfo{
			ID:       newID,
			Name:     name,
			Filename: filename,
			Line:     fn.StartLine(),
		}
		funcMapping[oldID] = newID
	}

	// Process location table for stack building
	locationFuncs := make(map[int64][]int64) // location index -> function IDs
	for i := 0; i < locTable.Len(); i++ {
		loc := locTable.At(i)
		var funcIDs []int64
		for j := 0; j < loc.Line().Len(); j++ {
			line := loc.Line().At(j)
			if newID, ok := funcMapping[int64(line.FunctionIndex())]; ok {
				funcIDs = append(funcIDs, newID)
			}
		}
		locationFuncs[int64(i)] = funcIDs
	}

	// Process samples using LocationIndices from profile
	samples := profile.Sample()
	locationIndices := profile.LocationIndices()

	for i := 0; i < samples.Len(); i++ {
		sample := samples.At(i)

		// Build stack from location indices
		var functionIDs []int64
		locStart := int(sample.LocationsStartIndex())
		locLen := int(sample.LocationsLength())

		for j := locStart; j < locStart+locLen && j < locationIndices.Len(); j++ {
			locIdx := locationIndices.At(j)
			if funcs, ok := locationFuncs[int64(locIdx)]; ok {
				functionIDs = append(functionIDs, funcs...)
			}
		}

		// Compute stack hash
		stackHash := computeStackHash(functionIDs)

		// Get sample value (first value or 1)
		sampleValue := int64(1)
		if sample.Value().Len() > 0 {
			sampleValue = sample.Value().At(0)
		}

		// Merge or create stack
		if existing, ok := aggProfile.Stacks[stackHash]; ok {
			existing.SampleCount += sampleValue
		} else {
			aggProfile.Stacks[stackHash] = &models.StackInfo{
				StackHash:   stackHash,
				FunctionIDs: functionIDs,
				SampleCount: sampleValue,
			}
		}

		bucket.SampleCount += sampleValue
		aggProfile.TotalSamples += sampleValue
	}
}

// computeStackHash computes a hash for a stack of function IDs.
func computeStackHash(functionIDs []int64) uint64 {
	var hash uint64 = 14695981039346656037 // FNV-1a offset basis
	for _, id := range functionIDs {
		hash ^= uint64(id)
		hash *= 1099511628211 // FNV-1a prime
	}
	return hash
}

// estimateProfileMemory estimates memory usage of an aggregated profile.
func (s *ProfileStore) estimateProfileMemory(profile *models.AggregatedProfile) int64 {
	var size int64

	// String table
	for _, str := range profile.StringTable {
		size += int64(len(str) + 16) // string + header
	}

	// Functions
	size += int64(len(profile.Functions) * 80) // estimated per-function

	// Stacks
	for _, stack := range profile.Stacks {
		size += int64(48 + len(stack.FunctionIDs)*8) // header + slice
	}

	return size
}

// enforceMemoryBudget evicts old data if memory budget is exceeded.
func (s *ProfileStore) enforceMemoryBudget() {
	for s.memoryManager.IsOverBudget() {
		s.mu.RLock()
		var oldestApp *AppBuffer
		var oldestTime time.Time

		for _, buffer := range s.apps {
			if !buffer.CanEvict() {
				continue
			}
			buckets := buffer.GetBuckets()
			if len(buckets) > 0 {
				if oldestApp == nil || buckets[0].StartTime.Before(oldestTime) {
					oldestApp = buffer
					oldestTime = buckets[0].StartTime
				}
			}
		}
		s.mu.RUnlock()

		if oldestApp == nil {
			// Cannot evict more
			break
		}

		evictedSize := oldestApp.EvictOldestBucket()
		s.memoryManager.RemoveMemory(evictedSize)
		log.Printf("Evicted bucket from %s, freed %d bytes", oldestApp.AppID().String(), evictedSize)
	}
}

// ListApps returns information about all apps with profiles.
func (s *ProfileStore) ListApps() []models.AppInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var apps []models.AppInfo
	for _, buffer := range s.apps {
		apps = append(apps, buffer.GetInfo())
	}
	return apps
}

// GetApp returns the buffer for a specific app.
func (s *ProfileStore) GetApp(appID models.AppID) *AppBuffer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.apps[appID.String()]
}

// GetBuckets returns bucket info for an app.
func (s *ProfileStore) GetBuckets(appID models.AppID) []models.BucketInfo {
	buffer := s.GetApp(appID)
	if buffer == nil {
		return nil
	}

	buckets := buffer.GetBuckets()
	var infos []models.BucketInfo
	for _, b := range buckets {
		infos = append(infos, models.BucketInfo{
			StartTime:   b.StartTime,
			EndTime:     b.EndTime,
			SampleCount: b.SampleCount,
		})
	}
	return infos
}

// GetBucketsInRange returns buckets within a time range for an app.
func (s *ProfileStore) GetBucketsInRange(appID models.AppID, start, end time.Time) []*models.TimeBucket {
	buffer := s.GetApp(appID)
	if buffer == nil {
		return nil
	}
	return buffer.GetBucketsInRange(start, end)
}

// Stats returns storage statistics.
func (s *ProfileStore) Stats() map[string]interface{} {
	return map[string]interface{}{
		"totalMemoryUsed": s.memoryManager.TotalUsed(),
		"memoryBudget":    UsableMemory,
		"numApps":         s.memoryManager.NumApps(),
		"budgetPerApp":    s.memoryManager.BudgetPerApp(),
		"availableMemory": s.memoryManager.AvailableMemory(),
	}
}
