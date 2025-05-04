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
		ItemsRequired:      []string{},
		ItemsRecommended:   []string{},
		ItemsAdvisory:      []string{},
		NoChangeCount:      0,
		NotApplicableCount: 0,
	}

	// Extract cluster and customer information
	summary.ClusterName = ExtractClusterName(lines)
	summary.CustomerName = ExtractCustomerName(lines)

	// Count items by status and category
	required, recommended, advisory, noChange, notApplicable := CountAllStatusItems(lines)

	// Set item counts
	summary.NoChangeCount = noChange
	summary.NotApplicableCount = notApplicable

	// Calculate overall score - exclude Not Applicable items from the calculation
	totalValidItems := required + recommended + advisory + noChange
	if totalValidItems > 0 {
		weightedSum := float64(noChange*100 + advisory*80 + recommended*50)
		summary.OverallScore = weightedSum / float64(totalValidItems)
	} else {
		summary.OverallScore = 0
	}

	// Calculate category scores
	categoryItems := CountStatusByCategory(lines)

	// Set category scores based on actual item counts by category
	// Infrastructure Setup
	infraItems := make(map[string]int)
	infraItems["required"] = categoryItemCount(categoryItems.Required, "Cluster Config")
	infraItems["recommended"] = categoryItemCount(categoryItems.Recommended, "Cluster Config")
	infraItems["advisory"] = categoryItemCount(categoryItems.Advisory, "Cluster Config")
	infraItems["nochange"] = categoryItemCount(categoryItems.NoChange, "Cluster Config")
	summary.ScoreInfra = CalculateCategoryScore(infraItems, "Infrastructure Setup")

	// Policy Governance
	govItems := make(map[string]int)
	govItems["required"] = categoryItemCount(categoryItems.Required, "Security")
	govItems["recommended"] = categoryItemCount(categoryItems.Recommended, "Security")
	govItems["advisory"] = categoryItemCount(categoryItems.Advisory, "Security")
	govItems["nochange"] = categoryItemCount(categoryItems.NoChange, "Security")
	summary.ScoreGovernance = CalculateCategoryScore(govItems, "Policy Governance")

	// Compliance Benchmarking
	compItems := make(map[string]int)
	compItems["required"] = 0 // Direct compliance items are less common
	compItems["recommended"] = categoryItemCount(categoryItems.Recommended, "Performance")
	compItems["advisory"] = categoryItemCount(categoryItems.Advisory, "Performance")
	compItems["nochange"] = categoryItemCount(categoryItems.NoChange, "Performance")
	summary.ScoreCompliance = CalculateCategoryScore(compItems, "Compliance Benchmarking")

	// Monitoring
	monItems := make(map[string]int)
	monItems["required"] = 0
	monItems["recommended"] = categoryItemCount(categoryItems.Recommended, "Op-Ready")
	monItems["advisory"] = categoryItemCount(categoryItems.Advisory, "Op-Ready")
	monItems["nochange"] = categoryItemCount(categoryItems.NoChange, "Op-Ready")
	summary.ScoreMonitoring = CalculateCategoryScore(monItems, "Monitoring")

	// Build/Deploy Security
	buildItems := make(map[string]int)
	buildItems["required"] = 0
	buildItems["recommended"] = categoryItemCount(categoryItems.Recommended, "Applications")
	buildItems["advisory"] = categoryItemCount(categoryItems.Advisory, "Applications")
	buildItems["nochange"] = categoryItemCount(categoryItems.NoChange, "Applications")
	summary.ScoreBuildSecurity = CalculateCategoryScore(buildItems, "Build/Deploy Security")

	// If calculated scores are still 0, try falling back to extracted scores
	if summary.ScoreInfra == 0 {
		summary.ScoreInfra = ExtractCategoryScore(lines, "Infrastructure Setup")
	}
	if summary.ScoreGovernance == 0 {
		summary.ScoreGovernance = ExtractCategoryScore(lines, "Policy Governance")
	}
	if summary.ScoreCompliance == 0 {
		summary.ScoreCompliance = ExtractCategoryScore(lines, "Compliance Benchmarking")
	}
	if summary.ScoreMonitoring == 0 {
		summary.ScoreMonitoring = ExtractCategoryScore(lines, "Central Monitoring")
		if summary.ScoreMonitoring == 0 {
			summary.ScoreMonitoring = ExtractCategoryScore(lines, "Monitoring")
		}
	}
	if summary.ScoreBuildSecurity == 0 {
		summary.ScoreBuildSecurity = ExtractCategoryScore(lines, "Build/Deploy Security")
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

	// Extract items from the Summary section
	summary.ItemsRequired = ExtractRequiredChanges(lines)
	summary.ItemsRecommended = ExtractRecommendedChanges(lines)
	summary.ItemsAdvisory = ExtractAdvisoryActions(lines)

	// If we have no items, use counts to create placeholder items
	if len(summary.ItemsRequired) == 0 && required > 0 {
		for i := 0; i < required; i++ {
			summary.ItemsRequired = append(summary.ItemsRequired, fmt.Sprintf("Required Item %d", i+1))
		}
	}

	if len(summary.ItemsRecommended) == 0 && recommended > 0 {
		for i := 0; i < recommended; i++ {
			summary.ItemsRecommended = append(summary.ItemsRecommended, fmt.Sprintf("Recommended Item %d", i+1))
		}
	}

	if len(summary.ItemsAdvisory) == 0 && advisory > 0 {
		for i := 0; i < advisory; i++ {
			summary.ItemsAdvisory = append(summary.ItemsAdvisory, fmt.Sprintf("Advisory Item %d", i+1))
		}
	}

	// Count "No Change" items if needed
	if summary.NoChangeCount == 0 {
		summary.NoChangeCount = CountNoChangeItems(lines)
	}

	log.Printf("Extracted summary data - Overall Score: %.1f%%, Required: %d, Recommended: %d, Advisory: %d, NoChange: %d, NotApplicable: %d",
		summary.OverallScore, len(summary.ItemsRequired), len(summary.ItemsRecommended), len(summary.ItemsAdvisory), summary.NoChangeCount, summary.NotApplicableCount)

	return summary, nil
}

// Helper function to count items for a specific category
func categoryItemCount(items map[string]int, category string) int {
	count := 0
	for cat, c := range items {
		if strings.Contains(cat, category) {
			count += c
		}
	}
	return count
}

// Enhanced item extraction from sections
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

	// Fallback based on status counts - exclude Not Applicable items
	required, recommended, advisory, noChange, _ := CountAllStatusItems(lines)
	total := required + recommended + advisory + noChange
	if total == 0 {
		return 75.0 // Default value if we can't calculate anything
	}

	// Weight calculation: Required=0%, Recommended=50%, Advisory=80%, NoChange=100%
	weightedSum := float64(noChange*100 + advisory*80 + recommended*50)
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
