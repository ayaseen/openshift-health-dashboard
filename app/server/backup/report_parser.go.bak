// app/server/utils/report_parser.go
package utils

import (
	"fmt"
	"log"
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
	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Convert content to string and split into lines
	fileContent := string(content)
	lines := strings.Split(fileContent, "\n")

	log.Printf("Processing AsciiDoc report with %d lines", len(lines))

	// Find the Summary section
	summaryStartIndex := -1
	summaryEndIndex := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == "= Summary" {
			summaryStartIndex = i
			break
		}
	}

	if summaryStartIndex == -1 {
		// No Summary section found
		log.Printf("No Summary section found in the document")
		return &types.ReportSummary{
			ItemsRequired:    []string{},
			ItemsRecommended: []string{},
			ItemsAdvisory:    []string{},
			NoChangeCount:    0,
		}, nil
	}

	// Find end of Summary section (next section heading or end of file)
	for i := summaryStartIndex + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "=") &&
			!strings.Contains(lines[i], "= Summary") {
			summaryEndIndex = i
			break
		}
	}

	if summaryEndIndex == -1 {
		summaryEndIndex = len(lines) // Use end of file if no next section
	}

	// Get only the Summary section content
	summaryLines := lines[summaryStartIndex:summaryEndIndex]
	summaryContent := strings.Join(summaryLines, "\n")

	// Extract all items by their status
	var requiredItems, recommendedItems, advisoryItems []string
	var noChangeCount int

	// Split the Summary section into ITEM blocks and process each
	itemBlocks := regexp.MustCompile(`(?s)// ------------------------ITEM START(.*?)// ------------------------ITEM END`).FindAllStringSubmatch(summaryContent, -1)
	log.Printf("Found %d ITEM blocks to process in Summary section", len(itemBlocks))

	for _, block := range itemBlocks {
		if len(block) < 2 {
			continue
		}

		itemContent := block[1]

		// Extract the item name
		itemNameMatch := regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(itemContent)
		itemName := ""
		if len(itemNameMatch) > 1 {
			itemName = strings.TrimSpace(itemNameMatch[1])
		}

		// Extract the observation (everything after the item name until the status)
		observationMatch := regexp.MustCompile(`(?s)\|\s*(.*?)(\{set:cellbgcolor:#[A-Fa-f0-9]+\})`).FindStringSubmatch(itemContent)
		observation := ""
		if len(observationMatch) > 1 {
			observation = strings.TrimSpace(observationMatch[1])
			// Clean up the observation - remove pipe characters and extra whitespace
			observation = regexp.MustCompile(`\|\s*`).ReplaceAllString(observation, "")
			observation = strings.TrimSpace(observation)
		}

		// Format the item text
		itemText := itemName
		if observation != "" {
			itemText = fmt.Sprintf("%s: %s", itemName, observation)
		}

		// Determine the status based on color code
		if strings.Contains(itemContent, "{set:cellbgcolor:#FF0000}") &&
			!strings.Contains(itemContent, "Indicates Changes Required") { // Required
			requiredItems = append(requiredItems, itemText)
		} else if strings.Contains(itemContent, "{set:cellbgcolor:#FEFE20}") &&
			!strings.Contains(itemContent, "Indicates Changes Recommended") { // Recommended
			recommendedItems = append(recommendedItems, itemText)
		} else if strings.Contains(itemContent, "{set:cellbgcolor:#80E5FF}") &&
			!strings.Contains(itemContent, "No advise given") { // Advisory
			advisoryItems = append(advisoryItems, itemText)
		} else if strings.Contains(itemContent, "{set:cellbgcolor:#00FF00}") &&
			!strings.Contains(itemContent, "No change required") { // No Change
			noChangeCount++
		}
	}

	// If we couldn't extract items from ITEM blocks, try direct color code counting
	if len(requiredItems) == 0 && len(recommendedItems) == 0 && len(advisoryItems) == 0 && noChangeCount == 0 {
		log.Printf("No items found in ITEM blocks, trying direct color code counting")

		inTable := false

		for _, line := range summaryLines {
			// Detect start of table
			if strings.Contains(line, "|===") && !inTable {
				inTable = true
				continue
			}

			// Detect end of table
			if strings.Contains(line, "|===") && inTable {
				break
			}

			// Skip header/legend rows
			if inTable && (strings.Contains(line, "*Category*") ||
				strings.Contains(line, "Indicates Changes Required") ||
				strings.Contains(line, "Indicates Changes Recommended") ||
				strings.Contains(line, "No advise given") ||
				strings.Contains(line, "No change required") ||
				strings.Contains(line, "Not yet evaluated")) {
				continue
			}

			// Extract item details and count by color codes
			if inTable {
				if strings.Contains(line, "{set:cellbgcolor:#FF0000}") &&
					!strings.Contains(line, "Indicates Changes Required") {

					// Try to extract item name and description from nearby lines
					requiredItems = append(requiredItems, fmt.Sprintf("Required Item %d", len(requiredItems)+1))
				} else if strings.Contains(line, "{set:cellbgcolor:#FEFE20}") &&
					!strings.Contains(line, "Indicates Changes Recommended") {

					recommendedItems = append(recommendedItems, fmt.Sprintf("Recommended Item %d", len(recommendedItems)+1))
				} else if strings.Contains(line, "{set:cellbgcolor:#80E5FF}") &&
					!strings.Contains(line, "No advise given") {

					advisoryItems = append(advisoryItems, fmt.Sprintf("Advisory Item %d", len(advisoryItems)+1))
				} else if strings.Contains(line, "{set:cellbgcolor:#00FF00}") &&
					!strings.Contains(line, "No change required") {

					noChangeCount++
				}
			}
		}
	}

	log.Printf("Extracted items - Required: %d, Recommended: %d, Advisory: %d, No Change: %d",
		len(requiredItems), len(recommendedItems), len(advisoryItems), noChangeCount)

	// Extract cluster and customer information
	clusterName := ExtractClusterName(lines)
	customerName := ExtractCustomerName(lines)

	// Extract scores
	overallScore := ExtractOverallScore(lines)
	scoreInfra := ExtractCategoryScore(lines, "Infrastructure Setup")
	scoreGovernance := ExtractCategoryScore(lines, "Policy Governance")
	scoreCompliance := ExtractCategoryScore(lines, "Compliance Benchmarking")
	scoreMonitoring := ExtractCategoryScore(lines, "Central Monitoring")
	scoreBuildSecurity := ExtractCategoryScore(lines, "Build/Deploy Security")

	// Extract or generate category descriptions
	infraDescription := ExtractCategoryDescription(lines, "Infrastructure Setup")
	governanceDescription := ExtractCategoryDescription(lines, "Policy Governance")
	complianceDescription := ExtractCategoryDescription(lines, "Compliance Benchmarking")
	monitoringDescription := ExtractCategoryDescription(lines, "Central Monitoring")
	buildSecurityDescription := ExtractCategoryDescription(lines, "Build/Deploy Security")

	// Create the summary
	summary := &types.ReportSummary{
		ClusterName:              clusterName,
		CustomerName:             customerName,
		OverallScore:             overallScore,
		ScoreInfra:               scoreInfra,
		ScoreGovernance:          scoreGovernance,
		ScoreCompliance:          scoreCompliance,
		ScoreMonitoring:          scoreMonitoring,
		ScoreBuildSecurity:       scoreBuildSecurity,
		InfraDescription:         infraDescription,
		GovernanceDescription:    governanceDescription,
		ComplianceDescription:    complianceDescription,
		MonitoringDescription:    monitoringDescription,
		BuildSecurityDescription: buildSecurityDescription,
		ItemsRequired:            requiredItems,
		ItemsRecommended:         recommendedItems,
		ItemsAdvisory:            advisoryItems,
		NoChangeCount:            noChangeCount,
	}

	return summary, nil
}
