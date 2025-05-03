// app/server/utils/report_parser.go
package utils

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ayaseen/openshift-health-dashboard/app/server/types"
)

// ReportItem represents a single item from the health check report
type ReportItem struct {
	Category       string
	ItemName       string
	Observation    string
	Recommendation string
	Status         string // No Change, Changes Required, Changes Recommended, Not Applicable, Advisory
	ColorCode      string // Color code from the report
}

// ParseAsciiDocExecutiveSummary parses an AsciiDoc file and extracts the executive summary
func ParseAsciiDocExecutiveSummary(filePath string) (*types.ReportSummary, error) {
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Parse all report items
	reportItems := ParseAllReportItems(lines)

	// Extract items by status
	var requiredItems, recommendedItems, advisoryItems []string

	for _, item := range reportItems {
		formattedItem := fmt.Sprintf("%s: %s", item.ItemName, item.Observation)

		switch item.Status {
		case "Changes Required":
			requiredItems = append(requiredItems, formattedItem)
		case "Changes Recommended":
			recommendedItems = append(recommendedItems, formattedItem)
		case "Advisory":
			advisoryItems = append(advisoryItems, formattedItem)
		}
	}

	// Extract overall score
	overallScore := ExtractOverallScore(lines)

	// Initialize the summary with extracted data
	summary := &types.ReportSummary{
		ClusterName:              ExtractClusterName(lines),
		CustomerName:             ExtractCustomerName(lines),
		OverallScore:             overallScore,
		ScoreInfra:               ExtractCategoryScore(lines, "Infrastructure Setup"),
		ScoreGovernance:          ExtractCategoryScore(lines, "Policy Governance"),
		ScoreCompliance:          ExtractCategoryScore(lines, "Compliance Benchmarking"),
		ScoreMonitoring:          ExtractCategoryScore(lines, "Central Monitoring"),
		ScoreBuildSecurity:       ExtractCategoryScore(lines, "Build/Deploy Security"),
		InfraDescription:         ExtractCategoryDescription(lines, "Infrastructure Setup"),
		GovernanceDescription:    ExtractCategoryDescription(lines, "Policy Governance"),
		ComplianceDescription:    ExtractCategoryDescription(lines, "Compliance Benchmarking"),
		MonitoringDescription:    ExtractCategoryDescription(lines, "Central Monitoring"),
		BuildSecurityDescription: ExtractCategoryDescription(lines, "Build/Deploy Security"),
		ItemsRequired:            requiredItems,
		ItemsRecommended:         recommendedItems,
		ItemsAdvisory:            advisoryItems,
	}

	// Set default category scores if not found
	if summary.ScoreInfra == 0 {
		summary.ScoreInfra = 75
	}
	if summary.ScoreGovernance == 0 {
		summary.ScoreGovernance = 75
	}
	if summary.ScoreCompliance == 0 {
		summary.ScoreCompliance = 75
	}
	if summary.ScoreMonitoring == 0 {
		summary.ScoreMonitoring = 75
	}
	if summary.ScoreBuildSecurity == 0 {
		summary.ScoreBuildSecurity = 75
	}

	// Ensure descriptions are present
	if summary.InfraDescription == "" {
		summary.InfraDescription = GenerateDescription("Infrastructure", summary.ScoreInfra)
	}
	if summary.GovernanceDescription == "" {
		summary.GovernanceDescription = GenerateDescription("Policy Governance", summary.ScoreGovernance)
	}
	if summary.ComplianceDescription == "" {
		summary.ComplianceDescription = GenerateDescription("Compliance", summary.ScoreCompliance)
	}
	if summary.MonitoringDescription == "" {
		summary.MonitoringDescription = GenerateDescription("Monitoring", summary.ScoreMonitoring)
	}
	if summary.BuildSecurityDescription == "" {
		summary.BuildSecurityDescription = GenerateDescription("Build/Deploy Security", summary.ScoreBuildSecurity)
	}

	// Logging for debugging purposes
	fmt.Printf("Found %d required items, %d recommended items, %d advisory items\n",
		len(summary.ItemsRequired), len(summary.ItemsRecommended), len(summary.ItemsAdvisory))

	return summary, nil
}

// ParseAllReportItems extracts all items from the report
func ParseAllReportItems(lines []string) []ReportItem {
	var items []ReportItem

	// Map of color codes to statuses
	colorStatusMap := map[string]string{
		"#FF0000": "Changes Required",
		"#FEFE20": "Changes Recommended",
		"#80E5FF": "Advisory",
		"#00FF00": "No Change",
		"#A6B9BF": "Not Applicable",
		"#FFFFFF": "To Be Evaluated",
	}

	// Find the table with report items
	tableStartIndex := -1
	tableEndIndex := -1

	// Look for the table header
	for i, line := range lines {
		if strings.Contains(line, "[cols=") &&
			strings.Contains(line, "options=header") &&
			i+1 < len(lines) && strings.Contains(lines[i+1], "|===") {
			tableStartIndex = i + 2 // +2 to skip the header row
		}

		// Find table end
		if tableStartIndex != -1 && strings.Contains(line, "|===") && i > tableStartIndex {
			tableEndIndex = i
			break
		}
	}

	// If we couldn't find the table properly, look for other indicators
	if tableStartIndex == -1 || tableEndIndex == -1 {
		// Try to find table boundaries by looking for column headers
		for i, line := range lines {
			if strings.Contains(line, "|*Category*") &&
				strings.Contains(line, "|*Item Evaluated*") &&
				strings.Contains(line, "|*Recommendation*") {
				tableStartIndex = i + 1 // +1 to skip the header row
			}
		}

		// Look for end marker or just set to end of file
		if tableStartIndex != -1 {
			for i := tableStartIndex; i < len(lines); i++ {
				if strings.Contains(lines[i], "|===") {
					tableEndIndex = i
					break
				}
			}

			// If we still couldn't find end, assume it goes to the end
			if tableEndIndex == -1 {
				tableEndIndex = len(lines) - 1
			}
		}
	}

	// If we found the table, process it
	if tableStartIndex != -1 && tableEndIndex != -1 {
		var currentItem ReportItem
		var itemStarted bool

		for i := tableStartIndex; i < tableEndIndex; i++ {
			line := strings.TrimSpace(lines[i])

			// Skip empty lines
			if line == "" {
				continue
			}

			// Check for category
			if strings.Contains(line, "{set:cellbgcolor!}") && !itemStarted {
				currentItem = ReportItem{} // Start a new item
				currentItem.Category = strings.TrimSpace(strings.Replace(line, "{set:cellbgcolor!}", "", -1))
				itemStarted = true
				continue
			}

			// Check for item name
			if strings.Contains(line, "<<") && strings.Contains(line, ">>") && itemStarted {
				itemMatches := regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(line)
				if len(itemMatches) > 1 {
					currentItem.ItemName = strings.TrimSpace(itemMatches[1])
				}
				continue
			}

			// Check for observation
			if line != "" && !strings.Contains(line, "{set:cellbgcolor") && itemStarted && currentItem.Observation == "" {
				// Clean up the line
				if strings.HasPrefix(line, "|") {
					line = strings.TrimSpace(line[1:])
				}

				if line != "" {
					currentItem.Observation = line
				}
				continue
			}

			// Check for recommendation status
			if strings.Contains(line, "{set:cellbgcolor:") && itemStarted {
				// Extract color code
				colorMatches := regexp.MustCompile(`{set:cellbgcolor:(#[A-Fa-f0-9]+)}`).FindStringSubmatch(line)
				if len(colorMatches) > 1 {
					colorCode := strings.ToUpper(colorMatches[1])
					currentItem.ColorCode = colorCode

					// Look up status from color
					if status, ok := colorStatusMap[colorCode]; ok {
						currentItem.Status = status
					} else {
						currentItem.Status = "Unknown"
					}

					// If this is a complete item, add it to results
					if currentItem.ItemName != "" {
						items = append(items, currentItem)
					}

					// Reset for next item
					itemStarted = false
				}
				continue
			}
		}
	} else {
		// Fallback approach: try to find items based on color codes directly
		for i, line := range lines {
			// Skip the legend section
			if strings.Contains(line, "Value") &&
				strings.Contains(line, "Description") {
				continue
			}

			// Look for color codes that indicate item status
			if strings.Contains(line, "{set:cellbgcolor:") {
				colorMatches := regexp.MustCompile(`{set:cellbgcolor:(#[A-Fa-f0-9]+)}`).FindStringSubmatch(line)
				if len(colorMatches) > 1 {
					colorCode := strings.ToUpper(colorMatches[1])

					// Skip if this is part of the legend
					if strings.Contains(line, "Changes Required") &&
						strings.Contains(lines[i+1], "Indicates Changes Required") {
						continue
					}

					if strings.Contains(line, "Changes Recommended") &&
						strings.Contains(lines[i+1], "Indicates Changes Recommended") {
						continue
					}

					// Try to find associated item name, observation
					var item ReportItem
					item.ColorCode = colorCode

					// Set status based on color
					if status, ok := colorStatusMap[colorCode]; ok {
						item.Status = status
					} else {
						item.Status = "Unknown"
					}

					// Look for item name in previous lines
					for j := i - 10; j < i; j++ {
						if j >= 0 && strings.Contains(lines[j], "<<") && strings.Contains(lines[j], ">>") {
							itemMatches := regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(lines[j])
							if len(itemMatches) > 1 {
								item.ItemName = strings.TrimSpace(itemMatches[1])
								break
							}
						}
					}

					// Look for observation in previous lines
					for j := i - 5; j < i; j++ {
						if j >= 0 && lines[j] != "" &&
							!strings.Contains(lines[j], "<<") &&
							!strings.Contains(lines[j], "{set:cellbgcolor") {

							observation := strings.TrimSpace(lines[j])
							if strings.HasPrefix(observation, "|") {
								observation = strings.TrimSpace(observation[1:])
							}

							if observation != "" {
								item.Observation = observation
								break
							}
						}
					}

					// Look for category in previous lines
					for j := i - 15; j < i; j++ {
						if j >= 0 && strings.Contains(lines[j], "Category") ||
							(strings.Contains(lines[j], "{set:cellbgcolor!}") &&
								lines[j] != "{set:cellbgcolor!}") {

							category := strings.TrimSpace(lines[j])
							if strings.HasPrefix(category, "|") {
								category = strings.TrimSpace(category[1:])
							}
							category = strings.Replace(category, "{set:cellbgcolor!}", "", -1)

							if category != "" && category != "Category" {
								item.Category = category
								break
							}
						}
					}

					// Add item if we have enough information
					if item.ItemName != "" {
						items = append(items, item)
					}
				}
			}
		}
	}

	// Deduplicate items
	return deduplicateItems(items)
}

// Deduplicate items by item name and status
func deduplicateItems(items []ReportItem) []ReportItem {
	unique := make(map[string]ReportItem)

	for _, item := range items {
		key := fmt.Sprintf("%s-%s", item.ItemName, item.Status)
		unique[key] = item
	}

	result := make([]ReportItem, 0, len(unique))
	for _, item := range unique {
		result = append(result, item)
	}

	return result
}

// extractBasicReport tries to extract basic information from any health check report
func extractBasicReport(lines []string) (*types.ReportSummary, error) {
	// Initialize with default values
	summary := &types.ReportSummary{
		ClusterName:        "Unknown Cluster",
		CustomerName:       "Unknown Customer",
		OverallScore:       CalculateScoreFromStatusCounts(lines),
		ScoreInfra:         ExtractGeneralCategoryScore(lines, "Cluster", "Infrastructure", "Infra"),
		ScoreGovernance:    ExtractGeneralCategoryScore(lines, "Security", "Policy", "Governance"),
		ScoreCompliance:    ExtractGeneralCategoryScore(lines, "Compliance", "Benchmarking"),
		ScoreMonitoring:    ExtractGeneralCategoryScore(lines, "Monitoring", "Logging", "Observability"),
		ScoreBuildSecurity: ExtractGeneralCategoryScore(lines, "Build", "Deploy", "Application", "App"),
	}

	// Get items from the report
	reportItems := ParseAllReportItems(lines)

	// Extract items by status
	var requiredItems, recommendedItems, advisoryItems []string

	for _, item := range reportItems {
		formattedItem := fmt.Sprintf("%s: %s", item.ItemName, item.Observation)

		switch item.Status {
		case "Changes Required":
			requiredItems = append(requiredItems, formattedItem)
		case "Changes Recommended":
			recommendedItems = append(recommendedItems, formattedItem)
		case "Advisory":
			advisoryItems = append(advisoryItems, formattedItem)
		}
	}

	summary.ItemsRequired = requiredItems
	summary.ItemsRecommended = recommendedItems
	summary.ItemsAdvisory = advisoryItems

	// Generate basic descriptions based on scores
	summary.InfraDescription = GenerateDescription("Infrastructure", summary.ScoreInfra)
	summary.GovernanceDescription = GenerateDescription("Policy Governance", summary.ScoreGovernance)
	summary.ComplianceDescription = GenerateDescription("Compliance", summary.ScoreCompliance)
	summary.MonitoringDescription = GenerateDescription("Monitoring and Logging", summary.ScoreMonitoring)
	summary.BuildSecurityDescription = GenerateDescription("Build/Deploy Security", summary.ScoreBuildSecurity)

	return summary, nil
}
