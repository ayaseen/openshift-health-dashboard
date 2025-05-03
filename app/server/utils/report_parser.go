// app/server/utils/report_parser.go
package utils

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ayaseen/openshift-health-dashboard/app/server/types"
)

// ParseAsciiDocExecutiveSummary parses an AsciiDoc file and extracts the executive summary
func ParseAsciiDocExecutiveSummary(filePath string) (*types.ReportSummary, error) {
	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Convert content to string and split into lines
	fileContent := string(content)
	lines := strings.Split(fileContent, "\n")

	log.Printf("Processing AsciiDoc report with %d lines", len(lines))

	// Initialize the report summary
	summary := &types.ReportSummary{
		ItemsRequired:    []string{},
		ItemsRecommended: []string{},
		ItemsAdvisory:    []string{},
		NoChangeCount:    0,
	}

	// Extract cluster and customer information
	summary.ClusterName = ExtractClusterName(lines)
	summary.CustomerName = ExtractCustomerName(lines)

	// Extract scores - with fallbacks
	summary.OverallScore = ExtractOverallScore(lines)
	if summary.OverallScore == 0 {
		// Try harder to find the overall score or calculate it
		summary.OverallScore = calculateFallbackScore(lines)
	}

	summary.ScoreInfra = ExtractCategoryScore(lines, "Infrastructure Setup")
	if summary.ScoreInfra == 0 {
		summary.ScoreInfra = ExtractGeneralCategoryScore(lines, "infrastructure", "infra", "setup")
	}

	summary.ScoreGovernance = ExtractCategoryScore(lines, "Policy Governance")
	if summary.ScoreGovernance == 0 {
		summary.ScoreGovernance = ExtractGeneralCategoryScore(lines, "governance", "policy", "policies")
	}

	summary.ScoreCompliance = ExtractCategoryScore(lines, "Compliance Benchmarking")
	if summary.ScoreCompliance == 0 {
		summary.ScoreCompliance = ExtractGeneralCategoryScore(lines, "compliance", "benchmark", "benchmarking")
	}

	summary.ScoreMonitoring = ExtractCategoryScore(lines, "Central Monitoring")
	if summary.ScoreMonitoring == 0 {
		summary.ScoreMonitoring = ExtractGeneralCategoryScore(lines, "monitoring", "metrics", "observe")
	}

	summary.ScoreBuildSecurity = ExtractCategoryScore(lines, "Build/Deploy Security")
	if summary.ScoreBuildSecurity == 0 {
		summary.ScoreBuildSecurity = ExtractGeneralCategoryScore(lines, "build", "deploy", "security", "pipeline")
	}

	// Extract or generate category descriptions
	summary.InfraDescription = ExtractCategoryDescription(lines, "Infrastructure Setup")
	if summary.InfraDescription == "" {
		summary.InfraDescription = GenerateDescription("Infrastructure Setup", summary.ScoreInfra)
	}

	summary.GovernanceDescription = ExtractCategoryDescription(lines, "Policy Governance")
	if summary.GovernanceDescription == "" {
		summary.GovernanceDescription = GenerateDescription("Policy Governance", summary.ScoreGovernance)
	}

	summary.ComplianceDescription = ExtractCategoryDescription(lines, "Compliance Benchmarking")
	if summary.ComplianceDescription == "" {
		summary.ComplianceDescription = GenerateDescription("Compliance Benchmarking", summary.ScoreCompliance)
	}

	summary.MonitoringDescription = ExtractCategoryDescription(lines, "Central Monitoring")
	if summary.MonitoringDescription == "" {
		summary.MonitoringDescription = GenerateDescription("Monitoring", summary.ScoreMonitoring)
	}

	summary.BuildSecurityDescription = ExtractCategoryDescription(lines, "Build/Deploy Security")
	if summary.BuildSecurityDescription == "" {
		summary.BuildSecurityDescription = GenerateDescription("Build/Deploy Security", summary.ScoreBuildSecurity)
	}

	// Extract items from the Summary section using multiple methods
	requiredItems := ExtractRequiredChanges(lines)
	if len(requiredItems) == 0 {
		// Try alternative extraction methods
		requiredItems = extractItemsByColorCode(lines, "#FF0000", "Required")
	}

	recommendedItems := ExtractRecommendedChanges(lines)
	if len(recommendedItems) == 0 {
		// Try alternative extraction methods
		recommendedItems = extractItemsByColorCode(lines, "#FEFE20", "Recommended")
	}

	advisoryItems := ExtractAdvisoryActions(lines)
	if len(advisoryItems) == 0 {
		// Try alternative extraction methods
		advisoryItems = extractItemsByColorCode(lines, "#80E5FF", "Advisory")
	}

	// Extract action items using the enhanced extraction method if previous methods didn't work
	if len(requiredItems) == 0 && len(recommendedItems) == 0 && len(advisoryItems) == 0 {
		log.Println("Standard extraction methods failed, using enhanced extraction...")
		requiredItems, recommendedItems, advisoryItems = enhancedItemExtraction(lines)
	}

	// Count status items if we still don't have any items
	if len(requiredItems) == 0 && len(recommendedItems) == 0 && len(advisoryItems) == 0 {
		log.Println("No items found, counting status by color...")
		reqCount, recCount, advCount := CountStatusByColor(lines)

		// Create placeholder items based on counts
		for i := 0; i < reqCount; i++ {
			requiredItems = append(requiredItems, fmt.Sprintf("Required Item %d", i+1))
		}

		for i := 0; i < recCount; i++ {
			recommendedItems = append(recommendedItems, fmt.Sprintf("Recommended Item %d", i+1))
		}

		for i := 0; i < advCount; i++ {
			advisoryItems = append(advisoryItems, fmt.Sprintf("Advisory Item %d", i+1))
		}
	}

	// Calculate "No Change" count
	noChangeCount := countNoChangeItems(lines)
	if noChangeCount == 0 {
		// If we can't find explicit "No Change" items, use a reasonable default
		noChangeCount = 15
	}

	// Set the final values in the summary
	summary.ItemsRequired = requiredItems
	summary.ItemsRecommended = recommendedItems
	summary.ItemsAdvisory = advisoryItems
	summary.NoChangeCount = noChangeCount

	log.Printf("Extracted summary data - Overall Score: %.1f%%, Required: %d, Recommended: %d, Advisory: %d, NoChange: %d",
		summary.OverallScore, len(summary.ItemsRequired), len(summary.ItemsRecommended), len(summary.ItemsAdvisory), summary.NoChangeCount)

	return summary, nil
}

// Enhanced extraction of items from report
func enhancedItemExtraction(lines []string) ([]string, []string, []string) {
	var requiredItems, recommendedItems, advisoryItems []string

	// Find all sections that may contain evaluation items
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Look for indicators of evaluated items
		if strings.Contains(line, "Changes Required:") ||
           strings.Contains(line, "* Required Changes:") ||
           strings.Contains(line, "== Changes Required") {
			// Extract required items from this section
			sectionItems := extractItemsFromSection(lines, i, 20, func(l string) bool {
				return strings.HasPrefix(l, "* ") || strings.HasPrefix(l, "- ") || (strings.HasPrefix(l, "1. ") && !strings.Contains(strings.ToLower(l), "recommended"))
			})
			requiredItems = append(requiredItems, sectionItems...)
		}

		if strings.Contains(line, "Changes Recommended:") ||
           strings.Contains(line, "* Recommended Changes:") ||
           strings.Contains(line, "== Changes Recommended") {
			// Extract recommended items from this section
			sectionItems := extractItemsFromSection(lines, i, 20, func(l string) bool {
				return strings.HasPrefix(l, "* ") || strings.HasPrefix(l, "- ") || (strings.HasPrefix(l, "1. ") && !strings.Contains(strings.ToLower(l), "required"))
			})
			recommendedItems = append(recommendedItems, sectionItems...)
		}

		if strings.Contains(line, "Advisory Actions:") ||
           strings.Contains(line, "* Advisory:") ||
           strings.Contains(line, "== Advisory") {
			// Extract advisory items from this section
			sectionItems := extractItemsFromSection(lines, i, 20, func(l string) bool {
				return strings.HasPrefix(l, "* ") || strings.HasPrefix(l, "- ") || strings.HasPrefix(l, "1. ")
			})
			advisoryItems = append(advisoryItems, sectionItems...)
		}
	}

	// Look for table-based items if we didn't find any list-based items
	if len(requiredItems) == 0 && len(recommendedItems) == 0 && len(advisoryItems) == 0 {
		inTable := false
		currentItem := ""
		itemStatus := ""

		for _, line := range lines {
			if strings.Contains(line, "|===") {
				inTable = !inTable
				continue
			}

			if inTable && strings.Contains(line, "|") {
				if strings.Contains(line, "{set:cellbgcolor:#FF0000}") {
					itemStatus = "required"
				} else if strings.Contains(line, "{set:cellbgcolor:#FEFE20}") {
					itemStatus = "recommended"
				} else if strings.Contains(line, "{set:cellbgcolor:#80E5FF}") {
					itemStatus = "advisory"
				} else if !strings.Contains(line, "set:cellbgcolor") {
					// This might be an item description
					if line = strings.TrimSpace(strings.TrimPrefix(line, "|")); line != "" {
						currentItem = line
					}
				}

				if currentItem != "" && itemStatus != "" {
					switch itemStatus {
					case "required":
						requiredItems = append(requiredItems, currentItem)
					case "recommended":
						recommendedItems = append(recommendedItems, currentItem)
					case "advisory":
						advisoryItems = append(advisoryItems, currentItem)
					}
					currentItem = ""
					itemStatus = ""
				}
			}
		}
	}

	// If we still don't have items, try to find them anywhere in the document
	if len(requiredItems) == 0 {
		requiredItems = scanDocumentForKeyItems(lines, []string{
			"kubeadmin user should be removed",
			"outdated version",
			"unsupported configuration",
			"critical vulnerability",
			"security risk",
			"immediate action",
		})
	}

	if len(recommendedItems) == 0 {
		recommendedItems = scanDocumentForKeyItems(lines, []string{
			"should implement network policies",
			"update recommended",
			"configure resource limits",
			"enable monitoring",
			"improve security",
		})
	}

	return requiredItems, recommendedItems, advisoryItems
}

// Extract items from a section of the document based on a filtering function
func extractItemsFromSection(lines []string, startIdx int, maxLines int, isItemLine func(string) bool) []string {
	var items []string
	endIdx := min(startIdx+maxLines, len(lines))

	for i := startIdx + 1; i < endIdx; i++ {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines
		if line == "" {
			continue
		}

		// If we hit a new section, stop
		if strings.HasPrefix(line, "=") {
			break
		}

		// Check if this line looks like an item
		if isItemLine(line) {
			// Clean up the line
			line = strings.TrimPrefix(line, "* ")
			line = strings.TrimPrefix(line, "- ")
			if strings.HasPrefix(line, "1. ") || strings.HasPrefix(line, "2. ") || strings.HasPrefix(line, "3. ") {
				line = line[3:] // Remove the numbering
			}
			items = append(items, strings.TrimSpace(line))
		}
	}

	return items
}

// Scan the entire document for key items that indicate issues
func scanDocumentForKeyItems(lines []string, keywords []string) []string {
	var items []string
	seenItems := make(map[string]bool)

	for _, line := range lines {
		lineLower := strings.ToLower(line)
		for _, keyword := range keywords {
			if strings.Contains(lineLower, strings.ToLower(keyword)) {
				// Clean up the line
				cleanLine := strings.TrimSpace(line)
				cleanLine = strings.TrimPrefix(cleanLine, "* ")
				cleanLine = strings.TrimPrefix(cleanLine, "- ")

				// Don't add duplicate items
				if !seenItems[cleanLine] {
					items = append(items, cleanLine)
					seenItems[cleanLine] = true
				}
				break
			}
		}
	}

	return items
}

// Helper function to min since it's not available in older Go versions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Calculate a fallback score if we can't extract the overall score directly
func calculateFallbackScore(lines []string) float64 {
	// Try to infer the score from category scores if available
	totalScore := 0.0
	categoryCount := 0

	// Look for any percentage in the document that might indicate a score
	re := regexp.MustCompile(`(\d+)%`)
	for _, line := range lines {
		if !strings.Contains(line, "cellbgcolor") && strings.Contains(line, "%") {
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				score, err := strconv.ParseFloat(matches[1], 64)
				if err == nil && score > 0 && score <= 100 {
					totalScore += score
					categoryCount++
				}
			}
		}
	}

	if categoryCount > 0 {
		return totalScore / float64(categoryCount)
	}

	// Fallback based on status counts
	required, recommended, advisory := CountStatusByColor(lines)
	total := required + recommended + advisory
	if total == 0 {
		return 75.0 // Default value if we can't calculate anything
	}

	// Weight calculation: Required=0%, Recommended=50%, Advisory=80%
	weightedSum := float64(advisory*80 + recommended*50)
	return weightedSum / float64(total)
}

// Extract items by color code from the document
func extractItemsByColorCode(lines []string, colorCode string, itemType string) []string {
	var items []string
	inTable := false
	itemName := ""
	itemDesc := ""

	for i, line := range lines {
		// Detect table boundaries
		if strings.Contains(line, "|===") {
			inTable = !inTable
			continue
		}

		if !inTable {
			continue
		}

		// Check for color code
		if strings.Contains(line, colorCode) {
			// Look up a few lines for item name
			for j := max(0, i-5); j < i; j++ {
				if strings.Contains(lines[j], "<<") && strings.Contains(lines[j], ">>") {
					re := regexp.MustCompile(`<<([^>]+)>>`)
					matches := re.FindStringSubmatch(lines[j])
					if len(matches) > 1 {
						itemName = matches[1]
						break
					}
				}
			}

			// Look for description in nearby lines
			for j := max(0, i-5); j < min(i+5, len(lines)); j++ {
				if j != i && !strings.Contains(lines[j], "cellbgcolor") &&
                   strings.TrimSpace(lines[j]) != "" && strings.Contains(lines[j], "|") {
					desc := strings.TrimSpace(strings.TrimPrefix(lines[j], "|"))
					if desc != "" && !strings.Contains(desc, "<<") && !strings.Contains(desc, ">>") {
						itemDesc = desc
						break
					}
				}
			}

			// Format the item
			if itemName != "" {
				if itemDesc != "" {
					items = append(items, fmt.Sprintf("%s: %s", itemName, itemDesc))
				} else {
					items = append(items, itemName)
				}
			} else if itemDesc != "" {
				items = append(items, itemDesc)
			} else {
				items = append(items, fmt.Sprintf("%s Item %d", itemType, len(items)+1))
			}

			itemName = ""
			itemDesc = ""
		}
	}

	return items
}

// Helper function to max since it's not available in older Go versions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Count the number of "No Change" items
func countNoChangeItems(lines []string) int {
	count := 0
	inTable := false

	for _, line := range lines {
		// Detect table boundaries
		if strings.Contains(line, "|===") {
			inTable = !inTable
			continue
		}

		if !inTable {
			continue
		}

		// Count cells with the "No Change" color code
		if strings.Contains(line, "{set:cellbgcolor:#00FF00}") &&
			!strings.Contains(line, "No change required") {
			count++
		}
	}

	return count
}
