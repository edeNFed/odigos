package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProfilingProxyHandler returns 503 Service Unavailable since profiling
// has been migrated to the OTel eBPF profiler in the node collector.
// This is a temporary placeholder until full UI integration is implemented.
func ProfilingProxyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusServiceUnavailable, "Profiling service not available")
	}
}
