package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/odigos-io/odigos/profiles-viewer/internal/aggregator"
	"github.com/odigos-io/odigos/profiles-viewer/internal/models"
	"github.com/odigos-io/odigos/profiles-viewer/internal/storage"
)

// Handlers contains HTTP request handlers.
type Handlers struct {
	store *storage.ProfileStore
}

// NewHandlers creates a new handlers instance.
func NewHandlers(store *storage.ProfileStore) *Handlers {
	return &Handlers{store: store}
}

// ListApps returns all applications with profiles.
func (h *Handlers) ListApps(c *gin.Context) {
	apps := h.store.ListApps()
	c.JSON(http.StatusOK, gin.H{
		"apps": apps,
	})
}

// GetBuckets returns available time buckets for an app.
func (h *Handlers) GetBuckets(c *gin.Context) {
	appID := models.AppID{
		Namespace: c.Param("ns"),
		Kind:      c.Param("kind"),
		Name:      c.Param("name"),
	}

	buckets := h.store.GetBuckets(appID)
	if buckets == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "app not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"buckets": buckets,
	})
}

// GetFlameGraph returns flame graph data for an app within a time range.
func (h *Handlers) GetFlameGraph(c *gin.Context) {
	appID := models.AppID{
		Namespace: c.Param("ns"),
		Kind:      c.Param("kind"),
		Name:      c.Param("name"),
	}

	// Parse time range
	startStr := c.Query("start")
	endStr := c.Query("end")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid start time format, use RFC3339",
			})
			return
		}
	} else {
		// Default to all time
		start = time.Time{}
	}

	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid end time format, use RFC3339",
			})
			return
		}
	} else {
		end = time.Now().Add(time.Hour) // Future to include current data
	}

	// Get buckets in range
	buckets := h.store.GetBucketsInRange(appID, start, end)
	if len(buckets) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "no profile data found for the specified time range",
		})
		return
	}

	// Merge profiles
	merged := aggregator.MergeProfiles(buckets)

	// Build flame graph
	flameGraph := BuildFlameGraph(merged)
	SortFlameGraph(flameGraph)

	c.JSON(http.StatusOK, flameGraph)
}

// HealthCheck returns health status.
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// ReadinessCheck returns readiness status.
func (h *Handlers) ReadinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// GetStats returns storage statistics.
func (h *Handlers) GetStats(c *gin.Context) {
	stats := h.store.Stats()
	c.JSON(http.StatusOK, stats)
}
