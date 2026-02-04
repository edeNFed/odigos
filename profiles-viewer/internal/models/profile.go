package models

import "time"

// FunctionInfo represents a function in the profile.
type FunctionInfo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Line     int64  `json:"line"`
}

// StackInfo represents an aggregated stack with sample count.
type StackInfo struct {
	StackHash   uint64  `json:"stackHash"`
	FunctionIDs []int64 `json:"functionIds"`
	SampleCount int64   `json:"sampleCount"`
}

// AggregatedProfile holds merged profile data for a time bucket.
type AggregatedProfile struct {
	StringTable  []string                `json:"stringTable"`
	Functions    map[int64]*FunctionInfo `json:"functions"`
	Stacks       map[uint64]*StackInfo   `json:"stacks"` // Hash -> aggregated stack
	TotalSamples int64                   `json:"totalSamples"`
}

// NewAggregatedProfile creates a new empty aggregated profile.
func NewAggregatedProfile() *AggregatedProfile {
	return &AggregatedProfile{
		StringTable: make([]string, 0),
		Functions:   make(map[int64]*FunctionInfo),
		Stacks:      make(map[uint64]*StackInfo),
	}
}

// TimeBucket represents a 30-second time window of aggregated profile data.
type TimeBucket struct {
	StartTime   time.Time          `json:"startTime"`
	EndTime     time.Time          `json:"endTime"`
	Profile     *AggregatedProfile `json:"profile"`
	SampleCount int64              `json:"sampleCount"`
	MemoryBytes int64              `json:"memoryBytes"`
}

// NewTimeBucket creates a new time bucket for the given time window.
func NewTimeBucket(start, end time.Time) *TimeBucket {
	return &TimeBucket{
		StartTime:   start,
		EndTime:     end,
		Profile:     NewAggregatedProfile(),
		SampleCount: 0,
		MemoryBytes: 0,
	}
}

// FlameGraphNode represents a node in the flame graph tree structure.
type FlameGraphNode struct {
	Name     string            `json:"name"`
	Value    int64             `json:"value"`
	Children []*FlameGraphNode `json:"children,omitempty"`
}

// BucketInfo provides metadata about a time bucket for the API.
type BucketInfo struct {
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	SampleCount int64     `json:"sampleCount"`
}

// AppInfo provides information about an app with profiles.
type AppInfo struct {
	AppID        AppID     `json:"appId"`
	BucketCount  int       `json:"bucketCount"`
	TotalSamples int64     `json:"totalSamples"`
	OldestData   time.Time `json:"oldestData"`
	NewestData   time.Time `json:"newestData"`
}
