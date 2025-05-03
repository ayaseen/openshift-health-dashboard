// app/server/server/server.go
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
	"regexp"
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
	// Set content type header and CORS headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check if the request method is POST
	if r.Method != "POST" {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if s.config.DebugMode {
		log.Printf("Handling report upload request")
	}

	// Parse the multipart form with 10MB max memory
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, `{"error":"Failed to parse form"}`, http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("report")
	if err != nil {
		log.Printf("Error getting file: %v", err)
		http.Error(w, `{"error":"Failed to get file"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Received file: %s, size: %d bytes", header.Filename, header.Size)

	// Check file extension
	if !utils.IsValidAsciiDocFile(header.Filename) {
		http.Error(w, `{"error":"Invalid file type. Only .adoc or .asciidoc files are allowed"}`, http.StatusBadRequest)
		return
	}

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "report-*.adoc")
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		http.Error(w, `{"error":"Failed to process file"}`, http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy the uploaded file to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		log.Printf("Error copying file: %v", err)
		http.Error(w, `{"error":"Failed to process file"}`, http.StatusInternalServerError)
		return
	}

	// Ensure file is flushed
	tempFile.Sync()

	// Parse the AsciiDoc file directly (without relying on utils)
	fileContent, err := os.ReadFile(tempFile.Name())
	if err != nil {
		log.Printf("Error reading file: %v", err)
		http.Error(w, `{"error":"Failed to read file"}`, http.StatusInternalServerError)
		return
	}

	// Extract data from the file
	summary, err := parseAsciiDocReport(string(fileContent))
	if err != nil {
		log.Printf("Error parsing report: %v", err)
		http.Error(w, fmt.Sprintf(`{"error":"Failed to parse report: %s"}`, err), http.StatusInternalServerError)
		return
	}

	// Return the summary as JSON
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(summary); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, `{"error":"Failed to encode response"}`, http.StatusInternalServerError)
		return
	}

	if s.config.DebugMode {
		log.Printf("Successfully processed report: %s", header.Filename)
		log.Printf("Found %d required changes, %d recommended changes, %d advisory items",
			len(summary.ItemsRequired), len(summary.ItemsRecommended), len(summary.ItemsAdvisory))
	}
}

// parseAsciiDocReport parses an AsciiDoc report directly
func parseAsciiDocReport(content string) (*types.ReportSummary, error) {
	// Split content into lines
	lines := strings.Split(content, "\n")

	// Initialize summary struct
	summary := &types.ReportSummary{
		ItemsRequired:    []string{},
		ItemsRecommended: []string{},
		ItemsAdvisory:    []string{},
		NoChangeCount:    0,
	}

	// Extract summary section
	var requiredItems, recommendedItems, advisoryItems []string
	var noChangeCount, notApplicableCount int

	// Find where the Summary section starts
	summaryStartIndex := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "= Summary" {
			summaryStartIndex = i
			break
		}
	}

	if summaryStartIndex == -1 {
		return summary, nil // No summary section found
	}

	// Scan the Summary section for the table
	inTable := false
	inKey := true // Assume we start in the key/legend section

	for i := summaryStartIndex; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// End of Summary section
		if line != "" && line[0] == '=' && !strings.Contains(line, "= Summary") {
			break
		}

		// Check for table start
		if strings.Contains(line, "|===") {
			if !inTable {
				inTable = true
				continue
			} else {
				// End of table
				inTable = false
				break
			}
		}

		if !inTable {
			continue
		}

		// Check if we're past the key/legend section
		if inKey && strings.Contains(line, "*Category*") &&
			strings.Contains(line, "*Item Evaluated*") {
			inKey = false
			continue
		}

		// Skip the key/legend rows
		if inKey || line == "" {
			continue
		}

		// Process table rows for status
		if strings.Contains(line, "{set:cellbgcolor:#FF0000}") &&
			!strings.Contains(line, "Indicates Changes Required") {
			// Get the item name and content
			var itemContent string

			// Look for name in previous or next lines
			for j := i - 5; j <= i+5 && j < len(lines); j++ {
				if j >= 0 && strings.Contains(lines[j], "<<") && strings.Contains(lines[j], ">>") {
					nameMatch := regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(lines[j])
					if len(nameMatch) > 1 {
						itemName := nameMatch[1]

						// Look for observation text
						for k := j + 1; k < i; k++ {
							obsLine := strings.TrimSpace(lines[k])
							if obsLine != "" && !strings.Contains(obsLine, "set:cellbgcolor") {
								if strings.HasPrefix(obsLine, "|") {
									obsLine = strings.TrimSpace(obsLine[1:])
								}
								itemContent = fmt.Sprintf("%s: %s", itemName, obsLine)
								break
							}
						}

						if itemContent == "" {
							itemContent = itemName
						}
						break
					}
				}
			}

			if itemContent == "" {
				itemContent = fmt.Sprintf("Required Item %d", len(requiredItems)+1)
			}

			requiredItems = append(requiredItems, itemContent)
		} else if strings.Contains(line, "{set:cellbgcolor:#FEFE20}") &&
			!strings.Contains(line, "Indicates Changes Recommended") {
			// Similar logic for recommended items
			var itemContent string

			// Look for name in previous or next lines
			for j := i - 5; j <= i+5 && j < len(lines); j++ {
				if j >= 0 && strings.Contains(lines[j], "<<") && strings.Contains(lines[j], ">>") {
					nameMatch := regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(lines[j])
					if len(nameMatch) > 1 {
						itemName := nameMatch[1]

						// Look for observation text
						for k := j + 1; k < i; k++ {
							obsLine := strings.TrimSpace(lines[k])
							if obsLine != "" && !strings.Contains(obsLine, "set:cellbgcolor") {
								if strings.HasPrefix(obsLine, "|") {
									obsLine = strings.TrimSpace(obsLine[1:])
								}
								itemContent = fmt.Sprintf("%s: %s", itemName, obsLine)
								break
							}
						}

						if itemContent == "" {
							itemContent = itemName
						}
						break
					}
				}
			}

			if itemContent == "" {
				itemContent = fmt.Sprintf("Recommended Item %d", len(recommendedItems)+1)
			}

			recommendedItems = append(recommendedItems, itemContent)
		} else if strings.Contains(line, "{set:cellbgcolor:#80E5FF}") &&
			!strings.Contains(line, "No advise given") {
			// Similar logic for advisory items
			var itemContent string

			// Look for name in previous or next lines
			for j := i - 5; j <= i+5 && j < len(lines); j++ {
				if j >= 0 && strings.Contains(lines[j], "<<") && strings.Contains(lines[j], ">>") {
					nameMatch := regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(lines[j])
					if len(nameMatch) > 1 {
						itemName := nameMatch[1]

						// Look for observation text
						for k := j + 1; k < i; k++ {
							obsLine := strings.TrimSpace(lines[k])
							if obsLine != "" && !strings.Contains(obsLine, "set:cellbgcolor") {
								if strings.HasPrefix(obsLine, "|") {
									obsLine = strings.TrimSpace(obsLine[1:])
								}
								itemContent = fmt.Sprintf("%s: %s", itemName, obsLine)
								break
							}
						}

						if itemContent == "" {
							itemContent = itemName
						}
						break
					}
				}
			}

			if itemContent == "" {
				itemContent = fmt.Sprintf("Advisory Item %d", len(advisoryItems)+1)
			}

			advisoryItems = append(advisoryItems, itemContent)
		} else if strings.Contains(line, "{set:cellbgcolor:#00FF00}") &&
			!strings.Contains(line, "No change required") {
			noChangeCount++
		} else if strings.Contains(line, "{set:cellbgcolor:#A6B9BF}") &&
			!strings.Contains(line, "No advise given") {
			notApplicableCount++
		}
	}

	// Fill in the rest of the summary data
	summary.ClusterName = extractClusterName(lines)
	summary.CustomerName = extractCustomerName(lines)
	summary.OverallScore = extractOverallScore(lines)
	summary.ScoreInfra = extractCategoryScore(lines, "Infrastructure Setup")
	summary.ScoreGovernance = extractCategoryScore(lines, "Policy Governance")
	summary.ScoreCompliance = extractCategoryScore(lines, "Compliance Benchmarking")
	summary.ScoreMonitoring = extractCategoryScore(lines, "Central Monitoring")
	summary.ScoreBuildSecurity = extractCategoryScore(lines, "Build/Deploy Security")

	// Get or generate category descriptions
	summary.InfraDescription = extractCategoryDescription(lines, "Infrastructure Setup")
	summary.GovernanceDescription = extractCategoryDescription(lines, "Policy Governance")
	summary.ComplianceDescription = extractCategoryDescription(lines, "Compliance Benchmarking")
	summary.MonitoringDescription = extractCategoryDescription(lines, "Central Monitoring")
	summary.BuildSecurityDescription = extractCategoryDescription(lines, "Build/Deploy Security")

	// Set the action items
	summary.ItemsRequired = requiredItems
	summary.ItemsRecommended = recommendedItems
	summary.ItemsAdvisory = advisoryItems
	summary.NoChangeCount = noChangeCount

	return summary, nil
}

// Helper functions for parsing AsciiDoc files

// extractClusterName extracts the cluster name from the document
func extractClusterName(lines []string) string {
	for _, line := range lines {
		if strings.Contains(line, "cluster") {
			re := regexp.MustCompile(`['"]([^'"]+)['"]|cluster\s+([a-zA-Z0-9_-]+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if matches[1] != "" {
					return matches[1]
				}
				if len(matches) > 2 && matches[2] != "" {
					return matches[2]
				}
			}
		}
	}
	return ""
}

// extractCustomerName extracts the customer name from the document
func extractCustomerName(lines []string) string {
	for _, line := range lines {
		if strings.Contains(line, "conducted") && strings.Contains(line, "health check") {
			re := regexp.MustCompile(`conducted.*?([A-Za-z0-9_\s]+)'s`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				return strings.TrimSpace(matches[1])
			}
		}
	}
	return ""
}

// extractOverallScore extracts the overall score from the document
func extractOverallScore(lines []string) float64 {
	var score float64

	// Look for explicit score notation
	scorePattern := regexp.MustCompile(`Overall\s+Cluster\s+Health:\s+(\d+\.?\d*)%`)
	for _, line := range lines {
		matches := scorePattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			fmt.Sscanf(matches[1], "%f", &score)
			return score
		}
	}

	// Check for alternative score format
	altScorePattern := regexp.MustCompile(`Overall Health Score.*?(\d+\.?\d*)%`)
	for _, line := range lines {
		matches := altScorePattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			fmt.Sscanf(matches[1], "%f", &score)
			return score
		}
	}

	return score
}

// extractCategoryScore extracts the score for a specific category
func extractCategoryScore(lines []string, categoryName string) int {
	var score int

	// Look for category score in various formats
	scorePattern := regexp.MustCompile(fmt.Sprintf(`\*%s\*:\s+(\d+)%%`, regexp.QuoteMeta(categoryName)))
	for _, line := range lines {
		matches := scorePattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			fmt.Sscanf(matches[1], "%d", &score)
			return score
		}
	}

	// Try partial matching if exact match not found
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(categoryName)) && strings.Contains(line, "%") {
			re := regexp.MustCompile(`(\d+)%`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				fmt.Sscanf(matches[1], "%d", &score)
				return score
			}
		}
	}

	return score
}

// extractCategoryDescription extracts or generates a description for a category
func extractCategoryDescription(lines []string, categoryName string) string {
	// Try to find an actual description in the document
	for i, line := range lines {
		if strings.Contains(line, categoryName) {
			// Look for description in next few lines
			for j := i + 1; j < i+10 && j < len(lines); j++ {
				if j < len(lines) && lines[j] != "" &&
					!strings.HasPrefix(lines[j], "*") &&
					!strings.HasPrefix(lines[j], "#") &&
					!strings.Contains(lines[j], "%") {
					return strings.TrimSpace(lines[j])
				}
			}
		}
	}

	return ""
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

	log.Printf("Server starting on port %s", s.config.Port)

	// Start the server
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
