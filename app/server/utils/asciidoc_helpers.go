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
	for _, line := range lines {
		if strings.Contains(line, "cluster") {
			// Look for quoted cluster name or after keywords
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
	return "OpenShift Cluster"
}

// ExtractCustomerName extracts the customer name from the report
func ExtractCustomerName(lines []string) string {
	for _, line := range lines {
		if strings.Contains(line, "conducted") && strings.Contains(line, "health check") {
			re := regexp.MustCompile(`conducted.*?([A-Za-z0-9_\s]+)'s`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				return strings.TrimSpace(matches[1])
			}
		}
	}
	return "Your Company"
}

// ExtractOverallScore extracts the overall score from the report
func ExtractOverallScore(lines []string) float64 {
	scorePattern := regexp.MustCompile(`Overall\s+Cluster\s+Health:\s+(\d+\.?\d*)%`)

	for _, line := range lines {
		matches := scorePattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			score, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				return score
			}
		}
	}

	// Check for a score in the health-check-report itself
	healthScorePattern := regexp.MustCompile(`Overall Health Score.*?(\d+\.?\d*)%`)
	for _, line := range lines {
		matches := healthScorePattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			score, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				return score
			}
		}
	}

	// If not found, try to calculate from status counts
	return CalculateScoreFromStatusCounts(lines)
}

// CalculateScoreFromStatusCounts calculates an approximate overall score based on status counts
func CalculateScoreFromStatusCounts(lines []string) float64 {
	totalItems := 0
	requiredChanges := 0
	recommendedChanges := 0
	advisory := 0
	noChanges := 0
	notApplicable := 0

	// Pattern to match status cells in the table
	statusCellPattern := regexp.MustCompile(`{set:cellbgcolor:(#[A-Fa-f0-9]+)}`)

	for _, line := range lines {
		if matches := statusCellPattern.FindStringSubmatch(line); len(matches) > 0 {
			colorCode := strings.ToUpper(matches[1])

			// Only count items that are actual evaluations (not section headers or notes)
			if strings.Contains(line, "Changes Required") ||
				strings.Contains(line, "Changes Recommended") ||
				strings.Contains(line, "N/A") ||
				strings.Contains(line, "Advisory") ||
				strings.Contains(line, "No Change") {
				continue // Skip the key description lines
			}

			totalItems++

			switch colorCode {
			case "#FF0000":
				requiredChanges++
			case "#FEFE20":
				recommendedChanges++
			case "#80E5FF":
				advisory++
			case "#00FF00":
				noChanges++
			case "#A6B9BF":
				notApplicable++
			}
		}
	}

	if totalItems == 0 {
		// Calculate from the results table
		requiredCount, recommendedCount, advisoryCount := CountStatusByColor(lines)
		totalRelevantItems := requiredCount + recommendedCount + advisoryCount + noChanges

		if totalRelevantItems == 0 {
			return 75.0 // Default fallback if no items found
		}

		// Weight calculation
		weightedSum := (noChanges * 100.0) + (advisoryCount * 80.0) + (recommendedCount * 50.0)
		return float64(weightedSum) / float64(totalRelevantItems)
	}

	// Calculate score based on weighted values
	// Required changes have the most negative impact, followed by recommended
	validItems := totalItems - notApplicable
	if validItems == 0 {
		return 100.0
	}

	// Weight: No changes = 100%, Advisory = 80%, Recommended = 50%, Required = 0%
	// Convert values to float64 to avoid integer division truncation
	weightedSum := float64(noChanges*100.0) + float64(advisory*80.0) + float64(recommendedChanges*50.0)
	return weightedSum / float64(validItems)
}

// CountStatusByColor counts items by their color status in the report
func CountStatusByColor(lines []string) (required, recommended, advisory int) {
	requiredPattern := regexp.MustCompile(`{set:cellbgcolor:#FF0000}`)
	recommendedPattern := regexp.MustCompile(`{set:cellbgcolor:#FEFE20}`)
	advisoryPattern := regexp.MustCompile(`{set:cellbgcolor:#80E5FF}`)

	for _, line := range lines {
		// Skip the key legend lines
		if strings.Contains(line, "Changes Required") && strings.Contains(line, "Description") {
			continue
		}
		if strings.Contains(line, "Changes Recommended") && strings.Contains(line, "Description") {
			continue
		}
		if strings.Contains(line, "Advisory") && strings.Contains(line, "Description") {
			continue
		}

		// Count actual items
		if requiredPattern.MatchString(line) {
			required++
		}
		if recommendedPattern.MatchString(line) {
			recommended++
		}
		if advisoryPattern.MatchString(line) {
			advisory++
		}
	}

	return required, recommended, advisory
}

// ExtractCategoryScore extracts the score for a specific category
func ExtractCategoryScore(lines []string, categoryName string) int {
	scorePattern := regexp.MustCompile(fmt.Sprintf(`\*%s\*:\s+(\d+)%%`, regexp.QuoteMeta(categoryName)))

	for _, line := range lines {
		matches := scorePattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			score, err := strconv.Atoi(matches[1])
			if err == nil {
				return score
			}
		}
	}

	// If not found with exact name, try partial matching
	return ExtractGeneralCategoryScore(lines, strings.Split(categoryName, " ")...)
}

// ExtractGeneralCategoryScore searches for a category score using keywords
func ExtractGeneralCategoryScore(lines []string, keywords ...string) int {
	// Default score
	defaultScore := 75

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
				score, err := strconv.Atoi(matches[1])
				if err == nil {
					return score
				}
			}
		}
	}

	return defaultScore
}

// ExtractCategoryDescription extracts the description for a specific category
func ExtractCategoryDescription(lines []string, categoryName string) string {
	var description string

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

	// If no description found, generate a generic one
	if description == "" {
		description = GenerateDescription(categoryName, ExtractCategoryScore(lines, categoryName))
	}

	return description
}

// GenerateDescription generates a generic description based on the category and score
func GenerateDescription(categoryName string, score int) string {
	if score >= 90 {
		return fmt.Sprintf("%s is excellent with best practices in place.", categoryName)
	} else if score >= 80 {
		return fmt.Sprintf("%s is well-configured with only minor improvements needed.", categoryName)
	} else if score >= 70 {
		return fmt.Sprintf("%s meets most requirements but has some areas that could be improved.", categoryName)
	} else if score >= 60 {
		return fmt.Sprintf("%s has several areas that need attention to meet best practices.", categoryName)
	} else {
		return fmt.Sprintf("%s requires significant improvements to ensure stability and security.", categoryName)
	}
}

// ExtractRequiredChanges extracts items marked as "Changes Required"
func ExtractRequiredChanges(lines []string) []string {
	var requiredItems []string
	var found bool

	// First, look for the table structure
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "Category") &&
			strings.Contains(lines[i], "Item Evaluated") &&
			strings.Contains(lines[i], "Observed Result") &&
			strings.Contains(lines[i], "Recommendation") {

			// Found the table header, now process rows
			for j := i + 1; j < len(lines); j++ {
				line := lines[j]

				// Check if this is a row with "Changes Required" status
				if strings.Contains(line, "{set:cellbgcolor:#FF0000}") && !strings.Contains(line, "Indicates Changes Required") {
					// This is a "Changes Required" row - extract the item name and observation

					// Find the Item Evaluated in the previous line(s)
					var itemName string
					for k := j - 1; k > j-5 && k >= 0; k-- {
						if strings.Contains(lines[k], "<<") && strings.Contains(lines[k], ">>") {
							// Extract text between << and >>
							matches := regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(lines[k])
							if len(matches) > 1 {
								itemName = strings.TrimSpace(matches[1])
								break
							}
						}
					}

					// Find the Observed Result
					var observation string
					for k := j - 1; k > j-5 && k >= 0; k-- {
						if !strings.Contains(lines[k], "<<") && !strings.Contains(lines[k], "Category") &&
							strings.TrimSpace(lines[k]) != "" && !strings.Contains(lines[k], "cellbgcolor") {
							observation = strings.TrimSpace(lines[k])
							break
						}
					}

					if itemName != "" {
						item := fmt.Sprintf("%s: %s", itemName, observation)
						requiredItems = append(requiredItems, item)
						found = true
					}
				}
			}

			// No need to continue if we found the table
			if found {
				break
			}
		}
	}

	// If we couldn't extract properly with the first method, try the fallback
	if !found {
		// Old method
		for i, line := range lines {
			if strings.Contains(line, "{set:cellbgcolor:#FF0000}") &&
				!strings.Contains(line, "Indicates Changes Required") {

				// Extract item name (usually in a <<item>> format)
				var itemName string
				itemMatches := regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(line)
				if len(itemMatches) > 1 {
					itemName = strings.TrimSpace(itemMatches[1])
				} else {
					// Try to look in adjacent lines
					for j := i - 3; j < i+3 && j < len(lines); j++ {
						if j >= 0 && strings.Contains(lines[j], "<<") {
							itemMatches = regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(lines[j])
							if len(itemMatches) > 1 {
								itemName = strings.TrimSpace(itemMatches[1])
								break
							}
						}
					}
				}

				// Extract observation (usually text after a pipe | character)
				var observation string
				parts := strings.Split(line, "|")
				if len(parts) >= 2 {
					for _, part := range parts {
						if !strings.Contains(part, "<<") &&
							!strings.Contains(part, "cellbgcolor") &&
							strings.TrimSpace(part) != "" {
							observation = strings.TrimSpace(part)
							break
						}
					}
				}

				// If we couldn't find a structured observation, use the next line
				if observation == "" && i+1 < len(lines) {
					observation = strings.TrimSpace(lines[i+1])
				}

				if itemName != "" {
					item := fmt.Sprintf("%s: %s", itemName, observation)
					requiredItems = append(requiredItems, item)
				}
			}
		}
	}

	return requiredItems
}

// ExtractRecommendedChanges extracts items marked as "Changes Recommended"
func ExtractRecommendedChanges(lines []string) []string {
	var recommendedItems []string
	var found bool

	// First, look for the table structure
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "Category") &&
			strings.Contains(lines[i], "Item Evaluated") &&
			strings.Contains(lines[i], "Observed Result") &&
			strings.Contains(lines[i], "Recommendation") {

			// Found the table header, now process rows
			for j := i + 1; j < len(lines); j++ {
				line := lines[j]

				// Check if this is a row with "Changes Recommended" status
				if strings.Contains(line, "{set:cellbgcolor:#FEFE20}") && !strings.Contains(line, "Indicates Changes Recommended") {
					// This is a "Changes Recommended" row - extract the item name and observation

					// Find the Item Evaluated in the previous line(s)
					var itemName string
					for k := j - 1; k > j-5 && k >= 0; k-- {
						if strings.Contains(lines[k], "<<") && strings.Contains(lines[k], ">>") {
							// Extract text between << and >>
							matches := regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(lines[k])
							if len(matches) > 1 {
								itemName = strings.TrimSpace(matches[1])
								break
							}
						}
					}

					// Find the Observed Result
					var observation string
					for k := j - 1; k > j-5 && k >= 0; k-- {
						if !strings.Contains(lines[k], "<<") && !strings.Contains(lines[k], "Category") &&
							strings.TrimSpace(lines[k]) != "" && !strings.Contains(lines[k], "cellbgcolor") {
							observation = strings.TrimSpace(lines[k])
							break
						}
					}

					if itemName != "" {
						item := fmt.Sprintf("%s: %s", itemName, observation)
						recommendedItems = append(recommendedItems, item)
						found = true
					}
				}
			}

			// No need to continue if we found the table
			if found {
				break
			}
		}
	}

	// If we couldn't extract properly with the first method, try the fallback
	if !found {
		// Old method
		for i, line := range lines {
			if strings.Contains(line, "{set:cellbgcolor:#FEFE20}") &&
				!strings.Contains(line, "Indicates Changes Recommended") {

				// Extract item name (usually in a <<item>> format)
				var itemName string
				itemMatches := regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(line)
				if len(itemMatches) > 1 {
					itemName = strings.TrimSpace(itemMatches[1])
				} else {
					// Try to look in adjacent lines
					for j := i - 3; j < i+3 && j < len(lines); j++ {
						if j >= 0 && strings.Contains(lines[j], "<<") {
							itemMatches = regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(lines[j])
							if len(itemMatches) > 1 {
								itemName = strings.TrimSpace(itemMatches[1])
								break
							}
						}
					}
				}

				// Extract observation (usually text after a pipe | character)
				var observation string
				parts := strings.Split(line, "|")
				if len(parts) >= 2 {
					for _, part := range parts {
						if !strings.Contains(part, "<<") &&
							!strings.Contains(part, "cellbgcolor") &&
							strings.TrimSpace(part) != "" {
							observation = strings.TrimSpace(part)
							break
						}
					}
				}

				// If we couldn't find a structured observation, use the next line
				if observation == "" && i+1 < len(lines) {
					observation = strings.TrimSpace(lines[i+1])
				}

				if itemName != "" {
					item := fmt.Sprintf("%s: %s", itemName, observation)
					recommendedItems = append(recommendedItems, item)
				}
			}
		}
	}

	return recommendedItems
}

// ExtractAdvisoryActions extracts items marked as "Advisory"
func ExtractAdvisoryActions(lines []string) []string {
	var advisoryItems []string
	var found bool

	// First, look for the table structure
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "Category") &&
			strings.Contains(lines[i], "Item Evaluated") &&
			strings.Contains(lines[i], "Observed Result") &&
			strings.Contains(lines[i], "Recommendation") {

			// Found the table header, now process rows
			for j := i + 1; j < len(lines); j++ {
				line := lines[j]

				// Check if this is a row with "Advisory" status
				if strings.Contains(line, "{set:cellbgcolor:#80E5FF}") && !strings.Contains(line, "No advise given") {
					// This is an "Advisory" row - extract the item name and observation

					// Find the Item Evaluated in the previous line(s)
					var itemName string
					for k := j - 1; k > j-5 && k >= 0; k-- {
						if strings.Contains(lines[k], "<<") && strings.Contains(lines[k], ">>") {
							// Extract text between << and >>
							matches := regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(lines[k])
							if len(matches) > 1 {
								itemName = strings.TrimSpace(matches[1])
								break
							}
						}
					}

					// Find the Observed Result
					var observation string
					for k := j - 1; k > j-5 && k >= 0; k-- {
						if !strings.Contains(lines[k], "<<") && !strings.Contains(lines[k], "Category") &&
							strings.TrimSpace(lines[k]) != "" && !strings.Contains(lines[k], "cellbgcolor") {
							observation = strings.TrimSpace(lines[k])
							break
						}
					}

					if itemName != "" {
						item := fmt.Sprintf("%s: %s", itemName, observation)
						advisoryItems = append(advisoryItems, item)
						found = true
					}
				}
			}

			// No need to continue if we found the table
			if found {
				break
			}
		}
	}

	// If we couldn't extract properly with the first method, try the fallback
	if !found {
		// Old method
		for i, line := range lines {
			if strings.Contains(line, "{set:cellbgcolor:#80E5FF}") &&
				!strings.Contains(line, "No advise given") {

				// Extract item name (usually in a <<item>> format)
				var itemName string
				itemMatches := regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(line)
				if len(itemMatches) > 1 {
					itemName = strings.TrimSpace(itemMatches[1])
				} else {
					// Try to look in adjacent lines
					for j := i - 3; j < i+3 && j < len(lines); j++ {
						if j >= 0 && strings.Contains(lines[j], "<<") {
							itemMatches = regexp.MustCompile(`<<([^>]+)>>`).FindStringSubmatch(lines[j])
							if len(itemMatches) > 1 {
								itemName = strings.TrimSpace(itemMatches[1])
								break
							}
						}
					}
				}

				// Extract observation (usually text after a pipe | character)
				var observation string
				parts := strings.Split(line, "|")
				if len(parts) >= 2 {
					for _, part := range parts {
						if !strings.Contains(part, "<<") &&
							!strings.Contains(part, "cellbgcolor") &&
							strings.TrimSpace(part) != "" {
							observation = strings.TrimSpace(part)
							break
						}
					}
				}

				// If we couldn't find a structured observation, use the next line
				if observation == "" && i+1 < len(lines) {
					observation = strings.TrimSpace(lines[i+1])
				}

				if itemName != "" {
					item := fmt.Sprintf("%s: %s", itemName, observation)
					advisoryItems = append(advisoryItems, item)
				}
			}
		}
	}

	return advisoryItems
}

// ExtractActionItems extracts action items from a specific section
func ExtractActionItems(lines []string, sectionName string) []string {
	// First try to use the new extraction methods based on the status color
	if sectionName == "Changes Required" {
		requiredItems := ExtractRequiredChanges(lines)
		if len(requiredItems) > 0 {
			return requiredItems
		}
	} else if sectionName == "Changes Recommended" {
		recommendedItems := ExtractRecommendedChanges(lines)
		if len(recommendedItems) > 0 {
			return recommendedItems
		}
	} else if sectionName == "Advisory Actions" {
		advisoryItems := ExtractAdvisoryActions(lines)
		if len(advisoryItems) > 0 {
			return advisoryItems
		}
	}

	// Fallback to the original method if the new methods don't return any results
	var items []string
	inSection := false

	// Look for section header and collect items
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for section start
		if strings.Contains(trimmed, sectionName) && !inSection {
			inSection = true
			continue
		}

		// Check for end of section (usually another section header)
		if inSection && (strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "=")) &&
			!strings.HasPrefix(trimmed, "- ") && !strings.HasPrefix(trimmed, "* ") {
			break
		}

		// Collect items that look like list entries
		if inSection && (strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ")) {
			item := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* "))
			if item != "" {
				items = append(items, item)
			}
		}
	}

	return items
}

// ExtractIssuesBySeverity extracts issues based on their severity
func ExtractIssuesBySeverity(lines []string, severities ...string) []string {
	var items []string

	for _, line := range lines {
		lowercase := strings.ToLower(line)

		// Check if line contains any severity marker
		found := false
		for _, severity := range severities {
			if strings.Contains(lowercase, strings.ToLower(severity)) {
				found = true
				break
			}
		}

		if found {
			// Extract the issue description - look for a colon or just take the whole line
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
				items = append(items, strings.TrimSpace(parts[1]))
			} else if strings.TrimSpace(line) != "" {
				items = append(items, strings.TrimSpace(line))
			}
		}
	}

	return items
}
