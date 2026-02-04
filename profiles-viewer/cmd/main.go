package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/odigos-io/odigos/profiles-viewer/internal/api"
	"github.com/odigos-io/odigos/profiles-viewer/internal/receiver"
	"github.com/odigos-io/odigos/profiles-viewer/internal/storage"
	"github.com/odigos-io/odigos/profiles-viewer/webapp"
)

func main() {
	httpAddr := flag.String("http-addr", "0.0.0.0:8080", "HTTP server listen address")
	otlpAddr := flag.String("otlp-addr", "0.0.0.0:4317", "OTLP gRPC receiver listen address")
	flag.Parse()

	log.Println("Starting profiles-viewer service")

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Initialize storage
	store := storage.NewProfileStore()

	// Create profiles consumer
	profilesConsumer := receiver.NewProfilesConsumer(store)

	// Create HTTP server
	server := api.NewServer(store, *httpAddr, webapp.StaticFS)

	// Start services
	var wg sync.WaitGroup

	// Start OTLP receiver
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := profilesConsumer.Run(ctx, *otlpAddr); err != nil {
			log.Printf("OTLP receiver error: %v", err)
		}
	}()

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Run(ctx); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	log.Printf("profiles-viewer ready: HTTP on %s, OTLP on %s", *httpAddr, *otlpAddr)

	// Wait for shutdown
	wg.Wait()
	log.Println("profiles-viewer shutdown complete")
}
