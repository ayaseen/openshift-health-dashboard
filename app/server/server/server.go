// app/server/server/server.go
package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config holds server configuration
type Config struct {
	StaticDir string
	Port      string
	DebugMode bool
}

// Server represents the HTTP server
type Server struct {
	config  Config
	handler http.Handler
}

// NewServer creates a new server instance
func NewServer(config Config) *Server {
	// Create the server
	s := &Server{
		config: config,
	}

	// Set up the HTTP handler
	s.setupHandler()

	return s
}

// setupHandler configures the HTTP handler
func (s *Server) setupHandler() {
	// Create a custom handler with logging
	mux := http.NewServeMux()

	// Add API endpoint for report upload and parsing
	mux.HandleFunc("/api/parse-report", s.HandleReportUpload)

	// Set up static file serving
	staticHandler := http.FileServer(http.Dir(s.config.StaticDir))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		log.Printf("%s - %s %s", r.RemoteAddr, r.Method, r.URL.Path)

		// Add headers to prevent caching
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// For API requests, let them be handled by specific handlers
		if strings.HasPrefix(r.URL.Path, "/api/") {
			return
		}

		// Check if the path exists
		path := filepath.Join(s.config.StaticDir, r.URL.Path)
		_, err := os.Stat(path)

		// Special handling for root path or index.html
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			indexPath := filepath.Join(s.config.StaticDir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				log.Println("Serving index.html for root path")
				http.ServeFile(w, r, indexPath)
				return
			}
		}

		// If path doesn't exist and it's not a file with extension, serve index.html for SPA routing
		if os.IsNotExist(err) && r.URL.Path != "/" {
			// If it's a file request with extension, return 404
			if filepath.Ext(r.URL.Path) != "" {
				log.Printf("File not found: %s, returning 404", path)
				http.NotFound(w, r)
				return
			}

			// Otherwise serve index.html for SPA routing
			log.Printf("Path not found: %s, serving index.html for SPA routing", path)
			http.ServeFile(w, r, filepath.Join(s.config.StaticDir, "index.html"))
			return
		}

		// Serve the file
		staticHandler.ServeHTTP(w, r)
	}))

	// Store the handler
	s.handler = mux
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create a custom server with timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", s.config.Port),
		Handler:      s.handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start the server
	return server.ListenAndServe()
}
