// app/server/main.go - Fixed unused variable
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayaseen/openshift-health-dashboard/app/server/server"
)

func main() {
	// Configure logging with file and line information
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Starting OpenShift Health Dashboard server")

	// Get configuration from environment variables
	config := server.Config{
		StaticDir: getEnv("STATIC_DIR", "./app/web/static"),
		Port:      getEnv("PORT", "8080"),
		DebugMode: getEnv("DEBUG", "false") == "true",
	}

	if config.DebugMode {
		log.Println("Debug mode enabled")
	}

	// Create and start the server
	s := server.NewServer(config)

	// Initialize server
	if err := s.Initialize(); err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	// Start the server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Server listening on port %s", config.Port)
		serverErrors <- s.Start()
	}()

	// Set up graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until shutdown or error
	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)

	case <-shutdown:
		log.Println("Shutting down gracefully...")

		// Create a timeout context for shutdown
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer timeoutCancel()

		if err := s.Shutdown(timeoutCtx); err != nil {
			log.Fatalf("Error during shutdown: %v", err)
		}

		log.Println("Server shutdown complete")
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
