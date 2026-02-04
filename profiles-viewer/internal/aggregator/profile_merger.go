package aggregator

import (
	"github.com/odigos-io/odigos/profiles-viewer/internal/models"
)

// MergeProfiles merges multiple time buckets into a single aggregated profile.
func MergeProfiles(buckets []*models.TimeBucket) *models.AggregatedProfile {
	if len(buckets) == 0 {
		return models.NewAggregatedProfile()
	}

	if len(buckets) == 1 {
		return buckets[0].Profile
	}

	merged := models.NewAggregatedProfile()

	// Build unified string table and function mapping
	stringMap := make(map[string]int) // string -> index in merged table
	funcMap := make(map[string]int64) // "name:file:line" -> new function ID

	for _, bucket := range buckets {
		profile := bucket.Profile

		// Process strings
		for _, str := range profile.StringTable {
			if _, exists := stringMap[str]; !exists {
				stringMap[str] = len(merged.StringTable)
				merged.StringTable = append(merged.StringTable, str)
			}
		}

		// Process functions and build mapping
		localFuncMap := make(map[int64]int64) // old ID -> new ID
		for oldID, fn := range profile.Functions {
			key := funcKey(fn)
			if newID, exists := funcMap[key]; exists {
				localFuncMap[oldID] = newID
			} else {
				newID := int64(len(merged.Functions))
				merged.Functions[newID] = &models.FunctionInfo{
					ID:       newID,
					Name:     fn.Name,
					Filename: fn.Filename,
					Line:     fn.Line,
				}
				funcMap[key] = newID
				localFuncMap[oldID] = newID
			}
		}

		// Process stacks with remapped function IDs
		for _, stack := range profile.Stacks {
			// Remap function IDs
			newFuncIDs := make([]int64, len(stack.FunctionIDs))
			for i, oldFuncID := range stack.FunctionIDs {
				if newID, ok := localFuncMap[oldFuncID]; ok {
					newFuncIDs[i] = newID
				} else {
					newFuncIDs[i] = oldFuncID
				}
			}

			// Compute new hash
			newHash := computeStackHash(newFuncIDs)

			// Merge or add
			if existing, ok := merged.Stacks[newHash]; ok {
				existing.SampleCount += stack.SampleCount
			} else {
				merged.Stacks[newHash] = &models.StackInfo{
					StackHash:   newHash,
					FunctionIDs: newFuncIDs,
					SampleCount: stack.SampleCount,
				}
			}
			merged.TotalSamples += stack.SampleCount
		}
	}

	return merged
}

// funcKey creates a unique key for a function.
func funcKey(fn *models.FunctionInfo) string {
	return fn.Name + ":" + fn.Filename + ":" + string(rune(fn.Line))
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
