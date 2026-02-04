package api

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/odigos-io/odigos/profiles-viewer/internal/storage"
)

// Server represents the HTTP API server.
type Server struct {
	engine   *gin.Engine
	handlers *Handlers
	addr     string
}

// NewServer creates a new HTTP server.
func NewServer(store *storage.ProfileStore, addr string, staticFS embed.FS) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(cors.Default())

	handlers := NewHandlers(store)

	// API routes
	api := engine.Group("/api")
	{
		api.GET("/apps", handlers.ListApps)
		api.GET("/apps/:ns/:kind/:name/buckets", handlers.GetBuckets)
		api.GET("/apps/:ns/:kind/:name/flamegraph", handlers.GetFlameGraph)
		api.GET("/stats", handlers.GetStats)
	}

	// Health endpoints
	engine.GET("/healthz", handlers.HealthCheck)
	engine.GET("/readyz", handlers.ReadinessCheck)

	// Serve static files
	setupStaticFiles(engine, staticFS)

	return &Server{
		engine:   engine,
		handlers: handlers,
		addr:     addr,
	}
}

// setupStaticFiles configures serving of embedded static files.
func setupStaticFiles(engine *gin.Engine, staticFS embed.FS) {
	// Try to get the dist subdirectory (embedded from webapp/dist)
	dist, err := fs.Sub(staticFS, "dist")
	if err != nil {
		log.Printf("Warning: Could not load static files: %v", err)
		return
	}

	// Create file server for static assets
	fileServer := http.FileServer(http.FS(dist))

	// Handle all non-API routes
	engine.NoRoute(func(c *gin.Context) {
		urlPath := c.Request.URL.Path

		// Don't serve for API routes
		if strings.HasPrefix(urlPath, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
			return
		}

		// Try to serve the file directly
		filePath := strings.TrimPrefix(urlPath, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		// Check if file exists
		f, err := dist.Open(filePath)
		if err == nil {
			f.Close()
			// Set correct content type based on extension
			ext := path.Ext(filePath)
			switch ext {
			case ".js":
				c.Header("Content-Type", "application/javascript")
			case ".css":
				c.Header("Content-Type", "text/css")
			case ".html":
				c.Header("Content-Type", "text/html; charset=utf-8")
			case ".svg":
				c.Header("Content-Type", "image/svg+xml")
			case ".json":
				c.Header("Content-Type", "application/json")
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// File not found, serve index.html for SPA routing
		data, err := fs.ReadFile(dist, "index.html")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
}

// Run starts the HTTP server.
func (s *Server) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:    s.addr,
		Handler: s.engine,
	}

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		log.Println("Shutting down HTTP server...")
		server.Shutdown(context.Background())
	}()

	log.Printf("Starting HTTP server on %s", s.addr)
	return server.ListenAndServe()
}
