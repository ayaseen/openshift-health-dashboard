// app/server/server/server.go
package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ayaseen/openshift-health-dashboard/app/server/utils"
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

// HandleReportUpload processes uploaded AsciiDoc reports
func (s *Server) HandleReportUpload(w http.ResponseWriter, r *http.Request) {
	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Check if the request method is POST
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.config.DebugMode {
		log.Printf("Handling report upload request")
	}

	// Parse the multipart form with 10MB max memory
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("report")
	if err != nil {
		log.Printf("Error getting file: %v", err)
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Received file: %s, size: %d bytes", header.Filename, header.Size)

	// Check file extension
	if !utils.IsValidAsciiDocFile(header.Filename) {
		http.Error(w, "Invalid file type. Only .adoc or .asciidoc files are allowed", http.StatusBadRequest)
		return
	}

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "report-*.adoc")
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		http.Error(w, "Failed to process file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy the uploaded file to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		log.Printf("Error copying file: %v", err)
		http.Error(w, "Failed to process file", http.StatusInternalServerError)
		return
	}

	// Parse the report
	summary, err := utils.ParseAsciiDocExecutiveSummary(tempFile.Name())
	if err != nil {
		log.Printf("Error parsing report: %v", err)
		http.Error(w, "Failed to parse report", http.StatusInternalServerError)
		return
	}

	// Return the summary as JSON
	if err := json.NewEncoder(w).Encode(summary); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	if s.config.DebugMode {
		log.Printf("Successfully processed report: %s", header.Filename)
	}
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
