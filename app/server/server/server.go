// app/server/server/server.go - Updated with improved file processing logic
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ayaseen/openshift-health-dashboard/app/server/types"
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
	config     Config
	handler    http.Handler
	httpServer *http.Server
	isReady    atomic.Bool
}

// NewServer creates a new server instance
func NewServer(config Config) *Server {
	// Create the server
	s := &Server{
		config: config,
	}

	// Set the server as not ready initially
	s.isReady.Store(false)

	// Set up the HTTP handler
	s.setupHandler()

	return s
}

// Initialize performs any necessary initialization before the server starts
func (s *Server) Initialize() error {
	// Check if static directory exists
	if _, err := os.Stat(s.config.StaticDir); os.IsNotExist(err) {
		return fmt.Errorf("static directory does not exist: %s", s.config.StaticDir)
	}

	// Check if index.html exists in static directory
	indexPath := filepath.Join(s.config.StaticDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return fmt.Errorf("index.html not found in static directory: %s", indexPath)
	}

	log.Printf("Initialization complete, server is ready")

	// Mark the server as ready
	s.isReady.Store(true)
	return nil
}

// setupHandler configures the HTTP handler
func (s *Server) setupHandler() {
	// Create a custom handler with logging
	mux := http.NewServeMux()

	// Add API endpoints
	mux.HandleFunc("/api/parse-report", s.HandleReportUpload)

	// Health check endpoint for liveness probe
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Readiness probe endpoint
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if s.isReady.Load() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ready"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"not ready"}`))
		}
	})

	// Set up static file serving
	staticHandler := http.FileServer(http.Dir(s.config.StaticDir))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		if s.config.DebugMode {
			log.Printf("%s - %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		}

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
				if s.config.DebugMode {
					log.Println("Serving index.html for root path")
				}
				http.ServeFile(w, r, indexPath)
				return
			}
		}

		// If path doesn't exist and it's not a file with extension, serve index.html for SPA routing
		if os.IsNotExist(err) && r.URL.Path != "/" {
			// If it's a file request with extension, return 404
			if filepath.Ext(r.URL.Path) != "" {
				if s.config.DebugMode {
					log.Printf("File not found: %s, returning 404", path)
				}
				http.NotFound(w, r)
				return
			}

			// Otherwise serve index.html for SPA routing
			if s.config.DebugMode {
				log.Printf("Path not found: %s, serving index.html for SPA routing", path)
			}
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

	// Ensure file is flushed and closed before parsing
	tempFile.Sync()
	tempFile.Close()

	// Parse the report
	summary, err := utils.ParseAsciiDocExecutiveSummary(tempFile.Name())
	if err != nil {
		log.Printf("Error parsing report: %v", err)
		http.Error(w, "Failed to parse report", http.StatusInternalServerError)
		return
	}

	// Validate and fix the summary data
	validateSummaryData(summary)

	// Return the summary as JSON
	if err := json.NewEncoder(w).Encode(summary); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	if s.config.DebugMode {
		log.Printf("Successfully processed report: %s", header.Filename)
		log.Printf("Found %d required changes, %d recommended changes, %d advisory items",
			len(summary.ItemsRequired), len(summary.ItemsRecommended), len(summary.ItemsAdvisory))
	}
}

// validateSummaryData ensures all summary data is valid and provides fallbacks where needed
func validateSummaryData(summary *types.ReportSummary) {
	// Ensure overall score is within range
	if summary.OverallScore < 0 {
		summary.OverallScore = 0
	} else if summary.OverallScore > 100 {
		summary.OverallScore = 100
	}

	// Ensure all category scores are within range and have fallbacks
	validateCategoryScore(&summary.ScoreInfra)
	validateCategoryScore(&summary.ScoreGovernance)
	validateCategoryScore(&summary.ScoreCompliance)
	validateCategoryScore(&summary.ScoreMonitoring)
	validateCategoryScore(&summary.ScoreBuildSecurity)

	// Ensure lists are initialized
	if summary.ItemsRequired == nil {
		summary.ItemsRequired = []string{}
	}
	if summary.ItemsRecommended == nil {
		summary.ItemsRecommended = []string{}
	}
	if summary.ItemsAdvisory == nil {
		summary.ItemsAdvisory = []string{}
	}

	// Ensure descriptions are set
	if summary.InfraDescription == "" {
		summary.InfraDescription = utils.GenerateDescription("Infrastructure Setup", summary.ScoreInfra)
	}
	if summary.GovernanceDescription == "" {
		summary.GovernanceDescription = utils.GenerateDescription("Policy Governance", summary.ScoreGovernance)
	}
	if summary.ComplianceDescription == "" {
		summary.ComplianceDescription = utils.GenerateDescription("Compliance Benchmarking", summary.ScoreCompliance)
	}
	if summary.MonitoringDescription == "" {
		summary.MonitoringDescription = utils.GenerateDescription("Central Monitoring", summary.ScoreMonitoring)
	}
	if summary.BuildSecurityDescription == "" {
		summary.BuildSecurityDescription = utils.GenerateDescription("Build/Deploy Security", summary.ScoreBuildSecurity)
	}
}

// validateCategoryScore ensures a category score is within valid range with a fallback
func validateCategoryScore(score *int) {
	if *score < 0 {
		*score = 0
	} else if *score > 100 {
		*score = 100
	} else if *score == 0 {
		*score = 75 // Default fallback if not found
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create a custom server with timeouts
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%s", s.config.Port),
		Handler:      s.handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start the server
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
