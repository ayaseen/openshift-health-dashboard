// app/server/utils/asciidoc_parser.go
package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ayaseen/openshift-health-dashboard/app/server/server"
	"github.com/ayaseen/openshift-health-dashboard/app/server/types"
	"github.com/ayaseen/openshift-health-dashboard/app/server/utils"
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

	// If not found, try to calculate from status counts
	return CalculateOverallScore(lines)
}

// CalculateOverallScore calculates an approximate overall score based on status counts
func CalculateOverallScore(lines []string) float64 {
	// Default score if we can't calculate
	defaultScore := 75.0

	// Count statuses
	ok := 0
	warning := 0
	critical := 0
	unknown := 0

	for _, line := range lines {
		lowercase := strings.ToLower(line)

		// Count based on common status keywords
		if strings.Contains(lowercase, "status: ok") ||
			strings.Contains(lowercase, "status: healthy") ||
			strings.Contains(lowercase, "status: pass") {
			ok++
		} else if strings.Contains(lowercase, "status: warning") ||
			strings.Contains(lowercase, "warning") {
			warning++
		} else if strings.Contains(lowercase, "status: critical") ||
			strings.Contains(lowercase, "status: error") ||
			strings.Contains(lowercase, "status: fail") {
			critical++
		} else if strings.Contains(lowercase, "status: unknown") {
			unknown++
		}
	}

	// If we found some statuses, calculate the score
	total := ok + warning + critical + unknown
	if total > 0 {
		// Weighted score: 100% for OK, 70% for Warning, 0% for Critical
		return (float64(ok)*100.0 + float64(warning)*70.0) / float64(total)
	}

	return defaultScore
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

// ExtractActionItems extracts action items from a specific section
func ExtractActionItems(lines []string, sectionName string) []string {
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
