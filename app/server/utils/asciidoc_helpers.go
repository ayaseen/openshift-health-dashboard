// app/server/utils/asciidoc_helpers.go
package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// IsValidAsciiDocFile checks if a filename has a valid AsciiDoc extension
func IsValidAsciiDocFile(filename string) bool {
	return strings.HasSuffix(filename, ".adoc") || strings.HasSuffix(filename, ".asciidoc")
}

// Helper functions for extracting data from AsciiDoc content

// ExtractClusterName extracts the cluster name from the report
func ExtractClusterName(lines []string) string {
	clusterName := ""

	for _, line := range lines {
		if strings.Contains(line, "cluster") {
			// Look for quoted cluster name or after keywords
			re := regexp.MustCompile(`['"]([^'"]+)['"]|cluster\s+([a-zA-Z0-9_-]+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if matches[1] != "" {
					clusterName = matches[1]
					break
				}
				if len(matches) > 2 && matches[2] != "" {
					clusterName = matches[2]
					break
				}
			}
		}
	}

	return clusterName
}

// ExtractCustomerName extracts the customer name from the report
func ExtractCustomerName(lines []string) string {
	customerName := ""

	for _, line := range lines {
		if strings.Contains(line, "conducted") && strings.Contains(line, "health check") {
			re := regexp.MustCompile(`conducted.*?([A-Za-z0-9_\s]+)'s`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				customerName = strings.TrimSpace(matches[1])
				break
			}
		}
	}

	return customerName
}

// ExtractOverallScore extracts the overall score from the report
func ExtractOverallScore(lines []string) float64 {
	var score float64

	scorePattern := regexp.MustCompile(`Overall\s+Cluster\s+Health:\s+(\d+\.?\d*)%`)
	for _, line := range lines {
		matches := scorePattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			score, _ = strconv.ParseFloat(matches[1], 64)
			return score
		}
	}

	// Check for a score in the health-check-report itself
	healthScorePattern := regexp.MustCompile(`Overall Health Score.*?(\d+\.?\d*)%`)
	for _, line := range lines {
		matches := healthScorePattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			score, _ = strconv.ParseFloat(matches[1], 64)
			return score
		}
	}

	// If no explicit score is found, calculate from status counts in the Summary section
	return CalculateScoreFromStatusCounts(lines)
}

// ItemsByCategory represents items grouped by category and status
type ItemsByCategory struct {
	Required      map[string]int
	Recommended   map[string]int
	Advisory      map[string]int
	NoChange      map[string]int
	NotApplicable map[string]int
}

// CalculateScoreFromStatusCounts calculates score based on status counts in Summary section
func CalculateScoreFromStatusCounts(lines []string) float64 {
	required, recommended, advisory, noChange, _ := CountAllStatusItems(lines)

	// Calculate score if we have valid items
	totalItems := required + recommended + advisory + noChange
	if totalItems == 0 {
		return 0.0
	}

	// Weight calculation based on status counts
	// Required = 0%, Recommended = 50%, Advisory = 80%, No Change = 100%
	weightedSum := float64(noChange*100 + advisory*80 + recommended*50)
	return weightedSum / float64(totalItems)
}

// CountAllStatusItems counts items by their color status in the Summary section
// Returns counts for required, recommended, advisory, noChange, and notApplicable
func CountAllStatusItems(lines []string) (int, int, int, int, int) {
	required := 0
	recommended := 0
	advisory := 0
	noChange := 0
	notApplicable := 0

	// Find summary section boundaries
	summaryStartIndex := -1
	summaryEndIndex := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == "= Summary" {
			summaryStartIndex = i
			break
		}
	}

	if summaryStartIndex == -1 {
		return 0, 0, 0, 0, 0 // Can't find summary section
	}

	// Find end of summary (next section or end of file)
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

	// Process the summary section for color codes
	inItem := false
	inTable := false

	for i := summaryStartIndex; i < summaryEndIndex; i++ {
		line := lines[i]

		// Start of an item block (if using item blocks)
		if strings.Contains(line, "// ------------------------ITEM START") {
			inItem = true
			continue
		}

		// End of an item block
		if strings.Contains(line, "// ------------------------ITEM END") {
			inItem = false
			continue
		}

		// Detect start of table
		if strings.Contains(line, "|===") && !inTable {
			inTable = true
			continue
		}

		// Detect end of table
		if strings.Contains(line, "|===") && inTable {
			inTable = false
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

		// Count by color codes - for table cells and items
		if (inTable || inItem) && !strings.Contains(line, "Description") {
			if strings.Contains(line, "{set:cellbgcolor:#FF0000}") {
				required++
			} else if strings.Contains(line, "{set:cellbgcolor:#FEFE20}") {
				recommended++
			} else if strings.Contains(line, "{set:cellbgcolor:#80E5FF}") {
				advisory++
			} else if strings.Contains(line, "{set:cellbgcolor:#00FF00}") {
				noChange++
			} else if strings.Contains(line, "{set:cellbgcolor:#A6B9BF}") {
				notApplicable++
			}
		}
	}

	return required, recommended, advisory, noChange, notApplicable
}

// CountStatusByCategory counts items by category and status
func CountStatusByCategory(lines []string) *ItemsByCategory {
	result := &ItemsByCategory{
		Required:      make(map[string]int),
		Recommended:   make(map[string]int),
		Advisory:      make(map[string]int),
		NoChange:      make(map[string]int),
		NotApplicable: make(map[string]int),
	}

	// Find summary section boundaries
	summaryStartIndex := -1
	summaryEndIndex := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == "= Summary" {
			summaryStartIndex = i
			break
		}
	}

	if summaryStartIndex == -1 {
		return result // Can't find summary section
	}

	// Find end of summary (next section or end of file)
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

	// Process the summary section
	var currentCategory string
	var currentStatus string
	inTable := false

	for i := summaryStartIndex; i < summaryEndIndex; i++ {
		line := strings.TrimSpace(lines[i])

		// Detect start/end of table
		if strings.Contains(line, "|===") {
			inTable = !inTable
			continue
		}

		if !inTable {
			continue
		}

		// Extract category
		if strings.HasPrefix(line, "|") && !strings.Contains(line, "cellbgcolor") {
			currentCategory = strings.TrimSpace(strings.TrimPrefix(line, "|"))
			continue
		}

		// Determine status by color code
		if strings.Contains(line, "{set:cellbgcolor:#FF0000}") {
			currentStatus = "required"
		} else if strings.Contains(line, "{set:cellbgcolor:#FEFE20}") {
			currentStatus = "recommended"
		} else if strings.Contains(line, "{set:cellbgcolor:#80E5FF}") {
			currentStatus = "advisory"
		} else if strings.Contains(line, "{set:cellbgcolor:#00FF00}") {
			currentStatus = "nochange"
		} else if strings.Contains(line, "{set:cellbgcolor:#A6B9BF}") {
			currentStatus = "notapplicable"
		}

		// Only count if we have both category and status
		if currentCategory != "" && currentStatus != "" {
			// Skip header/legend rows
			if strings.Contains(line, "Indicates Changes Required") ||
				strings.Contains(line, "Indicates Changes Recommended") ||
				strings.Contains(line, "No advise given") ||
				strings.Contains(line, "No change required") ||
				strings.Contains(line, "Not yet evaluated") {
				currentStatus = ""
				continue
			}

			// Count item by category and status
			switch currentStatus {
			case "required":
				result.Required[currentCategory]++
			case "recommended":
				result.Recommended[currentCategory]++
			case "advisory":
				result.Advisory[currentCategory]++
			case "nochange":
				result.NoChange[currentCategory]++
			case "notapplicable":
				result.NotApplicable[currentCategory]++
			}

			// Reset status to avoid double counting
			currentStatus = ""
		}
	}

	return result
}

// CalculateCategoryScore calculates score for a given category using item counts
func CalculateCategoryScore(categoryItems map[string]int, categoryName string) int {
	required := categoryItems["required"]
	recommended := categoryItems["recommended"]
	advisory := categoryItems["advisory"]
	noChange := categoryItems["nochange"]

	// Calculate total items (excluding Not Applicable)
	totalItems := required + recommended + advisory + noChange
	if totalItems == 0 {
		return 0
	}

	// Weight calculation:
	// Required = 0%, Recommended = 50%, Advisory = 80%, No Change = 100%
	weightedSum := float64(noChange*100 + advisory*80 + recommended*50)
	score := int(weightedSum / float64(totalItems))

	return score
}

// ExtractCategoryScore extracts the score for a specific category
func ExtractCategoryScore(lines []string, categoryName string) int {
	var score int

	scorePattern := regexp.MustCompile(fmt.Sprintf(`\*%s\*:\s+(\d+)%%`, regexp.QuoteMeta(categoryName)))
	for _, line := range lines {
		matches := scorePattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			score, _ = strconv.Atoi(matches[1])
			return score
		}
	}

	// If not found with exact name, try partial matching
	return ExtractGeneralCategoryScore(lines, strings.Split(categoryName, " ")...)
}

// ExtractGeneralCategoryScore searches for a category score using keywords
func ExtractGeneralCategoryScore(lines []string, keywords ...string) int {
	var score int

	// Search for lines containing any of the keywords and a percentage
	percentPattern := regexp.MustCompile(`(\d+)%`)

	for _, line := range lines {
		lowercase := strings.ToLower(line)

		// Check if line contains any keyword
		foundKeyword := false
		for _, keyword := range keywords {
			if strings.Contains(lowercase, strings.ToLower(keyword)) {
				foundKeyword = true
				break
			}
		}

		if foundKeyword {
			// Look for a percentage
			matches := percentPattern.FindStringSubmatch(line)
			if len(matches) > 1 {
				score, _ = strconv.Atoi(matches[1])
				return score
			}
		}
	}

	return score
}

// ExtractCategoryDescription extracts the description for a specific category
func ExtractCategoryDescription(lines []string, categoryName string) string {
	description := ""

	// Look for lines containing the category name followed by a description
	for i, line := range lines {
		// Check for category section
		if strings.Contains(line, categoryName) {
			// Look for description in next few lines
			for j := i + 1; j < i+10 && j < len(lines); j++ {
				// Skip empty lines
				if lines[j] == "" {
					continue
				}

				// Skip lines that look like headers or contain percentages
				if strings.HasPrefix(lines[j], "*") || strings.HasPrefix(lines[j], "#") ||
					strings.Contains(lines[j], "%") {
					continue
				}

				// Found a description
				description = lines[j]
				break
			}

			if description != "" {
				break
			}
		}
	}

	return description
}

// GenerateDescription generates a description based on the category and score
func GenerateDescription(categoryName string, score int) string {
	if score >= 90 {
		return fmt.Sprintf("%s is excellent with best practices in place.", categoryName)
	} else if score >= 80 {
		return fmt.Sprintf("%s is well-configured with only minor improvements needed.", categoryName)
	} else if score >= 70 {
		return fmt.Sprintf("%s meets most requirements but has some areas that could be improved.", categoryName)
	} else if score >= 60 {
		return fmt.Sprintf("%s has several areas that need attention to meet best practices.", categoryName)
	} else if score > 0 {
		return fmt.Sprintf("%s requires significant improvements to ensure stability and security.", categoryName)
	}

	return ""
}

// ExtractRequiredChanges extracts items marked as "Changes Required" from Summary section
func ExtractRequiredChanges(lines []string) []string {
	var requiredItems []string

	// Find summary section boundaries first
	summaryStartIndex := -1
	summaryEndIndex := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == "= Summary" {
			summaryStartIndex = i
			break
		}
	}

	if summaryStartIndex == -1 {
		return requiredItems // Empty list, summary not found
	}

	// Find end of summary (next section or end of file)
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

	// Now process only the lines in the Summary section
	summaryLines := lines[summaryStartIndex:summaryEndIndex]

	// Find ITEM blocks in the Summary section
	var currentItem string
	var itemName string
	var observation string
	inItem := false

	for _, line := range summaryLines {
		// Detect ITEM start
		if strings.Contains(line, "// ------------------------ITEM START") {
			inItem = true
			itemName = ""
			observation = ""
			continue
		}

		// Detect ITEM end
		if strings.Contains(line, "// ------------------------ITEM END") {
			if inItem && itemName != "" {
				if observation != "" {
					currentItem = fmt.Sprintf("%s: %s", itemName, observation)
				} else {
					currentItem = itemName
				}

				if currentItem != "" {
					requiredItems = append(requiredItems, currentItem)
				}
			}
			inItem = false
			continue
		}

		if !inItem {
			continue
		}

		// Extract item name
		if strings.Contains(line, "<<") && strings.Contains(line, ">>") {
			re := regexp.MustCompile(`<<([^>]+)>>`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				itemName = strings.TrimSpace(matches[1])
			}
			continue
		}

		// Extract observation
		if itemName != "" && observation == "" &&
			!strings.HasPrefix(line, "//") && !strings.Contains(line, "{set:cellbgcolor") {
			if strings.HasPrefix(line, "|") {
				line = strings.TrimSpace(line[1:])
			}
			if line != "" {
				observation = line
			}
			continue
		}

		// Check for required status
		if strings.Contains(line, "{set:cellbgcolor:#FF0000}") &&
			!strings.Contains(line, "Indicates Changes Required") {
			// This is a "Changes Required" item - keep it in the list
			continue
		} else if strings.Contains(line, "set:cellbgcolor:") {
			// This item has a different status - remove it from consideration
			inItem = false
		}
	}

	return requiredItems
}

// ExtractRecommendedChanges extracts items marked as "Changes Recommended" from Summary section
func ExtractRecommendedChanges(lines []string) []string {
	var recommendedItems []string

	// Find summary section boundaries first
	summaryStartIndex := -1
	summaryEndIndex := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == "= Summary" {
			summaryStartIndex = i
			break
		}
	}

	if summaryStartIndex == -1 {
		return recommendedItems // Empty list, summary not found
	}

	// Find end of summary (next section or end of file)
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

	// Now process only the lines in the Summary section
	summaryLines := lines[summaryStartIndex:summaryEndIndex]

	// Find ITEM blocks in the Summary section
	var currentItem string
	var itemName string
	var observation string
	inItem := false

	for _, line := range summaryLines {
		// Detect ITEM start
		if strings.Contains(line, "// ------------------------ITEM START") {
			inItem = true
			itemName = ""
			observation = ""
			continue
		}

		// Detect ITEM end
		if strings.Contains(line, "// ------------------------ITEM END") {
			if inItem && itemName != "" {
				if observation != "" {
					currentItem = fmt.Sprintf("%s: %s", itemName, observation)
				} else {
					currentItem = itemName
				}

				if currentItem != "" {
					recommendedItems = append(recommendedItems, currentItem)
				}
			}
			inItem = false
			continue
		}

		if !inItem {
			continue
		}

		// Extract item name
		if strings.Contains(line, "<<") && strings.Contains(line, ">>") {
			re := regexp.MustCompile(`<<([^>]+)>>`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				itemName = strings.TrimSpace(matches[1])
			}
			continue
		}

		// Extract observation
		if itemName != "" && observation == "" &&
			!strings.HasPrefix(line, "//") && !strings.Contains(line, "{set:cellbgcolor") {
			if strings.HasPrefix(line, "|") {
				line = strings.TrimSpace(line[1:])
			}
			if line != "" {
				observation = line
			}
			continue
		}

		// Check for recommended status
		if strings.Contains(line, "{set:cellbgcolor:#FEFE20}") &&
			!strings.Contains(line, "Indicates Changes Recommended") {
			// This is a "Changes Recommended" item - keep it in the list
			continue
		} else if strings.Contains(line, "set:cellbgcolor:") {
			// This item has a different status - remove it from consideration
			inItem = false
		}
	}

	return recommendedItems
}

// ExtractAdvisoryActions extracts items marked as "Advisory" from Summary section
func ExtractAdvisoryActions(lines []string) []string {
	var advisoryItems []string

	// Find summary section boundaries first
	summaryStartIndex := -1
	summaryEndIndex := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == "= Summary" {
			summaryStartIndex = i
			break
		}
	}

	if summaryStartIndex == -1 {
		return advisoryItems // Empty list, summary not found
	}

	// Find end of summary (next section or end of file)
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

	// Now process only the lines in the Summary section
	summaryLines := lines[summaryStartIndex:summaryEndIndex]

	// Find ITEM blocks in the Summary section
	var currentItem string
	var itemName string
	var observation string
	inItem := false

	for _, line := range summaryLines {
		// Detect ITEM start
		if strings.Contains(line, "// ------------------------ITEM START") {
			inItem = true
			itemName = ""
			observation = ""
			continue
		}

		// Detect ITEM end
		if strings.Contains(line, "// ------------------------ITEM END") {
			if inItem && itemName != "" {
				if observation != "" {
					currentItem = fmt.Sprintf("%s: %s", itemName, observation)
				} else {
					currentItem = itemName
				}

				if currentItem != "" {
					advisoryItems = append(advisoryItems, currentItem)
				}
			}
			inItem = false
			continue
		}

		if !inItem {
			continue
		}

		// Extract item name
		if strings.Contains(line, "<<") && strings.Contains(line, ">>") {
			re := regexp.MustCompile(`<<([^>]+)>>`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				itemName = strings.TrimSpace(matches[1])
			}
			continue
		}

		// Extract observation
		if itemName != "" && observation == "" &&
			!strings.HasPrefix(line, "//") && !strings.Contains(line, "{set:cellbgcolor") {
			if strings.HasPrefix(line, "|") {
				line = strings.TrimSpace(line[1:])
			}
			if line != "" {
				observation = line
			}
			continue
		}

		// Check for advisory status
		if strings.Contains(line, "{set:cellbgcolor:#80E5FF}") &&
			!strings.Contains(line, "No advise given") {
			// This is an "Advisory" item - keep it in the list
			continue
		} else if strings.Contains(line, "set:cellbgcolor:") {
			// This item has a different status - remove it from consideration
			inItem = false
		}
	}

	return advisoryItems
}

// CountNoChangeItems counts items marked as "No Change" in the Summary section
func CountNoChangeItems(lines []string) int {
	count := 0

	// Find summary section boundaries
	summaryStartIndex := -1
	summaryEndIndex := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == "= Summary" {
			summaryStartIndex = i
			break
		}
	}

	if summaryStartIndex == -1 {
		return count // Summary not found
	}

	// Find end of summary
	for i := summaryStartIndex + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "=") &&
			!strings.Contains(lines[i], "= Summary") {
			summaryEndIndex = i
			break
		}
	}

	if summaryEndIndex == -1 {
		summaryEndIndex = len(lines)
	}

	// Process summary section
	inItem := false
	inTable := false

	for i := summaryStartIndex; i < summaryEndIndex; i++ {
		line := lines[i]

		// Check for item blocks
		if strings.Contains(line, "// ------------------------ITEM START") {
			inItem = true
			continue
		}
		if strings.Contains(line, "// ------------------------ITEM END") {
			inItem = false
			continue
		}

		// Check for table
		if strings.Contains(line, "|===") {
			inTable = !inTable
			continue
		}

		// Count "No Change" items
		if (inTable || inItem) && strings.Contains(line, "{set:cellbgcolor:#00FF00}") &&
			!strings.Contains(line, "No change required") {
			count++
		}
	}

	return count
}
