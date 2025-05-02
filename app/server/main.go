// app/server/main.go
package main

import (
	"log"
	"os"

	"github.com/ayaseen/openshift-health-dashboard/app/server/server"
)

func main() {
	// Configure logging
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
	log.Printf("Starting server on port %s", config.Port)
	log.Printf("Serving static files from: %s", config.StaticDir)
	log.Fatal(s.Start())
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
